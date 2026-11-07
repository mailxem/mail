package handlers

import (
	"context"
	"encoding/base64"
	"kori/internal/utils"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"kori/internal/config"
	"kori/internal/models"
)

const browser = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15"

func TestMachineReason(t *testing.T) {
	for _, tc := range []struct {
		method, agent string
		event         models.EmailTrackingEvent
		kind          string
	}{
		{"GET", browser, models.EmailTrackingEventClick, ""},
		{"GET", browser, models.EmailTrackingEventOpen, ""},
		{"HEAD", browser, models.EmailTrackingEventClick, models.TrackingMachineBot},
		{"GET", "", models.EmailTrackingEventOpen, models.TrackingMachineBot},
		{"GET", "Mozilla/5.0 (compatible; Barracuda Sentinel)", models.EmailTrackingEventClick, models.TrackingMachineBot},
		{"GET", "python-requests/2.31", models.EmailTrackingEventClick, models.TrackingMachineBot},
		{"GET", "Mozilla/5.0 (compatible; bingbot/2.0)", models.EmailTrackingEventOpen, models.TrackingMachineBot},
		{"GET", "Mozilla/5.0", models.EmailTrackingEventOpen, models.TrackingMachineProxy},
		{"GET", "Mozilla/5.0 (Windows NT 5.1; rv:11.0) Gecko Firefox/11.0 (via ggpht.com GoogleImageProxy)", models.EmailTrackingEventOpen, models.TrackingMachineProxy},
		// Proxies only fetch images; a bare agent on a click is left to the timing rules.
		{"GET", "Mozilla/5.0", models.EmailTrackingEventClick, ""},
	} {
		kind, reason := machineReason(tc.method, tc.agent, tc.event)
		require.Equal(t, tc.kind, kind, "%s %q %s", tc.method, tc.agent, tc.event)
		require.Equal(t, kind == "", reason == "")
	}
}

func TestTrackingFlagsMachines(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	raw, err := db.DB()
	require.NoError(t, err)
	raw.SetMaxOpenConns(1) // one in-memory database, shared with the background writer
	require.NoError(t, db.AutoMigrate(&models.Campaign{}, &models.Email{}, &models.EmailTracking{}))
	config.GetConfig().JWT.Secret = "tracking-test"
	t.Setenv("TRACKING_FLUSH_INTERVAL", "10ms") // hits are saved in batches; flush quickly
	h := NewTrackingHandler(db)
	defer h.Close(context.Background())
	email := func(sent time.Time) string {
		e := models.Email{Base: models.Base{ID: uuid.NewString()}, To: "ada@example.com", SentAt: sent, Status: models.EmailStatusSent}
		require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&e).Error)
		return e.ID
	}
	hit := func(method, path, emailID, agent string, extra ...string) *httptest.ResponseRecorder {
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"mailId": emailID}).SignedString([]byte("tracking-test"))
		require.NoError(t, err)
		req := httptest.NewRequest(method, path+"?token="+token+strings.Join(extra, ""), nil)
		req.Header.Set("User-Agent", agent)
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)
		if path == "/t/open" {
			require.NoError(t, h.HandleEmailOpen(c))
		} else {
			require.NoError(t, h.HandleEmailClick(c))
		}
		return rec
	}
	saved := func(emailID string) int64 {
		var n int64
		require.NoError(t, db.Model(&models.EmailTracking{}).Where("email_id = ?", emailID).Count(&n).Error)
		return n
	}
	click := func(method, emailID, link, agent string) {
		before := saved(emailID)
		path := "/t/click/" + base64.StdEncoding.EncodeToString([]byte(link))
		rec := hit(method, path, emailID, agent, "&sig="+utils.ClickSignature("tracking-test", emailID, link))
		require.Equal(t, http.StatusFound, rec.Code, "the redirect works for machines too")
		require.Equal(t, link, rec.Header().Get("Location"))
		require.Eventually(t, func() bool { return saved(emailID) == before+1 }, 2*time.Second, 5*time.Millisecond)
	}
	events := func(emailID string) []models.EmailTracking {
		var rows []models.EmailTracking
		require.NoError(t, db.Order("timestamp").Find(&rows, "email_id = ?", emailID).Error)
		return rows
	}
	hourAgo := time.Now().Add(-time.Hour)

	// A person clicks one link, then another a few seconds later.
	person := email(hourAgo)
	click("GET", person, "https://example.com/a", browser)
	require.NoError(t, db.Model(&models.EmailTracking{}).Where("email_id = ?", person).Update("timestamp", time.Now().Add(-5*time.Second)).Error)
	click("GET", person, "https://example.com/b", browser)
	for _, e := range events(person) {
		require.True(t, e.Human(), e.URL)
	}

	// A HEAD probe still redirects and is flagged.
	probed := email(hourAgo)
	click("HEAD", probed, "https://example.com/a", browser)
	require.Equal(t, "HEAD request", events(probed)[0].MachineReason)

	// A named scanner.
	scanned := email(hourAgo)
	click("GET", scanned, "https://example.com/a", "Mozilla/5.0 (compatible; Proofpoint URL Defense)")
	require.Equal(t, models.TrackingMachineBot, events(scanned)[0].Machine)

	// A click right after delivery.
	fresh := email(time.Now())
	click("GET", fresh, "https://example.com/a", browser)
	require.Equal(t, "clicked within seconds of delivery", events(fresh)[0].MachineReason)

	// A gateway posing as a browser opens every link at once; the first click is flagged after the fact.
	burst := email(hourAgo)
	click("GET", burst, "https://example.com/a", browser)
	click("GET", burst, "https://example.com/b", browser)
	click("GET", burst, "https://example.com/c", browser)
	// The flush flags the burst just after it saves the row, so wait for the flag too.
	require.Eventually(t, func() bool {
		for _, e := range events(burst) {
			if e.MachineReason != "several links clicked at once" {
				return false
			}
		}
		return true
	}, 2*time.Second, 5*time.Millisecond)

	// Opens: a person, Apple Mail Privacy Protection and a scanner.
	opened := email(hourAgo)
	hit("GET", "/t/open", opened, browser)
	hit("GET", "/t/open", opened, "Mozilla/5.0")
	hit("GET", "/t/open", opened, "Mozilla/5.0 (compatible; Mimecast)")
	require.Eventually(t, func() bool { return saved(opened) == 3 }, 2*time.Second, 5*time.Millisecond)
	rows := events(opened)
	require.Len(t, rows, 3)
	stats := processEmailAnalytics(rows, "UTC")
	require.Equal(t, 1, stats.OpenCount)
	require.Equal(t, 1, stats.ProxyOpens)
	require.Equal(t, 1, stats.BotOpens)

	stats = processEmailAnalytics(events(burst), "UTC")
	require.Zero(t, stats.ClickCount)
	require.Equal(t, 3, stats.BotClicks)
	require.Empty(t, stats.ClickedLinks)
}
