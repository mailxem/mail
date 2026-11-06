package middleware

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
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

func TestFormSubmitRateLimiterCapsVisitorsAndForms(t *testing.T) {
	addr := os.Getenv("REDIS_TEST_ADDR")
	if addr == "" {
		t.Skip("set REDIS_TEST_ADDR to run against Redis")
	}
	client := redis.NewClient(&redis.Options{Addr: addr, DB: 15})
	defer client.Close()
	// The limiter lets requests through when Redis is down, so it can only be checked against a live one.
	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Skipf("Redis at %s is unreachable: %v", addr, err)
	}
	e := echo.New()
	e.POST("/public/forms/:slug", func(c echo.Context) error { return c.NoContent(200) }, FormSubmitRateLimiter(client))
	post := func(slug, ip string) int {
		req := httptest.NewRequest("POST", "/public/forms/"+slug, nil)
		req.Header.Set("X-Forwarded-For", ip)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec.Code
	}
	perHour := int(float64(formSubmitPerIP.Limit) * formSubmitPerIP.Window.Seconds())
	perMinute := int(float64(formSubmitPerForm.Limit) * formSubmitPerForm.Window.Seconds())

	// One visitor is capped across every form.
	ip := uuid.NewString()
	for i := 0; i < perHour; i++ {
		if code := post(uuid.NewString(), ip); code != 200 {
			t.Fatalf("submission %d from one visitor: got %d", i+1, code)
		}
	}
	if code := post(uuid.NewString(), ip); code != 429 {
		t.Fatalf("visitor over the limit: got %d, want 429", code)
	}

	// One form is capped across every visitor.
	slug := uuid.NewString()
	for i := 0; i < perMinute; i++ {
		if code := post(slug, uuid.NewString()); code != 200 {
			t.Fatalf("submission %d to one form: got %d", i+1, code)
		}
	}
	if code := post(slug, uuid.NewString()); code != 429 {
		t.Fatalf("form over the limit: got %d, want 429", code)
	}
}
