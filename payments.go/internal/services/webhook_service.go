package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dodopayments/dodopayments-go"
	"github.com/google/uuid"
	"github.com/mailxem/payments.go/internal/models"
	"github.com/mailxem/payments.go/internal/repositories"
	"github.com/shopspring/decimal"
	standardwebhooks "github.com/standard-webhooks/standard-webhooks/libraries/go"
)

type WebhookService interface {
	ProcessWebhook(ctx context.Context, body []byte, signature string, webhookID, webhookTimestamp string) error
	GetWebhookEvents(ctx context.Context, eventType string, limit int, offset int) ([]*models.WebhookEvent, error)
	GetWebhookLogs(ctx context.Context, eventID string) ([]*models.WebhookLog, error)
	RetryFailedEvents(ctx context.Context, limit int) error
	CleanupOldEvents(ctx context.Context, retentionDays int) error
}

type webhookService struct {
	webhookRepo      repositories.WebhookRepository
	webhookLogRepo   repositories.WebhookLogRepository
	subscriptionRepo repositories.SubscriptionRepository
	paymentRepo      repositories.PaymentRepository
	emailService     EmailService
	webhookSecret    string
}

func NewWebhookService(
	webhookRepo repositories.WebhookRepository,
	webhookLogRepo repositories.WebhookLogRepository,
	subscriptionRepo repositories.SubscriptionRepository,
	paymentRepo repositories.PaymentRepository,
	emailService EmailService,
	webhookSecret string,
) WebhookService {
	return &webhookService{
		webhookRepo:      webhookRepo,
		webhookLogRepo:   webhookLogRepo,
		subscriptionRepo: subscriptionRepo,
		paymentRepo:      paymentRepo,
		emailService:     emailService,
		webhookSecret:    webhookSecret,
	}
}

// ProcessWebhook handles incoming webhook events from Dodo Payments
func (s *webhookService) ProcessWebhook(ctx context.Context, payload []byte, signature string, webhookID, webhookTimestamp string) error {
	// Parse the webhook payload
	var rawEvent map[string]any
	if err := json.Unmarshal(payload, &rawEvent); err != nil {
		return fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	// Extract event type
	eventType, ok := rawEvent["type"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid event type in webhook payload")
	}

	// Validate webhook timestamp to prevent replay attacks
	// Reject webhooks older than 5 minutes
	if err := s.validateWebhookTimestamp(webhookTimestamp); err != nil {
		return fmt.Errorf("webhook timestamp validation failed: %w", err)
	}

	wh, err := standardwebhooks.NewWebhook(s.webhookSecret)
	if err != nil {
		return fmt.Errorf("failed to create webhook: %w", err)
	}

	header := http.Header{}

	header.Set(standardwebhooks.HeaderWebhookID, webhookID)
	header.Set(standardwebhooks.HeaderWebhookTimestamp, webhookTimestamp)
	header.Set(standardwebhooks.HeaderWebhookSignature, signature)

	if err := wh.Verify(payload, header); err != nil {
		return fmt.Errorf("failed to verify webhook: %w", err)
	}

	// Create webhook event record
	event := &models.WebhookEvent{
		ID:        uuid.New().String(),
		Type:      eventType,
		Data:      rawEvent,
		CreatedAt: time.Now(),
		Status:    models.WebhookPending,
		Signature: signature,
		RawBody:   string(payload),
	}

	// Save event to database
	if err := s.webhookRepo.SaveEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to save webhook event: %w", err)
	}

	// Log event reception
	s.logWebhookEvent(ctx, event.ID, models.LogLevelInfo, "Webhook event received", map[string]interface{}{
		"event_type": eventType,
		"signature":  signature,
	})

	// Process the event based on its type
	if err := s.processEventByType(ctx, event); err != nil {
		errorMsg := err.Error()
		s.webhookRepo.UpdateEventStatus(ctx, event.ID, models.WebhookFailed, &errorMsg)
		s.logWebhookEvent(ctx, event.ID, models.LogLevelError, "Failed to process webhook event", map[string]interface{}{
			"error": errorMsg,
		})
		return fmt.Errorf("failed to process webhook event: %w", err)
	}

	// Mark event as processed
	if err := s.webhookRepo.MarkEventProcessed(ctx, event.ID); err != nil {
		s.logWebhookEvent(ctx, event.ID, models.LogLevelWarning, "Failed to mark event as processed", map[string]interface{}{
			"error": err.Error(),
		})
	}

	s.logWebhookEvent(ctx, event.ID, models.LogLevelInfo, "Webhook event processed successfully", nil)
	return nil
}

