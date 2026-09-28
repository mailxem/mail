package sending

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestProviderRejectionAndUncertaintyNeverRetry(t *testing.T) {
	for _, tc := range []struct {
		code   string
		fault  smithy.ErrorFault
		status string
	}{
		{"AccessDeniedException", smithy.FaultUnknown, "FAILED"},
		{"AccessDenied", smithy.FaultUnknown, "FAILED"},
		{"MessageRejected", smithy.FaultClient, "FAILED"},
		{"InternalFailure", smithy.FaultServer, "DELIVERY_UNKNOWN"},
		{"UnrecognizedError", smithy.FaultUnknown, "DELIVERY_UNKNOWN"},
	} {
		t.Run(tc.code, func(t *testing.T) {
			s, team, d, p := setup(t)
			p.err = fmt.Errorf("SES SendEmail: %w", &smithy.GenericAPIError{Code: tc.code, Message: "private provider diagnostics", Fault: tc.fault})
			ctx := context.Background()
			m, err := s.Submit(ctx, input(team, d))
			require.NoError(t, err)
			require.NoError(t, s.ProcessOne(ctx))
			require.NoError(t, s.DB.First(&m, "id = ?", m.ID).Error)
			require.Equal(t, tc.status, m.Status)
			require.NotContains(t, m.Detail, "private provider diagnostics")
			if tc.status == "FAILED" {
				require.Equal(t, "Provider rejected the message: "+tc.code, m.Detail)
			}
			require.ErrorIs(t, s.ProcessOne(ctx), gorm.ErrRecordNotFound)
			require.Equal(t, 1, p.sends)
		})
	}
}
