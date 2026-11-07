package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestSendEmailNeedsASubjectWithoutATemplate(t *testing.T) {
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"to":"reader@example.com","html":"<p>Hi</p>","test":true}`))
	req.Header.Set("Content-Type", "application/json")
	c := echo.New().NewContext(req, httptest.NewRecorder())
	c.Set("teamID", "team")
	err := SendEmail(c)
	var he *echo.HTTPError
	require.ErrorAs(t, err, &he)
	require.Equal(t, 400, he.Code)
	require.Equal(t, "A subject is required", he.Message)
}
