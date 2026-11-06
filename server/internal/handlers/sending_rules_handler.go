package handlers

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"gorm.io/datatypes"
	"gorm.io/gorm/clause"
	"kori/internal/models"
	"kori/internal/workflowconfig"
)

// SendingRules reads or replaces the workspace defaults automations send by:
// its time zone, a default send window and the per-contact frequency cap.
func (h *MarketingHandler) SendingRules(c echo.Context) error {
	type rules struct {
		Timezone           string                     `json:"timezone"`
		SendWindow         *workflowconfig.SendWindow `json:"sendWindow"`
		FrequencyCap       int                        `json:"frequencyCap"`
		FrequencyCapPeriod string                     `json:"frequencyCapPeriod"`
	}
	db := h.service.DB.WithContext(c.Request().Context())
	if c.Request().Method == "PUT" {
		var in rules
		if c.Bind(&in) != nil {
			return bad("Invalid sending rules")
		}
		if !workflowconfig.ValidTimezone(in.Timezone) {
			return bad("Choose a time zone such as Europe/London")
		}
		if in.FrequencyCap < 0 || in.FrequencyCap > 100 || (in.FrequencyCapPeriod != "day" && in.FrequencyCapPeriod != "week") {
			return bad("The frequency cap must be 0–100 emails per day or week")
		}
		window := datatypes.JSON("null")
		if in.SendWindow != nil {
			if err := in.SendWindow.Validate(); err != nil {
				return bad(err.Error())
			}
			window, _ = json.Marshal(in.SendWindow)
		}
		row := models.TeamSendTimeDefaults{TeamID: team(c), DefaultTimezone: in.Timezone, AutomationWindow: window, FrequencyCap: in.FrequencyCap, FrequencyCapPeriod: in.FrequencyCapPeriod}
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "team_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"default_timezone", "automation_window", "frequency_cap", "frequency_cap_period", "updated_at"}),
		}).Create(&row).Error; err != nil {
			return err
		}
	}
	var row models.TeamSendTimeDefaults
	if err := db.Where("team_id = ?", team(c)).Limit(1).Find(&row).Error; err != nil {
		return err
	}
	out := rules{Timezone: row.DefaultTimezone, FrequencyCap: row.FrequencyCap, FrequencyCapPeriod: row.FrequencyCapPeriod}
	if out.Timezone == "" {
		out.Timezone = "UTC"
	}
	if out.FrequencyCapPeriod == "" {
		out.FrequencyCapPeriod = "day"
	}
	out.SendWindow, _ = workflowconfig.ParseSendWindow(row.AutomationWindow)
	return c.JSON(200, out)
}
