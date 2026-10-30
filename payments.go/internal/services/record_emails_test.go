package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mailxem/payments.go/internal/config"
	"github.com/mailxem/payments.go/internal/models"
	"github.com/mailxem/payments.go/internal/repositories"
	"github.com/mailxem/payments.go/internal/structs"
	"github.com/mailxem/payments.go/pkg/dodo"
)

type fakeUsage struct {
	repositories.UsageRepository
	total  int
	events map[string]bool
}

func (f *fakeUsage) EventSeen(_ context.Context, id string) (bool, error) { return f.events[id], nil }
func (f *fakeUsage) MarkEvent(_ context.Context, id string) error        { f.events[id] = true; return nil }
func (f *fakeUsage) UpdateLimit(_ context.Context, _ string, _ time.Time, n int, feature models.PlanFeature) error {
	if feature != models.PlanFeatureEmails {
		panic("wrong feature")
	}
	f.total += n
	return nil
}

func TestRecordEmailsBillsPaidPlansOnceAndNeverRefuses(t *testing.T) {
	var ingested []map[string]any
	dodoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct{ Events []map[string]any }
		_ = json.NewDecoder(r.Body).Decode(&body)
		ingested = append(ingested, body.Events...)
		w.WriteHeader(200)
	}))
	defer dodoServer.Close()
	cfg := &config.Config{DodoTestMode: true}
	cfg.Plans.FreePlanID = "xem-free"
	usage := &fakeUsage{events: map[string]bool{}}
	s := &subscriptionService{cfg: cfg, usageRepo: usage, dodoClient: dodo.NewClient(cfg).WithBaseURL(dodoServer.URL)}
	customer := "cus_123"
	paid := &models.Subscription{ID: "sub-1", PlanID: "xem-growth", DodoCustomerID: &customer}

	report := structs.RecordUsageRequest{Feature: models.PlanFeatureEmails, Usage: 250, EventID: "batch-1"}
	if err := s.recordEmails(context.Background(), paid, report); err != nil {
		t.Fatal(err)
	}
	if err := s.recordEmails(context.Background(), paid, report); err != nil { // retry
		t.Fatal(err)
	}
	if usage.total != 250 || len(ingested) != 1 {
		t.Fatalf("want 250 emails and one Dodo event, got %d and %d", usage.total, len(ingested))
	}
	e := ingested[0]
	if e["event_id"] != "batch-1" || e["customer_id"] != "cus_123" || e["event_name"] != dodo.EmailMeterEvent || e["metadata"].(map[string]any)["emails"].(float64) != 250 {
		t.Fatalf("unexpected Dodo event %v", e)
	}

	// The free plan is counted locally only, and over-quota sends are still recorded.
	free := &models.Subscription{ID: "sub-2", PlanID: "xem-free", DodoCustomerID: &customer}
	if err := s.recordEmails(context.Background(), free, structs.RecordUsageRequest{Feature: models.PlanFeatureEmails, Usage: 5000, EventID: "batch-2"}); err != nil {
		t.Fatal(err)
	}
	if usage.total != 5250 || len(ingested) != 1 {
		t.Fatalf("free plan must not reach Dodo: total %d, events %d", usage.total, len(ingested))
	}
}
