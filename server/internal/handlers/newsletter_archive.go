package handlers

import (
	"errors"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"kori/internal/marketing"
	"kori/internal/models"
)

type archiveIssue struct {
	ID      string    `json:"id"`
	Subject string    `json:"subject"`
	SentAt  time.Time `json:"sentAt"`
}

// publicNewsletter finds a newsletter its team chose to publish on the web.
func (h *MarketingHandler) publicNewsletter(c echo.Context) (models.Newsletter, error) {
	var n models.Newsletter
	err := h.service.DB.WithContext(c.Request().Context()).
		Where("public_slug = ? AND is_deleted = ?", c.Param("slug"), false).First(&n).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return n, missing()
	}
	return n, err
}

// sentEditions scopes campaigns to finished editions of one newsletter.
func (h *MarketingHandler) sentEditions(c echo.Context, n models.Newsletter) *gorm.DB {
	return h.service.DB.WithContext(c.Request().Context()).Model(&models.Campaign{}).
		Where("newsletter_id = ? AND team_id = ? AND status = ? AND is_deleted = ?", n.ID, n.TeamID, models.CampaignStatusCompleted, false)
}

// PublicNewsletter lists a published newsletter's sent editions, newest first.
func (h *MarketingHandler) PublicNewsletter(c echo.Context) error {
	n, err := h.publicNewsletter(c)
	if err != nil {
		return err
	}
	var rows []struct {
		ID           string
		Subject      string
		ScheduledFor time.Time
	}
	if err := h.sentEditions(c, n).Select("id, subject, scheduled_for").
		Order("scheduled_for DESC").Limit(200).Scan(&rows).Error; err != nil {
		return err
	}
	issues := make([]archiveIssue, 0, len(rows))
	for _, r := range rows {
		issues = append(issues, archiveIssue{ID: r.ID, Subject: marketing.PublicEditionSubject(r.Subject), SentAt: r.ScheduledFor})
	}
	// A published signup form for the same list becomes the page's subscribe link.
	var form struct{ Slug string }
	h.service.DB.WithContext(c.Request().Context()).Model(&models.Form{}).Select("slug").
		Where("team_id = ? AND add_to_list_id = ? AND status = ? AND is_deleted = ? AND published_at IS NOT NULL", n.TeamID, n.ListID, models.FormStatusPublished, false).
		Order("published_at DESC").Limit(1).Scan(&form)
	c.Response().Header().Set("Cache-Control", "public, max-age=300")
	// Explicit projection: no team, list, sender or audience details.
	return c.JSON(200, map[string]interface{}{"name": n.Name, "description": n.Description, "slug": *n.PublicSlug, "subscribeForm": form.Slug, "issues": issues})
}

// PublicNewsletterIssue returns one sent edition with merge fields blanked.
func (h *MarketingHandler) PublicNewsletterIssue(c echo.Context) error {
	n, err := h.publicNewsletter(c)
	if err != nil {
		return err
	}
	var e models.Campaign
	if err := h.sentEditions(c, n).Where("id = ?", c.Param("id")).First(&e).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return missing()
		}
		return err
	}
	c.Response().Header().Set("Cache-Control", "public, max-age=300")
	return c.JSON(200, map[string]interface{}{
		"newsletter": n.Name, "slug": *n.PublicSlug,
		"issue": archiveIssue{ID: e.ID, Subject: marketing.PublicEditionSubject(e.Subject), SentAt: e.ScheduledFor},
		"html":  marketing.PublicEditionHTML(e.HTMLBody),
	})
}
