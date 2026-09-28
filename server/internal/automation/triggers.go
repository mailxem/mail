package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"kori/internal/events"
	"kori/internal/models"
	"kori/internal/tasks"
	"kori/internal/utils/logger"

	"gorm.io/gorm"
)

var triggerLog = logger.New("AUTOMATION_TRIGGERS")

// TriggerManager manages event-based automation triggers
type TriggerManager struct {
	db         *gorm.DB
	taskClient *tasks.TaskClient
	eventBus   *events.EventBus
}

// TriggerConfig represents trigger configuration stored in Automation.Data
type TriggerConfig struct {
	Type   string                 `json:"type"`   // "manual", "event", "scheduled"
	Events []string               `json:"events"` // For event-based: ["contact.created", "email.opened"]
	Cron   string                 `json:"cron"`   // For scheduled: cron expression
	Filter map[string]interface{} `json:"filter"` // Optional filter conditions
}

// NewTriggerManager creates a new trigger manager
func NewTriggerManager(db *gorm.DB, taskClient *tasks.TaskClient) *TriggerManager {
	return &TriggerManager{
		db:         db,
		taskClient: taskClient,
		eventBus:   events.NewEventBus(),
	}
}

// Initialize sets up event listeners for all active automations
func (tm *TriggerManager) Initialize(ctx context.Context) error {
	triggerLog.Info("Initializing automation triggers...")

	// Register global event listeners
	tm.registerEventListeners()

	// Load all active automations and set up their triggers
	var automations []models.Automation
	if err := tm.db.Where("is_active = ? AND is_deleted = ?", true, false).Find(&automations).Error; err != nil {
		return fmt.Errorf("failed to load active automations: %w", err)
	}

	triggerLog.Info("Found %d active automations", len(automations))
	return nil
}

// registerEventListeners sets up global event listeners
func (tm *TriggerManager) registerEventListeners() {
	// Contact events
	events.On("contact.created", func(data interface{}) {
		if contact, ok := data.(*models.Contact); ok {
			tm.handleContactEvent("contact.created", contact)
		}
	})

	// Email events
	events.On("email.sent", func(data interface{}) {
		if email, ok := data.(*models.Email); ok {
			tm.handleEmailEvent("email.sent", email)
		}
	})

	events.On("email.opened", func(data interface{}) {
		if tracking, ok := data.(*models.EmailTracking); ok {
			tm.handleTrackingEvent("email.opened", tracking)
		}
	})

	events.On("email.clicked", func(data interface{}) {
		if tracking, ok := data.(*models.EmailTracking); ok {
			tm.handleTrackingEvent("email.clicked", tracking)
		}
	})

	events.On("email.bounced", func(data interface{}) {
		if tracking, ok := data.(*models.EmailTracking); ok {
			tm.handleTrackingEvent("email.bounced", tracking)
		}
	})

	// Campaign events
	events.On("campaign.completed", func(data interface{}) {
		if campaign, ok := data.(*models.Campaign); ok {
			tm.handleCampaignEvent("campaign.completed", campaign)
		}
	})

	// List events
	events.On("list.subscriber_added", func(data interface{}) {
		if subscriber, ok := data.(*models.Contact); ok {
			tm.handleContactEvent("list.subscriber_added", subscriber)
		}
	})

	triggerLog.Success("Registered automation event listeners")
}

// handleContactEvent processes contact-related events
func (tm *TriggerManager) handleContactEvent(eventName string, contact *models.Contact) {
	ctx := context.Background()

	// Find automations triggered by this event
	var automations []models.Automation
	if err := tm.db.Where("team_id = ? AND is_active = ? AND is_deleted = ?",
		contact.TeamID, true, false).Find(&automations).Error; err != nil {
		triggerLog.Error("Failed to load automations: %v", err)
		return
	}

	for _, automation := range automations {
		if automation.TriggerEvent != eventName {
			continue
		}
		// Parse trigger config from automation data/metadata
		// For now, we'll check if automation should be triggered
		// In production, you'd parse trigger config from Automation.Data field

		triggerLog.Info("Triggering automation %s for contact %s on event %s",
			automation.ID, contact.ID, eventName)

		// Enqueue automation execution
		task := tasks.AutomationExecuteTask{
			AutomationID: automation.ID,
			ContactID:    contact.ID,
			TriggerData: map[string]interface{}{
				"event":     eventName,
				"timestamp": fmt.Sprintf("%v", contact.CreatedAt),
			},
		}

		if err := tm.taskClient.EnqueueAutomationTask(ctx, task, 0); err != nil {
			triggerLog.Error("Failed to enqueue automation task: %v", err)
		}
	}
}

