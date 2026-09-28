package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/events"
	"kori/internal/marketing"
	"kori/internal/models"
)

type formAttribution struct {
	UTMSource   string `json:"utmSource,omitempty"`
	UTMMedium   string `json:"utmMedium,omitempty"`
	UTMCampaign string `json:"utmCampaign,omitempty"`
	UTMContent  string `json:"utmContent,omitempty"`
	UTMTerm     string `json:"utmTerm,omitempty"`
	Referrer    string `json:"referrer,omitempty"`
}

// Preserve known campaign tags when a resume visit supplies only a referrer.
func mergeFormAttribution(current, saved formAttribution) formAttribution {
	if current.UTMSource == "" {
		current.UTMSource = saved.UTMSource
	}
	if current.UTMMedium == "" {
		current.UTMMedium = saved.UTMMedium
	}
	if current.UTMCampaign == "" {
		current.UTMCampaign = saved.UTMCampaign
	}
	if current.UTMContent == "" {
		current.UTMContent = saved.UTMContent
	}
	if current.UTMTerm == "" {
		current.UTMTerm = saved.UTMTerm
	}
	if current.Referrer == "" {
		current.Referrer = saved.Referrer
	}
	return current
}

func (a formAttribution) validate() error {
	for _, v := range []string{a.UTMSource, a.UTMMedium, a.UTMCampaign, a.UTMContent, a.UTMTerm} {
		if len(v) > 200 || strings.ContainsAny(v, "\r\n\x00") {
			return bad("Invalid campaign attribution")
		}
	}
	if len(a.Referrer) > 2048 || strings.ContainsAny(a.Referrer, "\r\n\x00") {
		return bad("Invalid referrer")
	}
	return nil
}
func formDigest(value string) string {
	s := sha256.Sum256([]byte(value))
	return hex.EncodeToString(s[:])
}
func formVersion(f models.Form, version int, required bool) error {
	if version != f.Version && (required || version != 0) {
		return echo.NewHTTPError(409, "This form changed. Reload it before continuing.")
	}
	return nil
}
func formSession(in formSubmissionInput) (string, error) {
	if uuid.Validate(in.RequestID) != nil {
		return "", bad("A valid submission identifier is required")
	}
	session := in.SessionID
	if session == "" {
		session = in.RequestID
	}
	if uuid.Validate(session) != nil {
		return "", bad("A valid session identifier is required")
	}
	return session, nil
}
func (h *MarketingHandler) submitForm(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	in, err := bindFormSubmission(c)
	if err != nil {
		return bad("Invalid submission")
	}
	if in.Website != "" {
		return formSubmissionSuccess(c, "Thanks for your response!")
	}
	session, err := formSession(in)
	if err != nil {
		return err
	}
	if err = in.Attribution.validate(); err != nil {
		return err
	}
	var f models.Form
	var contact models.Contact
	created := false
	submissionID := ""
	redirect := ""
	err = h.service.DB.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
		var err error
		// Look up a committed receipt before checking the live revision. A lost
		// response must remain retryable after an owner edits or archives a form.
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Fields").Where("slug = ? AND is_deleted = ?", c.Param("slug"), false).First(&f).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return missing()
		}
		if err != nil {
			return err
		}
		var previousReceipt models.FormReceipt
		err = tx.First(&previousReceipt, "id = ? AND form_id = ?", formDigest(f.ID+":"+in.RequestID), f.ID).Error
		if err == nil {
			submissionID = previousReceipt.SubmissionID
			// Pre-revision receipts may not have a submission ID. They still
			// prevent duplicate writes, without exposing any submitted answers.
			if submissionID == "" {
				return nil
			}
			var previous models.FormSubmission
			if err := tx.First(&previous, "id = ? AND form_id = ?", submissionID, f.ID).Error; err != nil {
				return err
			}
			if previous.SessionID != session || in.Version != 0 && in.Version != previous.Version || (in.Fields["email"] != "" && strings.ToLower(strings.TrimSpace(in.Fields["email"])) != previous.EmailAddress) {
				return echo.NewHTTPError(409, "This submission identifier was already used for a different response.")
			}
			var revision models.FormRevision
			if err := tx.First(&revision, "form_id = ? AND version = ?", f.ID, previous.Version).Error; err == nil {
				var original marketing.FormDefinition
				var oldData map[string]string
				if err := json.Unmarshal(revision.Definition, &original); err != nil {
					return err
				}
				if err := json.Unmarshal(previous.FieldData, &oldData); err != nil {
					return err
				}
				data, _, answerErr := marketing.FormAnswers(original, in.Fields, false)
				if answerErr != nil || !reflect.DeepEqual(data, oldData) || previous.Consent != (in.Consent && original.Consent.Mode != "none") {
					return echo.NewHTTPError(409, "This submission identifier was already used for a different response.")
				}
				redirect = original.SuccessRedirectURL
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if f.Status != models.FormStatusPublished {
			return missing()
		}
		d, err := marketing.DefinitionForForm(f)
		if err != nil {
			return err
		}
		if err = formVersion(f, in.Version, !legacyEditableDefinition(d)); err != nil {
			return err
		}
		redirect = d.SuccessRedirectURL
		if d.Consent.Mode == "required" && !in.Consent {
			return bad("Consent is required")
		}
		consent := in.Consent && d.Consent.Mode != "none"
		data, pages, err := marketing.FormAnswers(d, in.Fields, false)
		if err != nil {
			return bad(err.Error())
		}
		if !marketing.ValidEmail(data["email"]) {
			return bad("Enter a valid email address")
		}
		if f.AddToListID == nil {
			return bad("This form is unavailable")
		}
		var audience models.MailingList
		if err = marketing.Scope(tx, f.TeamID).Clauses(clause.Locking{Strength: "UPDATE"}).First(&audience, "id = ?", *f.AddToListID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return bad("This form is unavailable")
			}
			return err
		}
		receipt := models.FormReceipt{ID: formDigest(f.ID + ":" + in.RequestID), FormID: f.ID, SubmissionID: uuid.NewString()}
		r := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&receipt)
		if r.Error != nil {
			return r.Error
		}
		if r.RowsAffected == 0 {
			var old models.FormReceipt
			if err := tx.First(&old, "id = ?", receipt.ID).Error; err != nil {
				return err
			}
			submissionID = old.SubmissionID
			return nil
		}
		var progress *models.FormProgress
		if in.ResumeToken != "" {
			p, err := validFormProgress(tx, f, in.ResumeToken)
			if err != nil {
				return err
			}
			if p.SessionID != session {
				return missing()
			}
			progress = &p
		}
		var contactID *string
		if consent {
			var suppressed int64
			if err := tx.Model(&models.SuppressionList{}).Where("team_id = ? AND email_address = ? AND is_active = ? AND is_deleted = ? AND (expires_at IS NULL OR expires_at > ?)", f.TeamID, data["email"], true, false, time.Now().UTC()).Count(&suppressed).Error; err != nil {
				return err
			}
			err := marketing.Scope(tx, f.TeamID).Where("email = ? AND list_id = ?", data["email"], audience.ID).First(&contact).Error
			if errors.Is(err, gorm.ErrRecordNotFound) && suppressed == 0 {
				created = true
				now := time.Now().UTC()
				contact = models.Contact{Base: models.Base{ID: uuid.NewString(), CreatedAt: now, UpdatedAt: now}, TeamID: f.TeamID, ListID: audience.ID, Email: data["email"], FirstName: data["first_name"], LastName: data["last_name"], Company: data["company"], Phone: data["phone"], Status: models.SubscriberStatusActive, LifecycleStage: "LEAD", Metadata: datatypes.JSON(`{"source":"form","consent":true}`)}
				if err = tx.Session(&gorm.Session{SkipHooks: true}).Create(&contact).Error; err != nil {
					return err
				}
			} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			// Suppression is never reversed by submitting a form; suppressed recipients are not eligible for a journey.
			if contact.ID != "" && contact.Status == models.SubscriberStatusActive && suppressed == 0 {
				contactID = &contact.ID
			}
		}
		encoded, _ := json.Marshal(data)
		a := in.Attribution
		if progress != nil {
			var saved formAttribution
			if err := json.Unmarshal(progress.Attribution, &saved); err != nil {
				return err
			}
			a = mergeFormAttribution(a, saved)
		}
		sub := models.FormSubmission{Base: models.Base{ID: receipt.SubmissionID}, FormID: f.ID, ContactID: contactID, EmailAddress: data["email"], FieldData: encoded, Version: f.Version, SessionID: session, Consent: consent, SubmittedAt: time.Now().UTC(), UTMSource: &a.UTMSource, UTMMedium: &a.UTMMedium, UTMCampaign: &a.UTMCampaign, UTMContent: &a.UTMContent, UTMTerm: &a.UTMTerm, Referrer: &a.Referrer}
		if err := tx.Create(&sub).Error; err != nil {
			return err
		}
		submissionID = sub.ID
		eventContact := ""
		if contactID != nil {
			eventContact = *contactID
		}
		payload, _ := json.Marshal(map[string]interface{}{"formId": f.ID, "submissionId": sub.ID, "version": f.Version, "fields": data, "consent": consent, "sessionId": session})
		if err := tx.Create(&models.FormCompletionEvent{ID: uuid.NewString(), TeamID: f.TeamID, FormID: f.ID, SubmissionID: sub.ID, ContactID: eventContact, Payload: payload}).Error; err != nil {
			return err
		}
		if err := recordFormEvent(tx, f, session, "complete", ""); err != nil {
			return err
		}
		for _, page := range pages {
			if err := recordFormEvent(tx, f, session, "step_complete", page); err != nil {
				return err
			}
		}
		// A completed session has no partial responses, including older saved
		// copies created before a retry or a return visit.
		if err := tx.Where("form_id = ? AND session_id = ?", f.ID, session).Delete(&models.FormProgress{}).Error; err != nil {
			return err
		}
		if created {
			if err := models.SyncSubscribersCountByID(tx, audience.ID); err != nil {
				return err
			}
		}
		return tx.Model(&f).UpdateColumn("submission_count", gorm.Expr("submission_count + 1")).Error
	})
	if err != nil {
		return err
	}
	if created {
		events.Emit("contact.created", &contact)
	}
	if isHTMLFormPost(c) {
		if redirect != "" {
			return c.Redirect(303, redirect)
		}
		return formSubmissionSuccess(c, f.SuccessMessage)
	}
	return c.JSON(200, map[string]interface{}{"ok": true, "message": f.SuccessMessage, "submissionId": submissionID, "successRedirectUrl": redirect})
}
