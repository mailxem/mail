package sending

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/smithy-go"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type checkProvider struct {
	Provider
	provisionErr, identityErr error
}

func (p checkProvider) Provision(ctx context.Context, team, domain string) error {
	if p.provisionErr != nil {
		return p.provisionErr
	}
	return p.Provider.Provision(ctx, team, domain)
}

func (p checkProvider) Identity(ctx context.Context, domain string) (Identity, error) {
	if p.identityErr != nil {
		return Identity{}, p.identityErr
	}
	return p.Provider.Identity(ctx, domain)
}

type checkDNS struct {
	Resolver
	name string
	err  error
}

func (r checkDNS) LookupTXT(ctx context.Context, name string) ([]string, error) {
	if name == r.name {
		return nil, r.err
	}
	return r.Resolver.LookupTXT(ctx, name)
}

func TestHTTPDomainCheckFailures(t *testing.T) {
	const sensitive = "private provider detail and ownership-token"
	providerErr := &smithy.OperationError{ServiceID: "SESv2", OperationName: "GetEmailIdentity", Err: &smithy.GenericAPIError{Code: "AccessDeniedException", Message: sensitive, Fault: smithy.FaultClient}}
	for _, tc := range []struct {
		name, stage, message string
		status               int
		configure            func(*Service, Domain)
	}{
		{"provider identity", "identity", "Unable to retrieve your domain verification status", 502, func(s *Service, d Domain) {
			s.Provider = checkProvider{Provider: s.Provider, identityErr: providerErr}
		}},
		{"provider setup", "provision", "Unable to prepare your sending domain", 502, func(s *Service, d Domain) {
			require.NoError(t, s.DB.Model(&Domain{}).Where("id = ?", d.ID).Update("provisioned", false).Error)
			s.Provider = checkProvider{Provider: s.Provider, provisionErr: errors.New(sensitive)}
		}},
		{"ownership DNS unavailable", "ownership_dns", "DNS lookup is temporarily unavailable", 503, func(s *Service, d Domain) {
			s.DNS = checkDNS{s.DNS, "_xem." + d.Name, &net.DNSError{Err: sensitive, IsTemporary: true}}
		}},
		{"DMARC DNS unavailable", "dmarc_dns", "DNS lookup is temporarily unavailable", 503, func(s *Service, d Domain) {
			s.DNS = checkDNS{s.DNS, "_dmarc." + d.Name, &net.DNSError{Err: sensitive, IsTemporary: true}}
		}},
		{"provider timeout", "identity", "The domain check timed out", 504, func(s *Service, d Domain) {
			s.Provider = checkProvider{Provider: s.Provider, identityErr: context.DeadlineExceeded}
		}},
		{"DNS timeout", "ownership_dns", "The domain check timed out", 504, func(s *Service, d Domain) {
			s.DNS = checkDNS{s.DNS, "_xem." + d.Name, &net.DNSError{Err: sensitive, IsTimeout: true}}
		}},
		{"database save", "save_domain", "Unable to complete domain verification", 500, func(s *Service, d Domain) {
			require.NoError(t, s.DB.Callback().Create().Before("gorm:create").Register("fail_check_audit", func(tx *gorm.DB) {
				if tx.Statement.Table == "managed_audits" {
					tx.AddError(errors.New(sensitive))
				}
			}))
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, d, _, token, secret := authenticatedSending(t)
			require.NoError(t, s.DB.Model(&Account{}).Where("team_id = ?", d.TeamID).Update("approved", false).Error)
			tc.configure(s, d)
			var logs bytes.Buffer
			oldOutput := log.Writer()
			log.SetOutput(&logs)
			t.Cleanup(func() { log.SetOutput(oldOutput) })
			e := echo.New()
			s.Register(e, secret)
			r := httptest.NewRequest(http.MethodPost, "/api/v1/sending/domains/"+d.ID+"/check", nil)
			r.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			e.ServeHTTP(w, r)
			require.Equal(t, tc.status, w.Code, w.Body.String())
			require.Contains(t, w.Body.String(), tc.message)
			require.NotContains(t, w.Body.String(), sensitive)
			require.NotContains(t, w.Body.String(), "AccessDeniedException")
			require.Contains(t, logs.String(), `stage="`+tc.stage+`"`)
			require.NotContains(t, logs.String(), sensitive)
			require.NotContains(t, logs.String(), token)
			if tc.name == "provider identity" {
				require.Contains(t, logs.String(), `provider_code="AccessDeniedException"`)
				require.Contains(t, logs.String(), `operation="GetEmailIdentity"`)
			}
			var stored Domain
			require.NoError(t, s.DB.First(&stored, "id = ?", d.ID).Error)
			require.False(t, stored.Ready)
			a, err := s.Account(context.Background(), d.TeamID)
			require.NoError(t, err)
			require.False(t, a.Approved)
		})
	}
}

func TestMissingDNSRemainsPendingAndCanRecover(t *testing.T) {
	for _, prefix := range []string{"_xem.", "_dmarc."} {
		t.Run(prefix, func(t *testing.T) {
			s, team, d, _ := unapprovedSending(t)
			resolver := s.DNS
			s.DNS = checkDNS{resolver, prefix + d.Name, &net.DNSError{IsNotFound: true}}
			result, err := s.RefreshDomain(context.Background(), team, d.ID)
			require.NoError(t, err)
			require.False(t, result.Ready)
			s.DNS = resolver
			result, err = s.RefreshDomain(context.Background(), team, d.ID)
			require.NoError(t, err)
			require.True(t, result.Ready)
		})
	}
}

func TestProviderFailureCanRecoverWithoutReprovisioning(t *testing.T) {
	s, team, d, p := setup(t)
	s.Provider = checkProvider{Provider: p, identityErr: errors.New("provider temporarily unavailable")}
	_, err := s.RefreshDomain(context.Background(), team, d.ID)
	require.Error(t, err)
	s.Provider = p
	result, err := s.RefreshDomain(context.Background(), team, d.ID)
	require.NoError(t, err)
	require.True(t, result.Ready)
	require.Equal(t, 1, p.provisions)
}
