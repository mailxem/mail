package handlers

import (
	"errors"
	"html"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type formTheme struct {
	Preset          string `json:"preset"`
	BackgroundColor string `json:"backgroundColor"`
	CardColor       string `json:"cardColor"`
	TextColor       string `json:"textColor"`
	ButtonColor     string `json:"buttonColor"`
	ButtonTextColor string `json:"buttonTextColor"`
	LogoURL         string `json:"logoUrl"`
	Font            string `json:"font"`
	Corners         string `json:"corners"`
}

func (t formTheme) validate() error {
	for _, color := range []string{t.BackgroundColor, t.CardColor, t.TextColor, t.ButtonColor, t.ButtonTextColor} {
		if color != "" && !regexp.MustCompile(`^#[0-9a-fA-F]{6}$`).MatchString(color) {
			return bad("Choose valid six-digit hex colors")
		}
	}
	if t.Preset != "" && t.Preset != "light" && t.Preset != "dark" && t.Preset != "warm" {
		return bad("Invalid theme preset")
	}
	if t.Font != "" && t.Font != "sans" && t.Font != "serif" && t.Font != "mono" {
		return bad("Invalid form font")
	}
	if t.Corners != "" && t.Corners != "rounded" && t.Corners != "square" {
		return bad("Invalid corner style")
	}
	if t.LogoURL != "" {
		u, err := url.Parse(t.LogoURL)
		if err != nil || len(t.LogoURL) > 2048 || u.Scheme != "https" || u.Hostname() == "" || u.User != nil {
			return bad("Use an HTTPS URL for your logo")
		}
	}
	return nil
}

type formSubmissionInput struct {
	Fields      map[string]string `json:"fields"`
	Consent     bool              `json:"consent"`
	Website     string            `json:"website"`
	RequestID   string            `json:"requestId"`
	Version     int               `json:"version"`
	SessionID   string            `json:"sessionId"`
	ResumeToken string            `json:"resumeToken"`
	PageID      string            `json:"pageId"`
	Attribution formAttribution   `json:"attribution"`
}

func isHTMLFormPost(c echo.Context) bool {
	media, _, _ := mime.ParseMediaType(c.Request().Header.Get(echo.HeaderContentType))
	return media == echo.MIMEApplicationForm || media == echo.MIMEMultipartForm
}
func bindFormSubmission(c echo.Context) (formSubmissionInput, error) {
	var in formSubmissionInput
	if !isHTMLFormPost(c) {
		err := c.Bind(&in)
		return in, err
	}
	// FormParams parses the body; use PostForm/MultipartForm directly to exclude query parameters.
	if _, err := c.FormParams(); err != nil {
		return in, err
	}
	values := c.Request().PostForm
	if c.Request().MultipartForm != nil {
		defer c.Request().MultipartForm.RemoveAll()
		if len(c.Request().MultipartForm.File) > 0 {
			return in, errors.New("file fields are unsupported")
		}
		values = c.Request().MultipartForm.Value
	}
	in.Fields = map[string]string{}
	for key := range values {
		in.Fields[key] = values.Get(key)
	}
	in.Consent = values.Get("consent") == "on" || values.Get("consent") == "true" || values.Get("consent") == "1"
	in.Website = values.Get("website")
	in.RequestID = values.Get("requestId")
	in.SessionID = values.Get("sessionId")
	in.ResumeToken = values.Get("resumeToken")
	if values.Get("version") != "" {
		var err error
		in.Version, err = strconv.Atoi(values.Get("version"))
		if err != nil {
			return in, err
		}
	}
	if in.RequestID == "" {
		in.RequestID = uuid.NewString()
	}
	return in, nil
}
func formHTMLResponse(c echo.Context, status int, title, message string) error {
	c.Response().Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; base-uri 'none'; frame-ancestors 'none'")
	c.Response().Header().Set("Cache-Control", "no-store")
	return c.HTML(status, `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>`+html.EscapeString(title)+`</title></head><body style="font:16px system-ui;background:#f5f2fa;color:#302839;padding:24px"><main style="max-width:480px;margin:12vh auto;background:white;padding:32px;border-radius:20px"><h1>`+html.EscapeString(title)+`</h1><p>`+html.EscapeString(message)+`</p><p style="font-size:12px">Powered by Xem</p></main></body></html>`)
}
func formSubmissionSuccess(c echo.Context, message string) error {
	if !isHTMLFormPost(c) {
		return c.JSON(200, map[string]interface{}{"ok": true, "message": message})
	}
	if strings.TrimSpace(message) == "" {
		message = "Thanks for your response!"
	}
	return formHTMLResponse(c, http.StatusOK, "Thanks for your response.", message)
}
func (h *MarketingHandler) SubmitForm(c echo.Context) error {
	err := h.submitForm(c)
	if err == nil || !isHTMLFormPost(c) {
		return err
	}
	status, message := 500, "We couldn’t submit your form. Please try again."
	var he *echo.HTTPError
	if errors.As(err, &he) && he.Code < 500 {
		status = he.Code
		if text, ok := he.Message.(string); ok {
			message = text
		}
	}
	return formHTMLResponse(c, status, "We couldn’t submit your form", message+" Go back to your form to try again.")
}
