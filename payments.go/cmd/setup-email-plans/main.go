// Command setup-email-plans creates Xem's per-email billing in Dodo Payments:
// one "Emails sent" meter and a monthly usage-based product per paid plan.
// It is idempotent: anything that already exists (by name) is reused. With
// -sql it then prints the plan seed for this mode's product IDs.
//
//	DODO_PAYMENTS_API_KEY=... DODO_TEST_MODE=true go run ./cmd/setup-email-plans -sql > plans.sql
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const meterEvent = "email.sent"

// plan prices are in US cents; overage is cents per email (Dodo allows fractions).
var plans = []struct {
	ID, Name       string
	MonthlyCents   int
	IncludedEmails int64
	OverageCents   string
}{
	{"xem-starter", "Xem Starter", 900, 20000, "0.04"},
	{"xem-growth", "Xem Growth", 2900, 100000, "0.03"},
	{"xem-scale", "Xem Scale", 9900, 500000, "0.02"},
}

// FreeEmails is the free plan's hard monthly cap; it has no Dodo product.
const FreeEmails = 3000

func main() {
	printSQL := flag.Bool("sql", false, "print the plan seed SQL to stdout; progress goes to stderr")
	flag.Parse()
	out := os.Stdout
	if *printSQL {
		out = os.Stderr
	}
	key := os.Getenv("DODO_PAYMENTS_API_KEY")
	if key == "" {
		key = os.Getenv("DODO_API_KEY")
	}
	if key == "" {
		log.Fatal("DODO_PAYMENTS_API_KEY is required")
	}
	base := "https://test.dodopayments.com"
	if os.Getenv("DODO_TEST_MODE") == "false" {
		base = "https://live.dodopayments.com"
	}
	call := func(method, path string, body, out any) {
		var r io.Reader
		if body != nil {
			raw, _ := json.Marshal(body)
			r = bytes.NewReader(raw)
		}
		req, _ := http.NewRequest(method, base+path, r)
		req.Header.Set("Authorization", "Bearer "+key)
		req.Header.Set("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()
		raw, _ := io.ReadAll(res.Body)
		if res.StatusCode >= 300 {
			log.Fatalf("%s %s: %d %s", method, path, res.StatusCode, raw)
		}
		if out != nil {
			if err := json.Unmarshal(raw, out); err != nil {
				log.Fatalf("%s %s: %v", method, path, err)
			}
		}
	}

	var meters struct {
		Items []struct {
			ID        string `json:"id"`
			EventName string `json:"event_name"`
		} `json:"items"`
	}
	call("GET", "/meters?page_size=100", nil, &meters)
	meterID := ""
	for _, m := range meters.Items {
		if m.EventName == meterEvent {
			meterID = m.ID
		}
	}
	if meterID == "" {
		var created struct{ ID string }
		call("POST", "/meters", map[string]any{
			"name": "Emails sent", "event_name": meterEvent, "measurement_unit": "emails",
			"description": "Emails sent from Xem; each event carries the count in metadata.emails",
			"aggregation": map[string]any{"type": "sum", "key": "emails"},
		}, &created)
		meterID = created.ID
	}
	fmt.Fprintf(out, "meter %s (%s)\n", meterID, meterEvent)

	var products struct {
		Items []struct {
			ProductID string `json:"product_id"`
			Name      string
		} `json:"items"`
	}
	call("GET", "/products?page_size=100", nil, &products)
	ids := map[string]string{}
	for _, p := range plans {
		id := ""
		for _, existing := range products.Items {
			if existing.Name == p.Name {
				id = existing.ProductID
			}
		}
		if id == "" {
			var created struct {
				ProductID string `json:"product_id"`
			}
			call("POST", "/products", map[string]any{
				"name":         p.Name,
				"description":  fmt.Sprintf("%d emails a month included, then $%.2f per 1,000.", p.IncludedEmails, mustFloat(p.OverageCents)*10),
				"tax_category": "saas",
				"price": map[string]any{
					"type": "usage_based_price", "currency": "USD", "fixed_price": p.MonthlyCents,
					"payment_frequency_count": 1, "payment_frequency_interval": "Month",
					"subscription_period_count": 20, "subscription_period_interval": "Year",
					"meters": []map[string]any{{"meter_id": meterID, "free_threshold": p.IncludedEmails, "price_per_unit": p.OverageCents}},
				},
			}, &created)
			id = created.ProductID
		}
		fmt.Fprintf(out, "%s %s\n", p.Name, id)
		ids[p.ID] = id
	}
	if *printSQL {
		fmt.Print(seedSQL(ids))
	}
}

// seedSQL upserts the per-email plans and stops older plans taking new
// subscriptions; existing subscribers keep their plan.
func seedSQL(ids map[string]string) string {
	sql := "BEGIN;\n"
	row := func(id, dodo, name, desc string, cents, order int, popular bool, quota int64) {
		sql += fmt.Sprintf(`INSERT INTO plans (id, dodo_plan_id, name, description, status, price, currency, price_per_seat, billing_cycle, trial_period_days, trial_enabled, display_order, is_popular, is_enterprise, metadata, created_at, updated_at)
VALUES ('%[1]s', '%[2]s', '%[3]s', '%[4]s', 'active', %[5]d/100.0, 'USD', %[5]d/100.0, 'MONTHLY', 0, false, %[6]d, %[7]t, false, '{}', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET dodo_plan_id = EXCLUDED.dodo_plan_id, name = EXCLUDED.name, description = EXCLUDED.description, status = 'active', price = EXCLUDED.price, price_per_seat = EXCLUDED.price_per_seat, display_order = EXCLUDED.display_order, is_popular = EXCLUDED.is_popular, updated_at = NOW();
INSERT INTO plan_features (id, plan_id, quota, feature, created_at, updated_at, trail)
VALUES ('%[1]s-emails', '%[1]s', %[8]d, 'emails_per_month', NOW(), NOW(), false)
ON CONFLICT (id) DO UPDATE SET quota = EXCLUDED.quota, updated_at = NOW();
`, id, dodo, name, desc, cents, order, popular, quota)
	}
	row("xem-free", "xem-free", "Free", fmt.Sprintf("%d emails a month. No card needed.", FreeEmails), 0, 1, false, FreeEmails)
	for i, p := range plans {
		// -1 = no hard cap: usage past the included emails is billed by Dodo.
		row(p.ID, ids[p.ID], p.Name, fmt.Sprintf("%d emails a month included, then $%.2f per 1,000.", p.IncludedEmails, mustFloat(p.OverageCents)*10), p.MonthlyCents, i+2, p.ID == "xem-growth", -1)
	}
	sql += "UPDATE plans SET status = 'inactive', updated_at = NOW() WHERE id NOT IN ('xem-free', 'xem-starter', 'xem-growth', 'xem-scale') AND status = 'active';\nCOMMIT;\n"
	return sql
}

func mustFloat(s string) float64 {
	var f float64
	if _, err := fmt.Sscan(s, &f); err != nil {
		log.Fatal(err)
	}
	return f
}
