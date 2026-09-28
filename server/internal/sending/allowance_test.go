package sending

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestStarterAllowanceUpgradePreservesDecisionsAndUsage(t *testing.T) {
	for _, scenario := range []string{"automatic", "operator", "custom", "paused", "suspended", "unapproved", "no-audit"} {
		t.Run(scenario, func(t *testing.T) {
			s, team, _, _ := setup(t)
			fields := map[string]any{"approved": true, "suspended": false, "paused": false, "daily_limit": 50, "monthly_limit": 200, "monthly_budget_micros": 200000, "daily_used": 17, "monthly_used": 123, "budget_used_micros": 123000}
			switch scenario {
			case "custom":
				fields["daily_limit"] = 7
			case "paused":
				fields["paused"] = true
			case "suspended":
				fields["suspended"] = true
			case "unapproved":
				fields["approved"] = false
			}
			require.NoError(t, s.DB.Model(&Account{}).Where("team_id = ?", team).Updates(fields).Error)
			if scenario != "no-audit" {
				require.NoError(t, s.DB.Create(&Audit{ID: uuid.NewString(), TeamID: team, Actor: "system:dns-verification", Action: "auto_approve", CreatedAt: time.Now()}).Error)
			}
			if scenario == "operator" {
				require.NoError(t, s.DB.Create(&Audit{ID: uuid.NewString(), TeamID: team, Actor: "operator:reviewer", Action: "approve"}).Error)
			}
			require.NoError(t, upgradeStarterAllowance(s.DB))
			require.NoError(t, upgradeStarterAllowance(s.DB))
			var a Account
			require.NoError(t, s.DB.First(&a, "team_id = ?", team).Error)
			require.EqualValues(t, 17, a.DailyUsed)
			require.EqualValues(t, 123, a.MonthlyUsed)
			require.EqualValues(t, 123000, a.BudgetUsedMicros)
			var count int64
			require.NoError(t, s.DB.Model(&Audit{}).Where("team_id = ? AND action = ?", team, "starter_allowance_v2").Count(&count).Error)
			if scenario == "automatic" {
				require.EqualValues(t, 100, a.DailyLimit)
				require.EqualValues(t, 3000, a.MonthlyLimit)
				require.EqualValues(t, 3000000, a.MonthlyBudgetMicros)
				require.EqualValues(t, 1, count)
			} else {
				require.EqualValues(t, fields["daily_limit"], a.DailyLimit)
				require.EqualValues(t, 200, a.MonthlyLimit)
				require.Zero(t, count)
			}
		})
	}
}

func TestFreePlanRecipientAndBudgetBoundary(t *testing.T) {
	s, team, d, _ := setup(t)
	require.NoError(t, s.DB.Model(&Account{}).Where("team_id = ?", team).Updates(map[string]any{"daily_limit": 100, "monthly_limit": 3000, "monthly_budget_micros": 3000000, "day": s.Now().Format("2006-01-02"), "month": s.Now().Format("2006-01"), "daily_used": 0, "monthly_used": 2999, "budget_used_micros": 2999000}).Error)
	_, err := s.Submit(context.Background(), input(team, d))
	require.NoError(t, err)
	_, err = s.Submit(context.Background(), input(team, d))
	require.ErrorIs(t, err, ErrLimit)
	var a Account
	require.NoError(t, s.DB.First(&a, "team_id = ?", team).Error)
	require.EqualValues(t, 3000, a.MonthlyUsed)
	require.EqualValues(t, 3000000, a.BudgetUsedMicros)
}
