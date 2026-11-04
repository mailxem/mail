package dodo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mailxem/payments.go/internal/config"
)

func TestCheckoutLeavesTheAddressToDodoAndTagsOurSubscription(t *testing.T) {
	var path string
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"session_id":"cks_1","checkout_url":"https://test.checkout.dodopayments.com/cks_1"}`))
	}))
	defer server.Close()
	client := NewClient(&config.Config{DodoTestMode: true}).WithBaseURL(server.URL)

	session, err := client.CreateCheckoutSession(context.Background(), CheckoutRequest{
		TeamID: "team-1", PlanID: "pdt_starter", CustomerID: "cus_1", Seats: 1,
		SuccessURL: "https://app.xem.email/settings/billing?success=true", XemSubscriptionID: "sub-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if path != "/checkouts" || session.URL != "https://test.checkout.dodopayments.com/cks_1" || session.ID != "cks_1" {
		t.Fatalf("got path %q and session %+v", path, session)
	}
	if _, ok := body["billing_address"]; ok {
		t.Fatalf("no saved address should leave billing_address to Dodo, got %v", body["billing_address"])
	}
	metadata, _ := body["metadata"].(map[string]any)
	if metadata[XemSubscriptionIDKey] != "sub-1" || metadata["team_id"] != "team-1" {
		t.Fatalf("metadata = %v", metadata)
	}
	trial, _ := body["subscription_data"].(map[string]any)
	if trial["trial_period_days"] != float64(7) {
		t.Fatalf("subscription_data = %v", body["subscription_data"])
	}

	_, _ = client.CreateCheckoutSession(context.Background(), CheckoutRequest{
		PlanID: "pdt_starter", CustomerID: "cus_1", Seats: 1, XemSubscriptionID: "sub-2",
		BillingDetails: BillingDetails{Country: "IN", City: "Pune"},
	})
	if address, _ := body["billing_address"].(map[string]any); address["country"] != "IN" || address["city"] != "Pune" {
		t.Fatalf("a saved address should prefill checkout, got %v", body["billing_address"])
	}
}
