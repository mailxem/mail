package receiving

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"kori/internal/sending"
)

const receiptTopic = "arn:aws:sns:us-east-2:123456789012:xem-receiving"

type memoryObjects struct {
	raw    []byte
	gets   int
	copies int
}

func (m *memoryObjects) Get(context.Context, string) ([]byte, error) {
	m.gets++
	return append([]byte(nil), m.raw...), nil
}

func (m *memoryObjects) Copy(context.Context, string, string) error {
	m.copies++
	return nil
}

func receivingTestService(t *testing.T) *Service {
	t.Helper()
	dsn := os.Getenv("POSTHOOT_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("POSTHOOT_TEST_DATABASE_URL required for PostgreSQL receiving integration tests")
	}
	admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	schema := "receiving_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	require.NoError(t, admin.Exec("CREATE SCHEMA "+pq.QuoteIdentifier(schema)).Error)
	t.Cleanup(func() {
		admin.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE")
		raw, _ := admin.DB()
		raw.Close()
	})
	db, err := gorm.Open(postgres.Open(dsn+" search_path="+schema), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true, Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&sending.Account{}, &sending.Domain{}))
	require.NoError(t, Migrate(db))
	return &Service{
		DB: db,
		Config: Config{
			Enabled:           true,
			Region:            "us-east-2",
			AccountID:         "123456789012",
			Bucket:            "xem-receiving-test",
			TopicARN:          receiptTopic,
			Prefix:            "incoming/",
			MaxMessageBytes:   10 * 1024 * 1024,
			MailboxQuotaBytes: 1024 * 1024 * 1024,
		},
		Now: func() time.Time { return time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC) },
	}
}

func officialReceipt(t *testing.T, recipients ...string) SESReceipt {
	t.Helper()
	// AWS SES receipt notifications document mail.source, not mail.sourceArn.
	// Keep this fixture shaped like the provider payload so a fabricated field
	// cannot become part of the trust boundary.
	raw := `{
  "notificationType":"Received",
  "mail":{"timestamp":"2026-10-01T00:00:00Z","source":"sender@example.net","messageId":"provider-message-1","destination":["inbox@example.com"]},
  "receipt":{"recipients":["inbox@example.com"],"spamVerdict":{"status":"PASS"},"virusVerdict":{"status":"PASS"},"action":{"type":"S3","topicArn":"arn:aws:sns:us-east-2:123456789012:xem-receiving","bucketName":"xem-receiving-test","objectKey":"incoming/provider-message-1"}}
}`
	var event SESReceipt
	require.NoError(t, json.Unmarshal([]byte(raw), &event))
	if recipients != nil {
		event.Receipt.Recipients = recipients
	}
	return event
}

func addMailbox(t *testing.T, service *Service, address string) Mailbox {
	t.Helper()
	team, domainID := uuid.NewString(), uuid.NewString()
	domain := sending.Domain{ID: domainID, TeamID: team, Name: strings.Split(address, "@")[1], Ownership: true, Ready: true, Provisioned: true, CreatedAt: service.Now(), UpdatedAt: service.Now()}
	require.NoError(t, service.DB.Create(&sending.Account{TeamID: team, Approved: true, CreatedAt: service.Now(), UpdatedAt: service.Now()}).Error)
	require.NoError(t, service.DB.Create(&domain).Error)
	box := Mailbox{ID: uuid.NewString(), TeamID: team, DomainID: domainID, Address: address, Active: true, SMTPConfigID: uuid.NewString(), Status: "active", UIDValidity: 1, QuotaBytes: service.Config.MailboxQuotaBytes, CreatedAt: service.Now(), UpdatedAt: service.Now()}
	require.NoError(t, service.DB.Create(&box).Error)
	return box
}

func plainMail(to string) []byte {
	return []byte("From: sender@example.net\r\nTo: " + to + "\r\nSubject: provider fixture\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nhello")
}

