package marketing

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"kori/internal/models"
)

type FormCondition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value,omitempty"`
}
type FormQuestion struct {
	Key          string         `json:"key"`
	Label        string         `json:"label"`
	Type         string         `json:"type"`
	Required     bool           `json:"required"`
	Placeholder  string         `json:"placeholder,omitempty"`
	HelpText     string         `json:"helpText,omitempty"`
	Options      []string       `json:"options,omitempty"`
	Condition    *FormCondition `json:"condition,omitempty"`
	DefaultValue string         `json:"defaultValue,omitempty"`
}
type FormPage struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description,omitempty"`
	Condition   *FormCondition `json:"condition,omitempty"`
	Fields      []FormQuestion `json:"fields"`
}
type FormConsent struct {
	Mode  string `json:"mode"`
	Label string `json:"label"`
}
type FormDefinition struct {
	SchemaVersion      int         `json:"schemaVersion"`
	Pages              []FormPage  `json:"pages"`
	Consent            FormConsent `json:"consent"`
	SaveProgress       bool        `json:"saveProgress"`
	SuccessRedirectURL string      `json:"successRedirectUrl,omitempty"`
}

var formKey = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{0,63}$`)
var formNumber = regexp.MustCompile(`^[+-]?(?:[0-9]+\.?[0-9]*|\.[0-9]+)(?:[eE][+-]?[0-9]+)?$`)
var reservedFormKeys = map[string]bool{"__proto__": true, "prototype": true, "constructor": true, "consent": true, "website": true, "requestId": true, "version": true, "sessionId": true, "resumeToken": true, "attribution": true, "fields": true, "pageId": true, "token": true}

// ValidateFormDefinition bounds payload complexity and guarantees acyclic forward branching.
func ValidateFormDefinition(d FormDefinition) error {
	if d.SchemaVersion != 1 || len(d.Pages) < 1 || len(d.Pages) > 12 {
		return errors.New("Use schema version 1 with 1 to 12 pages")
	}
	if d.Consent.Mode != "required" && d.Consent.Mode != "optional" && d.Consent.Mode != "none" {
		return errors.New("Invalid consent mode")
	}
	if strings.ContainsRune(d.Consent.Label, 0) || len(d.Consent.Label) > 500 || (d.Consent.Mode != "none" && strings.TrimSpace(d.Consent.Label) == "") {
		return errors.New("Provide consent wording of up to 500 characters")
	}
	if d.SuccessRedirectURL != "" {
		u, e := url.Parse(d.SuccessRedirectURL)
		if e != nil || len(d.SuccessRedirectURL) > 2048 || u.Scheme != "https" || u.Hostname() == "" || u.User != nil {
			return errors.New("Success redirects must use an HTTPS URL")
		}
	}
	seen, pages := map[string]bool{}, map[string]bool{}
	count := 0
	email := false
	for _, p := range d.Pages {
		if strings.ContainsRune(p.Title+p.Description, 0) || !formKey.MatchString(p.ID) || pages[p.ID] || len(p.Title) > 120 || strings.TrimSpace(p.Title) == "" || len(p.Description) > 1000 || len(p.Fields) < 1 {
			return errors.New("Each page needs a unique ID, a title, and questions")
		}
		pages[p.ID] = true
		if e := validateCondition(p.Condition, seen); e != nil {
			return e
		}
		for _, q := range p.Fields {
			count++
			if count > 50 {
				return errors.New("Use at most 50 questions")
			}
			if strings.ContainsRune(q.Label+q.Placeholder+q.HelpText+q.DefaultValue, 0) || !formKey.MatchString(q.Key) || reservedFormKeys[q.Key] || seen[q.Key] || len(q.Label) > 120 || strings.TrimSpace(q.Label) == "" || len(q.Placeholder) > 200 || len(q.HelpText) > 500 || len(q.DefaultValue) > 2000 {
				return errors.New("Questions need unique valid keys and bounded labels")
			}
			if e := validateCondition(q.Condition, seen); e != nil {
				return e
			}
			switch q.Type {
			case "TEXT", "EMAIL", "PHONE", "TEXTAREA", "SELECT", "RADIO", "CHECKBOX", "NUMBER", "DATE", "HIDDEN":
			default:
				return errors.New("Unsupported question type")
			}
			if len(q.Options) > 100 {
				return errors.New("Use at most 100 options")
			}
			options := map[string]bool{}
			for _, o := range q.Options {
				if strings.ContainsRune(o, 0) || strings.TrimSpace(o) == "" || strings.TrimSpace(o) != o || len(o) > 200 || options[o] {
					return errors.New("Options must be unique, trimmed, nonempty, and at most 200 characters")
				}
				options[o] = true
			}
			if (q.Type == "SELECT" || q.Type == "RADIO") && len(q.Options) == 0 {
				return errors.New("Choice questions need options")
			}
			if q.Type == "HIDDEN" && q.Required && strings.TrimSpace(q.DefaultValue) == "" {
				return errors.New("Required hidden questions need a default value")
			}
			if q.DefaultValue != "" {
				if e := validateFormValue(q, q.DefaultValue); e != nil {
					return e
				}
			}
			if q.Key == "email" {
				if q.Type != "EMAIL" || !q.Required || q.Condition != nil || p.Condition != nil {
					return errors.New("Email must be an unconditional required email question")
				}
				email = true
			}
			seen[q.Key] = true
		}
	}
	if !email {
		return errors.New("A required email question is needed")
	}
	return nil
}
func validateCondition(c *FormCondition, seen map[string]bool) error {
	if c == nil {
		return nil
	}
	if !seen[c.Field] || strings.ContainsRune(c.Value, 0) || len(c.Value) > 2000 {
		return errors.New("Conditions must reference an earlier question")
	}
	switch c.Operator {
	case "equals", "not_equals", "contains", "is_set":
		return nil
	}
	return errors.New("Unsupported condition operator")
}
func FormConditionMatches(c *FormCondition, answers map[string]string) bool {
	if c == nil {
		return true
	}
	v, ok := answers[c.Field]
	if !ok {
		return false
	}
	switch c.Operator {
	case "equals":
		return v == c.Value
	case "not_equals":
		return v != c.Value
	case "contains":
		return strings.Contains(v, c.Value)
	case "is_set":
		return strings.TrimSpace(v) != "" && v != "false"
	}
	return false
}
func validateFormValue(q FormQuestion, v string) error {
	if strings.ContainsRune(v, 0) {
		return fmt.Errorf("Remove invalid characters from %s", q.Label)
	}
	if len(v) > 2000 {
		return fmt.Errorf("%s must be at most 2,000 characters", q.Label)
	}
	if v == "" {
		return nil
	}
	switch q.Type {
	case "EMAIL":
		if !ValidEmail(v) {
			return fmt.Errorf("Enter a valid email for %s", q.Label)
		}
	case "NUMBER":
		n, e := strconv.ParseFloat(v, 64)
		if !formNumber.MatchString(v) || e != nil || math.IsNaN(n) || math.IsInf(n, 0) {
			return fmt.Errorf("Enter a valid number for %s", q.Label)
		}
	case "DATE":
		if _, e := time.Parse("2006-01-02", v); e != nil {
			return fmt.Errorf("Enter a valid date for %s", q.Label)
		}
	case "CHECKBOX":
		if v != "true" && v != "false" {
			return fmt.Errorf("Invalid checkbox answer for %s", q.Label)
		}
	case "SELECT", "RADIO":
		found := false
		for _, o := range q.Options {
			if o == v {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("Choose an available option for %s", q.Label)
		}
	}
	return nil
}

// FormAnswers evaluates conditions using only already accepted answers, preventing forged
// answers to unreachable questions from unlocking later pages. Hidden values come from the definition.
func FormAnswers(d FormDefinition, input map[string]string, partial bool) (map[string]string, []string, error) {
	if len(input) > 70 {
		return nil, nil, errors.New("Too many answers")
	}
	data := map[string]string{}
	pages := []string{}
	for _, p := range d.Pages {
		if !FormConditionMatches(p.Condition, data) {
			continue
		}
		pageHasAnswers := false
		for _, q := range p.Fields {
			if !FormConditionMatches(q.Condition, data) {
				continue
			}
			value, present := input[q.Key]
			if !present {
				value = q.DefaultValue
			}
			v := strings.TrimSpace(value)
			if q.Type == "HIDDEN" {
				v = q.DefaultValue
			}
			if q.Type == "EMAIL" {
				v = strings.ToLower(v)
			}
			if !partial && q.Required && (v == "" || q.Type == "CHECKBOX" && v != "true") {
				return nil, nil, fmt.Errorf("Complete %s", q.Label)
			}
			if e := validateFormValue(q, v); e != nil {
				return nil, nil, e
			}
			data[q.Key] = v
			pageHasAnswers = true
		}
		if pageHasAnswers {
			pages = append(pages, p.ID)
		}
	}
	return data, pages, nil
}

// DefinitionForForm gives older stored forms the same required-consent behavior they had before.
func DefinitionForForm(f models.Form) (FormDefinition, error) {
	if len(f.Definition) > 0 && string(f.Definition) != "null" && string(f.Definition) != "{}" {
		var d FormDefinition
		err := json.Unmarshal(f.Definition, &d)
		return d, err
	}
	d := FormDefinition{SchemaVersion: 1, Consent: FormConsent{Mode: "required", Label: "I agree to receive emails and can unsubscribe at any time."}, Pages: []FormPage{{ID: "details", Title: "Your details", Fields: []FormQuestion{}}}}
	for _, q := range f.Fields {
		if q.MapToContactField == nil {
			continue
		}
		d.Pages[0].Fields = append(d.Pages[0].Fields, FormQuestion{Key: *q.MapToContactField, Label: q.Label, Type: string(q.FieldType), Required: q.Required})
	}
	return d, nil
}
