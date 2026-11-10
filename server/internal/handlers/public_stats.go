package handlers

import (
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"kori/internal/models"
)

// PublicStats are platform-wide sending numbers for xem.email: no team,
// address or content, only totals. Test emails are left out.
type PublicStats struct {
	Delivered     int64     `json:"delivered"`     // all time
	Delivered30d  int64     `json:"delivered30d"`  // last 30 days
	DeliveryRate  float64   `json:"deliveryRate"`  // accepted / (accepted + bounced), last 30 days
	ComplaintRate float64   `json:"complaintRate"` // spam reports / accepted, last 30 days
	UpdatedAt     time.Time `json:"updatedAt"`
}

var accepted = []string{"SENT", "DELIVERED", "OPENED", "CLICKED"}

func publicStats(db *gorm.DB, now time.Time) (PublicStats, error) {
	s := PublicStats{UpdatedAt: now}
	since := now.Add(-30 * 24 * time.Hour)
	emails := func() *gorm.DB { return db.Model(&models.Email{}).Where("test = false") }
	var bounced, complaints int64
	for _, q := range []struct {
		dst *int64
		tx  *gorm.DB
	}{
		{&s.Delivered, emails().Where("status IN ?", accepted)},
		{&s.Delivered30d, emails().Where("status IN ? AND created_at >= ?", accepted, since)},
		{&bounced, emails().Where("status = ? AND created_at >= ?", "BOUNCED", since)},
		{&complaints, db.Model(&models.ComplaintReport{}).Where("created_at >= ?", since)},
	} {
		if err := q.tx.Count(q.dst).Error; err != nil {
			return s, err
		}
	}
	if total := s.Delivered30d + bounced; total > 0 {
		s.DeliveryRate = float64(s.Delivered30d) / float64(total)
	}
	if s.Delivered30d > 0 {
		s.ComplaintRate = float64(complaints) / float64(s.Delivered30d)
	}
	return s, nil
}

// PublicStatsHandler serves PublicStats to any site, recomputed at most hourly.
func PublicStatsHandler(db *gorm.DB) echo.HandlerFunc {
	var mu sync.Mutex
	var cached PublicStats
	return func(c echo.Context) error {
		mu.Lock()
		defer mu.Unlock()
		if time.Since(cached.UpdatedAt) > time.Hour {
			fresh, err := publicStats(db.WithContext(c.Request().Context()), time.Now().UTC())
			if err != nil {
				return echo.NewHTTPError(503, "Numbers are unavailable right now")
			}
			cached = fresh
		}
		c.Response().Header().Set("Access-Control-Allow-Origin", "*")
		c.Response().Header().Set("Cache-Control", "public, max-age=900")
		return c.JSON(200, cached)
	}
}
