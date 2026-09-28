package xem

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

// APIClient handles HTTP communication with the XEM API
type APIClient struct {
	client  *resty.Client
	apiKey  string
	baseURL string
}

// newAPIClient creates a new API client instance
func newAPIClient(baseURL string, timeout int) *APIClient {
	client := resty.New()
	client.SetBaseURL(baseURL)
	client.SetTimeout(time.Duration(timeout) * time.Millisecond)
	client.SetHeader("Content-Type", "application/json")

	return &APIClient{
		client:  client,
		baseURL: baseURL,
	}
}

// setAPIKey sets the API key for authentication
func (c *APIClient) setAPIKey(apiKey string) {
	c.apiKey = apiKey
}

// getAPIKey returns the current API key
func (c *APIClient) getAPIKey() string {
	return c.apiKey
}

// ensureAuthenticated checks if API key is set
func (c *APIClient) ensureAuthenticated() error {
	if c.apiKey == "" {
		return errors.New("API key not set. Please call Login() or pass apiKey in NewClient()")
	}
	return nil
}

// post makes a POST request to the API
func (c *APIClient) post(path string, body interface{}, result interface{}) error {
	if err := c.ensureAuthenticated(); err != nil {
		return err
	}

	resp, err := c.client.R().
		SetHeader("X-API-KEY", c.apiKey).
		SetBody(body).
		SetResult(result).
		SetError(&APIError{}).
		Post(path)

	if err != nil {
		return &APIError{
			Message: fmt.Sprintf("Request failed: %v", err),
		}
	}

	if !resp.IsSuccess() {
		if apiErr, ok := resp.Error().(*APIError); ok {
			apiErr.StatusCode = resp.StatusCode()
			return apiErr
		}
		return &APIError{
			Message:    "Request failed",
			StatusCode: resp.StatusCode(),
		}
	}

	return nil
}

// get makes a GET request to the API
func (c *APIClient) get(path string, result interface{}) error {
	if err := c.ensureAuthenticated(); err != nil {
		return err
	}

	resp, err := c.client.R().
		SetHeader("X-API-KEY", c.apiKey).
		SetResult(result).
		SetError(&APIError{}).
		Get(path)

	if err != nil {
		return &APIError{
			Message: fmt.Sprintf("Request failed: %v", err),
		}
	}

	if !resp.IsSuccess() {
		if apiErr, ok := resp.Error().(*APIError); ok {
			apiErr.StatusCode = resp.StatusCode()
			return apiErr
		}
		return &APIError{
			Message:    "Request failed",
			StatusCode: resp.StatusCode(),
		}
	}

	return nil
}

// put makes a PUT request to the API
func (c *APIClient) put(path string, body interface{}, result interface{}) error {
	if err := c.ensureAuthenticated(); err != nil {
		return err
	}

	resp, err := c.client.R().
		SetHeader("X-API-KEY", c.apiKey).
		SetBody(body).
		SetResult(result).
		SetError(&APIError{}).
		Put(path)

	if err != nil {
		return &APIError{
			Message: fmt.Sprintf("Request failed: %v", err),
		}
	}

	if !resp.IsSuccess() {
		if apiErr, ok := resp.Error().(*APIError); ok {
			apiErr.StatusCode = resp.StatusCode()
			return apiErr
		}
		return &APIError{
			Message:    "Request failed",
			StatusCode: resp.StatusCode(),
		}
	}

	return nil
}

// delete makes a DELETE request to the API
func (c *APIClient) delete(path string, result interface{}) error {
	if err := c.ensureAuthenticated(); err != nil {
		return err
	}

	resp, err := c.client.R().
		SetHeader("X-API-KEY", c.apiKey).
		SetResult(result).
		SetError(&APIError{}).
		Delete(path)

	if err != nil {
		return &APIError{
			Message: fmt.Sprintf("Request failed: %v", err),
		}
	}

	if !resp.IsSuccess() {
		if apiErr, ok := resp.Error().(*APIError); ok {
			apiErr.StatusCode = resp.StatusCode()
			return apiErr
		}
		return &APIError{
			Message:    "Request failed",
			StatusCode: resp.StatusCode(),
		}
	}

	return nil
}
