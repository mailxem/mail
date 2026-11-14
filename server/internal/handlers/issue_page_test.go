package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"kori/internal/models"
)

func issueFixture(t *testing.T) (*gorm.DB, *MarketingHandler, models.Newsletter, models.Campaign) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Newsletter{}, &models.Campaign{}, &models.Form{}, &models.Contact{}, &models.IssuePage{}, &models.Author{}, &models.SiteEvent{}))
	team, slug := uuid.NewString(), "founder-notes"
	n := models.Newsletter{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "Founder Notes", ListID: uuid.NewString(), PublicSlug: &slug}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&n).Error)
	e := models.Campaign{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "Edition one", Subject: "Edition one", HTMLBody: `<p>Hi {{first_name}}, we shipped.</p><img src="https://cdn.example.com/a.png">`, Status: models.CampaignStatusCompleted, NewsletterID: &n.ID, TemplateID: uuid.NewString(), ListID: n.ListID, ScheduledFor: time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC)}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&e).Error)
	return db, NewMarketingHandler(db, "secret", "https://xem.email"), n, e
}

// call runs a handler as the campaign's workspace, or publicly with a slug.
func call(t *testing.T, fn echo.HandlerFunc, method, body string, params ...string) (int, map[string]any) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c := echo.New().NewContext(req, rec)
	var names, values []string
	for i := 0; i+1 < len(params); i += 2 {
		if params[i] == "teamID" {
			c.Set("teamID", params[i+1])
			continue
		}
		names, values = append(names, params[i]), append(values, params[i+1])
	}
	c.SetParamNames(names...)
	c.SetParamValues(values...)
	if err := fn(c); err != nil {
		return err.(*echo.HTTPError).Code, nil
	}
	out := map[string]any{}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

func TestEditingASentIssuesWebPageKeepsTheEmailedVersion(t *testing.T) {
	db, h, n, e := issueFixture(t)
	slug := *n.PublicSlug

	code, _ := call(t, h.SaveIssueWebPage, "PUT", `{"title":"Edition one, revised","description":"What we shipped in October.","cover":"https://cdn.example.com/b.png","html":"<p>We shipped, and fixed a typo.</p>"}`, "id", e.ID, "teamID", e.TeamID)
	require.Equal(t, 200, code)

	// The website shows the edit, on the issue and in the archive.
	_, pub := call(t, h.PublicNewsletterIssue, "GET", "", "slug", slug, "id", e.ID)
	require.Equal(t, "Edition one, revised", pub["issue"].(map[string]any)["subject"])
	require.Equal(t, "https://cdn.example.com/b.png", pub["issue"].(map[string]any)["cover"])
	require.Equal(t, "What we shipped in October.", pub["description"])
	require.Equal(t, "<p>We shipped, and fixed a typo.</p>", pub["html"])
	_, list := call(t, h.PublicNewsletter, "GET", "", "slug", slug)
	require.Equal(t, "Edition one, revised", list["issues"].([]any)[0].(map[string]any)["subject"])

	// The emailed version is untouched and returned alongside the edit.
	var stored models.Campaign
	require.NoError(t, db.First(&stored, "id = ?", e.ID).Error)
	require.Equal(t, e.Subject, stored.Subject)
	require.Equal(t, e.HTMLBody, stored.HTMLBody)
	_, both := call(t, h.IssueWebPage, "GET", "", "id", e.ID, "teamID", e.TeamID)
	require.Equal(t, "Edition one", both["emailed"].(map[string]any)["title"])
	require.Equal(t, e.HTMLBody, both["emailed"].(map[string]any)["html"])
	require.Equal(t, "https://cdn.example.com/a.png", both["emailed"].(map[string]any)["cover"])
	require.Equal(t, "Edition one, revised", both["page"].(map[string]any)["title"])

	// Restoring puts the emailed version back on the website.
	code, _ = call(t, h.RestoreIssueWebPage, "DELETE", "", "id", e.ID, "teamID", e.TeamID)
	require.Equal(t, 204, code)
	_, pub = call(t, h.PublicNewsletterIssue, "GET", "", "slug", slug, "id", e.ID)
	require.Equal(t, "Edition one", pub["issue"].(map[string]any)["subject"])
	require.Equal(t, "<p>Hi, we shipped.</p><img src=\"https://cdn.example.com/a.png\">", pub["html"])

	// Only sent issues on a website, in the same workspace, can be edited.
	code, _ = call(t, h.SaveIssueWebPage, "PUT", `{"title":"x","html":"<p>x</p>"}`, "id", e.ID, "teamID", uuid.NewString())
	require.Equal(t, 404, code)
	code, _ = call(t, h.SaveIssueWebPage, "PUT", `{"title":"x","cover":"javascript:alert(1)","html":"<p>x</p>"}`, "id", e.ID, "teamID", e.TeamID)
	require.Equal(t, 400, code)
	require.NoError(t, db.Model(&stored).Update("status", models.CampaignStatusSending).Error)
	code, _ = call(t, h.SaveIssueWebPage, "PUT", `{"title":"x","html":"<p>x</p>"}`, "id", e.ID, "teamID", e.TeamID)
	require.Equal(t, 400, code)
}

