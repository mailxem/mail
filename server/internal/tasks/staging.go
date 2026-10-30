package tasks

import (
	"fmt"
	"os"
	"time"

	"gorm.io/gorm"
	"kori/internal/models"
)

// Staged sending protects a sender's reputation on lists that have never been
// mailed: a small first stage, a wait for bounces and complaints, then larger
// stages only while the results stay healthy.
const (
	stagedMinList  = 500  // smaller lists send in one go
	stagedMinFirst = 100  // first stage: 5% of the list, at least 100
	stagedMinMid   = 500  // second stage: 25% of the list, at least 500
	healthMinSends = 50   // judge failure rates only on a meaningful sample
	maxFailurePct  = 5    // pause at 5% failed or bounced
	maxComplaints  = 1000 // pause at 1 complaint per 1,000 (0.1%)...
	minComplaints  = 2    // ...and at least 2 complaints
)

// stageWait is how long results settle between stages (STAGED_SENDING_WAIT,
// a Go duration; default one hour so late bounces and complaints arrive).
func stageWait() time.Duration {
	if d, err := time.ParseDuration(os.Getenv("STAGED_SENDING_WAIT")); err == nil && d > 0 {
		return d
	}
	return time.Hour
}

// stageTarget returns how many contacts should have been sent once the next
// stage finishes, and whether that stage is the last.
func stageTarget(total, sent int) (target int, last bool) {
	for _, t := range []int{max(stagedMinFirst, total*5/100), max(stagedMinMid, total*25/100)} {
		if t > sent && t < total {
			return t, false
		}
	}
	return total, true
}

// stageHealth returns why a staged campaign should pause, or "" to continue.
func stageHealth(sent, failed, complained int64) string {
	if sent >= healthMinSends && failed*100 >= maxFailurePct*sent {
		return fmt.Sprintf("Paused after %d of %d emails failed or bounced (%.1f%%). Check the list's source and the sender before resuming.", failed, sent, float64(failed)*100/float64(sent))
	}
	if complained >= minComplaints && complained*maxComplaints >= sent {
		return fmt.Sprintf("Paused after %d of %d recipients reported the email as spam. Check that this list opted in before resuming.", complained, sent)
	}
	return ""
}

// useStages decides whether a campaign sends in stages: explicitly when set,
// otherwise for one-off campaigns to a large list that has never been mailed.
func useStages(db *gorm.DB, c *models.Campaign, total int) bool {
	if c.Schedule == models.CampaignScheduleRecurring || total <= stagedMinList {
		return false
	}
	if c.SendInStages != nil {
		return *c.SendInStages
	}
	var earlier int64
	db.Model(&models.Campaign{}).Where("list_id = ? AND id <> ? AND status = ? AND is_deleted = ?", c.ListID, c.ID, models.CampaignStatusCompleted, false).Count(&earlier)
	return earlier == 0
}

// campaignResults counts a campaign's emails that have an outcome.
func campaignResults(db *gorm.DB, campaignID string) (sent, failed, complained int64, err error) {
	var rows []struct {
		Status string
		N      int64
	}
	if err = db.Model(&models.Email{}).Select("status, count(*) AS n").Where("campaign_id = ? AND status NOT IN ?", campaignID, []string{"PENDING", "SENDING", "QUEUED"}).Group("status").Scan(&rows).Error; err != nil {
		return
	}
	for _, r := range rows {
		sent += r.N
		switch r.Status {
		case "FAILED", "BOUNCED":
			failed += r.N
		case "COMPLAINED":
			complained += r.N
		}
	}
	return
}
