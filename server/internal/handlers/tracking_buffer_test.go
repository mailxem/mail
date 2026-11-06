package handlers

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"kori/internal/config"
	"kori/internal/events"
	"kori/internal/models"
)

func trackingTestDB(t testing.TB) *gorm.DB {
	var dialect gorm.Dialector = sqlite.Open("file:" + uuid.NewString() + "?mode=memory&cache=shared")
	if dsn := os.Getenv("POSTHOOT_TEST_DATABASE_URL"); dsn != "" {
		admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
		require.NoError(t, err)
		schema := "tracking_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
		require.NoError(t, admin.Exec("CREATE SCHEMA "+schema).Error)
		t.Cleanup(func() { admin.Exec("DROP SCHEMA " + schema + " CASCADE"); raw, _ := admin.DB(); raw.Close() })
		dialect = postgres.Open(dsn + " search_path=" + schema)
	}
	db, err := gorm.Open(dialect, &gorm.Config{Logger: logger.Default.LogMode(logger.Silent), DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	raw, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { raw.Close() })
	if db.Dialector.Name() == "sqlite" {
		raw.SetMaxOpenConns(1)
	}
	require.NoError(t, db.AutoMigrate(&models.Contact{}, &models.Campaign{}, &models.Email{}, &models.EmailTracking{}))
	config.GetConfig().JWT.Secret = "tracking-buffer-test-secret"
	return db
}

func seedTrackedEmail(t testing.TB, db *gorm.DB, campaignID string) models.Email {
	email := models.Email{Base: models.Base{ID: uuid.NewString()}, TeamID: uuid.NewString(), ContactID: uuid.NewString(), CampaignID: campaignID, From: "news@example.com", To: "ada@example.net", Subject: "Hi", Body: "x", Status: models.EmailStatusSent, SMTPConfigID: uuid.NewString(), CategoryID: uuid.NewString()}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&email).Error)
	return email
}

func trackingToken(t testing.TB, emailID string) string {
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"mailId": emailID}).SignedString([]byte(config.GetConfig().JWT.Secret))
	require.NoError(t, err)
	return token
}

func hitOpen(t testing.TB, h *TrackingHandler, emailID string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	require.NoError(t, h.HandleEmailOpen(echo.New().NewContext(httptest.NewRequest("GET", "/t/open?token="+trackingToken(t, emailID), nil), rec)))
	return rec
}

func hitClick(t testing.TB, h *TrackingHandler, emailID, target string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	path := "/t/click/" + base64.StdEncoding.EncodeToString([]byte(target)) + "?token=" + trackingToken(t, emailID)
	require.NoError(t, h.HandleEmailClick(echo.New().NewContext(httptest.NewRequest("GET", path, nil), rec)))
	return rec
}

// countStatements counts every query and insert the database runs.
func countStatements(t testing.TB, db *gorm.DB) (queries, inserts *atomic.Int64) {
	queries, inserts = &atomic.Int64{}, &atomic.Int64{}
	name := "count_" + uuid.NewString()
	require.NoError(t, db.Callback().Query().Before("gorm:query").Register(name, func(*gorm.DB) { queries.Add(1) }))
	require.NoError(t, db.Callback().Create().Before("gorm:create").Register(name, func(*gorm.DB) { inserts.Add(1) }))
	return queries, inserts
}

// recordEmits counts email.opened and email.clicked per email for the given emails.
func recordEmits(emails ...models.Email) func(event, emailID string) int {
	watched := map[string]bool{}
	for _, e := range emails {
		watched[e.ID] = true
	}
	var mu sync.Mutex
	counts := map[string]int{}
	for _, event := range []string{"email.opened", "email.clicked"} {
		events.On(event, func(data interface{}) {
			if tr, ok := data.(*models.EmailTracking); ok && watched[tr.EmailID] {
				mu.Lock()
				counts[event+tr.EmailID]++
				mu.Unlock()
			}
		})
	}
	return func(event, emailID string) int {
		mu.Lock()
		defer mu.Unlock()
		return counts[event+emailID]
	}
}

func trackingRows(t testing.TB, db *gorm.DB, where ...interface{}) int64 {
	var n int64
	q := db.Model(&models.EmailTracking{})
	if len(where) > 0 {
		q = q.Where(where[0], where[1:]...)
	}
	require.NoError(t, q.Count(&n).Error)
	return n
}

