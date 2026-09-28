package automation

import (
	"context"
	"encoding/json"
	"kori/internal/events"
	"kori/internal/models"
	"kori/internal/tasks"
	"os"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Mock task client for testing
type MockTaskClient struct {
	enqueuedTasks []interface{}
}

func (m *MockTaskClient) EnqueueAutomationExecute(automationID, contactID string, triggerData map[string]interface{}, currentNodeID string, processIn time.Duration) error {
	m.enqueuedTasks = append(m.enqueuedTasks, map[string]interface{}{
		"automationID":  automationID,
		"contactID":     contactID,
		"triggerData":   triggerData,
		"currentNodeID": currentNodeID,
		"processIn":     processIn,
	})
	return nil
}

func setupTestDB(t *testing.T) *gorm.DB {
	// Set test mode to skip seeders
	os.Setenv("TEST_MODE", "true")

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&models.BrandingSettings{},
		&models.TeamSettings{},
		&models.Automation{},
		&models.AutomationNode{},
		&models.AutomationNodeEdge{},
		&models.AutomationExecution{},
		&models.Contact{},
		&models.Team{},
		&models.Tag{},
	)
	require.NoError(t, err)

	return db
}

func setupTestEngine(t *testing.T) (*Engine, *MockTaskClient, *gorm.DB) {
	db := setupTestDB(t)
	mockTaskClient := &MockTaskClient{enqueuedTasks: []interface{}{}}
	engine := NewEngine(db, mockTaskClient)

	// Register mock processors
	engine.RegisterProcessor(&MockStartProcessor{db: db})
	engine.RegisterProcessor(&MockEmailProcessor{db: db})
	engine.RegisterProcessor(&MockExitProcessor{})
	engine.RegisterProcessor(&MockWaitProcessor{db: db})
	engine.RegisterProcessor(&MockConditionProcessor{})

	return engine, mockTaskClient, db
}

// Mock Processors
type MockStartProcessor struct{ db *gorm.DB }

func (p *MockStartProcessor) Process(ctx *ExecutionContext, node *models.AutomationNode) (*ProcessResult, error) {
	return &ProcessResult{Complete: false, NextNodeIDs: testNextNodes(p.db, node)}, nil
}

func (p *MockStartProcessor) Validate(node *models.AutomationNode) error {
	return nil
}

func (p *MockStartProcessor) Type() models.NodeType {
	return models.NodeTypeStart
}

type MockEmailProcessor struct{ db *gorm.DB }

func (p *MockEmailProcessor) Process(ctx *ExecutionContext, node *models.AutomationNode) (*ProcessResult, error) {
	return &ProcessResult{Complete: false, NextNodeIDs: testNextNodes(p.db, node)}, nil
}

func (p *MockEmailProcessor) Validate(node *models.AutomationNode) error {
	return nil
}

func (p *MockEmailProcessor) Type() models.NodeType {
	return models.NodeTypeEmail
}

type MockExitProcessor struct{}

func (p *MockExitProcessor) Process(ctx *ExecutionContext, node *models.AutomationNode) (*ProcessResult, error) {
	return &ProcessResult{Complete: true}, nil
}

func (p *MockExitProcessor) Validate(node *models.AutomationNode) error {
	return nil
}

func (p *MockExitProcessor) Type() models.NodeType {
	return models.NodeTypeExit
}

type MockWaitProcessor struct{ db *gorm.DB }

func (p *MockWaitProcessor) Process(ctx *ExecutionContext, node *models.AutomationNode) (*ProcessResult, error) {
	duration := 5 * time.Minute
	return &ProcessResult{
		Complete:    false,
		Wait:        &duration,
		NextNodeIDs: testNextNodes(p.db, node),
	}, nil
}

func (p *MockWaitProcessor) Validate(node *models.AutomationNode) error {
	return nil
}

func (p *MockWaitProcessor) Type() models.NodeType {
	return models.NodeTypeWait
}

type MockConditionProcessor struct{}

func (p *MockConditionProcessor) Process(ctx *ExecutionContext, node *models.AutomationNode) (*ProcessResult, error) {
	// Parse node data
	var data struct {
		Branches map[string]string `json:"branches"`
	}
	if err := json.Unmarshal(node.Data, &data); err != nil {
		return nil, err
	}

	// Simple logic: route to true branch
	return &ProcessResult{
		Complete:    false,
		NextNodeIDs: []string{data.Branches["true"]},
	}, nil
}

func (p *MockConditionProcessor) Validate(node *models.AutomationNode) error {
	return nil
}

func (p *MockConditionProcessor) Type() models.NodeType {
	return models.NodeTypeCondition
}

