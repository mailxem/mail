package handlers

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"kori/internal/config"
	"kori/internal/models"
	"kori/internal/utils"
)

func TestClickRedirectsOnlyToTheEmailsOwnLinks(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	raw, err := db.DB()
	require.NoError(t, err)
	raw.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&models.Campaign{}, &models.Email{}, &models.EmailTracking{}))
	secret := "click-signing-test-secret"
	config.GetConfig().JWT.Secret = secret

	// An email sent before signing: its stored body holds the old, unsigned link.
	id := uuid.NewString()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"mailId": id}).SignedString([]byte(secret))
	require.NoError(t, err)
	enc := func(u string) string { return base64.StdEncoding.EncodeToString([]byte(u)) }
	body := `<a href="https://api.example.com/t/click/` + enc("https://example.com/pricing") + `?token=` + token + `">Pricing</a>`
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&models.Email{Base: models.Base{ID: id}, TeamID: uuid.NewString(), From: "news@example.com", To: "ada@example.net", Subject: "Hi", Body: base64.StdEncoding.EncodeToString([]byte(body)), Status: models.EmailStatusSent, SMTPConfigID: uuid.NewString(), CategoryID: uuid.NewString()}).Error)

	h := NewTrackingHandler(db)
	click := func(target, query string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		require.NoError(t, h.HandleEmailClick(echo.New().NewContext(httptest.NewRequest("GET", "/t/click/"+enc(target)+"?token="+token+query, nil), rec)))
		return rec
	}
	sign := func(target string) string { return "&sig=" + utils.ClickSignature(secret, id, target) }

	rec := click("https://example.com/new-link", sign("https://example.com/new-link"))
	require.Equal(t, http.StatusFound, rec.Code, "a signed link redirects")
	require.Equal(t, "https://example.com/new-link", rec.Header().Get("Location"))

	require.Equal(t, http.StatusNotFound, click("https://evil.example/login", sign("https://example.com/new-link")).Code, "a valid token can't carry another destination")
	require.Equal(t, http.StatusNotFound, click("https://evil.example/login", "&sig=forged").Code)

	rec = click("https://example.com/pricing", "")
	require.Equal(t, http.StatusFound, rec.Code, "an old unsigned link still works for the email's own destination")
	require.Equal(t, "https://example.com/pricing", rec.Header().Get("Location"))
	require.Equal(t, http.StatusNotFound, click("https://evil.example/login", "").Code, "an unsigned link to anywhere else is refused")

	for _, bad := range []string{"javascript:alert(1)", "data:text/html,hi", "//evil.example"} {
		require.Equal(t, http.StatusNotFound, click(bad, sign(bad)).Code, bad)
	}
}

func TestTrackedLinksCarryTheirSignature(t *testing.T) {
	html := utils.TrackEmailHTML(`<a href="https://example.com/a">A</a>`, "mail-1", "https://api.example.com", "secret-for-links-0123456789abcdef")
	require.Contains(t, html, "&amp;sig="+utils.ClickSignature("secret-for-links-0123456789abcdef", "mail-1", "https://example.com/a"))
}
