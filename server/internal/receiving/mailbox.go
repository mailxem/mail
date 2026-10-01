package receiving

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"kori/internal/utils"
)

type MailMessage struct {
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
	Attachments     []utils.EmailAttachment `json:"attachments,omitempty"`
	ReplyTo         string                  `json:"reply_to"`
}
type MailPage struct {
	FolderName    string        `json:"folder_name"`
	TotalEmails   int64         `json:"total_emails"`
	Limit         int           `json:"limit"`
	Offset        int           `json:"offset"`
	UIDValidity   uint32        `json:"uidValidity"`
	NextBeforeUID *uint32       `json:"next_before_uid,omitempty"`
	Emails        []MailMessage `json:"emails"`
}

func (s *Service) mailbox(ctx context.Context, team, id string) (Mailbox, error) {
	var b Mailbox
	err := s.DB.WithContext(ctx).Where("id = ? AND team_id = ?", id, team).First(&b).Error
	return b, err
}
func (s *Service) OwnsMailbox(ctx context.Context, team, id string) (bool, error) {
	var count int64
	err := s.DB.WithContext(ctx).Model(&Mailbox{}).Where("id = ? AND team_id = ?", id, team).Count(&count).Error
	return count == 1, err
}
func flags(m Message) []string {
	f := []string{}
	if m.Seen {
		f = append(f, "\\Seen")
	}
	if m.Starred {
		f = append(f, "\\Flagged")
	}
	return f
}
func metadata(m Message, box Mailbox) MailMessage {
	var attachments []utils.EmailAttachment
	_ = json.Unmarshal(m.AttachmentsJSON, &attachments)
	return MailMessage{ID: fmt.Sprintf("%s:%s:%d:%d", box.ID, m.Folder, m.UIDValidity, m.UID), UID: m.UID, UIDValidity: m.UIDValidity, Body: m.Snippet, Flags: flags(m), To: m.ToAddress, Cc: m.CcAddress, Bcc: m.BccAddress, From: m.FromAddress, Subject: m.Subject, Date: m.SentAt.Format(time.RFC3339), MessageID: m.RFCMessageID, LegacyMessageID: m.RFCMessageID, Attachments: attachments, ReplyTo: m.ReplyTo}
}

func canonicalFolder(folder string) string {
	if strings.EqualFold(folder, "inbox") {
		return "INBOX"
	}
	return folder
}
func (s *Service) Folders(ctx context.Context, team, id string) (Mailbox, []map[string]any, error) {
	box, err := s.mailbox(ctx, team, id)
	if err != nil {
		return box, nil, err
	}
	return box, []map[string]any{{"Name": "INBOX", "DisplayName": "Inbox", "Attributes": []string{}}, {"Name": "Starred", "DisplayName": "Starred", "Attributes": []string{}}, {"Name": "Spam", "DisplayName": "Spam", "Attributes": []string{}}}, nil
}
func (s *Service) ListMessages(ctx context.Context, team, id, folder, q string, limit, offset int, before uint32) (MailPage, error) {
	folder = canonicalFolder(folder)
	box, err := s.mailbox(ctx, team, id)
	if err != nil {
		return MailPage{}, err
	}
	if limit < 1 || limit > 100 {
		return MailPage{}, errors.New("invalid limit")
	}
	query := s.DB.WithContext(ctx).Model(&Message{}).Where("mailbox_id = ? AND team_id = ?", id, team)
	if folder == "Starred" {
		query = query.Where("starred = true")
	} else if folder == "INBOX" || folder == "Spam" {
		query = query.Where("folder = ?", folder)
	} else {
		return MailPage{}, errors.New("unsupported folder")
	}
	if q != "" {
		escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(strings.ToLower(q))
		needle := "%" + escaped + "%"
		query = query.Where("lower(subject) LIKE ? OR lower(from_address) LIKE ? OR lower(snippet) LIKE ? ESCAPE '\\'", needle, needle, needle)
	}
	var total int64
	if err = query.Count(&total).Error; err != nil {
		return MailPage{}, err
	}
	if before > 0 {
		query = query.Where("uid < ?", before)
	}
	rows := []Message{}
	if err = query.Order("uid DESC").Offset(offset).Limit(limit + 1).Find(&rows).Error; err != nil {
		return MailPage{}, err
	}
	page := MailPage{FolderName: folder, TotalEmails: total, Limit: limit, Offset: offset, UIDValidity: box.UIDValidity, Emails: []MailMessage{}}
	if len(rows) > limit {
		next := rows[limit-1].UID
		page.NextBeforeUID = &next
		rows = rows[:limit]
	}
	for _, m := range rows {
		page.Emails = append(page.Emails, metadata(m, box))
	}
	return page, nil
}
func (s *Service) Head(ctx context.Context, team, id, folder, q string) (map[string]any, error) {
	page, err := s.ListMessages(ctx, team, id, folder, q, 1, 0, 0)
	if err != nil {
		return nil, err
	}
	latest := uint32(0)
	if len(page.Emails) > 0 {
		latest = page.Emails[0].UID
	}
	return map[string]any{"total_emails": page.TotalEmails, "uidValidity": page.UIDValidity, "latest_uid": latest}, nil
}
func (s *Service) messageRow(ctx context.Context, team, id, folder string, uid, validity uint32) (Mailbox, Message, error) {
	folder = canonicalFolder(folder)
	box, err := s.mailbox(ctx, team, id)
	if err != nil {
		return box, Message{}, err
	}
	if box.UIDValidity != validity {
		return box, Message{}, ErrUIDValidity
	}
	var m Message
	q := s.DB.WithContext(ctx).Where("mailbox_id = ? AND team_id = ? AND uid = ?", id, team, uid)
	if folder == "Starred" {
		q = q.Where("starred = true")
	} else {
		q = q.Where("folder = ?", folder)
	}
	err = q.First(&m).Error
	return box, m, err
}

