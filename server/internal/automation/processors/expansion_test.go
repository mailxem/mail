package processors

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"kori/internal/automation"
	"kori/internal/models"
	"kori/internal/tasks"
	"kori/internal/workflowconfig"
	"testing"
	"time"
)

func configuredNode(kind string, data string) *models.AutomationNode {
	return &models.AutomationNode{Base: models.Base{ID: uuid.NewString()}, Type: models.NodeType(kind), Data: []byte(data)}
}
func connect(t *testing.T, db *gorm.DB, ctx *automation.ExecutionContext, node *models.AutomationNode, labels ...string) {
	for _, label := range labels {
		require.NoError(t, db.Create(&models.AutomationNodeEdge{AutomationID: ctx.AutomationID, SourceID: node.ID, TargetID: label + "-next", Label: label}).Error)
	}
}
func TestExpandedActions(t *testing.T) {
	db := setupDB(t)
	ctx := createContext(t, db)
	node := configuredNode("SET_VARIABLE", `{"variable":"workflow_stage","value":"qualified"}`)
	connect(t, db, ctx, node, "")
	result, err := NewSetVariableProcessor(db).Process(ctx, node)
	require.NoError(t, err)
	require.Equal(t, "qualified", result.UpdateVars["workflow_stage"])
	node = configuredNode("PERCENTAGE_SPLIT", `{"percentage":35}`)
	connect(t, db, ctx, node, "A", "B")
	split := NewPercentageSplitProcessor(db)
	first, err := split.Process(ctx, node)
	require.NoError(t, err)
	for i := 0; i < 10; i++ {
		retry, err := NewPercentageSplitProcessor(db).Process(ctx, node)
		require.NoError(t, err)
		require.Equal(t, first.NextNodeIDs, retry.NextNodeIDs)
	}
	node = configuredNode("UPDATE_SUBSCRIBER", `{"fields":{"company":"Acme","linkedin":"https://example.com/profile"}}`)
	connect(t, db, ctx, node, "")
	_, err = NewUpdateSubscriberProcessor(db).Process(ctx, node)
	require.NoError(t, err)
	require.Equal(t, "Acme", ctx.Contact.Company)
	require.Equal(t, "https://example.com/profile", ctx.Contact.LinkedIn)
	require.Equal(t, "Acme", ctx.Variables["contact_company"])
	node = configuredNode("UPDATE_SUBSCRIBER", `{"fields":{"status":"ACTIVE"}}`)
	_, err = NewUpdateSubscriberProcessor(db).Process(ctx, node)
	require.Error(t, err)
	node = configuredNode("TAG", `{"action":"add","tags":["qualified"]}`)
	connect(t, db, ctx, node, "")
	for i := 0; i < 2; i++ {
		_, err = NewTagProcessor(db).Process(ctx, node)
		require.NoError(t, err)
	}
	require.EqualValues(t, 1, db.Model(ctx.Contact).Association("Tags").Count())
	node.Data = []byte(`{"action":"remove","tags":["qualified"]}`)
	_, err = NewTagProcessor(db).Process(ctx, node)
	require.NoError(t, err)
	require.EqualValues(t, 0, db.Model(ctx.Contact).Association("Tags").Count())
}
func TestMoveListIsAtomicAndRetrySafe(t *testing.T) {
	db := setupDB(t)
	ctx := createContext(t, db)
	require.NoError(t, db.AutoMigrate(&models.MailingList{}))
	old := models.MailingList{Base: models.Base{ID: uuid.NewString()}, TeamID: ctx.TeamID, Name: "Old", SubscribersCount: 1}
	next := models.MailingList{Base: models.Base{ID: uuid.NewString()}, TeamID: ctx.TeamID, Name: "Next"}
	foreign := models.MailingList{Base: models.Base{ID: uuid.NewString()}, TeamID: uuid.NewString(), Name: "Foreign"}
	require.NoError(t, db.Create(&[]models.MailingList{old, next, foreign}).Error)
	require.NoError(t, db.Model(ctx.Contact).Update("list_id", old.ID).Error)
	node := configuredNode("ADD_TO_LIST", `{"listId":"`+foreign.ID+`"}`)
	_, err := NewAddToListProcessor(db).Process(ctx, node)
	require.Error(t, err)
	node.Data = []byte(`{"listId":"` + next.ID + `"}`)
	connect(t, db, ctx, node, "")
	for i := 0; i < 2; i++ {
		_, err = NewAddToListProcessor(db).Process(ctx, node)
		require.NoError(t, err)
	}
	require.NoError(t, db.First(&old, "id = ?", old.ID).Error)
	require.NoError(t, db.First(&next, "id = ?", next.ID).Error)
	require.EqualValues(t, 0, old.SubscribersCount)
	require.EqualValues(t, 1, next.SubscribersCount)
	require.Equal(t, next.ID, ctx.Contact.ListID)
}
func TestRichConditionsAndScheduledWait(t *testing.T) {
	db := setupDB(t)
	ctx := createContext(t, db)
	ctx.Variables["company"] = "Acme"
	ctx.Variables["blank"] = ""
	p := NewConditionProcessor(db)
	for _, tc := range []struct {
		operator, variable, value string
		want                      bool
	}{{"starts_with", "company", "AC", true}, {"ends_with", "company", "ME", true}, {"not_contains", "company", "other", true}, {"empty", "missing", "", true}, {"not_empty", "blank", "", false}} {
		got, err := p.evaluateCondition(Condition{Operator: tc.operator, Variable: tc.variable, Value: tc.value}, ctx)
		require.NoError(t, err)
		require.Equal(t, tc.want, got)
	}
	node := configuredNode("CONDITION", `{"operator":"OR","conditions":[{"variable":"company","operator":"==","value":"Other"},{"variable":"company","operator":"starts_with","value":"Ac"}]}`)
	connect(t, db, ctx, node, "true", "false")
	result, err := p.Process(ctx, node)
	require.NoError(t, err)
	require.Equal(t, []string{"true-next"}, result.NextNodeIDs)
	empty := configuredNode("CONDITION", `{"conditions":[]}`)
	_, err = p.Process(ctx, empty)
	require.Error(t, err)
	when := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	raw, _ := json.Marshal(map[string]string{"until": when})
	node = configuredNode("WAIT", string(raw))
	connect(t, db, ctx, node, "")
	result, err = NewWaitProcessor(db).Process(ctx, node)
	require.NoError(t, err)
	require.InDelta(t, 3600, result.Wait.Seconds(), 2)
	duration, err := workflowconfig.WaitDuration(workflowconfig.Wait{Until: time.Now().Add(-time.Hour).Format(time.RFC3339)}, time.Now())
	require.NoError(t, err)
	require.Zero(t, duration)
}

