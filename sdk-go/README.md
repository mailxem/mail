# XEM Email SDK for Go

[![Release](https://img.shields.io/github/v/release/mailxem/sdk-go)](https://github.com/mailxem/sdk-go/releases)

Official Go SDK for the XEM Email API.

## Installation

```bash
go get github.com/xem-email/sdk-go
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    
    xem "github.com/xem-email/sdk-go"
)

func main() {
    // Initialize with API key
    client := xem.NewClient(&xem.Config{
        APIKey: "your-api-key",
    })

    // Send an email
    result, err := client.Email.Send(&xem.SendEmailParams{
        To:      "user@example.com",
        Subject: "Welcome to XEM!",
        HTML:    xem.String("<h1>Hello World</h1>"),
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Email sent! Message ID: %s\n", *result.MessageID)
}
```

## Usage

### Initialization

There are multiple ways to initialize the SDK:

#### Option 1: Pass API key in constructor

```go
client := xem.NewClient(&xem.Config{
    APIKey: "your-api-key",
})
```

#### Option 2: Login after initialization

```go
client := xem.NewClient(nil)
err := client.Login("your-api-key")
if err != nil {
    log.Fatal(err)
}
```

#### Option 3: Custom configuration

```go
client := xem.NewClient(&xem.Config{
    APIKey:  "your-api-key",
    BaseURL: "https://api.xem.email/api/v1", // Optional
    Timeout: 30000,                           // Optional, in milliseconds
})
```

### Sending Emails

#### Basic email

```go
result, err := client.Email.Send(&xem.SendEmailParams{
    To:      "recipient@example.com",
    Subject: "Hello!",
    HTML:    xem.String("<p>This is a test email</p>"),
})
```

#### Email with CC and BCC

```go
result, err := client.Email.Send(&xem.SendEmailParams{
    To:      "recipient@example.com",
    CC:      xem.String("cc@example.com"),
    BCC:     xem.String("bcc@example.com"),
    Subject: "Team Update",
    HTML:    xem.String("<p>Important team announcement</p>"),
})
```

#### Using templates

```go
result, err := client.Email.Send(&xem.SendEmailParams{
    To:         "user@example.com",
    Subject:    "Welcome!",
    TemplateID: xem.String("welcome-template"),
    Data: []interface{}{
        map[string]interface{}{
            "name":  "John",
            "email": "john@example.com",
        },
    },
})
```

#### Scheduled emails

```go
result, err := client.Email.Send(&xem.SendEmailParams{
    To:         "user@example.com",
    Subject:    "Reminder",
    HTML:       xem.String("<p>This is your scheduled reminder</p>"),
    ScheduleAt: xem.String("2026-03-01T10:00:00Z"), // ISO 8601 format
})
```

#### Test mode

```go
result, err := client.Email.Send(&xem.SendEmailParams{
    To:      "test@example.com",
    Subject: "Test Email",
    HTML:    xem.String("<p>Testing</p>"),
    Test:    xem.Bool(true), // Email won't actually be sent
})
```

#### Custom provider

```go
result, err := client.Email.Send(&xem.SendEmailParams{
    To:       "user@example.com",
    Subject:  "Custom Provider",
    HTML:     xem.String("<p>Using custom email provider</p>"),
    Provider: xem.Provider(xem.EmailProviderCustom),
})
```

#### Reply-to address

```go
result, err := client.Email.Send(&xem.SendEmailParams{
    To:      "customer@example.com",
    Subject: "Support Ticket #123",
    HTML:    xem.String("<p>Your support ticket has been updated</p>"),
    ReplyTo: xem.String("support@example.com"),
})
```

## API Reference

### `Client`

Main SDK client struct.

#### Constructor

```go
func NewClient(config *Config) *Client
```

**Parameters:**
- `config.APIKey` (optional) - Your XEM API key
- `config.BaseURL` (optional) - Custom API base URL (default: `https://api.xem.email/api/v1`)
- `config.Timeout` (optional) - Request timeout in milliseconds (default: `30000`)

#### Methods

##### `Login(apiKey string) error`

Authenticate with your API key.

```go
err := client.Login("your-api-key")
```

##### `IsAuthenticated() bool`

Check if the client is authenticated.

```go
if client.IsAuthenticated() {
    // Ready to make API calls
}
```

### `client.Email`

Email resource for sending emails.

#### `Send(params *SendEmailParams) (*SendEmailResponse, error)`

Send an email.

**Parameters:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `To` | `string` | ✅ | Recipient email address |
| `Subject` | `string` | ✅ | Email subject |
| `HTML` | `*string` | * | HTML content (required if `TemplateID` not provided) |
| `TemplateID` | `*string` | * | Template ID (required if `HTML` not provided) |
| `Data` | `[]interface{}` | ❌ | Template data for dynamic content |
| `CC` | `*string` | ❌ | CC recipients |
| `BCC` | `*string` | ❌ | BCC recipients |
| `ReplyTo` | `*string` | ❌ | Reply-to address |
| `Provider` | `*EmailProvider` | ❌ | Email provider (default: `CUSTOM`) |
| `ScheduleAt` | `*string` | ❌ | Schedule for future delivery (ISO 8601) |
| `Test` | `*bool` | ❌ | Test mode - email won't be sent |

## Helper Functions

The SDK provides helper functions for creating pointers to values:

```go
xem.String("value")  // *string
xem.Bool(true)       // *bool
xem.Provider(xem.EmailProviderCustom) // *EmailProvider
```

## Error Handling

The SDK returns errors for invalid parameters and API errors. Always check for errors:

```go
result, err := client.Email.Send(&xem.SendEmailParams{
    To:      "user@example.com",
    Subject: "Test",
    HTML:    xem.String("<p>Hello</p>"),
})

if err != nil {
    // Handle error
    if apiErr, ok := err.(*xem.APIError); ok {
        fmt.Printf("API Error: %s (Status: %d)\n", apiErr.Message, apiErr.StatusCode)
    } else {
        fmt.Printf("Error: %v\n", err)
    }
    return
}

fmt.Printf("Success! Message ID: %s\n", *result.MessageID)
```

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build
```

## License

MIT
