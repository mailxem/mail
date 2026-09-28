package services

import (
	"context"
	"kori/internal/models"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	// Set test mode to skip seeders
	os.Setenv("TEST_MODE", "true")

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate models
	err = db.AutoMigrate(
		&models.Automation{},
		&models.AutomationNode{},
		&models.AutomationNodeEdge{},
		&models.AutomationExecution{},
	)
	require.NoError(t, err)

	return db
}

func TestAutomationService_Create(t *testing.T) {
	db := setupTestDB(t)
	service := NewAutomationService(db)
	ctx := context.Background()

	teamID := uuid.New().String()

	automation := &models.Automation{
		Name:        "Test Automation",
		Description: "Test Description",
		TeamID:      teamID,
	}

	err := service.Create(ctx, automation)
	require.NoError(t, err)
	assert.NotEmpty(t, automation.ID)
	assert.Equal(t, "Test Automation", automation.Name)
	assert.Equal(t, teamID, automation.TeamID)
}

func TestAutomationService_ValidateGraph_NoCycles(t *testing.T) {
	db := setupTestDB(t)
	service := NewAutomationService(db)

	// Create a valid linear graph: START -> EMAIL -> EXIT
	automation := &models.Automation{
		Nodes: []models.AutomationNode{
			{Base: models.Base{ID: "start"}, Type: models.NodeTypeStart},
			{Base: models.Base{ID: "email"}, Type: models.NodeTypeEmail},
			{Base: models.Base{ID: "exit"}, Type: models.NodeTypeExit},
		},
		Edges: []models.AutomationNodeEdge{
			{SourceID: "start", TargetID: "email"},
			{SourceID: "email", TargetID: "exit"},
		},
	}

	err := service.ValidateGraph(automation)
	assert.NoError(t, err)
}

func TestAutomationService_ValidateGraph_DetectsCycle(t *testing.T) {
	db := setupTestDB(t)
	service := NewAutomationService(db)

	// Create a graph with a cycle: A -> B -> C -> A
	automation := &models.Automation{
		Nodes: []models.AutomationNode{
			{Base: models.Base{ID: "a"}, Type: models.NodeTypeStart},
			{Base: models.Base{ID: "b"}, Type: models.NodeTypeEmail},
			{Base: models.Base{ID: "c"}, Type: models.NodeTypeWait},
		},
		Edges: []models.AutomationNodeEdge{
			{SourceID: "a", TargetID: "b"},
			{SourceID: "b", TargetID: "c"},
			{SourceID: "c", TargetID: "a"}, // Creates cycle
		},
	}

	err := service.ValidateGraph(automation)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "automation graph contains cycles")
}

func TestAutomationService_ValidateGraph_OrphanedNodes(t *testing.T) {
	db := setupTestDB(t)
	service := NewAutomationService(db)

	// Create a graph with orphaned node
	automation := &models.Automation{
		Nodes: []models.AutomationNode{
			{Base: models.Base{ID: "start"}, Type: models.NodeTypeStart},
			{Base: models.Base{ID: "email"}, Type: models.NodeTypeEmail},
			{Base: models.Base{ID: "orphan"}, Type: models.NodeTypeWait}, // Not connected
		},
		Edges: []models.AutomationNodeEdge{
			{SourceID: "start", TargetID: "email"},
		},
	}

	err := service.ValidateGraph(automation)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "all steps must be reachable")
}

func TestAutomationService_ValidateGraph_NoStartNode(t *testing.T) {
	db := setupTestDB(t)
	service := NewAutomationService(db)

	// Create a graph without START node
	automation := &models.Automation{
		Nodes: []models.AutomationNode{
			{Base: models.Base{ID: "email"}, Type: models.NodeTypeEmail},
			{Base: models.Base{ID: "exit"}, Type: models.NodeTypeExit},
		},
		Edges: []models.AutomationNodeEdge{
			{SourceID: "email", TargetID: "exit"},
		},
	}

	err := service.ValidateGraph(automation)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must have exactly one START node")
}

func TestAutomationService_ValidateGraph_EmptyGraph(t *testing.T) {
	db := setupTestDB(t)
	service := NewAutomationService(db)

	// Create an empty graph
	automation := &models.Automation{
		Nodes: []models.AutomationNode{},
		Edges: []models.AutomationNodeEdge{},
	}

	err := service.ValidateGraph(automation)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must have at least one node")
}

