package receiving

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	sestypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/stretchr/testify/require"
	"kori/internal/sending"
)

type fakeSES struct {
	describeErr error
	described   *sestypes.ReceiptRule
	created     *ses.CreateReceiptRuleInput
	updated     *ses.UpdateReceiptRuleInput
	deleted     *ses.DeleteReceiptRuleInput
}

func (f *fakeSES) DescribeActiveReceiptRuleSet(context.Context, *ses.DescribeActiveReceiptRuleSetInput, ...func(*ses.Options)) (*ses.DescribeActiveReceiptRuleSetOutput, error) {
	return &ses.DescribeActiveReceiptRuleSetOutput{Metadata: &sestypes.ReceiptRuleSetMetadata{Name: aws.String("xem-managed-receiving")}}, nil
}
func (f *fakeSES) DescribeReceiptRule(context.Context, *ses.DescribeReceiptRuleInput, ...func(*ses.Options)) (*ses.DescribeReceiptRuleOutput, error) {
	return &ses.DescribeReceiptRuleOutput{Rule: f.described}, f.describeErr
}
func (f *fakeSES) CreateReceiptRule(_ context.Context, input *ses.CreateReceiptRuleInput, _ ...func(*ses.Options)) (*ses.CreateReceiptRuleOutput, error) {
	f.created = input
	return &ses.CreateReceiptRuleOutput{}, nil
}
func (f *fakeSES) UpdateReceiptRule(_ context.Context, input *ses.UpdateReceiptRuleInput, _ ...func(*ses.Options)) (*ses.UpdateReceiptRuleOutput, error) {
	f.updated = input
	return &ses.UpdateReceiptRuleOutput{}, nil
}
func (f *fakeSES) DeleteReceiptRule(_ context.Context, input *ses.DeleteReceiptRuleInput, _ ...func(*ses.Options)) (*ses.DeleteReceiptRuleOutput, error) {
	f.deleted = input
	return &ses.DeleteReceiptRuleOutput{}, nil
}

func TestAWSRuleIsExactScannedTLSRequiredAndHasNoStopAction(t *testing.T) {
	rules := NewAWSRules(&fakeSES{}, Config{Bucket: "private-bucket", Prefix: "incoming/", TopicARN: receiptTopic})
	desired := rules.desired("xem-inbox-0123456789abcdef0123456789abcdef", []string{"two@example.com", "one@example.com"})
	require.Equal(t, []string{"one@example.com", "two@example.com"}, desired.Recipients)
	require.True(t, desired.Enabled)
	require.True(t, desired.ScanEnabled)
	require.Equal(t, sestypes.TlsPolicyRequire, desired.TlsPolicy)
	require.Len(t, desired.Actions, 1)
	require.NotNil(t, desired.Actions[0].S3Action)
	require.Nil(t, desired.Actions[0].StopAction)
	require.Equal(t, "private-bucket", aws.ToString(desired.Actions[0].S3Action.BucketName))
	require.Equal(t, "incoming/", aws.ToString(desired.Actions[0].S3Action.ObjectKeyPrefix))
	require.Equal(t, receiptTopic, aws.ToString(desired.Actions[0].S3Action.TopicArn))
}

func TestAWSRuleUpdateTouchesOnlyConfiguredOwnedRule(t *testing.T) {
	name := "xem-inbox-0123456789abcdef0123456789abcdef"
	fake := &fakeSES{described: &sestypes.ReceiptRule{Name: aws.String(name), Actions: []sestypes.ReceiptAction{{S3Action: &sestypes.S3Action{BucketName: aws.String("private-bucket"), ObjectKeyPrefix: aws.String("incoming/"), TopicArn: aws.String(receiptTopic)}}}}}
	rules := NewAWSRules(fake, Config{Bucket: "private-bucket", Prefix: "incoming/", TopicARN: receiptTopic, RuleSet: "xem-managed-receiving"})
	require.NoError(t, rules.PutRule(context.Background(), name, []string{"one@example.com"}))
	require.Nil(t, fake.created)
	require.NotNil(t, fake.updated)
	require.Equal(t, "xem-managed-receiving", aws.ToString(fake.updated.RuleSetName))
	require.Equal(t, name, aws.ToString(fake.updated.Rule.Name))
	require.Equal(t, []string{"one@example.com"}, fake.updated.Rule.Recipients)
}

func TestAWSRuleNoopsWhenReadyAndRejectsForeignProvenance(t *testing.T) {
	name := "xem-inbox-0123456789abcdef0123456789abcdef"
	rules := NewAWSRules(&fakeSES{}, Config{Bucket: "private-bucket", Prefix: "incoming/", TopicARN: receiptTopic, RuleSet: "xem-managed-receiving"})
	ready := rules.desired(name, []string{"one@example.com"})
	fake := &fakeSES{described: ready}
	rules.client = fake
	require.NoError(t, rules.PutRule(context.Background(), name, []string{"one@example.com"}))
	require.Nil(t, fake.updated)

	fake.described = &sestypes.ReceiptRule{Name: aws.String(name), Actions: []sestypes.ReceiptAction{{S3Action: &sestypes.S3Action{BucketName: aws.String("unrelated-bucket")}}}}
	require.EqualError(t, rules.PutRule(context.Background(), name, []string{"one@example.com"}), "refusing to overwrite an unowned receipt rule")
	require.EqualError(t, rules.DeleteRule(context.Background(), name), "refusing to delete an unowned receipt rule")
}

