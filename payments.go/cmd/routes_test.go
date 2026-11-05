package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/mailxem/payments.go/internal/config"
	"github.com/mailxem/payments.go/internal/handlers"
)

func TestWebhookAdminRoutesNeedTheAdminSecret(t *testing.T) {
	// A fresh router per call keeps the strict rate limiter from answering first.
	call := func(method, path, secret string) int {
		e := echo.New()
		setupRoutes(e, &handlers.SubscriptionHandlers{}, &handlers.PlanHandler{}, &handlers.WebhookHandler{}, &config.Config{AdminSecret: "an-admin-secret-that-is-long-enough"})
		req := httptest.NewRequest(method, path, nil)
		if secret != "" {
			req.Header.Set("X-Admin-Secret", secret)
		}
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec.Code
	}
	for _, r := range [][2]string{
		{http.MethodGet, "/api/v1/webhooks/events"},
		{http.MethodGet, "/api/v1/webhooks/events/evt_1/logs"},
		{http.MethodPost, "/api/v1/webhooks/retry-failed"},
		{http.MethodPost, "/api/v1/webhooks/cleanup"},
	} {
		for _, secret := range []string{"", "wrong"} {
			if got := call(r[0], r[1], secret); got != http.StatusUnauthorized {
				t.Errorf("%s %s with secret %q = %d, want 401", r[0], r[1], secret, got)
			}
		}
	}
	// Dodo's delivery endpoint stays public; it checks the webhook signature itself.
	if got := call(http.MethodPost, "/api/v1/webhooks/dodo", ""); got != http.StatusBadRequest {
		t.Errorf("POST /api/v1/webhooks/dodo = %d, want 400 for a missing signature", got)
	}
}
