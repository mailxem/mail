# Free sending comparison

Official pricing pages checked September 26, 2026. These are sending allowances, not a claim of complete feature parity.

| Provider | Free sending allowance | Source |
| --- | --- | --- |
| Resend | 3,000 emails/month; 100/day | https://resend.com/pricing |
| Brevo | 300 emails/day, subject to sending approval | https://www.brevo.com/pricing/ |
| Postmark | 100 emails/month on its developer plan | https://postmarkapp.com/pricing |
| Xem starter | 3,000 recipient sends/month; 100/day after domain verification | `managed-sending.md` |

Xem counts recipients, so a message to multiple recipients uses multiple sends. Its default 1,000 micro-USD reservation per recipient requires a 3,000,000 micro-USD ($3) monthly allowance. This is an internal delivery cap, not a charge to the free user or an estimate of the AWS invoice. Limits and remaining allowance are read from the backend account; operator overrides remain authoritative.
