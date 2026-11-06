package tasks

import (
	"context"

	"github.com/hibiken/asynq"
)

const TaskTypePostmasterTick = "postmaster:tick"

// Google publishes a day's Postmaster data late and at no fixed hour, so the
// tick runs a few times a day; the sync acts on each day only once.
func (h *TaskHandler) HandlePostmasterTick(ctx context.Context, _ *asynq.Task) error {
	if h.PostmasterSync == nil {
		return nil
	}
	return h.PostmasterSync(ctx)
}
