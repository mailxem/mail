package sending

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	em "github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/api/middleware"
	"kori/internal/models"
)

func workspace(c echo.Context) string { id, _ := c.Get("teamID").(string); return id }
func admin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		api, _ := c.Get("isAPIKey").(bool)
		allowed, _ := c.Get("hasAdminAccess").(bool)
		assistantWrite, _ := c.Get("assistantWrite").(bool)
		// Assistant credentials can inspect and manage sending, but never mint
		// reusable SMTP secrets. They have already passed actor and scope checks.
		if api {
			allowed = assistantWrite && !(c.Request().Method == "POST" && strings.HasSuffix(c.Path(), "/credentials"))
		}
		if (api && !assistantWrite) || !allowed || uuid.Validate(workspace(c)) != nil {
			return echo.NewHTTPError(403, "Workspace administrator access is required")
		}
		return next(c)
	}
}
func httpError(e error) error {
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return echo.NewHTTPError(404, "Sending resource not found")
	}
	if errors.Is(e, ErrDenied) {
		return echo.NewHTTPError(403, e.Error())
	}
	if errors.Is(e, ErrLimit) {
		return echo.NewHTTPError(429, e.Error())
	}
	if errors.Is(e, ErrSuppressed) {
		return echo.NewHTTPError(422, e.Error())
	}
	// Database/provider internals and message content must never leak in API errors.
	return echo.NewHTTPError(400, "Unable to complete this sending request. Check the domain, sender, message, and current limits.")
}
func (s *Service) Register(e *echo.Echo, jwtSecret string) {
	g := e.Group("/api/v1/sending", middleware.NewAuthMiddleware(jwtSecret).Middleware(), admin, em.BodyLimit("8M"), em.RateLimiterWithConfig(em.RateLimiterConfig{Store: em.NewRateLimiterMemoryStoreWithConfig(em.RateLimiterMemoryStoreConfig{Rate: rate.Every(time.Second), Burst: 10}), IdentifierExtractor: func(c echo.Context) (string, error) { return workspace(c), nil }}))
	g.GET("", s.status)
	g.PUT("/onboarding", s.onboarding)
	g.POST("/domains", s.addDomain)
	g.POST("/domains/:id/check", s.checkDomain)
	g.POST("/domains/:id/sender", s.activateSender)
	g.DELETE("/domains/:id", s.removeDomain)
	g.POST("/credentials", s.createCredential)
	g.DELETE("/credentials/:id", s.revokeCredential)
	g.PUT("/pause", s.pause)
	g.POST("/test", s.test)
	g.GET("/messages/:id/events", s.messageEvents)
	e.POST("/api/v1/sending/events/ses", s.feedback, em.BodyLimit("512K"), em.RateLimiter(em.NewRateLimiterMemoryStore(rate.Limit(50))))
}

type DNSRecord struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
}
type DomainView struct {
	Domain
	Records []DNSRecord `json:"records"`
}

