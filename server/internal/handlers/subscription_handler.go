package handlers

import (
	"net/http"
	"time"

	"kori/internal/events"
	"kori/internal/models"
	"kori/internal/utils"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type SubscriptionHandler struct {
	db *gorm.DB
}

func NewSubscriptionHandler(db *gorm.DB) *SubscriptionHandler {
	return &SubscriptionHandler{db: db}
}

type CreateSubscriptionRequest struct {
	ProductID string `json:"product_id" validate:"required"`
}

type GetPlansResponse struct {
	Plans []models.Product `json:"plans"`
}

type GetPlanResponse struct {
	Plan models.Product `json:"plan"`
}

// GetPlans returns all available plans
// @Summary Get all plans
// @Description Get all available plans
// @Tags plans
// @Accept json
// @Produce json
// @Success 200 {object} GetPlansResponse "Plans"
// @Router /plans [get]
func (h *SubscriptionHandler) GetPlans(c echo.Context) error {
	var plans []models.Product
	if err := h.db.Where("is_active = ?", true).Preload("Features").Find(&plans).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get plans"})
	}

	return c.JSON(http.StatusOK, GetPlansResponse{
		Plans: plans,
	})
}

// CreateSubscription initiates a subscription purchase
// @Summary Create a new subscription
// @Description Create a pending subscription and redirect to payment
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param request body CreateSubscriptionRequest true "Subscription details"
// @Success 200 {object} map[string]string "Payment URL"
// @Failure 400 {object} map[string]string "Validation error"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /subscriptions [post]
func (h *SubscriptionHandler) CreateSubscription(c echo.Context) error {
	teamID := c.Get("teamID").(string)
	userID := c.Get("userID").(string)

	// check if user is team admin or owner
	var user models.User
	if err := h.db.Where("id = ? AND team_id = ? AND role IN ?",
		userID, teamID, []models.UserRole{models.UserRoleAdmin, models.UserRoleSuperAdmin}).First(&user).Error; err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Not authorized"})
	}

	var req CreateSubscriptionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Start transaction
	tx := h.db.Begin()
	if tx.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction"})
	}

	// Get product details
	var product models.Product
	if err := tx.Preload("Features").First(&product, "id = ?", req.ProductID).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid product"})
	}

	// Create Dodo customer
	dodoCustomer, err := utils.CreateDodoCustomer(user.Email)
	if err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create customer"})
	}

	// Create subscription in Dodo
	dodoSub, err := utils.CreateDodoSubscription(dodoCustomer.ID, product.DodoID)
	if err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create subscription"})
	}

	// Create pending subscription
	subscription := models.Subscription{
		ProductID:          product.ID,
		Status:             models.SubscriptionStatusPending,
		Email:              user.Email,
		DodoCustomerID:     dodoCustomer.ID,
		DodoSubscriptionID: dodoSub.ID,
		TeamID:             teamID,
	}

	if err := tx.Create(&subscription).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create subscription"})
	}

	if err := tx.Commit().Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"payment_url": dodoSub.PaymentLink,
	})
}

// GetSubscription returns the current subscription for the team
// @Summary Get current subscription
// @Description Get the current subscription for the team
// @Tags subscriptions
// @Accept json
// @Produce json
// @Success 200 {object} models.Subscription "Subscription"
// @Failure 403 {object} map[string]string "Not authorized"
// @Router /subscriptions [get]
func (h *SubscriptionHandler) GetSubscription(c echo.Context) error {
	teamID := c.Get("teamID").(string)

	var subscription models.Subscription
	if err := h.db.Where("team_id = ?", teamID).First(&subscription).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No active subscription"})
	}

	return c.JSON(http.StatusOK, subscription)
}

// WebhookPayload represents the webhook payload from Dodo Payments
type WebhookPayload struct {
	Type       string `json:"type"`
	CustomerID string `json:"customer_id"`
	Data       struct {
		SubscriptionID string    `json:"subscription_id"`
		Status         string    `json:"status"`
		StartDate      time.Time `json:"start_date"`
		EndDate        time.Time `json:"end_date"`
	} `json:"data"`
}

