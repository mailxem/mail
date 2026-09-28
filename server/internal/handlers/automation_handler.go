package handlers

import (
	"context"
	"kori/internal/models"
	"kori/internal/services"
	"kori/internal/tasks"
	"kori/internal/utils/logger"
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

var automationHandlerLog = logger.New("AUTOMATION_HANDLER")

type AutomationHandler struct {
	service    *services.AutomationService
	taskClient tasks.AutomationEnqueuer
	db         *gorm.DB
}

func NewAutomationHandler(db *gorm.DB, taskClient tasks.AutomationEnqueuer) *AutomationHandler {
	return &AutomationHandler{
		service:    services.NewAutomationService(db),
		taskClient: taskClient,
		db:         db,
	}
}

// CreateAutomation handles POST /automations
func (h *AutomationHandler) CreateAutomation(c echo.Context) error {
	var automation models.Automation
	if err := c.Bind(&automation); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	automation.TeamID = c.Get("teamID").(string)
	automation.ID = ""
	automation.IsActive = false // Start as inactive

	if err := h.service.SaveGraph(c.Request().Context(), &automation); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	automationHandlerLog.Success("Created automation %s for team %s", automation.ID, automation.TeamID)

	return c.JSON(http.StatusCreated, automation)
}

// GetAutomation handles GET /automations/:id
func (h *AutomationHandler) GetAutomation(c echo.Context) error {
	id := c.Param("id")
	teamID := c.Get("teamID").(string)

	automation, err := h.service.GetWithNodesAndEdges(context.Background(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Automation not found"})
	}

	// Verify team ownership
	if automation.TeamID != teamID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
	}

	return c.JSON(http.StatusOK, automation)
}

// ListAutomations handles GET /automations
func (h *AutomationHandler) ListAutomations(c echo.Context) error {
	teamID := c.Get("teamID").(string)

	// TODO: Add pagination support
	automations, _, err := h.service.List(
		context.Background(),
		1,
		100,
		map[string]interface{}{"team_id": teamID},
		nil,
		[]string{"created_at"},
		"DESC",
		"Nodes",
		"Edges",
	)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, automations)
}

// UpdateAutomation handles PUT /automations/:id
func (h *AutomationHandler) UpdateAutomation(c echo.Context) error {
	id := c.Param("id")
	teamID := c.Get("teamID").(string)

	// Verify ownership
	existing, err := h.service.Get(context.Background(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Automation not found"})
	}

	if existing.TeamID != teamID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
	}

	var automation models.Automation
	if err := c.Bind(&automation); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	automation.ID = id
	automation.TeamID = teamID

	if err := h.service.SaveGraph(c.Request().Context(), &automation); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	automationHandlerLog.Success("Updated automation %s", id)

	return c.JSON(http.StatusOK, automation)
}

// DeleteAutomation handles DELETE /automations/:id
func (h *AutomationHandler) DeleteAutomation(c echo.Context) error {
	id := c.Param("id")
	teamID := c.Get("teamID").(string)

	// Verify ownership
	automation, err := h.service.Get(context.Background(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Automation not found"})
	}

	if automation.TeamID != teamID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
	}

	if err := h.service.Delete(context.Background(), id); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	automationHandlerLog.Success("Deleted automation %s", id)

	return c.JSON(http.StatusOK, map[string]string{"message": "Automation deleted successfully"})
}

// ActivateAutomation handles POST /automations/:id/activate
func (h *AutomationHandler) ActivateAutomation(c echo.Context) error {
	id := c.Param("id")
	teamID := c.Get("teamID").(string)

	// Verify ownership
	automation, err := h.service.Get(context.Background(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Automation not found"})
	}

	if automation.TeamID != teamID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
	}

	if err := h.service.Activate(context.Background(), id); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	automationHandlerLog.Success("Activated automation %s", id)

	return c.JSON(http.StatusOK, map[string]string{"message": "Automation activated successfully"})
}

