package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"kori/internal/models"
	"kori/internal/sitedomains"
)

func TestCustomDomainGoesLiveOnlyWhenCloudflareSaysSo(t *testing.T) {
	status := "pending"
	cf := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer cf-token", r.Header.Get("Authorization"))
		require.True(t, strings.HasPrefix(r.URL.Path, "/zones/zone-1/custom_hostnames"))
		fmt.Fprintf(w, `{"success":true,"result":{"id":"ch-1","status":%q,"ownership_verification":{"type":"txt","name":"_cf-custom-hostname.news.example.com","value":"abc"},"ssl":{"status":%q}}}`, status, status)
	}))
	defer cf.Close()
	old := siteDomains
	siteDomains = func() *sitedomains.Client {
		return &sitedomains.Client{Token: "cf-token", Zone: "zone-1", BaseURL: cf.URL, HTTP: cf.Client()}
	}
	t.Cleanup(func() { siteDomains = old })

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Newsletter{}))
	team, slug := uuid.NewString(), "founder-notes"
	n := models.Newsletter{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "Founder Notes", PublicSlug: &slug}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&n).Error)
	h := NewMarketingHandler(db, "secret", "https://api.xem.email")

	call := func(fn echo.HandlerFunc, method, body, param string) (int, map[string]any) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(method, "/", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		c := echo.New().NewContext(req, rec)
		c.Set("teamID", team)
		c.SetParamNames("id", "host")
		c.SetParamValues(n.ID, param)
		if err := fn(c); err != nil {
			return err.(*echo.HTTPError).Code, nil
		}
		var out map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &out)
		return rec.Code, out
	}

	code, _ := call(h.SetNewsletterDomain, "PUT", `{"domain":"blog.xem.email"}`, "")
	require.Equal(t, 400, code)
	code, out := call(h.SetNewsletterDomain, "PUT", `{"domain":"HTTPS://News.Example.com."}`, "")
	require.Equal(t, 200, code)
	require.Equal(t, "news.example.com", out["domain"])
	require.Equal(t, "pending", out["status"])
	require.Len(t, out["records"], 2, "the CNAME and Cloudflare's ownership TXT")

	// Not served until Cloudflare reports it active.
	code, _ = call(h.PublicSiteByDomain, "GET", "", "news.example.com")
	require.Equal(t, 404, code)
	status = "active"
	code, out = call(h.NewsletterDomain, "GET", "", "")
	require.Equal(t, 200, code)
	require.Equal(t, "active", out["status"])
	code, out = call(h.PublicSiteByDomain, "GET", "", "news.example.com")
	require.Equal(t, 200, code)
	require.Equal(t, slug, out["slug"])
	require.NoError(t, db.First(&n, "id = ?", n.ID).Error)
	require.Equal(t, "https://news.example.com", SiteURL(n))

	code, _ = call(h.SetNewsletterDomain, "PUT", `{"domain":""}`, "")
	require.Equal(t, 200, code)
	code, _ = call(h.PublicSiteByDomain, "GET", "", "news.example.com")
	require.Equal(t, 404, code)
}