// Helper function to create test automation
func createTestAutomation(t *testing.T, db *gorm.DB, nodes []models.AutomationNode, edges []models.AutomationNodeEdge) *models.Automation {
	teamID := uuid.New().String()
	team := &models.Team{
		Base: models.Base{ID: teamID},
		Name: "Test Team",
	}
	require.NoError(t, db.Create(team).Error)

	automation := &models.Automation{
		Base:        models.Base{ID: uuid.New().String()},
		TeamID:      teamID,
		Name:        "Test Automation",
		Description: "Test",
		IsActive:    true,
		Nodes:       nodes,
		Edges:       edges,
	}

	require.NoError(t, db.Create(automation).Error)
	return automation
}

func createTestContact(t *testing.T, db *gorm.DB, teamID string) *models.Contact {
	contact := &models.Contact{
		Base:      models.Base{ID: uuid.New().String()},
		TeamID:    teamID,
		Email:     "test@example.com",
		FirstName: "Test Contact",
	}
	require.NoError(t, db.Create(contact).Error)
	return contact
}

// ===== ENGINE TESTS =====

func TestEngine_Execute_SimpleLinearFlow(t *testing.T) {
	engine, _, db := setupTestEngine(t)

	// Create automation: START -> EMAIL -> EXIT
	nodes := []models.AutomationNode{
		{
			Base: models.Base{ID: "start"},
			Type: models.NodeTypeStart,
			Data: datatypes.JSON(`{}`),
		},
		{
			Base: models.Base{ID: "email"},
			Type: models.NodeTypeEmail,
			Data: datatypes.JSON(`{"templateId":"test"}`),
		},
		{
			Base: models.Base{ID: "exit"},
			Type: models.NodeTypeExit,
			Data: datatypes.JSON(`{}`),
		},
	}

	edges := []models.AutomationNodeEdge{
		{SourceID: "start", TargetID: "email"},
		{SourceID: "email", TargetID: "exit"},
	}

	automation := createTestAutomation(t, db, nodes, edges)
	contact := createTestContact(t, db, automation.TeamID)

	// Execute
	err := engine.Execute(context.Background(), automation.ID, contact.ID, nil, "")
	require.NoError(t, err)

	// Verify execution was created and completed
	var execution models.AutomationExecution
	err = db.Where("automation_id = ? AND contact_id = ?", automation.ID, contact.ID).First(&execution).Error
	require.NoError(t, err)
	assert.Equal(t, models.ExecutionStatusCompleted, execution.Status)
}

func TestEngine_Execute_WithWaitNode(t *testing.T) {
	engine, mockClient, db := setupTestEngine(t)

	// Create automation: START -> WAIT -> EXIT
	nodes := []models.AutomationNode{
		{
			Base: models.Base{ID: "start"},
			Type: models.NodeTypeStart,
			Data: datatypes.JSON(`{}`),
		},
		{
			Base: models.Base{ID: "wait"},
			Type: models.NodeTypeWait,
			Data: datatypes.JSON(`{"duration":"5m"}`),
		},
		{
			Base: models.Base{ID: "exit"},
			Type: models.NodeTypeExit,
			Data: datatypes.JSON(`{}`),
		},
	}

	edges := []models.AutomationNodeEdge{
		{SourceID: "start", TargetID: "wait"},
		{SourceID: "wait", TargetID: "exit"},
	}

	automation := createTestAutomation(t, db, nodes, edges)
	contact := createTestContact(t, db, automation.TeamID)

	// Execute
	err := engine.Execute(context.Background(), automation.ID, contact.ID, nil, "")
	require.NoError(t, err)

	// Verify execution is waiting
	var execution models.AutomationExecution
	err = db.Where("automation_id = ? AND contact_id = ?", automation.ID, contact.ID).First(&execution).Error
	require.NoError(t, err)
	assert.Equal(t, models.ExecutionStatusWaiting, execution.Status)
	assert.Equal(t, "wait", execution.CurrentNodeID)

	// Verify task was enqueued
	assert.Len(t, mockClient.enqueuedTasks, 1)
}

