package tasks

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"kori/internal/config"
	"kori/internal/models"
	"kori/internal/utils"
	"kori/internal/utils/base64"
	"kori/internal/utils/logger"

	"gorm.io/gorm"

	"maps"

	"github.com/hibiken/asynq"
)

// TaskHandler handles task processing with improved error handling and logging
type TaskHandler struct {
	db             *gorm.DB
	logger         *logger.Logger
	mailHandler    *utils.EmailHandler
	taskClient     *TaskClient
	storageHandler *utils.StorageHandler
	cfg            *config.Config
	engine         interface{} // Will be set to *automation.Engine to avoid circular dependency
}

// NewTaskHandler creates a new TaskHandler
func NewTaskHandler(db *gorm.DB, cfg *config.Config) *TaskHandler {
	return &TaskHandler{
		db:             db,
		logger:         logger.New("task_handler"),
		mailHandler:    utils.NewEmailHandler(5), // Rate limit of 5 emails per second
		taskClient:     NewTaskClient(cfg.Redis.Addr, cfg.Redis.Username, cfg.Redis.Password, cfg.Redis.DB, cfg.Redis.UseTLS),
		storageHandler: utils.NewStorageHandler(),
		cfg:            cfg,
	}
}

// SetAutomationEngine sets the automation engine (called from main.go)
func (h *TaskHandler) SetAutomationEngine(engine interface{}) {
	h.engine = engine
}

// HandleEmailSend processes an email sending task
func (h *TaskHandler) HandleEmailSend(ctx context.Context, t *asynq.Task) error {
	var task EmailTask
	if err := json.Unmarshal(t.Payload(), &task); err != nil {
		return fmt.Errorf("failed to unmarshal email task: %w", asynq.SkipRetry)
	}

	// Get email from db
	email, err := models.GetEmailByID(task.EmailID, h.db)
	if err != nil {
		return h.logger.Error("❌ failed to get email: %w", err)
	}

	// Suppression is checked at delivery time, including messages prepared before opt-out.
	if email.Status == models.EmailStatusSent || email.Status == models.EmailStatusOpened || email.Status == models.EmailStatusClicked || email.Status == "SUPPRESSED" || (email.Status == "QUEUED" && email.SMTPConfig != nil && email.SMTPConfig.Provider == "MANAGED") || email.Status == "DELIVERY_UNKNOWN" || email.Status == "DELIVERED" || email.Status == "BOUNCED" || email.Status == "COMPLAINED" {
		return nil
	}
	if email.ContactID != "" {
		var contact models.Contact
		if err := h.db.Where("id = ? AND team_id = ? AND is_deleted = ?", email.ContactID, email.TeamID, false).First(&contact).Error; err != nil {
			return err
		}
		if contact.Status != models.SubscriberStatusActive {
			return h.db.Model(email).UpdateColumn("status", "SUPPRESSED").Error
		}
	}
	h.logger.Info("📧 Processing email task ID: %s (Attempt: %d)", task.EmailID, task.AttemptNum)

	// Claim outbox deliveries atomically. Ambiguous deliveries require review.
	if email.DeliveryKey != nil && (email.SMTPConfig == nil || email.SMTPConfig.Provider != "MANAGED") {
		result := h.db.Model(&models.Email{}).Where("id = ? AND status IN ?", email.ID, []string{"PENDING", "FAILED"}).UpdateColumn("status", "SENDING")
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
	}
	// Send email using SMTP handler
	if err := h.mailHandler.SendEmail(email); err != nil {
		task.Error = err.Error()
		task.AttemptNum++
		return h.logger.Error("❌ failed to send email: %w", err)
	}

	h.logger.Success("✅ Email sent successfully")

	// Check for after function in payload
	var payload map[string]interface{}
	if err := json.Unmarshal(t.Payload(), &payload); err == nil {
		if fn, ok := payload["after_func"].(func(context.Context, *asynq.Task) error); ok {
			if err := fn(ctx, t); err != nil {
				return fmt.Errorf("after func failed: %w", err)
			}
		}
	}

	return nil
}

