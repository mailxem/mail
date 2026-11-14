package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"kori/internal/models"
)

func TestListEmptyFieldsCountsActiveReadersWithBlanks(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.MailingList{}, &models.Contact{}))
	team, list := uuid.NewString(), uuid.NewString()
	quiet := db.Session(&gorm.Session{SkipHooks: true})
	require.NoError(t, quiet.Create(&models.MailingList{Base: models.Base{ID: list}, TeamID: team, Name: "News"}).Error)
	for _, c := range []models.Contact{
		{FirstName: "Ravi", LastName: "Rao", City: "Pune", Status: models.SubscriberStatusActive},
		{FirstName: "  ", LastName: "Lee", Status: models.SubscriberStatusActive},
		{Status: models.SubscriberStatusActive},
		{Status: models.SubscriberStatusUnsubscribed}, // not a recipient
	} {
		c.ID, c.TeamID, c.ListID, c.Email = uuid.NewString(), team, list, uuid.NewString()+"@example.com"
		require.NoError(t, quiet.Create(&c).Error)
	}

	h := NewMarketingHandler(db, "secret", "https://api.xem.email")
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(httptest.NewRequest("GET", "/", nil), rec)
	c.Set("teamID", team)
	c.SetParamNames("id")
	c.SetParamValues(list)
	require.NoError(t, h.ListEmptyFields(c))
	var got struct {
		Total int64            `json:"total"`
		Empty map[string]int64 `json:"empty"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Equal(t, int64(3), got.Total)
	require.Equal(t, int64(2), got.Empty["first_name"])
	require.Equal(t, int64(1), got.Empty["last_name"])
	require.Equal(t, int64(1), got.Empty["name"])
	require.Equal(t, int64(2), got.Empty["city"])
	require.Equal(t, int64(3), got.Empty["company"])

	// Another team's list is not found.
	c.Set("teamID", uuid.NewString())
	require.Error(t, h.ListEmptyFields(c))
}
