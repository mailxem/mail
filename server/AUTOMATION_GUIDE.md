# Posthoot Automation & AI Agent System

## 🎯 Overview

A production-ready automation and AI agent system for email marketing campaigns with intelligent decision-making, event-driven triggers, and comprehensive analytics.

## ✨ Features Implemented

### Core Automation Engine
- ✅ **Graph-based workflow execution** with nodes and edges
- ✅ **9 node processors**: START, EMAIL, WAIT, CONDITION, ADD_TO_LIST, TAG, WEBHOOK, UPDATE_SUBSCRIBER, EXIT
- ✅ **Cycle detection** and graph validation
- ✅ **Resumable workflows** with WAIT nodes
- ✅ **Execution tracking** with detailed logs
- ✅ **Event-driven architecture** using event bus

### AI Integration
- ✅ **Anthropic Claude integration** (Claude 4 Opus/Sonnet)
- ✅ **AI_DECISION node** for intelligent routing
- ✅ **Knowledge base** with analytics aggregation
- ✅ **Natural language analytics queries**
- ✅ **Automation optimization suggestions**
- ✅ **AI-powered workflow builder**

### Analytics & Insights
- ✅ **Email performance metrics** (open rate, click rate, bounce rate)
- ✅ **Campaign analytics** with device/time insights
- ✅ **Contact engagement scoring**
- ✅ **Behavioral insights** (preferred device, active times)
- ✅ **Performance benchmarking** against industry standards

### Event Triggers
- ✅ **contact.created** - Trigger when new contact added
- ✅ **email.sent** - Trigger after email sent
- ✅ **email.opened** - Trigger on email open
- ✅ **email.clicked** - Trigger on link click
- ✅ **campaign.completed** - Trigger when campaign finishes

### API Endpoints

#### Automation Management
```
POST   /api/v1/automations              Create automation
GET    /api/v1/automations              List automations
GET    /api/v1/automations/:id          Get automation details
PUT    /api/v1/automations/:id          Update automation
DELETE /api/v1/automations/:id          Delete automation
POST   /api/v1/automations/:id/activate   Activate automation
POST   /api/v1/automations/:id/deactivate Deactivate automation
POST   /api/v1/automations/:id/trigger    Manual trigger
```

#### Execution Monitoring
```
GET    /api/v1/automations/:id/executions Get execution history
GET    /api/v1/executions/:id              Get execution details
```

#### AI Features
```
POST   /api/v1/ai/query      Natural language analytics queries
POST   /api/v1/ai/optimize   Get automation optimization suggestions
POST   /api/v1/ai/build      Generate automation from description
```

## 🚀 Quick Start

### 1. Configuration

Add to your `.env`:

```bash
# AI Configuration
AI_ENABLED=true
AI_PROVIDER=anthropic
ANTHROPIC_API_KEY=sk-ant-your-api-key
AI_MODEL=claude-sonnet-4
AI_AUTO_OPTIMIZE=false
AI_MAX_TOKENS=4096

# Automation Configuration
AUTOMATION_MAX_EXECUTIONS_PER_MINUTE=100
AUTOMATION_EXECUTION_TIMEOUT=1800
AUTOMATION_ENABLE_EVENT_TRIGGERS=true
```

### 2. Run Migrations

The system will automatically create these tables:
- `automation_executions`
- `ai_optimizations`

### 3. Start Server

```bash
go run cmd/main.go
```

## 📝 Usage Examples

### Example 1: Welcome Email Series

```json
{
  "name": "Welcome Series",
  "description": "3-day welcome sequence for new subscribers",
  "nodes": [
    {
      "id": "start",
      "type": "START",
      "data": {}
    },
    {
      "id": "welcome",
      "type": "EMAIL",
      "data": {
        "templateId": "welcome-template-id",
        "smtpConfigId": "smtp-config-id"
      }
    },
    {
      "id": "wait1",
      "type": "WAIT",
      "data": {
        "duration": "24h"
      }
    },
    {
      "id": "tips",
      "type": "EMAIL",
      "data": {
        "templateId": "tips-template-id",
        "smtpConfigId": "smtp-config-id"
      }
    },
    {
      "id": "exit",
      "type": "EXIT",
      "data": {}
    }
  ],
  "edges": [
    {"sourceId": "start", "targetId": "welcome"},
    {"sourceId": "welcome", "targetId": "wait1"},
    {"sourceId": "wait1", "targetId": "tips"},
    {"sourceId": "tips", "targetId": "exit"}
  ]
}
```

