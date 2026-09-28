package xem

// EmailProvider represents the email provider type
type EmailProvider string

const (
	// EmailProviderCustom is the default custom provider
	EmailProviderCustom   EmailProvider = "CUSTOM"
	EmailProviderSendGrid EmailProvider = "SENDGRID"
	EmailProviderMailgun  EmailProvider = "AMAZON"
)

// SendEmailParams contains parameters for sending an email
type SendEmailParams struct {
	// To is the recipient email address (required)
	To string `json:"to"`

	// Subject is the email subject (required)
	Subject string `json:"subject"`

	// HTML is the HTML content of the email
	HTML *string `json:"html,omitempty"`

	// TemplateID is the template ID to use (alternative to HTML)
	TemplateID *string `json:"templateId,omitempty"`

	// Data is the template data for dynamic content
	Data map[string]interface{} `json:"data,omitempty"`

	// BCC are the BCC recipients
	BCC *string `json:"bcc,omitempty"`

	// CC are the CC recipients
	CC *string `json:"cc,omitempty"`

	// ReplyTo is the reply-to address
	ReplyTo *string `json:"replyTo,omitempty"`

	// Provider is the email provider to use
	Provider *EmailProvider `json:"provider,omitempty"`

	// ScheduleAt schedules the email for future delivery (ISO 8601 format)
	ScheduleAt *string `json:"scheduleAt,omitempty"`

	// Test indicates test mode - email won't actually be sent
	Test *bool `json:"test,omitempty"`
}

// SendEmailResponse represents the response from sending an email
type SendEmailResponse struct {
	Success   bool                   `json:"success"`
	MessageID *string                `json:"messageId,omitempty"`
	Extra     map[string]interface{} `json:"-"`
}

// APIError represents an API error response
type APIError struct {
	Message    string                 `json:"message"`
	StatusCode int                    `json:"statusCode,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

// Error implements the error interface for APIError
func (e *APIError) Error() string {
	return e.Message
}

// Config represents SDK configuration options
type Config struct {
	// APIKey for authentication
	APIKey string

	// BaseURL for the API (default: https://api.xem.email/api/v1)
	BaseURL string

	// Timeout in milliseconds (default: 30000)
	Timeout int
}
