package utils

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/mail"
	"strings"

	"gopkg.in/gomail.v2"
	"kori/internal/config"
	"kori/internal/models"
)

// SendServiceEmail uses operator SMTP, so onboarding does not depend on the
// customer's still-unconfigured sender. It shares TLS and DATA outcome handling.
func SendServiceEmail(ctx context.Context, cfg config.SMTPConfig, key, recipient, subject, html, text string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	from, err := mail.ParseAddress(cfg.FromEmail)
	if err != nil {
		return err
	}
	if strings.ContainsAny(subject+recipient, "\r\n") {
		return fmt.Errorf("invalid service email header")
	}
	message := gomail.NewMessage()
	message.SetAddressHeader("From", from.Address, "Xem")
	message.SetHeader("To", recipient)
	message.SetHeader("Subject", subject)
	message.SetHeader("Auto-Submitted", "auto-generated")
	message.SetHeader("X-Auto-Response-Suppress", "All")
	message.SetHeader("Message-ID", fmt.Sprintf("<%x@%s>", sha256.Sum256([]byte(key)), strings.Split(from.Address, "@")[1]))
	message.SetBody("text/plain", text)
	message.AddAlternative("text/html", html)
	return sendSecureSMTP(message, &models.Email{From: from.Address, SMTPConfig: &models.SMTPConfig{Host: cfg.Host, Port: cfg.Port, Username: cfg.User, Password: cfg.Password}})
}