// DeactivateAutomation handles POST /automations/:id/deactivate
func (h *AutomationHandler) DeactivateAutomation(c echo.Context) error {
	id := c.Param("id")
	teamID := c.Get("teamID").(string)

	// Verify ownership
	automation, err := h.service.Get(context.Background(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Automation not found"})
	}

	if automation.TeamID != teamID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
	}

	if err := h.service.Deactivate(context.Background(), id); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	automationHandlerLog.Success("Deactivated automation %s", id)

	return c.JSON(http.StatusOK, map[string]string{"message": "Automation deactivated successfully"})
}

// TriggerAutomation handles POST /automations/:id/trigger
func (h *AutomationHandler) TriggerAutomation(c echo.Context) error {
	id := c.Param("id")
	teamID := c.Get("teamID").(string)

	// Verify ownership
	automation, err := h.service.Get(context.Background(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Automation not found"})
	}

	if automation.TeamID != teamID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
	}

	if !automation.IsActive {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Automation is not active"})
	}

	// Parse request body
	var req struct {
		ContactIDs  []string               `json:"contactIds" validate:"required,min=1"`
		TriggerData map[string]interface{} `json:"triggerData"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if len(req.ContactIDs) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "At least one contactId is required"})
	}

	if len(req.ContactIDs) > 100 {
		return bad("Select at most 100 contacts")
	}
	var count int64
	if err := h.db.Model(&models.Contact{}).Where("id IN ? AND team_id = ? AND is_deleted = ? AND status = ?", req.ContactIDs, teamID, false, models.SubscriberStatusActive).Count(&count).Error; err != nil {
		return err
	}
	unique := map[string]bool{}
	for _, id := range req.ContactIDs {
		unique[id] = true
	}
	if count != int64(len(unique)) {
		return bad("Select active contacts from this workspace")
	}
	if h.taskClient == nil {
		return echo.NewHTTPError(503, "Queue is unavailable")
	}
	// Enqueue execution tasks for each contact
	tasksEnqueued := 0
	for _, contactID := range req.ContactIDs {
		task := tasks.AutomationExecuteTask{
			AutomationID: id,
			ContactID:    contactID,
			TriggerData:  req.TriggerData,
		}

		if err := h.taskClient.EnqueueAutomationTask(context.Background(), task, 0); err != nil {
			automationHandlerLog.Error("Failed to enqueue task for contact %s: %v", err, contactID)
			continue
		}
		tasksEnqueued++
	}

	automationHandlerLog.Success("Triggered automation %s for %d contacts", id, tasksEnqueued)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":       "Automation triggered successfully",
		"tasksEnqueued": tasksEnqueued,
		"totalContacts": len(req.ContactIDs),
	})
}

// GetExecutions handles GET /automations/:id/executions
func (h *AutomationHandler) GetExecutions(c echo.Context) error {
	automationID := c.Param("id")
	teamID := c.Get("teamID").(string)

	// Verify automation ownership
	automation, err := h.service.Get(context.Background(), automationID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Automation not found"})
	}

	if automation.TeamID != teamID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
	}

	// Get executions
	var executions []models.AutomationExecution
	query := h.db.Where("automation_id = ?", automationID).
		Preload("Contact").
		Preload("CurrentNode").
		Order("created_at DESC")

	// Optional filters
	if status := c.QueryParam("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if contactID := c.QueryParam("contactId"); contactID != "" {
		query = query.Where("contact_id = ?", contactID)
	}

	if err := query.Limit(100).Find(&executions).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, executions)
}

// GetExecutionDetail handles GET /executions/:id
func (h *AutomationHandler) GetExecutionDetail(c echo.Context) error {
	executionID := c.Param("id")
	teamID := c.Get("teamID").(string)

	// Get execution with full details
	var execution models.AutomationExecution
	if err := h.db.Where("id = ?", executionID).
		Preload("Automation").
		Preload("Contact").
		Preload("CurrentNode").
		First(&execution).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Execution not found"})
	}

	// Verify team ownership through automation
	if execution.Automation == nil || execution.Automation.TeamID != teamID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
	}

	return c.JSON(http.StatusOK, execution)
}
