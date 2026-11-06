package utils

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/time/rate"
	"kori/internal/models"
)

type countingLimiter struct{ tokens atomic.Int64 }

func (l *countingLimiter) Wait(context.Context) error { l.tokens.Add(1); return nil }

func batch(n int) []*models.Email {
	emails := make([]*models.Email, n)
	for i := range emails {
		emails[i] = &models.Email{To: fmt.Sprintf("%d@example.com", i)}
	}
	return emails
}

func TestSendPooledCapsConcurrencyAndKeepsOrder(t *testing.T) {
	emails := batch(40)
	limiter := &countingLimiter{}
	var inFlight, peak atomic.Int64
	results := sendPooled(emails, 4, limiter, func(e *models.Email) error {
		n := inFlight.Add(1)
		for p := peak.Load(); n > p && !peak.CompareAndSwap(p, n); p = peak.Load() {
		}
		time.Sleep(5 * time.Millisecond)
		inFlight.Add(-1)
		if e.To == "7@example.com" {
			return errors.New("rejected")
		}
		return nil
	})
	if peak.Load() > 4 {
		t.Fatalf("%d sends at once, want at most 4", peak.Load())
	}
	if limiter.tokens.Load() != 40 {
		t.Fatalf("took %d tokens for 40 sends", limiter.tokens.Load())
	}
	for i, r := range results {
		if r.Email != emails[i] || (r.Error != nil) != (i == 7) {
			t.Fatalf("result %d is for %s (error %v)", i, r.Email.To, r.Error)
		}
	}
}

func TestSendPooledStaysUnderRate(t *testing.T) {
	const perSecond = 40
	var mu sync.Mutex
	var at []time.Time
	start := time.Now()
	sendPooled(batch(20), perSecond, rate.NewLimiter(perSecond, 1), func(*models.Email) error {
		mu.Lock()
		at = append(at, time.Now())
		mu.Unlock()
		return nil
	})
	// With a burst of one, send n can't start before n/perSecond.
	for n, sent := range at {
		if earliest := time.Duration(n) * time.Second / perSecond; sent.Sub(start) < earliest-5*time.Millisecond {
			t.Fatalf("send %d at %v, earlier than %v", n, sent.Sub(start), earliest)
		}
	}
}
