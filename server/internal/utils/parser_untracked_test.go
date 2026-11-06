package utils

import (
	"kori/internal/config"
	"kori/internal/utils/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTrackingOffKeepsLinksAndUnsubscribe(t *testing.T) {
	cfg := &config.Config{}
	cfg.JWT.Secret = strings.Repeat("s", 32)
	cfg.Server.PublicURL = "https://api.example.com"
	body := `<html><body><a href="https://example.com/story">Read</a></body></html>`

	off, err := base64.DecodeFromBase64(ReplaceVariables(body, nil, "message-id", cfg, false, true))
	require.NoError(t, err)
	require.NotContains(t, off, "/t/open")
	require.NotContains(t, off, "/t/click/")
	require.Contains(t, off, `href="https://example.com/story"`)
	require.Contains(t, off, "https://api.example.com/t/unsubscribe?token=")

	on, err := base64.DecodeFromBase64(ReplaceVariables(body, nil, "message-id", cfg, true, true))
	require.NoError(t, err)
	require.Contains(t, on, "/t/open?token=")
	require.Contains(t, on, "/t/click/")
	require.Contains(t, on, "https://api.example.com/t/unsubscribe?token=")
}
