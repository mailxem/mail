package sending

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func to(team string, d Domain, address string) Submission {
	return Submission{TeamID: team, DomainID: d.ID, Key: uuid.NewString(), From: "hello@example.com", Recipients: []string{address}, Raw: []byte("From: hello@example.com\r\nTo: " + address + "\r\nSubject: Hello\r\nContent-Type: text/plain\r\n\r\nHello!")}
}

func drain(t *testing.T, s *Service) {
	t.Helper()
	for {
		err := s.ProcessOne(context.Background())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return
		}
		require.NoError(t, err)
	}
}

func statuses(t *testing.T, s *Service, team string) map[string]int {
	t.Helper()
	var ms []Message
	require.NoError(t, s.DB.Where("team_id = ?", team).Find(&ms).Error)
	out := map[string]int{}
	for _, m := range ms {
		key := m.Status
		if gmailRecipients(m.Recipients) > 0 {
			key = "gmail " + key
		}
		out[key]++
	}
	return out
}

func TestGmailPaceHoldsGmailOverTheDailyAllowance(t *testing.T) {
	s, team, d, _ := setup(t)
	ctx := context.Background()
	defer func(n int64) { gmailFirstDay = n }(gmailFirstDay)
	gmailFirstDay = 2

	day1 := time.Date(2026, 11, 10, 23, 0, 0, 0, time.UTC)
	s.Now = func() time.Time { return day1 }
	d, err := s.RefreshDomain(ctx, team, d.ID) // keep the domain check fresh for the clock
	require.NoError(t, err)

	for _, a := range []string{"a@gmail.com", "b@googlemail.com", "c@gmail.com", "d@outlook.com"} {
		_, err := s.Submit(ctx, to(team, d, a))
		require.NoError(t, err)
	}
	drain(t, s)
	require.Equal(t, map[string]int{"gmail SENT": 2, "gmail QUEUED": 1, "SENT": 1}, statuses(t, s, team), "other providers are never held")

	var held Message
	require.NoError(t, s.DB.First(&held, "team_id = ? AND status = ?", team, "QUEUED").Error)
	require.NotNil(t, held.HoldUntil)
	require.Equal(t, time.Date(2026, 11, 11, 0, 0, 0, 0, time.UTC), held.HoldUntil.UTC())

	// The next UTC day the held message goes, and the allowance doubles.
	day2 := day1.Add(90 * time.Minute)
	s.Now = func() time.Time { return day2 }
	for _, a := range []string{"e@gmail.com", "f@gmail.com", "g@gmail.com", "h@gmail.com"} {
		_, err := s.Submit(ctx, to(team, d, a))
		require.NoError(t, err)
	}
	drain(t, s)
	require.Equal(t, map[string]int{"gmail SENT": 6, "gmail QUEUED": 1, "SENT": 1}, statuses(t, s, team), "day two: twice day one's 2")
}

func TestGmailPaceKeepsAnExistingDomainsVolume(t *testing.T) {
	s, team, d, _ := setup(t)
	defer func(n int64) { gmailFirstDay = n }(gmailFirstDay)
	gmailFirstDay = 2
	now := time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)
	s.Now = func() time.Time { return now }

	// Sent before pacing existed: 3 Gmail recipients on one day.
	for i := range 3 {
		m := Message{ID: uuid.NewString(), TeamID: team, DomainID: d.ID, IdempotencyKey: uuid.NewString(), Recipients: "r" + string(rune('a'+i)) + "@gmail.com", Status: "DELIVERED", CreatedAt: now.AddDate(0, 0, -3)}
		require.NoError(t, s.DB.Create(&m).Error)
	}
	for i := range 6 {
		ok, err := s.reserveGmail(context.Background(), d, 1)
		require.NoError(t, err)
		require.True(t, ok, "reservation %d of the 6 its best day allows", i+1)
	}
	ok, err := s.reserveGmail(context.Background(), d, 1)
	require.NoError(t, err)
	require.False(t, ok)
	require.EqualValues(t, 0, gmailRecipients("x@outlook.com, y@yahoo.com"))
}
