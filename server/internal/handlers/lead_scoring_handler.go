package handlers

import (
	"kori/internal/models"
	"kori/internal/services"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type LeadScoringHandler struct {
	scoringService *services.LeadScoringService
	ruleService    *services.ScoreRuleService
}

func NewLeadScoringHandler(db *gorm.DB) *LeadScoringHandler {
	return &LeadScoringHandler{
		scoringService: services.NewLeadScoringService(db),
		ruleService:    services.NewScoreRuleService(db),
	}
}

// GetContactScore retrieves the score for a specific contact
// @Summary Get contact score
// @Description Get the lead score for a specific contact
// @Tags Lead Scoring
// @Accept json
// @Produce json
// @Param id path string true "Contact ID"
// @Success 200 {object} models.LeadScore
// @Router /api/v1/contacts/{id}/score [get]
// @Security ApiKeyAuth
func (h *LeadScoringHandler) GetContactScore(c echo.Context) error {
	contactID := c.Param("id")

	score, err := h.scoringService.GetContactScore(c.Request().Context(), contactID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve contact score")
	}

	return c.JSON(http.StatusOK, score)
}

// UpdateContactScore manually updates a contact's score
// @Summary Update contact score
// @Description Manually add or subtract points from a contact's score
// @Tags Lead Scoring
// @Accept json
// @Produce json
// @Param id path string true "Contact ID"
// @Param body body object true "Score update request"
// @Success 200 {object} models.LeadScore
// @Router /api/v1/contacts/{id}/score [post]
// @Security ApiKeyAuth
func (h *LeadScoringHandler) UpdateContactScore(c echo.Context) error {
	contactID := c.Param("id")

	var req struct {
		Points      int                    `json:"points" validate:"required"`
		Description string                 `json:"description"`
		Metadata    map[string]interface{} `json:"metadata"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Add points
	err := h.scoringService.AddPoints(
		c.Request().Context(),
		contactID,
		"manual_adjustment",
		req.Points,
		req.Description,
		req.Metadata,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update score")
	}

	// Return updated score
	score, err := h.scoringService.GetContactScore(c.Request().Context(), contactID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve updated score")
	}

	return c.JSON(http.StatusOK, score)
}

// GetScoreHistory retrieves the score activity history for a contact
// @Summary Get score history
// @Description Get the score activity history for a specific contact
// @Tags Lead Scoring
// @Accept json
// @Produce json
// @Param id path string true "Contact ID"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} object
// @Router /api/v1/contacts/{id}/score/history [get]
// @Security ApiKeyAuth
func (h *LeadScoringHandler) GetScoreHistory(c echo.Context) error {
	contactID := c.Param("id")

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 50
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	activities, total, err := h.scoringService.GetScoreHistory(c.Request().Context(), contactID, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve score history")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"activities": activities,
		"total":      total,
		"limit":      limit,
		"offset":     offset,
	})
}

// GetTopScoredContacts retrieves contacts with highest scores
// @Summary Get top scored contacts
// @Description Get the contacts with the highest lead scores for the team
// @Tags Lead Scoring
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Success 200 {array} models.LeadScore
// @Router /api/v1/analytics/top-leads [get]
// @Security ApiKeyAuth
func (h *LeadScoringHandler) GetTopScoredContacts(c echo.Context) error {
	teamID := c.Get("teamID").(string)

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 10
	}

	scores, err := h.scoringService.GetTopScoredContacts(c.Request().Context(), teamID, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve top scored contacts")
	}

	return c.JSON(http.StatusOK, scores)
}

// GetContactsByGrade retrieves contacts by grade
// @Summary Get contacts by grade
// @Description Get all contacts with a specific grade (A, B, C, D, F)
// @Tags Lead Scoring
// @Accept json
// @Produce json
// @Param grade query string true "Grade" Enums(A, B, C, D, F)
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} object
// @Router /api/v1/analytics/contacts-by-grade [get]
// @Security ApiKeyAuth
func (h *LeadScoringHandler) GetContactsByGrade(c echo.Context) error {
	teamID := c.Get("teamID").(string)
	grade := models.ScoreGrade(c.QueryParam("grade"))

	// Validate grade
	validGrades := map[models.ScoreGrade]bool{
		models.GradeA: true,
		models.GradeB: true,
		models.GradeC: true,
		models.GradeD: true,
		models.GradeF: true,
	}

	if !validGrades[grade] {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid grade. Must be A, B, C, D, or F")
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 50
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	scores, total, err := h.scoringService.GetContactsByGrade(c.Request().Context(), teamID, grade, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve contacts")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"contacts": scores,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
		"grade":    grade,
	})
}

// GetScoreDistribution returns score distribution by grade
// @Summary Get score distribution
// @Description Get the distribution of contacts across different score grades
// @Tags Lead Scoring
// @Accept json
// @Produce json
// @Success 200 {object} map[string]int64
// @Router /api/v1/analytics/score-distribution [get]
// @Security ApiKeyAuth
func (h *LeadScoringHandler) GetScoreDistribution(c echo.Context) error {
	teamID := c.Get("teamID").(string)

	distribution, err := h.scoringService.GetScoreDistribution(c.Request().Context(), teamID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve score distribution")
	}

	return c.JSON(http.StatusOK, distribution)
}

// ListScoreRules retrieves all score rules for the team
// @Summary List score rules
// @Description Get all score rules for the team
// @Tags Lead Scoring
// @Accept json
// @Produce json
// @Param active_only query bool false "Show only active rules"
// @Success 200 {array} models.ScoreRule
// @Router /api/v1/score-rules [get]
// @Security ApiKeyAuth
func (h *LeadScoringHandler) ListScoreRules(c echo.Context) error {
	teamID := c.Get("teamID").(string)
	activeOnly := c.QueryParam("active_only") == "true"

	rules, err := h.ruleService.GetTeamRules(c.Request().Context(), teamID, activeOnly)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve score rules")
	}

	return c.JSON(http.StatusOK, rules)
}

// CreateScoreRule creates a new score rule
// @Summary Create score rule
// @Description Create a new score rule for the team
// @Tags Lead Scoring
// @Accept json
// @Produce json
// @Param body body models.ScoreRule true "Score rule data"
// @Success 201 {object} models.ScoreRule
// @Router /api/v1/score-rules [post]
// @Security ApiKeyAuth
func (h *LeadScoringHandler) CreateScoreRule(c echo.Context) error {
	teamID := c.Get("teamID").(string)

	var rule models.ScoreRule
	if err := c.Bind(&rule); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Set team ID from authenticated user
	rule.TeamID = teamID

	if err := h.ruleService.Create(c.Request().Context(), &rule); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create score rule")
	}

	return c.JSON(http.StatusCreated, rule)
}

// UpdateScoreRule updates an existing score rule
// @Summary Update score rule
// @Description Update an existing score rule
// @Tags Lead Scoring
// @Accept json
// @Produce json
// @Param id path string true "Rule ID"
// @Param body body object true "Update data"
// @Success 200 {object} models.ScoreRule
// @Router /api/v1/score-rules/{id} [put]
// @Security ApiKeyAuth
func (h *LeadScoringHandler) UpdateScoreRule(c echo.Context) error {
	id := c.Param("id")

	var updates map[string]interface{}
	if err := c.Bind(&updates); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Don't allow changing team_id
	delete(updates, "team_id")

	if err := h.ruleService.UpdateRule(c.Request().Context(), id, updates); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update score rule")
	}

	// Return updated rule
	rule, err := h.ruleService.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve updated rule")
	}

	return c.JSON(http.StatusOK, rule)
}

// DeleteScoreRule deletes a score rule
// @Summary Delete score rule
// @Description Delete a score rule (soft delete by setting is_active to false)
// @Tags Lead Scoring
// @Accept json
// @Produce json
// @Param id path string true "Rule ID"
// @Success 204
// @Router /api/v1/score-rules/{id} [delete]
// @Security ApiKeyAuth
func (h *LeadScoringHandler) DeleteScoreRule(c echo.Context) error {
	id := c.Param("id")

	if err := h.ruleService.DeleteRule(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete score rule")
	}

	return c.NoContent(http.StatusNoContent)
}

// InitializeScoreRules initializes default score rules for the team
// @Summary Initialize default score rules
// @Description Initialize default score rules for the team (called during team setup)
// @Tags Lead Scoring
// @Accept json
// @Produce json
// @Success 200 {array} models.ScoreRule
// @Router /api/v1/score-rules/initialize [post]
// @Security ApiKeyAuth
func (h *LeadScoringHandler) InitializeScoreRules(c echo.Context) error {
	teamID := c.Get("teamID").(string)

	if err := h.ruleService.InitializeDefaultRules(c.Request().Context(), teamID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to initialize score rules")
	}

	// Return all rules
	rules, err := h.ruleService.GetTeamRules(c.Request().Context(), teamID, false)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve score rules")
	}

	return c.JSON(http.StatusOK, rules)
}
