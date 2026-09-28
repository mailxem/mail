package handlers

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/marketing"
	"kori/internal/models"
)

type formInput struct {
	Theme          *formTheme                `json:"theme"`
	Definition     *marketing.FormDefinition `json:"definition"`
	Name           string                    `json:"name"`
	Description    string                    `json:"description"`
	ListID         string                    `json:"listId"`
	Status         string                    `json:"status"`
	SuccessMessage string                    `json:"successMessage"`
	ButtonText     string                    `json:"buttonText"`
	Fields         []marketing.FormQuestion  `json:"fields"`
}

func (h *MarketingHandler) Forms(c echo.Context) error {
	rows := []models.Form{}
	if err := h.scoped(c).Preload("Fields", func(db *gorm.DB) *gorm.DB { return db.Order("display_order") }).Order("created_at DESC").Limit(200).Find(&rows).Error; err != nil {
		return err
	}
	for i := range rows {
		d, err := marketing.DefinitionForForm(rows[i])
		if err != nil {
			return err
		}
		rows[i].Definition, _ = json.Marshal(d)
	}
	return c.JSON(200, rows)
}

func (h *MarketingHandler) SaveForm(c echo.Context) error {
	var in formInput
	if err := c.Bind(&in); err != nil {
		return bad("Invalid form")
	}
	if strings.ContainsRune(in.Name+in.Description+in.SuccessMessage+in.ButtonText, 0) || len(strings.TrimSpace(in.Name)) < 2 || len(in.Name) > 120 || len(in.Description) > 500 || len(in.SuccessMessage) > 500 || len(in.ButtonText) > 60 {
		return bad("Provide a name and valid form text")
	}
	if in.Status != "DRAFT" && in.Status != "PUBLISHED" && in.Status != "ARCHIVED" {
		return bad("Invalid form status")
	}
	if uuid.Validate(in.ListID) != nil || team(c) == "" {
		return bad("Choose an audience in this workspace")
	}
	if in.Theme != nil {
		if err := in.Theme.validate(); err != nil {
			return err
		}
	}
	var result models.Form
	err := h.service.DB.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
		f := models.Form{TeamID: team(c), FormType: models.FormTypeInline, Slug: uuid.NewString(), Theme: datatypes.JSON(`{}`), Version: 0}
		if id := c.Param("id"); id != "" {
			if err := marketing.Scope(tx, team(c)).Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Fields", func(db *gorm.DB) *gorm.DB { return db.Order("display_order") }).First(&f, "id = ?", id).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return missing()
				}
				return err
			}
		}
		// Recheck routing under the same transaction and lock order as capture.
		var audience models.MailingList
		if err := marketing.Scope(tx, team(c)).Clauses(clause.Locking{Strength: "UPDATE"}).First(&audience, "id = ?", in.ListID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return bad("Choose an audience in this workspace")
			}
			return err
		}
		var d marketing.FormDefinition
		if in.Definition != nil {
			d = *in.Definition
		} else if len(f.Definition) > 0 && string(f.Definition) != "null" && string(f.Definition) != "{}" {
			if err := json.Unmarshal(f.Definition, &d); err != nil {
				return err
			}
			// Keep legacy field edits working, but never erase richer question or consent
			// settings merely because an older client does not know about definitions.
			if legacyEditableDefinition(d) {
				d.Pages[0].Fields = in.Fields
			}
		} else {
			d = marketing.FormDefinition{SchemaVersion: 1, Pages: []marketing.FormPage{{ID: "details", Title: "Your details", Fields: in.Fields}}, Consent: marketing.FormConsent{Mode: "required", Label: "I agree to receive emails and can unsubscribe at any time."}}
		}
		if err := marketing.ValidateFormDefinition(d); err != nil {
			return bad(err.Error())
		}
		if in.Theme != nil {
			f.Theme, _ = json.Marshal(in.Theme)
		}
		f.Name = strings.TrimSpace(in.Name)
		f.Description = in.Description
		f.AddToListID = &in.ListID
		f.Status = models.FormStatus(in.Status)
		f.SuccessMessage = in.SuccessMessage
		f.SubmitButtonText = in.ButtonText
		f.Definition, _ = json.Marshal(d)
		f.Version++
		f.IsMultiStep = len(d.Pages) > 1
		if f.Status == models.FormStatusPublished {
			now := time.Now().UTC()
			f.PublishedAt = &now
		}
		if err := tx.Omit(clause.Associations).Save(&f).Error; err != nil {
			return err
		}
		rev := models.FormRevision{ID: uuid.NewString(), FormID: f.ID, Version: f.Version, Definition: f.Definition}
		if err := tx.Create(&rev).Error; err != nil {
			return err
		}
		if err := tx.Where("form_id = ?", f.ID).Delete(&models.FormField{}).Error; err != nil {
			return err
		}
		fields := []models.FormField{}
		for step, p := range d.Pages {
			for _, q := range p.Fields {
				key := q.Key
				options, _ := json.Marshal(q.Options)
				if q.Options == nil {
					options = []byte(`[]`)
				}
				condition, _ := json.Marshal(q.Condition)
				fields = append(fields, models.FormField{FormID: f.ID, StepNumber: step + 1, FieldType: models.FieldType(q.Type), Label: q.Label, Required: q.Required, DisplayOrder: len(fields), MapToContactField: &key, Options: options, ConditionalDisplay: condition, Placeholder: &q.Placeholder, HelpText: &q.HelpText, DefaultValue: &q.DefaultValue, IsHidden: q.Type == "HIDDEN"})
			}
		}
		if err := tx.Create(&fields).Error; err != nil {
			return err
		}
		f.Fields = fields
		result = f
		return nil
	})
	if err != nil {
		return err
	}
	return c.JSON(200, result)
}

