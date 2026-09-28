package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"kori/internal/models"
	"kori/internal/tasks"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Mock task client
type MockTaskClient struct{}

func (m *MockTaskClient) EnqueueAutomationExecute(automationID, contactID string, triggerData map[string]interface{}, currentNodeID string, processIn time.Duration) error {
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
		&models.Team{},
		&models.Automation{},
		&models.AutomationNode{},
		&models.AutomationNodeEdge{},
		&models.AutomationExecution{},
		&models.Contact{},
		&models.Tag{},
	)
	require.NoError(t, err)

	return db
}

func setupTestHandler(t *testing.T) (*AutomationHandler, *gorm.DB, *echo.Echo) {
	db := setupTestDB(t)
	mockClient := &MockTaskClient{}

	handler := NewAutomationHandler(db, mockClient)
	e := echo.New()

	return handler, db, e
}

func createTestTeam(t *testing.T, db *gorm.DB) *models.Team {
	team := &models.Team{
		Base: models.Base{ID: uuid.New().String()},
		Name: "Test Team",
	}
	require.NoError(t, db.Create(team).Error)
	return team
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

func TestAutomationHandler_CreateAutomation(t *testing.T) {
	handler, db, e := setupTestHandler(t)
	team := createTestTeam(t, db)

	reqBody := map[string]interface{}{
		"name":        "Test Automation",
		"description": "Test Description",
		"nodes": []map[string]interface{}{
			{
				"id":   "00000000-0000-4000-8000-000000000001",
				"type": "START",
				"data": map[string]interface{}{},
			},
			{
				"id":   "00000000-0000-4000-8000-000000000002",
				"type": "EXIT",
				"data": map[string]interface{}{},
			},
		},
		"edges": []map[string]interface{}{
			{
				"sourceId": "00000000-0000-4000-8000-000000000001",
				"targetId": "00000000-0000-4000-8000-000000000002",
			},
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/automations", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("teamID", team.ID)

	err := handler.CreateAutomation(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response
	assert.Equal(t, "Test Automation", data["name"])
	assert.NotEmpty(t, data["id"])
}

func TestAutomationHandler_GetAutomation(t *testing.T) {
	handler, db, e := setupTestHandler(t)
	team := createTestTeam(t, db)

	// Create automation
	automation := &models.Automation{
		Base:        models.Base{ID: uuid.New().String()},
		TeamID:      team.ID,
		Name:        "Test Automation",
		Description: "Test",
		Nodes: []models.AutomationNode{
			{
				Base: models.Base{ID: "00000000-0000-4000-8000-000000000001"},
				Type: models.NodeTypeStart,
				Data: datatypes.JSON(`{}`),
			},
		},
	}
	require.NoError(t, db.Create(automation).Error)

	// Get automation
	req := httptest.NewRequest(http.MethodGet, "/api/v1/automations/"+automation.ID, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/automations/:id")
	c.SetParamNames("id")
	c.SetParamValues(automation.ID)
	c.Set("teamID", team.ID)

	err := handler.GetAutomation(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response
	assert.Equal(t, automation.ID, data["id"])
	assert.Equal(t, "Test Automation", data["name"])
}

func TestAutomationHandler_ListAutomations(t *testing.T) {
	handler, db, e := setupTestHandler(t)
	team := createTestTeam(t, db)

	// Create multiple automations
	for i := 0; i < 3; i++ {
		automation := &models.Automation{
			Base:        models.Base{ID: uuid.New().String()},
			TeamID:      team.ID,
			Name:        "Test Automation " + string(rune(i)),
			Description: "Test",
		}
		require.NoError(t, db.Create(automation).Error)
	}

	// List automations
	req := httptest.NewRequest(http.MethodGet, "/api/v1/automations", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("teamID", team.ID)

	err := handler.ListAutomations(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var data []interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &data)
	require.NoError(t, err)
	assert.Len(t, data, 3)
}

func TestAutomationHandler_UpdateAutomation(t *testing.T) {
	handler, db, e := setupTestHandler(t)
	team := createTestTeam(t, db)

	// Create automation
	automation := &models.Automation{
		Base:        models.Base{ID: uuid.New().String()},
		TeamID:      team.ID,
		Name:        "Original Name",
		Description: "Original Description",
	}
	require.NoError(t, db.Create(automation).Error)

	// Update automation
	updateBody := map[string]interface{}{
		"name":        "Updated Name",
		"description": "Updated Description",
		"nodes":       []map[string]interface{}{{"id": "00000000-0000-4000-8000-000000000001", "type": "START", "data": map[string]interface{}{}}},
	}

	body, _ := json.Marshal(updateBody)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/automations/"+automation.ID, bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/automations/:id")
	c.SetParamNames("id")
	c.SetParamValues(automation.ID)
	c.Set("teamID", team.ID)

	err := handler.UpdateAutomation(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify update
	var updated models.Automation
	err = db.First(&updated, "id = ?", automation.ID).Error
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "Updated Description", updated.Description)
}

func TestAutomationHandler_DeleteAutomation(t *testing.T) {
	handler, db, e := setupTestHandler(t)
	team := createTestTeam(t, db)

	// Create automation
	automation := &models.Automation{
		Base:        models.Base{ID: uuid.New().String()},
		TeamID:      team.ID,
		Name:        "Test Automation",
		Description: "Test",
	}
	require.NoError(t, db.Create(automation).Error)

	// Delete automation
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/automations/"+automation.ID, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/automations/:id")
	c.SetParamNames("id")
	c.SetParamValues(automation.ID)
	c.Set("teamID", team.ID)

	err := handler.DeleteAutomation(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify deletion
	var count int64
	db.Model(&models.Automation{}).Where("id = ? AND is_deleted = ?", automation.ID, false).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestAutomationHandler_ActivateAutomation(t *testing.T) {
	handler, db, e := setupTestHandler(t)
	team := createTestTeam(t, db)

	// Create valid automation
	automation := &models.Automation{
		Base:        models.Base{ID: uuid.New().String()},
		TeamID:      team.ID,
		Name:        "Test Automation",
		Description: "Test",
		IsActive:    false,
		Nodes: []models.AutomationNode{
			{
				Base: models.Base{ID: "00000000-0000-4000-8000-000000000001"},
				Type: models.NodeTypeStart,
				Data: datatypes.JSON(`{}`),
			},
			{
				Base: models.Base{ID: "00000000-0000-4000-8000-000000000002"},
				Type: models.NodeTypeExit,
				Data: datatypes.JSON(`{}`),
			},
		},
		Edges: []models.AutomationNodeEdge{
			{SourceID: "00000000-0000-4000-8000-000000000001", TargetID: "00000000-0000-4000-8000-000000000002"},
		},
	}
	require.NoError(t, db.Create(automation).Error)

	// Activate automation
	req := httptest.NewRequest(http.MethodPost, "/api/v1/automations/"+automation.ID+"/activate", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/automations/:id/activate")
	c.SetParamNames("id")
	c.SetParamValues(automation.ID)
	c.Set("teamID", team.ID)

	err := handler.ActivateAutomation(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify activation
	var updated models.Automation
	err = db.First(&updated, "id = ?", automation.ID).Error
	require.NoError(t, err)
	assert.True(t, updated.IsActive)
}

func TestAutomationHandler_DeactivateAutomation(t *testing.T) {
	handler, db, e := setupTestHandler(t)
	team := createTestTeam(t, db)

	// Create active automation
	automation := &models.Automation{
		Base:        models.Base{ID: uuid.New().String()},
		TeamID:      team.ID,
		Name:        "Test Automation",
		Description: "Test",
		IsActive:    true,
	}
	require.NoError(t, db.Create(automation).Error)

	// Deactivate automation
	req := httptest.NewRequest(http.MethodPost, "/api/v1/automations/"+automation.ID+"/deactivate", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/automations/:id/deactivate")
	c.SetParamNames("id")
	c.SetParamValues(automation.ID)
	c.Set("teamID", team.ID)

	err := handler.DeactivateAutomation(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify deactivation
	var updated models.Automation
	err = db.First(&updated, "id = ?", automation.ID).Error
	require.NoError(t, err)
	assert.False(t, updated.IsActive)
}

func TestAutomationHandler_TriggerAutomation(t *testing.T) {
	handler, db, e := setupTestHandler(t)
	team := createTestTeam(t, db)
	contact := createTestContact(t, db, team.ID)

	// Create automation
	automation := &models.Automation{
		Base:        models.Base{ID: uuid.New().String()},
		TeamID:      team.ID,
		Name:        "Test Automation",
		Description: "Test",
		IsActive:    true,
		Nodes: []models.AutomationNode{
			{
				Base: models.Base{ID: "00000000-0000-4000-8000-000000000001"},
				Type: models.NodeTypeStart,
				Data: datatypes.JSON(`{}`),
			},
		},
	}
	require.NoError(t, db.Create(automation).Error)

	// Trigger automation
	triggerBody := map[string]interface{}{
		"contactIds": []string{contact.ID},
		"triggerData": map[string]interface{}{
			"source": "manual",
		},
	}

	body, _ := json.Marshal(triggerBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/automations/"+automation.ID+"/trigger", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/automations/:id/trigger")
	c.SetParamNames("id")
	c.SetParamValues(automation.ID)
	c.Set("teamID", team.ID)

	err := handler.TriggerAutomation(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Automation triggered successfully", response["message"])
}

func TestAutomationHandler_GetExecutions(t *testing.T) {
	handler, db, e := setupTestHandler(t)
	team := createTestTeam(t, db)
	contact := createTestContact(t, db, team.ID)

	// Create automation
	automation := &models.Automation{
		Base:        models.Base{ID: uuid.New().String()},
		TeamID:      team.ID,
		Name:        "Test Automation",
		Description: "Test",
	}
	require.NoError(t, db.Create(automation).Error)

	// Create executions
	for i := 0; i < 3; i++ {
		execution := &models.AutomationExecution{
			Base:         models.Base{ID: uuid.New().String()},
			AutomationID: automation.ID,
			ContactID:    contact.ID,
			Status:       models.ExecutionStatusCompleted,
			ExecutionLog: datatypes.JSON(`[]`),
			Variables:    datatypes.JSON(`{}`),
		}
		require.NoError(t, db.Create(execution).Error)
	}

	// Get executions
	req := httptest.NewRequest(http.MethodGet, "/api/v1/automations/"+automation.ID+"/executions", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/automations/:id/executions")
	c.SetParamNames("id")
	c.SetParamValues(automation.ID)
	c.Set("teamID", team.ID)

	err := handler.GetExecutions(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var data []interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &data)
	require.NoError(t, err)
	assert.Len(t, data, 3)
}

func (m *MockTaskClient) EnqueueAutomationTask(ctx context.Context, task tasks.AutomationExecuteTask, delay time.Duration) error {
	return nil
}
