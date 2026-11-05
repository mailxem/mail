package ai

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

// CopySlot is one writable heading, text block or button in a library design.
type CopySlot struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
	Text string `json:"text"`
}

// DesignImage is a small screenshot sent to the model as a data URL.
type DesignImage struct {
	Role string `json:"role"` // "reference" or "example"
	Data string `json:"data"`
	// Source names where the screenshot comes from, e.g. its page in the
	// public template gallery, so the model can cite it. Optional.
	Source string `json:"source,omitempty"`
}

const (
	maxDesignImages     = 4
	maxDesignImageBytes = 300 * 1024
	maxDesignImageTotal = 1024 * 1024
)

// designImageParts validates the screenshots and turns them into vision
// content parts, labelled so the model knows which image is which.
func designImageParts(images []DesignImage) ([]map[string]any, error) {
	if len(images) > maxDesignImages {
		return nil, errors.New("Send at most four reference images")
	}
	total := 0
	var parts []map[string]any
	for i, im := range images {
		ok := strings.HasPrefix(im.Data, "data:image/webp;base64,") || strings.HasPrefix(im.Data, "data:image/png;base64,") || strings.HasPrefix(im.Data, "data:image/jpeg;base64,")
		if !ok || len(im.Data) > maxDesignImageBytes || len(im.Source) > 200 || strings.ContainsAny(im.Source, "\r\n") || (im.Role != "reference" && im.Role != "example") || (im.Role == "reference" && i != 0) {
			return nil, errors.New("A reference image is invalid. Try again")
		}
		total += len(im.Data)
		if total > maxDesignImageTotal {
			return nil, errors.New("The reference images are too large. Try again")
		}
		label := "Image: the template you are adapting. Write copy that fits each slot's space and weight as shown."
		if im.Role == "example" {
			label = "Image: an example of the house style, for tone and structure only. Do not copy its words."
		}
		if im.Source != "" {
			label += " Source: " + im.Source + "."
		}
		parts = append(parts,
			map[string]any{"type": "text", "text": label},
			map[string]any{"type": "image_url", "image_url": map[string]string{"url": im.Data, "detail": "low"}},
		)
	}
	return parts, nil
}

// SlotRow lists the slot ids in one row of the design, so a whole section can be dropped.
type SlotRow struct {
	ID    string   `json:"id"`
	Slots []string `json:"slots"`
}

// humanVoice is shared by every copy prompt: concrete, plain, and free of the
// phrases readers recognise as machine-written.
const humanVoice = `Write like a thoughtful person, not a marketing bot: specific, plain, warm without gushing. Use contractions. Vary sentence length. Never use: "thrilled", "excited to announce", "delve", "elevate", "unlock", "supercharge", "game-changer", "seamless", "journey" (figurative), "whether you're X or Y", "in today's world", "look no further", "spark inspiration". No em dashes (—): use a comma, colon or full stop instead. No exclamation marks unless the user asks for them.`

//go:embed skills/email-design.md
var emailDesignSkill string

var slotsPrompt = `You are Xem's email copywriter, adapting a finished, professionally designed email template to a new brief. The design, layout, fonts, colours and images stay exactly as they are; you only write the words.
Input JSON: instruction (the brief), reference (what the template is), slots (each {id, kind, text}: kind is heading, text or button; text is the template's current copy), rows (each {id, slots}).
Return one JSON object: {subject: string (max 200), preheader: string (max 300), copy: {slotId: newText, ...}, remove: [rowId, ...]}.
Rules:
- Rewrite every slot that should change for the brief. Keep each new text close to the original's length (about 0.5x to 1.5x) and its role: headings stay short, buttons are 1 to 5 words, kickers stay short labels, footers and legal lines stay footers.
- Plain text only. Use \n for a line break and \n\n between paragraphs. No HTML, markdown, emoji bullets or URLs.
- Never invent facts: no prices, dates, numbers, names, addresses, offers or promises the brief doesn't give. Use a short [placeholder in brackets] for anything missing, e.g. [price], [date].
- Keep merge variables like {{first_name}} exactly as written.
- Use remove only for rows whose whole purpose doesn't fit the brief (for example a product grid in a plain announcement). Never remove the first row, the last row, or every row that contains a button.
- Only use slot and row ids from the input.
Follow this email design skill when deciding what each slot should say:

` + emailDesignSkill + "\n" + humanVoice

