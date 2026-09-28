package routes

import (
	"kori/internal/api/middleware"
	"kori/internal/config"
	"kori/internal/handlers"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func SetupAuthRoutes(e *echo.Echo, db *gorm.DB, cfg *config.Config) {
	authHandler := handlers.NewAuthHandler(db)

	base := e.Group("/api/v1")

	// Public auth routes group
	auth := base.Group("/auth")
	users := base.Group("/users")

	// Public routes (no auth required)
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.GET("/google/callback", authHandler.GoogleAuthCallback)

	auth.GET("/invite/:code", authHandler.CheckInvite)
	auth.POST("/accept/:code", authHandler.AcceptInvite)
	auth.POST("/password-reset", authHandler.RequestPasswordReset)
	auth.POST("/password-reset/verify", authHandler.VerifyResetCode)
	auth.POST("/refresh", authHandler.RefreshToken)

	// Protected auth routes (require authentication)
	protectedAuth := users.Group("")
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWT.Secret)
	protectedAuth.Use(authMiddleware.Middleware())

	// Invite user route (require admin permissions)
	protectedAuth.POST("/invite", authHandler.InviteUser)
	protectedAuth.DELETE("/invite/:code", authHandler.DeleteInvite)
	protectedAuth.POST("/invite/resend/:code", authHandler.ResendInvite)

	// User management routes (require admin permissions)
	// userManagement := protectedAuth.Group("/users")
	// userManagement.Use(middleware.RequirePermissions(db, "manage:users"))
	// // Add user management routes here when implemented
	// userManagement.GET("", authHandler.ListUsers)         // List all users
	// userManagement.GET("/:id", authHandler.GetUser)       // Get user details
	// userManagement.PUT("/:id", authHandler.UpdateUser)    // Update user
	// userManagement.DELETE("/:id", authHandler.DeleteUser) // Delete user
	protectedAuth.GET("/me", authHandler.GetMe) // Get current user - accessible to any authenticated user
}
