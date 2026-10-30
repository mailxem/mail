---
title: "Open-source email marketing software in 2026: Listmonk, Mautic and Xem"
description: "Three open-source email marketing tools compared on licence, what they include, what they need to run, and who each one suits. Plus where Sendy fits."
date: "2026-10-30"
updated: "2026-10-30"
category: "Comparisons"
tags: ["open-source email marketing", "self-hosted newsletter", "Listmonk", "Mautic", "open-source Mailchimp alternative"]
cover: "/images/blog/open-source-email-marketing.webp"
coverAlt: "Forest green cover with an open padlock beside four stacked cream cards labelled campaigns, automations, CRM and inbox, under the words Open by default."
coverSource: "https://xem.email/blog/editorial"
author: "Xem editorial"
---
Three actively developed open-source tools cover most email marketing needs in 2026. Listmonk is a fast, focused newsletter and campaign manager. Mautic is a full marketing-automation platform. Xem is an email marketing suite that adds a CRM and a shared inbox to campaigns, newsletters and automations. All three are free to run on your own server and send through Amazon SES or any SMTP provider.

Sendy often appears in the same lists, but it is not open source. It is a paid, one-time licence. This guide compares the three open-source tools, explains where Sendy fits, and lists what self-hosting asks of you. Facts were checked on October 30, 2026.

## How do they compare at a glance?

| | Listmonk | Mautic | Xem |
| --- | --- | --- | --- |
| Licence | AGPL-3.0 | GPL-3.0 | GPL-3.0 |
| GitHub stars | about 24,000 | about 10,900 | newer project |
| Main focus | Newsletters and campaigns | Marketing automation | Email marketing suite |
| Visual automations | Not listed on its site | Yes, campaign builder | Yes |
| Forms and landing pages | Subscription forms | Forms and landing pages | Signup forms |
| CRM | No | Contacts with lead scoring | Contact CRM with notes and stages |
| Shared inbox | No | No | Yes, over IMAP |
| Public newsletter archive | Yes | Not listed | Yes |
| Built with | Go and PostgreSQL | PHP | Go backend, Next.js app |
| Hosted option | Third parties only | Official trial through a partner | Hosted app or self-host |

"Not listed" means the project's own pages did not mention the feature when we checked. It does not mean the feature is impossible with plugins or custom work.

## Listmonk: when should you choose it?

Choose Listmonk if you want a newsletter and campaign manager that is quick to run and does one job well. It ships as a single binary backed by PostgreSQL, handles large lists efficiently, and supports a public archive of campaigns. It can process bounces from Amazon SES and SendGrid.

Its focus is also its limit. Listmonk's site does not describe visual automations, a CRM or an inbox. If you need those, you will be connecting other tools.

## Mautic: when should you choose it?

Choose Mautic if marketing automation is the centre of your work: multi-step campaigns, lead scoring and landing pages for a sales team. It is the most established of the three, with a large community and partner ecosystem.

It is also the heaviest to run. Mautic’s own documentation advises against cheap shared hosting, and it has many moving parts. Teams often need someone who knows Mautic to keep it healthy.

## Xem: when should you choose it?

Choose Xem if you want the day-to-day email work of a small team in one place: campaigns, recurring newsletters, visual automations, signup forms, a contact CRM, a shared inbox and analytics. It includes 200+ editable templates, AI help for drafting, an API with webhooks, and an MCP server so AI assistants can work with your campaigns and contacts.

Xem is younger than Listmonk and Mautic, with a smaller community. It does not handle paid subscriptions, and contact imports do not carry tags across yet. You can self-host it with a one-line installer, described in the [self-hosting guide](https://docs.xem.email/guides/self-hosting), or use the hosted app. In both cases sending goes through Amazon SES or your own SMTP provider.

## Where does Sendy fit?

Sendy is a self-hosted newsletter app sold as a $69 one-time licence per domain. It runs on your own PHP and MySQL host and sends mainly through Amazon SES, with support for a few other providers. It covers campaigns, autoresponders, segmentation and subscribe forms.

It belongs in the same budget conversation because it also sends at your provider's price. But the source is not openly licensed, so you cannot fork it or audit changes in public.

## What does self-hosting ask of you?

Self-hosting saves money on software and gives you control. It also hands you jobs a hosted platform would do:

- **A server and a database**, with backups you have tested by restoring.
- **Updates**, including security fixes, applied promptly.
- **Sending setup**: domain authentication with SPF, DKIM and DMARC, plus the Amazon SES production-access request if you use SES. New SES accounts are limited to 200 messages a day until AWS approves them.
- **Monitoring** of bounces and complaints after every send.

If that list looks heavy, choose a tool that also offers hosting, or start on a hosted plan and move once your list makes the savings worthwhile.

## How much do they cost to run?

The software is free. Your costs are hosting and sending. Through Amazon SES, sending costs $0.10 or $0.16 per 1,000 emails, so a weekly newsletter to 10,000 subscribers costs $4.00 to $6.40 a month. The worked examples are in [what it costs to send a newsletter through Amazon SES](/blog/amazon-ses-newsletter-cost).

Hosted newsletter platforms charged $89 to $139 a month for 10,000 subscribers when we checked. See [Mailchimp alternatives in 2026](/blog/mailchimp-alternative) for that comparison.

## Common questions

### What is the best open-source Mailchimp alternative?

It depends on scope. For newsletters only, Listmonk. For heavy marketing automation with lead scoring, Mautic. For campaigns, newsletters, automations, a CRM and a shared inbox in one app, Xem.

### Is Sendy open source?

No. Sendy is sold as a one-time commercial licence. It is self-hosted, but its source is not published under an open-source licence.

### Can I move my list into an open-source tool?

Yes. Export a CSV from your current platform, keep only subscribed contacts, and import it. The steps for beehiiv, Kit, Substack and Mailchimp are in [how to move your newsletter list](/blog/move-your-newsletter-list).

## Sources and further reading

- [Listmonk on GitHub](https://github.com/knadh/listmonk) and [listmonk.app](https://listmonk.app/), checked October 30, 2026.
- [Mautic on GitHub](https://github.com/mautic/mautic) and [mautic.org](https://www.mautic.org/), checked October 30, 2026.
- [Sendy](https://sendy.co/), checked October 30, 2026.
- [Xem on GitHub](https://github.com/mailxem/mail).
- [Amazon SES: Request production access](https://docs.aws.amazon.com/ses/latest/dg/request-production-access.html).
