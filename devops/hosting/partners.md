# Hosting partner links

The README's deployment buttons lead to provider-specific sections in [the hosting guide](README.md). The guide's three named signup-link definitions are the only provider destinations that need to change when Xem receives an approved affiliate URL. There are no affiliate IDs, enrollment claims or earnings configured today.

## Programs to consider

Research date: **September 28, 2026**. Check the linked program's current terms and eligibility before enrolling or describing an offer.

| Provider / program | What was verified | Fit for Xem |
| --- | --- | --- |
| [DigitalOcean Affiliate Program](https://www.digitalocean.com/affiliates) | The public page advertises 10% commission on a referred paying user's monthly spend for one year and links enrollment through CJ. Account approval, eligible traffic and payout terms still need review. | Cash-commission option, but Droplets' SMTP restrictions must stay prominent in the deployment guide. Do not recommend an incompatible email setup because it pays a commission. |
| [DigitalOcean Referral Program](https://www.digitalocean.com/referral-program) | Separate account-credit program. The page describes a $25 credit after a referred customer reaches $25 in billings. | Infrastructure credits are different from cash affiliate revenue. Do not describe credits as cash payouts. |
| [Vultr Creator Program](https://www.vultr.com/creator/) / [Referral Program](https://www.vultr.com/company/referral-program/) | Both programs are linked from the official [Vultr documentation footer](https://docs.vultr.com/what-ports-are-blocked). Their program pages denied automated access, so current eligibility, compensation and payout terms were not verified. | Worth reviewing because the published network block list names port 25, allowing a potential fit with authenticated SMTP on other ports. Confirm account connectivity and program terms independently. |
| Hetzner Cloud | The [official FAQ](https://docs.hetzner.com/cloud/general/faq/) says its old referral-credit program has been discontinued and old links cannot be used to participate. | Keep the ordinary hosting option for users who prefer it. Do not add an old referral code or promise signup credits. |

The root README avoids quoting commission amounts, temporary discounts or "free" offers so its deployment instructions do not depend on a promotion staying active.

## Activate an approved link

1. The project owner enrolls in the chosen program using the correct publisher/business identity and reviews its terms. Any contract acceptance, payment/tax details and account verification belong to that owner. Merely linking the program application does not create an affiliate relationship.
2. Obtain a public tracking URL from the approved program dashboard. Do not invent a referral ID, borrow another creator's code, or use a dashboard/session URL.
3. Confirm the program allows GitHub README/documentation placements. Generate a destination appropriate to a Linux VPS, not a managed app product that cannot run this Swarm stack. Use a sub-ID such as `github-readme` only if that program supports it.
4. Update the matching `vultr-signup`, `hetzner-signup` or `digitalocean-signup` definition at the bottom of [README.md](README.md). Keep the README buttons pointing to the setup sections so restrictions and prerequisites remain visible before purchase.
5. Add a clear disclosure immediately beside each compensated signup link, for example: **"Affiliate link: Xem may earn a commission if you sign up through this link."** For a credit program, say **"Referral link: Xem may receive account credit."** Do not say "at no extra cost" or promise a discount unless the actual offer supports that claim. Replace the guide's blanket "standard links" statement with an accurate list/status.
6. Check that the link reaches the intended provider and keeps the issued referral identifier intact, without creating a paid server. Confirm attribution through the program dashboard when available. A link opening successfully is not proof of a payable conversion.
7. Review the diff and publish through the normal PR process. Keep partner approval, private dashboard screenshots, tax/payment information and credentials out of the repository.

An approved public URL is the missing input for monetization. No hosting accounts, affiliate accounts or paid infrastructure are created by these documentation changes.
