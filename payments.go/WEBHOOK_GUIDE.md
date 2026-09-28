# Dodo Payments Webhook Integration Guide

This document explains the comprehensive webhook handling system implemented to process all Dodo Payments webhook events.

## Overview

The webhook system provides:
- ✅ **Complete Event Coverage**: Handles all Dodo Payments webhook events (payment, subscription, refund, dispute, license key)
- ✅ **Database Persistence**: Stores all webhook events and processing logs
- ✅ **Signature Verification**: Validates webhook authenticity using HMAC-SHA256
- ✅ **Error Handling & Retry**: Automatic retry for failed events
- ✅ **Comprehensive Logging**: Detailed logging for debugging and monitoring
- ✅ **API Management**: REST endpoints for webhook management

## Supported Webhook Events

### Payment Events
- `payment.succeeded` - Payment completed successfully
- `payment.failed` - Payment attempt failed
- `payment.processing` - Payment is being processed
- `payment.cancelled` - Payment was cancelled

### Subscription Events
- `subscription.active` - Subscription is now active
- `subscription.on_hold` - Subscription on hold due to failed payment
- `subscription.renewed` - Subscription successfully renewed
- `subscription.paused` - Subscription paused
- `subscription.plan_changed` - Subscription plan upgraded/downgraded
- `subscription.cancelled` - Subscription cancelled
- `subscription.failed` - Subscription creation failed
- `subscription.expired` - Subscription expired

### Refund Events
- `refund.succeeded` - Refund processed successfully
- `refund.failed` - Refund attempt failed

### Dispute Events
- `dispute.opened` - Customer initiated a dispute
- `dispute.expired` - Dispute expired without resolution
- `dispute.accepted` - Merchant accepted the dispute
- `dispute.cancelled` - Dispute was cancelled
- `dispute.challenged` - Merchant challenged the dispute
- `dispute.won` - Merchant won the dispute
- `dispute.lost` - Merchant lost the dispute

### License Key Events
- `license_key.created` - New license key created

## Architecture

### Core Components

1. **Webhook Models** (`internal/models/webhook.go`)
   - `WebhookEvent`: Stores webhook events
   - `WebhookLog`: Stores processing logs
   - Event-specific data structures for parsing

2. **Repositories** (`internal/repositories/webhook_repository.go`)
   - `WebhookRepository`: Manages webhook events
   - `WebhookLogRepository`: Manages webhook logs

3. **Service Layer** (`internal/services/webhook_service.go`)
   - `WebhookService`: Core webhook processing logic
   - Event routing and business logic handling

4. **API Handler** (`internal/handlers/webhook_handler.go`)
   - HTTP endpoints for webhook processing and management

5. **Database Migration** (`migrations/001_add_webhook_tables.sql`)
   - Creates necessary database tables

## Database Schema

### webhook_events Table
```sql
CREATE TABLE webhook_events (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(100) NOT NULL,
    data JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE NULL,
    status VARCHAR(20) DEFAULT 'pending',
    error TEXT NULL,
    signature VARCHAR(255) NOT NULL,
    raw_body TEXT NOT NULL
);
```

### webhook_logs Table
```sql
CREATE TABLE webhook_logs (
    id VARCHAR(255) PRIMARY KEY,
    event_id VARCHAR(255) NOT NULL,
    level VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    data JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed_by VARCHAR(100) NOT NULL,
    FOREIGN KEY (event_id) REFERENCES webhook_events(id)
);
```

## API Endpoints

### POST /webhooks/dodo
**Process Dodo Payments webhook**
- Headers: `X-Signature` or `X-Dodo-Signature` (webhook signature)
- Body: Raw webhook payload
- Response: `{"status": "success", "message": "Webhook processed successfully"}`

### GET /webhooks/events
**Get webhook events**
- Query params: `type`, `limit`, `offset`
- Response: List of webhook events with pagination

### GET /webhooks/events/:eventId/logs
**Get logs for specific event**
- Response: List of logs for the event

### POST /webhooks/retry-failed
**Retry failed webhook events**
- Query params: `limit` (max 50)
- Response: Confirmation of retry operation

