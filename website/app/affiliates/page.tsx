import type { Metadata } from "next";
import Link from "next/link";
import { ArrowUpRight } from "lucide-react";
import { SiteHeader, SiteFooter } from "@/components/site-chrome";
import { appLink } from "@/lib/site";
import { primaryButton } from "@/lib/brand";

const title = "Affiliates | Xem: earn 30% for as long as they pay";
const description =
  "Refer people to Xem Cloud and earn 30% of every payment they make, for as long as they stay. Paid monthly by bank transfer once you reach $100.";
export const metadata: Metadata = {
  title,
  description,
  alternates: { canonical: "/affiliates" },
  openGraph: { type: "website", url: "/affiliates", title, description },
  twitter: { card: "summary", title, description },
};

const steps = [
  ["Get your link", "Sign in to Xem, open Settings, then Affiliates. Your link works on any xem.email page: add ?via= and your code."],
  ["Share it", "Write about how you send your newsletter, record a walkthrough, or point a friend at the pricing page."],
  ["Earn on every payment", "When someone signs up through your link and pays for a plan, 30% of each payment is yours, every month or year they stay."],
];

const terms: [string, string][] = [
  ["What you earn", "30% of every payment a referred account makes for its plan, after discounts and without tax. Renewals, yearly plans, upgrades and usage past the plan's included emails all count, for as long as the account pays."],
  ["Which sign-ups count", "A new Xem Cloud account created within 60 days of a click on your link, in the same browser. If someone clicks more than one affiliate link, the last one counts."],
  ["When commissions unlock", "Each commission unlocks 60 days after its payment, so refunds and chargebacks can settle first."],
  ["How you're paid", "At the start of each month, everything unlocked goes out by bank transfer once it adds up to $100. Smaller amounts carry over. Payouts are in US dollars, and any transfer or currency conversion fee comes out of the payout."],
  ["Refunds and disputes", "A refunded or disputed payment earns nothing. If its commission was already paid, it comes off your next payout."],
  ["What isn't allowed", "Referring yourself or your own workspace, ads on searches for Xem, spam, and coupon or deal sites. Breaking these rules can close your affiliate account and cancel unpaid commissions."],
  ["Privacy", "Your dashboard shows clicks, sign-ups and earnings per customer. Customer names and emails stay private."],
  ["Changes", "We may change these terms or end the program. Commissions you've already earned keep the rules they were earned under."],
];

export default function AffiliatesPage() {
  return (
    <>
      <SiteHeader active="none" />
      <main id="main">
        <section className="mx-auto max-w-[1160px] px-5 pb-12 pt-16 md:pt-24">
          <p className="text-xs font-medium uppercase tracking-[0.17em] text-iris">Affiliates</p>
          <h1 className="mt-5 max-w-4xl font-editorial text-[clamp(3rem,6.5vw,6rem)] leading-[1.02] tracking-[-0.045em]">
            30% of every payment.
            <br />
            <em>For as long as they pay.</em>
          </h1>
          <p className="mt-7 max-w-xl text-lg leading-relaxed text-muted">
            Send people to Xem Cloud. When they pay for a plan, you get a third of it, every month they stay. No cap, and nothing to buy.
          </p>
          <a href={appLink("/settings/affiliates")} className={`${primaryButton} mt-9`}>
            Get your link <ArrowUpRight size={16} />
          </a>
          <p className="mt-4 text-sm text-muted">
            New to Xem? <a className="text-iris underline" href={appLink("/auth/register")}>Create a free account</a> first; you don't need a paid plan to be an affiliate.
          </p>
        </section>

        <section aria-label="How it works" className="mx-auto grid max-w-[1160px] gap-4 px-5 md:grid-cols-3">
          {steps.map(([heading, body], i) => (
            <article key={heading} className="rounded-2xl border border-ink/15 p-7">
              <p className="font-editorial text-4xl text-iris">{i + 1}</p>
              <h2 className="mt-4 font-editorial text-2xl tracking-tight">{heading}</h2>
              <p className="mt-3 text-sm leading-relaxed text-muted">{body}</p>
            </article>
          ))}
        </section>

        <section aria-labelledby="example-title" className="mt-16 bg-[#ebe7f2] px-5 py-16 md:py-20">
          <div className="mx-auto max-w-[1160px]">
            <h2 id="example-title" className="font-editorial text-4xl tracking-tight md:text-5xl">What it adds up to</h2>
            <p className="mt-5 max-w-2xl text-base leading-relaxed text-muted">
              Refer ten newsletters on Growth ($29 a month) and you earn $87 a month, or $1,044 a year, for as long as they stay. One business on Scale ($99 a month) is $29.70 a month on its own. See <Link href="/pricing" className="text-iris underline">pricing</Link>.
            </p>
          </div>
        </section>

        <section aria-labelledby="terms-title" className="mx-auto max-w-[860px] px-5 py-16">
          <h2 id="terms-title" className="font-editorial text-4xl tracking-tight">The terms</h2>
          <dl className="mt-8 divide-y divide-ink/10 border-y border-ink/10">
            {terms.map(([term, detail]) => (
              <div key={term} className="grid gap-2 py-5 md:grid-cols-[220px_1fr] md:gap-8">
                <dt className="font-medium">{term}</dt>
                <dd className="text-sm leading-relaxed text-muted">{detail}</dd>
              </div>
            ))}
          </dl>
          <p className="mt-6 text-sm text-muted">
            Questions? Email <a className="text-iris underline" href="mailto:team@xem.email">team@xem.email</a>.
          </p>
        </section>
      </main>
      <SiteFooter />
    </>
  );
}