// processEventByType routes events to appropriate handlers based on event type
func (s *webhookService) processEventByType(ctx context.Context, event *models.WebhookEvent) error {
	s.logWebhookEvent(ctx, event.ID, models.LogLevelDebug, "Processing event by type", map[string]interface{}{
		"event_type": event.Type,
	})

	switch {
	// Payment Events
	case strings.HasPrefix(event.Type, "payment."):
		return s.handlePaymentEvent(ctx, event)

	// Subscription Events
	case strings.HasPrefix(event.Type, "subscription."):
		return s.handleSubscriptionEvent(ctx, event)

	// Refund Events
	case strings.HasPrefix(event.Type, "refund."):
		return s.handleRefundEvent(ctx, event)

	// Dispute Events
	case strings.HasPrefix(event.Type, "dispute."):
		return s.handleDisputeEvent(ctx, event)

	// License Key Events
	case strings.HasPrefix(event.Type, "license_key."):
		return s.handleLicenseKeyEvent(ctx, event)

	default:
		s.logWebhookEvent(ctx, event.ID, models.LogLevelWarning, "Unknown event type", map[string]interface{}{
			"event_type": event.Type,
		})
		// Mark as skipped rather than failed for unknown events
		return s.webhookRepo.UpdateEventStatus(ctx, event.ID, models.WebhookSkipped, nil)
	}
}

// handlePaymentEvent processes payment-related webhook events
func (s *webhookService) handlePaymentEvent(ctx context.Context, event *models.WebhookEvent) error {
	paymentData, err := event.ParsePaymentData()
	if err != nil {
		return fmt.Errorf("failed to parse payment data: %w", err)
	}

	s.logWebhookEvent(ctx, event.ID, models.LogLevelInfo, "Processing payment event", map[string]interface{}{
		"payment_id": paymentData.PaymentID,
		"status":     paymentData.Status,
		"event_type": event.Type,
	})

	switch event.Type {
	case "payment.succeeded":
		return s.handlePaymentSucceeded(ctx, event.ID, paymentData)
	case "payment.failed":
		return s.handlePaymentFailed(ctx, event.ID, paymentData)
	case "payment.processing":
		return s.handlePaymentProcessing(ctx, event.ID, paymentData)
	case "payment.cancelled":
		return s.handlePaymentCancelled(ctx, event.ID, paymentData)
	default:
		return fmt.Errorf("unknown payment event type: %s", event.Type)
	}
}

// handleSubscriptionEvent processes subscription-related webhook events
func (s *webhookService) handleSubscriptionEvent(ctx context.Context, event *models.WebhookEvent) error {
	subscriptionData, err := event.ParseSubscriptionData()
	if err != nil {
		return fmt.Errorf("failed to parse subscription data: %w", err)
	}

	s.logWebhookEvent(ctx, event.ID, models.LogLevelInfo, "Processing subscription event", map[string]interface{}{
		"subscription_id": subscriptionData.SubscriptionID,
		"status":          subscriptionData.Status,
		"event_type":      event.Type,
	})

	subscription, err := s.findSubscriptionByDodoID(ctx, subscriptionData.SubscriptionID)
	if err != nil {
		return fmt.Errorf("failed to find subscription: %w", err)
	}

	go func() {
		time.Sleep(1 * time.Second)
		s.emailService.sendReminderEmail(context.Background(), subscription)
	}()

	switch event.Type {
	case "subscription.active":
		return s.handleSubscriptionActive(ctx, event.ID, subscriptionData)
	case "subscription.on_hold":
		return s.handleSubscriptionOnHold(ctx, event.ID, subscriptionData)
	case "subscription.renewed":
		return s.handleSubscriptionRenewed(ctx, event.ID, subscriptionData)
	case "subscription.paused":
		return s.handleSubscriptionPaused(ctx, event.ID, subscriptionData)
	case "subscription.plan_changed":
		return s.handleSubscriptionPlanChanged(ctx, event.ID, subscriptionData)
	case "subscription.cancelled":
		return s.handleSubscriptionCancelled(ctx, event.ID, subscriptionData)
	case "subscription.failed":
		return s.handleSubscriptionFailed(ctx, event.ID, subscriptionData)
	case "subscription.expired":
		return s.handleSubscriptionExpired(ctx, event.ID, subscriptionData)
	case "subscription.updated":
		return s.handleSubscriptionUpdated(ctx, event.ID, subscriptionData)
	default:
		return fmt.Errorf("unknown subscription event type: %s", event.Type)
	}
}

