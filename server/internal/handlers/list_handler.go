package handlers

import (
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"html"
	"kori/internal/config"
	"kori/internal/models"
	"kori/internal/templates"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

// HandleEmailUnsubscribe handles unsubscribe requests from email links
// @Summary Unsubscribe from email list
// @Description Unsubscribe from an email list
// @Accept json
// @Produce json
// @Param token query string true "Unsubscribe token"
// @Success 200 {object} map[string]string "Unsubscribed successfully"
// @Failure 400 {object} map[string]string "Missing token"
// @Failure 401 {object} map[string]string "Invalid token"
// @Router /t/unsubscribe [get]
func (h *TrackingHandler) HandleEmailUnsubscribe(c echo.Context) error {
	// Extract token from query params
	token := c.QueryParam("token")
	if token == "" {
		return c.String(http.StatusBadRequest, "Missing token")
	}

	// Parse JWT token
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("invalid signing method")
		}
		return []byte(config.GetConfig().JWT.Secret), nil
	})

	if err != nil {
		return c.String(http.StatusUnauthorized, "Invalid token")
	}

	// Extract email ID and recipient email from claims
	emailID, ok := claims["mailId"].(string)
	if !ok {
		return c.String(http.StatusBadRequest, "Invalid token claims - missing email ID")
	}

	// Get the email
	email, err := models.GetEmailByID(emailID, h.db)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to get email")
	}

	// Check if contact exists
	if email.Contact == nil {
		return c.String(http.StatusInternalServerError, "Contact not found for this email")
	}

	if c.Request().Method == http.MethodGet {
		return c.HTML(200, `<!doctype html><html><meta name="viewport" content="width=device-width"><title>Unsubscribe</title><body><h1>Unsubscribe</h1><p>Stop receiving messages for `+html.EscapeString(email.Contact.Email)+`?</p><form method="post"><button>Confirm unsubscribe</button></form></body></html>`)
	}
	err = h.db.Transaction(func(tx *gorm.DB) error {
		var contact models.Contact
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=? AND team_id=?", email.ContactID, email.TeamID).First(&contact).Error; err != nil {
			return err
		}
		if contact.Status == models.SubscriberStatusUnsubscribed {
			return nil
		}
		if err := tx.Model(&contact).UpdateColumn("status", models.SubscriberStatusUnsubscribed).Error; err != nil {
			return err
		}
		return tx.Create(&models.EmailTracking{EmailID: email.ID, ContactID: contact.ID, CampaignID: email.CampaignID, Event: models.EmailTrackingEventUnsubscribe, Timestamp: time.Now().UTC()}).Error
	})
	if err != nil {
		return err
	}

	// Return success page
	return c.HTML(http.StatusOK, templates.UnsubscribeTemplate(email.Contact.Email, token))
}

// HandleEmailResubscribe handles resubscribe requests from email links
// @Summary Resubscribe to email list
// @Description Resubscribe to an email list
// @Accept json
// @Produce json
// @Param email query string true "Resubscribe email"
// @Success 200 {object} map[string]string "Resubscribed successfully"
// @Failure 400 {object} map[string]string "Missing email"
// @Failure 401 {object} map[string]string "Invalid email"
// @Router /t/resubscribe [get]
func (h *TrackingHandler) HandleEmailResubscribe(c echo.Context) error {
	token := c.QueryParam("token")
	claims := jwt.MapClaims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("invalid signing method")
		}
		return []byte(config.GetConfig().JWT.Secret), nil
	})
	if err != nil || !parsed.Valid {
		return echo.NewHTTPError(401, "Invalid subscription link")
	}
	id, ok := claims["mailId"].(string)
	if !ok {
		return echo.NewHTTPError(400, "Invalid subscription link")
	}
	email, err := models.GetEmailByID(id, h.db)
	if err != nil || email.Contact == nil {
		return echo.NewHTTPError(404, "Subscription not found")
	}
	if c.Request().Method == http.MethodGet {
		return c.HTML(200, `<!doctype html><html><meta name="viewport" content="width=device-width"><title>Resubscribe</title><body><h1>Resubscribe</h1><p>Resume messages for `+html.EscapeString(email.Contact.Email)+`?</p><form method="post"><button>Confirm resubscription</button></form></body></html>`)
	}
	result := h.db.Model(&models.Contact{}).Where("id=? AND team_id=? AND status=? AND is_deleted=false", email.ContactID, email.TeamID, models.SubscriberStatusUnsubscribed).UpdateColumn("status", models.SubscriberStatusActive)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 && email.Contact.Status != models.SubscriberStatusActive {
		return echo.NewHTTPError(400, "This address cannot be resubscribed through this link")
	}
	// Historical opt-outs are retained; resubscription is a separate status transition.
	return c.HTML(200, templates.ResubscribeTemplate(email.Contact.Email))
}
