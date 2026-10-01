package receiving

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	sestypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"kori/internal/sending"
)

type sesAPI interface {
	DescribeActiveReceiptRuleSet(context.Context, *ses.DescribeActiveReceiptRuleSetInput, ...func(*ses.Options)) (*ses.DescribeActiveReceiptRuleSetOutput, error)
	DescribeReceiptRule(context.Context, *ses.DescribeReceiptRuleInput, ...func(*ses.Options)) (*ses.DescribeReceiptRuleOutput, error)
	CreateReceiptRule(context.Context, *ses.CreateReceiptRuleInput, ...func(*ses.Options)) (*ses.CreateReceiptRuleOutput, error)
	UpdateReceiptRule(context.Context, *ses.UpdateReceiptRuleInput, ...func(*ses.Options)) (*ses.UpdateReceiptRuleOutput, error)
	DeleteReceiptRule(context.Context, *ses.DeleteReceiptRuleInput, ...func(*ses.Options)) (*ses.DeleteReceiptRuleOutput, error)
}
type AWSRules struct {
	client sesAPI
	config Config
}

var ownedReceiptRulePattern = regexp.MustCompile(`^xem-inbox-[a-f0-9]{32}$`)

func NewAWSRules(client sesAPI, c Config) *AWSRules { return &AWSRules{client: client, config: c} }
func (r *AWSRules) desired(name string, recipients []string) *sestypes.ReceiptRule {
	sort.Strings(recipients)
	return &sestypes.ReceiptRule{Name: aws.String(name), Enabled: true, ScanEnabled: true, TlsPolicy: sestypes.TlsPolicyRequire, Recipients: recipients, Actions: []sestypes.ReceiptAction{{S3Action: &sestypes.S3Action{BucketName: aws.String(r.config.Bucket), ObjectKeyPrefix: aws.String(r.config.Prefix), TopicArn: aws.String(r.config.TopicARN)}}}}
}
func (r *AWSRules) ownsRule(rule *sestypes.ReceiptRule, name string) bool {
	if rule == nil || !ownedReceiptRulePattern.MatchString(name) || aws.ToString(rule.Name) != name || len(rule.Actions) != 1 || rule.Actions[0].S3Action == nil {
		return false
	}
	action := rule.Actions[0].S3Action
	return aws.ToString(action.BucketName) == r.config.Bucket && aws.ToString(action.ObjectKeyPrefix) == r.config.Prefix && aws.ToString(action.TopicArn) == r.config.TopicARN
}
func (r *AWSRules) ready(rule, want *sestypes.ReceiptRule) bool {
	if !r.ownsRule(rule, aws.ToString(want.Name)) {
		return false
	}
	gotRecipients := append([]string(nil), rule.Recipients...)
	sort.Strings(gotRecipients)
	return rule.Enabled && rule.ScanEnabled && rule.TlsPolicy == want.TlsPolicy && fmt.Sprint(gotRecipients) == fmt.Sprint(want.Recipients)
}
func (r *AWSRules) ActiveRuleSet(ctx context.Context) (string, error) {
	out, err := r.client.DescribeActiveReceiptRuleSet(ctx, &ses.DescribeActiveReceiptRuleSetInput{})
	if err != nil {
		return "", err
	}
	if out.Metadata == nil {
		return "", nil
	}
	return aws.ToString(out.Metadata.Name), nil
}
func awsNotFound(err error) bool {
	var api interface{ ErrorCode() string }
	return errors.As(err, &api) && (api.ErrorCode() == "RuleDoesNotExist" || api.ErrorCode() == "RuleSetDoesNotExist")
}
func (r *AWSRules) PutRule(ctx context.Context, name string, recipients []string) error {
	if !ownedReceiptRulePattern.MatchString(name) || len(recipients) == 0 || len(recipients) > 100 {
		return errors.New("invalid managed receipt rule")
	}
	rule := r.desired(name, append([]string(nil), recipients...))
	out, err := r.client.DescribeReceiptRule(ctx, &ses.DescribeReceiptRuleInput{RuleName: aws.String(name), RuleSetName: aws.String(r.config.RuleSet)})
	if awsNotFound(err) {
		_, err = r.client.CreateReceiptRule(ctx, &ses.CreateReceiptRuleInput{RuleSetName: aws.String(r.config.RuleSet), Rule: rule})
		return err
	}
	if err != nil {
		return err
	}
	if !r.ownsRule(out.Rule, name) {
		return errors.New("refusing to overwrite an unowned receipt rule")
	}
	if r.ready(out.Rule, rule) {
		return nil
	}
	_, err = r.client.UpdateReceiptRule(ctx, &ses.UpdateReceiptRuleInput{RuleSetName: aws.String(r.config.RuleSet), Rule: rule})
	return err
}
func (r *AWSRules) DeleteRule(ctx context.Context, name string) error {
	if !ownedReceiptRulePattern.MatchString(name) {
		return errors.New("invalid managed receipt rule")
	}
	out, err := r.client.DescribeReceiptRule(ctx, &ses.DescribeReceiptRuleInput{RuleName: aws.String(name), RuleSetName: aws.String(r.config.RuleSet)})
	if awsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !r.ownsRule(out.Rule, name) {
		return errors.New("refusing to delete an unowned receipt rule")
	}
	_, err = r.client.DeleteReceiptRule(ctx, &ses.DeleteReceiptRuleInput{RuleName: aws.String(name), RuleSetName: aws.String(r.config.RuleSet)})
	return err
}
func (r *AWSRules) RuleReady(ctx context.Context, name string, recipients []string) (bool, error) {
	if !ownedReceiptRulePattern.MatchString(name) || len(recipients) == 0 || len(recipients) > 100 {
		return false, errors.New("invalid managed receipt rule")
	}
	out, err := r.client.DescribeReceiptRule(ctx, &ses.DescribeReceiptRuleInput{RuleName: aws.String(name), RuleSetName: aws.String(r.config.RuleSet)})
	if awsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if out.Rule == nil {
		return false, nil
	}
	want := r.desired(name, append([]string(nil), recipients...))
	return r.ready(out.Rule, want), nil
}

