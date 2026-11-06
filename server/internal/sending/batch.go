package sending

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/billing"
	"kori/internal/hygiene"
	"kori/internal/models"
)

// BatchResult is one submission's outcome in SubmitBatch.
type BatchResult struct {
	Message Message
	Err     error
}

// prepared is a submission that passed the checks needing no database.
type prepared struct {
	raw                      []byte
	from, subject, hash, key string
	to                       []string
}

func (s *Service) prepare(in Submission) (prepared, error) {
	var p prepared
	if !s.Config.Enabled || len(in.Raw) == 0 || int64(len(in.Raw)) > s.Config.MaxBytes || len(in.Recipients) < 1 || len(in.Recipients) > 50 || len(in.Key) > 128 || len(in.Key) < 1 {
		return p, errors.New("invalid message, recipient count, or idempotency key")
	}
	raw, from, subject, e := NormalizeRaw(in.Raw, in.From)
	if e != nil {
		return p, e
	}
	if in.Marketing {
		if e = validateMarketing(raw); e != nil {
			return p, e
		}
	}
	to := []string{}
	seen := map[string]bool{}
	for _, r := range in.Recipients {
		a, e := address(r)
		if e != nil {
			return p, e
		}
		if !seen[a] {
			to = append(to, a)
			seen[a] = true
		}
	}
	sort.Strings(to)
	digest := sha256.Sum256(append(append([]byte(strings.Join(to, ",")+"\x00"+in.DomainID+"\x00"+in.CredentialID+"\x00"+in.EmailID+fmt.Sprint(in.IsTest)+"\x00"), raw...), byte(0)))
	return prepared{raw: raw, from: from, subject: subject, hash: hex.EncodeToString(digest[:]), key: in.TeamID + ":" + in.Key, to: to}, nil
}

