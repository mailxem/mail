# Plan Management System

This document describes the plan management system for the Diagonal Payments service. The system allows administrators to dynamically create, update, and manage subscription plans through a RESTful API.

## Overview

The plan management system provides:
- **Dynamic Plan Creation**: Create new subscription plans without code changes
- **Admin-Only Access**: Secure plan management using secret header authentication
- **Public Plan Display**: Public API for retrieving active plans
- **Comprehensive Features**: Rich plan configuration including pricing, limits, features, and metadata
- **Database-Backed**: All plans stored in database with proper indexing and constraints
- **Backward Compatibility**: Maintains compatibility with existing static plan definitions

## Authentication

Admin operations require authentication using a secret header:

```bash
curl -H "X-Admin-Secret: your-admin-secret" \
     -X POST https://api.diagonal.com/api/admin/plans
```

The admin secret is configured via the `ADMIN_SECRET` environment variable.

## API Endpoints

### Public Endpoints (No Authentication)

#### List Active Plans
```http
GET /api/plans
```

Returns all active plans visible to customers.

**Response:**
```json
{
  "plans": [
    {
      "id": "starter",
      "dodo_plan_id": "starter-plan-dodo",
      "name": "STARTER",
      "description": "Free forever for personal use",
      "status": "active",
      "price_monthly": "0.00",
      "price_yearly": "0.00",
      "currency": "USD",
      "queries_limit": 10,
      "data_sources_limit": 3,
      "max_team_members": 1,
      "trial_enabled": true,
      "trial_period_days": 7,
      "features": ["Single user", "3 data sources", "10 queries/week"],
      "is_popular": false,
      "display_order": 1
    }
  ]
}
```

#### Get Plan by ID
```http
GET /api/plans/{planId}
```

#### Get Plan by Dodo Plan ID
```http
GET /api/plans/dodo/{dodoPlanId}
```

### Admin Endpoints (Require X-Admin-Secret Header)

#### Create Plan
```http
POST /api/admin/plans
Content-Type: application/json
X-Admin-Secret: your-admin-secret
```

**Request Body:**
```json
{
  "dodo_plan_id": "pro-plan-2024",
  "name": "Professional",
  "description": "For growing teams and businesses",
  "price_monthly": "49.00",
  "price_yearly": "490.00",
  "currency": "USD",
  "queries_limit": null,
  "data_sources_limit": 50,
  "max_team_members": 10,
  "api_access_enabled": true,
  "advanced_analytics": true,
  "priority_support": true,
  "trial_enabled": true,
  "trial_period_days": 14,
  "display_order": 2,
  "is_popular": true,
  "features": [
    "Unlimited queries",
    "50 data sources",
    "Up to 10 team members",
    "Advanced analytics",
    "Priority support",
    "API access"
  ],
  "metadata": {
    "category": "professional",
    "target_audience": "growing_teams"
  }
}
```

#### Update Plan
```http
PUT /api/admin/plans/{planId}
Content-Type: application/json
X-Admin-Secret: your-admin-secret
```

**Request Body (partial updates supported):**
```json
{
  "name": "Professional Plus",
  "price_monthly": "59.00",
  "data_sources_limit": 75,
  "features": [
    "Unlimited queries",
    "75 data sources",
    "Up to 10 team members",
    "Advanced analytics",
    "Priority support",
    "API access",
    "Custom integrations"
  ]
}
```

#### List All Plans (Admin)
```http
GET /api/admin/plans?include_inactive=true
X-Admin-Secret: your-admin-secret
```

#### Get Plan with Statistics
```http
GET /api/admin/plans/{planId}/stats
X-Admin-Secret: your-admin-secret
```

**Response:**
```json
{
  "plan": {
    "id": "professional",
    "name": "Professional",
    "active_subscriptions": 25,
    // ... other plan fields
  }
}
```

#### Archive Plan
```http
POST /api/admin/plans/{planId}/archive
X-Admin-Secret: your-admin-secret
```

#### Delete Plan
```http
DELETE /api/admin/plans/{planId}
X-Admin-Secret: your-admin-secret
```

**Note:** Plans with active subscriptions cannot be deleted. Archive them instead.

## Plan Configuration

### Required Fields
- `dodo_plan_id`: Unique identifier for Dodo Payments integration
- `name`: Display name for the plan
- `price_monthly`: Monthly price in decimal format
- `price_yearly`: Yearly price in decimal format
- `currency`: 3-letter currency code (e.g., "USD")
- `data_sources_limit`: Maximum number of data sources

### Optional Fields
- `description`: Plan description
- `queries_limit`: Maximum queries per month (null = unlimited)
- `max_team_members`: Maximum team size (null = unlimited)
- `api_access_enabled`: Enable API access (default: true)
- `advanced_analytics`: Enable advanced analytics (default: false)
- `priority_support`: Enable priority support (default: false)
- `custom_integrations`: Enable custom integrations (default: false)
- `white_label_enabled`: Enable white labeling (default: false)
- `trial_enabled`: Enable trial period (default: true)
- `trial_period_days`: Trial period in days (default: 7)
- `display_order`: Display order for sorting (default: 0)
- `is_popular`: Mark as popular plan (default: false)
- `is_enterprise`: Mark as enterprise plan (default: false)
- `features`: Array of feature descriptions
- `metadata`: JSON object for additional data

