package models

import (
	"fmt"
	"kori/internal/events"
	"kori/internal/utils/crypto"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Team struct {
	Base
	Name            string          `gorm:"not null" json:"name" validate:"required,min=2"`
	Settings        []TeamSettings  `gorm:"foreignKey:TeamID;references:ID;constraint:OnDelete:CASCADE" json:"settings,omitempty"`
	Users           []User          `gorm:"foreignKey:TeamID;references:ID" json:"users,omitempty"`
	MailingLists    []MailingList   `gorm:"foreignKey:TeamID;references:ID;constraint:OnDelete:CASCADE" json:"mailingLists,omitempty"`
	Webhooks        []Webhook       `gorm:"foreignKey:TeamID;references:ID;constraint:OnDelete:CASCADE" json:"webhooks,omitempty"`
	SMTPConfigs     []SMTPConfig    `gorm:"foreignKey:TeamID;references:ID;constraint:OnDelete:CASCADE" json:"smtpConfigs,omitempty"`
	Domains         []Domain        `gorm:"foreignKey:TeamID;references:ID;constraint:OnDelete:CASCADE" json:"domains,omitempty"`
	Invites         []TeamInvite    `gorm:"foreignKey:TeamID;references:ID;constraint:OnDelete:CASCADE" json:"invites,omitempty"`
	APIKeys         []APIKey        `gorm:"foreignKey:TeamID;references:ID;constraint:OnDelete:CASCADE" json:"apiKeys,omitempty"`
	Automations     []Automation    `gorm:"foreignKey:TeamID;references:ID;constraint:OnDelete:CASCADE" json:"automations,omitempty"`
	Models          []Model         `gorm:"foreignKey:TeamID;references:ID;constraint:OnDelete:CASCADE" json:"models,omitempty"`
	EmailCategories []EmailCategory `gorm:"foreignKey:TeamID;references:ID;constraint:OnDelete:CASCADE" json:"emailCategories,omitempty"`
	Campaigns       []Campaign      `gorm:"foreignKey:TeamID;references:ID;constraint:OnDelete:CASCADE" json:"campaigns,omitempty"`
}

func (t *Team) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

func (t *Team) AfterCreate(tx *gorm.DB) error {
	// Create default branding settings
	branding := &BrandingSettings{
		DashboardName: t.Name,
		LogoURL:       "",
	}

	if err := tx.Create(branding).Error; err != nil {
		return err
	}

	// Create team settings with branding reference
	settings := &TeamSettings{
		InviteTemplateID:   "",
		WelcomeTemplateID:  "",
		BrandingSettingsID: branding.ID,
		TeamID:             t.ID,
	}

	if err := tx.Create(settings).Error; err != nil {
		return err
	}

	// Load initial data (skip in test mode)
	if os.Getenv("TEST_MODE") != "true" {
		if err := LoadInitialData(tx, t.ID); err != nil {
			return err
		}
	}

	// Emit team created event
	events.Emit("team.created", t)
	return nil
}

type TeamInvite struct {
	Base
	Email     string       `gorm:"not null" json:"email" validate:"required,email"`
	Name      string       `gorm:"not null" json:"name" validate:"required,min=2"`
	TeamID    string       `gorm:"type:uuid;not null" json:"teamId" validate:"required,uuid"`
	Team      *Team        `json:"team,omitempty"`
	InviterID string       `gorm:"type:uuid;not null" json:"inviterId" validate:"required,uuid"`
	Inviter   *User        `json:"inviter,omitempty"`
	Role      UserRole     `gorm:"not null;default:'MEMBER'" json:"role" validate:"required,oneof=MEMBER ADMIN"`
	Code      string       `gorm:"not null" json:"code" validate:"required=min=4"`
	Status    InviteStatus `gorm:"not null;default:'PENDING'" json:"status" validate:"required,oneof=PENDING ACCEPTED REJECTED"`
	ExpiresAt time.Time    `gorm:"not null" json:"expiresAt" validate:"required,gt=now"`
}

type Contact struct {
	Base
	LifecycleStage string           `gorm:"default:LEAD" json:"lifecycleStage"`
	Email          string           `gorm:"not null" json:"email" validate:"required,email"`
	FirstName      string           `json:"firstName" validate:"omitempty,min=2"`
	LastName       string           `json:"lastName" validate:"omitempty,min=2"`
	Metadata       datatypes.JSON   `gorm:"type:jsonb;default:'{}'" json:"metadata" validate:"omitempty,json"`
	LinkedIn       string           `json:"linkedin" validate:"omitempty,url"`
	Twitter        string           `json:"twitter" validate:"omitempty,url"`
	Facebook       string           `json:"facebook" validate:"omitempty,url"`
	Instagram      string           `json:"instagram" validate:"omitempty,url"`
	Tags           []Tag            `gorm:"many2many:contact_tags;" json:"tags"`
	Country        string           `json:"country" validate:"omitempty"`
	Phone          string           `json:"phone" validate:"omitempty"`
	City           string           `json:"city" validate:"omitempty"`
	State          string           `json:"state" validate:"omitempty"`
	Zip            string           `json:"zip" validate:"omitempty"`
	Address        string           `json:"address" validate:"omitempty"`
	Company        string           `json:"company" validate:"omitempty"`
	ListID         string           `gorm:"type:uuid;not null" json:"listId" validate:"required,uuid"`
	TeamID         string           `gorm:"type:uuid;not null" json:"teamId" validate:"required,uuid"`
	List           *MailingList     `json:"list,omitempty"`
	ImportID       string           `gorm:"type:uuid;default:NULL;" json:"importId" validate:"omitempty,uuid"`
	Import         *ContactImport   `json:"import,omitempty"`
	Status         SubscriberStatus `gorm:"not null;default:'ACTIVE'" json:"status" validate:"required,oneof=ACTIVE UNSUBSCRIBED BOUNCED COMPLAINED"`
}

type ContactImport struct {
	Base
	Status    ContactImportStatus `gorm:"not null;default:'PENDING'" json:"status" validate:"required,oneof=PENDING PROCESSING COMPLETED FAILED"`
	TeamID    string              `gorm:"type:uuid;not null" json:"teamId" validate:"required,uuid"`
	Team      *Team               `json:"team,omitempty"`
	FileID    string              `gorm:"default:NULL;type:uuid;" json:"fileId" validate:"omitempty,uuid"`
	File      *File               `json:"file,omitempty"`
	ListID    string              `gorm:"type:uuid;not null" json:"listId" validate:"required,uuid"`
	List      *MailingList        `json:"list,omitempty"`
	FieldsMap datatypes.JSON      `gorm:"type:jsonb;default:'{}'" json:"fieldsMap" validate:"required,json"`
	Contacts  []Contact           `gorm:"foreignKey:ImportID" json:"contacts,omitempty"`
}

type File struct {
	Base
	TeamID    string `gorm:"type:uuid" json:"teamId" validate:"omitempty,uuid"`
	Team      *Team  `json:"team,omitempty"`
	Path      string `gorm:"not null" json:"path" validate:"required"`
	UserID    string `gorm:"type:uuid;default:NULL" json:"userId" validate:"omitempty,uuid"`
	User      *User  `json:"user,omitempty"`
	Name      string `gorm:"not null" json:"name" validate:"required"`
	Size      int64  `gorm:"not null" json:"size" validate:"required,min=1"`
	Type      string `gorm:"not null" json:"type" validate:"required"`
	SignedURL string `gorm:"-" json:"signedUrl,omitempty"` // Virtual field
}

func (f *File) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		f.ID = uuid.New().String()
	}
	return nil
}

