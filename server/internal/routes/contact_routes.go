package routes

import (
	"kori/internal/api/middleware"
	"kori/internal/config"
	"kori/internal/models"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ContactImportRequest struct {
	FileID   string         `json:"fileId"`
	ListID   string         `json:"listId"`
	Mappings datatypes.JSON `json:"mappings" validate:"required,json"`
}

func SetupImportRoutes(e *echo.Echo, db *gorm.DB, cfg *config.Config) {
	importGroup := e.Group("/api/v1/imports")
	auth := middleware.NewAuthMiddleware(cfg.JWT.Secret)
	importGroup.Use(auth.Middleware())

	// Handle contact import
	// @Summary Import contacts from a file
	// @Description Import contacts from a file
	// @Accept json
	// @Produce json
	// @Param fileId path string true "File ID"
	// @Param listId path string true "List ID"
	// @Success 200 {object} map[string]string "Contact import queued successfully"
	// @Failure 400 {object} map[string]string "Validation error or file not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/imports/contact [post]
	importGroup.POST("/contact", func(c echo.Context) error {

		// content-type is application/json
		if c.Request().Header.Get("Content-Type") != "application/json" {
			return c.String(http.StatusBadRequest, "Content-Type must be application/json")
		}

		var req ContactImportRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		log.Info("🚀 Starting contact import")
		// we need get the fileId from the request
		fileId := req.FileID
		if fileId == "" {
			return c.String(http.StatusBadRequest, "fileId is required")
		}

		// listId is required
		listId := req.ListID
		if listId == "" {
			return c.String(http.StatusBadRequest, "listId is required")
		}

		// we need to get the file from the database
		var file models.File
		if err := db.Where("id = ?", fileId).First(&file).Error; err != nil {
			return c.String(http.StatusNotFound, "file not found")
		}

		// mappings is a json object
		mappings := req.Mappings

		if mappings == nil {
			return c.String(http.StatusBadRequest, "mappings is required")
		}

		contact_import := models.ContactImport{
			FileID:    fileId,
			TeamID:    c.Get("teamID").(string),
			ListID:    listId,
			Status:    models.ContactImportStatusPending,
			FieldsMap: mappings,
		}

		if err := db.Create(&contact_import).Error; err != nil {
			return c.String(http.StatusInternalServerError, "failed to create contact import")
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Contact import queued successfully"})
	})

}
