package handlers

import (
	"errors"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/marketing"
	"kori/internal/models"
	"strings"
)

type tagView struct {
	models.Tag
	ContactCount int64 `json:"contactCount"`
}

func (h *MarketingHandler) Tags(c echo.Context) error {
	var rows []tagView
	err := h.scoped(c).Model(&models.Tag{}).Select("tags.*, (SELECT COUNT(*) FROM contact_tags ct JOIN contacts ON contacts.id = ct.contact_id WHERE ct.tag_id = tags.id AND contacts.team_id = tags.team_id AND contacts.is_deleted = false) AS contact_count").Order("tags.name").Limit(500).Find(&rows).Error
	if err != nil {
		return err
	}
	return c.JSON(200, rows)
}
func (h *MarketingHandler) SaveTag(c echo.Context) error {
	var input struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	if c.Bind(&input) != nil {
		return bad("Invalid tag")
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || len(input.Name) > 80 || len(input.Value) > 120 {
		return bad("Tag names must contain 1–80 characters")
	}
	var tag models.Tag
	err := h.service.DB.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
		// Serialize changes for a workspace so simultaneous creates cannot duplicate names.
		var workspace models.Team
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&workspace, "id = ?", team(c)).Error; err != nil {
			return missing()
		}
		id := c.Param("id")
		if id != "" {
			if uuid.Validate(id) != nil {
				return missing()
			}
			if err := marketing.Scope(tx, team(c)).First(&tag, "id = ?", id).Error; err != nil {
				return missing()
			}
		}
		var count int64
		check := marketing.Scope(tx, team(c)).Model(&models.Tag{}).Where("LOWER(name) = LOWER(?)", input.Name)
		if id != "" {
			check = check.Where("id <> ?", id)
		}
		if err := check.Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return echo.NewHTTPError(409, "A tag with this name already exists")
		}
		tag.Name, tag.Value, tag.TeamID = input.Name, input.Value, team(c)
		return tx.Omit(clause.Associations).Save(&tag).Error
	})
	if err != nil {
		return err
	}
	return c.JSON(200, tag)
}
func (h *MarketingHandler) DeleteTag(c echo.Context) error {
	if uuid.Validate(c.Param("id")) != nil {
		return missing()
	}
	err := h.service.DB.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
		var tag models.Tag
		if err := marketing.Scope(tx, team(c)).Clauses(clause.Locking{Strength: "UPDATE"}).First(&tag, "id = ?", c.Param("id")).Error; err != nil {
			return missing()
		}
		if err := tx.Exec("DELETE FROM contact_tags WHERE tag_id = ?", tag.ID).Error; err != nil {
			return err
		}
		return tx.Model(&tag).Update("is_deleted", true).Error
	})
	if err != nil {
		return err
	}
	return c.NoContent(204)
}
func (h *MarketingHandler) ContactTags(c echo.Context) error {
	if uuid.Validate(c.Param("id")) != nil {
		return missing()
	}
	var input struct {
		TagIDs []string `json:"tagIds"`
	}
	if c.Request().Method == "PUT" && (c.Bind(&input) != nil || len(input.TagIDs) > 50) {
		return bad("Choose up to 50 tags")
	}
	var tags []models.Tag
	err := h.service.DB.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
		var contact models.Contact
		if err := marketing.Scope(tx, team(c)).Clauses(clause.Locking{Strength: "UPDATE"}).First(&contact, "id = ?", c.Param("id")).Error; err != nil {
			return missing()
		}
		if c.Request().Method == "GET" {
			return marketing.Scope(tx, team(c)).Where("id IN (SELECT tag_id FROM contact_tags WHERE contact_id = ?)", contact.ID).Find(&tags).Error
		}
		unique := map[string]bool{}
		for _, id := range input.TagIDs {
			if uuid.Validate(id) != nil || unique[id] {
				return bad("Invalid tag selection")
			}
			unique[id] = true
		}
		if len(input.TagIDs) > 0 {
			if err := marketing.Scope(tx, team(c)).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ?", input.TagIDs).Order("id").Find(&tags).Error; err != nil {
				return err
			}
			if len(tags) != len(input.TagIDs) {
				return missing()
			}
		}
		return tx.Model(&contact).Omit("Tags.*").Association("Tags").Replace(tags)
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return missing()
		}
		return err
	}
	if tags == nil {
		tags = []models.Tag{}
	}
	return c.JSON(200, tags)
}
