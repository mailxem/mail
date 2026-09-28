package routes

import (
	"kori/internal/api/middleware"
	"kori/internal/handlers"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func SetupSegmentRoutes(e *echo.Echo, db *gorm.DB, jwtSecret string) {
	segmentGroup := e.Group("/api/v1/segments")
	auth := middleware.NewAuthMiddleware(jwtSecret)
	segmentGroup.Use(auth.Middleware())

	handler := handlers.NewSegmentHandler(db)

	// Segment CRUD
	segmentGroup.POST("", handler.CreateSegment)
	segmentGroup.GET("", handler.ListSegments)
	segmentGroup.GET("/:id", handler.GetSegment)
	segmentGroup.PUT("/:id", handler.UpdateSegment)
	segmentGroup.DELETE("/:id", handler.DeleteSegment)

	// Segment operations
	segmentGroup.POST("/:id/refresh", handler.RefreshSegment)
	segmentGroup.POST("/preview", handler.PreviewSegment)

	// Segment contacts
	segmentGroup.GET("/:id/contacts", handler.GetSegmentContacts)
	segmentGroup.POST("/:id/contacts/:contactId", handler.AddContactToSegment)
	segmentGroup.DELETE("/:id/contacts/:contactId", handler.RemoveContactFromSegment)
}