func (f *File) AfterFind(tx *gorm.DB) error {
	registryMu.RLock()
	generator := urlGenerator
	registryMu.RUnlock()

	if generator != nil {
		// Generate URL with 1-hour expiry
		url, err := generator.GetSignedURL(tx.Statement.Context, f.Path, time.Hour)
		if err != nil {
			return fmt.Errorf("failed to generate signed URL: %w", err)
		}
		f.SignedURL = url
	}
	return nil
}

type Tag struct {
	Base
	TeamID string `gorm:"type:uuid;index" json:"teamId"`
	Name   string `gorm:"not null" json:"name" validate:"required,min=1"`
	Value  string `json:"value" validate:"omitempty"`
}

type MailingList struct {
	Base
	Name             string          `gorm:"not null" json:"name" validate:"required,min=2"`
	SubscribersCount int64           `gorm:"not null;default:0" json:"subscribersCount" validate:"min=0"`
	Description      string          `json:"description" validate:"omitempty"`
	TeamID           string          `gorm:"type:uuid;not null" json:"teamId" validate:"required,uuid"`
	Team             *Team           `json:"team,omitempty"`
	ContactImports   []ContactImport `gorm:"foreignKey:ListID" json:"contactImports,omitempty"`
	Contacts         []Contact       `gorm:"foreignKey:ListID" json:"contacts,omitempty"`
}

