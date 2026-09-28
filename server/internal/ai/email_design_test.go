package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func sampleLayout() EmailLayout {
	return EmailLayout{Subject: "Sunday Roast", Preheader: "A slower kind of morning", Background: "#F6EDDA", Accent: "#6D3424", Font: "serif", Sections: []LayoutSection{
		{Background: "#F6EDDA", Color: "#382719", Padding: 32, Weights: []int{2, 1}, Columns: [][]LayoutBlock{{{Type: "heading", Text: "Coffee, with a little perspective.", Size: 40}, {Type: "text", Text: "Hello {{first_name}},\nA little story from the roastery."}}, {{Type: "image", URL: "https://images.unsplash.com/photo-1495474472287-4d71bcdd2085?w=1200", Alt: "Fresh coffee"}}}},
		{Background: "#FFFFFF", Color: "#382719", Padding: 24, Columns: [][]LayoutBlock{{{Type: "heading", Text: "Your next good morning", Size: 28}, {Type: "button", Text: "Find your roast", URL: "https://example.com/coffee"}}}},
	}}
}
func encodedLayout(l EmailLayout) string { b, _ := json.Marshal(l); return string(b) }
func TestCompileEditableDesign(t *testing.T) {
	d, err := compileLayout(encodedLayout(sampleLayout()))
	if err != nil {
		t.Fatal(err)
	}
	if d.Design["schemaVersion"] != 18 || !strings.Contains(d.PreviewHTML, "Built with Xem") || !strings.Contains(d.Body, "{{first_name}}") {
		t.Fatal("missing schema, footer, or merge field")
	}
	body := d.Design["body"].(map[string]any)
	rows := body["rows"].([]any)
	if len(rows) != 3 {
		t.Fatalf("rows=%d", len(rows))
	}
	seen := map[string]bool{}
	types := map[string]bool{}
	for _, r := range rows {
		row := r.(map[string]any)
		for _, c := range row["columns"].([]any) {
			for _, v := range c.(map[string]any)["contents"].([]any) {
				b := v.(map[string]any)
				id := b["id"].(string)
				if seen[id] {
					t.Fatal("duplicate ID")
				}
				seen[id] = true
				types[b["type"].(string)] = true
			}
		}
	}
	for _, kind := range []string{"heading", "text", "image", "button"} {
		if !types[kind] {
			t.Fatalf("missing %s", kind)
		}
	}
}
func TestDesignOutputEscapesMarkup(t *testing.T) {
	l := sampleLayout()
	l.Sections[0].Columns[0][0].Text = `<img src=x onerror=alert(1)> & "quoted"`
	d, err := compileLayout(encodedLayout(l))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(d.PreviewHTML, `<img src=x`) || !strings.Contains(d.PreviewHTML, "&lt;img") {
		t.Fatal("unsafe model markup")
	}
	value := d.Design["body"].(map[string]any)["rows"].([]any)[0].(map[string]any)["columns"].([]any)[0].(map[string]any)["contents"].([]any)[0].(map[string]any)["values"].(map[string]any)["text"].(string)
	if !strings.Contains(value, "&lt;img") {
		t.Fatal("Unlayer content was not escaped")
	}
}
func TestDesignRejectsUnsafeAndUnboundedOutput(t *testing.T) {
	cases := map[string]func(*EmailLayout){
		"javascript":      func(l *EmailLayout) { l.Sections[1].Columns[0][1].URL = "javascript:alert(1)" },
		"data image":      func(l *EmailLayout) { l.Sections[0].Columns[1][0].URL = "data:image/svg+xml,<svg/>" },
		"css injection":   func(l *EmailLayout) { l.Accent = "#fff;background:url(https://example.com)" },
		"raw HTML":        func(l *EmailLayout) { l.Sections[0].Columns[0][0].Type = "html" },
		"oversized text":  func(l *EmailLayout) { l.Sections[0].Columns[0][0].Text = strings.Repeat("x", 5001) },
		"invalid weights": func(l *EmailLayout) { l.Sections[0].Weights = []int{0, 2} },
		"empty columns":   func(l *EmailLayout) { l.Sections[0].Columns = nil },
		"no sections":     func(l *EmailLayout) { l.Sections = nil },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			l := sampleLayout()
			mutate(&l)
			if _, err := compileLayout(encodedLayout(l)); err == nil {
				t.Fatal("accepted unsafe design")
			}
		})
	}
}
func TestWriterDesignModeCallsConfiguredProviderAndCompiles(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var p map[string]any
		json.NewDecoder(r.Body).Decode(&p)
		if p["model"] != "gpt-4.1" || p["max_tokens"] != float64(7000) {
			t.Error("incorrect model or design token limit")
		}
		messages := p["messages"].([]any)
		if !strings.Contains(messages[0].(map[string]any)["content"].(string), "art director") {
			t.Error("text prompt used for a design")
		}
		json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"content": encodedLayout(sampleLayout())}, "finish_reason": "stop"}}})
	}))
	defer server.Close()
	writer := NewEmailWriter(server.URL, "", "gpt-4.1")
	writer.Client = server.Client()
	d, err := writer.Draft(context.Background(), EmailDraftRequest{Instruction: "A beautiful coffee newsletter", Format: "design"})
	if err != nil {
		t.Fatal(err)
	}
	if d.Design == nil {
		t.Fatal("expected native design")
	}
}
func TestDraftImageValidation(t *testing.T) {
	if !validateImages([]DraftImage{{URL: "https://example.com/image.png", Alt: "A photo"}}) {
		t.Fatal("rejected https image")
	}
	for _, u := range []string{"javascript:alert(1)", "data:image/png;base64,x", "http://example.com/x", "https://user:pass@example.com/x"} {
		if validateImages([]DraftImage{{URL: u}}) {
			t.Fatalf("accepted %s", u)
		}
	}
}
