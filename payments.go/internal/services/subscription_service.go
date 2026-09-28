package services

import (
	"context"
	"fmt"
	"time"

	"github.com/dodopayments/dodopayments-go"
	"github.com/mailxem/payments.go/internal/config"
	"github.com/mailxem/payments.go/internal/models"
	"github.com/mailxem/payments.go/internal/repositories"
	"github.com/mailxem/payments.go/internal/structs"
	"github.com/mailxem/payments.go/pkg/dodo"
)

// ChangePlanResult holds the result of a plan change operation.
// When subscription is inactive (cancelled, expired, etc.), we delete the old subscription
// and either create a new one (free plan) or require checkout (paid plan).
type ChangePlanResult struct {
	Subscription     *models.Subscription  `json:"subscription,omitempty"`
	CheckoutRequired *CheckoutRequiredInfo `json:"checkout_required,omitempty"`
}

// CheckoutRequiredInfo is returned when user needs to complete checkout for a paid plan.
type CheckoutRequiredInfo struct {
	TeamID  string `json:"team_id"`
	PlanID  string `json:"plan_id"`
	Seats   int    `json:"seats"`
	Message string `json:"message"`
}

type SubscriptionService interface {
	CreateSubscription(ctx context.Context, teamID, planID string, seats int) (*models.Subscription, error)
	GetSubscriptionByTeamID(ctx context.Context, teamID string) (*models.Subscription, error)
	GetSubscriptionByID(ctx context.Context, id string) (*models.Subscription, error)
	UpdateSubscription(ctx context.Context, subscription *models.Subscription) error
	CancelSubscription(ctx context.Context, subscriptionID string) error
	UpgradeSubscription(ctx context.Context, subscriptionID, newPlanID string, newSeats int) error
	ChangePlan(ctx context.Context, subscriptionID, newPlanID string, newSeats int) (*ChangePlanResult, error)
	CheckUsageLimit(ctx context.Context, teamID string, feature models.PlanFeature) (bool, error)
	RecordUsage(ctx context.Context, teamID string, usage structs.RecordUsageRequest) error
	GetUsageStats(ctx context.Context, subscriptionID string) (*UsageStats, error)
	GetCustomerPortalSession(ctx context.Context, subscriptionID string) (string, error)
	RunDailyMaintenance(ctx context.Context) error
	CreateBillingDetails(ctx context.Context, billingDetails models.BillingDetails) error
	GetBillingDetails(ctx context.Context, teamID string) (*models.BillingDetails, error)
	UpdateBillingDetails(ctx context.Context, id string, billingDetails models.BillingDetails) error
}

type UsageStats struct {
	FeatureUsage  map[models.PlanFeature]int `json:"feature_usage"`
	CurrentPeriod time.Time                  `json:"current_period"`
	UsageHistory  []*models.UsageRecord      `json:"usage_history"`
}

type subscriptionService struct {
	cfg              *config.Config
	subscriptionRepo repositories.SubscriptionRepository
	paymentRepo      repositories.PaymentRepository
	usageRepo        repositories.UsageRepository
	dodoClient       *dodo.Client
	planService      PlanService
	emailService     EmailService
	teamService      TeamService
}

func (s *subscriptionService) CreateBillingDetails(ctx context.Context, billingDetails models.BillingDetails) error {
	return s.subscriptionRepo.CreateBillingDetails(ctx, billingDetails)
}

func (s *subscriptionService) GetBillingDetails(ctx context.Context, teamID string) (*models.BillingDetails, error) {
	return s.subscriptionRepo.GetBillingDetails(ctx, teamID)
}

func (s *subscriptionService) UpdateBillingDetails(ctx context.Context, id string, billingDetails models.BillingDetails) error {
	return s.subscriptionRepo.UpdateBillingDetails(ctx, id, billingDetails)
}

