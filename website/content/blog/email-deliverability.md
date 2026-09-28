---
title: "Email deliverability in 2026: the checks that matter"
description: "A practical guide to Gmail, Yahoo, and Outlook sender requirements, authentication, one-click unsubscribe, and the difference between acceptance and delivery."
date: "2026-09-10"
updated: "2026-09-10"
category: "Deliverability"
tags: ["email deliverability 2026", "Gmail sender requirements", "SPF DKIM DMARC"]
cover: "/images/blog/email-deliverability.webp"
coverAlt: "Forest green cover showing a cream envelope inside concentric delivery rings, with SPF, DKIM, and DMARC labels."
coverSource: "https://xem.email/blog/editorial"
author: "Xem editorial"
---
An email can be accepted by a mail server and still never appear in the primary inbox. That distinction is the starting point for useful deliverability work.

Your application knows that it submitted a message. Your SMTP provider may confirm acceptance. A receiving service decides whether to accept, defer, reject, or place that message in spam. A dashboard should not collapse those stages into one green “delivered” number.

As of this guide's September 2026 review, authentication and easy unsubscribe remain central to the published requirements from Gmail, Yahoo, and Outlook.com. They are necessary foundations, not a promise of inbox placement.

## Know which requirements apply

Provider rules are not identical. Check the current documentation for each destination rather than treating “bulk sender” as a universal threshold.

| Destination | Published requirements to check |
| --- | --- |
| Personal Gmail accounts | All senders need SPF or DKIM. Senders of 5,000 or more messages per day need SPF, DKIM, and DMARC, plus additional requirements including one-click unsubscribe for marketing and subscribed mail. |
| Yahoo | All senders need authentication and low complaint rates. Bulk senders need SPF and DKIM, a passing DMARC policy, and easy unsubscribe. Yahoo's requirements should be read directly rather than assigned Gmail's threshold. |
| Outlook.com | The postmaster announcement applies stricter SPF, DKIM, and DMARC standards to domains sending over 5,000 emails per day, beginning May 5, 2025. |

These are short summaries. Reverse DNS, message formatting, transport security, alignment, and reputation also matter. Google specifies TLS for transmission and recommends keeping Postmaster Tools spam rates below 0.1%, avoiding 0.3% or higher. Treat that as a provider-specific operating signal, not a target to approach.

## Check authentication on a real message

A DNS record can exist and still fail for the email you actually send. Inspect a message sent through every active provider.

**SPF** authorizes sending infrastructure for the envelope sender domain. **DKIM** applies a signature that a recipient can validate. **DMARC** connects an authenticated domain to the domain people see in the From header through alignment.

That last step is easy to miss. A message can pass SPF for a provider's domain while failing to align with your visible From domain. A correctly aligned DKIM signature can satisfy the other path. Check the authentication results in the received message, not just a DNS-checking website.

If several teams use the same domain, inventory every sender before tightening a DMARC policy. Start with visibility into legitimate mail, fix gaps, and move toward enforcement deliberately. Do not copy an aggressive policy into DNS without understanding the systems it will affect.

## Make leaving straightforward

A footer link and one-click unsubscribe headers solve related but different problems. The footer gives the reader a visible way to leave. Header-based one-click unsubscribe lets a supporting mailbox provider offer its own unsubscribe control.

RFC 8058 defines a POST-based mechanism. It exists partly because mail software can fetch URLs automatically; an ordinary GET request should not accidentally change someone's subscription. Google requires the appropriate one-click headers for applicable marketing and subscribed messages, along with a clearly visible body link.

Test the whole path. Can the provider reach the endpoint? Does the token identify the right subscription? Is the change recorded promptly? Will a queued automation respect it? A successful web response is not enough if the next email still sends.

Keep this separate from account access. A person should not need to remember a password to stop receiving your marketing.

## Protect reputation with audience discipline

Authentication does not make an unwanted message wanted. A bought list remains a poor choice even if every signature passes.

Send what the signup form promised. Avoid sudden jumps in volume to old, untested contacts. Investigate hard bounces and complaints, and preserve suppression records when moving data between systems. Do not treat an import as fresh consent.

If you change sending providers, carry over the audience rules and suppression state as carefully as the template. Routing flexibility is useful; it does not reset your responsibility to recipients or guarantee a fresh reputation.

## Build a weekly review you can act on

Review these measures by sending domain and campaign:

1. SMTP acceptance and explicit failures.
2. Deferrals or uncertain outcomes that require investigation.
3. Known bounces and complaints, with their coverage limits.
4. Unsubscribes and list sources.
5. Provider reputation signals where you have access to them.

Use Xem's delivery health view to review the outcomes the system actually records. A missing complaint feed must not look like proof that nobody complained. Similarly, a low open rate is not a reliable diagnosis of spam placement.

When something changes sharply, isolate the affected provider, domain, audience source, and send window. Preserve the actual SMTP response. “Our deliverability dropped” is much harder to debug than “this domain's messages began receiving a specific rejection after Tuesday's import.”

## A pre-send check for the next campaign

Send to test accounts at the providers your audience uses. Inspect authentication, follow the main link, and test the unsubscribe flow with a disposable test subscription. Review mobile rendering and the plain-text part. Confirm the audience excludes suppressed contacts.

Then watch the early operational signals after the campaign starts. A polished message and valid DNS do not remove the need to investigate errors while they are happening.

## Sources and further reading

- [Google: Email sender guidelines](https://support.google.com/a/answer/81126?hl=en).
- [Yahoo: Sender requirements and recommendations](https://senders.yahooinc.com/best-practices/).
- [Microsoft: Outlook.com Postmaster announcements](https://sendersupport.olc.protection.outlook.com/pm/postmaster.aspx).
- [IETF: RFC 8058, one-click unsubscribe](https://www.rfc-editor.org/rfc/rfc8058).
