package repositories

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/dodopayments/dodopayments-go"
	"github.com/mailxem/payments.go/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Needs Postgres: PAYMENTS_TEST_DATABASE_URL=postgres://... go test ./internal/repositories/
func TestTeamLookupPrefersAnActivePaidPlan(t *testing.T) {
	dsn := os.Getenv("PAYMENTS_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("PAYMENTS_TEST_DATABASE_URL is not set")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Plan{}, &models.PlanFeatures{}, &models.Subscription{}); err != nil {
		t.Fatal(err)
	}
	db.Exec("DELETE FROM subscriptions WHERE team_id = 'team-lookup'")
	for _, p := range []models.Plan{{ID: "lookup-free", DodoPlanID: "lookup-free", Price: decimal.Zero}, {ID: "lookup-paid", DodoPlanID: "pdt_lookup", Price: decimal.NewFromInt(9)}} {
		db.Where("id = ?", p.ID).FirstOrCreate(&p)
	}
	repo := NewSubscriptionRepository(db)
	ctx := context.Background()
	now := time.Now()
	add := func(plan string, status dodopayments.SubscriptionStatus, age time.Duration) {
		s := &models.Subscription{TeamID: "team-lookup", PlanID: plan, Status: status, BillingCycle: "MONTHLY", UpdatedAt: now.Add(-age)}
		if err := db.Create(s).Error; err != nil {
			t.Fatal(err)
		}
		db.Model(s).UpdateColumn("updated_at", now.Add(-age))
	}
	expect := func(plan string, status dodopayments.SubscriptionStatus) {
		t.Helper()
		got, err := repo.GetByTeamID(ctx, "team-lookup")
		if err != nil || got.PlanID != plan || got.Status != status {
			t.Fatalf("want %s %s, got %+v %v", plan, status, got, err)
		}
	}

	add("lookup-free", dodopayments.SubscriptionStatusActive, 3*time.Hour)
	add("lookup-paid", dodopayments.SubscriptionStatusPending, time.Hour)
	expect("lookup-free", dodopayments.SubscriptionStatusActive) // an unpaid checkout never replaces Free

	add("lookup-paid", dodopayments.SubscriptionStatusActive, 2*time.Hour)
	db.Model(&models.Subscription{}).Where("team_id = 'team-lookup' AND plan_id = 'lookup-free'").UpdateColumn("updated_at", now)
	expect("lookup-paid", dodopayments.SubscriptionStatusActive) // a newer Free row never hides the paid plan

	db.Model(&models.Subscription{}).Where("team_id = 'team-lookup' AND status = 'active' AND plan_id = 'lookup-paid'").UpdateColumn("status", dodopayments.SubscriptionStatusCancelled)
	expect("lookup-free", dodopayments.SubscriptionStatusActive) // cancelled plans no longer count
}