func NewSubscriptionService(
	cfg *config.Config,
	subscriptionRepo repositories.SubscriptionRepository,
	paymentRepo repositories.PaymentRepository,
	usageRepo repositories.UsageRepository,
	dodoClient *dodo.Client,
	planService PlanService,
	emailService EmailService,
	teamService TeamService,
) SubscriptionService {
	return &subscriptionService{
		cfg:              cfg,
		subscriptionRepo: subscriptionRepo,
		paymentRepo:      paymentRepo,
		usageRepo:        usageRepo,
		dodoClient:       dodoClient,
		planService:      planService,
		emailService:     emailService,
		teamService:      teamService,
	}
}

func (s *subscriptionService) CreateSubscription(ctx context.Context, teamID, planID string, seats int) (*models.Subscription, error) {
	// Get plan details
	plan, err := s.planService.GetPlan(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("plan %s not found", planID)
	}

	// Check if subscription already exists
	existing, err := s.subscriptionRepo.GetByTeamID(ctx, teamID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("subscription already exists for team %s", teamID)
	}

	// Create subscription
	subscription := &models.Subscription{
		TeamID:       teamID,
		PlanID:       planID,
		Status:       dodopayments.SubscriptionStatusActive,
		PricePerSeat: plan.Price,
		TotalSeats:   seats,
		BillingCycle: plan.BillingCycle,
		StartDate:    time.Now(),
	}

	// Set trial period if plan supports trials
	if plan.TrialEnabled && plan.TrialPeriodDays > 0 {
		trialEnd := time.Now().AddDate(0, 0, plan.TrialPeriodDays)
		subscription.TrialEndsAt = &trialEnd
		subscription.Status = dodopayments.SubscriptionStatusPending
	}

	// Calculate next billing date
	if !plan.Price.IsZero() {
		if plan.BillingCycle == models.BillingYearly {
			nextBilling := time.Now().AddDate(1, 0, 0) // Yearly billing
			subscription.NextBillingDate = &nextBilling
		} else {
			nextBilling := time.Now().AddDate(0, 1, 0) // Monthly billing
			subscription.NextBillingDate = &nextBilling
		}
	}

	err = s.subscriptionRepo.Create(ctx, subscription)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	go func() {
		time.Sleep(1 * time.Second)
		if subscription.PlanID != s.cfg.Plans.FreePlanID {
			s.emailService.sendReminderEmail(context.Background(), subscription)
		} else {
			s.emailService.ScheduleFreeTrialEmails(context.Background(), subscription, *subscription.TrialEndsAt)
		}
	}()

	return subscription, nil
}

func (s *subscriptionService) GetSubscriptionByTeamID(ctx context.Context, teamID string) (*models.Subscription, error) {
	return s.subscriptionRepo.GetByTeamID(ctx, teamID)
}

func (s *subscriptionService) GetSubscriptionByID(ctx context.Context, id string) (*models.Subscription, error) {
	return s.subscriptionRepo.GetByID(ctx, id)
}

func (s *subscriptionService) UpdateSubscription(ctx context.Context, subscription *models.Subscription) error {
	return s.subscriptionRepo.Update(ctx, subscription)
}

