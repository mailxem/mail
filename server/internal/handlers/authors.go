package handlers

import (
	"strings"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"kori/internal/marketing"
	"kori/internal/models"
)

// Authors lists the workspace's authors, the default first.
func (h *MarketingHandler) Authors(c echo.Context) error {
	authors := []models.Author{}
	if err := h.scoped(c).Order("is_default DESC, name").Limit(200).Find(&authors).Error; err != nil {
		return err
	}
	return c.JSON(200, authors)
}

// SaveAuthor adds an author, or with :id changes one. Making an author the
// default takes it from whoever had it.
func (h *MarketingHandler) SaveAuthor(c echo.Context) error {
	var in struct {
		Name, PhotoURL, Email             string
		BylineEmail, BylineWeb, IsDefault bool
	}
	if err := c.Bind(&in); err != nil {
		return bad("Invalid author")
	}
	a := models.Author{TeamID: team(c), Name: strings.TrimSpace(in.Name), PhotoURL: strings.TrimSpace(in.PhotoURL), Email: strings.TrimSpace(in.Email), BylineEmail: in.BylineEmail, BylineWeb: in.BylineWeb, IsDefault: in.IsDefault}
	if a.Name == "" || len(a.Name) > 80 || strings.ContainsAny(a.Name, "\r\n") {
		return bad("Enter a name of up to 80 characters")
	}
	if a.Email != "" && !marketing.ValidEmail(a.Email) {
		return bad("Enter a valid email for replies")
	}
	if !marketing.ValidSiteLogo(a.PhotoURL) {
		return bad("The photo must be an https:// image address")
	}
	err := h.service.DB.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
		if id := c.Param("id"); id != "" {
			var old models.Author
			if err := marketing.Scope(tx, a.TeamID).First(&old, "id = ?", id).Error; err != nil {
				return missing()
			}
			a.Base = old.Base
		}
		if a.IsDefault {
			if err := marketing.Scope(tx, a.TeamID).Model(&models.Author{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return tx.Save(&a).Error
	})
	if err != nil {
		return err
	}
	return c.JSON(200, a)
}

// DeleteAuthor removes an author. Campaigns they signed fall back to the
// workspace's default author.
func (h *MarketingHandler) DeleteAuthor(c echo.Context) error {
	result := h.scoped(c).Model(&models.Author{}).Where("id = ?", c.Param("id")).Updates(map[string]any{"is_deleted": true, "is_default": false})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return missing()
	}
	return c.NoContent(204)
}

// authorID checks that a picked author is the workspace's own; "" means none.
func (h *MarketingHandler) authorID(c echo.Context, raw string) (*string, error) {
	if raw = strings.TrimSpace(raw); raw == "" {
		return nil, nil
	}
	var n int64
	if err := h.scoped(c).Model(&models.Author{}).Where("id = ?", raw).Count(&n).Error; err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, bad("Choose one of your authors")
	}
	return &raw, nil
}

// AssignAuthor picks the author of a campaign or newsletter (model); an
// empty authorId goes back to the workspace's default.
func (h *MarketingHandler) AssignAuthor(model any) echo.HandlerFunc {
	return func(c echo.Context) error {
		var in struct{ AuthorID string }
		if err := c.Bind(&in); err != nil {
			return bad("Invalid author")
		}
		id, err := h.authorID(c, in.AuthorID)
		if err != nil {
			return err
		}
		result := h.scoped(c).Model(model).Where("id = ?", c.Param("id")).Update("author_id", id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return missing()
		}
		return c.JSON(200, map[string]any{"authorId": id})
	}
}