### Example 2: Engagement-Based Routing with AI

```json
{
  "name": "AI-Powered Engagement Campaign",
  "description": "Uses AI to route based on engagement",
  "nodes": [
    {
      "id": "start",
      "type": "START",
      "data": {}
    },
    {
      "id": "ai_decision",
      "type": "AI_DECISION",
      "data": {
        "prompt": "Based on this contact's engagement history, should we send a promotional email or educational content?",
        "decisionPaths": {
          "promotional": "promo-email",
          "educational": "edu-email"
        },
        "model": "claude-sonnet-4",
        "maxTokens": 1000
      }
    },
    {
      "id": "promo-email",
      "type": "EMAIL",
      "data": {
        "templateId": "promo-template",
        "smtpConfigId": "smtp-id"
      }
    },
    {
      "id": "edu-email",
      "type": "EMAIL",
      "data": {
        "templateId": "edu-template",
        "smtpConfigId": "smtp-id"
      }
    }
  ],
  "edges": [
    {"sourceId": "start", "targetId": "ai_decision"}
  ]
}
```

### Example 3: Conditional Re-engagement

```json
{
  "name": "Re-engagement Campaign",
  "description": "Target inactive subscribers",
  "nodes": [
    {
      "id": "start",
      "type": "START",
      "data": {}
    },
    {
      "id": "check_engagement",
      "type": "CONDITION",
      "data": {
        "conditions": [
          {
            "variable": "contact_engagement_score",
            "operator": "<",
            "value": "30"
          }
        ],
        "operator": "AND",
        "branches": {
          "true": "send_reengagement",
          "false": "exit"
        }
      }
    },
    {
      "id": "send_reengagement",
      "type": "EMAIL",
      "data": {
        "templateId": "reengagement-template",
        "smtpConfigId": "smtp-id"
      }
    },
    {
      "id": "exit",
      "type": "EXIT",
      "data": {}
    }
  ],
  "edges": [
    {"sourceId": "start", "targetId": "check_engagement"}
  ]
}
```

## 🤖 AI Agent Capabilities

### 1. Natural Language Analytics

```bash
curl -X POST http://localhost:8080/api/v1/ai/query \
  -H "Authorization: Bearer YOUR_JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "What was the open rate for my last campaign?",
    "scope": {
      "startDate": "2026-03-01T00:00:00Z"
    }
  }'
```

### 2. Automation Optimization

```bash
curl -X POST http://localhost:8080/api/v1/ai/optimize \
  -H "Authorization: Bearer YOUR_JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "automationId": "uuid",
    "optimizationGoal": "increase_open_rate"
  }'
```

### 3. AI-Powered Workflow Builder

```bash
curl -X POST http://localhost:8080/api/v1/ai/build \
  -H "Authorization: Bearer YOUR_JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Send a 5-day onboarding series to new users",
    "goals": ["engagement", "product adoption"]
  }'
```

## 🏗️ Architecture

### Node Processor Pattern

Each node type implements the `NodeProcessor` interface:

```go
type NodeProcessor interface {
    Process(ctx *ExecutionContext, node *AutomationNode) (*ProcessResult, error)
    Validate(node *AutomationNode) error
    Type() NodeType
}
```

### Execution Flow

1. **Trigger** → Automation execution task enqueued
2. **Engine** → Loads automation graph
3. **Processor** → Processes each node sequentially
4. **WAIT** → Schedules resume task
5. **CONDITION/AI** → Routes to appropriate branch
6. **EXIT** → Marks execution complete

