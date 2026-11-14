package handlers

import (
	"errors"
	"kori/internal/events"
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
	return c.JSON(200, map[string]interface{}{"lists": lists, "templates": templates, "senders": senders, "starters": marketing.Starters(), "turnstileSiteKey": marketing.TurnstileSiteKey()})
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
		// Absent keeps the current web archive setting; "" turns it off.
		PublicSlug *string
		// Absent keeps the current setting; tracking is on for new newsletters.
		TrackEngagement *bool
		// Absent keeps the website's current look.
		SiteTheme, SiteLogoURL *string
		// Absent keeps the current author; "" uses the workspace default.
		AuthorID *string
	}
	if err := c.Bind(&input); err != nil {
		return bad("Invalid newsletter")
	}
	n := models.Newsletter{TeamID: team(c), Name: input.Name, Subject: input.Subject, Description: input.Description, TemplateID: input.TemplateID, ListID: input.ListID, SMTPConfigID: input.SMTPConfigID, Status: input.Status, Cadence: input.Cadence, Timezone: input.Timezone, NextSendAt: input.NextSendAt, PostalAddress: input.PostalAddress, TrackEngagement: input.TrackEngagement}
	if err := h.service.ValidateNewsletter(c.Request().Context(), &n); err != nil {
		return bad(err.Error())
	}
	if input.PublicSlug != nil {
		slug, err := marketing.NormalizePublicSlug(*input.PublicSlug)
		if err != nil {
			return bad(err.Error())
		}
		n.PublicSlug = slug
	}
	n.SiteTheme = "paper"
	if input.SiteTheme != nil {
		if !marketing.SiteThemes[*input.SiteTheme] {
			return bad("Choose a website theme")
		}
		n.SiteTheme = *input.SiteTheme
	}
	if input.SiteLogoURL != nil {
		if !marketing.ValidSiteLogo(strings.TrimSpace(*input.SiteLogoURL)) {
			return bad("The logo must be an https:// image address")
		}
		n.SiteLogoURL = strings.TrimSpace(*input.SiteLogoURL)
	}
	if input.AuthorID != nil {
		id, err := h.authorID(c, *input.AuthorID)
		if err != nil {
			return err
		}
		n.AuthorID = id
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
			if input.PublicSlug == nil {
				n.PublicSlug = old.PublicSlug
			}
			if input.TrackEngagement == nil {
				n.TrackEngagement = old.TrackEngagement
			}
			if input.SiteTheme == nil {
				n.SiteTheme = old.SiteTheme
			}
			if input.SiteLogoURL == nil {
				n.SiteLogoURL = old.SiteLogoURL
			}
			if input.AuthorID == nil {
				n.AuthorID = old.AuthorID
			}
			if n.NextSendAt != nil && old.NextSendAt != nil && n.NextSendAt.Equal(*old.NextSendAt) && n.Timezone == old.Timezone {
				n.MonthDay = old.MonthDay
			}
		}
		return tx.Omit(clause.Associations).Save(&n).Error
	})
	if errors.Is(err, gorm.ErrDuplicatedKey) || (err != nil && strings.Contains(err.Error(), "idx_newsletters_public_slug")) {
		return bad("That web address is already taken; choose another")
	}
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

// ResumeCampaign continues a campaign that staged sending paused. The rest of
// the list is sent without further stages: the sender has seen the warning.
func (h *MarketingHandler) ResumeCampaign(c echo.Context) error {
	var campaign models.Campaign
	if err := h.scoped(c).Where("id = ? AND status = ?", c.Param("id"), models.CampaignStatusPaused).First(&campaign).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return missing()
		}
		return err
	}
	off := false
	if err := h.scoped(c).Model(&campaign).Updates(map[string]any{"send_in_stages": off, "pause_reason": "", "status": models.CampaignStatusSending}).Error; err != nil {
		return err
	}
	campaign.SendInStages, campaign.PauseReason, campaign.Status = &off, "", models.CampaignStatusSending
	if campaign.NewsletterID == nil {
		events.Emit("campaign.created", &campaign) // queues the send, as on launch
	} // the newsletter tick picks editions up again
	return c.JSON(200, map[string]string{"status": string(models.CampaignStatusSending)})
}

// DuplicateCampaign saves a draft copy of a campaign. With audience
// "not_opened" or "not_clicked" the copy goes only to the readers the original
// reached who did not open or click it.
func (h *MarketingHandler) DuplicateCampaign(c echo.Context) error {
	var in struct {
		Audience models.AudienceFilter `json:"audience"`
	}
	if err := c.Bind(&in); err != nil || (in.Audience != "" && in.Audience != models.AudienceNotOpened && in.Audience != models.AudienceNotClicked) {
		return bad("audience must be empty, not_opened or not_clicked")
	}
	campaign, err := h.service.DuplicateCampaign(c.Request().Context(), team(c), c.Param("id"), in.Audience)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return missing()
	}
	if errors.Is(err, marketing.ErrNotSent) {
		return bad(err.Error())
	}
	if err != nil {
		return err
	}
	return c.JSON(201, campaign)
}

