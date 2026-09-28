package marketing_test

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"kori/internal/handlers"
	"kori/internal/models"
	"kori/internal/services"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTagsLifecycleAndTenantIsolation(t *testing.T) {
	db := testDB(t)
	require.NoError(t, db.AutoMigrate(&models.Tag{}))
	owner, other := uuid.NewString(), uuid.NewString()
	seed(t, db, &models.Team{Base: models.Base{ID: owner}, Name: "Workspace"})
	contact := models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: owner, ListID: uuid.NewString(), Email: "reader@example.com"}
	foreign := models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: other, ListID: uuid.NewString(), Email: "other@example.com"}
	seed(t, db, &contact)
	seed(t, db, &foreign)
	h := handlers.NewMarketingHandler(db, "test-secret", "")
	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error { c.Set("teamID", owner); return next(c) }
	})
	e.GET("/tags", h.Tags)
	e.POST("/tags", h.SaveTag)
	e.PUT("/tags/:id", h.SaveTag)
	e.DELETE("/tags/:id", h.DeleteTag)
	e.GET("/contacts/:id/tags", h.ContactTags)
	e.PUT("/contacts/:id/tags", h.ContactTags)
	call := func(method, path, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		e.ServeHTTP(w, r)
		return w
	}
	result := call("POST", "/tags", `{"name":"VIP"}`)
	require.Equal(t, 200, result.Code, result.Body.String())
	var tag models.Tag
	require.NoError(t, json.Unmarshal(result.Body.Bytes(), &tag))
	require.Equal(t, owner, tag.TeamID)
	require.Equal(t, 409, call("POST", "/tags", `{"name":"vip"}`).Code)
	selected := `{"tagIds":["` + tag.ID + `"]}`
	require.Equal(t, 200, call("PUT", "/contacts/"+contact.ID+"/tags", selected).Code)
	require.Equal(t, 200, call("PUT", "/contacts/"+contact.ID+"/tags", selected).Code)
	require.Contains(t, call("GET", "/tags", "").Body.String(), `"contactCount":1`)
	require.Equal(t, 404, call("PUT", "/contacts/"+foreign.ID+"/tags", selected).Code)
	foreignTag := models.Tag{Base: models.Base{ID: uuid.NewString()}, TeamID: other, Name: "Private"}
	seed(t, db, &foreignTag)
	require.Equal(t, 404, call("PUT", "/contacts/"+contact.ID+"/tags", `{"tagIds":["`+foreignTag.ID+`"]}`).Code)
	require.Equal(t, 404, call("DELETE", "/tags/"+foreignTag.ID, "").Code)
	require.Equal(t, 200, call("PUT", "/tags/"+tag.ID, `{"name":"Customers"}`).Code)
	require.Contains(t, call("GET", "/contacts/"+contact.ID+"/tags", "").Body.String(), "Customers")
	require.Equal(t, 204, call("DELETE", "/tags/"+tag.ID, "").Code)
	require.Equal(t, "[]\n", call("GET", "/contacts/"+contact.ID+"/tags", "").Body.String())
	var count int64
	require.NoError(t, db.Model(&models.Contact{}).Where("id = ?", contact.ID).Count(&count).Error)
	require.EqualValues(t, 1, count)
}
func TestListCountRemainsStableOnLaterPages(t *testing.T) {
	db := testDB(t)
	owner := uuid.NewString()
	for _, name := range []string{"First", "Second", "Third"} {
		seed(t, db, &models.MailingList{Base: models.Base{ID: uuid.NewString()}, TeamID: owner, Name: name})
	}
	service := services.NewBaseService(db, models.MailingList{})
	rows, total, err := service.List(context.Background(), 2, 2, map[string]interface{}{"team_id": owner}, nil, []string{"name"}, "ASC")
	require.NoError(t, err)
	require.EqualValues(t, 3, total)
	require.Len(t, rows, 1)
}

func TestLegacyTagsAreSplitByWorkspaceAndMigrationIsIdempotent(t *testing.T) {
	db := testDB(t)
	first, second := uuid.NewString(), uuid.NewString()
	legacy := models.Tag{Base: models.Base{ID: uuid.NewString()}, Name: "Legacy VIP", Value: "gold"}
	require.NoError(t, db.Omit("TeamID").Create(&legacy).Error)
	for _, owner := range []string{first, second} {
		contact := models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: owner, ListID: uuid.NewString(), Email: owner + "@example.com"}
		seed(t, db, &contact)
		require.NoError(t, db.Exec("INSERT INTO contact_tags (contact_id, tag_id) VALUES (?, ?)", contact.ID, legacy.ID).Error)
	}
	require.NoError(t, db.Transaction(models.BackfillTagWorkspaces))
	require.NoError(t, db.Transaction(models.BackfillTagWorkspaces))
	var count int64
	require.NoError(t, db.Model(&models.Tag{}).Where("team_id IS NOT NULL").Count(&count).Error)
	require.EqualValues(t, 2, count)
	require.NoError(t, db.Table("contact_tags").Joins("JOIN contacts ON contacts.id = contact_tags.contact_id").Joins("JOIN tags ON tags.id = contact_tags.tag_id").Where("contacts.team_id = tags.team_id").Count(&count).Error)
	require.EqualValues(t, 2, count)
	require.NoError(t, db.Table("contact_tags").Where("tag_id = ?", legacy.ID).Count(&count).Error)
	require.Zero(t, count)
}
