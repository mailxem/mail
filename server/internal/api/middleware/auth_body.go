package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

// injectRequestTeamID scopes JSON objects to the authenticated workspace. Actions
// such as checking domain records have no payload; leave those requests empty.
func injectRequestTeamID(request *http.Request, teamID string) error {
	if request.Body == nil {
		request.Body = http.NoBody
		request.ContentLength = 0
		return nil
	}
	defer request.Body.Close()
	raw, err := io.ReadAll(request.Body)
	if err != nil {
		if errors.Is(err, echo.ErrStatusRequestEntityTooLarge) {
			return err
		}
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON body")
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		request.Body = http.NoBody
		request.ContentLength = 0
		return nil
	}

	// RawMessage preserves numbers and nested values while replacing only teamId.
	var body map[string]json.RawMessage
	if err := json.Unmarshal(raw, &body); err != nil || body == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON body")
	}
	body["teamId"], _ = json.Marshal(teamID)
	rewritten, err := json.Marshal(body)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to encode body")
	}
	request.Body = io.NopCloser(bytes.NewReader(rewritten))
	request.ContentLength = int64(len(rewritten))
	return nil
}