func (h *MarketingHandler) Submissions(c echo.Context) error {
	if !h.service.Owns(c.Request().Context(), &models.Form{}, c.Param("id"), team(c)) {
		return missing()
	}
	rows := []models.FormSubmission{}
	if err := h.service.DB.WithContext(c.Request().Context()).Where("form_id = ?", c.Param("id")).Select("id", "email_address", "field_data", "submitted_at", "contact_id", "version", "consent", "session_id", "utm_source", "utm_medium", "utm_campaign").Order("submitted_at DESC").Limit(200).Find(&rows).Error; err != nil {
		return err
	}
	return c.JSON(200, rows)
}
func (h *MarketingHandler) publicForm(c echo.Context, tx *gorm.DB, lock bool) (models.Form, error) {
	var f models.Form
	q := tx.WithContext(c.Request().Context()).Preload("Fields", func(db *gorm.DB) *gorm.DB { return db.Order("display_order") })
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := q.Where("slug = ? AND status = ? AND is_deleted = ?", c.Param("slug"), models.FormStatusPublished, false).First(&f).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return f, missing()
	}
	return f, err
}
func (h *MarketingHandler) PublicForm(c echo.Context) error {
	f, err := h.publicForm(c, h.service.DB, false)
	if err != nil {
		return err
	}
	d, err := marketing.DefinitionForForm(f)
	if err != nil {
		return err
	}
	c.Response().Header().Set("Cache-Control", "no-store")
	// Explicit projection never exposes audience routing, tokens, or submitted answers.
	return c.JSON(200, map[string]interface{}{"name": f.Name, "description": f.Description, "fields": f.Fields, "buttonText": f.SubmitButtonText, "successMessage": f.SuccessMessage, "theme": f.Theme, "definition": d, "version": f.Version})
}
func (h *MarketingHandler) FormManifest(c echo.Context) error {
	f, err := h.publicForm(c, h.service.DB, false)
	if err != nil {
		return err
	}
	d, err := marketing.DefinitionForForm(f)
	if err != nil {
		return err
	}
	c.Response().Header().Set("Cache-Control", "no-store")
	return c.JSON(200, map[string]interface{}{"name": f.Name, "definition": d, "version": f.Version})
}

// legacyEditableDefinition recognizes only the representation synthesized for old clients.
func legacyEditableDefinition(d marketing.FormDefinition) bool {
	if len(d.Pages) != 1 || d.Pages[0].ID != "details" || d.Pages[0].Condition != nil || d.Consent.Mode != "required" || d.Consent.Label != "I agree to receive emails and can unsubscribe at any time." || d.SaveProgress || d.SuccessRedirectURL != "" {
		return false
	}
	for _, q := range d.Pages[0].Fields {
		if q.Condition != nil || len(q.Options) > 0 || q.HelpText != "" || q.Placeholder != "" || q.DefaultValue != "" {
			return false
		}
		switch q.Type {
		case "TEXT", "EMAIL", "PHONE", "TEXTAREA":
		default:
			return false
		}
	}
	return true
}
