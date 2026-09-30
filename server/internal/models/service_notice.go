package models

import (
	"crypto/sha256"
	"fmt"
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
	CreatedAt time.Time `gorm:"index"`
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

// One notice per category/workspace/UTC hour, with a one-minute collection window.
// The event key deduplicates callbacks and retries even across hour boundaries.
// Call within the transaction that persists the sending outcome.
func QueueSendingNotice(tx *gorm.DB, team, source, resource, status string, now time.Time) error {
	kind := SendingNoticeKind(status)
	if kind == "" {
		return nil
	}
	if team == "" || resource == "" {
		return fmt.Errorf("sending notice requires workspace and event identifiers")
	}
	event := ServiceNoticeEvent{Key: noticeKey(team, source, resource, kind), CreatedAt: now}
	r := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&event)
	if r.Error != nil || r.RowsAffected == 0 {
		return r.Error
	}
	n := ServiceNotice{Key: noticeKey(team, kind, now.UTC().Format("2006-01-02T15")), Kind: kind, TeamID: team, Status: "QUEUED", EventCount: 1, NextAttemptAt: now.Add(time.Minute), CreatedAt: now, UpdatedAt: now}
	return tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "key"}}, DoUpdates: clause.Assignments(map[string]any{"event_count": gorm.Expr("service_notices.event_count + 1")})}).Create(&n).Error
}