func (s *subscriptionService) CancelSubscription(ctx context.Context, subscriptionID string) error {
	subscription, err := s.subscriptionRepo.GetByID(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	// Check if subscription can be cancelled
	switch subscription.Status {
	case dodopayments.SubscriptionStatusCancelled:
		return fmt.Errorf("subscription already cancelled")
	case dodopayments.SubscriptionStatusPending:
		return fmt.Errorf("subscription is pending, cannot cancel")
	case dodopayments.SubscriptionStatusExpired:
		return fmt.Errorf("subscription is expired, cannot cancel")
	case dodopayments.SubscriptionStatusFailed:
		return fmt.Errorf("subscription is failed, cannot cancel")
	case dodopayments.SubscriptionStatusPaused:
		return fmt.Errorf("subscription is paused, cannot cancel")
	}

	// Cancel with Dodo if external subscription exists
	if subscription.DodoSubscriptionID != nil {
		err = s.dodoClient.CancelSubscription(ctx, *subscription.DodoSubscriptionID)
		if err != nil {
			return fmt.Errorf("failed to cancel external subscription: %w", err)
		}
	}

	// Update subscription status
	subscription.Status = dodopayments.SubscriptionStatusCancelled
	endDate := time.Now()
	subscription.EndDate = &endDate

	go func() {
		s.emailService.sendReminderEmail(ctx, subscription)
	}()

	return s.subscriptionRepo.Update(ctx, subscription)
}

// isSubscriptionActiveAndUpgradable returns true if the subscription can be upgraded in-place
// (active status, has Dodo subscription ID, not on free plan).
func isSubscriptionActiveAndUpgradable(status dodopayments.SubscriptionStatus, dodoSubscriptionID *string, planID, freePlanID string) bool {
	if planID == freePlanID {
		return false
	}
	if dodoSubscriptionID == nil || *dodoSubscriptionID == "" {
		return false
	}
	return status == dodopayments.SubscriptionStatusActive
}

func (s *subscriptionService) UpgradeSubscription(ctx context.Context, subscriptionID, newPlanID string, newSeats int) error {
	result, err := s.ChangePlan(ctx, subscriptionID, newPlanID, newSeats)
	if err != nil {
		return err
	}
	if result.CheckoutRequired != nil {
		return fmt.Errorf("subscription is inactive; please complete checkout to activate your new plan")
	}
	return nil
}

func (s *subscriptionService) ChangePlan(ctx context.Context, subscriptionID, newPlanID string, newSeats int) (*ChangePlanResult, error) {
	subscription, err := s.subscriptionRepo.GetByID(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("subscription not found: %w", err)
	}

	// Resolve plan and seats - at least one must be provided
	if newPlanID == "" {
		newPlanID = subscription.PlanID
	}
	if newSeats <= 0 {
		newSeats = subscription.TotalSeats
	}
	if newSeats <= 0 {
		newSeats = 1
	}

	plan, err := s.planService.GetPlan(ctx, newPlanID)
	if err != nil {
		return nil, fmt.Errorf("plan %s not found", newPlanID)
	}

	// Case 1: Active subscription with Dodo, upgrading to another paid plan → upgrade in-place
	// (Cannot upgrade in-place when: switching to free plan, or subscription is inactive)
	if newPlanID != s.cfg.Plans.FreePlanID &&
		isSubscriptionActiveAndUpgradable(subscription.Status, subscription.DodoSubscriptionID, subscription.PlanID, s.cfg.Plans.FreePlanID) {
		err = s.dodoClient.UpdateSubscription(ctx, *subscription.DodoSubscriptionID, dodo.UpdateSubscriptionRequest{
			PlanID:   plan.DodoPlanID,
			Quantity: newSeats,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update subscription: %w", err)
		}
		subscription.PlanID = newPlanID
		subscription.PricePerSeat = plan.Price
		subscription.TotalSeats = newSeats
		if err := s.subscriptionRepo.Update(ctx, subscription); err != nil {
			return nil, fmt.Errorf("failed to save subscription: %w", err)
		}
		return &ChangePlanResult{Subscription: subscription}, nil
	}

	// Case 2: Inactive subscription (cancelled, expired, failed, paused, past_due, pending)
	// OR free plan (cannot upgrade in-place) → cancel and create new
	teamID := subscription.TeamID

	// Step 2a: Cancel Dodo subscription if it exists and is not already cancelled (best effort)
	if subscription.DodoSubscriptionID != nil && *subscription.DodoSubscriptionID != "" &&
		subscription.Status != dodopayments.SubscriptionStatusCancelled &&
		subscription.Status != dodopayments.SubscriptionStatusExpired {
		_ = s.dodoClient.CancelSubscription(ctx, *subscription.DodoSubscriptionID)
		// Ignore error - subscription may already be cancelled at Dodo
	}

	// Step 2b: Delete usage records for the old subscription
	if err := s.usageRepo.DeleteAll(ctx, subscriptionID); err != nil {
		return nil, fmt.Errorf("failed to delete usage records: %w", err)
	}

	// Step 2c: Delete payments (required before subscription due to FK constraint)
	if err := s.paymentRepo.DeleteBySubscriptionID(ctx, subscriptionID); err != nil {
		return nil, fmt.Errorf("failed to delete payments: %w", err)
	}

	// Step 2d: Delete the old subscription
	if err := s.subscriptionRepo.Delete(ctx, subscriptionID); err != nil {
		return nil, fmt.Errorf("failed to delete old subscription: %w", err)
	}

	// Step 2e: Create new subscription
	if newPlanID == s.cfg.Plans.FreePlanID {
		return s.createSubscriptionForTeam(ctx, teamID, newPlanID, newSeats)
	}

	// Paid plan: return checkout required - client must call CreateCheckoutSession
	return &ChangePlanResult{
		CheckoutRequired: &CheckoutRequiredInfo{
			TeamID:  teamID,
			PlanID:  newPlanID,
			Seats:   newSeats,
			Message: "Old subscription removed. Please complete checkout to activate your new subscription.",
		},
	}, nil
}

// createSubscriptionForTeam creates a new subscription for a team (used after deleting old subscription).
func (s *subscriptionService) createSubscriptionForTeam(ctx context.Context, teamID, planID string, seats int) (*ChangePlanResult, error) {
	sub, err := s.CreateSubscription(ctx, teamID, planID, seats)
	if err != nil {
		return nil, err
	}
	return &ChangePlanResult{Subscription: sub}, nil
}

func (s *subscriptionService) CheckUsageLimit(ctx context.Context, teamID string, feature models.PlanFeature) (bool, error) {
	subscription, err := s.subscriptionRepo.GetByTeamID(ctx, teamID)
	if err != nil {
		return false, fmt.Errorf("subscription not found: %w", err)
	}

	if subscription.Status == dodopayments.SubscriptionStatusCancelled || subscription.Status == dodopayments.SubscriptionStatusExpired || subscription.Status == dodopayments.SubscriptionStatusFailed || subscription.Status == dodopayments.SubscriptionStatusPaused {
		return false, fmt.Errorf("subscription is cancelled, expired, failed, or paused")
	}

	// get free plan
	freePlan, err := s.planService.GetPlan(ctx, s.cfg.Plans.FreePlanID)
	if err != nil {
		return false, fmt.Errorf("free plan not found: %w", err)
	}

	// Get current period usage
	currentPeriod := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)
	currentUsage, err := s.usageRepo.GetBySubscriptionAndPeriod(ctx, subscription.ID, currentPeriod, feature)
	if err != nil || currentUsage == nil {
		return true, nil // No usage yet, within limits
	}

	// Determine which features to check (trial or regular)
	features := subscription.Plan.Features
	if subscription.Plan.TrialEnabled && subscription.TrialEndsAt != nil && subscription.TrialEndsAt.After(currentPeriod) && subscription.Plan.ID == s.cfg.Plans.FreePlanID {
		features = subscription.Plan.TrailFeatures
	}

	if subscription.Plan.ID != s.cfg.Plans.FreePlanID && subscription.Status == dodopayments.SubscriptionStatusPending {
		features = make([]models.PlanFeatures, len(freePlan.Features))
		for i, feat := range freePlan.Features {
			features[i] = feat.ToPlanFeature()
		}
	}

	// Check feature quota
	for _, feat := range features {
		if feat.Feature == feature {
			if models.IsUnlimitedQuota(feat.Quota) {
				return true, nil // Unlimited usage
			}
			return currentUsage.Usage < (feat.Quota * subscription.TotalSeats), nil
		}
	}

	return true, nil // No limit found, allow usage
}

func (s *subscriptionService) RecordUsage(ctx context.Context, teamID string, usage structs.RecordUsageRequest) error {
	subscription, err := s.subscriptionRepo.GetByTeamID(ctx, teamID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	if subscription.Status == dodopayments.SubscriptionStatusCancelled || subscription.Status == dodopayments.SubscriptionStatusExpired || subscription.Status == dodopayments.SubscriptionStatusFailed || subscription.Status == dodopayments.SubscriptionStatusPaused {
		return fmt.Errorf("subscription is cancelled, expired, failed, or paused")
	}

	// if subscription is pending and trial ends at is in the future, let them unlimited usage
	if subscription.PlanID != s.cfg.Plans.FreePlanID && subscription.Status == dodopayments.SubscriptionStatusPending && subscription.TrialEndsAt != nil && subscription.TrialEndsAt.After(time.Now()) {
		// let them use without checking limits
		return nil
	}

	currentPeriod := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)

	// Find the feature quota and use atomic check-and-update to prevent race conditions
	for _, feature := range subscription.Plan.Features {
		if feature.Feature == usage.Feature {
			if models.IsUnlimitedQuota(feature.Quota) {
				// Unlimited usage - just record it without checking limits
				err = s.usageRepo.UpdateLimit(ctx, subscription.ID, currentPeriod, usage.Usage, usage.Feature)
				if err != nil {
					return fmt.Errorf("failed to record usage: %w", err)
				}
				return nil
			}

			// Use atomic update with limit check to prevent race conditions
			maxLimit := feature.Quota * subscription.TotalSeats
			err = s.usageRepo.UpdateLimitWithCheck(ctx, subscription.ID, currentPeriod, usage.Usage, usage.Feature, maxLimit)
			if err != nil {
				return err // Error message already includes details
			}
			return nil
		}
	}

	// No feature limit found - record usage anyway
	err = s.usageRepo.UpdateLimit(ctx, subscription.ID, currentPeriod, usage.Usage, usage.Feature)
	if err != nil {
		return fmt.Errorf("failed to record usage: %w", err)
	}

	return nil
}

func (s *subscriptionService) GetUsageStats(ctx context.Context, subscriptionID string) (*UsageStats, error) {
	_, err := s.subscriptionRepo.GetByID(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("subscription not found: %w", err)
	}

	// Get usage history (last 12 months)
	usageHistory, err := s.usageRepo.GetUsageHistory(ctx, subscriptionID, 12)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage history: %w", err)
	}

	currentPeriod := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)

	// Build feature usage map from current period
	featureUsage := make(map[models.PlanFeature]int)
	for _, record := range usageHistory {
		if record.Period.Equal(currentPeriod) {
			featureUsage[record.Feature] = record.Usage
		}
	}

	return &UsageStats{
		FeatureUsage:  featureUsage,
		CurrentPeriod: currentPeriod,
		UsageHistory:  usageHistory,
	}, nil
}

