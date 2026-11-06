package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/billing"
	"kori/internal/db"
	"kori/internal/events"
	"kori/internal/mailconnect"
	"kori/internal/models"
	"kori/internal/utils/base64"
	"net/mail"
	"os"
	"strings"
	"sync"
	"time"

	"kori/internal/utils/logger"

	"golang.org/x/time/rate"
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
	NotifyFailures bool
	rateLimiter    chan struct{}            // Global concurrency limiter
	smtpRateLimits map[string]chan struct{} // Per-SMTP server rate limiters
	logger         *logger.Logger
}

// NewEmailHandler creates a new EmailHandler with rate limiting
func NewEmailHandler(maxSendRate int) *EmailHandler {
	maxSendRate = validSendRate(maxSendRate)
	return &EmailHandler{
		NotifyFailures: os.Getenv("SERVICE_NOTIFICATIONS_ENABLED") == "true",
		rateLimiter:    make(chan struct{}, maxSendRate),
		smtpRateLimits: make(map[string]chan struct{}),
		logger:         logger.New("EMAIL_HANDLER"),
	}
}

func validSendRate(rate int) int {
	if rate < 1 {
		return 1
	}
	return rate
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
	// Managed sending enforces its own account limits.
	if !email.Test && email.SMTPConfig.Provider != "MANAGED" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := billing.AllowSend(ctx, email.TeamID, recipientCount(email.To, email.CC, email.BCC))
		cancel()
		if err != nil {
			email.Error = err.Error()
			email.Status = models.EmailStatusFailed
			if dbErr := h.UpdateEmail(email); dbErr != nil {
				return h.logger.Error("❌ Failed to update email status", dbErr)
			}
			return err
		}
	}

	h.logger.Info("📧 Sending email to: %s using SMTP server: %s", email.To, email.SMTPConfig.Host)

	// Create new message
	m := gomail.NewMessage()
	m.SetHeader("From", email.From)
	m.SetHeader("To", email.To)
	m.SetHeader("Subject", email.Subject)
	sender, err := mail.ParseAddress(email.From)
	if err != nil {
		return fmt.Errorf("invalid sender address")
	}
	messageDomain := sender.Address[strings.LastIndexByte(sender.Address, '@')+1:]
	m.SetHeader("Message-ID", "<"+email.ID+"@"+messageDomain+">")
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
	if email.InReplyTo != "" {
		m.SetHeader("In-Reply-To", email.InReplyTo)
		m.SetHeader("References", email.InReplyTo)
	}

	if email.CC != "" {
		addresses, err := mail.ParseAddressList(email.CC)
		if err != nil {
			return fmt.Errorf("invalid Cc recipients")
		}
		values := make([]string, len(addresses))
		for i, address := range addresses {
			values[i] = address.String()
		}
		m.SetHeader("Cc", values...)
	}

	if email.BCC != "" {
		addresses, err := mail.ParseAddressList(email.BCC)
		if err != nil {
			return fmt.Errorf("invalid Bcc recipients")
		}
		values := make([]string, len(addresses))
		for i, address := range addresses {
			values[i] = address.String()
		}
		m.SetHeader("Bcc", values...)
	}

	// Decode base64 body
	decodedBody, err := base64.DecodeFromBase64(email.Body)
	if err != nil {
		return fmt.Errorf("❌ failed to decode email body: %w", err)
	}
	m.SetBody("text/html", decodedBody)
	if err := addMailAttachments(m, email.Attachments); err != nil {
		return err
	}

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
	var deliveryErr error
	var providerStatus string
	if email.SMTPConfig.Provider == mailconnect.Cloudflare {
		providerStatus, deliveryErr = sendCloudflareEmail(email, m, decodedBody)
	} else {
		deliveryErr = sendSecureSMTP(m, email)
	}
	if err := deliveryErr; err != nil {
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
		tracked := strings.Contains(decodedBody, "/t/click/") || strings.Contains(decodedBody, "/t/open?token=")
		email.ClickTrackingEnabled = &tracked
	}
	email.SentAt = time.Now()
	email.Status = models.EmailStatusSent
	if providerStatus != "" {
		email.Status = models.EmailStatus(providerStatus)
	}
	email.Error = ""

	if err := h.UpdateEmail(email); err != nil {
		return fmt.Errorf("delivery outcome unknown: failed to persist SMTP acknowledgement: %w", err)
	}

	h.logger.Success("✅ Email sent successfully to: %s", email.To)

	return nil
}

