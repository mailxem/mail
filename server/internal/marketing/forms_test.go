package marketing_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"kori/internal/handlers"
	"kori/internal/models"
)

func TestFormBrandingAndHTMLAction(t *testing.T) {
	db := testDB(t)
	team, list := uuid.NewString(), uuid.NewString()
	seed(t, db, &models.MailingList{Base: models.Base{ID: list}, TeamID: team, Name: "Audience"})
	h := handlers.NewMarketingHandler(db, strings.Repeat("s", 32), "https://api.example.com")
	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error { c.Set("teamID", team); return next(c) }
	})
	e.POST("/forms", h.SaveForm)
	e.PUT("/forms/:id", h.SaveForm)
	e.GET("/forms", h.Forms)
	e.GET("/public/:slug", h.PublicForm)
	e.POST("/public/:slug", h.SubmitForm)
	call := func(method, path, contentType, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", contentType)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, r)
		return rec
	}
	input := map[string]interface{}{"name": "Custom signup", "listId": list, "status": "PUBLISHED", "successMessage": "Welcome <script>alert(1)</script>", "fields": []map[string]interface{}{{"label": "Email", "type": "EMAIL", "key": "email", "required": true}}, "theme": map[string]string{"preset": "dark", "backgroundColor": "#15151c", "cardColor": "#242430", "textColor": "#f4f2ff", "buttonColor": "#c4b5fd", "buttonTextColor": "#211736", "logoUrl": "https://example.com/logo.png", "font": "serif", "corners": "square"}}
	raw, _ := json.Marshal(input)
	rec := call("POST", "/forms", echo.MIMEApplicationJSON, string(raw))
	require.Equal(t, 200, rec.Code, rec.Body.String())
	var form models.Form
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &form))
	rec = call("GET", "/public/"+form.Slug, "", "")
	require.Equal(t, 200, rec.Code)
	var public map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &public))
	require.Equal(t, "#15151c", public["theme"].(map[string]interface{})["backgroundColor"])
	require.NotContains(t, public, "TeamID")
	// Older clients omitting theme must preserve customization.
	delete(input, "theme")
	raw, _ = json.Marshal(input)
	rec = call("PUT", "/forms/"+form.ID, echo.MIMEApplicationJSON, string(raw))
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Contains(t, rec.Body.String(), "https://example.com/logo.png")
	// Server rejects script URLs and CSS payloads.
	for _, theme := range []map[string]string{{"logoUrl": "javascript:alert(1)"}, {"buttonColor": "red;display:none"}, {"font": "url(evil)"}} {
		input["theme"] = theme
		raw, _ = json.Marshal(input)
		rec = call("PUT", "/forms/"+form.ID, echo.MIMEApplicationJSON, string(raw))
		require.Equal(t, 400, rec.Code)
	}
	path := "/public/" + form.Slug
	for _, payload := range []url.Values{{"email": {"reader@example.com"}}, {"email": {"invalid"}, "consent": {"true"}}, {"email": {""}, "consent": {"on"}}} {
		rec = call("POST", path, echo.MIMEApplicationForm, payload.Encode())
		require.Equal(t, 400, rec.Code, rec.Body.String())
		require.Contains(t, rec.Header().Get("Content-Type"), "text/html")
	}
	// Query values must not supply consent.
	rec = call("POST", path+"?consent=true", echo.MIMEApplicationForm, "email=reader%40example.com")
	require.Equal(t, 400, rec.Code)
	payload := url.Values{"email": {"reader@example.com"}, "consent": {"on"}, "requestId": {uuid.NewString()}}
	for i := 0; i < 2; i++ {
		rec = call("POST", path, echo.MIMEApplicationForm, payload.Encode())
		require.Equal(t, 200, rec.Code, rec.Body.String())
		require.Contains(t, rec.Body.String(), "&lt;script&gt;")
		require.NotContains(t, rec.Body.String(), "<script>")
	}
	var count int64
	require.NoError(t, db.Model(&models.FormSubmission{}).Count(&count).Error)
	require.EqualValues(t, 1, count)
	// Plain HTML needs no generated identifier, multipart works too.
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.WriteField("email", "second@example.com"))
	require.NoError(t, writer.WriteField("consent", "true"))
	require.NoError(t, writer.Close())
	rec = call("POST", path, writer.FormDataContentType(), body.String())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	var contact models.Contact
	require.NoError(t, db.Where("email = ?", "second@example.com").First(&contact).Error)
	require.Equal(t, list, contact.ListID)
	require.Equal(t, team, contact.TeamID)
	rec = call("POST", path, echo.MIMEApplicationForm, "email=bot%40example.com&consent=on&website=spam")
	require.Equal(t, 200, rec.Code)
	require.NoError(t, db.Model(&models.FormSubmission{}).Count(&count).Error)
	require.EqualValues(t, 2, count)
	require.NoError(t, db.Model(&form).Update("status", "DRAFT").Error)
	rec = call("POST", path, echo.MIMEApplicationForm, "email=third%40example.com&consent=on")
	require.Equal(t, 404, rec.Code)
}

