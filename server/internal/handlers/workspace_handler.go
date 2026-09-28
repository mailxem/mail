package handlers

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"kori/internal/models"
	"strconv"
	"strings"
	"time"
)

func (h *MarketingHandler) SaveProfile(c echo.Context) error {
	userID, _ := c.Get("userID").(string)
	if userID == "" {
		return echo.NewHTTPError(401, "Sign in to update your profile")
	}
	var input struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Bio       string `json:"bio"`
	}
	if c.Bind(&input) != nil || len(strings.TrimSpace(input.FirstName)) == 0 || len(input.FirstName) > 80 || len(input.LastName) > 80 || len(input.Bio) > 1000 {
		return bad("Enter a name of up to 80 characters and a bio of up to 1000 characters")
	}
	result := h.scoped(c).Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{"first_name": strings.TrimSpace(input.FirstName), "last_name": strings.TrimSpace(input.LastName), "bio": input.Bio})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return missing()
	}
	return c.NoContent(204)
}
func (h *MarketingHandler) WebhookStatus(c echo.Context) error {
	if uuid.Validate(c.Param("id")) != nil {
		return missing()
	}
	var input struct {
		Active bool `json:"isActive"`
	}
	if c.Bind(&input) != nil {
		return bad("Invalid status")
	}
	result := h.scoped(c).Model(&models.Webhook{}).Where("id = ?", c.Param("id")).Update("is_active", input.Active)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return missing()
	}
	return c.NoContent(204)
}
func (h *MarketingHandler) WebhookDeliveries(c echo.Context) error {
	id := c.Param("id")
	if uuid.Validate(id) != nil {
		return missing()
	}
	var webhook models.Webhook
	if h.scoped(c).First(&webhook, "id = ?", id).Error != nil {
		return missing()
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}
	query := h.service.DB.WithContext(c.Request().Context()).Model(&models.Delivery{}).Where("webhook_id = ? AND is_deleted = false", id)
	if status := c.QueryParam("status"); status != "" {
		code, err := strconv.Atoi(status)
		if err != nil || code < 100 || code > 599 {
			return bad("Invalid response status")
		}
		query = query.Where("response_code = ?", code)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return err
	}
	var rows []models.Delivery
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return err
	}
	output := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		output = append(output, map[string]interface{}{"id": row.ID, "webhookId": row.WebhookID, "eventType": row.Event, "status": row.ResponseCode, "payload": row.Payload, "response": row.ResponseBody, "error": row.Error, "createdAt": row.CreatedAt})
	}
	return c.JSON(200, map[string]interface{}{"deliveries": output, "pagination": map[string]interface{}{"total": total, "limit": limit, "offset": offset, "hasMore": int64(offset+len(rows)) < total}})
}

// APIKeyUsage scopes usage through its parent key; usage rows have no TeamID column.
func (h *MarketingHandler) APIKeyUsage(c echo.Context) error {
	query := h.service.DB.WithContext(c.Request().Context()).Model(&models.APIKeyUsage{}).
		Where("api_key_id IN (SELECT id FROM api_keys WHERE team_id = ? AND is_deleted = false) AND is_deleted = false", team(c))
	if id := c.Param("id"); id != "" {
		var row models.APIKeyUsage
		if uuid.Validate(id) != nil || query.First(&row, "id = ?", id).Error != nil {
			return missing()
		}
		return c.JSON(200, row)
	}
	if id := c.QueryParam("api_key_id"); id != "" {
		if uuid.Validate(id) != nil {
			return bad("Invalid API key")
		}
		var key models.APIKey
		if h.scoped(c).First(&key, "id = ?", id).Error != nil {
			return missing()
		}
		query = query.Where("api_key_id = ?", id)
	}
	if period := c.QueryParam("period"); period != "" {
		duration, ok := map[string]time.Duration{"1h": time.Hour, "24h": 24 * time.Hour, "7d": 7 * 24 * time.Hour, "30d": 30 * 24 * time.Hour}[period]
		if !ok {
			return bad("Invalid time period")
		}
		query = query.Where("timestamp >= ?", time.Now().UTC().Add(-duration))
	}
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 50
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return err
	}
	var rows []models.APIKeyUsage
	if err := query.Order("timestamp DESC").Limit(limit).Offset((page - 1) * limit).Find(&rows).Error; err != nil {
		return err
	}
	return c.JSON(200, map[string]interface{}{"data": rows, "total": total, "page": page, "limit": limit})
}