type SMTPConfig struct {
	Base
	Provider     string `gorm:"not null" json:"provider" validate:"required,oneof=CUSTOM GMAIL OUTLOOK AMAZON"`
	Host         string `gorm:"not null" json:"host" validate:"required,hostname"`
	Port         int    `gorm:"not null" json:"port" validate:"required,min=1,max=65535"`
	Username     string `json:"username" validate:"required"`
	FromEmail    string `json:"fromEmail" validate:"required"`
	Password     string `json:"password" validate:"required,min=8"`
	IsDefault    bool   `gorm:"not null;default:false" json:"isDefault"`
	IsActive     bool   `gorm:"not null;default:false" json:"isActive"`
	SupportsTLS  bool   `gorm:"not null;default:true" json:"supportsTls"`
	RequiresAuth bool   `gorm:"not null;default:true" json:"requiresAuth"`
	MaxSendRate  int    `gorm:"not null;default:10" json:"maxSendRate" validate:"required,min=1"`
	TeamID       string `gorm:"type:uuid;not null" json:"teamId" validate:"required,uuid"`
}

type IMAPConfig struct {
	Base
	Host     string `gorm:"not null" json:"host" validate:"required,hostname"`
	Port     int    `gorm:"not null" json:"port" validate:"required,min=1,max=65535"`
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
	IsActive bool   `gorm:"not null;default:false" json:"isActive"`
	TeamID   string `gorm:"type:uuid;not null" json:"teamId" validate:"required,uuid"`
	Team     *Team  `json:"team,omitempty"`
}

func (s *SMTPConfig) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	password, err := crypto.Encrypt(s.Password)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}
	s.Password = password
	return nil
}

func (s *IMAPConfig) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}

	password, err := crypto.Encrypt(s.Password)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}
	s.Password = password
	return nil
}

func (s *SMTPConfig) BeforeUpdate(tx *gorm.DB) error {
	password, err := crypto.Encrypt(s.Password)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}
	s.Password = password
	return nil
}

func (s *IMAPConfig) BeforeUpdate(tx *gorm.DB) error {
	password, err := crypto.Encrypt(s.Password)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}
	s.Password = password
	return nil
}

func (s *SMTPConfig) AfterFind(tx *gorm.DB) error {
	if s.Provider == "MANAGED" {
		return nil
	}
	password, err := crypto.Decrypt(s.Password)
	if err != nil {
		return fmt.Errorf("failed to decrypt password: %w", err)
	}
	s.Password = password
	return nil
}

func (s *IMAPConfig) AfterFind(tx *gorm.DB) error {
	password, err := crypto.Decrypt(s.Password)
	if err != nil {
		return fmt.Errorf("failed to decrypt password: %w", err)
	}
	s.Password = password
	return nil
}

type Domain struct {
	Base
	Domain     string `gorm:"uniqueIndex;not null" json:"domain" validate:"required,fqdn"`
	IsVerified bool   `gorm:"not null;default:false" json:"isVerified"`
	DNSRecord  string `json:"dnsRecord" validate:"omitempty"`
	TeamID     string `gorm:"type:uuid;not null" json:"teamId" validate:"required,uuid"`
}

type Webhook struct {
	Base
	Name       string         `gorm:"not null" json:"name" validate:"required,min=2"`
	URL        string         `gorm:"not null" json:"url" validate:"required,url"`
	Events     pq.StringArray `gorm:"type:text[]" json:"events" validate:"required,min=1,dive,oneof=click open reply bounce complaint"`
	IsActive   bool           `gorm:"not null;default:false" json:"isActive"`
	Secret     string         `json:"secret" validate:"required,min=16"`
	TeamID     string         `gorm:"type:uuid;not null" json:"teamId" validate:"required,uuid"`
	Deliveries []Delivery     `gorm:"foreignKey:WebhookID" json:"deliveries,omitempty"`
}

