package sending

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"kori/internal/models"
)

const testTopic = "arn:aws:sns:us-east-1:123456789012:xem-feedback"

type fakeEventQueue struct {
	batches [][]sqstypes.Message
	deleted []string
}

func (q *fakeEventQueue) ReceiveMessage(context.Context, *sqs.ReceiveMessageInput, ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
	out := &sqs.ReceiveMessageOutput{}
	if len(q.batches) > 0 {
		out.Messages, q.batches = q.batches[0], q.batches[1:]
	}
	return out, nil
}
func (q *fakeEventQueue) DeleteMessageBatch(_ context.Context, in *sqs.DeleteMessageBatchInput, _ ...func(*sqs.Options)) (*sqs.DeleteMessageBatchOutput, error) {
	for _, e := range in.Entries {
		q.deleted = append(q.deleted, aws.ToString(e.ReceiptHandle))
	}
	return &sqs.DeleteMessageBatchOutput{}, nil
}

func queueService(t *testing.T) (*Service, string, Domain, *fakeEventQueue) {
	s, team, d, _ := setup(t)
	q := &fakeEventQueue{}
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	s.Config.TopicARN, s.Events, s.Now = testTopic, q, func() time.Time { return now }
	enableAlerts(t, s)
	return s, team, d, q
}

// sentMessages stores messages SES has accepted, each with its source email.
func sentMessages(t *testing.T, s *Service, team string, d Domain, ids []string, test, noProvider bool) []Message {
	t.Helper()
	ms := make([]Message, len(ids))
	emails := make([]models.Email, len(ids))
	for i, id := range ids {
		emails[i] = models.Email{Base: models.Base{ID: uuid.NewSHA1(uuid.NameSpaceOID, []byte(id)).String()}, From: "hello@example.com", To: "reader@example.net", Subject: "Hello", Body: "Hello", Status: "SENDING", TeamID: team, SMTPConfigID: d.ID, CategoryID: uuid.NewString()}
		ms[i] = Message{ID: id, TeamID: team, DomainID: d.ID, EmailID: emails[i].ID, IdempotencyKey: uuid.NewString(), From: "hello@example.com", Recipients: "reader@example.net", Subject: "Hello", Status: "SENT", ProviderID: "ses-" + id, IsTest: test && i == 0}
		if noProvider && i == 0 {
			ms[i].ProviderID = ""
		}
		if ms[i].IsTest {
			ms[i].EmailID = ""
		}
	}
	db := s.DB.Session(&gorm.Session{SkipHooks: true})
	require.NoError(t, db.CreateInBatches(&emails, 200).Error)
	require.NoError(t, db.CreateInBatches(&ms, 200).Error)
	return ms
}

func sesEvent(m Message, kind string) []byte {
	recipients := []map[string]string{{"emailAddress": "reader@example.net"}}
	raw, _ := json.Marshal(map[string]any{
		"eventType": kind,
		"mail":      map[string]any{"messageId": "ses-" + m.ID, "source": m.From, "tags": map[string][]string{"xem_message": {m.ID}, "ses:configuration-set": {tenant(m.TeamID)}}},
		"bounce":    map[string]any{"bounceType": "Permanent", "bouncedRecipients": recipients},
		"complaint": map[string]any{"complainedRecipients": recipients},
	})
	return raw
}

func envelope(id string, raw []byte) sqstypes.Message {
	body, _ := json.Marshal(Notification{Type: "Notification", MessageID: id, TopicARN: testTopic, Message: string(raw)})
	return sqstypes.Message{Body: aws.String(string(body)), ReceiptHandle: aws.String(id)}
}

type feedbackState struct {
	Messages []struct{ ID, Status, Detail, ProviderID string }
	Events   []struct {
		ID, MessageID, Kind, Detail string
		CreatedAt                   time.Time
	}
	Emails []struct {
		ID, Status string
		SentAt     time.Time
	}
	Milestones, Notices, NoticeEvents, Suppressions, Audits []string
	Suspended                                               bool
}

