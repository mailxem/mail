# Email deliverability checker: launch drafts

Page: https://xem.email/tools/email-checker
Status: drafts, not posted. Check the page is live before posting anything.

What the tool does (keep every post inside this):

- Takes a domain and looks up its MX, SPF, DMARC and DKIM records from the
  browser, through Cloudflare's public DNS-over-HTTPS resolver. No backend.
- SPF: flags a missing record, more than one record, `+all`, `?all`, a missing
  `all`, `ptr`, includes that have no SPF record, and counts DNS lookups
  through every include and redirect against the limit of 10.
- DMARC: flags a missing or invalid record, more than one record, a missing or
  unknown `p=`, a missing `rua`, and `pct` below 100. Explains that `p=none`
  meets the Gmail/Yahoo requirement but doesn't stop spoofing.
- DKIM: checks a selector you type, or tries google, selector1, selector2, k1,
  s1, s2 and resend. Flags revoked keys and 1024-bit RSA keys. Says plainly that
  Amazon SES uses random selectors, so you have to paste one in.
- Each problem comes with the record to add or change.
- Doesn't check: From alignment, reverse DNS, blocklists, the one-click
  unsubscribe header. The page says so.

## X

> If you send newsletters from your own domain, Gmail and Yahoo check SPF, DKIM and DMARC before they look at anything else.
>
> I put up a free checker that reads those records and tells you what's missing, with the exact TXT record to add. It also counts SPF's DNS lookups through every include, since going past 10 breaks SPF quietly.
>
> Runs in your browser against Cloudflare DNS, nothing sent to us.
>
> https://xem.email/tools/email-checker

## r/emailmarketing

Title: Free SPF/DKIM/DMARC checker that gives you the record to paste, not just a red X

> Disclosure first: I work on Xem, an open-source email marketing tool. This checker is on our site, but it's free, has no signup and doesn't need Xem.
>
> Most deliverability questions here end up at the same three records, and a lot of checkers tell you something is wrong without telling you what to type into your DNS. This one tries to do the second part.
>
> You enter a domain and it looks up:
>
> - MX
> - SPF: missing or duplicate records, `+all`/`?all`, includes that point at domains with no SPF, and the total number of DNS lookups through every nested include (the limit is 10, and stacking Google + a CRM + a newsletter tool + a helpdesk gets there faster than people expect)
> - DMARC: missing record, policy, whether you get reports (`rua`), `pct`
> - DKIM for a selector you give it, or a handful of common ones (Google Workspace, Microsoft 365, SendGrid, Resend)
>
> For each problem it shows the record to add or the corrected version of yours.
>
> Two honest limits. If you send through Amazon SES, it can't guess your DKIM selector because SES generates random ones, so you paste one from the SES console or from the `s=` in a sent message's DKIM-Signature header. And passing all of this only gets you through the door; it says nothing about complaint rates or whether people want your email.
>
> The lookups run in your browser against Cloudflare's public DNS resolver, so we never see the domains you check.
>
> https://xem.email/tools/email-checker
>
> If something it says is wrong or unclear for your setup, I'd like to hear it.

## Show HN

Title: Show HN: SPF/DKIM/DMARC checker that runs in the browser over DNS-over-HTTPS

First comment:

> I work on Xem, an open-source (GPL-3.0) email marketing app. Every email-tool thread here turns to deliverability, and the first answer is always "check SPF, DKIM and DMARC", so we made a page that does that and tells you the exact record to fix.
>
> There's no backend. The page is a static export, and the browser queries Cloudflare's DoH JSON endpoint (cloudflare-dns.com/dns-query with accept: application/dns-json), which sends CORS headers. So we don't see the domains you type.
>
> The part I found most useful to get right is SPF's 10-lookup limit (RFC 7208 4.6.4). The checker walks every include and redirect recursively and counts include, a, mx, ptr, exists and redirect; ip4/ip6/all are free. Includes with macros are skipped because they expand per message.
>
> DKIM is the weak spot: there's no way to list a domain's selectors from DNS, so it tries a few common ones and otherwise asks you for the s= value from a sent message. Amazon SES uses random selectors, so for SES that's the only way.
>
> It doesn't check alignment, PTR records or blocklists, since those need a real message.
>
> The parsing is a small TypeScript module with tests in the repo: https://github.com/mailxem/mail/tree/undefined/website/lib
