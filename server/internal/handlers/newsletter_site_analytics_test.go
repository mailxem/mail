package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"kori/internal/models"
)

func TestSiteEventsFeedTheAnalyticsReport(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Newsletter{}, &models.SiteEvent{}, &models.NewsletterSignup{}))
	team, slug := uuid.NewString(), "founder-notes"
	n := models.Newsletter{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "Founder Notes", ListID: uuid.NewString(), PublicSlug: &slug}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&n).Error)
	h := NewMarketingHandler(db, "secret", "https://xem.email")

	send := func(body string) int {
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(httptest.NewRequest("POST", "/", strings.NewReader(body)), rec)
		c.SetParamNames("slug")
		c.SetParamValues(slug)
		if err := h.RecordSiteEvent(c); err != nil {
			return err.(*echo.HTTPError).Code
		}
		return rec.Code
	}
	require.Equal(t, 204, send(`{"t":"view","p":"/founder-notes?ref=x","r":"https://www.google.com/search","s":"abcdefgh1","w":1440,"h":2400}`))
	require.Equal(t, 204, send(`{"t":"view","p":"/founder-notes","s":"abcdefgh1","w":1440,"h":2400}`))
	require.Equal(t, 204, send(`{"t":"view","p":"/founder-notes/issue-1","s":"zyxwvuts2","w":390,"h":5000}`))
	require.Equal(t, 204, send(`{"t":"click","p":"/founder-notes","s":"abcdefgh1","w":1440,"x":0.5,"y":310,"h":2400,"l":"https://example.com/post"}`))
	require.Equal(t, 204, send(`{"t":"click","p":"/founder-notes","s":"abcdefgh1","w":1440,"x":0.51,"y":315,"h":2400,"l":"javascript:alert(1)"}`))
	require.Equal(t, 400, send(`{"t":"scroll","p":"/founder-notes","s":"abcdefgh1","w":1440}`))
	require.Equal(t, 400, send(`{"t":"view","p":"/founder-notes","s":"bad id!","w":1440}`))

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(httptest.NewRequest("GET", "/?days=7", nil), rec)
	c.SetParamNames("id")
	c.SetParamValues(n.ID)
	c.Set("teamID", team)
	require.NoError(t, h.SiteAnalytics(c))
	var r struct {
		Views, Visitors, Clicks int64
		Pages                   []siteCount
		Referrers               []siteCount
		Daily                   []siteCount
		Links                   []struct{ Link string }
		Heatmap                 struct {
			Path   string
			Clicks int
			Cells  []struct{ X, Y, N int }
		}
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &r))
	require.EqualValues(t, 3, r.Views)
	require.EqualValues(t, 2, r.Visitors)
	require.EqualValues(t, 2, r.Clicks)
	require.Equal(t, "/founder-notes", r.Pages[0].Key, "query strings are dropped")
	require.EqualValues(t, 2, r.Pages[0].Views)
	require.Equal(t, "google.com", r.Referrers[0].Key)
	require.Len(t, r.Daily, 1)
	require.Len(t, r.Links, 1, "only http(s) and mailto links are kept")
	require.Equal(t, "/founder-notes", r.Heatmap.Path)
	require.Equal(t, 2, r.Heatmap.Clicks)
	require.Equal(t, []struct{ X, Y, N int }{{20, 15, 2}}, r.Heatmap.Cells)

	// Another workspace can't read it.
	c = echo.New().NewContext(httptest.NewRequest("GET", "/", nil), httptest.NewRecorder())
	c.SetParamNames("id")
	c.SetParamValues(n.ID)
	c.Set("teamID", uuid.NewString())
	require.Error(t, h.SiteAnalytics(c))
}
