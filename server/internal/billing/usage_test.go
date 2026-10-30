package billing

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestEmailsAreCountedPerMinuteAndReportedOnce(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&EmailUsage{}))
	team := "11111111-1111-1111-1111-111111111111"
	at := time.Date(2026, 10, 30, 10, 15, 20, 0, time.UTC)

	Enabled = false
	require.NoError(t, CountEmails(db, team, 5, at))
	var rows int64
	require.NoError(t, db.Model(&EmailUsage{}).Count(&rows).Error)
	require.Zero(t, rows, "self-hosted installs must not count")

	Enabled = true
	t.Cleanup(func() { Enabled = false })
	require.NoError(t, CountEmails(db, team, 3, at))
	require.NoError(t, CountEmails(db, team, 2, at.Add(30*time.Second)))
	require.NoError(t, CountEmails(db, team, 7, at.Add(time.Minute))) // the next, unfinished minute

	var got []map[string]any
	fail := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "secret", r.Header.Get("X-API-Key"))
		require.Equal(t, "/api/v1/usage/record/"+team, r.URL.Path)
		if fail {
			w.WriteHeader(500)
			return
		}
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		got = append(got, body)
	}))
	defer server.Close()
	r := &Reporter{DB: db, URL: server.URL + "/api/v1", Key: "secret", HTTP: server.Client(), Now: func() time.Time { return at.Add(time.Minute) }}

	require.Error(t, r.Flush(context.Background()), "a failed report stays pending")
	fail = false
	require.NoError(t, r.Flush(context.Background()))
	require.NoError(t, r.Flush(context.Background())) // nothing left to send
	require.Len(t, got, 1)
	require.EqualValues(t, 5, got[0]["usage"])
	require.Equal(t, "emails_per_month", got[0]["feature"])
	require.Equal(t, "xem-emails-"+team+"-1793355300", got[0]["event_id"])
}
