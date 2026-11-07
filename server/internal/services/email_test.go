package services

import (
	"crypto/rand"
	"crypto/rsa"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"kori/internal/config"
	"kori/internal/db"
	"kori/internal/marketing"
	"kori/internal/models"
	"kori/internal/utils/base64"
	cryptoutil "kori/internal/utils/crypto"
)

// emailTestDB points the email service at an in-memory DB with a default SMTP config.
func emailTestDB(t *testing.T) (*gorm.DB, string) {
	testDB, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true, Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, testDB.AutoMigrate(&models.Template{}, &models.SMTPConfig{}, &models.EmailCategory{}, &models.Email{}, &models.File{}, &models.MailingList{}, &models.Contact{}, &models.ContactImport{}))
	prevDB, prevCfg := db.DB, cfg
	db.DB, cfg = testDB, &config.Config{}
	t.Cleanup(func() { db.DB, cfg = prevDB, prevCfg })

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	cryptoutil.PrivateKey, cryptoutil.PublicKey = key, &key.PublicKey

	teamID := uuid.NewString()
	smtp := &models.SMTPConfig{TeamID: teamID, IsDefault: true, FromEmail: "hi@example.com", Password: "secret"}
	smtp.ID = uuid.NewString()
	require.NoError(t, testDB.Create(smtp).Error)
	return testDB, teamID
}

// Editor templates keep raw HTML in HTMLBody and a raw subject, with no HtmlFile.
func TestSendEmail_EditorTemplateUsesHTMLBody(t *testing.T) {
	testDB, teamID := emailTestDB(t)
	seed := testDB.Session(&gorm.Session{SkipHooks: true})
	category := &models.EmailCategory{Name: "Transactional", TeamID: teamID}
	category.ID = uuid.NewString()
	template := &models.Template{TeamID: teamID, CategoryID: category.ID, Name: "editor", Subject: "Hello {{ name }}", HTMLBody: "<p>Hi {{ name }}</p>"}
	template.ID = uuid.NewString()
	require.NoError(t, seed.Create(category).Error)
	require.NoError(t, seed.Create(template).Error)

	require.NoError(t, sendEmail(&sendEmailHandlerBody{teamId: teamID, templateId: template.ID, to: "ada@example.com", variables: map[string]string{"name": "Ada"}, testMail: true}))

	var email models.Email
	require.NoError(t, testDB.First(&email, "template_id = ?", template.ID).Error)
	body, err := base64.DecodeFromBase64(email.Body)
	require.NoError(t, err)
	require.Contains(t, body, "<p>Hi Ada</p>")
	require.Equal(t, "Hello Ada", email.Subject)
}

// Marketing-category API mail gets the campaign-style one-click unsubscribe link; transactional mail does not.
func TestSendEmail_MarketingCategoryGetsUnsubscribeURL(t *testing.T) {
	testDB, teamID := emailTestDB(t)
	cfg.JWT.Secret = strings.Repeat("s", 32)
	t.Setenv("PUBLIC_API_URL", "https://api.example.com/")
	seed := testDB.Session(&gorm.Session{SkipHooks: true})
	list := &models.MailingList{Name: "All Users", TeamID: teamID}
	list.ID = uuid.NewString()
	require.NoError(t, seed.Create(list).Error)
	for _, name := range []string{"Marketing", "Transactional"} {
		category := &models.EmailCategory{Name: name, TeamID: teamID}
		category.ID = uuid.NewString()
		require.NoError(t, seed.Create(category).Error)
		require.NoError(t, sendEmail(&sendEmailHandlerBody{teamId: teamID, to: "ada@example.com", subject: name, body: "<p>Hi</p>", categoryId: category.ID}))

		var email models.Email
		require.NoError(t, testDB.First(&email, "category_id = ?", category.ID).Error)
		if name == "Marketing" {
			require.Equal(t, "https://api.example.com/public/unsubscribe/"+teamID+"/"+email.ContactID+"/"+marketing.New(nil, cfg.JWT.Secret, "").MessageToken(teamID, email.ContactID, email.ID)+"?message="+email.ID, email.UnsubscribeURL)
		} else {
			require.Empty(t, email.UnsubscribeURL)
		}
	}
}
