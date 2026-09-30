package handlers

import (
	"encoding/base64"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"kori/internal/receiving"
)

func managed(c echo.Context, db *gorm.DB) (*receiving.Service, bool, error) {
	s := receiving.Current()
	if s == nil || s.DB != db {
		return nil, false, nil
	}
	team, _ := c.Get("teamID").(string)
	owns, err := s.OwnsMailbox(c.Request().Context(), team, c.QueryParam("config_id"))
	if err != nil {
		return s, true, echo.NewHTTPError(500, "Unable to load mailbox")
	}
	return s, owns, nil
}
func managedError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return echo.NewHTTPError(404, "Message is unavailable")
	}
	if errors.Is(err, receiving.ErrUIDValidity) {
		return echo.NewHTTPError(409, "Mailbox changed. Refresh and try again.")
	}
	if errors.Is(err, receiving.ErrAttachmentLarge) {
		return echo.NewHTTPError(413, "Attachment exceeds the 10 MiB download limit")
	}
	return echo.NewHTTPError(502, "Managed mailbox is temporarily unavailable")
}
func (h *IMAPHandler) managedFolders(c echo.Context, s *receiving.Service) error {
	_, folders, err := s.Folders(c.Request().Context(), c.Get("teamID").(string), c.QueryParam("config_id"))
	if err != nil {
		return managedError(err)
	}
	return c.JSON(200, folders)
}
func (h *IMAPHandler) managedEmails(c echo.Context, s *receiving.Service, p pagination) error {
	for _, name := range []string{"since", "before", "subject", "from", "to", "cc", "bcc", "body"} {
		if c.QueryParam(name) != "" {
			return echo.NewHTTPError(400, "This filter is unavailable for managed mailboxes")
		}
	}
	if len(c.QueryParam("q")) > 200 {
		return echo.NewHTTPError(400, "Search is too long")
	}
	folder := c.QueryParam("folder")
	if folder == "" {
		return echo.NewHTTPError(400, "Folder is required")
	}
	page, err := s.ListMessages(c.Request().Context(), c.Get("teamID").(string), c.QueryParam("config_id"), folder, c.QueryParam("q"), p.Limit, p.Offset, p.BeforeUID)
	if err != nil {
		return managedError(err)
	}
	return c.JSON(200, page)
}
func (h *IMAPHandler) managedHead(c echo.Context, s *receiving.Service) error {
	result, err := s.Head(c.Request().Context(), c.Get("teamID").(string), c.QueryParam("config_id"), c.QueryParam("folder"), c.QueryParam("q"))
	if err != nil {
		return managedError(err)
	}
	return c.JSON(200, result)
}
func (h *IMAPHandler) managedMessage(c echo.Context, s *receiving.Service, q messageQuery) error {
	message, err := s.Message(c.Request().Context(), c.Get("teamID").(string), c.QueryParam("config_id"), q.Folder, q.UID, q.UIDValidity)
	if err != nil {
		return managedError(err)
	}
	return c.JSON(200, message)
}
func (h *IMAPHandler) managedFlags(c echo.Context, s *receiving.Service, request struct {
	Folder      string `json:"folder"`
	UID         uint32 `json:"uid"`
	UIDValidity uint32 `json:"uidValidity"`
	Flag        string `json:"flag"`
	Enabled     bool   `json:"enabled"`
}) error {
	if err := s.ChangeFlags(c.Request().Context(), c.Get("teamID").(string), c.QueryParam("config_id"), request.Folder, request.UID, request.UIDValidity, request.Flag, request.Enabled); err != nil {
		return managedError(err)
	}
	return c.NoContent(204)
}
func (h *IMAPHandler) managedAttachment(c echo.Context, s *receiving.Service) error {
	q, err := parseMessageQuery(c)
	if err != nil {
		return err
	}
	data, err := s.Attachment(c.Request().Context(), c.Get("teamID").(string), c.QueryParam("config_id"), q.Folder, q.UID, q.UIDValidity, c.QueryParam("attachment_id"))
	if err != nil {
		return managedError(err)
	}
	return c.JSON(http.StatusOK, map[string]string{"Data": base64.StdEncoding.EncodeToString(data)})
}
