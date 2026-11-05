package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEmailWriterUsesConfiguredModelAndRejectsInvalidOutput(t *testing.T) {
	for _, tc := range []struct {
		name, content string
		status        int
		valid         bool
	}{
		{"valid", `{"subject":"Welcome {{firstName}}","body":"Hello {{firstName}}, welcome aboard."}`, 200, true},
		{"malformed", "<script>bad</script>", 200, false},
		{"missing subject", `{"body":"hello"}`, 200, false},
		{"provider failure", "private provider diagnostic", 500, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				if r.URL.Path != "/v1/chat/completions" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				var input struct {
					Model    string    `json:"model"`
					Messages []Message `json:"messages"`
				}
				if json.NewDecoder(r.Body).Decode(&input) != nil || input.Model != "gpt-4.1" || len(input.Messages) != 2 {
					t.Error("invalid provider request")
				}
				w.WriteHeader(tc.status)
				json.NewEncoder(w).Encode(map[string]interface{}{"choices": []interface{}{map[string]interface{}{"message": map[string]string{"content": tc.content}, "finish_reason": "stop"}}})
			}))
			defer server.Close()
			writer := NewEmailWriter(server.URL+"/v1", "test-only", "gpt-4.1")
			writer.Client = server.Client()
			draft, err := writer.Draft(context.Background(), EmailDraftRequest{Instruction: "Write a welcome email"})
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v error=%v", tc.valid, err)
			}
			if !called {
				t.Fatal("provider was not called")
			}
			if tc.valid && !strings.Contains(draft.Body, "{{firstName}}") {
				t.Fatal("merge variable lost")
			}
			if err != nil && strings.Contains(err.Error(), "private provider") {
				t.Fatal("provider diagnostic leaked")
			}
		})
	}
}
func TestEmailWriterRejectsOversizedInputBeforeCallingProvider(t *testing.T) {
	writer := NewEmailWriter("https://example.com/v1", "", "gpt-4.1")
	_, err := writer.Draft(context.Background(), EmailDraftRequest{Instruction: strings.Repeat("x", 4001)})
	if err == nil {
		t.Fatal("expected oversized input rejection")
	}
}

func TestEmailWriterAssemblesStreamedDrafts(t *testing.T) {
	chunk := func(content, finish string) string {
		choice := map[string]interface{}{"delta": map[string]string{"content": content}}
		if finish != "" {
			choice["finish_reason"] = finish
		}
		b, _ := json.Marshal(map[string]interface{}{"choices": []interface{}{choice}})
		return "data: " + string(b) + "\n\n"
	}
	for _, tc := range []struct {
		name   string
		events []string
		valid  bool
	}{
		{"complete", []string{chunk(`{"subject":"Hi {{firstName}}",`, ""), chunk(`"body":"Welcome, {{firstName}}."}`, ""), chunk("", "stop"), "data: [DONE]\n\n"}, true},
		{"cut off by length", []string{chunk(`{"subject":"Hi",`, ""), chunk("", "length")}, false},
		{"connection dropped", []string{chunk(`{"subject":"Hi",`, "")}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var input struct {
					Stream bool `json:"stream"`
				}
				if json.NewDecoder(r.Body).Decode(&input) != nil || !input.Stream {
					t.Error("expected a streaming request")
				}
				w.Header().Set("Content-Type", "text/event-stream")
				for _, e := range tc.events {
					w.Write([]byte(e))
				}
			}))
			defer server.Close()
			writer := NewEmailWriter(server.URL+"/v1", "", "gpt-4.1")
			writer.Client = server.Client()
			draft, err := writer.Draft(context.Background(), EmailDraftRequest{Instruction: "Write a welcome email"})
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v error=%v", tc.valid, err)
			}
			if tc.valid && draft.Body != "Welcome, {{firstName}}." {
				t.Fatalf("body = %q", draft.Body)
			}
		})
	}
}

func TestEmailWriterRetriesOnceAfterProviderError(t *testing.T) {
	calls := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusGatewayTimeout)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"choices": []interface{}{map[string]interface{}{"message": map[string]string{"content": `{"subject":"Hi","body":"Hello."}`}, "finish_reason": "stop"}}})
	}))
	defer server.Close()
	writer := NewEmailWriter(server.URL+"/v1", "", "gpt-4.1")
	writer.Client = server.Client()
	if _, err := writer.Draft(context.Background(), EmailDraftRequest{Instruction: "Write a welcome email"}); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}