type Delivery struct {
	Base
	WebhookID    string         `gorm:"type:uuid;not null" json:"webhookId" validate:"required,uuid"`
	Event        string         `gorm:"not null" json:"event" validate:"required,oneof=click open reply bounce complaint"`
	Payload      datatypes.JSON `gorm:"type:jsonb;not null" json:"payload" validate:"required,json"`
	ResponseCode int            `json:"responseCode" validate:"omitempty,min=100,max=599"`
	ResponseBody string         `json:"responseBody" validate:"omitempty"`
	Error        string         `json:"error" validate:"omitempty"`
	Status       string         `gorm:"not null" json:"status" validate:"required,oneof=PENDING SUCCESS FAILED"`
}

type EmailCategory struct {
	Base
	Name        string     `gorm:"not null" json:"name" validate:"required,min=2"`
	Description string     `json:"description" validate:"omitempty"`
	Emails      []Email    `gorm:"foreignKey:CategoryID" json:"emails,omitempty"`
	Templates   []Template `gorm:"foreignKey:CategoryID" json:"templates,omitempty"`
	TeamID      string     `gorm:"type:uuid;not null" json:"teamId" validate:"required,uuid"`
	Team        *Team      `json:"team,omitempty"`
}

type Template struct {
	Base
	HTMLBody   string         `gorm:"type:text" json:"htmlBody"`
	StarterKey string         `json:"starterKey"`
	Name       string         `gorm:"not null" json:"name" validate:"required,min=2"`
	Subject    string         `gorm:"not null" json:"subject" validate:"required"`
	HtmlFileID string         `gorm:"type:uuid;default:NULL" json:"htmlFileId" validate:"omitempty,uuid"`
	HtmlFile   *File          `json:"htmlFile,omitempty"`
	DesignJSON string         `gorm:"not null;default:''" json:"designJson" validate:"omitempty"`
	TeamID     string         `gorm:"type:uuid;not null" json:"teamId" validate:"required,uuid"`
	Team       *Team          `json:"team,omitempty"`
	Emails     []Email        `gorm:"foreignKey:TemplateID" json:"emails,omitempty"`
	Variables  pq.StringArray `gorm:"type:text[]" json:"variables" validate:"omitempty,dive,min=1"`
	CategoryID string         `gorm:"type:uuid;not null" json:"categoryId" validate:"required,uuid"`
	Category   *EmailCategory `json:"category,omitempty"`
}

