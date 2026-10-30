package dodo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// EmailMeterEvent is the event name of the "Emails sent" meter that
// cmd/setup-email-plans creates; its sum over metadata.emails is billed.
const EmailMeterEvent = "email.sent"

// IngestEmails reports emails sent by a customer to the Dodo meter. Dodo
// deduplicates by eventID, so retrying with the same ID is safe. Events must
// arrive within an hour of the timestamp.
func (c *Client) IngestEmails(ctx context.Context, eventID, customerID string, emails int, at time.Time) error {
	key := os.Getenv("DODO_PAYMENTS_API_KEY")
	if key == "" {
		key = os.Getenv("DODO_API_KEY")
	}
	body, _ := json.Marshal(map[string]any{"events": []map[string]any{{
		"event_id": eventID, "customer_id": customerID, "event_name": EmailMeterEvent,
		"timestamp": at.UTC().Format(time.RFC3339), "metadata": map[string]any{"emails": emails},
	}}})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/events/ingest", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(res.Body, 512))
		return fmt.Errorf("dodo usage ingest: %d %s", res.StatusCode, raw)
	}
	return nil
}

// WithBaseURL points the client at another Dodo host, for tests.
func (c *Client) WithBaseURL(u string) *Client {
	c.baseURL = u
	return c
}
