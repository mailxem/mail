package utils

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"syscall"
	"time"

	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-sasl"
)

// DialIMAP validates the resolved destination and requires verified TLS before
// credentials are sent. Port 143 uses STARTTLS; other ports use implicit TLS.
func DialIMAP(ctx context.Context, host string, port int, username, password string) (*client.Client, error) {
	return DialIMAPAuth(ctx, host, port, username, sasl.NewPlainClient("", username, password))
}

func DialIMAPAuth(ctx context.Context, host string, port int, username string, auth sasl.Client) (*client.Client, error) {
	if host == "" || username == "" || port < 1 || port > 65535 {
		return nil, fmt.Errorf("host, valid port, and username are required")
	}
	dialer := &net.Dialer{Timeout: 15 * time.Second, Control: func(network, address string, _ syscall.RawConn) error {
		if os.Getenv("ALLOW_PRIVATE_IMAP") == "true" {
			return nil
		}
		ip, _, err := net.SplitHostPort(address)
		if err != nil || !publicSMTPAddress(ip) {
			return fmt.Errorf("IMAP must use a public server address")
		}
		return nil
	}}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(host, fmt.Sprint(port)))
	if err != nil {
		return nil, err
	}
	deadline := time.Now().Add(45 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	if err := conn.SetDeadline(deadline); err != nil {
		conn.Close()
		return nil, err
	}
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12, ServerName: host}
	if port != 143 {
		tlsConn := tls.Client(conn, tlsConfig)
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			conn.Close()
			return nil, err
		}
		conn = tlsConn
	}
	// go-imap clears deadlines before commands when Client.Timeout is zero.
	// Clamp even those resets to this request's overall budget.
	guarded := &imapDeadlineConn{Conn: conn, deadline: deadline}
	guarded.stop = context.AfterFunc(ctx, func() { _ = conn.Close() })
	im, err := client.New(guarded)
	if err != nil {
		guarded.Close()
		return nil, err
	}
	if port == 143 {
		if err := im.StartTLS(tlsConfig); err != nil {
			im.Close()
			return nil, err
		}
	}
	if err := im.Authenticate(auth); err != nil {
		im.Close()
		return nil, err
	}
	return im, nil
}

type imapDeadlineConn struct {
	net.Conn
	deadline time.Time
	stop     func() bool
}

func (c *imapDeadlineConn) SetDeadline(deadline time.Time) error {
	if deadline.IsZero() || deadline.After(c.deadline) {
		deadline = c.deadline
	}
	return c.Conn.SetDeadline(deadline)
}
func (c *imapDeadlineConn) Close() error {
	if c.stop != nil {
		c.stop()
	}
	return c.Conn.Close()
}