func (s *subscriptionService) GetCustomerPortalSession(ctx context.Context, subscriptionID string) (string, error) {
	subscription, err := s.subscriptionRepo.GetByID(ctx, subscriptionID)
	if err != nil {
		return "", fmt.Errorf("subscription not found: %w", err)
	}

	return s.dodoClient.GetCustomerPortalSession(ctx, *subscription.DodoCustomerID)
}

// RunDailyMaintenance expires overdue trial subscriptions and resets usage for current period
func (s *subscriptionService) RunDailyMaintenance(ctx context.Context) error {
	subs, err := s.subscriptionRepo.GetTrialSubscriptionsToday(ctx)
	if err != nil {
		return err
	}

	for _, sub := range subs {
		// DELETE ALL USAGE RECORDS FOR THE SUBSCRIPTION
		err = s.usageRepo.DeleteAll(ctx, sub.ID)
		if err != nil {
			return fmt.Errorf("failed to delete usage records: %w", err)
		}
	}

	return nil
}

type PaymentService interface {
	CreateCheckoutSession(ctx context.Context, req CheckoutSessionRequest) (*CheckoutSessionResponse, error)
	GetPaymentHistory(ctx context.Context, subscriptionID string) ([]*models.Payment, error)
	GetPaymentInvoicePDF(ctx context.Context, paymentID string) ([]byte, error)
}

