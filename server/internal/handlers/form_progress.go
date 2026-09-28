package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/marketing"
	"kori/internal/models"
)

const formProgressLifetime = 7 * 24 * time.Hour

func validFormProgress(tx *gorm.DB, f models.Form, token string) (models.FormProgress, error) {
	var p models.FormProgress
	if len(token) != 43 {
		return p, missing()
	}
	err := tx.Where("form_id = ? AND token_hash = ? AND expires_at > ?", f.ID, formDigest(token), time.Now().UTC()).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return p, missing()
	}
	if err != nil {
		return p, err
	}
	if !p.ExpiresAt.After(time.Now().UTC()) {
		return p, missing()
	}
	if err = formVersion(f, p.Version, true); err != nil {
		return p, err
	}
	return p, nil
}
func (h *MarketingHandler) SaveFormProgress(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	if len(h.service.Secret) < 32 {
		return echo.NewHTTPError(503, "Saved forms are temporarily unavailable.")
	}
	var in formSubmissionInput
	if err := c.Bind(&in); err != nil {
		return bad("Invalid progress")
	}
	if in.Website != "" {
		return bad("Unable to save progress")
	}
	session, err := formSession(in)
	if err != nil {
		return err
	}
	if in.SessionID == "" {
		return bad("A session identifier is required")
	}
	if err = in.Attribution.validate(); err != nil {
		return err
	}
	var token string
	var expires time.Time
	err = h.service.DB.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
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
		if !d.SaveProgress {
			return bad("Saving progress is disabled for this form")
		}
		data, pages, err := marketing.FormAnswers(d, in.Fields, true)
		if err != nil {
			return bad(err.Error())
		}
		if in.PageID != "" {
			found := false
			for _, page := range pages {
				if page == in.PageID {
					found = true
				}
			}
			if !found {
				return bad("Choose a reachable page")
			}
		}
		now := time.Now().UTC()
		p := models.FormProgress{}
		token = in.ResumeToken
		if token != "" {
			p, err = validFormProgress(tx, f, token)
			if err != nil {
				return err
			}
			if p.SessionID != session {
				return missing()
			}
		} else {
			id := formDigest(f.ID + ":" + in.RequestID)
			err := tx.First(&p, "id = ?", id).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				p = models.FormProgress{ID: id, FormID: f.ID, Version: f.Version, SessionID: session, ExpiresAt: now.Add(formProgressLifetime).Truncate(time.Microsecond)}
			} else if err != nil {
				return err
			} else if p.SessionID != session || p.Version != f.Version || !p.ExpiresAt.After(now) {
				return echo.NewHTTPError(409, "This saved session expired or changed. Start a new session.")
			}
			// Bind the retryable token to this saved record's lifetime. Reusing a
			// request ID after retention cleanup must never revive an old link.
			mac := hmac.New(sha256.New, []byte(h.service.Secret))
			mac.Write([]byte("form-progress:v2:" + f.ID + ":" + in.RequestID + ":" + session + ":" + strconv.Itoa(f.Version) + ":" + strconv.FormatInt(p.ExpiresAt.UnixMicro(), 10)))
			token = base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
			if p.TokenHash != "" && p.TokenHash != formDigest(token) {
				return echo.NewHTTPError(409, "This saved session changed. Start a new session.")
			}
			p.TokenHash = formDigest(token)
		}
		p.Fields, _ = json.Marshal(data)
		var saved formAttribution
		if len(p.Attribution) > 0 {
			if err := json.Unmarshal(p.Attribution, &saved); err != nil {
				return err
			}
		}
		p.Attribution, _ = json.Marshal(mergeFormAttribution(in.Attribution, saved))
		p.PageID = in.PageID
		if err := tx.Save(&p).Error; err != nil {
			return err
		}
		expires = p.ExpiresAt
		// Opportunistic retention: remove expired personal data whenever this form is used.
		return tx.Where("form_id = ? AND expires_at <= ?", f.ID, now).Delete(&models.FormProgress{}).Error
	})
	if err != nil {
		return err
	}
	return c.JSON(200, map[string]interface{}{"resumeToken": token, "expiresAt": expires})
}
func (h *MarketingHandler) ResumeForm(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	if len(h.service.Secret) < 32 {
		return echo.NewHTTPError(503, "Saved forms are temporarily unavailable.")
	}
	var in struct {
		Token string `json:"token"`
	}
	if err := c.Bind(&in); err != nil {
		return bad("Invalid resume token")
	}
	var p models.FormProgress
	err := h.service.DB.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
		f, err := h.publicForm(c, tx, true)
		if err != nil {
			return err
		}
		d, err := marketing.DefinitionForForm(f)
		if err != nil {
			return err
		}
		if !d.SaveProgress {
			return missing()
		}
		p, err = validFormProgress(tx.Clauses(clause.Locking{Strength: "UPDATE"}), f, in.Token)
		return err
	})
	if err != nil {
		return err
	}
	return c.JSON(200, map[string]interface{}{"fields": p.Fields, "version": p.Version, "pageId": p.PageID, "sessionId": p.SessionID})
}
