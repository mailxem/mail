package receiving

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"kori/internal/models"
)

// acceptanceStore models the provider staging object and the durable copy.
// Keeping both keys makes Message and Attachment exercise the same object-store
// boundary as production after ProcessReceipt has indexed the message.
type acceptanceStore struct {
	mu      sync.Mutex
	objects map[string][]byte
}

func newAcceptanceStore() *acceptanceStore {
	return &acceptanceStore{objects: map[string][]byte{}}
}

func (s *acceptanceStore) stage(key string, raw []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.objects[key] = append([]byte(nil), raw...)
}

func (s *acceptanceStore) Get(_ context.Context, key string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	raw, ok := s.objects[key]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return append([]byte(nil), raw...), nil
}

func (s *acceptanceStore) Copy(_ context.Context, source, destination string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	raw, ok := s.objects[source]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	s.objects[destination] = append([]byte(nil), raw...)
	return nil
}

func acceptanceReceipt(t *testing.T, id string, recipients ...string) SESReceipt {
	t.Helper()
	event := officialReceipt(t, recipients...)
	event.Mail.MessageID = id
	event.Receipt.Action.ObjectKey = "incoming/" + id
	return event
}

func acceptanceSibling(t *testing.T, service *Service, first Mailbox, local string) Mailbox {
	t.Helper()
	box := Mailbox{
		ID:           uuid.NewString(),
		TeamID:       first.TeamID,
		DomainID:     first.DomainID,
		Address:      local + "@" + first.Address[len(first.Address)-len("example.com"):],
		Active:       true,
		SMTPConfigID: uuid.NewString(),
		Status:       "active",
		UIDValidity:  1,
		QuotaBytes:   service.Config.MailboxQuotaBytes,
		CreatedAt:    service.Now(),
		UpdatedAt:    service.Now(),
	}
	require.NoError(t, service.DB.Create(&box).Error)
	return box
}

func TestAcceptanceDuplicateAtExactQuotaIsIdempotent(t *testing.T) {
	service := receivingTestService(t)
	box := addMailbox(t, service, "quota@example.com")
	raw := plainMail(box.Address)
	require.NoError(t, service.DB.Model(&box).Updates(map[string]any{
		"quota_bytes": int64(len(raw)),
		"usage_bytes": 0,
	}).Error)
	store := newAcceptanceStore()
	event := acceptanceReceipt(t, "exact-quota", box.Address)
	store.stage(event.Receipt.Action.ObjectKey, raw)

	require.NoError(t, service.ProcessReceipt(context.Background(), event, store))
	require.NoError(t, service.ProcessReceipt(context.Background(), event, store))

	var refreshed Mailbox
	require.NoError(t, service.DB.First(&refreshed, "id = ?", box.ID).Error)
	require.EqualValues(t, len(raw), refreshed.UsageBytes)
	require.Equal(t, "active", refreshed.Status)
	var messages, blobs int64
	require.NoError(t, service.DB.Model(&Message{}).Where("mailbox_id = ?", box.ID).Count(&messages).Error)
	require.NoError(t, service.DB.Model(&Blob{}).Where("provider_id = ?", event.Mail.MessageID).Count(&blobs).Error)
	require.EqualValues(t, 1, messages)
	require.EqualValues(t, 1, blobs)
}

