# Xem Payments Service

A comprehensive Go-based microservice for managing subscriptions, usage tracking, and payments integration with Dodo Payments. This service provides a clean REST API for handling billing operations in the mailxem platform.

## Features

- **Subscription Management**: Create, update, cancel, and manage team subscriptions
- **Usage Tracking**: Track query usage with automatic billing period calculations
- **Payment Processing**: Integration with Dodo Payments for secure payment handling
- **Plan Management**: Support for Free, Syne Scale, and Enterprise plans
- **RESTful API**: Clean HTTP endpoints for frontend integration
- **PostgreSQL Storage**: Reliable data persistence with proper schema design

## Architecture

```
├── cmd/                    # Application entrypoints
│   └── main.go            # Main server application
├── internal/              # Private application code
│   ├── config/           # Configuration management
│   ├── handlers/         # HTTP request handlers
│   ├── models/          # Data models and structures
│   ├── repositories/    # Data access layer
│   └── services/        # Business logic layer
├── pkg/                  # Public packages
│   └── dodo/            # Dodo Payments client
├── Dockerfile           # Container configuration
└── README.md           # This file
```

## Subscription Plans

### 🆓 Free (Starter)
- **Price**: $0/month
- **Query Limit**: 10 queries/week
- **Features**: 
  - Single user
  - 3 data sources
  - Basic support

### 💼 Syne Scale  
- **Price**: $29/seat/month
- **Query Limit**: Unlimited
- **Features**:
  - Unlimited queries per month
  - 24/7 support with founder
  - White glove onboarding
  - Custom features
  - 20+ data sources

### 🏢 Enterprise
- **Price**: Custom pricing
- **Query Limit**: Unlimited
- **Features**:
  - Everything in Syne Scale
  - Custom integrations
  - Dedicated support
  - SLA guarantees
  - Advanced security features

## API Endpoints

### Subscriptions

#### Create Subscription
```http
POST /api/v1/subscriptions
Content-Type: application/json

{
  "team_id": "team_123",
  "plan_type": "syne_scale",
  "total_seats": 5
}
```

#### Get Subscription
```http
GET /api/v1/subscriptions/team/{teamId}
```

#### Update Subscription
```http
PUT /api/v1/subscriptions/{subscriptionId}
Content-Type: application/json

{
  "total_seats": 10,
  "plan_type": "enterprise"
}
```

#### Cancel Subscription
```http
DELETE /api/v1/subscriptions/{subscriptionId}
```

### Usage Tracking

#### Record Usage
```http
POST /api/v1/usage/{teamId}
Content-Type: application/json

{
  "queries_count": 5
}
```

#### Check Usage Limit
```http
GET /api/v1/usage/{teamId}/check
```

#### Get Usage Stats
```http
GET /api/v1/usage/{subscriptionId}/stats
```

#### Get Usage History
```http
GET /api/v1/usage/{subscriptionId}/history
```

### Payments

#### Create Payment
```http
POST /api/v1/payments
Content-Type: application/json

{
  "subscription_id": "sub_123",
  "amount": 2900,
  "currency": "USD",
  "description": "Monthly subscription - 5 seats"
}
```

#### Get Payment History
```http
GET /api/v1/payments/subscription/{subscriptionId}
```

### Plans

#### Get Available Plans
```http
GET /api/v1/plans
```

## Environment Variables

Create a `.env` file based on `.env.example`:

```bash
# Server Configuration
PORT=8080
HOST=localhost

# Database Configuration
DATABASE_URL=postgres://username:password@localhost:5432/payments_db

# Dodo Payments Configuration
DODO_API_KEY=your_dodo_api_key
DODO_SECRET_KEY=your_dodo_secret_key
DODO_BASE_URL=https://api.dodo.com

# Environment
ENVIRONMENT=development
```

## Database Schema

