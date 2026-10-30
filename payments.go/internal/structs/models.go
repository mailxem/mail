package structs

import "github.com/mailxem/payments.go/internal/models"

// RecordUsageRequest represents the request to record usage
type RecordUsageRequest struct {
	Feature models.PlanFeature `json:"feature" validate:"required" example:"queries"`
	Usage   int                `json:"usage" validate:"omitempty,min=1" example:"10"`
	// EventID makes a report idempotent: a repeated ID is not counted again.
	EventID string `json:"event_id" validate:"omitempty,max=128" example:"usage-2026-10-30T10:00-team"`
}
