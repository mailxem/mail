package models

import (
	"crypto/sha256"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ServiceNotice is an internal transactional outbox, never a generic CRUD resource.
// It stores identifiers and a recipient snapshot, never passwords, reset tokens,
// message content, recipient lists, or raw provider errors.
type ServiceNotice struct {
	Key           string `gorm:"primaryKey"`
	Kind          string `gorm:"not null"`
	TeamID        string `gorm:"index;not null"`
	UserID        string
	AccountEmail  string
	ReferenceID   string
	Status        string `gorm:"index:service_notice_due;not null"`
	EventCount    int64
	Attempts      int
	NextAttemptAt time.Time `gorm:"index:service_notice_due"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ServiceNoticeEvent struct {
	Key       string    `gorm:"primaryKey"`
	TeamID    string    `gorm:"index:service_notice_event_day"`
	Kind      string    `gorm:"index:service_notice_event_day"`
	Reason    string    // short and sanitized; never message content or addresses
	CreatedAt time.Time `gorm:"index;index:service_notice_event_day"`
}

func noticeKey(parts ...string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%q", parts))))
}

func QueueAccountNotice(tx *gorm.DB, user User, kind, reference string, now time.Time) error {
	switch kind {
	case "welcome", "password_reset", "password_changed":
	default:
		return fmt.Errorf("unknown account notice")
	}
	if user.ID == "" || user.TeamID == "" || user.Email == "" {
		return fmt.Errorf("account notice requires a user")
	}
	n := ServiceNotice{Key: noticeKey(kind, user.ID, reference), Kind: kind, TeamID: user.TeamID, UserID: user.ID, AccountEmail: user.Email, ReferenceID: reference, Status: "QUEUED", EventCount: 1, NextAttemptAt: now, CreatedAt: now, UpdatedAt: now}
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&n).Error
}

// SendingNoticeKind maps only actionable sending outcomes, never opens/clicks.
func SendingNoticeKind(status string) string {
	return map[string]string{"FAILED": "sending_failed", "DELIVERY_UNKNOWN": "sending_unknown", "DELAYED": "sending_delayed", "BOUNCED": "sending_bounced", "COMPLAINED": "sending_complaint", "SUSPENDED": "sending_suspended"}[status]
}

// AlertThreshold is how many events of a kind a workspace may have in one UTC
// day before its owner is emailed. SERVICE_ALERT_DAILY_THRESHOLD overrides the
// default of 10 for routine delivery problems.
func AlertThreshold(kind string) int64 {
	routine := int64(10)
	if v, err := strconv.ParseInt(os.Getenv("SERVICE_ALERT_DAILY_THRESHOLD"), 10, 64); err == nil && v > 0 {
		routine = v
	}
	return map[string]int64{
		"sending_failed": routine, "sending_bounced": routine, "sending_unknown": routine, "sending_delayed": routine,
		"automation_failed": 5, "sending_complaint": 3, "sending_suspended": 1, "import_failed": 1,
	}[kind]
}

var reasonAddress = regexp.MustCompile(`[^\s<>"'@]+@[^\s<>"'@]+`)

// SanitizeReason keeps a short, single-line reason without email addresses.
func SanitizeReason(reason string) string {
	reason = strings.Join(strings.Fields(reasonAddress.ReplaceAllString(reason, "[address]")), " ")
	if r := []rune(reason); len(r) > 160 {
		reason = string(r[:157]) + "..."
	}
	return reason
}

// QueueSendingNotice records a sending outcome for the daily owner alert.
// Call within the transaction that persists the outcome.
func QueueSendingNotice(tx *gorm.DB, team, source, resource, status, reason string, now time.Time) error {
	kind := SendingNoticeKind(status)
	if kind == "" {
		return nil
	}
	return QueueAlert(tx, team, source, resource, kind, reason, now)
}

// QueueAlert counts one event per workspace, kind and UTC day. The owner's
// email is queued only once the day's count reaches AlertThreshold, so one
// failure does not page anyone, and at most one email per kind is sent a day.
// The event key deduplicates callbacks and retries.
func QueueAlert(tx *gorm.DB, team, source, resource, kind, reason string, now time.Time) error {
	threshold := AlertThreshold(kind)
	if threshold == 0 {
		return fmt.Errorf("unknown alert kind")
	}
	if team == "" || resource == "" {
		return fmt.Errorf("alert requires workspace and event identifiers")
	}
	event := ServiceNoticeEvent{Key: noticeKey(team, source, resource, kind), TeamID: team, Kind: kind, Reason: SanitizeReason(reason), CreatedAt: now}
	r := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&event)
	if r.Error != nil || r.RowsAffected == 0 {
		return r.Error
	}
	status := "COUNTING"
	if threshold <= 1 {
		status = "QUEUED"
	}
	due := now.Add(time.Minute)
	reached := "service_notices.status = 'COUNTING' AND service_notices.event_count + 1 >= ?"
	n := ServiceNotice{Key: noticeKey(team, kind, now.UTC().Format("2006-01-02")), Kind: kind, TeamID: team, Status: status, EventCount: 1, NextAttemptAt: due, CreatedAt: now, UpdatedAt: now}
	return tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "key"}}, DoUpdates: clause.Assignments(map[string]any{
		"event_count":     gorm.Expr("service_notices.event_count + 1"),
		"status":          gorm.Expr("CASE WHEN "+reached+" THEN 'QUEUED' ELSE service_notices.status END", threshold),
		"next_attempt_at": gorm.Expr("CASE WHEN "+reached+" THEN ? ELSE service_notices.next_attempt_at END", threshold, due),
	})}).Create(&n).Error
}

// TopAlertReasons returns up to three of the day's most common reasons for a
// workspace and kind, formatted for the owner's email.
func TopAlertReasons(db *gorm.DB, team, kind string, day time.Time) ([]string, error) {
	start := time.Date(day.UTC().Year(), day.UTC().Month(), day.UTC().Day(), 0, 0, 0, 0, time.UTC)
	var rows []struct {
		Reason string
		Total  int64
	}
	err := db.Model(&ServiceNoticeEvent{}).Select("reason, count(*) AS total").Where("team_id = ? AND kind = ? AND created_at >= ? AND created_at < ? AND reason <> ''", team, kind, start, start.Add(24*time.Hour)).Group("reason").Order("total DESC, reason").Limit(3).Scan(&rows).Error
	reasons := make([]string, 0, len(rows))
	for _, row := range rows {
		reasons = append(reasons, fmt.Sprintf("%d × %s", row.Total, row.Reason))
	}
	return reasons, err
}
