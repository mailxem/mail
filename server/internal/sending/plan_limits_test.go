package sending

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"kori/internal/billing"
)

// planServer answers the payments subscription lookup with each team's plan.
func planServer(t *testing.T, s *Service, plans map[string]string) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"plan_id":%q,"status":"active"}`, plans[strings.TrimPrefix(r.URL.Path, "/subscriptions/team/")])
	}))
	t.Cleanup(server.Close)
	require.NoError(t, s.DB.AutoMigrate(&billing.EmailUsage{}))
	billing.Enabled = true
	billing.Use(&billing.Reporter{DB: s.DB, URL: server.URL, HTTP: server.Client(), Now: time.Now})
	t.Cleanup(func() { billing.Enabled = false; billing.Use(nil) })
}

func useAutomaticAllowance(t *testing.T, s *Service, team string, used int64) {
	t.Helper()
	require.NoError(t, s.DB.Model(&Account{}).Where("team_id = ?", team).Updates(map[string]any{
		"daily_limit": autoApprovalDailyLimit, "monthly_limit": autoApprovalMonthlyLimit, "monthly_budget_micros": autoApprovalBudgetMicros,
		"day": time.Now().Format("2006-01-02"), "month": time.Now().Format("2006-01"),
		"daily_used": used, "monthly_used": used, "budget_used_micros": used * s.Config.CostMicros,
	}).Error)
}

func TestManagedLimitsFollowThePlan(t *testing.T) {
	ctx := context.Background()

	t.Run("a paid plan sends past the automatic 100 a day", func(t *testing.T) {
		s, team, d, _ := setup(t)
		planServer(t, s, map[string]string{team: "xem-growth"})
		useAutomaticAllowance(t, s, team, 100)
		_, err := s.Submit(ctx, input(team, d))
		require.NoError(t, err)
	})

	t.Run("a yearly plan gets its monthly plan's limit", func(t *testing.T) {
		s, team, d, _ := setup(t)
		planServer(t, s, map[string]string{team: "xem-starter-yearly"})
		useAutomaticAllowance(t, s, team, 200000-1) // Starter ceiling: 20,000 × 10
		_, err := s.Submit(ctx, input(team, d))
		require.NoError(t, err)
		_, err = s.Submit(ctx, input(team, d))
		require.ErrorIs(t, err, ErrLimit)
	})

	t.Run("limits an operator set are kept", func(t *testing.T) {
		s, team, d, _ := setup(t)
		planServer(t, s, map[string]string{team: "xem-scale"})
		require.NoError(t, s.DB.Model(&Account{}).Where("team_id = ?", team).Updates(map[string]any{
			"daily_limit": 5, "monthly_limit": 5, "monthly_budget_micros": 5000,
			"day": time.Now().Format("2006-01-02"), "month": time.Now().Format("2006-01"), "daily_used": 5, "monthly_used": 5,
		}).Error)
		_, err := s.Submit(ctx, input(team, d))
		require.ErrorIs(t, err, ErrLimit)
	})

	t.Run("Free stops at 3,000 emails a month, counting every channel", func(t *testing.T) {
		s, team, d, _ := setup(t)
		planServer(t, s, map[string]string{team: billing.FreePlanID})
		useAutomaticAllowance(t, s, team, 0)
		require.NoError(t, s.DB.Create(&billing.EmailUsage{TeamID: team, Minute: time.Now().UTC().Truncate(time.Minute), Emails: billing.FreeEmails}).Error)
		_, err := s.Submit(ctx, input(team, d))
		require.ErrorIs(t, err, ErrLimit)
		require.ErrorContains(t, err, "Free plan")
		in := input(team, d)
		in.IsTest = true // tests to yourself are free
		_, err = s.Submit(ctx, in)
		require.NoError(t, err)
	})
}