func TestLoadConfigReceivingDefaultsAndBounds(t *testing.T) {
	for key, value := range map[string]string{
		"MANAGED_RECEIVING_ENABLED":   "true",
		"MANAGED_SES_REGION":          "us-east-2",
		"MANAGED_AWS_ACCOUNT_ID":      "123456789012",
		"MANAGED_RECEIVING_BUCKET":    "bucket",
		"MANAGED_RECEIVING_TOPIC_ARN": receiptTopic,
		"MANAGED_RECEIVING_QUEUE_URL": "https://sqs.us-east-2.amazonaws.com/123456789012/queue",
		"MANAGED_RECEIVING_RULE_SET":  "xem-managed-receiving",
	} {
		t.Setenv(key, value)
	}
	config, err := LoadConfig()
	require.NoError(t, err)
	require.EqualValues(t, 10*1024*1024, config.MaxMessageBytes)
	require.EqualValues(t, 1024*1024*1024, config.MailboxQuotaBytes)
	require.Equal(t, "incoming/", config.Prefix)

	t.Setenv("MANAGED_RECEIVING_MAX_MESSAGE_BYTES", "10485761")
	_, err = LoadConfig()
	require.EqualError(t, err, "managed receiving size limits are invalid")
	t.Setenv("MANAGED_RECEIVING_MAX_MESSAGE_BYTES", "10485760")
	t.Setenv("MANAGED_RECEIVING_MAILBOX_QUOTA_BYTES", "1048575")
	_, err = LoadConfig()
	require.EqualError(t, err, "managed receiving size limits are invalid")
}

func TestLoadConfigDisabledDoesNotRequireAWSInfrastructure(t *testing.T) {
	t.Setenv("MANAGED_RECEIVING_ENABLED", "false")
	for _, key := range []string{"MANAGED_SES_REGION", "MANAGED_AWS_ACCOUNT_ID", "MANAGED_RECEIVING_BUCKET", "MANAGED_RECEIVING_TOPIC_ARN", "MANAGED_RECEIVING_QUEUE_URL", "MANAGED_RECEIVING_RULE_SET"} {
		t.Setenv(key, "")
	}
	config, err := LoadConfig()
	require.NoError(t, err)
	require.False(t, config.Enabled)
}

func TestOfficialSESReceiptNeedsNoMailSourceARN(t *testing.T) {
	service := receivingTestService(t)
	addMailbox(t, service, "inbox@example.com")
	objects := &memoryObjects{raw: plainMail("inbox@example.com")}
	require.NoError(t, service.ProcessReceipt(context.Background(), officialReceipt(t), objects))
	require.Equal(t, 1, objects.gets)
	require.Equal(t, 1, objects.copies)
}

func TestReceiptObjectKeyMustMatchProviderMessageID(t *testing.T) {
	service := receivingTestService(t)
	event := officialReceipt(t)
	event.Receipt.Action.ObjectKey = "incoming/a-different-message"
	objects := &memoryObjects{raw: plainMail("inbox@example.com")}
	require.Error(t, service.ProcessReceipt(context.Background(), event, objects))
	require.Zero(t, objects.gets)
	require.Zero(t, objects.copies)
}

func TestQuarantineKeepsDistinctEnvelopeRecipientSets(t *testing.T) {
	service := receivingTestService(t)
	first := officialReceipt(t, "one@example.com")
	first.Receipt.VirusVerdict.Status = "FAIL"
	second := officialReceipt(t, "two@example.com")
	second.Receipt.VirusVerdict.Status = "FAIL"
	require.NoError(t, service.ProcessReceipt(context.Background(), first, &memoryObjects{}))
	require.NoError(t, service.ProcessReceipt(context.Background(), second, &memoryObjects{}))
	require.NoError(t, service.ProcessReceipt(context.Background(), first, &memoryObjects{}))
	var rows []Rejection
	require.NoError(t, service.DB.Order("recipient_set_hash").Find(&rows).Error)
	require.Len(t, rows, 2)
	sets := map[string]bool{}
	for _, row := range rows {
		var recipients []string
		require.NoError(t, json.Unmarshal(row.RecipientsJSON, &recipients))
		require.Len(t, recipients, 1)
		sets[recipients[0]] = true
	}
	require.True(t, sets["one@example.com"])
	require.True(t, sets["two@example.com"])
}

