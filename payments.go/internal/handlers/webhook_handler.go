package handlers

import (
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/mailxem/payments.go/internal/services"
)

type WebhookHandler struct {
	webhookService services.WebhookService
}

func NewWebhookHandler(webhookService services.WebhookService) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
	}
}

// HandleDodoWebhook processes incoming webhook events from Dodo Payments
// @Summary Process Dodo Payments webhook
// @Description Process incoming webhook events from Dodo Payments for subscription and payment updates
// @Tags webhooks
// @Accept json
// @Produce json
// @Param X-Signature header string true "Webhook signature for verification (alternative: X-Dodo-Signature)"
// @Param webhook body object true "Webhook payload from Dodo Payments"
// @Success 200 {object} map[string]string "Webhook processed successfully"
// @Failure 400 {object} map[string]string "Bad request - missing signature or invalid payload"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /webhooks/dodo [post]
func (h *WebhookHandler) HandleDodoWebhook(c echo.Context) error {
	// Get the signature from headers
	signature := c.Request().Header.Get("webhook-signature")
	webhookID := c.Request().Header.Get("webhook-id")
	webhookTimestamp := c.Request().Header.Get("webhook-timestamp")

	if signature == "" || webhookID == "" || webhookTimestamp == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing webhook signature, ID, or timestamp",
		})
	}

	c.Request().Header.Del("content-length")

	// jhson
	rawBody := c.Request().Body

	body, err := io.ReadAll(rawBody)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to read request body",
		})
	}

	// Process the webhook using the comprehensive webhook service
	ctx := c.Request().Context()

	if err := h.webhookService.ProcessWebhook(ctx, body, signature, webhookID, webhookTimestamp); err != nil {
		// Log the error but return success to prevent retries for invalid webhooks
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to process webhook" + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Webhook processed successfully",
	})
}

// GetWebhookEvents retrieves webhook events with optional filtering
// @Summary Get webhook events
// @Description Retrieve webhook events with optional filtering by type and pagination
// @Tags webhooks
// @Produce json
// @Param type query string false "Filter by event type (e.g., payment.succeeded, subscription.active)"
// @Param limit query integer false "Maximum number of events to return (max: 100)" default(50)
// @Param offset query integer false "Number of events to skip for pagination" default(0)
// @Success 200 {object} map[string]interface{} "List of webhook events with pagination info"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/webhooks/events [get]
func (h *WebhookHandler) GetWebhookEvents(c echo.Context) error {
	ctx := c.Request().Context()

	// Parse query parameters
	eventType := c.QueryParam("type")
	limitStr := c.QueryParam("limit")
	offsetStr := c.QueryParam("offset")

	// Set defaults
	limit := 50
	offset := 0

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil {
			offset = parsedOffset
		}
	}

	// Limit the maximum number of events returned
	if limit > 100 {
		limit = 100
	}

	events, err := h.webhookService.GetWebhookEvents(ctx, eventType, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve webhook events",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"events": events,
		"pagination": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"count":  len(events),
		},
	})
}

// GetWebhookLogs retrieves webhook processing logs for debugging
// @Summary Get webhook logs
// @Description Retrieve webhook processing logs for debugging and monitoring purposes
// @Tags webhooks
// @Produce json
// @Param event_id query string false "Filter logs by specific webhook event ID"
// @Param status query string false "Filter logs by processing status (success, failed, pending)"
// @Param limit query integer false "Maximum number of logs to return (max: 100)" default(50)
// @Success 200 {object} map[string]interface{} "List of webhook processing logs"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/webhooks/logs [get]
func (h *WebhookHandler) GetWebhookLogs(c echo.Context) error {
	ctx := c.Request().Context()
	eventID := c.Param("eventId")

	if eventID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Event ID is required",
		})
	}

	logs, err := h.webhookService.GetWebhookLogs(ctx, eventID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve webhook logs",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"logs": logs,
	})
}

// RetryFailedWebhooks retries processing of failed webhooks
// @Summary Retry failed webhooks
// @Description Retry processing of failed webhook events within a specified time range
// @Tags webhooks
// @Accept json
// @Produce json
// @Param retry body map[string]interface{} false "Optional parameters for retry operation (hours_back, event_types)"
// @Success 200 {object} map[string]interface{} "Retry operation result with counts"
// @Failure 400 {object} map[string]string "Bad request - invalid parameters"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/webhooks/retry [post]
func (h *WebhookHandler) RetryFailedWebhooks(c echo.Context) error {
	ctx := c.Request().Context()

	limitStr := c.QueryParam("limit")
	limit := 10 // Default limit for retries

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	// Limit the maximum number of retries
	if limit > 50 {
		limit = 50
	}

	err := h.webhookService.RetryFailedEvents(ctx, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retry webhook events",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Failed webhook events have been retried",
	})
}

// CleanupOldWebhooks removes old webhook events and logs
// @Summary Cleanup old webhook data
// @Description Remove old webhook events and logs to free up storage space
// @Tags webhooks
// @Accept json
// @Produce json
// @Param cleanup body map[string]interface{} false "Cleanup parameters (days_to_keep, event_types)"
// @Success 200 {object} map[string]interface{} "Cleanup operation result with counts"
// @Failure 400 {object} map[string]string "Bad request - invalid parameters"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/webhooks/cleanup [post]
func (h *WebhookHandler) CleanupOldWebhooks(c echo.Context) error {
	ctx := c.Request().Context()

	retentionDaysStr := c.QueryParam("retention_days")
	retentionDays := 30 // Default retention period

	if retentionDaysStr != "" {
		if parsedDays, err := strconv.Atoi(retentionDaysStr); err == nil {
			retentionDays = parsedDays
		}
	}

	// Minimum retention period of 7 days
	if retentionDays < 7 {
		retentionDays = 7
	}

	err := h.webhookService.CleanupOldEvents(ctx, retentionDays)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to cleanup old webhook events",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":         "success",
		"message":        "Old webhook events have been cleaned up",
		"retention_days": retentionDays,
	})
}
