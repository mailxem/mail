package handlers

import (
	"slices"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"kori/internal/models"
	"kori/internal/sending"
)

// alertKinds gives each alert its title and the fix it links to. {ref} is the
// campaign ID; {list} is the list an import was for.
var alertKinds = map[string]struct{ title, action, href string }{
	"sending_paused":         {"A campaign was paused to protect your reputation", "Resume campaign", "/campaigns/{ref}"},
	"sending_suspended":      {"Managed sending is suspended", "Review sending", "/settings/sending"},
	"domain_unready":         {"Your sending domain stopped verifying", "Check DNS", "/settings/sending"},
	"sending_bounced":        {"More emails than usual are bouncing", "View bounces", "/developer/logs/emails?folder=BOUNCED"},
	"sending_complaint":      {"Recipients are reporting your emails as spam", "View delivery", "/analytics/delivery"},
	"sending_failed":         {"Emails are failing to send", "View failed sends", "/developer/logs/emails?folder=FAILED"},
	"sending_unknown":        {"Some delivery outcomes are uncertain", "Check before resending", "/developer/logs/emails?folder=DELIVERY_UNKNOWN"},
	"sending_delayed":        {"Deliveries are being delayed", "View delivery", "/analytics/delivery"},
	"gmail_spam_rate":        {"Gmail readers are reporting your emails as spam", "Review sending", "/settings/sending"},
	"gmail_domain_spam_rate": {"Gmail spam rate is high on managed sending", "Review sending", "/settings/sending"},
	"automation_failed":      {"Automations are failing", "Review automations", "/automations"},
	"import_failed":          {"A contact import failed", "Retry import", "/audience/lists/{list}"},
	"import_finished":        {"Your contact import finished", "View list", "/audience/lists/{list}"},
}

type alertView struct {
	models.Alert
	Title    string `json:"title"`
	Action   string `json:"action"`
	Href     string `json:"href"`
	Resolved bool   `json:"resolved"`
}

// Alerts lists the workspace's recent alerts, newest first, with the unread
// count for the bell. Ongoing problems show as resolved once they clear.
func (h *MarketingHandler) Alerts(c echo.Context) error {
	db := h.service.DB.WithContext(c.Request().Context())
	var alerts []models.Alert
	if err := db.Where("team_id = ?", team(c)).Order("created_at DESC, id").Limit(30).Find(&alerts).Error; err != nil {
		return err
	}
	var unread int64
	if err := db.Model(&models.Alert{}).Where("team_id = ? AND read_at IS NULL", team(c)).Count(&unread).Error; err != nil {
		return err
	}
	refs := map[string][]string{}
	for _, a := range alerts {
		refs[a.Kind] = append(refs[a.Kind], a.ReferenceID)
	}
	// Whatever the alert is about, scoped to this workspace.
	var paused, unready []string
	if len(refs["sending_paused"]) > 0 {
		if err := db.Model(&models.Campaign{}).Where("team_id = ? AND id IN ? AND status = ?", team(c), refs["sending_paused"], models.CampaignStatusPaused).Pluck("id", &paused).Error; err != nil {
			return err
		}
	}
	if len(refs["domain_unready"]) > 0 {
		if err := db.Model(&sending.Domain{}).Where("team_id = ? AND id IN ? AND ready = ?", team(c), refs["domain_unready"], false).Pluck("id", &unready).Error; err != nil {
			return err
		}
	}
	suspended := false
	if len(refs["sending_suspended"]) > 0 {
		var s []bool
		if err := db.Model(&sending.Account{}).Where("team_id = ?", team(c)).Pluck("suspended", &s).Error; err != nil {
			return err
		}
		suspended = len(s) == 1 && s[0]
	}
	var imports []models.ContactImport
	if ids := append(refs["import_failed"], refs["import_finished"]...); len(ids) > 0 {
		if err := db.Select("id", "list_id", "created_at").Where("team_id = ? AND id IN ?", team(c), ids).Find(&imports).Error; err != nil {
			return err
		}
	}
	list := map[string]models.ContactImport{}
	for _, i := range imports {
		list[i.ID] = i
	}
	out := make([]alertView, 0, len(alerts))
	for _, a := range alerts {
		k := alertKinds[a.Kind]
		v := alertView{Alert: a, Title: k.title, Action: k.action, Href: strings.ReplaceAll(k.href, "{ref}", a.ReferenceID)}
		switch a.Kind {
		case "sending_paused":
			v.Resolved = !slices.Contains(paused, a.ReferenceID)
		case "domain_unready":
			v.Resolved = !slices.Contains(unready, a.ReferenceID)
		case "sending_suspended":
			v.Resolved = !suspended
		case "import_failed", "import_finished":
			i, ok := list[a.ReferenceID]
			v.Href = strings.ReplaceAll(v.Href, "{list}", i.ListID)
			if !ok {
				v.Action, v.Href = "", ""
			} else if a.Kind == "import_failed" {
				var later int64
				if err := db.Model(&models.ContactImport{}).Where("team_id = ? AND list_id = ? AND status = ? AND created_at > ?", team(c), i.ListID, models.ContactImportStatusCompleted, i.CreatedAt).Count(&later).Error; err != nil {
					return err
				}
				v.Resolved = later > 0
			}
		}
		out = append(out, v)
	}
	return c.JSON(200, map[string]any{"alerts": out, "unread": unread})
}

// ReadAlerts marks one alert read, or every unread alert when no id is given.
func (h *MarketingHandler) ReadAlerts(c echo.Context) error {
	var in struct {
		ID string `json:"id"`
	}
	if c.Bind(&in) != nil {
		return bad("Invalid alert")
	}
	q := h.service.DB.WithContext(c.Request().Context()).Model(&models.Alert{}).Where("team_id = ?", team(c))
	if in.ID != "" {
		var n int64
		if err := q.Session(&gorm.Session{}).Where("id = ?", in.ID).Count(&n).Error; err != nil {
			return err
		}
		if n == 0 {
			return missing()
		}
		q = q.Where("id = ?", in.ID)
	}
	if err := q.Where("read_at IS NULL").Update("read_at", time.Now()).Error; err != nil {
		return err
	}
	return c.NoContent(204)
}
