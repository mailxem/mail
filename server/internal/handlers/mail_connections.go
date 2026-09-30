package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/mailconnect"
	"kori/internal/models"
)

type MailConnectionsHandler struct{ DB *gorm.DB }

func (h *MailConnectionsHandler) googleAvailable(c echo.Context) bool {
	var user models.User
	if err := h.DB.Select("email").Where("id = ? AND team_id = ? AND is_deleted = false", c.Get("userID"), c.Get("teamID")).First(&user).Error; err != nil {
		return false
	}
	return mailconnect.GoogleMailAccess(user.Email)
}

// Mailbox grants are workspace-wide in this release. Only a current workspace
// administrator can connect/disconnect them; API keys cannot manage credentials.
func MailConnectionAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		key, _ := c.Get("isAPIKey").(bool)
		role, _ := c.Get("role").(string)
		user, _ := c.Get("userID").(string)
		team, _ := c.Get("teamID").(string)
		if key || user == "" || team == "" || (role != string(models.UserRoleAdmin) && role != string(models.UserRoleSuperAdmin)) {
			return echo.NewHTTPError(403, "A workspace administrator must manage mail connections")
		}
		return next(c)
	}
}

func (h *MailConnectionsHandler) List(c echo.Context) error {
	rows := []models.MailConnection{}
	if err := h.DB.Where("team_id = ? AND active = true", c.Get("teamID")).Omit("secret").Find(&rows).Error; err != nil {
		return echo.NewHTTPError(500, "Unable to load connections")
	}
	_, err := mailconnect.GoogleConfig()
	return c.JSON(200, map[string]any{"connections": rows, "googleConfigured": err == nil, "googleAvailable": err == nil && h.googleAvailable(c)})
}

// These projections intentionally never load passwords or OAuth secrets.
func (h *MailConnectionsHandler) Mailboxes(c echo.Context) error {
	var rows = []struct {
		ID           string `json:"id"`
		Username     string `json:"username"`
		Host         string `json:"host"`
		SMTPConfigID string `json:"smtpConfigId,omitempty"`
		Provider     string `json:"provider"`
	}{}
	// Legacy IMAP forms never set is_active; preserve their existing behavior.
	// OAuth disconnection is authoritative on the linked connection row.
	err := h.DB.Table("imap_configs AS i").Select("i.id, i.username, i.host, m.smtp_config_id, CASE WHEN r.id IS NOT NULL THEN 'CLOUDFLARE' ELSE COALESCE(NULLIF(m.provider, ''), 'CUSTOM') END AS provider").Joins("LEFT JOIN mail_connections m ON m.imap_config_id = i.id AND m.team_id = i.team_id").Joins("LEFT JOIN cloudflare_relays r ON r.imap_config_id = i.id AND r.team_id = i.team_id").Where("i.team_id = ? AND i.is_deleted = false AND (m.id IS NULL OR m.active = true) AND (r.id IS NULL OR r.enabled = true)", c.Get("teamID")).Scan(&rows).Error
	if err != nil {
		return echo.NewHTTPError(500, "Unable to load mailboxes")
	}
	if h.DB.Migrator().HasTable("managed_receiving_mailboxes") {
		var managedRows = []struct {
			ID           string `json:"id"`
			Username     string `json:"username"`
			Host         string `json:"host"`
			SMTPConfigID string `json:"smtpConfigId,omitempty"`
			Provider     string `json:"provider"`
		}{}
		cutoff := time.Now().UTC().Add(-24 * time.Hour)
		if err := h.DB.Table("managed_receiving_mailboxes AS b").Select("b.id, b.address AS username, d.name AS host, CASE WHEN b.active = true AND b.status = 'active' AND d.ready = true AND d.ownership = true AND d.provisioned = true AND d.checked_at >= ? AND a.approved = true AND a.paused = false AND a.suspended = false THEN b.smtp_config_id ELSE '' END AS smtp_config_id, 'MANAGED' AS provider", cutoff).Joins("JOIN managed_domains d ON d.id = b.domain_id AND d.team_id = b.team_id").Joins("JOIN managed_accounts a ON a.team_id = b.team_id").Where("b.team_id = ?", c.Get("teamID")).Scan(&managedRows).Error; err != nil {
			return echo.NewHTTPError(500, "Unable to load mailboxes")
		}
		rows = append(rows, managedRows...)
	}
	return c.JSON(200, rows)
}
func (h *MailConnectionsHandler) Senders(c echo.Context) error {
	var rows = []struct {
		ID        string `json:"id"`
		FromEmail string `json:"fromEmail"`
		Provider  string `json:"provider"`
		IsDefault bool   `json:"isDefault"`
	}{}
	query := h.DB.Table("smtp_configs AS s").Select("s.id, s.from_email, s.provider, s.is_default").Joins("LEFT JOIN mail_connections m ON m.smtp_config_id = s.id AND m.team_id = s.team_id").Where("s.team_id = ? AND s.is_active = true AND s.is_deleted = false AND (m.id IS NULL OR m.active = true)", c.Get("teamID"))
	if h.DB.Migrator().HasTable("managed_receiving_mailboxes") {
		cutoff := time.Now().UTC().Add(-24 * time.Hour)
		query = query.Where("NOT EXISTS (SELECT 1 FROM managed_receiving_mailboxes b WHERE b.smtp_config_id = s.id) OR EXISTS (SELECT 1 FROM managed_receiving_mailboxes b JOIN managed_domains d ON d.id = b.domain_id AND d.team_id = b.team_id JOIN managed_accounts a ON a.team_id = b.team_id WHERE b.smtp_config_id = s.id AND b.team_id = s.team_id AND b.active = true AND b.status = 'active' AND d.ready = true AND d.ownership = true AND d.provisioned = true AND d.checked_at >= ? AND a.approved = true AND a.paused = false AND a.suspended = false)", cutoff)
	}
	err := query.Scan(&rows).Error
	if err != nil {
		return echo.NewHTTPError(500, "Unable to load senders")
	}
	return c.JSON(200, rows)
}

