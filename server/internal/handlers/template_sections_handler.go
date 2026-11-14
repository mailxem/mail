package handlers

import (
	"encoding/base64"
	"encoding/json"
	"slices"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/marketing"
	"kori/internal/models"
	"kori/internal/utils"
)

// TemplateSections lists a template's sections. With listId it counts how
// many of the list's readers get each one, and how many would get an email
// with nothing in it; with contactId it gives the sections that contact gets
// (as the utils.SectionsKey value), so a test email shows them that way.
func (h *MarketingHandler) TemplateSections(c echo.Context) error {
	tx := h.service.DB.WithContext(c.Request().Context())
	var template models.Template
	if err := h.scoped(c).First(&template, "id = ?", c.Param("id")).Error; err != nil {
		return missing()
	}
	body, err := marketing.TemplateHTML(tx, &template)
	if err != nil {
		return bad("This template has no content yet")
	}
	var reach marketing.SectionReach
	if list := c.QueryParam("listId"); list != "" {
		if uuid.Validate(list) != nil {
			return bad("Choose a list")
		}
		reach, err = marketing.ListReach(tx, team(c), list, body)
	} else {
		reach, err = marketing.Reach(tx, team(c), body)
	}
	if err != nil {
		return err
	}
	out := map[string]any{"sections": reach.Sections, "readers": reach.Readers, "empty": reach.Empty}
	if contact := c.QueryParam("contactId"); contact != "" {
		var n int64
		if uuid.Validate(contact) != nil || h.scoped(c).Model(&models.Contact{}).Where("id = ?", contact).Count(&n).Error != nil || n == 0 {
			return missing()
		}
		picker, err := marketing.NewSectionPicker(tx, team(c), body)
		if err != nil {
			return err
		}
		data := map[string]string{}
		out["contactEmpty"] = !picker.Pick(contact, data)
		out["plan"] = data[utils.SectionsKey]
	}
	if design, err := base64.StdEncoding.DecodeString(template.DesignJSON); err == nil && len(design) > 0 {
		if rows, err := utils.EditorRows(body, design); err == nil {
			out["rows"] = rows
		}
	}
	return c.JSON(200, out)
}

// SetTemplateSections replaces the sections of a template made in the editor,
// by row ID, and marks them in its HTML.
func (h *MarketingHandler) SetTemplateSections(c echo.Context) error {
	var in struct {
		Sections map[string]utils.Section `json:"sections"`
	}
	if c.Bind(&in) != nil || len(in.Sections) > 100 {
		return bad("Send sections by row ID")
	}
	var row models.Template
	err := h.service.DB.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
		if err := marketing.Scope(tx, team(c)).Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, "id = ?", c.Param("id")).Error; err != nil {
			return missing()
		}
		data, err := base64.StdEncoding.DecodeString(row.DesignJSON)
		var design map[string]json.RawMessage
		if err != nil || json.Unmarshal(data, &design) != nil || row.HTMLBody == "" {
			return bad("Only templates made in the editor have rows; mark sections in other HTML with <!--xem:section {...}--> and <!--/xem:section-->")
		}
		_, rows, err := utils.EditorSections(data)
		if err != nil {
			return bad("This template's design can't be read")
		}
		for id, s := range in.Sections {
			if !slices.Contains(rows, id) {
				return bad("This template has no row " + id)
			}
			if err := s.Validate(); err != nil {
				return bad(err.Error())
			}
		}
		if design["xemSections"], err = json.Marshal(in.Sections); err != nil {
			return err
		}
		if data, err = json.Marshal(design); err != nil {
			return err
		}
		body, err := utils.MarkSections(row.HTMLBody, data)
		if err != nil {
			return bad(err.Error())
		}
		row.HTMLBody, row.DesignJSON = body, base64.StdEncoding.EncodeToString(data)
		return tx.Model(&row).Updates(map[string]any{"html_body": row.HTMLBody, "design_json": row.DesignJSON}).Error
	})
	if err != nil {
		return err
	}
	return c.JSON(200, map[string]any{"sections": utils.ParseSections(row.HTMLBody)})
}
