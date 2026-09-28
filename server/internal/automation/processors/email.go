package processors

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm/clause"
	"html"
	"kori/internal/automation"
	"kori/internal/config"
	"kori/internal/marketing"
	"kori/internal/models"
	"kori/internal/tasks"
	"kori/internal/utils"
	"kori/internal/utils/base64"
	"maps"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"
)

// EmailProcessor handles EMAIL nodes
type EmailProcessor struct {
	db         *gorm.DB
	taskClient *tasks.TaskClient
	cfg        *config.Config
}

// EmailNodeData represents the data structure for EMAIL nodes
type EmailNodeData struct {
	TemplateID    string            `json:"templateId"`
	SMTPConfigID  string            `json:"smtpConfigId"`
	PostalAddress string            `json:"postalAddress,omitempty"`
	Subject       string            `json:"subject,omitempty"`    // Override template subject
	CustomBody    string            `json:"customBody,omitempty"` // Override template body
	Variables     map[string]string `json:"variables,omitempty"`  // Additional variables
}

// NewEmailProcessor creates a new EMAIL node processor
func NewEmailProcessor(db *gorm.DB, taskClient *tasks.TaskClient, cfg *config.Config) *EmailProcessor {
	return &EmailProcessor{
		db:         db,
		taskClient: taskClient,
		cfg:        cfg,
	}
}

// Type returns the node type this processor handles
func (p *EmailProcessor) Type() models.NodeType {
	return models.NodeTypeEmail
}

// Validate checks if the EMAIL node data is valid
func (p *EmailProcessor) Validate(node *models.AutomationNode) error {
	if node.Type != models.NodeTypeEmail {
		return fmt.Errorf("invalid node type for EmailProcessor")
	}

	var data EmailNodeData
	if err := json.Unmarshal(node.Data, &data); err != nil {
		return fmt.Errorf("invalid email node data: %w", err)
	}

	if data.TemplateID == "" && data.CustomBody == "" {
		return fmt.Errorf("either templateId or customBody must be specified")
	}

	if data.SMTPConfigID == "" {
		return fmt.Errorf("smtpConfigId is required")
	}

	return nil
}

