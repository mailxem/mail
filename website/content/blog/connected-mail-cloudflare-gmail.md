---
title: "Your inbox, your infrastructure: connected mail in Xem"
description: "A preview of Gmail, Google Workspace, and Cloudflare connections in Xem, with an open-source foundation and plans for hosted and self-hosted paid tools."
date: "2026-09-28"
updated: "2026-09-28"
category: "Deliverability"
tags: ["Cloudflare", "Gmail", "Google Workspace", "open source", "IMAP"]
cover: "/images/blog/managed-smtp-custom-domains.webp"
coverAlt: "Xem illustration of an envelope moving through an email delivery path."
coverSource: "https://xem.email/blog/editorial"
author: "Xem editorial"
featured: false
published: false
---

**Editorial draft. Keep unpublished until live provider validation is complete. Paid workflows below are a roadmap, not an available product bundle.**

Your business email already has a home. Your team should be able to work with it without rebuilding its infrastructure first.

We're adding Gmail and Google Workspace mailbox connections to Xem, improving the existing IMAP inbox, and connecting Cloudflare Email Sending for transactional messages. The foundation stays open source. We're also planning Pro, Team, and Company options, available through hosted subscriptions and self-hosted licenses.

## Bring the mailbox you already use

The Google connection links an inbox and a sender through a dedicated mailbox authorization flow. Google sign-in alone does not grant Xem access to your mail. You choose the mailbox, authorize it, and select it in your inbox and composer.

This first version connects a mailbox to the workspace under its existing permissions. It is intended for a business mailbox the workspace should access. Private mailboxes, delegation, and individual mailbox permissions need a dedicated access model and are part of the next stage.

For other providers, IMAP remains available. The current implementation improves folder listing, search, pagination, and stable message selection, and adds explicit read and star actions. Connections verify TLS certificates.

## Cloudflare handles transactional sending

Cloudflare Email Sending can submit transactional messages from your own account. Xem's Go backend calls the REST API directly, so your existing Xem installation can stay where it runs today.

Email Routing and Email Sending serve different purposes. Routing handles incoming mail; Sending submits outgoing messages. The connector in this release covers **sending**. It does not turn Xem into a Cloudflare-hosted mailbox or move your data into D1 and R2.

Cloudflare currently restricts Email Sending to transactional mail. Use a suitable SMTP provider or Xem's managed-sending path for newsletters and campaigns. Cloudflare's account allowance, paid-plan requirement, and usage charges belong to the provider and are separate from Xem. Read the current [Cloudflare pricing](https://developers.cloudflare.com/email-service/platform/pricing/) and [limits](https://developers.cloudflare.com/email-service/platform/limits/) before planning your usage.

## A useful core, with room to grow

Basic provider connections should be useful on their own. The paid roadmap focuses on the work around email: saved replies and follow-ups for individuals; shared ownership, assignments, and notes for teams; approval policies, audit exports, and access administration for companies.

We plan to offer these through managed hosting and self-hosted licenses. Existing free features will remain available. We will publish the final packages and prices as their features become ready; today's repository is not a promise that every planned workflow has shipped.

## Build with us

Start with the [Xem repository](https://github.com/mailxem/mail), follow the [connected-mail setup guide](https://github.com/mailxem/mail/blob/undefined/docs/connected-mail.md), or explore [hosted Xem](https://app.xem.email).

If you maintain a mailbox provider integration, want to improve the IMAP experience, or have a team workflow that deserves a better default, contributions and concrete examples help. We would especially like feedback on the steps between reading a message, deciding who owns it, and sending a reply.