// HandleCampaignProcess processes a campaign task
func (h *TaskHandler) HandleCampaignProcess(ctx context.Context, t *asynq.Task) error {
	var task CampaignTask
	if err := json.Unmarshal(t.Payload(), &task); err != nil {
		return fmt.Errorf("failed to unmarshal campaign task: %w", asynq.SkipRetry)
	}

	h.logger.Info("📧 Processing campaign task ID: %s with batch size %d and offset %d",
		task.CampaignID, task.BatchSize, task.Offset)

	// Get campaign details
	campaign, err := models.GetCampaignByID(task.CampaignID, h.db)
	if err != nil {
		return h.logger.Error("❌ failed to get campaign: %w", err)
	}

	if campaign.Status == models.CampaignStatusCompleted {
		h.logger.Info("✅ Campaign %s is already completed", task.CampaignID)
		return nil
	}

	// update campaign status to sending
	campaign.Status = models.CampaignStatusSending
	if err := h.db.Save(campaign).Error; err != nil {
		return h.logger.Error("❌ failed to update campaign status: %w", err)
	}

	// Get the campaign's SMTP config
	smtpConfig, err := models.GetSMTPConfig(campaign.TeamID, campaign.SMTPConfigID, "", h.db)
	if err != nil {
		return h.logger.Error("❌ failed to get smtp config: %w", err)
	}

	if smtpConfig == nil {
		return h.logger.Error("❌ failed to get smtp config: %w", fmt.Errorf("smtp config is nil"))
	}

	// Get the campaign's email list
	emailList, contactCount, err := models.GetEmailListByID(campaign.ListID, h.db)
	if err != nil {
		return h.logger.Error("❌ failed to get email list: %w", err)
	}

	// already processed contacts
	alreadyProcessedEmails, err := models.GetEmailsByCampaignID(campaign.ID, h.db)
	if err != nil {
		return h.logger.Error("❌ failed to get already processed emails: %w", err)
	}

	alreadyProcessedContacts := make([]string, len(alreadyProcessedEmails))
	for i, email := range alreadyProcessedEmails {
		alreadyProcessedContacts[i] = email.ContactID
	}

	// Get contacts for this batch, ordered by last email sent date
	var contacts []models.Contact
	query := h.db.Table("contacts").
		Select("contacts.*").
		Joins("LEFT JOIN emails ON emails.contact_id = contacts.id AND emails.campaign_id = ?", campaign.ID).
		Where("contacts.list_id = ? AND contacts.status = ?", emailList.ID, models.SubscriberStatusActive).
		Where("contacts.status != ?", models.SubscriberStatusUnsubscribed)

	if len(alreadyProcessedContacts) > 0 {
		query = query.Where("contacts.id NOT IN (?)", alreadyProcessedContacts)
	}

	if err := query.Group("contacts.id").
		Order("MAX(emails.sent_at) ASC NULLS FIRST"). // Contacts with no emails come first
		Offset(campaign.Processed).
		Limit(contactCount).
		Find(&contacts).Error; err != nil {
		return h.logger.Error("❌ failed to get contacts: %w", err)
	}

	// If no contacts in this batch, we're done
	if len(contacts) == 0 {
		h.logger.Info("✅ No more contacts to process for campaign %s", campaign.ID)
		campaign.Status = models.CampaignStatusCompleted
		if err := h.db.Save(campaign).Error; err != nil {
			return h.logger.Error("❌ failed to update campaign status: %w", err)
		}
		return nil
	}

	// Get HTML content outside transaction since it's an external operation
	htmlFromTemplate := campaign.Template.HTMLBody
	if htmlFromTemplate == "" {
		if campaign.Template.HtmlFile == nil {
			return fmt.Errorf("template has no HTML")
		}
		htmlFromTemplate, err = utils.GetHTMLFromURL(campaign.Template.HtmlFile.SignedURL)
	}
	if err != nil {
		return h.logger.Error("❌ failed to get html from template: %w", err)
	}

	// Create emails for each contact
	emails := make([]*models.Email, len(contacts))
	for i, contact := range contacts {
		// default variables
		defaultVariables := make(map[string]string)
		defaultVariables["email"] = contact.Email
		defaultVariables["first_name"] = contact.FirstName
		defaultVariables["last_name"] = contact.LastName
		fullName := fmt.Sprintf("%s %s", contact.FirstName, contact.LastName)
		defaultVariables["name"] = fullName
		defaultVariables["full_name"] = fullName
		defaultVariables["company"] = contact.Company
		defaultVariables["country"] = contact.Country
		defaultVariables["city"] = contact.City
		defaultVariables["state"] = contact.State
		defaultVariables["zip"] = contact.Zip
		defaultVariables["address"] = contact.Address
		defaultVariables["phone"] = contact.Phone
		defaultVariables["linkedin"] = contact.LinkedIn
		defaultVariables["twitter"] = contact.Twitter
		defaultVariables["facebook"] = contact.Facebook
		defaultVariables["instagram"] = contact.Instagram

		variables := make(map[string]string)
		maps.Copy(variables, defaultVariables)

		// read data from campaign.Data
		data, err := utils.JSONToMap(campaign.Data)
		if err != nil {
			return h.logger.Error("❌ failed to convert campaign data to map: %w", err)
		}
		maps.Copy(variables, data)

		messageID := uuid.NewString()
		parsedBody := utils.ReplaceVariables(htmlFromTemplate, variables, messageID, h.cfg, true, campaign.Template.Category.Name == "Marketing")
		parsedSubject := utils.ReplaceVariables(campaign.Template.Subject, variables, campaign.ID, h.cfg, false, false)

		parsedSubject, err = base64.DecodeFromBase64(parsedSubject)
		if err != nil {
			return h.logger.Error("❌ failed to decode subject: %w", err)
		}

		jsonData, err := utils.MapToJSON(variables)

		if err != nil {
			return h.logger.Error("❌ failed to convert variables to json: %w", err)
		}

		tracked := true
		email := &models.Email{
			Base: models.Base{ID: messageID}, ClickTrackingEnabled: &tracked,
			From:         smtpConfig.FromEmail,
			To:           contact.Email,
			Subject:      parsedSubject,
			Body:         parsedBody,
			Data:         jsonData,
			Status:       models.EmailStatusPending,
			TeamID:       campaign.TeamID,
			TemplateID:   campaign.TemplateID,
			ContactID:    contact.ID,
			SMTPConfigID: smtpConfig.ID,
			CategoryID:   campaign.Template.CategoryID,
			CampaignID:   campaign.ID,
		}
		emails[i] = email
	}

	// Save all emails in a transaction
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		for _, email := range emails {
			if err := tx.Create(email).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return h.logger.Error("❌ failed to create emails: %w", err)
	}

	h.mailHandler.SendCampaignEmails(emails, task.BatchSize, campaign.BatchDelay, smtpConfig)

	campaign.Processed += contactCount
	campaign.Status = models.CampaignStatusCompleted
	if err := h.db.Save(campaign).Error; err != nil {
		return h.logger.Error("❌ failed to update campaign processed: %w", err)
	}

	h.logger.Success("✅ Successfully processed campaign batch")

	// Check for after function in payload
	var payload map[string]any
	if err := json.Unmarshal(t.Payload(), &payload); err == nil {
		if fn, ok := payload["after_func"].(func(context.Context, *asynq.Task) error); ok {
			if err := fn(ctx, t); err != nil {
				return fmt.Errorf("after func failed: %w", err)
			}
		}
	}

	return nil
}