func (s *Service) view(d Domain) DomainView {
	records := []DNSRecord{{"TXT", "_xem." + d.Name, d.Token}}
	for _, t := range strings.Split(d.DKIMTokens, ",") {
		if t != "" {
			records = append(records, DNSRecord{"CNAME", t + "._domainkey." + d.Name, t + ".dkim.amazonses.com"})
		}
	}
	if d.Provisioned {
		records = append(records, DNSRecord{"MX", "bounce." + d.Name, "10 feedback-smtp." + s.Config.Region + ".amazonses.com"}, DNSRecord{"TXT", "bounce." + d.Name, "v=spf1 include:amazonses.com ~all"})
	}
	return DomainView{d, records}
}
func (s *Service) status(c echo.Context) error {
	ctx := c.Request().Context()
	team := workspace(c)
	a, err := s.Account(ctx, team)
	if err != nil {
		return httpError(err)
	}
	domains := []Domain{}
	credentials := []Credential{}
	messages := []Message{}
	for _, item := range []struct {
		v     any
		limit int
	}{{&domains, 10}, {&credentials, 100}, {&messages, 50}} {
		if err = s.DB.WithContext(ctx).Where("team_id = ?", team).Order("created_at DESC").Limit(item.limit).Find(item.v).Error; err != nil {
			return httpError(err)
		}
	}
	var byo, campaigns, sent int64
	if err := s.DB.WithContext(ctx).Model(&models.SMTPConfig{}).Where("team_id = ? AND is_deleted = false AND is_active = true AND provider <> ?", team, "MANAGED").Count(&byo).Error; err != nil {
		return httpError(err)
	}
	if err := s.DB.WithContext(ctx).Model(&models.Campaign{}).Where("team_id = ? AND is_deleted = false", team).Count(&campaigns).Error; err != nil {
		return httpError(err)
	}
	if err := s.DB.WithContext(ctx).Model(&models.Email{}).Where("team_id = ? AND status IN ?", team, []string{"SENT", "OPENED", "CLICKED", "DELIVERED"}).Count(&sent).Error; err != nil {
		return httpError(err)
	}
	testedDomains := []string{}
	if err := s.DB.WithContext(ctx).Model(&Message{}).Where("team_id = ? AND is_test = true AND status IN ?", team, []string{"SENT", "DELIVERED"}).Distinct("domain_id").Pluck("domain_id", &testedDomains).Error; err != nil {
		return httpError(err)
	}
	views := []DomainView{}
	for _, d := range domains {
		views = append(views, s.view(d))
	}
	c.Response().Header().Set("Cache-Control", "no-store")
	return c.JSON(200, map[string]any{"enabled": s.Config.Enabled, "account": a, "domains": views, "credentials": credentials, "messages": messages, "smtp": map[string]any{"enabled": s.Config.SMTPEnabled, "host": s.Config.SMTPHost, "port": 587, "security": "STARTTLS"}, "costPerRecipientMicros": s.Config.CostMicros, "maxMessageBytes": s.Config.MaxBytes, "journey": map[string]any{"byoSenders": byo, "campaigns": campaigns, "acceptedEmails": sent, "testedDomainIds": testedDomains}})
}
func (s *Service) addDomain(c echo.Context) error {
	var in struct {
		Name string `json:"name"`
	}
	if c.Bind(&in) != nil {
		return echo.NewHTTPError(400, "Invalid domain")
	}
	d, e := s.AddDomain(c.Request().Context(), workspace(c), in.Name)
	if e != nil {
		return httpError(e)
	}
	return c.JSON(201, s.view(d))
}
func (s *Service) checkDomain(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 25*time.Second)
	defer cancel()
	d, e := s.RefreshDomain(ctx, workspace(c), c.Param("id"))
	if e != nil {
		return domainCheckHTTPError(e)
	}
	return c.JSON(200, s.view(d))
}
func (s *Service) createCredential(c echo.Context) error {
	var in struct {
		DomainID string `json:"domainId"`
		Name     string `json:"name"`
	}
	if c.Bind(&in) != nil {
		return echo.NewHTTPError(400, "Invalid credential")
	}
	cred, secret, e := s.CreateCredential(c.Request().Context(), workspace(c), in.DomainID, in.Name)
	if e != nil {
		return httpError(e)
	}
	c.Response().Header().Set("Cache-Control", "no-store")
	return c.JSON(201, map[string]any{"credential": cred, "password": secret})
}
func (s *Service) revokeCredential(c echo.Context) error {
	r := s.DB.WithContext(c.Request().Context()).Model(&Credential{}).Where("id = ? AND team_id = ? AND revoked_at IS NULL", c.Param("id"), workspace(c)).Update("revoked_at", s.Now())
	if r.Error != nil {
		return httpError(r.Error)
	}
	if r.RowsAffected == 0 {
		return echo.NewHTTPError(404, "Credential not found")
	}
	return c.NoContent(204)
}
func (s *Service) pause(c echo.Context) error {
	var in struct {
		Paused bool `json:"paused"`
	}
	if c.Bind(&in) != nil {
		return echo.NewHTTPError(400, "Invalid pause setting")
	}
	r := s.DB.WithContext(c.Request().Context()).Model(&Account{}).Where("team_id = ?", workspace(c)).Update("paused", in.Paused)
	if r.Error != nil {
		return httpError(r.Error)
	}
	return c.NoContent(204)
}
func (s *Service) activateSender(c echo.Context) error {
	var in struct {
		From string `json:"from"`
	}
	if c.Bind(&in) != nil {
		return echo.NewHTTPError(400, "Invalid sender")
	}
	from, err := address(in.From)
	if err != nil {
		return httpError(err)
	}
	err = s.DB.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
		var d Domain
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&d, "id = ? AND team_id = ? AND ready = true", c.Param("id"), workspace(c)).Error; e != nil {
			return e
		}
		if strings.Split(from, "@")[1] != d.Name {
			return ErrDenied
		}
		if d.SMTPConfigID != "" {
			if e := tx.Session(&gorm.Session{SkipHooks: true}).Model(&models.SMTPConfig{}).Where("id = ? AND team_id = ? AND provider = ?", d.SMTPConfigID, d.TeamID, "MANAGED").Update("from_email", from).Error; e != nil {
				return e
			}
			return tx.Model(&d).Update("from_email", from).Error
		}
		id := uuid.NewString()
		sender := models.SMTPConfig{Base: models.Base{ID: id}, Provider: "MANAGED", Host: "managed.internal", Port: 587, Username: "managed", FromEmail: from, TeamID: d.TeamID, IsActive: true, SupportsTLS: true, RequiresAuth: true, MaxSendRate: 1}
		if e := tx.Session(&gorm.Session{SkipHooks: true}).Create(&sender).Error; e != nil {
			return e
		}
		return tx.Model(&d).Updates(map[string]any{"smtp_config_id": id, "from_email": from}).Error
	})
	if err != nil {
		return httpError(err)
	}
	return c.NoContent(204)
}
func (s *Service) removeDomain(c echo.Context) error {
	err := s.DB.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
		var d Domain
		if e := tx.First(&d, "id = ? AND team_id = ?", c.Param("id"), workspace(c)).Error; e != nil {
			return e
		}
		if e := tx.Model(&Credential{}).Where("domain_id = ? AND team_id = ?", d.ID, d.TeamID).Update("revoked_at", s.Now()).Error; e != nil {
			return e
		}
		if d.SMTPConfigID != "" {
			if e := tx.Session(&gorm.Session{SkipHooks: true}).Model(&models.SMTPConfig{}).Where("id = ? AND team_id = ?", d.SMTPConfigID, d.TeamID).Updates(map[string]any{"is_active": false, "is_deleted": true}).Error; e != nil {
				return e
			}
		}
		// Keep the claim/history and AWS identity: automatic domain transfer is unsafe.
		return tx.Model(&d).Updates(map[string]any{"ready": false, "ownership": false, "token": "xem-managed=" + randomSecret(), "smtp_config_id": "", "from_email": ""}).Error
	})
	if err != nil {
		return httpError(err)
	}
	return c.NoContent(204)
}
func (s *Service) test(c echo.Context) error {
	var in struct {
		DomainID string `json:"domainId"`
		From     string `json:"from"`
		Key      string `json:"key"`
	}
	if c.Bind(&in) != nil {
		return echo.NewHTTPError(400, "Invalid test")
	}
	to, _ := c.Get("email").(string)
	to, err := address(to)
	if err != nil {
		return httpError(err)
	}
	from, err := address(in.From)
	if err != nil {
		return httpError(err)
	}
	// Fixed content and the authenticated administrator's address keep this from
	// becoming an arbitrary unaudited send endpoint.
	raw := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: Your Xem sending test\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nYour verified domain is connected to Xem. Check delivery status in Sending.\r\n", from, to))
	m, err := s.Submit(c.Request().Context(), Submission{TeamID: workspace(c), DomainID: in.DomainID, From: from, Recipients: []string{to}, Key: in.Key, Raw: raw, IsTest: true})
	if err != nil {
		return httpError(err)
	}
	return c.JSON(202, m)
}
func (s *Service) messageEvents(c echo.Context) error {
	var m Message
	if e := s.DB.WithContext(c.Request().Context()).First(&m, "id = ? AND team_id = ?", c.Param("id"), workspace(c)).Error; e != nil {
		return httpError(e)
	}
	events := []Event{}
	if e := s.DB.WithContext(c.Request().Context()).Where("message_id = ? AND team_id = ?", m.ID, m.TeamID).Order("created_at DESC").Limit(100).Find(&events).Error; e != nil {
		return httpError(e)
	}
	return c.JSON(200, events)
}
func (s *Service) feedback(c echo.Context) error {
	if !s.Config.Enabled {
		return echo.NewHTTPError(503, "Managed sending is disabled")
	}
	raw, err := io.ReadAll(io.LimitReader(c.Request().Body, 512*1024+1))
	if err != nil || len(raw) > 512*1024 {
		return echo.NewHTTPError(413, "Notification too large")
	}
	var n Notification
	if json.Unmarshal(raw, &n) != nil {
		return echo.NewHTTPError(400, "Invalid notification")
	}
	if err = s.VerifyNotification(c.Request().Context(), n); err != nil {
		return echo.NewHTTPError(403, "Notification verification failed")
	}
	if n.Type == "SubscriptionConfirmation" {
		err = s.ConfirmSubscription(c.Request().Context(), n)
	} else {
		err = s.ApplyFeedback(c.Request().Context(), n.MessageID, []byte(n.Message))
	}
	if err != nil {
		return echo.NewHTTPError(503, "Notification processing failed; retry required")
	}
	return c.NoContent(204)
}

