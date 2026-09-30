package sending

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/models"
	"kori/internal/onboardingemails"
)

type MilestoneEmail struct {
	Key           string `gorm:"primaryKey"`
	TeamID        string `gorm:"index;not null"`
	DomainID      string
	Kind          string
	Status        string `gorm:"index"`
	Attempts      int
	NextAttemptAt time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (MilestoneEmail) TableName() string { return "managed_milestone_emails" }

// MilestoneSender returns uncertain=true only when SMTP might have accepted DATA.
type MilestoneSender func(context.Context, string, string, string, string, string) (uncertain bool, err error)

// Existing installations start from their current state. A routine DNS refresh
// after deployment must not send a backlog of historical onboarding notices.
func baselineMilestones(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		marker := MilestoneEmail{Key: "migration:milestones-v1", TeamID: "system", Kind: "migration", Status: "SKIPPED"}
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&marker)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
		for _, spec := range []struct{ kind, table, scope, condition string }{
			{"domain_added", "managed_domains", "id", "1=1"},
			{"ownership_verified", "managed_domains", "id", "ownership = true"},
			{"domain_ready", "managed_domains", "id", "ready = true"},
			{"sender_connected", "managed_domains", "id", "smtp_config_id <> ''"},
			{"approved", "managed_accounts", "team_id", "approved = true"},
		} {
			domain := spec.scope
			if spec.kind == "approved" {
				domain = "''"
			}
			query := "INSERT INTO managed_milestone_emails (key,team_id,domain_id,kind,status,attempts,next_attempt_at,created_at,updated_at) SELECT team_id || ':' || " + spec.scope + " || ':" + spec.kind + "',team_id," + domain + ",?,'SKIPPED',0,?,?,? FROM " + spec.table + " WHERE " + spec.condition + " ON CONFLICT (key) DO NOTHING"
			now := time.Now().UTC()
			if err := tx.Exec(query, spec.kind, now, now, now).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func QueueMilestone(tx *gorm.DB, team, domain, kind string) error {
	if _, ok := onboardingemails.Find(kind); !ok {
		return errors.New("unknown sending milestone")
	}
	scope := domain
	if kind == "approved" {
		scope = team
	}
	n := MilestoneEmail{Key: team + ":" + scope + ":" + kind, TeamID: team, DomainID: domain, Kind: kind, Status: "QUEUED", NextAttemptAt: time.Now().UTC()}
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&n).Error
}

// ProcessMilestone claims before network I/O. A lost acknowledgement is held for
// operator review; it is never blindly retried as a duplicate service email.
func (s *Service) ProcessMilestone(ctx context.Context) error {
	if s.Notify == nil {
		return nil
	}
	now := s.Now()
	if err := s.DB.WithContext(ctx).Model(&MilestoneEmail{}).Where("status = ? AND updated_at < ?", "SENDING", now.Add(-5*time.Minute)).Update("status", "DELIVERY_UNKNOWN").Error; err != nil {
		return err
	}
	var n MilestoneEmail
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Where("status = ? AND next_attempt_at <= ?", "QUEUED", now).Order("created_at, key").First(&n).Error; err != nil {
			return err
		}
		result := tx.Model(&n).Where("status = ?", "QUEUED").Updates(map[string]any{"status": "SENDING", "attempts": n.Attempts + 1, "updated_at": now})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
	if err != nil {
		return err
	}
	status := "SKIPPED"
	// Resolve the owner again at dispatch; an old queued notice must never leak to
	// someone who left the workspace. Legacy teams use their first current admin.
	var owner models.User
	var team models.Team
	err = s.DB.WithContext(ctx).Where("id = ? AND is_deleted = false", n.TeamID).First(&team).Error
	if err == nil {
		q := s.DB.WithContext(ctx).Where("team_id = ? AND is_deleted = false AND role IN ?", n.TeamID, []models.UserRole{models.UserRoleAdmin, models.UserRoleSuperAdmin})
		if team.OwnerUserID != "" {
			q = q.Where("id = ?", team.OwnerUserID)
		} else {
			q = q.Order("created_at, id")
		}
		err = q.First(&owner).Error
	}
	var domain Domain
	if err == nil && n.DomainID != "" {
		err = s.DB.WithContext(ctx).Where("id = ? AND team_id = ?", n.DomainID, n.TeamID).First(&domain).Error
	}
	var account Account
	if err == nil {
		err = s.DB.WithContext(ctx).Where("team_id = ?", n.TeamID).First(&account).Error
	}
	eligible := err == nil && !n.CreatedAt.Before(now.Add(-48*time.Hour))
	if n.Kind == "approved" || n.Kind == "sender_connected" {
		eligible = eligible && account.Approved && !account.Suspended && !account.Paused
	}
	if n.Kind == "sender_connected" {
		eligible = eligible && domain.Ready && domain.SMTPConfigID != ""
	}
	if n.Kind == "ownership_verified" {
		eligible = eligible && domain.Ownership
	}
	if n.Kind == "domain_ready" {
		eligible = eligible && domain.Ready
	}
	if eligible {
		var recipient string
		recipient, err = address(owner.Email)
		if err == nil {
			var blocked bool
			blocked, err = Suppressed(s.DB.WithContext(ctx), n.TeamID, recipient, false)
			if err == nil && !blocked {
				template, _ := onboardingemails.Find(n.Kind)
				label := domain.Name
				if label == "" {
					label = "See your workspace"
				}
				html, text := template.Render(team.Name, label, s.NotificationURL)
				uncertain := false
				uncertain, err = s.Notify(ctx, n.Key, recipient, template.Subject, html, text)
				status = "ACCEPTED"
				if err != nil {
					status = "QUEUED"
					if uncertain {
						status = "DELIVERY_UNKNOWN"
					} else if n.Attempts >= 4 {
						status = "FAILED"
					}
				}
			}
		}
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		status = "QUEUED"
	}
	// A failed suppression lookup is a transient dependency error, not permission to send.
	if err != nil && status == "SKIPPED" && !errors.Is(err, gorm.ErrRecordNotFound) {
		status = "QUEUED"
	}
	if status == "QUEUED" && n.Attempts >= 4 {
		status = "FAILED"
	}
	record, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	return s.DB.WithContext(record).Model(&n).Where("status = ?", "SENDING").Updates(map[string]any{"status": status, "next_attempt_at": now.Add(time.Minute * time.Duration(1<<min(n.Attempts+1, 6))), "updated_at": s.Now()}).Error
}
