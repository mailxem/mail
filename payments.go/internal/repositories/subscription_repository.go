package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/dodopayments/dodopayments-go"
	"github.com/google/uuid"
	"github.com/mailxem/payments.go/internal/models"
	"gorm.io/gorm"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, subscription *models.Subscription) error
	GetByID(ctx context.Context, id string) (*models.Subscription, error)
	GetByTeamID(ctx context.Context, teamID string) (*models.Subscription, error)
	GetByDodoSubscriptionID(ctx context.Context, dodoSubscriptionID string) (*models.Subscription, error)
	Update(ctx context.Context, subscription *models.Subscription) error
	Delete(ctx context.Context, id string) error
	GetActiveSubscriptions(ctx context.Context) ([]*models.Subscription, error)
	GetTrialSubscriptionsToday(ctx context.Context) ([]*models.Subscription, error)
	GetSubscriptionWithPayments(ctx context.Context, id string) (*models.Subscription, error)
	GetAll(ctx context.Context, limit int, offset int) ([]*models.Subscription, error)
	CreateBillingDetails(ctx context.Context, billingDetails models.BillingDetails) error
	GetBillingDetails(ctx context.Context, teamID string) (*models.BillingDetails, error)
	UpdateBillingDetails(ctx context.Context, id string, billingDetails models.BillingDetails) error
}

type subscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

func (r *subscriptionRepository) Create(ctx context.Context, subscription *models.Subscription) error {
	return r.db.WithContext(ctx).Create(subscription).Error
}

func (r *subscriptionRepository) GetByID(ctx context.Context, id string) (*models.Subscription, error) {
	var subscription models.Subscription
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&subscription).Error
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

func (r *subscriptionRepository) GetByTeamID(ctx context.Context, teamID string) (*models.Subscription, error) {
	var subscription models.Subscription
	err := r.db.WithContext(ctx).Where("team_id = ? AND (status != ? OR status != ?)", teamID, dodopayments.SubscriptionStatusCancelled, dodopayments.SubscriptionStatusPaused).Preload("Plan").Order("updated_at DESC").First(&subscription).Error
	if err != nil {
		return nil, err
	}

	subscription.Plan.Features = []models.PlanFeatures{}
	subscription.Plan.TrailFeatures = []models.PlanFeatures{}
	var allFeatures []models.PlanFeatures
	err = r.db.WithContext(ctx).Where("plan_id = ?", subscription.Plan.ID).Find(&allFeatures).Error
	if err != nil {
		return nil, err
	}

	for _, feature := range allFeatures {
		if feature.Trail {
			subscription.Plan.TrailFeatures = append(subscription.Plan.TrailFeatures, feature)
		} else {
			subscription.Plan.Features = append(subscription.Plan.Features, feature)
		}
	}

	return &subscription, nil
}

func (r *subscriptionRepository) GetByDodoSubscriptionID(ctx context.Context, dodoSubscriptionID string) (*models.Subscription, error) {
	var subscription models.Subscription
	err := r.db.WithContext(ctx).Where("dodo_subscription_id = ?", dodoSubscriptionID).First(&subscription).Error
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

func (r *subscriptionRepository) Update(ctx context.Context, subscription *models.Subscription) error {
	return r.db.WithContext(ctx).Save(subscription).Error
}

func (r *subscriptionRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Subscription{}, "id = ?", id).Error
}

func (r *subscriptionRepository) GetActiveSubscriptions(ctx context.Context) ([]*models.Subscription, error) {
	var subscriptions []*models.Subscription
	err := r.db.WithContext(ctx).
		Where("status IN ?", []dodopayments.SubscriptionStatus{
			dodopayments.SubscriptionStatusActive,
			dodopayments.SubscriptionStatusPending,
		}).
		Find(&subscriptions).Error
	return subscriptions, err
}

func (r *subscriptionRepository) GetTrialSubscriptionsToday(ctx context.Context) ([]*models.Subscription, error) {
	var subscriptions []*models.Subscription
	now := time.Now()
	// Query for subscriptions that are currently in trial:
	// 1. trial_ends_at IS NOT NULL - subscription has a trial period
	// 2. status IN (TRIAL, PENDING) - subscription is either explicitly in trial or pending activation
	// 3. trial_ends_at > now - trial period hasn't expired yet, but will expire today
	err := r.db.WithContext(ctx).Where("trial_ends_at IS NOT NULL AND status IN (?) AND trial_ends_at BETWEEN ? AND ?", []models.SubscriptionStatus{models.SubscriptionStatus(models.SubscriptionStatusTrial), models.SubscriptionStatus(dodopayments.SubscriptionStatusPending)}, now, now.Add(24*time.Hour)).Find(&subscriptions).Error
	return subscriptions, err
}

