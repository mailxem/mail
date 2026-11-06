package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

// owned returns 404 unless the automation exists in the caller's workspace.
func (h *AutomationHandler) owned(c echo.Context) error {
	automation, err := h.service.Get(c.Request().Context(), c.Param("id"))
	if err != nil || automation.IsDeleted || automation.TeamID != c.Get("teamID").(string) {
		return missing()
	}
	return nil
}

// PublishAutomation handles POST /automations/:id/publish
func (h *AutomationHandler) PublishAutomation(c echo.Context) error {
	if err := h.owned(c); err != nil {
		return err
	}
	var req struct {
		MoveRunning bool `json:"moveRunning"`
	}
	if err := c.Bind(&req); err != nil {
		return bad(err.Error())
	}
	userID, _ := c.Get("userID").(string)
	version, err := h.service.Publish(c.Request().Context(), c.Param("id"), userID, req.MoveRunning)
	if err != nil {
		return bad(err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"number": version.Number, "publishedAt": version.PublishedAt})
}

// RestoreDraft handles POST /automations/:id/discard (back to live) and
// POST /automations/:id/versions/:number/restore
func (h *AutomationHandler) RestoreDraft(c echo.Context) error {
	if err := h.owned(c); err != nil {
		return err
	}
	number := 0
	if raw := c.Param("number"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			return bad("Choose a published version")
		}
		number = n
	}
	if err := h.service.RestoreDraft(c.Request().Context(), c.Param("id"), number); err != nil {
		return bad(err.Error())
	}
	return h.GetAutomation(c)
}

// ListVersions handles GET /automations/:id/versions
func (h *AutomationHandler) ListVersions(c echo.Context) error {
	if err := h.owned(c); err != nil {
		return err
	}
	versions, err := h.service.Versions(c.Request().Context(), c.Param("id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, versions)
}

// GetStepStats handles GET /automations/:id/stats?from=&to= (RFC 3339, default last 30 days)
func (h *AutomationHandler) GetStepStats(c echo.Context) error {
	if err := h.owned(c); err != nil {
		return err
	}
	to, from := time.Now(), time.Now().AddDate(0, 0, -30)
	for name, target := range map[string]*time.Time{"from": &from, "to": &to} {
		if raw := c.QueryParam(name); raw != "" {
			t, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				return bad("Use RFC 3339 dates for from and to")
			}
			*target = t
		}
	}
	if !from.Before(to) {
		return bad("from must be before to")
	}
	steps, err := h.service.StepStats(c.Request().Context(), c.Param("id"), from, to)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"from": from, "to": to, "steps": steps})
}
