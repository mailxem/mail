package marketing_test

import (
	"context"
	"errors"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"kori/internal/marketing"
	"kori/internal/models"
)

func TestCopyCampaignKeepsContentNotState(t *testing.T) {
	on, newsletter, archive, other := true, uuid.NewString(), uuid.NewString(), uuid.NewString()
	next := time.Now()
	src := models.Campaign{
		Base: models.Base{ID: uuid.NewString(), CreatedAt: next}, TeamID: "team", Name: "Spring sale", Description: "April",
		TemplateID: "tpl", ListID: "list", SMTPConfigID: "sender", Subject: "Hello", HTMLBody: "<p>Hi</p>", ReplyTo: "replies@example.com",
		PostalAddress: "1 Main St", TrackEngagement: &on, ArchiveNewsletterID: &archive, Data: datatypes.JSON(`{"code":"SPRING"}`),
		BatchSize: 50, BatchDelay: models.Duration(time.Minute), Timezone: "Europe/Berlin",
		// None of these carry over.
		NewsletterID: &newsletter, Status: models.CampaignStatusCompleted, ScheduledFor: next, Schedule: models.CampaignScheduleRecurring,
		RecurringSchedule: models.CampaignRecurringScheduleWeekly, CronExpression: "0 0 * * 1", Processed: 900, SendInStages: &on,
		Stage: 2, NextStageAt: &next, PauseReason: "Paused", AudienceCampaignID: &other, AudienceFilter: models.AudienceNotOpened,
	}
	got := marketing.CopyCampaign(src)
	require.Equal(t, models.Campaign{
		TeamID: "team", Name: "Spring sale (copy)", Description: "April", TemplateID: "tpl", ListID: "list", SMTPConfigID: "sender",
		Subject: "Hello", HTMLBody: "<p>Hi</p>", ReplyTo: "replies@example.com", PostalAddress: "1 Main St", TrackEngagement: &on,
		ArchiveNewsletterID: &archive, Data: datatypes.JSON(`{"code":"SPRING"}`), BatchSize: 50, BatchDelay: models.Duration(time.Minute),
		Timezone: "Europe/Berlin", Status: models.CampaignStatusDraft, Schedule: models.CampaignScheduleOneTime,
	}, got)
}

func TestFollowUpAudienceSkipsEngagedUnsubscribedAndSuppressed(t *testing.T) {
	db := testDB(t)
	require.NoError(t, db.AutoMigrate(&models.EmailTracking{}))
	ctx := context.Background()
	s := marketing.New(db, strings.Repeat("s", 32), "https://api.example.com")
	team, list, tpl, sender := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	seed(t, db, &models.MailingList{Base: models.Base{ID: list}, TeamID: team, Name: "Customers"})
	src := models.Campaign{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "Launch", Subject: "We launched", TemplateID: tpl, ListID: list, SMTPConfigID: sender, ReplyTo: "team@example.com", Status: models.CampaignStatusCompleted, Schedule: models.CampaignScheduleOneTime}
	seed(t, db, &src)

	reader := func(name string, status models.SubscriberStatus, email models.EmailStatus, events ...models.EmailTracking) {
		contact := models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: team, ListID: list, Email: name + "@example.com", Status: status}
		seed(t, db, &contact)
		if email == "" {
			return
		}
		e := models.Email{Base: models.Base{ID: uuid.NewString()}, TeamID: team, CampaignID: src.ID, ContactID: contact.ID, TemplateID: tpl, SMTPConfigID: sender, CategoryID: uuid.NewString(), From: "hello@example.com", To: contact.Email, Subject: "We launched", Status: email}
		seed(t, db, &e)
		for _, ev := range events {
			ev.ID, ev.EmailID, ev.Timestamp = uuid.NewString(), e.ID, time.Now()
			seed(t, db, &ev)
		}
	}
	open := models.EmailTracking{Event: models.EmailTrackingEventOpen}
	click := models.EmailTracking{Event: models.EmailTrackingEventClick}
	botOpen := models.EmailTracking{Event: models.EmailTrackingEventOpen, Machine: models.TrackingMachineBot}
	active := models.SubscriberStatusActive
	reader("quiet", active, models.EmailStatus("DELIVERED"))
	reader("scanned", active, models.EmailStatusSent, botOpen) // only a link scanner opened it
	reader("opener", active, models.EmailStatusOpened, open)
	reader("clicker", active, models.EmailStatusClicked, click)
	reader("left", models.SubscriberStatusUnsubscribed, models.EmailStatus("DELIVERED"))
	reader("blocked", active, models.EmailStatus("DELIVERED"))
	reader("bounced", active, models.EmailStatusBounced)
	reader("joined-later", active, "")
	seed(t, db, &models.SuppressionList{Base: models.Base{ID: uuid.NewString()}, TeamID: team, EmailAddress: "BLOCKED@example.com", Reason: models.SuppressionReasonManual, AddedAt: time.Now(), IsActive: true})

	audience := func(filter models.AudienceFilter) []string {
		t.Helper()
		draft, err := s.DuplicateCampaign(ctx, team, src.ID, filter)
		require.NoError(t, err)
		require.Equal(t, models.CampaignStatusDraft, reload(t, db, draft.ID).Status)
		require.Equal(t, "team@example.com", draft.ReplyTo)
		old, fresh := marketing.Recipients(db, &draft)
		var a, b []string
		require.NoError(t, old.Pluck("email", &a).Error)
		require.NoError(t, fresh.Pluck("email", &b).Error)
		got := append(a, b...)
		sort.Strings(got)
		return got
	}
	require.Equal(t, []string{"quiet@example.com", "scanned@example.com"}, audience(models.AudienceNotOpened))
	require.Equal(t, []string{"opener@example.com", "quiet@example.com", "scanned@example.com"}, audience(models.AudienceNotClicked))
	require.Len(t, audience(""), 7, "a plain copy goes to every active reader on the list")

	unsent := models.Campaign{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "Draft", TemplateID: tpl, ListID: list, SMTPConfigID: sender, Status: models.CampaignStatusDraft}
	seed(t, db, &unsent)
	_, err := s.DuplicateCampaign(ctx, team, unsent.ID, models.AudienceNotOpened)
	require.ErrorIs(t, err, marketing.ErrNotSent)
	_, err = s.DuplicateCampaign(ctx, uuid.NewString(), src.ID, "")
	require.True(t, errors.Is(err, gorm.ErrRecordNotFound), "another workspace can't copy it")
}