// handleRefundEvent processes refund-related webhook events
func (s *webhookService) handleRefundEvent(ctx context.Context, event *models.WebhookEvent) error {
	refundData, err := event.ParseRefundData()
	if err != nil {
		return fmt.Errorf("failed to parse refund data: %w", err)
	}

	s.logWebhookEvent(ctx, event.ID, models.LogLevelInfo, "Processing refund event", map[string]interface{}{
		"refund_id":  refundData.RefundID,
		"payment_id": refundData.PaymentID,
		"status":     refundData.Status,
		"event_type": event.Type,
	})

	switch event.Type {
	case "refund.succeeded":
		return s.handleRefundSucceeded(ctx, event.ID, refundData)
	case "refund.failed":
		return s.handleRefundFailed(ctx, event.ID, refundData)
	default:
		return fmt.Errorf("unknown refund event type: %s", event.Type)
	}
}

// handleDisputeEvent processes dispute-related webhook events
func (s *webhookService) handleDisputeEvent(ctx context.Context, event *models.WebhookEvent) error {
	disputeData, err := event.ParseDisputeData()
	if err != nil {
		return fmt.Errorf("failed to parse dispute data: %w", err)
	}

	s.logWebhookEvent(ctx, event.ID, models.LogLevelInfo, "Processing dispute event", map[string]interface{}{
		"dispute_id": disputeData.DisputeID,
		"payment_id": disputeData.PaymentID,
		"status":     disputeData.DisputeStatus,
		"event_type": event.Type,
	})

	// For now, we'll just log dispute events
	// In a full implementation, you might want to update dispute records
	s.logWebhookEvent(ctx, event.ID, models.LogLevelInfo, "Dispute event logged", map[string]interface{}{
		"dispute_stage": disputeData.DisputeStage,
		"amount":        disputeData.Amount,
		"currency":      disputeData.Currency,
	})

	return nil
}

// handleLicenseKeyEvent processes license key-related webhook events
func (s *webhookService) handleLicenseKeyEvent(ctx context.Context, event *models.WebhookEvent) error {
	licenseKeyData, err := event.ParseLicenseKeyData()
	if err != nil {
		return fmt.Errorf("failed to parse license key data: %w", err)
	}

	s.logWebhookEvent(ctx, event.ID, models.LogLevelInfo, "Processing license key event", map[string]interface{}{
		"license_key_id": licenseKeyData.LicenseKeyID,
		"product_id":     licenseKeyData.ProductID,
		"customer_id":    licenseKeyData.CustomerID,
		"event_type":     event.Type,
	})

	// For now, we'll just log license key events
	// In a full implementation, you might want to store license keys
	return nil
}

// Payment event handlers
func (s *webhookService) handlePaymentSucceeded(ctx context.Context, eventID string, data *models.PaymentWebhookData) error {
	// Update payment status to completed if we have a local payment record
	if err := s.updatePaymentStatus(ctx, data.PaymentID, data, models.PaymentCompleted); err != nil {
		return err
	}

	s.logWebhookEvent(ctx, eventID, models.LogLevelInfo, "Payment succeeded", map[string]interface{}{
		"payment_id": data.PaymentID,
		"amount":     data.TotalAmount,
		"currency":   data.Currency,
	})
	return nil
}

func (s *webhookService) handlePaymentFailed(ctx context.Context, eventID string, data *models.PaymentWebhookData) error {
	if err := s.updatePaymentStatus(ctx, data.PaymentID, data, models.PaymentFailed); err != nil {
		// Log but don't fail webhook processing for status update errors
		s.logWebhookEvent(ctx, eventID, models.LogLevelWarning, "Failed to update payment status", map[string]interface{}{
			"payment_id": data.PaymentID,
			"error":      err.Error(),
		})
	}

	s.logWebhookEvent(ctx, eventID, models.LogLevelWarning, "Payment failed", map[string]interface{}{
		"payment_id":    data.PaymentID,
		"error_code":    data.ErrorCode,
		"error_message": data.ErrorMessage,
	})
	return nil
}

func (s *webhookService) handlePaymentProcessing(ctx context.Context, eventID string, data *models.PaymentWebhookData) error {
	if err := s.updatePaymentStatus(ctx, data.PaymentID, data, models.PaymentPending); err != nil {
		s.logWebhookEvent(ctx, eventID, models.LogLevelWarning, "Failed to update payment status", map[string]interface{}{
			"payment_id": data.PaymentID,
			"error":      err.Error(),
		})
	}

	s.logWebhookEvent(ctx, eventID, models.LogLevelInfo, "Payment processing", map[string]interface{}{
		"payment_id": data.PaymentID,
	})
	return nil
}

