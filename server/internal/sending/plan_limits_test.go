package sending

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"kori/internal/billing"
)

// planServer answers the payments subscription lookup with each team's plan.
func planServer(t *testing.T, s *Service, plans map[string]string) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		plan := plans[strings.TrimPrefix(r.URL.Path, "/subscriptions/team/")]
		paid := !strings.HasSuffix(plan, "+trial")
		fmt.Fprintf(w, `{"plan_id":%q,"status":"active","has_paid":%t}`, strings.TrimSuffix(plan, "+trial"), paid)
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
		"day": s.Now().Format("2006-01-02"), "month": s.Now().Format("2006-01"),
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

	t.Run("a paid plan has no daily limit below its month", func(t *testing.T) {
		s, team, d, _ := setup(t)
		planServer(t, s, map[string]string{team: "xem-starter-yearly"})
		useAutomaticAllowance(t, s, team, 40000-1) // all of Starter's month in one day
		_, err := s.Submit(ctx, input(team, d))
		require.NoError(t, err)
		_, err = s.Submit(ctx, input(team, d))
		require.ErrorIs(t, err, ErrLimit)
	})

	t.Run("a paid plan's month allows twice its emails", func(t *testing.T) {
		s, team, d, _ := setup(t)
		planServer(t, s, map[string]string{team: "xem-starter"})
		useAutomaticAllowance(t, s, team, 0)
		require.NoError(t, s.DB.Model(&Account{}).Where("team_id = ?", team).Update("monthly_used", 40000-1).Error)
		_, err := s.Submit(ctx, input(team, d))
		require.NoError(t, err)
		_, err = s.Submit(ctx, input(team, d))
		require.ErrorIs(t, err, ErrLimit)
	})

	t.Run("a trial sends only its included emails", func(t *testing.T) {
		s, team, d, _ := setup(t)
		planServer(t, s, map[string]string{team: "xem-starter+trial"})
		useAutomaticAllowance(t, s, team, 0)
		require.NoError(t, s.DB.Model(&Account{}).Where("team_id = ?", team).Update("monthly_used", 20000-1).Error)
		_, err := s.Submit(ctx, input(team, d))
		require.NoError(t, err)
		_, err = s.Submit(ctx, input(team, d))
		require.ErrorIs(t, err, ErrLimit)
	})

	t.Run("Free sends at most 100 a day", func(t *testing.T) {
		s, team, d, _ := setup(t)
		planServer(t, s, map[string]string{team: billing.FreePlanID})
		useAutomaticAllowance(t, s, team, billing.FreeDailyEmails)
		_, err := s.Submit(ctx, input(team, d))
		require.ErrorIs(t, err, ErrLimit)
	})

	t.Run("limits an operator set are kept", func(t *testing.T) {
		s, team, d, _ := setup(t)
		planServer(t, s, map[string]string{team: "xem-scale"})
		require.NoError(t, s.DB.Model(&Account{}).Where("team_id = ?", team).Updates(map[string]any{
			"daily_limit": 5, "monthly_limit": 5, "monthly_budget_micros": 5000,
			"day": s.Now().Format("2006-01-02"), "month": s.Now().Format("2006-01"), "daily_used": 5, "monthly_used": 5,
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

func TestStatusShowsThePlanLimits(t *testing.T) {
	s, team, _, _ := setup(t)
	planServer(t, s, map[string]string{team: "xem-starter+trial"})
	useAutomaticAllowance(t, s, team, 0)
	e := echo.New()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())
	c.Set("teamID", team)
	require.NoError(t, s.status(c))
	var body struct{ Account Account }
	require.NoError(t, json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &body))
	require.Equal(t, int64(20000), body.Account.MonthlyLimit)
	require.Equal(t, int64(20000), body.Account.DailyLimit)
	require.Equal(t, 20000*s.Config.CostMicros, body.Account.MonthlyBudgetMicros)
}
