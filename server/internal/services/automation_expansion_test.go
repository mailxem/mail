package services

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"kori/internal/models"
	"testing"
)

func TestExpandedWorkflowValidation(t *testing.T) {
	db := setupTestDB(t)
	s := NewAutomationService(db)
	require.NoError(t, db.AutoMigrate(&models.Contact{}, &models.MailingList{}))
	team := uuid.NewString()
	list := models.MailingList{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "Audience"}
	require.NoError(t, db.Create(&list).Error)
	for _, tc := range []struct {
		kind, data string
		valid      bool
	}{
		{"TAG", `{"action":"add","tags":["trial"]}`, true},
		{"TAG", `{"action":"add","tags":[""]}`, false},
		{"UPDATE_SUBSCRIBER", `{"fields":{"company":"Acme"}}`, true},
		{"UPDATE_SUBSCRIBER", `{"fields":{"status":"ACTIVE"}}`, false},
		{"UPDATE_SUBSCRIBER", `{"fields":{"company":null}}`, false},
		{"SET_VARIABLE", `{"variable":"workflow_stage","value":"trial"}`, true},
		{"SET_VARIABLE", `{"variable":"contact_email","value":"injected@example.com"}`, false},
		{"WAIT", `{"until":"2027-01-01T10:00:00Z"}`, true},
		{"WAIT", `{"until":"2027-01-01T10:00:00Z","duration":"1h"}`, false},
		{"WAIT", `{"duration":"-1h"}`, false},
		{"ADD_TO_LIST", `{"listId":"` + list.ID + `"}`, true},
		{"ADD_TO_LIST", `{"listId":"` + uuid.NewString() + `"}`, false},
	} {
		t.Run(tc.kind+tc.data, func(t *testing.T) {
			a := &models.Automation{TeamID: team, Nodes: []models.AutomationNode{{Base: models.Base{ID: "step"}, Type: models.NodeType(tc.kind), Data: []byte(tc.data)}, {Base: models.Base{ID: "exit"}, Type: models.NodeTypeExit}}, Edges: []models.AutomationNodeEdge{{SourceID: "step", TargetID: "exit"}}}
			err := s.ValidateConfiguration(a)
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
	a := &models.Automation{TeamID: team, Nodes: []models.AutomationNode{{Base: models.Base{ID: "split"}, Type: "PERCENTAGE_SPLIT", Data: []byte(`{"percentage":50}`)}, {Base: models.Base{ID: "a"}, Type: models.NodeTypeExit}, {Base: models.Base{ID: "b"}, Type: models.NodeTypeExit}}, Edges: []models.AutomationNodeEdge{{SourceID: "split", TargetID: "a", Label: "A"}, {SourceID: "split", TargetID: "b", Label: "B"}}}
	require.NoError(t, s.ValidateConfiguration(a))
	a.Edges[1].Label = "A"
	require.Error(t, s.ValidateConfiguration(a))
}
