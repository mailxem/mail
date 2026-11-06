package automation

import (
	"context"
	"encoding/json"
	"errors"
	"kori/internal/events"
	"kori/internal/models"
	"kori/internal/sending"
	"kori/internal/services"
	"kori/internal/tasks"
	"kori/internal/workflowconfig"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// errExited stops a run whose contact has left the workflow.
var errExited = errors.New("contact left the workflow")

var liveStatuses = []models.ExecutionStatus{models.ExecutionStatusRunning, models.ExecutionStatusWaiting, models.ExecutionStatusPaused}

// waitEvents are the bus events that can end a wait-until step, an exit check or a goal.
var waitEvents = []string{"email.opened", "email.clicked", "email.unsubscribed", "email.bounced", "email.complained", "contact.updated", "contact.deleted", "contact.tag_added", "event.received"}

// Listen subscribes the engine to the events running contacts react to.
func (e *Engine) Listen() {
	for _, name := range waitEvents {
		events.On(name, func(data interface{}) {
			if err := e.HandleEvent(context.Background(), name, data); err != nil {
				engineLog.Warn("automation event %s: %v", name, err)
			}
		})
	}
}

// HandleEvent exits contacts who may no longer be emailed or reached a goal, and
// sends contacts whose wait-until step was waiting for this event down its "happened" path.
func (e *Engine) HandleEvent(ctx context.Context, name string, data interface{}) error {
	var contactID, emailID, url, tag, appEvent string
	switch d := data.(type) {
	case *models.EmailTracking:
		contactID, emailID, url = d.ContactID, d.EmailID, d.URL
	case *models.EmailEvent:
		contactID = d.ContactID
	case *models.Email:
		contactID = d.ContactID
	case *models.Contact:
		contactID = d.ID
	default:
		// *models.TagAdded {Contact, Tag} and *models.ContactEvent {contactId, name},
		// read by shape so this works before and after the triggers change lands.
		var p struct {
			ContactID string `json:"contactId"`
			Name      string `json:"name"`
			Contact   struct{ ID string }
			Tag       struct{ Name string }
		}
		raw, _ := json.Marshal(data)
		json.Unmarshal(raw, &p)
		contactID, tag, appEvent = p.ContactID, p.Tag.Name, p.Name
		if p.Contact.ID != "" {
			contactID = p.Contact.ID
		}
	}
	if contactID == "" {
		return nil
	}
	db := e.db.WithContext(ctx)
	var runs []models.AutomationExecution
	if err := db.Where("contact_id = ? AND status IN ?", contactID, liveStatuses).Find(&runs).Error; err != nil || len(runs) == 0 {
		return err
	}

	reason := "contact_deleted"
	if name != "contact.deleted" {
		var err error
		if reason, err = e.contactExitReason(ctx, contactID); err != nil {
			return err
		}
	}
	if reason != "" {
		return e.exit(ctx, reason, "contact_id = ?", contactID)
	}

	var contact models.Contact
	if name == "contact.updated" {
		if err := db.First(&contact, "id = ?", contactID).Error; err != nil {
			return err
		}
	}
	var deliveryKey string
	if emailID != "" {
		db.Model(&models.Email{}).Where("id = ?", emailID).Pluck("delivery_key", &deliveryKey)
	}

	for _, run := range runs {
		// Opens and clicks on this run's own emails count towards the workflow's goals.
		if goal := map[string]models.GoalType{"email.opened": models.GoalTypeEmailOpen, "email.clicked": models.GoalTypeEmailClick}[name]; goal != "" && strings.HasPrefix(deliveryKey, "automation:"+run.ID+":") {
			left, err := e.checkGoal(ctx, &run, goal)
			if err != nil {
				return err
			}
			if left {
				continue
			}
		}
		if run.Status != models.ExecutionStatusWaiting {
			continue
		}
		var node models.AutomationNode
		if err := db.Where("id = ? AND automation_id = ? AND type = ?", run.CurrentNodeID, run.AutomationID, models.NodeTypeWaitUntil).First(&node).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return err
		}
		var w workflowconfig.WaitUntil
		if json.Unmarshal(node.Data, &w) != nil {
			continue
		}
		var vars map[string]interface{}
		json.Unmarshal(run.Variables, &vars)
		lastEmail, _ := vars["last_email_id"].(string)
		sameEmail := lastEmail == "" || lastEmail == emailID
		var happened bool
		switch w.Event {
		case "opened":
			happened = name == "email.opened" && sameEmail
		case "clicked":
			happened = name == "email.clicked" && sameEmail && (w.URL == "" || w.URL == url)
		case "tag_added":
			happened = name == "contact.tag_added" && strings.EqualFold(strings.TrimSpace(tag), strings.TrimSpace(w.Tag))
		case "field_changed":
			if name == "contact.updated" {
				fields := ExecutionContext{Contact: &contact}
				fields.RefreshContactVariables()
				value, _ := fields.Variables["contact_"+w.Field].(string)
				happened = strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(w.Value))
			}
		case "app_event":
			happened = name == "event.received" && strings.EqualFold(strings.TrimSpace(appEvent), strings.TrimSpace(w.Name))
		}
		if !happened {
			continue
		}
		var edge models.AutomationNodeEdge
		if err := db.Where("automation_id = ? AND source_id = ? AND label = ?", run.AutomationID, node.ID, "happened").First(&edge).Error; err != nil {
			return err
		}
		// The continuation claims the wait in Execute, so a timeout racing it runs at most one path.
		task := tasks.AutomationExecuteTask{ExecutionID: run.ID, AutomationID: run.AutomationID, ContactID: contactID, TriggerData: map[string]interface{}{}, CurrentNodeID: edge.TargetID}
		if err := e.taskClient.EnqueueAutomationTask(ctx, task, 0); err != nil {
			return err
		}
	}
	return nil
}

