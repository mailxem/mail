package mailconnect

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"
)

type CloudflareMessage struct {
	From    string            `json:"from"`
	To      []string          `json:"to"`
	CC      []string          `json:"cc,omitempty"`
	BCC     []string          `json:"bcc,omitempty"`
	Subject string            `json:"subject"`
	HTML    string            `json:"html"`
	Text    string            `json:"text"`
	ReplyTo string            `json:"reply_to,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}
type CloudflareResult struct {
	Delivered        []string `json:"delivered"`
	PermanentBounces []string `json:"permanent_bounces"`
	Queued           []string `json:"queued"`
}

func (r CloudflareResult) Status() string {
	if len(r.PermanentBounces) > 0 {
		if len(r.Delivered)+len(r.Queued) == 0 {
			return "BOUNCED"
		}
		return "PARTIAL"
	}
	if len(r.Queued) > 0 {
		return "ACCEPTED"
	}
	return "SENT"
}

func validateCloudflareMessage(m *CloudflareMessage) (map[string]bool, error) {
	if utf8.RuneCountInString(m.Subject) > 998 || strings.ContainsAny(m.Subject, "\r\n") {
		return nil, errors.New("invalid Cloudflare subject")
	}
	if len(m.HTML)+len(m.Text) > 5*1024*1024 {
		return nil, errors.New("Cloudflare message exceeds 5 MiB")
	}
	if m.HTML == "" && m.Text == "" {
		return nil, errors.New("message body is required")
	}
	if _, err := mail.ParseAddress(m.From); err != nil {
		return nil, errors.New("invalid sender")
	}
	if m.ReplyTo != "" {
		if _, err := mail.ParseAddress(m.ReplyTo); err != nil {
			return nil, errors.New("invalid reply address")
		}
	}
	count := len(m.To) + len(m.CC) + len(m.BCC)
	if len(m.To) == 0 || count > 50 {
		return nil, errors.New("Cloudflare supports 1 to 50 recipients")
	}
	expected := make(map[string]bool)
	for _, list := range [][]string{m.To, m.CC, m.BCC} {
		for _, raw := range list {
			a, err := mail.ParseAddress(raw)
			if err != nil {
				return nil, errors.New("invalid recipient")
			}
			key := strings.ToLower(a.Address)
			if expected[key] {
				return nil, errors.New("duplicate recipient")
			}
			expected[key] = true
		}
	}
	headerSize := 0
	for k, v := range m.Headers {
		headerSize += len(k) + len(v) + 4
		if strings.ContainsAny(k+v, "\r\n") {
			return nil, errors.New("invalid message headers")
		}
	}
	if headerSize > 16*1024 {
		return nil, errors.New("Cloudflare headers exceed 16 KiB")
	}
	return expected, nil
}

// There is deliberately no automatic retry: an HTTP timeout or partial result
// may mean recipients already received the message.
func SendCloudflare(ctx context.Context, client *http.Client, account, token string, m CloudflareMessage) (*CloudflareResult, error) {
	if !regexp.MustCompile(`^[a-fA-F0-9]{32}$`).MatchString(account) {
		return nil, errors.New("invalid Cloudflare account ID")
	}
	if m.Text == "" && len(m.HTML) <= 5*1024*1024 {
		m.Text = plainMailText(m.HTML)
	}
	expected, err := validateCloudflareMessage(&m)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	if len(raw) > 5*1024*1024 {
		return nil, errors.New("Cloudflare request exceeds 5 MiB")
	}
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.cloudflare.com/client/v4/accounts/"+account+"/email/sending/send", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return nil, errors.New("delivery outcome unknown: Cloudflare request interrupted")
	}
	defer response.Body.Close()
	if response.StatusCode >= 400 && response.StatusCode < 500 {
		return nil, fmt.Errorf("Cloudflare rejected email (HTTP %d); check sender setup, token permissions, and limits", response.StatusCode)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, errors.New("delivery outcome unknown: Cloudflare returned an unexpected status")
	}
	var envelope struct {
		Success bool              `json:"success"`
		Result  *CloudflareResult `json:"result"`
	}
	if json.NewDecoder(io.LimitReader(response.Body, 65536)).Decode(&envelope) != nil || !envelope.Success || envelope.Result == nil {
		return nil, errors.New("delivery outcome unknown: invalid Cloudflare response")
	}
	seen := make(map[string]bool)
	for _, list := range [][]string{envelope.Result.Delivered, envelope.Result.PermanentBounces, envelope.Result.Queued} {
		for _, address := range list {
			key := strings.ToLower(address)
			if !expected[key] || seen[key] {
				return nil, errors.New("delivery outcome unknown: inconsistent Cloudflare recipient results")
			}
			seen[key] = true
		}
	}
	if len(seen) != len(expected) {
		return nil, errors.New("delivery outcome unknown: incomplete Cloudflare recipient results")
	}
	return envelope.Result, nil
}

func plainMailText(body string) string {
	z := html.NewTokenizer(strings.NewReader(body))
	var result strings.Builder
	hidden := 0
	for {
		kind := z.Next()
		if kind == html.ErrorToken {
			break
		}
		token := z.Token()
		switch kind {
		case html.StartTagToken:
			if token.Data == "head" || token.Data == "script" || token.Data == "style" {
				hidden++
			}
			if token.Data == "br" {
				result.WriteByte(' ')
			}
		case html.EndTagToken:
			if token.Data == "head" || token.Data == "script" || token.Data == "style" {
				if hidden > 0 {
					hidden--
				}
			}
			if token.Data == "p" || token.Data == "div" || token.Data == "li" || token.Data == "h1" || token.Data == "h2" {
				result.WriteByte(' ')
			}
		case html.TextToken:
			if hidden == 0 {
				result.WriteString(token.Data)
			}
		}
	}
	return strings.Join(strings.Fields(result.String()), " ")
}
