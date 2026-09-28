package tasks

import (
	"context"
	"github.com/hibiken/asynq"
	"kori/internal/marketing"
	"kori/internal/models"
	"time"
)

const TaskTypeFormCompletionTick = "forms:completion-tick"

func (h *TaskHandler) HandleFormCompletionTick(ctx context.Context, _ *asynq.Task) error {
	// Bounded cleanup also expires saves on forms that no longer receive traffic.
	var expired []string
	if err := h.db.WithContext(ctx).Model(&models.FormProgress{}).Where("expires_at <= ?", time.Now().UTC()).Limit(500).Pluck("id", &expired).Error; err != nil {
		return err
	}
	if len(expired) > 0 {
		if err := h.db.WithContext(ctx).Where("id IN ? AND expires_at <= ?", expired, time.Now().UTC()).Delete(&models.FormProgress{}).Error; err != nil {
			return err
		}
	}
	return marketing.DispatchFormCompletions(ctx, h.db, func(ctx context.Context, executionID, automationID, contactID string, variables map[string]interface{}) error {
		return h.taskClient.EnqueueAutomationTask(ctx, AutomationExecuteTask{ExecutionID: executionID, AutomationID: automationID, ContactID: contactID, TriggerData: variables}, 0)
	}, 100)
}
