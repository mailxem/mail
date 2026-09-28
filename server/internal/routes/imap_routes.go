package routes

import (
	"kori/internal/api/middleware"
	"kori/internal/config"
	"kori/internal/handlers"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"
)

func SetupIMAPRoutes(e *echo.Echo, config *config.Config, db *gorm.DB) {
	imap := e.Group("/api/v1/imap", echoMiddleware.BodyLimit("16K"))

	// Create IMAP handler
	imapHandler := handlers.NewIMAPHandler(db)

	// Add authentication middleware
	auth := middleware.NewAuthMiddleware(config.JWT.Secret)
	imap.Use(auth.Middleware())

	// Setup routes
	imap.GET("/folders", imapHandler.GetFolders, middleware.RequirePermissions(db, "imap_configs:read"))

	imap.GET("/emails", imapHandler.GetEmails, middleware.RequirePermissions(db, "imap_configs:read"))
	imap.PATCH("/flags", imapHandler.ChangeFlags, middleware.RequirePermissions(db, "imap_configs:write"))

	// test imap connection
	imap.POST("/test", imapHandler.TestConnection, middleware.RequirePermissions(db, "imap_configs:write"))
}
