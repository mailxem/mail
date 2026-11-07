package handlers

import (
	"context"
	"encoding/base64"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"kori/internal/marketing"
	"kori/internal/models"
)

func TestSiteSignupNeedsConfirmationFromTheInbox(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Newsletter{}, &models.NewsletterSignup{}, &models.Contact{}, &models.MailingList{}, &models.SuppressionList{}))
	team := uuid.NewString()
	list := models.MailingList{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "Readers"}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&list).Error)
	slug := "founder-notes"
	n := models.Newsletter{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "Founder Notes", ListID: list.ID, SMTPConfigID: uuid.NewString(), PublicSlug: &slug}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&n).Error)

	var sent []models.Email
	oldQueue, oldCheck := queueSiteEmail, checkSiteEmail
	queueSiteEmail = func(e *models.Email, _, _ string) error { sent = append(sent, *e); return nil }
	checkSiteEmail = func(context.Context, string) string { return "" }
	t.Cleanup(func() { queueSiteEmail, checkSiteEmail = oldQueue, oldCheck })

	h := NewMarketingHandler(db, "secret", "https://api.xem.email")
	call := func(fn echo.HandlerFunc, body string) (int, error) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		c := echo.New().NewContext(req, rec)
		c.SetParamNames("slug")
		c.SetParamValues(slug)
		err := fn(c)
		if he, ok := err.(*echo.HTTPError); ok {
			return he.Code, nil
		}
		return rec.Code, err
	}
	active := func(email string) int64 {
		var count int64
		require.NoError(t, db.Model(&models.Contact{}).Where("list_id = ? AND email = ? AND status = ?", list.ID, email, models.SubscriberStatusActive).Count(&count).Error)
		return count
	}

	code, err := call(h.SubscribeNewsletter, `{"email":" Reader@Example.com "}`)
	require.NoError(t, err)
	require.Equal(t, 202, code)
	require.Len(t, sent, 1)
	require.Equal(t, "reader@example.com", sent[0].To)
	require.Equal(t, n.SMTPConfigID, sent[0].SMTPConfigID)
	require.Zero(t, active("reader@example.com"), "nobody joins before confirming")

	// Asking again straight away doesn't mail them twice.
	code, _ = call(h.SubscribeNewsletter, `{"email":"reader@example.com"}`)
	require.Equal(t, 202, code)
	require.Len(t, sent, 1)

	body, err := base64.StdEncoding.DecodeString(sent[0].Body)
	require.NoError(t, err)
	link := regexp.MustCompile(`https://xem\.email/founder-notes/confirm\?token=([0-9a-f]+)`).FindStringSubmatch(string(body))
	require.NotNil(t, link, string(body))

	code, _ = call(h.ConfirmNewsletter, `{"token":"not-a-real-token"}`)
	require.Equal(t, 410, code)
	code, err = call(h.ConfirmNewsletter, `{"token":"`+link[1]+`"}`)
	require.NoError(t, err)
	require.Equal(t, 200, code)
	require.EqualValues(t, 1, active("reader@example.com"))
	// The link works twice without adding them twice.
	code, _ = call(h.ConfirmNewsletter, `{"token":"`+link[1]+`"}`)
	require.Equal(t, 200, code)
	require.EqualValues(t, 1, active("reader@example.com"))

	// Someone already reading gets the same answer and no email.
	code, _ = call(h.SubscribeNewsletter, `{"email":"reader@example.com"}`)
	require.Equal(t, 202, code)
	require.Len(t, sent, 1)

	// A suppressed address confirms without joining.
	require.NoError(t, db.Create(&models.SuppressionList{TeamID: team, EmailAddress: "blocked@example.com", IsActive: true}).Error)
	_, _ = call(h.SubscribeNewsletter, `{"email":"blocked@example.com"}`)
	require.Len(t, sent, 2)
	body, _ = base64.StdEncoding.DecodeString(sent[1].Body)
	link = regexp.MustCompile(`token=([0-9a-f]+)`).FindStringSubmatch(string(body))
	code, _ = call(h.ConfirmNewsletter, `{"token":"`+link[1]+`"}`)
	require.Equal(t, 200, code)
	require.Zero(t, active("blocked@example.com"))

	// Links stop working after a week.
	require.NoError(t, db.Model(&models.NewsletterSignup{}).Where("1 = 1").Updates(map[string]any{"created_at": time.Now().Add(-8 * 24 * time.Hour), "confirmed_at": nil}).Error)
	code, _ = call(h.ConfirmNewsletter, `{"token":"`+link[1]+`"}`)
	require.Equal(t, 410, code)
}

func TestEditionCoverSkipsTrackingPixels(t *testing.T) {
	html := `<img src="https://t.example.com/open.gif" width="1" height="1"><p>Hi</p><img alt="" src="https://cdn.example.com/hero.jpg" width="600">`
	require.Equal(t, "https://cdn.example.com/hero.jpg", marketing.EditionCover(html))
	require.Equal(t, "", marketing.EditionCover(`<img src="http://insecure.example.com/a.png">`))
}

func TestReservedSiteSlugs(t *testing.T) {
	_, err := marketing.NormalizePublicSlug("pricing")
	require.Error(t, err)
	slug, err := marketing.NormalizePublicSlug("Founder-Notes")
	require.NoError(t, err)
	require.Equal(t, "founder-notes", *slug)
}
