package tasks

import (
	"context"
	"github.com/hibiken/asynq"
	"kori/internal/marketing"
	"os"
	"time"
)

const TaskTypeListCleanup = "lists:cleanup"

func (h *TaskHandler) HandleListCleanup(ctx context.Context, _ *asynq.Task) error {
	return marketing.New(h.db, h.cfg.JWT.Secret, os.Getenv("PUBLIC_API_URL")).CleanUpLists(ctx, time.Now())
}
