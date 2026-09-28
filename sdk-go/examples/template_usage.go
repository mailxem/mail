package main

import (
	"fmt"
	"log"

	xem "github.com/mailxem/sdk-go"
)

func main() {
	// Initialize with API key
	client := xem.NewClient(&xem.Config{
		APIKey: "your-api-key",
	})

	// Send email using a template
	result, err := client.Email.Send(&xem.SendEmailParams{
		To:         "user@example.com",
		Subject:    "Welcome to Our Platform!",
		TemplateID: xem.String("welcome-template"),
		Data: map[string]interface{}{
			"name":     "John Doe",
			"email":    "john@example.com",
			"company":  "Acme Corp",
			"role":     "Developer",
			"joinDate": "2026-02-16",
		},
	})

	if err != nil {
		log.Fatalf("Failed to send templated email: %v", err)
	}

	fmt.Printf("Templated email sent successfully!\n")
	if result.MessageID != nil {
		fmt.Printf("Message ID: %s\n", *result.MessageID)
	}

	// Send bulk emails using templates
	users := []map[string]interface{}{
		{"name": "Alice", "email": "alice@example.com"},
		{"name": "Bob", "email": "bob@example.com"},
		{"name": "Charlie", "email": "charlie@example.com"},
	}

	bulkResult, err := client.Email.Send(&xem.SendEmailParams{
		To:         "users@example.com",
		Subject:    "Monthly Newsletter",
		TemplateID: xem.String("newsletter-template"),
		Data:       map[string]interface{}{"users": users},
	})

	if err != nil {
		log.Fatalf("Failed to send bulk templated email: %v", err)
	}

	fmt.Printf("Bulk templated email sent successfully!\n")
	if bulkResult.MessageID != nil {
		fmt.Printf("Message ID: %s\n", *bulkResult.MessageID)
	}
}
