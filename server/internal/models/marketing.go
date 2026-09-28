package models

import "time"

// Newsletter is reusable configuration. Each scheduled edition is an immutable Campaign.
type Newsletter struct {
	Base
	TeamID        string     `gorm:"type:uuid;not null;index" json:"teamId"`
	Name          string     `json:"name"`
	Subject       string     `json:"subject"`
	Description   string     `json:"description"`
	TemplateID    string     `gorm:"type:uuid;not null" json:"templateId"`
	Template      *Template  `json:"template,omitempty"`
	ListID        string     `gorm:"type:uuid;not null" json:"listId"`
	SMTPConfigID  string     `gorm:"type:uuid;not null" json:"smtpConfigId"`
	Status        string     `gorm:"not null;default:DRAFT;index" json:"status"`
	MonthDay      int        `json:"-"` // Original monthly day, preserved through shorter months.
	Cadence       string     `gorm:"not null;default:ONCE" json:"cadence"`
	Timezone      string     `gorm:"not null;default:UTC" json:"timezone"`
	NextSendAt    *time.Time `gorm:"index" json:"nextSendAt"`
	LastSentAt    *time.Time `json:"lastSentAt"`
	Editions      int        `gorm:"default:0" json:"editions"`
	PostalAddress string     `json:"postalAddress"`
}

type ContactNote struct {
	Base
	TeamID    string `gorm:"type:uuid;not null;index" json:"teamId"`
	ContactID string `gorm:"type:uuid;not null;index" json:"contactId"`
	Body      string `json:"body"`
	AuthorID  string `json:"authorId"`
}

// FormReceipt makes retries idempotent without exposing whether an email is subscribed.
type FormReceipt struct {
	ID           string `gorm:"primaryKey"`
	FormID       string `gorm:"type:uuid;index"`
	SubmissionID string `gorm:"type:uuid"`
	CreatedAt    time.Time
}
