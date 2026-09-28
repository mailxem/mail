package utils

import (
	"encoding/json"
	"fmt"
	"html"
	"kori/internal/config"
	"kori/internal/utils/base64"
	"kori/internal/utils/logger"
	"regexp"
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
	for variable, value := range variables {
		re := regexp.MustCompile(`{{\s*` + regexp.QuoteMeta(variable) + `(?:\.\w+)*\s*}}`)
		input = re.ReplaceAllStringFunc(input, func(string) string { return value })
	}

	if trackLinks {
		input = ReplaceLinksWithRedirect(input, mailId, cfg, isMarketing)
	}

	return base64.EncodeToBase64(input)
}

// JSONToMap convert datatypes.JSON to map[string]string
func JSONToMap(jsonData datatypes.JSON) (map[string]string, error) {
	var result map[string]string
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
// so we can track the number of clicks
func ReplaceLinksWithRedirect(body string, mailID string, cfg *config.Config, isMarketing bool) string {
	if isMarketing {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"mailId": mailID})
		signed, err := token.SignedString([]byte(cfg.JWT.Secret))
		if err == nil {
			footer := fmt.Sprintf(`<p><a href="%s/t/unsubscribe?token=%s">Unsubscribe from this list</a></p>`, cfg.Server.PublicURL, signed)
			if strings.Contains(body, "</body>") {
				body = strings.Replace(body, "</body>", footer+"</body>", 1)
			} else {
				body += footer
			}
		}
	}
	return TrackEmailHTML(body, mailID, cfg.Server.PublicURL, cfg.JWT.Secret)
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
		tracking := fmt.Sprintf(` href="%s/t/click/%s?token=%s"`, publicURL, base64.EncodeToBase64(target), signed)
		return href.ReplaceAllStringFunc(tag, func(string) string { return tracking })
	})
	return body + fmt.Sprintf(`<img alt="" src="%s/t/open?token=%s" width="1" height="1">`, publicURL, signed)
}
