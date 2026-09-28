package models

import (
	"time"

	"gorm.io/datatypes"
)

// BounceType represents the type of bounce
type BounceType string

const (
	BounceTypeHard    BounceType = "HARD"    // Permanent failure
	BounceTypeSoft    BounceType = "SOFT"    // Temporary failure
	BounceTypeBlock   BounceType = "BLOCK"   // Blocked by recipient server
	BounceTypeGeneral BounceType = "GENERAL" // Unknown type
)

// BounceCategory represents the category of bounce
type BounceCategory string

const (
	BounceCategoryUndetermined    BounceCategory = "UNDETERMINED"
	BounceCategoryMailboxFull     BounceCategory = "MAILBOX_FULL"
	BounceCategoryMessageTooLarge BounceCategory = "MESSAGE_TOO_LARGE"
	BounceCategoryContentRejected BounceCategory = "CONTENT_REJECTED"
	BounceCategoryAttachmentRejected BounceCategory = "ATTACHMENT_REJECTED"
	BounceCategoryInvalidAddress  BounceCategory = "INVALID_ADDRESS"
	BounceCategoryNoMailbox       BounceCategory = "NO_MAILBOX"
	BounceCategoryDomainNotFound  BounceCategory = "DOMAIN_NOT_FOUND"
	BounceCategorySpamDetected    BounceCategory = "SPAM_DETECTED"
)

// SuppressionReason represents why a contact was suppressed
type SuppressionReason string

const (
	SuppressionReasonHardBounce  SuppressionReason = "HARD_BOUNCE"
	SuppressionReasonComplaint   SuppressionReason = "COMPLAINT"
	SuppressionReasonManual      SuppressionReason = "MANUAL"
	SuppressionReasonUnsubscribe SuppressionReason = "UNSUBSCRIBE"
	SuppressionReasonMultipleSoftBounces SuppressionReason = "MULTIPLE_SOFT_BOUNCES"
)

// EmailBounce represents a bounce event
type EmailBounce struct {
	Base
	TeamID         string         `gorm:"type:uuid;not null;index"`
	Team           *Team          `json:"team,omitempty"`
	EmailID        *string        `gorm:"type:uuid;index"`
	Email          *Email         `json:"email,omitempty"`
	ContactID      string         `gorm:"type:uuid;not null;index"`
	Contact        *Contact       `json:"contact,omitempty"`
	EmailAddress   string         `gorm:"not null;index"`
	BounceType     BounceType     `gorm:"not null;index"`
	BounceCategory BounceCategory `gorm:"not null"`
	BounceCode     string         `json:"bounceCode"`     // SMTP error code
	BounceMessage  string         `gorm:"type:text" json:"bounceMessage"` // Full error message
	DiagnosticCode string         `gorm:"type:text" json:"diagnosticCode"`
	ShouldSuppress bool           `gorm:"not null;default:false"` // Should this bounce trigger suppression?
	ProcessedAt    *time.Time     `json:"processedAt"`
	Metadata       datatypes.JSON `gorm:"type:jsonb;default:'{}'"`
}

// SuppressionList represents the suppression list (do not send)
type SuppressionList struct {
	Base
	TeamID        string            `gorm:"type:uuid;not null;index"`
	Team          *Team             `json:"team,omitempty"`
	EmailAddress  string            `gorm:"not null;index"`
	ContactID     *string           `gorm:"type:uuid;index"`
	Contact       *Contact          `json:"contact,omitempty"`
	Reason        SuppressionReason `gorm:"not null;index"`
	Source        string            `json:"source"` // "bounce", "complaint", "manual", "webhook"
	Description   string            `json:"description"`
	AddedAt       time.Time         `gorm:"not null;default:CURRENT_TIMESTAMP;index"`
	ExpiresAt     *time.Time        `json:"expiresAt"` // For temporary suppressions (soft bounces)
	IsActive      bool              `gorm:"not null;default:true;index"`
	ReactivatedAt *time.Time        `json:"reactivatedAt"`
}

