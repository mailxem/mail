package sending

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestAddDomainTakenByAnotherWorkspace(t *testing.T) {
	s, owner, _, _ := setup(t)
	other := uuid.NewString()
	_, e := s.AddDomain(context.Background(), other, " Example.com. ")
	var taken DomainTakenError
	require.True(t, errors.As(e, &taken))
	var he *echo.HTTPError
	require.True(t, errors.As(httpError(e), &he))
	require.Equal(t, 409, he.Code)
	require.Equal(t, "example.com is already used by another Xem workspace. Add a subdomain instead, like mail.example.com, or ask that workspace to remove it.", he.Message)
	require.False(t, strings.Contains(he.Message.(string), owner))
	// A subdomain is still free.
	_, e = s.AddDomain(context.Background(), other, "mail.example.com")
	require.NoError(t, e)
}