// SubmitBatch submits one team's messages with the rules of Submit, but locks
// the account, counts the queue and checks suppressions once for the batch,
// and inserts the messages together. Each submission gets its own result; a
// refused one doesn't stop the others. Submit is SubmitBatch of one.
func (s *Service) SubmitBatch(ctx context.Context, items []Submission) []BatchResult {
	out := make([]BatchResult, len(items))
	if len(items) == 0 {
		return out
	}
	team := items[0].TeamID
	ps := make([]prepared, len(items))
	var pending []int
	for i, in := range items {
		p, err := s.prepare(in)
		if err == nil && in.TeamID != team {
			err = ErrDenied // a batch is one team's, like the account lock
		}
		if err == nil && !in.IsTest {
			if e := billing.AllowSend(ctx, team, len(p.to)); e != nil {
				err = fmt.Errorf("%w: %v", ErrLimit, e)
			}
		}
		if err != nil {
			out[i].Err = err
			continue
		}
		ps[i] = p
		pending = append(pending, i)
	}
	if len(pending) == 0 {
		return out
	}
	// The plan sets the limits of accounts still on the automatic allowance;
	// limits an operator set stay as they are. Looked up outside the lock.
	planDaily, planMonthly, planKnown := billing.ManagedLimits(ctx, team)
	// Results are only kept if the transaction commits.
	results := make([]BatchResult, len(items))
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var a Account
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&a, "team_id = ?", team).Error; err != nil {
			return ErrDenied
		}
		if !a.Approved || a.Suspended || a.Paused {
			return ErrDenied
		}
		keys := make([]string, 0, len(pending))
		var recipients []string
		for _, i := range pending {
			keys = append(keys, ps[i].key)
			recipients = append(recipients, ps[i].to...)
		}
		var existing []Message
		if err := tx.Where("idempotency_key IN ?", keys).Find(&existing).Error; err != nil {
			return err
		}
		byKey := map[string]Message{}
		for _, m := range existing {
			byKey[m.IdempotencyKey] = m
		}
		blocked, err := suppressedSet(tx, team, recipients, true)
		if err != nil {
			return err
		}
		limits, queue := a, s.Config.QueueLimit
		if s.applyPlan(&limits, planDaily, planMonthly, planKnown) {
			queue = max(queue, planDaily) // a day's whole send may wait in the queue
		}
		daily, monthly, budget := limits.DailyLimit, limits.MonthlyLimit, limits.MonthlyBudgetMicros
		var queued int64
		if err = tx.Model(&Message{}).Where("team_id = ? AND status IN ?", team, []string{"QUEUED", "SENDING"}).Count(&queued).Error; err != nil {
			return err
		}
		day, month := s.Now().Format("2006-01-02"), s.Now().Format("2006-01")
		if a.Day != day {
			a.Day = day
			a.DailyUsed = 0
		}
		if a.Month != month {
			a.Month = month
			a.MonthlyUsed = 0
			a.BudgetUsedMicros = 0
		}
		domains := map[string]*Domain{}
		credentials := map[string]bool{}
		var fresh []Message
		var emailIDs []string
		for _, i := range pending {
			in, p := items[i], ps[i]
			if m, ok := byKey[p.key]; ok {
				if m.PayloadHash != p.hash {
					results[i].Err = errors.New("idempotency key was already used for a different message")
				} else {
					results[i].Message = m
				}
				continue
			}
			d, ok := domains[in.DomainID]
			if !ok {
				d = &Domain{}
				if tx.First(d, "id = ? AND team_id = ? AND ready = true", in.DomainID, team).Error != nil {
					d = nil
				}
				domains[in.DomainID] = d
			}
			if d == nil {
				results[i].Err = ErrDenied
				continue
			}
			if strings.Split(p.from, "@")[1] != d.Name {
				results[i].Err = errors.New("From address must belong to the selected verified domain")
				continue
			}
			if d.CheckedAt == nil || s.Now().Sub(*d.CheckedAt) > 24*time.Hour {
				results[i].Err = errors.New("domain verification is stale; check DNS again")
				continue
			}
			if in.CredentialID != "" {
				ck := in.CredentialID + ":" + d.ID
				valid, ok := credentials[ck]
				if !ok {
					var c Credential
					valid = tx.First(&c, "id = ? AND team_id = ? AND domain_id = ? AND revoked_at IS NULL AND expires_at > ?", in.CredentialID, team, d.ID, s.Now()).Error == nil
					credentials[ck] = valid
				}
				if !valid {
					results[i].Err = ErrDenied
					continue
				}
			}
			suppressed := false
			for _, r := range p.to {
				suppressed = suppressed || blocked[r]
			}
			if suppressed {
				results[i].Err = ErrSuppressed
				continue
			}
			if queued >= queue {
				results[i].Err = ErrLimit
				continue
			}
			count := int64(len(p.to))
			cost := count * s.Config.CostMicros
			if count > daily-a.DailyUsed || count > monthly-a.MonthlyUsed || cost > budget-a.BudgetUsedMicros {
				results[i].Err = ErrLimit
				continue
			}
			a.DailyUsed += count
			a.MonthlyUsed += count
			a.BudgetUsedMicros += cost
			queued++
			var raw []byte
			if !in.Transient { // a transient message is rebuilt from its email at dispatch (RenderRaw)
				raw = append([]byte("Feedback-ID: "+feedbackID(in)+"\r\n"), p.raw...)
			}
			m := Message{ID: uuid.NewString(), TeamID: team, DomainID: d.ID, CredentialID: in.CredentialID, EmailID: in.EmailID, IsTest: in.IsTest, IdempotencyKey: p.key, PayloadHash: p.hash, From: p.from, Recipients: strings.Join(p.to, ","), Subject: p.subject, Raw: raw, Status: "QUEUED"}
			byKey[p.key] = m
			results[i].Message = m
			fresh = append(fresh, m)
			if in.EmailID != "" {
				emailIDs = append(emailIDs, in.EmailID)
			}
		}
		if len(fresh) == 0 {
			return nil
		}
		if err = tx.Save(&a).Error; err != nil {
			return err
		}
		if len(emailIDs) > 0 {
			if err = tx.Model(&models.Email{}).Where("id IN ? AND team_id = ?", emailIDs, team).Updates(map[string]any{"status": "QUEUED", "error": ""}).Error; err != nil {
				return err
			}
		}
		if err = tx.CreateInBatches(fresh, 500).Error; err != nil {
			return err
		}
		for _, m := range fresh {
			if m.IsTest {
				if err = QueueMilestone(tx, m.TeamID, m.DomainID, "test_queued"); err != nil {
					return err
				}
			}
		}
		return nil
	})
	for _, i := range pending {
		if err != nil {
			out[i].Err = err
		} else {
			out[i] = results[i]
		}
	}
	return out
}

// suppressedSet is Suppressed for many recipients of one team, in four
// queries whatever their number. It returns the suppressed addresses.
func suppressedSet(tx *gorm.DB, team string, recipients []string, marketing bool) (map[string]bool, error) {
	blocked, e := hygiene.Blocked(tx, recipients...) // the never-email list every workspace shares
	if e != nil {
		return nil, e
	}
	var found []string
	if e := tx.Model(&Suppression{}).Where("team_id = ? AND address IN ?", team, recipients).Pluck("address", &found).Error; e != nil {
		return nil, e
	}
	for _, a := range found {
		blocked[a] = true
	}
	found = nil
	if e := tx.Model(&models.SuppressionList{}).Where("team_id = ? AND LOWER(email_address) IN ? AND is_active = true AND is_deleted = false AND (expires_at IS NULL OR expires_at > ?)", team, recipients, time.Now()).Pluck("LOWER(email_address)", &found).Error; e != nil {
		return nil, e
	}
	for _, a := range found {
		blocked[a] = true
	}
	states := []string{"BOUNCED", "COMPLAINED"}
	if marketing {
		states = append(states, "UNSUBSCRIBED")
	}
	found = nil
	if e := tx.Model(&models.Contact{}).Where("team_id = ? AND LOWER(email) IN ? AND status IN ?", team, recipients, states).Pluck("LOWER(email)", &found).Error; e != nil {
		return nil, e
	}
	for _, a := range found {
		blocked[a] = true
	}
	return blocked, nil
}