type CheckoutSessionRequest struct {
	TeamID         string                `json:"team_id"`
	PlanID         string                `json:"plan_id"`
	Seats          int                   `json:"seats"`
	SuccessURL     string                `json:"success_url"`
	CustomerName   string                `json:"customer_name"`
	CustomerEmail  string                `json:"customer_email"`
	BillingDetails models.BillingDetails `json:"billing_details" required:"true"`
}

type CheckoutSessionResponse struct {
	CheckoutURL string `json:"checkout_url"`
	SessionID   string `json:"session_id"`
}

type paymentService struct {
	cfg              *config.Config
	subscriptionRepo repositories.SubscriptionRepository
	paymentRepo      repositories.PaymentRepository
	emailService     EmailService
	dodoClient       *dodo.Client
	webhookService   WebhookService
	planService      PlanService
}

func NewPaymentService(
	cfg *config.Config,
	subscriptionRepo repositories.SubscriptionRepository,
	paymentRepo repositories.PaymentRepository,
	emailService EmailService,
	dodoClient *dodo.Client,
	webhookService WebhookService,
	planService PlanService,
) PaymentService {
	return &paymentService{
		cfg:              cfg,
		subscriptionRepo: subscriptionRepo,
		paymentRepo:      paymentRepo,
		emailService:     emailService,
		dodoClient:       dodoClient,
		webhookService:   webhookService,
		planService:      planService,
	}
}

