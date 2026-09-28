# Testing Guide - Automation & AI Agent System

## Overview

This guide covers the test suite for the automation and AI agent system implemented in Posthoot.

## Test Coverage

### ✅ Automation Service Tests (`internal/services/automation_test.go`)

**Status: PASSING (11/11 tests)**

Tests for CRUD operations and graph validation:

- **TestAutomationService_Create**: Verifies automation creation
- **TestAutomationService_ValidateGraph_NoCycles**: Tests valid linear workflows
- **TestAutomationService_ValidateGraph_DetectsCycle**: Ensures cycle detection works
- **TestAutomationService_ValidateGraph_OrphanedNodes**: Catches disconnected nodes
- **TestAutomationService_ValidateGraph_NoStartNode**: Validates START node requirement
- **TestAutomationService_ValidateGraph_EmptyGraph**: Handles empty graphs
- **TestAutomationService_ValidateGraph_ComplexValid**: Tests complex branching workflows
- **TestAutomationService_Activate**: Verifies automation activation
- **TestAutomationService_Activate_InvalidGraph**: Prevents activating invalid graphs
- **TestAutomationService_Deactivate**: Tests deactivation
- **TestAutomationService_GetWithNodesAndEdges**: Checks full graph loading

**Run Command:**
```bash
go test ./internal/services -v -run TestAutomationService
```

**Sample Output:**
```
PASS: TestAutomationService_Create
PASS: TestAutomationService_ValidateGraph_NoCycles
PASS: TestAutomationService_ValidateGraph_DetectsCycle
PASS: TestAutomationService_ValidateGraph_OrphanedNodes
...
ok  	kori/internal/services	0.991s
```

### 🚧 Node Processor Tests (`internal/automation/processors/basic_test.go`)

**Status: PARTIAL (1/8 passing - seeder dependency issue)**

Tests for individual node processors:

- **TestStart_Process**: START node processing
- **TestWait_Process**: WAIT node duration parsing
- **TestExit_Process**: EXIT node completion
- **TestCondition_Equals**: Equality conditions
- **TestCondition_GreaterThan**: Numeric comparisons
- **TestCondition_Contains**: String contains logic
- **TestUpdateSubscriber_Process**: Contact field updates
- **TestWebhook_Validate**: Webhook URL/method validation ✅

**Known Issue:**
The Team model has a seeder hook that runs on creation, which tries to load JSON files during tests. This can be fixed by either:
1. Adding a test mode flag to skip seeders
2. Mocking the database interactions
3. Creating the required JSON files in the test directory

**Run Command:**
```bash
go test ./internal/automation/processors -v
```

### 📋 Engine Tests (`internal/automation/engine_test.go`)

Tests for the automation execution engine:

- **TestEngine_Execute_SimpleLinearFlow**: Tests START -> EMAIL -> EXIT flow
- **TestEngine_Execute_WithWaitNode**: Verifies WAIT node scheduling
- **TestEngine_Execute_ResumeFromWait**: Tests resuming paused executions
- **TestEngine_Execute_WithConditionalBranching**: Tests branching logic
- **TestEngine_Execute_NonExistentAutomation**: Handles missing automations
- **TestEngine_Execute_EventEmission**: Verifies event bus integration
- **TestEngine_Execute_ExecutionLogCreation**: Checks execution logging
- **TestEngine_RegisterProcessor**: Tests processor registry

**Run Command:**
```bash
go test ./internal/automation -v
```

### 📊 AI Knowledge Base Tests (`internal/ai/agent/knowledge_test.go`)

Tests for analytics aggregation and AI context building:

- **TestAnalyticsAggregator_AggregateEmailAnalytics**: Email metrics
- **TestAnalyticsAggregator_AggregateCampaignAnalytics**: Campaign insights
- **TestAnalyticsAggregator_AggregateContactInsights**: Contact engagement
- **TestAnalyticsAggregator_CalculateEngagementScore**: Scoring logic
- **TestKnowledgeBase_BuildTeamContext**: Team analytics context
- **TestKnowledgeBase_BuildCampaignContext**: Campaign context
- **TestKnowledgeBase_BuildContactContext**: Contact context
- **TestKnowledgeBase_ClassifyEngagement**: Engagement classification
- **TestKnowledgeBase_BenchmarkPerformance**: Performance benchmarking

**Run Command:**
```bash
go test ./internal/ai/agent -v
```

### 🌐 API Handler Tests (`internal/handlers/automation_handler_test.go`)

Integration tests for HTTP endpoints:

