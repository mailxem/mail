package handlers

import (
	"kori/internal/models"
	"kori/internal/services"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type GoalTrackingHandler struct {
	goalService *services.GoalTrackingService
}

func NewGoalTrackingHandler(db *gorm.DB) *GoalTrackingHandler {
	return &GoalTrackingHandler{
		goalService: services.NewGoalTrackingService(db),
	}
}

// CreateGoal creates a new goal for an automation
// @Summary Create goal
// @Description Create a new goal for an automation
// @Tags Goal Tracking
// @Accept json
// @Produce json
// @Param body body models.AutomationGoal true "Goal data"
// @Success 201 {object} models.AutomationGoal
// @Router /api/v1/goals [post]
// @Security ApiKeyAuth
func (h *GoalTrackingHandler) CreateGoal(c echo.Context) error {
	var goal models.AutomationGoal
	if err := c.Bind(&goal); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := h.goalService.CreateGoal(c.Request().Context(), &goal); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create goal")
	}

	return c.JSON(http.StatusCreated, goal)
}

// GetGoal retrieves a goal by ID
// @Summary Get goal
// @Description Get details of a specific goal
// @Tags Goal Tracking
// @Accept json
// @Produce json
// @Param id path string true "Goal ID"
// @Success 200 {object} models.AutomationGoal
// @Router /api/v1/goals/{id} [get]
// @Security ApiKeyAuth
func (h *GoalTrackingHandler) GetGoal(c echo.Context) error {
	id := c.Param("id")

	goal, err := h.goalService.GetGoal(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Goal not found")
	}

	return c.JSON(http.StatusOK, goal)
}

// ListGoals retrieves all goals for an automation
// @Summary List goals
// @Description Get all goals for an automation
// @Tags Goal Tracking
// @Accept json
// @Produce json
// @Param automationId query string true "Automation ID"
// @Param active_only query bool false "Show only active goals"
// @Success 200 {array} models.AutomationGoal
// @Router /api/v1/goals [get]
// @Security ApiKeyAuth
func (h *GoalTrackingHandler) ListGoals(c echo.Context) error {
	automationID := c.QueryParam("automationId")
	if automationID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "automationId is required")
	}

	activeOnly := c.QueryParam("active_only") == "true"

	goals, err := h.goalService.ListGoals(c.Request().Context(), automationID, activeOnly)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list goals")
	}

	return c.JSON(http.StatusOK, goals)
}

// UpdateGoal updates a goal
// @Summary Update goal
// @Description Update an existing goal
// @Tags Goal Tracking
// @Accept json
// @Produce json
// @Param id path string true "Goal ID"
// @Param body body object true "Update data"
// @Success 200 {object} models.AutomationGoal
// @Router /api/v1/goals/{id} [put]
// @Security ApiKeyAuth
func (h *GoalTrackingHandler) UpdateGoal(c echo.Context) error {
	id := c.Param("id")

	var updates map[string]interface{}
	if err := c.Bind(&updates); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := h.goalService.UpdateGoal(c.Request().Context(), id, updates); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update goal")
	}

	goal, err := h.goalService.GetGoal(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve updated goal")
	}

	return c.JSON(http.StatusOK, goal)
}

// DeleteGoal deletes a goal
// @Summary Delete goal
// @Description Delete a goal and all its conversions
// @Tags Goal Tracking
// @Accept json
// @Produce json
// @Param id path string true "Goal ID"
// @Success 204
// @Router /api/v1/goals/{id} [delete]
// @Security ApiKeyAuth
func (h *GoalTrackingHandler) DeleteGoal(c echo.Context) error {
	id := c.Param("id")

	if err := h.goalService.DeleteGoal(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete goal")
	}

	return c.NoContent(http.StatusNoContent)
}

// GetGoalStats retrieves statistics for a goal
// @Summary Get goal stats
// @Description Get detailed statistics for a goal
// @Tags Goal Tracking
// @Accept json
// @Produce json
// @Param id path string true "Goal ID"
// @Success 200 {object} models.GoalStats
// @Router /api/v1/goals/{id}/stats [get]
// @Security ApiKeyAuth
func (h *GoalTrackingHandler) GetGoalStats(c echo.Context) error {
	id := c.Param("id")

	stats, err := h.goalService.GetGoalStats(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve goal stats")
	}

	return c.JSON(http.StatusOK, stats)
}

