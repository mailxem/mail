package automation

import (
	"context"
	"kori/internal/models"
	"kori/internal/sending"
	"kori/internal/tasks"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type taskRecorder struct{ queued []tasks.AutomationExecuteTask }

func (r *taskRecorder) EnqueueAutomationTask(_ context.Context, task tasks.AutomationExecuteTask, _ time.Duration) error {
	if task.ExecutionID == "" {
		task.ExecutionID = uuid.NewString()
	}
	r.queued = append(r.queued, task)
	return nil
}

// sentEmails counts EMAIL steps that reached their processor.
type sentEmails struct {
	db   *gorm.DB
	sent int
}

func (p *sentEmails) Process(ctx *ExecutionContext, node *models.AutomationNode) (*ProcessResult, error) {
	p.sent++
	return &ProcessResult{NextNodeIDs: testNextNodes(p.db, node), UpdateVars: map[string]interface{}{"last_email_id": ctx.Variables["test_email_id"]}}, nil
}
func (p *sentEmails) Validate(*models.AutomationNode) error { return nil }
func (p *sentEmails) Type() models.NodeType                 { return models.NodeTypeEmail }

// waitUntil mirrors processors.WaitUntilProcessor, which this package cannot import.
type waitUntil struct{ db *gorm.DB }

func (p *waitUntil) Process(ctx *ExecutionContext, node *models.AutomationNode) (*ProcessResult, error) {
	var edge models.AutomationNodeEdge
	if err := p.db.Where("source_id = ? AND label = ?", node.ID, "timeout").First(&edge).Error; err != nil {
		return nil, err
	}
	d := time.Hour
	return &ProcessResult{NextNodeIDs: []string{edge.TargetID}, Wait: &d}, nil
}
func (p *waitUntil) Validate(*models.AutomationNode) error { return nil }
func (p *waitUntil) Type() models.NodeType                 { return models.NodeTypeWaitUntil }

type flowFixture struct {
	t       *testing.T
	db      *gorm.DB
	engine  *Engine
	queue   *taskRecorder
	email   *sentEmails
	auto    *models.Automation
	contact *models.Contact
	ids     map[string]string
}

func flowDB(t *testing.T) *gorm.DB {
	var dialect gorm.Dialector = sqlite.Open("file:" + uuid.NewString() + "?mode=memory&cache=shared")
	if url := os.Getenv("POSTHOOT_TEST_DATABASE_URL"); url != "" {
		admin, err := gorm.Open(postgres.Open(url), &gorm.Config{})
		require.NoError(t, err)
		schema := "flow_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
		require.NoError(t, admin.Exec("CREATE SCHEMA "+schema).Error)
		t.Cleanup(func() { admin.Exec("DROP SCHEMA " + schema + " CASCADE"); raw, _ := admin.DB(); raw.Close() })
		dialect = postgres.Open(url + " search_path=" + schema)
	}
	db, err := gorm.Open(dialect, &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true, Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Automation{}, &models.AutomationNode{}, &models.AutomationNodeEdge{}, &models.AutomationExecution{}, &models.Contact{}, &models.Email{}, &models.SuppressionList{}, &sending.Suppression{}, &models.AutomationGoal{}, &models.GoalConversion{}))
	raw, err := db.DB()
	require.NoError(t, err)
	raw.SetMaxOpenConns(1)
	t.Cleanup(func() { raw.Close() })
	return db
}

// newFlow saves a workflow whose nodes and {source, target, label} edges are named; f.ids maps names to IDs.
func newFlow(t *testing.T, reentry string, nodes []models.AutomationNode, edges [][3]string) *flowFixture {
	os.Setenv("TEST_MODE", "true")
	db := flowDB(t)
	f := &flowFixture{t: t, db: db, queue: &taskRecorder{}, email: &sentEmails{db: db}, ids: map[string]string{}}
	f.engine = NewEngine(db, f.queue)
	f.engine.RegisterProcessor(&MockStartProcessor{db: db})
	f.engine.RegisterProcessor(f.email)
	f.engine.RegisterProcessor(&MockExitProcessor{})
	f.engine.RegisterProcessor(&MockWaitProcessor{db: db})
	f.engine.RegisterProcessor(&waitUntil{db: db})
	for i := range nodes {
		name := nodes[i].ID
		f.ids[name] = uuid.NewString()
		nodes[i].ID = f.ids[name]
	}
	var es []models.AutomationNodeEdge
	for _, e := range edges {
		es = append(es, models.AutomationNodeEdge{SourceID: f.ids[e[0]], TargetID: f.ids[e[1]], Label: e[2]})
	}
	f.auto = &models.Automation{Base: models.Base{ID: uuid.NewString()}, TeamID: uuid.NewString(), Name: "Flow", IsActive: true, ReentryMode: reentry, Nodes: nodes, Edges: es}
	require.NoError(t, db.Create(f.auto).Error)
	f.contact = &models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: f.auto.TeamID, ListID: uuid.NewString(), Email: "reader@example.com", Status: models.SubscriberStatusActive}
	require.NoError(t, db.Create(f.contact).Error)
	return f
}