// handleEmailEvent processes email-related events
func (tm *TriggerManager) handleEmailEvent(eventName string, email *models.Email) {
	ctx := context.Background()

	// Find automations triggered by this event
	var automations []models.Automation
	if err := tm.db.Where("team_id = ? AND is_active = ? AND is_deleted = ?",
		email.TeamID, true, false).Find(&automations).Error; err != nil {
		triggerLog.Error("Failed to load automations: %v", err)
		return
	}

	for _, automation := range automations {
		if automation.TriggerEvent != eventName {
			continue
		}
		if email.ContactID == "" {
			continue // Skip if no contact associated
		}

		triggerLog.Info("Triggering automation %s for contact %s on event %s",
			automation.ID, email.ContactID, eventName)

		task := tasks.AutomationExecuteTask{
			AutomationID: automation.ID,
			ContactID:    email.ContactID,
			TriggerData: map[string]interface{}{
				"event":   eventName,
				"emailId": email.ID,
			},
		}

		if err := tm.taskClient.EnqueueAutomationTask(ctx, task, 0); err != nil {
			triggerLog.Error("Failed to enqueue automation task: %v", err)
		}
	}
}

// handleTrackingEvent processes email tracking events
func (tm *TriggerManager) handleTrackingEvent(eventName string, tracking *models.EmailTracking) {
	ctx := context.Background()

	if tracking.ContactID == "" {
		return
	}

	// Get contact to find team
	var contact models.Contact
	if err := tm.db.Where("id = ?", tracking.ContactID).First(&contact).Error; err != nil {
		triggerLog.Error("Failed to load contact: %v", err)
		return
	}

	// Find automations triggered by this event
	var automations []models.Automation
	if err := tm.db.Where("team_id = ? AND is_active = ? AND is_deleted = ?",
		contact.TeamID, true, false).Find(&automations).Error; err != nil {
		triggerLog.Error("Failed to load automations: %v", err)
		return
	}

	for _, automation := range automations {
		if automation.TriggerEvent != eventName {
			continue
		}
		triggerLog.Info("Triggering automation %s for contact %s on event %s",
			automation.ID, tracking.ContactID, eventName)

		task := tasks.AutomationExecuteTask{
			AutomationID: automation.ID,
			ContactID:    tracking.ContactID,
			TriggerData: map[string]interface{}{
				"event":     eventName,
				"emailId":   tracking.EmailID,
				"url":       tracking.URL,
				"timestamp": tracking.Timestamp,
			},
		}

		if err := tm.taskClient.EnqueueAutomationTask(ctx, task, 0); err != nil {
			triggerLog.Error("Failed to enqueue automation task: %v", err)
		}
	}
}

// handleCampaignEvent processes campaign-related events
func (tm *TriggerManager) handleCampaignEvent(eventName string, campaign *models.Campaign) {
	ctx := context.Background()

	// Find automations triggered by this event
	var automations []models.Automation
	if err := tm.db.Where("team_id = ? AND is_active = ? AND is_deleted = ?",
		campaign.TeamID, true, false).Find(&automations).Error; err != nil {
		triggerLog.Error("Failed to load automations: %v", err)
		return
	}

	// Get all contacts from the campaign's list
	var contacts []models.Contact
	if err := tm.db.Where("list_id = ?", campaign.ListID).Find(&contacts).Error; err != nil {
		triggerLog.Error("Failed to load contacts: %v", err)
		return
	}

	for _, automation := range automations {
		if automation.TriggerEvent != eventName {
			continue
		}
		for _, contact := range contacts {
			task := tasks.AutomationExecuteTask{
				AutomationID: automation.ID,
				ContactID:    contact.ID,
				TriggerData: map[string]interface{}{
					"event":      eventName,
					"campaignId": campaign.ID,
				},
			}

			if err := tm.taskClient.EnqueueAutomationTask(ctx, task, 0); err != nil {
				triggerLog.Error("Failed to enqueue automation task: %v", err)
			}
		}
	}
}

// ParseTriggerConfig parses trigger configuration from automation data
func ParseTriggerConfig(data []byte) (*TriggerConfig, error) {
	var config TriggerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
