package handlers

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"kori/internal/automation"
	"kori/internal/automation/processors"
	"kori/internal/models"
	"kori/internal/services"
)

// executions lists the distinct execution IDs queued for an automation. A
// repeated delivery of one occurrence reuses its ID, which the queue and the
// engine both treat as the same run, so distinct IDs are the real starts.
func (r *recordedAutomations) executions(automationID string) map[string]map[string]interface{} {
	r.mu.Lock()
	defer r.mu.Unlock()
	runs := map[string]map[string]interface{}{}
	for _, task := range r.tasks {
		if task.AutomationID == automationID {
			runs[task.ExecutionID] = task.TriggerData
		}
	}
	return runs
}

func triggerTestDB(t *testing.T) *gorm.DB {
	var dialect gorm.Dialector = sqlite.Open(":memory:")
	if dsn := os.Getenv("POSTHOOT_TEST_DATABASE_URL"); dsn != "" {
		admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		require.NoError(t, err)
		schema := "trigger_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
		require.NoError(t, admin.Exec("CREATE SCHEMA "+schema).Error)
		t.Cleanup(func() { admin.Exec("DROP SCHEMA " + schema + " CASCADE"); raw, _ := admin.DB(); raw.Close() })
		dialect = postgres.Open(dsn + " search_path=" + schema)
	}
	db, err := gorm.Open(dialect, &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true, Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	raw, err := db.DB()
	require.NoError(t, err)
	raw.SetMaxOpenConns(1)
	t.Cleanup(func() { raw.Close() })
	require.NoError(t, db.AutoMigrate(&models.Contact{}, &models.Tag{}, &models.MailingList{}, &models.Automation{}, &models.AutomationNode{}, &models.AutomationNodeEdge{},
		&models.Segment{}, &models.SegmentContact{}, &models.ContactEvent{}, &models.SuppressionList{}, &models.TeamSendTimeDefaults{}))
	return db
}

// liveAutomation creates an active automation whose START step has settings.
func liveAutomation(t *testing.T, db *gorm.DB, team, trigger string, settings map[string]interface{}) models.Automation {
	data, err := json.Marshal(settings)
	require.NoError(t, err)
	a := models.Automation{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: trigger, TriggerEvent: trigger, IsActive: true}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&a).Error)
	require.NoError(t, db.Create(&models.AutomationNode{Base: models.Base{ID: uuid.NewString()}, AutomationID: a.ID, Type: models.NodeTypeStart, Data: data}).Error)
	return a
}

func settle() { time.Sleep(150 * time.Millisecond) } // event handlers run on goroutines

