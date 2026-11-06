package sending

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"kori/internal/config"
	"kori/internal/db"
	"kori/internal/hygiene"
	"kori/internal/models"
	"kori/internal/utils"
)

// batchScenario seeds suppressions and one earlier message, then returns
// submissions that hit every rule Submit applies, in order.
func batchScenario(t *testing.T, s *Service, team string, d Domain) []Submission {
	quiet := s.DB.Session(&gorm.Session{SkipHooks: true})
	require.NoError(t, s.DB.AutoMigrate(&hygiene.NeverEmailEntry{}))
	require.NoError(t, hygiene.Block(s.DB, "Never@example.net", time.Now()))
	require.NoError(t, quiet.Create(&Suppression{TeamID: team, Address: "managed@example.net", Reason: "bounce", CreatedAt: time.Now()}).Error)
	require.NoError(t, quiet.Create(&models.SuppressionList{Base: models.Base{ID: uuid.NewString()}, TeamID: team, EmailAddress: "Listed@Example.net", Reason: "manual", AddedAt: time.Now(), IsActive: true}).Error)
	require.NoError(t, quiet.Create(&models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: team, ListID: uuid.NewString(), Email: "Left@example.net", Status: models.SubscriberStatusUnsubscribed}).Error)
	sub := func(key string, to ...string) Submission {
		in := input(team, d)
		in.Key, in.Recipients = key, to
		return in
	}
	earlier := sub("earlier", "reader@example.net")
	_, err := s.Submit(context.Background(), earlier)
	require.NoError(t, err)

	changed := sub("first", "reader@example.net")
	changed.Raw = append(append([]byte{}, changed.Raw...), '!')
	badDomain := sub("bad-domain", "reader@example.net")
	badDomain.DomainID = uuid.NewString()
	otherFrom := sub("other-from", "reader@example.net")
	otherFrom.From, otherFrom.Raw = "hello@other.com", []byte("From: hello@other.com\r\nTo: reader@example.net\r\n\r\nhi")
	empty := sub("empty", "reader@example.net")
	empty.Raw = nil
	transient := sub("transient", "reader@example.net")
	transient.Transient = true
	test := sub("test", "reader@example.net")
	test.IsTest = true
	otherTeam := sub("other-team", "reader@example.net")
	otherTeam.TeamID = uuid.NewString()
	items := []Submission{
		sub("first", "reader@example.net"),
		sub("pair", "A@example.net", "a@example.net", "b@example.net"),
		sub("managed", "managed@example.net"),
		sub("listed", "listed@example.net"),
		sub("left", "reader2@example.net", "left@example.net"),
		sub("first", "reader@example.net"), // the same message again
		changed, earlier, badDomain, otherFrom, empty, transient, test, otherTeam,
		sub("never", "reader3@example.net", "never@example.net"),
	}
	for i := range 8 { // runs past the daily limit of 10
		items = append(items, sub(fmt.Sprintf("fill-%d", i), fmt.Sprintf("fill%d@example.net", i)))
	}
	return items
}

