package xem

import "errors"

// EmailResource provides methods for email operations
type EmailResource struct {
	client *APIClient
}

// newEmailResource creates a new email resource
func newEmailResource(client *APIClient) *EmailResource {
	return &EmailResource{
		client: client,
	}
}

// Send sends an email
//
// Example:
//
//	result, err := client.Email.Send(&xem.SendEmailParams{
//	    To:      "user@example.com",
//	    Subject: "Hello",
//	    HTML:    xem.String("<h1>Welcome!</h1>"),
//	})
func (e *EmailResource) Send(params *SendEmailParams) (*SendEmailResponse, error) {
	// Validate required fields
	if params.To == "" {
		return nil, errors.New("email recipient (to) is required")
	}

	if params.Subject == "" {
		return nil, errors.New("email subject is required")
	}

	if params.HTML == nil && params.TemplateID == nil {
		return nil, errors.New("either HTML or TemplateID must be provided")
	}

	// Make API request
	var result SendEmailResponse
	err := e.client.post("/emails", params, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
