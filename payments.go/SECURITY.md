# Security Guide

## Overview

This document outlines the security measures implemented in the payments service and best practices for deployment.

## Security Features Implemented

### 1. Rate Limiting

Three levels of rate limiting are implemented to prevent DoS attacks:

- **Generous (5 req/sec, burst 20)**: Applied globally to all endpoints
- **Moderate (1 req/sec, burst 10)**: Applied to payment endpoints
- **Strict (0.166 req/sec, burst 3)**: Applied to admin endpoints

Location: `internal/middleware/ratelimit.go`

### 2. CORS Protection

Restrictive CORS policy that:
- Only allows configured frontend origins (no wildcards)
- Limits HTTP methods to GET, POST, PUT, DELETE, OPTIONS
- Restricts headers to necessary ones only
- Credentials allowed only from trusted origins
- 1-hour max age for preflight caching

Configuration: `cmd/main.go:setupMiddleware()`

### 3. Security Headers

All responses include:
- `X-Frame-Options: DENY` - Prevents clickjacking
- `X-Content-Type-Options: nosniff` - Prevents MIME sniffing
- `X-XSS-Protection: 1; mode=block` - XSS protection for legacy browsers
- `Referrer-Policy: strict-origin-when-cross-origin` - Controls referrer leakage
- `Content-Security-Policy: default-src 'none'` - Strict CSP for API
- `Permissions-Policy` - Disables unnecessary browser features
- `Strict-Transport-Security` - HTTPS enforcement (production only)

Location: `internal/middleware/security.go`

### 4. Authentication & Authorization

#### Admin Authentication
- Constant-time comparison to prevent timing attacks
- Secure admin secret validation (min 32 chars in production)
- Failed auth attempts logged with IP addresses
- Rate limited to prevent brute force

#### Team-scoped Access Control
- X-Team-ID header required for team-specific operations
- Validates team ID matches resource being accessed
- Prevents unauthorized cross-team access

Location: `internal/middleware/auth.go`, `internal/middleware/security.go`

### 5. Webhook Security

- **Signature verification**: Using standard-webhooks library
- **Replay protection**: Rejects webhooks older than 5 minutes
- **Timestamp validation**: Prevents replay attacks
- **Clock skew tolerance**: 1 minute future tolerance

Location: `internal/services/webhook_service.go`

### 6. Input Validation & Sanitization

Multi-layer input protection:
- Struct validation using go-playground/validator
- Input sanitization (HTML escaping, XSS prevention)
- SQL injection pattern detection (defense in depth)
- Path traversal prevention
- UUID/ID sanitization

Location: `internal/middleware/sanitize.go`

### 7. Error Handling

- Sanitized error messages for clients (no stack traces/internal details)
- Detailed errors logged server-side only
- Custom error handler prevents information disclosure
- Production mode further restricts 5xx error messages

Location: `internal/middleware/errors.go`

### 8. Request Size Limits

- Maximum request body: 10MB
- Prevents memory exhaustion attacks
- Applied globally to all endpoints

### 9. Audit Logging

Security events logged:
- Admin authentication attempts (success/failure)
- Admin operations with IP addresses
- Failed authentication attempts
- Sensitive endpoint access

Location: `internal/middleware/security.go`

### 10. Configuration Security

Production environment enforces:
- JWT_SECRET: minimum 32 characters, no weak defaults
- ADMIN_SECRET: minimum 32 characters, no weak defaults
- DODO_WEBHOOK_SECRET: required
- Database SSL: warning if not using sslmode=require
- No default/weak secrets allowed (blocks: "password", "secret", "test", etc.)

Location: `internal/config/config.go`

## Deployment Security Checklist

### Environment Configuration

**Required in Production:**

```bash
# Secrets - Generate strong random strings (32+ characters)
JWT_SECRET=<strong-random-secret-min-32-chars>
ADMIN_SECRET=<strong-random-secret-min-32-chars>
DODO_WEBHOOK_SECRET=<from-dodo-payments-dashboard>

# Database - Always use SSL in production
DATABASE_URL=postgresql://user:pass@host:5432/db?sslmode=require

# Environment
ENVIRONMENT=production

# CORS - Restrict to your actual frontend domains
FRONTEND_URL=https://yourdomain.com
```

**Generate Strong Secrets:**

```bash
# Generate 32-byte random secrets
openssl rand -base64 32

# Or use uuidgen for simpler secrets
uuidgen
```

### Network Security

1. **TLS/HTTPS**: Always deploy behind TLS-terminating reverse proxy (nginx, Cloudflare, AWS ALB)
2. **Firewall**: Restrict database access to application servers only
3. **Private Networks**: Run database in private subnet, not internet-accessible
4. **DDoS Protection**: Use Cloudflare or AWS Shield

### Database Security

1. **SSL Required**: Use `sslmode=require` or `sslmode=verify-full` in DATABASE_URL
2. **Least Privilege**: Database user should have minimum necessary permissions
3. **Connection Pooling**: Current limits (10 idle, 100 max) are reasonable
4. **Backups**: Enable automated encrypted backups
5. **Network Isolation**: Database should not be publicly accessible

### Monitoring & Alerting

Set up alerts for:
- Failed admin authentication attempts (potential brute force)
- Rate limit violations (potential DoS)
- Webhook signature failures (potential attack)
- 5xx error rate spikes
- Unusual traffic patterns

### Secrets Management

**Do NOT:**
- Commit `.env` files to git (already in `.gitignore`)
- Use default/weak secrets in production
- Share secrets via email/Slack
- Hard-code secrets in source code

**DO:**
- Use environment variables for all secrets
- Rotate secrets regularly (quarterly recommended)
- Use secret management service (AWS Secrets Manager, HashiCorp Vault, etc.)
- Different secrets per environment (dev/staging/production)

### Regular Security Maintenance

- [ ] Update dependencies monthly: `go get -u ./... && go mod tidy`
- [ ] Review security logs weekly
- [ ] Rotate secrets quarterly
- [ ] Audit user access quarterly
- [ ] Review and update CORS origins when adding new frontends
- [ ] Test disaster recovery procedures

### Incident Response

If security incident suspected:

1. **Contain**: Rotate compromised secrets immediately
2. **Investigate**: Review audit logs, identify scope
3. **Notify**: Inform affected users if PII/payment data exposed
4. **Remediate**: Fix vulnerability, deploy patches
5. **Document**: Post-mortem to prevent recurrence

## Security Testing

### Before Production Deployment

Run security checks:

```bash
# Check for known vulnerabilities in dependencies
go list -json -m all | nancy sleuth

# Static analysis
gosec ./...

# Check for secrets in code
gitleaks detect --source . --verbose

# Dependency audit
go mod verify
```

### Penetration Testing

Recommended annual penetration testing covering:
- Authentication bypass attempts
- SQL injection (should be blocked by GORM)
- XSS attempts (should be sanitized)
- CSRF attacks (CORS protection)
- Rate limit bypass attempts
- Webhook signature bypass attempts

## PCI DSS Considerations

This service handles payment subscriptions. Key requirements:

1. **No Card Data Storage**: We use Dodo Payments - card data never touches our servers ✓
2. **TLS 1.2+**: Enforce HTTPS in production ✓
3. **Access Control**: Admin authentication with audit logging ✓
4. **Logging**: Security event logging implemented ✓
5. **Vulnerability Management**: Keep dependencies updated

## Contact

For security issues, contact: security@synehq.com

**Do not** open public GitHub issues for security vulnerabilities.

---

Last Updated: 2026-04-16
