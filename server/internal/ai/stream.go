package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	errProviderUnavailable = errors.New("provider unavailable")
	errProviderStatus      = errors.New("provider returned an error status")
	errProviderResponse    = errors.New("provider returned an invalid response")
	errProviderIncomplete  = errors.New("provider stopped before finishing")
)

const maxCompletionBytes = 128 * 1024

// complete asks the provider to stream its answer and assembles the content.
// Streaming matters: the hosted proxy cuts non-streaming requests off at 30s
// (HTTP 504), which long drafts and designs regularly exceed. Providers that
// ignore "stream" and answer with a plain JSON body are still accepted.
func (w *EmailWriter) complete(ctx context.Context, payload map[string]any) (string, error) {
	content, err := w.completeOnce(ctx, payload)
	// The hosted proxy occasionally drops a stream or returns 5xx under load;
	// one retry recovers most of these without the user noticing.
	if (errors.Is(err, errProviderUnavailable) || errors.Is(err, errProviderStatus)) && ctx.Err() == nil {
		select {
		case <-ctx.Done():
			return "", errProviderUnavailable
		case <-time.After(800 * time.Millisecond):
		}
		return w.completeOnce(ctx, payload)
	}
	return content, err
}

func (w *EmailWriter) completeOnce(ctx context.Context, payload map[string]any) (string, error) {
	payload["stream"] = true
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.Endpoint+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	if w.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+w.APIKey)
	}
	resp, err := w.Client.Do(req)
	if err != nil {
		return "", errProviderUnavailable
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errProviderStatus
	}
	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream") {
		return readEnvelope(resp.Body)
	}
	return readStream(resp.Body)
}

func readEnvelope(r io.Reader) (string, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxCompletionBytes+1))
	if err != nil || len(data) > maxCompletionBytes {
		return "", errProviderResponse
	}
	var envelope struct {
		Choices []struct {
			Message      Message `json:"message"`
			FinishReason string  `json:"finish_reason"`
		} `json:"choices"`
	}
	if json.Unmarshal(data, &envelope) != nil || len(envelope.Choices) == 0 {
		return "", errProviderResponse
	}
	if envelope.Choices[0].FinishReason == "length" {
		return "", errProviderIncomplete
	}
	return envelope.Choices[0].Message.Content, nil
}

func readStream(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	var out strings.Builder
	finish := ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			break
		}
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
				FinishReason *string `json:"finish_reason"`
			} `json:"choices"`
		}
		if json.Unmarshal([]byte(data), &chunk) != nil {
			return "", errProviderResponse
		}
		for _, c := range chunk.Choices {
			out.WriteString(c.Delta.Content)
			if c.FinishReason != nil && *c.FinishReason != "" {
				finish = *c.FinishReason
			}
		}
		if out.Len() > maxCompletionBytes {
			return "", errProviderResponse
		}
	}
	if scanner.Err() != nil {
		return "", errProviderUnavailable
	}
	if finish == "length" {
		return "", errProviderIncomplete
	}
	if finish == "" {
		// The connection ended without a finish reason: treat as cut off.
		return "", errProviderIncomplete
	}
	return out.String(), nil
}
