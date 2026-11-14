package marketing_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"kori/internal/marketing"
	"kori/internal/models"
	cryptoutil "kori/internal/utils/crypto"
)

// stagedEdition seeds an edition to fresh never-emailed contacts plus old
// readers the workspace emailed before, sent through provider.
func stagedEdition(t *testing.T, db *gorm.DB, provider string, fresh, old int) models.Campaign {
	t.Helper()
	require.NoError(t, db.AutoMigrate(&models.ServiceNotice{}, &models.ServiceNoticeEvent{}, &models.Alert{}))
	team, list, tpl, sender := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	seed(t, db, &models.MailingList{Base: models.Base{ID: list}, TeamID: team, Name: "Imported"})
	seed(t, db, &models.Template{Base: models.Base{ID: tpl}, TeamID: team, CategoryID: uuid.NewString(), Name: "Editorial", Subject: "Hello", HTMLBody: "<p>Hi</p>"})
	password, err := cryptoutil.Encrypt("fixture-password")
	require.NoError(t, err)
	seed(t, db, &models.SMTPConfig{Base: models.Base{ID: sender}, TeamID: team, FromEmail: "hello@example.com", SupportsTLS: true, Password: password, Provider: provider})
	contacts := make([]models.Contact, 0, fresh+old)
	for i := 0; i < fresh+old; i++ {
		contacts = append(contacts, models.Contact{Base: models.Base{ID: uuid.NewString(), CreatedAt: time.Now()}, TeamID: team, ListID: list, Email: fmt.Sprintf("Reader%d@example.com", i), Status: models.SubscriberStatusActive})
	}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).CreateInBatches(contacts, 200).Error)
	earlier := uuid.NewString()
	for _, c := range contacts[fresh:] {
		seed(t, db, &models.Email{Base: models.Base{ID: uuid.NewString()}, TeamID: team, CampaignID: earlier, ContactID: uuid.NewString(), TemplateID: tpl, SMTPConfigID: sender, CategoryID: uuid.NewString(), From: "hello@example.com", To: strings.ToLower(c.Email), Subject: "Before", Status: models.EmailStatusSent})
	}
	newsletter := uuid.NewString()
	edition := models.Campaign{Base: models.Base{ID: uuid.NewString()}, NewsletterID: &newsletter, TeamID: team, Name: "Edition", Subject: "Hello", HTMLBody: "<p>Hi</p>", TemplateID: tpl, ListID: list, SMTPConfigID: sender, Status: models.CampaignStatusSending, Schedule: models.CampaignScheduleOneTime}
	seed(t, db, &edition)
	return edition
}

type editionCounts struct{ Total, Fresh, Contacts int64 }

func counts(t *testing.T, db *gorm.DB, c models.Campaign) editionCounts {
	t.Helper()
	var n editionCounts
	require.NoError(t, db.Model(&models.Email{}).Where("campaign_id = ?", c.ID).Count(&n.Total).Error)
	require.NoError(t, db.Model(&models.Email{}).Where("campaign_id = ? AND LOWER(\"to\") NOT IN (SELECT LOWER(\"to\") FROM emails WHERE campaign_id <> ?)", c.ID, c.ID).Count(&n.Fresh).Error)
	require.NoError(t, db.Model(&models.Email{}).Where("campaign_id = ?", c.ID).Distinct("contact_id").Count(&n.Contacts).Error)
	return n
}

func reload(t *testing.T, db *gorm.DB, id string) models.Campaign {
	t.Helper()
	var c models.Campaign
	require.NoError(t, db.First(&c, "id = ?", id).Error)
	return c
}

// settle marks the edition's emails delivered and lets the stage wait pass.
func settle(t *testing.T, db *gorm.DB, c models.Campaign) {
	t.Helper()
	require.NoError(t, db.Model(&models.Email{}).Where("campaign_id = ? AND status = ?", c.ID, models.EmailStatusPending).Update("status", models.EmailStatusSent).Error)
	require.NoError(t, db.Model(&models.Campaign{}).Where("id = ? AND next_stage_at IS NOT NULL", c.ID).Update("next_stage_at", time.Now().Add(-time.Minute)).Error)
}