func TestAcceptanceMixedQuotaAndPausedRecipientsKeepSuccessfulDeliveryIdempotent(t *testing.T) {
	service := receivingTestService(t)
	free := addMailbox(t, service, "free@example.com")
	full := acceptanceSibling(t, service, free, "full")
	paused := acceptanceSibling(t, service, free, "paused")
	require.NoError(t, service.DB.Model(&full).Updates(map[string]any{"quota_bytes": 1, "usage_bytes": 1}).Error)
	require.NoError(t, service.DB.Model(&paused).Updates(map[string]any{"active": false, "status": "inactive"}).Error)
	raw := plainMail(free.Address)
	store := newAcceptanceStore()
	event := acceptanceReceipt(t, "partial-delivery", free.Address, full.Address, paused.Address)
	store.stage(event.Receipt.Action.ObjectKey, raw)

	err := service.ProcessReceipt(context.Background(), event, store)
	require.Error(t, err, "provider must retry while any envelope recipient was not accepted")
	err = service.ProcessReceipt(context.Background(), event, store)
	require.Error(t, err, "retry remains required for the unavailable recipients")

	var freeMessages, fullMessages, pausedMessages int64
	require.NoError(t, service.DB.Model(&Message{}).Where("mailbox_id = ?", free.ID).Count(&freeMessages).Error)
	require.NoError(t, service.DB.Model(&Message{}).Where("mailbox_id = ?", full.ID).Count(&fullMessages).Error)
	require.NoError(t, service.DB.Model(&Message{}).Where("mailbox_id = ?", paused.ID).Count(&pausedMessages).Error)
	require.EqualValues(t, 1, freeMessages)
	require.Zero(t, fullMessages)
	require.Zero(t, pausedMessages)
	var refreshed Mailbox
	require.NoError(t, service.DB.First(&refreshed, "id = ?", free.ID).Error)
	require.EqualValues(t, len(raw), refreshed.UsageBytes)
	refreshed = Mailbox{}
	require.NoError(t, service.DB.First(&refreshed, "id = ?", full.ID).Error)
	require.Equal(t, "storage_full", refreshed.Status)
	var blob Blob
	require.NoError(t, service.DB.First(&blob, "provider_id = ?", event.Mail.MessageID).Error)
	require.EqualValues(t, 1, blob.References)
}

func TestAcceptanceNativeMailboxRoundTripAndTenantGuards(t *testing.T) {
	service := receivingTestService(t)
	box := addMailbox(t, service, "alex@example.com")
	store := newAcceptanceStore()
	service.Store = store
	attachment := "exact attachment bytes"
	for index, id := range []string{"native-one", "native-two", "native-three"} {
		raw := []byte("From: Sender <sender@example.net>\r\nTo: Alex <alex@example.com>\r\nSubject: native " + id + "\r\nMessage-ID: <" + id + "@example.net>\r\nMIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=x\r\n\r\n--x\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nhello " + id + "\r\n--x\r\nContent-Type: text/plain\r\nContent-Disposition: attachment; filename=note.txt\r\n\r\n" + attachment + "\r\n--x--\r\n")
		event := acceptanceReceipt(t, id, box.Address)
		store.stage(event.Receipt.Action.ObjectKey, raw)
		require.NoError(t, service.ProcessReceipt(context.Background(), event, store), index)
	}

	first, err := service.ListMessages(context.Background(), box.TeamID, box.ID, "inbox", "", 2, 0, 0)
	require.NoError(t, err)
	require.Equal(t, "INBOX", first.FolderName)
	require.EqualValues(t, 3, first.TotalEmails)
	require.Len(t, first.Emails, 2)
	require.NotNil(t, first.NextBeforeUID)
	second, err := service.ListMessages(context.Background(), box.TeamID, box.ID, "INBOX", "", 2, 0, *first.NextBeforeUID)
	require.NoError(t, err)
	require.Len(t, second.Emails, 1)
	require.Less(t, second.Emails[0].UID, first.Emails[1].UID)

	selected := first.Emails[0]
	detail, err := service.Message(context.Background(), box.TeamID, box.ID, "INBOX", selected.UID, selected.UIDValidity)
	require.NoError(t, err)
	require.Contains(t, detail.Body, "hello native-three")
	require.Len(t, detail.Attachments, 1)
	require.Equal(t, "part:0", detail.Attachments[0].AttachmentID)
	data, err := service.Attachment(context.Background(), box.TeamID, box.ID, "INBOX", selected.UID, selected.UIDValidity, "part:0")
	require.NoError(t, err)
	require.Equal(t, []byte(attachment), data)

	require.NoError(t, service.ChangeFlags(context.Background(), box.TeamID, box.ID, "INBOX", selected.UID, selected.UIDValidity, "\\Seen", true))
	require.NoError(t, service.ChangeFlags(context.Background(), box.TeamID, box.ID, "INBOX", selected.UID, selected.UIDValidity, "\\Flagged", true))
	flagged, err := service.Message(context.Background(), box.TeamID, box.ID, "Starred", selected.UID, selected.UIDValidity)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"\\Seen", "\\Flagged"}, flagged.Flags)

	_, err = service.Message(context.Background(), uuid.NewString(), box.ID, "INBOX", selected.UID, selected.UIDValidity)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	_, err = service.Message(context.Background(), box.TeamID, box.ID, "INBOX", selected.UID, selected.UIDValidity+1)
	require.ErrorIs(t, err, ErrUIDValidity)
}

