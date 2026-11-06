package sending

import (
	"context"
	"errors"
	"kori/internal/billing"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/aws/smithy-go"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/models"
)

// ProcessOne claims durably before making a network request. A crashed SENDING
// claim is quarantined, never automatically replayed: SES has no idempotency key.
func (s *Service) ProcessOne(ctx context.Context) error {
	claimed, err := s.claim(ctx, 1)
	if err != nil {
		return err
	}
	// An empty queue is normal.
	if len(claimed) == 0 {
		return gorm.ErrRecordNotFound
	}
	return s.send(ctx, claimed[0])
}

// claim moves up to n eligible messages from QUEUED to SENDING in one
// transaction. Workspaces take turns (each team's oldest first, then each
// team's second oldest, ...), so one large send cannot starve the others.
// Postgres rejects FOR UPDATE beside a window function, so ids are ranked in a
// subquery and locked outside it; the outer status check is re-evaluated after
// waiting on a row another replica claimed.
// ponytail: ranks every QUEUED row per claim; fine at ~1e5 queued, keep a
// per-team cursor if the queue grows far past that.
func (s *Service) claim(ctx context.Context, n int) ([]Message, error) {
	var claimed []Message
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ranked := tx.Model(&Message{}).Select("id, created_at, ROW_NUMBER() OVER (PARTITION BY team_id ORDER BY created_at) AS team_turn").Where("status = ?", "QUEUED").Where("EXISTS (SELECT 1 FROM managed_accounts a WHERE a.team_id = managed_messages.team_id AND a.paused = false)").Where("EXISTS (SELECT 1 FROM managed_domains d WHERE d.id = managed_messages.domain_id AND d.team_id = managed_messages.team_id AND d.ready = true)")
		fair := tx.Table("(?) AS ranked", ranked).Select("id").Order("team_turn, created_at").Limit(n)
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Where("status = ? AND id IN (?)", "QUEUED", fair).Find(&claimed).Error; e != nil {
			return e
		}
		if len(claimed) == 0 {
			return nil
		}
		ids := make([]string, len(claimed))
		for i, m := range claimed {
			ids[i] = m.ID
		}
		r := tx.Model(&Message{}).Where("id IN ? AND status = ?", ids, "QUEUED").Updates(map[string]any{"status": "SENDING", "updated_at": s.Now()})
		if r.Error != nil {
			return r.Error
		}
		if r.RowsAffected != int64(len(ids)) {
			return errors.New("managed claim raced another worker")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return claimed, nil
}

// release returns claimed messages that were never handed to the provider.
func (s *Service) release(ms []Message) {
	if len(ms) == 0 {
		return
	}
	ids := make([]string, len(ms))
	for i, m := range ms {
		ids[i] = m.ID
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if s.DB.WithContext(ctx).Model(&Message{}).Where("id IN ? AND status = ?", ids, "SENDING").Update("status", "QUEUED").Error != nil {
		log.Print("managed delivery could not release unsent claims; they will be quarantined")
	}
}

// send dispatches one message claim() moved to SENDING and records the outcome.
func (s *Service) send(ctx context.Context, m Message) error {
	status, detail := "FAILED", "Sending is paused, unapproved, or its domain or credential is no longer eligible."
	var d Domain
	var a Account
	err := s.DB.WithContext(ctx).First(&a, "team_id = ? AND approved = true AND suspended = false AND paused = false", m.TeamID).Error
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
		if r.RowsAffected > 0 && m.IsTest {
			kind := "test_attention"
			if status == "SENT" {
				kind = "test_accepted"
			}
			if err := QueueMilestone(tx, m.TeamID, m.DomainID, kind); err != nil {
				return err
			}
		}
		if r.RowsAffected > 0 && s.NotifyAlerts && !m.IsTest {
			if err := models.QueueSendingNotice(tx, m.TeamID, "managed", m.ID, status, detail, s.Now()); err != nil {
				return err
			}
		}
		if r.RowsAffected > 0 && status == "SENT" && !m.IsTest {
			if err := billing.CountEmails(tx, m.TeamID, len(strings.Split(m.Recipients, ",")), s.Now()); err != nil {
				return err
			}
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
	notificationsDone := make(chan struct{})
	go func() {
		defer close(notificationsDone)
		tick := time.NewTicker(2 * time.Second)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				if err := s.ProcessMilestone(ctx); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
					log.Print("managed milestone worker failed; inspect notification outbox")
				}
			}
		}
	}()
	defer func() { <-notificationsDone }()
	deliveryDone := make(chan struct{})
	go func() {
		defer close(deliveryDone)
		// The rate is per API process: replicas each get their own bucket, so
		// MANAGED_SEND_RATE x replicas must stay within the SES account's max send rate.
		limiter := rate.NewLimiter(rate.Limit(max(s.Config.SendRate, 1)), 1)
		s.deliver(ctx, limiter, int(max(s.Config.Workers, 1)), 100)
	}()
	defer func() { <-deliveryDone }()
	maintenance := time.NewTicker(time.Minute)
	defer maintenance.Stop()
	for {
		select {
		case <-ctx.Done():
			return
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

// deliver claims up to batch messages at a time and hands them to workers,
// paced by limiter. It claims again at once while claims come back full and
// waits a second when the queue is drained. On cancellation it stops claiming,
// lets in-flight sends finish, and returns claims never handed to a worker to
// QUEUED. Replicas stay safe because claims use SKIP LOCKED. A batch of 100 at
// 1/s or more drains in under two minutes, inside MaintainMessages' 5-minute
// quarantine of stale SENDING claims.
func (s *Service) deliver(ctx context.Context, limiter *rate.Limiter, workers, batch int) {
	jobs := make(chan Message)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for m := range jobs {
				// A started SES request is finished and recorded even during shutdown.
				if e := s.send(context.WithoutCancel(ctx), m); e != nil {
					log.Print("managed delivery worker failed; inspect database/provider availability")
				}
			}
		}()
	}
	defer wg.Wait()
	defer close(jobs)
	for ctx.Err() == nil {
		claimed, err := s.claim(ctx, batch)
		if err != nil && ctx.Err() == nil {
			log.Print("managed delivery claim failed; inspect database availability")
		}
		for i, m := range claimed {
			if limiter.Wait(ctx) != nil {
				s.release(claimed[i:])
				return
			}
			select {
			case jobs <- m:
			case <-ctx.Done():
				s.release(claimed[i:])
				return
			}
		}
		if len(claimed) < batch {
			select {
			case <-ctx.Done():
			case <-time.After(time.Second):
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
			if s.NotifyAlerts && !m.IsTest {
				if err := models.QueueSendingNotice(tx, m.TeamID, "managed", m.ID, status, detail, s.Now()); err != nil {
					return err
				}
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
