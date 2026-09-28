package processors

import (
	"encoding/json"
	"kori/internal/automation"
	"kori/internal/models"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupDB(t *testing.T) *gorm.DB {
	// Set test mode to skip seeders
	os.Setenv("TEST_MODE", "true")

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// AutoMigrate all necessary tables
	err = db.AutoMigrate(
		&models.BrandingSettings{},
		&models.TeamSettings{},
		&models.Team{},
		&models.Contact{},
		&models.Tag{},
		&models.Automation{},
		&models.AutomationNode{},
		&models.AutomationNodeEdge{},
		&models.AutomationExecution{},
	)
	require.NoError(t, err)

	return db
}

func createContext(t *testing.T, db *gorm.DB) *automation.ExecutionContext {
	teamID := uuid.New().String()
	contactID := uuid.New().String()
	automationID := uuid.New().String()

	team := &models.Team{
		Base: models.Base{ID: teamID},
		Name: "Test Team",
	}
	require.NoError(t, db.Create(team).Error)

	contact := &models.Contact{
		Base:      models.Base{ID: contactID},
		TeamID:    teamID,
		Email:     "test@example.com",
		FirstName: "Test",
	}
	require.NoError(t, db.Create(contact).Error)

	// Create automation
	autom := &models.Automation{
		Base:   models.Base{ID: automationID},
		TeamID: teamID,
		Name:   "Test Automation",
	}
	require.NoError(t, db.Create(autom).Error)

	return &automation.ExecutionContext{
		AutomationID: automationID,
		ContactID:    contactID,
		TeamID:       teamID,
		Contact:      contact,
		Variables:    make(map[string]interface{}),
		ExecutionLog: []automation.LogEntry{},
	}
}

func createContextWithEdges(t *testing.T, db *gorm.DB, sourceID, targetID string) *automation.ExecutionContext {
	ctx := createContext(t, db)

	// Create nodes
	startNode := &models.AutomationNode{
		Base:         models.Base{ID: sourceID},
		AutomationID: ctx.AutomationID,
		Type:         models.NodeTypeStart,
		Data:         datatypes.JSON(`{}`),
	}
	require.NoError(t, db.Create(startNode).Error)

	nextNode := &models.AutomationNode{
		Base:         models.Base{ID: targetID},
		AutomationID: ctx.AutomationID,
		Type:         models.NodeTypeExit,
		Data:         datatypes.JSON(`{}`),
	}
	require.NoError(t, db.Create(nextNode).Error)

	// Create edge
	edge := &models.AutomationNodeEdge{
		AutomationID: ctx.AutomationID,
		SourceID:     sourceID,
		TargetID:     targetID,
	}
	require.NoError(t, db.Create(edge).Error)

	return ctx
}

// Test StartProcessor
func TestStart_Process(t *testing.T) {
	db := setupDB(t)
	processor := NewStartProcessor(db)
	ctx := createContextWithEdges(t, db, "start", "next")

	node := &models.AutomationNode{
		Base: models.Base{ID: "start"},
		Type: models.NodeTypeStart,
		Data: datatypes.JSON(`{}`),
	}

	result, err := processor.Process(ctx, node)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Complete)
	assert.Len(t, result.NextNodeIDs, 1)
	assert.Equal(t, "next", result.NextNodeIDs[0])
}

// Test WaitProcessor
func TestWait_Process(t *testing.T) {
	db := setupDB(t)
	processor := NewWaitProcessor(db)
	ctx := createContextWithEdges(t, db, "wait", "next")

	data := map[string]interface{}{"duration": "5m"}
	dataJSON, _ := json.Marshal(data)

	node := &models.AutomationNode{
		Base: models.Base{ID: "wait"},
		Type: models.NodeTypeWait,
		Data: datatypes.JSON(dataJSON),
	}

	result, err := processor.Process(ctx, node)
	require.NoError(t, err)
	assert.NotNil(t, result.Wait)
	assert.Equal(t, 5*time.Minute, *result.Wait)
	assert.Len(t, result.NextNodeIDs, 1)
}

// Test ExitProcessor
func TestExit_Process(t *testing.T) {
	db := setupDB(t)
	processor := NewExitProcessor(db)
	ctx := createContext(t, db)

	node := &models.AutomationNode{
		Base: models.Base{ID: "exit"},
		Type: models.NodeTypeExit,
		Data: datatypes.JSON(`{}`),
	}

	result, err := processor.Process(ctx, node)
	require.NoError(t, err)
	assert.True(t, result.Complete)
}

