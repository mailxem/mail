# Connected mail: Gmail, Workspace, IMAP, and Cloudflare

This change adds Google mailbox OAuth and Cloudflare transactional sending to Xem's existing Go backend. It also fixes IMAP pagination, search, folder listing, and message identity. Deploy the backend migration before the client. These connectors do not require a paid Xem entitlement.

## Google mailbox setup

Google mailbox authorization is separate from Google sign-in. Configure a dedicated Google OAuth web client, enable the Gmail API in its Google Cloud project, and register this exact redirect URI:

```text
https://YOUR_XEM_APP/settings/imap/google/callback
```

Set these variables on the **backend**, through your deployment's secret management:

```dotenv
GOOGLE_MAIL_CLIENT_ID=your-oauth-client-id
GOOGLE_MAIL_CLIENT_SECRET=your-oauth-client-secret
GOOGLE_MAIL_REDIRECT_URI=https://YOUR_XEM_APP/settings/imap/google/callback
```

For local development, an exact `http://localhost:PORT/settings/imap/google/callback` URI is allowed. Never put the client secret in a `NEXT_PUBLIC_` variable. Back up the existing installation RSA encryption key; losing it also loses access to stored connection credentials.

1. Sign in as a workspace administrator and open **Settings → IMAP**.
2. Select **Connect Google mailbox** and authorize the intended Gmail or Workspace mailbox.
3. Return to Xem. The connection creates a linked `imap.gmail.com:993` mailbox and `smtp.gmail.com:465` sender.
4. Select the mailbox in **Inbox** and the sender in **Compose**. Replies from a Google mailbox select its linked sender and preserve the parent Message-ID in `In-Reply-To` and `References`.

The connection grants **workspace-wide** mailbox access under existing IMAP permissions. This release does not provide private personal mailboxes, delegated send-as, per-mailbox roles, or shared inbox assignments. The consent screen in Xem states this before authorization. Connect a business mailbox intended for your workspace.

Google's IMAP/SMTP XOAUTH2 flow requires the restricted `https://mail.google.com/` scope. Public hosted distribution needs the applicable Google OAuth verification and security assessment; Workspace organizations may need to allow the application. Self-hosters configure their own OAuth application. Test-mode grants may be short-lived. Google receiving/sending limits still apply. This is not a way to bypass Workspace licensing or bulk-sending restrictions.

OAuth uses PKCE, a ten-minute state bound to the current user and team, and a single-use completion endpoint. The browser receives no access or refresh tokens. Credentials use AES-GCM with a per-secret key wrapped by the installation RSA key and bound to the workspace/connection. Refreshes lock the connection row and persist rotated tokens.

Disconnect removes local credentials and disables both linked configurations. Already submitted messages cannot be recalled. Google account permissions can additionally be revoked in the Google account's third-party access settings; Xem does not revoke the whole Google application grant, which may be shared with another connection.

## Cloudflare sender setup

Enable **Email Sending**, including required domain authentication, in your own Cloudflare account. Email Routing alone is not outbound Email Sending. Create a token with Email Sending permission scoped to that account.

In **Settings → SMTP**, save the Cloudflare account ID, sending token, and sender address. Saving does not send a message or claim domain verification succeeded. Choose this sender explicitly in Compose or pass its `smtpConfigId` to the existing email API. New connections do not replace your workspace's default sender.

```json
{
  "smtpConfigId": "YOUR_CLOUDFLARE_SENDER_ID",
  "to": "a-recipient-you-control@example.com",
  "subject": "Your receipt",
  "html": "<p>Your payment was received.</p>",
  "data": {}
}
```

The endpoint remains `POST /api/v1/emails`. A successful response now means the message was committed to the database outbox. The existing worker's minute-based dispatcher enqueues due messages and recovers from temporary Redis unavailability. It respects `scheduleAt`. Keep the task worker and scheduler running.

Cloudflare currently supports **transactional email only**. Campaigns/newsletters are rejected at message creation and delivery; use a marketing-capable SMTP provider or Xem's existing managed-sending path for them. The API cannot determine the intent of arbitrary user-written content: senders must use this connector for eligible transactional messages.

Cloudflare's documented limits, checked September 28, 2026: 50 combined recipients, 5 MiB message size, 998-character subject, and 16 KiB custom headers. Arbitrary recipients require Workers Paid. The documented account allowance is 3,000 sends/month, then $0.35 per 1,000; Workers plan charges and provider usage are separate from Xem. Recheck the provider links below before quoting prices to customers.

Xem calls Cloudflare's REST API directly. The Go backend, database, and dashboard stay on the existing deployment. No Worker/D1/R2 migration is required. This release does **not** add Cloudflare inbound routing, stored Cloudflare mailboxes, or outgoing attachments.

### Delivery status

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
- `GET /imap/emails?config_id=...&folder=INBOX&offset=0&limit=20&q=receipt` uses IMAP TEXT search. Limits are 1–100; legacy zero-based `page` remains supported. Search totals reflect the filtered results. Empty pages return an empty list with HTTP 200.
- Message IDs combine config, folder, UIDVALIDITY, and UID; `messageId` and the legacy `message_id` both expose the RFC Message-ID. Listing uses `BODY.PEEK` and does not mark messages read.
- `PATCH /imap/flags?config_id=...` accepts `{folder, uid, uidValidity, flag, enabled}` for `\\Seen` or `\\Flagged`. A changed UIDVALIDITY returns 409. No expunge or permanent-delete operation is added.
- Message sizes are fetched before bodies, with a cumulative 25 MiB raw-message budget per response. A smaller page is required for large messages. There is no background full-mailbox mirror or attachment indexing yet.

API keys need existing `imap_configs:read/create` or `smtp_configs:read` permissions for mailbox reads/flags/sender listing. Connection credential management requires a current administrator session and denies API keys.

## Release acceptance

Automated checks cover state expiry/replay/tenant binding, authorization, encryption and credential redaction, pagination/search, recipient outcomes, size/header validation, and ambiguous request handling using local mocks. Browser preview uses sample data and blocks real sending.

Before making a production availability claim, use an authorized test Workspace/Gmail account to check consent, refresh after token expiry, inbox folders, reply threading, and disconnect. Use an enabled Cloudflare domain and recipients you control to verify transactional delivery, queued recipients, and provider logs. These live provider checks require actual deployment configuration and are not implied by local tests.

## Research sources

- [Google Gmail scopes](https://developers.google.com/workspace/gmail/api/auth/scopes)
- [Google IMAP/SMTP OAuth](https://developers.google.com/workspace/gmail/imap/xoauth2-protocol)
- [Cloudflare REST sending](https://developers.cloudflare.com/email-service/api/send-emails/rest-api/)
- [Cloudflare FAQ](https://developers.cloudflare.com/email-service/reference/faq/)
- [Cloudflare limits](https://developers.cloudflare.com/email-service/platform/limits/)
- [Cloudflare pricing](https://developers.cloudflare.com/email-service/platform/pricing/)
- [Mailflare mailer architecture](https://github.com/hieunc229/mailflare/blob/main/server/runtime/mailer.ts)
- [Mailflare mailbox access](https://github.com/hieunc229/mailflare/blob/main/src/lib/mailboxes/access.ts)

Mailflare was studied for its adapter, mailbox-access, inbound relay, and entitlement architecture. Its code was not copied. Xem retains its existing GPL license.
