package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"kori/internal/models"
	"strings"
	"time"

	"gorm.io/gorm"
)

// FormService handles form operations
type FormService struct {
	db *gorm.DB
}

// NewFormService creates a new form service
func NewFormService(db *gorm.DB) *FormService {
	return &FormService{
		db: db,
	}
}

// CreateForm creates a new form
func (s *FormService) CreateForm(ctx context.Context, form *models.Form) error {
	// Generate unique slug if not provided
	if form.Slug == "" {
		form.Slug = s.generateSlug(form.Name)
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create form
		if err := tx.Create(form).Error; err != nil {
			return err
		}

		// Create default fields if none provided
		if len(form.Fields) == 0 {
			defaultFields := models.DefaultFormFields()
			for i, template := range defaultFields {
				field := &models.FormField{
					FormID:            form.ID,
					StepNumber:        1,
					FieldType:         template.FieldType,
					Label:             template.Label,
					Required:          template.Required,
					DisplayOrder:      i + 1,
					MapToContactField: &template.MapTo,
				}
				if err := tx.Create(field).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// GetForm retrieves a form by ID
func (s *FormService) GetForm(ctx context.Context, id string) (*models.Form, error) {
	var form models.Form
	err := s.db.WithContext(ctx).
		Preload("Fields", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Preload("AddToList").
		Preload("AddToSegment").
		Preload("TriggerAutomation").
		Where("id = ?", id).
		First(&form).Error

	return &form, err
}

// GetFormBySlug retrieves a form by slug
func (s *FormService) GetFormBySlug(ctx context.Context, slug string) (*models.Form, error) {
	var form models.Form
	err := s.db.WithContext(ctx).
		Preload("Fields", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Where("slug = ? AND status = ?", slug, models.FormStatusPublished).
		First(&form).Error

	return &form, err
}

// ListForms retrieves all forms for a team
func (s *FormService) ListForms(ctx context.Context, teamID string, status models.FormStatus, page, limit int) ([]models.Form, int64, error) {
	var forms []models.Form
	var total int64

	query := s.db.WithContext(ctx).
		Where("team_id = ?", teamID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Model(&models.Form{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&forms).Error

	return forms, total, err
}

// UpdateForm updates a form
func (s *FormService) UpdateForm(ctx context.Context, id string, updates map[string]interface{}) error {
	return s.db.WithContext(ctx).
		Model(&models.Form{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// DeleteForm deletes a form
func (s *FormService) DeleteForm(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete submissions
		if err := tx.Where("form_id = ?", id).Delete(&models.FormSubmission{}).Error; err != nil {
			return err
		}

		// Delete fields
		if err := tx.Where("form_id = ?", id).Delete(&models.FormField{}).Error; err != nil {
			return err
		}

		// Delete form
		return tx.Where("id = ?", id).Delete(&models.Form{}).Error
	})
}

// PublishForm publishes a form
func (s *FormService) PublishForm(ctx context.Context, id string) error {
	now := time.Now()
	return s.db.WithContext(ctx).
		Model(&models.Form{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       models.FormStatusPublished,
			"published_at": now,
		}).Error
}

// UnpublishForm unpublishes a form
func (s *FormService) UnpublishForm(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).
		Model(&models.Form{}).
		Where("id = ?", id).
		Update("status", models.FormStatusDraft).Error
}

// RecordView increments the view count for a form
func (s *FormService) RecordView(ctx context.Context, formID string) error {
	return s.db.WithContext(ctx).
		Model(&models.Form{}).
		Where("id = ?", formID).
		UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}

// SubmitForm processes a form submission
func (s *FormService) SubmitForm(ctx context.Context, submission *models.FormSubmission) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get form
		var form models.Form
		if err := tx.Preload("Fields").Where("id = ?", submission.FormID).First(&form).Error; err != nil {
			return err
		}

		// Validate submission data
		if err := s.validateSubmission(&form, submission); err != nil {
			return err
		}

		// Extract email from field data
		var fieldData map[string]interface{}
		if err := json.Unmarshal(submission.FieldData, &fieldData); err != nil {
			return err
		}

		email, _ := fieldData["email"].(string)
		if email == "" {
			return fmt.Errorf("email is required")
		}
		submission.EmailAddress = strings.ToLower(email)

		// Generate confirmation token if double opt-in required
		if form.DoubleOptIn {
			submission.RequiresConfirmation = true
			token, err := generateToken()
			if err != nil {
				return err
			}
			submission.ConfirmationToken = &token
		}

		// Save submission
		if err := tx.Create(submission).Error; err != nil {
			return err
		}

		// Increment submission count
		if err := tx.Model(&models.Form{}).
			Where("id = ?", form.ID).
			UpdateColumn("submission_count", gorm.Expr("submission_count + ?", 1)).Error; err != nil {
			return err
		}

		// Update conversion rate
		if err := tx.Model(&models.Form{}).Where("id = ?", form.ID).
			UpdateColumn("conversion_rate", gorm.Expr("CAST(submission_count AS FLOAT) / NULLIF(view_count, 0)")).Error; err != nil {
			return err
		}

		// Process submission if not requiring confirmation
		if !form.DoubleOptIn {
			if err := s.processSubmissionData(ctx, tx, &form, submission); err != nil {
				return err
			}
		}

		return nil
	})
}

// ConfirmSubmission confirms a double opt-in submission
func (s *FormService) ConfirmSubmission(ctx context.Context, token string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var submission models.FormSubmission
		if err := tx.Where("confirmation_token = ?", token).First(&submission).Error; err != nil {
			return err
		}

		if submission.ConfirmedAt != nil {
			return fmt.Errorf("submission already confirmed")
		}

		// Mark as confirmed
		now := time.Now()
		submission.ConfirmedAt = &now

		if err := tx.Save(&submission).Error; err != nil {
			return err
		}

		// Process submission data
		var form models.Form
		if err := tx.Preload("Fields").Where("id = ?", submission.FormID).First(&form).Error; err != nil {
			return err
		}

		return s.processSubmissionData(ctx, tx, &form, &submission)
	})
}

// processSubmissionData creates/updates contact and triggers automations
func (s *FormService) processSubmissionData(ctx context.Context, tx *gorm.DB, form *models.Form, submission *models.FormSubmission) error {
	// Parse field data
	var fieldData map[string]interface{}
	if err := json.Unmarshal(submission.FieldData, &fieldData); err != nil {
		return err
	}

	// Find or create contact
	contact, err := s.findOrCreateContact(ctx, tx, form.TeamID, fieldData)
	if err != nil {
		return err
	}

	submission.ContactID = &contact.ID

	// Add to list if configured
	if form.AddToListID != nil {
		// Add contact to list logic here
	}

	// Add to segment if configured
	if form.AddToSegmentID != nil {
		// Add contact to segment logic here
	}

	// Trigger automation if configured
	if form.TriggerAutomationID != nil {
		// Trigger automation logic here
		// This would enqueue an automation execution task
	}

	// Mark as processed
	now := time.Now()
	submission.ProcessedAt = &now

	return tx.Save(submission).Error
}

// findOrCreateContact finds existing contact or creates new one
func (s *FormService) findOrCreateContact(ctx context.Context, tx *gorm.DB, teamID string, fieldData map[string]interface{}) (*models.Contact, error) {
	email, _ := fieldData["email"].(string)
	email = strings.ToLower(email)

	// Try to find existing contact
	var contact models.Contact
	err := tx.Where("team_id = ? AND email = ?", teamID, email).First(&contact).Error

	if err == gorm.ErrRecordNotFound {
		// Create new contact
		contact = models.Contact{
			TeamID: teamID,
			Email:  email,
			Status: models.SubscriberStatusActive,
		}

		// Map fields
		if firstName, ok := fieldData["first_name"].(string); ok {
			contact.FirstName = firstName
		}
		if lastName, ok := fieldData["last_name"].(string); ok {
			contact.LastName = lastName
		}
		if phone, ok := fieldData["phone"].(string); ok {
			contact.Phone = phone
		}
		if company, ok := fieldData["company"].(string); ok {
			contact.Company = company
		}

		// Store other fields in metadata
		metadataJSON, _ := json.Marshal(fieldData)
		contact.Metadata = metadataJSON

		if err := tx.Create(&contact).Error; err != nil {
			return nil, err
		}

		return &contact, nil
	}

	if err != nil {
		return nil, err
	}

	// Update existing contact with new data
	if firstName, ok := fieldData["first_name"].(string); ok && firstName != "" {
		contact.FirstName = firstName
	}
	if lastName, ok := fieldData["last_name"].(string); ok && lastName != "" {
		contact.LastName = lastName
	}

	// Merge metadata
	var existingMeta map[string]interface{}
	if contact.Metadata != nil {
		json.Unmarshal(contact.Metadata, &existingMeta)
	} else {
		existingMeta = make(map[string]interface{})
	}

	for k, v := range fieldData {
		existingMeta[k] = v
	}

	metadataJSON, _ := json.Marshal(existingMeta)
	contact.Metadata = metadataJSON

	if err := tx.Save(&contact).Error; err != nil {
		return nil, err
	}

	return &contact, nil
}

// validateSubmission validates submission data against form fields
func (s *FormService) validateSubmission(form *models.Form, submission *models.FormSubmission) error {
	var fieldData map[string]interface{}
	if err := json.Unmarshal(submission.FieldData, &fieldData); err != nil {
		return err
	}

	for _, field := range form.Fields {
		if !field.Required {
			continue
		}

		// Check if required field is present
		key := field.Label
		if field.MapToContactField != nil {
			key = *field.MapToContactField
		}

		value, exists := fieldData[key]
		if !exists || value == nil || value == "" {
			return fmt.Errorf("required field '%s' is missing", field.Label)
		}
	}

	return nil
}

// GetSubmissions retrieves submissions for a form
func (s *FormService) GetSubmissions(ctx context.Context, formID string, limit int, offset int) ([]models.FormSubmission, int64, error) {
	var submissions []models.FormSubmission
	var total int64

	query := s.db.WithContext(ctx).
		Where("form_id = ?", formID).
		Preload("Contact")

	if err := query.Model(&models.FormSubmission{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("submitted_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&submissions).Error

	return submissions, total, err
}

// generateSlug generates a URL-friendly slug
func (s *FormService) generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove special characters (simplified)
	return slug
}

// generateToken generates a random token
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
