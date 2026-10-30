package marketing

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	publicSlugPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{1,46}[a-z0-9])$`)
	// The space before a tag goes too, so "Hi {{first_name}}," reads "Hi,".
	mergeField = regexp.MustCompile(`[ \t]*\{\{\s*[A-Za-z0-9_.-]+\s*\}\}`)
)

// NormalizePublicSlug lower-cases a web archive address; "" means private (nil).
func NormalizePublicSlug(raw string) (*string, error) {
	slug := strings.ToLower(strings.TrimSpace(raw))
	if slug == "" {
		return nil, nil
	}
	if !publicSlugPattern.MatchString(slug) {
		return nil, fmt.Errorf("Web address must be 3 to 48 lowercase letters, numbers or hyphens")
	}
	return &slug, nil
}

// PublicEditionHTML is an edition as anyone may read it on the web: merge
// fields are blanked so no reader's details, or placeholders, show.
// Tracking and unsubscribe links are added per recipient at send time and are
// never part of the stored edition body.
func PublicEditionHTML(body string) string {
	return mergeField.ReplaceAllString(body, "")
}

// PublicEditionSubject blanks merge fields in a subject the same way.
func PublicEditionSubject(subject string) string {
	return strings.Join(strings.Fields(PublicEditionHTML(subject)), " ")
}
