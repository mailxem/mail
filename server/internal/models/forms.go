package models

import (
	"fmt"
	"time"

	"gorm.io/datatypes"
)

// FormType represents the type of form
type FormType string

const (
	FormTypeInline     FormType = "INLINE"      // Embedded in page
	FormTypePopup      FormType = "POPUP"       // Pop-up overlay
	FormTypeLanding    FormType = "LANDING"     // Full landing page
	FormTypeExitIntent FormType = "EXIT_INTENT" // Shows when user tries to leave
	FormTypeScroll     FormType = "SCROLL"      // Shows after scroll percentage
)

// FormStatus represents the status of a form
type FormStatus string

const (
	FormStatusDraft     FormStatus = "DRAFT"
	FormStatusPublished FormStatus = "PUBLISHED"
	FormStatusArchived  FormStatus = "ARCHIVED"
)

// FieldType represents the type of form field
type FieldType string

const (
	FieldTypeText     FieldType = "TEXT"
	FieldTypeEmail    FieldType = "EMAIL"
	FieldTypePhone    FieldType = "PHONE"
	FieldTypeNumber   FieldType = "NUMBER"
	FieldTypeTextarea FieldType = "TEXTAREA"
	FieldTypeSelect   FieldType = "SELECT"
	FieldTypeRadio    FieldType = "RADIO"
	FieldTypeCheckbox FieldType = "CHECKBOX"
	FieldTypeDate     FieldType = "DATE"
	FieldTypeHidden   FieldType = "HIDDEN"
	FieldTypeFile     FieldType = "FILE"
)

// Form represents a form or landing page
type Form struct {
	Base
	TeamID             string     `gorm:"type:uuid;not null;index"`
	Team               *Team      `json:"team,omitempty"`
	Name               string     `gorm:"not null"`
	Description        string     `json:"description"`
	FormType           FormType   `gorm:"not null"`
	Status             FormStatus `gorm:"not null;default:'DRAFT';index"`
	Slug               string     `gorm:"uniqueIndex"`  // For hosted pages: forms.posthoot.com/{slug}
	CustomDomain       *string    `json:"customDomain"` // Optional custom domain
	IsMultiStep        bool       `gorm:"default:false"`
	SuccessMessage     string     `gorm:"type:text" json:"successMessage"`
	SuccessRedirectURL *string    `json:"successRedirectUrl"`

	Definition datatypes.JSON `gorm:"type:jsonb;default:'null'" json:"definition"`
	Version    int            `gorm:"not null;default:1" json:"version"`

	// Styling
	Theme     datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"theme"` // Colors, fonts, etc.
	CustomCSS *string        `gorm:"type:text" json:"customCss"`
	CustomJS  *string        `gorm:"type:text" json:"customJs"`

	// Behavior
	SubmitButtonText   string  `gorm:"default:'Submit'"`
	DoubleOptIn        bool    `gorm:"default:false"`
	DoubleOptInEmailID *string `gorm:"type:uuid"`
	DoubleOptInEmail   *Email  `json:"doubleOptInEmail,omitempty"`

	// Popup/Trigger Settings
	TriggerDelay      *int `json:"triggerDelay"`  // Seconds delay for popup
	TriggerScroll     *int `json:"triggerScroll"` // Scroll percentage (0-100)
	ExitIntentEnabled bool `gorm:"default:false"`
	ShowOnce          bool `gorm:"default:false"` // Show only once per visitor

	// Integration
	AddToListID         *string      `gorm:"type:uuid"`
	AddToList           *MailingList `json:"addToList,omitempty"`
	AddToSegmentID      *string      `gorm:"type:uuid"`
	AddToSegment        *Segment     `json:"addToSegment,omitempty"`
	TriggerAutomationID *string      `gorm:"type:uuid"`
	TriggerAutomation   *Automation  `json:"triggerAutomation,omitempty"`

	// Analytics
	ViewCount       int     `gorm:"default:0"`
	SubmissionCount int     `gorm:"default:0"`
	ConversionRate  float64 `gorm:"type:decimal(5,4);default:0"`

	// Relations
	Fields      []FormField      `json:"fields,omitempty" gorm:"foreignKey:FormID"`
	Submissions []FormSubmission `json:"submissions,omitempty" gorm:"foreignKey:FormID"`

	PublishedAt *time.Time `json:"publishedAt"`
	ArchivedAt  *time.Time `json:"archivedAt"`
}

// FormField represents a field in a form
type FormField struct {
	Base
	FormID       string    `gorm:"type:uuid;not null;index"`
	Form         *Form     `json:"form,omitempty"`
	StepNumber   int       `gorm:"default:1"` // For multi-step forms
	FieldType    FieldType `gorm:"not null"`
	Label        string    `gorm:"not null"`
	Placeholder  *string   `json:"placeholder"`
	HelpText     *string   `json:"helpText"`
	Required     bool      `gorm:"default:false"`
	DisplayOrder int       `gorm:"not null"`

	// Validation
	MinLength         *int    `json:"minLength"`
	MaxLength         *int    `json:"maxLength"`
	Pattern           *string `json:"pattern"` // Regex pattern
	ValidationMessage *string `json:"validationMessage"`

	// Options (for select, radio, checkbox)
	Options datatypes.JSON `gorm:"type:jsonb;default:'[]'" json:"options"` // ["Option 1", "Option 2"]

	// Mapping
	MapToContactField *string `json:"mapToContactField"` // Maps to Contact field (email, first_name, etc.)
	CustomFieldKey    *string `json:"customFieldKey"`    // Store as custom field

	// Conditional Logic
	ConditionalDisplay datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"conditionalDisplay"` // Show/hide based on other fields

	DefaultValue *string `json:"defaultValue"`
	IsHidden     bool    `gorm:"default:false"`
}

