// Generates the dashboard's editable copies from the exact transactional designs.
// Run from server: go run ./cmd/onboarding-templates
package main

import (
	"bytes"
	"encoding/json"
	"kori/internal/onboardingemails"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	brandDir := "../client/public/assets/template-starters/brand"
	must(os.MkdirAll(brandDir, 0755))
	for name, source := range map[string]string{
		"xem-mark.png":            "../website/public/brand/xem-mark.png",
		"dm-sans.woff2":           "../website/design/blog/fonts/dm-sans.woff2",
		"eb-garamond.woff2":       "../website/design/blog/fonts/eb-garamond.woff2",
		"dm-sans-LICENSE.txt":     "../website/design/blog/fonts/dm-sans-LICENSE.txt",
		"eb-garamond-LICENSE.txt": "../website/design/blog/fonts/eb-garamond-LICENSE.txt",
	} {
		data, err := os.ReadFile(source)
		must(err)
		must(os.WriteFile(filepath.Join(brandDir, name), data, 0644))
	}
	must(os.WriteFile(filepath.Join(brandDir, "fonts.css"), []byte("@font-face{font-family:'DM Sans';font-style:normal;font-weight:100 1000;src:url('./dm-sans.woff2') format('woff2');font-display:swap}@font-face{font-family:'EB Garamond';font-style:normal;font-weight:400 800;src:url('./eb-garamond.woff2') format('woff2');font-display:swap}\n"), 0644))
	manifestPath := "../client/lib/template-starters/manifest.json"
	raw, err := os.ReadFile(manifestPath)
	must(err)
	var rows []json.RawMessage
	must(json.Unmarshal(raw, &rows))
	keep := rows[:0]
	for _, row := range rows {
		var existing struct {
			Source string `json:"source"`
		}
		must(json.Unmarshal(row, &existing))
		if existing.Source != "xem-managed-onboarding" && existing.Source != "xem-service-notifications" {
			keep = append(keep, row)
		}
	}
	for _, t := range onboardingemails.Templates {
		key := "xem-" + t.Key
		dir := filepath.Join("../client/public/assets/template-starters", key)
		must(os.MkdirAll(dir, 0755))
		must(os.WriteFile(filepath.Join(dir, "design.json"), t.Design(), 0644))
		html, _ := t.Render("{{workspace_name}}", "{{domain}}", "https://app.xem.email/settings/sending")
		// Gallery iframes use the same public assets as this installation's editor.
		html = strings.ReplaceAll(html, "https://app.xem.email/assets/template-starters/", "/assets/template-starters/")
		must(os.WriteFile(filepath.Join(dir, "preview.html"), []byte(html), 0644))
		entry, err := json.Marshal(map[string]any{"key": key, "name": t.Subject, "category": "Transactional", "description": t.Description, "subject": t.Subject, "preheader": t.Next, "tags": []string{"managed sending", "onboarding", "xem"}, "marketing": false, "collection": "Xem originals", "source": "xem-managed-onboarding", "reference": "", "editingMode": "blocks", "version": 2, "designUrl": "/assets/template-starters/" + key + "/design.json", "previewUrl": "/assets/template-starters/" + key + "/preview.html"})
		must(err)
		keep = append(keep, entry)
	}
	for _, t := range onboardingemails.ServiceTemplates {
		key := "xem-" + t.Key
		dir := filepath.Join("../client/public/assets/template-starters", key)
		must(os.MkdirAll(dir, 0755))
		must(os.WriteFile(filepath.Join(dir, "design.json"), t.Design(), 0644))
		html, _ := t.Render(onboardingemails.ServiceData{Name: "{{first_name}}", Email: "{{account_email}}", Workspace: "{{workspace_name}}", Time: "{{event_time}}", Count: "{{event_count}}"}, "https://app.xem.email"+t.Path)
		html = strings.ReplaceAll(html, "https://app.xem.email/assets/template-starters/", "/assets/template-starters/")
		must(os.WriteFile(filepath.Join(dir, "preview.html"), []byte(html), 0644))
		entry, err := json.Marshal(map[string]any{"key": key, "name": t.Subject, "category": "Transactional", "description": t.Description, "subject": t.Subject, "preheader": t.Next, "tags": []string{"account", "security", "sending alerts", "xem"}, "marketing": false, "collection": "Xem originals", "source": "xem-service-notifications", "reference": "", "editingMode": "blocks", "version": 1, "designUrl": "/assets/template-starters/" + key + "/design.json", "previewUrl": "/assets/template-starters/" + key + "/preview.html"})
		must(err)
		keep = append(keep, entry)
	}

	var output bytes.Buffer
	encoder := json.NewEncoder(&output)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	must(encoder.Encode(keep))
	must(os.WriteFile(manifestPath, output.Bytes(), 0644))
}
func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
