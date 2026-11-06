package utils

import (
	"net/url"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/require"
	"kori/internal/config"
)

func TestCampaignUnsubscribeLinkIsTheFooterLink(t *testing.T) {
	cfg := &config.Config{}
	cfg.JWT.Secret = strings.Repeat("u", 32)
	cfg.Server.PublicURL = "https://api.example.com"

	link, err := UnsubscribeLink("mail-1", cfg)
	require.NoError(t, err)
	u, err := url.Parse(link)
	require.NoError(t, err)
	require.Equal(t, "https", u.Scheme, "managed sending only accepts an https List-Unsubscribe address")
	require.Equal(t, "/t/unsubscribe", u.Path)
	token, err := jwt.Parse(u.Query().Get("token"), func(*jwt.Token) (interface{}, error) { return []byte(cfg.JWT.Secret), nil })
	require.NoError(t, err)
	require.Equal(t, "mail-1", token.Claims.(jwt.MapClaims)["mailId"])

	body := ReplaceLinksWithRedirect(`<html><body><p>Hello</p></body></html>`, "mail-1", cfg, true)
	require.Contains(t, body, `<a href="`+link+`">Unsubscribe from this list</a></p></body>`, "the footer and the header point at the same address, and the footer isn't click-tracked")
}
