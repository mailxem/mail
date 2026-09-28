package handlers

import (
	"io"
	"kori/internal/db"
	"kori/internal/models"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"kori/internal/utils/logger"

	"github.com/labstack/echo/v4"
)

type UploadHandler struct {
	log *logger.Logger
	acl types.ObjectCannedACL
}

func NewUploadHandler(acl types.ObjectCannedACL) *UploadHandler {
	if acl == "" {
		acl = types.ObjectCannedACLPublicRead
	}
	return &UploadHandler{
		log: logger.New("upload_handler"),
		acl: acl,
	}
}

// UploadFile handles file uploads to S3
// @Summary Upload a file
// @Description Upload a file to the server
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Success 200 {object} map[string]interface{} "File uploaded successfully"
// @Failure 400 {object} map[string]string "Validation error or file not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/files/upload [post]
// @Tags Files
// @Security BearerAuth
func (h *UploadHandler) UploadFile(c echo.Context) error {

	contentType := c.Request().Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Content-Type must be multipart/form-data",
		})
	}

	storage := GetStorageHandler()
	if storage == nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Storage handler not configured",
		})
	}

	// Get file from request
	file, err := c.FormFile("file")
	if err != nil {
		log.Error("Failed to get file from request", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "No file provided",
		})
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to open file",
		})
	}

	content, err := io.ReadAll(src)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to read file",
		})
	}

	// Upload file to S3
	url, err := storage.UploadFile(c.Request().Context(), content, file.Filename, h.acl, file.Header.Get("Content-Type"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to upload file",
		})
	}

	h.log.Success("File uploaded successfully: %s", url)

	fileModel := &models.File{
		TeamID: c.Get("teamID").(string),
		UserID: c.Get("userID").(string),
		Path:   url[strings.LastIndex(url, "/")+1:],
		Name:   file.Filename,
		Size:   file.Size,
		Type:   file.Header.Get("Content-Type"),
	}

	getDb := db.GetDB()

	// Insert file into database
	err = getDb.Create(fileModel).Error

	if err != nil {
		err := h.log.Error("Failed to insert file into database", err)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to insert file into database",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "File uploaded successfully",
		"file":    fileModel.ID,
	})
}