func node(name string, kind models.NodeType) models.AutomationNode {
	return models.AutomationNode{Base: models.Base{ID: name}, Type: kind, Data: datatypes.JSON(`{}`)}
}

func waitNode(data string) models.AutomationNode {
	return models.AutomationNode{Base: models.Base{ID: "wait"}, Type: models.NodeTypeWaitUntil, Data: datatypes.JSON(data)}
}

// waitFlow: START → EMAIL → WAIT_UNTIL —happened→ yes(EXIT) / —timeout→ no(EXIT)
func waitFlow(t *testing.T, data string) *flowFixture {
	return newFlow(t, "once",
		[]models.AutomationNode{node("start", models.NodeTypeStart), node("email", models.NodeTypeEmail), waitNode(data), node("yes", models.NodeTypeExit), node("no", models.NodeTypeExit)},
		[][3]string{{"start", "email", ""}, {"email", "wait", ""}, {"wait", "yes", "happened"}, {"wait", "no", "timeout"}})
}

func (f *flowFixture) start(data map[string]interface{}) string {
	id := uuid.NewString()
	if data == nil {
		data = map[string]interface{}{}
	}
	data["_execution_id"] = id
	require.NoError(f.t, f.engine.Execute(context.Background(), f.auto.ID, f.contact.ID, data, ""))
	return id
}

// resume runs a queued continuation the way the task handler does.
func (f *flowFixture) resume(task tasks.AutomationExecuteTask) {
	require.NoError(f.t, f.engine.Execute(context.Background(), task.AutomationID, task.ContactID, map[string]interface{}{"_execution_id": task.ExecutionID}, task.CurrentNodeID))
}

func (f *flowFixture) run(id string) models.AutomationExecution {
	var run models.AutomationExecution
	require.NoError(f.t, f.db.First(&run, "id = ?", id).Error)
	return run
}

func (f *flowFixture) ranStep(run models.AutomationExecution, name string) bool {
	return strings.Contains(string(run.ExecutionLog), f.ids[name])
}

func TestWaitUntilHappenedBeatsTimeout(t *testing.T) {
	f := waitFlow(t, `{"event":"opened","timeout":"72h"}`)
	emailID := uuid.NewString()
	id := f.start(map[string]interface{}{"test_email_id": emailID})
	require.Equal(t, models.ExecutionStatusWaiting, f.run(id).Status)
	require.Len(t, f.queue.queued, 1)
	timeout := f.queue.queued[0]
	require.Equal(t, f.ids["no"], timeout.CurrentNodeID)

	// Opening some other email does not count; opening the previous one does.
	require.NoError(t, f.engine.HandleEvent(context.Background(), "email.opened", &models.EmailTracking{ContactID: f.contact.ID, EmailID: uuid.NewString()}))
	require.Len(t, f.queue.queued, 1)
	require.NoError(t, f.engine.HandleEvent(context.Background(), "email.opened", &models.EmailTracking{ContactID: f.contact.ID, EmailID: emailID}))
	require.Len(t, f.queue.queued, 2)
	happened := f.queue.queued[1]
	require.Equal(t, f.ids["yes"], happened.CurrentNodeID)

	f.resume(happened)
	f.resume(timeout)  // the timeout fires later and must not run
	f.resume(happened) // nor may a duplicate delivery of the winner
	run := f.run(id)
	require.Equal(t, models.ExecutionStatusCompleted, run.Status)
	require.True(t, f.ranStep(run, "yes"))
	require.False(t, f.ranStep(run, "no"))
}

