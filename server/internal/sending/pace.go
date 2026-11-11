package sending

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Gmail trusts a domain more as its volume grows steadily. So each sending
// domain reaches at most gmailFirstDay Gmail recipients on its first day, then
// twice its best day of the last gmailWindow days (never less than the first
// day's). Recipients at other providers are never held back. Over the
// allowance, a message waits in the queue for the next UTC day.
var gmailFirstDay int64 = 2000

const gmailWindow = 45

type GmailPace struct {
	Domain string `gorm:"primaryKey"`
	Day    string `gorm:"primaryKey"` // UTC, 2006-01-02
	Count  int64  `gorm:"not null;default:0"`
}

func (GmailPace) TableName() string { return "managed_gmail_pace" }

func gmailRecipients(recipients string) (n int64) {
	for r := range strings.SplitSeq(recipients, ",") {
		r = strings.ToLower(strings.TrimSpace(r))
		if strings.HasSuffix(r, "@gmail.com") || strings.HasSuffix(r, "@googlemail.com") {
			n++
		}
	}
	return n
}

// paceKey is the domain Gmail sees: shared addresses all sign as SharedDomain.
func (s *Service) paceKey(d Domain) string {
	if d.Shared && s.Config.SharedDomain != "" {
		return s.Config.SharedDomain
	}
	return d.Name
}

// reserveGmail takes n from the domain's Gmail allowance for today. false
// means the message should wait for tomorrow.
func (s *Service) reserveGmail(ctx context.Context, d Domain, n int64) (bool, error) {
	if n == 0 {
		return true, nil
	}
	key, now := s.paceKey(d), s.Now()
	today := now.Format("2006-01-02")
	since := now.AddDate(0, 0, -gmailWindow).Format("2006-01-02")
	ok := false
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Today's row is the per-domain lock, so replicas reserve one at a time.
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&GmailPace{Domain: key, Day: today}).Error; err != nil {
			return err
		}
		var row GmailPace
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, "domain = ? AND day = ?", key, today).Error; err != nil {
			return err
		}
		var history []GmailPace
		if err := tx.Where("domain = ? AND day >= ? AND day < ?", key, since, today).Find(&history).Error; err != nil {
			return err
		}
		if len(history) == 0 && row.Count == 0 {
			var err error
			if history, err = s.backfillGmail(tx, key, now); err != nil {
				return err
			}
		}
		allowance := gmailFirstDay
		for _, h := range history {
			allowance = max(allowance, 2*h.Count)
		}
		// A message alone bigger than the allowance still goes, or it would wait forever.
		if row.Count > 0 && row.Count+n > allowance {
			return nil
		}
		ok = true
		return tx.Model(&row).Where("domain = ? AND day = ?", key, today).Update("count", row.Count+n).Error
	})
	return ok, err
}

// backfillGmail records a domain's Gmail days from messages it sent before
// pacing existed, so a domain already sending keeps its volume.
func (s *Service) backfillGmail(tx *gorm.DB, key string, now time.Time) ([]GmailPace, error) {
	domains := tx.Model(&Domain{}).Select("id").Where("name = ?", key)
	if key == s.Config.SharedDomain {
		domains = tx.Model(&Domain{}).Select("id").Where("shared = true")
	}
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	var sent []struct {
		Recipients string
		CreatedAt  time.Time
	}
	if err := tx.Model(&Message{}).Select("recipients, created_at").Where("domain_id IN (?) AND status IN ? AND created_at >= ? AND created_at < ?", domains, []string{"SENT", "DELIVERED"}, start.AddDate(0, 0, -gmailWindow), start).Find(&sent).Error; err != nil {
		return nil, err
	}
	byDay := map[string]int64{}
	for _, m := range sent {
		if n := gmailRecipients(m.Recipients); n > 0 {
			byDay[m.CreatedAt.UTC().Format("2006-01-02")] += n
		}
	}
	rows := make([]GmailPace, 0, len(byDay))
	for day, n := range byDay {
		rows = append(rows, GmailPace{Domain: key, Day: day, Count: n})
	}
	if len(rows) == 0 {
		return rows, nil
	}
	return rows, tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
}

// hold puts a claimed message back in the queue until the next UTC day.
func (s *Service) hold(ctx context.Context, m Message) error {
	now := s.Now()
	next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	return s.DB.WithContext(context.WithoutCancel(ctx)).Model(&Message{}).Where("id = ? AND status = ?", m.ID, "SENDING").Updates(map[string]any{"status": "QUEUED", "hold_until": next, "detail": "Waiting for tomorrow's Gmail allowance, so Gmail sees this domain's volume grow steadily.", "updated_at": now}).Error
}
