package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// WebhookEvent represents a webhook event received from Dodo Payments
type WebhookEvent struct {
	ID          string         `json:"id" gorm:"primaryKey"`
	Type        string         `json:"type" gorm:"not null;index"`
	Data        map[string]any `json:"data" gorm:"type:jsonb"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	ProcessedAt *time.Time     `json:"processed_at,omitempty"`
	Status      WebhookStatus  `json:"status" gorm:"default:pending"`
	Error       *string        `json:"error,omitempty"`
	Signature   string         `json:"signature"`
	RawBody     string         `json:"raw_body" gorm:"type:text"`
}

type WebhookStatus string

const (
	WebhookPending   WebhookStatus = "pending"
	WebhookProcessed WebhookStatus = "processed"
	WebhookFailed    WebhookStatus = "failed"
	WebhookSkipped   WebhookStatus = "skipped"
)

// PaymentWebhookData represents the payload for payment webhook events
type PaymentWebhookData struct {
	PaymentID                string                 `json:"payment_id"`
	Status                   string                 `json:"status"`
	Currency                 string                 `json:"currency"`
	TotalAmount              int64                  `json:"total_amount"`
	SettlementAmount         int64                  `json:"settlement_amount"`
	SettlementCurrency       string                 `json:"settlement_currency"`
	Tax                      *int64                 `json:"tax,omitempty"`
	SettlementTax            *int64                 `json:"settlement_tax,omitempty"`
	BusinessID               string                 `json:"business_id"`
	BrandID                  string                 `json:"brand_id"`
	SubscriptionID           *string                `json:"subscription_id,omitempty"`
	PaymentMethod            *string                `json:"payment_method,omitempty"`
	PaymentMethodType        *string                `json:"payment_method_type,omitempty"`
	CardNetwork              *string                `json:"card_network,omitempty"`
	CardType                 *string                `json:"card_type,omitempty"`
	CardLastFour             *string                `json:"card_last_four,omitempty"`
	CardIssuingCountry       *string                `json:"card_issuing_country,omitempty"`
	ErrorCode                *string                `json:"error_code,omitempty"`
	ErrorMessage             *string                `json:"error_message,omitempty"`
	PaymentLink              *string                `json:"payment_link,omitempty"`
	DigitalProductsDelivered bool                   `json:"digital_products_delivered"`
	DiscountID               *string                `json:"discount_id,omitempty"`
	CreatedAt                string                 `json:"created_at"`
	UpdatedAt                *string                `json:"updated_at,omitempty"`
	Customer                 CustomerWebhookData    `json:"customer"`
	Billing                  BillingWebhookData     `json:"billing"`
	Metadata                 map[string]interface{} `json:"metadata"`
	ProductCart              []ProductCartItem      `json:"product_cart,omitempty"`
	Refunds                  []RefundWebhookData    `json:"refunds"`
	Disputes                 []DisputeWebhookData   `json:"disputes"`
}

// SubscriptionWebhookData represents the payload for subscription webhook events
type SubscriptionWebhookData struct {
	SubscriptionID             string                 `json:"subscription_id"`
	Status                     string                 `json:"status"`
	ProductID                  string                 `json:"product_id"`
	Quantity                   int                    `json:"quantity"`
	Currency                   string                 `json:"currency"`
	RecurringPreTaxAmount      int64                  `json:"recurring_pre_tax_amount"`
	TaxInclusive               bool                   `json:"tax_inclusive"`
	PaymentFrequencyCount      int                    `json:"payment_frequency_count"`
	PaymentFrequencyInterval   string                 `json:"payment_frequency_interval"`
	SubscriptionPeriodCount    int                    `json:"subscription_period_count"`
	SubscriptionPeriodInterval string                 `json:"subscription_period_interval"`
	TrialPeriodDays            int                    `json:"trial_period_days"`
	OnDemand                   bool                   `json:"on_demand"`
	CancelAtNextBillingDate    bool                   `json:"cancel_at_next_billing_date"`
	NextBillingDate            string                 `json:"next_billing_date"`
	PreviousBillingDate        string                 `json:"previous_billing_date"`
	CreatedAt                  string                 `json:"created_at"`
	CancelledAt                *string                `json:"cancelled_at,omitempty"`
	DiscountID                 *string                `json:"discount_id,omitempty"`
	Customer                   CustomerWebhookData    `json:"customer"`
	Billing                    BillingWebhookData     `json:"billing"`
	Metadata                   map[string]interface{} `json:"metadata"`
	Addons                     []AddonWebhookData     `json:"addons"`
}

// RefundWebhookData represents the payload for refund webhook events
type RefundWebhookData struct {
	RefundID   string  `json:"refund_id"`
	PaymentID  string  `json:"payment_id"`
	BusinessID string  `json:"business_id"`
	Status     string  `json:"status"`
	Amount     *int64  `json:"amount,omitempty"`
	Currency   *string `json:"currency,omitempty"`
	IsPartial  bool    `json:"is_partial"`
	Reason     *string `json:"reason,omitempty"`
	CreatedAt  string  `json:"created_at"`
}

// DisputeWebhookData represents the payload for dispute webhook events
type DisputeWebhookData struct {
	DisputeID     string  `json:"dispute_id"`
	PaymentID     string  `json:"payment_id"`
	BusinessID    string  `json:"business_id"`
	Amount        string  `json:"amount"`
	Currency      string  `json:"currency"`
	DisputeStage  string  `json:"dispute_stage"`
	DisputeStatus string  `json:"dispute_status"`
	Remarks       *string `json:"remarks,omitempty"`
	CreatedAt     string  `json:"created_at"`
}

// LicenseKeyWebhookData represents the payload for license key webhook events
type LicenseKeyWebhookData struct {
	LicenseKeyID string                 `json:"license_key_id"`
	ProductID    string                 `json:"product_id"`
	Key          string                 `json:"key"`
	CustomerID   string                 `json:"customer_id"`
	PaymentID    *string                `json:"payment_id,omitempty"`
	CreatedAt    string                 `json:"created_at"`
	ExpiresAt    *string                `json:"expires_at,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// CustomerWebhookData represents customer data in webhook payloads
type CustomerWebhookData struct {
	CustomerID string  `json:"customer_id"`
	Name       string  `json:"name"`
	Email      string  `json:"email"`
	Phone      *string `json:"phone,omitempty"`
	TaxID      *string `json:"tax_id,omitempty"`
}

// BillingWebhookData represents billing address data in webhook payloads
type BillingWebhookData struct {
	Street  string  `json:"street"`
	City    string  `json:"city"`
	State   string  `json:"state"`
	Country string  `json:"country"`
	Zipcode string  `json:"zipcode"`
	Name    *string `json:"name,omitempty"`
	Email   *string `json:"email,omitempty"`
	Phone   *string `json:"phone,omitempty"`
	TaxID   *string `json:"tax_id,omitempty"`
	Company *string `json:"company,omitempty"`
}

// ProductCartItem represents a product in the cart
type ProductCartItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Price     int64  `json:"price"`
}

