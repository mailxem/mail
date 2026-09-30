package handlers

import (
	"encoding/base64"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/marketing"
	"kori/internal/models"
)

// SaveTemplate serializes editor saves with ordinary CRUD updates via a row lock.
// The expected timestamp prevents a slow AI revision from overwriting newer work.
func (h *MarketingHandler) SaveTemplate(c echo.Context) error {
	var in struct {
		CreationID        string     `json:"creationId"`
		Name              string     `json:"name"`
		Subject           string     `json:"subject"`
		CategoryID        string     `json:"categoryId"`
		HTMLBody          string     `json:"htmlBody"`
		DesignJSON        string     `json:"designJson"`
		ExpectedUpdatedAt *time.Time `json:"expectedUpdatedAt"`
	}
	if c.Bind(&in) != nil {
		return bad("Invalid template")
	}
	in.Name, in.Subject = strings.TrimSpace(in.Name), strings.TrimSpace(in.Subject)
	data, err := base64.StdEncoding.DecodeString(in.DesignJSON)
	var design struct {
		SchemaVersion int `json:"schemaVersion"`
		Body          struct {
			Rows []json.RawMessage `json:"rows"`
		} `json:"body"`
	}
	if len(in.Name) < 2 || len(in.Name) > 200 || len(in.Subject) == 0 || len(in.Subject) > 200 || strings.ContainsAny(in.Subject, "\r\n") || uuid.Validate(in.CategoryID) != nil || len(in.HTMLBody) == 0 || len(in.HTMLBody) > 1_000_000 || len(data) > 1_000_000 || err != nil || json.Unmarshal(data, &design) != nil || design.SchemaVersion < 1 || design.Body.Rows == nil {
		return bad("A name, subject, category and editable design are required")
	}
	id := c.Param("id")
	if id == "" && uuid.Validate(in.CreationID) != nil {
		return bad("A template creation ID is required")
	}
	if id != "" && (uuid.Validate(id) != nil || in.ExpectedUpdatedAt == nil) {
		return bad("The saved template version is required")
	}
	var row models.Template
	variables := pq.StringArray{}
	seen := map[string]bool{}
	for _, match := range regexp.MustCompile(`\{\{([^}]+)\}\}`).FindAllStringSubmatch(in.HTMLBody, -1) {
		name := strings.TrimSpace(match[1])
		if name != "" && !seen[name] {
			variables = append(variables, name)
			seen[name] = true
		}
	}
	err = h.service.DB.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
		if id == "" {
			var existing models.Template
			if e := tx.First(&existing, "id = ?", in.CreationID).Error; e == nil {
				if existing.TeamID == team(c) && !existing.IsDeleted && existing.Name == in.Name && existing.Subject == in.Subject && existing.CategoryID == in.CategoryID && existing.HTMLBody == in.HTMLBody && existing.DesignJSON == in.DesignJSON {
					row = existing
					return nil
				}
				return echo.NewHTTPError(409, "Template creation was already used; check your saved templates")
			} else if e != gorm.ErrRecordNotFound {
				return e
			}
		}
		if id != "" {
			if err := marketing.Scope(tx, team(c)).Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, "id = ?", id).Error; err != nil {
				return missing()
			}
			if !row.UpdatedAt.Equal(*in.ExpectedUpdatedAt) {
				return echo.NewHTTPError(409, "Template changed; reload before saving")
			}
		}
		var category models.EmailCategory
		if err := marketing.Scope(tx, team(c)).First(&category, "id = ?", in.CategoryID).Error; err != nil {
			return bad("Choose a category in this workspace")
		}
		if id == "" {
			row = models.Template{Base: models.Base{ID: in.CreationID}, TeamID: team(c), Name: in.Name, Subject: in.Subject, CategoryID: in.CategoryID, HTMLBody: in.HTMLBody, DesignJSON: in.DesignJSON, Variables: variables}
			result := tx.Omit(clause.Associations).Clauses(clause.OnConflict{DoNothing: true}).Create(&row)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				var existing models.Template
				if err := tx.First(&existing, "id = ?", in.CreationID).Error; err != nil {
					return err
				}
				if existing.TeamID != team(c) || existing.IsDeleted || existing.Name != in.Name || existing.Subject != in.Subject || existing.CategoryID != in.CategoryID || existing.HTMLBody != in.HTMLBody || existing.DesignJSON != in.DesignJSON {
					return echo.NewHTTPError(409, "Template creation was already used")
				}
				row = existing
			}
			return nil
		}
		// Clear the legacy file reference so previews and sends use the new HTML.
		return tx.Model(&row).Updates(map[string]any{"name": in.Name, "subject": in.Subject, "category_id": in.CategoryID, "html_body": in.HTMLBody, "design_json": in.DesignJSON, "html_file_id": nil, "variables": variables}).Error
	})
	if err != nil {
		return err
	}
	status := 200
	if id == "" {
		status = 201
	}
	return c.JSON(status, row)
}
