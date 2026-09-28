package services

import (
	"crypto/rand"
	"crypto/rsa"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"kori/internal/config"
	"kori/internal/db"
	"kori/internal/models"
	"kori/internal/utils/crypto"
	"testing"
	"time"
)

func TestAPIEmailPersistsOutboxAndThreadingBeforeAcknowledgement(t *testing.T) {
	oldDB, oldCfg, oldPrivate, oldPublic := db.DB, cfg, crypto.PrivateKey, crypto.PublicKey
	t.Cleanup(func() { db.DB, cfg, crypto.PrivateKey, crypto.PublicKey = oldDB, oldCfg, oldPrivate, oldPublic })
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	crypto.PrivateKey, crypto.PublicKey = key, &key.PublicKey
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, database.AutoMigrate(&models.Email{}, &models.SMTPConfig{}, &models.EmailCategory{}))
	db.DB = database
	cfg = &config.Config{Server: config.ServerConfig{PublicURL: "https://api.example.com"}, JWT: config.JWTConfig{Secret: "test"}}
	team := uuid.NewString()
	sender := models.SMTPConfig{TeamID: team, Provider: "CUSTOM", Host: "smtp.example.com", Port: 465, Username: "sender@example.com", FromEmail: "sender@example.com", Password: "secret-password", IsActive: true}
	require.NoError(t, database.Create(&sender).Error)
	category := models.EmailCategory{TeamID: team, Name: "Transactional"}
	require.NoError(t, database.Create(&category).Error)
	scheduled := time.Now().Add(time.Hour).UTC().Truncate(time.Second)
	input := &models.Email{TeamID: team, SMTPConfigID: sender.ID, To: "reader@example.com", Subject: "Reply", Body: "<p>Hello</p>", InReplyTo: "<parent@example.com>", Test: true, SendAt: scheduled}
	require.NoError(t, QueueAPIEmail(input))
	var stored models.Email
	require.NoError(t, database.First(&stored).Error)
	require.NotNil(t, stored.DeliveryKey)
	require.Equal(t, "api:"+stored.ID, *stored.DeliveryKey)
	require.Equal(t, models.EmailStatusPending, stored.Status)
	require.Equal(t, input.InReplyTo, stored.InReplyTo)
	require.Equal(t, sender.ID, stored.SMTPConfigID)
	require.Equal(t, scheduled, stored.SendAt)
	input.TeamID = uuid.NewString()
	require.Error(t, QueueAPIEmail(input))
	var count int64
	require.NoError(t, database.Model(&models.Email{}).Count(&count).Error)
	require.EqualValues(t, 1, count)
}