func TestEngine_Execute_ResumeFromWait(t *testing.T) {
	engine, _, db := setupTestEngine(t)

	// Create automation: START -> WAIT -> EXIT
	nodes := []models.AutomationNode{
		{
			Base: models.Base{ID: "start"},
			Type: models.NodeTypeStart,
			Data: datatypes.JSON(`{}`),
		},
		{
			Base: models.Base{ID: "wait"},
			Type: models.NodeTypeWait,
			Data: datatypes.JSON(`{"duration":"5m"}`),
		},
		{
			Base: models.Base{ID: "exit"},
			Type: models.NodeTypeExit,
			Data: datatypes.JSON(`{}`),
		},
	}

	edges := []models.AutomationNodeEdge{
		{SourceID: "start", TargetID: "wait"},
		{SourceID: "wait", TargetID: "exit"},
	}

	automation := createTestAutomation(t, db, nodes, edges)
	contact := createTestContact(t, db, automation.TeamID)

	// Execute initial flow (will pause at WAIT)
	err := engine.Execute(context.Background(), automation.ID, contact.ID, nil, "")
	require.NoError(t, err)

	// Resume at the continuation node selected by the WAIT processor.
	err = engine.Execute(context.Background(), automation.ID, contact.ID, nil, "exit")
	require.NoError(t, err)

	// Verify execution is now completed
	var execution models.AutomationExecution
	err = db.Where("automation_id = ? AND contact_id = ?", automation.ID, contact.ID).First(&execution).Error
	require.NoError(t, err)
	assert.Equal(t, models.ExecutionStatusCompleted, execution.Status)
}

func TestEngine_Execute_WithConditionalBranching(t *testing.T) {
	engine, _, db := setupTestEngine(t)

	// Create automation: START -> CONDITION -> (EMAIL_YES or EMAIL_NO) -> EXIT
	branchesData := map[string]interface{}{
		"true":  "email-yes",
		"false": "email-no",
	}
	branchesJSON, _ := json.Marshal(map[string]interface{}{"branches": branchesData})

	nodes := []models.AutomationNode{
		{
			Base: models.Base{ID: "start"},
			Type: models.NodeTypeStart,
			Data: datatypes.JSON(`{}`),
		},
		{
			Base: models.Base{ID: "condition"},
			Type: models.NodeTypeCondition,
			Data: datatypes.JSON(branchesJSON),
		},
		{
			Base: models.Base{ID: "email-yes"},
			Type: models.NodeTypeEmail,
			Data: datatypes.JSON(`{"templateId":"yes"}`),
		},
		{
			Base: models.Base{ID: "email-no"},
			Type: models.NodeTypeEmail,
			Data: datatypes.JSON(`{"templateId":"no"}`),
		},
		{
			Base: models.Base{ID: "exit"},
			Type: models.NodeTypeExit,
			Data: datatypes.JSON(`{}`),
		},
	}

	edges := []models.AutomationNodeEdge{
		{SourceID: "start", TargetID: "condition"},
		{SourceID: "condition", TargetID: "email-yes"},
		{SourceID: "condition", TargetID: "email-no"},
		{SourceID: "email-yes", TargetID: "exit"},
		{SourceID: "email-no", TargetID: "exit"},
	}

	automation := createTestAutomation(t, db, nodes, edges)
	contact := createTestContact(t, db, automation.TeamID)

	// Execute
	err := engine.Execute(context.Background(), automation.ID, contact.ID, nil, "")
	require.NoError(t, err)

	// Verify execution completed (took true branch)
	var execution models.AutomationExecution
	err = db.Where("automation_id = ? AND contact_id = ?", automation.ID, contact.ID).First(&execution).Error
	require.NoError(t, err)
	assert.Equal(t, models.ExecutionStatusCompleted, execution.Status)
}

