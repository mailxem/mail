package models

import (
	"time"

	"gorm.io/datatypes"
)

// FormRevision preserves the exact questions and consent wording visitors saw.
// Revisions are append-only; resumed sessions must match the current version.
type FormRevision struct {
	ID         string         `gorm:"primaryKey"`
	FormID     string         `gorm:"type:uuid;not null;uniqueIndex:idx_form_revision"`
	Version    int            `gorm:"not null;uniqueIndex:idx_form_revision"`
	Definition datatypes.JSON `gorm:"type:jsonb;not null"`
	CreatedAt  time.Time
}

// FormProgress is saved only on an explicit visitor request. Tokens are never stored raw.
type FormProgress struct {
	ID          string `gorm:"primaryKey"`
	FormID      string `gorm:"type:uuid;not null;index"`
	TokenHash   string `gorm:"not null;uniqueIndex" json:"-"`
	Version     int
	SessionID   string
	PageID      string
	Fields      datatypes.JSON `gorm:"type:jsonb;not null"`
	Attribution datatypes.JSON `gorm:"type:jsonb;not null"`
	ExpiresAt   time.Time      `gorm:"index"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// FormAnalyticsEvent contains no answers or visitor identifiers beyond a random session UUID.
type FormAnalyticsEvent struct {
	ID        string `gorm:"primaryKey"`
	FormID    string `gorm:"type:uuid;not null;index"`
	SessionID string `gorm:"not null"`
	Version   int
	Event     string
	PageID    string
	CreatedAt time.Time
}

// FormCompletionEvent is an outbox record committed atomically with a completed submission.
// The dispatcher may retry delivery; consumers must deduplicate using this ID.
type FormCompletionEvent struct {
	ID           string         `gorm:"primaryKey"`
	TeamID       string         `gorm:"type:uuid;not null;index"`
	FormID       string         `gorm:"type:uuid;not null;index"`
	SubmissionID string         `gorm:"type:uuid;not null;uniqueIndex"`
	ContactID    string         // Empty when marketing consent was not granted.
	Payload      datatypes.JSON `gorm:"type:jsonb;not null"`
	CreatedAt    time.Time
	DispatchedAt *time.Time `gorm:"index"`
}