// GetAutomationGoalSummary retrieves goal summary for an automation
// @Summary Get automation goal summary
// @Description Get summary of all goals for an automation
// @Tags Goal Tracking
// @Accept json
// @Produce json
// @Param automationId query string true "Automation ID"
// @Success 200 {object} models.AutomationGoalSummary
// @Router /api/v1/goals/summary [get]
// @Security ApiKeyAuth
func (h *GoalTrackingHandler) GetAutomationGoalSummary(c echo.Context) error {
	automationID := c.QueryParam("automationId")
	if automationID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "automationId is required")
	}

	summary, err := h.goalService.GetAutomationGoalSummary(c.Request().Context(), automationID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve summary")
	}

	return c.JSON(http.StatusOK, summary)
}

// GetConversions retrieves conversions for a goal
// @Summary Get goal conversions
// @Description Get all conversions for a specific goal
// @Tags Goal Tracking
// @Accept json
// @Produce json
// @Param id path string true "Goal ID"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} object
// @Router /api/v1/goals/{id}/conversions [get]
// @Security ApiKeyAuth
func (h *GoalTrackingHandler) GetConversions(c echo.Context) error {
	id := c.Param("id")

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 50
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	conversions, total, err := h.goalService.GetConversions(c.Request().Context(), id, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve conversions")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"conversions": conversions,
		"total":       total,
		"limit":       limit,
		"offset":      offset,
	})
}

// GetContactGoalHistory retrieves goal conversions for a contact
// @Summary Get contact goal history
// @Description Get all goal conversions for a specific contact
// @Tags Goal Tracking
// @Accept json
// @Produce json
// @Param contactId path string true "Contact ID"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} object
// @Router /api/v1/contacts/{contactId}/goals [get]
// @Security ApiKeyAuth
func (h *GoalTrackingHandler) GetContactGoalHistory(c echo.Context) error {
	contactID := c.Param("contactId")

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 50
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	conversions, total, err := h.goalService.GetContactGoalHistory(c.Request().Context(), contactID, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve goal history")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"conversions": conversions,
		"total":       total,
		"limit":       limit,
		"offset":      offset,
	})
}

// GetGoalsByDateRange retrieves conversions within a date range
// @Summary Get goals by date range
// @Description Get goal conversions within a specific date range
// @Tags Goal Tracking
// @Accept json
// @Produce json
// @Param automationId query string true "Automation ID"
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {array} models.GoalConversion
// @Router /api/v1/goals/date-range [get]
// @Security ApiKeyAuth
func (h *GoalTrackingHandler) GetGoalsByDateRange(c echo.Context) error {
	automationID := c.QueryParam("automationId")
	if automationID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "automationId is required")
	}

	startDateStr := c.QueryParam("start_date")
	endDateStr := c.QueryParam("end_date")

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid start_date format (use YYYY-MM-DD)")
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid end_date format (use YYYY-MM-DD)")
	}

	conversions, err := h.goalService.GetGoalsByDateRange(c.Request().Context(), automationID, startDate, endDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve conversions")
	}

	return c.JSON(http.StatusOK, conversions)
}

// GetTopConvertingContacts retrieves contacts with most conversions
// @Summary Get top converting contacts
// @Description Get contacts with the most goal conversions
// @Tags Goal Tracking
// @Accept json
// @Produce json
// @Param automationId query string true "Automation ID"
// @Param limit query int false "Limit" default(10)
// @Success 200 {array} services.ContactConversionStats
// @Router /api/v1/goals/top-contacts [get]
// @Security ApiKeyAuth
func (h *GoalTrackingHandler) GetTopConvertingContacts(c echo.Context) error {
	automationID := c.QueryParam("automationId")
	if automationID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "automationId is required")
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 10
	}

	stats, err := h.goalService.GetTopConvertingContacts(c.Request().Context(), automationID, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve top contacts")
	}

	return c.JSON(http.StatusOK, stats)
}

// GetGoalProgress retrieves progress towards a goal target
// @Summary Get goal progress
// @Description Get progress towards a goal target
// @Tags Goal Tracking
// @Accept json
// @Produce json
// @Param id path string true "Goal ID"
// @Success 200 {object} services.GoalProgress
// @Router /api/v1/goals/{id}/progress [get]
// @Security ApiKeyAuth
func (h *GoalTrackingHandler) GetGoalProgress(c echo.Context) error {
	id := c.Param("id")

	progress, err := h.goalService.CalculateGoalProgress(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, progress)
}