func (s *webhookService) handlePaymentCancelled(ctx context.Context, eventID string, data *models.PaymentWebhookData) error {
	s.logWebhookEvent(ctx, eventID, models.LogLevelInfo, "Payment cancelled", map[string]interface{}{
		"payment_id": data.PaymentID,
	})
	return nil
}

// Helper method to update payment status (DRY pattern)
func (s *webhookService) updatePaymentStatus(ctx context.Context, dodoPaymentID string, data *models.PaymentWebhookData, status models.PaymentStatus) error {

	subscription, err := s.subscriptionRepo.GetByDodoSubscriptionID(ctx, *data.SubscriptionID)

	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	payment, err := s.paymentRepo.GetByDodoPaymentID(ctx, dodoPaymentID)

	if err != nil {
		// create a new payment record
		payment = &models.Payment{
			DodoPaymentID:  dodoPaymentID,
			SubscriptionID: subscription.ID,
			Status:         status,
			FailureReason:  data.ErrorMessage,
			Amount:         decimal.NewFromInt(data.TotalAmount),
			Currency:       data.Currency,
			CreatedAt:      time.Now(),
		}
		if err := s.paymentRepo.Create(ctx, payment); err != nil {
			return fmt.Errorf("failed to create payment: %w", err)
		}

		return nil
	}

	payment.Status = status
	payment.SubscriptionID = subscription.ID
	payment.FailureReason = data.ErrorMessage
	payment.Amount = decimal.NewFromInt(data.TotalAmount)
	payment.Currency = data.Currency
	payment.UpdatedAt = time.Now()

	if err := s.paymentRepo.Upsert(ctx, payment); err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}
	return nil
}

// Subscription event handlers
func (s *webhookService) handleSubscriptionActive(ctx context.Context, eventID string, data *models.SubscriptionWebhookData) error {
	// Find and update local subscription
	subscription, err := s.findSubscriptionByDodoID(ctx, data.SubscriptionID)
	if err != nil {
		s.logWebhookEvent(ctx, eventID, models.LogLevelWarning, "Subscription not found locally", map[string]interface{}{
			"subscription_id": data.SubscriptionID,
			"error":           err.Error(),
		})
		return nil // Don't fail the webhook processing
	}

	subscription.Status = dodopayments.SubscriptionStatusActive
	if data.NextBillingDate != "" {
		if nextBilling, err := time.Parse(time.RFC3339, data.NextBillingDate); err == nil {
			subscription.NextBillingDate = &nextBilling
		}
	}

	if err := s.subscriptionRepo.Update(ctx, subscription); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	s.logWebhookEvent(ctx, eventID, models.LogLevelInfo, "Subscription activated", map[string]interface{}{
		"subscription_id": data.SubscriptionID,
		"local_id":        subscription.ID,
	})

	return nil
}

func (s *webhookService) handleSubscriptionOnHold(ctx context.Context, eventID string, data *models.SubscriptionWebhookData) error {
	subscription, err := s.findSubscriptionByDodoID(ctx, data.SubscriptionID)
	if err != nil {
		s.logWebhookEvent(ctx, eventID, models.LogLevelWarning, "Subscription not found locally", map[string]interface{}{
			"subscription_id": data.SubscriptionID,
		})
		return nil
	}

	subscription.Status = dodopayments.SubscriptionStatusOnHold
	if err := s.subscriptionRepo.Update(ctx, subscription); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	s.logWebhookEvent(ctx, eventID, models.LogLevelInfo, "Subscription put on hold", map[string]interface{}{
		"subscription_id": data.SubscriptionID,
	})
	return nil
}

func (s *webhookService) handleSubscriptionRenewed(ctx context.Context, eventID string, data *models.SubscriptionWebhookData) error {
	subscription, err := s.findSubscriptionByDodoID(ctx, data.SubscriptionID)
	if err != nil {
		return nil
	}

	subscription.Status = dodopayments.SubscriptionStatusActive
	if data.NextBillingDate != "" {
		if nextBilling, err := time.Parse(time.RFC3339, data.NextBillingDate); err == nil {
			subscription.NextBillingDate = &nextBilling
		}
	}

	if err := s.subscriptionRepo.Update(ctx, subscription); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	s.logWebhookEvent(ctx, eventID, models.LogLevelInfo, "Subscription renewed", map[string]interface{}{
		"subscription_id": data.SubscriptionID,
	})
	return nil
}

