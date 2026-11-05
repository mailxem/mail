package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mailxem/payments.go/internal/models"
)

// AffiliateReporter tells the Xem API about plan payments, refunds and
// disputes, so it can credit or reverse affiliate commissions. It uses the
// same shared key the API uses to record usage here.
type AffiliateReporter struct {
	URL  string // Xem API base, e.g. https://api.internal/api/v1
	Key  string
	HTTP *http.Client
}

// Affiliates is nil unless XEM_API_INTERNAL_URL and PAYMENTS_INTERNAL_KEY are set.
var Affiliates = AffiliatesFromEnv()

func AffiliatesFromEnv() *AffiliateReporter {
	u, key := strings.TrimRight(os.Getenv("XEM_API_INTERNAL_URL"), "/"), os.Getenv("PAYMENTS_INTERNAL_KEY")
	if u == "" || key == "" {
		return nil
	}
	return &AffiliateReporter{URL: u, Key: key, HTTP: &http.Client{Timeout: 10 * time.Second}}
}

// commissionBase is what the customer paid for the plan after discounts and
// without tax, in the settlement currency when Dodo provides one.
func commissionBase(p *models.PaymentWebhookData) (int64, string) {
	if p.SettlementCurrency != "" && p.SettlementAmount > 0 {
		base := p.SettlementAmount
		if p.SettlementTax != nil {
			base -= *p.SettlementTax
		}
		return base, p.SettlementCurrency
	}
	base := p.TotalAmount
	if p.Tax != nil {
		base -= *p.Tax
	}
	return base, p.Currency
}

func (a *AffiliateReporter) post(ctx context.Context, path string, body any) error {
	b, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.URL+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", a.Key)
	res, err := a.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return fmt.Errorf("affiliate %s returned %d", path, res.StatusCode)
	}
	return nil
}

// Payment reports a successful payment by a workspace.
func (a *AffiliateReporter) Payment(ctx context.Context, teamID string, p *models.PaymentWebhookData) error {
	base, currency := commissionBase(p)
	if base <= 0 {
		return nil
	}
	paidAt, err := time.Parse(time.RFC3339, p.CreatedAt)
	if err != nil {
		paidAt = time.Now()
	}
	return a.post(ctx, "/internal/affiliates/payments", map[string]any{
		"payment_id": p.PaymentID, "team_id": teamID, "base": base, "currency": currency, "paid_at": paidAt,
	})
}

// Reversal reports a refunded or disputed payment.
func (a *AffiliateReporter) Reversal(ctx context.Context, paymentID string) error {
	return a.post(ctx, "/internal/affiliates/reversals", map[string]string{"payment_id": paymentID})
}
