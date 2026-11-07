package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"kori/internal/config"
	"kori/internal/db"
	"kori/internal/events"
	"kori/internal/mailconnect"
	"kori/internal/models"
	"kori/internal/tasks"
	"kori/internal/utils"
	"kori/internal/utils/base64"
	"kori/internal/utils/logger"
	"os"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	log        = logger.New("EMAIL")
	cfg        *config.Config
	taskClient *tasks.TaskClient
)

// Initialize sets up email services after Infisical and config are loaded in main.
func Initialize(appCfg *config.Config) {
	cfg = appCfg
	taskClient = tasks.NewTaskClient(
		cfg.Redis.Addr,
		cfg.Redis.Username,
		cfg.Redis.Password,
		cfg.Redis.DB,
		cfg.Redis.UseTLS,
	)

	registerEmailEventHandlers()
}

type sendEmailHandlerBody struct {
	teamId         string
	templateId     string
	to             string
	SMTPProvider   string
	categoryId     string
	variables      map[string]string
	subject        string
	listId         string
	campaignId     string
	body           string
	cc             string
	bcc            string
	replyTo        string
	inReplyTo      string
	attachments    []models.MailAttachment
	outboxID       uuid.UUID
	requestHash    string
	durable        bool
	testMail       bool
	sendAt         time.Time
	isWelcomeEmail bool
}

// QueueAPIEmail persists an outbox record before acknowledging the request.
// The existing periodic dispatcher retries enqueue when Redis is unavailable.
func QueueAPIEmail(email *models.Email) error {
	return QueueAPIEmailWithKey(email, "")
}

var ErrEmailRequestConflict = errors.New("this request key was already used for a different message")
var ErrEmailSenderUnavailable = errors.New("selected sender is unavailable or inactive")
var ErrEmailTemplateUnavailable = errors.New("template is unavailable")

// Retries of an identical request return its original durable outbox ID, even
// after delivery. A changed payload must use a new key. The unique email ID is
// the final concurrency boundary, so two simultaneous requests cannot enqueue
// two messages.
func QueueAPIEmailWithKey(email *models.Email, key string) error {
	return QueueAPIEmailRequest(email, key, "")
}

// The receipt uses the requested sender selection, before resolving a default
// or provider. A retry still finds it after a disconnect or a default change.
func QueueAPIEmailRequest(email *models.Email, key, provider string) error {
	if err := models.ValidateMailAttachments(email.Attachments); err != nil {
		return err
	}
	variables := map[string]string{}
	if email.Data != nil {
		var err error
		variables, err = utils.JSONToMap(email.Data)
		if err != nil {
			return err
		}
	}
	handler := &sendEmailHandlerBody{teamId: email.TeamID, templateId: email.TemplateID, to: email.To, SMTPProvider: email.SMTPConfigID, categoryId: email.CategoryID, variables: variables, subject: email.Subject, body: email.Body, cc: email.CC, bcc: email.BCC, replyTo: email.ReplyTo, inReplyTo: email.InReplyTo, attachments: email.Attachments, testMail: email.Test, sendAt: email.SendAt, durable: true, outboxID: uuid.New()}
	if key != "" {
		handler.outboxID = uuid.NewSHA1(uuid.NameSpaceOID, []byte("xem:outbox:"+email.TeamID+":"+key))
		// Only user-controlled, immutable request fields belong in this digest.
		payload, err := json.Marshal(struct {
			Template, To, Sender, Provider, Subject, Body, CC, BCC, ReplyTo, InReplyTo string
			Variables                                                                  map[string]string
			Attachments                                                                []models.MailAttachment
			Test                                                                       bool
			SendAt                                                                     time.Time
		}{email.TemplateID, email.To, email.SMTPConfigID, provider, email.Subject, email.Body, email.CC, email.BCC, email.ReplyTo, email.InReplyTo, variables, email.Attachments, email.Test, email.SendAt})
		if err != nil {
			return err
		}
		hash := sha256.Sum256(payload)
		handler.requestHash = hex.EncodeToString(hash[:])
		if found, err := existingAPIEmail(db.DB, handler.outboxID, handler.requestHash); found || err != nil {
			email.ID = handler.outboxID.String()
			return err
		}
	}
	sender, err := models.GetSMTPConfig(email.TeamID, email.SMTPConfigID, provider, db.DB)
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && !sender.IsActive) {
		return ErrEmailSenderUnavailable
	}
	if err != nil {
		return err
	}
	handler.SMTPProvider = sender.ID
	if email.TemplateID != "" {
		var count int64
		if err := db.DB.Model(&models.Template{}).Where("id = ? AND team_id = ? AND is_deleted = false", email.TemplateID, email.TeamID).Count(&count).Error; err != nil {
			return err
		}
		if count != 1 {
			return ErrEmailTemplateUnavailable
		}
	}
	if err := sendEmail(handler); err != nil {
		return err
	}
	email.ID = handler.outboxID.String()
	return nil
}