func TestWaitUntilTimesOut(t *testing.T) {
	f := waitFlow(t, `{"event":"clicked","url":"https://example.com/pricing","timeout":"72h"}`)
	id := f.start(nil)
	timeout := f.queue.queued[0]

	// A click on another link is not the one we wait for.
	require.NoError(t, f.engine.HandleEvent(context.Background(), "email.clicked", &models.EmailTracking{ContactID: f.contact.ID, URL: "https://example.com/blog"}))
	require.Len(t, f.queue.queued, 1)

	f.resume(timeout)
	run := f.run(id)
	require.Equal(t, models.ExecutionStatusCompleted, run.Status)
	require.True(t, f.ranStep(run, "no"))

	// The awaited click arriving after the timeout changes nothing.
	require.NoError(t, f.engine.HandleEvent(context.Background(), "email.clicked", &models.EmailTracking{ContactID: f.contact.ID, URL: "https://example.com/pricing"}))
	require.Len(t, f.queue.queued, 1)
	require.False(t, f.ranStep(f.run(id), "yes"))
}

func TestWaitUntilRaceHasOneWinner(t *testing.T) {
	f := waitFlow(t, `{"event":"tag_added","tag":"VIP","timeout":"1h"}`)
	id := f.start(nil)
	// Both continuations load the same waiting run before either claims it.
	first, second := f.run(id), f.run(id)
	won, err := f.engine.claimResume(context.Background(), &first, f.auto, f.ids["yes"])
	require.NoError(t, err)
	require.True(t, won)
	won, err = f.engine.claimResume(context.Background(), &second, f.auto, f.ids["no"])
	require.NoError(t, err)
	require.False(t, won)
	// A continuation for a step that does not follow the wait is ignored.
	third := f.run(id)
	won, err = f.engine.claimResume(context.Background(), &third, f.auto, f.ids["email"])
	require.NoError(t, err)
	require.False(t, won)
}

func TestWaitUntilTagAndFieldAndAppEvent(t *testing.T) {
	for _, c := range []struct {
		data  string
		event string
		send  func(*flowFixture) interface{}
	}{
		{`{"event":"tag_added","tag":"VIP","timeout":"1h"}`, "contact.tag_added", func(f *flowFixture) interface{} {
			// Same shape as models.TagAdded.
			return &struct {
				Contact *models.Contact
				Tag     models.Tag
			}{f.contact, models.Tag{Name: "vip"}}
		}},
		{`{"event":"field_changed","field":"company","value":"Acme","timeout":"1h"}`, "contact.updated", func(f *flowFixture) interface{} {
			require.NoError(t, f.db.Model(f.contact).UpdateColumn("company", "Acme").Error)
			return f.contact
		}},
		{`{"event":"app_event","name":"trial_started","timeout":"1h"}`, "event.received", func(f *flowFixture) interface{} {
			// Same shape as models.ContactEvent.
			return &struct {
				ContactID string `json:"contactId"`
				Name      string `json:"name"`
			}{f.contact.ID, "trial_started"}
		}},
	} {
		f := waitFlow(t, c.data)
		f.start(nil)
		require.NoError(t, f.engine.HandleEvent(context.Background(), c.event, c.send(f)), c.event)
		require.Len(t, f.queue.queued, 2, c.event)
		require.Equal(t, f.ids["yes"], f.queue.queued[1].CurrentNodeID, c.event)
	}
}

// linearFlow: START → WAIT(5m) → EMAIL → EXIT
func linearFlow(t *testing.T, reentry string) *flowFixture {
	return newFlow(t, reentry,
		[]models.AutomationNode{node("start", models.NodeTypeStart), node("wait", models.NodeTypeWait), node("email", models.NodeTypeEmail), node("exit", models.NodeTypeExit)},
		[][3]string{{"start", "wait", ""}, {"wait", "email", ""}, {"email", "exit", ""}})
}

func (f *flowFixture) runs() int64 {
	var n int64
	require.NoError(f.t, f.db.Model(&models.AutomationExecution{}).Where("automation_id = ? AND contact_id = ?", f.auto.ID, f.contact.ID).Count(&n).Error)
	return n
}

