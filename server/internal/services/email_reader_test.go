package services

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"kori/internal/config"
	"kori/internal/models"
	"kori/internal/utils"
	"kori/internal/utils/base64"
)

// The email log reads bulk emails' full HTML although they store no body.
func TestEmailLogShowsRenderedBulkEmails(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, database.AutoMigrate(&models.Email{}, &models.EmailContent{}))
	old := utils.RenderConfig
	t.Cleanup(func() { utils.RenderConfig = old })
	utils.RenderConfig = &config.Config{JWT: config.JWTConfig{Secret: "s"}, Server: config.ServerConfig{PublicURL: "https://api.example.com"}}

	team := uuid.NewString()
	contentID, err := utils.StoreEmailContent(database, team, utils.ContentCampaign, `<p>Hi {{first_name}}</p><a href="https://example.com">x</a>`, "Hi", "")
	require.NoError(t, err)
	data, _ := json.Marshal(map[string]string{"first_name": "Asha"})
	tracked := true
	bulk := models.Email{Base: models.Base{ID: uuid.NewString()}, ClickTrackingEnabled: &tracked, ContentID: &contentID, Data: data, TeamID: team, From: "a@example.com", To: "b@example.com", Subject: "Hi", Status: models.EmailStatusSent, SMTPConfigID: uuid.NewString(), CategoryID: uuid.NewString()}
	single := bulk
	single.ID, single.ContentID, single.Body = uuid.NewString(), nil, base64.EncodeToBase64("<p>API email</p>")
	require.NoError(t, database.Session(&gorm.Session{SkipHooks: true}).Create([]models.Email{bulk, single}).Error)

	content := models.EmailContent{}
	require.NoError(t, database.First(&content, "id = ?", contentID).Error)
	html, err := utils.RenderEmail(&bulk, &content, utils.RenderConfig)
	require.NoError(t, err)
	require.Contains(t, html, "Hi Asha")
	require.Contains(t, html, "/t/click/")

	reader := NewEmailReader(database)
	got, err := reader.Get(context.Background(), bulk.ID)
	require.NoError(t, err)
	require.Equal(t, base64.EncodeToBase64(html), got.Body)

	list, total, err := reader.List(context.Background(), 1, 10, map[string]interface{}{"team_id": team}, nil, []string{"created_at"}, "ASC")
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	bodies := map[string]string{}
	for _, e := range list {
		bodies[e.ID] = e.Body
	}
	require.Equal(t, base64.EncodeToBase64(html), bodies[bulk.ID])
	require.Equal(t, single.Body, bodies[single.ID], "a stored body is returned as it is")
}
