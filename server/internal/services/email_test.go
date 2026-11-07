package services

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"kori/internal/config"
	"kori/internal/db"
	"kori/internal/models"
	"kori/internal/utils/base64"
	cryptoutil "kori/internal/utils/crypto"
)

// Editor templates keep raw HTML in HTMLBody and a raw subject, with no HtmlFile.
func TestSendEmail_EditorTemplateUsesHTMLBody(t *testing.T) {
	testDB, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true, Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, testDB.AutoMigrate(&models.Template{}, &models.SMTPConfig{}, &models.EmailCategory{}, &models.Email{}, &models.File{}))
	prevDB, prevCfg := db.DB, cfg
	db.DB, cfg = testDB, &config.Config{}
	t.Cleanup(func() { db.DB, cfg = prevDB, prevCfg })

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	cryptoutil.PrivateKey, cryptoutil.PublicKey = key, &key.PublicKey

	teamID := uuid.NewString()
	seed := testDB.Session(&gorm.Session{SkipHooks: true})
	category := &models.EmailCategory{Name: "Transactional", TeamID: teamID}
	category.ID = uuid.NewString()
	smtp := &models.SMTPConfig{TeamID: teamID, IsDefault: true, FromEmail: "hi@example.com", Password: "secret"}
	smtp.ID = uuid.NewString()
	template := &models.Template{TeamID: teamID, CategoryID: category.ID, Name: "editor", Subject: "Hello {{ name }}", HTMLBody: "<p>Hi {{ name }}</p>"}
	template.ID = uuid.NewString()
	require.NoError(t, seed.Create(category).Error)
	require.NoError(t, testDB.Create(smtp).Error)
	require.NoError(t, seed.Create(template).Error)

	require.NoError(t, sendEmail(&sendEmailHandlerBody{teamId: teamID, templateId: template.ID, to: "ada@example.com", variables: map[string]string{"name": "Ada"}, testMail: true}))

	var email models.Email
	require.NoError(t, testDB.First(&email, "template_id = ?", template.ID).Error)
	body, err := base64.DecodeFromBase64(email.Body)
	require.NoError(t, err)
	require.Contains(t, body, "<p>Hi Ada</p>")
	require.Equal(t, "Hello Ada", email.Subject)
}