type resumeQueue struct {
	task  tasks.AutomationExecuteTask
	delay time.Duration
}

func (q *resumeQueue) EnqueueAutomationTask(_ context.Context, task tasks.AutomationExecuteTask, delay time.Duration) error {
	q.task = task
	q.delay = delay
	return nil
}
func TestVariablesAndProfileRulesSurviveWaitResume(t *testing.T) {
	db := setupDB(t)
	ctx := createContext(t, db)
	queue := &resumeQueue{}
	engine := automation.NewEngine(db, queue)
	engine.RegisterProcessor(NewStartProcessor(db))
	engine.RegisterProcessor(NewSetVariableProcessor(db))
	engine.RegisterProcessor(NewWaitProcessor(db))
	engine.RegisterProcessor(NewConditionProcessor(db))
	engine.RegisterProcessor(NewUpdateSubscriberProcessor(db))
	engine.RegisterProcessor(NewExitProcessor(db))
	nodes := []*models.AutomationNode{configuredNode("START", `{}`), configuredNode("SET_VARIABLE", `{"variable":"workflow_stage","value":"qualified"}`), configuredNode("WAIT", `{"duration":"1h"}`), configuredNode("CONDITION", `{"operator":"AND","conditions":[{"variable":"workflow_stage","operator":"==","value":"qualified"},{"variable":"event_source","operator":"==","value":"signup"}]}`), configuredNode("UPDATE_SUBSCRIBER", `{"fields":{"company":"Qualified"}}`), configuredNode("EXIT", `{}`)}
	for _, n := range nodes {
		n.AutomationID = ctx.AutomationID
		require.NoError(t, db.Create(n).Error)
	}
	for i := 0; i < len(nodes)-1; i++ {
		label := ""
		if i == 3 {
			label = "true"
		}
		require.NoError(t, db.Create(&models.AutomationNodeEdge{AutomationID: ctx.AutomationID, SourceID: nodes[i].ID, TargetID: nodes[i+1].ID, Label: label}).Error)
	}
	require.NoError(t, db.Create(&models.AutomationNodeEdge{AutomationID: ctx.AutomationID, SourceID: nodes[3].ID, TargetID: nodes[5].ID, Label: "false"}).Error)
	require.NoError(t, db.Model(&models.Automation{}).Where("id = ?", ctx.AutomationID).Update("is_active", true).Error)
	require.NoError(t, engine.Execute(context.Background(), ctx.AutomationID, ctx.ContactID, map[string]interface{}{"event_source": "signup"}, ""))
	require.Equal(t, time.Hour, queue.delay)
	require.Equal(t, nodes[3].ID, queue.task.CurrentNodeID)
	var execution models.AutomationExecution
	require.NoError(t, db.First(&execution, "id = ?", queue.task.ExecutionID).Error)
	require.Equal(t, models.ExecutionStatusWaiting, execution.Status)
	require.NoError(t, engine.Execute(context.Background(), ctx.AutomationID, ctx.ContactID, map[string]interface{}{"_execution_id": queue.task.ExecutionID}, queue.task.CurrentNodeID))
	require.NoError(t, db.First(&execution, "id = ?", execution.ID).Error)
	require.Equal(t, models.ExecutionStatusCompleted, execution.Status)
	var contact models.Contact
	require.NoError(t, db.First(&contact, "id = ?", ctx.ContactID).Error)
	require.Equal(t, "Qualified", contact.Company)
	require.Contains(t, string(execution.Variables), "workflow_stage")
	require.Contains(t, string(execution.ExecutionLog), "Condition evaluated to true")
}
