package marketing_test

import (
	"encoding/json"
	"fmt"
	"kori/internal/automation"
	"kori/internal/automation/processors"
	"kori/internal/config"
	"kori/internal/handlers"
	"kori/internal/models"
	cryptoutil "kori/internal/utils/crypto"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type sendRulesFixture struct {
	db        *gorm.DB
	processor *processors.EmailProcessor
	ctx       *automation.ExecutionContext
	teamID    string
	nextID    string
	step      map[string]interface{}
}

func newSendRulesFixture(t *testing.T) *sendRulesFixture {
	db := testDB(t)
	teamID, senderID, categoryID, templateID, flowID, stepID, nextID := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	password, err := cryptoutil.Encrypt("fixture")
	require.NoError(t, err)
	seed(t, db, &models.EmailCategory{Base: models.Base{ID: categoryID}, TeamID: teamID, Name: "Transactional"})
	seed(t, db, &models.SMTPConfig{Base: models.Base{ID: senderID}, TeamID: teamID, Password: password, FromEmail: "hello@example.com", SupportsTLS: true})
	seed(t, db, &models.Template{Base: models.Base{ID: templateID}, TeamID: teamID, CategoryID: categoryID, HTMLBody: "<p>Hi</p>", Subject: "Hello"})
	seed(t, db, &models.Automation{Base: models.Base{ID: flowID}, TeamID: teamID, Name: "Welcome"})
	seed(t, db, &models.AutomationNodeEdge{Base: models.Base{ID: uuid.NewString()}, AutomationID: flowID, SourceID: stepID, TargetID: nextID})
	cfg := &config.Config{}
	cfg.JWT.Secret = strings.Repeat("x", 32)
	contact := &models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: teamID, Email: "reader@example.com", Status: models.SubscriberStatusActive}
	return &sendRulesFixture{
		db: db, processor: processors.NewEmailProcessor(db, nil, cfg), teamID: teamID, nextID: nextID,
		step: map[string]interface{}{"id": stepID, "templateId": templateID, "smtpConfigId": senderID},
		ctx:  &automation.ExecutionContext{TeamID: teamID, ContactID: contact.ID, Contact: contact, AutomationID: flowID, Execution: &models.AutomationExecution{Base: models.Base{ID: uuid.NewString()}}},
	}
}

func (f *sendRulesFixture) rules(t *testing.T, row models.TeamSendTimeDefaults) {
	row.Base, row.TeamID = models.Base{ID: uuid.NewString()}, f.teamID
	seed(t, f.db, &row)
}

// sent records an automation email (or another kind when key is empty) created ago.
func (f *sendRulesFixture) sent(t *testing.T, ago time.Duration, automationEmail bool) {
	var key *string
	if automationEmail {
		k := "automation:" + uuid.NewString() + ":" + uuid.NewString()
		key = &k
	}
	email := &models.Email{Base: models.Base{ID: uuid.NewString()}, DeliveryKey: key, TeamID: f.teamID, ContactID: f.ctx.ContactID, To: "reader@example.com", From: "hello@example.com", SMTPConfigID: uuid.NewString(), CategoryID: uuid.NewString(), Status: models.EmailStatusSent}
	seed(t, f.db, email)
	require.NoError(t, f.db.Model(email).UpdateColumn("created_at", time.Now().UTC().Add(-ago)).Error)
}

func (f *sendRulesFixture) run(t *testing.T, extra map[string]interface{}) *automation.ProcessResult {
	data := map[string]interface{}{}
	for k, v := range f.step {
		data[k] = v
	}
	for k, v := range extra {
		data[k] = v
	}
	raw, _ := json.Marshal(data)
	result, err := f.processor.Process(f.ctx, &models.AutomationNode{Base: models.Base{ID: f.step["id"].(string)}, Type: models.NodeTypeEmail, Data: raw})
	require.NoError(t, err)
	return result
}

func (f *sendRulesFixture) queued(t *testing.T) int64 {
	var n int64
	require.NoError(t, f.db.Model(&models.Email{}).Where("delivery_key = ?", "automation:"+f.ctx.Execution.ID+":"+f.step["id"].(string)).Count(&n).Error)
	return n
}

// closedWindow opens only three days from today, so it is shut right now in any time zone.
func closedWindow() datatypes.JSON {
	day := (int(time.Now().UTC().Weekday()) + 3) % 7
	return datatypes.JSON(fmt.Sprintf(`{"enabled":true,"start":"09:00","end":"10:00","days":[%d]}`, day))
}

func TestAutomationEmailFrequencyCapDefers(t *testing.T) {
	f := newSendRulesFixture(t)
	f.rules(t, models.TeamSendTimeDefaults{FrequencyCap: 2, FrequencyCapPeriod: "day"})
	f.sent(t, 2*time.Hour, true)
	f.sent(t, time.Hour, true)
	f.sent(t, 30*time.Minute, false) // campaign emails do not count
	f.sent(t, 30*time.Hour, true)    // outside the period

	result := f.run(t, nil)
	require.NotNil(t, result.Wait)
	require.InDelta(t, (22 * time.Hour).Seconds(), result.Wait.Seconds(), 60)
	require.Equal(t, []string{f.step["id"].(string)}, result.NextNodeIDs, "the step resumes itself")
	require.Zero(t, f.queued(t))

	// Under a weekly cap of 4, the same history is still capped; a cap of 5 is not.
	require.NoError(t, f.db.Model(&models.TeamSendTimeDefaults{}).Where("team_id = ?", f.teamID).Updates(map[string]interface{}{"frequency_cap": 5, "frequency_cap_period": "week"}).Error)
	result = f.run(t, nil)
	require.Nil(t, result.Wait)
	require.EqualValues(t, 1, f.queued(t))
}

