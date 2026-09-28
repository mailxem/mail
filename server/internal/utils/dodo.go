package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	dodoProdBaseURL = "https://live.dodopayments.com"
	dodoTestBaseURL = "https://test.dodopayments.com"
)

type DodoCustomer struct {
	ID    string `json:"customer_id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type DodoSubscription struct {
	ID          string    `json:"subscription_id"`
	CustomerID  string    `json:"customer_id"`
	ProductID   string    `json:"product_id"`
	Status      string    `json:"status"`
	PaymentLink string    `json:"payment_link"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
}

// getDodoBaseURL returns the appropriate base URL based on environment
func getDodoBaseURL() string {
	if os.Getenv("APP_ENV") == "production" {
		return dodoProdBaseURL
	}
	return dodoTestBaseURL
}

// CreateDodoCustomer creates a new customer in Dodo Payments
func CreateDodoCustomer(email string) (*DodoCustomer, error) {
	url := fmt.Sprintf("%s/customers", getDodoBaseURL())
	payload := map[string]string{
		"email": email,
	}

	resp, err := makeRequest("POST", url, payload)
	if err != nil {
		return nil, err
	}

	var customer DodoCustomer
	if err := json.Unmarshal(resp, &customer); err != nil {
		return nil, err
	}

	return &customer, nil
}

// CreateDodoSubscription creates a new subscription in Dodo Payments
func CreateDodoSubscription(customerID, productID string) (*DodoSubscription, error) {
	url := fmt.Sprintf("%s/subscriptions", getDodoBaseURL())
	payload := map[string]interface{}{
		"customer": map[string]string{
			"customer_id": customerID,
		},
		"product_id": productID,
		"quantity":   1,
	}

	resp, err := makeRequest("POST", url, payload)
	if err != nil {
		return nil, err
	}

	var subscription DodoSubscription
	if err := json.Unmarshal(resp, &subscription); err != nil {
		return nil, err
	}

	return &subscription, nil
}

// CreateDodoPortalSession creates a customer portal session
func CreateDodoPortalSession(customerID string) (string, error) {
	url := fmt.Sprintf("%s/customers/%s/customer-portal/session", getDodoBaseURL(), customerID)

	resp, err := makeRequest("POST", url, nil)
	if err != nil {
		return "", err
	}

	var result struct {
		Link string `json:"link"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", err
	}

	return result.Link, nil
}

// VerifyDodoWebhook verifies the webhook signature from Dodo Payments
func VerifyDodoWebhook(r *http.Request) error {
	signature := r.Header.Get("Dodo-Signature")
	if signature == "" {
		return fmt.Errorf("missing webhook signature")
	}

	// TODO: Implement proper signature verification
	// This should use the webhook secret to verify the signature
	return nil
}

// makeRequest makes an HTTP request to the Dodo Payments API
func makeRequest(method, url string, payload interface{}) ([]byte, error) {
	var body []byte
	var err error

	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("DODO_API_KEY")))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return responseBody, nil
}