// SubmitBatch must accept and refuse exactly what the same submissions sent
// one at a time through Submit would.
func TestSubmitBatchMatchesSubmit(t *testing.T) {
	t.Setenv("SUPPRESSION_HASH_KEY", "test-key")
	for _, tc := range []struct {
		name  string
		limit func(*Service, string)
	}{
		{"daily limit", func(*Service, string) {}},
		{"queue limit", func(s *Service, _ string) { s.Config.QueueLimit = 4 }},
		{"budget", func(s *Service, team string) {
			require.NoError(t, s.DB.Model(&Account{}).Where("team_id = ?", team).Update("monthly_budget_micros", 5500).Error)
		}},
		{"paused", func(s *Service, team string) {
			require.NoError(t, s.DB.Model(&Account{}).Where("team_id = ?", team).Update("paused", true).Error)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run := func(batch bool) ([]BatchResult, Account, int64) {
				s, team, d, _ := setup(t)
				items := batchScenario(t, s, team, d)
				tc.limit(s, team)
				var results []BatchResult
				if batch {
					results = s.SubmitBatch(context.Background(), items)
				} else {
					for _, in := range items {
						m, err := s.Submit(context.Background(), in)
						results = append(results, BatchResult{m, err})
					}
				}
				a, err := s.Account(context.Background(), team)
				require.NoError(t, err)
				var n int64
				require.NoError(t, s.DB.Model(&Message{}).Where("team_id = ?", team).Count(&n).Error)
				return results, a, n
			}
			one, oneAccount, oneCount := run(false)
			batch, batchAccount, batchCount := run(true)
			for i := range one {
				a, b := one[i], batch[i]
				require.Equal(t, fmt.Sprint(a.Err), fmt.Sprint(b.Err), "submission %d", i)
				require.Equal(t, a.Message.ID == "", b.Message.ID == "", "submission %d", i)
				require.Equal(t, a.Message.Recipients, b.Message.Recipients, "submission %d", i)
				require.Equal(t, a.Message.Status, b.Message.Status, "submission %d", i)
				// The runs have their own team and domain IDs, so compare sizes.
				require.Equal(t, len(a.Message.Raw), len(b.Message.Raw), "submission %d", i)
			}
			require.Equal(t, oneCount, batchCount)
			require.Equal(t, [3]int64{oneAccount.DailyUsed, oneAccount.MonthlyUsed, oneAccount.BudgetUsedMicros}, [3]int64{batchAccount.DailyUsed, batchAccount.MonthlyUsed, batchAccount.BudgetUsedMicros})
			if tc.name == "daily limit" {
				// Sanity: the scenario does reach every rule.
				require.ErrorIs(t, batch[2].Err, ErrSuppressed)
				require.ErrorIs(t, batch[3].Err, ErrSuppressed)
				require.ErrorIs(t, batch[4].Err, ErrSuppressed)
				require.Equal(t, batch[0].Message.ID, batch[5].Message.ID)
				require.Error(t, batch[6].Err)
				require.ErrorIs(t, batch[8].Err, ErrDenied)
				require.Nil(t, batch[11].Message.Raw, "a transient message keeps no MIME")
				require.ErrorIs(t, batch[13].Err, ErrDenied)
				require.ErrorIs(t, batch[14].Err, ErrSuppressed, "the shared never-email list applies")
				require.ErrorIs(t, batch[len(batch)-1].Err, ErrLimit)
				require.EqualValues(t, 10, batchAccount.DailyUsed)
			}
			if tc.name == "paused" {
				require.ErrorIs(t, batch[0].Err, ErrDenied)
			}
		})
	}
}