func TestEnvelopeRecipientsOverrideConflictingMIMEHeaders(t *testing.T) {
	service := receivingTestService(t)
	owned := addMailbox(t, service, "owned@example.com")
	other := addMailbox(t, service, "header-only@example.org")
	objects := &memoryObjects{raw: plainMail("header-only@example.org")}
	require.NoError(t, service.ProcessReceipt(context.Background(), officialReceipt(t, owned.Address), objects))
	var ownedCount, otherCount int64
	require.NoError(t, service.DB.Model(&Message{}).Where("mailbox_id = ?", owned.ID).Count(&ownedCount).Error)
	require.NoError(t, service.DB.Model(&Message{}).Where("mailbox_id = ?", other.ID).Count(&otherCount).Error)
	require.EqualValues(t, 1, ownedCount)
	require.Zero(t, otherCount)
	var stored Message
	require.NoError(t, service.DB.Where("mailbox_id = ?", owned.ID).First(&stored).Error)
	require.True(t, stored.SentAt.Equal(time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)))
}

func TestVerdictsFailClosedAndSpamPassesToSpamFolder(t *testing.T) {
	for _, status := range []string{"", "FAIL", "GRAY", "PROCESSING_FAILED"} {
		t.Run("virus_"+status, func(t *testing.T) {
			service := receivingTestService(t)
			event := officialReceipt(t)
			event.Receipt.VirusVerdict.Status = status
			objects := &memoryObjects{raw: plainMail("inbox@example.com")}
			require.NoError(t, service.ProcessReceipt(context.Background(), event, objects))
			require.Zero(t, objects.gets)
			var rejection Rejection
			require.NoError(t, service.DB.First(&rejection).Error)
			require.Equal(t, "quarantined", rejection.Reason)
		})
	}

	service := receivingTestService(t)
	box := addMailbox(t, service, "inbox@example.com")
	event := officialReceipt(t)
	event.Receipt.SpamVerdict.Status = "FAIL"
	require.NoError(t, service.ProcessReceipt(context.Background(), event, &memoryObjects{raw: plainMail(box.Address)}))
	var message Message
	require.NoError(t, service.DB.First(&message).Error)
	require.Equal(t, "Spam", message.Folder)
}

func TestDuplicateCrossWorkspaceReceiptIsIdempotentAndKeepsAttachmentMetadata(t *testing.T) {
	service := receivingTestService(t)
	first := addMailbox(t, service, "one@example.com")
	second := addMailbox(t, service, "two@example.org")
	raw := []byte("From: sender@example.net\r\nTo: one@example.com, two@example.org\r\nSubject: shared\r\nMIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=x\r\n\r\n--x\r\nContent-Type: text/plain\r\n\r\nhello\r\n--x\r\nContent-Type: application/octet-stream\r\nContent-Disposition: attachment; filename=a.bin\r\nContent-Transfer-Encoding: base64\r\n\r\nYWJj\r\n--x--\r\n")
	event := officialReceipt(t, first.Address, second.Address)
	objects := &memoryObjects{raw: raw}
	require.NoError(t, service.ProcessReceipt(context.Background(), event, objects))
	require.NoError(t, service.ProcessReceipt(context.Background(), event, objects))
	var messages []Message
	require.NoError(t, service.DB.Order("mailbox_id").Find(&messages).Error)
	require.Len(t, messages, 2)
	for _, message := range messages {
		var attachments []map[string]any
		require.NoError(t, json.Unmarshal(message.AttachmentsJSON, &attachments))
		require.Len(t, attachments, 1)
		require.EqualValues(t, 3, attachments[0]["Size"])
	}
	var blob Blob
	require.NoError(t, service.DB.First(&blob).Error)
	require.EqualValues(t, 2, blob.References)
}