func (h *MailConnectionsHandler) GoogleStart(c echo.Context) error {
	if !h.googleAvailable(c) {
		return echo.NewHTTPError(403, "Google mailbox connections are in a limited test release. Ask the installation administrator for access.")
	}
	cfg, err := mailconnect.GoogleConfig()
	if err != nil {
		return echo.NewHTTPError(503, err.Error())
	}
	state := oauth2.GenerateVerifier()
	row := models.MailOAuthState{Hash: mailconnect.StateHash(state), TeamID: c.Get("teamID").(string), UserID: c.Get("userID").(string), Verifier: oauth2.GenerateVerifier(), ExpiresAt: time.Now().Add(10 * time.Minute)}
	err = h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("expires_at < ? OR (team_id = ? AND user_id = ?)", time.Now(), row.TeamID, row.UserID).Delete(&models.MailOAuthState{}).Error; err != nil {
			return err
		}
		return tx.Create(&row).Error
	})
	if err != nil {
		return echo.NewHTTPError(500, "Unable to start Google authorization")
	}
	return c.JSON(200, map[string]string{"url": cfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent"), oauth2.S256ChallengeOption(row.Verifier))})
}

func consumeMailState(tx *gorm.DB, hash, team, user string) (*models.MailOAuthState, error) {
	var state models.MailOAuthState
	err := tx.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("hash = ? AND team_id = ? AND user_id = ? AND expires_at > ?", hash, team, user, time.Now()).First(&state).Error; err != nil {
			return err
		}
		result := tx.Where("hash = ?", state.Hash).Delete(&models.MailOAuthState{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return errors.New("state already used")
		}
		return nil
	})
	return &state, err
}

func (h *MailConnectionsHandler) GoogleComplete(c echo.Context) error {
	var input struct {
		Code  string `json:"code"`
		State string `json:"state"`
	}
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, 16384)
	if err := c.Bind(&input); err != nil || input.Code == "" || len(input.State) < 32 {
		return echo.NewHTTPError(400, "Invalid Google authorization response")
	}
	cfg, err := mailconnect.GoogleConfig()
	if err != nil {
		return echo.NewHTTPError(503, err.Error())
	}
	state, err := consumeMailState(h.DB, mailconnect.StateHash(input.State), c.Get("teamID").(string), c.Get("userID").(string))
	if err != nil {
		return echo.NewHTTPError(400, "Authorization expired or already used. Start again.")
	}
	ctx := mailconnect.OAuthContext(c.Request().Context())
	token, err := cfg.Exchange(ctx, input.Code, oauth2.VerifierOption(state.Verifier))
	if err != nil {
		return echo.NewHTTPError(400, "Google authorization failed. Start again.")
	}
	scope, _ := token.Extra("scope").(string)
	granted := false
	for _, s := range strings.Fields(scope) {
		if s == mailconnect.MailScope {
			granted = true
		}
	}
	if token.RefreshToken == "" || !granted {
		return echo.NewHTTPError(400, "Grant mailbox access and offline access to connect Gmail")
	}
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://gmail.googleapis.com/gmail/v1/users/me/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	response, err := mailconnect.HTTPClient().Do(req)
	if err != nil {
		return echo.NewHTTPError(502, "Unable to identify Google mailbox")
	}
	defer response.Body.Close()
	var profile struct {
		Email string `json:"emailAddress"`
	}
	if response.StatusCode != 200 || json.NewDecoder(io.LimitReader(response.Body, 16384)).Decode(&profile) != nil {
		return echo.NewHTTPError(502, "Unable to identify Google mailbox")
	}
	address, err := mail.ParseAddress(profile.Email)
	if err != nil || address.Address != profile.Email {
		return echo.NewHTTPError(502, "Google returned an invalid mailbox address")
	}
	if !h.googleAvailable(c) || !mailconnect.GoogleMailAccess(address.Address) {
		return echo.NewHTTPError(403, "This Google mailbox is not approved for the test release. No mailbox connection was saved.")
	}
	connection := models.MailConnection{Base: models.Base{ID: uuid.NewString()}, TeamID: state.TeamID, Provider: mailconnect.Google, Address: address.Address, Active: true}
	if err := mailconnect.Seal(&connection, token); err != nil {
		return echo.NewHTTPError(500, "Unable to protect credentials")
	}
	if err := h.createConnection(&connection); err != nil {
		return echo.NewHTTPError(500, "Unable to save mailbox connection")
	}
	return c.JSON(201, connection)
}