func TestTagAndListTriggersStartOncePerRealChange(t *testing.T) {
	db := triggerTestDB(t)
	queue := &recordedAutomations{}
	require.NoError(t, automation.NewTriggerManager(db, queue).Initialize(context.Background()))
	team := uuid.NewString()
	listA := models.MailingList{Base: models.Base{ID: uuid.NewString()}, Name: "Trial", TeamID: team}
	listB := models.MailingList{Base: models.Base{ID: uuid.NewString()}, Name: "Customers", TeamID: team}
	require.NoError(t, db.Create(&listA).Error)
	require.NoError(t, db.Create(&listB).Error)
	vip := models.Tag{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "vip"}
	beta := models.Tag{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "beta"}
	require.NoError(t, db.Create(&vip).Error)
	require.NoError(t, db.Create(&beta).Error)
	onVIP := liveAutomation(t, db, team, automation.TriggerTagAdded, map[string]interface{}{"tagId": vip.ID})
	onBeta := liveAutomation(t, db, team, automation.TriggerTagAdded, map[string]interface{}{"tagId": beta.ID})
	unset := liveAutomation(t, db, team, automation.TriggerTagAdded, map[string]interface{}{})
	joinedA := liveAutomation(t, db, team, automation.TriggerListJoined, map[string]interface{}{"listId": listA.ID})
	joinedB := liveAutomation(t, db, team, automation.TriggerListJoined, map[string]interface{}{"listId": listB.ID})

	// Joining a list: creating the contact in list A.
	contact := models.Contact{TeamID: team, ListID: listA.ID, Email: "ada@example.net", Status: models.SubscriberStatusActive}
	require.NoError(t, db.Create(&contact).Error)
	require.Eventually(t, func() bool { return len(queue.executions(joinedA.ID)) == 1 }, 2*time.Second, 10*time.Millisecond)

	// Tags from the contact page: only newly attached tags start automations.
	h := NewMarketingHandler(db, "secret", "")
	e := echo.New()
	setTags := func(ids ...string) {
		body, _ := json.Marshal(map[string]interface{}{"tagIds": ids})
		req := httptest.NewRequest("PUT", "/", strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		c := e.NewContext(req, httptest.NewRecorder())
		c.Set("teamID", team)
		c.SetParamNames("id")
		c.SetParamValues(contact.ID)
		require.NoError(t, h.ContactTags(c))
	}
	setTags(vip.ID)
	setTags(vip.ID, beta.ID)
	setTags(vip.ID, beta.ID)
	require.Eventually(t, func() bool { return len(queue.executions(onBeta.ID)) == 1 }, 2*time.Second, 10*time.Millisecond)
	settle()
	require.Len(t, queue.executions(onVIP.ID), 1, "keeping a tag does not start again")
	require.Len(t, queue.executions(onBeta.ID), 1)
	require.Empty(t, queue.executions(unset.ID), "a tag trigger needs an explicit tag")
	for _, data := range queue.executions(onVIP.ID) {
		require.Equal(t, "vip", data["tag_name"])
	}

	// The TAG step: a tag the contact already has is not a change.
	ctx := &automation.ExecutionContext{AutomationID: uuid.NewString(), ContactID: contact.ID, TeamID: team, Contact: &contact}
	node := &models.AutomationNode{Base: models.Base{ID: uuid.NewString()}, Type: models.NodeTypeTag, Data: datatypes.JSON(`{"action":"add","tags":["vip","new"]}`)}
	tagStep := processors.NewTagProcessor(db)
	_, err := tagStep.Process(ctx, node)
	require.NoError(t, err)
	settle()
	require.Len(t, queue.executions(onVIP.ID), 1)
	setTags() // removing and re-adding is a new change
	_, err = tagStep.Process(ctx, node)
	require.NoError(t, err)
	require.Eventually(t, func() bool { return len(queue.executions(onVIP.ID)) == 2 }, 2*time.Second, 10*time.Millisecond)

	// Moving lists with the ADD_TO_LIST step; repeating it is not a change.
	listStep := processors.NewAddToListProcessor(db)
	move := &models.AutomationNode{Base: models.Base{ID: uuid.NewString()}, Type: models.NodeTypeAddToList, Data: datatypes.JSON(`{"listId":"` + listB.ID + `"}`)}
	for range 2 {
		_, err = listStep.Process(ctx, move)
		require.NoError(t, err)
	}
	require.Eventually(t, func() bool { return len(queue.executions(joinedB.ID)) == 1 }, 2*time.Second, 10*time.Millisecond)
	// An edit that keeps the list is not a join.
	edit := models.Contact{Base: models.Base{ID: contact.ID}, ListID: listB.ID, FirstName: "Ada"}
	require.NoError(t, db.Model(&edit).Where("id = ?", contact.ID).Updates(&edit).Error)
	settle()
	require.Len(t, queue.executions(joinedB.ID), 1)
	require.Len(t, queue.executions(joinedA.ID), 1)
}

func TestSegmentEntryStartsOnlyNewMembersOnce(t *testing.T) {
	db := triggerTestDB(t)
	queue := &recordedAutomations{}
	triggers := automation.NewTriggerManager(db, queue)
	require.NoError(t, triggers.Initialize(context.Background()))
	team, other := uuid.NewString(), uuid.NewString()
	plain := db.Session(&gorm.Session{SkipHooks: true})
	early := models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: team, ListID: uuid.NewString(), Email: "early@example.net", Company: "Acme"}
	foreign := models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: other, ListID: uuid.NewString(), Email: "spy@example.net", Company: "Acme"}
	require.NoError(t, plain.Create(&early).Error)
	require.NoError(t, plain.Create(&foreign).Error)
	segment := models.Segment{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "Acme", IsDynamic: true,
		Rules: datatypes.JSON(`{"matchType":"ANY","rules":[{"field":"company","operator":"equals","value":"Acme"},{"field":"city","operator":"equals","value":"Pune"}]}`)}
	require.NoError(t, db.Create(&segment).Error)
	entered := liveAutomation(t, db, team, automation.TriggerSegmentEntered, map[string]interface{}{"segmentId": segment.ID})
	require.NoError(t, db.Model(&models.Automation{}).Where("id = ?", entered.ID).Update("updated_at", time.Now().Add(-time.Hour)).Error)

	service := services.NewSegmentService(db)
	require.NoError(t, triggers.RefreshSegmentTriggers(context.Background())) // baseline
	var members []string
	require.NoError(t, db.Model(&models.SegmentContact{}).Where("segment_id = ?", segment.ID).Pluck("contact_id", &members).Error)
	require.Equal(t, []string{early.ID}, members, "ANY rules stay inside the workspace")

	late := models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: team, ListID: uuid.NewString(), Email: "late@example.net", City: "Pune"}
	require.NoError(t, plain.Create(&late).Error)
	// A workflow switched on after the last refresh must not inherit this entry.
	switchedOn := liveAutomation(t, db, team, automation.TriggerSegmentEntered, map[string]interface{}{"segmentId": segment.ID})
	require.NoError(t, service.RefreshSegment(context.Background(), segment.ID))
	require.NoError(t, service.RefreshSegment(context.Background(), segment.ID))
	require.Eventually(t, func() bool { return len(queue.executions(entered.ID)) == 1 }, 2*time.Second, 10*time.Millisecond)
	settle()
	runs := queue.executions(entered.ID)
	require.Len(t, runs, 1, "the first refresh is a baseline and a repeat refresh adds nobody")
	for _, task := range queue.tasks {
		if task.AutomationID == entered.ID {
			require.Equal(t, late.ID, task.ContactID)
		}
	}
	require.Empty(t, queue.executions(switchedOn.ID))

	// Leaving and coming back is a new entry.
	require.NoError(t, plain.Model(&models.Contact{}).Where("id = ?", late.ID).Update("city", "Delhi").Error)
	require.NoError(t, service.RefreshSegment(context.Background(), segment.ID))
	require.NoError(t, plain.Model(&models.Contact{}).Where("id = ?", late.ID).Update("city", "Pune").Error)
	require.NoError(t, service.RefreshSegment(context.Background(), segment.ID))
	require.Eventually(t, func() bool { return len(queue.executions(entered.ID)) == 2 }, 2*time.Second, 10*time.Millisecond)

	bad := models.Segment{TeamID: team, Rules: datatypes.JSON(`{"matchType":"ALL","rules":[{"field":"email = email OR 1=1 --","operator":"is_set"}]}`)}
	require.Error(t, service.ValidateRules(&bad), "rule fields are contact columns, never SQL")
}

