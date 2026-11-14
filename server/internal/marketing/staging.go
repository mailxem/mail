package marketing

import (
	"fmt"
	"os"
	"time"

	"gorm.io/gorm"
	"kori/internal/models"
)

// Staged sending protects the shared managed-sending reputation from lists
// nobody has mailed yet. Readers the workspace has emailed before go out
// normally; never-emailed readers go out as a random 2%, then 5% more, then
// the rest, with a wait before each later stage to check bounces and
// complaints. Sends with fewer new readers than StagedMinNew go out at once.
const (
	StagedMinNew       = 500
	firstStagePct      = 2
	secondStagePct     = 7  // cumulative: 2% + 5%
	maxHardBouncePerK  = 25 // pause past 2.5% hard bounces...
	maxComplaintPer10K = 8  // ...or past 0.08% complaints
)

// reachedStatuses are emails that actually reached a reader.
var reachedStatuses = []string{"SENT", "DELIVERED", "OPENED", "CLICKED"}

// StageWait is how long results settle before the next stage
// (STAGED_SENDING_WAIT, a Go duration; default 30 minutes).
func StageWait() time.Duration {
	if d, err := time.ParseDuration(os.Getenv("STAGED_SENDING_WAIT")); err == nil && d > 0 {
		return d
	}
	return 30 * time.Minute
}

// emailedBefore lists addresses the workspace reached outside send c.
// ponytail: scans the team's emails per call; add a per-team address table if that gets slow.
func emailedBefore(db *gorm.DB, c *models.Campaign) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Table("emails").Select(`LOWER("to")`).
		Where("team_id = ? AND (campaign_id IS NULL OR campaign_id <> ?) AND status IN ?", c.TeamID, c.ID, reachedStatuses)
}

// Recipients returns the active contacts of send c that it has not emailed
// yet, split into readers the workspace emailed before and never-emailed ones.
func Recipients(db *gorm.DB, c *models.Campaign) (old, fresh *gorm.DB) {
	base := func() *gorm.DB {
		q := db.Session(&gorm.Session{NewDB: true}).Model(&models.Contact{}).
			Where("team_id = ? AND list_id = ? AND status = ? AND is_deleted = ?", c.TeamID, c.ListID, models.SubscriberStatusActive, false).
			Where("id NOT IN (?)", db.Session(&gorm.Session{NewDB: true}).Table("emails").Select("contact_id").Where("campaign_id = ? AND contact_id IS NOT NULL", c.ID))
		if c.AudienceCutoff != nil {
			q = q.Where("created_at <= ?", *c.AudienceCutoff)
		}
		if c.AudienceCampaignID != nil {
			q = followUp(db, q, c)
		}
		return q.Session(&gorm.Session{})
	}
	return base().Where("LOWER(email) IN (?)", emailedBefore(db, c)), base().Where("LOWER(email) NOT IN (?)", emailedBefore(db, c))
}

// followUp narrows q to the readers send c's source campaign reached who did
// not open (or click) it. Opens and clicks count as reports do: people only,
// not bots or privacy proxies. Suppressed addresses are left out too.
func followUp(db *gorm.DB, q *gorm.DB, c *models.Campaign) *gorm.DB {
	events := []models.EmailTrackingEvent{models.EmailTrackingEventClick}
	if c.AudienceFilter != models.AudienceNotClicked {
		events = append(events, models.EmailTrackingEventOpen)
	}
	sub := func() *gorm.DB { return db.Session(&gorm.Session{NewDB: true}) }
	reached := sub().Table("emails").Select("contact_id").
		Where("team_id = ? AND campaign_id = ? AND test = ? AND status IN ? AND contact_id IS NOT NULL", c.TeamID, *c.AudienceCampaignID, false, reachedStatuses)
	engaged := sub().Table("emails e").Joins("JOIN email_trackings t ON t.email_id = e.id").Select("e.contact_id").
		Where("e.team_id = ? AND e.campaign_id = ? AND e.contact_id IS NOT NULL AND t.event IN ? AND t.machine = '' AND t.is_deleted = ?", c.TeamID, *c.AudienceCampaignID, events, false)
	suppressed := sub().Model(&models.SuppressionList{}).Select("LOWER(email_address)").
		Where("team_id = ? AND is_active = ? AND is_deleted = ? AND (expires_at IS NULL OR expires_at > ?)", c.TeamID, true, false, time.Now())
	return q.Where("id IN (?)", reached).Where("id NOT IN (?)", engaged).Where("LOWER(email) NOT IN (?)", suppressed)
}

// StagePlan is how a staged send splits its audience: returning readers go
// first, never-emailed readers in three stages ending at Bounds and New.
type StagePlan struct {
	ReturningLeft int64  `json:"returningLeft"` // returning readers not emailed yet
	New           int64  `json:"new"`           // never-emailed readers in all
	SentToNew     int64  `json:"sentToNew"`
	Bounds        [2]int `json:"bounds"` // cumulative new-reader sends after stages 1 and 2
}