### Subscriptions Table
```sql
CREATE TABLE subscriptions (
    id VARCHAR(255) PRIMARY KEY,
    team_id VARCHAR(255) NOT NULL,
    plan_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    total_seats INTEGER NOT NULL,
    price_per_seat INTEGER NOT NULL,
    billing_cycle VARCHAR(50) NOT NULL,
    next_billing_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Usage History Table
```sql
CREATE TABLE usage_history (
    id SERIAL PRIMARY KEY,
    subscription_id VARCHAR(255) NOT NULL,
    period VARCHAR(50) NOT NULL,
    queries_count INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (subscription_id) REFERENCES subscriptions(id)
);
```

### Payments Table
```sql
CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    subscription_id VARCHAR(255) NOT NULL,
    dodo_payment_id VARCHAR(255) UNIQUE NOT NULL,
    amount INTEGER NOT NULL,
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP,
    FOREIGN KEY (subscription_id) REFERENCES subscriptions(id)
);
```

## Running the Service

### Development
```bash
# Install dependencies
go mod tidy

# Run the service
go run cmd/main.go
```

### Using Docker
```bash
# Build the image
docker build -t payments-service .

# Run the container
docker run -p 8080:8080 --env-file .env payments-service
```

### Using Docker Compose
```bash
# Start all services (includes PostgreSQL)
docker-compose up -d
```

## Frontend Integration

The service includes a TypeScript client for easy integration with the mailxem frontend:

```typescript
import { usePaymentsClient } from '@/lib/payments-client';

const paymentsClient = usePaymentsClient();

// Create a subscription
const subscription = await paymentsClient.createSubscription({
  team_id: 'team_123',
  plan_type: 'syne_scale',
  total_seats: 5
});

// Record usage
await paymentsClient.recordUsage('team_123', 3);

// Check usage limits
const canQuery = await paymentsClient.checkUsageLimit('team_123');
```

### React Components

The service comes with pre-built React components for the billing interface:

- `SubscriptionOverview`: Display current subscription details
- `PlanCard`: Show available plans with upgrade options
- `UsageChart`: Visualize usage over time
- `PaymentHistory`: Display payment transactions

### Usage Tracking Hook

Automatically track query usage in your application:

```typescript
import { useQueryWithUsageTracking } from '@/hooks/useSubscriptionCheck';

const { executeQueryWithTracking, canExecuteQuery } = useQueryWithUsageTracking({
  teamId: 'team_123',
  onUsageLimitReached: () => {
    // Show upgrade prompt
  }
});

// Execute a query with automatic usage tracking
const result = await executeQueryWithTracking(async () => {
  return await executeSQL(query);
});
```

## Testing

### Unit Tests
```bash
go test ./...
```

### Integration Tests
```bash
# Requires running PostgreSQL instance
go test ./... -tags=integration
```

### API Testing
Use the provided test scripts:
```bash
# Test subscription creation
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Content-Type: application/json" \
  -d '{"team_id":"team_123","plan_type":"syne_scale","total_seats":5}'
```

## Monitoring and Logging

The service includes structured logging with request tracing:

```go
log.WithFields(log.Fields{
    "team_id": teamID,
    "subscription_id": subscription.ID,
    "queries": count,
}).Info("Usage recorded successfully")
```

## Security Considerations

- All API endpoints should be protected with authentication
- Sensitive payment data is handled securely through Dodo Payments
- Database credentials are stored in environment variables
- Input validation is performed on all endpoints
- Rate limiting should be implemented for production use

## Deployment

### Production Checklist
- [ ] Set up production database with proper backup strategy
- [ ] Configure environment variables securely
- [ ] Set up monitoring and alerting
- [ ] Implement health checks
- [ ] Configure load balancing if needed
- [ ] Set up SSL/TLS certificates
- [ ] Configure rate limiting and DDoS protection

### Scaling Considerations
- Database connection pooling
- Redis caching for frequently accessed data
- Horizontal scaling with load balancers
- Queue-based processing for heavy operations
- Database read replicas for improved performance

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

This project is part of the mailxem platform. All rights reserved.

## Support

For technical support or questions about the payments service:
- Create an issue in the repository
- Contact the mailxem development team
- Check the API documentation for detailed endpoint specifications

---

Built with ❤️ by the mailxem team using Go, Echo, PostgreSQL, and Dodo Payments.
