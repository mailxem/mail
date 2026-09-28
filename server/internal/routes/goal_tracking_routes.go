package routes

import (
	"kori/internal/api/middleware"
	"kori/internal/handlers"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func SetupGoalTrackingRoutes(e *echo.Echo, db *gorm.DB, jwtSecret string) {
	handler := handlers.NewGoalTrackingHandler(db)
	auth := middleware.NewAuthMiddleware(jwtSecret)

	// Goal endpoints
	goalGroup := e.Group("/api/v1/goals")
	goalGroup.Use(auth.Middleware())

	goalGroup.POST("", handler.CreateGoal)
	goalGroup.GET("", handler.ListGoals)
	goalGroup.GET("/summary", handler.GetAutomationGoalSummary)
	goalGroup.GET("/date-range", handler.GetGoalsByDateRange)
	goalGroup.GET("/top-contacts", handler.GetTopConvertingContacts)
	goalGroup.GET("/:id", handler.GetGoal)
	goalGroup.PUT("/:id", handler.UpdateGoal)
	goalGroup.DELETE("/:id", handler.DeleteGoal)
	goalGroup.GET("/:id/stats", handler.GetGoalStats)
	goalGroup.GET("/:id/conversions", handler.GetConversions)
	goalGroup.GET("/:id/progress", handler.GetGoalProgress)

	// Contact goal history endpoint
	contactGoalGroup := e.Group("/api/v1/contacts/:contactId/goals")
	contactGoalGroup.Use(auth.Middleware())
	contactGoalGroup.GET("", handler.GetContactGoalHistory)
}