// StageProgress reports c's stage plan and how far it has got.
func StageProgress(db *gorm.DB, c *models.Campaign) (p StagePlan, err error) {
	old, fresh := Recipients(db, c)
	if err = old.Count(&p.ReturningLeft).Error; err != nil {
		return
	}
	var left int64
	if err = fresh.Count(&left).Error; err != nil {
		return
	}
	err = db.Session(&gorm.Session{NewDB: true}).Table("emails e").Where(`e.campaign_id = ? AND e.test = false AND LOWER(e."to") NOT IN (?)`, c.ID, emailedBefore(db, c)).Count(&p.SentToNew).Error
	p.New = left + p.SentToNew
	p.Bounds = stageBounds(int(p.New))
	return
}

// stageBounds are the cumulative never-emailed sends after stages 1 and 2.
func stageBounds(total int) [2]int {
	first := max(1, total*firstStagePct/100)
	return [2]int{first, max(first+1, total*secondStagePct/100)}
}

// stageHealth returns why a staged send should pause, or "" to continue.
func stageHealth(sent, bounced, complained int64) string {
	if bounced*1000 > maxHardBouncePerK*sent {
		return fmt.Sprintf("Paused: %.1f%% of the first %d emails bounced. Check this list before sending to the rest.", float64(bounced)*100/float64(sent), sent)
	}
	if complained*10000 > maxComplaintPer10K*sent {
		return fmt.Sprintf("Paused: %.2f%% of the first %d emails were reported as spam. Check that this list opted in before sending to the rest.", float64(complained)*100/float64(sent), sent)
	}
	return ""
}

// Stage decides how many never-emailed readers send c may email now: -1 for
// no limit, otherwise that many (0 with wait > 0: come back after wait). It
// saves c's staging state, and pauses c and alerts the owner when the stages
// sent so far bounced or drew complaints. managed says the sender is Xem's
// managed sending, the only one staged automatically.
func Stage(db *gorm.DB, c *models.Campaign, managed bool, now time.Time) (limit int, wait time.Duration, err error) {
	if c.Schedule == models.CampaignScheduleRecurring || (c.SendInStages != nil && !*c.SendInStages) {
		return -1, 0, nil
	}
	_, fresh := Recipients(db, c)
	var remaining, sent int64
	if err = fresh.Count(&remaining).Error; err != nil {
		return
	}
	sentToNew := db.Session(&gorm.Session{NewDB: true}).Table("emails e").Where(`e.campaign_id = ? AND e.test = false AND LOWER(e."to") NOT IN (?)`, c.ID, emailedBefore(db, c)).Session(&gorm.Session{})
	if err = sentToNew.Count(&sent).Error; err != nil {
		return
	}
	total := int(remaining + sent)
	if c.SendInStages == nil {
		staged := managed && total >= StagedMinNew
		c.SendInStages = &staged
		if err = db.Model(c).Update("send_in_stages", staged).Error; err != nil || !staged {
			return -1, 0, err
		}
	}
	if remaining == 0 {
		return -1, 0, nil
	}
	bounds := stageBounds(total)
	reached := 0
	for _, b := range bounds {
		if int(sent) >= b {
			reached++
		}
	}
	if reached > c.Stage {
		// A stage finished: wait, then check its results before the next.
		if c.NextStageAt == nil {
			at := now.Add(StageWait())
			c.NextStageAt = &at
			return 0, StageWait(), db.Model(c).Update("next_stage_at", at).Error
		}
		if now.Before(*c.NextStageAt) {
			return 0, c.NextStageAt.Sub(now), nil
		}
		var r struct{ Waiting, Bounced, Complained int64 }
		if err = sentToNew.Joins("LEFT JOIN contacts ON contacts.id = e.contact_id").Select(`
			SUM(CASE WHEN e.status IN ('PENDING','QUEUED','SENDING') THEN 1 ELSE 0 END) AS waiting,
			SUM(CASE WHEN e.status = 'BOUNCED' AND contacts.status = 'BOUNCED' THEN 1 ELSE 0 END) AS bounced,
			SUM(CASE WHEN e.status = 'COMPLAINED' THEN 1 ELSE 0 END) AS complained`).Scan(&r).Error; err != nil {
			return
		}
		if r.Waiting*10 > sent { // judge a stage once most of it has an outcome
			at := now.Add(StageWait())
			c.NextStageAt = &at
			return 0, StageWait(), db.Model(c).Update("next_stage_at", at).Error
		}
		if reason := stageHealth(sent, r.Bounced, r.Complained); reason != "" {
			c.Status, c.PauseReason, c.NextStageAt = models.CampaignStatusPaused, reason, nil
			return 0, 0, db.Transaction(func(tx *gorm.DB) error {
				if e := tx.Model(c).Updates(map[string]any{"status": c.Status, "pause_reason": reason, "next_stage_at": nil}).Error; e != nil {
					return e
				}
				return models.QueueAlert(tx, c.TeamID, "staged-sending", c.ID, "sending_paused", reason, now)
			})
		}
		c.Stage, c.NextStageAt = reached, nil
		if err = db.Model(c).Updates(map[string]any{"stage": reached, "next_stage_at": nil}).Error; err != nil {
			return
		}
	}
	if c.Stage >= len(bounds) {
		return -1, 0, nil
	}
	return max(0, bounds[c.Stage]-int(sent)), 0, nil
}
