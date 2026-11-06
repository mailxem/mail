package tasks

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"kori/internal/affiliates"
	"kori/internal/config"
	"kori/internal/hygiene"
	"kori/internal/marketing"
	"kori/internal/models"
	"kori/internal/utils"
	"kori/internal/utils/base64"
	"kori/internal/utils/logger"
	"kori/internal/workflowconfig"
	"os"
	"strconv"
	"time"

	"gorm.io/gorm"

	"maps"
	"strings"

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
	// PostmasterSync reads Gmail spam rates; set by main when managed sending is on.
	PostmasterSync func(context.Context) error
	triggers       AutomationTriggers
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
	if email.Status == "ACCEPTED" || email.Status == "PARTIAL" {
		return nil
	}
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
	if campaign.Status == models.CampaignStatusPaused {
		h.logger.Info("⏸️ Campaign %s is paused", task.CampaignID)
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

	// Staged sending: readers emailed before go out now; never-emailed
	// readers only as far as the current stage allows.
	limit, wait, err := marketing.Stage(h.db, campaign, smtpConfig.Provider == "MANAGED", time.Now())
	if err != nil {
		return h.logger.Error("❌ failed to plan staged sending: %w", err)
	}
	if campaign.Status == models.CampaignStatusPaused {
		h.logger.Info("⏸️ Paused campaign %s: %s", campaign.ID, campaign.PauseReason)
		return nil
	}
	nextStage := func(wait time.Duration) error {
		next := CampaignTask{CampaignID: campaign.ID, BatchSize: task.BatchSize, Stage: task.Stage + 1}
		// A retry re-queues the same task ID, so it never runs twice.
		if err := h.taskClient.EnqueueCampaignTask(ctx, next, wait); err != nil && !errors.Is(err, asynq.ErrTaskIDConflict) {
			return h.logger.Error("❌ failed to queue the next stage: %w", err)
		}
		h.logger.Info("⏳ Campaign %s: next stage in %s", campaign.ID, wait)
		return nil
	}

	// Fetched and stored on first use, so a list with no one left completes
	// without it. Every email shares it and keeps only its own variables.
	var contentID string
	badge := affiliates.Badge(ctx, h.db, campaign.TeamID)
	// Builds one contact's email; the pages below call it for each contact.
	build := func(contact models.Contact) (*models.Email, error) {
		if contentID == "" {
			htmlFromTemplate := campaign.Template.HTMLBody
			if htmlFromTemplate == "" {
				if campaign.Template.HtmlFile == nil {
					return nil, fmt.Errorf("template has no HTML")
				}
				html, err := utils.GetHTMLFromURL(campaign.Template.HtmlFile.SignedURL)
				if err != nil {
					return nil, h.logger.Error("❌ failed to get html from template: %w", err)
				}
				htmlFromTemplate = html
			}
			id, err := utils.StoreEmailContent(h.db, models.EmailContent{TeamID: campaign.TeamID, Kind: utils.ContentCampaign, HTML: htmlFromTemplate, Subject: campaign.Template.Subject, Badge: badge})
			if err != nil {
				return nil, h.logger.Error("❌ failed to store campaign content: %w", err)
			}
			contentID = id
		}

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
			return nil, h.logger.Error("❌ failed to convert campaign data to map: %w", err)
		}
		maps.Copy(variables, data)

		messageID := uuid.NewString()
		// A campaign goes to a list, so it's marketing: every email gets an
		// unsubscribe footer and the one-click header that managed sending,
		// Gmail and Yahoo require, whether or not it's tracked. The body is
		// rendered from the shared content when sent (utils.RenderEmail).
		tracked := models.ShouldTrack(h.db, campaign.TeamID, contact.Email, campaign.TrackEngagement, true)
		unsubscribe, err := utils.UnsubscribeLink(messageID, h.cfg)
		if err != nil {
			return nil, h.logger.Error("❌ failed to sign unsubscribe link: %w", err)
		}
		parsedSubject := utils.ReplaceVariables(campaign.Template.Subject, variables, campaign.ID, h.cfg, false, false)

		parsedSubject, err = base64.DecodeFromBase64(parsedSubject)
		if err != nil {
			return nil, h.logger.Error("❌ failed to decode subject: %w", err)
		}

		jsonData, err := utils.MapToJSON(variables)
		if err != nil {
			return nil, h.logger.Error("❌ failed to convert variables to json: %w", err)
		}

		return &models.Email{
			Base: models.Base{ID: messageID}, ClickTrackingEnabled: &tracked, UnsubscribeURL: unsubscribe,
			From:         smtpConfig.FromEmail,
			To:           contact.Email,
			Subject:      parsedSubject,
			ContentID:    &contentID,
			Data:         jsonData,
			Status:       models.EmailStatusPending,
			TeamID:       campaign.TeamID,
			TemplateID:   campaign.TemplateID,
			ContactID:    contact.ID,
			SMTPConfigID: smtpConfig.ID,
			CategoryID:   campaign.Template.CategoryID,
			CampaignID:   campaign.ID,
		}, nil
	}
	send := func(emails []*models.Email) error {
		h.mailHandler.SendCampaignEmails(emails, task.BatchSize, time.Duration(campaign.BatchDelay), smtpConfig)
		campaign.Processed += len(emails)
		return h.db.Model(campaign).UpdateColumn("processed", campaign.Processed).Error
	}

	// Readers emailed before go out now, a page at a time; never-emailed
	// readers only as far as the current stage allows, picked at random.
	old, fresh := marketing.Recipients(h.db, campaign)
	sent, err := sendCampaignPages(h.db, old, false, campaignPageSize(), -1, build, send)
	if err != nil {
		return h.logger.Error("❌ failed to send campaign emails: %w", err)
	}
	if limit != 0 {
		n, err := sendCampaignPages(h.db, fresh, limit > 0, campaignPageSize(), limit, build, send)
		sent += n
		if err != nil {
			return h.logger.Error("❌ failed to send campaign emails: %w", err)
		}
	}
	if sent == 0 && wait > 0 {
		return nextStage(wait)
	}

	// If there was no one left to email, we're done
	if sent == 0 {
		h.logger.Info("✅ No more contacts to process for campaign %s", campaign.ID)
		campaign.Status = models.CampaignStatusCompleted
		if err := h.db.Save(campaign).Error; err != nil {
			return h.logger.Error("❌ failed to update campaign status: %w", err)
		}
		return nil
	}

	// Starts the wait when this run finished a stage.
	if _, wait, err := marketing.Stage(h.db, campaign, smtpConfig.Provider == "MANAGED", time.Now()); err != nil || wait > 0 {
		if err != nil {
			return h.logger.Error("❌ failed to plan staged sending: %w", err)
		}
		return nextStage(wait)
	}
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

