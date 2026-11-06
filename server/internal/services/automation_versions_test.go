package services

import (
	"context"
	"encoding/json"
	"kori/internal/models"
	"kori/internal/sending"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// versionsDB uses PostgreSQL when POSTHOOT_TEST_DATABASE_URL is set, else SQLite.
func versionsDB(t *testing.T) *gorm.DB {
	t.Helper()
	os.Setenv("TEST_MODE", "true")
	var dialect gorm.Dialector = sqlite.Open("file:" + uuid.NewString() + "?mode=memory&cache=shared")
	if url := os.Getenv("POSTHOOT_TEST_DATABASE_URL"); url != "" {
		admin, err := gorm.Open(postgres.Open(url), &gorm.Config{})
		require.NoError(t, err)
		schema := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
		require.NoError(t, admin.Exec("CREATE SCHEMA "+schema).Error)
		t.Cleanup(func() { admin.Exec("DROP SCHEMA " + schema + " CASCADE"); raw, _ := admin.DB(); raw.Close() })
		dialect = postgres.Open(url + " search_path=" + schema)
	}
	db, err := gorm.Open(dialect, &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true, Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.User{}, &models.Automation{}, &models.AutomationNode{}, &models.AutomationNodeEdge{}, &models.AutomationExecution{}, &models.AutomationVersion{},
		&models.Email{}, &models.EmailTracking{}, &models.AutomationGoal{}, &models.GoalConversion{}, &sending.Message{}))
	raw, err := db.DB()
	require.NoError(t, err)
	raw.SetMaxOpenConns(1)
	t.Cleanup(func() { raw.Close() })
	return db
}

func step(id string, kind models.NodeType, data string) models.AutomationNode {
	return models.AutomationNode{Base: models.Base{ID: id}, Type: kind, Data: datatypes.JSON(data)}
}

func link(from, to string) models.AutomationNodeEdge {
	return models.AutomationNodeEdge{SourceID: from, TargetID: to}
}

