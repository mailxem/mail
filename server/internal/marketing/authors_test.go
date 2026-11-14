package marketing_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"kori/internal/marketing"
	"kori/internal/models"
	cryptoutil "kori/internal/utils/crypto"
)

func TestRepliesGoToTheAuthorUnlessTheCampaignSetsReplyTo(t *testing.T) {
	db := testDB(t)
	team := uuid.NewString()
	editor := models.Author{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "Asha Rao", Email: "asha@example.com", IsDefault: true, BylineEmail: true}
	guest := models.Author{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "Guest <Writer>", Email: "guest@example.com", PhotoURL: "https://cdn.example.com/g.png", BylineEmail: true, BylineWeb: true}
	gone := models.Author{Base: models.Base{ID: uuid.NewString(), IsDeleted: true}, TeamID: team, Name: "Left", Email: "left@example.com"}
	other := models.Author{Base: models.Base{ID: uuid.NewString()}, TeamID: uuid.NewString(), Name: "Elsewhere", Email: "else@example.com"}
	for _, a := range []*models.Author{&editor, &guest, &gone, &other} {
		seed(t, db, a)
	}

	author := func(id *string) *models.Author {
		a, err := marketing.CampaignAuthor(db, team, id)
		require.NoError(t, err)
		return a
	}
	require.Equal(t, guest.ID, author(&guest.ID).ID, "the picked author signs")
	require.Equal(t, editor.ID, author(nil).ID, "no pick uses the workspace default")
	require.Equal(t, editor.ID, author(&gone.ID).ID, "a removed author falls back to the default")
	require.Equal(t, editor.ID, author(&other.ID).ID, "another workspace's author is never used")
	none, err := marketing.CampaignAuthor(db, uuid.NewString(), nil)
	require.NoError(t, err)
	require.Nil(t, none, "a workspace without a default has no byline")

	require.Equal(t, "guest@example.com", marketing.ReplyTo("", &guest))
	require.Equal(t, "desk@example.com", marketing.ReplyTo(" desk@example.com ", &guest), "the campaign's own Reply-To wins")
	require.Equal(t, "desk@example.com", marketing.ReplyTo("desk@example.com", nil))
	require.Equal(t, "", marketing.ReplyTo("", nil), "empty keeps the sender's own")

	body := marketing.WithByline(`<html><body style="margin:0"><p>Hi</p></body></html>`, &guest)
	require.True(t, strings.HasPrefix(body, `<html><body style="margin:0"><table`), "the byline goes at the top of the body")
	require.Contains(t, body, "By <strong style=\"color:#222222\">Guest &lt;Writer&gt;</strong>")
	require.Contains(t, body, `src="https://cdn.example.com/g.png"`)
	guest.BylineEmail = false
	require.Equal(t, "<p>Hi</p>", marketing.WithByline("<p>Hi</p>", &guest), "the author can keep it out of emails")
	require.Equal(t, map[string]string{"name": "Guest <Writer>", "photo": "https://cdn.example.com/g.png"}, marketing.Byline(&guest))
	require.Nil(t, marketing.Byline(&editor), "the editor keeps their byline off the web")
}

func TestNewsletterEditionsAreSignedAndRepliedToByTheirAuthor(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	team, tplID, listID, senderID := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	seed(t, db, &models.MailingList{Base: models.Base{ID: listID}, TeamID: team, Name: "Audience"})
	seed(t, db, &models.Template{Base: models.Base{ID: tplID}, TeamID: team, CategoryID: uuid.NewString(), Name: "Editorial", Subject: "Hello", HTMLBody: "<body><p>Hello</p></body>"})
	password, err := cryptoutil.Encrypt("fixture-password")
	require.NoError(t, err)
	seed(t, db, &models.SMTPConfig{Base: models.Base{ID: senderID}, TeamID: team, FromEmail: "hello@example.com", SupportsTLS: true, Password: password})
	seed(t, db, &models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: team, ListID: listID, Email: "reader@example.com", Status: models.SubscriberStatusActive})
	author := models.Author{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "Asha Rao", Email: "asha@example.com", BylineEmail: true}
	seed(t, db, &author)
	due := time.Now().UTC().Add(-time.Hour)
	n := models.Newsletter{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "Weekly", Subject: "An edition", TemplateID: tplID, ListID: listID, SMTPConfigID: senderID, Status: "SCHEDULED", Cadence: "WEEKLY", Timezone: "UTC", NextSendAt: &due, PostalAddress: "42 Market Street, Bengaluru", AuthorID: &author.ID}
	seed(t, db, &n)

	service := marketing.New(db, strings.Repeat("s", 32), "https://api.example.com")
	require.NoError(t, service.MaterializeDue(ctx, time.Now().UTC()))
	var edition models.Campaign
	require.NoError(t, db.First(&edition, "newsletter_id = ?", n.ID).Error)
	require.Equal(t, author.ID, *edition.AuthorID, "editions keep the newsletter's author")
	require.NoError(t, service.PrepareEdition(ctx, edition.ID))

	var message models.Email
	require.NoError(t, db.First(&message, "campaign_id = ?", edition.ID).Error)
	require.Equal(t, "asha@example.com", message.ReplyTo)
	var content models.EmailContent
	require.NoError(t, db.First(&content, "id = ?", *message.ContentID).Error)
	require.Contains(t, content.HTML, "By <strong")
	require.NotContains(t, edition.HTMLBody, "By <strong", "the edition keeps its body without the byline")
}
