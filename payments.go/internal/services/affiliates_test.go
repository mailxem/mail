package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mailxem/payments.go/internal/models"
)

func TestAffiliateReporterSendsPaymentWithoutTaxInSettlementCurrency(t *testing.T) {
	var got map[string]any
	var key, path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key, path = r.Header.Get("X-API-Key"), r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	a := &AffiliateReporter{URL: srv.URL + "/api/v1", Key: "secret", HTTP: srv.Client()}

	tax, settlementTax := int64(160), int64(180)
	p := &models.PaymentWebhookData{PaymentID: "pay_1", TotalAmount: 1060, Tax: &tax, Currency: "EUR",
		SettlementAmount: 1080, SettlementTax: &settlementTax, SettlementCurrency: "USD", CreatedAt: "2026-11-06T10:00:00Z"}
	if err := a.Payment(context.Background(), "team-1", p); err != nil {
		t.Fatal(err)
	}
	if key != "secret" || path != "/api/v1/internal/affiliates/payments" {
		t.Fatalf("sent to %s with key %q", path, key)
	}
	if got["payment_id"] != "pay_1" || got["team_id"] != "team-1" || got["base"] != float64(900) || got["currency"] != "USD" || got["paid_at"] != "2026-11-06T10:00:00Z" {
		t.Fatalf("payload = %v", got)
	}

	// Without settlement figures, the charged amount minus tax is used.
	if base, cur := commissionBase(&models.PaymentWebhookData{TotalAmount: 1060, Tax: &tax, Currency: "USD"}); base != 900 || cur != "USD" {
		t.Fatalf("base = %d %s", base, cur)
	}
}

func TestAffiliateReporterErrorsSoTheWebhookIsRetried(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusBadGateway) }))
	defer srv.Close()
	a := &AffiliateReporter{URL: srv.URL, Key: "k", HTTP: srv.Client()}
	if err := a.Reversal(context.Background(), "pay_1"); err == nil {
		t.Fatal("expected an error")
	}
}
