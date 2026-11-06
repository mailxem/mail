package tasks

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
)

const (
	TaskTypeAutomationDateTick    = "automation:date-tick"
	TaskTypeAutomationSegmentTick = "automation:segment-tick"
)

// AutomationTriggers runs the time-based automation triggers. It is
// *automation.TriggerManager, set from main.go to avoid a circular import.
type AutomationTriggers interface {
	RunDateTriggers(context.Context, time.Time) error
	RefreshSegmentTriggers(context.Context) error
}

func (h *TaskHandler) SetAutomationTriggers(triggers AutomationTriggers) {
	h.triggers = triggers
}

func (h *TaskHandler) HandleAutomationDateTick(ctx context.Context, _ *asynq.Task) error {
	if h.triggers == nil {
		return nil
	}
	return h.triggers.RunDateTriggers(ctx, time.Now())
}

func (h *TaskHandler) HandleAutomationSegmentTick(ctx context.Context, _ *asynq.Task) error {
	if h.triggers == nil {
		return nil
	}
	return h.triggers.RefreshSegmentTriggers(ctx)
}
