package utils

import (
	"fmt"
	"kori/internal/db"
	"kori/internal/models"
	"kori/internal/utils/base64"
	"strings"
	"sync"
	"time"

	"kori/internal/utils/logger"

	"gopkg.in/gomail.v2"
)

// EmailConfig holds the configuration for the email server
type EmailConfig struct {
	Host         string
	Port         int
	Username     string
	Password     string
	MaxSendRate  int
	SupportsTLS  bool
	RequiresAuth bool
}

// BatchEmailResult represents the result of sending a batch of emails
type BatchEmailResult struct {
	Email *models.Email
	Error error
}

// EmailHandler handles sending emails via SMTP
type EmailHandler struct {
	rateLimiter    chan struct{}            // Global concurrency limiter
	smtpRateLimits map[string]chan struct{} // Per-SMTP server rate limiters
	logger         *logger.Logger
}

// NewEmailHandler creates a new EmailHandler with rate limiting
func NewEmailHandler(maxSendRate int) *EmailHandler {
	return &EmailHandler{
		rateLimiter:    make(chan struct{}, maxSendRate),
		smtpRateLimits: make(map[string]chan struct{}),
		logger:         logger.New("EMAIL_HANDLER"),
	}
}

// Registered once during startup; all delivery paths share recipient policy.
var ManagedDelivery func(*models.Email, *gomail.Message) error
var RecipientPolicy func(string, []string) error

// SendEmail sends a single email using the configured SMTP server
func (h *EmailHandler) SendEmail(email *models.Email) error {
	if email == nil {
		return fmt.Errorf("email is nil")
	}

	if email.Status != models.EmailStatusPending && email.Status != models.EmailStatusFailed {
		return fmt.Errorf("email is not pending or failed")
	}

	if email.SMTPConfig == nil || email.SMTPConfig.TeamID != email.TeamID || !email.SMTPConfig.IsActive {
		return fmt.Errorf("SMTP config is nil")
	}

	h.logger.Info("📧 Sending email to: %s using SMTP server: %s", email.To, email.SMTPConfig.Host)

	// Create new message
	m := gomail.NewMessage()
	m.SetHeader("From", email.From)
	m.SetHeader("To", email.To)
	m.SetHeader("Subject", email.Subject)
	m.SetHeader("Message-ID", "<"+email.ID+"@posthoot.local>")
	if !email.CreatedAt.IsZero() {
		m.SetHeader("Date", email.CreatedAt.UTC().Format(time.RFC1123Z))
	}
	if email.UnsubscribeURL != "" {
		m.SetHeader("List-Unsubscribe", "<"+email.UnsubscribeURL+">")
		m.SetHeader("List-Unsubscribe-Post", "List-Unsubscribe=One-Click")
	}

	if email.ReplyTo != "" {
		m.SetHeader("Reply-To", email.ReplyTo)
	}

	if email.CC != "" {
		m.SetHeader("Cc", strings.Split(email.CC, ",")...)
	}

	if email.BCC != "" {
		m.SetHeader("Bcc", strings.Split(email.BCC, ",")...)
	}

	// Decode base64 body
	decodedBody, err := base64.DecodeFromBase64(email.Body)
	if err != nil {
		return fmt.Errorf("❌ failed to decode email body: %w", err)
	}
	m.SetBody("text/html", decodedBody)

	if RecipientPolicy != nil {
		var recipients []string
		for _, header := range []string{"To", "Cc", "Bcc"} {
			recipients = append(recipients, m.GetHeader(header)...)
		}
		if err := RecipientPolicy(email.TeamID, recipients); err != nil {
			return err
		}
	}
	if email.SMTPConfig.Provider == "MANAGED" {
		if ManagedDelivery == nil {
			return fmt.Errorf("managed sending is disabled")
		}
		return ManagedDelivery(email, m)
	}
	// Send email
	if err := sendSecureSMTP(m, email); err != nil {
		email.Error = err.Error()
		email.Status = models.EmailStatusFailed
		if strings.Contains(err.Error(), "delivery outcome unknown") {
			email.Status = "DELIVERY_UNKNOWN"
		}
		if dbErr := h.UpdateEmail(email); dbErr != nil {
			return h.logger.Error("❌ Failed to update email status", dbErr)
		}
		return h.logger.Error("❌ failed to send email: %w", err)
	}

	if email.ClickTrackingEnabled == nil {
		tracked := strings.Contains(decodedBody, "/t/click/")
		email.ClickTrackingEnabled = &tracked
	}
	email.SentAt = time.Now()
	email.Status = models.EmailStatusSent
	email.Error = ""

	if err := h.UpdateEmail(email); err != nil {
		return fmt.Errorf("delivery outcome unknown: failed to persist SMTP acknowledgement: %w", err)
	}

	h.logger.Success("✅ Email sent successfully to: %s", email.To)

	return nil
}

// SendBatchEmails sends multiple emails in parallel with rate limiting
func (h *EmailHandler) SendBatchEmails(emails []*models.Email, smtpConfig *models.SMTPConfig) []BatchEmailResult {
	results := make([]BatchEmailResult, len(emails))
	var wg sync.WaitGroup

	h.logger.Info("📤 Starting to send batch emails, total: %d", len(emails))

	safeBatchSize := min(len(emails), smtpConfig.MaxSendRate)

	for i := 0; i < len(emails); i += safeBatchSize {
		end := min(i+safeBatchSize, len(emails))
		batchEmails := emails[i:end]

		for _, email := range batchEmails {
			wg.Add(1)
			go func(index int, e *models.Email) {
				defer wg.Done()
				h.logger.Info("📧 Sending email to: %s", e.To)
				e.SMTPConfig = smtpConfig
				err := h.SendEmail(e)
				if err != nil {
					h.logger.Error("❌ Failed to send email, error: %v", err)
				} else {
					h.logger.Success("✅ Email sent successfully to: %s", e.To)
				}
				results[index] = BatchEmailResult{
					Email: e,
					Error: err,
				}
				time.Sleep(time.Second * 1)
			}(i, email)
		}
	}

	wg.Wait()
	h.logger.Info("✅ Finished sending batch emails")
	return results
}

// SendCampaignEmails sends campaign emails in batches
func (h *EmailHandler) SendCampaignEmails(emails []*models.Email, batchSize int, delay time.Duration, smtpConfig *models.SMTPConfig) []BatchEmailResult {
	totalEmails := len(emails)
	results := make([]BatchEmailResult, totalEmails)

	h.logger.Info("📊 Starting to send campaign emails in batches, total: %d, batch size: %d", totalEmails, batchSize)

	// Process in batches
	for i := 0; i < totalEmails; i += batchSize {
		end := min(i+batchSize, totalEmails)

		h.logger.Info("📦 Sending batch from %d to %d", i, end)
		// Send batch
		batchResults := h.SendBatchEmails(emails[i:end], smtpConfig)

		// Copy batch results to final results
		copy(results[i:end], batchResults)

		// Optional: Add delay between batches to prevent overwhelming the SMTP server
		if end < totalEmails {
			h.logger.Info("⏳ Waiting for %v before sending the next batch", delay)
			time.Sleep(delay)
		}
	}

	h.logger.Info("✅ Finished sending campaign emails")
	return results
}

func (h *EmailHandler) UpdateEmail(email *models.Email) error {
	if email == nil {
		return fmt.Errorf("email is nil")
	}

	return db.GetDB().Updates(email).Error
}
