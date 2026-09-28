package routes

import (
	"kori/internal/api/middleware"
	"kori/internal/config"
	"kori/internal/handlers"
	"kori/internal/tasks"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func SetupAutomationRoutes(e *echo.Echo, db *gorm.DB, cfg *config.Config, taskClient *tasks.TaskClient) {
	automationGroup := e.Group("/api/v1/automations")
	auth := middleware.NewAuthMiddleware(cfg.JWT.Secret)
	automationGroup.Use(auth.Middleware())

	handler := handlers.NewAutomationHandler(db, taskClient)

	automationGroup.POST("/validate", handler.ValidateAutomation)

	// CRUD operations
	automationGroup.POST("", handler.CreateAutomation)
	automationGroup.GET("", handler.ListAutomations)
	automationGroup.GET("/:id", handler.GetAutomation)
	automationGroup.PUT("/:id", handler.UpdateAutomation)
	automationGroup.DELETE("/:id", handler.DeleteAutomation)

	// Activation/Deactivation
	automationGroup.POST("/:id/activate", handler.ActivateAutomation)
	automationGroup.POST("/:id/deactivate", handler.DeactivateAutomation)

	// Trigger
	automationGroup.POST("/:id/trigger", handler.TriggerAutomation)

	// Executions
	automationGroup.GET("/:id/executions", handler.GetExecutions)

	// Execution detail
	executionGroup := e.Group("/api/v1/executions")
	executionGroup.Use(auth.Middleware())
	executionGroup.GET("/:id", handler.GetExecutionDetail)
}
