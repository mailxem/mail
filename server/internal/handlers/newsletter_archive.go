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
	Cover   string    `json:"cover,omitempty"`
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

// sentEditions scopes campaigns to one newsletter's finished editions and the
// one-off campaigns its team chose to publish in the same archive.
func (h *MarketingHandler) sentEditions(c echo.Context, n models.Newsletter) *gorm.DB {
	return h.service.DB.WithContext(c.Request().Context()).Model(&models.Campaign{}).
		Where("(newsletter_id = ? OR archive_newsletter_id = ?) AND team_id = ? AND status = ? AND is_deleted = ?", n.ID, n.ID, n.TeamID, models.CampaignStatusCompleted, false)
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
		HTMLBody     string
		// An edited web page's title and card image, if any.
		PageTitle, PageCover, PageHTML *string
	}
	if err := h.sentEditions(c, n).Select("campaigns.id, subject, scheduled_for, html_body, issue_pages.title AS page_title, issue_pages.cover AS page_cover, issue_pages.html AS page_html").
		Joins("LEFT JOIN issue_pages ON issue_pages.campaign_id = campaigns.id").
		Order("scheduled_for DESC").Limit(200).Scan(&rows).Error; err != nil {
		return err
	}
	issues := make([]archiveIssue, 0, len(rows))
	for _, r := range rows {
		if r.PageTitle != nil {
			r.Subject, r.HTMLBody = *r.PageTitle, *r.PageHTML
		}
		cover := marketing.EditionCover(r.HTMLBody)
		if r.PageCover != nil && *r.PageCover != "" {
			cover = *r.PageCover
		}
		issues = append(issues, archiveIssue{ID: r.ID, Subject: marketing.PublicEditionSubject(r.Subject), SentAt: r.ScheduledFor, Cover: cover})
	}
	var readers int64
	if err := h.service.DB.WithContext(c.Request().Context()).Model(&models.Contact{}).
		Where("team_id = ? AND list_id = ? AND status = ? AND is_deleted = ?", n.TeamID, n.ListID, models.SubscriberStatusActive, false).
		Count(&readers).Error; err != nil {
		return err
	}
	// A published signup form for the same list becomes the page's subscribe link.
	var form struct{ Slug string }
	h.service.DB.WithContext(c.Request().Context()).Model(&models.Form{}).Select("slug").
		Where("team_id = ? AND add_to_list_id = ? AND status = ? AND is_deleted = ? AND published_at IS NOT NULL", n.TeamID, n.ListID, models.FormStatusPublished, false).
		Order("published_at DESC").Limit(1).Scan(&form)
	c.Response().Header().Set("Cache-Control", "public, max-age=300")
	// Explicit projection: no team, list, sender or audience details.
	return c.JSON(200, map[string]interface{}{"name": n.Name, "description": n.Description, "slug": *n.PublicSlug, "subscribeForm": form.Slug, "issues": issues,
		"theme": n.SiteTheme, "logo": n.SiteLogoURL, "readers": readers, "cadence": n.Cadence})
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
	db := h.service.DB.WithContext(c.Request().Context())
	// An edited web page replaces the emailed subject and body here only.
	title, description, body := e.Subject, "", e.HTMLBody
	cover := marketing.EditionCover(body)
	page, err := issuePage(db, e.ID)
	if err != nil {
		return err
	}
	if page != nil {
		title, description, body = page.Title, page.Description, page.HTML
		cover = page.Cover
		if cover == "" {
			cover = marketing.EditionCover(body)
		}
	}
	author, err := marketing.CampaignAuthor(db, e.TeamID, e.AuthorID)
	if err != nil {
		return err
	}
	c.Response().Header().Set("Cache-Control", "public, max-age=300")
	return c.JSON(200, map[string]interface{}{
		"newsletter": n.Name, "slug": *n.PublicSlug, "theme": n.SiteTheme, "logo": n.SiteLogoURL,
		"issue":       archiveIssue{ID: e.ID, Subject: marketing.PublicEditionSubject(title), SentAt: e.ScheduledFor, Cover: cover},
		"description": description,
		"author":      marketing.Byline(author),
		"html":        marketing.PublicEditionHTML(body),
	})
}
