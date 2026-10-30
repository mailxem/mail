package notifications

import (
	"time"

	"gorm.io/gorm"
	"kori/internal/events"
	"kori/internal/models"
)

// WatchAutomationFailures records each failed automation run for the daily
// owner alert. The reason names the automation so the owner knows where to look.
func WatchAutomationFailures(db *gorm.DB) {
	events.On("automation.execution.failed", func(data interface{}) {
		var execution models.AutomationExecution
		switch e := data.(type) {
		case models.AutomationExecution:
			execution = e
		case *models.AutomationExecution:
			if e == nil {
				return
			}
			execution = *e
		default:
			return
		}
		_ = RecordAutomationFailure(db, execution, time.Now())
	})
}

// RecordAutomationFailure queues the alert event for one failed run.
func RecordAutomationFailure(db *gorm.DB, execution models.AutomationExecution, now time.Time) error {
	var automation models.Automation
	if err := db.Select("id", "team_id", "name").First(&automation, "id = ?", execution.AutomationID).Error; err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		return models.QueueAlert(tx, automation.TeamID, "automation", execution.ID, "automation_failed", automation.Name+": "+execution.Error, now)
	})
}