func TestStagedEditionSendsNewReadersInStages(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	s := marketing.New(db, strings.Repeat("s", 32), "https://api.example.com")
	c := stagedEdition(t, db, "MANAGED", 1000, 50)

	// Stage 1: every old reader, and 2% of the 1,000 new ones.
	require.NoError(t, s.PrepareEdition(ctx, c.ID))
	require.Equal(t, editionCounts{70, 20, 70}, counts(t, db, c))
	got := reload(t, db, c.ID)
	require.NotNil(t, got.SendInStages)
	require.True(t, *got.SendInStages)

	// The next tick only starts the wait.
	require.NoError(t, s.PrepareEdition(ctx, c.ID))
	require.Equal(t, int64(70), counts(t, db, c).Total)
	got = reload(t, db, c.ID)
	require.NotNil(t, got.NextStageAt)
	require.Equal(t, models.CampaignStatusSending, got.Status)

	// Stage 2 after the wait: 5% more.
	settle(t, db, c)
	require.NoError(t, s.PrepareEdition(ctx, c.ID))
	require.Equal(t, int64(70), counts(t, db, c).Fresh)
	require.NoError(t, s.PrepareEdition(ctx, c.ID))
	require.Equal(t, int64(70), counts(t, db, c).Fresh)

	// Then the rest, 100 per tick, until the edition completes.
	settle(t, db, c)
	for i := 0; i < 20 && reload(t, db, c.ID).Status == models.CampaignStatusSending; i++ {
		require.NoError(t, s.PrepareEdition(ctx, c.ID))
	}
	require.Equal(t, models.CampaignStatusCompleted, reload(t, db, c.ID).Status)
	require.Equal(t, editionCounts{1050, 1000, 1050}, counts(t, db, c))
}

func TestSmallOrSelfHostedEditionsAreNotStaged(t *testing.T) {
	for _, tc := range []struct {
		name     string
		provider string
		fresh    int
	}{{"under 500 new readers", "MANAGED", 499}, {"own SMTP sender", "CUSTOM", 1000}} {
		t.Run(tc.name, func(t *testing.T) {
			db := testDB(t)
			s := marketing.New(db, strings.Repeat("s", 32), "https://api.example.com")
			c := stagedEdition(t, db, tc.provider, tc.fresh, 10)
			for i := 0; i < 20 && reload(t, db, c.ID).Status == models.CampaignStatusSending; i++ {
				require.NoError(t, s.PrepareEdition(context.Background(), c.ID))
			}
			got := reload(t, db, c.ID)
			require.Equal(t, models.CampaignStatusCompleted, got.Status)
			require.False(t, *got.SendInStages)
			require.Nil(t, got.NextStageAt)
			require.Equal(t, int64(tc.fresh+10), counts(t, db, c).Contacts)
		})
	}
}

func TestBadFirstStagePausesAndResumeSendsTheRestOnce(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	s := marketing.New(db, strings.Repeat("s", 32), "https://api.example.com")
	c := stagedEdition(t, db, "MANAGED", 5000, 0)
	require.NoError(t, s.PrepareEdition(ctx, c.ID)) // the 2% stage: 100 readers
	require.NoError(t, s.PrepareEdition(ctx, c.ID)) // starts the wait
	require.NoError(t, s.PrepareEdition(ctx, c.ID)) // still waiting
	require.Equal(t, int64(100), counts(t, db, c).Fresh)

	// Three permanent bounces in the first 100: 3% is past 2.5%.
	var bounced []models.Email
	require.NoError(t, db.Where("campaign_id = ?", c.ID).Limit(3).Find(&bounced).Error)
	for _, e := range bounced {
		require.NoError(t, db.Table("emails").Where("id = ?", e.ID).Update("status", models.EmailStatusBounced).Error)
		require.NoError(t, db.Table("contacts").Where("id = ?", e.ContactID).Update("status", models.SubscriberStatusBounced).Error)
	}
	settle(t, db, c)
	require.NoError(t, s.PrepareEdition(ctx, c.ID))
	got := reload(t, db, c.ID)
	require.Equal(t, models.CampaignStatusPaused, got.Status)
	require.Equal(t, "Paused: 3.0% of the first 100 emails bounced. Check this list before sending to the rest.", got.PauseReason)
	var alerts int64
	require.NoError(t, db.Model(&models.ServiceNotice{}).Where("team_id = ? AND kind = ? AND status = ?", c.TeamID, "sending_paused", "QUEUED").Count(&alerts).Error)
	require.Equal(t, int64(1), alerts)
	require.NoError(t, s.PrepareEdition(ctx, c.ID))
	require.Equal(t, int64(100), counts(t, db, c).Total, "a paused edition sends nothing more")

	// Resuming (as the resume endpoint does) sends everyone else exactly once.
	require.NoError(t, db.Model(&models.Campaign{}).Where("id = ?", c.ID).Updates(map[string]any{"send_in_stages": false, "pause_reason": "", "status": models.CampaignStatusSending}).Error)
	for i := 0; i < 80 && reload(t, db, c.ID).Status == models.CampaignStatusSending; i++ {
		require.NoError(t, s.PrepareEdition(ctx, c.ID))
	}
	require.Equal(t, models.CampaignStatusCompleted, reload(t, db, c.ID).Status)
	n := counts(t, db, c)
	require.Equal(t, int64(5000), n.Total)
	require.Equal(t, n.Total, n.Contacts, "nobody got the edition twice")
	require.Equal(t, 5000, reload(t, db, c.ID).Processed)
}
