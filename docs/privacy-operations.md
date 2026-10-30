# Privacy publication and Google verification readiness

Status: **draft; not a deployed retention policy or completed assessment**. The user selected `team@xem.email` and asked for concrete commitments. The website drafts propose 30-day verified deletion and 90-day residual backup expiry. Do not publish these as effective promises until the controls below have evidence and an accountable operator.

This is a hosted-service policy. Self-hosted operators control their own installations and need their own policy. The open-source license does not automatically bind independent operators to Xem's hosted retention promises.

## Commitments and present evidence

| Proposed commitment | Present implementation | Required before policy publication |
| --- | --- | --- |
| Remove Google credentials on disconnect | `MailConnectionsHandler.Disconnect` blanks `secret` and disables linked SMTP/IMAP configs in one transaction. Token refresh holds a row lock and cannot restore a disconnected connection. | Real disconnect/reconnect validation; cancellation of pending sends; restore procedure that does not resurrect old grants. In-flight requests can finish. |
| Delete verified requested data from active systems within 30 days | Generic model deletion is soft deletion. Disconnect retains the address/config references and outbox content. A comprehensive purge is not implemented. | Tested deletion procedure covering the complete subject/workspace graph, files, queues, caches, logs, integrations and provider requests; named request owner and deadline tracking. |
| Residual backups expire within 90 days after active-system deletion | No deployment-wide retention evidence collected. Redis conversation TTL is not backup expiry. | Inventory all snapshots, replicas, dumps, object versions, AOF/RDB files and manual exports; enforce bounded retention; test a restore and reapply completed deletions before reopening service. |
| Assistant history expires seven days after last save | `client/lib/assistant/store.ts` uses seven-day expiry and deletes active conversations on request. | Verify deployed Redis settings, expired data visibility and backup/provider handling. Conversation indexes, audit rows and provider logs are distinct. |
| No sale, ads or general AI training with Google data | No training pipeline identified, but messages/drafts and outgoing-mail tool results can reach an external AI proxy. Upstream provider and contract unknown. | Identify every processor, intermediary and model host; confirm compatible purpose limits, no-general-training terms, retention and deletion. If unavailable, disable affected Google-data-to-AI paths end to end, including existing history/tool results. A UI warning or prompt instruction alone is insufficient. |
| Human access is purpose-limited | Code enforces workspace permissions; production operator access was not audited. | Least-privilege operator access, documented message-specific support consent, access logs, security-incident procedure and staff/provider restrictions. |

Do not use these proposed windows as an excuse to retain data longer than necessary or where applicable law requires an earlier response. Any exception needs a specific reason, restricted access and an end/review date.

## Data inventory to cover

The two purposes of the Google connector are SMTP sending and IMAP receiving/reading. Ordinary sign-in is separate. The following are code-backed data paths, not a claim that production has been inspected:

- **Credentials and identity:** `mail_connections` retains mailbox address, provider, encrypted secret, sender/inbox IDs and team. OAuth state contains user/team, a hashed state and PKCE verifier until consumption/expiry cleanup. Disconnect clears the secret but does not delete the record.
- **Incoming mail:** `IMAPHandler` fetches messages and attachments on demand and returns them to the browser; it does not mirror the entire inbox to PostgreSQL. Browser memory, caches, downloads and explicitly copied content are separate from database storage.
- **Outgoing mail:** `models.Email` includes body, subject, recipients, attachments and related delivery/status information. Follow campaign/newsletter/tracking/bounce relationships when identifying Google-derived data. Do not infer provenance solely from a currently active sender: disconnected configurations and copied content can remain.
- **AI:** writing posts instructions, subject and draft body to `marketing/email-draft`; chat sends user text and relevant MCP tool results to the configured proxy. API keys issued for assistant use include outgoing-email access. The configured model name does not identify all processors. Tool and conversation records may contain Google-derived data.
- **Copies:** database replicas/snapshots/dumps, queue contents, object storage and versions, Redis/AOF/RDB, support tickets, provider logs, webhook payloads and exported files all require inventory. The seven-day SES MIME cleanup is not a general message-retention policy.

## Request handling procedure to validate

1. Receive requests at `team@xem.email`. Record a request ID, received date, requested scope and responsible operator; avoid copying message bodies into the ticket.
2. Verify identity and authority using the signed-in account or proportionate confirmation. Do not ask for passwords, client secrets or a full inbox export. Resolve shared-workspace ownership and preserve other members' rights. Offer an export before irreversible deletion if requested.
3. Confirm exactly which account, connection or workspace records are affected. Map references and derived data before deletion. Log only identifiers/counts necessary for evidence. A personal-account deletion must not silently destroy a team's workspace.
4. Disconnect relevant providers, stop pending automations/sends and invalidate affected grants. Revoke Google's grant only with an explicit understanding of the other clients/sign-in connections it affects. Check already-running work and prevent new writes during deletion.
5. Apply the tested purge to active databases, files and caches. Soft-deleted rows must be included; deleting a flag or UI entry is insufficient. Handle linked provider/processor requests and any required limited exceptions. **A reviewed purge implementation and a synthetic fixture test are still required; this document is not an executable purge.**
6. Verify absence through both storage queries and authenticated API/UI reads. Confirm another workspace's fixtures remain intact. Record completion date and affected record counts without retaining deleted content as "evidence."
7. Record only the minimum deletion manifest needed to prevent resurrection. Restrict it and set its own retention. Track the final expiry date for every backup class; manifests/restore gates must cover encrypted tokens as well as message content. Test the longest-retained backup in isolation.
8. Reply with what was removed, backup-expiry date, any narrowly retained categories and external-provider limitations. Close only when the active-system obligation is met; keep backup expiry as a tracked obligation.

## Publication gate

- [ ] Verify the operator identity for the public notice and that `team@xem.email` is monitored and usable as Google's support contact.
- [ ] Inventory deployed processors and countries/regions as applicable; disclose actual AI providers and handling terms. Do not claim consent alone overrides Limited Use.
- [ ] Validate hosted storage encryption, access control, diagnostics/log content, cookies, infrastructure providers, security reporting and limited human access.
- [ ] Implement and exercise the deletion procedure with synthetic cross-workspace fixtures, disconnected senders, soft-deleted rows, attachments, assistant history and an in-flight send.
- [ ] Enforce and test the backup window, including object versioning, copied snapshots and post-restore deletion replay. Avoid production purge experiments.
- [ ] Assign owners, deadline tracking and request verification for the 30-day commitment. Document bounded legal/security exceptions.
- [ ] Revise the draft copy to describe the exact implemented process; resolve any remaining placeholders. Remove the draft notice/noindex and add an effective date in a reviewed change. Add finalized public pages to the sitemap.
- [ ] Deploy the policy pages through the normal website release; verify HTTPS, unauthenticated access, footer links, readable mobile/desktop layout, and correct canonical URLs.
- [ ] Keep Google public access disabled until brand review, scope review, the genuine demo, applicable annual independent assessment and Google approval are complete.

## Evidence boundaries

Policy drafts, code tests and a security checklist cannot substitute for an assessor's Letter of Assessment or Google's approval. No retention infrastructure, public policy, OAuth console setting or hosted availability is changed by this document. The privacy notice is a reviewable proposal until the gate is completed.

Related: [Google verification packet](google-mail-verification.md), [connected-mail implementation](connected-mail.md), [assistant data handling](https://github.com/mailxem/mail/blob/undefined/client/docs/assistant.md).