func TestAutomationEmailFrequencyCapSkip(t *testing.T) {
	f := newSendRulesFixture(t)
	f.rules(t, models.TeamSendTimeDefaults{FrequencyCap: 1, FrequencyCapPeriod: "week"})
	f.sent(t, 3*24*time.Hour, true)

	result := f.run(t, map[string]interface{}{"skipIfCapped": true})
	require.Nil(t, result.Wait)
	require.Equal(t, []string{f.nextID}, result.NextNodeIDs)
	require.Zero(t, f.queued(t))
}

func TestTransactionalAutomationEmailIgnoresSendingRules(t *testing.T) {
	f := newSendRulesFixture(t)
	f.rules(t, models.TeamSendTimeDefaults{FrequencyCap: 1, FrequencyCapPeriod: "day", AutomationWindow: closedWindow()})
	f.sent(t, time.Hour, true)

	result := f.run(t, map[string]interface{}{"transactional": true, "skipIfCapped": true})
	require.Nil(t, result.Wait)
	require.Equal(t, []string{f.nextID}, result.NextNodeIDs)
	require.EqualValues(t, 1, f.queued(t))
}

func TestAutomationEmailSendWindow(t *testing.T) {
	f := newSendRulesFixture(t)
	f.rules(t, models.TeamSendTimeDefaults{DefaultTimezone: "Asia/Tokyo", AutomationWindow: closedWindow()})
	f.ctx.Contact.Timezone = "America/Los_Angeles"

	result := f.run(t, nil)
	require.NotNil(t, result.Wait)
	resumeAt, err := time.Parse(time.RFC3339, result.Data["resumeAt"].(string))
	require.NoError(t, err)
	la, _ := time.LoadLocation("America/Los_Angeles")
	local := resumeAt.In(la)
	require.Equal(t, 9, local.Hour(), "opens at 09:00 in the reader's time zone")
	require.Equal(t, (int(time.Now().UTC().Weekday())+3)%7, int(local.Weekday()))
	require.Zero(t, f.queued(t))

	// The automation's own window replaces the workspace default; disabled means any time.
	require.NoError(t, f.db.Model(&models.Automation{}).Where("id = ?", f.ctx.AutomationID).Update("send_window", datatypes.JSON(`{"enabled":false}`)).Error)
	result = f.run(t, nil)
	require.Nil(t, result.Wait)
	require.EqualValues(t, 1, f.queued(t))

	// Once queued, a retried step is never deferred again.
	require.NoError(t, f.db.Model(&models.Automation{}).Where("id = ?", f.ctx.AutomationID).Update("send_window", closedWindow()).Error)
	result = f.run(t, nil)
	require.Nil(t, result.Wait)
	require.EqualValues(t, 1, f.queued(t))
}

func TestWorkspaceSendingRulesAPI(t *testing.T) {
	db := testDB(t)
	handler := handlers.NewMarketingHandler(db, strings.Repeat("x", 32), "https://api.example.com")
	teamID := uuid.NewString()
	call := func(method, body string) (int, string) {
		e := echo.New()
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(method, "/sending-rules", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c := e.NewContext(req, rec)
		c.Set("teamID", teamID)
		if err := handler.SendingRules(c); err != nil {
			e.HTTPErrorHandler(err, c)
		}
		return rec.Code, rec.Body.String()
	}
	code, body := call("GET", "")
	require.Equal(t, 200, code)
	require.JSONEq(t, `{"timezone":"UTC","sendWindow":null,"frequencyCap":0,"frequencyCapPeriod":"day"}`, body)
	for _, bad := range []string{
		`{"timezone":"Mars/Base","frequencyCapPeriod":"day"}`,
		`{"timezone":"UTC","frequencyCap":3,"frequencyCapPeriod":"month"}`,
		`{"timezone":"UTC","frequencyCapPeriod":"day","sendWindow":{"enabled":true,"start":"09:00","end":"17:00","days":[]}}`,
	} {
		code, _ = call("PUT", bad)
		require.Equal(t, 400, code, bad)
	}
	rules := `{"timezone":"Europe/London","sendWindow":{"enabled":true,"start":"09:00","end":"17:00","days":[1,2,3,4,5]},"frequencyCap":3,"frequencyCapPeriod":"week"}`
	for range 2 { // the second save updates the same row
		code, body = call("PUT", rules)
		require.Equal(t, 200, code, body)
		require.JSONEq(t, rules, body)
	}
	var rows int64
	require.NoError(t, db.Model(&models.TeamSendTimeDefaults{}).Where("team_id = ?", teamID).Count(&rows).Error)
	require.EqualValues(t, 1, rows)
}
