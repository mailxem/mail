package marketing_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"kori/internal/handlers"
	"kori/internal/models"
)

func TestCRMContactsPaginationAndIsolation(t *testing.T) {
	db := testDB(t)
	teamID, otherTeam, listID, foreignList := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	seed(t, db, &models.MailingList{Base: models.Base{ID: listID}, TeamID: teamID, Name: "Community"})
	seed(t, db, &models.MailingList{Base: models.Base{ID: foreignList}, TeamID: otherTeam, Name: "Private audience"})
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 1; i <= 215; i++ {
		contact := models.Contact{
			Base:   models.Base{ID: fmt.Sprintf("00000000-0000-0000-0000-%012d", i), CreatedAt: created},
			TeamID: teamID, ListID: listID, Email: fmt.Sprintf("person%d@example.com", i),
			FirstName: "Ada", LastName: "Lovelace", Company: "Acme", LifecycleStage: "LEAD", Status: models.SubscriberStatusActive,
		}
		if i <= 20 {
			contact.LifecycleStage = "QUALIFIED"
		}
		if i > 20 && i <= 30 {
			contact.LifecycleStage = "CUSTOMER"
			contact.Status = models.SubscriberStatusUnsubscribed
		}
		if i == 215 {
			contact.Company = "100%_Real!"
			contact.ListID = foreignList
		}
		require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&contact).Error)
	}
	seed(t, db, &models.Contact{Base: models.Base{ID: uuid.NewString(), IsDeleted: true}, TeamID: teamID, ListID: listID, Email: "deleted@example.com", LifecycleStage: "QUALIFIED"})
	seed(t, db, &models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: otherTeam, ListID: foreignList, Email: "foreign@example.com", LifecycleStage: "QUALIFIED"})
	// Legacy stages also belong in the Lead filter.
	require.NoError(t, db.Model(&models.Contact{}).Where("email = ?", "person213@example.com").UpdateColumn("lifecycle_stage", nil).Error)
	require.NoError(t, db.Model(&models.Contact{}).Where("email = ?", "person214@example.com").UpdateColumn("lifecycle_stage", "").Error)
	h := handlers.NewMarketingHandler(db, "", "")
	e := echo.New()
	e.GET("/contacts", h.CRMContacts, func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error { c.Set("teamID", teamID); return next(c) }
	})
	call := func(query string) handlers.CRMPage {
		t.Helper()
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/contacts"+query, nil))
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		var result handlers.CRMPage
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
		require.NotNil(t, result.Data)
		require.Equal(t, handlers.CRMCounts{Total: 215, Qualified: 20, Customers: 10, Subscribed: 205}, result.Summary)
		return result
	}
	first := call("?page=1&limit=100")
	require.EqualValues(t, 215, first.Total)
	require.Len(t, first.Data, 100)
	require.Equal(t, "person215@example.com", first.Data[0].Email)
	require.Empty(t, first.Data[0].ListName, "foreign list names must not leak")
	require.Equal(t, "Community", first.Data[1].ListName)
	second := call("?page=2&limit=100")
	last := call("?page=3&limit=100")
	require.Len(t, second.Data, 100)
	require.Len(t, last.Data, 15)
	seen := map[string]bool{}
	for _, page := range []handlers.CRMPage{first, second, last} {
		for _, contact := range page.Data {
			require.False(t, seen[contact.ID], "page overlap")
			seen[contact.ID] = true
		}
	}
	require.Len(t, seen, 215, "contacts past the former 200-row cap must be reachable")
	require.Equal(t, last, call("?page=999&limit=100"))
	qualified := call("?stage=QUALIFIED&search=" + url.QueryEscape("ada lovelace") + "&limit=10")
	require.EqualValues(t, 20, qualified.Total)
	require.Len(t, qualified.Data, 10)
	leads := call("?stage=lead")
	require.EqualValues(t, 185, leads.Total)
	require.Equal(t, "LEAD", leads.Data[1].LifecycleStage)
	require.Equal(t, "LEAD", leads.Data[2].LifecycleStage)
	for _, search := range []string{"100%_Real!", "%", "_", "!", "PERSON215@EXAMPLE.COM"} {
		result := call("?search=" + url.QueryEscape(search))
		require.EqualValues(t, 1, result.Total, search)
		require.Equal(t, "person215@example.com", result.Data[0].Email)
	}
	missing := call("?page=7&search=absent")
	require.Zero(t, missing.Total)
	require.Empty(t, missing.Data)
	require.Equal(t, 1, missing.Page)
}

func TestCRMContactsValidation(t *testing.T) {
	db := testDB(t)
	h := handlers.NewMarketingHandler(db, "", "")
	e := echo.New()
	for _, query := range []string{"page=0", "page=-1", "page=abc", "page=1000001", "limit=0", "limit=101", "limit=1.5", "stage=unknown", "search=" + strings.Repeat("a", 201)} {
		c := e.NewContext(httptest.NewRequest(http.MethodGet, "/?"+query, nil), httptest.NewRecorder())
		c.Set("teamID", uuid.NewString())
		err := h.CRMContacts(c)
		var httpErr *echo.HTTPError
		require.ErrorAs(t, err, &httpErr, query)
		require.Equal(t, http.StatusBadRequest, httpErr.Code, query)
	}
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())
	var httpErr *echo.HTTPError
	require.ErrorAs(t, h.CRMContacts(c), &httpErr)
	require.Equal(t, http.StatusUnauthorized, httpErr.Code)
}