func TestTrackingHitsAnswerWithoutTouchingTheDatabase(t *testing.T) {
	t.Setenv("TRACKING_FLUSH_INTERVAL", "1h")
	db := trackingTestDB(t)
	email := seedTrackedEmail(t, db, uuid.NewString())
	h := NewTrackingHandler(db)
	queries, inserts := countStatements(t, db)

	open := hitOpen(t, h, email.ID)
	require.Equal(t, http.StatusOK, open.Code)
	require.Equal(t, "image/gif", open.Header().Get("Content-Type"))
	click := hitClick(t, h, email.ID, "https://example.com/pricing?a=1")
	require.Equal(t, http.StatusFound, click.Code)
	require.Equal(t, "https://example.com/pricing?a=1", click.Header().Get("Location"))
	require.Zero(t, queries.Load()+inserts.Load(), "the request path does no database work")

	require.NoError(t, h.Close(context.Background()), "shutdown flushes the in-process buffer")
	var rows []models.EmailTracking
	require.NoError(t, db.Order("event").Find(&rows).Error)
	require.Len(t, rows, 2)
	require.Equal(t, models.EmailTrackingEventClick, rows[0].Event)
	require.Equal(t, "https://example.com/pricing?a=1", rows[0].URL)
	require.Equal(t, email.CampaignID, rows[1].CampaignID)
	require.Equal(t, email.ContactID, rows[1].ContactID)
}

func TestTrackingBatchIsOneInsertAndOpensEmitOncePerEmail(t *testing.T) {
	t.Setenv("TRACKING_FLUSH_INTERVAL", "1h")
	t.Setenv("TRACKING_BATCH_SIZE", "1000")
	db := trackingTestDB(t)
	campaign := func(i int) string {
		// Campaign and transactional mail share one insert; SQLite cannot write DEFAULT in a multi-row insert, Postgres can.
		if i%2 == 0 && db.Dialector.Name() == "postgres" {
			return ""
		}
		return uuid.NewString()
	}
	var fresh []models.Email
	for i := range 10 {
		fresh = append(fresh, seedTrackedEmail(t, db, campaign(i)))
	}
	seen := seedTrackedEmail(t, db, campaign(0))
	require.NoError(t, db.Create(&models.EmailTracking{EmailID: seen.ID, ContactID: seen.ContactID, Event: models.EmailTrackingEventOpen, Timestamp: time.Now()}).Error)
	emits := recordEmits(append(fresh, seen)...)
	h := NewTrackingHandler(db)
	_, inserts := countStatements(t, db)

	for i := range 990 {
		hitOpen(t, h, fresh[i%10].ID) // 99 opens of each fresh email
	}
	for range 9 {
		hitOpen(t, h, seen.ID) // opened before this batch
	}
	hitOpen(t, h, uuid.NewString()) // no such email: dropped, as before

	require.Eventually(t, func() bool { return trackingRows(t, db) == 1000 }, 5*time.Second, 10*time.Millisecond)
	require.EqualValues(t, 1, inserts.Load(), "1,000 hits are one multi-row insert")
	require.Eventually(t, func() bool {
		for _, e := range fresh {
			if emits("email.opened", e.ID) != 1 {
				return false
			}
		}
		return true
	}, 2*time.Second, 10*time.Millisecond, "each fresh email emits its first open once, though it was opened 99 times in the batch")
	require.EqualValues(t, 99, trackingRows(t, db, "email_id = ? AND contact_id = ?", fresh[3].ID, fresh[3].ContactID))

	hitOpen(t, h, fresh[0].ID)
	hitClick(t, h, fresh[0].ID, "https://example.com/a")
	hitClick(t, h, fresh[0].ID, "https://example.com/b")
	require.NoError(t, h.Close(context.Background()))
	require.Eventually(t, func() bool { return emits("email.clicked", fresh[0].ID) == 2 }, 2*time.Second, 10*time.Millisecond, "every click emits")
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, 1, emits("email.opened", fresh[0].ID), "an open in a later batch does not emit again")
	require.Zero(t, emits("email.opened", seen.ID), "an email already opened before the batch does not emit")
	require.EqualValues(t, 1003, trackingRows(t, db))
}

