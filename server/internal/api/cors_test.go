package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/require"
)

func TestCORSOrigins(t *testing.T) {
	for _, tc := range []struct {
		name, configured, origin string
		allowed                  bool
	}{
		{"hosted app without environment", "", "https://app.xem.email", true},
		{"development", "", "http://localhost:3000", true},
		{"untrusted site", "", "https://attacker.example", false},
		{"lookalike site", "", "https://app.xem.email.attacker.example", false},
		{"insecure hosted origin", "", "http://app.xem.email", false},
		{"trim configured list", " https://workspace.example , , https://app.xem.email ", "https://app.xem.email", true},
		{"self-hosted override", "https://workspace.example", "https://workspace.example", true},
		{"override excludes hosted app", "https://workspace.example", "https://app.xem.email", false},
		{"explicit empty list denies all", ", ,", "https://app.xem.email", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("CORS_ALLOWED_ORIGINS", tc.configured)
			e := echo.New()
			e.Use(echomiddleware.CORSWithConfig(corsConfig()))
			e.POST("/api/v1/protected", func(c echo.Context) error { return echo.ErrUnauthorized })
			for _, method := range []string{http.MethodOptions, http.MethodPost} {
				req := httptest.NewRequest(method, "/api/v1/protected", nil)
				req.Header.Set(echo.HeaderOrigin, tc.origin)
				if method == http.MethodOptions {
					req.Header.Set(echo.HeaderAccessControlRequestMethod, http.MethodPost)
					req.Header.Set(echo.HeaderAccessControlRequestHeaders, "content-type,authorization")
				}
				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, req)
				expected := ""
				if tc.allowed {
					expected = tc.origin
				}
				require.Equal(t, expected, rec.Header().Get(echo.HeaderAccessControlAllowOrigin), method)
				if method == http.MethodOptions && tc.allowed {
					require.Equal(t, http.StatusNoContent, rec.Code)
					require.Contains(t, rec.Header().Get(echo.HeaderAccessControlAllowHeaders), "Authorization")
					require.Contains(t, rec.Header().Get(echo.HeaderAccessControlAllowMethods), "POST")
				}
				if method == http.MethodPost {
					require.Equal(t, http.StatusUnauthorized, rec.Code)
				}
			}
		})
	}
}
