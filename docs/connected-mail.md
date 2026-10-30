# Connected mail: Gmail, Workspace, IMAP, and Cloudflare

This change adds Google mailbox OAuth, Cloudflare transactional sending, and an optional customer-owned Cloudflare inbound mailbox. It also fixes IMAP pagination, search, folder listing, and message identity. Deploy the backend migration before the client. These connectors do not require a paid Xem entitlement.

## Google mailbox setup

Google mailbox authorization is separate from Google sign-in. Configure a Google OAuth web client, enable the Gmail API in its Google Cloud project, and register this exact redirect URI. An existing web client can be reused; preserve its sign-in redirects and keep the mailbox grant as a separate flow:

```text
https://YOUR_XEM_APP/settings/imap/google/callback
```

Set these variables on the **backend**, through your deployment's secret management:

```dotenv
GOOGLE_MAIL_CLIENT_ID=your-oauth-client-id
GOOGLE_MAIL_CLIENT_SECRET=your-oauth-client-secret
GOOGLE_MAIL_REDIRECT_URI=https://YOUR_XEM_APP/settings/imap/google/callback
GOOGLE_MAIL_ACCESS=testing
GOOGLE_MAIL_TEST_USERS=authorized-tester@example.com
```

For local development, an exact `http://localhost:PORT/settings/imap/google/callback` URI is allowed. Never put the client secret in a `NEXT_PUBLIC_` variable. Back up the existing installation RSA encryption key; losing it also loses access to stored connection credentials.

Credentials alone do not enable public mailbox connections. The default `testing` mode requires an exact, case-insensitive address in `GOOGLE_MAIL_TEST_USERS`, for both the initiating Xem user and the mailbox Google returns. Unknown modes and `disabled` deny new grants. Set `GOOGLE_MAIL_ACCESS=public` only after applicable Google verification is complete; self-hosters using their own OAuth application can explicitly choose `self-hosted`. This gate controls new connections, not revocation of existing grants. See [the verification checklist](google-mail-verification.md).

1. Sign in as a workspace administrator and open **Settings → IMAP**.
2. Select **Connect Google mailbox** and authorize the intended Gmail or Workspace mailbox.
3. Return to Xem. The connection creates a Gmail API inbox and linked `smtp.gmail.com:465` sender.
4. Select the mailbox in **Inbox** and the sender in **Compose**. Replies from a Google mailbox select its linked sender and preserve the parent Message-ID in `In-Reply-To` and `References`.

The connection grants **workspace-wide** mailbox access under existing IMAP permissions. This release does not provide private personal mailboxes, delegated send-as, per-mailbox roles, or shared inbox assignments. The consent screen in Xem states this before authorization. Connect a business mailbox intended for your workspace.

The current OAuth grant retains the full `https://mail.google.com/` scope because SMTP sending still uses XOAUTH2. Inbox reads and read/starred changes use Gmail's HTTPS API and do not open an IMAP connection. Public hosted distribution needs the applicable Google OAuth verification and security assessment; Workspace organizations may need to allow the application. Self-hosters configure their own OAuth application. Test-mode grants may be short-lived. Google receiving/sending limits still apply. This is not a way to bypass Workspace licensing or bulk-sending restrictions.

Gmail lists fetch metadata and snippets with at most 20 messages per page using Gmail's opaque `page_token`; full message bodies load only when a message is opened. Attachments remain metadata until explicitly downloaded, and each attachment download is limited to 10 MiB. Xem exposes a stable MIME part selector for each attachment and resolves Google's current attachment ID only at download time because Gmail can rotate that provider ID between reads. Gmail totals are estimates. Label IDs remain the request identifiers even when Xem shows friendly display names.

Foreground refresh checks use Gmail history IDs and reset the current listing if Google reports that a saved history baseline is stale. They skip hidden or offline views and do not overlap. Pub/Sub watch delivery and durable background mailbox synchronization are future infrastructure and are not implemented here. Switching mailbox or search context starts a new pagination and history baseline.

OAuth uses PKCE, a ten-minute state bound to the current user and team, and a single-use completion endpoint. The browser receives no access or refresh tokens. Credentials use AES-GCM with a per-secret key wrapped by the installation RSA key and bound to the workspace/connection. Refreshes lock the connection row and persist rotated tokens.

Disconnect removes local credentials and disables both linked configurations. Already submitted messages cannot be recalled. Google account permissions can additionally be revoked in the Google account's third-party access settings; Xem does not revoke the whole Google application grant, which may be shared with another connection.

## Cloudflare sender setup

