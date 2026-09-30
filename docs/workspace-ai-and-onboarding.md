# Workspace AI and managed sending updates

Xem AI opens in the bottom-right corner of the dashboard. Its Workspace tab keeps
the existing assistant, history, and explicit approval cards for workspace actions.
The Email design tab works on the open template, including new templates.

Describe the email, then ask for revisions. Each request uses the current editor
content. A successful revision replaces the editable design and subject, saves it,
and stays on the editor page. A new design gets its own template URL. No AI design
operation sends an email.

## Editing safeguards

- The app authenticates design and save requests on the server and uses
  `INTERNAL_API_URL`. Tokens are never included in generated email content.
- Ordinary backend requests retain the 30-second timeout. Email generation has a
  95-second request context around the writer's 90-second upstream timeout; the
  frontend proxy allows 100 seconds. Configure any deployment gateway to permit
  at least 120 seconds for `/api/assistant/design`.
- Template writes verify permissions, workspace ownership, category ownership,
  and the expected `updatedAt` version. Repeated creation requests use a stable
  client-generated UUID to avoid duplicates after a lost response.
- Navigating to another template or workspace cancels pending generation. Edits
  made during generation prevent automatic replacement. A generated revision
  remains available for preview, download, or explicit application.
- Failed saves retain the local design. A conflict requires checking the latest
  saved template; it is never silently overwritten. Download a held revision
  before reloading if you want to keep both versions.
- Undo restores the previous design and subject only while there are no later
  manual edits. It saves through the same version check.
- Long email content is rejected rather than silently truncated. Export-only
  styles, scripts, and comments are removed from AI context first. The writer
  accepts up to 16,000 characters of actual email HTML.
- Design conversations stay in memory for the current editor; navigating away
  clears them. Workspace conversations retain their existing seven-day policy.

## Owner notifications

Managed sending records these one-time milestones:

1. Domain added.
2. Ownership verified.
3. Domain authentication ready.
4. Workspace approved (automatic or operator approval).
5. Approved managed sender connected.
6. First test queued.
7. Test accepted by the provider.
8. Test accepted by the receiving server, based on provider feedback.
9. Test needing attention (failed, delayed, bounced, complained, or uncertain).

These are transactional workspace updates, not marketing subscriptions. They do
not include SMTP credentials, DNS ownership tokens, or customer email content.
Acceptance is not called delivery, and delivery is not called inbox placement.
Provider feedback must already be configured for the delivery milestone.

New workspaces record their creator as `teams.owner_user_id` during password or
Google registration. For older workspaces without that field, the first current
admin by creation time is the legacy notification contact. Operators can set
`owner_user_id` to a verified existing member to remove that ambiguity. Dispatch
rechecks membership and admin role; removed owners are not replaced implicitly
when an explicit owner exists. Suppressed recipients do not receive these emails.

Each state change and its outbox entry commit together. A unique workspace/domain
milestone key prevents repeated DNS checks, retries, or concurrent workers from
duplicating notifications. Existing completed steps are recorded as historical
during migration, so deployment does not email a backlog. Notices expire after
48 hours and eligibility is rechecked before delivery.

## Deployment

Deploy the backend migration **before** the client. It adds `owner_user_id` and
`managed_milestone_emails`; it does not send emails during migration.

Owner notifications use the operator's platform SMTP account, independently of
the customer's managed sender. After checking that account and a controlled test
recipient, enable:

```dotenv
MANAGED_NOTIFICATIONS_ENABLED=true
DASHBOARD_URL=https://app.xem.email
SMTP_HOST=your-platform-smtp-host
SMTP_PORT=587
SMTP_USER=your-platform-smtp-user
SMTP_PASSWORD=your-platform-smtp-password
SMTP_FROM_EMAIL=your-verified-platform-sender
```

Use your deployment secret store for the SMTP password. STARTTLS or implicit TLS
on port 465 is required; certificate checks remain enabled. Notifications are
off unless explicitly enabled. Managed sending itself must also be enabled.

The outbox records `QUEUED`, `SENDING`, `ACCEPTED`, `SKIPPED`, `FAILED`, or
`DELIVERY_UNKNOWN`. Definite pre-acceptance failures retry with backoff, up to five
attempts. Interrupted claims and uncertain DATA acknowledgements are quarantined
as `DELIVERY_UNKNOWN` and are not automatically resent. Operators should inspect
platform SMTP logs using the deterministic Message-ID before considering replay.
`ACCEPTED` in this outbox means platform SMTP accepted the notification; it does
not imply recipient delivery. Disabling the flag stops dispatch without changing
managed sender eligibility.

```sql
SELECT kind, status, count(*)
FROM managed_milestone_emails
WHERE kind <> 'migration'
GROUP BY kind, status;
```

## Editable branded templates

The Templates library includes nine Xem originals under **Transactional**, searchable
by “managed sending” or “onboarding”. Opening one creates an editable copy. The
system notification copy is maintained separately in version control; changing a
workspace copy does not change emails sent to other workspace owners.

The server's `internal/onboardingemails` package generates both HTML/plaintext
notifications and the library's native editor designs. Regenerate after changes:

```sh
cd server
go run ./cmd/onboarding-templates
```

Local automated and browser fixture checks do not establish production delivery
or live AI provider availability. Validate a controlled workspace end to end after
deploying: generate, revise, reopen the saved template, exercise each sender
milestone, and inspect platform SMTP/provider feedback before enabling broadly.