func (r *subscriptionRepository) GetSubscriptionWithPayments(ctx context.Context, id string) (*models.Subscription, error) {
	var subscription models.Subscription
	err := r.db.WithContext(ctx).
		Preload("Payments").
		Preload("UsageRecords").
		Where("id = ?", id).
		First(&subscription).Error
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

func (r *subscriptionRepository) GetAll(ctx context.Context, limit int, offset int) ([]*models.Subscription, error) {
	var subscriptions []*models.Subscription
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&subscriptions).Error
	return subscriptions, err
}

func (r *subscriptionRepository) CreateBillingDetails(ctx context.Context, billingDetails models.BillingDetails) error {
	billingDetails.ID = uuid.New().String()
	return r.db.WithContext(ctx).Create(&billingDetails).Error
}

func (r *subscriptionRepository) GetBillingDetails(ctx context.Context, teamID string) (*models.BillingDetails, error) {
	var billingDetails models.BillingDetails
	err := r.db.WithContext(ctx).Where("team_id = ?", teamID).First(&billingDetails).Error
	if err != nil {
		return nil, err
	}
	return &billingDetails, nil
}

func (r *subscriptionRepository) UpdateBillingDetails(ctx context.Context, id string, billingDetails models.BillingDetails) error {
	return r.db.WithContext(ctx).Model(&models.BillingDetails{}).Where("id = ?", id).Updates(billingDetails).Error
}

type PaymentRepository interface {
	Create(ctx context.Context, payment *models.Payment) error
	GetByID(ctx context.Context, id string) (*models.Payment, error)
	GetByDodoPaymentID(ctx context.Context, dodoPaymentID string) (*models.Payment, error)
	GetBySubscriptionID(ctx context.Context, subscriptionID string) ([]*models.Payment, error)
	DeleteBySubscriptionID(ctx context.Context, subscriptionID string) error
	Upsert(ctx context.Context, payment *models.Payment) error
	UpdateStatus(ctx context.Context, id string, status models.PaymentStatus) error
}

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(ctx context.Context, payment *models.Payment) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

func (r *paymentRepository) GetByID(ctx context.Context, id string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) GetByDodoPaymentID(ctx context.Context, dodoPaymentID string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.WithContext(ctx).Where("dodo_payment_id = ?", dodoPaymentID).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) GetBySubscriptionID(ctx context.Context, subscriptionID string) ([]*models.Payment, error) {
	var payments []*models.Payment
	err := r.db.WithContext(ctx).
		Where("subscription_id = ?", subscriptionID).
		Order("created_at DESC").
		Find(&payments).Error
	return payments, err
}

func (r *paymentRepository) DeleteBySubscriptionID(ctx context.Context, subscriptionID string) error {
	return r.db.WithContext(ctx).Where("subscription_id = ?", subscriptionID).Delete(&models.Payment{}).Error
}

// Upsert creates a new payment record if it doesn't exist, otherwise updates the existing record
func (r *paymentRepository) Upsert(ctx context.Context, payment *models.Payment) error {
	return r.db.WithContext(ctx).Where("dodo_payment_id = ?", payment.DodoPaymentID).FirstOrCreate(payment).Error
}

func (r *paymentRepository) UpdateStatus(ctx context.Context, id string, status models.PaymentStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == models.PaymentCompleted {
		updates["processed_at"] = time.Now()
	}

	return r.db.WithContext(ctx).Model(&models.Payment{}).
		Where("id = ?", id).
		Updates(updates).Error
}

type UsageRepository interface {
	Create(ctx context.Context, usage *models.UsageRecord) error
	GetBySubscriptionAndPeriod(ctx context.Context, subscriptionID string, period time.Time, feature models.PlanFeature) (*models.UsageRecord, error)
	UpdateLimit(ctx context.Context, subscriptionID string, period time.Time, usage int, feature models.PlanFeature) error
	UpdateLimitWithCheck(ctx context.Context, subscriptionID string, period time.Time, usage int, feature models.PlanFeature, maxLimit int) error
	GetUsageHistory(ctx context.Context, subscriptionID string, limit int, feature ...models.PlanFeature) ([]*models.UsageRecord, error)
	DeleteAll(ctx context.Context, subscriptionID string) error
}

