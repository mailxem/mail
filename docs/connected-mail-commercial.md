# Connected mail commercial plan

Decision: offer **hosted subscriptions and self-hosted commercial licenses**. Existing free features keep working. This document is a product and pricing proposal, not an available price list or a set of implemented paid features.

## What customers pay for

The free connectors make Xem useful and give developers a reason to star, deploy, and contribute. Revenue comes from saving people time, coordinating teams, and operating business email reliably. A customer bringing Cloudflare or Google credentials pays that provider directly; Xem charges for its software, hosting, and support value. Do not promise unlimited sending or count provider allowances as Xem-owned credits.

| Package   | Proposed new value                                                                                                                               | Hosted price to test                                                        | Self-hosted price to test                                        |
| --------- | ------------------------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------- | ---------------------------------------------------------------- |
| Community | Existing free features, secure IMAP, Gmail/Workspace connection, Cloudflare transactional sender, basic composer and inbox                       | Existing free plan and limits                                               | Free GPL core                                                    |
| Pro       | Saved replies/signatures, snooze and follow-up reminders, mailbox rules, optional metered AI assistance                                          | $19/workspace/month, 1 seat                                                 | $99/year, 1 production installation, 1 seat                      |
| Team      | Pro plus private/shared mailboxes, per-mailbox read/send/delegate permissions, assignments, internal notes, collision warnings, shared templates | $59/workspace/month including 5 seats; test $8/additional seat              | $299/year including 5 seats; test $30/additional seat/year       |
| Company   | Team plus SSO, provisioning, approval policies, audit export, retention controls, and contracted support                                         | From $199/workspace/month including 20 seats, quote for service commitments | From $999/year including 20 seats, quote for service commitments |

All new paid capabilities above are **roadmap items**. Existing collaboration or automation features cannot be reclassified as paid retroactively. Cloudflare inbound routing is also a separate roadmap item, not an included capability in this patch. AI credits, storage, retained history, and support scope must have explicit allowances before selling any tier.

Start with recurring annual self-hosted licenses, not a lifetime promise that must fund indefinite security updates and support. Include a staging installation for validation; define recovery/migration between production installations without charging a second time. A paid setup service can be a separate fixed-scope offer after a repeatable setup runbook and labor-cost estimate exist.

## Revenue and margin checks

Illustrative mix, not traction or a forecast: 50 Pro + 20 Team + 5 Company hosted customers at the proposed base prices produces **$3,125 monthly recurring revenue**. Self-hosted licenses are tracked as annual recurring revenue separately; do not present their full annual receipts as monthly revenue.

For each tier, measure hosting, storage, payment fees, support time, and included AI costs. Provider send charges are separate when customers bring credentials. Use a 75% gross-margin target as a planning constraint, not a claim; expand included usage only after measuring its cost. Google verification/security-review expense belongs in the launch budget. No price or checkout product is changed by this proposal.

## Entitlements for both purchase paths

The current dashboard uses `payments.go` through the client billing integration. The Go backend also has older subscription models; these are not interchangeable. Build one server-side entitlement resolver backed by the authoritative billing service, rather than treating a browser-selected plan or the legacy `HasFeature` helper as authorization.

Recommended contract:

```text
Entitlement {
  workspace_id, source: hosted | license,
  plan: pro | team | company,
  features, seat_limit, installation_id,
  valid_from, paid_through, grace_until, revision
}
```

- Hosted: verify billing webhooks, use idempotent event IDs, reject stale revisions, and fetch current subscription state server-side when needed. Grant only paid/trial entitlements within their valid window; canceled subscriptions retain access through the paid period. Refund, suspension, and failed-renewal policies need explicit states.
- Self-hosted: issue an Ed25519-signed entitlement for a customer installation and workspace after successful purchase. Keep the signing key only in the licensing service. The installation stores a public verification key and verifies signature, issuer, audience, expiry, feature allowlist, seat limit, and installation binding. A pasted license string is never sufficient without verification.
- Renewal: propose a 30-day grace window with clear owner notices, then stop new paid automation/administration actions. Keep basic mail access, export, existing free capabilities, and customer data intact. Never delete mail as a billing action. Verify time-bound entitlements locally so a temporary billing-service outage does not immediately break a paid customer's work.
- Enforce limits in API handlers and background workers, inside the same transaction that creates a seat, delegation, or paid rule. The frontend mirrors these decisions but does not enforce them alone.
- Keep license activation separate from mailbox OAuth. Support must not require customers to disclose email content, API tokens, or Google refresh tokens.

Commercial licensing must respect the existing GPL distribution and contributors' rights. Do not retroactively restrict the public core. Before distributing proprietary paid modules, establish ownership and license compatibility for their exact code and integration. Paid maintenance/support and hosted service contracts can fund GPL software; a runtime entitlement check does not change recipients' GPL rights or make public code proprietary. No repository license changes are made here.

## Build and launch sequence

1. Validate this connector foundation against real authorized provider accounts; complete Google verification appropriate to the hosted service. Keep the public integration labels in preview until then.
2. Deliver and test the first Pro workflow (saved replies and follow-up reminders), then the first Team workflow (private/shared mailbox access plus assignment). Entitlements and authorization must land with the paid action, not as a cosmetic upgrade button.
3. Add authoritative billing events, signed self-hosted issuance/activation/renewal, license-key rotation, seat checks, and cancellation/grace tests. Test hosted and self-hosted purchases independently before opening checkout.
4. Publish the capability page and a short real product demo. Enable only tiers whose advertised features work. Company remains contact-sales until its controls and support commitments are deliverable.
5. Measure repository visits → stars/deployments → connected mailbox → first successful send → weekly use → upgrade. Track hosted trials and self-hosted license activations separately. Use aggregate product events; never collect message content for marketing analytics.

## Positioning and campaign drafts

Lead: **Your inbox, your infrastructure. A workspace your team can grow into.**

Supporting copy, after provider verification: “Connect Gmail or Google Workspace, bring an IMAP mailbox, and send transactional email through your own Cloudflare account. Start with Xem's open-source core. Choose managed hosting or keep the deployment on your own servers.”

Paid follow-on, only as features launch: “Upgrade for the work around the inbox: saved replies and follow-ups for individuals, shared ownership for teams, and access controls for companies.”

Use two primary calls to action: **Start with GitHub** and **Try hosted Xem**. Put “Found a missing provider or a rough edge? Contributions welcome.” beside the technical setup guide. The repository remains a useful entry point, not a disguised license checkout.

Draft founder post, suitable while this is in development:

> We're building connected mail into Xem: Gmail/Workspace inboxes, safer IMAP, and Cloudflare for transactional sending. The core stays open source. We also plan paid Pro, Team, and Company options, available hosted or self-hosted, for the work around the inbox. Follow the implementation on GitHub and tell us which team workflow costs you the most time today.

Draft technical post: “What Cloudflare Email Sending does—and why it doesn't replace your marketing sender.” Explain routing versus sending, bring-your-own account costs, recipient outcomes, and Google's separate mailbox consent. Link to official limits. Avoid competitor-price comparisons from unverified store copy and avoid claiming this is a free replacement for Google Workspace.

Keep launch messages as drafts until live provider acceptance and the advertised paid workflows are complete. No public posts or marketing emails are sent by this change.
