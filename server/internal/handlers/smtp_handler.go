package handlers

import (
	"kori/internal/models"
	"kori/internal/utils"
	"net/http"

	"kori/internal/utils/logger"

	"github.com/labstack/echo/v4"
	"gopkg.in/gomail.v2"
)

var log = logger.New("smtp_handler")

type SMTPHandler struct{}

type SMTPTestRequest struct {
	Host       string `json:"host" validate:"required"`
	Port       int    `json:"port" validate:"required"`
	Username   string `json:"username" validate:"required"`
	Password   string `json:"password" validate:"required"`
	From       string `json:"from" validate:"required"`
	RequireTLS bool   `json:"requireTls" default:"false"`
	RequireSSL bool   `json:"requireSSL" default:"false"`
	To         string `json:"to" validate:"omitempty"`
}

func NewSMTPHandler() *SMTPHandler {
	return &SMTPHandler{}
}

// TestSMTPConnection tests SMTP connection with provided credentials
func (h *SMTPHandler) TestSMTPConnection(c echo.Context) error {
	var req SMTPTestRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Create a new message
	m := gomail.NewMessage()
	m.SetHeader("From", req.From)
	if req.To != "" {
		m.SetHeader("To", req.To)
	} else {
		m.SetHeader("To", req.From)
	}
	m.SetHeader("Subject", "Test Email from Xem")
	m.SetBody("text/html", "Hello, this is a test email from Xem!")

	if req.Port != 465 && req.Port != 587 {
		return echo.NewHTTPError(400, "Use port 587 with STARTTLS or port 465 with TLS")
	}
	config := &models.SMTPConfig{Host: req.Host, Port: req.Port, Username: req.Username, Password: req.Password}
	if err := utils.TestSecureSMTP(m, &models.Email{From: req.From, SMTPConfig: config}); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "SMTP test failed. Check hostname, credentials, and TLS configuration.")
	}

	log.Success("SMTP connection test successful")
	return c.JSON(http.StatusOK, map[string]string{
		"message": "SMTP connection test successful",
	})
}