func existingAPIEmail(tx *gorm.DB, id uuid.UUID, hash string) (bool, error) {
	var row models.Email
	err := tx.Select("id", "request_hash").Where("id = ?", id.String()).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if row.RequestHash != hash {
		return true, ErrEmailRequestConflict
	}
	return true, nil
}

func registerEmailEventHandlers() {
	events.On("invite.created", func(data interface{}) {
		invite := data.(*models.TeamInvite)
		log.Info("Sending invite email to %s", invite.Email)
		if err := sendTeamInviteEmail(invite); err != nil {
			err := log.Error("Failed to send invite email: %v", err)
			if err != nil {
				return
			}
		}
	})

	events.On("email.created", func(data interface{}) {
		email := data.(*models.Email)
		log.Info("Enqueueing email task for %s", email.ID)

		task := tasks.EmailTask{
			EmailID:      email.ID,
			AttemptNum:   1,
			SMTPConfigID: email.SMTPConfigID,
			MaxSendRate:  email.SMTPConfig.MaxSendRate,
			SendAt:       email.SendAt,
		}

		log.Info("Enqueueing email task for %s", task.EmailID)

		if err := taskClient.EnqueueEmailTask(context.Background(), task); err != nil {
			err := log.Error("Failed to enqueue email task: %v", err)
			if err != nil {
				return
			}
		}
	})

	events.On("campaign.created", func(data interface{}) {
		campaign := data.(*models.Campaign)
		log.Info("Enqueueing campaign task for %s", campaign.ID)

		if campaign.Schedule == models.CampaignScheduleRecurring {

			cron := ""

			if campaign.RecurringSchedule == models.CampaignRecurringScheduleDaily {
				cron = "0 0 * * *"
			} else if campaign.RecurringSchedule == models.CampaignRecurringScheduleWeekly {
				cron = "0 0 * * 1"
			} else if campaign.RecurringSchedule == models.CampaignRecurringScheduleMonthly {
				cron = "0 0 1 * *"
			} else if campaign.RecurringSchedule == models.CampaignRecurringScheduleCustom {
				cron = campaign.CronExpression
			} else {
				log.Error("%s", fmt.Errorf("invalid recurring schedule: %s", campaign.RecurringSchedule))
				return
			}

			if err := taskClient.EnqueueCampaignTask(context.Background(), tasks.CampaignTask{
				CampaignID:     campaign.ID,
				BatchSize:      campaign.BatchSize,
				CronExpression: cron,
				Offset:         0,
			}, 0); err != nil {
				log.Error("Failed to enqueue campaign task: %v", err)
			}
		} else {
			if err := taskClient.EnqueueCampaignTask(context.Background(), tasks.CampaignTask{
				CampaignID:  campaign.ID,
				BatchSize:   campaign.BatchSize,
				Offset:      0,
				ScheduledAt: campaign.ScheduledFor,
			}, 0); err != nil {
				log.Error("Failed to enqueue campaign task: %v", err)
			}
		}
	})

	events.On("users.created", func(data interface{}) {
		if os.Getenv("SERVICE_NOTIFICATIONS_ENABLED") == "true" {
			return
		}
		user := data.(*models.User)
		log.Info("Sending welcome email to %s", user.Email)
		if err := sendWelcomeEmail(user); err != nil {
			err := log.Error("Failed to send welcome email: %v", err)
			if err != nil {
				return
			}
		}
	})

	events.On("password.reset", func(data interface{}) {
		if os.Getenv("SERVICE_NOTIFICATIONS_ENABLED") == "true" {
			return
		}
		reset := data.(*models.PasswordReset)
		log.Info("Sending password reset email to %s", reset.User.Email)
		if err := sendPasswordResetEmail(reset); err != nil {
			err := log.Error("Failed to send password reset email: %v", err)
			if err != nil {
				return
			}
		}
	})

	events.On("email.send", func(data interface{}) {
		email := data.(*models.Email)
		var emailData map[string]string
		var err error
		if email.Data != nil {
			emailData, err = utils.JSONToMap(email.Data)
			if err != nil {
				log.Error("Failed to convert data to map: %v", err)
				return
			}
		}

		handler := &sendEmailHandlerBody{
			teamId:       email.TeamID,
			templateId:   email.TemplateID,
			to:           email.To,
			SMTPProvider: email.SMTPConfigID,
			categoryId:   email.CategoryID,
			variables:    emailData,
			subject:      email.Subject,
			listId:       "",
			campaignId:   email.CampaignID,
			body:         email.Body,
			cc:           email.CC,
			bcc:          email.BCC,
			replyTo:      email.ReplyTo,
			inReplyTo:    email.InReplyTo,
			testMail:     email.Test,
			sendAt:       email.SendAt,
		}

		if err := sendEmail(handler); err != nil {
			err := log.Error("Failed to send email: %v", err)
			if err != nil {
				return
			}
		}
	})
}

