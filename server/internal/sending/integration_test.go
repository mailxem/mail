package sending

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
	"kori/internal/config"
	"kori/internal/models"
	"kori/internal/tasks"
	"kori/internal/utils"
)

func TestCampaignTaskSubmissionCanRetryBeforeDurableAcceptance(t *testing.T) {
	s, team, d, p := setup(t)
	c, _ := apiContext(team, "POST", `{"from":"hello@example.com"}`)
	c.SetParamNames("id")
	c.SetParamValues(d.ID)
	require.NoError(t, s.activateSender(c))
	require.NoError(t, s.DB.First(&d, "id = ?", d.ID).Error)
	key := uuid.NewString()
	email := models.Email{Base: models.Base{ID: uuid.NewString()}, DeliveryKey: &key, TeamID: team, SMTPConfigID: d.SMTPConfigID, CategoryID: uuid.NewString(), From: "hello@example.com", To: "reader@example.net", Subject: "Campaign test", Body: base64.StdEncoding.EncodeToString([]byte("<p>Hello</p>")), Status: models.EmailStatusPending}
	require.NoError(t, s.DB.Session(&gorm.Session{SkipHooks: true}).Create(&email).Error)
	oldDelivery, oldPolicy := utils.ManagedDelivery, utils.RecipientPolicy
	t.Cleanup(func() { utils.ManagedDelivery = oldDelivery; utils.RecipientPolicy = oldPolicy })
	utils.RecipientPolicy = nil
	utils.ManagedDelivery = func(e *models.Email, m *gomail.Message) error {
		var raw bytes.Buffer
		if _, err := m.WriteTo(&raw); err != nil {
			return err
		}
		return s.SubmitEmail(context.Background(), e, raw.Bytes())
	}
	handler := tasks.NewTaskHandler(s.DB, &config.Config{})
	payload, err := json.Marshal(tasks.EmailTask{EmailID: email.ID})
	require.NoError(t, err)
	task := asynq.NewTask("email:send", payload)
	require.NoError(t, s.DB.Model(&Account{}).Where("team_id = ?", team).Update("paused", true).Error)
	require.Error(t, handler.HandleEmailSend(context.Background(), task))
	require.NoError(t, s.DB.First(&email, "id = ?", email.ID).Error)
	require.Equal(t, models.EmailStatusPending, email.Status, "a pre-queue refusal must remain retryable")
	require.NoError(t, s.DB.Model(&Account{}).Where("team_id = ?", team).Update("paused", false).Error)
	require.NoError(t, handler.HandleEmailSend(context.Background(), task))
	require.NoError(t, handler.HandleEmailSend(context.Background(), task))
	var count int64
	require.NoError(t, s.DB.Model(&Message{}).Count(&count).Error)
	require.EqualValues(t, 1, count)
	require.NoError(t, s.ProcessOne(context.Background()))
	require.Equal(t, 1, p.sends)
	require.NoError(t, s.DB.First(&email, "id = ?", email.ID).Error)
	require.Equal(t, models.EmailStatusSent, email.Status)
}
