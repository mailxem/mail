package handlers

import (
	"context"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"kori/internal/models"
	"net"
	"regexp"
	"strings"
	"time"
)

var domainName = regexp.MustCompile(`^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}$`)

func (h *MarketingHandler) AddDomain(c echo.Context) error {
	var input struct {
		Domain string `json:"domain"`
	}
	if c.Bind(&input) != nil {
		return bad("Invalid domain")
	}
	name := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(input.Domain), "."))
	if len(name) > 253 || !domainName.MatchString(name) {
		return bad("Enter a domain such as example.com")
	}
	var count int64
	if err := h.service.DB.Model(&models.Domain{}).Where("domain = ?", name).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return echo.NewHTTPError(409, "This domain has already been registered")
	}
	domain := models.Domain{Domain: name, TeamID: team(c), DNSRecord: "xem-verification=" + uuid.NewString()}
	if err := h.service.DB.WithContext(c.Request().Context()).Create(&domain).Error; err != nil {
		return echo.NewHTTPError(409, "Unable to register this domain")
	}
	return c.JSON(201, domain)
}
func (h *MarketingHandler) VerifyDomain(c echo.Context) error {
	var domain models.Domain
	if uuid.Validate(c.Param("id")) != nil || h.scoped(c).First(&domain, "id = ?", c.Param("id")).Error != nil {
		return missing()
	}
	if domain.DNSRecord == "" {
		domain.DNSRecord = "xem-verification=" + uuid.NewString()
		if err := h.scoped(c).Model(&models.Domain{}).Where("id = ?", domain.ID).Update("dns_record", domain.DNSRecord).Error; err != nil {
			return err
		}
		return echo.NewHTTPError(409, "A verification record was created. Refresh your domains and add the TXT record before trying again.")
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()
	records, err := net.DefaultResolver.LookupTXT(ctx, "_xem."+domain.Domain)
	if err != nil {
		return bad("Verification TXT record was not found. Check DNS and try again.")
	}
	for _, record := range records {
		if record == domain.DNSRecord {
			if err := h.scoped(c).Model(&models.Domain{}).Where("id = ?", domain.ID).Update("is_verified", true).Error; err != nil {
				return err
			}
			return c.NoContent(204)
		}
	}
	return bad("The TXT record does not match. Check DNS and try again.")
}