func TestEngine_Execute_NonExistentAutomation(t *testing.T) {
	engine, _, db := setupTestEngine(t)

	fakeAutomationID := uuid.New().String()
	fakeContactID := uuid.New().String()

	err := engine.Execute(context.Background(), fakeAutomationID, fakeContactID, nil, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load automation")

	// Verify no execution was created
	var count int64
	db.Model(&models.AutomationExecution{}).Where("automation_id = ?", fakeAutomationID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestEngine_Execute_EventEmission(t *testing.T) {
	engine, _, db := setupTestEngine(t)

	// Track emitted events
	var emittedEvents []string
	var eventMu sync.Mutex
	events.On("automation.execution.started", func(data interface{}) {
		eventMu.Lock()
		defer eventMu.Unlock()
		emittedEvents = append(emittedEvents, "started")
	})
	events.On("automation.execution.completed", func(data interface{}) {
		eventMu.Lock()
		defer eventMu.Unlock()
		emittedEvents = append(emittedEvents, "completed")
	})
	events.On("automation.node.executed", func(data interface{}) {
		eventMu.Lock()
		defer eventMu.Unlock()
		emittedEvents = append(emittedEvents, "node_executed")
	})

	// Create simple automation
	nodes := []models.AutomationNode{
		{
			Base: models.Base{ID: "start"},
			Type: models.NodeTypeStart,
			Data: datatypes.JSON(`{}`),
		},
		{
			Base: models.Base{ID: "exit"},
			Type: models.NodeTypeExit,
			Data: datatypes.JSON(`{}`),
		},
	}

	edges := []models.AutomationNodeEdge{
		{SourceID: "start", TargetID: "exit"},
	}

	automation := createTestAutomation(t, db, nodes, edges)
	contact := createTestContact(t, db, automation.TeamID)

	// Execute
	err := engine.Execute(context.Background(), automation.ID, contact.ID, nil, "")
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		eventMu.Lock()
		defer eventMu.Unlock()
		return slices.Contains(emittedEvents, "started") && slices.Contains(emittedEvents, "completed") && slices.Contains(emittedEvents, "node_executed")
	}, time.Second, 5*time.Millisecond)

}

func TestEngine_Execute_ExecutionLogCreation(t *testing.T) {
	engine, _, db := setupTestEngine(t)

	// Create automation: START -> EMAIL -> EXIT
	nodes := []models.AutomationNode{
		{
			Base: models.Base{ID: "start"},
			Type: models.NodeTypeStart,
			Data: datatypes.JSON(`{}`),
		},
		{
			Base: models.Base{ID: "email"},
			Type: models.NodeTypeEmail,
			Data: datatypes.JSON(`{"templateId":"test"}`),
		},
		{
			Base: models.Base{ID: "exit"},
			Type: models.NodeTypeExit,
			Data: datatypes.JSON(`{}`),
		},
	}

	edges := []models.AutomationNodeEdge{
		{SourceID: "start", TargetID: "email"},
		{SourceID: "email", TargetID: "exit"},
	}

	automation := createTestAutomation(t, db, nodes, edges)
	contact := createTestContact(t, db, automation.TeamID)

	// Execute
	err := engine.Execute(context.Background(), automation.ID, contact.ID, nil, "")
	require.NoError(t, err)

	// Verify execution log
	var execution models.AutomationExecution
	err = db.Where("automation_id = ? AND contact_id = ?", automation.ID, contact.ID).First(&execution).Error
	require.NoError(t, err)

	var executionLog []LogEntry
	err = json.Unmarshal(execution.ExecutionLog, &executionLog)
	require.NoError(t, err)

	// Should have 3 log entries (start, email, exit)
	assert.Len(t, executionLog, 3)
	assert.Equal(t, "start", executionLog[0].NodeID)
	assert.Equal(t, "email", executionLog[1].NodeID)
	assert.Equal(t, "exit", executionLog[2].NodeID)
}

func TestEngine_RegisterProcessor(t *testing.T) {
	engine, _, _ := setupTestEngine(t)

	// Verify processors are registered
	processor, _ := engine.registry.Get(models.NodeTypeStart)
	assert.NotNil(t, processor)

	processor, _ = engine.registry.Get(models.NodeTypeEmail)
	assert.NotNil(t, processor)

	processor, _ = engine.registry.Get(models.NodeTypeExit)
	assert.NotNil(t, processor)
}

func (m *MockTaskClient) EnqueueAutomationTask(ctx context.Context, task tasks.AutomationExecuteTask, delay time.Duration) error {
	return m.EnqueueAutomationExecute(task.AutomationID, task.ContactID, task.TriggerData, task.CurrentNodeID, delay)
}

func testNextNodes(db *gorm.DB, node *models.AutomationNode) []string {
	var edges []models.AutomationNodeEdge
	db.Where("source_id = ? AND automation_id = ?", node.ID, node.AutomationID).Find(&edges)
	ids := []string{}
	for _, edge := range edges {
		ids = append(ids, edge.TargetID)
	}
	return ids
}

func TestQueueRetryRetainsExecutionIdentity(t *testing.T) {
	db := setupTestDB(t)
	engine := NewEngine(db, &MockTaskClient{})
	id, workflow, contact := uuid.NewString(), uuid.NewString(), uuid.NewString()
	data := map[string]interface{}{"_execution_id": id}
	first, created, err := engine.getOrCreateExecution(context.Background(), workflow, contact, data, "")
	require.NoError(t, err)
	require.True(t, created)
	require.NoError(t, db.Model(first).Update("status", models.ExecutionStatusFailed).Error)
	retried, created, err := engine.getOrCreateExecution(context.Background(), workflow, contact, data, "")
	require.NoError(t, err)
	require.False(t, created)
	require.Equal(t, first.ID, retried.ID)
	_, _, err = engine.getOrCreateExecution(context.Background(), workflow, uuid.NewString(), data, "")
	require.Error(t, err)
}
