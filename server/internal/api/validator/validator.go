package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	playgroundvalidator "github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// ValidationErrors wraps the validator's ValidationErrors
type ValidationErrors []playgroundvalidator.FieldError

// CustomValidator wraps go-playground/validator
type CustomValidator struct {
	validator *playgroundvalidator.Validate
}

// NewValidator creates a new validator instance
func NewValidator() echo.Validator {
	v := playgroundvalidator.New()

	// Register custom validation tags
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// Register custom validations
	err := v.RegisterValidation("user_role", validateUserRole)
	if err != nil {
		return nil
	}
	err = v.RegisterValidation("invite_status", validateInviteStatus)
	if err != nil {
		return nil
	}
	err = v.RegisterValidation("email_tracking_event", validateEmailTrackingEvent)
	if err != nil {
		return nil
	}
	err = v.RegisterValidation("campaign_status", validateCampaignStatus)
	if err != nil {
		return nil
	}

	return &CustomValidator{validator: v}
}

// Custom validation functions
func validateUserRole(fl playgroundvalidator.FieldLevel) bool {
	role := fl.Field().String()
	return role == "admin" || role == "member"
}

func validateInviteStatus(fl playgroundvalidator.FieldLevel) bool {
	status := fl.Field().String()
	return status == "PENDING" || status == "ACCEPTED" || status == "REJECTED"
}

func validateEmailTrackingEvent(fl playgroundvalidator.FieldLevel) bool {
	event := fl.Field().String()
	validEvents := map[string]bool{
		"click":     true,
		"open":      true,
		"reply":     true,
		"bounce":    true,
		"complaint": true,
	}
	return validEvents[event]
}

func validateCampaignStatus(fl playgroundvalidator.FieldLevel) bool {
	status := fl.Field().String()
	return status == "DRAFT" || status == "SCHEDULED" || status == "RUNNING" || status == "COMPLETED" || status == "FAILED"
}

// Validate implements echo.Validator interface
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		var validationErrors playgroundvalidator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return ValidationErrors(validationErrors)
		}
		return err
	}
	return nil
}

// Error implements the error interface for ValidationErrors
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}
	var fields []string
	for _, err := range ve {
		fields = append(fields, err.Field())
	}
	return fmt.Sprintf("validation failed on fields: %s", strings.Join(fields, ", "))
}

// UserRequest Request validation structs based on models
type UserRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Role      string `json:"role" validate:"required,user_role"`
	TeamID    string `json:"teamId" validate:"required,uuid"`
}

type TeamSettingsRequest struct {
	LogoURL        string `json:"logoUrl"`
	PrimaryColor   string `json:"primaryColor"`
	SecondaryColor string `json:"secondaryColor"`
}

type TeamRequest struct {
	Name     string               `json:"name" validate:"required"`
	Settings *TeamSettingsRequest `json:"settings"`
}

type TeamInviteRequest struct {
	Email     string    `json:"email" validate:"required,email"`
	Name      string    `json:"name" validate:"required"`
	TeamID    string    `json:"teamId" validate:"required,uuid"`
	ExpiresAt time.Time `json:"expiresAt" validate:"required,gt=now"`
}

type ContactRequest struct {
	Email     string                 `json:"email" validate:"required,email"`
	FirstName string                 `json:"firstName"`
	LastName  string                 `json:"lastName"`
	Metadata  map[string]interface{} `json:"metadata"`
	ListID    string                 `json:"listId" validate:"required,uuid"`
	TeamID    string                 `json:"teamId" validate:"required,uuid"`
}

type MailingListRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	TeamID      string `json:"teamId" validate:"required,uuid"`
}

type SMTPConfigRequest struct {
	Provider     string `json:"provider" validate:"required,oneof=CUSTOM GMAIL OUTLOOK AMAZON"`
	Host         string `json:"host" validate:"required,hostname"`
	Port         int    `json:"port" validate:"required,min=1,max=65535"`
	Username     string `json:"username" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=8"`
	IsActive     bool   `json:"isActive"`
	SupportsTLS  bool   `json:"supportsTls"`
	RequiresAuth bool   `json:"requiresAuth"`
	MaxSendRate  int    `json:"maxSendRate" validate:"required,min=1"`
	TeamID       string `json:"teamId" validate:"required,uuid"`
}

type DomainRequest struct {
	Domain string `json:"domain" validate:"required,fqdn"`
	TeamID string `json:"teamId" validate:"required,uuid"`
}

type WebhookRequest struct {
	Name   string   `json:"name" validate:"required"`
	URL    string   `json:"url" validate:"required,url"`
	Events []string `json:"events" validate:"required,min=1,dive,oneof=SENT OPENED CLICKED BOUNCED UNSUBSCRIBED"`
	TeamID string   `json:"teamId" validate:"required,uuid"`
}

type TemplateRequest struct {
	Name       string   `json:"name" validate:"required"`
	Subject    string   `json:"subject" validate:"required"`
	HtmlFile   string   `json:"htmlFile" validate:"required"`
	DesignJSON string   `json:"designJson" validate:"required,json"`
	Variables  []string `json:"variables"`
	CategoryID string   `json:"categoryId" validate:"required,uuid"`
	TeamID     string   `json:"teamId" validate:"required,uuid"`
}

type CampaignRequest struct {
	Name         string    `json:"name" validate:"required"`
	Description  string    `json:"description"`
	TemplateID   string    `json:"templateId" validate:"required,uuid"`
	TeamID       string    `json:"teamId" validate:"required,uuid"`
	Status       string    `json:"status" validate:"required,campaign_status"`
	ScheduledFor time.Time `json:"scheduledFor" validate:"required_if=Status SCHEDULED"`
	ListID       string    `json:"listId" validate:"required,uuid"`
	SMTPConfigID string    `json:"smtpConfigId" validate:"required,uuid"`
}

type APIKeyRequest struct {
	TeamID      string    `json:"teamId" validate:"required,uuid"`
	ExpiresAt   time.Time `json:"expiresAt" validate:"required,gt=now"`
	Permissions []string  `json:"permissions" validate:"required,min=1,dive,oneof=READ WRITE DELETE ADMIN"`
}

type AutomationRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	TeamID      string `json:"teamId" validate:"required,uuid"`
	IsActive    bool   `json:"isActive"`
}

type ModelRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	TeamID      string `json:"teamId" validate:"required,uuid"`
	Provider    string `json:"provider" validate:"required,oneof=OPENAI ANTHROPIC GOOGLE AZURE"`
}