// campaignInsertBatch is how many campaign emails go in one INSERT.
const campaignInsertBatch = 500

// campaignPageSize is how many contacts a campaign loads at a time
// (CAMPAIGN_PAGE_SIZE; default 1000).
func campaignPageSize() int {
	if n, err := strconv.Atoi(os.Getenv("CAMPAIGN_PAGE_SIZE")); err == nil && n > 0 {
		return n
	}
	return 1000
}

// sendCampaignPages emails up to limit (or, when limit < 0, every) contact
// matched by eligible, one page at a time, so only one page is in memory.
// eligible must leave out contacts that already have an email for the
// campaign (marketing.Recipients does), so a retried task carries on where
// the last one stopped. random picks each page at random instead of in id
// order, for staged samples. It returns how many emails it created.
func sendCampaignPages(db *gorm.DB, eligible *gorm.DB, random bool, pageSize, limit int, build func(models.Contact) (*models.Email, error), send func([]*models.Email) error) (int, error) {
	done, after := 0, ""
	for limit < 0 || done < limit {
		size := pageSize
		if limit >= 0 {
			size = min(size, limit-done)
		}
		query := eligible
		if random {
			query = query.Order("RANDOM()")
		} else {
			if after != "" {
				query = query.Where("contacts.id > ?", after)
			}
			query = query.Order("contacts.id")
		}
		var contacts []models.Contact
		if err := query.Limit(size).Find(&contacts).Error; err != nil {
			return done, err
		}
		if len(contacts) == 0 {
			break
		}
		after = contacts[len(contacts)-1].ID

		emails := make([]*models.Email, len(contacts))
		for i, contact := range contacts {
			email, err := build(contact)
			if err != nil {
				return done, err
			}
			emails[i] = email
		}
		if err := db.CreateInBatches(emails, campaignInsertBatch).Error; err != nil {
			return done, err
		}
		done += len(emails)
		if err := send(emails); err != nil {
			return done, err
		}
		if len(contacts) < size {
			break
		}
	}
	return done, nil
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
// failImport marks an import failed and, when service notifications are on,
// records an alert for the workspace owner in the same transaction.
func (h *TaskHandler) failImport(ci *models.ContactImport, reason string) error {
	ci.Status = models.ContactImportStatusFailed
	return h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(ci).Error; err != nil {
			return err
		}
		if os.Getenv("SERVICE_NOTIFICATIONS_ENABLED") != "true" {
			return nil
		}
		return models.QueueAlert(tx, ci.TeamID, "import", ci.ID, "import_failed", reason, time.Now())
	})
}

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
		if err := h.failImport(contact_import, "The import file could not be downloaded"); err != nil {
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
		if err := h.failImport(contact_import, "The file could not be read as CSV"); err != nil {
			return h.logger.Error("❌ failed to update contact import status: %w", err)
		}
		return h.logger.Error("❌ failed to read file: %w", err)
	}

	// Skip header row
	if len(records) <= 1 {
		if err := h.failImport(contact_import, "The file has no data rows"); err != nil {
			return h.logger.Error("❌ failed to update contact import status: %w", err)
		}
		return h.logger.Error("❌ file has no data rows", fmt.Errorf("no data rows in file"))
	}

	h.logger.Info("📋 Records: %v", records)

	fieldsMap, err := utils.JSONToMap(contact_import.FieldsMap)
	if err != nil {
		if err := h.failImport(contact_import, "The column mapping could not be read"); err != nil {
			return h.logger.Error("❌ failed to update contact import status: %w", err)
		}
		return h.logger.Error("❌ failed to parse fields map from json: %w", err)
	}

	h.logger.Info("📋 Fields map: %v", fieldsMap)

	// contacts to create
	var contacts []models.Contact
	skipped := map[string]int{}
	checker := hygiene.Checker{Import: true}

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
		contact.Email = strings.TrimSpace(getFieldValue("email"))
		if contact.Email == "" {
			continue
		}
		if reason := checker.Reason(ctx, contact.Email); reason != "" {
			skipped[reason]++
			continue
		}
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
		// Unknown time zones are dropped rather than failing the row.
		if tz := strings.TrimSpace(getFieldValue("timezone")); workflowconfig.ValidTimezone(tz) {
			contact.Timezone = tz
		}

		// Store all fields in metadata
		contact.Metadata, err = utils.MapToJSON(recordMap)
		if err != nil {
			return h.logger.Error("❌ failed to convert record map to json: %w", err)
		}

		contacts = append(contacts, contact)

		h.logger.Info("processing record %v", contact)
	}

	emails := make([]string, len(contacts))
	for i, contact := range contacts {
		emails[i] = contact.Email
	}
	blocked, err := hygiene.Blocked(h.db, emails...)
	if err != nil {
		return h.logger.Error("❌ failed to check the never-email list: %w", err)
	}
	kept := contacts[:0]
	for _, contact := range contacts {
		if blocked[contact.Email] {
			skipped[hygiene.NeverEmail]++
			continue
		}
		kept = append(kept, contact)
	}
	contacts = kept
	if contact_import.Skipped, err = json.Marshal(skipped); err != nil {
		return h.logger.Error("❌ failed to encode skipped counts: %w", err)
	}

	// save the contacts
	if len(contacts) == 0 {
		// Every row was left out; nothing to save.
	} else if err := h.db.CreateInBatches(&contacts, 100).Error; err != nil {
		if err := h.failImport(contact_import, "The contacts could not be saved"); err != nil {
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