type s3API interface {
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	CopyObject(context.Context, *s3.CopyObjectInput, ...func(*s3.Options)) (*s3.CopyObjectOutput, error)
}
type AWSStore struct {
	client s3API
	bucket string
	max    int64
}

func NewAWSStore(client s3API, c Config) *AWSStore {
	return &AWSStore{client: client, bucket: c.Bucket, max: c.MaxMessageBytes}
}
func (s *AWSStore) Get(ctx context.Context, key string) ([]byte, error) {
	if !validObjectKey(key, "incoming/") && !validObjectKey(key, "mail/") {
		return nil, errors.New("invalid managed message object key")
	}
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()
	data, err := io.ReadAll(io.LimitReader(out.Body, s.max+1))
	if err != nil {
		return nil, err
	}
	return data, nil
}
func (s *AWSStore) Copy(ctx context.Context, from, to string) error {
	if !validObjectKey(from, "incoming/") || !validObjectKey(to, "mail/") {
		return errors.New("invalid managed message object key")
	}
	source := url.PathEscape(s.bucket + "/" + from)
	_, err := s.client.CopyObject(ctx, &s3.CopyObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(to), CopySource: aws.String(source)})
	return err
}

func validObjectKey(key, prefix string) bool {
	remainder := strings.TrimPrefix(key, prefix)
	return remainder != key && providerIDPattern.MatchString(remainder)
}

type sqsAPI interface {
	ReceiveMessage(context.Context, *sqs.ReceiveMessageInput, ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(context.Context, *sqs.DeleteMessageInput, ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
	GetQueueAttributes(context.Context, *sqs.GetQueueAttributesInput, ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error)
}
type Worker struct {
	Service  *Service
	Queue    sqsAPI
	Store    ObjectStore
	QueueURL string
	process  func(context.Context, sending.Notification, ObjectStore) error
}

func (w *Worker) processEnvelope(ctx context.Context, envelope sending.Notification) error {
	if w.process != nil {
		return w.process(ctx, envelope, w.Store)
	}
	return w.Service.ProcessEnvelope(ctx, envelope, w.Store)
}

func (w *Worker) Run(ctx context.Context) {
	for ctx.Err() == nil {
		out, err := w.Queue.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: aws.String(w.QueueURL), MaxNumberOfMessages: 1, WaitTimeSeconds: 20, VisibilityTimeout: 300})
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Print("managed receiving queue poll failed; retrying")
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}
			continue
		}
		for _, m := range out.Messages {
			if len(aws.ToString(m.Body)) > 256*1024 {
				log.Print("managed receiving queue message exceeded the SNS envelope limit; leaving it for redrive")
				continue
			}
			var envelope sending.Notification
			if json.Unmarshal([]byte(aws.ToString(m.Body)), &envelope) != nil {
				log.Print("managed receiving queue message was malformed; leaving it for redrive")
				continue
			}
			messageCtx, cancel := context.WithTimeout(ctx, 240*time.Second)
			err := w.processEnvelope(messageCtx, envelope)
			cancel()
			if err != nil {
				log.Print("managed receiving queue message processing failed; leaving it for retry")
				continue
			}
			deleteCtx, deleteCancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
			_, err = w.Queue.DeleteMessage(deleteCtx, &sqs.DeleteMessageInput{QueueUrl: aws.String(w.QueueURL), ReceiptHandle: m.ReceiptHandle})
			deleteCancel()
			if err != nil {
				log.Print("managed receiving queue message commit succeeded but acknowledgement failed; duplicate delivery is expected")
			}
		}
	}
}

func verifyQueue(ctx context.Context, client sqsAPI, c Config) error {
	queueURL, err := url.Parse(c.QueueURL)
	if err != nil || queueURL.Scheme != "https" || queueURL.User != nil || queueURL.RawQuery != "" || queueURL.Fragment != "" || queueURL.Hostname() != "sqs."+c.Region+".amazonaws.com" {
		return errors.New("managed receiving queue URL does not match the configured AWS region")
	}
	parts := strings.Split(strings.Trim(queueURL.Path, "/"), "/")
	if len(parts) != 2 || parts[0] != c.AccountID || parts[1] == "" {
		return errors.New("managed receiving queue URL does not match the configured AWS account")
	}
	out, err := client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{QueueUrl: aws.String(c.QueueURL), AttributeNames: []sqstypes.QueueAttributeName{sqstypes.QueueAttributeNameQueueArn}})
	if err != nil {
		return err
	}
	expected := "arn:aws:sqs:" + c.Region + ":" + c.AccountID + ":" + parts[1]
	if out.Attributes[string(sqstypes.QueueAttributeNameQueueArn)] != expected {
		return errors.New("managed receiving queue ARN does not match its configured URL")
	}
	return nil
}

func NewAWSRuntime(ctx context.Context, c Config) (RuleProvider, ObjectStore, *Worker, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(c.Region))
	if err != nil {
		return nil, nil, nil, err
	}
	rules := NewAWSRules(ses.NewFromConfig(cfg), c)
	store := NewAWSStore(s3.NewFromConfig(cfg), c)
	queue := sqs.NewFromConfig(cfg)
	if err := verifyQueue(ctx, queue, c); err != nil {
		return nil, nil, nil, err
	}
	worker := &Worker{Queue: queue, Store: store, QueueURL: c.QueueURL}
	return rules, store, worker, nil
}
