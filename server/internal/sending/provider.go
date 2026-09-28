package sending

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/aws/smithy-go"
)

type Identity struct {
	Verified               bool
	Status, DKIM, MAILFROM string
	Tokens                 []string
}
type Provider interface {
	Provision(context.Context, string, string) error
	Identity(context.Context, string) (Identity, error)
	Send(context.Context, string, string, string, []string, []byte, string) (string, error)
}
type SES struct {
	client *sesv2.Client
	config Config
}

func NewSES(ctx context.Context, c Config) (*SES, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(c.Region))
	if err != nil {
		return nil, err
	}
	// SES has no SendEmail idempotency token: automatic SDK retries can duplicate mail.
	return &SES{client: sesv2.NewFromConfig(cfg, func(o *sesv2.Options) { o.RetryMaxAttempts = 1 }), config: c}, nil
}
func tenant(team string) string { return "xem-" + team }
func already(err error) error {
	var e smithy.APIError
	if errors.As(err, &e) && e.ErrorCode() == "AlreadyExistsException" {
		return nil
	}
	return err
}
func (s *SES) Provision(ctx context.Context, team, domain string) error {
	name := tenant(team)
	_, err := s.client.CreateTenant(ctx, &sesv2.CreateTenantInput{TenantName: aws.String(name), SuppressionAttributes: &types.TenantSuppressionAttributes{SuppressionScope: types.SuppressionListScopeTenant, SuppressedReasons: []types.SuppressionListReason{types.SuppressionListReasonBounce, types.SuppressionListReasonComplaint}}})
	if err = already(err); err != nil {
		return err
	}
	_, err = s.client.CreateConfigurationSet(ctx, &sesv2.CreateConfigurationSetInput{ConfigurationSetName: aws.String(name), DeliveryOptions: &types.DeliveryOptions{TlsPolicy: types.TlsPolicyRequire}})
	if err = already(err); err != nil {
		return err
	}
	dest := &types.EventDestinationDefinition{Enabled: true, MatchingEventTypes: []types.EventType{types.EventTypeSend, types.EventTypeDelivery, types.EventTypeBounce, types.EventTypeComplaint, types.EventTypeReject, types.EventTypeDeliveryDelay}, SnsDestination: &types.SnsDestination{TopicArn: aws.String(s.config.TopicARN)}}
	_, err = s.client.CreateConfigurationSetEventDestination(ctx, &sesv2.CreateConfigurationSetEventDestinationInput{ConfigurationSetName: aws.String(name), EventDestinationName: aws.String("xem-events"), EventDestination: dest})
	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) && ae.ErrorCode() == "AlreadyExistsException" {
			_, err = s.client.UpdateConfigurationSetEventDestination(ctx, &sesv2.UpdateConfigurationSetEventDestinationInput{ConfigurationSetName: aws.String(name), EventDestinationName: aws.String("xem-events"), EventDestination: dest})
		}
		if err != nil {
			return err
		}
	}
	_, err = s.client.CreateEmailIdentity(ctx, &sesv2.CreateEmailIdentityInput{EmailIdentity: aws.String(domain), DkimSigningAttributes: &types.DkimSigningAttributes{NextSigningKeyLength: types.DkimSigningKeyLengthRsa2048Bit}})
	if err = already(err); err != nil {
		return err
	}
	_, err = s.client.PutEmailIdentityMailFromAttributes(ctx, &sesv2.PutEmailIdentityMailFromAttributesInput{EmailIdentity: aws.String(domain), MailFromDomain: aws.String("bounce." + domain), BehaviorOnMxFailure: types.BehaviorOnMxFailureRejectMessage})
	if err != nil {
		return err
	}
	for _, resource := range []string{"identity/" + domain, "configuration-set/" + name} {
		_, err = s.client.CreateTenantResourceAssociation(ctx, &sesv2.CreateTenantResourceAssociationInput{TenantName: aws.String(name), ResourceArn: aws.String(fmt.Sprintf("arn:aws:ses:%s:%s:%s", s.config.Region, s.config.AccountID, resource))})
		if err = already(err); err != nil {
			return err
		}
	}
	return nil
}
func (s *SES) Identity(ctx context.Context, domain string) (Identity, error) {
	out, err := s.client.GetEmailIdentity(ctx, &sesv2.GetEmailIdentityInput{EmailIdentity: aws.String(domain)})
	if err != nil {
		return Identity{}, err
	}
	i := Identity{Verified: out.VerifiedForSendingStatus, Status: string(out.VerificationStatus)}
	if out.DkimAttributes != nil {
		i.DKIM = string(out.DkimAttributes.Status)
		i.Tokens = out.DkimAttributes.Tokens
	}
	if out.MailFromAttributes != nil {
		i.MAILFROM = string(out.MailFromAttributes.MailFromDomainStatus)
	}
	return i, nil
}
func (s *SES) Send(ctx context.Context, team, domain, from string, to []string, raw []byte, id string) (string, error) {
	out, err := s.client.SendEmail(ctx, &sesv2.SendEmailInput{TenantName: aws.String(tenant(team)), ConfigurationSetName: aws.String(tenant(team)), FromEmailAddress: aws.String(from), Destination: &types.Destination{ToAddresses: to}, Content: &types.EmailContent{Raw: &types.RawMessage{Data: raw}}, EmailTags: []types.MessageTag{{Name: aws.String("xem_message"), Value: aws.String(id)}}})
	if err != nil {
		return "", err
	}
	return aws.ToString(out.MessageId), nil
}
