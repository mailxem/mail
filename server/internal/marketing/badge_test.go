package marketing_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"html"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"kori/internal/billing"
	"kori/internal/config"
	"kori/internal/marketing"
	"kori/internal/models"
	"kori/internal/utils"
	cryptoutil "kori/internal/utils/crypto"
)

// A newsletter edition from a Free workspace ends with the "Sent with Xem"
// referral footer; one from a paying workspace doesn't.
func TestFreeWorkspaceEditionsCarryTheReferralBadge(t *testing.T) {
	db := testDB(t)
	require.NoError(t, db.AutoMigrate(&models.User{}, &models.Affiliate{}))
	free, paid := uuid.NewString(), uuid.NewString()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		plan := "xem-starter"
		if strings.HasSuffix(r.URL.Path, free) {
			plan = "xem-free"
		}
		fmt.Fprintf(w, `{"plan_id":%q,"status":"active"}`, plan)
	}))
	defer srv.Close()
	billing.Enabled = true
	billing.Use(&billing.Reporter{URL: srv.URL, HTTP: srv.Client(), Now: time.Now})
	t.Cleanup(func() { billing.Enabled = false; billing.Use(nil) })
	t.Setenv("PUBLIC_API_URL", "https://api.example.com")
	oldRender := utils.RenderConfig
	t.Cleanup(func() { utils.RenderConfig = oldRender })
	utils.RenderConfig = &config.Config{JWT: config.JWTConfig{Secret: strings.Repeat("s", 32)}}

	body := func(team string) string {
		owner := uuid.NewString()
		seed(t, db, &models.Team{Base: models.Base{ID: team}, Name: "Team", OwnerUserID: owner})
		seed(t, db, &models.User{Base: models.Base{ID: owner}, Email: owner + "@example.com", Password: "x", Role: models.UserRoleAdmin, TeamID: team})
		tplID, listID, senderID := uuid.NewString(), uuid.NewString(), uuid.NewString()
		seed(t, db, &models.MailingList{Base: models.Base{ID: listID}, TeamID: team, Name: "Audience"})
		seed(t, db, &models.Template{Base: models.Base{ID: tplID}, TeamID: team, CategoryID: uuid.NewString(), Name: "Issue", Subject: "Hi", HTMLBody: "<p>Hello</p>"})
		password, err := cryptoutil.Encrypt("fixture-password")
		require.NoError(t, err)
		seed(t, db, &models.SMTPConfig{Base: models.Base{ID: senderID}, TeamID: team, FromEmail: "hello@example.com", SupportsTLS: true, Password: password})
		seed(t, db, &models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: team, ListID: listID, Email: "reader@example.com", Status: models.SubscriberStatusActive})
		due := time.Now().UTC().Add(-time.Hour)
		n := models.Newsletter{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "Weekly", Subject: "An edition", TemplateID: tplID, ListID: listID, SMTPConfigID: senderID, Status: "SCHEDULED", Cadence: "WEEKLY", Timezone: "UTC", NextSendAt: &due, PostalAddress: "42 Market Street, Bengaluru"}
		seed(t, db, &n)
		service := marketing.New(db, strings.Repeat("s", 32), "https://api.example.com")
		ctx := context.Background()
		require.NoError(t, service.MaterializeDue(ctx, time.Now().UTC()))
		var edition models.Campaign
		require.NoError(t, db.Where("newsletter_id = ?", n.ID).First(&edition).Error)
		require.NoError(t, service.PrepareEdition(ctx, edition.ID))
		var email models.Email
		require.NoError(t, db.Where("team_id = ?", team).First(&email).Error)
		html, err := utils.EmailHTML(db, &email)
		require.NoError(t, err)
		return html
	}

	freeBody := body(free)
	require.Contains(t, freeBody, "Sent with")
	var a models.Affiliate
	require.NoError(t, db.First(&a).Error, "the Free workspace's owner became an affiliate")
	target := "https://xem.email/?utm_source=email-badge&via=" + a.Code
	// The link carries the owner's code whether or not click tracking rewrote it.
	tracked := strings.Contains(freeBody, base64.RawURLEncoding.EncodeToString([]byte(target))) || strings.Contains(freeBody, base64.StdEncoding.EncodeToString([]byte(target)))
	require.True(t, tracked || strings.Contains(freeBody, html.EscapeString(target)), "the badge links to %s", target)
	require.NotContains(t, body(paid), "Sent with")
}
