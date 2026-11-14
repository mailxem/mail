package utils

import (
	"encoding/json"
	"strings"
	"testing"
)

func mark(s Section, inner string) string {
	meta, _ := json.Marshal(s)
	return "<!--xem:section " + string(meta) + "-->" + inner + "<!--/xem:section-->"
}

func TestSectionVisibility(t *testing.T) {
	vip := &SectionReader{Tags: map[string]bool{"vip": true}, Segments: map[string]bool{}}
	inSegment := &SectionReader{Tags: map[string]bool{}, Segments: map[string]bool{"6f1c3a52-2c1d-4a8f-9a57-0d0e3b1c2a11": true}}
	nobody := &SectionReader{}
	only := Section{Audience: "only", Tags: []string{"VIP"}, Segments: []string{"6f1c3a52-2c1d-4a8f-9a57-0d0e3b1c2a11"}}
	except := Section{Audience: "except", Tags: []string{"vip"}}
	cases := []struct {
		name    string
		s       Section
		channel string
		r       *SectionReader
		want    bool
	}{
		{"everyone, email", Section{}, SectionEmail, nil, true},
		{"everyone, web", Section{}, SectionWeb, nobody, true},
		{"email only, on the web", Section{Show: SectionEmail}, SectionWeb, nobody, false},
		{"web only, in an email", Section{Show: SectionWeb}, SectionEmail, vip, false},
		{"web only, on the web", Section{Show: SectionWeb}, SectionWeb, nobody, true},
		{"only tagged, tag in another case", only, SectionEmail, vip, true},
		{"only tagged or in segment, in segment", only, SectionEmail, inSegment, true},
		{"only tagged, untagged", only, SectionEmail, nobody, false},
		{"except tagged, tagged", except, SectionEmail, vip, false},
		{"except tagged, untagged", except, SectionEmail, nobody, true},
		{"only, reader never looked up", only, SectionEmail, nil, false},
		{"except, reader never looked up", except, SectionEmail, nil, false},
		{"only, web visitor", only, SectionWeb, nobody, false},
		{"except, web visitor", except, SectionWeb, nobody, true},
	}
	for _, c := range cases {
		if got := c.s.Visible(c.channel, c.r); got != c.want {
			t.Errorf("%s: visible = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestEmailAndWebSectionsDecideOnTheServer(t *testing.T) {
	body := "<p>Hi</p>" +
		mark(Section{Name: "VIP", Audience: "only", Tags: []string{"vip"}}, "<p>VIP offer</p>") +
		mark(Section{Name: "Others", Audience: "except", Tags: []string{"vip"}}, "<p>Join VIP</p>") +
		mark(Section{Name: "Web", Show: SectionWeb}, "<p>Read online</p>") +
		"<!--xem:section {not json}--><p>Broken</p><!--/xem:section-->"
	sections := ParseSections(body)
	if len(sections) != 4 || sections[3].Name != "Unreadable section" {
		t.Fatalf("parsed %+v", sections)
	}
	vip := &SectionReader{Tags: map[string]bool{"vip": true}}
	plan := SectionPlan(sections, vip)
	if plan != "0" {
		t.Fatalf("plan = %q, want 0", plan)
	}
	got := EmailSections(body, map[string]string{SectionsKey: plan})
	if got != "<p>Hi</p><p>VIP offer</p>" {
		t.Fatalf("vip email = %q", got)
	}
	// No plan: only sections for everyone, and never one limited to some readers.
	if got := EmailSections(body, nil); got != "<p>Hi</p>" {
		t.Fatalf("unplanned email = %q", got)
	}
	if got := WebSections(body); got != "<p>Hi</p><p>Join VIP</p><p>Read online</p>" {
		t.Fatalf("web = %q", got)
	}
	// Markers carry tag names; no rendered email may keep them.
	if got := fillVariables(body, map[string]string{SectionsKey: "0,1"}, "id", nil, false, false); strings.Contains(got, "xem:section") || !strings.Contains(got, "Join VIP") {
		t.Fatalf("filled = %q", got)
	}
}

func TestHasContent(t *testing.T) {
	for body, want := range map[string]bool{
		"": false,
		"<html><head><title>T</title><style>p{}</style></head><body><table><tr><td>&nbsp;</td></tr></table></body></html>": false,
		`<div><img src="https://x/p.gif" width="1" height="1"></div>`:                                                      false,
		"<!--[if mso]><p>Only Outlook</p><![endif]-->":                                                                     false,
		`<img src="https://x/hero.png" width="600">`:                                                                       true,
		"<p>Hello</p>": true,
	} {
		if got := HasContent(body); got != want {
			t.Errorf("HasContent(%q) = %v, want %v", body, got, want)
		}
	}
}

const editorHTML = `<html><body><table id="u_body"><tr><td>` +
	`<div class="u-row-container" style="padding:0"><div class="u-row"><div><p>Header</p></div></div></div>` +
	`<div class="u-row-container" style="padding:0"><div class="u-row"><div><div><p>Offer</p></div></div></div></div>` +
	`<div class="u-row-container" style="padding:0"><div class="u-row"><p>Footer</p></div></div>` +
	`</td></tr></table></body></html>`

func design(sections string) []byte {
	return []byte(`{"body":{"rows":[{"values":{"_meta":{"htmlID":"u_row_1"}}},{"values":{"_meta":{"htmlID":"u_row_2"}}},{"values":{"_meta":{"htmlID":"u_row_3"}}}]},"xemSections":` + sections + `}`)
}

func TestMarkSectionsWrapsEditorRows(t *testing.T) {
	body, err := MarkSections(editorHTML, design(`{"u_row_2":{"name":"Offer","audience":"only","tags":["vip"]},"u_row_3":{"show":"web"}}`))
	if err != nil {
		t.Fatal(err)
	}
	sections := ParseSections(body)
	if len(sections) != 2 || sections[0].ID != "u_row_2" || sections[1].Name != "Row 3" {
		t.Fatalf("sections = %+v", sections)
	}
	got := EmailSections(body, nil)
	if !strings.Contains(got, "Header") || strings.Contains(got, "Offer") || strings.Contains(got, "Footer") || strings.Count(got, "<div") != strings.Count(got, "</div>") {
		t.Fatalf("email = %q", got)
	}
	// Marking again replaces the markers instead of nesting them.
	again, err := MarkSections(body, design(`{"u_row_1":{"audience":"except","tags":["vip"]}}`))
	if err != nil || len(ParseSections(again)) != 1 || ParseSections(again)[0].ID != "u_row_1" {
		t.Fatalf("again = %q, %v", again, err)
	}
	// Rows are found by ID when the HTML carries them.
	withIDs := strings.Replace(editorHTML, `<div class="u-row-container" style="padding:0"><div class="u-row"><p>Footer`, `<div id="u_row_3" class="u-row-container"><div class="u-row"><p>Footer`, 1)
	withIDs = strings.Replace(withIDs, `<div class="u-row-container" style="padding:0"><div class="u-row"><div><p>Header</p></div></div></div>`, "", 1)
	body, err = MarkSections(withIDs, design(`{"u_row_3":{"audience":"only","tags":["vip"]}}`))
	if err != nil || strings.Contains(EmailSections(body, nil), "Footer") || !strings.Contains(EmailSections(body, nil), "Offer") {
		t.Fatalf("by id = %q, %v", body, err)
	}
	if _, err := MarkSections(withIDs, design(`{"u_row_1":{"audience":"only","tags":["vip"]}}`)); err == nil {
		t.Fatal("a row missing from the HTML must not be guessed")
	}
	if _, err := MarkSections(editorHTML, design(`{"u_row_1":{"audience":"only"}}`)); err == nil {
		t.Fatal("a section limited to nobody named must be refused")
	}
	rows, err := EditorRows(editorHTML, design(`{}`))
	if err != nil || len(rows) != 3 || rows[1].Text != "Offer" {
		t.Fatalf("rows = %+v, %v", rows, err)
	}
}
