package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mailxem/payments.go/internal/config"
	"github.com/mailxem/payments.go/internal/models"
	"github.com/mailxem/payments.go/internal/services"
	"github.com/mailxem/payments.go/internal/structs"
	"gorm.io/gorm"
)

type SubscriptionHandlers struct {
	subscriptionService services.SubscriptionService
	paymentService      services.PaymentService
	webhookService      services.WebhookService
	planService         services.PlanService
	cfg                 *config.Config
}

func NewSubscriptionHandlers(
	subscriptionService services.SubscriptionService,
	paymentService services.PaymentService,
	webhookService services.WebhookService,
	planService services.PlanService,
	cfg *config.Config,
) *SubscriptionHandlers {
	return &SubscriptionHandlers{
		subscriptionService: subscriptionService,
		paymentService:      paymentService,
		webhookService:      webhookService,
		planService:         planService,
		cfg:                 cfg,
	}
}

// GetPlans returns all available subscription plans
// @Summary Get available subscription plans (legacy)
// @Description Get all available subscription plans. This is a legacy endpoint, use /api/plans instead.
// @Tags subscriptions
// @Produce json
// @Success 200 {object} map[string]interface{} "Available plans"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/plans [get]
// @Deprecated
func (h *SubscriptionHandlers) GetPlans(c echo.Context) error {
	plans, err := h.planService.ListPlans(c.Request().Context(), false)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to get plans",
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"plans": plans,
	})
}

// GetSubscription returns subscription details for a team
// @Summary Get team subscription
// @Description Get subscription details for a specific team by team ID
// @Tags subscriptions
// @Produce json
// @Param teamId path string true "Team ID"
// @Success 200 {object} models.Subscription "Subscription details"
// @Failure 400 {object} map[string]string "Bad request - missing team ID"
// @Failure 404 {object} map[string]string "Subscription not found"
// @Router /api/v1/subscriptions/team/{teamId} [get]
func (h *SubscriptionHandlers) GetSubscription(c echo.Context) error {
	teamID := c.Param("teamId")
	if teamID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "team_id is required",
		})
	}

	subscription, err := h.subscriptionService.GetSubscriptionByTeamID(c.Request().Context(), teamID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// create a free subscription
			subscription, err = h.subscriptionService.CreateSubscription(c.Request().Context(), teamID, h.cfg.Plans.FreePlanID, 1)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": err.Error(),
				})
			}
			return c.JSON(http.StatusOK, subscription)
		}
	}

	return c.JSON(http.StatusOK, subscription)
}

// UpdateSubscriptionRequest represents the request to update a subscription
type UpdateSubscriptionRequest struct {
	PlanID string `json:"plan_id" example:"syne_scale"`
	Seats  int    `json:"seats" validate:"min=1" example:"5"`
}

// UpdateSubscription updates an existing subscription
// @Summary Update subscription
// @Description Update an existing subscription with new plan or seat count. When the current subscription is cancelled, expired, failed, paused, or otherwise inactive, the old subscription is removed and a new one is created (free plan) or checkout is required (paid plan).
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Param subscription body UpdateSubscriptionRequest true "Subscription update request"
// @Success 200 {object} models.Subscription "Updated subscription"
// @Success 202 {object} services.ChangePlanResult "Checkout required for paid plan (when previous subscription was inactive)"
// @Failure 400 {object} map[string]string "Bad request - validation error"
// @Failure 404 {object} map[string]string "Subscription not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/subscriptions/{id} [put]
func (h *SubscriptionHandlers) UpdateSubscription(c echo.Context) error {
	subscriptionID := c.Param("id")
	if subscriptionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "subscription_id is required",
		})
	}

	var req UpdateSubscriptionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	// At least one of plan_id or seats must be provided
	if req.PlanID == "" && req.Seats <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "at least one of plan_id or seats is required",
		})
	}

	result, err := h.subscriptionService.ChangePlan(
		c.Request().Context(),
		subscriptionID,
		req.PlanID,
		req.Seats,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// When checkout is required (inactive subscription + paid plan), return 202 with checkout info
	if result.CheckoutRequired != nil {
		return c.JSON(http.StatusAccepted, result)
	}

	return c.JSON(http.StatusOK, result.Subscription)
}

