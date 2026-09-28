package routes

import (
	"kori/internal/api/middleware"
	"kori/internal/handlers"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func SetupABTestingRoutes(e *echo.Echo, db *gorm.DB, jwtSecret string) {
	handler := handlers.NewABTestingHandler(db)
	auth := middleware.NewAuthMiddleware(jwtSecret)

	// A/B Testing endpoints
	testGroup := e.Group("/api/v1/ab-tests")
	testGroup.Use(auth.Middleware())

	testGroup.POST("", handler.CreateTest)
	testGroup.GET("", handler.ListTests)
	testGroup.GET("/:id", handler.GetTest)
	testGroup.PUT("/:id", handler.UpdateTest)
	testGroup.DELETE("/:id", handler.DeleteTest)
	testGroup.POST("/:id/start", handler.StartTest)
	testGroup.POST("/:id/stop", handler.StopTest)
	testGroup.GET("/:id/results", handler.GetResults)
	testGroup.GET("/:id/analyze", handler.AnalyzeTest)
	testGroup.POST("/:id/declare-winner", handler.DeclareWinner)
}