// CancelCampaign stops a paused campaign or edition: the rest is never sent.
// CancelCampaign stops a paused campaign's remaining sends, or a scheduled one
// before it starts. The queued task skips completed campaigns.
func (h *MarketingHandler) CancelCampaign(c echo.Context) error {
	result := h.scoped(c).Model(&models.Campaign{}).Where("id = ? AND status IN ?", c.Param("id"), []models.CampaignStatus{models.CampaignStatusPaused, models.CampaignStatusScheduled}).Update("status", models.CampaignStatusCompleted)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return missing()
	}
	return c.JSON(200, map[string]string{"status": string(models.CampaignStatusCompleted)})
}
func (h *MarketingHandler) NewsletterEditions(c echo.Context) error {
	if !h.service.Owns(c.Request().Context(), &models.Newsletter{}, c.Param("id"), team(c)) {
		return missing()
	}
	var rows []struct {
		ID           string     `json:"id"`
		Name         string     `json:"name"`
		ScheduledFor time.Time  `json:"scheduledFor"`
		Status       string     `json:"status"`
		Processed    int        `json:"processed"`
		Sent         int64      `json:"sent"`
		Failed       int64      `json:"failed"`
		NeedsReview  int64      `json:"needsReview"`
		SendInStages *bool      `json:"sendInStages"`
		PauseReason  string     `json:"pauseReason"`
		NextStageAt  *time.Time `json:"nextStageAt"`
	}
	err := h.service.DB.Table("campaigns c").Select("c.id,c.name,c.scheduled_for,c.status,c.processed,(SELECT count(*) FROM emails e WHERE e.campaign_id=c.id AND e.status='SENT') as sent,(SELECT count(*) FROM emails e WHERE e.campaign_id=c.id AND e.status='FAILED') as failed,(SELECT count(*) FROM emails e WHERE e.campaign_id=c.id AND e.status IN ('DELIVERY_UNKNOWN','SENDING')) as needs_review,c.send_in_stages,c.pause_reason,c.next_stage_at").Where("c.newsletter_id = ? AND c.team_id = ?", c.Param("id"), team(c)).Order("c.created_at DESC").Limit(50).Scan(&rows).Error
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
	var contact models.Contact
	if err := marketing.Scope(h.service.DB, c.Param("team")).First(&contact, "id = ?", c.Param("contact")).Error; err != nil {
		return missing()
	}
	consent, done, err := trackingPreference(c, h.service.DB, contact.TeamID, contact.Email)
	if done {
		return err
	}
	if c.Request().Method == http.MethodGet {
		c.Response().Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; form-action 'self'; base-uri 'none'")
		return c.HTML(200, `<!doctype html><html><meta name="viewport" content="width=device-width"><title>Unsubscribe</title><body style="font:16px Arial;max-width:480px;margin:80px auto;padding:24px"><h1>Unsubscribe from this audience</h1><p>You can stop receiving these emails at any time.</p><form method="post"><button style="padding:12px 24px">Confirm unsubscribe</button></form>`+consent+`</body></html>`)
	}
	var unsubscribed *models.EmailEvent
	err = h.service.DB.Transaction(func(tx *gorm.DB) error {
		if err := marketing.Scope(tx, c.Param("team")).Clauses(clause.Locking{Strength: "UPDATE"}).First(&contact, "id = ?", c.Param("contact")).Error; err != nil {
			return err
		}
		if contact.Status != models.SubscriberStatusUnsubscribed {
			unsubscribed = &models.EmailEvent{TeamID: contact.TeamID, ContactID: contact.ID, Address: contact.Email, Source: unsubscribeSource(c)}
		}
		if messageID != "" && contact.Status != models.SubscriberStatusUnsubscribed {
			var message models.Email
			if err := tx.Where("id=? AND team_id=? AND contact_id=?", messageID, contact.TeamID, contact.ID).First(&message).Error; err != nil {
				return err
			}
			unsubscribed.EmailID, unsubscribed.CampaignID = message.ID, message.CampaignID
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
	if unsubscribed != nil {
		events.Emit("email.unsubscribed", unsubscribed)
	}
	return c.HTML(200, "<p>You have been unsubscribed.</p>")
}

// KeepSending answers the list cleanup question. GET only shows a button, so link scanners can't answer for the reader.
func (h *MarketingHandler) KeepSending(c echo.Context) error {
	if !h.service.CheckKeepToken(c.Param("team"), c.Param("contact"), c.Param("token")) {
		return missing()
	}
	if c.Request().Method == http.MethodGet {
		c.Response().Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; form-action 'self'; base-uri 'none'")
		return c.HTML(200, `<!doctype html><html><meta name="viewport" content="width=device-width"><title>Keep sending</title><body style="font:16px Arial;max-width:480px;margin:80px auto;padding:24px"><h1>Still want these emails?</h1><p>Confirm and you'll keep getting them.</p><form method="post"><button style="padding:12px 24px">Yes, keep sending</button></form></body></html>`)
	}
	err := h.service.DB.Transaction(func(tx *gorm.DB) error {
		var contact models.Contact
		if err := marketing.Scope(tx, c.Param("team")).Clauses(clause.Locking{Strength: "UPDATE"}).First(&contact, "id = ?", c.Param("contact")).Error; err != nil {
			return err
		}
		return marketing.KeepSending(tx, &contact, time.Now())
	})
	if err != nil {
		return missing()
	}
	return c.HTML(200, "<p>Thanks, you'll keep getting these emails.</p>")
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