var accountIDPattern = regexp.MustCompile(`^[a-fA-F0-9]{32}$`)

func (h *MailConnectionsHandler) CloudflareCreate(c echo.Context) error {
	var input struct {
		AccountID string `json:"accountId"`
		Token     string `json:"token"`
		Address   string `json:"address"`
	}
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, 16384)
	if c.Bind(&input) != nil || !accountIDPattern.MatchString(input.AccountID) || len(input.Token) < 16 || len(input.Token) > 4096 || strings.ContainsAny(input.Token, "\r\n") {
		return echo.NewHTTPError(400, "A Cloudflare account ID and sending API token are required")
	}
	address, err := mail.ParseAddress(input.Address)
	if err != nil || address.Address != input.Address {
		return echo.NewHTTPError(400, "Enter a sender email address on a domain enabled for Cloudflare Email Sending")
	}
	connection := models.MailConnection{Base: models.Base{ID: uuid.NewString()}, TeamID: c.Get("teamID").(string), Provider: mailconnect.Cloudflare, Address: address.Address, AccountID: input.AccountID, Active: true}
	if err := mailconnect.Seal(&connection, input.Token); err != nil {
		return echo.NewHTTPError(500, "Unable to protect credentials")
	}
	// Save only: connecting must never send a test email without a send request.
	if err := h.createConnection(&connection); err != nil {
		return echo.NewHTTPError(500, "Unable to save Cloudflare sender")
	}
	return c.JSON(201, connection)
}

func (h *MailConnectionsHandler) createConnection(c *models.MailConnection) error {
	return h.DB.Transaction(func(tx *gorm.DB) error {
		sender := models.SMTPConfig{Provider: c.Provider, Host: "api.cloudflare.com", Port: 443, Username: c.Address, FromEmail: c.Address, IsActive: true, MaxSendRate: 1, SupportsTLS: true, RequiresAuth: true, TeamID: c.TeamID}
		if c.Provider == mailconnect.Google {
			sender.Host = "smtp.gmail.com"
			sender.Port = 465
		}
		if err := tx.Create(&sender).Error; err != nil {
			return err
		}
		c.SMTPConfigID = sender.ID
		if c.Provider == mailconnect.Google {
			inbox := models.IMAPConfig{Host: "imap.gmail.com", Port: 993, Username: c.Address, IsActive: true, TeamID: c.TeamID}
			if err := tx.Create(&inbox).Error; err != nil {
				return err
			}
			c.IMAPConfigID = &inbox.ID
		}
		return tx.Create(c).Error
	})
}

func (h *MailConnectionsHandler) Disconnect(c echo.Context) error {
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		var connection models.MailConnection
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND team_id = ?", c.Param("id"), c.Get("teamID")).First(&connection).Error; err != nil {
			return err
		}
		if err := tx.Model(&connection).Updates(map[string]any{"active": false, "secret": ""}).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.SMTPConfig{}).Where("id = ? AND team_id = ?", connection.SMTPConfigID, connection.TeamID).UpdateColumn("is_active", false).Error; err != nil {
			return err
		}
		if connection.IMAPConfigID != nil {
			return tx.Model(&models.IMAPConfig{}).Where("id = ? AND team_id = ?", *connection.IMAPConfigID, connection.TeamID).UpdateColumn("is_active", false).Error
		}
		return nil
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return echo.NewHTTPError(404, "Connection not found")
	}
	if err != nil {
		return echo.NewHTTPError(500, "Unable to disconnect")
	}
	return c.NoContent(204)
}
