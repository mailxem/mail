# Managed sending validation — 12 September 2026

## Application

Backend implementation commit: `f1af4d6`. Frontend implementation commit: `3b906b5` in the client repository. Both are on `codex/managed-sending`, based on `origin/sudo` at the start of this work. The website remote has no `sudo` branch and was left unchanged.

Passed:

- Managed sending tests against an isolated PostgreSQL 16 database with the Go race detector, including concurrent reservations, domain/workspace isolation, revocation, SMTP TLS/authentication, durable acceptance, suppression, recovery, and source campaign task retry.
- Changed backend package tests and targeted `go vet`.
- All backend command binaries build.
- Frontend: all 27 Jest tests, TypeScript checking, and Next.js production build.
- Browser preview: managed choice/domain flow, DNS waiting, successful simulated test, completion, BYO, disabled managed sending, one-time credential dialog and dismissal; desktop and 390px mobile layout. No real email sent through the preview.
- Docker image `xem-managed-sending:local` builds with the application and operator CLI in a nonroot distroless runtime. `/app/sending-admin -h` runs from the image.
- CloudFormation template passed cfn-lint 1.56.3 and AWS `validate-template`. Its resolved identity policy passed AWS IAM Access Analyzer validation with no findings.

The full pre-existing `go test ./...` suite remains blocked by compile errors in `internal/ai/agent/knowledge_test.go`: those tests use fields/status constants absent from the current Contact and EmailTracking models. This file is unchanged by this feature. The new integration suite is separate and passes; a passing feature suite does not imply a green repository-wide CI run.

## Live activation

The `xem` AWS CLI profile was refreshed. Its region is `us-east-2`. At inspection, SES is healthy with sending enabled but production access disabled; quota is 200 recipients/day and one/second. Both bounce and complaint account suppression are enabled. There were no SES identities before this setup.

Created the `xem.email` SES identity with 2048-bit DKIM and configured `bounce.xem.email` as custom MAIL FROM with reject-on-MX-failure. No real email has been sent. Completed DNS verification and production approval are recorded below.

The authoritative nameservers for `xem.email` are Cloudflare. The old Route53 zone is not authoritative; changing it would not verify the domain. Receiving MX points to iCloud. Wrangler can read the active zone but its OAuth token lacks DNS record permissions, so the owner signed into the Cloudflare dashboard and the records were published through that authenticated UI.

A running `xem-server` EC2 instance exists in this region with IMDSv2 required and no IAM instance profile. The repository's old EKS deployment workflow does not match the discovered hosting: no EKS clusters were found in us-east-1 or us-east-2. Before deploying, establish the actual container/service rollout process and attach a dedicated least-privilege runtime role; never put the interactive root login credentials on the application server.

The application is committed locally; no application deployment, IAM role attachment, SMTP port exposure, or CloudFormation stack creation has occurred. Domain verification and an SES production access request do not by themselves enable customer sending. Public SMTP still requires TLS certificates, the listener deployment, DNS, firewall configuration, operator approvals, and verified provider feedback.

### Completed DNS and SES activation

On 12 September 2026 (Asia/Kolkata), all six records in `xem-email-dns.json` were published in Cloudflare: three DNS-only DKIM CNAMEs, bounce MX/SPF, and `_dmarc.xem.email` with `v=DMARC1; p=none`. Each has a 300-second TTL. No existing DMARC record existed at the exact root sending domain. Existing iCloud receiving MX, root SPF, iCloud DKIM, and other records were preserved.

At 2026-09-11 19:52 UTC, all six records matched the intended values on both authoritative nameservers (`marek.ns.cloudflare.com`, `paige.ns.cloudflare.com`) and public resolvers `1.1.1.1` and `8.8.8.8`. Root MX still resolves to `mx01.mail.icloud.com` and `mx02.mail.icloud.com`; root SPF remains `v=spf1 include:icloud.com ~all`.

SES subsequently confirmed `VerifiedForSendingStatus=true`, identity verification `SUCCESS`, DKIM `SUCCESS`, and custom MAIL FROM `SUCCESS` in `us-east-2`. The explicitly authorized production-access request was then submitted via `sesv2 put-account-details`. AWS granted it under case `178915636800118`; a follow-up `get-account` confirmed `ProductionAccessEnabled=true` and `SendingEnabled=true`. The assigned quota is 50,000 recipients per 24 hours and 14 recipients/second. Bounce and complaint account suppression remain enabled. No real email was sent.

This AWS quota is provider capacity, not the initial pilot operating limit. Retain the request's controlled, manually approved pilot target of at most 200 recipient sends/day until live feedback, monitoring, and operator procedures are validated. Production service rollout prerequisites above remain outstanding.