// checkGoal records goal conversions for run and exits it when the workflow says to.
func (e *Engine) checkGoal(ctx context.Context, run *models.AutomationExecution, goal models.GoalType) (bool, error) {
	if err := services.NewGoalTrackingService(e.db).CheckAndRecordGoal(ctx, run.AutomationID, run.ContactID, run.ID, goal, 0, nil); err != nil {
		return false, err
	}
	var a models.Automation
	if err := e.db.WithContext(ctx).Select("exit_on_goal").First(&a, "id = ?", run.AutomationID).Error; err != nil {
		return false, err
	}
	if a.ExitOnGoal != nil && !*a.ExitOnGoal {
		return false, nil
	}
	var n int64
	if err := e.db.WithContext(ctx).Model(&models.GoalConversion{}).Where("execution_id = ?", run.ID).Count(&n).Error; err != nil || n == 0 {
		return false, err
	}
	return true, e.exit(ctx, "goal_reached", "id = ?", run.ID)
}

// contactExitReason says why a contact must leave every workflow, or "" when they may stay.
func (e *Engine) contactExitReason(ctx context.Context, contactID string) (string, error) {
	db := e.db.WithContext(ctx)
	var c models.Contact
	err := db.First(&c, "id = ?", contactID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && c.IsDeleted) {
		return "contact_deleted", nil
	}
	if err != nil {
		return "", err
	}
	switch c.Status {
	case models.SubscriberStatusUnsubscribed:
		return "unsubscribed", nil
	case models.SubscriberStatusBounced:
		return "bounced", nil
	case models.SubscriberStatusComplained:
		return "complained", nil
	}
	suppressed, err := sending.Suppressed(db, c.TeamID, strings.ToLower(strings.TrimSpace(c.Email)), true)
	if err != nil {
		return "", err
	}
	if suppressed {
		return "suppressed", nil
	}
	return "", nil
}

// exit moves the matching live runs to EXITED with reason.
func (e *Engine) exit(ctx context.Context, reason string, where string, args ...interface{}) error {
	var ids []string
	err := e.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		q := tx.Model(&models.AutomationExecution{}).Where("status IN ?", liveStatuses).Where(where, args...)
		if err := q.Pluck("id", &ids).Error; err != nil || len(ids) == 0 {
			return err
		}
		return tx.Model(&models.AutomationExecution{}).Where("id IN ? AND status IN ?", ids, liveStatuses).
			Updates(map[string]interface{}{"status": models.ExecutionStatusExited, "exit_reason": reason, "completed_at": time.Now()}).Error
	})
	for _, id := range ids {
		events.Emit("automation.execution.exited", map[string]interface{}{"executionId": id, "reason": reason})
	}
	return err
}

// claimResume lets exactly one continuation leave a wait: the timeout, or the event that came first.
func (e *Engine) claimResume(ctx context.Context, execution *models.AutomationExecution, automation *models.Automation, nodeID string) (bool, error) {
	if execution.Status != models.ExecutionStatusWaiting {
		// Only a queue retry of the continuation that already won may run again.
		return execution.ResumeNodeID == nodeID, nil
	}
	follows, exists := false, false
	for _, edge := range automation.Edges {
		follows = follows || (edge.SourceID == execution.CurrentNodeID && edge.TargetID == nodeID)
	}
	for _, node := range automation.Nodes {
		exists = exists || node.ID == nodeID
	}
	// A step missing from the contact's version is still claimed, so the run
	// exits with a reason instead of waiting forever.
	if exists && !follows {
		return false, nil
	}
	res := e.db.WithContext(ctx).Model(&models.AutomationExecution{}).
		Where("id = ? AND status = ? AND current_node_id = ?", execution.ID, models.ExecutionStatusWaiting, execution.CurrentNodeID).
		Updates(map[string]interface{}{"status": models.ExecutionStatusRunning, "resume_node_id": nodeID})
	if res.Error != nil || res.RowsAffected == 0 {
		return false, res.Error
	}
	execution.Status, execution.ResumeNodeID = models.ExecutionStatusRunning, nodeID
	return true, nil
}

// admit applies the workflow's re-entry rule and creates the run; it returns false when the contact may not enter.
func (e *Engine) admit(ctx context.Context, execution *models.AutomationExecution) (bool, error) {
	err := e.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Serialises entries per contact so two triggers cannot both start a run.
		// ponytail: contact row lock; also blocks the contact's other workflows briefly.
		var locked []models.Contact
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").Where("id = ?", execution.ContactID).Find(&locked).Error; err != nil {
			return err
		}
		var a models.Automation
		if err := tx.Select("reentry_mode", "reentry_days").Where("id = ?", execution.AutomationID).Find(&a).Error; err != nil {
			return err
		}
		previous := tx.Model(&models.AutomationExecution{}).Where("automation_id = ? AND contact_id = ?", execution.AutomationID, execution.ContactID)
		switch a.ReentryMode {
		case "once":
		case "after_days":
			previous = previous.Where("status IN ? OR started_at > ?", liveStatuses, time.Now().AddDate(0, 0, -a.ReentryDays))
		default:
			previous = previous.Where("status IN ?", liveStatuses)
		}
		var n int64
		if err := previous.Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			execution.ID = ""
			return nil
		}
		if execution.CurrentNodeID == "" {
			tx = tx.Omit("CurrentNodeID") // Postgres rejects '' as a uuid
		}
		return tx.Create(execution).Error
	})
	return execution.ID != "", err
}
