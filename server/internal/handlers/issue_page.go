package handlers

import (
	"errors"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/marketing"
	"kori/internal/models"
)

// issueSite finds a workspace campaign and the newsletter website it's
// published on. sent requires that it has finished sending.
func (h *MarketingHandler) issueSite(c echo.Context, sent bool) (models.Campaign, models.Newsletter, error) {
	var e models.Campaign
	var n models.Newsletter
	if err := h.scoped(c).First(&e, "id = ?", c.Param("id")).Error; err != nil {
		return e, n, missing()
	}
	site := e.NewsletterID
	if site == nil {
		site = e.ArchiveNewsletterID
	}
	if site == nil || (sent && e.Status != models.CampaignStatusCompleted) {
		return e, n, bad("This campaign isn’t on a newsletter website")
	}
	if err := h.scoped(c).Where("public_slug IS NOT NULL").First(&n, "id = ?", *site).Error; err != nil {
		return e, n, bad("This campaign isn’t on a newsletter website")
	}
	return e, n, nil
}

// issuePage is a sent issue's edited web page, or nil if it reads as emailed.
func issuePage(db *gorm.DB, campaignID string) (*models.IssuePage, error) {
	var p models.IssuePage
	err := db.First(&p, "campaign_id = ?", campaignID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// IssueWebPage returns a sent issue's web page next to the version that was
// emailed, which never changes.
func (h *MarketingHandler) IssueWebPage(c echo.Context) error {
	e, n, err := h.issueSite(c, true)
	if err != nil {
		return err
	}
	page, err := issuePage(h.service.DB.WithContext(c.Request().Context()), e.ID)
	if err != nil {
		return err
	}
	return c.JSON(200, map[string]any{
		"newsletter": map[string]any{"id": n.ID, "name": n.Name, "publicSlug": n.PublicSlug, "siteDomain": n.SiteDomain, "siteDomainStatus": n.SiteDomainStatus},
		"emailed":    models.IssuePage{CampaignID: e.ID, Title: marketing.PublicEditionSubject(e.Subject), Cover: marketing.EditionCover(e.HTMLBody), HTML: e.HTMLBody},
		"page":       page,
	})
}

// SaveIssueWebPage changes a sent issue's title, description, card image and
// content on the newsletter website. The email stays as it was sent.
func (h *MarketingHandler) SaveIssueWebPage(c echo.Context) error {
	e, _, err := h.issueSite(c, true)
	if err != nil {
		return err
	}
	var in struct{ Title, Description, Cover, HTML string }
	if err := c.Bind(&in); err != nil {
		return bad("Invalid web page")
	}
	p := models.IssuePage{CampaignID: e.ID, Title: strings.TrimSpace(in.Title), Description: strings.TrimSpace(in.Description), Cover: strings.TrimSpace(in.Cover), HTML: in.HTML, UpdatedAt: time.Now()}
	switch {
	case p.Title == "" || len(p.Title) > 200 || strings.ContainsAny(p.Title, "\r\n"):
		return bad("Enter a title of up to 200 characters")
	case len(p.Description) > 300:
		return bad("Keep the description to 300 characters")
	case !marketing.ValidSiteLogo(p.Cover):
		return bad("The card image must be an https:// image address")
	case strings.TrimSpace(p.HTML) == "" || len(p.HTML) > 1<<20:
		return bad("The page needs some content, up to 1 MB")
	}
	if err := h.service.DB.WithContext(c.Request().Context()).Clauses(clause.OnConflict{UpdateAll: true}).Create(&p).Error; err != nil {
		return err
	}
	return c.JSON(200, p)
}

// RestoreIssueWebPage puts the emailed version back on the website.
func (h *MarketingHandler) RestoreIssueWebPage(c echo.Context) error {
	e, _, err := h.issueSite(c, true)
	if err != nil {
		return err
	}
	if err := h.service.DB.WithContext(c.Request().Context()).Delete(&models.IssuePage{}, "campaign_id = ?", e.ID).Error; err != nil {
		return err
	}
	return c.NoContent(204)
}

// issueViewDays is how many days the day-by-day list goes back.
const issueViewDays = 90

type issueViews struct {
	Total int64       `json:"total"`
	Today int64       `json:"today"`
	Week  int64       `json:"week"`
	Best  *siteCount  `json:"best"`
	Daily []siteCount `json:"daily"`
}

// countIssueViews totals per-day views (Key "YYYY-MM-DD", UTC) up to today,
// and lists every day from the first view, or at most issueViewDays back,
// with days nobody visited as zero.
func countIssueViews(days []siteCount, today time.Time) issueViews {
	today = today.UTC().Truncate(24 * time.Hour)
	week := today.AddDate(0, 0, -6).Format(time.DateOnly)
	byDay := map[string]siteCount{}
	v := issueViews{Daily: []siteCount{}}
	first := today
	for _, d := range days {
		v.Total += d.Views
		if d.Key >= week {
			v.Week += d.Views
		}
		if v.Best == nil || d.Views > v.Best.Views {
			best := d
			v.Best = &best
		}
		byDay[d.Key] = d
		if t, err := time.Parse(time.DateOnly, d.Key); err == nil && t.Before(first) {
			first = t
		}
	}
	v.Today = byDay[today.Format(time.DateOnly)].Views
	if len(days) == 0 {
		return v
	}
	if oldest := today.AddDate(0, 0, 1-issueViewDays); first.Before(oldest) {
		first = oldest
	}
	for t := first; !t.After(today); t = t.AddDate(0, 0, 1) {
		key := t.Format(time.DateOnly)
		d, ok := byDay[key]
		if !ok {
			d = siteCount{Key: key}
		}
		v.Daily = append(v.Daily, d)
	}
	return v
}

// IssueWebViews reports how often a sent issue's web page was viewed: in
// all, today, over the last 7 days, on its best day and day by day (UTC).
func (h *MarketingHandler) IssueWebViews(c echo.Context) error {
	e, n, err := h.issueSite(c, false)
	if err != nil {
		return err
	}
	db := h.service.DB.WithContext(c.Request().Context())
	day := siteDay(db)
	days := []siteCount{}
	// The website tracker records an issue's page as "/{campaign id}".
	if err := db.Model(&models.SiteEvent{}).Where("newsletter_id = ? AND kind = 'view' AND path = ?", n.ID, "/"+e.ID).
		Select(day + " AS key, COUNT(*) AS views, COUNT(DISTINCT session) AS visitors").Group(day).Order("key").Scan(&days).Error; err != nil {
		return err
	}
	return c.JSON(200, countIssueViews(days, time.Now()))
}