// FormSubmission represents a form submission
type FormSubmission struct {
	Base
	FormID       string   `gorm:"type:uuid;not null;index"`
	Form         *Form    `json:"form,omitempty"`
	ContactID    *string  `gorm:"type:uuid;index"`
	Contact      *Contact `json:"contact,omitempty"`
	EmailAddress string   `gorm:"index"`

	Version   int    `gorm:"not null;default:1" json:"version"`
	SessionID string `gorm:"index" json:"sessionId"`
	Consent   bool   `json:"consent"`

	// Data
	FieldData datatypes.JSON `gorm:"type:jsonb;not null" json:"fieldData"` // All form field values

	// Tracking
	IPAddress   *string `json:"ipAddress"`
	UserAgent   *string `json:"userAgent"`
	Referrer    *string `json:"referrer"`
	UTMSource   *string `json:"utmSource"`
	UTMMedium   *string `json:"utmMedium"`
	UTMCampaign *string `json:"utmCampaign"`
	UTMContent  *string `json:"utmContent"`
	UTMTerm     *string `json:"utmTerm"`

	// Double Opt-In
	RequiresConfirmation bool       `gorm:"default:false"`
	ConfirmedAt          *time.Time `json:"confirmedAt"`
	ConfirmationToken    *string    `gorm:"index" json:"confirmationToken"`

	// Processing
	ProcessedAt     *time.Time `json:"processedAt"`
	ProcessingError *string    `gorm:"type:text" json:"processingError"`

	SubmittedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;index"`
}

// LandingPage represents a full landing page (extends Form)
type LandingPage struct {
	Base
	FormID string `gorm:"type:uuid;not null;uniqueIndex"`
	Form   *Form  `json:"form,omitempty"`

	// Page Content
	Title       string         `gorm:"not null"`
	Headline    string         `json:"headline"`
	Subheadline string         `json:"subheadline"`
	HeroImage   *string        `json:"heroImage"`
	VideoURL    *string        `json:"videoUrl"`
	Sections    datatypes.JSON `gorm:"type:jsonb;default:'[]'" json:"sections"` // Rich content sections

	// SEO
	MetaTitle       *string `json:"metaTitle"`
	MetaDescription *string `json:"metaDescription"`
	MetaKeywords    *string `json:"metaKeywords"`
	OGImage         *string `json:"ogImage"`

	// Analytics
	TrackingScripts datatypes.JSON `gorm:"type:jsonb;default:'[]'" json:"trackingScripts"` // GA, FB Pixel, etc.
}

// FormABTest represents an A/B test for forms
type FormABTest struct {
	Base
	TeamID          string     `gorm:"type:uuid;not null;index"`
	Team            *Team      `json:"team,omitempty"`
	Name            string     `gorm:"not null"`
	ControlFormID   string     `gorm:"type:uuid;not null"`
	ControlForm     *Form      `json:"controlForm,omitempty"`
	VariantFormID   string     `gorm:"type:uuid;not null"`
	VariantForm     *Form      `json:"variantForm,omitempty"`
	SplitPercentage int        `gorm:"default:50"` // % traffic to variant
	Status          TestStatus `gorm:"not null;default:'DRAFT'"`
	WinnerFormID    *string    `gorm:"type:uuid"`
	StartedAt       *time.Time `json:"startedAt"`
	EndedAt         *time.Time `json:"endedAt"`
}

// TableName for Form
func (Form) TableName() string {
	return "forms"
}

// TableName for FormField
func (FormField) TableName() string {
	return "form_fields"
}

// TableName for FormSubmission
func (FormSubmission) TableName() string {
	return "form_submissions"
}

// TableName for LandingPage
func (LandingPage) TableName() string {
	return "landing_pages"
}

// TableName for FormABTest
func (FormABTest) TableName() string {
	return "form_ab_tests"
}

// CalculateConversionRate updates the conversion rate
func (f *Form) CalculateConversionRate() {
	if f.ViewCount > 0 {
		f.ConversionRate = float64(f.SubmissionCount) / float64(f.ViewCount)
	}
}

// IsPublished checks if the form is published
func (f *Form) IsPublished() bool {
	return f.Status == FormStatusPublished && f.PublishedAt != nil
}

// GetEmbedCode returns the embed code for the form
func (f *Form) GetEmbedCode(baseURL string) string {
	return fmt.Sprintf(`<script src="%s/embed/%s.js"></script>
<div id="posthoot-form-%s"></div>`, baseURL, f.ID, f.ID)
}

// IsConfirmed checks if the submission is confirmed (for double opt-in)
func (s *FormSubmission) IsConfirmed() bool {
	return !s.RequiresConfirmation || s.ConfirmedAt != nil
}

// DefaultFormFields returns common form field templates
func DefaultFormFields() []struct {
	Label     string
	FieldType FieldType
	Required  bool
	MapTo     string
} {
	return []struct {
		Label     string
		FieldType FieldType
		Required  bool
		MapTo     string
	}{
		{Label: "Email", FieldType: FieldTypeEmail, Required: true, MapTo: "email"},
		{Label: "First Name", FieldType: FieldTypeText, Required: false, MapTo: "first_name"},
		{Label: "Last Name", FieldType: FieldTypeText, Required: false, MapTo: "last_name"},
		{Label: "Company", FieldType: FieldTypeText, Required: false, MapTo: "company"},
		{Label: "Phone", FieldType: FieldTypePhone, Required: false, MapTo: "phone"},
	}
}
