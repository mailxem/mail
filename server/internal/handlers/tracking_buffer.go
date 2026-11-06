package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"kori/internal/events"
	"kori/internal/models"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	trackingStreamKey   = "xem:tracking:hits"
	trackingStreamGroup = "tracking-writers"
	trackingStreamCap   = 1_000_000   // approximate MAXLEN; a backlog this deep means the writers are down
	trackingClaimIdle   = time.Minute // hits a writer read but never acknowledged are taken over after this
	trackingChannelCap  = 100_000     // in-process buffer when Redis is not configured
	trackingAddTimeout  = 250 * time.Millisecond
)

// trackingStream is the Redis Stream opens and clicks queue on; tests use a fake.
type trackingStream interface {
	add(ctx context.Context, payload []byte) error
	read(ctx context.Context, count int64) ([]redis.XMessage, error)
	ack(ctx context.Context, ids ...string) error
}

type redisTrackingStream struct {
	rdb      *redis.Client
	consumer string
}

func newRedisTrackingStream(rdb *redis.Client) (*redisTrackingStream, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.XGroupCreateMkStream(ctx, trackingStreamKey, trackingStreamGroup, "0").Err(); err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return nil, err
	}
	host, _ := os.Hostname()
	return &redisTrackingStream{rdb: rdb, consumer: fmt.Sprintf("%s-%d", host, os.Getpid())}, nil
}

func (s *redisTrackingStream) add(ctx context.Context, payload []byte) error {
	return s.rdb.XAdd(ctx, &redis.XAddArgs{Stream: trackingStreamKey, MaxLen: trackingStreamCap, Approx: true, Values: []any{"t", payload}}).Err()
}

// read first takes over hits a stopped or crashed writer left unacknowledged, then reads new ones without blocking.
func (s *redisTrackingStream) read(ctx context.Context, count int64) ([]redis.XMessage, error) {
	claimed, _, err := s.rdb.XAutoClaim(ctx, &redis.XAutoClaimArgs{Stream: trackingStreamKey, Group: trackingStreamGroup, Consumer: s.consumer, MinIdle: trackingClaimIdle, Start: "0-0", Count: count}).Result()
	if err != nil || len(claimed) > 0 {
		return claimed, err
	}
	streams, err := s.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{Group: trackingStreamGroup, Consumer: s.consumer, Streams: []string{trackingStreamKey, ">"}, Count: count, Block: -1}).Result()
	if err == redis.Nil || len(streams) == 0 {
		return nil, nil
	}
	return streams[0].Messages, err
}

func (s *redisTrackingStream) ack(ctx context.Context, ids ...string) error {
	return s.rdb.XAck(ctx, trackingStreamKey, trackingStreamGroup, ids...).Err()
}

// trackingBuffer takes opens and clicks off the request path and saves them in batches.
type trackingBuffer struct {
	db       *gorm.DB
	stream   trackingStream // nil: the in-process channel holds hits instead
	hits     chan *models.EmailTracking
	size     int
	interval time.Duration
	stop     context.CancelFunc
	done     chan struct{}
}

func newTrackingBuffer(db *gorm.DB, stream trackingStream) *trackingBuffer {
	size, _ := strconv.Atoi(os.Getenv("TRACKING_BATCH_SIZE"))
	if size <= 0 {
		size = 1000
	}
	interval, _ := time.ParseDuration(os.Getenv("TRACKING_FLUSH_INTERVAL"))
	if interval <= 0 {
		interval = time.Second
	}
	b := &trackingBuffer{db: db, stream: stream, hits: make(chan *models.EmailTracking, trackingChannelCap), size: size, interval: interval, done: make(chan struct{})}
	ctx, stop := context.WithCancel(context.Background())
	b.stop = stop
	if stream != nil {
		go b.consumeStream(ctx)
	} else {
		go b.consumeChannel(ctx)
	}
	return b
}

// add queues a hit; when the queue cannot take it, the hit is saved right away so none is lost.
func (b *trackingBuffer) add(t *models.EmailTracking) error {
	if t.ID == "" {
		t.ID = uuid.NewString() // a retried stream batch skips rows already saved
	}
	if b.stream != nil {
		payload, err := json.Marshal(t)
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), trackingAddTimeout)
			err = b.stream.add(ctx, payload)
			cancel()
		}
		if err == nil {
			return nil
		}
		trackingLog.Warn("Tracking stream unavailable, saving the hit directly: %v", err)
	} else {
		select {
		case b.hits <- t:
			return nil
		default:
		}
	}
	return b.write([]*models.EmailTracking{t})
}

