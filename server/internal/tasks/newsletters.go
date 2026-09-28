package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/hibiken/asynq"
	"kori/internal/marketing"
	"kori/internal/models"
	"os"
	"time"
)

const TaskTypeNewsletterTick = "newsletter:tick"

func (h *TaskHandler) HandleNewsletterTick(ctx context.Context, t *asynq.Task) error {
	var failures []error
	s := marketing.New(h.db, h.cfg.JWT.Secret, os.Getenv("PUBLIC_API_URL"))
	// Unconfigured installations may use transactional mail without newsletters.
	var count int64
	if err := h.db.Model(&models.Newsletter{}).Where("status = ?", "SCHEDULED").Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		if err := s.MaterializeDue(ctx, time.Now()); err != nil {
			failures = append(failures, err)
		}
	}
	var editions []models.Campaign
	if err := h.db.Where("newsletter_id IS NOT NULL AND status IN ?", []string{"SCHEDULED", "SENDING"}).Limit(100).Find(&editions).Error; err != nil {
		return err
	}
	for _, edition := range editions {
		if err := s.PrepareEdition(ctx, edition.ID); err != nil {
			failures = append(failures, err)
		}
	}
	// Durable database outbox: retry enqueue after a Redis outage without recreating mail.
	var messages []models.Email
	if err := h.db.Where("delivery_key IS NOT NULL AND status = ?", models.EmailStatusPending).Order("created_at").Limit(500).Find(&messages).Error; err != nil {
		return err
	}
	for _, message := range messages {
		payload, _ := json.Marshal(EmailTask{EmailID: message.ID, SMTPConfigID: message.SMTPConfigID})
		_, err := h.taskClient.client.EnqueueContext(ctx, asynq.NewTask(TaskTypeEmailSend, payload), asynq.TaskID(message.ID), asynq.Queue(QueueCritical), asynq.MaxRetry(3), asynq.Retention(24*time.Hour))
		if err != nil && !errors.Is(err, asynq.ErrTaskIDConflict) && !errors.Is(err, asynq.ErrDuplicateTask) {
			return err
		}
	}
	return errors.Join(failures...)
}
