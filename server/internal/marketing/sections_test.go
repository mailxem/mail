package marketing_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"kori/internal/config"
	"kori/internal/marketing"
	"kori/internal/models"
	"kori/internal/utils"
	cryptoutil "kori/internal/utils/crypto"
)

// sectionEdition sends a newsletter edition of body to a list with one
// reader tagged "VIP" and one in no tag, and returns the edition and each
// reader's email HTML by address.
func sectionEdition(t *testing.T, body string) (*gorm.DB, models.Campaign, map[string]string) {
	db := testDB(t)
	require.NoError(t, db.AutoMigrate(&models.Tag{}, &models.Contact{}, &models.Segment{}, &models.SegmentContact{}, &models.User{}, &models.Affiliate{}))
	oldRender := utils.RenderConfig
	t.Cleanup(func() { utils.RenderConfig = oldRender })
	utils.RenderConfig = &config.Config{JWT: config.JWTConfig{Secret: strings.Repeat("s", 32)}}
	team, tplID, listID, senderID := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	seed(t, db, &models.Team{Base: models.Base{ID: team}, Name: "Team", OwnerUserID: uuid.NewString()})
	seed(t, db, &models.MailingList{Base: models.Base{ID: listID}, TeamID: team, Name: "Audience"})
	seed(t, db, &models.Template{Base: models.Base{ID: tplID}, TeamID: team, CategoryID: uuid.NewString(), Name: "Issue", Subject: "Hi", HTMLBody: body})
	password, err := cryptoutil.Encrypt("fixture-password")
	require.NoError(t, err)
	seed(t, db, &models.SMTPConfig{Base: models.Base{ID: senderID}, TeamID: team, FromEmail: "hello@example.com", SupportsTLS: true, Password: password})
	vip := models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: team, ListID: listID, Email: "vip@example.com", Status: models.SubscriberStatusActive}
	seed(t, db, &vip)
	seed(t, db, &models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: team, ListID: listID, Email: "other@example.com", Status: models.SubscriberStatusActive})
	tag := models.Tag{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "VIP"}
	seed(t, db, &tag)
	require.NoError(t, db.Exec("INSERT INTO contact_tags (contact_id, tag_id) VALUES (?, ?)", vip.ID, tag.ID).Error)

	due := time.Now().UTC().Add(-time.Hour)
	n := models.Newsletter{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "Weekly", Subject: "An edition", TemplateID: tplID, ListID: listID, SMTPConfigID: senderID, Status: "SCHEDULED", Cadence: "WEEKLY", Timezone: "UTC", NextSendAt: &due, PostalAddress: "42 Market Street"}
	seed(t, db, &n)
	service := marketing.New(db, strings.Repeat("s", 32), "https://api.example.com")
	ctx := context.Background()
	require.NoError(t, service.MaterializeDue(ctx, time.Now().UTC()))
	var edition models.Campaign
	require.NoError(t, db.Where("newsletter_id = ?", n.ID).First(&edition).Error)
	require.NoError(t, service.PrepareEdition(ctx, edition.ID))
	require.NoError(t, db.First(&edition, "id = ?", edition.ID).Error)
	var emails []models.Email
	require.NoError(t, db.Where("campaign_id = ?", edition.ID).Find(&emails).Error)
	got := map[string]string{}
	for _, e := range emails {
		html, err := utils.EmailHTML(db, &e)
		require.NoError(t, err)
		got[e.To] = html
	}
	return db, edition, got
}

func section(name, rule string) string {
	return `<!--xem:section {"name":"` + name + `",` + rule + `}-->`
}

// Each reader's email holds only their sections, decided when it is rendered.
func TestEditionSectionsFollowEachReader(t *testing.T) {
	body := `<p>For everyone</p>` +
		section("VIP", `"audience":"only","tags":["vip"]`) + `<p>VIP offer</p><!--/xem:section-->` +
		section("Upgrade", `"audience":"except","tags":["vip"]`) + `<p>Become a VIP</p><!--/xem:section-->` +
		section("Online", `"show":"web"`) + `<p>Read online</p><!--/xem:section-->`
	db, edition, emails := sectionEdition(t, body)
	require.Len(t, emails, 2)
	require.Contains(t, emails["vip@example.com"], "VIP offer")
	require.NotContains(t, emails["vip@example.com"], "Become a VIP")
	require.Contains(t, emails["other@example.com"], "Become a VIP")
	require.NotContains(t, emails["other@example.com"], "VIP offer")
	for _, html := range emails {
		require.NotContains(t, html, "Read online")
		require.NotContains(t, html, "xem:section", "markers name tags; no reader may see them")
	}
	web := marketing.PublicEditionHTML(utils.WebSections(edition.HTMLBody))
	require.Contains(t, web, "Read online")
	require.Contains(t, web, "Become a VIP")
	require.NotContains(t, web, "VIP offer")

	reach, err := marketing.ListReach(db, edition.TeamID, edition.ListID, body)
	require.NoError(t, err)
	require.Equal(t, 2, reach.Readers)
	require.Equal(t, 0, reach.Empty)
	require.Equal(t, []int{1, 1, 0}, []int{reach.Sections[0].Readers, reach.Sections[1].Readers, reach.Sections[2].Readers})
}

// An edition nobody but VIPs has anything in is paused before anyone gets it.
func TestEditionWithAnEmptyEmailIsPaused(t *testing.T) {
	body := `<table><tr><td>` + section("VIP", `"audience":"only","tags":["vip"]`) + `<p>VIP offer</p><!--/xem:section--></td></tr></table>`
	db, edition, emails := sectionEdition(t, body)
	require.Empty(t, emails)
	require.Equal(t, models.CampaignStatusPaused, edition.Status)
	require.Equal(t, marketing.EmptyEmailReason(1), edition.PauseReason)
	reach, err := marketing.ListReach(db, edition.TeamID, edition.ListID, body)
	require.NoError(t, err)
	require.Equal(t, 1, reach.Empty)
}
