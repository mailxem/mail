package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/events"
	"kori/internal/marketing"
	"kori/internal/models"
)

// decodeMarketingInput rejects mass assignment and trailing JSON values.
func decodeMarketingInput(c echo.Context, target any) error {
	d := json.NewDecoder(io.LimitReader(c.Request().Body, 2<<20))
	d.DisallowUnknownFields()
	if err := d.Decode(target); err != nil {
		return bad("Invalid request fields")
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return bad("Expected one JSON object")
	}
	return nil
}

type contactInput struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Phone     string `json:"phone"`
	Company   string `json:"company"`
	Country   string `json:"country"`
	City      string `json:"city"`
}

// ImportContactBatch validates the entire batch and commits it atomically. Locking
// the list serializes these imports with public form capture for the same list.
func (h *MarketingHandler) ImportContactBatch(c echo.Context) error {
	if team(c) == "" {
		return echo.NewHTTPError(401, "Authentication required")
	}
	var in struct {
		ListID   string         `json:"listId"`
		Contacts []contactInput `json:"contacts"`
		DryRun   *bool          `json:"dryRun"`
	}
	if err := decodeMarketingInput(c, &in); err != nil {
		return err
	}
	if uuid.Validate(in.ListID) != nil || len(in.Contacts) < 1 || len(in.Contacts) > 500 {
		return bad("Choose a list and provide 1 to 500 contacts")
	}
	for i := range in.Contacts {
		v := &in.Contacts[i]
		v.Email = strings.ToLower(strings.TrimSpace(v.Email))
		if len(v.Email) > 254 || !marketing.ValidEmail(v.Email) {
			return bad("Every row must contain a valid email")
		}
		for _, value := range []string{v.FirstName, v.LastName, v.Phone, v.Company, v.Country, v.City} {
			if len(value) > 500 || strings.ContainsRune(value, '\x00') {
				return bad("Contact fields must be at most 500 bytes without null characters")
			}
		}
	}
	dryRun := in.DryRun == nil || *in.DryRun
	created := []models.Contact{}
	skipped := 0
	err := h.service.DB.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
		var list models.MailingList
		if err := marketing.Scope(tx, team(c)).Clauses(clause.Locking{Strength: "UPDATE"}).First(&list, "id = ?", in.ListID).Error; err != nil {
			return missing()
		}
		// One lookup for the bounded batch avoids one query per CSV row.
		emails := make([]string, 0, len(in.Contacts))
		for _, row := range in.Contacts {
			emails = append(emails, row.Email)
		}
		var existing []string
		// Include deleted/suppressed records: imports never silently resubscribe.
		if err := tx.Model(&models.Contact{}).Where("team_id = ? AND list_id = ? AND LOWER(email) IN ?", team(c), in.ListID, emails).Pluck("LOWER(email)", &existing).Error; err != nil {
			return err
		}
		seen := map[string]bool{}
		for _, email := range existing {
			seen[email] = true
		}
		for _, row := range in.Contacts {
			if seen[row.Email] {
				skipped++
				continue
			}
			seen[row.Email] = true
			now := time.Now().UTC()
			contact := models.Contact{Base: models.Base{ID: uuid.NewString(), CreatedAt: now, UpdatedAt: now}, TeamID: team(c), ListID: in.ListID, Email: row.Email, FirstName: row.FirstName, LastName: row.LastName, Phone: row.Phone, Company: row.Company, Country: row.Country, City: row.City, Status: models.SubscriberStatusActive, LifecycleStage: "LEAD"}
			if !dryRun {
				if err := tx.Session(&gorm.Session{SkipHooks: true}).Omit(clause.Associations).Create(&contact).Error; err != nil {
					return err
				}
			}
			created = append(created, contact)
		}
		if !dryRun {
			return models.SyncSubscribersCountByID(tx, in.ListID)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if !dryRun {
		for i := range created {
			events.Emit("contact.created", &created[i])
		}
	}
	createdCount := 0
	if !dryRun {
		createdCount = len(created)
	}
	return c.JSON(200, map[string]any{"dryRun": dryRun, "created": createdCount, "wouldCreate": len(created), "skipped": skipped, "total": len(in.Contacts)})
}

// CreateCampaignDraft never queues delivery, including through model hooks.
func (h *MarketingHandler) CreateCampaignDraft(c echo.Context) error {
	if team(c) == "" {
		return echo.NewHTTPError(401, "Authentication required")
	}
	var in struct {
		Name          string `json:"name"`
		Subject       string `json:"subject"`
		Description   string `json:"description"`
		TemplateID    string `json:"templateId"`
		ListID        string `json:"listId"`
		SMTPConfigID  string `json:"smtpConfigId"`
		HTMLBody      string `json:"htmlBody"`
		PostalAddress string `json:"postalAddress"`
	}
	if err := decodeMarketingInput(c, &in); err != nil {
		return err
	}
	n := models.Newsletter{TeamID: team(c), Name: in.Name, Subject: in.Subject, TemplateID: in.TemplateID, ListID: in.ListID, SMTPConfigID: in.SMTPConfigID, Status: "DRAFT", Cadence: "ONCE", Timezone: "UTC"}
	if err := h.service.ValidateNewsletter(c.Request().Context(), &n); err != nil {
		return bad(err.Error())
	}
	if len(in.HTMLBody) > 500000 || len(in.Description) > 2000 || len(in.PostalAddress) > 500 {
		return bad("Campaign content exceeds limits")
	}
	now := time.Now().UTC()
	campaign := models.Campaign{Base: models.Base{ID: uuid.NewString(), CreatedAt: now, UpdatedAt: now}, TeamID: team(c), Name: n.Name, Subject: n.Subject, Description: in.Description, TemplateID: in.TemplateID, ListID: in.ListID, SMTPConfigID: in.SMTPConfigID, HTMLBody: in.HTMLBody, PostalAddress: in.PostalAddress, Status: models.CampaignStatusDraft, Schedule: models.CampaignScheduleOneTime, Timezone: "UTC"}
	if err := h.service.DB.WithContext(c.Request().Context()).Session(&gorm.Session{SkipHooks: true}).Omit(clause.Associations).Create(&campaign).Error; err != nil {
		return err
	}
	return c.JSON(201, campaign)
}

// UnsubscribeContact preserves historical records and updates the audience count.
func (h *MarketingHandler) UnsubscribeContact(c echo.Context) error {
	if team(c) == "" {
		return echo.NewHTTPError(401, "Authentication required")
	}
	err := h.service.DB.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
		var contact models.Contact
		if err := marketing.Scope(tx, team(c)).Clauses(clause.Locking{Strength: "UPDATE"}).First(&contact, "id = ?", c.Param("id")).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return missing()
			}
			return err
		}
		if contact.Status == models.SubscriberStatusActive {
			if err := tx.Model(&contact).UpdateColumn("status", models.SubscriberStatusUnsubscribed).Error; err != nil {
				return err
			}
		}
		if err := models.SyncSubscribersCountByID(tx, contact.ListID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return c.JSON(200, map[string]bool{"unsubscribed": true})
}