type usageRepository struct {
	db *gorm.DB
}

func NewUsageRepository(db *gorm.DB) UsageRepository {
	return &usageRepository{db: db}
}

func (r *usageRepository) DeleteAll(ctx context.Context, subscriptionID string) error {
	return r.db.WithContext(ctx).Delete(&models.UsageRecord{}, "subscription_id = ?", subscriptionID).Error
}

func (r *usageRepository) Create(ctx context.Context, usage *models.UsageRecord) error {
	return r.db.WithContext(ctx).Create(usage).Error
}

func (r *usageRepository) GetBySubscriptionAndPeriod(ctx context.Context, subscriptionID string, period time.Time, feature models.PlanFeature) (*models.UsageRecord, error) {
	var usage models.UsageRecord
	err := r.db.WithContext(ctx).
		Where("subscription_id = ? AND period = ? AND feature = ?", subscriptionID, period, feature).
		First(&usage).Error
	if err != nil {
		return nil, err
	}
	return &usage, nil
}

func (r *usageRepository) UpdateLimit(ctx context.Context, subscriptionID string, period time.Time, usage int, feature models.PlanFeature) error {
	// Try to update existing record first
	// First try to find existing record for this feature and period
	var existing models.UsageRecord
	err := r.db.WithContext(ctx).
		Where("subscription_id = ? AND period = ? AND feature = ?", subscriptionID, period, feature).
		First(&existing).Error

	switch err {
	case nil:
		// Update existing record
		return r.db.WithContext(ctx).Model(&existing).
			UpdateColumn("usage", gorm.Expr("usage + ?", usage)).Error
	case gorm.ErrRecordNotFound:
		// Create new record
		usage := &models.UsageRecord{
			SubscriptionID: subscriptionID,
			Period:         period,
			Feature:        feature,
			Usage:          usage,
		}
		return r.Create(ctx, usage)
	}
	return err
}

// UpdateLimitWithCheck atomically updates usage only if it doesn't exceed the maxLimit
// Returns an error if the update would exceed the limit, preventing race conditions
func (r *usageRepository) UpdateLimitWithCheck(ctx context.Context, subscriptionID string, period time.Time, usage int, feature models.PlanFeature, maxLimit int) error {
	// Try to find existing record
	var existing models.UsageRecord
	err := r.db.WithContext(ctx).
		Where("subscription_id = ? AND period = ? AND feature = ?", subscriptionID, period, feature).
		First(&existing).Error

	switch err {
	case nil:
		// Update existing record only if it won't exceed the limit
		// This is atomic at the database level, preventing race conditions
		result := r.db.WithContext(ctx).Model(&existing).
			Where("usage + ? <= ?", usage, maxLimit).
			UpdateColumn("usage", gorm.Expr("usage + ?", usage))

		if result.Error != nil {
			return result.Error
		}

		// If no rows were affected, it means the limit would be exceeded
		if result.RowsAffected == 0 {
			return fmt.Errorf("usage limit exceeded: current usage + %d would exceed limit of %d", usage, maxLimit)
		}

		return nil

	case gorm.ErrRecordNotFound:
		// Create new record only if initial usage doesn't exceed limit
		if usage > maxLimit {
			return fmt.Errorf("usage limit exceeded: %d exceeds limit of %d", usage, maxLimit)
		}

		newRecord := &models.UsageRecord{
			SubscriptionID: subscriptionID,
			Period:         period,
			Feature:        feature,
			Usage:          usage,
		}
		return r.Create(ctx, newRecord)
	}

	return err
}

func (r *usageRepository) GetUsageHistory(ctx context.Context, subscriptionID string, limit int, features ...models.PlanFeature) ([]*models.UsageRecord, error) {
	var usage []*models.UsageRecord
	query := r.db.WithContext(ctx).
		Where("subscription_id = ?", subscriptionID).
		Order("period DESC").
		Limit(limit)

	if len(features) > 0 {
		query = query.Where("feature IN ?", features)
	}

	// do not populate subscription relation
	query = query.Omit("Subscription")

	err := query.Find(&usage).Error
	return usage, err
}