func TestDateTriggersFireOncePerContactAndDate(t *testing.T) {
	db := triggerTestDB(t)
	queue := &recordedAutomations{}
	triggers := automation.NewTriggerManager(db, queue)
	team := uuid.NewString()
	require.NoError(t, db.Create(&models.TeamSendTimeDefaults{TeamID: team, DefaultTimezone: "Asia/Kolkata"}).Error)
	plain := db.Session(&gorm.Session{SkipHooks: true})
	contact := func(email, metadata string, status models.SubscriberStatus) models.Contact {
		c := models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: team, ListID: uuid.NewString(), Email: email, Status: status, Metadata: datatypes.JSON(metadata)}
		require.NoError(t, plain.Create(&c).Error)
		return c
	}
	birthday := contact("bday@example.net", `{"birthday":"1990-11-09"}`, models.SubscriberStatusActive)
	contact("other@example.net", `{"birthday":"1990-11-10"}`, models.SubscriberStatusActive)
	contact("gone@example.net", `{"birthday":"1991-11-09"}`, models.SubscriberStatusUnsubscribed)
	leap := contact("leap@example.net", `{"birthday":"2000-02-29"}`, models.SubscriberStatusActive)
	trial := contact("trial@example.net", `{"trial_end":"2026-11-08T00:00:00Z"}`, models.SubscriberStatusActive)
	contact("future@example.net", `{"birthday":"2030-11-09"}`, models.SubscriberStatusActive)

	beforeBirthday := liveAutomation(t, db, team, automation.TriggerDateReached, map[string]interface{}{"field": "birthday", "offsetDays": -3, "yearly": true})
	trialEnding := liveAutomation(t, db, team, automation.TriggerDateReached, map[string]interface{}{"field": "trial_end", "offsetDays": -2})
	onBirthday := liveAutomation(t, db, team, automation.TriggerDateReached, map[string]interface{}{"field": "birthday", "offsetDays": 0, "yearly": true})

	// 08:00 in Kolkata: too early to start anything.
	require.NoError(t, triggers.RunDateTriggers(context.Background(), time.Date(2026, 11, 6, 2, 30, 0, 0, time.UTC)))
	require.Empty(t, queue.tasks)
	// 10:30 in Kolkata on 6 November, then the hourly runs that follow.
	for hour := 5; hour < 10; hour++ {
		require.NoError(t, triggers.RunDateTriggers(context.Background(), time.Date(2026, 11, 6, hour, 0, 0, 0, time.UTC)))
	}
	runs := queue.executions(beforeBirthday.ID)
	require.Len(t, runs, 1, "one start per contact and date, however many runs")
	for _, data := range runs {
		require.Equal(t, "2026-11-09", data["date"])
	}
	require.Len(t, queue.executions(trialEnding.ID), 1)
	require.Empty(t, queue.executions(onBirthday.ID))
	for _, task := range queue.tasks {
		if task.AutomationID == beforeBirthday.ID {
			require.Equal(t, birthday.ID, task.ContactID)
		}
		if task.AutomationID == trialEnding.ID {
			require.Equal(t, trial.ID, task.ContactID)
		}
	}
	// Next year it comes round again.
	require.NoError(t, triggers.RunDateTriggers(context.Background(), time.Date(2027, 11, 6, 6, 0, 0, 0, time.UTC)))
	require.Len(t, queue.executions(beforeBirthday.ID), 2)
	require.Len(t, queue.executions(trialEnding.ID), 1, "a one-off date does not repeat")
	// A 29 February birthday is celebrated on 28 February in other years.
	require.NoError(t, triggers.RunDateTriggers(context.Background(), time.Date(2027, 2, 28, 6, 0, 0, 0, time.UTC)))
	require.Len(t, queue.executions(onBirthday.ID), 1)
	for _, task := range queue.tasks {
		if task.AutomationID == onBirthday.ID {
			require.Equal(t, leap.ID, task.ContactID)
		}
	}
}

