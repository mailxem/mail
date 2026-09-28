package utils

import (
	"crypto/tls"
	"fmt"
	"gopkg.in/gomail.v2"
	"kori/internal/models"
	"net"
	"net/mail"
	"net/netip"
	"net/smtp"
	"os"
	"syscall"
	"time"
)

// sendSecureSMTP requires TLS before authentication and never skips certificate checks.
func sendSecureSMTP(message *gomail.Message, email *models.Email) error {
	config := email.SMTPConfig
	address := net.JoinHostPort(config.Host, fmt.Sprint(config.Port))
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12, ServerName: config.Host}
	dialer := &net.Dialer{Timeout: 20 * time.Second, Control: func(network, addr string, conn syscall.RawConn) error {
		if os.Getenv("ALLOW_PRIVATE_SMTP") == "true" {
			return nil
		}
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			return err
		}
		if !publicSMTPAddress(host) {
			return fmt.Errorf("SMTP must use a public server address")
		}
		return nil
	}}
	var conn net.Conn
	var err error
	if config.Port == 465 {
		conn, err = tls.DialWithDialer(dialer, "tcp", address, tlsConfig)
	} else {
		conn, err = dialer.Dial("tcp", address)
	}
	if err != nil {
		return err
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(45 * time.Second)); err != nil {
		return err
	}
	client, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		return err
	}
	defer client.Close()
	if config.Port != 465 {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return fmt.Errorf("SMTP server must support STARTTLS")
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return err
		}
	}
	if config.Username != "" {
		if err := client.Auth(smtp.PlainAuth("", config.Username, config.Password, config.Host)); err != nil {
			return err
		}
	}
	return sendSMTPMessage(client, message, email.From)
}

func sendSMTPMessage(client *smtp.Client, message *gomail.Message, from string) error {
	// The envelope requires a mailbox, while the From header may include a display name.
	sender, err := mail.ParseAddress(from)
	if err != nil {
		return fmt.Errorf("invalid SMTP sender address: %w", err)
	}
	if err := client.Mail(sender.Address); err != nil {
		return err
	}
	for _, header := range []string{"To", "Cc", "Bcc"} {
		for _, recipient := range message.GetHeader(header) {
			addr, err := mail.ParseAddress(recipient)
			if err != nil {
				return err
			}
			if err := client.Rcpt(addr.Address); err != nil {
				return err
			}
		}
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err = message.WriteTo(writer); err != nil {
		return fmt.Errorf("delivery outcome unknown: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("delivery outcome unknown: %w", err)
	}
	// DATA was acknowledged; a QUIT failure must not cause a duplicate delivery.
	_ = client.Quit()
	return nil
}

// TestSecureSMTP shares the same TLS and destination policy as production sends.
func TestSecureSMTP(message *gomail.Message, email *models.Email) error {
	return sendSecureSMTP(message, email)
}

// Public destination checks run after DNS resolution in Dialer.Control, so a
// hostname cannot bypass them by changing its answer between lookup and dial.
func publicSMTPAddress(host string) bool {
	ip, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}
	ip = ip.Unmap()
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return false
	}
	for _, reserved := range []string{"100.64.0.0/10", "198.18.0.0/15", "192.0.0.0/24"} {
		if netip.MustParsePrefix(reserved).Contains(ip) {
			return false
		}
	}
	return true
}
