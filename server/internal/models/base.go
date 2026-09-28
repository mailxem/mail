package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Base contains common columns for all tables
type Base struct {
	ID        string    `gorm:"type:uuid;primary_key" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt time.Time `gorm:"index;default:NULL" json:"-" validate:"omitempty"`
	IsDeleted bool      `json:"isDeleted" default:"false"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (base *Base) BeforeCreate(tx *gorm.DB) error {
	if base.ID == "" {
		base.ID = uuid.New().String()
	}
	return nil
}

// CampaignStatus Status enums
type CampaignStatus string
type CampaignSchedule string
type CampaignRecurringSchedule string
type EmailStatus string
type JobStatus string
type SubscriberStatus string
type TrackingType string
type SMTPProvider string
type NodeType string
type ExecutionStatus string

// Campaign status constants
const (
	CampaignStatusDraft     CampaignStatus = "DRAFT"
	CampaignStatusScheduled CampaignStatus = "SCHEDULED"
	CampaignStatusSending   CampaignStatus = "SENDING"
	CampaignStatusCompleted CampaignStatus = "COMPLETED"
	CampaignStatusFailed    CampaignStatus = "FAILED"
	CampaignStatusPaused    CampaignStatus = "PAUSED"
)

// Campaign schedule constants
const (
	CampaignScheduleOneTime   CampaignSchedule = "ONE_TIME"
	CampaignScheduleRecurring CampaignSchedule = "RECURRING"
)

// Campaign recurring schedule constants
const (
	CampaignRecurringScheduleDaily   CampaignRecurringSchedule = "DAILY"
	CampaignRecurringScheduleWeekly  CampaignRecurringSchedule = "WEEKLY"
	CampaignRecurringScheduleMonthly CampaignRecurringSchedule = "MONTHLY"
	CampaignRecurringScheduleCustom  CampaignRecurringSchedule = "CUSTOM"
)

// Email status constants
const (
	EmailStatusPending EmailStatus = "PENDING"
	EmailStatusSent    EmailStatus = "SENT"
	EmailStatusFailed  EmailStatus = "FAILED"
	EmailStatusBounced EmailStatus = "BOUNCED"
	EmailStatusOpened  EmailStatus = "OPENED"
	EmailStatusClicked EmailStatus = "CLICKED"
)

// Job status constants
const (
	JobStatusQueued     JobStatus = "QUEUED"
	JobStatusProcessing JobStatus = "PROCESSING"
	JobStatusCompleted  JobStatus = "COMPLETED"
	JobStatusFailed     JobStatus = "FAILED"
	JobStatusCancelled  JobStatus = "CANCELLED"
)

// Subscriber status constants
const (
	SubscriberStatusActive       SubscriberStatus = "ACTIVE"
	SubscriberStatusUnsubscribed SubscriberStatus = "UNSUBSCRIBED"
	SubscriberStatusBounced      SubscriberStatus = "BOUNCED"
	SubscriberStatusComplained   SubscriberStatus = "COMPLAINED"
)

// Tracking type constants
const (
	TrackingTypeClicked TrackingType = "CLICKED"
	TrackingTypeOpened  TrackingType = "OPENED"
	TrackingTypeBounced TrackingType = "BOUNCED"
	TrackingTypeFailed  TrackingType = "FAILED"
)

// SMTP provider constants
const (
	SMTPProviderCustom  SMTPProvider = "CUSTOM"
	SMTPProviderGmail   SMTPProvider = "GMAIL"
	SMTPProviderOutlook SMTPProvider = "OUTLOOK"
	SMTPProviderAmazon  SMTPProvider = "AMAZON"
)

// Node type constants
const (
	NodeTypeStart            NodeType = "START"
	NodeTypeEmail            NodeType = "EMAIL"
	NodeTypeWait             NodeType = "WAIT"
	NodeTypeCondition        NodeType = "CONDITION"
	NodeTypeWebhook          NodeType = "WEBHOOK"
	NodeTypeAddToList        NodeType = "ADD_TO_LIST"
	NodeTypeRemoveFromList   NodeType = "REMOVE_FROM_LIST"
	NodeTypeUpdateSubscriber NodeType = "UPDATE_SUBSCRIBER"
	NodeTypeCheckEngagement  NodeType = "CHECK_ENGAGEMENT"
	NodeTypeSegment          NodeType = "SEGMENT"
	NodeTypeTag              NodeType = "TAG"
	NodeTypeUnsubscribe      NodeType = "UNSUBSCRIBE"
	NodeTypeCustomCode       NodeType = "CUSTOM_CODE"
	NodeTypeExit             NodeType = "EXIT"
	NodeTypeAIDecision       NodeType = "AI_DECISION"
	NodeTypeAISuggest        NodeType = "AI_SUGGEST"
)

type UserRole string

const (
	UserRoleSuperAdmin UserRole = "SUPER_ADMIN"
	UserRoleAdmin      UserRole = "ADMIN"
	UserRoleMember     UserRole = "MEMBER"
)

type InviteStatus string

const (
	InviteStatusPending  InviteStatus = "PENDING"
	InviteStatusAccepted InviteStatus = "ACCEPTED"
	InviteStatusRejected InviteStatus = "REJECTED"
)

type TeamSettings struct {
	Base
	InviteTemplateID   string            `json:"inviteTemplateId"`
	WelcomeTemplateID  string            `json:"welcomeTemplateId"`
	BrandingSettingsID string            `gorm:"type:uuid;not null" json:"brandingSettingsId"`
	BrandingSettings   *BrandingSettings `gorm:"constraint:OnDelete:CASCADE" json:"branding,omitempty"`
	TeamID             string            `gorm:"type:uuid;uniqueIndex;not null" json:"teamId"`
}

type BrandingSettings struct {
	Base
	DashboardName string         `gorm:"not null" json:"dashboardName"`
	LogoURL       string         `json:"logoUrl"`
	TeamSettings  []TeamSettings `gorm:"foreignKey:BrandingSettingsID;references:ID" json:"-"`
}

type ContactImportStatus string

const (
	ContactImportStatusPending   ContactImportStatus = "PENDING"
	ContactImportStatusCompleted ContactImportStatus = "COMPLETED"
	ContactImportStatusFailed    ContactImportStatus = "FAILED"
)

type EmailTrackingEvent string

const (
	EmailTrackingEventClick       EmailTrackingEvent = "click"
	EmailTrackingEventOpen        EmailTrackingEvent = "open"
	EmailTrackingEventReply       EmailTrackingEvent = "reply"
	EmailTrackingEventBounce      EmailTrackingEvent = "bounce"
	EmailTrackingEventComplaint   EmailTrackingEvent = "complaint"
	EmailTrackingEventUnsubscribe EmailTrackingEvent = "unsubscribe"
)

// Execution status constants
const (
	ExecutionStatusRunning   ExecutionStatus = "RUNNING"
	ExecutionStatusCompleted ExecutionStatus = "COMPLETED"
	ExecutionStatusFailed    ExecutionStatus = "FAILED"
	ExecutionStatusPaused    ExecutionStatus = "PAUSED"
	ExecutionStatusWaiting   ExecutionStatus = "WAITING"
)