func TestClickRedirectChecks(t *testing.T) {
	t.Setenv("TRACKING_FLUSH_INTERVAL", "1h")
	db := trackingTestDB(t)
	email := seedTrackedEmail(t, db, "")
	h := NewTrackingHandler(db)
	defer h.Close(context.Background())
	call := func(path string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		require.NoError(t, h.HandleEmailClick(echo.New().NewContext(httptest.NewRequest("GET", path, nil), rec)))
		return rec
	}
	link := func(target string) string { return "/t/click/" + base64.StdEncoding.EncodeToString([]byte(target)) }

	require.Equal(t, "http://example.com/x", hitClick(t, h, email.ID, "http://example.com/x").Header().Get("Location"))
	require.Equal(t, http.StatusBadRequest, call(link("https://example.com")).Code, "a link without a token is refused")
	require.Equal(t, http.StatusUnauthorized, call(link("https://example.com")+"?token=forged").Code)
	forged, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"mailId": email.ID}).SignedString([]byte("another-secret"))
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, call(link("https://evil.example")+"?token="+forged).Code, "a token signed elsewhere is refused")
	token := trackingToken(t, email.ID)
	require.Equal(t, http.StatusBadRequest, call("/t/click/!!!?token="+token).Code)
	for _, target := range []string{"javascript:alert(1)", "data:text/html,hi", "//evil.example"} {
		rec := call(link(target) + "?token=" + token)
		require.Equal(t, http.StatusBadRequest, rec.Code, target)
		require.Empty(t, rec.Header().Get("Location"))
	}
}

// fakeStream behaves like the Redis Stream: read hands back unacknowledged entries first (as XAUTOCLAIM does once they go idle), then new ones.
type fakeStream struct {
	mu      sync.Mutex
	queued  []redis.XMessage
	pending []redis.XMessage
	next    int
	addErr  error
	ackErr  error
}

func (f *fakeStream) add(_ context.Context, payload []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.addErr != nil {
		return f.addErr
	}
	f.next++
	f.queued = append(f.queued, redis.XMessage{ID: fmt.Sprintf("%d-0", f.next), Values: map[string]interface{}{"t": string(payload)}})
	return nil
}

func (f *fakeStream) read(_ context.Context, count int64) ([]redis.XMessage, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.pending) > 0 {
		return append([]redis.XMessage(nil), f.pending...), nil
	}
	n := min(int(count), len(f.queued))
	out := f.queued[:n:n]
	f.queued = f.queued[n:]
	f.pending = append(f.pending, out...)
	return out, nil
}

func (f *fakeStream) ack(_ context.Context, ids ...string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.ackErr != nil {
		err := f.ackErr
		f.ackErr = nil
		return err
	}
	acked := map[string]bool{}
	for _, id := range ids {
		acked[id] = true
	}
	kept := f.pending[:0]
	for _, m := range f.pending {
		if !acked[m.ID] {
			kept = append(kept, m)
		}
	}
	f.pending = kept
	return nil
}

func TestTrackingStreamAcksOnlyAfterCommitAndFallsBackWhenDown(t *testing.T) {
	t.Setenv("TRACKING_FLUSH_INTERVAL", "1h")
	db := trackingTestDB(t)
	email := seedTrackedEmail(t, db, uuid.NewString())
	emits := recordEmits(email)
	stream := &fakeStream{}
	h := &TrackingHandler{db: db, buffer: newTrackingBuffer(db, stream)}
	defer h.Close(context.Background())
	queries, inserts := countStatements(t, db)

	hitOpen(t, h, email.ID)
	hitClick(t, h, email.ID, "https://example.com/a")
	require.Zero(t, queries.Load()+inserts.Load(), "hits go to the stream, not the database")
	require.Len(t, stream.queued, 2)

	var failInsert atomic.Bool
	failInsert.Store(true)
	require.NoError(t, db.Callback().Create().Before("gorm:create").Register("fail_insert", func(tx *gorm.DB) {
		if failInsert.Load() {
			tx.AddError(errors.New("database unavailable"))
		}
	}))
	_, err := h.buffer.flushStream(context.Background())
	require.Error(t, err)
	require.Len(t, stream.pending, 2, "a batch that did not commit stays unacknowledged")
	require.Zero(t, trackingRows(t, db))

	failInsert.Store(false)
	stream.ackErr = errors.New("redis blip")
	_, err = h.buffer.flushStream(context.Background())
	require.Error(t, err, "committed, but the ack was lost")
	require.Len(t, stream.pending, 2)
	n, err := h.buffer.flushStream(context.Background())
	require.NoError(t, err)
	require.Equal(t, 2, n)
	require.Empty(t, stream.pending)
	require.EqualValues(t, 2, trackingRows(t, db), "a redelivered batch is not saved twice")
	var saved models.EmailTracking
	require.NoError(t, db.First(&saved, "event = ?", models.EmailTrackingEventOpen).Error)
	require.Equal(t, email.ContactID, saved.ContactID)
	require.Equal(t, email.CampaignID, saved.CampaignID)
	require.Eventually(t, func() bool {
		return emits("email.opened", email.ID) == 1 && emits("email.clicked", email.ID) == 1
	}, 2*time.Second, 10*time.Millisecond)
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, 1, emits("email.clicked", email.ID), "a redelivered batch does not emit twice")

	stream.addErr = errors.New("connection refused")
	hitClick(t, h, email.ID, "https://example.com/b")
	require.EqualValues(t, 1, trackingRows(t, db, "url = ?", "https://example.com/b"), "with Redis down the hit is saved before the redirect")
}

