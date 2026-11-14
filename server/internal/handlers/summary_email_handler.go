package handlers

import (
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"kori/internal/models"
	"kori/internal/notifications"
)

// SummaryEmail reads or saves how often the workspace owner gets a sending
// summary, and whether the signed-in admin gets a copy too.
func (h *MarketingHandler) SummaryEmail(c echo.Context) error {
	db := h.service.DB.WithContext(c.Request().Context())
	user, _ := c.Get("userID").(string)
	if c.Request().Method == "PUT" {
		var in struct {
			Frequency string `json:"frequency"`
			IncludeMe bool   `json:"includeMe"`
		}
		if c.Bind(&in) != nil || (in.Frequency != "daily" && in.Frequency != "weekly" && in.Frequency != "off") {
			return bad("Choose daily, weekly or off")
		}
		result := db.Model(&models.Team{}).Where("id = ? AND is_deleted = false", team(c)).Update("summary_frequency", in.Frequency)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return missing()
		}
		if err := db.Model(&models.User{}).Where("id = ? AND team_id = ?", user, team(c)).Update("summary_emails", in.IncludeMe).Error; err != nil {
			return err
		}
	}
	var out struct {
		Frequency string `json:"frequency"`
		IncludeMe bool   `json:"includeMe"`
		IsOwner   bool   `json:"isOwner"`
		Timezone  string `json:"timezone"`
	}
	if err := db.Model(&models.Team{}).Where("id = ?", team(c)).Pluck("summary_frequency", &out.Frequency).Error; err != nil {
		return err
	}
	if err := db.Model(&models.User{}).Where("id = ? AND team_id = ?", user, team(c)).Pluck("summary_emails", &out.IncludeMe).Error; err != nil {
		return err
	}
	owner, _ := notifications.SummaryOwner(db, team(c))
	out.IsOwner = owner.ID != "" && owner.ID == user
	out.Timezone = workspaceTimezone(h, c)
	return c.JSON(200, out)
}

func workspaceTimezone(h *MarketingHandler, c echo.Context) string {
	var tz []string
	h.service.DB.WithContext(c.Request().Context()).Model(&models.TeamSendTimeDefaults{}).Where("team_id = ?", team(c)).Pluck("default_timezone", &tz)
	if len(tz) == 0 || tz[0] == "" {
		return "UTC"
	}
	return tz[0]
}

// SummaryEmailPreview renders the summary for the last completed period with
// the workspace's real numbers, and says whether it would have been skipped.
func (h *MarketingHandler) SummaryEmailPreview(c echo.Context) error {
	ctx := c.Request().Context()
	db := h.service.DB.WithContext(ctx)
	frequency := c.QueryParam("frequency")
	if frequency != "daily" {
		frequency = "weekly"
	}
	period, _ := notifications.LastPeriod(frequency, workspaceTimezone(h, c), time.Now())
	summary, err := notifications.BuildSummary(ctx, db, team(c), period)
	if err != nil {
		return err
	}
	var name string
	if err := db.Model(&models.Team{}).Where("id = ?", team(c)).Pluck("name", &name).Error; err != nil {
		return err
	}
	base := os.Getenv("DASHBOARD_URL")
	if base == "" {
		base = c.Scheme() + "://" + c.Request().Host
	}
	subject, html, _ := summary.Email(name).Render(base + "/dashboard")
	return c.JSON(200, map[string]any{"subject": subject, "html": html, "quiet": summary.Quiet(), "from": period.From, "to": period.To})
}