func TestPublishMigrateAndRestore(t *testing.T) {
	db := versionsDB(t)
	s := NewAutomationService(db)
	ctx := context.Background()
	start, wait, exit, later := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	user := models.User{Email: "pat@example.com", FirstName: "Pat", Password: "x", TeamID: uuid.NewString()}
	require.NoError(t, db.Create(&user).Error)

	a := &models.Automation{Name: "Welcome", TeamID: user.TeamID, Nodes: []models.AutomationNode{step(start, models.NodeTypeStart, `{}`), step(wait, models.NodeTypeWait, `{"duration":"1h"}`), step(exit, models.NodeTypeExit, `{}`)},
		Edges: []models.AutomationNodeEdge{link(start, wait), link(wait, exit)}}
	require.NoError(t, s.SaveGraph(ctx, a))
	v1, err := s.Publish(ctx, a.ID, user.ID, false)
	require.NoError(t, err)
	assert.Equal(t, 1, v1.Number)

	// Two contacts inside v1: one waiting on the wait step, one on the exit step.
	onWait := models.AutomationExecution{AutomationID: a.ID, Version: 1, ContactID: uuid.NewString(), Status: models.ExecutionStatusWaiting, CurrentNodeID: wait, StartedAt: time.Now()}
	onExit := models.AutomationExecution{AutomationID: a.ID, Version: 1, ContactID: uuid.NewString(), Status: models.ExecutionStatusWaiting, CurrentNodeID: exit, StartedAt: time.Now()}
	require.NoError(t, db.Create(&onWait).Error)
	require.NoError(t, db.Create(&onExit).Error)

	// Editing a live automation changes only the draft.
	a.TriggerEvent = "contact.created"
	a.Nodes = []models.AutomationNode{step(start, models.NodeTypeStart, `{}`), step(wait, models.NodeTypeWait, `{"duration":"2h"}`), step(later, models.NodeTypeExit, `{}`)}
	a.Edges = []models.AutomationNodeEdge{link(start, wait), link(wait, later)}
	require.NoError(t, s.SaveGraph(ctx, a))
	var stored models.Automation
	require.NoError(t, db.First(&stored, "id = ?", a.ID).Error)
	assert.True(t, stored.HasDraft)
	assert.Equal(t, 1, stored.LiveVersion)
	assert.Equal(t, "manual", stored.TriggerEvent, "the live trigger waits for publish")
	live := models.Automation{Base: models.Base{ID: a.ID}}
	require.NoError(t, models.LoadAutomationGraph(db, &live, 1))
	assert.Len(t, live.Nodes, 3)
	assert.Equal(t, exit, live.Edges[1].TargetID, "the published graph is untouched")

	// Publishing without moving keeps running contacts on v1.
	v2, err := s.Publish(ctx, a.ID, user.ID, false)
	require.NoError(t, err)
	assert.Equal(t, 2, v2.Number)
	require.NoError(t, db.First(&onWait, "id = ?", onWait.ID).Error)
	assert.Equal(t, 1, onWait.Version)
	require.NoError(t, db.First(&stored, "id = ?", a.ID).Error)
	assert.False(t, stored.HasDraft)
	assert.Equal(t, "contact.created", stored.TriggerEvent)

	// Publishing with "move running contacts": mapped by step ID, removed steps exit.
	_, err = s.Publish(ctx, a.ID, user.ID, true)
	require.NoError(t, err)
	require.NoError(t, db.First(&onWait, "id = ?", onWait.ID).Error)
	require.NoError(t, db.First(&onExit, "id = ?", onExit.ID).Error)
	assert.Equal(t, 3, onWait.Version)
	assert.Equal(t, models.ExecutionStatusWaiting, onWait.Status)
	assert.Equal(t, models.ExecutionStatusCompleted, onExit.Status)
	assert.Contains(t, onExit.Error, "removed in version 3")

	// Restore v1 as the draft, then discard back to live.
	require.NoError(t, s.RestoreDraft(ctx, a.ID, 1))
	draft, err := s.draft(ctx, a.ID)
	require.NoError(t, err)
	assert.True(t, draft.HasDraft)
	assert.Equal(t, "manual", draft.TriggerEvent)
	ids := []string{}
	for _, n := range draft.Nodes {
		ids = append(ids, n.ID)
	}
	assert.ElementsMatch(t, []string{start, wait, exit}, ids)
	require.NoError(t, s.RestoreDraft(ctx, a.ID, 0))
	draft, err = s.draft(ctx, a.ID)
	require.NoError(t, err)
	assert.False(t, draft.HasDraft)
	assert.Equal(t, "contact.created", draft.TriggerEvent)
	assert.Len(t, draft.Nodes, 3)

	versions, err := s.Versions(ctx, a.ID)
	require.NoError(t, err)
	require.Len(t, versions, 3)
	assert.Equal(t, 3, versions[0].Number)
	assert.Equal(t, "Pat", versions[0].PublishedBy)
}

func TestBackfillPinsExistingAutomationsToVersionOne(t *testing.T) {
	db := versionsDB(t)
	start, exit := uuid.NewString(), uuid.NewString()
	a := models.Automation{Name: "Old", TeamID: uuid.NewString(), IsActive: true, Nodes: []models.AutomationNode{step(start, models.NodeTypeStart, `{}`), step(exit, models.NodeTypeExit, `{}`)}, Edges: []models.AutomationNodeEdge{link(start, exit)}}
	require.NoError(t, db.Create(&a).Error)
	running := models.AutomationExecution{AutomationID: a.ID, ContactID: uuid.NewString(), Status: models.ExecutionStatusWaiting, CurrentNodeID: start, StartedAt: time.Now()}
	require.NoError(t, db.Create(&running).Error)

	require.NoError(t, models.BackfillAutomationVersions(db))
	require.NoError(t, db.First(&a, "id = ?", a.ID).Error)
	require.NoError(t, db.First(&running, "id = ?", running.ID).Error)
	assert.Equal(t, 1, a.LiveVersion)
	assert.Equal(t, 1, running.Version)
	assert.True(t, a.IsActive)
	require.NoError(t, models.LoadAutomationGraph(db, &a, 1))
	assert.Len(t, a.Nodes, 2)
	assert.Len(t, a.Edges, 1)
}

