package marketing_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"kori/internal/ai"
	"kori/internal/handlers"
	"kori/internal/marketing"
	"kori/internal/models"
	"kori/internal/services"
)

type formTestHarness struct {
	t    *testing.T
	db   *gorm.DB
	app  *echo.Echo
	team string
	list string
}

func newFormHarness(t *testing.T) *formTestHarness {
	db := testDB(t)
	require.NoError(t, db.AutoMigrate(&models.Automation{}, &models.AutomationNode{}, &models.AutomationExecution{}))
	h := &formTestHarness{t: t, db: db, app: echo.New(), team: uuid.NewString(), list: uuid.NewString()}
	seed(t, db, &models.MailingList{Base: models.Base{ID: h.list}, TeamID: h.team, Name: "Audience"})
	handler := handlers.NewMarketingHandler(db, strings.Repeat("s", 32), "https://api.example.com")
	h.app.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			team := c.Request().Header.Get("X-Test-Team")
			if team == "" {
				team = h.team
			}
			c.Set("teamID", team)
			return next(c)
		}
	})
	h.app.POST("/forms", handler.SaveForm)
	h.app.PUT("/forms/:id", handler.SaveForm)
	h.app.POST("/forms/:id/journey", handler.CreateFormJourney)
	h.app.GET("/forms/:id/analytics", handler.FormAnalytics)
	h.app.GET("/public/:slug", handler.PublicForm)
	h.app.POST("/public/:slug", handler.SubmitForm)
	h.app.POST("/public/:slug/progress", handler.SaveFormProgress)
	h.app.POST("/public/:slug/resume", handler.ResumeForm)
	h.app.POST("/public/:slug/events", handler.RecordFormEvent)
	return h
}
func (h *formTestHarness) call(method, path string, body any) *httptest.ResponseRecorder {
	h.t.Helper()
	raw, err := json.Marshal(body)
	require.NoError(h.t, err)
	req := httptest.NewRequest(method, path, strings.NewReader(string(raw)))
	req.Header.Set("Content-Type", echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	h.app.ServeHTTP(rec, req)
	return rec
}
func (h *formTestHarness) definition(consent string) map[string]any {
	return map[string]any{"schemaVersion": 1, "saveProgress": true, "consent": map[string]any{"mode": consent, "label": "I agree to receive updates"}, "pages": []any{
		map[string]any{"id": "details", "title": "Details", "fields": []any{map[string]any{"key": "email", "label": "Email", "type": "EMAIL", "required": true}, map[string]any{"key": "interest", "label": "Interest", "type": "SELECT", "required": true, "options": []string{"business", "personal"}}, map[string]any{"key": "source", "label": "Source", "type": "HIDDEN", "defaultValue": "official"}}},
		map[string]any{"id": "company", "title": "Company", "condition": map[string]any{"field": "interest", "operator": "equals", "value": "business"}, "fields": []any{map[string]any{"key": "company", "label": "Company", "type": "TEXT", "required": true}}},
	}}
}
func (h *formTestHarness) create(consent string) models.Form {
	h.t.Helper()
	rec := h.call("POST", "/forms", map[string]any{"name": "Journey form", "listId": h.list, "status": "PUBLISHED", "definition": h.definition(consent)})
	require.Equal(h.t, 200, rec.Code, rec.Body.String())
	var f models.Form
	require.NoError(h.t, json.Unmarshal(rec.Body.Bytes(), &f))
	return f
}
func TestFormJourneyConsentIdempotencyAndSuppression(t *testing.T) {
	h := newFormHarness(t)
	f := h.create("optional")
	path := "/public/" + f.Slug
	input := map[string]any{"fields": map[string]string{"email": "reader@example.com", "interest": "personal", "company": "forged", "source": "forged"}, "consent": false, "version": f.Version, "requestId": uuid.NewString(), "sessionId": uuid.NewString()}
	for i := 0; i < 2; i++ {
		r := h.call("POST", path, input)
		require.Equal(t, 200, r.Code, r.Body.String())
	}
	var count int64
	require.NoError(t, h.db.Model(&models.Contact{}).Count(&count).Error)
	require.Zero(t, count)
	var subs []models.FormSubmission
	require.NoError(t, h.db.Find(&subs).Error)
	require.Len(t, subs, 1)
	require.Nil(t, subs[0].ContactID)
	require.False(t, subs[0].Consent)
	require.NotContains(t, string(subs[0].FieldData), "forged")
	var outbox []models.FormCompletionEvent
	require.NoError(t, h.db.Find(&outbox).Error)
	require.Len(t, outbox, 1)
	require.Empty(t, outbox[0].ContactID)
	input["consent"] = true
	input["requestId"] = uuid.NewString()
	r := h.call("POST", path, input)
	require.Equal(t, 200, r.Code, r.Body.String())
	var contact models.Contact
	require.NoError(t, h.db.First(&contact).Error)
	require.Equal(t, models.SubscriberStatusActive, contact.Status)
	require.NoError(t, h.db.Model(&contact).Update("status", models.SubscriberStatusUnsubscribed).Error)
	input["requestId"] = uuid.NewString()
	r = h.call("POST", path, input)
	require.Equal(t, 200, r.Code, r.Body.String())
	require.NoError(t, h.db.First(&contact, "id = ?", contact.ID).Error)
	require.Equal(t, models.SubscriberStatusUnsubscribed, contact.Status)
	require.NoError(t, h.db.Find(&outbox).Error)
	require.Len(t, outbox, 3) // New submissions by existing contacts create new form events.
	seed(t, h.db, &models.SuppressionList{Base: models.Base{ID: uuid.NewString()}, TeamID: h.team, EmailAddress: "blocked@example.com", IsActive: true, Reason: models.SuppressionReason("MANUAL")})
	input["fields"] = map[string]string{"email": "blocked@example.com", "interest": "personal"}
	input["requestId"] = uuid.NewString()
	r = h.call("POST", path, input)
	require.Equal(t, 200, r.Code, r.Body.String())
	require.NoError(t, h.db.Model(&models.Contact{}).Where("email = ?", "blocked@example.com").Count(&count).Error)
	require.Zero(t, count)
}
func TestFormProgressScopeExpiryVersionAndCompletion(t *testing.T) {
	h := newFormHarness(t)
	f := h.create("optional")
	other := h.create("optional")
	path := "/public/" + f.Slug
	session := uuid.NewString()
	input := map[string]any{"fields": map[string]string{"email": "reader@example.com", "interest": "business"}, "version": f.Version, "requestId": uuid.NewString(), "sessionId": session, "pageId": "company", "attribution": map[string]string{"utmSource": "newsletter"}}
	r := h.call("POST", path+"/progress", input)
	require.Equal(t, 200, r.Code, r.Body.String())
	var saved struct {
		ResumeToken string `json:"resumeToken"`
	}
	require.NoError(t, json.Unmarshal(r.Body.Bytes(), &saved))
	require.Len(t, saved.ResumeToken, 43)
	r = h.call("POST", path+"/progress", input)
	require.Equal(t, 200, r.Code)
	require.Contains(t, r.Body.String(), saved.ResumeToken)
	var progress models.FormProgress
	require.NoError(t, h.db.First(&progress).Error)
	require.NotEqual(t, saved.ResumeToken, progress.TokenHash)
	r = h.call("POST", "/public/"+other.Slug+"/resume", map[string]string{"token": saved.ResumeToken})
	require.Equal(t, 404, r.Code)
	r = h.call("POST", path+"/resume", map[string]string{"token": saved.ResumeToken})
	require.Equal(t, 200, r.Code, r.Body.String())
	require.Contains(t, r.Body.String(), "reader@example.com")
	require.Equal(t, "no-store", r.Header().Get("Cache-Control"))
	require.NoError(t, h.db.Model(&progress).Update("expires_at", time.Now().UTC().Add(-time.Hour)).Error)
	r = h.call("POST", path+"/resume", map[string]string{"token": saved.ResumeToken})
	require.Equal(t, 404, r.Code)
	require.NoError(t, h.db.Model(&progress).Update("expires_at", time.Now().UTC().Add(time.Hour)).Error)
	require.NoError(t, h.db.Model(&f).Update("version", f.Version+1).Error)
	r = h.call("POST", path+"/resume", map[string]string{"token": saved.ResumeToken})
	require.Equal(t, 409, r.Code)
	require.NoError(t, h.db.Model(&f).Update("version", progress.Version).Error)
	input["resumeToken"] = saved.ResumeToken
	input["requestId"] = uuid.NewString()
	input["fields"] = map[string]string{"email": "reader@example.com", "interest": "business", "company": "Acme"}
	r = h.call("POST", path, input)
	require.Equal(t, 200, r.Code, r.Body.String())
	r = h.call("POST", path, input)
	require.Equal(t, 200, r.Code, r.Body.String())
	var count int64
	require.NoError(t, h.db.Model(&models.FormProgress{}).Count(&count).Error)
	require.Zero(t, count)
	var sub models.FormSubmission
	require.NoError(t, h.db.First(&sub).Error)
	require.Equal(t, "newsletter", *sub.UTMSource)
}
func TestFormAnalyticsCountsSessionsAndScopesWorkspace(t *testing.T) {
	h := newFormHarness(t)
	f := h.create("optional")
	path := "/public/" + f.Slug
	session := uuid.NewString()
	for _, event := range []string{"view", "view", "start", "step", "step"} {
		in := map[string]any{"sessionId": session, "event": event, "version": f.Version}
		if event == "step" {
			in["pageId"] = "details"
		}
		r := h.call("POST", path+"/events", in)
		require.Equal(t, 204, r.Code, r.Body.String())
	}
	input := map[string]any{"fields": map[string]string{"email": "reader@example.com", "interest": "personal"}, "version": f.Version, "requestId": uuid.NewString(), "sessionId": session, "attribution": map[string]string{"utmSource": "blog"}}
	for i := 0; i < 2; i++ {
		input["requestId"] = uuid.NewString()
		r := h.call("POST", path, input)
		require.Equal(t, 200, r.Code, r.Body.String())
	}
	r := h.call("GET", "/forms/"+f.ID+"/analytics", nil)
	require.Equal(t, 200, r.Code, r.Body.String())
	var metrics map[string]any
	require.NoError(t, json.Unmarshal(r.Body.Bytes(), &metrics))
	require.EqualValues(t, 1, metrics["views"])
	require.EqualValues(t, 1, metrics["starts"])
	require.EqualValues(t, 1, metrics["completions"])
	require.EqualValues(t, 1, metrics["completionRate"])
	req := httptest.NewRequest("GET", "/forms/"+f.ID+"/analytics", nil)
	req.Header.Set("X-Test-Team", uuid.NewString())
	rec := httptest.NewRecorder()
	h.app.ServeHTTP(rec, req)
	require.Equal(t, 404, rec.Code)
	input["version"] = f.Version + 1
	r = h.call("POST", path, input)
	require.Equal(t, 409, r.Code)
}

// A published form and an active workflow must be explicitly linked in the
// same workspace before the submission happened.
func seedFormWorkflow(t *testing.T, db *gorm.DB, teamID, formID string) models.Automation {
	t.Helper()
	raw, err := json.Marshal(map[string]string{"formId": formID})
	require.NoError(t, err)
	workflow := models.Automation{Base: models.Base{ID: uuid.NewString(), UpdatedAt: time.Now().UTC().Add(-time.Hour)}, TeamID: teamID, Name: "Form follow-up", TriggerEvent: "form.completed", IsActive: true}
	seed(t, db, &workflow)
	node := models.AutomationNode{Base: models.Base{ID: uuid.NewString()}, AutomationID: workflow.ID, Type: models.NodeTypeStart, Data: datatypes.JSON(raw)}
	seed(t, db, &node)
	return workflow
}
func seedFormCompletion(t *testing.T, db *gorm.DB, teamID, formID, contactID string, consent bool) models.FormCompletionEvent {
	t.Helper()
	payload, err := json.Marshal(map[string]any{"consent": consent, "version": 7, "fields": map[string]any{"first_name": "Reader", "id": "untrusted-answer", "version": "untrusted-version", "boolean": true}})
	require.NoError(t, err)
	event := models.FormCompletionEvent{ID: uuid.NewString(), TeamID: teamID, FormID: formID, SubmissionID: uuid.NewString(), ContactID: contactID, Payload: datatypes.JSON(payload)}
	seed(t, db, &event)
	return event
}
func TestFormCompletionDispatchRetriesStableExecutions(t *testing.T) {
	h := newFormHarness(t)
	f := h.create("optional")
	contact := models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: h.team, ListID: h.list, Email: "reader@example.com", Status: models.SubscriberStatusActive}
	seed(t, h.db, &contact)
	first := seedFormWorkflow(t, h.db, h.team, f.ID)
	second := seedFormWorkflow(t, h.db, h.team, f.ID)
	event := seedFormCompletion(t, h.db, h.team, f.ID, contact.ID, true)
	seen := map[string]int{}
	enqueued := map[string]bool{}
	fail := true
	enqueue := func(_ context.Context, executionID, automationID, contactID string, variables map[string]interface{}) error {
		require.Contains(t, []string{first.ID, second.ID}, automationID)
		require.Equal(t, contact.ID, contactID)
		require.Equal(t, f.ID, variables["form_id"])
		require.Equal(t, float64(7), variables["form_version"])
		require.Equal(t, event.SubmissionID, variables["submission_id"])
		require.Equal(t, "Reader", variables["form_first_name"])
		require.NotContains(t, variables, "form_boolean")
		require.Equal(t, uuid.NewSHA1(uuid.NameSpaceOID, []byte("xem:form:"+event.ID+":"+automationID)).String(), executionID)
		seen[executionID]++
		// Simulate a Redis enqueue succeeding but its response getting lost. A retry
		// reaches the same queue identity instead of creating another execution.
		enqueued[executionID] = true
		if fail && len(enqueued) == 2 {
			return errors.New("queue response lost")
		}
		return nil
	}
	require.Error(t, marketing.DispatchFormCompletions(context.Background(), h.db, enqueue, 100))
	var pending models.FormCompletionEvent
	require.NoError(t, h.db.First(&pending, "id = ?", event.ID).Error)
	require.Nil(t, pending.DispatchedAt)
	fail = false
	require.NoError(t, marketing.DispatchFormCompletions(context.Background(), h.db, enqueue, 100))
	require.Len(t, enqueued, 2)
	for _, attempts := range seen {
		require.Equal(t, 2, attempts)
	}
	require.NoError(t, h.db.First(&pending, "id = ?", event.ID).Error)
	require.NotNil(t, pending.DispatchedAt)
	require.NoError(t, marketing.DispatchFormCompletions(context.Background(), h.db, func(context.Context, string, string, string, map[string]interface{}) error {
		t.Fatal("acknowledged event dispatched again")
		return nil
	}, 100))
}
func TestFormCompletionDispatchRequiresCurrentConsentScopeAndEligibility(t *testing.T) {
	for _, scenario := range []string{"no-consent", "missing-contact", "foreign-contact", "unsubscribed", "suppressed-after-submit", "inactive-workflow", "foreign-workflow", "different-form", "newly-activated", "expired-suppression", "foreign-suppression"} {
		t.Run(scenario, func(t *testing.T) {
			h := newFormHarness(t)
			f := h.create("optional")
			contact := models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: h.team, ListID: h.list, Email: "reader@example.com", Status: models.SubscriberStatusActive}
			if scenario == "foreign-contact" {
				contact.TeamID = uuid.NewString()
			}
			if scenario == "unsubscribed" {
				contact.Status = models.SubscriberStatusUnsubscribed
			}
			seed(t, h.db, &contact)
			workflow := seedFormWorkflow(t, h.db, h.team, f.ID)
			if scenario == "inactive-workflow" {
				require.NoError(t, h.db.Model(&workflow).Update("is_active", false).Error)
			}
			if scenario == "foreign-workflow" {
				require.NoError(t, h.db.Model(&workflow).UpdateColumn("team_id", uuid.NewString()).Error)
			}
			if scenario == "different-form" {
				require.NoError(t, h.db.Model(&models.AutomationNode{}).Where("automation_id = ?", workflow.ID).Update("data", `{"formId":"`+uuid.NewString()+`"}`).Error)
			}
			contactID := contact.ID
			if scenario == "missing-contact" {
				contactID = ""
			}
			event := seedFormCompletion(t, h.db, h.team, f.ID, contactID, scenario != "no-consent")
			if scenario == "newly-activated" {
				require.NoError(t, h.db.Model(&workflow).Update("updated_at", event.CreatedAt.Add(time.Second)).Error)
			}
			if scenario == "suppressed-after-submit" || scenario == "expired-suppression" || scenario == "foreign-suppression" {
				suppression := models.SuppressionList{Base: models.Base{ID: uuid.NewString()}, TeamID: h.team, EmailAddress: "READER@example.com", IsActive: true, Reason: models.SuppressionReason("MANUAL")}
				if scenario == "expired-suppression" {
					expired := time.Now().UTC().Add(-time.Minute)
					suppression.ExpiresAt = &expired
				}
				if scenario == "foreign-suppression" {
					suppression.TeamID = uuid.NewString()
				}
				seed(t, h.db, &suppression)
			}
			calls := 0
			require.NoError(t, marketing.DispatchFormCompletions(context.Background(), h.db, func(context.Context, string, string, string, map[string]interface{}) error { calls++; return nil }, 100))
			expected := 0
			if scenario == "expired-suppression" || scenario == "foreign-suppression" {
				expected = 1
			}
			require.Equal(t, expected, calls)
			require.NoError(t, h.db.First(&event, "id = ?", event.ID).Error)
			require.NotNil(t, event.DispatchedAt)
		})
	}
}
func TestFormJourneyCreatesEditableInactiveDraftAndReplaysWithoutDuplicates(t *testing.T) {
	h := newFormHarness(t)
	f := h.create("optional")
	sender := models.SMTPConfig{Base: models.Base{ID: uuid.NewString()}, TeamID: h.team, Host: "smtp.example.com", FromEmail: "hello@example.com"}
	seed(t, h.db, &sender)
	input := map[string]any{"requestId": uuid.NewString(), "name": "Welcome series", "smtpConfigId": sender.ID, "postalAddress": "42 Market Street", "emails": []ai.JourneyEmailDraft{{Name: "Welcome", Subject: "Hello {{first_name}}", Body: "Hello {{first_name}}\n\n<script>alert(1)</script>", DelayHours: 24}, {Name: "A follow-up", Subject: "Your interest", Body: "You chose {{form_interest}}."}}}
	path := "/forms/" + f.ID + "/journey"
	response := h.call("POST", path, input)
	require.Equal(t, 201, response.Code, response.Body.String())
	var workflow models.Automation
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &workflow))
	require.False(t, workflow.IsActive)
	require.Equal(t, "form.completed", workflow.TriggerEvent)
	require.Equal(t, h.team, workflow.TeamID)
	require.Len(t, workflow.Nodes, 5)
	require.Len(t, workflow.Edges, 4)
	require.NoError(t, services.NewAutomationService(h.db).ValidateConfiguration(&workflow))
	var templates []models.Template
	require.NoError(t, h.db.Preload("Category").Find(&templates).Error)
	require.Len(t, templates, 2)
	for _, template := range templates {
		require.Equal(t, h.team, template.TeamID)
		require.Equal(t, "Marketing", template.Category.Name)
		require.NotContains(t, template.HTMLBody, "<script>")
		design, err := base64.StdEncoding.DecodeString(template.DesignJSON)
		require.NoError(t, err)
		require.True(t, json.Valid(design))
	}
	// A lost-response retry returns the saved result after a sender is removed.
	require.NoError(t, h.db.Model(&sender).Update("is_deleted", true).Error)
	// Soft-deleted graph nodes must not reappear in a repeated create response.
	seed(t, h.db, &models.AutomationNode{Base: models.Base{ID: uuid.NewString(), IsDeleted: true}, AutomationID: workflow.ID, Type: models.NodeTypeExit, Data: datatypes.JSON(`{}`)})
	response = h.call("POST", path, input)
	require.Equal(t, 201, response.Code, response.Body.String())
	var replay models.Automation
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &replay))
	require.Equal(t, workflow.ID, replay.ID)
	require.Len(t, replay.Nodes, 5)
	for model, expected := range map[any]int64{&models.Template{}: 2, &models.Automation{}: 1, &models.Email{}: 0} {
		var count int64
		require.NoError(t, h.db.Model(model).Count(&count).Error)
		require.Equal(t, expected, count)
	}
}
func TestFormJourneyRejectsForeignSenderAndRollsBackPartialGraph(t *testing.T) {
	h := newFormHarness(t)
	f := h.create("optional")
	sender := models.SMTPConfig{Base: models.Base{ID: uuid.NewString()}, TeamID: uuid.NewString(), Host: "smtp.example.com", FromEmail: "hello@example.com"}
	seed(t, h.db, &sender)
	input := map[string]any{"requestId": uuid.NewString(), "name": "Welcome series", "smtpConfigId": sender.ID, "postalAddress": "42 Market Street", "emails": []ai.JourneyEmailDraft{{Name: "Welcome", Subject: "Hello", Body: "Hello there."}, {Name: "Follow-up", Subject: "Hello again", Body: "More to share."}}}
	response := h.call("POST", "/forms/"+f.ID+"/journey", input)
	require.Equal(t, 400, response.Code, response.Body.String())
	require.NoError(t, h.db.Model(&sender).UpdateColumn("team_id", h.team).Error)
	creates := 0
	require.NoError(t, h.db.Callback().Create().Before("gorm:create").Register("journey:fail-second-template", func(tx *gorm.DB) {
		if tx.Statement.Table == "templates" {
			creates++
			if creates == 2 {
				tx.AddError(errors.New("template storage unavailable"))
			}
		}
	}))
	response = h.call("POST", "/forms/"+f.ID+"/journey", input)
	require.Equal(t, 500, response.Code, response.Body.String())
	require.NoError(t, h.db.Callback().Create().Remove("journey:fail-second-template"))
	for _, model := range []any{&models.Template{}, &models.Automation{}, &models.AutomationNode{}, &models.AutomationNodeEdge{}, &models.EmailCategory{}, &models.Email{}} {
		var count int64
		require.NoError(t, h.db.Model(model).Count(&count).Error)
		require.Zero(t, count)
	}
	response = h.call("POST", "/forms/"+f.ID+"/journey", input)
	require.Equal(t, 201, response.Code, response.Body.String())
}
