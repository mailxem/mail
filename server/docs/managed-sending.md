# Managed sending and onboarding

Xem can send through SES using a customer's verified domain, or keep using an existing SMTP provider. Customers keep their existing mailboxes and receiving MX records. Xem's optional SMTP submission endpoint accepts authenticated mail from external applications and queues it for SES; it is not an inbound mailbox or a direct-delivery MTA.

The feature is off by default. This release provides the application, queue, feedback handling, operator controls, and onboarding. Live AWS provisioning and real inbox delivery must be validated in the deployment account before launch. A successful local test is not proof of AWS production access or inbox placement.

## Activation

1. Use a commercial AWS region where SES tenants are available. Request SES production access, review the account's rate/recipient quotas, and enable account-level bounce and complaint suppression. Use a dedicated AWS account for Xem managed sending to bound reputation and financial exposure. SES tenant separation does not eliminate shared account risk.
2. Back up PostgreSQL. Deploy the backend first with managed sending disabled; startup creates additive `managed_*` tables. Deploy the matching frontend, which adds `/onboarding` and `/settings/sending`. Admin sessions can configure sending; ordinary members and general API keys cannot access these management endpoints.
3. Create a change set from `deploy/managed-sending/cloudformation.json`, providing the existing backend IAM role name and leaving `FeedbackEndpoint` blank. Review and execute it in the intended region. It creates an SNS topic and attaches a scoped managed policy; it creates no AWS access keys and does not change role trust. Prefer workload identity (EKS role/EC2 role) over environment credentials.
4. Set `MANAGED_SES_REGION`, `MANAGED_AWS_ACCOUNT_ID`, and `MANAGED_SNS_TOPIC_ARN` from the stack outputs, then set `MANAGED_SENDING_ENABLED=true`. Redeploy the backend. The API must be reachable by SNS over HTTPS on `/api/v1/sending/events/ses`; preserve the signed body. Do not disable signature validation or rewrite the body in a proxy.
5. Update the stack's `FeedbackEndpoint` to the exact HTTPS route. The backend authenticates and confirms SNS subscriptions itself. Verify the subscription is confirmed. Failed notifications must be investigated; create CloudWatch alerts for SNS failed notifications, SES reputation/complaints/bounces, AWS spend, queue age, and `DELIVERY_UNKNOWN`. No alarms or billing resources are automatically provisioned by this template.
6. Sign in as a workspace administrator, open Getting started, and choose managed sending. Add a domain and publish the `_xem` TXT ownership challenge. The account starts unapproved. On **Check my records**, verified ownership automatically provisions the SES tenant/configuration set, domain identity with 2048-bit DKIM, and custom MAIL FROM `bounce.DOMAIN`. Unverified domains do not create AWS resources; suspended accounts cannot provision new ones.
7. Add the returned DKIM CNAMEs, bounce MX (priority 10), bounce SPF TXT, and an explicit DMARC TXT at `_dmarc.DOMAIN`, then check again. Preserve receiving MX and existing DMARC policies. DMARC must exist at the exact sending domain, including subdomains; organizational-domain inheritance is not implemented. Ownership, SES identity verification, DKIM, MAIL FROM, and DMARC must all pass before sending is enabled.
8. A fully verified domain automatically approves its workspace at **at most 100 recipient sends/day, 3,000/month, and 3,000,000 micro-USD/month of reserved delivery allowance**. Lower existing limits remain lower, and usage is preserved. Already approved accounts retain their operator-set limits. A paused or suspended account is never automatically approved or unblocked. The approval and verified domain state commit together, and an `auto_approve` audit records the decision.
9. Save a sender and send a fixed-content test to the authenticated administrator's address. Check SES acceptance and authenticated Delivery feedback, then create a campaign draft. Checklist progress comes from stored configuration and successful test events.

Automatic onboarding applies to existing unapproved workspaces too: checking records, or the periodic DNS worker, can provision an ownership-verified domain and approve a fully verified workspace. There is no new environment switch. `MANAGED_SENDING_ENABLED=false` still disables the feature. The migration adds an internal `managed_domains.check_id`; stale concurrent DNS results and disconnected-domain results cannot approve an account.

### Operator review and abuse controls

For higher limits, review the customer's use case, consent practices, volumes, domain, and abuse risk. From a trusted backend environment with the correct database configuration, run:

```sh
go run ./cmd/sending-admin -team WORKSPACE_UUID -action approve -daily 100 -monthly 3000 -budget-micros 3000000
```

The CLI preserves usage and records an audit. It can restore a suspended account after investigation; it never clears the customer's pause. Limits remain positive and bounded at 1 billion, with budget expressed in micro-USD. `-action suspend` blocks an account. Workspace administrators cannot approve themselves or raise their own limits through the API.

