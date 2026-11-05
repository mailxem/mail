package affiliates

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	em "github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"kori/internal/api/middleware"
	"kori/internal/billing"
	"kori/internal/models"
)

type handler struct {
	db  *gorm.DB
	now func() time.Time
}

// Register adds the affiliate routes. They answer 404 unless Xem Cloud billing
// is configured, since self-hosted installs have nothing to pay commissions on.
func Register(e *echo.Echo, db *gorm.DB, jwtSecret string) {
	h := &handler{db: db, now: time.Now}
	small := em.BodyLimit("4K")

	// The website reports link visits; nothing about the visitor is stored.
	e.POST("/api/v1/affiliates/clicks", h.click, cloudOnly, small,
		em.RateLimiter(em.NewRateLimiterMemoryStore(rate.Limit(5))))

	g := e.Group("/api/v1/affiliates", cloudOnly, small, middleware.NewAuthMiddleware(jwtSecret).Middleware(), people)
	g.GET("/me", h.me)
	g.POST("", h.join)
	g.POST("/attribute", h.attribute)

	// Called by the payments service, and by an operator recording payouts.
	in := e.Group("/api/v1/internal/affiliates", cloudOnly, em.BodyLimit("16K"), internalKey)
	in.POST("/payments", h.payment)
	in.POST("/reversals", h.reversal)
	in.GET("/payouts/due", h.due)
	in.POST("/:id/payouts", h.pay)
}

func cloudOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !billing.Enabled {
			return echo.ErrNotFound
		}
		return next(c)
	}
}

// people refuses API keys: affiliate accounts belong to a signed-in person.
func people(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		key, _ := c.Get("isAPIKey").(bool)
		user, _ := c.Get("userID").(string)
		if key || user == "" {
			return echo.NewHTTPError(http.StatusForbidden, "Sign in to use the affiliate program")
		}
		return next(c)
	}
}

func internalKey(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		want := os.Getenv("PAYMENTS_INTERNAL_KEY")
		got := c.Request().Header.Get("X-API-Key")
		if got == "" {
			got = strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")
		}
		if want == "" || subtle.ConstantTimeCompare([]byte(got), []byte(want)) != 1 {
			return echo.ErrUnauthorized
		}
		return next(c)
	}
}

type codeBody struct {
	Code string `json:"code"`
}

func (h *handler) click(c echo.Context) error {
	var b codeBody
	// The website sends a beacon, which arrives as text/plain.
	if json.NewDecoder(c.Request().Body).Decode(&b) == nil && b.Code != "" {
		if err := Click(h.db, b.Code); err != nil {
			return err
		}
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *handler) me(c echo.Context) error {
	a, err := find(h.db, "user_id = ?", c.Get("userID"))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON(http.StatusOK, map[string]any{"joined": false})
	}
	if err != nil {
		return err
	}
	return h.summary(c, a)
}

func (h *handler) join(c echo.Context) error {
	a, err := Join(h.db, c.Get("userID").(string))
	if err != nil {
		return err
	}
	return h.summary(c, a)
}

func (h *handler) summary(c echo.Context, a *models.Affiliate) error {
	s, err := Summarize(h.db, a, h.now())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"joined": true, "summary": s})
}

func (h *handler) attribute(c echo.Context) error {
	var b codeBody
	if err := c.Bind(&b); err != nil || b.Code == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing referral code")
	}
	team, _ := c.Get("teamID").(string)
	if err := Attribute(h.db, team, c.Get("userID").(string), b.Code, h.now()); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *handler) payment(c echo.Context) error {
	var p Payment
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid payment")
	}
	if err := RecordPayment(h.db, p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *handler) reversal(c echo.Context) error {
	var b struct {
		PaymentID string `json:"payment_id"`
	}
	if err := c.Bind(&b); err != nil || b.PaymentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing payment_id")
	}
	if err := Reverse(h.db, b.PaymentID, h.now()); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *handler) due(c echo.Context) error {
	rows, err := DuePayouts(h.db, h.now())
	if err != nil {
		return err
	}
	if rows == nil {
		rows = []Due{}
	}
	return c.JSON(http.StatusOK, rows)
}

func (h *handler) pay(c echo.Context) error {
	var b struct {
		Currency string `json:"currency"`
	}
	_ = c.Bind(&b)
	if b.Currency == "" {
		b.Currency = "USD"
	}
	p, err := Pay(h.db, c.Param("id"), b.Currency, h.now())
	if errors.Is(err, ErrBelowMinimum) {
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	}
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, p)
}
