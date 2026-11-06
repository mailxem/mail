package processors

import (
	"fmt"
	"kori/internal/automation"
	"kori/internal/models"
	"kori/internal/workflowconfig"
	"time"

	"gorm.io/gorm"
)

// sendAt applies the workspace frequency cap and the send window (the automation's,
// else the workspace default) in the contact's time zone, falling back to the
// workspace's and then UTC. It returns now when the email may go out immediately.
func sendAt(db *gorm.DB, ctx *automation.ExecutionContext, now time.Time) (time.Time, bool, error) {
	var rules models.TeamSendTimeDefaults
	if err := db.Where("team_id = ?", ctx.TeamID).Limit(1).Find(&rules).Error; err != nil {
		return now, false, err
	}
	var flow models.Automation
	if err := db.Select("id", "send_window").Where("id = ?", ctx.AutomationID).Limit(1).Find(&flow).Error; err != nil {
		return now, false, err
	}
	window, err := workflowconfig.ParseSendWindow(flow.SendWindow)
	if err == nil && window == nil {
		window, err = workflowconfig.ParseSendWindow(rules.AutomationWindow)
	}
	if err != nil {
		return now, false, err
	}

	at, capped := now, false
	if rules.FrequencyCap > 0 {
		period := workflowconfig.CapPeriod(rules.FrequencyCapPeriod)
		// The cap'th most recent automation email; the cap lifts when it leaves the period.
		// The literal LIKE matches the partial index automation_emails_contact.
		var sent []time.Time
		if err := db.Model(&models.Email{}).
			Where("contact_id = ? AND delivery_key LIKE 'automation:%' AND created_at > ?", ctx.ContactID, now.Add(-period)).
			Order("created_at DESC").Offset(rules.FrequencyCap-1).Limit(1).
			Pluck("created_at", &sent).Error; err != nil {
			return now, false, fmt.Errorf("count automation emails: %w", err)
		}
		if len(sent) > 0 {
			at, capped = sent[0].Add(period), true
		}
	}
	if window != nil {
		tz := ""
		if ctx.Contact != nil {
			tz = ctx.Contact.Timezone
		}
		at = window.Next(at, workflowconfig.Location(tz, rules.DefaultTimezone))
	}
	return at, capped, nil
}

func nextNodes(db *gorm.DB, ctx *automation.ExecutionContext, node *models.AutomationNode) ([]string, error) {
	var edges []models.AutomationNodeEdge
	if err := db.Where("automation_id = ? AND source_id = ?", ctx.AutomationID, node.ID).Find(&edges).Error; err != nil {
		return nil, fmt.Errorf("failed to load edges: %w", err)
	}
	next := make([]string, len(edges))
	for i, edge := range edges {
		next[i] = edge.TargetID
	}
	return next, nil
}
