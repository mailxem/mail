# Release: Cloudflare inbox for self-hosters

Drafts only. Nothing here has been posted.

## Headline

**Free custom-domain inbox on your own Cloudflare account**

The headline we wanted was "Free custom-domain inbox and newsletters on your own Cloudflare account". The newsletter half is not true: Xem only uses Cloudflare Email Sending for transactional mail and rejects campaigns on that sender, and Cloudflare sending is not on the Workers Free plan at all. Newsletters still need an SMTP provider such as Amazon SES. So the headline covers receiving and reading only.

## What is true (checked November 5, 2026)

- Email Routing inbound is unlimited on Workers Free. https://developers.cloudflare.com/email-service/platform/pricing/
- Workers Free: 100,000 requests a day, 10 ms CPU per invocation. https://developers.cloudflare.com/workers/platform/limits/
- R2 free: 10 GB-month storage, 1M Class A and 10M Class B operations a month, free egress. R2 needs a payment method and subscription checkout first. https://developers.cloudflare.com/r2/pricing/ and https://developers.cloudflare.com/r2/get-started/
- D1 free: 5M rows read and 100k written a day, 500 MB per database. https://developers.cloudflare.com/d1/platform/pricing/ and https://developers.cloudflare.com/d1/platform/limits/
- Cloudflare Email Sending: not on Workers Free. Workers Paid is $5 a month with 3,000 emails included, then $0.35 per 1,000.
- Self-hosted Xem with no billing configured no longer shows a paid-plan wall for this feature. On Xem Cloud it stays a paid feature.

Not verified yet: live Email Routing delivery into a deployed Worker, and the Deploy to Cloudflare button run end to end. Do not post until someone has done both on a real domain, and update the drafts if anything differs.

Do not claim: newsletters on Cloudflare, free replies, any number of users, any speed, "unlimited storage".

## X post

> Xem can now keep an inbox for an address on your own domain in your own Cloudflare account.
>
> Email Routing hands mail to a Worker, the Worker stores it in your R2 bucket and D1 database, and Xem reads it over a signed API. Receiving and reading fit Cloudflare's free tier.
>
> Replies need a sender (Cloudflare's is on Workers Paid, or use SES). Guide: https://docs.xem.email/guides/cloudflare-inbox

## r/CloudFlare post

**Title:** Email Routing + a Worker + R2 + D1 as a free inbox for your own domain (open source)

> I work on Xem, an open-source (GPL-3.0) email marketing app with a shared inbox. We just finished making one part of it usable on its own Cloudflare account, and the setup leans on a few Cloudflare products together, so I figured people here might find it interesting or poke holes in it.
>
> How it works: you route an address like hello@yourdomain.com to a Worker with Email Routing. The Worker writes the raw message to a private R2 bucket, parses it, and indexes it in D1. Xem (self-hosted) reads the mailbox through a small HMAC-signed API on the Worker. Xem never gets an R2 or D1 token, and the bucket stays private.
>
> There is a Deploy to Cloudflare button that creates the D1 database and R2 bucket and asks for two values (the address and a shared secret). Or it's `wrangler deploy --secrets-file .dev.vars` plus one D1 migration.
>
> What's free and what isn't, from Cloudflare's pricing pages:
>
> - Email Routing inbound: unlimited on Workers Free.
> - Workers Free: 100k requests a day and 10 ms CPU per invocation. I haven't measured how big a message can get before parsing hits that 10 ms, so very large messages may need Workers Paid.
> - R2: 10 GB-month free, but you have to add a payment method and do the R2 subscription checkout first.
> - D1: 500 MB per database on Free. If you hit the daily write limit, the raw message stays in R2 and a cron retries indexing every minute.
> - Replying is not free through Cloudflare. Email Sending isn't on Workers Free; Workers Paid is $5/month with 3,000 sends included. You can reply through any other SMTP sender instead.
> - It doesn't send newsletters through Cloudflare. That still needs something like SES.
>
> One address per Worker deployment. It also points your domain's MX at Cloudflare, so don't turn it on for a domain that already gets mail somewhere else.
>
> Guide with the pricing links: https://docs.xem.email/guides/cloudflare-inbox
> Worker source: https://github.com/mailxem/mail/tree/undefined/devops/cloudflare-mailbox
>
> Happy to answer questions, and I'd like to hear if anyone has run Workers Free email handlers on large attachments and knows where the CPU limit bites.
