package sending

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/emersion/go-sasl"
	smtp "github.com/emersion/go-smtp"
	"github.com/google/uuid"
	"golang.org/x/net/netutil"
	"golang.org/x/time/rate"
)

type authBucket struct {
	limiter *rate.Limiter
	used    time.Time
}
type SMTPBackend struct {
	Service *Service
	mu      sync.Mutex
	buckets map[string]authBucket
}

func (b *SMTPBackend) allow(ip string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.buckets == nil {
		b.buckets = map[string]authBucket{}
	}
	now := time.Now()
	for k, v := range b.buckets {
		if now.Sub(v.used) > 10*time.Minute {
			delete(b.buckets, k)
		}
	}
	bucket, ok := b.buckets[ip]
	if !ok {
		if len(b.buckets) >= 10000 {
			return false
		}
		bucket = authBucket{limiter: rate.NewLimiter(rate.Every(10*time.Second), 6)}
	}
	bucket.used = now
	b.buckets[ip] = bucket
	return bucket.limiter.Allow()
}
func (b *SMTPBackend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	ip, _, _ := net.SplitHostPort(c.Conn().RemoteAddr().String())
	if !b.allow(ip) {
		return nil, &smtp.SMTPError{Code: 421, Message: "Too many connections; try later"}
	}
	return &smtpSession{backend: b, conn: c, ip: ip}, nil
}

type smtpSession struct {
	backend    *SMTPBackend
	conn       *smtp.Conn
	ip         string
	credential Credential
	from       string
	to         []string
}

func (s *smtpSession) AuthMechanisms() []string { return []string{sasl.Plain} }
func (s *smtpSession) Auth(mechanism string) (sasl.Server, error) {
	if mechanism != sasl.Plain {
		return nil, smtp.ErrAuthUnsupported
	}
	return sasl.NewPlainServer(func(identity, username, password string) error {
		if _, ok := s.conn.TLSConnectionState(); !ok {
			return smtp.ErrAuthRequired
		}
		if (identity != "" && identity != username) || !s.backend.allow(s.ip) {
			return smtp.ErrAuthFailed
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		c, e := s.backend.Service.Authenticate(ctx, username, password)
		if e != nil {
			return smtp.ErrAuthFailed
		}
		s.credential = c
		return nil
	}), nil
}
func (s *smtpSession) Reset()        { s.from = ""; s.to = nil }
func (s *smtpSession) Logout() error { s.Reset(); s.credential = Credential{}; return nil }
func (s *smtpSession) Mail(from string, _ *smtp.MailOptions) error {
	if s.credential.ID == "" {
		return smtp.ErrAuthRequired
	}
	addr, e := address(from)
	if e != nil {
		return &smtp.SMTPError{Code: 553, Message: "A valid nonempty sender is required"}
	}
	s.from = addr
	s.to = nil
	return nil
}
func (s *smtpSession) Rcpt(to string, _ *smtp.RcptOptions) error {
	if s.credential.ID == "" {
		return smtp.ErrAuthRequired
	}
	addr, e := address(to)
	if e != nil {
		return &smtp.SMTPError{Code: 553, Message: "Invalid recipient"}
	}
	if len(s.to) >= 50 {
		return &smtp.SMTPError{Code: 452, Message: "Too many recipients"}
	}
	s.to = append(s.to, addr)
	return nil
}
func (s *smtpSession) Data(r io.Reader) error {
	if s.credential.ID == "" {
		return smtp.ErrAuthRequired
	}
	raw, e := io.ReadAll(io.LimitReader(r, s.backend.Service.Config.MaxBytes+1))
	if e != nil {
		return e
	}
	if int64(len(raw)) > s.backend.Service.Config.MaxBytes {
		return &smtp.SMTPError{Code: 552, Message: "Message too large"}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	_, e = s.backend.Service.Submit(ctx, Submission{TeamID: s.credential.TeamID, DomainID: s.credential.DomainID, CredentialID: s.credential.ID, Key: uuid.NewString(), From: s.from, Recipients: s.to, Raw: raw})
	if e != nil {
		if errors.Is(e, ErrLimit) {
			return &smtp.SMTPError{Code: 452, Message: "Sending quota or queue limit reached"}
		}
		if errors.Is(e, ErrDenied) || errors.Is(e, ErrSuppressed) {
			return &smtp.SMTPError{Code: 550, Message: "Sender or recipient policy rejected this message"}
		}
		return &smtp.SMTPError{Code: 451, Message: "Message not accepted; check sending settings and retry later"}
	}
	return nil
}
func newSMTPServer(service *Service) (*smtp.Server, error) {
	cfg := service.Config
	if _, e := tls.LoadX509KeyPair(cfg.TLSCert, cfg.TLSKey); e != nil {
		return nil, e
	}
	server := smtp.NewServer(&SMTPBackend{Service: service})
	server.Addr = cfg.SMTPAddr
	server.Domain = cfg.SMTPHost
	server.MaxRecipients = 50
	server.MaxMessageBytes = cfg.MaxBytes
	server.MaxLineLength = 1000
	server.ReadTimeout = 30 * time.Second
	server.WriteTimeout = 30 * time.Second
	server.AllowInsecureAuth = false
	server.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12, GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
		cert, e := tls.LoadX509KeyPair(cfg.TLSCert, cfg.TLSKey)
		return &cert, e
	}}
	return server, nil
}

func StartSMTP(service *Service) (*smtp.Server, error) {
	server, e := newSMTPServer(service)
	if e != nil {
		return nil, e
	}
	listener, e := net.Listen("tcp", service.Config.SMTPAddr)
	if e != nil {
		return nil, e
	}
	go func() { _ = server.Serve(netutil.LimitListener(listener, 100)) }()
	return server, nil
}
