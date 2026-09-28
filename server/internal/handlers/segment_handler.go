package handlers

import (
	"encoding/json"
	"kori/internal/models"
	"kori/internal/services"
	"kori/internal/utils/logger"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

var segmentHandlerLog = logger.New("SEGMENT_HANDLER")

type SegmentHandler struct {
	db             *gorm.DB
	segmentService *services.SegmentService
}

func NewSegmentHandler(db *gorm.DB) *SegmentHandler {
	return &SegmentHandler{
		db:             db,
		segmentService: services.NewSegmentService(db),
	}
}

// CreateSegment handles POST /segments
func (h *SegmentHandler) CreateSegment(c echo.Context) error {
	teamID := c.Get("teamID").(string)

	var req struct {
		Name        string                 `json:"name" validate:"required,min=2,max=100"`
		Description string                 `json:"description"`
		Rules       models.SegmentRules    `json:"rules" validate:"required"`
		IsDynamic   bool                   `json:"isDynamic"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Marshal rules
	rulesJSON, err := json.Marshal(req.Rules)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid rules format"})
	}

	segment := &models.Segment{
		TeamID:      teamID,
		Name:        req.Name,
		Description: req.Description,
		Rules:       rulesJSON,
		MatchType:   req.Rules.MatchType,
		IsDynamic:   req.IsDynamic,
	}

	if err := h.segmentService.Create(c.Request().Context(), segment); err != nil {
		segmentHandlerLog.Error("Failed to create segment", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create segment"})
	}

	segmentHandlerLog.Success("Created segment %s for team %s", segment.ID, teamID)
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Segment created successfully",
		"data":    segment,
	})
}

// ListSegments handles GET /segments
func (h *SegmentHandler) ListSegments(c echo.Context) error {
	teamID := c.Get("teamID").(string)

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	filters := map[string]interface{}{
		"team_id": teamID,
	}

	// Filter by dynamic/static
	if isDynamic := c.QueryParam("isDynamic"); isDynamic != "" {
		filters["is_dynamic"] = isDynamic == "true"
	}

	segments, total, err := h.segmentService.List(
		c.Request().Context(),
		page,
		limit,
		filters,
		nil,
		[]string{"created_at"},
		"DESC",
	)

	if err != nil {
		segmentHandlerLog.Error("Failed to list segments", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to list segments"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": segments,
		"pagination": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetSegment handles GET /segments/:id
func (h *SegmentHandler) GetSegment(c echo.Context) error {
	teamID := c.Get("teamID").(string)
	segmentID := c.Param("id")

	segment, err := h.segmentService.Get(c.Request().Context(), segmentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Segment not found"})
		}
		segmentHandlerLog.Error("Failed to get segment", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get segment"})
	}

	// Verify ownership
	if segment.TeamID != teamID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": segment,
	})
}

// UpdateSegment handles PUT /segments/:id
func (h *SegmentHandler) UpdateSegment(c echo.Context) error {
	teamID := c.Get("teamID").(string)
	segmentID := c.Param("id")

	// Verify ownership
	segment, err := h.segmentService.Get(c.Request().Context(), segmentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Segment not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get segment"})
	}

	if segment.TeamID != teamID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
	}

	var req struct {
		Name        string              `json:"name"`
		Description string              `json:"description"`
		Rules       *models.SegmentRules `json:"rules"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Update fields
	if req.Name != "" {
		segment.Name = req.Name
	}
	if req.Description != "" {
		segment.Description = req.Description
	}
	if req.Rules != nil {
		rulesJSON, err := json.Marshal(req.Rules)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid rules format"})
		}
		segment.Rules = rulesJSON
		segment.MatchType = req.Rules.MatchType
	}

	if err := h.segmentService.Update(c.Request().Context(), segmentID, segment); err != nil {
		segmentHandlerLog.Error("Failed to update segment", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update segment"})
	}

	segmentHandlerLog.Success("Updated segment %s", segmentID)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Segment updated successfully",
		"data":    segment,
	})
}