// CancelSubscription cancels a subscription
// @Summary Cancel subscription
// @Description Cancel an active subscription
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} map[string]string "Cancellation confirmation"
// @Failure 400 {object} map[string]string "Bad request - missing subscription ID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/subscriptions/{id} [delete]
func (h *SubscriptionHandlers) CancelSubscription(c echo.Context) error {
	subscriptionID := c.Param("id")
	if subscriptionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "subscription_id is required",
		})
	}

	err := h.subscriptionService.CancelSubscription(c.Request().Context(), subscriptionID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "subscription cancelled successfully",
	})
}

// GetUsageStats returns usage statistics for a subscription
// @Summary Get subscription usage statistics
// @Description Get usage statistics for a specific subscription
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} models.UsageRecord "Usage statistics"
// @Failure 400 {object} map[string]string "Bad request - missing subscription ID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/subscriptions/{id}/usage [get]
func (h *SubscriptionHandlers) GetUsageStats(c echo.Context) error {
	subscriptionID := c.Param("id")
	if subscriptionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "subscription_id is required",
		})
	}

	stats, err := h.subscriptionService.GetUsageStats(c.Request().Context(), subscriptionID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, stats)
}

// CheckUsageLimit checks if team can perform queries
// @Summary Check usage limit
// @Description Check if a team can perform queries based on their subscription limits
// @Tags usage
// @Produce json
// @Param teamId path string true "Team ID"
// @Param feature path string true "Feature"
// @Success 200 {object} map[string]bool "Usage limit check result"
// @Failure 400 {object} map[string]string "Bad request - missing team ID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/usage/check/{teamId}/{feature} [get]
func (h *SubscriptionHandlers) CheckUsageLimit(c echo.Context) error {
	teamID := c.Param("teamId")
	if teamID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "team_id is required",
		})
	}

	feature := c.Param("feature")
	if feature == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "feature is required",
		})
	}

	canQuery, err := h.subscriptionService.CheckUsageLimit(c.Request().Context(), teamID, models.PlanFeature(feature))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"allowed": canQuery,
		"feature": feature,
	})
}

// RecordUsage records query usage for a team
// @Summary Record usage
// @Description Record query usage for a team
// @Tags usage
// @Accept json
// @Produce json
// @Param teamId path string true "Team ID"
// @Param usage body structs.RecordUsageRequest true "Usage recording request"
// @Success 200 {object} map[string]string "Usage recorded successfully"
// @Failure 400 {object} map[string]string "Bad request - validation error"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/usage/record/{teamId} [post]
func (h *SubscriptionHandlers) RecordUsage(c echo.Context) error {
	teamID := c.Param("teamId")
	if teamID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "team_id is required",
		})
	}

	var req structs.RecordUsageRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	err := h.subscriptionService.RecordUsage(c.Request().Context(), teamID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "usage recorded successfully",
	})
}

// CreateCheckoutSession creates a payment checkout session
// @Summary Create payment checkout session
// @Description Create a new payment checkout session for subscription purchase
// @Tags payments
// @Accept json
// @Produce json
// @Param checkout body services.CheckoutSessionRequest true "Checkout session request"
// @Success 201 {object} services.CheckoutSessionResponse "Checkout session created"
// @Failure 400 {object} map[string]string "Bad request - validation error"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/payments/checkout [post]
func (h *SubscriptionHandlers) CreateCheckoutSession(c echo.Context) error {
	var req services.CheckoutSessionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	response, err := h.paymentService.CreateCheckoutSession(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, response)
}

// GetPaymentHistory returns payment history for a subscription
// @Summary Get payment history
// @Description Get payment history for a specific subscription
// @Tags payments
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} map[string][]models.Payment "Payment history"
// @Failure 400 {object} map[string]string "Bad request - missing subscription ID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/subscriptions/{id}/payments [get]
func (h *SubscriptionHandlers) GetPaymentHistory(c echo.Context) error {
	subscriptionID := c.Param("id")
	if subscriptionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "subscription_id is required",
		})
	}

	payments, err := h.paymentService.GetPaymentHistory(c.Request().Context(), subscriptionID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"payments": payments,
	})
}

