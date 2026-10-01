package handlers

import (
	"context"
	"encoding/base64"
	"errors"
	"html"
	"net/http"
	"net/mail"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/labstack/echo/v4"
	"kori/internal/mailconnect"
	"kori/internal/models"
	"kori/internal/utils"
)

type gmailMailboxAPI interface {
	Labels(context.Context) ([]mailconnect.GmailLabel, error)
	Messages(context.Context, string, string, string, int) (*mailconnect.GmailMessagePage, error)
	Message(context.Context, string, string) (*mailconnect.GmailMessage, error)
	ModifyLabels(context.Context, string, []string, []string) error
	Attachment(context.Context, string, string) ([]byte, error)
	History(context.Context, string, string) (mailconnect.GmailHistoryResult, error)
}

func (h *IMAPHandler) gmail(c echo.Context) (gmailMailboxAPI, *models.IMAPConfig, bool, error) {
	team, _ := c.Get("teamID").(string)
	if team == "" {
		return nil, nil, true, echo.NewHTTPError(401, "Authentication required")
	}
	if h.gmailConnect != nil {
		return h.gmailConnect(c)
	}
	client, cfg, handled, err := mailconnect.NewGmailClient(c.Request().Context(), h.db, team, c.QueryParam("config_id"))
	if err != nil {
		return nil, cfg, handled, echo.NewHTTPError(409, "Google authorization needs attention. Reconnect in settings.")
	}
	return client, cfg, handled, nil
}

func gmailError(c echo.Context, err error) error {
	var apiErr *mailconnect.GmailAPIError
	if errors.As(err, &apiErr) {
		if apiErr.Status == 403 && (apiErr.Reason == "accessNotConfigured" || apiErr.Reason == "serviceDisabled") {
			return echo.NewHTTPError(503, "The Gmail API is not enabled for this OAuth project")
		}
		if apiErr.Status == 403 && (apiErr.Reason == "rateLimitExceeded" || apiErr.Reason == "userRateLimitExceeded" || apiErr.Reason == "quotaExceeded" || apiErr.Reason == "dailyLimitExceeded") {
			if apiErr.RetryAfter != "" {
				c.Response().Header().Set("Retry-After", apiErr.RetryAfter)
			}
			return echo.NewHTTPError(429, "Gmail quota is temporarily exhausted. Try again later.")
		}
		switch apiErr.Status {
		case 404:
			return echo.NewHTTPError(404, "Gmail message is unavailable")
		case 429:
			if apiErr.RetryAfter != "" {
				c.Response().Header().Set("Retry-After", apiErr.RetryAfter)
			}
			return echo.NewHTTPError(429, "Gmail is temporarily rate limited. Try again shortly.")
		case 401, 403:
			return echo.NewHTTPError(409, "Google authorization needs attention. Reconnect in settings.")
		}
	}
	return echo.NewHTTPError(502, "Gmail is temporarily unavailable")
}

func gmailLabelDisplay(label mailconnect.GmailLabel) string {
	if friendly := map[string]string{
		"INBOX": "Inbox", "STARRED": "Starred", "YELLOW_STAR": "Yellow star", "SENT": "Sent", "DRAFT": "Drafts",
		"IMPORTANT": "Important", "UNREAD": "Unread", "SPAM": "Spam", "TRASH": "Trash", "CHAT": "Chats",
		"CATEGORY_PERSONAL": "Primary", "CATEGORY_SOCIAL": "Social", "CATEGORY_PROMOTIONS": "Promotions",
		"CATEGORY_UPDATES": "Updates", "CATEGORY_FORUMS": "Forums",
	}[label.ID]; friendly != "" {
		return friendly
	}
	if label.Name != "" {
		return label.Name
	}
	return strings.Title(strings.ToLower(strings.ReplaceAll(label.ID, "_", " ")))
}

func gmailLabelRank(id string) int {
	order := map[string]int{
		"INBOX": 0, "STARRED": 1, "YELLOW_STAR": 1, "SENT": 2, "DRAFT": 3, "IMPORTANT": 4,
		"UNREAD": 5, "CATEGORY_PERSONAL": 6, "CATEGORY_SOCIAL": 7, "CATEGORY_PROMOTIONS": 8,
		"CATEGORY_UPDATES": 9, "CATEGORY_FORUMS": 10, "CHAT": 11, "SPAM": 12, "TRASH": 13,
	}
	if rank, ok := order[id]; ok {
		return rank
	}
	return 100
}

func gmailFlags(labels []string) []string {
	flags, seen := []string{}, true
	for _, label := range labels {
		if label == "UNREAD" {
			seen = false
		}
		if label == "STARRED" {
			flags = append(flags, imap.FlaggedFlag)
		}
	}
	if seen {
		flags = append(flags, imap.SeenFlag)
	}
	return flags
}

