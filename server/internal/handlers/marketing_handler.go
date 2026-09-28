package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/marketing"
	"kori/internal/models"
)

type MarketingHandler struct{ service *marketing.Service }

func NewMarketingHandler(db *gorm.DB, secret, url string) *MarketingHandler {
	return &MarketingHandler{marketing.New(db, secret, url)}
}
func team(c echo.Context) string { v, _ := c.Get("teamID").(string); return v }
func (h *MarketingHandler) scoped(c echo.Context) *gorm.DB {
	return marketing.Scope(h.service.DB.WithContext(c.Request().Context()), team(c))
}
func bad(message string) error { return echo.NewHTTPError(400, message) }
func missing() error           { return echo.NewHTTPError(404, "Not found") }
func (h *MarketingHandler) Options(c echo.Context) error {
	var lists []models.MailingList
	var templates []models.Template
	if err := h.scoped(c).Select("id", "name", "subscribers_count").Find(&lists).Error; err != nil {
		return err
	}
	if err := h.scoped(c).Select("id", "name", "subject", "html_body", "html_file_id", "starter_key").Find(&templates).Error; err != nil {
		return err
	}
	var senders []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		FromEmail   string `json:"fromEmail"`
		SupportsTLS bool   `json:"supportsTLS"`
	}
	if err := h.scoped(c).Model(&models.SMTPConfig{}).Select("id", "provider AS name", "from_email", "supports_tls").Find(&senders).Error; err != nil {
		return err
	}
	return c.JSON(200, map[string]interface{}{"lists": lists, "templates": templates, "senders": senders, "starters": marketing.Starters()})
}
func (h *MarketingHandler) Newsletters(c echo.Context) error {
	var rows []models.Newsletter
	if err := h.scoped(c).Order("created_at DESC").Limit(200).Find(&rows).Error; err != nil {
		return err
	}
	return c.JSON(200, rows)
}
func (h *MarketingHandler) SaveNewsletter(c echo.Context) error {
	var input struct {
		Name, Subject, Description, TemplateID, ListID, SMTPConfigID, Status, Cadence, Timezone, PostalAddress string
		NextSendAt                                                                                             *time.Time
	}
	if err := c.Bind(&input); err != nil {
		return bad("Invalid newsletter")
	}
	n := models.Newsletter{TeamID: team(c), Name: input.Name, Subject: input.Subject, Description: input.Description, TemplateID: input.TemplateID, ListID: input.ListID, SMTPConfigID: input.SMTPConfigID, Status: input.Status, Cadence: input.Cadence, Timezone: input.Timezone, NextSendAt: input.NextSendAt, PostalAddress: input.PostalAddress}
	if err := h.service.ValidateNewsletter(c.Request().Context(), &n); err != nil {
		return bad(err.Error())
	}
	if n.NextSendAt != nil {
		loc, _ := time.LoadLocation(n.Timezone) // Validated above.
		n.MonthDay = n.NextSendAt.In(loc).Day()
	}
	err := h.service.DB.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
		if id := c.Param("id"); id != "" {
			var old models.Newsletter
			if err := marketing.Scope(tx, team(c)).Clauses(clause.Locking{Strength: "UPDATE"}).First(&old, "id = ?", id).Error; err != nil {
				return missing()
			}
			n.Base, n.Editions, n.LastSentAt = old.Base, old.Editions, old.LastSentAt
			if n.NextSendAt != nil && old.NextSendAt != nil && n.NextSendAt.Equal(*old.NextSendAt) && n.Timezone == old.Timezone {
				n.MonthDay = old.MonthDay
			}
		}
		return tx.Omit(clause.Associations).Save(&n).Error
	})
	if err != nil {
		return err
	}
	return c.JSON(200, n)
}
func (h *MarketingHandler) PauseNewsletter(c echo.Context) error {
	result := h.scoped(c).Model(&models.Newsletter{}).Where("id = ?", c.Param("id")).Update("status", "PAUSED")
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return missing()
	}
	return c.JSON(200, map[string]string{"status": "PAUSED"})
}
func (h *MarketingHandler) NewsletterEditions(c echo.Context) error {
	if !h.service.Owns(c.Request().Context(), &models.Newsletter{}, c.Param("id"), team(c)) {
		return missing()
	}
	var rows []struct {
		ID           string    `json:"id"`
		Name         string    `json:"name"`
		ScheduledFor time.Time `json:"scheduledFor"`
		Status       string    `json:"status"`
		Processed    int       `json:"processed"`
		Sent         int64     `json:"sent"`
		Failed       int64     `json:"failed"`
		NeedsReview  int64     `json:"needsReview"`
	}
	err := h.service.DB.Table("campaigns c").Select("c.id,c.name,c.scheduled_for,c.status,c.processed,(SELECT count(*) FROM emails e WHERE e.campaign_id=c.id AND e.status='SENT') as sent,(SELECT count(*) FROM emails e WHERE e.campaign_id=c.id AND e.status='FAILED') as failed,(SELECT count(*) FROM emails e WHERE e.campaign_id=c.id AND e.status IN ('DELIVERY_UNKNOWN','SENDING')) as needs_review").Where("c.newsletter_id = ? AND c.team_id = ?", c.Param("id"), team(c)).Order("c.created_at DESC").Limit(50).Scan(&rows).Error
	if err != nil {
		return err
	}
	return c.JSON(200, rows)
}

