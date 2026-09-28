package marketing

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func testFormDefinition() FormDefinition {
	return FormDefinition{SchemaVersion: 1, Consent: FormConsent{Mode: "optional", Label: "Send me updates"}, Pages: []FormPage{
		{ID: "details", Title: "Details", Fields: []FormQuestion{{Key: "email", Label: "Email", Type: "EMAIL", Required: true}, {Key: "interest", Label: "Interest", Type: "SELECT", Required: true, Options: []string{"business", "personal"}}, {Key: "source", Label: "Source", Type: "HIDDEN", DefaultValue: "website"}}},
		{ID: "business", Title: "Business", Condition: &FormCondition{Field: "interest", Operator: "equals", Value: "business"}, Fields: []FormQuestion{{Key: "company", Label: "Company", Type: "TEXT", Required: true}}},
		{ID: "followup", Title: "Follow-up", Condition: &FormCondition{Field: "company", Operator: "not_equals", Value: ""}, Fields: []FormQuestion{{Key: "employees", Label: "Employees", Type: "NUMBER", Required: true}}},
	}}
}
func TestFormAnswersRejectBranchSpoofing(t *testing.T) {
	d := testFormDefinition()
	require.NoError(t, ValidateFormDefinition(d))
	data, pages, err := FormAnswers(d, map[string]string{"email": " READER@example.com ", "interest": "personal", "company": "injected", "employees": "123", "source": "forged", "admin": "yes"}, false)
	require.NoError(t, err)
	require.Equal(t, []string{"details"}, pages)
	require.NotContains(t, data, "company")
	require.NotContains(t, data, "employees")
	require.NotContains(t, data, "admin")
	require.Equal(t, "website", data["source"])
	require.Equal(t, "reader@example.com", data["email"])
	_, _, err = FormAnswers(d, map[string]string{"email": "reader@example.com", "interest": "business"}, false)
	require.ErrorContains(t, err, "Company")
	_, _, err = FormAnswers(d, map[string]string{"email": "reader@example.com", "interest": "business", "company": "Acme", "employees": "NaN"}, false)
	require.ErrorContains(t, err, "number")
	_, _, err = FormAnswers(d, map[string]string{"interest": "other"}, true)
	require.ErrorContains(t, err, "available option")
	_, _, err = FormAnswers(d, map[string]string{"interest": "business"}, true)
	require.NoError(t, err)
}
func TestFormDefinitionValidation(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*FormDefinition)
	}{
		{"forward reference", func(d *FormDefinition) { d.Pages[0].Condition = &FormCondition{Field: "company", Operator: "is_set"} }},
		{"duplicate key", func(d *FormDefinition) { d.Pages[0].Fields[1].Key = "email" }},
		{"protocol key", func(d *FormDefinition) { d.Pages[0].Fields[1].Key = "resumeToken" }},
		{"prototype", func(d *FormDefinition) { d.Pages[0].Fields[1].Key = "constructor" }},
		{"conditional email", func(d *FormDefinition) {
			d.Pages[0].Fields[0].Condition = &FormCondition{Field: "company", Operator: "is_set"}
		}},
		{"script redirect", func(d *FormDefinition) { d.SuccessRedirectURL = "javascript:alert(1)" }},
		{"missing consent wording", func(d *FormDefinition) { d.Consent.Label = "" }},
		{"unsupported operator", func(d *FormDefinition) { d.Pages[1].Condition.Operator = "regex" }},
		{"duplicate option", func(d *FormDefinition) { d.Pages[0].Fields[1].Options = []string{"business", "business"} }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := testFormDefinition()
			tc.mutate(&d)
			require.Error(t, ValidateFormDefinition(d))
		})
	}
}

func TestFormAnswersDefaultsAndEmptyConditionalPages(t *testing.T) {
	d := testFormDefinition()
	d.Pages[0].Fields[1].DefaultValue = "personal"
	d.Pages = append(d.Pages, FormPage{ID: "unused", Title: "Unused", Fields: []FormQuestion{{Key: "role", Label: "Role", Type: "TEXT", Condition: &FormCondition{Field: "interest", Operator: "equals", Value: "business"}}}})
	require.NoError(t, ValidateFormDefinition(d))
	data, pages, err := FormAnswers(d, map[string]string{"email": "reader@example.com"}, false)
	require.NoError(t, err)
	require.Equal(t, "personal", data["interest"])
	require.Equal(t, []string{"details"}, pages)
	// Clearing a supplied value must not silently restore a default.
	_, _, err = FormAnswers(d, map[string]string{"email": "reader@example.com", "interest": ""}, false)
	require.ErrorContains(t, err, "Interest")
}

func TestFormDefinitionRejectsUnusableOptionsAndNullCharacters(t *testing.T) {
	d := testFormDefinition()
	d.Pages[0].Fields[1].Options = []string{" business "}
	require.ErrorContains(t, ValidateFormDefinition(d), "trimmed")
	d = testFormDefinition()
	d.Pages[0].Title = "Details\x00"
	require.Error(t, ValidateFormDefinition(d))
	d = testFormDefinition()
	_, _, err := FormAnswers(d, map[string]string{"email": "reader@example.com", "interest": "business", "company": "Acme\x00"}, false)
	require.ErrorContains(t, err, "invalid characters")
}

func TestFormNumberAnswersMatchBrowserDecimalSyntax(t *testing.T) {
	d := testFormDefinition()
	for _, value := range []string{"0x1p2", "1_000", "Infinity", "NaN"} {
		_, _, err := FormAnswers(d, map[string]string{"email": "reader@example.com", "interest": "business", "company": "Acme", "employees": value}, false)
		require.ErrorContains(t, err, "number", value)
	}
	for _, value := range []string{"1000", "1e3", "1.5", ".5"} {
		_, _, err := FormAnswers(d, map[string]string{"email": "reader@example.com", "interest": "business", "company": "Acme", "employees": value}, false)
		require.NoError(t, err, value)
	}
}