func sendCloudflareEmail(email *models.Email, message *gomail.Message, html string) (string, error) {
	if email.CampaignID != "" || email.UnsubscribeURL != "" {
		return "", fmt.Errorf("Cloudflare does not support marketing or campaign email")
	}
	var category models.EmailCategory
	if err := db.GetDB().Where("id = ? AND team_id = ?", email.CategoryID, email.TeamID).First(&category).Error; err != nil || category.Name != "Transactional" {
		return "", fmt.Errorf("Cloudflare only supports transactional email")
	}
	c, err := mailconnect.Find(db.GetDB(), email.TeamID, email.SMTPConfigID, false)
	if err != nil || c == nil || c.Provider != mailconnect.Cloudflare {
		return "", fmt.Errorf("Cloudflare sender is disconnected")
	}
	if email.From != c.Address {
		return "", fmt.Errorf("sender must match the connected Cloudflare sender")
	}
	var token string
	if err := mailconnect.Open(c, &token); err != nil {
		return "", fmt.Errorf("unable to load Cloudflare credentials")
	}
	headers := map[string]string{}
	for _, name := range []string{"Message-ID", "In-Reply-To", "References"} {
		if values := message.GetHeader(name); len(values) > 0 {
			headers[name] = values[0]
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()
	attachments := make([]mailconnect.CloudflareAttachment, 0, len(email.Attachments))
	for _, file := range email.Attachments {
		attachments = append(attachments, mailconnect.CloudflareAttachment{Filename: file.Filename, Content: file.Content, Type: file.ContentType, Disposition: "attachment"})
	}
	result, err := mailconnect.SendCloudflare(ctx, mailconnect.HTTPClient(), c.AccountID, token, mailconnect.CloudflareMessage{From: c.Address, To: message.GetHeader("To"), CC: message.GetHeader("Cc"), BCC: message.GetHeader("Bcc"), Subject: email.Subject, HTML: html, ReplyTo: email.ReplyTo, Headers: headers, Attachments: attachments})
	if err != nil {
		return "", err
	}
	raw, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("delivery outcome unknown: unable to encode provider response")
	}
	email.ProviderResult = raw
	return result.Status(), nil
}

// maxSendWorkers caps how many emails one batch sends at once.
const maxSendWorkers = 50

// SendBatchEmails sends multiple emails in parallel, at most MaxSendRate per
// second and at most MaxSendRate (capped at maxSendWorkers) at a time.
func (h *EmailHandler) SendBatchEmails(emails []*models.Email, smtpConfig *models.SMTPConfig) []BatchEmailResult {
	h.logger.Info("📤 Starting to send batch emails, total: %d", len(emails))
	sendRate := validSendRate(smtpConfig.MaxSendRate)
	// ponytail: one SMTP connection per email (sendSecureSMTP dials each time);
	// reuse per worker if connection setup dominates.
	results := sendPooled(emails, min(sendRate, maxSendWorkers), rate.NewLimiter(rate.Limit(sendRate), 1), func(e *models.Email) error {
		h.logger.Info("📧 Sending email to: %s", e.To)
		e.SMTPConfig = smtpConfig
		err := h.SendEmail(e)
		if err != nil {
			h.logger.Error("❌ Failed to send email, error: %v", err)
		} else {
			h.logger.Success("✅ Email sent successfully to: %s", e.To)
		}
		return err
	})
	h.logger.Info("✅ Finished sending batch emails")
	return results
}

// sendPooled sends emails with a fixed number of workers, each waiting on the
// limiter before every send. Results keep the order of emails.
func sendPooled(emails []*models.Email, workers int, limiter interface{ Wait(context.Context) error }, send func(*models.Email) error) []BatchEmailResult {
	results := make([]BatchEmailResult, len(emails))
	next := make(chan int)
	var wg sync.WaitGroup
	for range max(workers, 1) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range next {
				err := limiter.Wait(context.Background())
				if err == nil {
					err = send(emails[i])
				}
				results[i] = BatchEmailResult{Email: emails[i], Error: err}
			}
		}()
	}
	for i := range emails {
		next <- i
	}
	close(next)
	wg.Wait()
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

	err := db.GetDB().Transaction(func(tx *gorm.DB) error {
		// Save only the email: saving its loaded SMTPConfig would re-encrypt the
		// stored password and fail, leaving failed emails stuck as PENDING.
		if err := tx.Omit(clause.Associations).Updates(email).Error; err != nil {
			return err
		}
		// Managed sends are counted by the managed worker, which sees the real outcome.
		if email.Status == models.EmailStatusSent && !email.Test && (email.SMTPConfig == nil || email.SMTPConfig.Provider != "MANAGED") {
			if err := billing.CountEmails(tx, email.TeamID, recipientCount(email.To, email.CC, email.BCC), time.Now()); err != nil {
				return err
			}
		}
		if h.NotifyFailures && !email.Test && (email.SMTPConfig == nil || email.SMTPConfig.Provider != "MANAGED") {
			return models.QueueSendingNotice(tx, email.TeamID, "smtp", email.ID, string(email.Status), email.Error, time.Now())
		}
		return nil
	})
	// Managed sends report their outcome through provider feedback instead.
	if name := map[models.EmailStatus]string{models.EmailStatusSent: "email.sent", models.EmailStatusFailed: "email.failed", models.EmailStatusBounced: "email.bounced"}[email.Status]; err == nil && name != "" && !email.Test {
		sent := *email
		events.Emit(name, &sent)
	}
	return err
}

// recipientCount counts the addresses an email went to, which is what is billed.
func recipientCount(lists ...string) int {
	n := 0
	for _, list := range lists {
		for _, a := range strings.Split(list, ",") {
			if strings.TrimSpace(a) != "" {
				n++
			}
		}
	}
	return n
}
