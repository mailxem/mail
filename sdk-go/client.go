package xem

import "errors"

const (
	// DefaultBaseURL is the default XEM API base URL
	DefaultBaseURL = "https://api.xem.email/api/v1"
	// DefaultTimeout is the default request timeout in milliseconds
	DefaultTimeout = 30000
)

// Client is the main XEM SDK client
//
// Example:
//
//	// Initialize with API key
//	client := xem.NewClient(&xem.Config{
//	    APIKey: "your-api-key",
//	})
//
//	// Or initialize and login later
//	client := xem.NewClient(nil)
//	client.Login("your-api-key")
//
//	// Send an email
//	result, err := client.Email.Send(&xem.SendEmailParams{
//	    To:      "user@example.com",
//	    Subject: "Hello",
//	    HTML:    xem.String("<h1>Welcome!</h1>"),
//	})
type Client struct {
	apiClient *APIClient

	// Email resource for sending emails
	Email *EmailResource
}

// NewClient creates a new XEM client
//
// Example:
//
//	client := xem.NewClient(&xem.Config{
//	    APIKey: "your-api-key",
//	})
func NewClient(config *Config) *Client {
	if config == nil {
		config = &Config{}
	}

	// Set defaults
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	// Create API client
	apiClient := newAPIClient(baseURL, timeout)

	// Set API key if provided
	if config.APIKey != "" {
		apiClient.setAPIKey(config.APIKey)
	}

	// Initialize client
	client := &Client{
		apiClient: apiClient,
	}

	// Initialize resources
	client.Email = newEmailResource(apiClient)

	return client
}

// Login authenticates with an API key
//
// Example:
//
//	client := xem.NewClient(nil)
//	err := client.Login("your-api-key")
func (c *Client) Login(apiKey string) error {
	if apiKey == "" {
		return errors.New("API key must be a non-empty string")
	}

	c.apiClient.setAPIKey(apiKey)
	return nil
}

// IsAuthenticated checks if the client is authenticated
//
// Example:
//
//	if client.IsAuthenticated() {
//	    // Ready to make API calls
//	}
func (c *Client) IsAuthenticated() bool {
	return c.apiClient.getAPIKey() != ""
}

// Helper functions for pointer conversions

// String returns a pointer to the string value
func String(v string) *string {
	return &v
}

// Bool returns a pointer to the bool value
func Bool(v bool) *bool {
	return &v
}

// Provider returns a pointer to the EmailProvider value
func Provider(v EmailProvider) *EmailProvider {
	return &v
}