func TestStepStats(t *testing.T) {
	if os.Getenv("POSTHOOT_TEST_DATABASE_URL") == "" {
		t.Skip("POSTHOOT_TEST_DATABASE_URL required: step stats use PostgreSQL JSON functions")
	}
	db := versionsDB(t)
	s := NewAutomationService(db)
	automationID, start, email, exit := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	goal := models.AutomationGoal{AutomationID: automationID, Name: "Buy", GoalType: models.GoalTypePurchase}
	require.NoError(t, db.Create(&goal).Error)
	now := time.Now()
	execution := func(status models.ExecutionStatus, at string, path ...string) models.AutomationExecution {
		log := []map[string]string{}
		for _, id := range path {
			log = append(log, map[string]string{"nodeId": id})
		}
		raw, _ := json.Marshal(log)
		x := models.AutomationExecution{AutomationID: automationID, Version: 1, ContactID: uuid.NewString(), Status: status, CurrentNodeID: at, ExecutionLog: raw, StartedAt: now}
		require.NoError(t, db.Create(&x).Error)
		return x
	}
	sendEmail := func(x models.AutomationExecution, status models.EmailStatus, events ...models.EmailTrackingEvent) string {
		key := "automation:" + x.ID + ":" + email
		e := models.Email{Base: models.Base{ID: uuid.NewString()}, DeliveryKey: &key, From: "a@example.com", To: "b@example.com", Subject: "Hi", Body: "Hi", Status: status, TeamID: uuid.NewString(), ContactID: x.ContactID, TemplateID: uuid.NewString(), CampaignID: uuid.NewString(), SMTPConfigID: uuid.NewString(), CategoryID: uuid.NewString()}
		require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&e).Error)
		for _, ev := range events {
			// Repeat opens count once.
			for range 2 {
				require.NoError(t, db.Create(&models.EmailTracking{EmailID: e.ID, Event: ev, Timestamp: now}).Error)
			}
		}
		return e.ID
	}

	a := execution(models.ExecutionStatusCompleted, exit, start, email, exit)
	b := execution(models.ExecutionStatusCompleted, exit, start, email, exit)
	c := execution(models.ExecutionStatusWaiting, email, start, email)
	execution(models.ExecutionStatusCompleted, start, start) // suppressed before the email
	old := execution(models.ExecutionStatusCompleted, exit, start, email, exit)
	require.NoError(t, db.Model(&old).UpdateColumn("started_at", now.AddDate(0, -3, 0)).Error)

	delivered := sendEmail(a, models.EmailStatusSent, models.EmailTrackingEventOpen, models.EmailTrackingEventClick)
	sendEmail(b, models.EmailStatusSent, models.EmailTrackingEventOpen, models.EmailTrackingEventUnsubscribe)
	sendEmail(c, models.EmailStatusFailed)
	require.NoError(t, db.Create(&sending.Message{ID: uuid.NewString(), TeamID: uuid.NewString(), DomainID: uuid.NewString(), IdempotencyKey: "k", EmailID: delivered, Status: "DELIVERED"}).Error)
	require.NoError(t, db.Create(&models.GoalConversion{GoalID: goal.ID, AutomationID: automationID, ContactID: a.ContactID, ExecutionID: &a.ID, ConvertedAt: now}).Error)

	stats, err := s.StepStats(context.Background(), automationID, now.AddDate(0, 0, -30), now.Add(time.Minute))
	require.NoError(t, err)
	byNode := map[string]StepStats{}
	for _, st := range stats {
		byNode[st.NodeID] = st
	}
	require.Len(t, byNode, 3)
	assert.Equal(t, StepStats{NodeID: start, Entered: 4, MovedOn: 3, Exited: 1, Conversions: 1}, byNode[start])
	e := byNode[email]
	assert.Equal(t, [5]int64{3, 1, 2, 0, 1}, [5]int64{e.Entered, e.Waiting, e.MovedOn, e.Exited, e.Conversions})
	assert.Equal(t, [4]int64{2, 2, 1, 1}, [4]int64{e.Sent, e.Opens, e.Clicks, e.Unsubscribes})
	require.NotNil(t, e.Delivered)
	assert.Equal(t, int64(1), *e.Delivered)
	assert.InDelta(t, 1.0, e.OpenRate, 0.001)
	assert.InDelta(t, 0.5, e.ClickRate, 0.001)
	assert.InDelta(t, 0.5, e.UnsubscribeRate, 0.001)
	assert.Equal(t, int64(2), byNode[exit].Exited)
	assert.Nil(t, byNode[start].Delivered)
}
