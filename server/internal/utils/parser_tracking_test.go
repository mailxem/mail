package utils

import (
	"encoding/base64"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/require"
	"regexp"
	"strings"
	"testing"
)

func TestTrackingPreservesDesignAndMessageIdentity(t *testing.T) {
	body := `<a class="button" target="_blank" href='https://example.com/story?a=1&amp;b=2'>Read</a><a href="https://example.com/public/unsubscribe/token">Unsubscribe</a><a href="mailto:help@example.com">Help</a>`
	result := TrackEmailHTML(body, "message-id", "https://api.example.com", strings.Repeat("s", 32))
	require.Contains(t, result, `class="button" target="_blank"`)
	require.Contains(t, result, `href="https://example.com/public/unsubscribe/token"`)
	require.Contains(t, result, `href="mailto:help@example.com"`)
	match := regexp.MustCompile(`/t/click/([^?]+)\?token=([^"']+)`).FindStringSubmatch(result)
	require.Len(t, match, 3)
	destination, err := base64.StdEncoding.DecodeString(match[1])
	require.NoError(t, err)
	require.Equal(t, "https://example.com/story?a=1&b=2", string(destination))
	token, err := jwt.Parse(match[2], func(*jwt.Token) (interface{}, error) { return []byte(strings.Repeat("s", 32)), nil })
	require.NoError(t, err)
	require.Equal(t, "message-id", token.Claims.(jwt.MapClaims)["mailId"])
}