### Plan Status
Plans can have the following statuses:
- `active`: Available for new subscriptions
- `inactive`: Hidden but existing subscriptions continue
- `archived`: Permanently hidden, no new subscriptions

## Database Schema

The plans are stored in the `plans` table with the following structure:

```sql
CREATE TABLE plans (
    id VARCHAR(255) PRIMARY KEY,
    dodo_plan_id VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) DEFAULT 'active',
    
    -- Pricing
    price_monthly DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    price_yearly DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(3) DEFAULT 'USD',
    
    -- Features & Limits
    queries_limit INTEGER NULL,
    data_sources_limit INTEGER NOT NULL DEFAULT 0,
    max_team_members INTEGER NULL,
    api_access_enabled BOOLEAN DEFAULT true,
    advanced_analytics BOOLEAN DEFAULT false,
    priority_support BOOLEAN DEFAULT false,
    custom_integrations BOOLEAN DEFAULT false,
    white_label_enabled BOOLEAN DEFAULT false,
    
    -- Trial settings
    trial_period_days INTEGER DEFAULT 7,
    trial_enabled BOOLEAN DEFAULT true,
    
    -- Display & Ordering
    display_order INTEGER DEFAULT 0,
    is_popular BOOLEAN DEFAULT false,
    is_enterprise BOOLEAN DEFAULT false,
    features TEXT[],
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

## Examples

### Creating a Basic Plan
```bash
curl -X POST https://api.diagonal.com/api/admin/plans \
  -H "Content-Type: application/json" \
  -H "X-Admin-Secret: your-admin-secret" \
  -d '{
    "dodo_plan_id": "basic-2024",
    "name": "Basic",
    "description": "Perfect for individuals",
    "price_monthly": "19.00",
    "price_yearly": "190.00",
    "currency": "USD",
    "queries_limit": 1000,
    "data_sources_limit": 5,
    "max_team_members": 1,
    "features": [
      "1,000 queries/month",
      "5 data sources",
      "Email support"
    ]
  }'
```

### Creating an Enterprise Plan
```bash
curl -X POST https://api.diagonal.com/api/admin/plans \
  -H "Content-Type: application/json" \
  -H "X-Admin-Secret: your-admin-secret" \
  -d '{
    "dodo_plan_id": "enterprise-2024",
    "name": "Enterprise",
    "description": "For large organizations",
    "price_monthly": "0.00",
    "price_yearly": "0.00",
    "currency": "USD",
    "queries_limit": null,
    "data_sources_limit": -1,
    "max_team_members": null,
    "api_access_enabled": true,
    "advanced_analytics": true,
    "priority_support": true,
    "custom_integrations": true,
    "white_label_enabled": true,
    "trial_enabled": false,
    "is_enterprise": true,
    "features": [
      "Unlimited queries",
      "Unlimited data sources",
      "Unlimited team members",
      "Dedicated support",
      "Custom integrations",
      "White labeling",
      "SLA guarantee"
    ],
    "metadata": {
      "contact_required": true,
      "custom_pricing": true
    }
  }'
```

### Updating Plan Pricing
```bash
curl -X PUT https://api.diagonal.com/api/admin/plans/professional \
  -H "Content-Type: application/json" \
  -H "X-Admin-Secret: your-admin-secret" \
  -d '{
    "price_monthly": "59.00",
    "price_yearly": "590.00"
  }'
```

## Environment Variables

Add these environment variables to your `.env` file:

```env
# Admin secret for plan management
ADMIN_SECRET=your-secure-admin-secret-here

# Other existing variables...
```

## Security Considerations

1. **Admin Secret**: Use a strong, randomly generated admin secret
2. **HTTPS Only**: Always use HTTPS in production for admin operations
3. **Access Logging**: Monitor admin API usage
4. **Rate Limiting**: Consider implementing rate limiting for admin endpoints
5. **Audit Trail**: All plan changes are logged with timestamps

## Migration from Static Plans

The system includes a migration that preserves existing static plans as database records. The legacy functions (`GetPlan`, `GetAllPlans`) are kept for backward compatibility but should be replaced with service calls in new code.

## Error Handling

The API returns appropriate HTTP status codes:
- `200`: Success
- `201`: Created (for new plans)
- `400`: Bad Request (validation errors)
- `401`: Unauthorized (missing admin secret)
- `403`: Forbidden (invalid admin secret)
- `404`: Not Found
- `500`: Internal Server Error

Error responses include descriptive messages:
```json
{
  "error": "Plan with Dodo Plan ID pro-plan-2024 already exists"
}
``` 