func (w *EmailWriter) draftSlots(ctx context.Context, input EmailDraftRequest) (*EmailDraft, error) {
	if len(input.Slots) == 0 || len(input.Slots) > 80 || len(input.Rows) > 60 || len(input.Reference) > 600 {
		return nil, errors.New("Choose a template to design from and try again")
	}
	allowed := make(map[string]CopySlot, len(input.Slots))
	for _, s := range input.Slots {
		if s.ID == "" || len(s.ID) > 80 || len(s.Text) > 1200 {
			return nil, errors.New("This template is too large to adapt. Try a simpler one")
		}
		allowed[s.ID] = s
	}
	rows := map[string]bool{}
	for _, r := range input.Rows {
		rows[r.ID] = true
	}
	images, err := designImageParts(input.Images)
	if err != nil {
		return nil, err
	}
	request, _ := json.Marshal(map[string]any{"instruction": input.Instruction, "reference": input.Reference, "slots": input.Slots, "rows": input.Rows})
	var user any = string(request)
	if len(images) > 0 {
		user = append([]map[string]any{{"type": "text", "text": string(request)}}, images...)
	}
	content, err := w.complete(ctx, map[string]any{
		"model": w.Model, "max_tokens": 6000, "response_format": map[string]string{"type": "json_object"},
		"messages": []map[string]any{{"role": "system", "content": slotsPrompt}, {"role": "user", "content": user}},
	})
	if err != nil {
		switch {
		case errors.Is(err, errProviderUnavailable):
			return nil, errors.New("The writing service is unavailable. Please try again")
		case errors.Is(err, errProviderIncomplete):
			return nil, errors.New("The writing service returned an incomplete draft. Try a shorter request")
		default:
			return nil, errors.New("The writing service could not generate a draft. Please try again")
		}
	}
	var out struct {
		Subject   string            `json:"subject"`
		Preheader string            `json:"preheader"`
		Copy      map[string]string `json:"copy"`
		Remove    []string          `json:"remove"`
	}
	if json.Unmarshal([]byte(content), &out) != nil || strings.TrimSpace(out.Subject) == "" || utf8.RuneCountInString(out.Subject) > 200 || utf8.RuneCountInString(out.Preheader) > 300 {
		return nil, errors.New("The writing service returned an invalid draft. Please try again")
	}
	copyOut := map[string]string{}
	var body strings.Builder
	for id, text := range out.Copy {
		slot, ok := allowed[id]
		text = noDashes(strings.TrimSpace(text))
		if !ok || text == "" || strings.ContainsAny(text, "<>") || utf8.RuneCountInString(text) > 2400 {
			continue
		}
		if slot.Kind == "button" && utf8.RuneCountInString(text) > 60 {
			continue
		}
		copyOut[id] = text
		body.WriteString(text)
		body.WriteString("\n\n")
	}
	if len(copyOut) == 0 {
		return nil, errors.New("The writing service returned an invalid draft. Please try again")
	}
	var remove []string
	for _, id := range out.Remove {
		if rows[id] && len(remove) < len(input.Rows)/2 {
			remove = append(remove, id)
		}
	}
	return &EmailDraft{Subject: noDashes(strings.TrimSpace(out.Subject)), Preheader: noDashes(strings.TrimSpace(out.Preheader)), Body: strings.TrimSpace(body.String()), Copy: copyOut, Remove: remove}, nil
}

// Models keep writing em dashes mid-sentence even when told not to, and they
// read as machine-written. A leading dash ("— Jo, founder") is a sign-off and stays.
var midDash = regexp.MustCompile(`(\S)\s*—\s*(\S)`)

func noDashes(s string) string { return midDash.ReplaceAllString(s, "$1, $2") }