func TestIssueWebPageShowsItsAuthorsByline(t *testing.T) {
	db, h, n, e := issueFixture(t)
	a := models.Author{Base: models.Base{ID: uuid.NewString()}, TeamID: e.TeamID, Name: "Asha Rao", PhotoURL: "https://cdn.example.com/asha.png", BylineWeb: true, IsDefault: true}
	require.NoError(t, db.Create(&a).Error)
	_, pub := call(t, h.PublicNewsletterIssue, "GET", "", "slug", *n.PublicSlug, "id", e.ID)
	require.Equal(t, map[string]any{"name": "Asha Rao", "photo": "https://cdn.example.com/asha.png"}, pub["author"])
}

func TestIssueViewsCountOnlyThatIssuesPage(t *testing.T) {
	db, h, n, e := issueFixture(t)
	now := time.Now().UTC()
	view := func(path string, at time.Time, session string) {
		require.NoError(t, db.Create(&models.SiteEvent{NewsletterID: n.ID, Kind: "view", Path: path, Session: session, CreatedAt: at}).Error)
	}
	issue := "/" + e.ID
	view(issue, now, "a1")
	view(issue, now, "a2")
	view(issue, now.AddDate(0, 0, -3), "b1")
	for i := range 4 {
		view(issue, now.AddDate(0, 0, -10), "c"+string(rune('0'+i)))
	}
	view("/", now, "a1")                                                                                                                  // the home page
	view("/"+uuid.NewString(), now, "a1")                                                                                                 // another issue
	require.NoError(t, db.Create(&models.SiteEvent{NewsletterID: n.ID, Kind: "click", Path: issue, Session: "a1", CreatedAt: now}).Error) // not a view

	code, v := call(t, h.IssueWebViews, "GET", "", "id", e.ID, "teamID", e.TeamID)
	require.Equal(t, 200, code)
	require.EqualValues(t, 7, v["total"])
	require.EqualValues(t, 2, v["today"])
	require.EqualValues(t, 3, v["week"])
	require.Equal(t, now.AddDate(0, 0, -10).Format(time.DateOnly), v["best"].(map[string]any)["key"])
	require.EqualValues(t, 4, v["best"].(map[string]any)["views"])
	daily := v["daily"].([]any)
	require.Len(t, daily, 11, "every day from the first view to today")
	require.EqualValues(t, 0, daily[1].(map[string]any)["views"], "quiet days count as zero")
	require.EqualValues(t, 2, daily[10].(map[string]any)["views"])

	code, _ = call(t, h.IssueWebViews, "GET", "", "id", e.ID, "teamID", uuid.NewString())
	require.Equal(t, 404, code)
}

func TestCountIssueViewsCapsTheDailyList(t *testing.T) {
	today := time.Date(2026, 11, 14, 15, 0, 0, 0, time.UTC)
	v := countIssueViews([]siteCount{{Key: "2025-01-01", Views: 9}, {Key: "2026-11-14", Views: 1}}, today)
	require.EqualValues(t, 10, v.Total)
	require.EqualValues(t, 1, v.Today)
	require.EqualValues(t, 1, v.Week)
	require.Equal(t, "2025-01-01", v.Best.Key)
	require.Len(t, v.Daily, issueViewDays)
	require.Equal(t, "2026-11-14", v.Daily[len(v.Daily)-1].Key)
	require.Empty(t, countIssueViews(nil, today).Daily)
}