// stateOf reads everything feedback writes, without workspace and domain IDs,
// which differ between the two databases compared.
func stateOf(t *testing.T, s *Service, team string) feedbackState {
	t.Helper()
	var st feedbackState
	require.NoError(t, s.DB.Model(&Message{}).Select("id, status, detail, provider_id").Order("id").Scan(&st.Messages).Error)
	require.NoError(t, s.DB.Model(&Event{}).Select("id, message_id, kind, detail, created_at").Order("id").Scan(&st.Events).Error)
	require.NoError(t, s.DB.Model(&models.Email{}).Select("id, status, sent_at").Order("id").Scan(&st.Emails).Error)
	for i := range st.Emails {
		st.Emails[i].SentAt = st.Emails[i].SentAt.UTC()
	}
	for i := range st.Events {
		st.Events[i].CreatedAt = st.Events[i].CreatedAt.UTC()
	}
	require.NoError(t, s.DB.Model(&MilestoneEmail{}).Order("kind").Pluck("kind || ':' || status", &st.Milestones).Error)
	require.NoError(t, s.DB.Model(&models.ServiceNotice{}).Order("kind").Pluck("kind || ':' || status || ':' || event_count", &st.Notices).Error)
	require.NoError(t, s.DB.Model(&models.ServiceNoticeEvent{}).Order("kind, reason").Pluck("kind || ':' || reason", &st.NoticeEvents).Error)
	require.NoError(t, s.DB.Model(&Suppression{}).Order("address").Pluck("address || ':' || reason", &st.Suppressions).Error)
	require.NoError(t, s.DB.Model(&Audit{}).Order("action").Pluck("action", &st.Audits).Error)
	a, err := s.Account(context.Background(), team)
	require.NoError(t, err)
	st.Suspended = a.Suspended
	return st
}

func TestQueuedFeedbackWritesWhatApplyFeedbackWrites(t *testing.T) {
	type step struct {
		event string
		msg   int
		kind  string
	}
	for _, sc := range []struct {
		name             string
		test, noProvider bool
		steps            []step
	}{
		{"delivery", false, false, []step{{"e1", 0, "Delivery"}}},
		{"delay then delivery", false, false, []step{{"e1", 0, "DeliveryDelay"}, {"e2", 0, "Delivery"}}},
		{"delay after delivery does not downgrade", false, false, []step{{"e1", 0, "Delivery"}, {"e2", 0, "DeliveryDelay"}}},
		{"duplicate event in one batch", false, false, []step{{"e1", 0, "Delivery"}, {"e1", 0, "Delivery"}}},
		{"late send event", false, false, []step{{"e1", 0, "Delivery"}, {"e2", 0, "Send"}}},
		{"test message milestones", true, false, []step{{"e1", 0, "DeliveryDelay"}, {"e2", 0, "Delivery"}}},
		{"provider id recorded", false, true, []step{{"e1", 0, "Delivery"}}},
		{"several messages", false, false, []step{{"e1", 0, "Delivery"}, {"e2", 1, "Delivery"}, {"e3", 2, "DeliveryDelay"}, {"e4", 2, "Delivery"}}},
		{"delay then permanent bounce", false, false, []step{{"e1", 0, "DeliveryDelay"}, {"e2", 0, "Bounce"}}},
		{"delivery and complaint", false, false, []step{{"e1", 0, "Delivery"}, {"e2", 1, "Complaint"}}},
		{"reject", false, false, []step{{"e1", 0, "Reject"}}},
		{"unsubscribed event type", false, false, []step{{"e1", 0, "Open"}}},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ctx := context.Background()
			ids := []string{uuid.NewString(), uuid.NewString(), uuid.NewString()}
			run := func(batched bool) feedbackState {
				s, team, d, q := queueService(t)
				ms := sentMessages(t, s, team, d, ids, sc.test, sc.noProvider)
				batch := []sqstypes.Message{}
				for _, st := range sc.steps {
					raw := sesEvent(ms[st.msg], st.kind)
					if batched {
						batch = append(batch, envelope(st.event, raw))
					} else {
						require.NoError(t, s.ApplyFeedback(ctx, st.event, raw))
					}
				}
				if !batched {
					return stateOf(t, s, team)
				}
				q.batches = [][]sqstypes.Message{batch, batch} // the second is a redelivery
				require.NoError(t, s.PollEvents(ctx))
				first := stateOf(t, s, team)
				require.NoError(t, s.PollEvents(ctx))
				require.Equal(t, first, stateOf(t, s, team), "redelivery must change nothing")
				require.Len(t, q.deleted, 2*len(batch), "every applied event is acknowledged")
				return first
			}
			single, batched := run(false), run(true)
			require.Equal(t, single, batched)
			if sc.name == "delivery and complaint" {
				require.True(t, batched.Suspended, "complaints still go through ApplyFeedback")
			}
		})
	}
}

