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

// reservedSlugs are xem.email's own pages and paths, which a newsletter
// website at xem.email/{slug} must never shadow.
var reservedSlugs = map[string]bool{
	"about": true, "admin": true, "affiliates": true, "api": true, "app": true, "auth": true,
	"blog": true, "brand": true, "compare": true, "data-deletion": true, "docs": true,
	"feed": true, "fonts": true, "google-mail": true, "help": true, "images": true,
	"login": true, "logos": true, "mcp": true, "migrate": true, "pricing": true,
	"privacy": true, "signup": true, "sitemap": true, "status": true, "support": true,
	"templates": true, "terms": true, "tools": true, "videos": true, "www": true, "xem": true,
}

// NormalizePublicSlug lower-cases a web archive address; "" means private (nil).
func NormalizePublicSlug(raw string) (*string, error) {
	slug := strings.ToLower(strings.TrimSpace(raw))
	if slug == "" {
		return nil, nil
	}
	if !publicSlugPattern.MatchString(slug) {
		return nil, fmt.Errorf("Web address must be 3 to 48 lowercase letters, numbers or hyphens")
	}
	if reservedSlugs[slug] {
		return nil, fmt.Errorf("That web address is reserved; choose another")
	}
	return &slug, nil
}

// SiteThemes are the looks a newsletter website can take.
var SiteThemes = map[string]bool{"paper": true, "broadsheet": true, "midnight": true}

// ValidSiteLogo accepts "" or an HTTPS image address.
func ValidSiteLogo(raw string) bool {
	return raw == "" || len(raw) <= 500 && strings.HasPrefix(raw, "https://") && !strings.ContainsAny(raw, "\"<> ")
}

var editionImage = regexp.MustCompile(`(?i)<img\b[^>]*>`)
var imageSrc = regexp.MustCompile(`(?i)\ssrc\s*=\s*["'](https://[^"']+)["']`)
var pixelSize = regexp.MustCompile(`(?i)\s(width|height)\s*=\s*["']?[01]["'\s>]`)

// EditionCover is an edition's first real image, for its card on the
// newsletter website; tracking pixels and spacers are skipped.
func EditionCover(body string) string {
	for _, tag := range editionImage.FindAllString(body, 20) {
		if pixelSize.MatchString(tag) {
			continue
		}
		if m := imageSrc.FindStringSubmatch(tag); m != nil {
			return m[1]
		}
	}
	return ""
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