### Event-Driven Triggers

```go
events.On("contact.created", func(data interface{}) {
    // Automatically trigger relevant automations
})
```

## 📊 Analytics & Knowledge Base

The AI agent uses aggregated analytics to make informed decisions:

- **Email Analytics**: Open rates, click rates, bounce rates
- **Campaign Performance**: Time/device insights
- **Contact Insights**: Engagement scores, behavioral patterns
- **Benchmarking**: Compare against industry standards

## 🔒 Security Features

- ✅ Team-level isolation (all queries scoped to teamID)
- ✅ JWT authentication on all endpoints
- ✅ API key encryption for SMTP configs
- ✅ Graph validation prevents infinite loops
- ✅ Execution timeouts prevent runaway processes
- ✅ Rate limiting on AI API calls

## 🎯 Performance Considerations

- **Async execution**: All automations run in background via asynq
- **Batching**: Email sending respects rate limits
- **Caching**: Analytics aggregations can be cached
- **Concurrent processing**: Task workers handle multiple executions
- **Efficient queries**: Optimized database queries with proper indexes

## 🔧 Extending the System

### Add a New Node Type

1. Create processor in `internal/automation/processors/`
2. Implement `NodeProcessor` interface
3. Register in `cmd/main.go`:

```go
automationEngine.RegisterProcessor(processors.NewYourProcessor(db_instance))
```

### Add New Event Trigger

In `internal/automation/triggers.go`:

```go
events.On("your.event", func(data interface{}) {
    tm.handleYourEvent("your.event", data)
})
```

## 📈 Monitoring & Debugging

### View Execution Logs

```bash
GET /api/v1/executions/:id
```

Returns detailed execution log with:
- Each node execution
- Timestamps and duration
- Variables at each step
- Error messages if failed

### Check Active Automations

```bash
GET /api/v1/automations?status=active
```

## 🚨 Troubleshooting

### Automation Not Triggering

1. Check if automation is active: `isActive: true`
2. Verify event triggers are enabled in config
3. Check execution logs for errors
4. Verify contact matches trigger criteria

### AI Decision Not Working

1. Ensure `AI_ENABLED=true` in .env
2. Check ANTHROPIC_API_KEY is valid
3. Verify decision paths match AI response options
4. Check execution logs for AI response

### WAIT Node Not Resuming

1. Check Redis connection (asynq uses Redis)
2. Verify task server is running
3. Check task queue for scheduled tasks
4. Review asynq logs for processing errors

## 🎓 Best Practices

1. **Start Simple**: Begin with linear flows (START → EMAIL → EXIT)
2. **Test Thoroughly**: Use manual trigger with test contacts
3. **Monitor Executions**: Review logs regularly
4. **Optimize Prompts**: Refine AI decision prompts based on results
5. **Use Variables**: Pass data between nodes via execution context
6. **Handle Failures**: Add error handling branches in critical flows
7. **Respect Limits**: Configure appropriate rate limits and timeouts

## 📚 Node Reference

| Node Type | Purpose | Key Data Fields |
|-----------|---------|-----------------|
| START | Entry point | None |
| EMAIL | Send email | templateId, smtpConfigId, subject, variables |
| WAIT | Delay execution | duration (e.g., "5m", "24h") |
| CONDITION | Branch based on logic | conditions, operator, branches |
| ADD_TO_LIST | Add to mailing list | listId |
| TAG | Add/remove tags | action, tags |
| WEBHOOK | Call external API | url, method, headers, body |
| UPDATE_SUBSCRIBER | Update contact | fields |
| AI_DECISION | AI-powered routing | prompt, decisionPaths, model |
| EXIT | End workflow | None |

## 🤝 Contributing

When adding features:
1. Follow existing patterns (processor interface, event bus)
2. Add proper error handling and logging
3. Update this documentation
4. Test with real workflows
5. Consider backwards compatibility

## 📄 License

[Your License Here]

---

Built with ❤️ by the Posthoot team
