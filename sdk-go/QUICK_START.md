# Quick Start Guide - XEM Email SDK for Go

Get started with the XEM Email SDK for Go in just a few minutes!

## Installation

```bash
go get github.com/xem-email/sdk-go
```

## Basic Usage

### 1. Import and Initialize

```go
package main

import (
    "fmt"
    "log"
    
    xem "github.com/xem-email/sdk-go"
)

func main() {
    // Initialize the client with your API key
    client := xem.NewClient(&xem.Config{
        APIKey: "your-api-key",
    })
    
    // Or initialize without API key and login later
    // client := xem.NewClient(nil)
    // client.Login("your-api-key")
}
```

### 2. Send Your First Email

```go
result, err := client.Email.Send(&xem.SendEmailParams{
    To:      "recipient@example.com",
    Subject: "Welcome to XEM!",
    HTML:    xem.String("<h1>Hello World</h1>"),
})

if err != nil {
    log.Fatalf("Failed to send email: %v", err)
}

fmt.Printf("Email sent! Message ID: %s\n", *result.MessageID)
```

### 3. Handle Errors

```go
result, err := client.Email.Send(&xem.SendEmailParams{
    To:      "user@example.com",
    Subject: "Test",
    HTML:    xem.String("<p>Hello</p>"),
})

if err != nil {
    // Handle API errors
    if apiErr, ok := err.(*xem.APIError); ok {
        fmt.Printf("API Error: %s (Status: %d)\n", 
            apiErr.Message, apiErr.StatusCode)
    } else {
        fmt.Printf("Error: %v\n", err)
    }
    return
}

fmt.Println("Success!")
```

## Common Patterns

### Email with CC and BCC

```go
client.Email.Send(&xem.SendEmailParams{
    To:      "primary@example.com",
    CC:      xem.String("team@example.com"),
    BCC:     xem.String("archive@example.com"),
    Subject: "Team Update",
    HTML:    xem.String("<p>Important announcement</p>"),
})
```

### Using Templates

```go
client.Email.Send(&xem.SendEmailParams{
    To:         "user@example.com",
    Subject:    "Welcome!",
    TemplateID: xem.String("welcome-template"),
    Data: []interface{}{
        map[string]interface{}{
            "name":  "John Doe",
            "email": "john@example.com",
        },
    },
})
```

### Scheduled Emails

```go
client.Email.Send(&xem.SendEmailParams{
    To:         "user@example.com",
    Subject:    "Reminder",
    HTML:       xem.String("<p>Don't forget!</p>"),
    ScheduleAt: xem.String("2026-03-01T10:00:00Z"), // ISO 8601
})
```

### Test Mode

```go
client.Email.Send(&xem.SendEmailParams{
    To:      "test@example.com",
    Subject: "Test Email",
    HTML:    xem.String("<p>Testing</p>"),
    Test:    xem.Bool(true), // Won't actually send
})
```

## Helper Functions

The SDK provides helper functions for creating pointers:

```go
xem.String("value")    // Returns *string
xem.Bool(true)         // Returns *bool
xem.Provider(xem.EmailProviderCustom) // Returns *EmailProvider
```

## Complete Example

```go
package main

import (
    "fmt"
    "log"
    
    xem "github.com/xem-email/sdk-go"
)

func main() {
    // Initialize client
    client := xem.NewClient(&xem.Config{
        APIKey: "your-api-key",
    })
    
    // Check authentication
    if !client.IsAuthenticated() {
        log.Fatal("Not authenticated")
    }
    
    // Send email
    result, err := client.Email.Send(&xem.SendEmailParams{
        To:      "user@example.com",
        Subject: "Welcome to XEM!",
        HTML:    xem.String(`
            <h1>Welcome!</h1>
            <p>Thank you for signing up.</p>
        `),
        ReplyTo: xem.String("support@example.com"),
    })
    
    if err != nil {
        if apiErr, ok := err.(*xem.APIError); ok {
            log.Fatalf("API Error: %s (Status: %d)", 
                apiErr.Message, apiErr.StatusCode)
        }
        log.Fatalf("Error: %v", err)
    }
    
    fmt.Printf("Success! Message ID: %s\n", *result.MessageID)
}
```

## Next Steps

- Check out the [full examples](./examples/) directory
- Read the complete [documentation](./README.md)
- Learn about [advanced features](./README.md#api-reference)

## Need Help?

- 📖 [Full Documentation](./README.md)
- 🔍 [API Reference](./README.md#api-reference)
- 💬 Contact support: support@xem.email

## Pro Tips

1. **Always check errors**: Go's error handling is explicit - always check the returned error
2. **Use helper functions**: Use `xem.String()`, `xem.Bool()` for cleaner code
3. **Reuse client**: Create one client instance and reuse it across your application
4. **Authentication check**: Use `IsAuthenticated()` to verify before making API calls
5. **Error types**: Check for `*xem.APIError` to get detailed API error information
