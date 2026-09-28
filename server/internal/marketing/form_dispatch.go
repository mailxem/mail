package marketing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/models"
)

// FormAutomationEnqueue must deduplicate the supplied execution ID. Keeping this
// boundary independent of Redis makes an outage/retry testable without a queue.
type FormAutomationEnqueue func(context.Context, string, string, string, map[string]interface{}) error

// DispatchFormCompletions drains the durable submission outbox. A database lock
// serializes dispatchers; stable execution IDs also cover a crash after enqueue
// but before commit. An event is never acknowledged when enqueue fails.
func DispatchFormCompletions(ctx context.Context, db *gorm.DB, enqueue FormAutomationEnqueue, limit int) error {
	if enqueue == nil {
		return errors.New("form automation queue is unavailable")
	}
	if limit < 1 || limit > 500 {
		limit = 100
	}
	var pending []models.FormCompletionEvent
	if err := db.WithContext(ctx).Where("dispatched_at IS NULL").Order("created_at, id").Limit(limit).Find(&pending).Error; err != nil {
		return err
	}
	var failures []error
	for _, candidate := range pending {
		err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			var event models.FormCompletionEvent
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND dispatched_at IS NULL", candidate.ID).First(&event).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil
				}
				return err
			}
			var payload map[string]interface{}
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				return fmt.Errorf("invalid completion event %s", event.ID)
			}
			consent, _ := payload["consent"].(bool)
			if event.ContactID != "" && consent {
				var contact models.Contact
				err := tx.Where("id = ? AND team_id = ? AND is_deleted = ?", event.ContactID, event.TeamID, false).First(&contact).Error
				if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
				eligible := err == nil && contact.Status == models.SubscriberStatusActive
				if eligible {
					// A recipient can opt out after submitting but before this durable
					// event is dispatched. Never start a journey for that address.
					var suppressed int64
					if err := tx.Model(&models.SuppressionList{}).
						Where("team_id = ? AND LOWER(email_address) = ? AND is_active = ? AND is_deleted = ? AND (expires_at IS NULL OR expires_at > ?)", event.TeamID, strings.ToLower(strings.TrimSpace(contact.Email)), true, false, time.Now().UTC()).
						Count(&suppressed).Error; err != nil {
						return err
					}
					eligible = suppressed == 0
				}
				if eligible {
					var workflows []models.Automation
					// Newly activated/edited workflows must not consume historical submissions.
					if err := tx.Preload("Nodes", "type = ? AND is_deleted = ?", models.NodeTypeStart, false).
						Where("team_id = ? AND is_deleted = ? AND is_active = ? AND trigger_event = ? AND updated_at <= ?", event.TeamID, false, true, "form.completed", event.CreatedAt).
						Order("id").Find(&workflows).Error; err != nil {
						return err
					}
					for _, workflow := range workflows {
						if !formWorkflowMatches(workflow, event.FormID) {
							continue
						}
						variables := map[string]interface{}{"event": "form.completed", "form_id": event.FormID, "submission_id": event.SubmissionID, "form_version": payload["version"]}
						if fields, ok := payload["fields"].(map[string]interface{}); ok {
							for key, value := range fields {
								name := "form_" + key
								if _, reserved := variables[name]; reserved {
									continue // Answers must not replace trusted event metadata.
								}
								if text, ok := value.(string); ok {
									variables[name] = text
								}
							}
						}
						executionID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("xem:form:"+event.ID+":"+workflow.ID)).String()
						if err := enqueue(ctx, executionID, workflow.ID, contact.ID, variables); err != nil {
							return err
						}
					}
				}
			}
			return tx.Model(&models.FormCompletionEvent{}).Where("id = ?", event.ID).Update("dispatched_at", time.Now().UTC()).Error
		})
		if err != nil {
			failures = append(failures, fmt.Errorf("dispatch form event %s: %w", candidate.ID, err))
		}
	}
	return errors.Join(failures...)
}

func formWorkflowMatches(workflow models.Automation, formID string) bool {
	if len(workflow.Nodes) != 1 {
		return false
	}
	var config struct {
		FormID string `json:"formId"`
	}
	if len(workflow.Nodes[0].Data) > 0 && json.Unmarshal(workflow.Nodes[0].Data, &config) != nil {
		return false
	}
	return config.FormID == formID // An explicit form prevents accidental workspace-wide mail.
}
