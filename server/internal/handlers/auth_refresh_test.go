package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"kori/internal/models"
	"kori/internal/utils"
)

// Exercise the real refresh handler and PostgreSQL SQL generation. Dry-run
// callbacks supply explicit fixture rows without opening a database connection.
func TestRefreshTokenUUID(t *testing.T) {
	t.Setenv("JWT_SECRET", "development-only-refresh-regression-signing-key")
	db, err := gorm.Open(postgres.Open("host=127.0.0.1 user=fixture dbname=fixture sslmode=disable"), &gorm.Config{
		DryRun: true, DisableAutomaticPing: true, SkipDefaultTransaction: true,
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { raw.Close() })
	user := models.User{
		Base:   models.Base{ID: "a034798b-3e6c-448b-8749-ea4e8f72dc08"},
		TeamID: "63a43d1e-963d-4d5d-b978-d19dc573c672",
		Email:  "admin@example.test", Role: models.UserRoleSuperAdmin,
	}
	refresh, err := utils.GenerateRefreshToken(user)
	if err != nil {
		t.Fatal(err)
	}
	session := models.AuthTransaction{
		Base:   models.Base{ID: "42715582-da98-42b4-bb1a-89ca3a4b2cbb"},
		UserID: user.ID, TeamID: user.TeamID, Refresh: refresh,
		Token: "previous-access-token", ExpiresAt: time.Now().Add(time.Hour),
	}
	queriedUser, savedSession := false, false
	err = db.Callback().Query().After("gorm:query").Register("test:refresh-fixtures", func(tx *gorm.DB) {
		switch row := tx.Statement.Dest.(type) {
		case *models.AuthTransaction:
			*row = session
			tx.RowsAffected = 1
		case *models.User:
			queriedUser = true
			if strings.Contains(tx.Statement.SQL.String(), user.ID) {
				t.Fatal("refresh UUID appeared as SQL text instead of a bound value")
			}
			bound := false
			for _, value := range tx.Statement.Vars {
				if text, ok := value.(string); ok && text == user.ID {
					bound = true
				}
			}
			if !bound {
				t.Fatal("refresh lookup did not bind its UUID")
			}
			*row = user
			tx.RowsAffected = 1
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	err = db.Callback().Update().After("gorm:update").Register("test:refresh-save", func(tx *gorm.DB) {
		if row, ok := tx.Statement.Dest.(*models.AuthTransaction); ok {
			if row.ID != session.ID || row.UserID != user.ID || row.Refresh != refresh || row.Token == "" || row.Token == session.Token {
				t.Fatal("refresh changed the session identity or failed to replace its access token")
			}
			savedSession = true
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]string{"refresh_token": refresh})
	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h := &AuthHandler{db: db}
	if err := h.RefreshToken(echo.New().NewContext(req, rec)); err != nil {
		t.Fatal(err)
	}
	if rec.Code != 200 || !queriedUser || !savedSession {
		t.Fatalf("refresh failed: HTTP %d; queried=%t saved=%t", rec.Code, queriedUser, savedSession)
	}
	var response struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	claims, err := utils.ParseJWT(response.Token)
	if err != nil || claims.UserID != user.ID || claims.TeamID != user.TeamID {
		t.Fatal("refresh returned a token for the wrong identity", err)
	}
}