Authenticated SES feedback automatically suspends a workspace on **one complaint affecting a known recipient**, or **five distinct permanently bounced recipients recorded within the last 24 hours**. These conservative absolute thresholds apply to managed accounts, including accounts with reviewed limits; they are not a calculation of SES reputation rates. Transient bounces, duplicate recipients/events, other workspaces, and older suppressions do not add to the hard-bounce threshold. Suspension, suppression, message feedback, and the `auto_suspend_complaint`/`auto_suspend_bounces` audit commit atomically. Invalid routing metadata or recipients cannot suspend an account. Concurrent feedback is serialized per workspace.

New submissions and credentials are blocked immediately after suspension. Queued messages are rechecked before dispatch; an already in-flight SES request may complete. DNS rechecks never clear suspension. Investigate the cause before using the operator CLI to restore access, and retain address suppressions. Fleet-wide rate control and SES/CloudWatch monitoring remain necessary: per-workspace starter caps do not bound aggregate signup volume or AWS account spend.

AWS IAM reference used for the template: [SES v2 actions and resource types](https://docs.aws.amazon.com/service-authorization/latest/reference/list_sesv2.html). Identity permissions cover domains in this region/account because customer domains are not known in advance. Xem enforces domain ownership and tenant scope. Tenant and configuration-set permissions are restricted to `xem-*`; sending requires that tenant prefix. SNS permits SES publication only from this account's `xem-*` configuration sets. Isolate this role and account: a compromised runtime can still act across its managed customer domains.

## Optional external SMTP submission

The raw MIME transport requires `ses:SendRawEmail` even though the SDK operation
is SESv2 `SendEmail`. Keep both sending actions in the scoped runtime policy and
update both the backend and devops CloudFormation template copies together.
Explicit IAM `AccessDenied`/`AccessDeniedException` responses are recorded as
FAILED even when the SDK reports an unknown fault classification. Transport
failures with uncertain acceptance remain DELIVERY_UNKNOWN and are never retried
automatically. This classification does not rewrite historical message outcomes.

Set `MANAGED_SMTP_ENABLED=true`, a listening address in `MANAGED_SMTP_ADDR` (default `:587`), the public hostname in `MANAGED_SMTP_HOST`, and readable PEM certificate/key paths in `MANAGED_SMTP_TLS_CERT` and `MANAGED_SMTP_TLS_KEY`. The hostname must resolve to the listener. Expose TCP port 587 through a layer-4 load balancer/firewall; ordinary HTTP ingress is insufficient. The UI advertises port 587; map the public port to your internal bind port. Use a publicly trusted certificate for the hostname. Certificate files reload on each TLS handshake; replace them atomically during renewal.

Customers create a named credential for a verified domain. The username is its UUID, the password is displayed once, and only a bcrypt hash is stored. Credentials expire in 90 days and are revocable. They cannot be used as AWS credentials or management API keys. Clients must use STARTTLS and PLAIN authentication inside TLS; unauthenticated relay and plaintext authentication are rejected. Responses acknowledge only a committed outbox entry. SMTP itself does not provide exactly-once delivery: a client retry after losing the final 250 response can create a duplicate. Internal Xem sends use a stable idempotency key.

The listener limits connections (100), per-IP connection/auth attempts, recipients (50), message size, line lengths, and read/write duration. If using a proxy, retain meaningful source-IP controls at the edge: a TCP proxy may cause application limits to apply to its shared IP. Sender headers must match the envelope and the credential's verified domain. Xem removes Bcc and provider-routing headers. Customer apps must implement consent and unsubscribe correctly for marketing mail; generic SMTP cannot reliably infer whether arbitrary content is marketing. Xem campaign submissions enforce HTTPS one-click unsubscribe headers.

## Delivery and recovery semantics

- PostgreSQL is the durable outbox. Workers claim with row locks and `SKIP LOCKED`, then recheck workspace, domain, credentials, and recipient suppression. `QUEUED` means stored, `SENT` means SES accepted, and `DELIVERED` means SES reported recipient-server acceptance. None proves placement in the inbox.
- Daily/monthly limits count recipient sends, including tests, when queued. A configurable conservative cost per recipient reserves the monthly allowance. This is a safety cap, not an invoice or reconciliation with AWS pricing; tenant fees, data charges, and extras need separate billing. Reservations are not refunded after failure or suppression.
- The initial worker sends one message at a time per process, with a one-second interval. Keep a single replica during the pilot. Multiple replicas safely share claims but multiply throughput; set account limits and monitor the account-wide SES rate quota before scaling.
- SES SDK automatic sending retries are disabled. A timeout, network error, or ambiguous provider response produces `DELIVERY_UNKNOWN`. Crashed `SENDING` rows are quarantined after five minutes. They are never automatically replayed. Search authenticated provider events/SES records before deciding whether a new send is warranted.
- Raw MIME is removed after seven days. Old queued messages expire as failed before content deletion. Recovery updates both the outbox and its source Xem email. Metadata (addresses, subject, events, domain history) remains for support; define an operator retention policy and secure database/backups accordingly. No automatic metadata deletion is supplied.
- Permanent bounces and complaints suppress the address within its workspace; existing suppressions and unsubscribed contacts also block sending. Dispatch rechecks prevent a queued message bypassing a newly recorded suppression. Do not remove suppressions merely to force a retry.
- Domain checks run periodically. Verification older than 24 hours blocks sending. Missing DNS or provider errors clear readiness. Disconnect revokes credentials and invalidates the ownership token while retaining history and the AWS identity. Transfers and abandoned domain claims require operator review; there is no automatic domain transfer/deletion.
- A missing DNS record is a successful check with pending verification. Resolver failures return 503, check timeouts return 504, SES setup/status failures return 502, and internal verification failures return 500. These responses describe domain verification rather than message sending. For HTTP and periodic checks, search backend logs for `managed domain check failed`; correlate `domain_id`, `stage`, `region`, `provider_code`, `operation`, `request_id`, and `sql_state`. Raw provider messages, credentials, ownership tokens, and SQL are excluded from this diagnostic log. Confirm the running backend version before expecting these diagnostics.
- Paused accounts and domains waiting for readiness keep messages queued until they become eligible or expire after seven days. Pause, suspension, revocation, and domain disconnection stop future eligible dispatches. A request already submitted to SES may complete. Policy-rejected queued messages become failed/suppressed and retain their quota reservation.
- SES requires TLS for delivery through each new managed configuration set. Servers without compatible TLS will not receive these messages; monitor delays/bounces. Do not reuse unrelated pre-existing `xem-*` resources.

## Verification and release gate

Local tests use fake DNS/SES with real PostgreSQL and a real localhost SMTP/TLS exchange; they never send external email. Run:

```sh
go test -race ./internal/sending ./internal/utils ./internal/tasks ./internal/api/controllers ./internal/handlers
go build ./cmd/...
POSTHOOT_TEST_DATABASE_URL='host=127.0.0.1 port=5432 user=TEST_USER dbname=TEST_DB sslmode=disable' go test -race ./internal/sending
```

Use an isolated test database and credentials that can create/drop test schemas. With no DSN, sending tests use SQLite. In the frontend run `bunx tsc --noEmit` and `bun run build`. The development-only `/preview` route includes managed onboarding scenarios, with simulated transport and no real sends; production rejects the preview route.

Before enabling customer accounts, validate the CloudFormation change set and IAM behavior in the real AWS account, SES tenant/resource association, SNS confirmation and retries, domain propagation, the SES mailbox simulator's success/bounce/complaint paths, self-test acceptance/delivery, TLS certificate rotation, operator suspension, credential revocation, and outage recovery. Real sends require an explicitly authorized test recipient or the SES simulator. Review spend/reputation alerting and abuse operations. Automatic starter approval does not replace these deployment checks or operator monitoring.

To roll back, disable `MANAGED_SMTP_ENABLED` and `MANAGED_SENDING_ENABLED` and redeploy. Existing BYO SMTP remains usable. Keep the additive tables and SNS topic until queued/ambiguous messages and delayed feedback have been reconciled. Do not downgrade to a backend that treats `MANAGED` sender records as ordinary SMTP records. An unavailable feedback route causes SNS retries; restore verification/processing for reconciliation.

## Contribution-sized follow-ups

Prioritize measured activation and delivery reliability over a general assistant. Good first contributions: DNS-provider instructions and copyable record names, localized onboarding copy, accessibility audits, and additional deterministic credential/feedback fixtures. Larger contributions: organizational DMARC inheritance with public-suffix tests, a domain transfer workflow, durable delivery metrics/alerts, account-wide rate scheduling, and a feedback dead-letter/replay operator tool. Require workspace-isolation tests and redacted examples for every provider integration. Never include live keys or customer messages in issues.

### Free allowance upgrade

At startup, the migration upgrades only approved, unpaused, unsuspended accounts with the exact former 50/day, 200/month, 200,000 micro-USD bundle, an automatic DNS approval audit, and no operator audit. It preserves all usage and records `starter_allowance_v2` atomically. Custom limits and operator decisions require explicit review. Explicit `MANAGED_*_LIMIT` and budget environment overrides still apply to new accounts. The $3 reserved monthly allowance is an internal delivery cap, not a customer charge; at the default 1,000 micro-USD per recipient it covers 3,000 sends.