// A campaign page goes to managed sending in one SubmitBatch, is queued
// without its MIME, and is dispatched with the MIME it was submitted with.
func TestManagedCampaignPageIsQueuedWithoutRawAndDispatchedUnchanged(t *testing.T) {
	s, team, d, p := setup(t)
	require.NoError(t, s.DB.AutoMigrate(&models.EmailContent{}))
	c, _ := apiContext(team, "POST", `{"from":"hello@example.com"}`)
	c.SetParamNames("id")
	c.SetParamValues(d.ID)
	require.NoError(t, s.activateSender(c))
	require.NoError(t, s.DB.First(&d, "id = ?", d.ID).Error)
	var sender models.SMTPConfig
	require.NoError(t, s.DB.First(&sender, "id = ?", d.SMTPConfigID).Error)

	cfg := &config.Config{JWT: config.JWTConfig{Secret: "campaign-secret"}, Server: config.ServerConfig{PublicURL: "https://api.example.com"}}
	oldDB, oldCfg, oldBatch, oldPolicy, oldRender := db.DB, utils.RenderConfig, utils.ManagedDeliveryBatch, utils.RecipientPolicy, RenderRaw
	t.Cleanup(func() {
		db.DB, utils.RenderConfig, utils.ManagedDeliveryBatch, utils.RecipientPolicy, RenderRaw = oldDB, oldCfg, oldBatch, oldPolicy, oldRender
	})
	db.DB, utils.RenderConfig, utils.RecipientPolicy, RenderRaw = s.DB, cfg, nil, utils.RenderRaw
	submitted := map[string][]byte{}
	calls := 0
	utils.ManagedDeliveryBatch = func(emails []*models.Email, raws [][]byte) []error {
		calls++
		for i, e := range emails {
			submitted[e.ID] = raws[i]
		}
		return s.SubmitEmails(context.Background(), emails, raws)
	}

	contentID, err := utils.StoreEmailContent(s.DB, models.EmailContent{TeamID: team, Kind: utils.ContentCampaign, HTML: `<p>Hi {{first_name}}, <a href="https://example.com/x">read</a></p>`, Subject: "Hi {{first_name}}"})
	require.NoError(t, err)
	campaign, tracked := uuid.NewString(), true
	var emails []*models.Email
	for i := range 5 {
		data, _ := json.Marshal(map[string]string{"first_name": fmt.Sprintf("Reader <%d>", i)})
		e := &models.Email{Base: models.Base{ID: uuid.NewString()}, ClickTrackingEnabled: &tracked, ContentID: &contentID, Data: data, UnsubscribeURL: "https://api.example.com/u/" + uuid.NewString(),
			From: "hello@example.com", To: fmt.Sprintf("reader%d@example.net", i), Subject: fmt.Sprintf("Hi Reader <%d>", i), Status: models.EmailStatusPending,
			TeamID: team, ContactID: uuid.NewString(), SMTPConfigID: sender.ID, CategoryID: uuid.NewString(), CampaignID: campaign}
		emails = append(emails, e)
	}
	// One email from before content was shared keeps its own body.
	old := *emails[0]
	old.ID, old.To, old.ContentID, old.Body = uuid.NewString(), "old@example.net", nil, "PHA+b2xkPC9wPg=="
	emails = append(emails, &old)
	require.NoError(t, s.DB.Session(&gorm.Session{SkipHooks: true}).Create(emails).Error)

	for _, r := range utils.NewEmailHandler(1).SendCampaignEmails(emails, 2, time.Hour, &sender) {
		require.NoError(t, r.Error)
	}
	require.Equal(t, 1, calls, "the page goes in one batch, without waiting between batches")
	var messages []Message
	require.NoError(t, s.DB.Find(&messages).Error)
	require.Len(t, messages, len(emails))
	for _, m := range messages {
		require.Empty(t, m.Raw, "managed_messages keeps no copy of an in-process email")
		require.NotEmpty(t, m.EmailID)
	}
	a, err := s.Account(context.Background(), team)
	require.NoError(t, err)
	require.EqualValues(t, len(emails), a.DailyUsed)

	for range messages {
		require.NoError(t, s.ProcessOne(context.Background()))
	}
	require.NoError(t, s.DB.Find(&messages).Error)
	feedback := "Feedback-ID: " + feedbackID(Submission{TeamID: team, FeedbackRef: campaign, FeedbackKind: "campaign"}) + "\r\n"
	for _, m := range messages {
		require.Equal(t, "SENT", m.Status)
		require.Empty(t, m.Raw)
		want, _, _, err := NormalizeRaw(submitted[m.EmailID], m.From)
		require.NoError(t, err)
		require.Equal(t, feedback+string(want), string(p.raws[m.ID]), "dispatched MIME differs from the submitted one")
	}
	var sent int64
	require.NoError(t, s.DB.Model(&models.Email{}).Where("status = ?", models.EmailStatusSent).Count(&sent).Error)
	require.EqualValues(t, len(emails), sent)
}

// Messages submitted with their MIME (API, SMTP relay) drop it once final.
func TestRawIsClearedAtTerminalStatus(t *testing.T) {
	s, team, d, p := setup(t)
	ctx := context.Background()
	sent, err := s.Submit(ctx, input(team, d))
	require.NoError(t, err)
	require.NotEmpty(t, sent.Raw)
	stuck, err := s.Submit(ctx, input(team, d))
	require.NoError(t, err)
	require.NoError(t, s.ProcessOne(ctx))
	require.Equal(t, 1, p.sends)
	require.NoError(t, s.DB.First(&sent, "id = ?", sent.ID).Error)
	require.Equal(t, "SENT", sent.Status)
	require.Empty(t, sent.Raw)

	// A worker that died mid-send leaves SENDING; maintenance closes it.
	require.NoError(t, s.DB.Model(&Message{}).Where("id = ?", stuck.ID).UpdateColumns(map[string]any{"status": "SENDING", "updated_at": s.Now().Add(-time.Hour)}).Error)
	require.NoError(t, s.MaintainMessages(ctx))
	require.NoError(t, s.DB.First(&stuck, "id = ?", stuck.ID).Error)
	require.Equal(t, "DELIVERY_UNKNOWN", stuck.Status)
	require.Empty(t, stuck.Raw)
}
