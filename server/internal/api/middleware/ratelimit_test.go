package middleware

import (
	"testing"
	"time"
)

func TestDefaultLimitsMatchRealRoutes(t *testing.T) {
	config := RateLimitConfig{DefaultLimit: 100.0 / 60.0, DefaultBurst: 50, EndpointLimits: defaultEndpointLimits}
	perWindow := func(l EndpointLimit) int { return int(float64(l.Limit) * l.Window.Seconds()) }
	for _, tc := range []struct {
		request, counter string
		max              int
		window           time.Duration
	}{
		{"POST:/api/v1/auth/login", "POST:/api/v1/auth/login", 5, time.Minute},
		{"POST:/api/v1/auth/register", "POST:/api/v1/auth/register", 3, time.Hour},
		{"POST:/api/v1/auth/password-reset", "POST:/api/v1/auth/password-reset", 3, time.Hour},
		{"POST:/api/v1/auth/password-reset/verify", "POST:/api/v1/auth/password-reset/verify", 10, time.Hour},
		{"POST:/api/v1/auth/refresh", "POST:/api/v1/auth/refresh", 10, time.Minute},
		{"POST:/api/v1/emails", "POST:/api/v1/emails", 100, time.Minute},
		{"GET:/api/v1/analytics/campaign", "GET:/api/v1/analytics/", 100, time.Minute},
		{"GET:/t/click/abc123", "GET:/t/click/", 1000, time.Minute},
		{"GET:/t/click/def456", "GET:/t/click/", 1000, time.Minute},
		// Other routes keep their own counter at the default limit.
		{"GET:/api/v1/campaigns", "GET:/api/v1/campaigns", 100, time.Minute},
		// A GET to an auth path is not the login limit.
		{"GET:/api/v1/auth/login", "GET:/api/v1/auth/login", 100, time.Minute},
	} {
		counter, limit := getLimitConfig(tc.request, config)
		if counter != tc.counter || perWindow(limit) != tc.max || limit.Window != tc.window {
			t.Errorf("%s: got counter %q, %d per %s; want %q, %d per %s",
				tc.request, counter, perWindow(limit), limit.Window, tc.counter, tc.max, tc.window)
		}
	}
}