func TestReentryRules(t *testing.T) {
	t.Run("once ever", func(t *testing.T) {
		f := linearFlow(t, "once")
		f.start(nil)
		f.start(nil) // while waiting
		require.EqualValues(t, 1, f.runs())
		f.resume(f.queue.queued[0])
		f.start(nil) // after finishing
		require.EqualValues(t, 1, f.runs())
	})
	t.Run("again after it finishes", func(t *testing.T) {
		f := linearFlow(t, "after_finish")
		f.start(nil)
		f.start(nil)
		require.EqualValues(t, 1, f.runs(), "never two runs at once")
		f.resume(f.queue.queued[0])
		f.start(nil)
		require.EqualValues(t, 2, f.runs())
	})
	t.Run("again after N days", func(t *testing.T) {
		f := linearFlow(t, "after_days")
		require.NoError(t, f.db.Model(f.auto).Update("reentry_days", 7).Error)
		first := f.start(nil)
		f.resume(f.queue.queued[0])
		f.start(nil)
		require.EqualValues(t, 1, f.runs(), "too soon")
		require.NoError(t, f.db.Model(&models.AutomationExecution{}).Where("id = ?", first).Update("started_at", time.Now().AddDate(0, 0, -8)).Error)
		f.start(nil)
		require.EqualValues(t, 2, f.runs())
	})
}

func TestUnsubscribeMidWaitExits(t *testing.T) {
	f := linearFlow(t, "once")
	id := f.start(nil)
	require.NoError(t, f.db.Model(f.contact).UpdateColumn("status", models.SubscriberStatusUnsubscribed).Error)
	require.NoError(t, f.engine.HandleEvent(context.Background(), "email.unsubscribed", &models.EmailEvent{ContactID: f.contact.ID}))
	run := f.run(id)
	require.Equal(t, models.ExecutionStatusExited, run.Status)
	require.Equal(t, "unsubscribed", run.ExitReason)
	f.resume(f.queue.queued[0])
	require.Equal(t, models.ExecutionStatusExited, f.run(id).Status)
	require.Zero(t, f.email.sent)
}

func TestNoEmailAfterMissedUnsubscribe(t *testing.T) {
	for reason, optOut := range map[string]func(*flowFixture){
		"unsubscribed": func(f *flowFixture) {
			require.NoError(t, f.db.Model(f.contact).UpdateColumn("status", models.SubscriberStatusUnsubscribed).Error)
		},
		"suppressed": func(f *flowFixture) {
			require.NoError(t, f.db.Create(&sending.Suppression{TeamID: f.auto.TeamID, Address: "reader@example.com", Reason: "Bounce"}).Error)
		},
		"contact_deleted": func(f *flowFixture) {
			require.NoError(t, f.db.Model(f.contact).UpdateColumn("is_deleted", true).Error)
		},
	} {
		f := linearFlow(t, "once")
		id := f.start(nil)
		optOut(f) // no event reaches the engine
		f.resume(f.queue.queued[0])
		run := f.run(id)
		require.Equal(t, models.ExecutionStatusExited, run.Status, reason)
		require.Equal(t, reason, run.ExitReason)
		require.Zero(t, f.email.sent, reason)
	}
}

func TestReachingGoalExits(t *testing.T) {
	f := waitFlow(t, `{"event":"opened","timeout":"72h"}`)
	require.NoError(t, f.db.Create(&models.AutomationGoal{AutomationID: f.auto.ID, Name: "Clicked", GoalType: models.GoalTypeEmailClick, IsActive: true}).Error)
	id := f.start(nil)
	emailID := uuid.NewString()
	key := "automation:" + id + ":" + f.ids["email"]
	require.NoError(t, f.db.Create(&models.Email{Base: models.Base{ID: emailID}, DeliveryKey: &key, TeamID: f.auto.TeamID, ContactID: f.contact.ID, SMTPConfigID: uuid.NewString(), CategoryID: uuid.NewString()}).Error)
	require.NoError(t, f.engine.HandleEvent(context.Background(), "email.clicked", &models.EmailTracking{ContactID: f.contact.ID, EmailID: emailID}))
	run := f.run(id)
	require.Equal(t, models.ExecutionStatusExited, run.Status)
	require.Equal(t, "goal_reached", run.ExitReason)
}
