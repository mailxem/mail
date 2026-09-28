---
title: "Managed SMTP, your domain: how Xem sends through Amazon SES"
description: "Xem supports managed SMTP through Amazon SES. Learn how domain verification, workspace controls, the delivery queue, and feedback fit together."
date: "2026-09-12"
updated: "2026-09-22"
category: "Deliverability"
tags:
  ["managed SMTP", "custom email domains", "Amazon SES", "email infrastructure"]
cover: "/images/blog/managed-smtp-custom-domains.webp"
coverAlt: "Original Xem illustration: Your domain. A clear path. An envelope travels through Xem’s queue to Amazon SES, with feedback returning to Xem."
coverSource: "https://xem.email/blog/editorial"
author: "Xem editorial"
featured: true
---

You should be able to send a newsletter from your own domain without first becoming an email infrastructure operator. Connect the domain, understand the checks, send yourself a test, and start writing to the people who asked to hear from you.

Xem supports managed sending through Amazon Simple Email Service (SES). SES handles delivery to receiving mail servers. Xem adds workspace controls, domain setup, a durable submission queue, and delivery feedback.

**Availability update — September 22, 2026:** managed SMTP through SES is available in Xem. Complete domain verification and workspace approval before sending, and check the allowances shown in your workspace. You can also continue using your own SMTP provider. This updates the original September 12 article’s pre-launch status; the architecture and status distinctions below explain the sending path.

The illustrations below explain the design. They are not customer traffic, benchmarks, or measured inbox-placement results.

## Your address, your domain, your existing inbox

Imagine your team owns `example.com` and wants to send from `hello@example.com`. That visible From address is your identity. The SMTP hostname is where another application submits a message. Those names do not have to match.

The managed flow uses one configured Xem submission hostname, with credentials scoped to a workspace and verified domain. It does not create a branded `smtp.example.com` endpoint for every customer; that would need separate hostname and certificate management, not just a DNS alias.

It also does not create an inbox. Your mailbox can stay with Google Workspace, Microsoft 365, iCloud, or another provider. Keep its existing receiving MX records. Replies can go to that mailbox through your Reply-To address.

![Three addresses for one brand: hello@example.com is the visible sender, bounce.example.com handles delivery failures through SES, and replies stay with the existing mailbox. All addresses are examples.](/images/blog/managed-smtp-domain-map.svg)

