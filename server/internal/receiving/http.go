package receiving

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	em "github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"
	"kori/internal/api/middleware"
	"kori/internal/models"
	"kori/internal/sending"
)

func admin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		api, _ := c.Get("isAPIKey").(bool)
		role, _ := c.Get("role").(string)
		team, _ := c.Get("teamID").(string)
		user, _ := c.Get("userID").(string)
		if api || team == "" || user == "" || (role != string(models.UserRoleAdmin) && role != string(models.UserRoleSuperAdmin)) {
			return echo.NewHTTPError(403, "A workspace administrator must manage receiving")
		}
		return next(c)
	}
}
func team(c echo.Context) string { v, _ := c.Get("teamID").(string); return v }

func (s *Service) Register(e *echo.Echo, jwt string) {
	g := e.Group("/api/v1/mail-connections/receiving", em.BodyLimit("16K"), middleware.NewAuthMiddleware(jwt).Middleware(), admin)
	g.GET("", s.status)
	g.POST("/mailboxes", s.create)
	g.PATCH("/mailboxes/:id", s.patch)
	g.POST("/domains/:id/check", s.check)
}

type domainView struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Ownership       bool   `json:"ownership"`
	SendingReady    bool   `json:"sendingReady"`
	ReceivingStatus string `json:"receivingStatus"`
	MX              struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		Value    string `json:"value"`
		Priority uint16 `json:"priority"`
	} `json:"mx"`
	ExistingMX []MX   `json:"existingMX"`
	CheckedAt  any    `json:"checkedAt,omitempty"`
	Detail     string `json:"detail,omitempty"`
}

func (s *Service) status(c echo.Context) error {
	domains, states, boxes, err := s.List(c.Request().Context(), team(c))
	if err != nil {
		return echo.NewHTTPError(500, "Unable to load managed receiving")
	}
	var account sending.Account
	accountReady := s.DB.Where("team_id = ?", team(c)).First(&account).Error == nil && account.Approved && !account.Paused && !account.Suspended
	byID := map[string]DomainState{}
	for _, v := range states {
		byID[v.DomainID] = v
	}
	domainReady := map[string]bool{}
	views := make([]domainView, 0, len(domains))
	for _, d := range domains {
		state, ok := byID[d.ID]
		status := "unconfigured"
		if ok {
			status = state.Status
		}
		if s.Config.Enabled && !d.Ownership {
			status = "needs_verification"
		}
		if s.Config.Enabled && account.Suspended {
			status = "error"
			state.Detail = "Workspace receiving is suspended"
		}
		ready := accountReady && d.Ready && d.Ownership && d.Provisioned && d.CheckedAt != nil && s.Now().Sub(*d.CheckedAt) <= 24*time.Hour
		domainReady[d.ID] = ready
		v := domainView{ID: d.ID, Name: d.Name, Ownership: d.Ownership, SendingReady: ready, ReceivingStatus: status, ExistingMX: []MX{}, Detail: state.Detail}
		_ = json.Unmarshal(state.ExistingMXJSON, &v.ExistingMX)
		v.MX.Name = d.Name
		v.MX.Type = "MX"
		v.MX.Value = "inbound-smtp." + s.Config.Region + ".amazonaws.com"
		v.MX.Priority = 10
		if state.CheckedAt != nil {
			v.CheckedAt = state.CheckedAt
		}
		views = append(views, v)
	}
	for i := range boxes {
		if !boxes[i].Active || boxes[i].Status != "active" || !domainReady[boxes[i].DomainID] {
			boxes[i].SMTPConfigID = ""
		}
	}
	return c.JSON(200, map[string]any{"enabled": s.Config.Enabled, "region": s.Config.Region, "maxMailboxesPerDomain": s.Config.MaxMailboxesPerDomain, "maxMessageBytes": s.Config.MaxMessageBytes, "domains": views, "mailboxes": boxes})
}

func (s *Service) create(c echo.Context) error {
	var in struct {
		DomainID    string `json:"domainId"`
		LocalPart   string `json:"localPart"`
		DisplayName string `json:"displayName"`
	}
	if c.Bind(&in) != nil || uuid.Validate(in.DomainID) != nil {
		return echo.NewHTTPError(400, "Invalid mailbox")
	}
	box, err := s.CreateMailbox(c.Request().Context(), team(c), in.DomainID, in.LocalPart, in.DisplayName)
	if err != nil {
		return receivingHTTPError(err)
	}
	if box.Status != "active" {
		box.SMTPConfigID = ""
	}
	return c.JSON(201, box)
}
func (s *Service) patch(c echo.Context) error {
	var in struct {
		Active *bool `json:"active"`
	}
	if c.Bind(&in) != nil || in.Active == nil || uuid.Validate(c.Param("id")) != nil {
		return echo.NewHTTPError(400, "Invalid mailbox setting")
	}
	box, err := s.SetActive(c.Request().Context(), team(c), c.Param("id"), *in.Active)
	if err != nil {
		return receivingHTTPError(err)
	}
	if !box.Active || box.Status != "active" {
		box.SMTPConfigID = ""
	}
	return c.JSON(200, box)
}
func (s *Service) check(c echo.Context) error {
	if uuid.Validate(c.Param("id")) != nil {
		return echo.NewHTTPError(400, "Invalid domain")
	}
	state, mx, err := s.CheckDomain(c.Request().Context(), team(c), c.Param("id"))
	if err != nil {
		return receivingHTTPError(err)
	}
	return c.JSON(200, map[string]any{"receivingStatus": state.Status, "existingMX": mx, "checkedAt": state.CheckedAt, "detail": state.Detail})
}
func receivingHTTPError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return echo.NewHTTPError(404, "Receiving resource not found")
	}
	if err == ErrDisabled {
		return echo.NewHTTPError(503, "Managed receiving is disabled")
	}
	if errors.Is(err, ErrMailboxExists) {
		return echo.NewHTTPError(http.StatusConflict, "A mailbox already exists for this address")
	}
	for _, safe := range []string{"invalid mailbox local part", "display name is too long", "domain ownership must be verified", "mailbox limit reached"} {
		if err.Error() == safe {
			return echo.NewHTTPError(http.StatusBadRequest, safe)
		}
	}
	return echo.NewHTTPError(503, "Unable to update managed receiving; retry later")
}
