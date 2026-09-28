package routes

import (
	"kori/internal/api/middleware"
	"kori/internal/config"
	"kori/internal/handlers"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func SetupEMAILRoutes(e *echo.Echo, config *config.Config, db *gorm.DB) {
	// Redirect singular to plural for consistency
	e.POST("/api/v1/email", func(c echo.Context) error {
		// Forward the request body to the plural endpoint
		req := c.Request()
		req.URL.Path = "/api/v1/emails"
		e.ServeHTTP(c.Response(), req)
		return nil
	})

	// Create EMAIL routes group
	email := e.Group("/api/v1/emails")

	// Add authentication middleware
	auth := middleware.NewAuthMiddleware(config.JWT.Secret)
	email.Use(auth.Middleware())

	email.Use(middleware.RequirePermissions(db, "emails:create"))

	// @Summary Send an email
	// @Description Send an email to a list of contacts
	// @Accept json
	// @Produce json
	// @Param email body handlers.Email true "Email details"
	// @Success 200 {object} map[string]string "Email sent successfully"
	// @Failure 400 {object} map[string]string "Validation error or email not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/emails [post]
	email.POST("", handlers.SendEmail)

}
