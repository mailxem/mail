package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDraftSlotsKeepsOnlySafeCopyForKnownSlots(t *testing.T) {
	answer := `{"subject":"Field Notes Plus","preheader":"Deep dives, weekly.","copy":{
		"h1":"Field Notes Plus",
		"t1":"One long essay a week, for [price] a month.",
		"b1":"Read the first essay",
		"b2":"This button text is far too long to be a button and should be rejected by the server",
		"t2":"<script>alert(1)</script>",
		"ghost":"not a slot"},
		"remove":["r2","r3","nope"]}`
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"content": answer}, "finish_reason": "stop"}}})
	}))
	defer server.Close()
	writer := NewEmailWriter(server.URL+"/v1", "", "gpt-4.1")
	writer.Client = server.Client()
	draft, err := writer.Draft(context.Background(), EmailDraftRequest{
		Instruction: "Announce a paid tier",
		Format:      "slots",
		Reference:   "Field Notes: an editorial newsletter",
		Slots:       []CopySlot{{"h1", "heading", "Old"}, {"t1", "text", "Old"}, {"b1", "button", "Old"}, {"b2", "button", "Old"}, {"t2", "text", "Old"}},
		Rows:        []SlotRow{{"r1", []string{"h1"}}, {"r2", []string{"t1"}}, {"r3", []string{"b1"}}, {"r4", []string{"b2", "t2"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"h1": "Field Notes Plus", "t1": "One long essay a week, for [price] a month.", "b1": "Read the first essay"}
	if len(draft.Copy) != len(want) {
		t.Fatalf("copy = %v", draft.Copy)
	}
	for id, text := range want {
		if draft.Copy[id] != text {
			t.Fatalf("copy[%s] = %q", id, draft.Copy[id])
		}
	}
	if len(draft.Remove) != 2 || draft.Remove[0] != "r2" || draft.Remove[1] != "r3" {
		t.Fatalf("remove = %v, want at most half the rows and only known ids", draft.Remove)
	}
}

func TestDraftSlotsNeedsAReference(t *testing.T) {
	writer := NewEmailWriter("https://example.invalid/v1", "", "gpt-4.1")
	if _, err := writer.Draft(context.Background(), EmailDraftRequest{Instruction: "Announce a paid tier", Format: "slots"}); err == nil {
		t.Fatal("expected an error without slots")
	}
}
