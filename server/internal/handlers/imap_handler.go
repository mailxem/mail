package handlers

import (
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/mail"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"kori/internal/mailconnect"
	"kori/internal/models"
	"kori/internal/utils"
)

type IMAPHandler struct{ db *gorm.DB }

func mailboxAddresses(addresses []*mail.Address) string {
	if len(addresses) == 0 {
		return ""
	}
	return utils.FormatAddresses(addresses)
}

func NewIMAPHandler(db *gorm.DB) *IMAPHandler { return &IMAPHandler{db: db} }

type IMAPCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Server   string `json:"host"`
	Port     int    `json:"port"`
}

func (h *IMAPHandler) TestConnection(c echo.Context) error {
	var credentials IMAPCredentials
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, 16*1024)
	if err := c.Bind(&credentials); err != nil {
		return echo.NewHTTPError(400, "Invalid connection details")
	}
	im, err := utils.DialIMAP(c.Request().Context(), credentials.Server, credentials.Port, credentials.Username, credentials.Password)
	if err != nil {
		return echo.NewHTTPError(400, "IMAP connection failed. Check host, credentials, and TLS configuration.")
	}
	defer im.Close()
	return c.JSON(200, map[string]string{"message": "Connection successful"})
}
func (h *IMAPHandler) connect(c echo.Context) (*client.Client, *models.IMAPConfig, error) {
	teamID, _ := c.Get("teamID").(string)
	if teamID == "" {
		return nil, nil, echo.NewHTTPError(401, "Authentication required")
	}
	cfg, err := models.GetIMAPConfig(teamID, c.QueryParam("config_id"), h.db.WithContext(c.Request().Context()))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, echo.NewHTTPError(404, "Connect a mailbox in IMAP settings")
	}
	if err != nil {
		return nil, nil, echo.NewHTTPError(500, "Unable to load mailbox")
	}
	connection, err := mailconnect.Find(h.db, teamID, cfg.ID, true)
	if err != nil {
		return nil, nil, echo.NewHTTPError(409, "Mailbox disconnected. Reconnect in settings.")
	}
	var im *client.Client
	if connection != nil {
		if connection.Provider != mailconnect.Google {
			return nil, nil, echo.NewHTTPError(400, "Unsupported mailbox provider")
		}
		token, tokenErr := mailconnect.GoogleToken(c.Request().Context(), h.db, connection)
		if tokenErr != nil {
			return nil, nil, echo.NewHTTPError(409, "Google authorization needs attention. Reconnect in settings.")
		}
		im, err = utils.DialIMAPAuth(c.Request().Context(), "imap.gmail.com", 993, connection.Address, mailconnect.XOAUTH2{Username: connection.Address, Token: token})
	} else {
		im, err = utils.DialIMAP(c.Request().Context(), cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	}
	if err != nil {
		return nil, nil, echo.NewHTTPError(502, "Mailbox connection failed. Check credentials and TLS settings.")
	}
	return im, cfg, nil
}
func (h *IMAPHandler) GetFolders(c echo.Context) error {
	im, _, err := h.connect(c)
	if err != nil {
		return err
	}
	defer im.Close()
	folders := make(chan *imap.MailboxInfo)
	done := make(chan error, 1)
	go func() { done <- im.List("", "*", folders) }()
	result := make([]imap.MailboxInfo, 0)
	for folder := range folders {
		if folder != nil {
			result = append(result, *folder)
		}
	}
	if err := <-done; err != nil {
		return echo.NewHTTPError(502, "Unable to list mailbox folders")
	}
	return c.JSON(200, result)
}

type pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func parseMailPagination(c echo.Context) (pagination, error) {
	p := pagination{Limit: 20}
	if raw := c.QueryParam("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > 100 {
			return p, echo.NewHTTPError(400, "Limit must be between 1 and 100")
		}
		p.Limit = n
	}
	if raw := c.QueryParam("offset"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 {
			return p, echo.NewHTTPError(400, "Offset must be non-negative")
		}
		p.Offset = n
	} else if raw := c.QueryParam("page"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 || n > 1000000 {
			return p, echo.NewHTTPError(400, "Invalid page")
		}
		p.Offset = n * p.Limit
	}
	return p, nil
}
func mailCriteria(c echo.Context) (*imap.SearchCriteria, error) {
	criteria := imap.NewSearchCriteria()
	for name, target := range map[string]*time.Time{"since": &criteria.Since, "before": &criteria.Before} {
		if raw := c.QueryParam(name); raw != "" {
			t, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				return nil, echo.NewHTTPError(400, "Invalid "+name+" time")
			}
			*target = t
		}
	}
	for _, name := range []string{"subject", "from", "to", "cc", "bcc"} {
		if value := c.QueryParam(name); value != "" {
			criteria.Header.Add(strings.ToUpper(name), value)
		}
	}
	if value := c.QueryParam("body"); value != "" {
		criteria.Body = []string{value}
	}
	if value := c.QueryParam("q"); value != "" {
		criteria.Text = []string{value}
	}
	return criteria, nil
}
func (h *IMAPHandler) GetEmails(c echo.Context) error {
	p, err := parseMailPagination(c)
	if err != nil {
		return err
	}
	criteria, err := mailCriteria(c)
	if err != nil {
		return err
	}
	folder := c.QueryParam("folder")
	if folder == "" {
		return echo.NewHTTPError(400, "Folder is required")
	}
	im, cfg, err := h.connect(c)
	if err != nil {
		return err
	}
	defer im.Close()
	status, err := im.Select(folder, true)
	if err != nil {
		return echo.NewHTTPError(502, "Unable to open mailbox folder")
	}
	uids, err := im.UidSearch(criteria)
	if err != nil {
		return echo.NewHTTPError(502, "Unable to search mailbox")
	}
	sort.Slice(uids, func(i, j int) bool { return uids[i] > uids[j] })
	response := FolderData{FolderName: folder, TotalEmails: len(uids), Limit: p.Limit, Offset: p.Offset, UIDValidity: status.UidValidity, Emails: []EmailMessage{}}
	if p.Offset >= len(uids) {
		return c.JSON(200, response)
	}
	uids = uids[p.Offset:min(p.Offset+p.Limit, len(uids))]
	set := new(imap.SeqSet)
	set.AddNum(uids...)
	// Fetch sizes first so a legitimate large message is rejected before its
	// literal is buffered by go-imap. BODY.PEEK keeps listing from marking read.
	sizes := make(chan *imap.Message)
	sizeDone := make(chan error, 1)
	go func() { sizeDone <- im.UidFetch(set, []imap.FetchItem{imap.FetchUid, imap.FetchRFC822Size}, sizes) }()
	var totalBytes uint64
	for message := range sizes {
		if message != nil {
			totalBytes += uint64(message.Size)
		}
	}
	if err := <-sizeDone; err != nil {
		return echo.NewHTTPError(502, "Unable to load message sizes")
	}
	if totalBytes > 25*1024*1024 {
		return echo.NewHTTPError(413, "Messages exceed the 25 MiB response limit. Use a smaller page size.")
	}
	section := &imap.BodySectionName{Peek: true}
	messages := make(chan *imap.Message)
	done := make(chan error, 1)
	go func() {
		done <- im.UidFetch(set, []imap.FetchItem{imap.FetchUid, imap.FetchFlags, section.FetchItem()}, messages)
	}()
	indexed := make(map[uint32]EmailMessage, len(uids))
	var readErr error
	budget := int64(25 * 1024 * 1024)
	for message := range messages {
		if readErr != nil || message == nil {
			continue
		}
		literal := message.GetBody(section)
		if literal == nil {
			readErr = errors.New("Missing message body")
			continue
		}
		if int64(literal.Len()) > budget {
			readErr = echo.NewHTTPError(413, "Messages exceed the 25 MiB response limit. Use a smaller page size.")
			continue
		}
		budget -= int64(literal.Len())
		parsed, err := utils.ParseEmail(io.LimitReader(literal, int64(literal.Len())))
		if err != nil {
			readErr = echo.NewHTTPError(502, "Unable to parse mailbox message")
			continue
		}
		body := parsed.BodyHTML
		if body == "" {
			body = "<pre>" + html.EscapeString(parsed.BodyText) + "</pre>"
		}
		flags := message.Flags
		if flags == nil {
			flags = []string{}
		}
		indexed[message.Uid] = EmailMessage{ID: fmt.Sprintf("%s:%s:%d:%d", cfg.ID, folder, status.UidValidity, message.Uid), UID: message.Uid, UIDValidity: status.UidValidity, Body: body, Flags: flags, Attachments: parsed.Attachments, From: mailboxAddresses(parsed.From), To: mailboxAddresses(parsed.To), Cc: mailboxAddresses(parsed.Cc), Bcc: mailboxAddresses(parsed.Bcc), ReplyTo: mailboxAddresses(parsed.ReplyTo), Subject: parsed.Subject, Date: parsed.Date.Format(time.RFC3339), MessageID: parsed.MessageID, LegacyMessageID: parsed.MessageID}
	}
	if err := <-done; err != nil {
		return echo.NewHTTPError(502, "Unable to fetch mailbox messages")
	}
	if readErr != nil {
		return readErr
	}
	for _, uid := range uids {
		if message, ok := indexed[uid]; ok {
			response.Emails = append(response.Emails, message)
		}
	}
	return c.JSON(200, response)
}

