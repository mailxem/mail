package handlers

import (
	"github.com/labstack/echo/v4"
	"kori/internal/models"
)

func (h *AutomationHandler) ValidateAutomation(c echo.Context) error {
	var a models.Automation
	if err := c.Bind(&a); err != nil {
		return bad("Invalid workflow")
	}
	a.TeamID = team(c)
	if len(a.Nodes) > 100 || len(a.Edges) > 200 {
		return bad("Workflow is too large")
	}
	if err := h.service.ValidateGraph(&a); err != nil {
		return bad(err.Error())
	}
	if err := h.service.ValidateConfiguration(&a); err != nil {
		return bad(err.Error())
	}
	return c.JSON(200, map[string]bool{"valid": true})
}