// HandleWebhookDelivery processes a webhook delivery task
func (h *TaskHandler) HandleWebhookDelivery(ctx context.Context, t *asynq.Task) error {
	var task WebhookDeliveryTask
	if err := json.Unmarshal(t.Payload(), &task); err != nil {
		return fmt.Errorf("failed to unmarshal webhook task: %w", asynq.SkipRetry)
	}

	h.logger.Info("processing webhook task %s with event %s and attempt %d", task.WebhookID, task.Event, task.AttemptNum)

	// TODO: Implement webhook delivery logic
	return nil
}

// HandleDomainVerification processes a domain verification task
func (h *TaskHandler) HandleDomainVerification(ctx context.Context, t *asynq.Task) error {
	var task DomainVerificationTask
	if err := json.Unmarshal(t.Payload(), &task); err != nil {
		return fmt.Errorf("failed to unmarshal domain verification task: %w", asynq.SkipRetry)
	}

	h.logger.Info("processing domain verification task %s", task.DomainID)

	// TODO: Implement domain verification logic
	return nil
}

// HandleContactImport processes a contact import task
func (h *TaskHandler) HandleContactImport(ctx context.Context, t *asynq.Task) error {
	h.logger.Info("🚀 Starting contact import task")

	var task ContactImportTask
	if err := json.Unmarshal(t.Payload(), &task); err != nil {
		return h.logger.Error("❌ failed to unmarshal contact import task: %w", asynq.SkipRetry)
	}

	h.logger.Info("📥 Processing contact import task %s", task.ImportID)

	// get the contact import from the database
	contact_import, err := models.GetContactImportByID(task.ImportID, h.db)
	if err != nil {
		return h.logger.Error("❌ failed to get contact import: %w", err)
	}
	h.logger.Info("📋 Found contact import record: %v", contact_import)

	// get the file from the database
	file, err := models.GetFileByID(contact_import.FileID, h.db)
	if err != nil {
		return h.logger.Error("❌ failed to get file: %w", err)
	}
	h.logger.Info("📁 Found file record: %v", file)

	// we need to download the file from the storage provider
	h.logger.Info("⬇️ Downloading file from: %s", file.SignedURL)
	fileContent, err := h.storageHandler.DownloadFile(file.SignedURL)
	if err != nil {
		contact_import.Status = models.ContactImportStatusFailed
		if err := h.db.Save(contact_import).Error; err != nil {
			return h.logger.Error("❌ failed to update contact import status: %w", err)
		}
		return h.logger.Error("❌ failed to download file: %w", err)
	}
	h.logger.Info("✅ File downloaded successfully")

	// now we need to get the file content
	fileData := bytes.NewReader(fileContent)

	// we need to parse the file content
	reader := csv.NewReader(fileData)
	reader.Comma = ','
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	// now we need to parse the file content
	records, err := reader.ReadAll()
	if err != nil {
		contact_import.Status = models.ContactImportStatusFailed
		if err := h.db.Save(contact_import).Error; err != nil {
			return h.logger.Error("❌ failed to update contact import status: %w", err)
		}
		return h.logger.Error("❌ failed to read file: %w", err)
	}

	// Skip header row
	if len(records) <= 1 {
		contact_import.Status = models.ContactImportStatusFailed
		if err := h.db.Save(contact_import).Error; err != nil {
			return h.logger.Error("❌ failed to update contact import status: %w", err)
		}
		return h.logger.Error("❌ file has no data rows", fmt.Errorf("no data rows in file"))
	}

	h.logger.Info("📋 Records: %v", records)

	fieldsMap, err := utils.JSONToMap(contact_import.FieldsMap)
	if err != nil {
		contact_import.Status = models.ContactImportStatusFailed
		if err := h.db.Save(contact_import).Error; err != nil {
			return h.logger.Error("❌ failed to update contact import status: %w", err)
		}
		return h.logger.Error("❌ failed to parse fields map from json: %w", err)
	}

	h.logger.Info("📋 Fields map: %v", fieldsMap)

	// contacts to create
	var contacts []models.Contact

	// Process each record (skipping header)
	for _, record := range records[1:] {
		// Create a map of field values from the record
		recordMap := make(map[string]string)
		headers := records[0]

		for j, value := range record {
			if j < len(headers) {
				recordMap[headers[j]] = value
			}
		}

		// Map fields according to the fieldsMap
		contact := models.Contact{
			TeamID:   contact_import.TeamID,
			ListID:   contact_import.ListID,
			ImportID: contact_import.ID,
		}

		// Create reverse mapping (database fields to CSV fields)
		var reverseMap = make(map[string]string)
		for k, v := range fieldsMap {
			reverseMap[v] = k
		}

		// Helper function to get the CSV field name from database field
		getFieldValue := func(dbField string) string {
			// Check if the database field exists in reverse mapping
			if csvField, exists := reverseMap[dbField]; exists {
				return recordMap[csvField]
			}
			return ""
		}

		// Map all fields using the helper function
		contact.Email = getFieldValue("email")
		contact.FirstName = getFieldValue("first_name")
		contact.LastName = getFieldValue("last_name")
		contact.LinkedIn = getFieldValue("linkedin")
		contact.Twitter = getFieldValue("twitter")
		contact.Facebook = getFieldValue("facebook")
		contact.Instagram = getFieldValue("instagram")
		contact.Country = getFieldValue("country")
		contact.Phone = getFieldValue("phone")
		contact.City = getFieldValue("city")
		contact.State = getFieldValue("state")
		contact.Zip = getFieldValue("zip")
		contact.Address = getFieldValue("address")
		contact.Company = getFieldValue("company")

		// Store all fields in metadata
		contact.Metadata, err = utils.MapToJSON(recordMap)
		if err != nil {
			return h.logger.Error("❌ failed to convert record map to json: %w", err)
		}

		contacts = append(contacts, contact)

		h.logger.Info("processing record %v", contact)
	}

	// save the contacts
	if err := h.db.CreateInBatches(&contacts, 100).Error; err != nil {
		contact_import.Status = models.ContactImportStatusFailed
		if err := h.db.Save(contact_import).Error; err != nil {
			return h.logger.Error("❌ failed to update contact import status: %w", err)
		}
		return h.logger.Error("❌ failed to create contacts: %w", err)
	}

	contact_import.Status = models.ContactImportStatusCompleted
	if err := h.db.Save(contact_import).Error; err != nil {
		return h.logger.Error("❌ failed to update contact import status: %w", err)
	}

	return nil
}

