package utils

import (
	"bufio"
	"errors"
	"net"
	"net/smtp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/gomail.v2"
)

// fakeSMTP accepts one message and answers its end of DATA with reply.
func fakeSMTP(t *testing.T, reply string, drop bool) *smtp.Client {
	t.Helper()
	client, server := net.Pipe()
	go func() {
		defer server.Close()
		r := bufio.NewReader(server)
		write := func(s string) { server.Write([]byte(s + "\r\n")) }
		write("220 relay ready")
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				return
			}
			switch cmd := strings.ToUpper(strings.TrimSpace(line)); {
			case strings.HasPrefix(cmd, "EHLO"), strings.HasPrefix(cmd, "HELO"):
				write("250 relay")
			case strings.HasPrefix(cmd, "MAIL"), strings.HasPrefix(cmd, "RCPT"), strings.HasPrefix(cmd, "RSET"), strings.HasPrefix(cmd, "NOOP"):
				write("250 ok")
			case cmd == "DATA":
				write("354 go ahead")
				for {
					body, err := r.ReadString('\n')
					if err != nil || body == ".\r\n" {
						break
					}
				}
				if drop {
					return
				}
				write(reply)
			case cmd == "QUIT":
				write("221 bye")
				return
			}
		}
	}()
	c, err := smtp.NewClient(client, "relay.test")
	require.NoError(t, err)
	t.Cleanup(func() { c.Close() })
	return c
}

func message() *gomail.Message {
	m := gomail.NewMessage()
	m.SetHeader("From", "hello@example.com")
	m.SetHeader("To", "reader@example.net")
	m.SetHeader("Subject", "Hello")
	m.SetBody("text/plain", "Hello!")
	return m
}

func TestARefusedMessageIsAFailureNotAnUnknownDelivery(t *testing.T) {
	err := sendSMTPMessage(fakeSMTP(t, "452 4.0.0 Sending quota or queue limit reached", false), message(), "hello@example.com")
	require.Error(t, err)
	require.False(t, errors.Is(err, ErrSMTPDeliveryUnknown), "a 452 reply means the relay refused the message")
	require.Contains(t, err.Error(), "452")

	err = sendSMTPMessage(fakeSMTP(t, "", true), message(), "hello@example.com")
	require.ErrorIs(t, err, ErrSMTPDeliveryUnknown, "a dropped connection after DATA leaves the outcome unknown")

	require.NoError(t, sendSMTPMessage(fakeSMTP(t, "250 queued", false), message(), "hello@example.com"))
}
