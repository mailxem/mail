package structs

import "github.com/mailxem/payments.go/internal/models"

// RecordUsageRequest represents the request to record usage
type RecordUsageRequest struct {
	Feature models.PlanFeature `json:"feature" validate:"required" example:"queries"`
	Usage   int                `json:"usage" validate:"omitempty,min=1" example:"10"`
}