// DeleteSegment handles DELETE /segments/:id
func (h *SegmentHandler) DeleteSegment(c echo.Context) error {
	teamID := c.Get("teamID").(string)
	segmentID := c.Param("id")

	// Verify ownership
	segment, err := h.segmentService.Get(c.Request().Context(), segmentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Segment not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get segment"})
	}

	if segment.TeamID != teamID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
	}

	if err := h.segmentService.Delete(c.Request().Context(), segmentID); err != nil {
		segmentHandlerLog.Error("Failed to delete segment", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete segment"})
	}

	segmentHandlerLog.Success("Deleted segment %s", segmentID)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Segment deleted successfully",
	})
}

// GetSegmentContacts handles GET /segments/:id/contacts
func (h *SegmentHandler) GetSegmentContacts(c echo.Context) error {
	teamID := c.Get("teamID").(string)
	segmentID := c.Param("id")

	// Verify ownership
	segment, err := h.segmentService.Get(c.Request().Context(), segmentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Segment not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get segment"})
	}

	if segment.TeamID != teamID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	contacts, total, err := h.segmentService.GetSegmentContacts(c.Request().Context(), segmentID, page, limit)
	if err != nil {
		segmentHandlerLog.Error("Failed to get segment contacts", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get contacts"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": contacts,
		"pagination": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// RefreshSegment handles POST /segments/:id/refresh
func (h *SegmentHandler) RefreshSegment(c echo.Context) error {
	teamID := c.Get("teamID").(string)
	segmentID := c.Param("id")

	// Verify ownership
	segment, err := h.segmentService.Get(c.Request().Context(), segmentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Segment not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get segment"})
	}

	if segment.TeamID != teamID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
	}

	if !segment.IsDynamic {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot refresh static segment"})
	}

	if err := h.segmentService.RefreshSegment(c.Request().Context(), segmentID); err != nil {
		segmentHandlerLog.Error("Failed to refresh segment", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to refresh segment"})
	}

	// Get updated segment
	segment, _ = h.segmentService.Get(c.Request().Context(), segmentID)

	segmentHandlerLog.Success("Refreshed segment %s: %d contacts", segmentID, segment.ContactCount)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":      "Segment refreshed successfully",
		"contactCount": segment.ContactCount,
	})
}

// PreviewSegment handles POST /segments/preview
func (h *SegmentHandler) PreviewSegment(c echo.Context) error {
	teamID := c.Get("teamID").(string)

	var req struct {
		Rules models.SegmentRules `json:"rules" validate:"required"`
		Limit int                 `json:"limit"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if req.Limit == 0 {
		req.Limit = 10
	}

	contacts, total, err := h.segmentService.PreviewSegment(c.Request().Context(), teamID, req.Rules, req.Limit)
	if err != nil {
		segmentHandlerLog.Error("Failed to preview segment", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to preview segment"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  contacts,
		"total": total,
		"limit": req.Limit,
	})
}

// AddContactToSegment handles POST /segments/:id/contacts/:contactId
func (h *SegmentHandler) AddContactToSegment(c echo.Context) error {
	teamID := c.Get("teamID").(string)
	segmentID := c.Param("id")
	contactID := c.Param("contactId")

	// Verify ownership
	segment, err := h.segmentService.Get(c.Request().Context(), segmentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Segment not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get segment"})
	}

	if segment.TeamID != teamID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
	}

	if err := h.segmentService.AddContactToSegment(c.Request().Context(), segmentID, contactID); err != nil {
		segmentHandlerLog.Error("Failed to add contact to segment", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	segmentHandlerLog.Success("Added contact %s to segment %s", contactID, segmentID)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Contact added to segment successfully",
	})
}

// RemoveContactFromSegment handles DELETE /segments/:id/contacts/:contactId
func (h *SegmentHandler) RemoveContactFromSegment(c echo.Context) error {
	teamID := c.Get("teamID").(string)
	segmentID := c.Param("id")
	contactID := c.Param("contactId")

	// Verify ownership
	segment, err := h.segmentService.Get(c.Request().Context(), segmentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Segment not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get segment"})
	}

	if segment.TeamID != teamID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
	}

	if err := h.segmentService.RemoveContactFromSegment(c.Request().Context(), segmentID, contactID); err != nil {
		segmentHandlerLog.Error("Failed to remove contact from segment", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	segmentHandlerLog.Success("Removed contact %s from segment %s", contactID, segmentID)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Contact removed from segment successfully",
	})
}
