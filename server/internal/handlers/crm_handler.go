package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"kori/internal/models"
)

type CRMContact struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	FirstName      string    `json:"firstName"`
	LastName       string    `json:"lastName"`
	Company        string    `json:"company"`
	Phone          string    `json:"phone"`
	ListID         string    `json:"listId"`
	ListName       string    `json:"listName"`
	LifecycleStage string    `json:"lifecycleStage"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
}
type CRMCounts struct {
	Total      int64 `json:"total"`
	Qualified  int64 `json:"qualified"`
	Customers  int64 `json:"customers"`
	Subscribed int64 `json:"subscribed"`
}
type CRMPage struct {
	Data    []CRMContact `json:"data"`
	Total   int64        `json:"total"`
	Page    int          `json:"page"`
	Limit   int          `json:"limit"`
	Summary CRMCounts    `json:"summary"`
}

// CRMContacts counts the filtered collection before fetching a page. The summary
// covers the workspace, so its values do not change as the user changes pages.
func (h *MarketingHandler) CRMContacts(c echo.Context) error {
	if team(c) == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "Sign in to view contacts")
	}
	parsePage := func(name string, fallback, maximum int) (int, error) {
		raw := c.QueryParam(name)
		if raw == "" {
			return fallback, nil
		}
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > maximum {
			return 0, bad("Invalid " + name)
		}
		return value, nil
	}
	page, err := parsePage("page", 1, 1000000)
	if err != nil {
		return err
	}
	limit, err := parsePage("limit", 20, 100)
	if err != nil {
		return err
	}
	search := strings.TrimSpace(c.QueryParam("search"))
	if len(search) > 200 {
		return bad("Search must be at most 200 characters")
	}
	stage := strings.ToUpper(strings.TrimSpace(c.QueryParam("stage")))
	switch stage {
	case "", "LEAD", "QUALIFIED", "CUSTOMER", "LOST":
	default:
		return bad("Invalid lifecycle stage")
	}
	result := CRMPage{Data: []CRMContact{}, Page: page, Limit: limit}
	err = h.service.DB.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
		base := func() *gorm.DB {
			return tx.Model(&models.Contact{}).Where("contacts.team_id = ? AND contacts.is_deleted = false", team(c))
		}
		if err := base().Select(`COUNT(*) AS total,
   COALESCE(SUM(CASE WHEN lifecycle_stage = 'QUALIFIED' THEN 1 ELSE 0 END),0) AS qualified,
   COALESCE(SUM(CASE WHEN lifecycle_stage = 'CUSTOMER' THEN 1 ELSE 0 END),0) AS customers,
   COALESCE(SUM(CASE WHEN status = 'ACTIVE' THEN 1 ELSE 0 END),0) AS subscribed`).Scan(&result.Summary).Error; err != nil {
			return err
		}
		filtered := base()
		if stage != "" {
			filtered = filtered.Where("COALESCE(NULLIF(contacts.lifecycle_stage, ''), 'LEAD') = ?", stage)
		}
		if search != "" {
			// Search literal text, including % and _, rather than accepting SQL wildcards.
			escaped := strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(strings.ToLower(search))
			filtered = filtered.Where("LOWER(COALESCE(contacts.first_name, '') || ' ' || COALESCE(contacts.last_name, '') || ' ' || contacts.email || ' ' || COALESCE(contacts.company, '')) LIKE ? ESCAPE '!'", "%"+escaped+"%")
		}
		if err := filtered.Count(&result.Total).Error; err != nil {
			return err
		}
		// Clamp pages after deletions or stage changes so the UI never gets stranded
		// on an empty page while matching contacts still exist.
		pages := (result.Total + int64(limit) - 1) / int64(limit)
		if pages < 1 {
			pages = 1
		}
		if int64(result.Page) > pages {
			result.Page = int(pages)
		}
		return filtered.Select(`contacts.id, contacts.email, contacts.first_name, contacts.last_name,
   contacts.company, contacts.phone, contacts.list_id, contacts.status, contacts.created_at,
   COALESCE(NULLIF(contacts.lifecycle_stage, ''), 'LEAD') AS lifecycle_stage,
   COALESCE((SELECT name FROM mailing_lists WHERE mailing_lists.id = contacts.list_id AND mailing_lists.team_id = contacts.team_id AND mailing_lists.is_deleted = false), '') AS list_name`).
			Order("contacts.created_at DESC, contacts.id DESC").Offset((result.Page - 1) * limit).Limit(limit).Scan(&result.Data).Error
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, result)
}