Enable **Email Sending**, including required domain authentication, in your own Cloudflare account. Email Routing alone is not outbound Email Sending. Create a token with Email Sending permission scoped to that account.

In **Settings → SMTP**, save the Cloudflare account ID, sending token, and sender address. Saving does not send a message or claim domain verification succeeded. Choose this sender explicitly in Compose or pass its `smtpConfigId` to the existing email API. New connections do not replace your workspace's default sender.

```json
{
  "smtpConfigId": "YOUR_CLOUDFLARE_SENDER_ID",
  "requestId": "87f7e850-33b9-4f15-9b42-9aabdb0daee2",
  "to": "a-recipient-you-control@example.com",
  "subject": "Your receipt",
  "html": "<p>Your payment was received.</p>",
  "data": {}
}
```

The endpoint remains `POST /api/v1/emails`. A successful response now means the message was committed to the database outbox. The existing worker's minute-based dispatcher enqueues due messages and recovers from temporary Redis unavailability. It respects `scheduleAt`. Keep the task worker and scheduler running.

The response includes the persisted message `id`. Use a new UUID `requestId` for each intended message and reuse the identical key and payload when retrying an uncertain request. The key is scoped to the workspace. Replays return the original receipt even after sending, soft deletion, a sender disconnect, or a changed default sender; a changed payload returns HTTP 409. This deduplicates saving to Xem's outbox, not provider delivery. It never retries an ambiguous SMTP or Cloudflare submission.

Outgoing attachments are available in Compose and through `attachments: [{filename, content, contentType}]`, where `content` is standard base64. Up to ten files and 3 MiB of decoded data are allowed in total. Names cannot contain paths or hidden/control characters; MIME types cannot contain parameters. Bytes are stored with the durable email row, attached through SMTP or Cloudflare, and exposed for download in Outbox. Provider message-size limits still apply. Arbitrary remote attachment URLs and filesystem paths are not accepted.

Cloudflare currently supports **transactional email only**. Campaigns/newsletters are rejected at message creation and delivery; use a marketing-capable SMTP provider or Xem's existing managed-sending path for them. The API cannot determine the intent of arbitrary user-written content: senders must use this connector for eligible transactional messages.

Cloudflare's documented limits, checked September 28, 2026: 50 combined recipients, 5 MiB message size, 998-character subject, and 16 KiB custom headers. Arbitrary recipients require Workers Paid. The documented account allowance is 3,000 sends/month, then $0.35 per 1,000; Workers plan charges and provider usage are separate from Xem. Recheck the provider links below before quoting prices to customers.

Xem calls Cloudflare's REST API directly for outbound sending. That sender does not require a Worker, D1, or R2.

## Customer-owned Cloudflare inbound mailbox

The optional inbound mailbox is a separate deployment. Cloudflare Email Routing invokes a Worker, private R2 stores raw MIME, bodies, and attachments, and D1 stores the mailbox index and flags. Xem's backend accesses it through a signed API; the browser receives neither the mailbox secret nor a Cloudflare API token.

