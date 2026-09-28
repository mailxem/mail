package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mailxem/payments.go/internal/models"
	"github.com/mailxem/payments.go/internal/services"
)

type PlanHandler struct {
	planService services.PlanService
	adminSecret string
}

func NewPlanHandler(planService services.PlanService, adminSecret string) *PlanHandler {
	return &PlanHandler{
		planService: planService,
		adminSecret: adminSecret,
	}
}

// Middleware to check admin authentication
func (h *PlanHandler) requireAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Check for admin secret in header
		providedSecret := c.Request().Header.Get("X-Admin-Secret")
		if providedSecret == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Admin secret required in X-Admin-Secret header",
			})
		}

		if providedSecret != h.adminSecret {
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "Invalid admin secret",
			})
		}

		return next(c)
	}
}

// CreatePlan creates a new plan (admin only)
// @Summary Create a new subscription plan
// @Description Create a new subscription plan with comprehensive configuration options. Requires admin authentication.
// @Tags plans
// @Accept json
// @Produce json
// @Security AdminAuth
// @Param X-Admin-Secret header string true "Admin secret for authentication"
// @Param plan body models.CreatePlanRequest true "Plan creation request"
// @Success 201 {object} map[string]models.PlanResponse "Plan created successfully"
// @Failure 400 {object} map[string]string "Bad request - validation error"
// @Failure 401 {object} map[string]string "Unauthorized - missing admin secret"
// @Failure 403 {object} map[string]string "Forbidden - invalid admin secret"
// @Router /api/admin/plans [post]
func (h *PlanHandler) CreatePlan(c echo.Context) error {
	ctx := c.Request().Context()

	var req models.CreatePlanRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	plan, err := h.planService.CreatePlan(ctx, &req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"plan": plan,
	})
}

// UpdatePlan updates an existing plan (admin only)
// @Summary Update an existing subscription plan
// @Description Update an existing subscription plan. All fields are optional for partial updates. Requires admin authentication.
// @Tags plans
// @Accept json
// @Produce json
// @Security AdminAuth
// @Param X-Admin-Secret header string true "Admin secret for authentication"
// @Param planId path string true "Plan ID"
// @Param plan body models.UpdatePlanRequest true "Plan update request"
// @Success 200 {object} map[string]models.PlanResponse "Plan updated successfully"
// @Failure 400 {object} map[string]string "Bad request - validation error"
// @Failure 401 {object} map[string]string "Unauthorized - missing admin secret"
// @Failure 403 {object} map[string]string "Forbidden - invalid admin secret"
// @Failure 404 {object} map[string]string "Plan not found"
// @Router /api/admin/plans/{planId} [put]
func (h *PlanHandler) UpdatePlan(c echo.Context) error {
	ctx := c.Request().Context()
	planID := c.Param("planId")

	var req models.UpdatePlanRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	plan, err := h.planService.UpdatePlan(ctx, planID, &req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"plan": plan,
	})
}

// GetPlan retrieves a specific plan (public)
// @Summary Get a subscription plan by ID
// @Description Retrieve a specific subscription plan by its ID. Available to all users.
// @Tags plans
// @Produce json
// @Param planId path string true "Plan ID"
// @Success 200 {object} map[string]models.PlanResponse "Plan details"
// @Failure 404 {object} map[string]string "Plan not found"
// @Router /api/plans/{planId} [get]
func (h *PlanHandler) GetPlan(c echo.Context) error {
	ctx := c.Request().Context()
	planID := c.Param("planId")

	plan, err := h.planService.GetPlan(ctx, planID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Plan not found",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"plan": plan,
	})
}

// GetPlanByDodoID retrieves a plan by Dodo Plan ID (public)
// @Summary Get a subscription plan by Dodo Plan ID
// @Description Retrieve a specific subscription plan by its Dodo Payments plan ID. Available to all users.
// @Tags plans
// @Produce json
// @Param dodoPlanId path string true "Dodo Plan ID"
// @Success 200 {object} map[string]models.PlanResponse "Plan details"
// @Failure 404 {object} map[string]string "Plan not found"
// @Router /api/plans/dodo/{dodoPlanId} [get]
func (h *PlanHandler) GetPlanByDodoID(c echo.Context) error {
	ctx := c.Request().Context()
	dodoPlanID := c.Param("dodoPlanId")

	plan, err := h.planService.GetPlanByDodoID(ctx, dodoPlanID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Plan not found",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"plan": plan,
	})
}

// ListPlans lists all plans (admin only - includes inactive)
// @Summary List all subscription plans (admin)
// @Description List all subscription plans including inactive ones. Requires admin authentication.
// @Tags plans
// @Produce json
// @Security AdminAuth
// @Param X-Admin-Secret header string true "Admin secret for authentication"
// @Param include_inactive query boolean false "Include inactive plans in the response"
// @Success 200 {object} map[string][]models.PlanResponse "List of plans with subscription counts"
// @Failure 401 {object} map[string]string "Unauthorized - missing admin secret"
// @Failure 403 {object} map[string]string "Forbidden - invalid admin secret"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/admin/plans [get]
func (h *PlanHandler) ListPlans(c echo.Context) error {
	ctx := c.Request().Context()

	// Check if we should include inactive plans
	includeInactive := c.QueryParam("include_inactive") == "true"

	plans, err := h.planService.ListPlans(ctx, includeInactive)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve plans",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"plans": plans,
	})
}

