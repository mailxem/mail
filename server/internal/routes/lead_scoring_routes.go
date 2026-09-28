package routes

import (
	"kori/internal/api/middleware"
	"kori/internal/handlers"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func SetupLeadScoringRoutes(e *echo.Echo, db *gorm.DB, jwtSecret string) {
	handler := handlers.NewLeadScoringHandler(db)
	auth := middleware.NewAuthMiddleware(jwtSecret)

	// Contact score endpoints
	contactScoreGroup := e.Group("/api/v1/contacts/:id/score")
	contactScoreGroup.Use(auth.Middleware())
	contactScoreGroup.GET("", handler.GetContactScore)
	contactScoreGroup.POST("", handler.UpdateContactScore)
	contactScoreGroup.GET("/history", handler.GetScoreHistory)

	// Analytics endpoints
	analyticsGroup := e.Group("/api/v1/analytics")
	analyticsGroup.Use(auth.Middleware())
	analyticsGroup.GET("/top-leads", handler.GetTopScoredContacts)
	analyticsGroup.GET("/contacts-by-grade", handler.GetContactsByGrade)
	analyticsGroup.GET("/score-distribution", handler.GetScoreDistribution)

	// Score rules endpoints
	rulesGroup := e.Group("/api/v1/score-rules")
	rulesGroup.Use(auth.Middleware())
	rulesGroup.GET("", handler.ListScoreRules)
	rulesGroup.POST("", handler.CreateScoreRule)
	rulesGroup.POST("/initialize", handler.InitializeScoreRules)
	rulesGroup.PUT("/:id", handler.UpdateScoreRule)
	rulesGroup.DELETE("/:id", handler.DeleteScoreRule)
}
