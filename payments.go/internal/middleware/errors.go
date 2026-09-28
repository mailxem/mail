package middleware

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ErrorResponse represents a sanitized error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

// SanitizeError returns a safe error message for clients
// Logs the detailed error server-side
func SanitizeError(c echo.Context, err error, userMessage string) error {
	// Log detailed error server-side
	log.Printf("ERROR: path=%s ip=%s error=%v",
		c.Request().URL.Path,
		c.RealIP(),
		err)

	// Return sanitized error to client
	return c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error:   "internal_server_error",
		Message: userMessage,
	})
}

// CustomErrorHandler handles errors in a secure way
func CustomErrorHandler(production bool) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		message := "An error occurred"

		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			// Only expose message for client errors (4xx)
			if code >= 400 && code < 500 {
				if msg, ok := he.Message.(string); ok {
					message = msg
				}
			}
		}

		// Log all errors server-side
		log.Printf("HTTP_ERROR: code=%d path=%s ip=%s error=%v",
			code,
			c.Request().URL.Path,
			c.RealIP(),
			err)

		// In production, sanitize all 5xx errors
		if production && code >= 500 {
			message = "An internal error occurred"
		}

		// Don't send error response if already sent
		if !c.Response().Committed {
			if c.Request().Method == http.MethodHead {
				err = c.NoContent(code)
			} else {
				err = c.JSON(code, ErrorResponse{
					Error:   http.StatusText(code),
					Message: message,
				})
			}
			if err != nil {
				log.Printf("ERROR: failed to send error response: %v", err)
			}
		}
	}
}