func TestQueuedFeedbackRejectsForgedEventsAlone(t *testing.T) {
	s, team, d, q := queueService(t)
	ctx := context.Background()
	ms := sentMessages(t, s, team, d, []string{uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()}, false, false)
	other := ms[1]
	other.TeamID = uuid.NewString()
	spoofed := ms[2]
	spoofed.From = "attacker@example.com"
	wrongProvider := sesEvent(ms[3], "Delivery")
	wrongProvider = []byte(strings.Replace(string(wrongProvider), "ses-"+ms[3].ID, "ses-other", 1))
	missing := ms[0]
	missing.ID = uuid.NewString()
	notUUID := ms[0]
	notUUID.ID = "not-a-uuid"
	wrongTopic := envelope("wrong-topic", sesEvent(ms[0], "Delivery"))
	wrongTopic.Body = aws.String(strings.Replace(aws.ToString(wrongTopic.Body), testTopic, testTopic+"-other", 1))
	q.batches = [][]sqstypes.Message{{
		envelope("good", sesEvent(ms[0], "Delivery")),
		envelope("wrong-tenant", sesEvent(other, "Delivery")),
		envelope("wrong-source", sesEvent(spoofed, "Delivery")),
		envelope("wrong-provider", wrongProvider),
		envelope("missing-message", sesEvent(missing, "Delivery")),
		envelope("not-a-uuid", sesEvent(notUUID, "Delivery")),
		wrongTopic,
		{Body: aws.String("{"), ReceiptHandle: aws.String("malformed")},
	}}
	require.NoError(t, s.PollEvents(ctx))
	require.Equal(t, []string{"good"}, q.deleted)
	st := stateOf(t, s, team)
	require.Len(t, st.Events, 1)
	for _, m := range st.Messages {
		want := "SENT"
		if m.ID == ms[0].ID {
			want = "DELIVERED"
		}
		require.Equal(t, want, m.Status, m.ID)
	}
}

func TestFailedDeliveryBatchIsRetriedNotAcknowledged(t *testing.T) {
	s, team, d, q := queueService(t)
	ctx := context.Background()
	ms := sentMessages(t, s, team, d, []string{uuid.NewString(), uuid.NewString()}, false, false)
	batch := []sqstypes.Message{envelope("e1", sesEvent(ms[0], "Delivery")), envelope("e2", sesEvent(ms[1], "Delivery"))}
	q.batches = [][]sqstypes.Message{batch, batch}
	require.NoError(t, s.DB.Migrator().DropTable(&Event{}))
	require.NoError(t, s.PollEvents(ctx))
	require.Empty(t, q.deleted)
	require.NoError(t, s.DB.AutoMigrate(&Event{}))
	for _, m := range stateOf(t, s, team).Messages {
		require.Equal(t, "SENT", m.Status, "a failed batch writes nothing")
	}
	require.NoError(t, s.PollEvents(ctx)) // SQS redelivers after the visibility timeout
	require.ElementsMatch(t, []string{"e1", "e2"}, q.deleted)
}

func TestEventQueueConfiguration(t *testing.T) {
	t.Setenv("MANAGED_SENDING_ENABLED", "true")
	t.Setenv("MANAGED_SES_REGION", "us-east-1")
	t.Setenv("MANAGED_AWS_ACCOUNT_ID", "123456789012")
	t.Setenv("MANAGED_SNS_TOPIC_ARN", testTopic)
	t.Setenv("SES_EVENTS_QUEUE_URL", "https://sqs.us-east-1.amazonaws.com/123456789012/xem-feedback")
	c, err := LoadConfig()
	require.NoError(t, err)
	require.EqualValues(t, 4, c.EventsPollers)
	t.Setenv("SES_EVENTS_QUEUE_URL", "https://sqs.us-east-1.amazonaws.com/999999999999/xem-feedback")
	_, err = LoadConfig()
	require.Error(t, err, "a queue in another account is refused")
}

