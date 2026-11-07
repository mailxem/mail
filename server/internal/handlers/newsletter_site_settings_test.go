package handlers

import (
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

func TestSiteSettingsLeaveTheScheduleAlone(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Newsletter{}))
	team, taken := uuid.NewString(), "taken-name"
	// A scheduled newsletter whose send time has just passed: a full save would fail validation.
	due := time.Now().Add(-time.Minute)
	n := models.Newsletter{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "Notes", Status: "SCHEDULED", NextSendAt: &due}
	other := models.Newsletter{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "Other", PublicSlug: &taken}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&n).Error)
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&other).Error)
	h := NewMarketingHandler(db, "secret", "https://api.xem.email")
	put := func(body string) int {
		req := httptest.NewRequest("PUT", "/", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)
		c.Set("teamID", team)
		c.SetParamNames("id")
		c.SetParamValues(n.ID)
		if err := h.SetNewsletterSite(c); err != nil {
			return err.(*echo.HTTPError).Code
		}
		return rec.Code
	}

	require.Equal(t, 400, put(`{"publicSlug":"weekly","siteTheme":"nope"}`))
	require.Equal(t, 400, put(`{"publicSlug":"taken-name","siteTheme":"zine"}`))
	require.Equal(t, 200, put(`{"publicSlug":"Weekly","siteTheme":"zine","siteLogoUrl":"https://example.com/l.png"}`))
	var got models.Newsletter
	require.NoError(t, db.First(&got, "id = ?", n.ID).Error)
	require.Equal(t, "weekly", *got.PublicSlug)
	require.Equal(t, "zine", got.SiteTheme)
	require.Equal(t, "SCHEDULED", got.Status)
	require.True(t, got.NextSendAt.Equal(due))

	require.Equal(t, 200, put(`{"publicSlug":"","siteTheme":"zine"}`))
	require.NoError(t, db.First(&got, "id = ?", n.ID).Error)
	require.Nil(t, got.PublicSlug)
}
