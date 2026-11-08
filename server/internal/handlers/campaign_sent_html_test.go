package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"kori/internal/models"
)

func TestCampaignSentHTMLPrefersWhatWasSent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Campaign{}, &models.Email{}, &models.EmailContent{}, &models.Template{}))
	team := uuid.NewString()
	quiet := db.Session(&gorm.Session{SkipHooks: true})
	tpl := models.Template{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "T", HTMLBody: "<p>edited later</p>"}
	require.NoError(t, quiet.Create(&tpl).Error)
	camp := models.Campaign{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "Launch", TemplateID: tpl.ID}
	require.NoError(t, quiet.Create(&camp).Error)
	h := NewMarketingHandler(db, "secret", "https://api.xem.email")
	get := func(teamID string) (int, map[string]string) {
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(httptest.NewRequest("GET", "/", nil), rec)
		c.SetParamNames("id")
		c.SetParamValues(camp.ID)
		c.Set("teamID", teamID)
		if err := h.CampaignSentHTML(c); err != nil {
			return err.(*echo.HTTPError).Code, nil
		}
		var out map[string]string
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
		return rec.Code, out
	}

	// Not sent yet: the template.
	_, out := get(team)
	require.Equal(t, "template", out["source"])

	content := models.EmailContent{ID: uuid.NewString(), TeamID: team, Kind: "campaign", HTML: `<a href="https://example.com/a">A</a>`}
	require.NoError(t, quiet.Create(&content).Error)
	require.NoError(t, quiet.Create(&models.Email{Base: models.Base{ID: uuid.NewString()}, TeamID: team, CampaignID: camp.ID, To: "a@example.com", ContentID: &content.ID}).Error)
	_, out = get(team)
	require.Equal(t, "sent", out["source"])
	require.Equal(t, content.HTML, out["html"])

	code, _ := get(uuid.NewString())
	require.Equal(t, 404, code, "another workspace can't read it")
}