func gmailDate(message mailconnect.GmailMessage) string {
	if parsed, err := mail.ParseDate(message.Header("Date")); err == nil {
		return parsed.Format(time.RFC3339)
	}
	if millis, err := strconv.ParseInt(message.InternalDate, 10, 64); err == nil && millis > 0 {
		return time.UnixMilli(millis).Format(time.RFC3339)
	}
	return ""
}

func gmailAttachments(message mailconnect.GmailMessage) []utils.EmailAttachment {
	metadata := message.Attachments()
	if len(metadata) == 0 {
		return nil
	}
	result := make([]utils.EmailAttachment, 0, len(metadata))
	for _, item := range metadata {
		result = append(result, utils.EmailAttachment{Filename: item.Filename, MIMEType: item.MIMEType, Size: item.Size, AttachmentID: item.AttachmentID})
	}
	return result
}

func gmailEmail(cfgID string, message mailconnect.GmailMessage, body, warning string) EmailMessage {
	rfcID := message.Header("Message-ID")
	return EmailMessage{ID: cfgID + ":gmail:" + message.ID, ProviderMessageID: message.ID, Body: body, Flags: gmailFlags(message.LabelIDs), Attachments: gmailAttachments(message), From: message.Header("From"), To: message.Header("To"), Cc: message.Header("Cc"), Bcc: message.Header("Bcc"), ReplyTo: message.Header("Reply-To"), Subject: message.Header("Subject"), Date: gmailDate(message), MessageID: rfcID, LegacyMessageID: rfcID, Warning: warning}
}

func gmailFolderMatches(message *mailconnect.GmailMessage, folder string) bool {
	for _, label := range message.LabelIDs {
		if label == folder {
			return true
		}
	}
	return false
}

func (h *IMAPHandler) gmailFolders(c echo.Context, client gmailMailboxAPI) error {
	labels, err := client.Labels(c.Request().Context())
	if err != nil {
		return gmailError(c, err)
	}
	type folder struct {
		Name        string   `json:"Name"`
		DisplayName string   `json:"DisplayName"`
		Attributes  []string `json:"Attributes"`
	}
	result := make([]folder, 0, len(labels))
	for _, label := range labels {
		if label.ID != "" {
			result = append(result, folder{Name: label.ID, DisplayName: gmailLabelDisplay(label), Attributes: []string{}})
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		ri, rj := gmailLabelRank(result[i].Name), gmailLabelRank(result[j].Name)
		if ri != rj {
			return ri < rj
		}
		if result[i].DisplayName != result[j].DisplayName {
			return strings.ToLower(result[i].DisplayName) < strings.ToLower(result[j].DisplayName)
		}
		return result[i].Name < result[j].Name
	})
	return c.JSON(200, result)
}

func (h *IMAPHandler) gmailEmails(c echo.Context, client gmailMailboxAPI, cfg *models.IMAPConfig) error {
	folder, query, pageToken := c.QueryParam("folder"), c.QueryParam("q"), c.QueryParam("page_token")
	if !validGmailIDParam(folder) || len(query) > 1024 || strings.ContainsRune(query, '\x00') || (pageToken != "" && !validOpaqueParam(pageToken)) {
		return echo.NewHTTPError(400, "Invalid Gmail mailbox query")
	}
	if c.QueryParam("offset") != "" || c.QueryParam("page") != "" || c.QueryParam("before_uid") != "" {
		return echo.NewHTTPError(400, "Gmail pagination uses page_token only")
	}
	limit := 20
	if raw := c.QueryParam("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 20 {
			return echo.NewHTTPError(400, "Gmail limit must be between 1 and 20")
		}
		limit = parsed
	}
	page, err := client.Messages(c.Request().Context(), folder, query, pageToken, limit)
	if err != nil {
		return gmailError(c, err)
	}
	response := FolderData{FolderName: folder, TotalEmails: page.ResultSizeEstimate, Limit: limit, Emails: []EmailMessage{}, NextPageToken: page.NextPageToken, HistoryID: page.HistoryID, TotalIsEstimate: true}
	for _, message := range page.Messages {
		response.Emails = append(response.Emails, gmailEmail(cfg.ID, message, html.EscapeString(message.Snippet), ""))
		if response.HistoryID == "" {
			response.HistoryID = message.HistoryID
		}
	}
	return c.JSON(200, response)
}

