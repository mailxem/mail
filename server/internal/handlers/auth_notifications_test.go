package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"kori/internal/api/validator"
	"kori/internal/models"
)

func authNoticeFixture(t *testing.T) (*AuthHandler, models.User) {
	t.Helper()
	var dialect gorm.Dialector = sqlite.Open(":memory:")
	if dsn := os.Getenv("POSTHOOT_TEST_DATABASE_URL"); dsn != "" {
		admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		require.NoError(t, err)
		schema := "auth_notice_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
		require.NoError(t, admin.Exec("CREATE SCHEMA "+schema).Error)
		t.Cleanup(func() { admin.Exec("DROP SCHEMA " + schema + " CASCADE"); raw, _ := admin.DB(); raw.Close() })
		dialect = postgres.Open(dsn + " search_path=" + schema)
	}
	db, err := gorm.Open(dialect, &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true, Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.User{}, &models.Team{}, &models.PasswordReset{}, &models.AuthTransaction{}, &models.ServiceNotice{}, &models.TeamInvite{}, &models.TeamSettings{}, &models.BrandingSettings{}, &models.ResourcePermission{}, &models.UserPermission{}))
	raw, err := db.DB()
	require.NoError(t, err)
	if os.Getenv("POSTHOOT_TEST_DATABASE_URL") == "" {
		raw.SetMaxOpenConns(1)
	}
	t.Cleanup(func() { raw.Close() })
	hash, err := bcrypt.GenerateFromPassword([]byte("old-password"), bcrypt.MinCost)
	require.NoError(t, err)
	u := models.User{Base: models.Base{ID: uuid.NewString()}, Email: "ada@example.net", FirstName: "Ada", TeamID: uuid.NewString(), Role: models.UserRoleAdmin, Password: string(hash)}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&models.Team{Base: models.Base{ID: u.TeamID}, Name: "Studio", OwnerUserID: u.ID}).Error)
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&u).Error)
	return &AuthHandler{db: db, serviceEmails: true}, u
}
func authCall(t *testing.T, fn echo.HandlerFunc, body any) *httptest.ResponseRecorder {
	t.Helper()
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	e := echo.New()
	e.Validator = validator.NewValidator()
	req := httptest.NewRequest("POST", "/", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	require.NoError(t, fn(e.NewContext(req, rec)))
	return rec
}
func TestResetNoticeAndPasswordChangeCommitTogether(t *testing.T) {
	h, u := authNoticeFixture(t)
	first := authCall(t, h.RequestPasswordReset, map[string]string{"email": u.Email})
	require.Equal(t, 200, first.Code)
	unknown := authCall(t, h.RequestPasswordReset, map[string]string{"email": "unknown@example.net"})
	require.Equal(t, first.Body.String(), unknown.Body.String())
	again := authCall(t, h.RequestPasswordReset, map[string]string{"email": u.Email})
	require.Equal(t, first.Body.String(), again.Body.String())
	var n int64
	require.NoError(t, h.db.Model(&models.PasswordReset{}).Count(&n).Error)
	require.EqualValues(t, 1, n)
	var reset models.PasswordReset
	require.NoError(t, h.db.First(&reset).Error)
	require.Len(t, reset.Code, 32)
	var notice models.ServiceNotice
	require.NoError(t, h.db.First(&notice).Error)
	require.Equal(t, "password_reset", notice.Kind)
	require.Equal(t, u.ID, notice.UserID)
	require.Equal(t, reset.ID, notice.ReferenceID)
	session := models.AuthTransaction{UserID: u.ID, TeamID: u.TeamID, Token: "old-token", Refresh: "old-refresh", ExpiresAt: time.Now().Add(time.Hour)}
	require.NoError(t, h.db.Create(&session).Error)
	other := models.PasswordReset{UserID: u.ID, Code: "another-unused-code", ExpiresAt: time.Now().Add(time.Hour)}
	require.NoError(t, h.db.Create(&other).Error)
	response := authCall(t, h.VerifyResetCode, map[string]string{"code": reset.Code, "new_password": "new-password"})
	require.Equal(t, 200, response.Code, response.Body.String())
	require.NoError(t, h.db.First(&u, "id = ?", u.ID).Error)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("new-password")))
	require.NoError(t, h.db.Model(&models.PasswordReset{}).Where("used = false").Count(&n).Error)
	require.Zero(t, n)
	require.NoError(t, h.db.Model(&models.AuthTransaction{}).Count(&n).Error)
	require.Zero(t, n)
	require.NoError(t, h.db.Model(&models.ServiceNotice{}).Where("kind = ?", "password_changed").Count(&n).Error)
	require.EqualValues(t, 1, n)
	replay := authCall(t, h.VerifyResetCode, map[string]string{"code": reset.Code, "new_password": "different-password"})
	require.Equal(t, 400, replay.Code)
}
func TestPasswordChangeRollsBackIfConfirmationCannotBeQueued(t *testing.T) {
	h, u := authNoticeFixture(t)
	reset := models.PasswordReset{UserID: u.ID, Code: "valid-reset-code", ExpiresAt: time.Now().Add(time.Hour)}
	require.NoError(t, h.db.Create(&reset).Error)
	require.NoError(t, h.db.Migrator().DropTable(&models.ServiceNotice{}))
	response := authCall(t, h.VerifyResetCode, map[string]string{"code": reset.Code, "new_password": "new-password"})
	require.Equal(t, 500, response.Code)
	require.NoError(t, h.db.First(&u, "id = ?", u.ID).Error)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("old-password")))
	require.NoError(t, h.db.First(&reset, "id = ?", reset.ID).Error)
	require.False(t, reset.Used)
}
func TestNewResetInvalidatesPreviousLinkAndExpiredCodesFail(t *testing.T) {
	h, u := authNoticeFixture(t)
	old := models.PasswordReset{Base: models.Base{CreatedAt: time.Now().Add(-2 * time.Minute)}, UserID: u.ID, Code: "previous", ExpiresAt: time.Now().Add(time.Hour)}
	require.NoError(t, h.db.Create(&old).Error)
	response := authCall(t, h.RequestPasswordReset, map[string]string{"email": u.Email})
	require.Equal(t, 200, response.Code)
	require.NoError(t, h.db.First(&old, "id = ?", old.ID).Error)
	require.True(t, old.Used)
	expired := models.PasswordReset{UserID: u.ID, Code: "expired", ExpiresAt: time.Now().Add(-time.Second)}
	require.NoError(t, h.db.Create(&expired).Error)
	require.Equal(t, 400, authCall(t, h.VerifyResetCode, map[string]string{"code": expired.Code, "new_password": "new-password"}).Code)
}
func TestWelcomeQueuedOnlyForSuccessfulSignup(t *testing.T) {
	t.Setenv("TEST_MODE", "true")
	h, _ := authNoticeFixture(t)
	body := map[string]string{"email": "new@example.net", "password": "good-password", "first_name": "New", "last_name": "User"}
	response := authCall(t, h.Register, body)
	require.Equal(t, 201, response.Code, response.Body.String())
	var n int64
	require.NoError(t, h.db.Model(&models.ServiceNotice{}).Where("kind = ?", "welcome").Count(&n).Error)
	require.EqualValues(t, 1, n)
	require.Equal(t, 400, authCall(t, h.Register, body).Code)
	require.NoError(t, h.db.Model(&models.ServiceNotice{}).Where("kind = ?", "welcome").Count(&n).Error)
	require.EqualValues(t, 1, n)
}

func TestConcurrentResetConsumptionChangesPasswordOnlyOnce(t *testing.T) {
	h, u := authNoticeFixture(t)
	r := models.PasswordReset{UserID: u.ID, Code: "single-use-code", ExpiresAt: time.Now().Add(time.Hour)}
	require.NoError(t, h.db.Create(&r).Error)
	codes := make(chan int, 2)
	for _, password := range []string{"one-password", "two-password"} {
		go func(p string) {
			codes <- authCall(t, h.VerifyResetCode, map[string]string{"code": r.Code, "new_password": p}).Code
		}(password)
	}
	first, second := <-codes, <-codes
	require.ElementsMatch(t, []int{200, 400}, []int{first, second})
	var count int64
	require.NoError(t, h.db.Model(&models.ServiceNotice{}).Where("kind = ?", "password_changed").Count(&count).Error)
	require.EqualValues(t, 1, count)
}