func TestExactQuotaDuplicateRemainsIdempotent(t *testing.T) {
	service := receivingTestService(t)
	box := addMailbox(t, service, "inbox@example.com")
	raw := plainMail(box.Address)
	require.NoError(t, service.DB.Model(&box).Update("quota_bytes", int64(len(raw))).Error)
	objects := &memoryObjects{raw: raw}
	event := officialReceipt(t, box.Address)

	require.NoError(t, service.ProcessReceipt(context.Background(), event, objects))
	require.NoError(t, service.ProcessReceipt(context.Background(), event, objects))

	var messages, blobs int64
	require.NoError(t, service.DB.Model(&Message{}).Where("mailbox_id = ?", box.ID).Count(&messages).Error)
	require.NoError(t, service.DB.Model(&Blob{}).Where("provider_id = ?", event.Mail.MessageID).Count(&blobs).Error)
	require.EqualValues(t, 1, messages)
	require.EqualValues(t, 1, blobs)
	require.NoError(t, service.DB.First(&box, "id = ?", box.ID).Error)
	require.EqualValues(t, len(raw), box.UsageBytes)
	require.Equal(t, "active", box.Status)
}

func TestMultiRecipientIndexesEligibleMailboxWhenOthersAreFullOrPaused(t *testing.T) {
	service := receivingTestService(t)
	service.Rules = permissiveRules{ruleSet: service.Config.RuleSet}
	healthy := addMailbox(t, service, "healthy@example.com")
	full := addMailbox(t, service, "full@example.org")
	paused := addMailbox(t, service, "paused@example.net")
	raw := plainMail("healthy@example.com, full@example.org, paused@example.net")
	require.NoError(t, service.DB.Model(&full).Update("quota_bytes", int64(len(raw)-1)).Error)
	require.NoError(t, service.DB.Model(&paused).Updates(map[string]any{"active": false, "status": "inactive"}).Error)
	event := officialReceipt(t, healthy.Address, full.Address, paused.Address)
	objects := &memoryObjects{raw: raw}

	require.ErrorIs(t, service.ProcessReceipt(context.Background(), event, objects), ErrMailboxQuota)

	var healthyCount, fullCount, pausedCount int64
	require.NoError(t, service.DB.Model(&Message{}).Where("mailbox_id = ?", healthy.ID).Count(&healthyCount).Error)
	require.NoError(t, service.DB.Model(&Message{}).Where("mailbox_id = ?", full.ID).Count(&fullCount).Error)
	require.NoError(t, service.DB.Model(&Message{}).Where("mailbox_id = ?", paused.ID).Count(&pausedCount).Error)
	require.EqualValues(t, 1, healthyCount)
	require.Zero(t, fullCount)
	require.Zero(t, pausedCount)
	require.NoError(t, service.DB.First(&full, "id = ?", full.ID).Error)
	require.Equal(t, "storage_full", full.Status)

	// SQS retry must leave the already indexed recipient idempotent while the
	// unavailable recipients continue to prevent acknowledgement.
	require.ErrorIs(t, service.ProcessReceipt(context.Background(), event, objects), ErrMailboxQuota)
	require.NoError(t, service.DB.Model(&Message{}).Where("mailbox_id = ?", healthy.ID).Count(&healthyCount).Error)
	require.EqualValues(t, 1, healthyCount)
}

type permissiveRules struct{ ruleSet string }

func (p permissiveRules) ActiveRuleSet(context.Context) (string, error) { return p.ruleSet, nil }
func (permissiveRules) PutRule(context.Context, string, []string) error { return nil }
func (permissiveRules) DeleteRule(context.Context, string) error        { return nil }
func (permissiveRules) RuleReady(context.Context, string, []string) (bool, error) {
	return false, nil
}

func TestObjectStoreFailureRemainsRetryable(t *testing.T) {
	service := receivingTestService(t)
	addMailbox(t, service, "inbox@example.com")
	retry := errors.New("temporary object failure")
	store := failingObjects{err: retry}
	require.ErrorIs(t, service.ProcessReceipt(context.Background(), officialReceipt(t), store), retry)
}

type failingObjects struct{ err error }

func (f failingObjects) Get(context.Context, string) ([]byte, error) { return nil, f.err }
func (f failingObjects) Copy(context.Context, string, string) error  { return f.err }
