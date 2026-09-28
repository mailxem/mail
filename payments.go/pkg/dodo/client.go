package dodo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/dodopayments/dodopayments-go"
	"github.com/dodopayments/dodopayments-go/option"
	"github.com/mailxem/payments.go/internal/config"
)

type Client struct {
	client *dodopayments.Client
}

type BillingDetails struct {
	City    string `json:"city" required:"true"`
	State   string `json:"state" required:"true"`
	Street  string `json:"street" required:"true"`
	Zip     string `json:"zipcode" required:"true"`
	Country string `json:"country" required:"true"`
}

type Address struct {
	Line1 string `json:"line1"`
}

type UpdateSubscriptionRequest struct {
	CancelAtNextBillingDate bool   `json:"cancel_at_next_billing_date"`
	PlanID                  string `json:"plan_id" required:"true"`
	Quantity                int    `json:"quantity" required:"true"`
}

type CheckoutRequest struct {
	TeamID         string         `json:"team_id"`
	PlanID         string         `json:"plan_id"`
	CustomerID     string         `json:"customer_id"`
	Seats          int            `json:"seats"`
	SuccessURL     string         `json:"success_url"`
	CustomerName   string         `json:"customer_name"`
	CustomerEmail  string         `json:"customer_email"`
	BillingDetails BillingDetails `json:"billing_details" required:"true"`
}

type CheckoutSession struct {
	ID         string `json:"id"`
	URL        string `json:"url"`
	CustomerID string `json:"customer_id"`
}

type Customer struct {
	ID    string `json:"customer_id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	TaxID string `json:"business_id"`
}

type Product struct {
	ID          string `json:"product_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	Currency    string `json:"currency"`
	IsRecurring bool   `json:"is_recurring"`
}

type Subscription struct {
	ID           string                          `json:"id"`
	Customer     Customer                        `json:"customer"`
	ProductID    string                          `json:"product_id"`
	Status       dodopayments.SubscriptionStatus `json:"status"`
	Quantity     int                             `json:"quantity"`
	PaymentID    string                          `json:"payment_id"`
	PaymentLink  string                          `json:"payment_link"`
	ClientSecret string                          `json:"client_secret"`
}

type Payment struct {
	ID          string `json:"id"`
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
	Status      string `json:"status"`
	CheckoutURL string `json:"checkout_url"`
}

type WebhookEvent struct {
	ID     string                 `json:"id"`
	Type   string                 `json:"type"`
	Object map[string]interface{} `json:"object"`
}

func (c *Client) GetPaymentInvoicePDF(ctx context.Context, paymentID string) ([]byte, error) {
	payment, err := c.client.Invoices.Payments.Get(ctx, paymentID, option.WithHeader("Accept", "application/pdf"))
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	// this returns a pdf file, we need to convert it to a string
	pdfBytes, err := io.ReadAll(payment.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read payment: %w", err)
	}

	return pdfBytes, nil
}

func NewClient(cfg *config.Config) *Client {
	var client *dodopayments.Client
	if cfg.DodoTestMode {
		client = dodopayments.NewClient(
			option.WithEnvironmentTestMode(),
		)
	} else {
		client = dodopayments.NewClient(
			option.WithEnvironmentLiveMode(),
		)
	}

	return &Client{
		client: client,
	}
}

func (c *Client) CreateCustomer(ctx context.Context, teamID, name, email string) (*Customer, error) {
	// check if customer already exists
	customer, err := c.GetCustomer(ctx, teamID)
	if err == nil {
		return &Customer{
			ID:    customer.CustomerID,
			Name:  customer.Name,
			Email: customer.Email,
		}, nil
	}

	customer, err = c.client.Customers.New(ctx, dodopayments.CustomerNewParams{
		Name:  dodopayments.F(name),
		Email: dodopayments.F(email),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	return &Customer{
		ID:    customer.CustomerID,
		Name:  customer.Name,
		Email: customer.Email,
	}, nil
}

func (c *Client) GetCustomer(ctx context.Context, customerID string) (*dodopayments.Customer, error) {
	customer, err := c.client.Customers.Get(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return customer, nil
}

func (c *Client) GetCustomerPortalSession(ctx context.Context, customerID string) (string, error) {
	session, err := c.client.Customers.CustomerPortal.New(ctx, customerID, dodopayments.CustomerCustomerPortalNewParams{})
	if err != nil {
		return "", fmt.Errorf("failed to get customer portal session: %w", err)
	}

	return session.Link, nil
}

func (c *Client) CreateCheckoutSession(ctx context.Context, req CheckoutRequest) (*CheckoutSession, error) {
	// Create subscription directly
	subscription, err := c.client.Subscriptions.New(ctx, dodopayments.SubscriptionNewParams{
		Customer: dodopayments.F[dodopayments.CustomerRequestUnionParam](dodopayments.AttachExistingCustomerParam{
			CustomerID: dodopayments.F(req.CustomerID),
		}),
		ProductID:   dodopayments.F(req.PlanID),
		Quantity:    dodopayments.F(int64(req.Seats)),
		ReturnURL:   dodopayments.F(req.SuccessURL),
		PaymentLink: dodopayments.F(true),
		Billing: dodopayments.F(dodopayments.BillingAddressParam{
			Street:  dodopayments.F(req.BillingDetails.Street),
			City:    dodopayments.F(req.BillingDetails.City),
			State:   dodopayments.F(req.BillingDetails.State),
			Zipcode: dodopayments.F(req.BillingDetails.Zip),
			Country: dodopayments.F(dodopayments.CountryCode(req.BillingDetails.Country)),
		}),
		TrialPeriodDays: dodopayments.F(int64(7)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	// Get the payment link from the subscription
	paymentLink := subscription.PaymentLink

	return &CheckoutSession{
		ID:         subscription.SubscriptionID,
		URL:        paymentLink,
		CustomerID: req.CustomerID,
	}, nil
}

func (c *Client) CancelSubscription(ctx context.Context, subscriptionID string) error {
	_, err := c.client.Subscriptions.Update(ctx, subscriptionID, dodopayments.SubscriptionUpdateParams{
		Status: dodopayments.F(dodopayments.SubscriptionStatusCancelled),
	})
	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	return nil
}

func (c *Client) GetSubscription(ctx context.Context, subscriptionID string) (*Subscription, error) {
	subscription, err := c.client.Subscriptions.Get(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return &Subscription{
		ID:        subscription.SubscriptionID,
		ProductID: subscription.ProductID,
		Status:    subscription.Status,
		Quantity:  int(subscription.Quantity),
	}, nil
}

func (c *Client) UpdateSubscription(ctx context.Context, subscriptionID string, req UpdateSubscriptionRequest) error {
	err := c.client.Subscriptions.ChangePlan(ctx, subscriptionID, dodopayments.SubscriptionChangePlanParams{
		ProductID:            dodopayments.F(req.PlanID),
		ProrationBillingMode: dodopayments.F(dodopayments.SubscriptionChangePlanParamsProrationBillingModeProratedImmediately),
		Quantity:             dodopayments.F(int64(req.Quantity)),
	})
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}

func (c *Client) ProcessWebhook(ctx context.Context, payload []byte, signature string) (*WebhookEvent, error) {
	// Parse the webhook payload to extract event type and data
	var rawEvent map[string]interface{}
	if err := json.Unmarshal(payload, &rawEvent); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	// Extract event type
	eventType, ok := rawEvent["type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid event type in webhook payload")
	}

	// Create webhook event
	event := &WebhookEvent{
		Type:   eventType,
		Object: rawEvent,
	}

	return event, nil
}