func sendPasswordResetEmail(reset *models.PasswordReset) error {
	tx := db.DB.Begin()
	team := &models.Team{}
	if err := tx.Where("name =?", os.Getenv("SUPERADMIN_TEAM_NAME")).First(team).Error; err != nil {
		tx.Rollback()
		return log.Error("failed to get team details", err)
	}

	// Get default SMTP config
	smtpConfig := &models.SMTPConfig{}
	if err := tx.Where("team_id = ? AND is_default = ?", team.ID, true).First(smtpConfig).Error; err != nil {
		tx.Rollback()
		return log.Error("failed to get default smtp config", err)
	}

	template := &models.Template{}
	if err := tx.Where("name = ? AND team_id = ?", "Password Reset", team.ID).First(template).Error; err != nil {
		tx.Rollback()
		return log.Error("failed to get template", err)
	}

	mailingList := &models.MailingList{}
	if err := tx.Where("name = ? AND team_id = ?", "All Users", team.ID).First(mailingList).Error; err != nil {
		tx.Rollback()
		return log.Error("failed to get mailing list", err)
	}

	handler := &sendEmailHandlerBody{
		teamId:       team.ID,
		templateId:   template.ID,
		to:           reset.User.Email,
		SMTPProvider: smtpConfig.ID,
		categoryId:   template.CategoryID,
		variables:    map[string]string{"name": reset.User.FirstName, "code": reset.Code, "url": fmt.Sprintf("%s/auth/reset-password/%s", os.Getenv("OFFICE_URL"), reset.Code)},
		subject:      fmt.Sprintf("Hey %s 👋🏻! We've received a request to reset your password on Posthoot", reset.User.FirstName),
		listId:       mailingList.ID,
		campaignId:   "",
		body:         "",
		cc:           "",
		bcc:          "",
		replyTo:      "",
		testMail:     false,
	}

	return sendEmail(handler)
}