// ComplaintReport represents a spam complaint
type ComplaintReport struct {
	Base
	TeamID           string         `gorm:"type:uuid;not null;index"`
	Team             *Team          `json:"team,omitempty"`
	EmailID          *string        `gorm:"type:uuid;index"`
	Email            *Email         `json:"email,omitempty"`
	ContactID        *string        `gorm:"type:uuid;index"`
	Contact          *Contact       `json:"contact,omitempty"`
	EmailAddress     string         `gorm:"not null;index"`
	ComplaintType    string         `json:"complaintType"` // "abuse", "fraud", "not-spam", "virus"
	UserAgent        string         `json:"userAgent"`
	FeedbackType     string         `json:"feedbackType"`
	ArrivalDate      *time.Time     `json:"arrivalDate"`
	ProcessedAt      *time.Time     `json:"processedAt"`
	AutoSuppressed   bool           `gorm:"default:true"`
	Metadata         datatypes.JSON `gorm:"type:jsonb;default:'{}'"`
}

// BounceRule represents rules for auto-suppression
type BounceRule struct {
	Base
	TeamID              string     `gorm:"type:uuid;not null;index"`
	Team                *Team      `json:"team,omitempty"`
	Name                string     `gorm:"not null"`
	Description         string     `json:"description"`
	BounceType          BounceType `gorm:"not null"`
	ThresholdCount      int        `gorm:"not null"` // Number of bounces before suppression
	ThresholdPeriodDays int        `gorm:"not null"` // Time period in days
	AutoSuppress        bool       `gorm:"default:true"`
	SuppressionDuration *int       `json:"suppressionDuration"` // Days (null = permanent)
	IsActive            bool       `gorm:"default:true;index"`
}

// TableName for EmailBounce
func (EmailBounce) TableName() string {
	return "email_bounces"
}

// TableName for SuppressionList
func (SuppressionList) TableName() string {
	return "suppression_list"
}

// TableName for ComplaintReport
func (ComplaintReport) TableName() string {
	return "complaint_reports"
}

// TableName for BounceRule
func (BounceRule) TableName() string {
	return "bounce_rules"
}

// IsPermanent checks if the bounce is permanent
func (b *EmailBounce) IsPermanent() bool {
	return b.BounceType == BounceTypeHard ||
		b.BounceCategory == BounceCategoryInvalidAddress ||
		b.BounceCategory == BounceCategoryNoMailbox ||
		b.BounceCategory == BounceCategoryDomainNotFound
}

// IsActive checks if the suppression is currently active
func (s *SuppressionList) IsActiveNow() bool {
	if !s.IsActive {
		return false
	}

	if s.ExpiresAt != nil && s.ExpiresAt.Before(time.Now()) {
		return false
	}

	return true
}

// DefaultBounceRules returns default bounce rules for a new team
func DefaultBounceRules(teamID string) []BounceRule {
	return []BounceRule{
		{
			TeamID:              teamID,
			Name:                "Hard Bounce - Immediate Suppression",
			Description:         "Suppress contacts immediately after a hard bounce",
			BounceType:          BounceTypeHard,
			ThresholdCount:      1,
			ThresholdPeriodDays: 1,
			AutoSuppress:        true,
			SuppressionDuration: nil, // Permanent
			IsActive:            true,
		},
		{
			TeamID:              teamID,
			Name:                "Soft Bounce - Multiple Failures",
			Description:         "Suppress contacts after 3 soft bounces in 7 days",
			BounceType:          BounceTypeSoft,
			ThresholdCount:      3,
			ThresholdPeriodDays: 7,
			AutoSuppress:        true,
			SuppressionDuration: func() *int { v := 30; return &v }(), // 30 days
			IsActive:            true,
		},
		{
			TeamID:              teamID,
			Name:                "Block - Immediate Suppression",
			Description:         "Suppress contacts immediately after being blocked",
			BounceType:          BounceTypeBlock,
			ThresholdCount:      1,
			ThresholdPeriodDays: 1,
			AutoSuppress:        true,
			SuppressionDuration: func() *int { v := 90; return &v }(), // 90 days
			IsActive:            true,
		},
	}
}

// ClassifyBounce analyzes bounce message and classifies it
func ClassifyBounce(smtpCode string, message string) (BounceType, BounceCategory) {
	// Simplified classification logic
	// In production, use more sophisticated pattern matching

	switch smtpCode {
	case "550", "551", "553":
		return BounceTypeHard, BounceCategoryNoMailbox
	case "552":
		return BounceTypeSoft, BounceCategoryMailboxFull
	case "554":
		return BounceTypeHard, BounceCategorySpamDetected
	default:
		if smtpCode[0:1] == "5" {
			return BounceTypeHard, BounceCategoryUndetermined
		}
		return BounceTypeSoft, BounceCategoryUndetermined
	}
}