func (s *Service) onboarding(c echo.Context) error {
	var in struct {
		Mode      string `json:"mode"`
		Dismissed bool   `json:"dismissed"`
	}
	if c.Bind(&in) != nil || (in.Mode != "MANAGED" && in.Mode != "BYO") {
		return echo.NewHTTPError(400, "Choose managed sending or your own provider")
	}
	if in.Mode == "MANAGED" && !s.Config.Enabled {
		return echo.NewHTTPError(400, "Managed sending is not enabled on this installation")
	}
	// Onboarding preferences are available even when delivery is disabled.
	a := Account{TeamID: workspace(c), DailyLimit: s.Config.DailyLimit, MonthlyLimit: s.Config.MonthlyLimit, MonthlyBudgetMicros: s.Config.BudgetMicros}
	if e := s.DB.WithContext(c.Request().Context()).Clauses(clause.OnConflict{DoNothing: true}).Create(&a).Error; e != nil {
		return httpError(e)
	}
	if e := s.DB.WithContext(c.Request().Context()).Model(&Account{}).Where("team_id = ?", workspace(c)).Updates(map[string]any{"sending_mode": in.Mode, "onboarding_dismissed": in.Dismissed}).Error; e != nil {
		return httpError(e)
	}
	return c.NoContent(204)
}