func sendTeamInviteEmail(invite *models.TeamInvite) error {
	// Start transaction
	tx := db.DB.Begin()
	if tx.Error != nil {
		return log.Error("failed to begin transaction", tx.Error)
	}

	// Get team details with settings
	team := &models.Team{}
	if err := tx.Preload("Settings").First(team, "id = ?", invite.TeamID).Error; err != nil {
		tx.Rollback()
		return log.Error("failed to get team details", err)
	}

	// Get inviter details
	inviter := &models.User{}
	if err := tx.First(inviter, "id = ?", invite.InviterID).Error; err != nil {
		tx.Rollback()
		return log.Error("failed to get inviter details", err)
	}

	var template *models.Template

	inviteTemplateID := ""
	if len(team.Settings) > 0 {
		inviteTemplateID = team.Settings[0].InviteTemplateID
	}

	if inviteTemplateID == "" {
		// Get invite template
		template = &models.Template{}
		if err := tx.Where("name = ? AND team_id = ?", "PLATFORM INVITE", invite.TeamID).First(template).Error; err != nil {
			tx.Rollback()
			return log.Error("failed to get invite template", err)
		}
	} else {
		template = &models.Template{}
		if err := tx.Where("id = ? AND team_id = ?", inviteTemplateID, invite.TeamID).First(template).Error; err != nil {
			tx.Rollback()
			return log.Error("failed to get invite template", err)
		}
	}

	// Get invite mailing list
	mailingList := &models.MailingList{}
	if err := tx.Where("name = ? AND team_id = ?", "Invited Users", invite.TeamID).First(mailingList).Error; err != nil {
		tx.Rollback()
		return log.Error("failed to get invite mailing list", err)
	}

	// Get or create contact
	contact := &models.Contact{}
	if err := tx.Where("email = ? AND team_id = ? AND list_id = ?", invite.Email, invite.TeamID, mailingList.ID).First(contact).Error; err != nil {
		contact = &models.Contact{
			Email:     invite.Email,
			FirstName: invite.Name,
			TeamID:    invite.TeamID,
			ListID:    mailingList.ID,
		}
		if err := tx.Create(contact).Error; err != nil {
			tx.Rollback()
			return log.Error("failed to create contact", err)
		}
	}

	// get superadmin team
	superadminTeam := &models.Team{}
	if err := tx.Where("name = ?", os.Getenv("SUPERADMIN_TEAM_NAME")).First(superadminTeam).Error; err != nil {
		tx.Rollback()
		return log.Error("failed to get superadmin team", err)
	}

	// Get default SMTP config
	smtpConfig := &models.SMTPConfig{}
	if err := tx.Where("team_id = ?", superadminTeam.ID).First(smtpConfig).Error; err != nil {
		tx.Rollback()
		return log.Error("failed to get default smtp config", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return log.Error("failed to commit transaction", err)
	}

	handler := &sendEmailHandlerBody{
		teamId:       team.ID,
		templateId:   template.ID,
		to:           invite.Email,
		SMTPProvider: smtpConfig.ID,
		categoryId:   template.CategoryID,
		variables:    map[string]string{"inviter": fmt.Sprintf("%s %s", inviter.FirstName, inviter.LastName), "name": invite.Name, "role": string(invite.Role), "url": fmt.Sprintf("%s/auth/accept-invite/%s", os.Getenv("OFFICE_URL"), invite.Code)},
		subject:      "Hey {{ name }} 👋🏻! You've been invited to join a team on Posthoot",
		listId:       mailingList.ID,
		campaignId:   "",
		body:         "",
		cc:           "",
		bcc:          "",
		replyTo:      "",
	}

	// Send email outside transaction since it's an external operation
	return sendEmail(handler)
}

func sendWelcomeEmail(user *models.User) error {
	tx := db.DB.Begin()
	team := &models.Team{}
	if err := tx.Where("name =?", os.Getenv("SUPERADMIN_TEAM_NAME")).First(team).Preload("Settings").Preload("Settings.BrandingSettings").Error; err != nil {
		tx.Rollback()
		return log.Error("failed to get team details", err)
	}

	// Get default SMTP config
	smtpConfig := &models.SMTPConfig{}
	if err := tx.Where("team_id = ?", team.ID).First(smtpConfig).Error; err != nil {
		tx.Rollback()
		return log.Error("failed to get default smtp config", err)
	}

	// mailing list

	mailingList := &models.MailingList{}
	if err := tx.Where("team_id = ? AND name = ?", user.TeamID, "PLATFORM WELCOME").First(mailingList).Error; err != nil {
		tx.Rollback()
		return log.Error("failed to get mailing list", err)
	}

	welcomeTemplateID := ""
	if len(team.Settings) > 0 {
		welcomeTemplateID = team.Settings[0].WelcomeTemplateID
	}

	handler := &sendEmailHandlerBody{
		teamId:         user.TeamID,
		templateId:     welcomeTemplateID,
		to:             user.Email,
		SMTPProvider:   smtpConfig.ID,
		categoryId:     "",
		variables:      map[string]string{"name": user.FirstName},
		subject:        "Hey {{ name }} 👋🏻! We're glad to have you onboard 🎉",
		listId:         mailingList.ID,
		campaignId:     "",
		body:           "",
		cc:             "",
		bcc:            "",
		replyTo:        "",
		isWelcomeEmail: true,
	}

	err := sendEmail(handler)
	if err != nil {
		return err
	}
	return nil
}

func sendEmail(
	handler *sendEmailHandlerBody,
) error {
	// Start transaction
	tx := db.DB.Begin()
	if tx.Error != nil {
		return log.Error("failed to begin transaction", tx.Error)
	}

	definedID := handler.outboxID
	if definedID == uuid.Nil {
		definedID = uuid.New()
	}

	// Get SMTP config
	smtpConfig, err := models.GetSMTPConfig(handler.teamId, handler.SMTPProvider, "", tx)
	if err != nil {
		tx.Rollback()
		return log.Error("failed to get team smtp config ❌", err)
	}

	// Get template
	template := &models.Template{}
	if handler.templateId != "" {
		if err := tx.Where("id = ? AND team_id = ? AND is_deleted = false", handler.templateId, handler.teamId).Preload("HtmlFile").First(template).Error; err != nil {
			tx.Rollback()
			return log.Error("failed to get template ❌", err)
		}
	}

	// Get or use default mailing list
	if handler.listId == "" && !handler.testMail {
		mailingList := &models.MailingList{}
		if err := tx.Where("team_id = ? AND name = ?", handler.teamId, "All Users").First(mailingList).Error; err != nil {
			tx.Rollback()
			return log.Error("failed to get default mailing list ❌", err)
		}
		handler.listId = mailingList.ID
	}

	contact := &models.Contact{}

	if !handler.testMail {
		// Get or create contact
		if err := tx.Where("email = ? AND team_id = ? AND list_id = ?", handler.to, handler.teamId, handler.listId).First(contact).Error; err != nil {
			contactImport := &models.ContactImport{}
			contactImport.TeamID = handler.teamId
			contactImport.ListID = handler.listId
			contactImport.Status = models.ContactImportStatusCompleted
			if err := tx.Create(contactImport).Error; err != nil {
				tx.Rollback()
				return log.Error("failed to create contact import ❌", err)
			}
			contact = &models.Contact{
				Email:     handler.to,
				TeamID:    handler.teamId,
				ListID:    handler.listId,
				ImportID:  contactImport.ID,
				FirstName: handler.variables["name"],
			}
			if err := tx.Create(contact).Error; err != nil {
				tx.Rollback()
				return log.Error("failed to create contact ❌", err)
			}
		}
	}

	if handler.categoryId == "" {
		category := &models.EmailCategory{}
		if err := tx.Where("name = ? AND team_id = ?", "Transactional", handler.teamId).First(category).Error; err != nil {
			tx.Rollback()
			return log.Error("failed to get category ❌", err)
		}
		handler.categoryId = category.ID
	}
	// Get category
	category := &models.EmailCategory{}
	if err := tx.Where("id = ? AND team_id = ?", handler.categoryId, handler.teamId).First(category).Error; err != nil {
		tx.Rollback()
		return log.Error("failed to get category ❌", err)
	}
	if smtpConfig.Provider == mailconnect.Cloudflare && (handler.campaignId != "" || category.Name != "Transactional") {
		tx.Rollback()
		return errors.New("Cloudflare Email Sending only supports transactional email")
	}

	htmlFromTemplate := handler.body
	if htmlFromTemplate == "" {
		// Editor templates keep raw HTML in HTMLBody; older ones only have a file.
		htmlFromTemplate = template.HTMLBody
	}

	if htmlFromTemplate == "" && template.HtmlFile != nil {
		// Get HTML content outside transaction since it's an external operation
		htmlFromTemplate, err = utils.GetHTMLFromURL(template.HtmlFile.SignedURL)
		if err != nil {
			tx.Rollback()
			return log.Error("failed to get html from template ❌", err)
		}
	} else if htmlFromTemplate == "" {
		tx.Rollback()
		return log.Error("failed to get html from template ❌", errors.New("body is empty"))
	}

	parsedBody := utils.ReplaceVariables(htmlFromTemplate, handler.variables, definedID.String(), cfg, true, category.Name == "Marketing")
	parsedSubject := handler.subject
	if handler.subject == "" {
		parsedSubject = utils.ReplaceVariables(template.Subject, handler.variables, definedID.String(), cfg, false, false)
		// Editor templates store the subject raw; legacy ones store it base64.
		// ponytail: a raw subject that happens to be valid base64 gets decoded; store an encoding flag if that bites.
		if decoded, err := base64.DecodeFromBase64(parsedSubject); err == nil {
			parsedSubject = decoded
		}
	}

	jsonData, err := utils.MapToJSON(handler.variables)

	if err != nil {
		tx.Rollback()
		return log.Error("failed to convert variables to json ❌", err)
	}

	// Create email
	email := &models.Email{
		From:         smtpConfig.FromEmail,
		TeamID:       handler.teamId,
		TemplateID:   handler.templateId,
		To:           handler.to,
		Subject:      parsedSubject,
		Body:         parsedBody,
		Status:       models.EmailStatusPending,
		Data:         jsonData,
		ContactID:    contact.ID,
		CategoryID:   category.ID,
		SMTPConfigID: smtpConfig.ID,
		CampaignID:   handler.campaignId,
		CC:           handler.cc,
		BCC:          handler.bcc,
		ReplyTo:      handler.replyTo,
		InReplyTo:    handler.inReplyTo,
		Attachments:  handler.attachments,
		RequestHash:  handler.requestHash,
		Test:         handler.testMail,
		SendAt:       handler.sendAt,
	}

	email.ID = definedID.String()
	if handler.durable {
		key := "api:" + email.ID
		email.DeliveryKey = &key
	}
	if err := tx.Create(email).Error; err != nil {
		tx.Rollback()
		if handler.requestHash != "" {
			if found, lookupErr := existingAPIEmail(db.DB, definedID, handler.requestHash); found || lookupErr != nil {
				return lookupErr
			}
		}
		return log.Error("failed to create email ❌", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return log.Error("failed to commit transaction", err)
	}

	return nil // ✅
}

// EmailReader serves the emails API. Bulk emails store no body, so it renders
// their shared content: readers see the HTML that was sent.
type EmailReader struct {
	BaseService[models.Email]
	db *gorm.DB
}

func NewEmailReader(db *gorm.DB) EmailReader {
	return EmailReader{NewBaseService(db, models.Email{}), db}
}

func (s EmailReader) Get(ctx context.Context, id string, includes ...string) (*models.Email, error) {
	email, err := s.BaseService.Get(ctx, id, includes...)
	if err == nil {
		err = s.fillBody(ctx, email)
	}
	return email, err
}

func (s EmailReader) List(ctx context.Context, page, limit int, filters map[string]interface{}, excludes map[string]bool, sortFields []string, order string, includes ...string) ([]models.Email, int64, error) {
	emails, total, err := s.BaseService.List(ctx, page, limit, filters, excludes, sortFields, order, includes...)
	for i := 0; err == nil && !excludes["body"] && i < len(emails); i++ {
		err = s.fillBody(ctx, &emails[i])
	}
	return emails, total, err
}

func (s EmailReader) fillBody(ctx context.Context, email *models.Email) error {
	if email.Body != "" || email.ContentID == nil {
		return nil
	}
	html, err := utils.EmailHTML(s.db.WithContext(ctx), email)
	if err != nil {
		return err
	}
	email.Body = base64.EncodeToBase64(html)
	return nil
}
