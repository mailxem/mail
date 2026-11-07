package billing

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestSelfHostedFeaturesFollowTheLicense(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Key string `json:"license_key"`
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		if in.Key == "unknown" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]bool{"valid": in.Key == "good"})
	}))
	defer server.Close()
	old := LicenseURL
	LicenseURL = server.URL
	t.Cleanup(func() { LicenseURL = old; licenseRequired.Store(false); licensed.Store(false) })

	paidOnly := func() error {
		c := echo.New().NewContext(httptest.NewRequest(http.MethodPost, "/", nil), httptest.NewRecorder())
		c.Set("teamID", "team")
		return PaidOnly(func(echo.Context) error { return nil })(c)
	}
	ctx := context.Background()

	// Before WatchLicense runs (tests, Xem Cloud), nothing is gated.
	require.NoError(t, paidOnly())
	require.False(t, IsFree(ctx, "team"))

	// A self-hosted install without a key gets Free-plan limits.
	WatchLicense(ctx, "", t.Logf)
	var he *echo.HTTPError
	require.ErrorAs(t, paidOnly(), &he)
	require.Equal(t, http.StatusPaymentRequired, he.Code)
	require.True(t, IsFree(ctx, "team"))
	// Newsletters, like every paid feature, name the license on self-hosted installs.
	newsletterErr := PaidFeature(ErrNewsletters)(func(echo.Context) error { return nil })(echo.New().NewContext(httptest.NewRequest(http.MethodPost, "/", nil), httptest.NewRecorder()))
	require.ErrorAs(t, newsletterErr, &he)
	require.Equal(t, ErrLicenseFeature, he.Message)

	for key, want := range map[string]bool{"good": true, "revoked": false, "unknown": false} {
		valid, err := validateLicense(ctx, key)
		require.NoError(t, err, key)
		require.Equal(t, want, valid, key)
	}

	watch, stop := context.WithCancel(ctx)
	defer stop()
	go WatchLicense(watch, "good", t.Logf)
	require.Eventually(t, Licensed, time.Second, 10*time.Millisecond)
	require.NoError(t, paidOnly())
	require.False(t, IsFree(ctx, "team"))

	// A failed check keeps a licensed install licensed.
	LicenseURL = "http://127.0.0.1:1"
	_, err := validateLicense(ctx, "good")
	require.Error(t, err)
	require.True(t, Licensed())
}