type Email struct {
	Base
	ClickTrackingEnabled *bool          `json:"-"`
	DeliveryKey          *string        `gorm:"uniqueIndex" json:"-"`
	UnsubscribeURL       string         `json:"-"`
	From                 string         `gorm:"not null" json:"from" validate:"required,email"`
	To                   string         `gorm:"not null" json:"to" validate:"required,email"`
	Subject              string         `gorm:"not null" json:"subject" validate:"required"`
	Body                 string         `gorm:"not null" json:"body" validate:"required"`
	Status               EmailStatus    `gorm:"not null" json:"status" validate:"required,oneof=DRAFT QUEUED SENDING SENT FAILED"`
	Error                string         `json:"error" validate:"omitempty"`
	Data                 datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"data" validate:"omitempty,json"`
	TemplateID           string         `gorm:"type:uuid;default:NULL" json:"templateId" validate:"omitempty,uuid"`
	Template             *Template      `json:"template,omitempty"`
	TeamID               string         `gorm:"type:uuid;not null" json:"teamId" validate:"required,uuid"`
	Team                 *Team          `json:"team,omitempty"`
	ContactID            string         `gorm:"type:uuid;default:NULL" json:"contactId" validate:"omitempty,uuid"`
	Contact              *Contact       `json:"contact,omitempty"`
	SMTPConfigID         string         `gorm:"type:uuid;not null" json:"smtpConfigId" validate:"required,uuid"`
	SMTPConfig           *SMTPConfig    `json:"smtpConfig,omitempty"`
	SentAt               time.Time      `json:"sentAt" validate:"omitempty"`
	SendAt               time.Time      `json:"sendAt" validate:"omitempty"`
	CategoryID           string         `gorm:"type:uuid;not null" json:"categoryId" validate:"required,uuid"`
	Category             *EmailCategory `json:"category,omitempty"`
	CampaignID           string         `gorm:"type:uuid;default:NULL" json:"campaignId" validate:"omitempty,uuid"`
	Campaign             *Campaign      `json:"campaign,omitempty"`
	CC                   string         `json:"cc" validate:"omitempty,email"`
	BCC                  string         `json:"bcc" validate:"omitempty,email"`
	ReplyTo              string         `json:"replyTo" validate:"omitempty,email"`
	Test                 bool           `gorm:"not null;default:false" json:"test"`
}

func (e *Email) BeforeUpdate(tx *gorm.DB) error {
	e.UpdatedAt = time.Now()
	return nil
}

func (e *Email) AfterUpdate(tx *gorm.DB) error {
	events.Emit("email.updated", e)
	return nil
}

func (e *Email) AfterCreate(tx *gorm.DB) error {
	if e.CampaignID != "" || e.DeliveryKey != nil {
		// If the email is part of a campaign, we don't need to send it immediately this is handled in the campaign handler
		return nil
	}
	smtp, err := GetSMTPConfig(e.TeamID, e.SMTPConfigID, "", tx)
	if err != nil {
		return err
	}
	e.SMTPConfig = smtp
	events.Emit("email.created", e)
	return nil
}

func (e *Email) AfterDelete(tx *gorm.DB) error {
	events.Emit("email.deleted", e)
	return nil
}

type EmailTracking struct {
	Base
	EmailID    string             `gorm:"type:uuid;not null" json:"emailId" validate:"required,uuid"`
	Email      *Email             `json:"email,omitempty"`
	CampaignID string             `gorm:"type:uuid;default:NULL" json:"campaignId" validate:"omitempty,uuid"`
	Campaign   *Campaign          `json:"campaign,omitempty"`
	ContactID  string             `gorm:"type:uuid;default:NULL" json:"contactId" validate:"omitempty,uuid"`
	Contact    *Contact           `json:"contact,omitempty"`
	Event      EmailTrackingEvent `gorm:"not null" json:"event" validate:"required,oneof=click open reply bounce complaint"`
	Timestamp  time.Time          `json:"timestamp" validate:"required"`
	// 🌍 Geographic Data
	IPAddress string `json:"ipAddress" validate:"omitempty,ip"`
	Country   string `json:"country" validate:"omitempty"`
	City      string `json:"city" validate:"omitempty"`
	Region    string `json:"region" validate:"omitempty"`
	// 📱 Device Information
	UserAgent  string `json:"userAgent" validate:"omitempty"`
	DeviceType string `json:"deviceType" validate:"omitempty,oneof=desktop mobile tablet other"`
	Browser    string `json:"browser" validate:"omitempty"`
	OS         string `json:"os" validate:"omitempty"`
	// 🔗 Click Specific Data (for click events)
	URL string `json:"url" validate:"omitempty,url"`
	// 📊 Additional Metadata
	Metadata datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"metadata" validate:"omitempty,json"`
}

type APIKey struct {
	Base
	AssistantUserID string             `gorm:"index" json:"-"`
	AssistantWrite  bool               `json:"-"`
	Name            string             `gorm:"not null" json:"name"`
	Key             string             `gorm:"not null" json:"key"`
	TeamID          string             `gorm:"type:uuid;not null" json:"teamId" validate:"required,uuid"`
	Team            *Team              `json:"team,omitempty"`
	CreatedAt       time.Time          `json:"createdAt"`
	LastUsedAt      time.Time          `json:"lastUsedAt"`
	ExpiresAt       time.Time          `json:"expiresAt"`
	Permissions     []APIKeyPermission `gorm:"foreignKey:KeyID" json:"permissions,omitempty"`
}

func (a *APIKey) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	if a.Key == "" {
		keyWithoutDashes := strings.ReplaceAll(uuid.New().String(), "-", "")
		a.Key = "kori_" + keyWithoutDashes
	}
	if a.ExpiresAt.IsZero() {
		a.ExpiresAt = time.Now().Add(24 * 90 * time.Hour)
	}
	return nil
}

