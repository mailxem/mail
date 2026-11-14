package tasks

import (
	"context"

	"github.com/hibiken/asynq"
)

const TaskTypeSummaryTick = "summaries:tick"

// Summaries go out from 08:00 in each workspace's time zone, so the tick runs
// often enough to catch half-hour zones; queuing is idempotent per period.
func (h *TaskHandler) HandleSummaryTick(ctx context.Context, _ *asynq.Task) error {
	if h.QueueSummaries == nil {
		return nil
	}
	return h.QueueSummaries(ctx)
}
