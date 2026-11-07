package handlers

import (
	"errors"
	"kori/internal/models"
	"kori/internal/services"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/datatypes"
)

type SendEmailRequest struct {
	TemplateID         string                  `json:"templateId"`
	To                 string                  `json:"to" validate:"required,email"`
	Variables          datatypes.JSON          `json:"data" validate:"required,json"`
	SMTPConfigProvider string                  `json:"provider" validate:"omitempty,oneof=CUSTOM GMAIL OUTLOOK AMAZON"`
	SMTPConfigID       string                  `json:"smtpConfigId"`
	Subject            string                  `json:"subject"`
	Body               string                  `json:"html"`
	CC                 string                  `json:"cc"`
	BCC                string                  `json:"bcc"`
	ReplyTo            string                  `json:"replyTo"`
	InReplyTo          string                  `json:"inReplyTo"`
	Attachments        []models.MailAttachment `json:"attachments"`
	RequestID          string                  `json:"requestId"`
	Test               bool                    `json:"test"`
	SendAt             time.Time               `json:"scheduleAt"`
}

// SendEmail sends an email using the provided template and variables
// @Summary Send an email
// @Description Send an email using the provided template and variables
// @Tags Email
// @Accept json
// @Produce json
// @Param request body SendEmailRequest true "Email request"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Success 200 {object} map[string]string
// @Router /emails [post]
func SendEmail(c echo.Context) error {
	var req SendEmailRequest
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, 6*1024*1024)
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}
	if req.RequestID != "" {
		if _, err := uuid.Parse(req.RequestID); err != nil {
			return echo.NewHTTPError(400, "requestId must be a UUID")
		}
	}
	if err := models.ValidateMailAttachments(req.Attachments); err != nil {
		return echo.NewHTTPError(400, err.Error())
	}
	if _, err := mail.ParseAddress(req.To); err != nil {
		return echo.NewHTTPError(400, "A valid recipient is required")
	}
	for _, value := range []string{req.CC, req.BCC, req.ReplyTo} {
		if value != "" {
			if _, err := mail.ParseAddressList(value); err != nil {
				return echo.NewHTTPError(400, "Invalid recipient or reply address")
			}
		}
	}
	if strings.ContainsAny(req.Subject+req.InReplyTo+req.ReplyTo+req.To+req.CC+req.BCC, "\r\n") || len(req.InReplyTo) > 998 {
		return echo.NewHTTPError(400, "Invalid email headers")
	}
	if req.InReplyTo != "" && (!strings.HasPrefix(req.InReplyTo, "<") || !strings.HasSuffix(req.InReplyTo, ">")) {
		return echo.NewHTTPError(400, "Invalid reply Message-ID")
	}
	if req.Body == "" && req.TemplateID == "" {
		return echo.NewHTTPError(400, "Email content or a template is required")
	}
	// Without a template there is no subject to fall back on, and providers
	// reject the message only later, in the queue.
	if req.TemplateID == "" && strings.TrimSpace(req.Subject) == "" {
		return echo.NewHTTPError(400, "A subject is required")
	}

	// Get teamID from context (set by auth middleware)
	teamID := c.Get("teamID").(string)

	email := models.Email{
		TeamID:       teamID,
		TemplateID:   req.TemplateID,
		To:           req.To,
		SMTPConfigID: req.SMTPConfigID,
		Subject:      req.Subject,
		Data:         req.Variables,
		Body:         req.Body,
		CC:           req.CC,
		BCC:          req.BCC,
		ReplyTo:      req.ReplyTo,
		InReplyTo:    req.InReplyTo,
		Attachments:  req.Attachments,
		Test:         req.Test,
		SendAt:       req.SendAt,
	}

	if err := services.QueueAPIEmailRequest(&email, req.RequestID, req.SMTPConfigProvider); err != nil {
		if errors.Is(err, services.ErrEmailRequestConflict) {
			return echo.NewHTTPError(409, err.Error())
		}
		if errors.Is(err, services.ErrEmailSenderUnavailable) {
			return echo.NewHTTPError(400, err.Error())
		}
		if errors.Is(err, services.ErrEmailTemplateUnavailable) {
			return echo.NewHTTPError(404, err.Error())
		}
		return echo.NewHTTPError(500, "Unable to save email to the outbox. Check the sender, template, and workspace configuration.")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "Email saved to outbox",
		"id":     email.ID,
	})
}