// HandleWebhook processes subscription webhooks from Dodo Payments
// @Summary Handle subscription webhooks
// @Description Process subscription status updates from Dodo Payments
// @Tags subscriptions
// @Accept json
// @Success 200 {object} map[string]string "Success"
// @Failure 400 {object} map[string]string "Invalid webhook"
// @Router /subscriptions/webhook [post]
func (h *SubscriptionHandler) HandleWebhook(c echo.Context) error {
	var payload WebhookPayload
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid webhook payload"})
	}

	// Verify webhook signature
	if err := utils.VerifyDodoWebhook(c.Request()); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid webhook signature"})
	}

	tx := h.db.Begin()
	if tx.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction"})
	}

	var subscription models.Subscription
	if err := tx.Where("dodo_subscription_id = ?", payload.Data.SubscriptionID).First(&subscription).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Subscription not found"})
	}

	switch payload.Type {
	case "subscription.activated":
		subscription.Status = models.SubscriptionStatusActive
		subscription.CurrentPeriodStart = payload.Data.StartDate
		subscription.CurrentPeriodEnd = payload.Data.EndDate
	case "subscription.canceled":
		subscription.Status = models.SubscriptionStatusCanceled
		now := time.Now()
		subscription.CanceledAt = &now
	case "subscription.failed":
		subscription.Status = models.SubscriptionStatusFailed
	}

	if err := tx.Save(&subscription).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update subscription"})
	}

	if err := tx.Commit().Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction"})
	}

	events.Emit("subscription.updated", &subscription)

	return c.JSON(http.StatusOK, map[string]string{"status": "success"})
}

// GetManagementPortal generates a URL for subscription management
// @Summary Get subscription management portal URL
// @Description Generate a URL for managing subscription settings
// @Tags subscriptions
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Portal URL"
// @Failure 403 {object} map[string]string "Not authorized"
// @Router /subscriptions/portal [get]
func (h *SubscriptionHandler) GetManagementPortal(c echo.Context) error {
	userID := c.Get("userID").(string)
	teamID := c.Get("teamID").(string)

	// Check if user is team admin or owner
	var user models.User
	if err := h.db.Where("id = ? AND team_id = ? AND role IN ?",
		userID, teamID, []models.UserRole{models.UserRoleAdmin}).First(&user).Error; err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Not authorized"})
	}

	// Get subscription
	var subscription models.Subscription
	if err := h.db.Where("team_id = ?", teamID).First(&subscription).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No active subscription"})
	}

	// Generate portal URL
	portalURL, err := utils.CreateDodoPortalSession(subscription.DodoCustomerID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create portal session"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"portal_url": portalURL,
	})
}

// GetFeatures returns the features enabled for the current team
// @Summary Get enabled features
// @Description Get list of features enabled for the team's subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Features"
// @Router /subscriptions/features [get]
func (h *SubscriptionHandler) GetFeatures(c echo.Context) error {
	teamID := c.Get("teamID").(string)

	var subscription models.Subscription
	if err := h.db.Preload("Product.Features").Where("team_id = ?", teamID).First(&subscription).Error; err != nil {
		// Return free tier features if no subscription found
		return c.JSON(http.StatusOK, map[string]interface{}{
			"features": getFreeTierFeatures(),
		})
	}

	features := make(map[string]interface{})
	for _, feature := range subscription.Product.Features {
		features[string(feature.Feature)] = map[string]interface{}{
			"enabled": feature.Enabled,
			"limit":   feature.Limit,
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"features": features,
	})
}

func getFreeTierFeatures() map[string]interface{} {
	return map[string]interface{}{
		string(models.FeatureEmailCampaigns): map[string]interface{}{
			"enabled": true,
			"limit":   100, // Free tier limit
		},
		string(models.FeatureTemplateLibrary): map[string]interface{}{
			"enabled": true,
			"limit":   5, // Free tier limit
		},
	}
}
