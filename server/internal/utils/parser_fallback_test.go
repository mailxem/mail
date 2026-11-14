package utils

import (
	"kori/internal/config"
	"kori/internal/utils/base64"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFieldFallbacks(t *testing.T) {
	cfg := &config.Config{}
	reader := map[string]string{"first_name": "Ravi", "last_name": "", "name": " ", "company": "Acme"}
	for in, want := range map[string]string{
		"Hi {{first_name|there}}":                 "Hi Ravi",
		"Hi {{ first_name | there }}":             "Hi Ravi",
		"Hi {{last_name|there}}":                  "Hi there",
		"Hi {{ last_name | dear reader }}!":       "Hi dear reader!",
		"Hi {{last_name|}}.":                      "Hi .",
		"Hi {{ last_name | }}.":                   "Hi .",
		"Hi {{name|friend}}":                      "Hi friend",
		"Hi {{city|there}}":                       "Hi there",
		"Hi {{ city.town | there }}":              "Hi there",
		"Hi {{last_name}}.":                       "Hi .",
		"Hi {{name}}.":                            "Hi  .",
		"Hi {{city}}":                             "Hi {{city}}",
		"{{first_name|a}} at {{company|nowhere}}": "Ravi at Acme",
	} {
		got, err := base64.DecodeFromBase64(ReplaceVariables(in, reader, "id", cfg, false, false))
		require.NoError(t, err)
		require.Equal(t, want, got, in)
	}
}