func (h *IMAPHandler) gmailMessage(c echo.Context, client gmailMailboxAPI, cfg *models.IMAPConfig) error {
	folder, id := c.QueryParam("folder"), c.QueryParam("message_id")
	if !validGmailIDParam(folder) || !validGmailIDParam(id) {
		return echo.NewHTTPError(400, "Folder and Gmail message ID are required")
	}
	message, err := client.Message(c.Request().Context(), id, "full")
	if err != nil {
		return gmailError(c, err)
	}
	if !gmailFolderMatches(message, folder) {
		return echo.NewHTTPError(404, "Gmail message is unavailable in this folder")
	}
	body, isHTML, warning := message.VisibleBody(10 * 1024 * 1024)
	if body == "" {
		body = "<pre>Message content could not be displayed.</pre>"
	} else if !isHTML {
		body = "<pre>" + html.EscapeString(body) + "</pre>"
	}
	return c.JSON(200, gmailEmail(cfg.ID, *message, body, warning))
}

func (h *IMAPHandler) gmailFlags(c echo.Context, client gmailMailboxAPI) error {
	var request struct {
		Folder            string `json:"folder"`
		ProviderMessageID string `json:"providerMessageId"`
		Flag              string `json:"flag"`
		Enabled           bool   `json:"enabled"`
	}
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, 4096)
	if c.Bind(&request) != nil || !validGmailIDParam(request.Folder) || !validGmailIDParam(request.ProviderMessageID) {
		return echo.NewHTTPError(400, "Folder and Gmail message ID are required")
	}
	label := ""
	switch request.Flag {
	case imap.SeenFlag:
		label = "UNREAD"
	case imap.FlaggedFlag:
		label = "STARRED"
	default:
		return echo.NewHTTPError(400, "Only read and starred flags can be changed")
	}
	add, remove := []string{}, []string{}
	if (request.Flag == imap.SeenFlag && request.Enabled) || (request.Flag == imap.FlaggedFlag && !request.Enabled) {
		remove = []string{label}
	} else {
		add = []string{label}
	}
	if err := client.ModifyLabels(c.Request().Context(), request.ProviderMessageID, add, remove); err != nil {
		return gmailError(c, err)
	}
	return c.NoContent(204)
}

func (h *IMAPHandler) gmailHead(c echo.Context, client gmailMailboxAPI) error {
	folder, historyID := c.QueryParam("folder"), c.QueryParam("history_id")
	if !validGmailIDParam(folder) {
		return echo.NewHTTPError(400, "Folder is required")
	}
	if historyID == "" {
		return echo.NewHTTPError(400, "Gmail history ID is required")
	}
	result, err := client.History(c.Request().Context(), folder, historyID)
	if err != nil {
		return gmailError(c, err)
	}
	if result.Stale {
		return c.JSON(200, map[string]any{"history_id": historyID, "changed": false, "reset_required": true})
	}
	return c.JSON(200, map[string]any{"history_id": result.HistoryID, "changed": result.Changed})
}

func (h *IMAPHandler) GetAttachment(c echo.Context) error {
	if service, handled, managedErr := managed(c, h.db); handled {
		if managedErr != nil {
			return managedErr
		}
		return h.managedAttachment(c, service)
	}
	if h.gmailConnect == nil {
		if relay, _, relayErr := h.cloudflareRelay(c); relayErr != nil {
			return relayErr
		} else if relay != nil {
			return echo.NewHTTPError(400, "Attachments are downloaded with the message for this mailbox provider")
		}
	}
	client, _, handled, err := h.gmail(c)
	if handled && err != nil {
		return err
	}
	if !handled {
		return echo.NewHTTPError(400, "Attachments are downloaded with the message for this mailbox provider")
	}
	folder, messageID, attachmentID := c.QueryParam("folder"), c.QueryParam("message_id"), c.QueryParam("attachment_id")
	if !validGmailIDParam(folder) || !validGmailIDParam(messageID) || !validOpaqueParam(attachmentID) {
		return echo.NewHTTPError(400, "Folder, Gmail message ID, and attachment ID are required")
	}
	message, err := client.Message(c.Request().Context(), messageID, "metadata")
	if err != nil {
		return gmailError(c, err)
	}
	if !gmailFolderMatches(message, folder) {
		return echo.NewHTTPError(404, "Gmail message is unavailable in this folder")
	}
	data, err := client.Attachment(c.Request().Context(), messageID, attachmentID)
	if mailconnect.IsGmailAttachmentNotFound(err) {
		return echo.NewHTTPError(404, "Attachment is unavailable")
	}
	if mailconnect.IsGmailAttachmentTooLarge(err) {
		return echo.NewHTTPError(413, "Attachment exceeds the 10 MiB download limit")
	}
	if err != nil {
		return gmailError(c, err)
	}
	return c.JSON(200, map[string]string{"Data": base64.StdEncoding.EncodeToString(data)})
}

var gmailIDParamPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,256}$`)

func validGmailIDParam(value string) bool { return gmailIDParamPattern.MatchString(value) }
func validOpaqueParam(value string) bool {
	return value != "" && len(value) <= 4096 && !strings.ContainsAny(value, "\x00\r\n")
}