// HandleLLMEmailWriter processes an LLM email writer task
func (h *TaskHandler) HandleLLMEmailWriter(ctx context.Context, t *asynq.Task) error {
	var task LLMEmailWriterTask
	if err := json.Unmarshal(t.Payload(), &task); err != nil {
		return fmt.Errorf("failed to unmarshal LLM email writer task: %w", asynq.SkipRetry)
	}

	h.logger.Info("processing LLM email writer task %s with template %s and model %s and attempt %d", task.EmailID, task.TemplateID, task.ModelID, task.AttemptNum)

	// TODO: Implement LLM email generation logic
	return nil
}

// HandleAutomationExecute processes an automation execution task
func (h *TaskHandler) HandleAutomationExecute(ctx context.Context, t *asynq.Task) error {
	var task AutomationExecuteTask
	if err := json.Unmarshal(t.Payload(), &task); err != nil {
		return fmt.Errorf("failed to unmarshal automation execute task: %w", asynq.SkipRetry)
	}

	h.logger.Info("🤖 Processing automation execution task for automation %s, contact %s", task.AutomationID, task.ContactID)

	// Execute the automation using the engine
	// The engine is set via SetAutomationEngine from main.go
	if h.engine != nil {
		// Type assert to get the Execute method
		type AutomationEngine interface {
			Execute(ctx context.Context, automationID, contactID string, triggerData map[string]interface{}, currentNodeID string) error
		}

		if engine, ok := h.engine.(AutomationEngine); ok {
			if task.TriggerData == nil {
				task.TriggerData = map[string]interface{}{}
			}
			delete(task.TriggerData, "_execution_id")
			if task.ExecutionID != "" {
				task.TriggerData["_execution_id"] = task.ExecutionID
			}
			if err := engine.Execute(ctx, task.AutomationID, task.ContactID, task.TriggerData, task.CurrentNodeID); err != nil {
				h.logger.Error("❌ Automation execution failed: %v", err)
				return fmt.Errorf("automation execution failed: %w", err)
			}
			h.logger.Success("✅ Automation execution completed successfully")
			return nil
		}
	}

	h.logger.Warn("⚠️  Automation engine not initialized, skipping execution")
	return nil
}
