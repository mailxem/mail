package marketing_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"kori/internal/handlers"
	"kori/internal/models"
)

func TestMCPContactBatchValidationAndSuppression(t *testing.T) {
	db := testDB(t)
	team, other, listID, foreign := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	seed(t, db, &models.MailingList{Base: models.Base{ID: listID}, TeamID: team, Name: "Our list"})
	seed(t, db, &models.MailingList{Base: models.Base{ID: foreign}, TeamID: other, Name: "Other list"})
	seed(t, db, &models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: team, ListID: listID, Email: "suppressed@example.com", Status: models.SubscriberStatusUnsubscribed})
	h := handlers.NewMarketingHandler(db, "", "")
	e := echo.New()
	e.POST("/batch", h.ImportContactBatch, func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error { c.Set("teamID", team); return next(c) }
	})
	call := func(body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", "/batch", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		e.ServeHTTP(w, r)
		return w
	}
	payload := fmt.Sprintf(`{"listId":%q,"contacts":[{"email":"NEW@example.com"},{"email":"new@example.com"},{"email":"suppressed@example.com"}]}`, listID)
	w := call(payload)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), `"wouldCreate":1`)
	require.Contains(t, w.Body.String(), `"skipped":2`)
	var count int64
	require.NoError(t, db.Model(&models.Contact{}).Count(&count).Error)
	require.EqualValues(t, 1, count, "preview must not write")
	for _, badBody := range []string{
		fmt.Sprintf(`{"listId":%q,"contacts":[{"email":"a@example.com"}],"dryRun":false}`, foreign),
		fmt.Sprintf(`{"listId":%q,"contacts":[{"email":"valid@example.com"},{"email":"bad"}],"dryRun":false}`, listID),
		fmt.Sprintf(`{"listId":%q,"contacts":[{"email":"a@example.com","teamId":%q}],"dryRun":false}`, listID, other),
		payload + ` {}`,
	} {
		w = call(badBody)
		require.GreaterOrEqual(t, w.Code, 400, w.Body.String())
	}
	require.NoError(t, db.Model(&models.Contact{}).Count(&count).Error)
	require.EqualValues(t, 1, count)
	payload = strings.TrimSuffix(payload, "}") + `,"dryRun":false}`
	w = call(payload)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), `"created":1`)
	w = call(payload)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), `"created":0`)
	var contact models.Contact
	require.NoError(t, db.Where("email = ?", "suppressed@example.com").First(&contact).Error)
	require.Equal(t, models.SubscriberStatusUnsubscribed, contact.Status)
	require.NoError(t, db.Model(&models.Contact{}).Count(&count).Error)
	require.EqualValues(t, 2, count)
}

func TestMCPCampaignDraftOwnershipAndStatus(t *testing.T) {
	db := testDB(t)
	team, listID, templateID, senderID := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	seed(t, db, &models.MailingList{Base: models.Base{ID: listID}, TeamID: team, Name: "Audience"})
	seed(t, db, &models.Template{Base: models.Base{ID: templateID}, TeamID: team, Name: "Template"})
	seed(t, db, &models.SMTPConfig{Base: models.Base{ID: senderID}, TeamID: team})
	h := handlers.NewMarketingHandler(db, "", "")
	e := echo.New()
	e.POST("/drafts", h.CreateCampaignDraft, func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error { c.Set("teamID", team); return next(c) }
	})
	payload := map[string]any{"name": "Launch", "subject": "Hello", "templateId": templateID, "listId": listID, "smtpConfigId": senderID}
	call := func() *httptest.ResponseRecorder {
		b, _ := json.Marshal(payload)
		r := httptest.NewRequest("POST", "/drafts", strings.NewReader(string(b)))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		e.ServeHTTP(w, r)
		return w
	}
	w := call()
	require.Equal(t, 201, w.Code, w.Body.String())
	var campaign models.Campaign
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &campaign))
	require.Equal(t, models.CampaignStatusDraft, campaign.Status)
	require.Equal(t, team, campaign.TeamID)
	require.NotEmpty(t, campaign.ID)
	payload["status"] = "SENDING"
	require.Equal(t, 400, call().Code)
	delete(payload, "status")
	payload["templateId"] = uuid.NewString()
	require.Equal(t, 400, call().Code)
	var count int64
	require.NoError(t, db.Model(&models.Campaign{}).Count(&count).Error)
	require.EqualValues(t, 1, count)
	require.NoError(t, db.Model(&models.Email{}).Count(&count).Error)
	require.Zero(t, count)
}

func TestMCPBatchRollsBackOnStorageFailure(t *testing.T) {
	db := testDB(t)
	team, listID := uuid.NewString(), uuid.NewString()
	seed(t, db, &models.MailingList{Base: models.Base{ID: listID}, TeamID: team, Name: "Audience"})
	require.NoError(t, db.Callback().Create().Before("gorm:create").Register("test:fail_contact", func(tx *gorm.DB) {
		if contact, ok := tx.Statement.Dest.(*models.Contact); ok && contact.Email == "fail@example.com" {
			tx.AddError(errors.New("simulated storage failure"))
		}
	}))
	h := handlers.NewMarketingHandler(db, "", "")
	e := echo.New()
	e.POST("/batch", h.ImportContactBatch, func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error { c.Set("teamID", team); return next(c) }
	})
	r := httptest.NewRequest("POST", "/batch", strings.NewReader(fmt.Sprintf(`{"listId":%q,"dryRun":false,"contacts":[{"email":"first@example.com"},{"email":"fail@example.com"}]}`, listID)))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)
	require.Equal(t, 500, w.Code)
	var count int64
	require.NoError(t, db.Model(&models.Contact{}).Count(&count).Error)
	require.Zero(t, count, "first row must roll back when second row fails")
}

func TestMCPAuthorizeChecksKeyExpiryAndDeletion(t *testing.T) {
	db := testDB(t)
	require.NoError(t, db.AutoMigrate(&models.APIKey{}))
	team := uuid.NewString()
	valid := models.APIKey{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Key: "valid", Name: "MCP"}
	seed(t, db, &valid)
	expired := models.APIKey{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Key: "expired", Name: "Expired", ExpiresAt: time.Now().Add(-time.Hour)}
	seed(t, db, &expired)
	expiredEast := models.APIKey{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Key: "expired-east", Name: "Expired east", ExpiresAt: time.Now().Add(-time.Hour).In(time.FixedZone("east", 12*3600))}
	seed(t, db, &expiredEast)
	validWest := models.APIKey{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Key: "valid-west", Name: "Valid west", ExpiresAt: time.Now().Add(time.Hour).In(time.FixedZone("west", -12*3600))}
	seed(t, db, &validWest)
	deleted := models.APIKey{Base: models.Base{ID: uuid.NewString(), IsDeleted: true}, TeamID: team, Key: "deleted", Name: "Deleted"}
	seed(t, db, &deleted)
	e := echo.New()
	e.GET("/authorize", handlers.MCPAuthorize(db))
	for key, status := range map[string]int{"valid": 204, "valid-west": 204, "expired": 401, "expired-east": 401, "deleted": 401, "unknown": 401, "": 401} {
		r := httptest.NewRequest("GET", "/authorize", nil)
		r.Header.Set("X-API-Key", key)
		w := httptest.NewRecorder()
		e.ServeHTTP(w, r)
		require.Equal(t, status, w.Code, w.Body.String())
		require.NotContains(t, w.Body.String(), team)
	}
}