type APIKeyUsage struct {
	Base
	APIKeyID  string    `gorm:"type:uuid;not null" json:"apiKeyId" validate:"required,uuid"`
	APIKey    *APIKey   `json:"apiKey,omitempty" validate:"required"`
	Endpoint  string    `gorm:"not null" json:"endpoint" validate:"required"`
	Method    string    `gorm:"not null" json:"method" validate:"required,oneof=GET POST PUT DELETE"`
	Timestamp time.Time `json:"timestamp" validate:"required"`
	Success   bool      `gorm:"not null;default:true" json:"success" validate:"required"`
	Error     string    `json:"error" validate:"omitempty"`
	IPAddress string    `json:"ipAddress" validate:"omitempty"`
	UserAgent string    `json:"userAgent" validate:"omitempty"`
}

type Campaign struct {
	Base
	NewsletterID      *string                   `gorm:"type:uuid;index" json:"newsletterId,omitempty"`
	HTMLBody          string                    `gorm:"type:text" json:"htmlBody,omitempty"`
	Subject           string                    `json:"subject"`
	PostalAddress     string                    `json:"postalAddress"`
	AudienceCursor    string                    `json:"-"`
	AudienceCutoff    *time.Time                `json:"-"`
	Name              string                    `gorm:"not null" json:"name"`
	Description       string                    `json:"description"`
	TemplateID        string                    `gorm:"type:uuid;not null" json:"templateId"`
	Template          *Template                 `json:"template,omitempty"`
	TeamID            string                    `gorm:"type:uuid;not null" json:"teamId"`
	Team              *Team                     `json:"team,omitempty"`
	Status            CampaignStatus            `gorm:"not null;default:'DRAFT'" json:"status"`
	ScheduledFor      time.Time                 `json:"scheduledFor"`
	Schedule          CampaignSchedule          `json:"schedule"`
	ListID            string                    `gorm:"type:uuid;not null" json:"listId"`
	List              *MailingList              `json:"list,omitempty"`
	RecurringSchedule CampaignRecurringSchedule `json:"recurringSchedule"`
	CronExpression    string                    `json:"cronExpression"`
	SentEmails        []Email                   `gorm:"foreignKey:CampaignID" json:"sentEmails,omitempty"`
	Analytics         []EmailTracking           `gorm:"foreignKey:CampaignID" json:"analytics,omitempty"`
	SMTPConfigID      string                    `gorm:"type:uuid;not null" json:"smtpConfigId"`
	SMTPConfig        *SMTPConfig               `json:"smtpConfig,omitempty"`
	BatchSize         int                       `gorm:"not null;default:100" json:"batchSize"`
	Processed         int                       `gorm:"not null;default:0" json:"processed"`
	BatchDelay        time.Duration             `gorm:"not null;default:3600" json:"batchDelay"` // 1 hour delay between batches
	Timezone          string                    `gorm:"not null;default:'America/New_York'" json:"timezone"`
	Data              datatypes.JSON            `gorm:"type:jsonb;default:'{}'" json:"data" validate:"omitempty,json"`
}
type RateLimit struct {
	Base
	APIKeyID  string        `gorm:"type:uuid;not null" json:"apiKeyId"`
	APIKey    *APIKey       `json:"apiKey,omitempty"`
	Limit     int           `gorm:"not null" json:"limit"`
	Period    time.Duration `gorm:"not null" json:"period"`
	CreatedAt time.Time     `json:"createdAt"`
}

type Automation struct {
	Base
	TriggerEvent string               `gorm:"default:manual" json:"triggerEvent"`
	Name         string               `gorm:"not null" json:"name"`
	Description  string               `json:"description"`
	TeamID       string               `gorm:"type:uuid;not null" json:"teamId"`
	Team         *Team                `json:"team,omitempty"`
	Nodes        []AutomationNode     `gorm:"foreignKey:AutomationID" json:"nodes,omitempty"`
	Edges        []AutomationNodeEdge `gorm:"foreignKey:AutomationID" json:"edges,omitempty"`
	IsActive     bool                 `gorm:"not null;default:false" json:"isActive"`
}

type Model struct {
	Base
	Name        string `gorm:"not null" json:"name"`
	Description string `json:"description"`
	TeamID      string `gorm:"type:uuid;not null" json:"teamId"`
	Team        *Team  `json:"team,omitempty"`
	Provider    string `gorm:"not null" json:"provider"`
}

