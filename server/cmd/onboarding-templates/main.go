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
)

func main() {
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
		if existing.Source != "xem-managed-onboarding" {
			keep = append(keep, row)
		}
	}
	for _, t := range onboardingemails.Templates {
		key := "xem-" + t.Key
		dir := filepath.Join("../client/public/assets/template-starters", key)
		must(os.MkdirAll(dir, 0755))
		must(os.WriteFile(filepath.Join(dir, "design.json"), t.Design(), 0644))
		html, _ := t.Render("{{workspace_name}}", "{{domain}}", "https://app.xem.email/settings/sending")
		must(os.WriteFile(filepath.Join(dir, "preview.html"), []byte(html), 0644))
		entry, err := json.Marshal(map[string]any{"key": key, "name": t.Subject, "category": "Transactional", "description": t.Description, "subject": t.Subject, "preheader": t.Next, "tags": []string{"managed sending", "onboarding", "xem"}, "marketing": false, "collection": "Xem originals", "source": "xem-managed-onboarding", "reference": "", "editingMode": "blocks", "version": 1, "designUrl": "/assets/template-starters/" + key + "/design.json", "previewUrl": "/assets/template-starters/" + key + "/preview.html"})
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
