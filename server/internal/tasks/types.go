package tasks

import (
	"context"
	"time"
)

type AutomationEnqueuer interface {
	EnqueueAutomationTask(context.Context, AutomationExecuteTask, time.Duration) error
}

// Task Types
const (
	// Email related tasks
	TaskTypeEmailSend  = "email:send"
	TaskTypeEmailRetry = "email:retry"

	// Campaign related tasks
	TaskTypeCampaignProcess  = "campaign:process"
	TaskTypeCampaignSchedule = "campaign:schedule"

	// Contact related tasks
	TaskTypeContactImport = "contact:import"
	TaskTypeContactSync   = "contact:sync"

	// Webhook related tasks
	TaskTypeWebhookDelivery = "webhook:delivery"
	TaskTypeWebhookRetry    = "webhook:retry"

	// Domain related tasks
	TaskTypeDomainVerification = "domain:verify"
	TaskTypeDomainCheck        = "domain:check"

	// LLM related tasks
	TaskTypeLLMEmailWriter = "llm:email_writer"

	// Automation related tasks
	TaskTypeAutomationExecute = "automation:execute"
	TaskTypeAutomationNode    = "automation:node"

	// Queue related tasks
	TaskTypeQueueConfig = "queue:config"
)

// Task Queues
const (
	QueueCritical = "critical" // For time-sensitive tasks like email sending
	QueueDefault  = "default"  // For regular tasks
	QueueLow      = "low"      // For background tasks like cleanup
)

// Task Priorities (1-10, higher is more important)
const (
	PriorityCritical = 10
	PriorityHigh     = 8
	PriorityNormal   = 5
	PriorityLow      = 3
	PriorityBG       = 1
)

// Task Timeouts
const (
	TimeoutShort  = 1 * time.Minute
	TimeoutMedium = 5 * time.Minute
	TimeoutLong   = 30 * time.Minute
)

// Task Retry Settings
const (
	RetryMax     = 5
	RetryDefault = 3
	RetryMin     = 1
)

// Task Payloads
type EmailTask struct {
	EmailID      string    `json:"email_id"`
	AttemptNum   int       `json:"attempt_num"`
	LastAttempt  time.Time `json:"last_attempt,omitempty"`
	Error        string    `json:"error,omitempty"`
	SMTPConfigID string    `json:"smtp_config_id"`
	MaxSendRate  int       `json:"max_send_rate"`
	SendAt       time.Time `json:"send_at,omitempty"`
}

type CampaignTask struct {
	CampaignID     string                 `json:"campaign_id"`
	BatchSize      int                    `json:"batch_size"`
	Offset         int                    `json:"offset"`
	Parameters     map[string]interface{} `json:"parameters,omitempty"`
	ScheduledAt    time.Time              `json:"scheduled_at,omitempty"`
	CronExpression string                 `json:"cron_expression,omitempty"`
}

type WebhookDeliveryTask struct {
	WebhookID   string                 `json:"webhook_id"`
	Event       string                 `json:"event"`
	Payload     map[string]interface{} `json:"payload"`
	AttemptNum  int                    `json:"attempt_num"`
	LastAttempt time.Time              `json:"last_attempt,omitempty"`
}

type DomainVerificationTask struct {
	DomainID    string    `json:"domain_id"`
	LastChecked time.Time `json:"last_checked,omitempty"`
	DNSRecord   string    `json:"dns_record,omitempty"`
}

type ContactImportTask struct {
	ImportID string `json:"import_id"`
}

type LLMEmailWriterTask struct {
	EmailID     string                 `json:"email_id"`
	TemplateID  string                 `json:"template_id"`
	UserID      string                 `json:"user_id"`
	TeamID      string                 `json:"team_id"`
	ModelID     string                 `json:"model_id"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	AttemptNum  int                    `json:"attempt_num"`
	LastAttempt time.Time              `json:"last_attempt,omitempty"`
}

type AutomationExecuteTask struct {
	ExecutionID   string                 `json:"execution_id"`
	AutomationID  string                 `json:"automation_id"`
	ContactID     string                 `json:"contact_id"`
	TriggerData   map[string]interface{} `json:"trigger_data"`
	CurrentNodeID string                 `json:"current_node_id"` // For resuming from WAIT
}
