package utils

import (
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// A section is a part of an email kept for some readers or for one place.
// It is marked in the HTML as
//
//	<!--xem:section {"name":"Offer","audience":"only","tags":["vip"]}-->...<!--/xem:section-->
//
// and decided on the server when rendering, so no reader's mail client ever
// gets a section that isn't theirs. Sections don't nest.
type Section struct {
	ID       string   `json:"id,omitempty"`       // the editor row it wraps
	Name     string   `json:"name"`               // shown to the sender only
	Audience string   `json:"audience,omitempty"` // "only" or "except" the readers below; "" is everyone
	Tags     []string `json:"tags,omitempty"`     // tag names
	Segments []string `json:"segments,omitempty"` // segment IDs
	Show     string   `json:"show,omitempty"`     // "email", "web" or "" for both
}

// SectionsKey is the email data value holding the sections one reader gets,
// as indexes into the content's sections; the campaign decides it per reader.
const SectionsKey = "xem_sections"

// Section channels.
const (
	SectionEmail = "email"
	SectionWeb   = "web"
)

// SectionReader is what a section's audience depends on. Tags are lower-case.
type SectionReader struct{ Tags, Segments map[string]bool }

var sectionMarkup = regexp.MustCompile(`(?s)<!--xem:section\s*(.*?)\s*-->(.*?)<!--/xem:section-->`)

// Restricted reports whether the section is for some readers only.
func (s Section) Restricted() bool { return s.Audience != "" }

// Visible reports whether a reader gets the section in channel. A nil reader
// is one nobody looked up: they get no section limited to some readers.
func (s Section) Visible(channel string, r *SectionReader) bool {
	if s.Show != "" && s.Show != channel {
		return false
	}
	if !s.Restricted() {
		return true
	}
	if r == nil {
		return false
	}
	in := slices.ContainsFunc(s.Tags, func(t string) bool { return r.Tags[strings.ToLower(t)] }) ||
		slices.ContainsFunc(s.Segments, func(id string) bool { return r.Segments[id] })
	return in == (s.Audience == "only")
}

// Validate checks a section the sender wrote.
func (s Section) Validate() error {
	switch {
	case len(s.Name) > 80:
		return fmt.Errorf("Section names are up to 80 characters")
	case s.Show != "" && s.Show != SectionEmail && s.Show != SectionWeb:
		return fmt.Errorf("Show a section in the email, on the website, or both")
	case s.Audience != "" && s.Audience != "only" && s.Audience != "except":
		return fmt.Errorf("A section is for everyone, only some readers, or everyone except some")
	case (s.Audience == "") != (len(s.Tags)+len(s.Segments) == 0):
		return fmt.Errorf("Section %q: pick its readers by tag or segment, or make it for everyone", s.Name)
	case len(s.Tags)+len(s.Segments) > 50:
		return fmt.Errorf("A section can name up to 50 tags and segments")
	}
	for _, id := range s.Segments {
		if uuid.Validate(id) != nil {
			return fmt.Errorf("Section %q: %q isn't a segment ID", s.Name, id)
		}
	}
	return nil
}

func parseSection(raw string) Section {
	var s Section
	if json.Unmarshal([]byte(raw), &s) != nil || s.Validate() != nil {
		// Never guess who an unreadable section is for: nobody gets it.
		return Section{Name: "Unreadable section", Audience: "only"}
	}
	return s
}

// ParseSections lists the sections marked in body, in order.
func ParseSections(body string) []Section {
	var out []Section
	for _, m := range sectionMarkup.FindAllStringSubmatch(body, -1) {
		out = append(out, parseSection(m[1]))
	}
	return out
}

// FilterSections keeps the sections keep accepts, without their markers, and
// drops the rest.
func FilterSections(body string, keep func(i int, s Section) bool) string {
	if !strings.Contains(body, "<!--xem:section") {
		return body
	}
	i := -1
	return sectionMarkup.ReplaceAllStringFunc(body, func(m string) string {
		i++
		parts := sectionMarkup.FindStringSubmatch(m)
		if keep(i, parseSection(parts[1])) {
			return parts[2]
		}
		return ""
	})
}

// WebSections is body as the public website shows it: a visitor is nobody in
// particular, so they get sections for everyone, and those for everyone
// except some readers.
func WebSections(body string) string {
	return FilterSections(body, func(_ int, s Section) bool { return s.Visible(SectionWeb, &SectionReader{}) })
}

// EmailSections is body as one email's reader gets it: the sections the send
// picked for them (data[SectionsKey]) or, when it picked none, those for
// everyone.
func EmailSections(body string, data map[string]string) string {
	picked, ok := data[SectionsKey]
	if !ok {
		return FilterSections(body, func(_ int, s Section) bool { return s.Visible(SectionEmail, nil) })
	}
	keep := map[int]bool{}
	for f := range strings.SplitSeq(picked, ",") {
		if n, err := strconv.Atoi(f); err == nil {
			keep[n] = true
		}
	}
	return FilterSections(body, func(i int, _ Section) bool { return keep[i] })
}

// SectionPlan is the SectionsKey value for a reader of sections.
func SectionPlan(sections []Section, r *SectionReader) string {
	var keep []string
	for i, s := range sections {
		if s.Visible(SectionEmail, r) {
			keep = append(keep, strconv.Itoa(i))
		}
	}
	return strings.Join(keep, ",")
}

var (
	invisibleMarkup = regexp.MustCompile(`(?is)<!--.*?-->|<head\b.*?</head>|<style\b.*?</style>|<script\b.*?</script>|<title\b.*?</title>`)
	anyTag          = regexp.MustCompile(`(?s)<[^>]*>`)
	realImage       = regexp.MustCompile(`(?i)<img\b[^>]*>`)
)

// HasContent reports whether body shows a reader any text or image.
func HasContent(body string) bool {
	body = invisibleMarkup.ReplaceAllString(body, "")
	for _, img := range realImage.FindAllString(body, -1) {
		if !pixelImage.MatchString(img) {
			return true
		}
	}
	return visibleText(body) != ""
}

func visibleText(body string) string {
	text := html.UnescapeString(anyTag.ReplaceAllString(invisibleMarkup.ReplaceAllString(body, ""), " "))
	return strings.Join(strings.Fields(strings.ReplaceAll(text, "\u00a0", " ")), " ")
}

var pixelImage = regexp.MustCompile(`(?i)\s(width|height)\s*=\s*["']?[01](px)?["'\s/>]`)

// SectionAudience knows which contacts have the tags and segments sections
// name, so a send can pick each reader's sections without a query per reader.
type SectionAudience struct {
	tags     map[string][]string // contact -> lower-case tag names
	segments map[string][]string // contact -> segment IDs
}

// LoadSectionAudience loads who in team has the tags and segments sections name.
// ponytail: holds every matching contact ID in memory; page it if one tag or
// segment ever spans millions of contacts.
func LoadSectionAudience(tx *gorm.DB, team string, sections []Section) (*SectionAudience, error) {
	a := &SectionAudience{tags: map[string][]string{}, segments: map[string][]string{}}
	var tags, segments []string
	for _, s := range sections {
		for _, t := range s.Tags {
			tags = append(tags, strings.ToLower(t))
		}
		segments = append(segments, s.Segments...)
	}
	var rows []struct{ ContactID, Name string }
	if len(tags) > 0 {
		if err := tx.Table("contact_tags").Select("contact_tags.contact_id, LOWER(tags.name) AS name").
			Joins("JOIN tags ON tags.id = contact_tags.tag_id").
			Where("tags.team_id = ? AND tags.is_deleted = ? AND LOWER(tags.name) IN ?", team, false, tags).Scan(&rows).Error; err != nil {
			return nil, err
		}
		for _, r := range rows {
			a.tags[r.ContactID] = append(a.tags[r.ContactID], r.Name)
		}
	}
	if len(segments) > 0 {
		rows = nil
		if err := tx.Table("segment_contacts").Select("segment_contacts.contact_id, segment_contacts.segment_id AS name").
			Joins("JOIN segments ON segments.id = segment_contacts.segment_id").
			Where("segments.team_id = ? AND segments.is_deleted = ? AND segments.id IN ?", team, false, segments).Scan(&rows).Error; err != nil {
			return nil, err
		}
		for _, r := range rows {
			a.segments[r.ContactID] = append(a.segments[r.ContactID], r.Name)
		}
	}
	return a, nil
}

// Reader is one contact as sections see them.
func (a *SectionAudience) Reader(contact string) *SectionReader {
	r := &SectionReader{Tags: map[string]bool{}, Segments: map[string]bool{}}
	for _, t := range a.tags[contact] {
		r.Tags[t] = true
	}
	for _, s := range a.segments[contact] {
		r.Segments[s] = true
	}
	return r
}

var (
	rowContainer = regexp.MustCompile(`<div\b[^>]*\bu-row-container\b[^>]*>`)
	divTag       = regexp.MustCompile(`(?i)<div\b|</div\s*>`)
	sectionMark  = regexp.MustCompile(`<!--/?xem:section\b.*?-->`)
	htmlIDAttr   = regexp.MustCompile(`\bid\s*=\s*["']([^"']+)["']`)
)

// EditorSections reads the sections the template editor keeps in a design,
// by row ID, and the design's row IDs in order.
func EditorSections(design []byte) (map[string]Section, []string, error) {
	var d struct {
		Body struct {
			Rows []struct {
				Values struct {
					Meta struct {
						HTMLID string `json:"htmlID"`
					} `json:"_meta"`
				} `json:"values"`
			} `json:"rows"`
		} `json:"body"`
		Sections map[string]Section `json:"xemSections"`
	}
	if err := json.Unmarshal(design, &d); err != nil {
		return nil, nil, err
	}
	rows := make([]string, len(d.Body.Rows))
	for i, r := range d.Body.Rows {
		rows[i] = r.Values.Meta.HTMLID
	}
	return d.Sections, rows, nil
}

// MarkSections marks, in the editor's exported HTML, the rows that design
// gives sections. Markers already there are replaced.
func MarkSections(body string, design []byte) (string, error) {
	sections, rows, err := EditorSections(design)
	if err != nil {
		return "", err
	}
	body = sectionMark.ReplaceAllString(body, "")
	containers := rowContainer.FindAllStringIndex(body, -1)
	type span struct {
		start, end int
		s          Section
	}
	var spans []span
	for i, id := range rows {
		s, ok := sections[id]
		if !ok || (s.Audience == "" && s.Show == "") {
			continue
		}
		s.ID = id
		if s.Name == "" {
			s.Name = "Row " + strconv.Itoa(i+1)
		}
		if err := s.Validate(); err != nil {
			return "", err
		}
		start, end, ok := findRow(body, containers, rows, i)
		if !ok {
			return "", fmt.Errorf("Couldn’t find %s in the template's HTML; save the design again", s.Name)
		}
		spans = append(spans, span{start, end, s})
	}
	for _, sp := range slices.Backward(spans) {
		meta, _ := json.Marshal(sp.s) // escapes < and >, so "-->" can't appear
		body = body[:sp.start] + "<!--xem:section " + string(meta) + "-->" + body[sp.start:sp.end] + "<!--/xem:section-->" + body[sp.end:]
	}
	return body, nil
}

// findRow finds row i of rows in body by its ID, or by position when the HTML
// has one container per row.
func findRow(body string, containers [][]int, rows []string, i int) (start, end int, ok bool) {
	at := slices.IndexFunc(containers, func(c []int) bool {
		m := htmlIDAttr.FindStringSubmatch(body[c[0]:c[1]])
		return m != nil && m[1] == rows[i]
	})
	if at < 0 && len(containers) == len(rows) {
		at = i
	}
	if at < 0 {
		return 0, 0, false
	}
	start, depth := containers[at][0], 0
	for _, t := range divTag.FindAllStringIndex(body[start:], -1) {
		if body[start+t[0]+1] == '/' {
			depth--
		} else {
			depth++
		}
		if depth == 0 {
			return start, start + t[1], true
		}
	}
	return 0, 0, false
}

// EditorRow is a row of a template editor design, for picking sections.
type EditorRow struct {
	ID   string `json:"id"`
	Text string `json:"text"` // the start of what it says
}

// EditorRows lists a design's rows with the start of each one's text.
func EditorRows(body string, design []byte) ([]EditorRow, error) {
	_, rows, err := EditorSections(design)
	if err != nil {
		return nil, err
	}
	containers := rowContainer.FindAllStringIndex(body, -1)
	out := make([]EditorRow, len(rows))
	for i, id := range rows {
		out[i].ID = id
		if start, end, ok := findRow(body, containers, rows, i); ok {
			text := []rune(visibleText(body[start:end]))
			out[i].Text = string(text[:min(len(text), 80)])
		}
	}
	return out, nil
}
