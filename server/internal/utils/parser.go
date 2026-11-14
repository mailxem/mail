package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	stdbase64 "encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"kori/internal/config"
	"kori/internal/utils/base64"
	"kori/internal/utils/logger"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/golang-jwt/jwt"
	"gorm.io/datatypes"
)

var console = logger.New("console")

// ParseVariables input is html text with variables in the form of {{variable}} or {{ variable.subvariable }} or {{ varible }}
// output is a map of variables and their values
// input: <p>Hello, {{name}}! {{name.first}} {{name.last}}</p>
// output: map[name:John, name.first:John, name.last:Doe]
func ParseVariables(html string) (map[string]string, error) {
	variables := make(map[string]string)

	re := regexp.MustCompile(`{{\s*(\w+(?:\.\w+)*)\s*}}`)
	matches := re.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		variables[match[1]] = match[1]
	}

	return variables, nil
}

// ReplaceVariables input is html text with variables in the form of {{variable}} or {{ variable.subvariable }} or {{ varible }}
// output is a string with the variables replaced by their values
func ReplaceVariables(input string, variables map[string]string, mailId string, cfg *config.Config, trackLinks bool, isMarketing bool) string {
	return base64.EncodeToBase64(fillVariables(input, variables, mailId, cfg, trackLinks, isMarketing))
}

// fallbackField is {{field|fallback}} for any field; the fallback is whatever
// sits between the bar and the closing braces, trimmed.
var fallbackField = regexp.MustCompile(`{{\s*\w+(?:\.\w+)*\s*\|([^{}]*)}}`)

// fillVariables is ReplaceVariables without the base64. Variables go in key
// order, so a value that itself holds a {{variable}} comes out the same every
// time: bulk emails are rendered again when sent and when shown.
// {{first_name|there}} prints "there" when the reader has no first name;
// {{first_name}} prints the value as it is, even blank.
func fillVariables(input string, variables map[string]string, mailId string, cfg *config.Config, trackLinks bool, isMarketing bool) string {
	input = EmailSections(input, variables)
	for _, variable := range slices.Sorted(maps.Keys(variables)) {
		value := variables[variable]
		re := regexp.MustCompile(`{{\s*` + regexp.QuoteMeta(variable) + `(?:\.\w+)*\s*(\|[^{}]*)?}}`)
		input = re.ReplaceAllStringFunc(input, func(match string) string {
			if fallback := re.FindStringSubmatch(match)[1]; fallback != "" && strings.TrimSpace(value) == "" {
				return strings.TrimSpace(fallback[1:])
			}
			return value
		})
	}
	// A field this reader has no value for at all.
	input = fallbackField.ReplaceAllStringFunc(input, func(match string) string {
		return strings.TrimSpace(fallbackField.FindStringSubmatch(match)[1])
	})

	if trackLinks || isMarketing {
		input = ReplaceLinksWithRedirect(input, mailId, cfg, isMarketing, trackLinks)
	}
	return input
}

// JSONToMap convert datatypes.JSON to map[string]string
func JSONToMap(jsonData datatypes.JSON) (map[string]string, error) {
	result := map[string]string{}
	// A campaign saved without variables stores NULL: that's no values, not an error.
	if len(jsonData) == 0 {
		return result, nil
	}
	err := json.Unmarshal(jsonData, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// MapToJSON convert map[string]string to datatypes.JSON
func MapToJSON(data map[string]string) (datatypes.JSON, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

// ReplaceLinksWithRedirect usecase is to replace all the links in the html with our redirect url
// so we can track the number of clicks. Marketing email always gets the
// unsubscribe footer; track=false leaves links and the open pixel out.
func ReplaceLinksWithRedirect(body string, mailID string, cfg *config.Config, isMarketing bool, track bool) string {
	if isMarketing {
		link, err := UnsubscribeLink(mailID, cfg)
		if err == nil {
			// Inline styles: email clients drop stylesheets, and a bare link renders
			// left-aligned in default blue under a centered template footer.
			footer := fmt.Sprintf(`<p style="margin:0;padding:12px 24px 32px;text-align:center;font-family:Helvetica,Arial,sans-serif;font-size:12px;line-height:1.5;color:#6b6b6b"><a href="%s" style="color:#6b6b6b;text-decoration:underline">Unsubscribe from this list</a></p>`, link)
			if strings.Contains(body, "</body>") {
				body = strings.Replace(body, "</body>", footer+"</body>", 1)
			} else {
				body += footer
			}
		}
	}
	if !track {
		return body
	}
	return TrackEmailHTML(body, mailID, cfg.Server.PublicURL, cfg.JWT.Secret)
}

// UnsubscribeLink is the signed one-click unsubscribe address for one email; the
// same address goes in the footer and in the List-Unsubscribe header.
func UnsubscribeLink(mailID string, cfg *config.Config) (string, error) {
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"mailId": mailID}).SignedString([]byte(cfg.JWT.Secret))
	if err != nil {
		return "", err
	}
	return cfg.Server.PublicURL + "/t/unsubscribe?token=" + signed, nil
}

// ClickSignature binds a tracked link's destination to its email, so a token
// from one email can't be reused to redirect anywhere else.
func ClickSignature(secret, messageID, target string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("click\x00" + messageID + "\x00" + target))
	return stdbase64.RawURLEncoding.EncodeToString(mac.Sum(nil)[:16])
}

// TrackEmailHTML keeps original attributes and does not count opt-out links as engagement.
func TrackEmailHTML(body, messageID, publicURL, secret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"mailId": messageID})
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return body
	}
	links := regexp.MustCompile(`(?i)<a\b[^>]*\bhref\s*=\s*(?:"[^"]*"|'[^']*')[^>]*>`)
	href := regexp.MustCompile(`(?i)\shref\s*=\s*(?:"([^"]*)"|'([^']*)')`)
	body = links.ReplaceAllStringFunc(body, func(tag string) string {
		parts := href.FindStringSubmatch(tag)
		if len(parts) < 3 {
			return tag
		}
		target := parts[1]
		if target == "" {
			target = parts[2]
		}
		target = html.UnescapeString(target)
		lower := strings.ToLower(target)
		if (!strings.HasPrefix(lower, "https://") && !strings.HasPrefix(lower, "http://")) || strings.Contains(lower, "unsubscribe") || strings.Contains(lower, "/t/click/") {
			return tag
		}
		tracking := fmt.Sprintf(` href="%s/t/click/%s?token=%s&amp;sig=%s"`, publicURL, base64.EncodeToBase64(target), signed, ClickSignature(secret, messageID, target))
		return href.ReplaceAllStringFunc(tag, func(string) string { return tracking })
	})
	return body + fmt.Sprintf(`<img alt="" src="%s/t/open?token=%s" width="1" height="1">`, publicURL, signed)
}
