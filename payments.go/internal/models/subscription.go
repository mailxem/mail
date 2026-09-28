package models

import (
	"time"

	"github.com/dodopayments/dodopayments-go"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type SubscriptionStatus string
type BillingCycle string
type PaymentStatus string

const (
	BillingMonthly BillingCycle = "MONTHLY"
	BillingYearly  BillingCycle = "YEARLY"
)

const (
	PaymentPending   PaymentStatus = "PENDING"
	PaymentCompleted PaymentStatus = "COMPLETED"
	PaymentFailed    PaymentStatus = "FAILED"
	PaymentRefunded  PaymentStatus = "REFUNDED"
)

// lets add trial status to dodopayments.SubscriptionStatus
const (
	SubscriptionStatusTrial dodopayments.SubscriptionStatus = "TRIAL"
)

type Subscription struct {
	ID     string                          `json:"id" gorm:"primaryKey;type:uuid"`
	TeamID string                          `json:"team_id" gorm:"not null;index"`
	Status dodopayments.SubscriptionStatus `json:"status" gorm:"default:active" swaggertype:"string" enums:"active,canceled,past_due,paused,pending,TRIAL"`

	// Dodo Payments Integration
	DodoSubscriptionID *string `json:"dodo_subscription_id,omitempty"`
	DodoCustomerID     *string `json:"dodo_customer_id,omitempty"`

	// Billing
	PricePerSeat decimal.Decimal `json:"price_per_seat" gorm:"type:decimal(10,2)"`
	TotalSeats   int             `json:"total_seats" gorm:"default:1"`
	BillingCycle BillingCycle    `json:"billing_cycle" gorm:"default:MONTHLY"`

	// Dates
	StartDate       time.Time  `json:"start_date"`
	EndDate         *time.Time `json:"end_date,omitempty"`
	NextBillingDate *time.Time `json:"next_billing_date,omitempty"`
	TrialEndsAt     *time.Time `json:"trial_ends_at,omitempty"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`

	// Relations
	Payments     []Payment     `json:"payments,omitempty"`
	UsageRecords []UsageRecord `json:"usage_records,omitempty"`

	Plan   Plan   `json:"plan,omitempty" gorm:"foreignKey:PlanID"`
	PlanID string `json:"plan_id,omitempty" gorm:"not null;index"`
}

func (s *Subscription) IsTrial() bool {
	return s.TrialEndsAt != nil && s.TrialEndsAt.After(time.Now()) && (s.Status == SubscriptionStatusTrial || s.Status == dodopayments.SubscriptionStatusPending)
}

// BillingDetails is used to store the billing details for a subscription
type BillingDetails struct {
	ID      string `json:"id" gorm:"primaryKey"`
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
	Country string `json:"country"`
	TeamID  string `json:"team_id" gorm:"not null;index"`
}

type Payment struct {
	ID             string `json:"id" gorm:"primaryKey;type:uuid"`
	SubscriptionID string `json:"subscription_id" gorm:"type:uuid;not null;index"`

	// Dodo Payments
	DodoPaymentID string          `json:"dodo_payment_id" gorm:"unique;not null"`
	Amount        decimal.Decimal `json:"amount" gorm:"type:decimal(10,2)"`
	Currency      string          `json:"currency" gorm:"default:USD"`
	Status        PaymentStatus   `json:"status"`

	// Metadata
	Description   *string    `json:"description,omitempty"`
	FailureReason *string    `json:"failure_reason,omitempty"`
	ProcessedAt   *time.Time `json:"processed_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Subscription Subscription `json:"subscription,omitempty" gorm:"foreignKey:SubscriptionID"`
}

type UsageRecord struct {
	ID             string `json:"id" gorm:"primaryKey;type:uuid"`
	SubscriptionID string `json:"subscription_id" gorm:"type:uuid;not null;index"`

	Period  time.Time   `json:"period"` // Billing period (month/year)
	Feature PlanFeature `json:"feature" gorm:"not null"`
	Usage   int         `json:"usage" gorm:"default:0"` // Actual usage amount

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Subscription Subscription `json:"subscription,omitempty" gorm:"foreignKey:SubscriptionID"`
}

// BeforeCreate hook to generate UUID
func (s *Subscription) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	if s.StartDate.IsZero() {
		s.StartDate = time.Now()
	}
	return nil
}

func (p *Payment) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

func (u *UsageRecord) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}