func TestFormReceiptSurvivesRevisionAndArchiveWithoutDuplicateCapture(t *testing.T) {
	h := newFormHarness(t)
	f := h.create("optional")
	path := "/public/" + f.Slug
	input := map[string]any{"fields": map[string]string{"email": "reader@example.com", "interest": "personal"}, "version": f.Version, "requestId": uuid.NewString(), "sessionId": uuid.NewString()}
	first := h.call("POST", path, input)
	require.Equal(t, 200, first.Code, first.Body.String())
	var original map[string]any
	require.NoError(t, json.Unmarshal(first.Body.Bytes(), &original))
	updated := h.definition("required")
	updated["successRedirectUrl"] = "https://example.com/new-thanks"
	r := h.call("PUT", "/forms/"+f.ID, map[string]any{"name": "Updated signup", "listId": h.list, "status": "PUBLISHED", "definition": updated})
	require.Equal(t, 200, r.Code, r.Body.String())
	for _, status := range []string{"PUBLISHED", "ARCHIVED"} {
		require.NoError(t, h.db.Model(&f).Update("status", status).Error)
		r = h.call("POST", path, input)
		require.Equal(t, 200, r.Code, r.Body.String())
		var replay map[string]any
		require.NoError(t, json.Unmarshal(r.Body.Bytes(), &replay))
		require.Equal(t, original["submissionId"], replay["submissionId"])
		require.Equal(t, "", replay["successRedirectUrl"])
	}
	input["fields"] = map[string]string{"email": "different@example.com", "interest": "personal"}
	r = h.call("POST", path, input)
	require.Equal(t, 409, r.Code, r.Body.String())
	input["requestId"] = uuid.NewString()
	r = h.call("POST", path, input)
	require.Equal(t, 404, r.Code)
	var count int64
	require.NoError(t, h.db.Model(&models.FormSubmission{}).Count(&count).Error)
	require.EqualValues(t, 1, count)
	require.NoError(t, h.db.Model(&models.FormCompletionEvent{}).Count(&count).Error)
	require.EqualValues(t, 1, count)
}

func TestRichFormRequiresCurrentVersionAndDoesNotEnrollUnsubscribedContacts(t *testing.T) {
	h := newFormHarness(t)
	f := h.create("required")
	input := map[string]any{"fields": map[string]string{"email": "reader@example.com", "interest": "personal"}, "consent": true, "requestId": uuid.NewString(), "sessionId": uuid.NewString()}
	r := h.call("POST", "/public/"+f.Slug, input)
	require.Equal(t, 409, r.Code, r.Body.String())
	seed(t, h.db, &models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: h.team, ListID: h.list, Email: "reader@example.com", Status: models.SubscriberStatusUnsubscribed})
	input["version"] = f.Version
	r = h.call("POST", "/public/"+f.Slug, input)
	require.Equal(t, 200, r.Code, r.Body.String())
	var submission models.FormSubmission
	require.NoError(t, h.db.First(&submission).Error)
	require.Nil(t, submission.ContactID)
	var completion models.FormCompletionEvent
	require.NoError(t, h.db.First(&completion).Error)
	require.Empty(t, completion.ContactID)
	var contact models.Contact
	require.NoError(t, h.db.First(&contact).Error)
	require.Equal(t, models.SubscriberStatusUnsubscribed, contact.Status)
}

func TestFormProgressTokenDoesNotReviveAfterRetentionAndRequiresStrongSecret(t *testing.T) {
	h := newFormHarness(t)
	f := h.create("optional")
	path := "/public/" + f.Slug
	input := map[string]any{"fields": map[string]string{"email": "reader@example.com"}, "version": f.Version, "requestId": uuid.NewString(), "sessionId": uuid.NewString()}
	readToken := func(r *httptest.ResponseRecorder) string {
		t.Helper()
		require.Equal(t, 200, r.Code, r.Body.String())
		var saved map[string]any
		require.NoError(t, json.Unmarshal(r.Body.Bytes(), &saved))
		return saved["resumeToken"].(string)
	}
	first := readToken(h.call("POST", path+"/progress", input))
	require.Equal(t, first, readToken(h.call("POST", path+"/progress", input)))
	require.NoError(t, h.db.Where("form_id = ?", f.ID).Delete(&models.FormProgress{}).Error)
	second := readToken(h.call("POST", path+"/progress", input))
	require.NotEqual(t, first, second)
	require.Equal(t, 404, h.call("POST", path+"/resume", map[string]string{"token": first}).Code)
	require.Equal(t, 200, h.call("POST", path+"/resume", map[string]string{"token": second}).Code)
	weak := handlers.NewMarketingHandler(h.db, "weak", "https://api.example.com")
	h.app.POST("/weak/:slug/progress", weak.SaveFormProgress)
	h.app.POST("/weak/:slug/resume", weak.ResumeForm)
	require.Equal(t, 503, h.call("POST", "/weak/"+f.Slug+"/progress", input).Code)
	require.Equal(t, 503, h.call("POST", "/weak/"+f.Slug+"/resume", map[string]string{"token": second}).Code)
}

