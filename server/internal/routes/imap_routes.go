package routes

import (
	"kori/internal/api/middleware"
	"kori/internal/config"
	"kori/internal/handlers"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func SetupIMAPRoutes(e *echo.Echo, config *config.Config, db *gorm.DB) {
	imap := e.Group("api/v1/imap")

	// Create IMAP handler
	imapHandler := handlers.NewIMAPHandler(db)

	// Add authentication middleware
	auth := middleware.NewAuthMiddleware(config.JWT.Secret)
	imap.Use(auth.Middleware())

	imap.Use(middleware.RequirePermissions(db, "imap:read"))

	// Setup routes
	imap.GET("/folders", imapHandler.GetFolders)

	imap.GET("/emails", imapHandler.GetEmails)

	// test imap connection
	imap.POST("/test", imapHandler.TestConnection)
}
