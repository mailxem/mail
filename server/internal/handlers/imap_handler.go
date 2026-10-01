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

type messageIMAPClient interface {
	Select(string, bool) (*imap.MailboxStatus, error)
	UidFetch(*imap.SeqSet, []imap.FetchItem, chan *imap.Message) error
	Close() error
}

type headIMAPClient interface {
	Select(string, bool) (*imap.MailboxStatus, error)
	UidSearch(*imap.SearchCriteria) ([]uint32, error)
	Close() error
}

type IMAPHandler struct {
	db             *gorm.DB
	messageConnect func(echo.Context) (messageIMAPClient, *models.IMAPConfig, error)
	headConnect    func(echo.Context) (headIMAPClient, error)
	gmailConnect   func(echo.Context) (gmailMailboxAPI, *models.IMAPConfig, bool, error)
}

func mailboxAddresses(addresses []*mail.Address) string {
	if len(addresses) == 0 {
		return ""
	}
	return utils.FormatAddresses(addresses)
}

func mailboxDate(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339)
}

func NewIMAPHandler(db *gorm.DB) *IMAPHandler { return &IMAPHandler{db: db} }

type IMAPCredentials struct {
	ID       string `json:"id"`
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
	if credentials.Password == "" && credentials.ID != "" {
		var count int64
		if err := h.db.Model(&models.CloudflareRelay{}).Where("imap_config_id = ? AND team_id = ?", credentials.ID, c.Get("teamID")).Count(&count).Error; err != nil {
			return echo.NewHTTPError(500, "Unable to load mailbox")
		}
		if count > 0 {
			return echo.NewHTTPError(400, "Cloudflare Worker mailboxes are tested from mail connection settings")
		}
		stored, err := models.GetIMAPConfig(c.Get("teamID").(string), credentials.ID, h.db.WithContext(c.Request().Context()))
		if err != nil {
			return echo.NewHTTPError(404, "Connection not found")
		}
		if stored.Host != credentials.Server || stored.Port != credentials.Port || stored.Username != credentials.Username {
			return echo.NewHTTPError(400, "Enter the password again when changing the host, port or username")
		}
		credentials.Password = stored.Password
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
		return nil, nil, echo.NewHTTPError(500, "Google mailbox must use the Gmail API")
	} else {
		im, err = utils.DialIMAP(c.Request().Context(), cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	}
	if err != nil {
		return nil, nil, echo.NewHTTPError(502, "Mailbox connection failed. Check credentials and TLS settings.")
	}
	return im, cfg, nil
}
func (h *IMAPHandler) GetFolders(c echo.Context) error {
	if service, handled, err := managed(c, h.db); handled {
		if err != nil {
			return err
		}
		return h.managedFolders(c, service)
	}
	if h.gmailConnect != nil {
		gmail, _, _, err := h.gmail(c)
		if err != nil {
			return err
		}
		return h.gmailFolders(c, gmail)
	}
	if handled, err := h.cloudflareFolders(c); handled {
		return err
	}
	if gmail, _, handled, err := h.gmail(c); handled {
		if err != nil {
			return err
		}
		return h.gmailFolders(c, gmail)
	}
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
	Limit     int
	Offset    int
	BeforeUID uint32
}

type messageQuery struct {
	Folder      string
	UID         uint32
	UIDValidity uint32
}

func parseMessageQuery(c echo.Context) (messageQuery, error) {
	q := messageQuery{Folder: c.QueryParam("folder")}
	uid, err := strconv.ParseUint(c.QueryParam("uid"), 10, 32)
	if err != nil || uid == 0 || q.Folder == "" || len(q.Folder) > 1024 || strings.ContainsRune(q.Folder, '\x00') {
		return q, echo.NewHTTPError(400, "Folder, UID, and UIDVALIDITY are required")
	}
	validity, err := strconv.ParseUint(c.QueryParam("uid_validity"), 10, 32)
	if err != nil || validity == 0 {
		return q, echo.NewHTTPError(400, "Folder, UID, and UIDVALIDITY are required")
	}
	q.UID, q.UIDValidity = uint32(uid), uint32(validity)
	return q, nil
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
	if raw := c.QueryParam("before_uid"); raw != "" {
		n, err := strconv.ParseUint(raw, 10, 32)
		if err != nil || n == 0 {
			return p, echo.NewHTTPError(400, "before_uid must be a positive 32-bit UID")
		}
		for _, name := range []string{"offset", "page"} {
			if value := c.QueryParam(name); value != "" {
				position, positionErr := strconv.Atoi(value)
				if positionErr != nil || position != 0 {
					return p, echo.NewHTTPError(400, "before_uid cannot be combined with offset or page")
				}
			}
		}
		p.BeforeUID = uint32(n)
	}
	return p, nil
}

// mailUIDWindow returns descending UIDs for one page. total is deliberately
// measured before cursor filtering so polling and pagination share a stable
// count of all messages matching the folder/search criteria.
func mailUIDWindow(uids []uint32, p pagination) (window []uint32, total int, next uint32) {
	sort.Slice(uids, func(i, j int) bool { return uids[i] > uids[j] })
	total = len(uids)
	if p.BeforeUID != 0 {
		start := sort.Search(len(uids), func(i int) bool { return uids[i] < p.BeforeUID })
		uids = uids[start:]
	} else if p.Offset >= len(uids) {
		return nil, total, 0
	} else {
		uids = uids[p.Offset:]
	}
	if len(uids) == 0 {
		return nil, total, 0
	}
	window = uids[:min(p.Limit, len(uids))]
	if len(uids) > len(window) {
		next = window[len(window)-1]
	}
	return window, total, next
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
	if h.gmailConnect != nil {
		gmail, cfg, _, gmailErr := h.gmail(c)
		if gmailErr != nil {
			return gmailErr
		}
		return h.gmailEmails(c, gmail, cfg)
	}
	p, err := parseMailPagination(c)
	if err != nil {
		return err
	}
	if service, handled, managedErr := managed(c, h.db); handled {
		if managedErr != nil {
			return managedErr
		}
		return h.managedEmails(c, service, p)
	}
	criteria, err := mailCriteria(c)
	if err != nil {
		return err
	}
	folder := c.QueryParam("folder")
	if folder == "" {
		return echo.NewHTTPError(400, "Folder is required")
	}
	if handled, err := h.cloudflareEmails(c, p); handled {
		return err
	}
	if gmail, cfg, handled, gmailErr := h.gmail(c); handled {
		if gmailErr != nil {
			return gmailErr
		}
		return h.gmailEmails(c, gmail, cfg)
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
	uids, total, next := mailUIDWindow(uids, p)
	response := FolderData{FolderName: folder, TotalEmails: total, Limit: p.Limit, Offset: p.Offset, UIDValidity: status.UidValidity, Emails: []EmailMessage{}}
	if next != 0 {
		response.NextBeforeUID = &next
	}
	if len(uids) == 0 {
		return c.JSON(200, response)
	}
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
		indexed[message.Uid] = EmailMessage{ID: fmt.Sprintf("%s:%s:%d:%d", cfg.ID, folder, status.UidValidity, message.Uid), UID: message.Uid, UIDValidity: status.UidValidity, Body: body, Flags: flags, Attachments: parsed.Attachments, From: mailboxAddresses(parsed.From), To: mailboxAddresses(parsed.To), Cc: mailboxAddresses(parsed.Cc), Bcc: mailboxAddresses(parsed.Bcc), ReplyTo: mailboxAddresses(parsed.ReplyTo), Subject: parsed.Subject, Date: mailboxDate(parsed.Date), MessageID: parsed.MessageID, LegacyMessageID: parsed.MessageID}
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

// GetHead returns only mailbox identity/count information for foreground
// polling. It intentionally performs no FETCH and never reads message bodies.
func (h *IMAPHandler) GetHead(c echo.Context) error {
	folder := c.QueryParam("folder")
	if folder == "" || len(folder) > 1024 || strings.ContainsRune(folder, '\x00') {
		return echo.NewHTTPError(400, "Folder is required")
	}
	query := c.QueryParam("q")
	if len(query) > 1024 || strings.ContainsRune(query, '\x00') {
		return echo.NewHTTPError(400, "Search query is invalid")
	}
	if service, handled, err := managed(c, h.db); handled {
		if err != nil {
			return err
		}
		return h.managedHead(c, service)
	}
	if h.headConnect == nil {
		if h.gmailConnect != nil {
			gmail, _, _, err := h.gmail(c)
			if err != nil {
				return err
			}
			return h.gmailHead(c, gmail)
		}
		if handled, err := h.cloudflareHead(c); handled {
			return err
		}
		if gmail, _, handled, err := h.gmail(c); handled {
			if err != nil {
				return err
			}
			return h.gmailHead(c, gmail)
		}
	}
	var im headIMAPClient
	var err error
	if h.headConnect != nil {
		im, err = h.headConnect(c)
	} else {
		connected, _, connectErr := h.connect(c)
		im, err = connected, connectErr
	}
	if err != nil {
		return err
	}
	defer im.Close()
	status, err := im.Select(folder, true)
	if err != nil {
		return echo.NewHTTPError(502, "Unable to open mailbox folder")
	}
	criteria := imap.NewSearchCriteria()
	if query != "" {
		criteria.Text = []string{query}
	}
	uids, err := im.UidSearch(criteria)
	if err != nil {
		return echo.NewHTTPError(502, "Unable to search mailbox")
	}
	var latest uint32
	for _, uid := range uids {
		if uid > latest {
			latest = uid
		}
	}
	return c.JSON(200, map[string]any{"total_emails": len(uids), "uidValidity": status.UidValidity, "latest_uid": latest})
}

// GetMessage retrieves canonical server-side content for one stable IMAP UID.
// It never accepts message content from the caller and does not expose
// attachment bytes. BODY.PEEK and a read-only mailbox keep retrieval inert.
func (h *IMAPHandler) GetMessage(c echo.Context) error {
	if service, handled, managedErr := managed(c, h.db); handled {
		if managedErr != nil {
			return managedErr
		}
		query, err := parseMessageQuery(c)
		if err != nil {
			return err
		}
		return h.managedMessage(c, service, query)
	}
	if h.messageConnect == nil {
		if h.gmailConnect != nil {
			gmail, cfg, _, err := h.gmail(c)
			if err != nil {
				return err
			}
			return h.gmailMessage(c, gmail, cfg)
		}
		if relay, _, relayErr := h.cloudflareRelay(c); relayErr != nil {
			return relayErr
		} else if relay != nil {
			query, err := parseMessageQuery(c)
			if err != nil {
				return err
			}
			_, err = h.cloudflareMessage(c, query)
			return err
		}
		if gmail, cfg, handled, gmailErr := h.gmail(c); handled {
			if gmailErr != nil {
				return gmailErr
			}
			return h.gmailMessage(c, gmail, cfg)
		}
	}
	query, err := parseMessageQuery(c)
	if err != nil {
		return err
	}
	var im messageIMAPClient
	var cfg *models.IMAPConfig
	if h.messageConnect != nil {
		im, cfg, err = h.messageConnect(c)
	} else {
		im, cfg, err = h.connect(c)
	}
	if err != nil {
		return err
	}
	defer im.Close()
	status, err := im.Select(query.Folder, true)
	if err != nil {
		return echo.NewHTTPError(502, "Unable to open mailbox folder")
	}
	if status.UidValidity != query.UIDValidity {
		return echo.NewHTTPError(409, "Mailbox changed. Refresh before loading this message.")
	}
	set := new(imap.SeqSet)
	set.AddNum(query.UID)
	sizes := make(chan *imap.Message)
	sizeDone := make(chan error, 1)
	go func() { sizeDone <- im.UidFetch(set, []imap.FetchItem{imap.FetchUid, imap.FetchRFC822Size}, sizes) }()
	var found, oversized bool
	for message := range sizes {
		if message != nil && message.Uid == query.UID {
			found = true
			if message.Size > 10*1024*1024 {
				// Keep draining: go-imap's producer must finish before returning.
				oversized = true
			}
		}
	}
	if err := <-sizeDone; err != nil {
		return echo.NewHTTPError(502, "Unable to load message size")
	}
	if oversized {
		return echo.NewHTTPError(413, "Message exceeds the 10 MiB summary limit")
	}
	if !found {
		return echo.NewHTTPError(404, "Message is unavailable")
	}
	section := &imap.BodySectionName{Peek: true}
	messages := make(chan *imap.Message)
	done := make(chan error, 1)
	go func() {
		done <- im.UidFetch(set, []imap.FetchItem{imap.FetchUid, imap.FetchFlags, section.FetchItem()}, messages)
	}()
	var result *EmailMessage
	for message := range messages {
		if message == nil || message.Uid != query.UID || result != nil {
			continue
		}
		literal := message.GetBody(section)
		if literal == nil || literal.Len() > 10*1024*1024 {
			continue
		}
		parsed, parseErr := utils.ParseEmail(io.LimitReader(literal, 10*1024*1024+1))
		if parseErr != nil {
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
		value := EmailMessage{ID: fmt.Sprintf("%s:%s:%d:%d", cfg.ID, query.Folder, status.UidValidity, message.Uid), UID: message.Uid, UIDValidity: status.UidValidity, Body: body, Flags: flags, From: mailboxAddresses(parsed.From), To: mailboxAddresses(parsed.To), Cc: mailboxAddresses(parsed.Cc), Bcc: mailboxAddresses(parsed.Bcc), ReplyTo: mailboxAddresses(parsed.ReplyTo), Subject: parsed.Subject, Date: mailboxDate(parsed.Date), MessageID: parsed.MessageID, LegacyMessageID: parsed.MessageID}
		result = &value
	}
	if err := <-done; err != nil {
		return echo.NewHTTPError(502, "Unable to fetch mailbox message")
	}
	if result == nil {
		return echo.NewHTTPError(404, "Message is unavailable")
	}
	return c.JSON(200, result)
}

// ChangeFlags addresses a message by UID and UIDVALIDITY, never by its changing
// sequence number. It cannot permanently delete messages or expunge a mailbox.
func (h *IMAPHandler) ChangeFlags(c echo.Context) error {
	if service, handled, managedErr := managed(c, h.db); handled {
		if managedErr != nil {
			return managedErr
		}
		var request struct {
			Folder      string `json:"folder"`
			UID         uint32 `json:"uid"`
			UIDValidity uint32 `json:"uidValidity"`
			Flag        string `json:"flag"`
			Enabled     bool   `json:"enabled"`
		}
		c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, 4096)
		if c.Bind(&request) != nil || request.Folder == "" || request.UID == 0 || request.UIDValidity == 0 {
			return echo.NewHTTPError(400, "Folder, UID, and UIDVALIDITY are required")
		}
		return h.managedFlags(c, service, request)
	}
	if h.gmailConnect == nil {
		if relay, _, relayErr := h.cloudflareRelay(c); relayErr != nil {
			return relayErr
		} else if relay != nil {
			return h.changeLegacyFlags(c, true)
		}
	}
	if gmail, _, handled, gmailErr := h.gmail(c); handled {
		if gmailErr != nil {
			return gmailErr
		}
		return h.gmailFlags(c, gmail)
	}
	return h.changeLegacyFlags(c, false)
}

func (h *IMAPHandler) changeLegacyFlags(c echo.Context, cloudflare bool) error {
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
	if cloudflare {
		_, err := h.cloudflareFlags(c, request)
		return err
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
	FolderName      string         `json:"folder_name"`
	TotalEmails     int            `json:"total_emails"`
	Limit           int            `json:"limit"`
	Offset          int            `json:"offset"`
	UIDValidity     uint32         `json:"uidValidity"`
	NextBeforeUID   *uint32        `json:"next_before_uid,omitempty"`
	NextPageToken   string         `json:"next_page_token,omitempty"`
	HistoryID       string         `json:"history_id,omitempty"`
	TotalIsEstimate bool           `json:"total_is_estimate,omitempty"`
	Emails          []EmailMessage `json:"emails"`
}
type EmailMessage struct {
	ID                string                  `json:"id"`
	UID               uint32                  `json:"uid"`
	UIDValidity       uint32                  `json:"uidValidity"`
	Body              string                  `json:"body"`
	Flags             []string                `json:"flags"`
	To                string                  `json:"to"`
	Cc                string                  `json:"cc"`
	Bcc               string                  `json:"bcc"`
	From              string                  `json:"from"`
	Subject           string                  `json:"subject"`
	Date              string                  `json:"date"`
	MessageID         string                  `json:"messageId"`
	LegacyMessageID   string                  `json:"message_id"`
	ProviderMessageID string                  `json:"providerMessageId,omitempty"`
	Warning           string                  `json:"warning,omitempty"`
	Attachments       []utils.EmailAttachment `json:"attachments,omitempty"`
	ReplyTo           string                  `json:"reply_to"`
}