// BenchmarkTrackingOpen measures opens end to end, including saving every buffered hit.
func BenchmarkTrackingOpen(b *testing.B) {
	db := trackingTestDB(b)
	require.NoError(b, db.Exec("CREATE INDEX IF NOT EXISTS analytics_tracking_message ON email_trackings(email_id,timestamp,event)").Error)
	email := seedTrackedEmail(b, db, uuid.NewString())
	h := NewTrackingHandler(db)
	token := trackingToken(b, email.ID)
	b.ResetTimer()
	for range b.N {
		if err := h.HandleEmailOpen(echo.New().NewContext(httptest.NewRequest("GET", "/t/open?token="+token, nil), httptest.NewRecorder())); err != nil {
			b.Fatal(err)
		}
	}
	require.NoError(b, h.Close(context.Background()))
}

// BenchmarkTrackingOpenStream measures the handler when hits go to the stream (Redis replaced by a fake).
func BenchmarkTrackingOpenStream(b *testing.B) {
	b.Setenv("TRACKING_FLUSH_INTERVAL", "1h")
	db := trackingTestDB(b)
	email := seedTrackedEmail(b, db, uuid.NewString())
	h := &TrackingHandler{db: db, buffer: newTrackingBuffer(db, discardStream{})}
	b.Cleanup(func() { h.Close(context.Background()) })
	token := trackingToken(b, email.ID)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		e := echo.New()
		for pb.Next() {
			if err := h.HandleEmailOpen(e.NewContext(httptest.NewRequest("GET", "/t/open?token="+token, nil), httptest.NewRecorder())); err != nil {
				b.Fatal(err)
			}
		}
	})
}

type discardStream struct{}

func (discardStream) add(context.Context, []byte) error { return nil }
func (discardStream) read(context.Context, int64) ([]redis.XMessage, error) {
	return nil, nil
}
func (discardStream) ack(context.Context, ...string) error { return nil }

// TestRedisTrackingStream runs the stream against a real Redis when TRACKING_TEST_REDIS_ADDR is set.
func TestRedisTrackingStream(t *testing.T) {
	addr := os.Getenv("TRACKING_TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("TRACKING_TEST_REDIS_ADDR not set")
	}
	t.Setenv("TRACKING_FLUSH_INTERVAL", "50ms")
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	defer rdb.Close()
	require.NoError(t, rdb.Del(ctx, trackingStreamKey).Err())
	db := trackingTestDB(t)
	email := seedTrackedEmail(t, db, uuid.NewString())

	// A writer read a hit and stopped before acknowledging it, long enough ago to be taken over.
	crashed, err := newRedisTrackingStream(rdb)
	require.NoError(t, err)
	crashed.consumer = "crashed"
	require.NoError(t, crashed.add(ctx, []byte(`{"id":"`+uuid.NewString()+`","emailId":"`+email.ID+`","event":"open"}`)))
	msgs, err := crashed.read(ctx, 10)
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	require.NoError(t, rdb.Do(ctx, "XCLAIM", trackingStreamKey, trackingStreamGroup, "crashed", 0, msgs[0].ID, "IDLE", (2*trackingClaimIdle).Milliseconds()).Err())

	h := NewTrackingHandler(db, rdb)
	hitClick(t, h, email.ID, "https://example.com/a")
	require.Eventually(t, func() bool { return trackingRows(t, db) == 2 }, 5*time.Second, 20*time.Millisecond, "the stale hit is taken over and the new one saved")
	require.NoError(t, h.Close(ctx))
	pending, err := rdb.XPending(ctx, trackingStreamKey, trackingStreamGroup).Result()
	require.NoError(t, err)
	require.Zero(t, pending.Count, "everything saved is acknowledged")
}