// GetInvoicePaymentPDF returns a payment PDF for a specific invoice
// @Summary Get invoice payment PDF
// @Description Get a payment PDF for a specific invoice ID
// @Tags payments
// @Produce json
// @Param payment_id path string true "Payment ID"
// @Success 200 {file} []byte "Invoice payment PDF"
// @Failure 400 {object} map[string]string "Bad request - missing invoice ID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/subscriptions/{id}/payments/{payment_id}/invoice [get]
func (h *SubscriptionHandlers) GetInvoicePaymentPDF(c echo.Context) error {
	subscriptionID := c.Param("id")
	if subscriptionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "subscription_id is required",
		})
	}

	paymentID := c.Param("payment_id")
	if paymentID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "payment_id is required",
		})
	}

	paymentInvoicePDF, err := h.paymentService.GetPaymentInvoicePDF(c.Request().Context(), paymentID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=invoice-payment-%s.pdf", paymentID))
	return c.Blob(http.StatusOK, "application/pdf", paymentInvoicePDF)
}

// GetCustomerPortalSession returns a customer portal session
// @Summary Get customer portal session
// @Description Get a customer portal session for a specific subscription
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} map[string]string "Customer portal session"
// @Failure 400 {object} map[string]string "Bad request - missing subscription ID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/subscriptions/{id}/session [get]
func (h *SubscriptionHandlers) GetCustomerPortalSession(c echo.Context) error {
	subscriptionID := c.Param("id")
	if subscriptionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "subscription_id is required",
		})
	}

	session, err := h.subscriptionService.GetCustomerPortalSession(c.Request().Context(), subscriptionID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"session": session,
	})
}

// RunMaintenance triggers daily maintenance tasks manually (admin only)
// @Summary Run daily maintenance
// @Description Manually trigger daily maintenance to expire overdue trials and reset usage
// @Tags admin
// @Security AdminAuth
// @Produce json
// @Success 200 {object} map[string]string "Maintenance run status"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/admin/maintenance/run [post]
func (h *SubscriptionHandlers) RunMaintenance(c echo.Context) error {
	if err := h.subscriptionService.RunDailyMaintenance(c.Request().Context()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

// Health check endpoint
// @Summary Health check
// @Description Check the health status of the payments service
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{} "Service health status"
// @Router /api/v1/health [get]
func (h *SubscriptionHandlers) Health(c echo.Context) error {
	timestampStr := strconv.FormatInt(time.Now().Unix(), 10)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"service":   "payments",
		"timestamp": timestampStr,
	})
}

// CreateBillingDetails creates a billing details record
// @Summary Create billing details
// @Description Create a billing details record
// @Tags billing
// @Accept json
// @Produce json
// @Param billingDetails body models.BillingDetails true "Billing details"
// @Success 201 {object} models.BillingDetails "Billing details created"
// @Failure 400 {object} map[string]string "Bad request - validation error"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/billing/details [post]
func (h *SubscriptionHandlers) CreateBillingDetails(c echo.Context) error {
	var req models.BillingDetails
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	existingBillingDetails, _ := h.subscriptionService.GetBillingDetails(c.Request().Context(), req.TeamID)
	if existingBillingDetails != nil {
		return c.JSON(http.StatusOK, existingBillingDetails)
	}

	err := h.subscriptionService.CreateBillingDetails(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, req)
}

// GetBillingDetails returns a billing details record
// @Summary Get billing details
// @Description Get a billing details record
// @Tags billing
// @Produce json
// @Param teamId path string true "Team ID"
// @Success 200 {object} models.BillingDetails "Billing details"
// @Failure 400 {object} map[string]string "Bad request - missing billing details ID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/billing/details/{teamId} [get]
func (h *SubscriptionHandlers) GetBillingDetails(c echo.Context) error {
	teamID := c.Param("teamId")
	if teamID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "team_id is required",
		})
	}

	billingDetails, err := h.subscriptionService.GetBillingDetails(c.Request().Context(), teamID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, billingDetails)
}

// UpdateBillingDetails updates a billing details record
// @Summary Update billing details
// @Description Update a billing details record
// @Tags billing
// @Accept json
// @Produce json
// @Param id path string true "Billing details ID"
// @Param billingDetails body models.BillingDetails true "Billing details"
// @Success 200 {object} models.BillingDetails "Billing details updated"
// @Failure 400 {object} map[string]string "Bad request - validation error"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/billing/details/{id} [put]
func (h *SubscriptionHandlers) UpdateBillingDetails(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "billing_details_id is required",
		})
	}

	var req models.BillingDetails
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	err := h.subscriptionService.UpdateBillingDetails(c.Request().Context(), id, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, req)
}