// Process executes the EMAIL node logic
func (p *EmailProcessor) Process(ctx *automation.ExecutionContext, node *models.AutomationNode) (*automation.ProcessResult, error) {
	var data EmailNodeData
	if err := json.Unmarshal(node.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse email node data: %w", err)
	}

	if ctx.Contact == nil || ctx.Contact.Status != models.SubscriberStatusActive {
		return &automation.ProcessResult{Complete: true, Message: "Contact is suppressed"}, nil
	}

	// Get SMTP config
	smtpConfig, err := models.GetSMTPConfig(ctx.TeamID, data.SMTPConfigID, "", p.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get SMTP config: %w", err)
	}

	// Get template if specified
	var template *models.Template
	var htmlBody string
	var subject string

	if data.TemplateID != "" {
		if err := p.db.Preload("HtmlFile").Preload("Category").Where("id = ? AND team_id = ? AND is_deleted = ?", data.TemplateID, ctx.TeamID, false).First(&template).Error; err != nil {
			return nil, fmt.Errorf("failed to load template: %w", err)
		}

		// Get HTML content from template
		if template.HTMLBody != "" {
			htmlBody = template.HTMLBody
		} else if template.HtmlFile != nil {
			htmlBody, err = utils.GetHTMLFromURL(template.HtmlFile.SignedURL)
		} else {
			return nil, fmt.Errorf("template has no HTML")
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get HTML from template: %w", err)
		}

		subject = template.Subject
	} else {
		htmlBody = data.CustomBody
		subject = data.Subject
	}

	if data.Subject != "" {
		subject = data.Subject
	}

	// Build variables for email personalization
	variables := make(map[string]string)

	// Default contact variables
	if ctx.Contact != nil {
		variables["email"] = ctx.Contact.Email
		variables["first_name"] = ctx.Contact.FirstName
		variables["last_name"] = ctx.Contact.LastName
		variables["name"] = fmt.Sprintf("%s %s", ctx.Contact.FirstName, ctx.Contact.LastName)
		variables["full_name"] = fmt.Sprintf("%s %s", ctx.Contact.FirstName, ctx.Contact.LastName)
		variables["company"] = ctx.Contact.Company
		variables["country"] = ctx.Contact.Country
		variables["city"] = ctx.Contact.City
		variables["state"] = ctx.Contact.State
		variables["zip"] = ctx.Contact.Zip
		variables["address"] = ctx.Contact.Address
		variables["phone"] = ctx.Contact.Phone
	}

	// Add custom variables from node data
	if data.Variables != nil {
		maps.Copy(variables, data.Variables)
	}

	// Add variables from execution context
	for key, val := range ctx.Variables {
		if strVal, ok := val.(string); ok {
			variables[key] = strVal
		}
	}

	// One immutable email per execution step, even after queue retries.
	if ctx.Execution == nil || ctx.Execution.ID == "" {
		return nil, fmt.Errorf("execution identity is required")
	}
	deliveryKey := "automation:" + ctx.Execution.ID + ":" + node.ID
	emailID := uuid.NewString()
	bodyVariables := make(map[string]string, len(variables))
	for k, v := range variables {
		bodyVariables[k] = html.EscapeString(v)
	}
	// Replace variables in body and subject
	isMarketing := template != nil && template.Category != nil && template.Category.Name == "Marketing"
	unsub := ""
	if isMarketing {
		service := marketing.New(p.db, p.cfg.JWT.Secret, os.Getenv("PUBLIC_API_URL"))
		if len(service.Secret) < 32 || !strings.HasPrefix(service.PublicURL, "https://") || len(strings.TrimSpace(data.PostalAddress)) < 8 {
			return nil, fmt.Errorf("marketing email requires HTTPS public URL, signing secret and sender postal address")
		}
		unsub = service.PublicURL + "/public/unsubscribe/" + ctx.TeamID + "/" + ctx.ContactID + "/" + service.Token(ctx.TeamID, ctx.ContactID)
		htmlBody += `<footer style="padding:24px;text-align:center;font:12px Arial">` + html.EscapeString(data.PostalAddress) + `<br><a href="` + unsub + `">Unsubscribe</a></footer>`
	}
	parsedBody := utils.ReplaceVariables(htmlBody, bodyVariables, emailID, p.cfg, true, false)
	parsedSubject := utils.ReplaceVariables(subject, variables, emailID, p.cfg, false, false)

	parsedSubject, err = base64.DecodeFromBase64(parsedSubject)
	if err != nil {
		return nil, err
	}

	// Serialize variables for email data
	jsonData, err := utils.MapToJSON(variables)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize variables: %w", err)
	}

	// Create email record
	email := &models.Email{
		Base:           models.Base{ID: emailID},
		DeliveryKey:    &deliveryKey,
		UnsubscribeURL: unsub,
		From:           smtpConfig.FromEmail,
		To:             ctx.Contact.Email,
		Subject:        parsedSubject,
		Body:           parsedBody,
		Data:           jsonData,
		Status:         models.EmailStatusPending,
		TeamID:         ctx.TeamID,
		TemplateID:     data.TemplateID,
		ContactID:      ctx.ContactID,
		SMTPConfigID:   smtpConfig.ID,
		CategoryID:     "",
	}

	if template != nil {
		email.CategoryID = template.CategoryID
	}

	// The database outbox dispatches committed emails and survives Redis outages.
	if err := p.db.Omit(clause.Associations).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "delivery_key"}}, DoNothing: true}).Create(email).Error; err != nil {
		return nil, fmt.Errorf("failed to persist email: %w", err)
	}
	email = &models.Email{}
	if err := p.db.Where("delivery_key = ?", deliveryKey).First(email).Error; err != nil {
		return nil, err
	}

	// Get next nodes
	var edges []models.AutomationNodeEdge
	if err := p.db.Where("automation_id = ? AND source_id = ?", ctx.AutomationID, node.ID).Find(&edges).Error; err != nil {
		return nil, fmt.Errorf("failed to load edges: %w", err)
	}

	nextNodeIDs := make([]string, len(edges))
	for i, edge := range edges {
		nextNodeIDs[i] = edge.TargetID
	}

	// Update variables with email info
	updateVars := map[string]interface{}{
		"last_email_id":      email.ID,
		"last_email_subject": email.Subject,
		"last_email_sent_at": time.Now().Format(time.RFC3339),
	}

	return &automation.ProcessResult{
		NextNodeIDs: nextNodeIDs,
		Message:     fmt.Sprintf("Email queued for sending to %s", ctx.Contact.Email),
		UpdateVars:  updateVars,
		Data: map[string]interface{}{
			"emailId": email.ID,
		},
	}, nil
}