// AddonWebhookData represents addon data in subscription webhooks
type AddonWebhookData struct {
	AddonID  string `json:"addon_id"`
	Quantity int    `json:"quantity"`
	Price    int64  `json:"price"`
}

// WebhookLog represents a log entry for webhook processing
type WebhookLog struct {
	ID          string                 `gorm:"primaryKey"`
	EventID     string                 `gorm:"index"` // References WebhookEvent.ID
	Level       LogLevel               `gorm:"not null"`
	Message     string                 `gorm:"not null"`
	Data        map[string]interface{} `gorm:"type:jsonb"`
	CreatedAt   time.Time              `gorm:"autoCreateTime"`
	ProcessedBy string                 `gorm:"not null"` // Service or handler name
}

type LogLevel string

const (
	LogLevelInfo    LogLevel = "info"
	LogLevelWarning LogLevel = "warning"
	LogLevelError   LogLevel = "error"
	LogLevelDebug   LogLevel = "debug"
)

// PaymentWebhookUpdate represents an update to payment status from webhook
type PaymentWebhookUpdate struct {
	PaymentID        string                 `json:"payment_id"`
	Status           string                 `json:"status"`
	ErrorCode        *string                `json:"error_code,omitempty"`
	ErrorMessage     *string                `json:"error_message,omitempty"`
	SettlementAmount *int64                 `json:"settlement_amount,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	ProcessedAt      time.Time              `json:"processed_at"`
}

// SubscriptionWebhookUpdate represents an update to subscription status from webhook
type SubscriptionWebhookUpdate struct {
	SubscriptionID  string                 `json:"subscription_id"`
	Status          string                 `json:"status"`
	NextBillingDate *string                `json:"next_billing_date,omitempty"`
	CancelledAt     *string                `json:"cancelled_at,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	ProcessedAt     time.Time              `json:"processed_at"`
}

// Helper methods for parsing webhook data
func (w *WebhookEvent) ParsePaymentData() (*PaymentWebhookData, error) {
	nestedData, ok := w.Data["data"]
	if !ok {
		return nil, fmt.Errorf("'data' field not found in webhook payload")
	}

	dataBytes, err := json.Marshal(nestedData)
	if err != nil {
		return nil, err
	}

	var paymentData PaymentWebhookData
	err = json.Unmarshal(dataBytes, &paymentData)
	return &paymentData, err
}

func (w *WebhookEvent) ParseSubscriptionData() (*SubscriptionWebhookData, error) {
	// Extract the nested "data" field
	nestedData, ok := w.Data["data"]
	if !ok {
		return nil, fmt.Errorf("'data' field not found in webhook payload")
	}

	dataBytes, err := json.Marshal(nestedData)
	if err != nil {
		return nil, err
	}

	var subscriptionData SubscriptionWebhookData
	err = json.Unmarshal(dataBytes, &subscriptionData)

	return &subscriptionData, err
}

func (w *WebhookEvent) ParseRefundData() (*RefundWebhookData, error) {
	nestedData, ok := w.Data["data"]
	if !ok {
		return nil, fmt.Errorf("'data' field not found in webhook payload")
	}

	dataBytes, err := json.Marshal(nestedData)
	if err != nil {
		return nil, err
	}

	var refundData RefundWebhookData
	err = json.Unmarshal(dataBytes, &refundData)
	return &refundData, err
}

func (w *WebhookEvent) ParseDisputeData() (*DisputeWebhookData, error) {
	nestedData, ok := w.Data["data"]
	if !ok {
		return nil, fmt.Errorf("'data' field not found in webhook payload")
	}

	dataBytes, err := json.Marshal(nestedData)
	if err != nil {
		return nil, err
	}

	var disputeData DisputeWebhookData
	err = json.Unmarshal(dataBytes, &disputeData)
	return &disputeData, err
}

func (w *WebhookEvent) ParseLicenseKeyData() (*LicenseKeyWebhookData, error) {
	nestedData, ok := w.Data["data"]
	if !ok {
		return nil, fmt.Errorf("'data' field not found in webhook payload")
	}

	dataBytes, err := json.Marshal(nestedData)
	if err != nil {
		return nil, err
	}

	var licenseKeyData LicenseKeyWebhookData
	err = json.Unmarshal(dataBytes, &licenseKeyData)
	return &licenseKeyData, err
}