type fakeS3 struct {
	body      string
	copyInput *s3.CopyObjectInput
}

func (f *fakeS3) GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return &s3.GetObjectOutput{Body: io.NopCloser(strings.NewReader(f.body))}, nil
}
func (f *fakeS3) CopyObject(_ context.Context, input *s3.CopyObjectInput, _ ...func(*s3.Options)) (*s3.CopyObjectOutput, error) {
	f.copyInput = input
	return &s3.CopyObjectOutput{}, nil
}

func TestAWSStoreBoundsReadsAndPromotesOnlyToRequestedDurableKey(t *testing.T) {
	fake := &fakeS3{body: "123456"}
	store := &AWSStore{client: fake, bucket: "private-bucket", max: 5}
	data, err := store.Get(context.Background(), "incoming/message")
	require.NoError(t, err)
	require.Equal(t, []byte("123456"), data, "the max+1 sentinel lets ingestion ledger and acknowledge oversized mail")

	fake.body = "12345"
	data, err = store.Get(context.Background(), "incoming/message")
	require.NoError(t, err)
	require.Equal(t, []byte("12345"), data)
	require.NoError(t, store.Copy(context.Background(), "incoming/message", "mail/message"))
	require.Equal(t, "private-bucket", aws.ToString(fake.copyInput.Bucket))
	require.Equal(t, "mail/message", aws.ToString(fake.copyInput.Key))
	require.NotEmpty(t, aws.ToString(fake.copyInput.CopySource))
}

type fakeQueue struct {
	cancel      context.CancelFunc
	receive     *sqs.ReceiveMessageInput
	deleteCalls int
	body        string
	deleteErr   error
	queueARN    string
}

func (f *fakeQueue) ReceiveMessage(ctx context.Context, input *sqs.ReceiveMessageInput, _ ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
	if f.receive != nil {
		return nil, context.Canceled
	}
	f.receive = input
	f.cancel()
	return &sqs.ReceiveMessageOutput{Messages: []sqstypes.Message{{Body: aws.String(f.body), ReceiptHandle: aws.String("receipt")}}}, nil
}
func (f *fakeQueue) DeleteMessage(context.Context, *sqs.DeleteMessageInput, ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	f.deleteCalls++
	return &sqs.DeleteMessageOutput{}, f.deleteErr
}
func (f *fakeQueue) GetQueueAttributes(context.Context, *sqs.GetQueueAttributesInput, ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error) {
	return &sqs.GetQueueAttributesOutput{Attributes: map[string]string{string(sqstypes.QueueAttributeNameQueueArn): f.queueARN}}, nil
}

func TestWorkerUsesSafeVisibilityAndDoesNotAckMalformedMessages(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	queue := &fakeQueue{cancel: cancel, body: "not-json"}
	worker := Worker{Service: &Service{Config: Config{Enabled: false}}, Queue: queue, Store: &memoryObjects{}, QueueURL: "queue-url"}
	worker.Run(ctx)
	require.NotNil(t, queue.receive)
	require.EqualValues(t, 1, queue.receive.MaxNumberOfMessages)
	require.GreaterOrEqual(t, queue.receive.VisibilityTimeout, int32(300))
	require.EqualValues(t, 20, queue.receive.WaitTimeSeconds)
	require.Zero(t, queue.deleteCalls)
}

func TestWorkerAcknowledgesOnlySuccessfulProcessing(t *testing.T) {
	for _, tc := range []struct {
		processingError error
		deleteError     error
		wantDeleteCalls int
		wantLog         string
	}{{nil, nil, 1, ""}, {context.DeadlineExceeded, nil, 0, "processing failed"}, {nil, errors.New("delete unavailable"), 1, "acknowledgement failed"}} {
		ctx, cancel := context.WithCancel(context.Background())
		queue := &fakeQueue{cancel: cancel, body: `{}`}
		queue.deleteErr = tc.deleteError
		var logs bytes.Buffer
		previous := log.Writer()
		log.SetOutput(&logs)
		worker := Worker{Queue: queue, Store: &memoryObjects{}, QueueURL: "queue-url", process: func(context.Context, sending.Notification, ObjectStore) error { return tc.processingError }}
		worker.Run(ctx)
		log.SetOutput(previous)
		require.Equal(t, tc.wantDeleteCalls, queue.deleteCalls)
		if tc.wantLog != "" {
			require.Contains(t, logs.String(), tc.wantLog)
		}
	}
}

func TestQueueIdentityMustMatchConfiguredRegionAndAccount(t *testing.T) {
	config := Config{Region: "us-east-2", AccountID: "123456789012", QueueURL: "https://sqs.us-east-2.amazonaws.com/123456789012/xem-receiving"}
	queue := &fakeQueue{queueARN: "arn:aws:sqs:us-east-2:123456789012:xem-receiving"}
	require.NoError(t, verifyQueue(context.Background(), queue, config))

	queue.queueARN = "arn:aws:sqs:us-east-2:999999999999:xem-receiving"
	require.EqualError(t, verifyQueue(context.Background(), queue, config), "managed receiving queue ARN does not match its configured URL")
	config.QueueURL = "https://sqs.us-east-1.amazonaws.com/123456789012/xem-receiving"
	require.EqualError(t, verifyQueue(context.Background(), queue, config), "managed receiving queue URL does not match the configured AWS region")
}