type acceptanceRules struct {
	fail  bool
	ready bool
}

func (r *acceptanceRules) ActiveRuleSet(context.Context) (string, error) {
	if r.fail {
		return "", errors.New("temporary rule provider failure")
	}
	return "acceptance-rules", nil
}
func (r *acceptanceRules) PutRule(context.Context, string, []string) error {
	if r.fail {
		return errors.New("temporary rule provider failure")
	}
	r.ready = true
	return nil
}
func (r *acceptanceRules) DeleteRule(context.Context, string) error {
	if r.fail {
		return errors.New("temporary rule provider failure")
	}
	return nil
}
func (r *acceptanceRules) RuleReady(context.Context, string, []string) (bool, error) {
	if r.fail {
		return false, errors.New("temporary rule provider failure")
	}
	return r.ready, nil
}

type acceptanceResolver struct{ host string }

func (r acceptanceResolver) LookupMX(context.Context, string) ([]*net.MX, error) {
	return []*net.MX{{Host: r.host, Pref: 10}}, nil
}

func TestAcceptanceLifecycleRecoversAfterRuleProviderFailure(t *testing.T) {
	service := receivingTestService(t)
	require.NoError(t, service.DB.AutoMigrate(&models.SMTPConfig{}))
	seed := addMailbox(t, service, "lifecycle@example.com")
	rules := &acceptanceRules{fail: true}
	service.Rules = rules
	service.Config.RuleSet = "acceptance-rules"
	service.Config.MaxMailboxesPerDomain = 10
	service.DNS = acceptanceResolver{host: "inbound-smtp." + service.Config.Region + ".amazonaws.com."}

	created, err := service.CreateMailbox(context.Background(), seed.TeamID, seed.DomainID, "created", "Created mailbox")
	require.NoError(t, err)
	require.Equal(t, "created@example.com", created.Address)
	require.Equal(t, "error", created.Status)
	_, err = service.CreateMailbox(context.Background(), seed.TeamID, seed.DomainID, "created", "Duplicate")
	require.ErrorIs(t, err, ErrMailboxExists)
	var senderCount int64
	require.NoError(t, service.DB.Model(&models.SMTPConfig{}).Count(&senderCount).Error)
	require.EqualValues(t, 1, senderCount)

	failed, err := service.SetActive(context.Background(), seed.TeamID, created.ID, false)
	require.NoError(t, err)
	require.False(t, failed.Active)
	require.Equal(t, "error", failed.Status)

	rules.fail = false
	resumed, err := service.SetActive(context.Background(), seed.TeamID, created.ID, true)
	require.NoError(t, err)
	require.True(t, resumed.Active)
	require.Equal(t, "active", resumed.Status)
	require.NoError(t, service.ReconcileAll(context.Background()))

	var persisted Mailbox
	require.NoError(t, service.DB.First(&persisted, "id = ?", created.ID).Error)
	require.True(t, persisted.Active)
	require.Equal(t, "active", persisted.Status)
}
