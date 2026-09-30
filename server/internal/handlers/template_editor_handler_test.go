package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"kori/internal/models"
)

func TestTemplateEditorIsolationAndVersion(t *testing.T) {
	var dialect gorm.Dialector = sqlite.Open(":memory:")
	if dsn := os.Getenv("POSTHOOT_TEST_DATABASE_URL"); dsn != "" {
		admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		require.NoError(t, err)
		schema := "template_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
		require.NoError(t, admin.Exec("CREATE SCHEMA "+schema).Error)
		t.Cleanup(func() { admin.Exec("DROP SCHEMA " + schema + " CASCADE"); raw, _ := admin.DB(); raw.Close() })
		dialect = postgres.Open(dsn + " search_path=" + schema)
	}
	db, err := gorm.Open(dialect, &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	rawDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { rawDB.Close() })
	require.NoError(t, db.AutoMigrate(&models.Template{}, &models.EmailCategory{}))
	team, other := uuid.NewString(), uuid.NewString()
	category := models.EmailCategory{TeamID: team, Name: "Transactional"}
	require.NoError(t, db.Create(&category).Error)
	foreign := models.EmailCategory{TeamID: other, Name: "Foreign"}
	require.NoError(t, db.Create(&foreign).Error)
	row := models.Template{TeamID: team, CategoryID: category.ID, Name: "Welcome", Subject: "Hello", HTMLBody: "old", HtmlFileID: uuid.NewString()}
	require.NoError(t, db.Create(&row).Error)
	require.NoError(t, db.First(&row, "id = ?", row.ID).Error)
	h := NewMarketingHandler(db, "secret", "https://xem.email")
	input := map[string]any{"name": "New welcome", "subject": "A new hello", "categoryId": category.ID, "htmlBody": "<p>New</p>", "designJson": base64.StdEncoding.EncodeToString([]byte(`{"schemaVersion":18,"body":{"rows":[]}}`)), "expectedUpdatedAt": row.UpdatedAt}
	var response models.Template
	call := func(workspace, id string) int {
		raw, _ := json.Marshal(input)
		req := httptest.NewRequest("PUT", "/", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		e := echo.New()
		c := e.NewContext(req, rec)
		c.Set("teamID", workspace)
		c.SetParamNames("id")
		c.SetParamValues(id)
		if err := h.SaveTemplate(c); err != nil {
			e.HTTPErrorHandler(err, c)
		}
		if rec.Code == 200 || rec.Code == 201 {
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
		}
		return rec.Code
	}
	require.Equal(t, 404, call(other, row.ID))
	input["categoryId"] = foreign.ID
	require.Equal(t, 400, call(team, row.ID))
	input["categoryId"] = category.ID
	require.Equal(t, 200, call(team, row.ID))
	var saved models.Template
	require.NoError(t, db.First(&saved, "id = ?", row.ID).Error)
	require.Equal(t, "<p>New</p>", saved.HTMLBody)
	require.Empty(t, saved.HtmlFileID)
	require.Equal(t, "New welcome", saved.Name)
	require.True(t, response.UpdatedAt.Equal(saved.UpdatedAt), "The returned version must exactly match PostgreSQL's stored timestamp")
	require.Zero(t, response.UpdatedAt.Nanosecond()%1000)
	creationID := uuid.NewString()
	input["creationId"] = creationID
	require.Equal(t, 201, call(team, ""))
	require.Equal(t, 201, call(team, ""))
	var created int64
	require.NoError(t, db.Model(&models.Template{}).Where("id = ?", creationID).Count(&created).Error)
	require.EqualValues(t, 1, created)
	input["subject"] = "Changed on retry"
	require.Equal(t, 409, call(team, ""))
	input["subject"] = "A new hello"
	require.Equal(t, 409, call(team, row.ID))
	input["expectedUpdatedAt"] = time.Now().Add(time.Hour)
	require.Equal(t, 409, call(team, row.ID))
	input["expectedUpdatedAt"] = saved.UpdatedAt
	require.Equal(t, 200, call(team, row.ID))
	input["expectedUpdatedAt"] = response.UpdatedAt
	require.Equal(t, 200, call(team, row.ID), "Consecutive edits must work using the returned version")
	input["designJson"] = "broken"
	require.Equal(t, 400, call(team, row.ID))
	require.NoError(t, db.Model(&saved).Update("is_deleted", true).Error)
	require.Equal(t, 404, func() int {
		input["designJson"] = base64.StdEncoding.EncodeToString([]byte(`{"schemaVersion":18,"body":{"rows":[]}}`))
		return call(team, row.ID)
	}())
}
