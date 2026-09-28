# XEM Email SDK - Go Implementation

This is the official Go SDK for the XEM Email API, providing feature parity with the TypeScript/Node.js SDK.

## Project Structure

```
go/
├── README.md              # Full documentation
├── QUICK_START.md         # Quick start guide
├── LICENSE                # MIT License
├── .gitignore            # Git ignore rules
├── go.mod                # Go module definition
├── go.sum                # Go dependencies lock
├── client.go             # Main client implementation
├── api_client.go         # HTTP client with error handling
├── email.go              # Email resource methods
├── types.go              # Type definitions and structs
└── examples/             # Example usage
    ├── basic_usage.go    # Basic email sending
    ├── advanced_usage.go # Advanced features
    └── template_usage.go # Template-based emails
```

## Key Features

✅ **Complete Feature Parity** with TypeScript SDK
- Send emails with HTML content
- Template-based emails with dynamic data
- CC and BCC support
- Reply-to addresses
- Scheduled emails
- Test mode
- Custom email providers
- Flexible authentication (constructor or login method)

✅ **Idiomatic Go Design**
- Strong typing with structs
- Proper error handling
- Pointer helpers for optional fields
- Clean, documented API
- Standard Go module structure

✅ **Production Ready**
- Comprehensive error handling with custom error types
- Request timeout support
- HTTP client with retries (via resty)
- Full API validation

## Code Comparison: TypeScript vs Go

### Initialization

**TypeScript:**
```typescript
import { Xem } from '@xem.email/sdk';

const xem = new Xem({ apiKey: 'your-api-key' });
```

**Go:**
```go
import xem "github.com/xem-email/sdk-go"

client := xem.NewClient(&xem.Config{
    APIKey: "your-api-key",
})
```

### Sending Email

**TypeScript:**
```typescript
await xem.email.send({
  to: 'user@example.com',
  subject: 'Hello',
  html: '<h1>Welcome!</h1>',
});
```

**Go:**
```go
result, err := client.Email.Send(&xem.SendEmailParams{
    To:      "user@example.com",
    Subject: "Hello",
    HTML:    xem.String("<h1>Welcome!</h1>"),
})
```

### Error Handling

**TypeScript:**
```typescript
try {
  await xem.email.send(params);
} catch (error) {
  console.error('Failed:', error.message);
}
```

**Go:**
```go
result, err := client.Email.Send(params)
if err != nil {
    if apiErr, ok := err.(*xem.APIError); ok {
        fmt.Printf("API Error: %s (Status: %d)\n", 
            apiErr.Message, apiErr.StatusCode)
    }
}
```

## Architecture

### Type System

The Go SDK uses pointer types for optional fields to distinguish between "not provided" and "zero value":

```go
type SendEmailParams struct {
    To      string   // Required
    Subject string   // Required
    HTML    *string  // Optional
    CC      *string  // Optional
    Test    *bool    // Optional
}
```

Helper functions make this ergonomic:
```go
HTML: xem.String("content"),  // Instead of &"content"
Test: xem.Bool(true),          // Instead of &true
```

### Error Types

Custom error type for API errors:
```go
type APIError struct {
    Message    string
    StatusCode int
    Details    map[string]interface{}
}
```

### HTTP Client

Uses the popular `resty` HTTP client library for:
- Automatic retries
- Request/response logging
- Timeouts and cancellation
- JSON marshaling/unmarshaling

## Dependencies

- `github.com/go-resty/resty/v2` - HTTP client
- Standard library packages only

## Installation & Development

### Install Dependencies
```bash
cd go
go mod download
```

### Build
```bash
go build
```

### Run Examples
```bash
# Basic usage
go run examples/basic_usage.go

# Advanced features
go run examples/advanced_usage.go

# Template usage
go run examples/template_usage.go
```

## Testing

```bash
go test ./...
```

## Publishing

To publish this SDK to a Go module repository:

1. Create a GitHub repository at `github.com/xem-email/sdk-go`
2. Push the code
3. Tag a release: `git tag v1.0.0 && git push origin v1.0.0`
4. Users can then install with: `go get github.com/xem-email/sdk-go@v1.0.0`

## Design Decisions

### Why Pointers for Optional Fields?

In Go, we can't distinguish between a field being omitted vs. being set to its zero value. Using pointers allows:
- `nil` = field not provided (won't be sent in JSON)
- `&value` = field explicitly set (will be sent in JSON)

### Why Helper Functions?

Taking the address of literals is not allowed in Go (`&"string"` is invalid). Helper functions provide a clean way to create pointers:
```go
xem.String("value") // Returns *string
```

### Why Separate Resource Packages?

Following the TypeScript SDK's architecture, email operations are under `client.Email.*` rather than `client.SendEmail()`. This allows future expansion with other resources like `client.Templates.*`, `client.Contacts.*`, etc.

## Future Enhancements

Potential additions to match future TypeScript SDK features:
- [ ] Template management methods
- [ ] Contact/list management
- [ ] Webhook verification helpers
- [ ] Batch email operations
- [ ] Email analytics/tracking
- [ ] Rate limiting helpers
- [ ] Comprehensive test suite
- [ ] Mock server for testing
- [ ] CLI tool for quick testing

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) in the parent directory.

## License

MIT - See [LICENSE](./LICENSE)
