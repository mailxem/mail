package utils

import (
	"bytes"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/gomail.v2"
)

func TestSMTPEnvelopeSenderPreservesFromHeader(t *testing.T) {
	for _, from := range []string{
		"human@updates.example.com",
		"Humans from Zunofy <human@updates.example.com>",
		`"Zunofy, Humans" <human@updates.example.com>`,
	} {
		t.Run(from, func(t *testing.T) {
			clientConn, serverConn := net.Pipe()
			t.Cleanup(func() { clientConn.Close(); serverConn.Close() })
			require.NoError(t, clientConn.SetDeadline(time.Now().Add(5*time.Second)))
			require.NoError(t, serverConn.SetDeadline(time.Now().Add(5*time.Second)))
			result := make(chan error, 1)
			go func() {
				result <- checkSMTPTransaction(serverConn, from)
				serverConn.Close()
			}()
			client, err := smtp.NewClient(clientConn, "localhost")
			require.NoError(t, err)
			defer client.Close()
			message := gomail.NewMessage()
			message.SetHeader("From", from)
			message.SetHeader("To", "Reader <reader@example.net>")
			message.SetHeader("Cc", "Copy <copy@example.net>")
			message.SetHeader("Bcc", "Hidden <hidden@example.net>")
			message.SetBody("text/plain", "SMTP regression test")
			err = sendSMTPMessage(client, message, from)
			require.NoError(t, <-result)
			require.NoError(t, err)
		})
	}
}

// Exercise the actual SMTP commands and serialized message, without an external relay.
func checkSMTPTransaction(conn net.Conn, from string) error {
	server := textproto.NewConn(conn)
	if err := server.PrintfLine("220 localhost ESMTP"); err != nil {
		return err
	}
	for _, step := range []struct{ command, reply string }{
		{"EHLO localhost", "250 localhost"},
		{"MAIL FROM:<human@updates.example.com>", "250 OK"},
		{"RCPT TO:<reader@example.net>", "250 OK"},
		{"RCPT TO:<copy@example.net>", "250 OK"},
		{"RCPT TO:<hidden@example.net>", "250 OK"},
		{"DATA", "354 Send message"},
	} {
		line, err := server.ReadLine()
		if err != nil {
			return err
		}
		if line != step.command {
			return fmt.Errorf("SMTP command = %q, want %q", line, step.command)
		}
		if err := server.PrintfLine("%s", step.reply); err != nil {
			return err
		}
	}
	raw, err := server.ReadDotBytes()
	if err != nil {
		return err
	}
	message, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return err
	}
	got, err := mail.ParseAddress(message.Header.Get("From"))
	if err != nil {
		return err
	}
	want, err := mail.ParseAddress(from)
	if err != nil {
		return err
	}
	if *got != *want {
		return fmt.Errorf("From header = %v, want %v", got, want)
	}
	if message.Header.Get("Bcc") != "" {
		return fmt.Errorf("message exposed Bcc header")
	}
	if err := server.PrintfLine("250 Accepted"); err != nil {
		return err
	}
	line, err := server.ReadLine()
	if err != nil {
		return err
	}
	if line != "QUIT" {
		return fmt.Errorf("SMTP command = %q, want QUIT", line)
	}
	return server.PrintfLine("221 Bye")
}

func TestSMTPRejectsMalformedSenderBeforeMailCommand(t *testing.T) {
	for _, from := range []string{"", "not-an-email", "Name <broken", "one@example.com, two@example.com", "sender@example.com\r\nRCPT TO:<other@example.com>"} {
		t.Run(from, func(t *testing.T) {
			// A nil client proves invalid input is rejected before sending SMTP commands.
			err := sendSMTPMessage(nil, gomail.NewMessage(), from)
			require.Error(t, err)
			require.True(t, strings.HasPrefix(err.Error(), "invalid SMTP sender address:"))
		})
	}
}