var ErrUIDValidity = errors.New("mailbox UIDVALIDITY changed")

func (s *Service) Message(ctx context.Context, team, id, folder string, uid, validity uint32) (MailMessage, error) {
	if s.Store == nil {
		return MailMessage{}, errors.New("managed mailbox object store unavailable")
	}
	box, row, err := s.messageRow(ctx, team, id, folder, uid, validity)
	if err != nil {
		return MailMessage{}, err
	}
	raw, err := s.Store.Get(ctx, row.RawKey)
	if err != nil {
		return MailMessage{}, err
	}
	parsed, err := utils.ParseEmail(bytes.NewReader(raw))
	if err != nil {
		return MailMessage{}, err
	}
	result := metadata(row, box)
	if parsed.BodyHTML != "" {
		result.Body = parsed.BodyHTML
	} else {
		result.Body = "<pre>" + html.EscapeString(parsed.BodyText) + "</pre>"
	}
	return result, nil
}
func (s *Service) Attachment(ctx context.Context, team, id, folder string, uid, validity uint32, selector string) ([]byte, error) {
	if s.Store == nil {
		return nil, errors.New("managed mailbox object store unavailable")
	}
	_, row, err := s.messageRow(ctx, team, id, folder, uid, validity)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(selector, "part:") {
		return nil, gorm.ErrRecordNotFound
	}
	index, err := strconv.Atoi(strings.TrimPrefix(selector, "part:"))
	if err != nil || index < 0 {
		return nil, gorm.ErrRecordNotFound
	}
	raw, err := s.Store.Get(ctx, row.RawKey)
	if err != nil {
		return nil, err
	}
	parsed, err := utils.ParseEmail(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	if index >= len(parsed.Attachments) {
		return nil, gorm.ErrRecordNotFound
	}
	if len(parsed.Attachments[index].Data) > 10*1024*1024 {
		return nil, ErrAttachmentLarge
	}
	return parsed.Attachments[index].Data, nil
}

var ErrAttachmentLarge = errors.New("attachment exceeds 10 MiB")

func (s *Service) ChangeFlags(ctx context.Context, team, id, folder string, uid, validity uint32, flag string, enabled bool) error {
	_, row, err := s.messageRow(ctx, team, id, folder, uid, validity)
	if err != nil {
		return err
	}
	column := ""
	switch flag {
	case "\\Seen":
		column = "seen"
	case "\\Flagged":
		column = "starred"
	default:
		return errors.New("unsupported flag")
	}
	return s.DB.WithContext(ctx).Model(&row).Where("mailbox_id = ? AND team_id = ?", id, team).Update(column, enabled).Error
}