func TestProvisionDoesNotSubscribeToSendEvents(t *testing.T) {
	var mu sync.Mutex
	destinations := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if strings.HasSuffix(r.URL.Path, "/event-destinations") {
			mu.Lock()
			destinations = append(destinations, string(body))
			mu.Unlock()
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()
	client := sesv2.NewFromConfig(aws.Config{Region: "us-east-1", Credentials: aws.AnonymousCredentials{}, BaseEndpoint: aws.String(server.URL)})
	p := &SES{client: client, config: Config{Region: "us-east-1", AccountID: "123456789012", TopicARN: testTopic}}
	require.NoError(t, p.Provision(context.Background(), uuid.NewString(), "example.com"))
	require.Len(t, destinations, 1)
	require.Contains(t, destinations[0], `"DELIVERY"`)
	require.NotContains(t, destinations[0], `"SEND"`)
}

func TestDomainRefreshUpdatesOlderEventDestinations(t *testing.T) {
	s, team, d, p := setup(t)
	ctx := context.Background()
	require.NoError(t, s.DB.Model(&Domain{}).Where("id = ?", d.ID).Update("events_version", 0).Error)
	p.provisions = 0
	_, err := s.RefreshDomain(ctx, team, d.ID)
	require.NoError(t, err)
	_, err = s.RefreshDomain(ctx, team, d.ID)
	require.NoError(t, err)
	require.Equal(t, 1, p.provisions, "an older configuration set is updated once")
	require.NoError(t, s.DB.First(&d, "id = ?", d.ID).Error)
	require.Equal(t, eventDestinationVersion, d.EventsVersion)
}

// Run with -v for the numbers. One transaction per batch of ten instead of one
// per event.
func TestDeliveryBatchesVersusOneByOne(t *testing.T) {
	const n = 1000
	ctx := context.Background()
	measure := func(batched bool) (statements int64, transactions int, elapsed time.Duration) {
		s, team, d, q := queueService(t)
		ids := make([]string, n)
		for i := range ids {
			ids[i] = uuid.NewString()
		}
		ms := sentMessages(t, s, team, d, ids, false, false)
		count := func(*gorm.DB) { atomic.AddInt64(&statements, 1) }
		cb := s.DB.Callback()
		require.NoError(t, cb.Create().After("gorm:create").Register("count", count))
		require.NoError(t, cb.Query().After("gorm:query").Register("count", count))
		require.NoError(t, cb.Update().After("gorm:update").Register("count", count))
		require.NoError(t, cb.Raw().After("gorm:raw").Register("count", count))
		start := time.Now()
		if batched {
			for i := 0; i < n; i += 10 {
				batch := []sqstypes.Message{}
				for _, m := range ms[i : i+10] {
					batch = append(batch, envelope("event-"+m.ID, sesEvent(m, "Delivery")))
				}
				q.batches = [][]sqstypes.Message{batch}
				require.NoError(t, s.PollEvents(ctx))
				transactions++
			}
			require.Len(t, q.deleted, n)
		} else {
			for _, m := range ms {
				require.NoError(t, s.ApplyFeedback(ctx, "event-"+m.ID, sesEvent(m, "Delivery")))
				transactions++
			}
		}
		elapsed = time.Since(start)
		var delivered int64
		require.NoError(t, s.DB.Model(&Message{}).Where("status = ?", "DELIVERED").Count(&delivered).Error)
		require.EqualValues(t, n, delivered)
		require.NoError(t, s.DB.Model(&models.Email{}).Where("status = ?", "SENT").Count(&delivered).Error)
		require.EqualValues(t, n, delivered)
		return
	}
	oneStatements, oneTx, oneTime := measure(false)
	batchStatements, batchTx, batchTime := measure(true)
	t.Log(fmt.Sprintf("%d Delivery events one by one: %d transactions, %d statements, %v", n, oneTx, oneStatements, oneTime))
	t.Log(fmt.Sprintf("%d Delivery events in batches of 10: %d transactions, %d statements, %v", n, batchTx, batchStatements, batchTime))
	require.Less(t, batchStatements, oneStatements)
}
