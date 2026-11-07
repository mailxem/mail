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
	// TrackEngagement: nil or true tracks opens and clicks in each edition.
	TrackEngagement *bool `json:"trackEngagement,omitempty"`
	// PublicSlug publishes sent editions at /n/{slug}; nil keeps the archive private.
	PublicSlug *string `gorm:"uniqueIndex" json:"publicSlug"`
	// SiteTheme and SiteLogoURL style the newsletter's public website.
	SiteTheme   string `gorm:"not null;default:paper" json:"siteTheme"`
	SiteLogoURL string `json:"siteLogoUrl"`
	// SiteDomain serves the website on the team's own domain once
	// SiteDomainStatus is "active"; SiteDomainID is its Cloudflare hostname.
	SiteDomain       *string `gorm:"uniqueIndex" json:"siteDomain"`
	SiteDomainID     string  `json:"-"`
	SiteDomainStatus string  `json:"siteDomainStatus"`
}

// NewsletterSignup is a reader who asked to subscribe on a newsletter's
// website. They join the list only once they confirm from their inbox.
type NewsletterSignup struct {
	ID           string `gorm:"type:uuid;primaryKey"`
	TeamID       string `gorm:"type:uuid;not null"`
	NewsletterID string `gorm:"type:uuid;not null;index:idx_signup_newsletter_email"`
	Email        string `gorm:"not null;index:idx_signup_newsletter_email"`
	// TokenHash is the SHA-256 of the emailed confirmation token.
	TokenHash   string `gorm:"not null;uniqueIndex"`
	CreatedAt   time.Time
	ConfirmedAt *time.Time
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