// ListPublicPlans lists active plans (public)
// @Summary List active subscription plans (public)
// @Description List all active subscription plans available for customer subscriptions. Available to all users.
// @Tags plans
// @Produce json
// @Success 200 {object} map[string][]models.PlanResponse "List of active plans"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/plans [get]
func (h *PlanHandler) ListPublicPlans(c echo.Context) error {
	ctx := c.Request().Context()

	plans, err := h.planService.ListPublicPlans(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve plans",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"plans": plans,
	})
}

// GetPlanWithStats retrieves a plan with subscription statistics (admin only)
// @Summary Get plan with subscription statistics
// @Description Retrieve a specific plan with detailed subscription statistics. Requires admin authentication.
// @Tags plans
// @Produce json
// @Security AdminAuth
// @Param X-Admin-Secret header string true "Admin secret for authentication"
// @Param planId path string true "Plan ID"
// @Success 200 {object} map[string]models.PlanResponse "Plan details with statistics"
// @Failure 401 {object} map[string]string "Unauthorized - missing admin secret"
// @Failure 403 {object} map[string]string "Forbidden - invalid admin secret"
// @Failure 404 {object} map[string]string "Plan not found"
// @Router /api/admin/plans/{planId}/stats [get]
func (h *PlanHandler) GetPlanWithStats(c echo.Context) error {
	ctx := c.Request().Context()
	planID := c.Param("planId")

	plan, err := h.planService.GetPlanWithStats(ctx, planID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Plan not found",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"plan": plan,
	})
}

// DeletePlan deletes a plan (admin only)
// @Summary Delete a subscription plan
// @Description Delete a subscription plan. Plans with active subscriptions cannot be deleted. Requires admin authentication.
// @Tags plans
// @Produce json
// @Security AdminAuth
// @Param X-Admin-Secret header string true "Admin secret for authentication"
// @Param planId path string true "Plan ID"
// @Success 200 {object} map[string]string "Plan deleted successfully"
// @Failure 400 {object} map[string]string "Bad request - plan has active subscriptions"
// @Failure 401 {object} map[string]string "Unauthorized - missing admin secret"
// @Failure 403 {object} map[string]string "Forbidden - invalid admin secret"
// @Failure 404 {object} map[string]string "Plan not found"
// @Router /api/admin/plans/{planId} [delete]
func (h *PlanHandler) DeletePlan(c echo.Context) error {
	ctx := c.Request().Context()
	planID := c.Param("planId")

	err := h.planService.DeletePlan(ctx, planID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Plan deleted successfully",
	})
}

// ArchivePlan archives a plan (admin only)
// @Summary Archive a subscription plan
// @Description Archive a subscription plan. Archived plans are hidden from customers but existing subscriptions continue. Requires admin authentication.
// @Tags plans
// @Produce json
// @Security AdminAuth
// @Param X-Admin-Secret header string true "Admin secret for authentication"
// @Param planId path string true "Plan ID"
// @Success 200 {object} map[string]string "Plan archived successfully"
// @Failure 400 {object} map[string]string "Bad request - validation error"
// @Failure 401 {object} map[string]string "Unauthorized - missing admin secret"
// @Failure 403 {object} map[string]string "Forbidden - invalid admin secret"
// @Failure 404 {object} map[string]string "Plan not found"
// @Router /api/admin/plans/{planId}/archive [post]
func (h *PlanHandler) ArchivePlan(c echo.Context) error {
	ctx := c.Request().Context()
	planID := c.Param("planId")

	err := h.planService.ArchivePlan(ctx, planID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Plan archived successfully",
	})
}

// RegisterRoutes registers all plan routes
func (h *PlanHandler) RegisterRoutes(group *echo.Group) {
	// Public routes (no authentication required)
	publicGroup := group.Group("/plans")
	publicGroup.GET("/:planId", h.GetPlan)
	publicGroup.GET("/dodo/:dodoPlanId", h.GetPlanByDodoID)

	// Admin routes (require admin secret)
	adminGroup := group.Group("/admin/plans")
	adminGroup.Use(h.requireAdmin)
	adminGroup.POST("", h.CreatePlan)
	adminGroup.PUT("/:planId", h.UpdatePlan)
	adminGroup.GET("", h.ListPlans)
	adminGroup.GET("/:planId/stats", h.GetPlanWithStats)
	adminGroup.DELETE("/:planId", h.DeletePlan)
	adminGroup.POST("/:planId/archive", h.ArchivePlan)
}
