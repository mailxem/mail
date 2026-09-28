package api

import (
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func corsConfig() echomiddleware.CORSConfig {
	// A missing deployment variable must not lock out the hosted Xem app.
	// Explicit configuration replaces the defaults for self-hosted installations.
	configured := os.Getenv("CORS_ALLOWED_ORIGINS")
	if strings.TrimSpace(configured) == "" {
		configured = "https://app.xem.email,http://localhost:3000"
	}
	origins := []string{}
	for _, origin := range strings.Split(configured, ",") {
		if origin = strings.TrimSpace(origin); origin != "" {
			origins = append(origins, origin)
		}
	}
	return echomiddleware.CORSConfig{
		// Exact matching also prevents Echo from turning an empty list into "*".
		AllowOriginFunc: func(origin string) (bool, error) {
			for _, allowed := range origins {
				if origin == allowed {
					return true, nil
				}
			}
			return false, nil
		},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderContentLength},
	}
}