The dedicated `bounce.example.com` subdomain is the **custom MAIL FROM domain**, used for the envelope sender and delivery failures. Its SES MX record does not replace the MX records that receive ordinary mail for `example.com`. AWS explains this distinction in its [custom MAIL FROM guide](https://docs.aws.amazon.com/ses/latest/dg/mail-from.html).

## Why Amazon SES sits behind the send button

[Amazon SES](https://docs.aws.amazon.com/ses/latest/dg/Welcome.html) provides delivery infrastructure, domain authentication, sending APIs, and delivery feedback. Using it lets us concentrate on the product and operating controls around a send.

Operating a direct-delivery mail server would put more work on our team: reverse DNS, IP reputation, deferrals, blocklist investigations, and delivery queues across remote servers. Established open-source mail servers can do valuable work here. Installing one does not give a new operator an established sending reputation.

Our implementation uses the **SES v2 API through the AWS SDK for Go**. External tools can speak SMTP to Xem; Xem submits the accepted message to SES through the API. These are two separate connections.

Customers receive Xem credentials, not shared AWS SMTP keys. Every submission goes through Xem’s policy checks before reaching SES. The deployment runbook calls for a scoped workload IAM role, rather than an administrator’s credentials on the application server.

SES is an infrastructure choice, not an inbox-placement guarantee. Xem remains responsible for operating the service. Senders remain responsible for consent, relevant content, and audience quality.

## Follow a message through the system

![Message flow: a campaign or authenticated SMTP client enters Xem’s policy checks and PostgreSQL outbox, then goes through the SES API to a recipient server. Signed feedback returns through Amazon SNS to delivery history and suppression checks.](/images/blog/managed-smtp-flow.svg)

1. **Submit.** Campaigns use Xem’s internal sending path. External tools use authenticated SMTP with STARTTLS on the configured public port, 587. The listener uses the open-source `emersion/go-smtp` library and rejects plaintext authentication and unauthenticated relay.
2. **Check.** Xem verifies workspace approval, domain readiness, sender identity, credentials where applicable, recipient policy, and remaining allowance. The SMTP envelope and visible sender must agree with the allowed domain.
3. **Store.** PostgreSQL commits an outbox entry before Xem returns successful SMTP acceptance. Workers claim queued entries with row locks so concurrent workers do not take the same entry.
4. **Deliver.** The worker checks eligibility again before calling SES. The request identifies the workspace’s SES tenant and configuration set, plus a Xem message identifier for matching feedback.
5. **Listen.** SES publishes events to Amazon SNS. Xem authenticates the notification, records the outcome, and uses permanent bounces and complaints to suppress later sends.

PostgreSQL is the durable queue for this managed path. Acceptance does not depend on an in-memory handoff surviving a restart. Normal database operations still matter: backups, access controls, monitoring, and recovery procedures.

## Domain checks with a reason behind each record

Proving that you control a DNS name is necessary, but it is not the same as proving that outgoing mail is configured correctly. The new setup flow separates these checks.

| Check                     | What it establishes                                                                                           |
| ------------------------- | ------------------------------------------------------------------------------------------------------------- |
| Xem ownership TXT         | The workspace can publish a challenge on the requested domain.                                                |
| SES identity verification | SES recognizes the sending identity in the configured region.                                                 |
| DKIM                      | Recipients can validate signatures for the domain; new identities use 2048-bit keys.                          |
| Custom MAIL FROM          | The bounce subdomain has the SES MX and SPF records needed for the envelope sender.                           |
| DMARC                     | The sending domain publishes a policy connecting authentication to the visible From domain through alignment. |
| Workspace approval        | An operator has reviewed the sending use case and assigned limits.                                            |

The interface provides exact records to publish with your authoritative DNS provider, whether that is Cloudflare or another service. Preserve existing receiving MX and DMARC policies. DKIM CNAMEs should resolve as DNS records, without an HTTP proxy.

The first implementation requires an explicit DMARC record at the exact sending domain, including a sending subdomain. It does not yet implement organizational-domain DMARC inheritance. That is a Xem limitation, not a requirement of DMARC itself.

We configure SES to reject messages if the custom MAIL FROM setup fails, rather than silently fall back to a provider MAIL FROM domain. A `p=none` DMARC policy is a starting point for observation, not enforcement against spoofed mail. Authentication and alignment still need checking on real messages before launch.

## The controls Xem adds around SES

**Credentials with a limited purpose.** SMTP secrets are shown once, stored as bcrypt hashes, expire after 90 days, and can be revoked. They belong to one workspace and domain. They do not grant access to AWS or the management API. The listener also limits connections, authentication attempts, message size, recipients, and connection duration.

**Workspace separation.** Each workspace gets an [SES tenant](https://docs.aws.amazon.com/ses/latest/dg/tenants.html), associated resources, and a configuration set for event publication. Xem enforces workspace scope in its own application and database paths too. Tenants do not mean separate AWS accounts or dedicated IPs for every customer, and they do not remove all shared-account reputation risk.

**Recipient and budget controls.** Daily and monthly allowances count recipients, including tests, when queued. A conservative cost estimate caps the workspace’s reserved spending allowance; it is not a complete AWS bill. SES also applies [regional account quotas](https://docs.aws.amazon.com/ses/latest/dg/manage-sending-quotas.html). Check your workspace’s current allowances before planning a campaign.

**Authenticated feedback.** A public webhook cannot trust a JSON body merely because it says “Amazon SNS.” Xem checks the expected topic and verifies the SNS signature using a certificate from the expected regional AWS endpoint. Duplicate notifications are processed idempotently. [AWS’s signature guidance](https://docs.aws.amazon.com/sns/latest/dg/sns-verify-signature-of-message.html) explains why verification must precede processing.

**Suppression at dispatch.** Permanent bounces, complaints, and existing unsubscribes block later sends. The worker checks again before dispatch, so a newly recorded suppression can stop an already queued message. A pause cannot recall a request that has already reached SES.

**Consent and unsubscribe.** Xem campaign submissions enforce HTTPS one-click unsubscribe headers. A generic SMTP listener cannot reliably infer whether arbitrary content is marketing. External applications must implement consent, visible unsubscribe links, and appropriate headers for their messages. Our [deliverability guide](/blog/email-deliverability) covers those responsibilities.

## Queued, accepted, delivered: three different promises

![Status chart: queued means Xem stored the message; accepted means SES accepted it; delivered means SES reported recipient-server acceptance. Inbox placement and human attention are not proved by these events. These are status meanings, not conversion rates.](/images/blog/managed-smtp-states.svg)

An SMTP `250` response from Xem means the outbox entry is committed. SES acceptance means the provider accepted the submission. A delivery event means the recipient’s mail server accepted it. None proves that a person saw the message or that it reached the primary inbox.

This distinction shapes retry behavior. If a provider request times out after transmission, Xem may not know whether SES accepted it. Sending it again automatically risks a duplicate. The implementation disables automatic SDK send retries and quarantines ambiguous outcomes as `DELIVERY_UNKNOWN` for investigation using authenticated events and provider records.

This can require operator intervention and delay a legitimate message. It is not exactly-once email delivery: an external SMTP client that loses the final acceptance response and retries can still submit a duplicate. Internal Xem sends use stable idempotency keys to handle their own repeated submissions.

## Engineering checks and operating responsibilities

At the time of the original implementation review, sending tests passed against real isolated PostgreSQL and a local SMTP/TLS exchange, including concurrent quota reservations, workspace isolation, credential revocation, durable acceptance, and suppression. The dedicated integration job also passed on GitHub. These are engineering checks, not an independent security audit or production deliverability study.

Our `xem.email` identity has passed SES domain, DKIM, and custom MAIL FROM verification, with DNS published on Cloudflare. AWS granted production access in `us-east-2`. That removes the account’s SES sandbox restriction; it is not an AWS endorsement or certification of Xem’s application.

Operating managed sending requires a scoped runtime role, a valid public SMTP certificate and listener, live feedback handling, monitoring, and operator response procedures. Service availability does not remove those ongoing responsibilities or guarantee inbox placement.

The implementation removes stored raw message content after seven days; addresses, subjects, and event metadata remain for support. Operators still need an explicit metadata-retention policy and secured backups. These operational requirements belong alongside the code.

## Help make the first send better

Open source gives you a way to inspect the claims and improve the implementation. Review the [backend sending PR](https://github.com/mailxem/xem.go/pull/10), [onboarding PR](https://github.com/mailxem/xem-app.ts/pull/4), and [deployment runbook at this implementation](https://github.com/mailxem/xem.go/blob/2b09fa5/docs/managed-sending.md).

Useful contributions include DNS-provider instructions, accessible onboarding copy, SMTP recipes, and deterministic bounce or complaint fixtures. Larger follow-ups include DMARC inheritance, account-wide scheduling, and feedback replay tools. Discuss larger changes with maintainers first, and use placeholder domains and messages in examples. Never put credentials or customer email in an issue.

The aim is your domain, a setup you can understand, and delivery outcomes you can investigate. SES provides the delivery foundation. Xem’s job is to make the path to it clear, controlled, and inspectable.