// ImportTemplateStarter adds a predefined design to the existing template library.
func (h *MarketingHandler) ImportTemplateStarter(c echo.Context) error {
	var input struct {
		StarterKey string `json:"starterKey"`
	}
	if err := c.Bind(&input); err != nil {
		return bad("Choose a starter template")
	}
	var starter *marketing.Starter
	for _, candidate := range marketing.Starters() {
		if candidate.Key == input.StarterKey {
			starter = &candidate
			break
		}
	}
	if starter == nil {
		return bad("Unknown starter template")
	}
	var template models.Template
	err := h.service.DB.Transaction(func(tx *gorm.DB) error {
		var category models.EmailCategory
		if err := marketing.Scope(tx, team(c)).Where("name = ?", "Marketing").FirstOrCreate(&category, models.EmailCategory{Name: "Marketing", TeamID: team(c)}).Error; err != nil {
			return err
		}
		template = models.Template{TeamID: team(c), CategoryID: category.ID, Name: starter.Name, Subject: starter.Subject, HTMLBody: starter.HTMLBody, StarterKey: starter.Key, DesignJSON: starter.DesignJSON}
		return tx.Omit(clause.Associations).Create(&template).Error
	})
	if err != nil {
		return err
	}
	return c.JSON(201, template)
}

func (h *MarketingHandler) Unsubscribe(c echo.Context) error {
	messageID := c.QueryParam("message")
	valid := h.service.CheckToken(c.Param("team"), c.Param("contact"), c.Param("token"))
	if messageID != "" {
		valid = h.service.CheckMessageToken(c.Param("team"), c.Param("contact"), messageID, c.Param("token"))
	}
	if !valid {
		return missing()
	}
	if c.Request().Method == http.MethodGet {
		c.Response().Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; form-action 'self'; base-uri 'none'")
		return c.HTML(200, `<!doctype html><html><meta name="viewport" content="width=device-width"><title>Unsubscribe</title><body style="font:16px Arial;max-width:480px;margin:80px auto;padding:24px"><h1>Unsubscribe from this audience</h1><p>You can stop receiving these emails at any time.</p><form method="post"><button style="padding:12px 24px">Confirm unsubscribe</button></form></body></html>`)
	}
	var contact models.Contact
	err := h.service.DB.Transaction(func(tx *gorm.DB) error {
		if err := marketing.Scope(tx, c.Param("team")).Clauses(clause.Locking{Strength: "UPDATE"}).First(&contact, "id = ?", c.Param("contact")).Error; err != nil {
			return err
		}
		if messageID != "" && contact.Status != models.SubscriberStatusUnsubscribed {
			var message models.Email
			if err := tx.Where("id=? AND team_id=? AND contact_id=?", messageID, contact.TeamID, contact.ID).First(&message).Error; err != nil {
				return err
			}
			if err := tx.Create(&models.EmailTracking{EmailID: message.ID, CampaignID: message.CampaignID, ContactID: contact.ID, Event: models.EmailTrackingEventUnsubscribe, Timestamp: time.Now().UTC()}).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&contact).UpdateColumn("status", models.SubscriberStatusUnsubscribed).Error; err != nil {
			return err
		}
		return models.SyncSubscribersCountByID(tx, contact.ListID)
	})
	if err != nil {
		return missing()
	}
	return c.HTML(200, "<p>You have been unsubscribed.</p>")
}
func (h *MarketingHandler) Notes(c echo.Context) error {
	id := c.Param("id")
	if !h.service.Owns(c.Request().Context(), &models.Contact{}, id, team(c)) {
		return missing()
	}
	if c.Request().Method == "POST" {
		var in struct{ Body string }
		if err := c.Bind(&in); err != nil || len(strings.TrimSpace(in.Body)) == 0 || len(in.Body) > 5000 {
			return bad("A note of up to 5,000 characters is required")
		}
		author, _ := c.Get("userID").(string)
		n := models.ContactNote{TeamID: team(c), ContactID: id, Body: strings.TrimSpace(in.Body), AuthorID: author}
		if err := h.service.DB.Create(&n).Error; err != nil {
			return err
		}
		return c.JSON(201, n)
	}
	var notes []models.ContactNote
	if err := h.scoped(c).Where("contact_id = ?", id).Order("created_at DESC").Limit(100).Find(&notes).Error; err != nil {
		return err
	}
	return c.JSON(200, notes)
}

func (h *MarketingHandler) TemplatePreview(c echo.Context) error {
	var template models.Template
	if err := h.scoped(c).First(&template, "id = ?", c.Param("id")).Error; err != nil {
		return missing()
	}
	body, err := marketing.TemplateHTML(h.service.DB.WithContext(c.Request().Context()), &template)
	if err != nil {
		return bad("Template preview unavailable; open the template editor to check its content")
	}
	return c.JSON(200, map[string]string{"htmlBody": body})
}

// UpdateContactStage extends existing contacts without replacing contact CRUD.
func (h *MarketingHandler) UpdateContactStage(c echo.Context) error {
	var in struct {
		LifecycleStage string `json:"lifecycleStage"`
	}
	if err := c.Bind(&in); err != nil {
		return bad("Choose a lifecycle stage")
	}
	switch in.LifecycleStage {
	case "LEAD", "QUALIFIED", "CUSTOMER", "LOST":
	default:
		return bad("Choose a lifecycle stage")
	}
	result := h.scoped(c).Model(&models.Contact{}).Where("id = ?", c.Param("id")).UpdateColumn("lifecycle_stage", in.LifecycleStage)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return missing()
	}
	return c.JSON(200, map[string]string{"lifecycleStage": in.LifecycleStage})
}
