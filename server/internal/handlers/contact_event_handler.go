package handlers

import (
	"encoding/json"
	"errors"
	"kori/internal/events"
	"kori/internal/models"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// ContactEventInput is the body of POST /api/v1/events.
type ContactEventInput struct {
	Email      string                 `json:"email" example:"ada@example.com"`
	ContactID  string                 `json:"contactId"`
	Event      string                 `json:"event" example:"trial_started"`
	Properties map[string]interface{} `json:"properties"`
	// ListID creates the contact in this list when no contact there has the email.
	ListID string `json:"listId"`
}

// RecordContactEvent stores an event from the customer's app on a contact and
// starts the automations triggered by that event name.
// @Summary Record a contact event
// @Description Records something a contact did in your app (for example trial_started). Automations whose trigger is "Event named X" start for that contact, with each scalar property available as {{event_<name>}}. Identify the contact by contactId, or by email; with listId, an unknown email is added to that list first.
// @Tags Events
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param event body ContactEventInput true "Event"
// @Success 202 {object} map[string]interface{} "Event recorded"
// @Failure 400 {object} map[string]string "Validation error"
// @Failure 404 {object} map[string]string "Contact or list not found"
// @Failure 422 {object} map[string]string "Address is suppressed"
// @Failure 429 {object} map[string]string "Rate limit exceeded"
// @Router /api/v1/events [post]
func RecordContactEvent(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var in ContactEventInput
		if err := c.Bind(&in); err != nil {
			return bad("Send a JSON body")
		}
		in.Event = strings.TrimSpace(in.Event)
		in.Email = strings.ToLower(strings.TrimSpace(in.Email))
		if in.Event == "" || utf8.RuneCountInString(in.Event) > 100 || strings.ContainsAny(in.Event, "\r\n\t") {
			return bad("event must be a name of 1–100 characters")
		}
		if (in.Email == "") == (in.ContactID == "") {
			return bad("Send either email or contactId")
		}
		if in.ContactID != "" && uuid.Validate(in.ContactID) != nil {
			return bad("contactId must be a UUID")
		}
		if in.Email != "" {
			if address, err := mail.ParseAddress(in.Email); err != nil || address.Address != in.Email {
				return bad("email is not a valid address")
			}
		}
		if in.ListID != "" && uuid.Validate(in.ListID) != nil {
			return bad("listId must be a UUID")
		}
		if len(in.Properties) > 100 {
			return bad("Send at most 100 properties")
		}
		if in.Properties == nil {
			in.Properties = map[string]interface{}{}
		}
		properties, err := json.Marshal(in.Properties)
		if err != nil {
			return bad("properties must be a JSON object")
		}
		teamID := team(c)
		created := false
		event := models.ContactEvent{TeamID: teamID, Name: in.Event, Properties: properties}
		err = db.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
			var contact models.Contact
			query := tx.Where("team_id = ? AND is_deleted = ?", teamID, false)
			if in.ContactID != "" {
				query = query.Where("id = ?", in.ContactID)
			} else {
				query = query.Where("LOWER(email) = ?", in.Email)
				if in.ListID != "" {
					query = query.Where("list_id = ?", in.ListID)
				}
			}
			err := query.Order("created_at, id").First(&contact).Error
			if errors.Is(err, gorm.ErrRecordNotFound) && in.Email != "" && in.ListID != "" {
				var lists int64
				if err := tx.Model(&models.MailingList{}).Where("id = ? AND team_id = ? AND is_deleted = ?", in.ListID, teamID, false).Count(&lists).Error; err != nil {
					return err
				}
				if lists != 1 {
					return echo.NewHTTPError(404, "List not found")
				}
				var suppressed int64
				if err := tx.Model(&models.SuppressionList{}).Where("team_id = ? AND LOWER(email_address) = ? AND is_active = ? AND is_deleted = ? AND (expires_at IS NULL OR expires_at > ?)", teamID, in.Email, true, false, time.Now().UTC()).Count(&suppressed).Error; err != nil {
					return err
				}
				if suppressed > 0 {
					return echo.NewHTTPError(422, "This address is suppressed")
				}
				contact = models.Contact{TeamID: teamID, ListID: in.ListID, Email: in.Email, Status: models.SubscriberStatusActive, LifecycleStage: "LEAD", Metadata: []byte(`{"source":"api_event"}`)}
				if err := tx.Create(&contact).Error; err != nil {
					return err
				}
				created = true
			} else if errors.Is(err, gorm.ErrRecordNotFound) {
				return echo.NewHTTPError(404, "Contact not found; send listId to add them")
			} else if err != nil {
				return err
			}
			event.ContactID = contact.ID
			return tx.Create(&event).Error
		})
		if err != nil {
			return err
		}
		events.Emit("event.received", &event)
		return c.JSON(202, map[string]interface{}{"id": event.ID, "contactId": event.ContactID, "contactCreated": created})
	}
}