func (s *webhookService) handleSubscriptionPaused(ctx context.Context, eventID string, data *models.SubscriptionWebhookData) error {
	subscription, err := s.findSubscriptionByDodoID(ctx, data.SubscriptionID)
	if err != nil {
		return nil
	}

	subscription.Status = dodopayments.SubscriptionStatusPaused // You might want a separate "paused" status
	if err := s.subscriptionRepo.Update(ctx, subscription); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	s.logWebhookEvent(ctx, eventID, models.LogLevelInfo, "Subscription paused", map[string]interface{}{
		"subscription_id": data.SubscriptionID,
	})
	return nil
}

func (s *webhookService) handleSubscriptionPlanChanged(ctx context.Context, eventID string, data *models.SubscriptionWebhookData) error {
	subscription, err := s.findSubscriptionByDodoID(ctx, data.SubscriptionID)
	if err != nil {
		return nil
	}

	// Update subscription details
	subscription.TotalSeats = data.Quantity
	if data.Quantity > 0 {
		pricePerSeat := decimal.NewFromFloat(float64(data.RecurringPreTaxAmount) / float64(data.Quantity))
		subscription.PricePerSeat = pricePerSeat
	}

	if err := s.subscriptionRepo.Update(ctx, subscription); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	s.logWebhookEvent(ctx, eventID, models.LogLevelInfo, "Subscription plan changed", map[string]interface{}{
		"subscription_id": data.SubscriptionID,
		"new_quantity":    data.Quantity,
	})
	return nil
}

func (s *webhookService) handleSubscriptionCancelled(ctx context.Context, eventID string, data *models.SubscriptionWebhookData) error {
	subscription, err := s.findSubscriptionByDodoID(ctx, data.SubscriptionID)
	if err != nil {
		return nil
	}

	subscription.Status = dodopayments.SubscriptionStatusCancelled
	// Note: The subscription model doesn't have a CancelledAt field
	// If needed, you could add EndDate field or create a separate audit log

	if err := s.subscriptionRepo.Update(ctx, subscription); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	s.logWebhookEvent(ctx, eventID, models.LogLevelInfo, "Subscription cancelled", map[string]interface{}{
		"subscription_id": data.SubscriptionID,
		"cancelled_at":    data.CancelledAt,
	})
	return nil
}

func (s *webhookService) handleSubscriptionFailed(ctx context.Context, eventID string, data *models.SubscriptionWebhookData) error {
	subscription, err := s.findSubscriptionByDodoID(ctx, data.SubscriptionID)
	if err != nil {
		return nil
	}

	subscription.Status = dodopayments.SubscriptionStatusFailed
	if err := s.subscriptionRepo.Update(ctx, subscription); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	s.logWebhookEvent(ctx, eventID, models.LogLevelWarning, "Subscription failed", map[string]interface{}{
		"subscription_id": data.SubscriptionID,
	})

	return nil
}

func (s *webhookService) handleSubscriptionExpired(ctx context.Context, eventID string, data *models.SubscriptionWebhookData) error {
	subscription, err := s.findSubscriptionByDodoID(ctx, data.SubscriptionID)
	if err != nil {
		return nil
	}

	subscription.Status = dodopayments.SubscriptionStatusExpired
	if err := s.subscriptionRepo.Update(ctx, subscription); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	s.logWebhookEvent(ctx, eventID, models.LogLevelInfo, "Subscription expired", map[string]interface{}{
		"subscription_id": data.SubscriptionID,
	})
	return nil
}

func (s *webhookService) handleSubscriptionUpdated(ctx context.Context, eventID string, data *models.SubscriptionWebhookData) error {
	subscription, err := s.findSubscriptionByDodoID(ctx, data.SubscriptionID)
	if err != nil {
		return nil
	}

	subscription.Status = dodopayments.SubscriptionStatusActive
	if data.NextBillingDate != "" {
		if nextBilling, err := time.Parse(time.RFC3339, data.NextBillingDate); err == nil {
			subscription.NextBillingDate = &nextBilling
		}
	}

	if err := s.subscriptionRepo.Update(ctx, subscription); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	s.logWebhookEvent(ctx, eventID, models.LogLevelInfo, "Subscription updated", map[string]interface{}{
		"subscription_id": data.SubscriptionID,
	})
	return nil
}

