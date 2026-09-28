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

	// Send a basic email
	result, err := client.Email.Send(&xem.SendEmailParams{
		To:      "user@example.com",
		Subject: "Welcome to XEM!",
		HTML:    xem.String("<h1>Hello World</h1><p>This is a test email from the XEM Go SDK.</p>"),
	})

	if err != nil {
		log.Fatalf("Failed to send email: %v", err)
	}

	fmt.Printf("Email sent successfully!\n")
	if result.MessageID != nil {
		fmt.Printf("Message ID: %s\n", *result.MessageID)
	}
}