func TestAutomationService_ValidateGraph_ComplexValid(t *testing.T) {
	db := setupTestDB(t)
	service := NewAutomationService(db)

	// Create a complex valid graph with branching
	automation := &models.Automation{
		Nodes: []models.AutomationNode{
			{Base: models.Base{ID: "start"}, Type: models.NodeTypeStart},
			{Base: models.Base{ID: "condition"}, Type: models.NodeTypeCondition},
			{Base: models.Base{ID: "email1"}, Type: models.NodeTypeEmail},
			{Base: models.Base{ID: "email2"}, Type: models.NodeTypeEmail},
			{Base: models.Base{ID: "wait"}, Type: models.NodeTypeWait},
			{Base: models.Base{ID: "exit"}, Type: models.NodeTypeExit},
		},
		Edges: []models.AutomationNodeEdge{
			{SourceID: "start", TargetID: "condition"},
			{SourceID: "condition", TargetID: "email1"},
			{SourceID: "condition", TargetID: "email2"},
			{SourceID: "email1", TargetID: "wait"},
			{SourceID: "email2", TargetID: "wait"},
			{SourceID: "wait", TargetID: "exit"},
		},
	}

	err := service.ValidateGraph(automation)
	assert.NoError(t, err)
}

func TestAutomationService_Activate(t *testing.T) {
	db := setupTestDB(t)
	service := NewAutomationService(db)
	ctx := context.Background()

	// Create valid automation
	automation := &models.Automation{
		TeamID:      uuid.New().String(),
		Name:        "Test",
		Description: "Test",
		IsActive:    false,
		Nodes: []models.AutomationNode{
			{Base: models.Base{ID: "start"}, Type: models.NodeTypeStart},
			{Base: models.Base{ID: "exit"}, Type: models.NodeTypeExit},
		},
		Edges: []models.AutomationNodeEdge{
			{SourceID: "start", TargetID: "exit"},
		},
	}

	err := service.Create(ctx, automation)
	require.NoError(t, err)

	// Activate it
	err = service.Activate(ctx, automation.ID)
	assert.NoError(t, err)

	// Verify it's active
	fetched, err := service.Get(ctx, automation.ID)
	require.NoError(t, err)
	assert.True(t, fetched.IsActive)
}

func TestAutomationService_Activate_InvalidGraph(t *testing.T) {
	db := setupTestDB(t)
	service := NewAutomationService(db)
	ctx := context.Background()

	// Create automation with cycle
	automation := &models.Automation{
		TeamID:      uuid.New().String(),
		Name:        "Test",
		Description: "Test",
		IsActive:    false,
		Nodes: []models.AutomationNode{
			{Base: models.Base{ID: "a"}, Type: models.NodeTypeStart},
			{Base: models.Base{ID: "b"}, Type: models.NodeTypeEmail},
		},
		Edges: []models.AutomationNodeEdge{
			{SourceID: "a", TargetID: "b"},
			{SourceID: "b", TargetID: "a"}, // Cycle
		},
	}

	err := service.Create(ctx, automation)
	require.NoError(t, err)

	// Try to activate - should fail
	err = service.Activate(ctx, automation.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "automation graph contains cycles")
}

func TestAutomationService_Deactivate(t *testing.T) {
	db := setupTestDB(t)
	service := NewAutomationService(db)
	ctx := context.Background()

	// Create and activate automation
	automation := &models.Automation{
		TeamID:      uuid.New().String(),
		Name:        "Test",
		Description: "Test",
		IsActive:    true,
		Nodes: []models.AutomationNode{
			{Base: models.Base{ID: "start"}, Type: models.NodeTypeStart},
			{Base: models.Base{ID: "exit"}, Type: models.NodeTypeExit},
		},
		Edges: []models.AutomationNodeEdge{
			{SourceID: "start", TargetID: "exit"},
		},
	}

	err := service.Create(ctx, automation)
	require.NoError(t, err)

	// Deactivate it
	err = service.Deactivate(ctx, automation.ID)
	assert.NoError(t, err)

	// Verify it's inactive
	fetched, err := service.Get(ctx, automation.ID)
	require.NoError(t, err)
	assert.False(t, fetched.IsActive)
}

func TestAutomationService_GetWithNodesAndEdges(t *testing.T) {
	db := setupTestDB(t)
	service := NewAutomationService(db)
	ctx := context.Background()

	// Create automation with nodes and edges
	automation := &models.Automation{
		TeamID:      uuid.New().String(),
		Name:        "Test",
		Description: "Test",
		Nodes: []models.AutomationNode{
			{Base: models.Base{ID: "start"}, Type: models.NodeTypeStart},
			{Base: models.Base{ID: "email"}, Type: models.NodeTypeEmail},
			{Base: models.Base{ID: "exit"}, Type: models.NodeTypeExit},
		},
		Edges: []models.AutomationNodeEdge{
			{SourceID: "start", TargetID: "email"},
			{SourceID: "email", TargetID: "exit"},
		},
	}

	err := service.Create(ctx, automation)
	require.NoError(t, err)

	// Fetch with relations
	fetched, err := service.GetWithNodesAndEdges(ctx, automation.ID)
	require.NoError(t, err)
	assert.Len(t, fetched.Nodes, 3)
	assert.Len(t, fetched.Edges, 2)
}
