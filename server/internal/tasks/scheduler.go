package tasks

import (
	"crypto/tls"
	"fmt"
	"time"

	"kori/internal/utils/logger"

	"github.com/hibiken/asynq"
)

// Scheduler handles periodic task scheduling
type Scheduler struct {
	scheduler *asynq.Scheduler
	logger    *logger.Logger
}

// NewScheduler creates a new task scheduler
func NewScheduler(redisAddr, username, password string, db int, logger *logger.Logger, useTLS ...bool) *Scheduler {
	opt := asynq.RedisClientOpt{
		Addr:     redisAddr,
		Username: username,
		Password: password,
		DB:       db,
	}
	if len(useTLS) > 0 && useTLS[0] {
		opt.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	scheduler := asynq.NewScheduler(opt,
		&asynq.SchedulerOpts{},
	)

	return &Scheduler{
		scheduler: scheduler,
		logger:    logger,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	if err := s.registerTasks(); err != nil {
		return fmt.Errorf("failed to register tasks: %w", err)
	}

	s.logger.Info("starting task scheduler")
	return s.scheduler.Run()
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.scheduler.Shutdown()
	s.logger.Info("task scheduler stopped")
}

// registerTasks registers all periodic tasks
func (s *Scheduler) registerTasks() error {
	if _, err := s.scheduler.Register("@every 15s", asynq.NewTask(TaskTypeFormCompletionTick, nil), asynq.Queue(QueueDefault), asynq.Unique(14*time.Second)); err != nil {
		return err
	}
	if _, err := s.scheduler.Register("@every 1m", asynq.NewTask(TaskTypeNewsletterTick, nil), asynq.Queue(QueueDefault), asynq.Unique(50*time.Second)); err != nil {
		return err
	}
	// // Campaign scheduling (every minute)
	// entryID, err := s.scheduler.Register("*/1 * * * *", asynq.NewTask(
	// 	TaskTypeCampaignSchedule,
	// 	nil,
	// 	asynq.Queue(QueueDefault),
	// 	asynq.MaxRetry(RetryDefault),
	// 	asynq.Timeout(TimeoutMedium),
	// ))
	// if err != nil {
	// 	return fmt.Errorf("failed to register campaign scheduler: %w", err)
	// }
	// s.logger.Debug("registered campaign scheduler %s", entryID)

	// // Email retry (every 5 minutes)
	// entryID, err = s.scheduler.Register("*/5 * * * *", asynq.NewTask(
	// 	TaskTypeEmailRetry,
	// 	nil,
	// 	asynq.Queue(QueueDefault),
	// 	asynq.MaxRetry(RetryDefault),
	// 	asynq.Timeout(TimeoutMedium),
	// ))
	// if err != nil {
	// 	return fmt.Errorf("failed to register email retry scheduler: %w", err)
	// }
	// s.logger.Debug("registered email retry scheduler %s", entryID)

	// // Webhook retry (every 5 minutes)
	// entryID, err = s.scheduler.Register("*/5 * * * *", asynq.NewTask(
	// 	TaskTypeWebhookRetry,
	// 	nil,
	// 	asynq.Queue(QueueDefault),
	// 	asynq.MaxRetry(RetryDefault),
	// 	asynq.Timeout(TimeoutMedium),
	// ))
	// if err != nil {
	// 	return fmt.Errorf("failed to register webhook retry scheduler: %w", err)
	// }
	// s.logger.Debug("registered webhook retry scheduler %s", entryID)

	// // Domain verification (daily at midnight)
	// entryID, err = s.scheduler.Register("0 0 * * *", asynq.NewTask(
	// 	TaskTypeDomainCheck,
	// 	nil,
	// 	asynq.Queue(QueueLow),
	// 	asynq.MaxRetry(RetryMin),
	// 	asynq.Timeout(TimeoutLong),
	// ))
	// if err != nil {
	// 	return fmt.Errorf("failed to register domain verification scheduler: %w", err)
	// }
	// s.logger.Debug("registered domain verification scheduler %s", entryID)

	// // Contact sync (every 15 minutes)
	// entryID, err = s.scheduler.Register("*/15 * * * *", asynq.NewTask(
	// 	TaskTypeContactSync,
	// 	nil,
	// 	asynq.Queue(QueueDefault),
	// 	asynq.MaxRetry(RetryDefault),
	// 	asynq.Timeout(TimeoutMedium),
	// ))
	// if err != nil {
	// 	return fmt.Errorf("failed to register contact sync scheduler: %w", err)
	// }
	// s.logger.Debug("registered contact sync scheduler %s", entryID)

	s.logger.Info("registered all periodic tasks")
	return nil
}

// RegisterCustomTask registers a custom periodic task
func (s *Scheduler) RegisterCustomTask(spec string, taskType string, payload []byte, opts ...asynq.Option) error {
	entryID, err := s.scheduler.Register(spec, asynq.NewTask(taskType, payload, opts...))
	if err != nil {
		return fmt.Errorf("failed to register custom task: %w", err)
	}

	s.logger.Info("registered custom task %s %s %s", taskType, spec, entryID)
	return nil
}
