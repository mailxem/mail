package utils

import (
	"fmt"
	"io"
	"log"
	"net/mail"
	"strings"
	"time"

	"github.com/DusanKasan/parsemail"
)

// EmailAttachment represents a single attachment in an email.
type EmailAttachment struct {
	Filename string // The original filename of the attachment.
	Data     []byte // The raw byte data of the attachment.
	MIMEType string // The MIME type of the attachment (e.g., "application/pdf").
}

// ParsedMail represents the structured data extracted from an email.
type ParsedMail struct {
	BodyText      string                   // Plain text version of the email body.
	BodyHTML      string                   // HTML version of the email body.
	ReplyTo       []*mail.Address          // List of Reply-To addresses.
	Cc            []*mail.Address          // List of CC'd addresses.
	To            []*mail.Address          // List of To addresses.
	Bcc           []*mail.Address          // List of BCC'd addresses (usually not available in received mail).
	From          []*mail.Address          // List of From addresses.
	Subject       string                   // The subject of the email.
	Date          time.Time                // The date the email was sent.
	MessageID     string                   // The unique Message-ID of the email.
	Attachments   []EmailAttachment        // A slice of attachments found in the email.
	EmbeddedFiles []parsemail.EmbeddedFile // A slice of embedded images found in the email.
}

// ParseEmail takes an io.Reader containing raw email data and parses it
// into a ParsedMail struct.
func ParseEmail(emailReader io.Reader) (*ParsedMail, error) {
	parsedMail := &ParsedMail{} // Use a pointer to modify it

	msg, err := parsemail.Parse(emailReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read email message: %w", err)
	}

	// Parse basic headers
	parsedMail.Subject = msg.Header.Get("Subject")
	parsedMail.MessageID = msg.Header.Get("Message-ID")

	dateStr := msg.Header.Get("Date")
	if dateStr != "" {
		parsedMail.Date, err = mail.ParseDate(dateStr)
		if err != nil {
			// Log non-fatal error, or decide if this should be fatal
			log.Printf("Warning: Failed to parse date '%s': %v", dateStr, err)
		}
	}

	fromStr := msg.Header.Get("From")
	if fromStr != "" {
		parsedMail.From, err = mail.ParseAddressList(fromStr)
		if err != nil {
			log.Printf("Warning: Failed to parse 'From' addresses '%s': %v", fromStr, err)
		}
	}

	toStr := msg.Header.Get("To")
	if toStr != "" {
		parsedMail.To, err = mail.ParseAddressList(toStr)
		if err != nil {
			log.Printf("Warning: Failed to parse 'To' addresses '%s': %v", toStr, err)
		}
	}

	ccStr := msg.Header.Get("Cc")
	if ccStr != "" {
		parsedMail.Cc, err = mail.ParseAddressList(ccStr)
		if err != nil {
			log.Printf("Warning: Failed to parse 'Cc' addresses '%s': %v", ccStr, err)
		}
	}

	replyToStr := msg.Header.Get("Reply-To")
	if replyToStr != "" {
		parsedMail.ReplyTo, err = mail.ParseAddressList(replyToStr)
		if err != nil {
			log.Printf("Warning: Failed to parse 'Reply-To' addresses '%s': %v", replyToStr, err)
		}
	}

	body := msg.Content

	if msg.HTMLBody != "" {
		parsedMail.BodyHTML = msg.HTMLBody
	} else if msg.TextBody != "" {
		parsedMail.BodyText = msg.TextBody
	} else if body != nil {
		bodyBytes, err := io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("failed to read email body: %w", err)
		}
		parsedMail.BodyText = string(bodyBytes)
	}

	for _, attachment := range msg.Attachments {
		data, err := io.ReadAll(attachment.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to read attachment: %w", err)
		}
		parsedMail.Attachments = append(parsedMail.Attachments, EmailAttachment{
			Filename: attachment.Filename,
			Data:     data,
			MIMEType: attachment.ContentType,
		})
	}

	parsedMail.EmbeddedFiles = msg.EmbeddedFiles
	return parsedMail, nil
}

// Helper function to format addresses for printing (optional)
func FormatAddresses(addrs []*mail.Address) string {
	if len(addrs) == 0 {
		return "N/A"
	}
	var out []string
	for _, a := range addrs {
		out = append(out, a.String())
	}
	return strings.Join(out, ", ")
}

// Helper function to truncate strings for printing (optional)
func Truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}
