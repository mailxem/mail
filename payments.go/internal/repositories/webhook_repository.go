package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mailxem/payments.go/internal/models"
	"gorm.io/gorm"
)

type WebhookRepository interface {
	SaveEvent(ctx context.Context, event *models.WebhookEvent) error
	GetEvent(ctx context.Context, id string) (*models.WebhookEvent, error)
	GetEventsByType(ctx context.Context, eventType string, limit int, offset int) ([]*models.WebhookEvent, error)
	GetPendingEvents(ctx context.Context, limit int) ([]*models.WebhookEvent, error)
	UpdateEventStatus(ctx context.Context, id string, status models.WebhookStatus, errorMsg *string) error
	MarkEventProcessed(ctx context.Context, id string) error
	GetEventsByStatus(ctx context.Context, status models.WebhookStatus, limit int, offset int) ([]*models.WebhookEvent, error)
	DeleteOldEvents(ctx context.Context, olderThan time.Time) error
}

type WebhookLogRepository interface {
	SaveLog(ctx context.Context, log *models.WebhookLog) error
	GetLogsByEvent(ctx context.Context, eventID string) ([]*models.WebhookLog, error)
	GetLogsByLevel(ctx context.Context, level models.LogLevel, limit int, offset int) ([]*models.WebhookLog, error)
	DeleteOldLogs(ctx context.Context, olderThan time.Time) error
}

type webhookRepository struct {
	db *gorm.DB
}

type webhookLogRepository struct {
	db *gorm.DB
}

func NewWebhookRepository(db *gorm.DB) WebhookRepository {
	return &webhookRepository{db: db}
}

func NewWebhookLogRepository(db *gorm.DB) WebhookLogRepository {
	return &webhookLogRepository{db: db}
}

// WebhookRepository implementations
func (r *webhookRepository) SaveEvent(ctx context.Context, event *models.WebhookEvent) error {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}

	return r.db.WithContext(ctx).Create(event).Error
}

func (r *webhookRepository) GetEvent(ctx context.Context, id string) (*models.WebhookEvent, error) {
	var event models.WebhookEvent
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *webhookRepository) GetEventsByType(ctx context.Context, eventType string, limit int, offset int) ([]*models.WebhookEvent, error) {
	var events []*models.WebhookEvent
	err := r.db.WithContext(ctx).
		Where("type = ?", eventType).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error
	return events, err
}

func (r *webhookRepository) GetPendingEvents(ctx context.Context, limit int) ([]*models.WebhookEvent, error) {
	var events []*models.WebhookEvent
	err := r.db.WithContext(ctx).
		Where("status = ?", models.WebhookPending).
		Order("created_at ASC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

func (r *webhookRepository) UpdateEventStatus(ctx context.Context, id string, status models.WebhookStatus, errorMsg *string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if status == models.WebhookProcessed {
		now := time.Now()
		updates["processed_at"] = &now
	}

	if errorMsg != nil {
		updates["error"] = *errorMsg
	}

	return r.db.WithContext(ctx).
		Model(&models.WebhookEvent{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *webhookRepository) MarkEventProcessed(ctx context.Context, id string) error {
	return r.UpdateEventStatus(ctx, id, models.WebhookProcessed, nil)
}

func (r *webhookRepository) GetEventsByStatus(ctx context.Context, status models.WebhookStatus, limit int, offset int) ([]*models.WebhookEvent, error) {
	var events []*models.WebhookEvent
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error
	return events, err
}

func (r *webhookRepository) DeleteOldEvents(ctx context.Context, olderThan time.Time) error {
	return r.db.WithContext(ctx).
		Where("created_at < ? AND status IN (?)", olderThan, []models.WebhookStatus{
			models.WebhookProcessed,
			models.WebhookSkipped,
		}).
		Delete(&models.WebhookEvent{}).Error
}

// WebhookLogRepository implementations
func (r *webhookLogRepository) SaveLog(ctx context.Context, log *models.WebhookLog) error {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}

	return r.db.WithContext(ctx).Create(log).Error
}

func (r *webhookLogRepository) GetLogsByEvent(ctx context.Context, eventID string) ([]*models.WebhookLog, error) {
	var logs []*models.WebhookLog
	err := r.db.WithContext(ctx).
		Where("event_id = ?", eventID).
		Order("created_at ASC").
		Find(&logs).Error
	return logs, err
}

func (r *webhookLogRepository) GetLogsByLevel(ctx context.Context, level models.LogLevel, limit int, offset int) ([]*models.WebhookLog, error) {
	var logs []*models.WebhookLog
	err := r.db.WithContext(ctx).
		Where("level = ?", level).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

func (r *webhookLogRepository) DeleteOldLogs(ctx context.Context, olderThan time.Time) error {
	return r.db.WithContext(ctx).
		Where("created_at < ?", olderThan).
		Delete(&models.WebhookLog{}).Error
}
