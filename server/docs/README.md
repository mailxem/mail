# Posthoot API Documentation

This directory contains the OpenAPI/Swagger documentation for the Posthoot API server.

## 📁 Directory Structure

```
docs/
├── README.md           # This file
├── swagger/           # Generated Swagger 2.0 documentation
│   ├── docs.go        # Go package with embedded documentation
│   ├── swagger.json   # Swagger 2.0 specification (JSON format)
│   └── swagger.yaml   # Swagger 2.0 specification (YAML format)
└── internal/
    └── docs/
        └── swagger.go  # Swagger configuration and metadata

# Root directory also contains:
openapi.json          # OpenAPI 3.0 specification (JSON format)
```

## 🚀 Quick Start

### Generate Documentation

```bash
# Generate Swagger 2.0 documentation
make docs

# Generate OpenAPI 3.0 specification (openapi.json)
make openapi

# Using the scripts
./scripts/swagger.sh generate
./scripts/generate-openapi.sh generate

# Direct command
export PATH=$PATH:$(go env GOPATH)/bin
swag init -g cmd/main.go -o docs/swagger --parseDependency --parseInternal
```

### Serve Documentation

```bash
# Using Makefile
make docs-serve

# Using the script
./scripts/swagger.sh serve

# Direct command
export PATH=$PATH:$(go env GOPATH)/bin
swag serve -F=swagger docs/swagger/swagger.json
```

### Validate Documentation

```bash
# Using Makefile
make docs-validate

# Using the script
./scripts/swagger.sh validate

# Direct command
export PATH=$PATH:$(go env GOPATH)/bin
swag validate docs/swagger/swagger.json
```

## 📖 API Overview

The Posthoot API provides comprehensive endpoints for:

- **Authentication** - User registration, login, and token management
- **Teams** - Team management and collaboration features
- **Campaigns** - Email campaign creation, management, and tracking
- **Analytics** - Campaign analytics, audience insights, and performance metrics
- **Contacts** - Contact management and mailing list operations
- **Templates** - Email template management and customization
- **Automations** - Email automation workflows and triggers
- **SMTP** - SMTP configuration and email delivery settings
- **IMAP** - IMAP configuration for email inbox management
- **Webhooks** - Webhook management for real-time event notifications
- **Files** - File upload and management for attachments and media
- **Domains** - Domain management for email authentication
- **API Keys** - API key management for programmatic access

## 🔐 Authentication

The API supports two authentication methods:

1. **Bearer Token Authentication** - JWT tokens for user sessions
2. **API Key Authentication** - API keys for programmatic access

### Bearer Token
```
Authorization: Bearer <your-jwt-token>
```

### API Key
```
X-API-KEY: <your-api-key>
```

## 📝 Adding Documentation

To add or update API documentation:

1. **Add Swagger Annotations** to your handler functions:

```go
// @Summary Create a new user
// @Description Create a new user account
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "User data"
// @Success 201 {object} User
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users [post]
func CreateUser(c echo.Context) error {
    // Your handler code
}
```

2. **Define Request/Response Models**:

```go
type CreateUserRequest struct {
    Email     string `json:"email" validate:"required,email"`
    Password  string `json:"password" validate:"required,min=8"`
    FirstName string `json:"first_name" validate:"required"`
    LastName  string `json:"last_name" validate:"required"`
}

type User struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    FirstName string    `json:"first_name"`
    LastName  string    `json:"last_name"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

3. **Regenerate Documentation**:

```bash
make docs
```

## 🏷️ Available Tags

Use these tags to organize your API endpoints:

- `@Tags Authentication` - User authentication and authorization
- `@Tags Teams` - Team management and collaboration
- `@Tags Campaigns` - Email campaign management
- `@Tags Analytics` - Analytics and reporting
- `@Tags Contacts` - Contact management
- `@Tags Templates` - Email templates
- `@Tags Automations` - Email automation workflows
- `@Tags SMTP` - SMTP configuration
- `@Tags IMAP` - IMAP configuration
- `@Tags Webhooks` - Webhook management
- `@Tags Files` - File management
- `@Tags Domains` - Domain management
- `@Tags API Keys` - API key management

## 🔧 Configuration

The main swagger configuration is in `internal/docs/swagger.go`:

```go
// @title Posthoot API
// @version 1.0
// @description Comprehensive API server for Posthoot email marketing platform
// @host backyard.posthoot.com
// @BasePath /api/v1
// @schemes https
```

## 📊 Available Formats

The documentation is available in multiple formats:

- **Swagger UI** - Interactive web interface (accessible via `/swagger/`)
- **JSON** - `docs/swagger/swagger.json`
- **YAML** - `docs/swagger/swagger.yaml`
- **Go Package** - `docs/swagger/docs.go`

## 🛠️ Development Workflow

1. **Add/Update Endpoints** - Modify your handlers and add swagger annotations
2. **Update Models** - Define or update request/response models
3. **Regenerate Docs** - Run `make docs` to update documentation
4. **Test Locally** - Use `make docs-serve` to view in Swagger UI
5. **Validate** - Run `make docs-validate` to check for errors

## 🚨 Common Issues

### Swag CLI Not Found
```bash
go install github.com/swaggo/swag/cmd/swag@latest
export PATH=$PATH:$(go env GOPATH)/bin
```

### Documentation Not Updating
- Ensure you've added proper swagger annotations
- Check that your models are properly exported
- Verify that your handlers are being imported

### Validation Errors
- Check for missing required fields in annotations
- Ensure all referenced models exist
- Verify that response schemas are correct

## 📚 Additional Resources

- [Swaggo Documentation](https://github.com/swaggo/swag)
- [Echo Swagger](https://github.com/swaggo/echo-swagger)
- [OpenAPI Specification](https://swagger.io/specification/)
- [Swagger UI](https://swagger.io/tools/swagger-ui/)

## 🤝 Contributing

When contributing to the API:

1. Add comprehensive swagger annotations to new endpoints
2. Update this README if adding new features
3. Test the generated documentation locally
4. Ensure all examples are accurate and up-to-date
