package handlers

import (
	"github.com/labstack/echo/v4"
	"kori/internal/ai"
	"net/http"
)

func EmailDraftHandler(writer *ai.EmailWriter) echo.HandlerFunc {
	return func(c echo.Context) error {
		if team(c) == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Sign in to write an email")
		}
		var input ai.EmailDraftRequest
		if err := c.Bind(&input); err != nil {
			return bad("Invalid draft request")
		}
		if (input.Format != "" && input.Format != "text" && input.Format != "design") || len(input.Instruction) < 3 || len(input.Instruction) > 4000 || len(input.Body) > 16000 || len(input.Subject) > 200 || len(input.Tone) > 80 {
			return bad("Provide instructions of 3–4000 characters and a draft under 16000 characters")
		}
		draft, err := writer.Draft(c.Request().Context(), input)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadGateway, err.Error())
		}
		return c.JSON(200, draft)
	}
}
