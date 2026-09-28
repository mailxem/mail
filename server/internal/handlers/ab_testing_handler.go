package handlers

import (
	"kori/internal/models"
	"kori/internal/services"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type ABTestingHandler struct {
	testService *services.ABTestService
}

func NewABTestingHandler(db *gorm.DB) *ABTestingHandler {
	return &ABTestingHandler{
		testService: services.NewABTestService(db),
	}
}

// CreateTest creates a new A/B test
// @Summary Create A/B test
// @Description Create a new A/B test
// @Tags A/B Testing
// @Accept json
// @Produce json
// @Param body body models.ABTest true "Test data"
// @Success 201 {object} models.ABTest
// @Router /api/v1/ab-tests [post]
// @Security ApiKeyAuth
func (h *ABTestingHandler) CreateTest(c echo.Context) error {
	teamID := c.Get("teamID").(string)

	var test models.ABTest
	if err := c.Bind(&test); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	test.TeamID = teamID
	test.Status = models.TestStatusDraft

	if err := h.testService.Create(c.Request().Context(), &test); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create test")
	}

	return c.JSON(http.StatusCreated, test)
}

// ListTests retrieves all A/B tests for a team
// @Summary List A/B tests
// @Description Get all A/B tests for the team
// @Tags A/B Testing
// @Accept json
// @Produce json
// @Param status query string false "Filter by status"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} object
// @Router /api/v1/ab-tests [get]
// @Security ApiKeyAuth
func (h *ABTestingHandler) ListTests(c echo.Context) error {
	teamID := c.Get("teamID").(string)
	status := models.TestStatus(c.QueryParam("status"))

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 {
		limit = 20
	}

	tests, total, err := h.testService.List(c.Request().Context(), teamID, status, page, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list tests")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"tests": tests,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetTest retrieves a single A/B test
// @Summary Get A/B test
// @Description Get details of a specific A/B test
// @Tags A/B Testing
// @Accept json
// @Produce json
// @Param id path string true "Test ID"
// @Success 200 {object} models.ABTest
// @Router /api/v1/ab-tests/{id} [get]
// @Security ApiKeyAuth
func (h *ABTestingHandler) GetTest(c echo.Context) error {
	id := c.Param("id")

	test, err := h.testService.Get(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Test not found")
	}

	return c.JSON(http.StatusOK, test)
}

// UpdateTest updates an A/B test
// @Summary Update A/B test
// @Description Update an existing A/B test
// @Tags A/B Testing
// @Accept json
// @Produce json
// @Param id path string true "Test ID"
// @Param body body object true "Update data"
// @Success 200 {object} models.ABTest
// @Router /api/v1/ab-tests/{id} [put]
// @Security ApiKeyAuth
func (h *ABTestingHandler) UpdateTest(c echo.Context) error {
	id := c.Param("id")

	var updates map[string]interface{}
	if err := c.Bind(&updates); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Don't allow changing team_id or status directly
	delete(updates, "team_id")
	delete(updates, "status")

	if err := h.testService.Update(c.Request().Context(), id, updates); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update test")
	}

	test, err := h.testService.Get(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve updated test")
	}

	return c.JSON(http.StatusOK, test)
}

// DeleteTest deletes an A/B test
// @Summary Delete A/B test
// @Description Delete an A/B test and all its data
// @Tags A/B Testing
// @Accept json
// @Produce json
// @Param id path string true "Test ID"
// @Success 204
// @Router /api/v1/ab-tests/{id} [delete]
// @Security ApiKeyAuth
func (h *ABTestingHandler) DeleteTest(c echo.Context) error {
	id := c.Param("id")

	if err := h.testService.Delete(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete test")
	}

	return c.NoContent(http.StatusNoContent)
}

// StartTest starts an A/B test
// @Summary Start A/B test
// @Description Start running an A/B test
// @Tags A/B Testing
// @Accept json
// @Produce json
// @Param id path string true "Test ID"
// @Success 200 {object} models.ABTest
// @Router /api/v1/ab-tests/{id}/start [post]
// @Security ApiKeyAuth
func (h *ABTestingHandler) StartTest(c echo.Context) error {
	id := c.Param("id")

	if err := h.testService.StartTest(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	test, err := h.testService.Get(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve test")
	}

	return c.JSON(http.StatusOK, test)
}

// StopTest stops an A/B test
// @Summary Stop A/B test
// @Description Stop a running A/B test
// @Tags A/B Testing
// @Accept json
// @Produce json
// @Param id path string true "Test ID"
// @Success 200 {object} models.ABTest
// @Router /api/v1/ab-tests/{id}/stop [post]
// @Security ApiKeyAuth
func (h *ABTestingHandler) StopTest(c echo.Context) error {
	id := c.Param("id")

	if err := h.testService.StopTest(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to stop test")
	}

	test, err := h.testService.Get(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve test")
	}

	return c.JSON(http.StatusOK, test)
}

// GetResults retrieves results for an A/B test
// @Summary Get test results
// @Description Get detailed results for an A/B test
// @Tags A/B Testing
// @Accept json
// @Produce json
// @Param id path string true "Test ID"
// @Success 200 {array} models.TestResult
// @Router /api/v1/ab-tests/{id}/results [get]
// @Security ApiKeyAuth
func (h *ABTestingHandler) GetResults(c echo.Context) error {
	id := c.Param("id")

	results, err := h.testService.GetResults(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve results")
	}

	return c.JSON(http.StatusOK, results)
}

// AnalyzeTest performs statistical analysis on a test
// @Summary Analyze test
// @Description Perform statistical analysis to determine winner
// @Tags A/B Testing
// @Accept json
// @Produce json
// @Param id path string true "Test ID"
// @Success 200 {object} services.TestAnalysis
// @Router /api/v1/ab-tests/{id}/analyze [get]
// @Security ApiKeyAuth
func (h *ABTestingHandler) AnalyzeTest(c echo.Context) error {
	id := c.Param("id")

	analysis, err := h.testService.AnalyzeTest(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, analysis)
}

// DeclareWinner manually declares a winner
// @Summary Declare winner
// @Description Manually declare a winning variant
// @Tags A/B Testing
// @Accept json
// @Produce json
// @Param id path string true "Test ID"
// @Param body body object true "Winner data"
// @Success 200 {object} models.ABTest
// @Router /api/v1/ab-tests/{id}/declare-winner [post]
// @Security ApiKeyAuth
func (h *ABTestingHandler) DeclareWinner(c echo.Context) error {
	id := c.Param("id")

	var req struct {
		VariantID string `json:"variantId" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := h.testService.DeclareWinner(c.Request().Context(), id, req.VariantID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to declare winner")
	}

	test, err := h.testService.Get(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve test")
	}

	return c.JSON(http.StatusOK, test)
}
