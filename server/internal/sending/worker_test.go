package sending

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
	"kori/internal/models"
)

func queueMessages(t *testing.T, s *Service, team, domain string, n int, from time.Time) []string {
	t.Helper()
	ids := make([]string, n)
	for i := range ids {
		ids[i] = uuid.NewString()
		require.NoError(t, s.DB.Create(&Message{ID: ids[i], TeamID: team, DomainID: domain, IdempotencyKey: ids[i], From: "hello@example.com", Recipients: "reader@example.net", Raw: []byte("Subject: Hi\r\n\r\nHi"), Status: "QUEUED", CreatedAt: from.Add(time.Duration(i) * time.Second)}).Error)
	}
	return ids
}

func statusCounts(t *testing.T, s *Service) map[string]int {
	t.Helper()
	var ms []Message
	require.NoError(t, s.DB.Find(&ms).Error)
	out := map[string]int{}
	for _, m := range ms {
		out[m.Status]++
	}
	return out
}

func TestConcurrentClaimsAreBoundedAndDisjoint(t *testing.T) {
	s, team, d, _ := setup(t)
	queueMessages(t, s, team, d.ID, 60, time.Now().UTC().Add(-time.Hour))
	var mu sync.Mutex
	seen := map[string]int{}
	var wg sync.WaitGroup
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				claimed, err := s.claim(context.Background(), 7)
				if !assert.NoError(t, err) || !assert.LessOrEqual(t, len(claimed), 7) {
					return
				}
				mu.Lock()
				for _, m := range claimed {
					seen[m.ID]++
				}
				mu.Unlock()
				if len(claimed) == 0 {
					var left int64
					if !assert.NoError(t, s.DB.Model(&Message{}).Where("status = ?", "QUEUED").Count(&left).Error) || left == 0 {
						return
					}
				}
			}
		}()
	}
	wg.Wait()
	require.Len(t, seen, 60)
	for id, n := range seen {
		require.Equal(t, 1, n, id)
	}
	require.Equal(t, map[string]int{"SENDING": 60}, statusCounts(t, s))
}

func TestClaimTakesTurnsAcrossWorkspaces(t *testing.T) {
	s, team, d, _ := setup(t)
	other := uuid.NewString()
	require.NoError(t, s.DB.Create(&Account{TeamID: other, Approved: true}).Error)
	otherDomain := Domain{ID: uuid.NewString(), TeamID: other, Name: "other.example", Ready: true}
	require.NoError(t, s.DB.Create(&otherDomain).Error)
	base := time.Now().UTC().Add(-time.Hour)
	big := queueMessages(t, s, team, d.ID, 5, base)                               // a large send queued first
	small := queueMessages(t, s, other, otherDomain.ID, 2, base.Add(time.Minute)) // a later, small send

	claimed, err := s.claim(context.Background(), 4)
	require.NoError(t, err)
	var got []string
	for _, m := range claimed {
		got = append(got, m.ID)
	}
	require.ElementsMatch(t, []string{big[0], big[1], small[0], small[1]}, got)
	claimed, err = s.claim(context.Background(), 100)
	require.NoError(t, err)
	require.Len(t, claimed, 3)
}

func TestDeliveryPoolIsPacedByTheSharedLimiter(t *testing.T) {
	s, team, d, p := setup(t)
	queueMessages(t, s, team, d.ID, 21, time.Now().UTC().Add(-time.Hour))
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	start := time.Now()
	// Burst 1 at 40/s: 21 sends need at least 20 refills, i.e. 500ms.
	go func() { defer close(done); s.deliver(ctx, rate.NewLimiter(40, 1), 4, 5) }()
	require.Eventually(t, func() bool { p.mu.Lock(); defer p.mu.Unlock(); return p.sends == 21 }, 5*time.Second, 5*time.Millisecond)
	elapsed := time.Since(start)
	cancel()
	<-done
	require.GreaterOrEqual(t, elapsed, 480*time.Millisecond)
	require.Less(t, elapsed, 3*time.Second) // the old worker needed 21 seconds
	require.Equal(t, map[string]int{"SENT": 21}, statusCounts(t, s))
}

type gatedProvider struct {
	*fakeProvider
	started, release chan struct{}
}

func (p gatedProvider) Send(ctx context.Context, team, domain, from string, to []string, raw []byte, id string) (string, error) {
	p.started <- struct{}{}
	<-p.release
	return p.fakeProvider.Send(ctx, team, domain, from, to, raw, id)
}

func TestShutdownFinishesInFlightAndRequeuesTheRest(t *testing.T) {
	s, team, d, p := setup(t)
	gate := gatedProvider{p, make(chan struct{}), make(chan struct{})}
	s.Provider = gate
	queueMessages(t, s, team, d.ID, 3, time.Now().UTC().Add(-time.Hour))
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); s.deliver(ctx, rate.NewLimiter(rate.Inf, 1), 1, 10) }()
	<-gate.started
	cancel()
	// The feeder returns unsent claims while the only worker is still mid-send.
	require.Eventually(t, func() bool { return statusCounts(t, s)["QUEUED"] == 2 }, 5*time.Second, 5*time.Millisecond)
	close(gate.release)
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("delivery did not stop")
	}
	require.Equal(t, 1, p.sends)
	require.Equal(t, map[string]int{"SENT": 1, "QUEUED": 2}, statusCounts(t, s))
}

func TestMigrateRaisesOnlyTheOldManagedHandoffRate(t *testing.T) {
	s, team, _, _ := setup(t)
	old := models.SMTPConfig{Base: models.Base{ID: uuid.NewString()}, Provider: "MANAGED", Host: "managed.internal", Port: 587, Username: "managed", FromEmail: "a@example.com", TeamID: team, MaxSendRate: 1}
	byo := models.SMTPConfig{Base: models.Base{ID: uuid.NewString()}, Provider: "SMTP", Host: "smtp.example.com", Port: 587, Username: "u", FromEmail: "b@example.com", TeamID: team, MaxSendRate: 1}
	require.NoError(t, s.DB.Session(&gorm.Session{SkipHooks: true}).Create(&old).Error)
	require.NoError(t, s.DB.Session(&gorm.Session{SkipHooks: true}).Create(&byo).Error)
	require.NoError(t, Migrate(s.DB))
	var rates []int
	require.NoError(t, s.DB.Model(&models.SMTPConfig{}).Where("id IN ?", []string{old.ID, byo.ID}).Order("provider").Pluck("max_send_rate", &rates).Error)
	require.Equal(t, []int{HandoffRate, 1}, rates) // MANAGED raised, SMTP untouched
}
