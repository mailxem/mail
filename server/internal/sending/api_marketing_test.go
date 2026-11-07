package sending

import (
	"bytes"
	"context"
	"encoding/base64"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gopkg.in/gomail.v2"
	"kori/internal/models"
	"kori/internal/utils"
)

// API sends always carry a contact; only campaign, unsubscribe link or the Marketing category make them marketing.
func TestAPIEmailMarketingFollowsCategory(t *testing.T) {
	s, team, d, _ := setup(t)
	require.NoError(t, s.DB.AutoMigrate(&models.EmailCategory{}))
	categories := map[string]string{}
	for _, name := range []string{"Transactional", "Marketing"} {
		c := models.EmailCategory{Name: name, TeamID: team}
		c.ID = uuid.NewString()
		require.NoError(t, s.DB.Create(&c).Error)
		categories[name] = c.ID
	}
	var raw []byte
	oldDelivery, oldPolicy := utils.ManagedDelivery, utils.RecipientPolicy
	t.Cleanup(func() { utils.ManagedDelivery = oldDelivery; utils.RecipientPolicy = oldPolicy })
	utils.RecipientPolicy = nil
	utils.ManagedDelivery = func(e *models.Email, m *gomail.Message) error {
		var buf bytes.Buffer
		if _, err := m.WriteTo(&buf); err != nil {
			return err
		}
		raw = buf.Bytes()
		return s.SubmitEmail(context.Background(), e, raw)
	}
	send := func(category, unsub string) error {
		email := &models.Email{Base: models.Base{ID: uuid.NewString()}, TeamID: team, SMTPConfigID: d.SMTPConfigID, SMTPConfig: &models.SMTPConfig{TeamID: team, IsActive: true, Provider: "MANAGED"}, ContactID: uuid.NewString(), CategoryID: categories[category], UnsubscribeURL: unsub, From: "hello@example.com", To: "reader@example.net", Subject: "Hi", Body: base64.StdEncoding.EncodeToString([]byte("<p>Hi</p>")), Status: models.EmailStatusPending}
		return utils.NewEmailHandler(1).SendEmail(email)
	}

	require.NoError(t, send("Transactional", ""), "transactional API mail with a contact is not marketing")
	require.NotContains(t, string(raw), "List-Unsubscribe")

	require.ErrorContains(t, send("Marketing", ""), "one-click unsubscribe")

	require.NoError(t, send("Marketing", "https://api.example.com/public/unsubscribe/x"))
	require.Contains(t, string(raw), "List-Unsubscribe: <https://api.example.com/public/unsubscribe/x>")
	require.Contains(t, string(raw), "List-Unsubscribe-Post: List-Unsubscribe=One-Click")
}