### POST /webhooks/cleanup
**Clean up old webhook data**
- Query params: `retention_days` (min 7)
- Response: Confirmation of cleanup operation

## Usage Examples

### Setting Up Webhook Processing

```go
// Initialize repositories
webhookRepo := repositories.NewWebhookRepository(db)
webhookLogRepo := repositories.NewWebhookLogRepository(db)
subscriptionRepo := repositories.NewSubscriptionRepository(db)
paymentRepo := repositories.NewPaymentRepository(db)

// Create webhook service
webhookService := services.NewWebhookService(
    webhookRepo,
    webhookLogRepo,
    subscriptionRepo,
    paymentRepo,
    webhookSecret, // From config
)

// Create webhook handler
webhookHandler := handlers.NewWebhookHandler(webhookService)

// Register routes
e.POST("/webhooks/dodo", webhookHandler.HandleDodoWebhook)
e.GET("/webhooks/events", webhookHandler.GetWebhookEvents)
e.GET("/webhooks/events/:eventId/logs", webhookHandler.GetWebhookLogs)
e.POST("/webhooks/retry-failed", webhookHandler.RetryFailedWebhooks)
e.POST("/webhooks/cleanup", webhookHandler.CleanupOldWebhooks)
```

### Webhook Configuration in Dodo Payments

1. Set webhook endpoint: `https://your-domain.com/webhooks/dodo`
2. Configure webhook secret for signature verification
3. Select events to receive (recommend all for comprehensive logging)

### Monitoring and Maintenance

```go
// Retry failed events (run periodically)
err := webhookService.RetryFailedEvents(ctx, 10)

// Cleanup old data (run daily/weekly)
err := webhookService.CleanupOldEvents(ctx, 30) // Keep 30 days

// Get event statistics
events, err := webhookService.GetWebhookEvents(ctx, "", 100, 0)
```

## Event Processing Flow

1. **Webhook Received**
   - Parse JSON payload
   - Extract event type
   - Verify signature

2. **Event Storage**
   - Save raw event to database
   - Set status to `pending`
   - Log reception

3. **Event Processing**
   - Route to appropriate handler
   - Update local records (subscriptions, payments)
   - Log processing steps

4. **Completion**
   - Mark event as `processed`
   - Log success/failure
   - Return HTTP 200 to Dodo

## Error Handling

- **Invalid Signature**: HTTP 400, event not saved
- **Parse Error**: HTTP 400, event not saved
- **Processing Error**: HTTP 500, event marked as `failed`
- **Unknown Event**: Event marked as `skipped`

## Security

- ✅ HMAC-SHA256 signature verification
- ✅ Raw payload preservation for verification
- ✅ Configurable webhook secret
- ✅ Input validation and sanitization

## Performance Considerations

- Database indexes on commonly queried fields
- Configurable retention policies
- Batch processing for retries
- Async processing capability (can be added)

## Monitoring & Debugging

### Check Event Status
```sql
SELECT type, status, COUNT(*) 
FROM webhook_events 
GROUP BY type, status;
```

### View Failed Events
```sql
SELECT id, type, error, created_at 
FROM webhook_events 
WHERE status = 'failed'
ORDER BY created_at DESC;
```

### Event Processing Logs
```sql
SELECT level, message, created_at 
FROM webhook_logs 
WHERE event_id = 'event-id'
ORDER BY created_at ASC;
```

## Configuration

Add to your environment variables:
```bash
DODO_WEBHOOK_SECRET=your-webhook-secret-from-dodo-dashboard
```

The webhook secret is used for signature verification to ensure webhooks are from Dodo Payments.

## Migration

Run the database migration to create the webhook tables:
```bash
# Apply migration
psql -d your_database -f migrations/001_add_webhook_tables.sql
```

## Next Steps

1. Deploy the webhook endpoint
2. Configure the webhook URL in Dodo Payments dashboard
3. Test with a few transactions
4. Monitor webhook events and logs
5. Set up alerting for failed events
6. Configure automated cleanup jobs

The system is now ready to handle all Dodo Payments webhook events comprehensively! 