func (s *paymentService) CreateCheckoutSession(ctx context.Context, req CheckoutSessionRequest) (*CheckoutSessionResponse, error) {
	// Get plan details from the plan service to validate the plan exists and get pricing information
	plan, err := s.planService.GetPlan(ctx, req.PlanID)
	if err != nil {
		return nil, fmt.Errorf("plan %s not found", req.PlanID)
	}

	// check billing details
	if req.BillingDetails.Street == "" || req.BillingDetails.City == "" || req.BillingDetails.State == "" || req.BillingDetails.Zip == "" || req.BillingDetails.Country == "" {
		return nil, fmt.Errorf("billing details are required")
	}

	// Create or retrieve existing customer in Dodo Payments using team ID as identifier
	customer, err := s.dodoClient.CreateCustomer(ctx, req.TeamID, req.CustomerName, req.CustomerEmail)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	// Check if team already has an active subscription (excluding free plan)
	// This prevents duplicate paid subscriptions for the same team
	subscription, err := s.subscriptionRepo.GetByTeamID(ctx, req.TeamID)
	if err == nil && subscription != nil {
		// Allow creating new subscription only in these scenarios:
		// 1. Current subscription is for the free plan (can upgrade to paid)
		// 2. Current subscription is cancelled (can resubscribe)
		// 3. Current subscription is expired (can renew)
		// 4. Current subscription has failed (can retry with new subscription)
		// 5. Current subscription is pending (can retry with new subscription)

		// Block subscription creation if:
		// - Team has an active paid subscription
		// - Team has a pending paid subscription (waiting for payment)
		// - Team has a trial subscription on a paid plan
		// - Team has a paused subscription (should resume instead)
		// - Team has a past_due subscription (should update payment method)
		// - Team has a trial subscription on a paid plan
		if subscription.PlanID != s.cfg.Plans.FreePlanID &&
			subscription.Status != dodopayments.SubscriptionStatusCancelled &&
			subscription.Status != dodopayments.SubscriptionStatusExpired &&
			subscription.Status != dodopayments.SubscriptionStatusFailed &&
			subscription.Status != dodopayments.SubscriptionStatusPending {
			return nil, fmt.Errorf("you already have a subscription")
		}
	}

	// Create new subscription model with plan details and customer information
	subscription = &models.Subscription{
		TeamID:         req.TeamID,
		PlanID:         req.PlanID,
		PricePerSeat:   plan.Price,
		TotalSeats:     req.Seats,
		BillingCycle:   plan.BillingCycle,
		DodoCustomerID: &customer.ID,
	}

	// Handle free plan scenario - create trial subscription without payment processing
	if plan.ID == s.cfg.Plans.FreePlanID {
		// Calculate trial end date based on plan's trial period
		trialEnd := time.Now().AddDate(0, 0, plan.TrialPeriodDays)
		subscription.Status = models.SubscriptionStatusTrial
		subscription.TrialEndsAt = &trialEnd

		// Save trial subscription to database
		if err = s.subscriptionRepo.Create(ctx, subscription); err != nil {
			return nil, fmt.Errorf("failed to create subscription: %w", err)
		}

		go func() {
			time.Sleep(1 * time.Second)
			s.emailService.ScheduleFreeTrialEmails(context.Background(), subscription, *subscription.TrialEndsAt)
		}()

		// Return success URL directly since no payment is required for free plan
		return &CheckoutSessionResponse{
			CheckoutURL: req.SuccessURL,
			SessionID:   subscription.ID,
		}, nil
	}

	// Handle paid plan scenario - create checkout session through Dodo Payments
	dodoSubscription, err := s.dodoClient.CreateCheckoutSession(ctx, dodo.CheckoutRequest{
		TeamID:       req.TeamID,
		PlanID:       plan.DodoPlanID, // Use Dodo-specific plan ID
		Seats:        req.Seats,
		CustomerName: req.CustomerName,
		CustomerID:   customer.ID,
		// Map billing details from request to Dodo format
		BillingDetails: dodo.BillingDetails{
			Street:  req.BillingDetails.Street,
			City:    req.BillingDetails.City,
			State:   req.BillingDetails.State,
			Zip:     req.BillingDetails.Zip,
			Country: req.BillingDetails.Country,
		},
		SuccessURL:    req.SuccessURL,
		CustomerEmail: req.CustomerEmail,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create dodo subscription: %w", err)
	}

	// Set subscription as inactive until payment is confirmed via webhook
	subscription.Status = dodopayments.SubscriptionStatusPending
	subscription.DodoSubscriptionID = &dodoSubscription.ID

	// Save subscription to database with inactive status
	if err = s.subscriptionRepo.Create(ctx, subscription); err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	go func() {
		time.Sleep(1 * time.Second)
		s.emailService.ScheduleFreeTrialEmails(context.Background(), subscription, *subscription.TrialEndsAt)
	}()

	// Return checkout URL from Dodo Payments for user to complete payment
	return &CheckoutSessionResponse{
		CheckoutURL: dodoSubscription.URL,
		SessionID:   dodoSubscription.ID,
	}, nil
}

func (s *paymentService) GetPaymentHistory(ctx context.Context, subscriptionID string) ([]*models.Payment, error) {
	return s.paymentRepo.GetBySubscriptionID(ctx, subscriptionID)
}

func (s *paymentService) GetPaymentInvoicePDF(ctx context.Context, paymentID string) ([]byte, error) {
	payment, err := s.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	if payment.DodoPaymentID == "" {
		return nil, fmt.Errorf("payment does not have a dodo payment id")
	}

	pdfBytes, err := s.dodoClient.GetPaymentInvoicePDF(ctx, payment.DodoPaymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment invoice PDF: %w", err)
	}

	return pdfBytes, nil
}
