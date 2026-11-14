package handlers

import (
	"fmt"
	"kori/internal/models"
	"maps"
	"slices"
	"strings"

	"github.com/labstack/echo/v4"
)

// emptyFieldColumns maps each contact field a template can use to the text
// it is filled from (see HandleCampaignProcess); blank text is an empty field.
var emptyFieldColumns = map[string]string{
	"first_name": "first_name",
	"last_name":  "last_name",
	"name":       "COALESCE(first_name, '') || COALESCE(last_name, '')",
	"full_name":  "COALESCE(first_name, '') || COALESCE(last_name, '')",
	"company":    "company",
	"city":       "city",
	"state":      "state",
	"country":    "country",
	"zip":        "zip",
	"address":    "address",
	"phone":      "phone",
}

// ListEmptyFields counts the active readers of a list with no value for each
// contact field, so the review step can say who sees a fallback or a blank.
func (h *MarketingHandler) ListEmptyFields(c echo.Context) error {
	var list models.MailingList
	if err := h.scoped(c).Select("id").First(&list, "id = ?", c.Param("id")).Error; err != nil {
		return missing()
	}
	fields := slices.Sorted(maps.Keys(emptyFieldColumns))
	sums := []string{"COUNT(*)"}
	for _, field := range fields {
		sums = append(sums, fmt.Sprintf("COALESCE(SUM(CASE WHEN TRIM(COALESCE(%s, '')) = '' THEN 1 ELSE 0 END), 0)", emptyFieldColumns[field]))
	}
	counts := make([]int64, len(sums))
	targets := make([]any, len(counts))
	for i := range counts {
		targets[i] = &counts[i]
	}
	if err := h.scoped(c).Model(&models.Contact{}).Where("list_id = ? AND status = ?", list.ID, models.SubscriberStatusActive).
		Select(strings.Join(sums, ", ")).Row().Scan(targets...); err != nil {
		return err
	}
	empty := map[string]int64{}
	for i, field := range fields {
		empty[field] = counts[i+1]
	}
	return c.JSON(200, map[string]any{"total": counts[0], "empty": empty})
}
