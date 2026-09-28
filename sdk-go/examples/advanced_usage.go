package main

import (
	"fmt"
	"log"

	xem "github.com/mailxem/sdk-go"
)

func main() {
	// Initialize client without API key
	client := xem.NewClient(nil)

	// Login later
	err := client.Login("your-api-key")
	if err != nil {
		log.Fatalf("Failed to login: %v", err)
	}

	// Check authentication
	if !client.IsAuthenticated() {
		log.Fatal("Client is not authenticated")
	}

	fmt.Println("Client authenticated successfully!")

	// Send email with all optional fields
	result, err := client.Email.Send(&xem.SendEmailParams{
		To:       "recipient@example.com",
		CC:       xem.String("team@example.com"),
		BCC:      xem.String("archive@example.com"),
		Subject:  "Important Update",
		HTML:     xem.String("<h1>Team Update</h1><p>Here's what's new this week...</p>"),
		ReplyTo:  xem.String("noreply@example.com"),
		Provider: xem.Provider(xem.EmailProviderCustom),
	})

	if err != nil {
		if apiErr, ok := err.(*xem.APIError); ok {
			log.Fatalf("API Error: %s (Status: %d)\n", apiErr.Message, apiErr.StatusCode)
		}
		log.Fatalf("Failed to send email: %v", err)
	}

	fmt.Printf("Email sent successfully!\n")
	if result.MessageID != nil {
		fmt.Printf("Message ID: %s\n", *result.MessageID)
	}

	// Send scheduled email
	scheduledResult, err := client.Email.Send(&xem.SendEmailParams{
		To:         "user@example.com",
		Subject:    "Scheduled Reminder",
		HTML:       xem.String("<p>This is your scheduled reminder.</p>"),
		ScheduleAt: xem.String("2026-03-01T10:00:00Z"),
	})

	if err != nil {
		log.Fatalf("Failed to schedule email: %v", err)
	}

	fmt.Printf("Email scheduled successfully!\n")
	if scheduledResult.MessageID != nil {
		fmt.Printf("Message ID: %s\n", *scheduledResult.MessageID)
	}

	// Send test email (won't actually be sent)
	testResult, err := client.Email.Send(&xem.SendEmailParams{
		To:      "test@example.com",
		Subject: "Test Email",
		HTML:    xem.String("<p>This is a test.</p>"),
		Test:    xem.Bool(true),
	})

	if err != nil {
		log.Fatalf("Failed to send test email: %v", err)
	}

	fmt.Println("Test email processed successfully!")
	fmt.Printf("Test mode result: %+v\n", testResult)
}
