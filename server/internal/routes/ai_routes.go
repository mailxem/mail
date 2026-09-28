package routes

import (
	"kori/internal/ai"
	"kori/internal/api/middleware"
	"kori/internal/config"
	"kori/internal/handlers"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func SetupAIRoutes(e *echo.Echo, db *gorm.DB, cfg *config.Config, aiClient ai.AIClient) {
	if !cfg.AI.Enabled {
		return // Don't register routes if AI is disabled
	}

	aiGroup := e.Group("/api/v1/ai")
	auth := middleware.NewAuthMiddleware(cfg.JWT.Secret)
	aiGroup.Use(auth.Middleware())

	handler := handlers.NewAIHandler(db, aiClient, cfg)

	// Query analytics with natural language
	aiGroup.POST("/query", handler.QueryAnalytics)

	// Get automation optimization suggestions
	aiGroup.POST("/optimize", handler.OptimizeAutomation)

	// Build automation from description
	aiGroup.POST("/build", handler.BuildAutomation)
}