// Test ConditionProcessor - Simple Equals
func TestCondition_Equals(t *testing.T) {
	db := setupDB(t)
	processor := NewConditionProcessor(db)

	ctx := createContext(t, db)
	ctx.Variables["score"] = "100"

	data := map[string]interface{}{
		"conditions": []map[string]interface{}{
			{"variable": "score", "operator": "==", "value": "100"},
		},
		"operator": "AND",
	}
	dataJSON, _ := json.Marshal(data)

	node := &models.AutomationNode{
		Base: models.Base{ID: "cond"},
		Type: models.NodeTypeCondition,
		Data: datatypes.JSON(dataJSON),
	}

	require.NoError(t, db.Create(&[]models.AutomationNodeEdge{{AutomationID: ctx.AutomationID, SourceID: node.ID, TargetID: "yes-node", Label: "true"}, {AutomationID: ctx.AutomationID, SourceID: node.ID, TargetID: "no-node", Label: "false"}}).Error)

	result, err := processor.Process(ctx, node)
	require.NoError(t, err)
	assert.Len(t, result.NextNodeIDs, 1)
	assert.Equal(t, "yes-node", result.NextNodeIDs[0])
}

// Test ConditionProcessor - Greater Than
func TestCondition_GreaterThan(t *testing.T) {
	db := setupDB(t)
	processor := NewConditionProcessor(db)

	ctx := createContext(t, db)
	ctx.Variables["count"] = "75"

	data := map[string]interface{}{
		"conditions": []map[string]interface{}{
			{"variable": "count", "operator": ">", "value": "50"},
		},
		"operator": "AND",
	}
	dataJSON, _ := json.Marshal(data)

	node := &models.AutomationNode{
		Base: models.Base{ID: "cond"},
		Type: models.NodeTypeCondition,
		Data: datatypes.JSON(dataJSON),
	}

	require.NoError(t, db.Create(&[]models.AutomationNodeEdge{{AutomationID: ctx.AutomationID, SourceID: node.ID, TargetID: "high", Label: "true"}, {AutomationID: ctx.AutomationID, SourceID: node.ID, TargetID: "low", Label: "false"}}).Error)

	result, err := processor.Process(ctx, node)
	require.NoError(t, err)
	assert.Equal(t, "high", result.NextNodeIDs[0])
}

// Test ConditionProcessor - Contains
func TestCondition_Contains(t *testing.T) {
	db := setupDB(t)
	processor := NewConditionProcessor(db)

	ctx := createContext(t, db)
	ctx.Variables["tags"] = "vip,premium,active"

	data := map[string]interface{}{
		"conditions": []map[string]interface{}{
			{"variable": "tags", "operator": "contains", "value": "vip"},
		},
		"operator": "AND",
	}
	dataJSON, _ := json.Marshal(data)

	node := &models.AutomationNode{
		Base: models.Base{ID: "cond"},
		Type: models.NodeTypeCondition,
		Data: datatypes.JSON(dataJSON),
	}

	require.NoError(t, db.Create(&[]models.AutomationNodeEdge{{AutomationID: ctx.AutomationID, SourceID: node.ID, TargetID: "vip-flow", Label: "true"}, {AutomationID: ctx.AutomationID, SourceID: node.ID, TargetID: "regular-flow", Label: "false"}}).Error)

	result, err := processor.Process(ctx, node)
	require.NoError(t, err)
	assert.Equal(t, "vip-flow", result.NextNodeIDs[0])
}

// Test UpdateSubscriberProcessor
func TestUpdateSubscriber_Process(t *testing.T) {
	db := setupDB(t)
	processor := NewUpdateSubscriberProcessor(db)
	ctx := createContext(t, db)

	data := map[string]interface{}{
		"fields": map[string]interface{}{
			"first_name": "Updated",
			"last_name":  "Name",
		},
	}
	dataJSON, _ := json.Marshal(data)

	node := &models.AutomationNode{
		Base: models.Base{ID: "update"},
		Type: models.NodeTypeUpdateSubscriber,
		Data: datatypes.JSON(dataJSON),
	}

	result, err := processor.Process(ctx, node)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify update
	var contact models.Contact
	err = db.First(&contact, "id = ?", ctx.ContactID).Error
	require.NoError(t, err)
	assert.Equal(t, "Updated", contact.FirstName)
	assert.Equal(t, "Name", contact.LastName)
}

// Test WebhookProcessor Validation
func TestWebhook_Validate(t *testing.T) {
	db := setupDB(t)
	processor := NewWebhookProcessor(db)

	tests := []struct {
		name   string
		url    string
		method string
		valid  bool
	}{
		{"Valid GET", "https://example.com/hook", "GET", true},
		{"Valid POST", "https://example.com/hook", "POST", true},
		{"Invalid method", "https://example.com/hook", "INVALID", false},
		{"Missing URL", "", "POST", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := map[string]interface{}{
				"url":    tt.url,
				"method": tt.method,
			}
			dataJSON, _ := json.Marshal(data)

			node := &models.AutomationNode{
				Base: models.Base{ID: "webhook"},
				Type: models.NodeTypeWebhook,
				Data: datatypes.JSON(dataJSON),
			}

			err := processor.Validate(node)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