// Refund event handlers
func (s *webhookService) handleRefundSucceeded(ctx context.Context, eventID string, data *models.RefundWebhookData) error {
	s.logWebhookEvent(ctx, eventID, models.LogLevelInfo, "Refund succeeded", map[string]interface{}{
		"refund_id":  data.RefundID,
		"payment_id": data.PaymentID,
		"amount":     data.Amount,
	})
	return nil
}

func (s *webhookService) handleRefundFailed(ctx context.Context, eventID string, data *models.RefundWebhookData) error {
	s.logWebhookEvent(ctx, eventID, models.LogLevelWarning, "Refund failed", map[string]interface{}{
		"refund_id":  data.RefundID,
		"payment_id": data.PaymentID,
		"reason":     data.Reason,
	})
	return nil
}

// Helper methods
func (s *webhookService) findSubscriptionByDodoID(ctx context.Context, dodoSubscriptionID string) (*models.Subscription, error) {
	return s.subscriptionRepo.GetByDodoSubscriptionID(ctx, dodoSubscriptionID)
}

func (s *webhookService) logWebhookEvent(ctx context.Context, eventID string, level models.LogLevel, message string, data map[string]interface{}) {
	log := &models.WebhookLog{
		ID:          uuid.New().String(),
		EventID:     eventID,
		Level:       level,
		Message:     message,
		Data:        data,
		CreatedAt:   time.Now(),
		ProcessedBy: "webhook_service",
	}

	s.webhookLogRepo.SaveLog(ctx, log)
}

// Additional service methods
func (s *webhookService) GetWebhookEvents(ctx context.Context, eventType string, limit int, offset int) ([]*models.WebhookEvent, error) {
	if eventType == "" {
		return s.webhookRepo.GetEventsByStatus(ctx, models.WebhookProcessed, limit, offset)
	}
	return s.webhookRepo.GetEventsByType(ctx, eventType, limit, offset)
}

func (s *webhookService) GetWebhookLogs(ctx context.Context, eventID string) ([]*models.WebhookLog, error) {
	return s.webhookLogRepo.GetLogsByEvent(ctx, eventID)
}

func (s *webhookService) RetryFailedEvents(ctx context.Context, limit int) error {
	events, err := s.webhookRepo.GetEventsByStatus(ctx, models.WebhookFailed, limit, 0)
	if err != nil {
		return err
	}

	for _, event := range events {
		s.logWebhookEvent(ctx, event.ID, models.LogLevelInfo, "Retrying failed event", nil)

		if err := s.processEventByType(ctx, event); err != nil {
			errorMsg := err.Error()
			s.webhookRepo.UpdateEventStatus(ctx, event.ID, models.WebhookFailed, &errorMsg)
			s.logWebhookEvent(ctx, event.ID, models.LogLevelError, "Retry failed", map[string]interface{}{
				"error": errorMsg,
			})
		} else {
			s.webhookRepo.MarkEventProcessed(ctx, event.ID)
			s.logWebhookEvent(ctx, event.ID, models.LogLevelInfo, "Retry succeeded", nil)
		}
	}

	return nil
}

func (s *webhookService) CleanupOldEvents(ctx context.Context, retentionDays int) error {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	if err := s.webhookLogRepo.DeleteOldLogs(ctx, cutoff); err != nil {
		return fmt.Errorf("failed to cleanup old logs: %w", err)
	}

	if err := s.webhookRepo.DeleteOldEvents(ctx, cutoff); err != nil {
		return fmt.Errorf("failed to cleanup old events: %w", err)
	}

	return nil
}

// validateWebhookTimestamp validates webhook timestamp to prevent replay attacks
// Rejects webhooks older than 5 minutes
func (s *webhookService) validateWebhookTimestamp(timestampStr string) error {
	// Parse timestamp (Unix timestamp in seconds)
	var timestamp int64
	_, err := fmt.Sscanf(timestampStr, "%d", &timestamp)
	if err != nil {
		return fmt.Errorf("invalid timestamp format: %w", err)
	}

	webhookTime := time.Unix(timestamp, 0)
	now := time.Now()

	// Reject webhooks from the future (allow 1 minute clock skew)
	if webhookTime.After(now.Add(1 * time.Minute)) {
		return fmt.Errorf("webhook timestamp is in the future")
	}

	// Reject webhooks older than 5 minutes
	if now.Sub(webhookTime) > 5*time.Minute {
		return fmt.Errorf("webhook timestamp is too old (replay attack protection)")
	}

	return nil
}