type LLMEmailWriterJob struct {
	Base
	AutomationID string      `gorm:"type:uuid;not null" json:"automationId"`
	Automation   *Automation `json:"automation,omitempty"`
	EmailID      string      `gorm:"type:uuid;not null" json:"emailId"`
	Email        *Email      `json:"email,omitempty"`
	Status       string      `gorm:"not null" json:"status"`
	CreatedAt    time.Time   `json:"createdAt"`
	StartedAt    time.Time   `json:"startedAt"`
	CompletedAt  time.Time   `json:"completedAt"`
	UpdatedAt    time.Time   `json:"updatedAt"`
	Error        string      `json:"error"`
	Output       string      `json:"output"`
	Input        string      `json:"input"`
	Prompt       string      `json:"prompt"`
	Model        *Model      `json:"model,omitempty"`
	ModelID      string      `gorm:"type:uuid;not null" json:"modelId" validate:"required,uuid"`
}

type AutomationNode struct {
	Base
	AutomationID string               `gorm:"type:uuid;not null" json:"automationId"`
	Automation   *Automation          `json:"automation,omitempty"`
	Type         NodeType             `gorm:"not null" json:"type" validate:"required,oneof=EMAIL_WRITER"`
	Data         datatypes.JSON       `gorm:"type:jsonb" json:"data" validate:"required,json"`
	EdgesFrom    []AutomationNodeEdge `gorm:"foreignKey:SourceID" json:"edgesFrom,omitempty"`
	EdgesTo      []AutomationNodeEdge `gorm:"foreignKey:TargetID" json:"edgesTo,omitempty"`
}

type AutomationNodeEdge struct {
	Base
	AutomationID string          `gorm:"type:uuid;not null" json:"automationId"`
	Automation   *Automation     `json:"automation,omitempty"`
	SourceID     string          `gorm:"type:uuid;not null" json:"sourceId"`
	Source       *AutomationNode `json:"source,omitempty"`
	TargetID     string          `gorm:"type:uuid;not null" json:"targetId"`
	Target       *AutomationNode `json:"target,omitempty"`
	Label        string          `json:"label"`
	Animated     bool            `gorm:"not null;default:true" json:"animated"`
}

type AutomationExecution struct {
	Base
	AutomationID  string          `gorm:"type:uuid;not null" json:"automationId"`
	Automation    *Automation     `json:"automation,omitempty"`
	ContactID     string          `gorm:"type:uuid;not null" json:"contactId"`
	Contact       *Contact        `json:"contact,omitempty"`
	Status        ExecutionStatus `gorm:"not null;default:'RUNNING'" json:"status"`
	CurrentNodeID string          `gorm:"type:uuid" json:"currentNodeId"`
	CurrentNode   *AutomationNode `json:"currentNode,omitempty"`
	ExecutionLog  datatypes.JSON  `gorm:"type:jsonb;default:'[]'" json:"executionLog"`
	Variables     datatypes.JSON  `gorm:"type:jsonb;default:'{}'" json:"variables"`
	StartedAt     time.Time       `json:"startedAt"`
	CompletedAt   time.Time       `json:"completedAt"`
	Error         string          `json:"error"`
	TriggerType   string          `json:"triggerType"`
	TriggerData   datatypes.JSON  `gorm:"type:jsonb;default:'{}'" json:"triggerData"`
}

type AIOptimization struct {
	Base
	AutomationID  string         `gorm:"type:uuid;not null" json:"automationId"`
	Automation    *Automation    `json:"automation,omitempty"`
	Suggestions   datatypes.JSON `gorm:"type:jsonb" json:"suggestions"`
	Applied       bool           `gorm:"not null;default:false" json:"applied"`
	AppliedAt     time.Time      `json:"appliedAt"`
	ImpactMetrics datatypes.JSON `gorm:"type:jsonb" json:"impactMetrics"`
}

// IsValidUserRole checks if a given role is valid
func IsValidUserRole(role UserRole) bool {
	switch role {
	case UserRoleAdmin, UserRoleMember, UserRoleSuperAdmin:
		return true
	default:
		return false
	}
}

func GetEmailByID(id string, db *gorm.DB) (*Email, error) {
	var email Email
	if err := db.Where("id = ?", id).Preload("SMTPConfig").Preload("Contact").First(&email).Error; err != nil {
		return nil, err
	}
	return &email, nil
}
