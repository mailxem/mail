// Command setup-email-plans creates Xem's per-email billing in Dodo Payments:
// one "Emails sent" meter and a monthly and a yearly usage-based product per
// paid plan, all under the "Xem" brand (set -brand to use another).
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
	"strings"
)

const meterEvent = "email.sent"

type plan struct {
	ID, Name       string
	MonthlyCents   int
	IncludedEmails int64 // per month
	OverageCents   string
	Yearly         bool
}

// Prices are in US cents; overage is cents per email (Dodo allows fractions).
var monthly = []plan{
	{"xem-starter", "Xem Starter", 900, 20000, "0.04", false},
	{"xem-growth", "Xem Growth", 2900, 100000, "0.03", false},
	{"xem-scale", "Xem Scale", 9900, 500000, "0.02", false},
}

// Yearly plans cost ten months and include twelve months of emails as one
// pool for the year: Dodo meters usage over the billing period.
var plans = func() []plan {
	all := append([]plan{}, monthly...)
	for _, p := range monthly {
		all = append(all, plan{p.ID + "-yearly", p.Name + " (Yearly)", p.MonthlyCents, p.IncludedEmails, p.OverageCents, true})
	}
	return all
}()

func (p plan) cents() int {
	if p.Yearly {
		return p.MonthlyCents * 10
	}
	return p.MonthlyCents
}

func (p plan) included() int64 {
	if p.Yearly {
		return p.IncludedEmails * 12
	}
	return p.IncludedEmails
}

func (p plan) description() string {
	period := "a month"
	if p.Yearly {
		period = "a year"
	}
	return fmt.Sprintf("%d emails %s included, then $%.2f per 1,000.", p.included(), period, mustFloat(p.OverageCents)*10)
}

// FreeEmails is the free plan's hard monthly cap; it has no Dodo product.
const FreeEmails = 3000

func main() {
	printSQL := flag.Bool("sql", false, "print the plan seed SQL to stdout; progress goes to stderr")
	brandName := flag.String("brand", "Xem", "Dodo brand the products are sold under")
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

	var brands struct {
		Items []struct {
			BrandID string `json:"brand_id"`
			Name    string `json:"name"`
		} `json:"items"`
	}
	call("GET", "/brands", nil, &brands)
	brandID := ""
	for _, b := range brands.Items {
		if b.Name == *brandName {
			brandID = b.BrandID
		}
	}
	if brandID == "" {
		log.Fatalf("no Dodo brand named %q: create it in the Dodo dashboard first", *brandName)
	}
	fmt.Fprintf(out, "brand %s (%s)\n", brandID, *brandName)

	var products struct {
		Items []struct {
			ProductID string `json:"product_id"`
			Name      string
		} `json:"items"`
	}
	call("GET", "/products?page_size=100", nil, &products)
	ids := map[string]string{}
	for _, p := range plans {
		interval := "Month"
		if p.Yearly {
			interval = "Year"
		}
		// Reuse a product only when its price is exactly this plan's: an older
		// product can share the name with a different price.
		id := ""
		for _, existing := range products.Items {
			if existing.Name != p.Name {
				continue
			}
			var full struct {
				Price struct {
					Type       string `json:"type"`
					FixedPrice int    `json:"fixed_price"`
					Interval   string `json:"payment_frequency_interval"`
					Meters     []struct {
						MeterID       string `json:"meter_id"`
						FreeThreshold int64  `json:"free_threshold"`
					} `json:"meters"`
				} `json:"price"`
			}
			call("GET", "/products/"+existing.ProductID, nil, &full)
			pr := full.Price
			if pr.Type == "usage_based_price" && pr.FixedPrice == p.cents() && pr.Interval == interval && len(pr.Meters) == 1 && pr.Meters[0].MeterID == meterID && pr.Meters[0].FreeThreshold == p.included() {
				id = existing.ProductID
			}
		}
		if id == "" {
			var created struct {
				ProductID string `json:"product_id"`
			}
			call("POST", "/products", map[string]any{
				"name":         p.Name,
				"description":  p.description(),
				"tax_category": "saas",
				"brand_id":     brandID,
				"price": map[string]any{
					"type": "usage_based_price", "currency": "USD", "fixed_price": p.cents(),
					"payment_frequency_count": 1, "payment_frequency_interval": interval,
					"subscription_period_count": 20, "subscription_period_interval": "Year",
					"meters": []map[string]any{{"meter_id": meterID, "free_threshold": p.included(), "price_per_unit": p.OverageCents}},
				},
			}, &created)
			id = created.ProductID
		} else {
			// Products made before the brand existed move under it.
			call("PATCH", "/products/"+id, map[string]any{"brand_id": brandID}, nil)
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
	row := func(id, dodo, name, desc string, cents, order int, popular bool, quota int64, cycle string) {
		sql += fmt.Sprintf(`INSERT INTO plans (id, dodo_plan_id, name, description, status, price, currency, price_per_seat, billing_cycle, trial_period_days, trial_enabled, display_order, is_popular, is_enterprise, metadata, created_at, updated_at)
VALUES ('%[1]s', '%[2]s', '%[3]s', '%[4]s', 'active', %[5]d/100.0, 'USD', %[5]d/100.0, '%[9]s', 0, false, %[6]d, %[7]t, false, '{}', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET dodo_plan_id = EXCLUDED.dodo_plan_id, name = EXCLUDED.name, description = EXCLUDED.description, status = 'active', price = EXCLUDED.price, price_per_seat = EXCLUDED.price_per_seat, billing_cycle = EXCLUDED.billing_cycle, display_order = EXCLUDED.display_order, is_popular = EXCLUDED.is_popular, updated_at = NOW();
INSERT INTO plan_features (id, plan_id, quota, feature, created_at, updated_at, trail)
VALUES ('%[1]s-emails', '%[1]s', %[8]d, 'emails_per_month', NOW(), NOW(), false)
ON CONFLICT (id) DO UPDATE SET quota = EXCLUDED.quota, updated_at = NOW();
`, id, dodo, name, desc, cents, order, popular, quota, cycle)
	}
	row("xem-free", "xem-free", "Free", fmt.Sprintf("%d emails a month. No card needed.", FreeEmails), 0, 1, false, FreeEmails, "MONTHLY")
	active := "'xem-free'"
	for i, p := range plans {
		cycle := "MONTHLY"
		if p.Yearly {
			cycle = "YEARLY"
		}
		// -1 = no hard cap: usage past the included emails is billed by Dodo.
		row(p.ID, ids[p.ID], p.Name, p.description(), p.cents(), i+2, strings.HasPrefix(p.ID, "xem-growth"), -1, cycle)
		active += ", '" + p.ID + "'"
	}
	// Teams on the retired free plan move to Free, so its 3,000-email cap applies.
	// updated_at stays: lookups pick a team's latest row, and a paid one must win.
	sql += "UPDATE subscriptions SET plan_id = 'xem-free' WHERE plan_id = 'solo-plan-001';\n"
	sql += "UPDATE plans SET status = 'inactive', updated_at = NOW() WHERE id NOT IN (" + active + ") AND status = 'active';\nCOMMIT;\n"
	return sql
}

func mustFloat(s string) float64 {
	var f float64
	if _, err := fmt.Sscan(s, &f); err != nil {
		log.Fatal(err)
	}
	return f
}
