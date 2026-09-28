package utils

import (
	"context"
	"github.com/stretchr/testify/require"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestIMAPRejectsPrivateAddressUnlessOperatorOptsIn(t *testing.T) {
	t.Setenv("ALLOW_PRIVATE_IMAP", "false")
	_, err := DialIMAP(context.Background(), "127.0.0.1", 993, "user", "password")
	require.ErrorContains(t, err, "public server address")
}

func TestIMAPLibraryCannotClearRequestDeadline(t *testing.T) {
	left, right := net.Pipe()
	defer left.Close()
	defer right.Close()
	conn := &imapDeadlineConn{Conn: left, deadline: time.Now().Add(20 * time.Millisecond)}
	require.NoError(t, conn.SetDeadline(time.Time{}))
	_, err := conn.Write([]byte("blocked write"))
	require.Error(t, err)
	var networkError net.Error
	require.ErrorAs(t, err, &networkError)
	require.True(t, networkError.Timeout())
}

func TestIMAPStillRejectsUntrustedTLSWithPrivateOptIn(t *testing.T) {
	t.Setenv("ALLOW_PRIVATE_IMAP", "true")
	s := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("untrusted TLS must fail before application traffic")
	}))
	defer s.Close()
	host, rawPort, err := net.SplitHostPort(s.Listener.Addr().String())
	require.NoError(t, err)
	port, err := strconv.Atoi(rawPort)
	require.NoError(t, err)
	_, err = DialIMAP(context.Background(), host, port, "user", "password")
	require.ErrorContains(t, err, "certificate")
}
