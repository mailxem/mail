package mailconnect

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestCloudflareRecipientOutcomes(t *testing.T) {
	for _, test := range []struct {
		name, response, status string
		unknown                bool
	}{
		{"delivered", `{"delivered":["reader@example.com"],"queued":[],"permanent_bounces":[]}`, "SENT", false},
		{"queued", `{"delivered":[],"queued":["reader@example.com"],"permanent_bounces":[]}`, "ACCEPTED", false},
		{"bounce", `{"delivered":[],"queued":[],"permanent_bounces":["reader@example.com"]}`, "BOUNCED", false},
		{"missing recipient", `{"delivered":[],"queued":[],"permanent_bounces":[]}`, "", true},
		{"unexpected recipient", `{"delivered":["other@example.com"],"queued":[],"permanent_bounces":[]}`, "", true},
		{"duplicate result", `{"delivered":["reader@example.com"],"queued":["reader@example.com"]}`, "", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				require.Equal(t, "https://api.cloudflare.com/client/v4/accounts/"+strings.Repeat("a", 32)+"/email/sending/send", r.URL.String())
				require.Equal(t, "POST", r.Method)
				require.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
				var body map[string]any
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				require.Equal(t, "reply@example.com", body["reply_to"])
				require.Equal(t, "Received", body["text"])
				require.NotContains(t, body, "replyTo")
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"success":true,"result":` + test.response + `}`))}, nil
			})}
			result, err := SendCloudflare(context.Background(), client, strings.Repeat("a", 32), "test-token", CloudflareMessage{From: "sender@example.com", To: []string{"reader@example.com"}, Subject: "Receipt", HTML: "<p>Received</p>", ReplyTo: "reply@example.com"})
			if test.unknown {
				require.ErrorContains(t, err, "delivery outcome unknown")
			} else {
				require.NoError(t, err)
				require.Equal(t, test.status, result.Status())
			}
			require.Equal(t, 1, calls)
		})
	}
	require.Equal(t, "PARTIAL", (CloudflareResult{Delivered: []string{"a"}, PermanentBounces: []string{"b"}}).Status())
}

func TestCloudflareDoesNotRetryOrLeakProviderErrors(t *testing.T) {
	for _, status := range []int{0, 400, 401, 403, 429, 500, 302} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			calls := 0
			client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				calls++
				if status == 0 {
					return nil, errors.New("secret-provider-error")
				}
				return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader("secret-provider-error"))}, nil
			})}
			_, err := SendCloudflare(context.Background(), client, strings.Repeat("a", 32), "token", CloudflareMessage{From: "sender@example.com", To: []string{"reader@example.com"}, HTML: "hello"})
			require.Error(t, err)
			require.NotContains(t, err.Error(), "secret-provider-error")
			require.Equal(t, 1, calls)
			if status == 0 || status >= 500 || status == 302 {
				require.ErrorContains(t, err, "delivery outcome unknown")
			}
		})
	}
}

func TestCloudflareValidationBeforeNetwork(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { t.Fatal("unexpected network call"); return nil, nil })}
	for _, m := range []CloudflareMessage{
		{From: "sender@example.com", To: []string{"reader@example.com"}, HTML: "hi", Subject: "bad\r\nBcc: victim@example.com"},
		{From: "sender@example.com", To: []string{"reader@example.com"}, BCC: []string{"reader@example.com"}, HTML: "hi"},
		{From: "sender@example.com", To: []string{"reader@example.com"}, HTML: strings.Repeat("x", 5*1024*1024+1)},
	} {
		_, err := SendCloudflare(context.Background(), client, strings.Repeat("a", 32), "token", m)
		require.Error(t, err)
	}
}
