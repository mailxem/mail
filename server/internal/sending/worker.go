package sending

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/aws/smithy-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/models"
)

// ProcessOne claims durably before making a network request. A crashed SENDING
// claim is quarantined, never automatically replayed: SES has no idempotency key.
func (s *Service) ProcessOne(ctx context.Context) error {
	var m Message
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Where("status = ?", "QUEUED").Where("EXISTS (SELECT 1 FROM managed_accounts a WHERE a.team_id = managed_messages.team_id AND a.paused = false)").Where("EXISTS (SELECT 1 FROM managed_domains d WHERE d.id = managed_messages.domain_id AND d.team_id = managed_messages.team_id AND d.ready = true)").Order("created_at").First(&m).Error; e != nil {
			return e
		}
		r := tx.Model(&Message{}).Where("id = ? AND status = ?", m.ID, "QUEUED").Updates(map[string]any{"status": "SENDING", "updated_at": s.Now()})
		if r.Error != nil {
			return r.Error
		}
		if r.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
	if err != nil {
		return err
	}
	status, detail := "FAILED", "Sending is paused, unapproved, or its domain or credential is no longer eligible."
	var d Domain
	var a Account
	err = s.DB.WithContext(ctx).First(&a, "team_id = ? AND approved = true AND suspended = false AND paused = false", m.TeamID).Error
	if err == nil {
		err = s.DB.WithContext(ctx).First(&d, "id = ? AND team_id = ? AND ready = true", m.DomainID, m.TeamID).Error
	}
	if err == nil && (d.CheckedAt == nil || s.Now().Sub(*d.CheckedAt) > 24*time.Hour) {
		err = ErrDenied
	}
	if err == nil && m.CredentialID != "" {
		var c Credential
		err = s.DB.WithContext(ctx).First(&c, "id = ? AND team_id = ? AND domain_id = ? AND revoked_at IS NULL AND expires_at > ?", m.CredentialID, m.TeamID, m.DomainID, s.Now()).Error
	}
	if err == nil {
		for _, r := range strings.Split(m.Recipients, ",") {
			var blocked bool
			blocked, err = Suppressed(s.DB.WithContext(ctx), m.TeamID, r, true)
			if blocked && err == nil {
				err = ErrSuppressed
				status = "SUPPRESSED"
				detail = "Recipient was suppressed before dispatch."
			}
			if err != nil {
				break
			}
		}
	}
	providerID := ""
	if err == nil && (len(m.Raw) == 0 || m.CreatedAt.Before(s.Now().Add(-7*24*time.Hour))) {
		err = ErrDenied
		detail = "Queued message expired before dispatch."
	}
	if err == nil {
		sendCtx, cancel := context.WithTimeout(ctx, 25*time.Second)
		providerID, err = s.Provider.Send(sendCtx, m.TeamID, d.Name, m.From, strings.Split(m.Recipients, ","), m.Raw, m.ID)
		cancel()
		if err == nil {
			status = "SENT"
			detail = "Accepted by SES; delivery is not yet confirmed."
		} else {
			// Keep provider diagnostics in operator logs; persisted status is customer-facing.
			log.Printf("managed delivery provider send failed: message_id=%s error=%v", m.ID, err)
			var ae smithy.APIError
			// SES can decode IAM denials as GenericAPIError with FaultUnknown.
			// These explicit rejections still prove the message was not accepted.
			if errors.As(err, &ae) && (ae.ErrorFault() == smithy.FaultClient || ae.ErrorCode() == "AccessDenied" || ae.ErrorCode() == "AccessDeniedException") {
				detail = "Provider rejected the message: " + ae.ErrorCode()
			} else {
				status = "DELIVERY_UNKNOWN"
				detail = "Provider acknowledgement is uncertain. Review provider events before resending."
			}
		}
	}
	// A provider event can arrive before Send returns. Never overwrite that event.
	recordCtx, cancelRecord := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancelRecord()
	return s.DB.WithContext(recordCtx).Transaction(func(tx *gorm.DB) error {
		r := tx.Model(&Message{}).Where("id = ? AND status = ?", m.ID, "SENDING").Updates(map[string]any{"status": status, "detail": detail, "provider_id": providerID, "updated_at": s.Now()})
		if r.Error != nil {
			return r.Error
		}
		if r.RowsAffected > 0 && m.EmailID != "" {
			fields := map[string]any{"status": status, "error": detail}
			if status == "SENT" {
				fields["sent_at"] = s.Now()
				fields["error"] = ""
			}
			return tx.Model(&models.Email{}).Where("id = ? AND team_id = ?", m.EmailID, m.TeamID).Updates(fields).Error
		}
		return nil
	})
}
func (s *Service) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	maintenance := time.NewTicker(time.Minute)
	defer maintenance.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if e := s.ProcessOne(ctx); e != nil && !errors.Is(e, gorm.ErrRecordNotFound) {
				log.Print("managed delivery worker failed; inspect database/provider availability")
			}
		case <-maintenance.C:
			if e := s.MaintainMessages(ctx); e != nil {
				log.Print("managed delivery maintenance failed; inspect database availability")
			}
			var domains []Domain
			if s.DB.WithContext(ctx).Where("(ready = false AND checked_at < ?) OR (ready = true AND checked_at < ?) OR checked_at IS NULL", s.Now().Add(-5*time.Minute), s.Now().Add(-12*time.Hour)).Order("checked_at NULLS FIRST").Limit(5).Find(&domains).Error == nil {
				for _, d := range domains {
					refreshCtx, cancel := context.WithTimeout(ctx, 25*time.Second)
					_, _ = s.RefreshDomain(refreshCtx, d.TeamID, d.ID)
					cancel()
				}
			}
		}
	}
}

// MaintainMessages updates the outbox and its source email atomically. Row locks
// prevent retention work racing a dispatch claim or authenticated provider event.
func (s *Service) MaintainMessages(ctx context.Context) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var messages []Message
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Where("(status = ? AND updated_at < ?) OR (status = ? AND created_at < ?)", "SENDING", s.Now().Add(-5*time.Minute), "QUEUED", s.Now().Add(-7*24*time.Hour)).Limit(500).Find(&messages).Error; err != nil {
			return err
		}
		for _, m := range messages {
			status, detail := "FAILED", "Queued message expired before dispatch."
			if m.Status == "SENDING" {
				status, detail = "DELIVERY_UNKNOWN", "Worker stopped before outcome was recorded. Review events before resending."
			}
			if err := tx.Model(&m).Updates(map[string]any{"status": status, "detail": detail}).Error; err != nil {
				return err
			}
			if m.EmailID != "" {
				if err := tx.Model(&models.Email{}).Where("id = ? AND team_id = ?", m.EmailID, m.TeamID).Updates(map[string]any{"status": status, "error": detail}).Error; err != nil {
					return err
				}
			}
		}
		return tx.Model(&Message{}).Where("created_at < ? AND status NOT IN ?", s.Now().Add(-7*24*time.Hour), []string{"SENDING", "QUEUED"}).Update("raw", nil).Error
	})
}