func TestCustomEventsStartNamedAutomationsWithProperties(t *testing.T) {
	db := triggerTestDB(t)
	queue := &recordedAutomations{}
	require.NoError(t, automation.NewTriggerManager(db, queue).Initialize(context.Background()))
	team := uuid.NewString()
	list := models.MailingList{Base: models.Base{ID: uuid.NewString()}, Name: "Signups", TeamID: team}
	require.NoError(t, db.Create(&list).Error)
	started := liveAutomation(t, db, team, automation.TriggerEvent, map[string]interface{}{"eventName": "Trial_Started"})
	other := liveAutomation(t, db, team, automation.TriggerEvent, map[string]interface{}{"eventName": "order_placed"})
	elsewhere := liveAutomation(t, db, uuid.NewString(), automation.TriggerEvent, map[string]interface{}{"eventName": "trial_started"})

	handler := RecordContactEvent(db)
	e := echo.New()
	send := func(body string) (*httptest.ResponseRecorder, error) {
		req := httptest.NewRequest("POST", "/api/v1/events", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("teamID", team)
		return rec, handler(c)
	}
	_, err := send(`{"email":"new@example.net","event":"trial_started"}`)
	require.Equal(t, 404, err.(*echo.HTTPError).Code, "an unknown email needs a list")
	_, err = send(`{"email":"new@example.net"}`)
	require.Equal(t, 400, err.(*echo.HTTPError).Code)
	_, err = send(`{"email":"new@example.net","contactId":"` + uuid.NewString() + `","event":"x"}`)
	require.Equal(t, 400, err.(*echo.HTTPError).Code)

	rec, err := send(`{"email":"New@Example.net","event":"trial_started","listId":"` + list.ID + `","properties":{"plan":"pro","seats":5,"annual":true,"nested":{"x":1}}}`)
	require.NoError(t, err)
	require.Equal(t, 202, rec.Code)
	var out struct {
		ID             string `json:"id"`
		ContactID      string `json:"contactId"`
		ContactCreated bool   `json:"contactCreated"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.True(t, out.ContactCreated)
	// The same contact is found by email afterwards.
	rec, err = send(`{"email":"new@example.net","event":"order_placed"}`)
	require.NoError(t, err)
	var second struct {
		ContactID      string `json:"contactId"`
		ContactCreated bool   `json:"contactCreated"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &second))
	require.Equal(t, out.ContactID, second.ContactID)
	require.False(t, second.ContactCreated)

	require.Eventually(t, func() bool { return len(queue.executions(started.ID)) == 1 && len(queue.executions(other.ID)) == 1 }, 2*time.Second, 10*time.Millisecond)
	settle()
	require.Empty(t, queue.executions(elsewhere.ID))
	for _, data := range queue.executions(started.ID) {
		require.Equal(t, "pro", data["event_plan"])
		require.Equal(t, "5", data["event_seats"])
		require.Equal(t, "true", data["event_annual"])
		require.NotContains(t, data, "event_nested")
		require.Equal(t, out.ID, data["event_id"])
	}
	var stored models.ContactEvent
	require.NoError(t, db.First(&stored, "id = ?", out.ID).Error)
	require.Equal(t, out.ContactID, stored.ContactID)
	require.JSONEq(t, `{"plan":"pro","seats":5,"annual":true,"nested":{"x":1}}`, string(stored.Properties))
}