// close stops the writer; the in-process buffer is flushed first. Stream hits stay in Redis for the next writer.
func (b *trackingBuffer) close(ctx context.Context) error {
	b.stop()
	select {
	case <-b.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (b *trackingBuffer) consumeChannel(ctx context.Context) {
	defer close(b.done)
	tick := time.NewTicker(b.interval)
	defer tick.Stop()
	var batch []*models.EmailTracking
	flush := func() {
		if len(batch) > 0 {
			if err := b.write(batch); err != nil {
				trackingLog.Error("Failed to save tracking batch: %v", err)
			}
			batch = nil
		}
	}
	for {
		select {
		case t := <-b.hits:
			if batch = append(batch, t); len(batch) >= b.size {
				flush()
			}
		case <-tick.C:
			flush()
		case <-ctx.Done():
			for {
				select {
				case t := <-b.hits:
					if batch = append(batch, t); len(batch) >= b.size {
						flush()
					}
				default:
					flush()
					return
				}
			}
		}
	}
}

func (b *trackingBuffer) consumeStream(ctx context.Context) {
	defer close(b.done)
	tick := time.NewTicker(b.interval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
		}
		// Drain whatever piled up since the last tick, a full batch at a time.
		for ctx.Err() == nil {
			n, err := b.flushStream(ctx)
			if err != nil {
				trackingLog.Error("Failed to save tracking batch, it stays pending and is retried: %v", err)
			}
			if err != nil || n < b.size {
				break
			}
		}
	}
}

// flushStream saves one batch from the stream and acknowledges it only after the insert commits.
func (b *trackingBuffer) flushStream(ctx context.Context) (int, error) {
	msgs, err := b.stream.read(ctx, int64(b.size))
	if err != nil || len(msgs) == 0 {
		return 0, err
	}
	rows := make([]*models.EmailTracking, 0, len(msgs))
	ids := make([]string, 0, len(msgs))
	for _, m := range msgs {
		ids = append(ids, m.ID)
		raw, _ := m.Values["t"].(string)
		var t models.EmailTracking
		if err := json.Unmarshal([]byte(raw), &t); err != nil {
			trackingLog.Warn("Dropping unreadable tracking hit %s: %v", m.ID, err)
			continue
		}
		if string(t.Metadata) == "null" {
			t.Metadata = nil // keep the column default, as a direct insert does
		}
		rows = append(rows, &t)
	}
	if err := b.write(rows); err != nil {
		return len(msgs), err
	}
	return len(msgs), b.stream.ack(context.Background(), ids...)
}

// write saves a batch in one multi-row insert, then emits email.clicked for every click and
// email.opened for the first open of each email.
func (b *trackingBuffer) write(rows []*models.EmailTracking) error {
	if len(rows) == 0 {
		return nil
	}
	emailIDs := make([]string, 0, len(rows))
	rowIDs := make([]string, 0, len(rows))
	for _, t := range rows {
		emailIDs = append(emailIDs, t.EmailID)
		rowIDs = append(rowIDs, t.ID)
	}
	var emails []models.Email
	if err := b.db.Select("id", "campaign_id", "contact_id", "sent_at").Where("id IN ?", emailIDs).Find(&emails).Error; err != nil {
		return err
	}
	var saved []string
	if err := b.db.Model(&models.EmailTracking{}).Where("id IN ?", rowIDs).Pluck("id", &saved).Error; err != nil {
		return err
	}
	byID := make(map[string]models.Email, len(emails))
	for _, e := range emails {
		byID[e.ID] = e
	}
	skip := make(map[string]bool, len(saved))
	for _, id := range saved {
		skip[id] = true
	}

	kept := make([]*models.EmailTracking, 0, len(rows))
	var opens []string
	for _, t := range rows {
		email, ok := byID[t.EmailID]
		if !ok || skip[t.ID] { // unknown email, or saved by an earlier attempt at this batch
			continue
		}
		t.CampaignID, t.ContactID = email.CampaignID, email.ContactID
		// A click within seconds of delivery is a link scanner, not a reader.
		if t.Machine == "" && t.Event == models.EmailTrackingEventClick && !email.SentAt.IsZero() && t.Timestamp.Sub(email.SentAt) < clickAfterDelivery {
			t.Machine, t.MachineReason = models.TrackingMachineBot, "clicked within seconds of delivery"
		}
		kept = append(kept, t)
		if t.Event == models.EmailTrackingEventOpen {
			opens = append(opens, t.EmailID)
		}
	}
	if len(kept) == 0 {
		return nil
	}

	// Only the first open of an email starts automations and webhooks; image proxies and re-reads reload the pixel.
	opened := map[string]bool{}
	if len(opens) > 0 {
		var earlier []string
		if err := b.db.Model(&models.EmailTracking{}).Where("event = ? AND email_id IN ?", models.EmailTrackingEventOpen, opens).Distinct().Pluck("email_id", &earlier).Error; err != nil {
			return err
		}
		for _, id := range earlier {
			opened[id] = true
		}
	}

	if err := b.db.CreateInBatches(kept, 1000).Error; err != nil {
		return err
	}
	// Several links in one email clicked within a second is a link scanner.
	// ponytail: one query per click; batch it per email if clicks dominate.
	for _, t := range kept {
		if t.Event == models.EmailTrackingEventClick {
			if err := flagClickBurst(b.db, t); err != nil {
				return err
			}
		}
	}

	for _, t := range kept {
		if t.Event == models.EmailTrackingEventClick {
			events.Emit("email.clicked", t)
		} else if !opened[t.EmailID] {
			opened[t.EmailID] = true
			events.Emit("email.opened", t)
		}
	}
	return nil
}
