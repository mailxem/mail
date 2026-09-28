package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"kori/internal/marketing"
)

func formWriterDraft() FormDraft {
	return FormDraft{
		Name: "Product feedback", Description: "Tell us what could be better.", ButtonText: "Share feedback", SuccessMessage: "Thanks for your feedback.",
		Definition: marketing.FormDefinition{SchemaVersion: 1, Consent: marketing.FormConsent{Mode: "optional", Label: "Send me product updates. I can unsubscribe at any time."}, Pages: []marketing.FormPage{{ID: "details", Title: "Your feedback", Fields: []marketing.FormQuestion{{Key: "email", Label: "Email", Type: "EMAIL", Required: true}, {Key: "feedback", Label: "Feedback", Type: "TEXTAREA"}}}}},
		Emails:     []JourneyEmailDraft{{Name: "A thank-you", Subject: "Thanks for sharing", Body: "Hello {{first_name}}, thanks for sharing {{form_feedback}}."}},
	}
}

func TestFormWriterValidatesProviderProposal(t *testing.T) {
	for _, scenario := range []string{"valid", "unknown-field", "trailing-json", "missing-email", "invalid-delay", "subject-newline", "redirect", "truncated", "provider-failure", "oversized"} {
		t.Run(scenario, func(t *testing.T) {
			draft := formWriterDraft()
			switch scenario {
			case "missing-email":
				draft.Definition.Pages[0].Fields = draft.Definition.Pages[0].Fields[1:]
			case "invalid-delay":
				draft.Emails[0].DelayHours = -1
			case "subject-newline":
				draft.Emails[0].Subject = "Hello\r\nBcc: attacker@example.com"
			case "redirect":
				draft.Definition.SuccessRedirectURL = "https://unrequested.example.com"
			}
			raw, err := json.Marshal(draft)
			require.NoError(t, err)
			content := string(raw)
			if scenario == "unknown-field" {
				content = strings.TrimSuffix(content, "}") + `,"publish":true}`
			}
			if scenario == "trailing-json" {
				content += `{}`
			}
			if scenario == "oversized" {
				content = strings.Repeat("x", 128*1024+1)
			}
			called := false
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				require.Equal(t, "/v1/chat/completions", r.URL.Path)
				require.Equal(t, "Bearer test-only", r.Header.Get("Authorization"))
				var request struct {
					Model    string    `json:"model"`
					Messages []Message `json:"messages"`
				}
				require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
				require.Equal(t, "form-test-model", request.Model)
				require.Len(t, request.Messages, 2)
				require.Equal(t, "user", request.Messages[1].Role)
				require.Equal(t, "Build a feedback form", request.Messages[1].Content)
				if scenario == "provider-failure" {
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte("private provider diagnostics"))
					return
				}
				finish := "stop"
				if scenario == "truncated" {
					finish = "length"
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": Message{Role: "assistant", Content: content}, "finish_reason": finish}}})
			}))
			defer server.Close()
			writer := NewEmailWriter(server.URL+"/v1", "test-only", "form-test-model")
			writer.Client = server.Client()
			proposal, err := writer.DraftForm(context.Background(), FormDraftRequest{Instruction: "Build a feedback form"})
			require.True(t, called)
			if scenario == "valid" {
				require.NoError(t, err)
				require.Equal(t, draft, *proposal)
			} else {
				require.Error(t, err)
				require.Nil(t, proposal)
				require.NotContains(t, err.Error(), "private provider")
			}
		})
	}
}

func TestFormWriterRejectsInvalidRequestBeforeProvider(t *testing.T) {
	for _, instruction := range []string{"", "ab", strings.Repeat("x", 4001)} {
		writer := NewEmailWriter("https://unused.example.com/v1", "", "test")
		writer.Client = nil // Invalid input must be rejected before any network call.
		_, err := writer.DraftForm(context.Background(), FormDraftRequest{Instruction: instruction})
		require.Error(t, err)
	}
	for _, endpoint := range []string{"http://example.com/v1", "https://user:secret@example.com/v1", "https://example.com/v1?secret=1", "https://example.com/v1#fragment"} {
		writer := NewEmailWriter(endpoint, "", "test")
		writer.Client = nil
		_, err := writer.DraftForm(context.Background(), FormDraftRequest{Instruction: "Build a feedback form"})
		require.Error(t, err)
	}
}