Follow [the mailbox setup guide](https://github.com/mailxem/mail/blob/undefined/devops/cloudflare-mailbox/README.md). Apply its D1 migration and the Xem backend migration before connecting the mailbox or switching routes in the UI. Each deployment accepts one exact `MAILBOX_ADDRESS`, rejects raw messages larger than 10 MiB, and does not automatically delete received or quarantined mail. This implementation has local tests and a Wrangler dry run; a production claim still requires an account-specific deployment and live Email Routing validation.

## Outbound delivery status

| Xem status         | Meaning                                                                                                                                        |
| ------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| `SENT`             | Cloudflare reported all recipients delivered, or the SMTP server acknowledged submission. SMTP acknowledgement is not proof of inbox delivery. |
| `ACCEPTED`         | Cloudflare queued at least one recipient and reported no permanent bounces. Final delivery remains unconfirmed.                                |
| `PARTIAL`          | Cloudflare reported a mix of accepted/delivered recipients and permanent bounces. Do not resend the whole message.                             |
| `BOUNCED`          | Cloudflare reported every recipient permanently bounced.                                                                                       |
| `DELIVERY_UNKNOWN` | Request interrupted, provider server error, malformed response, or incomplete recipient results; review with the provider before retrying.     |

The raw recipient groups are stored as `providerResult` on the email record. This release does not subscribe to Cloudflare delivery events; queued outcomes remain `ACCEPTED` until an operator has external evidence. Terminal and ambiguous states are not automatically resent. A worker crash during submission can leave `SENDING`; investigate before resetting it.

## IMAP changes and compatibility

- Verified TLS 1.2 or newer. Port 143 upgrades with STARTTLS; other ports use implicit TLS. No insecure certificate bypass.
- Public destinations by default, checked after DNS resolution. `ALLOW_PRIVATE_IMAP=true` is an explicit operator opt-in for internal servers, while TLS validation still applies.
- `GET /imap/folders?config_id=...` returns the existing array of folder objects, now consumed correctly by the client. Nonselectable folders are omitted from the picker.
- For generic IMAP, `GET /imap/emails?config_id=...&folder=INBOX&offset=0&limit=20&q=receipt` uses IMAP TEXT search. Limits are 1–100; legacy zero-based `page` remains supported. Search totals reflect the filtered results. Empty pages return an empty list with HTTP 200. Gmail connections instead use the native metadata and page-token flow described above.
- Stable infinite scrolling can omit `offset` and pass the prior response's `next_before_uid` back as `before_uid`. Cursor pages contain only UIDs below that value; `total_emails` remains the total count before cursor filtering, and `next_before_uid` is omitted when exhausted. A cursor cannot be combined with a nonzero `offset` or `page`.
- For generic IMAP, `GET /imap/head?config_id=...&folder=INBOX&q=receipt` supports lightweight foreground polling. It performs a read-only UID search without fetching message bodies and returns `total_emails`, `uidValidity`, and the highest matching `latest_uid` (zero for an empty result). Gmail uses bounded history checks instead.
- Message IDs combine config, folder, UIDVALIDITY, and UID; `messageId` and the legacy `message_id` both expose the RFC Message-ID. Listing uses `BODY.PEEK` and does not mark messages read.
- `PATCH /imap/flags?config_id=...` accepts `{folder, uid, uidValidity, flag, enabled}` for `\\Seen` or `\\Flagged`. A changed UIDVALIDITY returns 409. No expunge or permanent-delete operation is added.
- Message sizes are fetched before bodies, with a cumulative 25 MiB raw-message budget per response. A smaller page is required for large messages. There is no background full-mailbox mirror or attachment indexing yet.

API keys need existing `imap_configs:read/create` or `smtp_configs:read` permissions for mailbox reads/flags/sender listing. Connection credential management requires a current administrator session and denies API keys.

Legacy SMTP/IMAP passwords are write-only, including when configurations are loaded through related records. Leaving the password blank while editing or testing an unchanged login endpoint keeps the stored password. Changing host, port or username requires re-entering it, so an edit cannot forward an existing secret to another endpoint.

## On-demand AI summaries

Inbox summaries are disabled by default. To offer them, configure the client server with `XEM_MAIL_SUMMARY_ENABLED=true` and the existing assistant settings: `XEM_ASSISTANT_ENABLED`, `AI_PROXY_BASE_URL`, `AI_PROXY_MODEL`, and `AI_PROXY_API_KEY`. Production also requires `ASSISTANT_REDIS_URL`; Xem uses Redis for shared rate limits, the daily team allowance, and a lock that prevents overlapping assistant work across replicas. Keep all of these variables server-only.

A signed-in user must select **Summarize** for a specific inbox message. The browser sends the mailbox configuration ID and folder plus either the Gmail provider message ID or the generic IMAP UID and UIDVALIDITY. The Xem server retrieves that exact message through the authenticated, workspace-scoped read endpoint and verifies the returned identity before sending bounded visible body text plus the From, To, Cc, Date, and Subject headers to the configured AI provider. Attachment bytes, remote images, and linked resources are not sent or fetched. Long visible bodies are shortened and the result says when this happened.

The summary response is returned to the requesting browser and is not written to Xem's assistant conversation store. Normal infrastructure and provider processing may still apply, so operators must review their AI provider's data handling, recipient-data restrictions, regional requirements, retention terms, and acceptable-use policies before enabling the feature. Xem does not determine whether mailbox participants consented to third-party AI processing.

The summary model has no tools and cannot send, reply, change flags, follow links, or modify workspace data. Email content is handled as untrusted quoted data, and output is limited to a summary, supported key points, and explicit action items. These controls reduce prompt-injection risk but do not guarantee factual model output; users should verify summaries against the original message.

`XEM_MAIL_SUMMARY_ENABLED` is separate from `GOOGLE_MAIL_ACCESS`. Keep Google mailbox access in `testing` or `disabled` until the applicable Google verification is complete. Enabling summaries does not make Google authorization production-ready, and enabling Google mailbox access does not enable summaries.

## Root-admin email templates

On backend startup, every workspace containing a super admin receives saved, editable copies of the 18 branded Xem system emails: nine managed-sending onboarding messages and nine welcome, account-security, and delivery-alert messages. This also backfills existing root accounts; bootstrap credentials do not need to be set again. Open **Templates** after the updated backend starts.

Seeding uses stable identities, preserves existing edits and intentionally deleted templates, and does not populate ordinary workspaces or send email. These are library copies: editing them does not override the canonical templates used by automatic system notifications.

If the Transactional category was intentionally deleted, startup reports a seeding warning and preserves that deletion. Create or restore an active Transactional category and restart the backend to populate missing templates.

## Release acceptance

Automated checks cover state expiry/replay/tenant binding, authorization, encryption and credential redaction, pagination/search, recipient outcomes, size/header validation, and ambiguous request handling using local mocks. Browser preview uses sample data and blocks real sending.

An authorized own-account Gmail acceptance run verified consent, refresh-token use, SMTP send, Gmail API receive, reply headers, two unique 20-message cursor pages, exact attachment bytes through both the API and UI, all four read/unread and star/unstar actions, and stale-history reset. A real Gmail flag update was detected by history polling: the generic mailbox-change banner appeared after 27 seconds, and refreshing showed the changed message. This verifies change detection for a flag update; it is not evidence of a new-arrival-specific banner. In that run, the first 20-message metadata page took about 2.2 seconds and transferred about 16 KiB; an opened message detail took about 0.3 seconds. These measurements describe one test account and network, not a latency guarantee.

Disconnect returned HTTP 204, cleared the stored secret, made the linked inbox and sender inactive, and caused subsequent reads to return HTTP 409 and sending attempts to return HTTP 400. Reconnecting through the real Google **Allow** screen produced one HTTP 201 completion, a success screen with a clean callback URL, and a 20-message inbox. Real Gmail API UI checks covered mobile and desktop layouts in light and dark themes with no horizontal overflow. Screenshots and acceptance records must not include private mailbox content or provider identifiers.

Before making a production availability claim, complete applicable Google public verification and security assessment. Use an enabled Cloudflare domain and recipients you control to verify transactional delivery, queued recipients, and provider logs. If summaries are enabled, separately verify the configured AI provider with an authorized test message and review its output against the source. The browser preview uses mock summary data; this acceptance run did not contact Cloudflare or an AI provider. No production deployment or live Cloudflare delivery is implied by the Gmail acceptance run.

### Isolated local provider checks

`go run ./cmd/mailcheck` from `server` starts the real mail routes, worker and scheduler without requiring object storage. Configure the normal backend environment, a random `JWT_SECRET` of at least 32 characters, an installation RSA key, local Redis, and a dedicated PostgreSQL database with a name beginning `xem_mailcheck_`. `POSTGRES_HOST` must be `127.0.0.1`. The API binds only to loopback. `MAILCHECK_RECIPIENT` must be one address you control; the delivery boundary rejects every other recipient, including CC and BCC.

Set `SUPERADMIN_EMAIL`, `SUPERADMIN_PASSWORD`, `SUPERADMIN_NAME` and `SUPERADMIN_TEAM_NAME` for the isolated local workspace. Keep these values and the OAuth client secret in protected local configuration outside Git. Point the client API URLs at this runner and register its exact callback URI with Google. This runner is for provider acceptance, not a production deployment or full storage/marketing-runtime test.

Database connection values are quoted, including empty passwords, and the backend verifies `current_database()` before running migrations. SQL logging omits parameter values to avoid recording credentials and message bodies.

## Research sources

- [Google Gmail scopes](https://developers.google.com/workspace/gmail/api/auth/scopes)
- [Gmail API message listing](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.messages/list)
- [Gmail API message retrieval](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.messages/get)
- [Gmail API history listing](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.history/list)
- [Gmail API attachment retrieval](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.messages.attachments/get)
- [Google IMAP/SMTP OAuth](https://developers.google.com/workspace/gmail/imap/xoauth2-protocol)
- [Cloudflare REST sending](https://developers.cloudflare.com/email-service/api/send-emails/rest-api/)
- [Cloudflare FAQ](https://developers.cloudflare.com/email-service/reference/faq/)
- [Cloudflare limits](https://developers.cloudflare.com/email-service/platform/limits/)
- [Cloudflare pricing](https://developers.cloudflare.com/email-service/platform/pricing/)
- [Mailflare mailer architecture](https://github.com/hieunc229/mailflare/blob/main/server/runtime/mailer.ts)
- [Mailflare mailbox access](https://github.com/hieunc229/mailflare/blob/main/src/lib/mailboxes/access.ts)

Mailflare was studied for its adapter, mailbox-access, inbound relay, and entitlement architecture. Its code was not copied. Xem retains its existing GPL license.