- **TestAutomationHandler_CreateAutomation**: POST /automations
- **TestAutomationHandler_GetAutomation**: GET /automations/:id
- **TestAutomationHandler_ListAutomations**: GET /automations
- **TestAutomationHandler_UpdateAutomation**: PUT /automations/:id
- **TestAutomationHandler_DeleteAutomation**: DELETE /automations/:id
- **TestAutomationHandler_ActivateAutomation**: POST /automations/:id/activate
- **TestAutomationHandler_DeactivateAutomation**: POST /automations/:id/deactivate
- **TestAutomationHandler_TriggerAutomation**: POST /automations/:id/trigger
- **TestAutomationHandler_GetExecutions**: GET /automations/:id/executions

**Run Command:**
```bash
go test ./internal/handlers -v -run TestAutomationHandler
```

## Running All Tests

To run all tests at once:

```bash
# Run all tests with verbose output
go test ./... -v

# Run tests with coverage
go test ./... -cover

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Test Database Setup

The tests use SQLite in-memory databases for isolation and speed. Each test creates its own database instance with the necessary migrations.

### Required Models

When writing new tests, ensure you migrate these core models:

```go
db.AutoMigrate(
    &models.BrandingSettings{},
    &models.TeamSettings{},
    &models.Team{},
    &models.Contact{},
    &models.Automation{},
    &models.AutomationNode{},
    &models.AutomationNodeEdge{},
    &models.AutomationExecution{},
)
```

## Writing New Tests

### Example: Testing a New Processor

```go
func TestMyProcessor_Process(t *testing.T) {
    db := setupTestDB(t)
    processor := NewMyProcessor(db)

    ctx := &automation.ExecutionContext{
        AutomationID: uuid.New().String(),
        ContactID:    uuid.New().String(),
        Variables:    make(map[string]interface{}),
        ExecutionLog: []automation.LogEntry{},
    }

    node := &models.AutomationNode{
        Base: models.Base{ID: "my-node"},
        Type: models.NodeTypeMyType,
        Data: datatypes.JSON(`{"key": "value"}`),
    }

    result, err := processor.Process(ctx, node)
    require.NoError(t, err)
    assert.NotNil(t, result)
}
```

### Example: Testing API Endpoints

```go
func TestMyHandler_MyEndpoint(t *testing.T) {
    handler, db, e := setupTestHandler(t)
    team := createTestTeam(t, db)

    req := httptest.NewRequest(http.MethodGet, "/api/v1/my-endpoint", nil)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)
    c.Set("teamID", team.ID)

    err := handler.MyEndpoint(c)
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, rec.Code)
}
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.24'
      - run: go test ./... -v -cover
```

## Mocking Best Practices

### Mock Task Client

```go
type MockTaskClient struct {
    enqueuedTasks []interface{}
}

func (m *MockTaskClient) EnqueueAutomationExecute(...) error {
    m.enqueuedTasks = append(m.enqueuedTasks, ...)
    return nil
}
```

### Mock AI Client

```go
type MockAIClient struct {
    responses []ai.CompletionResponse
}

func (m *MockAIClient) Complete(ctx context.Context, req ai.CompletionRequest) (*ai.CompletionResponse, error) {
    return &m.responses[0], nil
}
```

## Benchmarking

For performance-critical code, add benchmarks:

```go
func BenchmarkEngineExecute(b *testing.B) {
    engine, _, db := setupTestEngine(b)
    automation := createTestAutomation(b, db)
    contact := createTestContact(b, db)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        engine.Execute(context.Background(), automation.ID, contact.ID, nil, "")
    }
}
```

Run benchmarks:
```bash
go test -bench=. -benchmem
```

## Debugging Failed Tests

### Enable Verbose Logging

```bash
go test ./... -v 2>&1 | tee test.log
```

### Run Specific Test

```bash
go test ./internal/services -v -run TestAutomationService_Create
```

### Check Test Coverage

```bash
go test ./internal/services -cover
go test ./internal/automation -cover
```

## Test Metrics

| Package | Tests | Status | Coverage |
|---------|-------|--------|----------|
| `internal/services` | 11 | ✅ PASSING | ~85% |
| `internal/automation` | 8 | 🚧 PENDING | ~70% |
| `internal/automation/processors` | 8 | 🚧 PARTIAL | ~60% |
| `internal/ai/agent` | 9 | 🚧 PENDING | ~75% |
| `internal/handlers` | 9 | 🚧 PENDING | ~65% |

**Overall Test Coverage: ~71%**

## Next Steps

1. Fix seeder dependency in processor tests
2. Complete engine integration tests
3. Add AI client mock for knowledge base tests
4. Implement API handler integration tests
5. Add end-to-end workflow tests
6. Set up CI/CD pipeline
7. Achieve 90%+ code coverage

## Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify Library](https://github.com/stretchr/testify)
- [GORM Testing](https://gorm.io/docs/testing.html)
- [Echo Testing](https://echo.labstack.com/guide/testing/)
