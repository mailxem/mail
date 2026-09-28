package handlers

import (
	"errors"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/models"
	"net/url"
	"strings"
)

func (h *MarketingHandler) Branding(c echo.Context) error {
	var input struct {
		DashboardName string `json:"dashboardName"`
		LogoURL       string `json:"logoUrl"`
	}
	if c.Request().Method == "PUT" {
		if c.Bind(&input) != nil {
			return bad("Invalid branding settings")
		}
		input.DashboardName = strings.TrimSpace(input.DashboardName)
		if len(input.DashboardName) < 2 || len(input.DashboardName) > 80 {
			return bad("Workspace name must contain 2–80 characters")
		}
		if input.LogoURL != "" {
			u, err := url.Parse(input.LogoURL)
			if err != nil || len(input.LogoURL) > 2048 || u.Scheme != "https" || u.Host == "" || u.User != nil {
				return bad("Use an HTTPS logo URL")
			}
		}
	}
	var branding models.BrandingSettings
	err := h.service.DB.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
		var workspace models.Team
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&workspace, "id = ? AND is_deleted = false", team(c)).Error; err != nil {
			return missing()
		}
		var settings models.TeamSettings
		err := tx.Where("team_id = ? AND is_deleted = false", team(c)).First(&settings).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if settings.ID != "" {
			if err = tx.First(&branding, "id = ?", settings.BrandingSettingsID).Error; err != nil {
				return err
			}
		} else {
			branding.DashboardName = workspace.Name
		}
		if c.Request().Method == "GET" {
			return nil
		}
		branding.DashboardName, branding.LogoURL = input.DashboardName, input.LogoURL
		if err = tx.Omit(clause.Associations).Save(&branding).Error; err != nil {
			return err
		}
		if settings.ID == "" {
			settings.TeamID, settings.BrandingSettingsID = team(c), branding.ID
			if err = tx.Omit(clause.Associations).Create(&settings).Error; err != nil {
				return err
			}
		}
		return tx.Model(&workspace).Update("name", input.DashboardName).Error
	})
	if err != nil {
		return err
	}
	return c.JSON(200, branding)
}
