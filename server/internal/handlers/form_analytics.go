package handlers

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/marketing"
	"kori/internal/models"
)

func recordFormEvent(tx *gorm.DB, f models.Form, session, event, page string) error {
	// A page/event contributes once per revision and session across retries.
	// Lifetime summary counts remain distinct by session across revisions.
	row := models.FormAnalyticsEvent{ID: formDigest(f.ID + ":" + strconv.Itoa(f.Version) + ":" + session + ":" + event + ":" + page), FormID: f.ID, SessionID: session, Version: f.Version, Event: event, PageID: page}
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error
}
func (h *MarketingHandler) RecordFormEvent(c echo.Context) error {
	var in struct {
		SessionID string `json:"sessionId"`
		Event     string `json:"event"`
		PageID    string `json:"pageId"`
		Version   int    `json:"version"`
		Website   string `json:"website"`
	}
	if err := c.Bind(&in); err != nil {
		return bad("Invalid event")
	}
	if in.Website != "" {
		return c.NoContent(204)
	}
	if uuid.Validate(in.SessionID) != nil || (in.Event != "view" && in.Event != "start" && in.Event != "step") {
		return bad("Invalid event")
	}
	err := h.service.DB.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
		f, err := h.publicForm(c, tx, true)
		if err != nil {
			return err
		}
		if err = formVersion(f, in.Version, true); err != nil {
			return err
		}
		d, err := marketing.DefinitionForForm(f)
		if err != nil {
			return err
		}
		if in.Event == "step" {
			found := false
			for _, p := range d.Pages {
				if p.ID == in.PageID {
					found = true
				}
			}
			if !found {
				return bad("Unknown page")
			}
		} else if in.PageID != "" {
			return bad("Only step events accept a page")
		}
		return recordFormEvent(tx, f, in.SessionID, in.Event, in.PageID)
	})
	if err != nil {
		return err
	}
	return c.NoContent(204)
}

type formStepAnalytics struct {
	PageID      string `json:"pageId"`
	Title       string `json:"title"`
	Views       int64  `json:"views"`
	Completions int64  `json:"completions"`
}
type formSourceAnalytics struct {
	Source      string `json:"source"`
	Completions int64  `json:"completions"`
}

func (h *MarketingHandler) FormAnalytics(c echo.Context) error {
	var f models.Form
	if err := h.scoped(c).Preload("Fields").First(&f, "id = ?", c.Param("id")).Error; err != nil {
		return missing()
	}
	db := h.service.DB.WithContext(c.Request().Context())
	var views, starts, completions, partials int64
	for event, dest := range map[string]*int64{"view": &views, "start": &starts, "complete": &completions} {
		if err := db.Model(&models.FormAnalyticsEvent{}).Where("form_id = ? AND event = ?", f.ID, event).Distinct("session_id").Count(dest).Error; err != nil {
			return err
		}
	}
	if err := db.Model(&models.FormProgress{}).Where("form_id = ? AND expires_at > ? AND version = ?", f.ID, time.Now().UTC(), f.Version).Distinct("session_id").Count(&partials).Error; err != nil {
		return err
	}
	d, err := marketing.DefinitionForForm(f)
	if err != nil {
		return err
	}
	steps := []formStepAnalytics{}
	for _, p := range d.Pages {
		s := formStepAnalytics{PageID: p.ID, Title: p.Title}
		if err := db.Model(&models.FormAnalyticsEvent{}).Where("form_id = ? AND version = ? AND page_id = ? AND event = ?", f.ID, f.Version, p.ID, "step").Distinct("session_id").Count(&s.Views).Error; err != nil {
			return err
		}
		if err := db.Model(&models.FormAnalyticsEvent{}).Where("form_id = ? AND version = ? AND page_id = ? AND event = ?", f.ID, f.Version, p.ID, "step_complete").Distinct("session_id").Count(&s.Completions).Error; err != nil {
			return err
		}
		steps = append(steps, s)
	}
	sources := []formSourceAnalytics{}
	if err := db.Model(&models.FormSubmission{}).Select("COALESCE(NULLIF(utm_source, ''), 'Direct / unknown') AS source, COUNT(DISTINCT session_id) AS completions").Where("form_id = ?", f.ID).Group("COALESCE(NULLIF(utm_source, ''), 'Direct / unknown')").Order("completions DESC").Limit(50).Scan(&sources).Error; err != nil {
		return err
	}
	rate := float64(0)
	if views > 0 {
		rate = float64(completions) / float64(views)
	}
	return c.JSON(200, map[string]interface{}{"views": views, "starts": starts, "completions": completions, "completionRate": rate, "partials": partials, "steps": steps, "sources": sources})
}