// ChangeFlags addresses a message by UID and UIDVALIDITY, never by its changing
// sequence number. It cannot permanently delete messages or expunge a mailbox.
func (h *IMAPHandler) ChangeFlags(c echo.Context) error {
	var request struct {
		Folder      string `json:"folder"`
		UID         uint32 `json:"uid"`
		UIDValidity uint32 `json:"uidValidity"`
		Flag        string `json:"flag"`
		Enabled     bool   `json:"enabled"`
	}
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, 4096)
	if err := c.Bind(&request); err != nil || request.UID == 0 || request.UIDValidity == 0 || request.Folder == "" {
		return echo.NewHTTPError(400, "Folder, UID, and UIDVALIDITY are required")
	}
	if request.Flag != imap.SeenFlag && request.Flag != imap.FlaggedFlag {
		return echo.NewHTTPError(400, "Only read and starred flags can be changed")
	}
	im, _, err := h.connect(c)
	if err != nil {
		return err
	}
	defer im.Close()
	status, err := im.Select(request.Folder, false)
	if err != nil {
		return echo.NewHTTPError(502, "Unable to open mailbox folder")
	}
	if status.UidValidity != request.UIDValidity {
		return echo.NewHTTPError(409, "Mailbox changed. Refresh before updating this message.")
	}
	set := new(imap.SeqSet)
	set.AddNum(request.UID)
	op := imap.FlagsOp(imap.RemoveFlags)
	if request.Enabled {
		op = imap.AddFlags
	}
	if err := im.UidStore(set, imap.FormatFlagsOp(op, true), []interface{}{request.Flag}, nil); err != nil {
		return echo.NewHTTPError(502, "Unable to update message")
	}
	return c.NoContent(204)
}

type FolderData struct {
	FolderName  string         `json:"folder_name"`
	TotalEmails int            `json:"total_emails"`
	Limit       int            `json:"limit"`
	Offset      int            `json:"offset"`
	UIDValidity uint32         `json:"uidValidity"`
	Emails      []EmailMessage `json:"emails"`
}
type EmailMessage struct {
	ID              string                  `json:"id"`
	UID             uint32                  `json:"uid"`
	UIDValidity     uint32                  `json:"uidValidity"`
	Body            string                  `json:"body"`
	Flags           []string                `json:"flags"`
	To              string                  `json:"to"`
	Cc              string                  `json:"cc"`
	Bcc             string                  `json:"bcc"`
	From            string                  `json:"from"`
	Subject         string                  `json:"subject"`
	Date            string                  `json:"date"`
	MessageID       string                  `json:"messageId"`
	LegacyMessageID string                  `json:"message_id"`
	Attachments     []utils.EmailAttachment `json:"attachments"`
	ReplyTo         string                  `json:"reply_to"`
}
