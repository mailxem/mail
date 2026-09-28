package services

import (
	"context"
	"errors"
	"fmt"
	"kori/internal/config"
	"kori/internal/db"
	"kori/internal/events"
	"kori/internal/models"
	"kori/internal/tasks"
	"kori/internal/utils"
	"kori/internal/utils/base64"
	"kori/internal/utils/logger"
	"os"
	"time"

	"github.com/google/uuid"
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
	testMail       bool
	sendAt         time.Time
	isWelcomeEmail bool
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

	definedID := uuid.New()

	// Get SMTP config
	smtpConfig, err := models.GetSMTPConfig(handler.teamId, handler.SMTPProvider, "", tx)
	if err != nil {
		tx.Rollback()
		return log.Error("failed to get team smtp config ❌", err)
	}

	// Get template
	template := &models.Template{}
	if handler.templateId != "" {
		if err := tx.Where("id = ? AND team_id = ?", handler.templateId, handler.teamId).Preload("HtmlFile").First(template).Error; err != nil {
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

	htmlFromTemplate := handler.body

	if handler.body == "" && template.HtmlFile != nil {
		// Get HTML content outside transaction since it's an external operation
		htmlFromTemplate, err = utils.GetHTMLFromURL(template.HtmlFile.SignedURL)
		if err != nil {
			tx.Rollback()
			return log.Error("failed to get html from template ❌", err)
		}
	} else if handler.body == "" && template.HtmlFile == nil {
		tx.Rollback()
		return log.Error("failed to get html from template ❌", errors.New("body is empty"))
	}

	parsedBody := utils.ReplaceVariables(htmlFromTemplate, handler.variables, definedID.String(), cfg, true, category.Name == "Marketing")
	parsedSubject := handler.subject
	if handler.subject == "" {
		parsedSubject = utils.ReplaceVariables(template.Subject, handler.variables, definedID.String(), cfg, false, false)
		parsedSubject, err = base64.DecodeFromBase64(parsedSubject)
		if err != nil {
			tx.Rollback()
			return log.Error("failed to decode subject ❌", err)
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
		SendAt:       handler.sendAt,
	}

	email.ID = definedID.String()
	if err := tx.Create(email).Error; err != nil {
		tx.Rollback()
		return log.Error("failed to create email ❌", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return log.Error("failed to commit transaction", err)
	}

	return nil // ✅
}
