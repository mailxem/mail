// Package billing reports emails sent on Xem Cloud to the payments service,
// which bills them through the Dodo meter. It does nothing unless
// PAYMENTS_INTERNAL_URL and PAYMENTS_INTERNAL_KEY are set, so self-hosted
// installs never count or report usage.
package billing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// EmailUsage counts emails sent by a workspace in one UTC minute.
type EmailUsage struct {
	TeamID     string     `gorm:"primaryKey;type:uuid"`
	Minute     time.Time  `gorm:"primaryKey"`
	Emails     int64      `gorm:"not null"`
	ReportedAt *time.Time `gorm:"index"`
}

// Enabled is set at startup when the payments service is configured.
var Enabled bool

// CountEmails adds n sent emails to the workspace's current minute. Call it in
// the transaction that records the send, so a rolled-back send is not billed.
func CountEmails(tx *gorm.DB, team string, n int, now time.Time) error {
	if !Enabled || team == "" || n <= 0 {
		return nil
	}
	row := EmailUsage{TeamID: team, Minute: now.UTC().Truncate(time.Minute), Emails: int64(n)}
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "team_id"}, {Name: "minute"}},
		DoUpdates: clause.Assignments(map[string]any{"emails": gorm.Expr("email_usages.emails + ?", n)}),
	}).Create(&row).Error
}

// Reporter sends finished minutes to the payments service.
type Reporter struct {
	DB   *gorm.DB
	URL  string // payments API base, e.g. https://payments.internal/api/v1
	Key  string
	Now  func() time.Time
	HTTP *http.Client
}

// FromEnv returns a reporter, or nil when billing is not configured.
func FromEnv(db *gorm.DB) *Reporter {
	u, key := os.Getenv("PAYMENTS_INTERNAL_URL"), os.Getenv("PAYMENTS_INTERNAL_KEY")
	if u == "" || key == "" {
		return nil
	}
	return &Reporter{DB: db, URL: u, Key: key, Now: time.Now, HTTP: &http.Client{Timeout: 15 * time.Second}}
}

// Flush reports every finished, unreported minute. Each report's event ID is
// derived from the workspace and minute, so a retry is never billed twice.
func (r *Reporter) Flush(ctx context.Context) error {
	var rows []EmailUsage
	current := r.Now().UTC().Truncate(time.Minute)
	if err := r.DB.WithContext(ctx).Where("reported_at IS NULL AND minute < ?", current).Order("minute").Limit(500).Find(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		body, _ := json.Marshal(map[string]any{
			"feature": "emails_per_month", "usage": row.Emails,
			"event_id": fmt.Sprintf("xem-emails-%s-%d", row.TeamID, row.Minute.Unix()),
		})
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.URL+"/usage/record/"+url.PathEscape(row.TeamID), bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", r.Key)
		res, err := r.HTTP.Do(req)
		if err != nil {
			return err
		}
		res.Body.Close()
		if res.StatusCode >= 300 {
			return fmt.Errorf("payments usage report: %d for team %s", res.StatusCode, row.TeamID)
		}
		now := r.Now()
		if err := r.DB.WithContext(ctx).Model(&EmailUsage{}).Where("team_id = ? AND minute = ?", row.TeamID, row.Minute).Update("reported_at", now).Error; err != nil {
			return err
		}
	}
	return nil
}

// Run flushes every 30 seconds until ctx is cancelled.
func (r *Reporter) Run(ctx context.Context, logErr func(error)) {
	t := time.NewTicker(30 * time.Second)
	defer t.Stop()
	for {
		if err := r.Flush(ctx); err != nil && ctx.Err() == nil {
			logErr(err)
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}