func TestFormCompletionClearsAllSavedCopiesOnlyForItsSession(t *testing.T) {
	h := newFormHarness(t)
	f := h.create("optional")
	path := "/public/" + f.Slug
	session, otherSession := uuid.NewString(), uuid.NewString()
	for _, id := range []string{session, session, otherSession} {
		r := h.call("POST", path+"/progress", map[string]any{"fields": map[string]string{"email": "reader@example.com"}, "version": f.Version, "requestId": uuid.NewString(), "sessionId": id})
		require.Equal(t, 200, r.Code, r.Body.String())
	}
	r := h.call("POST", path, map[string]any{"fields": map[string]string{"email": "reader@example.com", "interest": "personal"}, "version": f.Version, "requestId": uuid.NewString(), "sessionId": session})
	require.Equal(t, 200, r.Code, r.Body.String())
	var progress []models.FormProgress
	require.NoError(t, h.db.Find(&progress).Error)
	require.Len(t, progress, 1)
	require.Equal(t, otherSession, progress[0].SessionID)
}

func TestFormAnalyticsSeparatesRevisedStepsAndDeduplicatesLifetimeVisitors(t *testing.T) {
	h := newFormHarness(t)
	f := h.create("optional")
	path := "/public/" + f.Slug
	session := uuid.NewString()
	track := func(version int) {
		for _, event := range []string{"view", "start", "step"} {
			input := map[string]any{"sessionId": session, "version": version, "event": event}
			if event == "step" {
				input["pageId"] = "details"
			}
			r := h.call("POST", path+"/events", input)
			require.Equal(t, 204, r.Code, r.Body.String())
		}
	}
	track(f.Version)
	r := h.call("PUT", "/forms/"+f.ID, map[string]any{"name": "Updated signup", "listId": h.list, "status": "PUBLISHED", "definition": h.definition("optional")})
	require.Equal(t, 200, r.Code, r.Body.String())
	var updated models.Form
	require.NoError(t, json.Unmarshal(r.Body.Bytes(), &updated))
	r = h.call("GET", "/forms/"+f.ID+"/analytics", nil)
	var metrics struct {
		Views, Starts int64
		Steps         []struct{ Views int64 }
	}
	require.NoError(t, json.Unmarshal(r.Body.Bytes(), &metrics))
	require.EqualValues(t, 1, metrics.Views)
	require.Zero(t, metrics.Steps[0].Views)
	track(updated.Version)
	r = h.call("GET", "/forms/"+f.ID+"/analytics", nil)
	require.NoError(t, json.Unmarshal(r.Body.Bytes(), &metrics))
	require.EqualValues(t, 1, metrics.Views)
	require.EqualValues(t, 1, metrics.Starts)
	require.EqualValues(t, 1, metrics.Steps[0].Views)
}

func TestFormResumeKeepsCampaignAttributionWhenReferrerChanges(t *testing.T) {
	h := newFormHarness(t)
	f := h.create("optional")
	path := "/public/" + f.Slug
	input := map[string]any{"fields": map[string]string{"email": "reader@example.com"}, "version": f.Version, "requestId": uuid.NewString(), "sessionId": uuid.NewString(), "attribution": map[string]string{"utmSource": "newsletter"}}
	r := h.call("POST", path+"/progress", input)
	require.Equal(t, 200, r.Code, r.Body.String())
	var saved map[string]any
	require.NoError(t, json.Unmarshal(r.Body.Bytes(), &saved))
	input["resumeToken"] = saved["resumeToken"]
	input["requestId"] = uuid.NewString()
	input["attribution"] = map[string]string{"referrer": "https://example.com/return"}
	r = h.call("POST", path+"/progress", input)
	require.Equal(t, 200, r.Code, r.Body.String())
	input["requestId"] = uuid.NewString()
	input["fields"] = map[string]string{"email": "reader@example.com", "interest": "personal"}
	input["attribution"] = map[string]string{"referrer": "https://example.com/return"}
	r = h.call("POST", path, input)
	require.Equal(t, 200, r.Code, r.Body.String())
	var submission models.FormSubmission
	require.NoError(t, h.db.First(&submission).Error)
	require.Equal(t, "newsletter", *submission.UTMSource)
	require.Equal(t, "https://example.com/return", *submission.Referrer)
}
