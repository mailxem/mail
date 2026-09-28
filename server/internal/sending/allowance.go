package sending

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

// Upgrade only the recognizable legacy automatic grant. Never infer a free plan
// from a custom limit or an operator-managed account, and never reset usage.
func upgradeStarterAllowance(db *gorm.DB) error {
	var teams []string
	if err := db.Model(&Account{}).Where("approved = ? AND suspended = ? AND paused = ? AND daily_limit = ? AND monthly_limit = ? AND monthly_budget_micros = ?", true, false, false, 50, 200, 200000).Pluck("team_id", &teams).Error; err != nil {
		return err
	}
	for _, team := range teams {
		if err := db.Transaction(func(tx *gorm.DB) error {
			var a Account
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&a, "team_id = ?", team).Error; err != nil {
				return err
			}
			if !a.Approved || a.Paused || a.Suspended || a.DailyLimit != 50 || a.MonthlyLimit != 200 || a.MonthlyBudgetMicros != 200000 {
				return nil
			}
			var automatic, operator int64
			if err := tx.Model(&Audit{}).Where("team_id = ? AND actor = ? AND action = ?", team, "system:dns-verification", "auto_approve").Count(&automatic).Error; err != nil {
				return err
			}
			if err := tx.Model(&Audit{}).Where("team_id = ? AND (actor LIKE ? OR action = ?)", team, "operator:%", "starter_allowance_v2").Count(&operator).Error; err != nil {
				return err
			}
			if automatic == 0 || operator != 0 {
				return nil
			}
			if err := tx.Model(&a).Updates(map[string]any{"daily_limit": autoApprovalDailyLimit, "monthly_limit": autoApprovalMonthlyLimit, "monthly_budget_micros": autoApprovalBudgetMicros}).Error; err != nil {
				return err
			}
			return tx.Create(&Audit{ID: uuid.NewString(), TeamID: team, Actor: "system:free-plan", Action: "starter_allowance_v2", CreatedAt: time.Now()}).Error
		}); err != nil {
			return err
		}
	}
	return nil
}
