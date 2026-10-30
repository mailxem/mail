import type { Metadata } from "next";
import Link from "next/link";
import { ArrowUpRight, Check } from "lucide-react";
import { SiteHeader, SiteFooter } from "@/components/site-chrome";
import { appLink, docsUrl, siteUrl } from "@/lib/site";
import { primaryButton } from "@/lib/brand";

const title = "Pricing | Xem — pay for the emails you send";
const description =
  "Xem Cloud is priced per email: free for 3,000 a month, then from $9 for 20,000. Every plan includes campaigns, newsletters, automations, forms, CRM and a shared inbox. Self-hosting is free.";
export const metadata: Metadata = {
  title,
  description,
  alternates: { canonical: "/pricing" },
  openGraph: { type: "website", url: "/pricing", title, description },
  twitter: { card: "summary", title, description },
};

const plans = [
  {
    name: "Free",
    price: 0,
    included: 3000,
    overage: null,
    note: "No card needed. Sending pauses at 3,000 until next month.",
  },
  {
    name: "Starter",
    price: 9,
    included: 20000,
    overage: 0.4,
    note: "For a weekly newsletter to a few thousand readers.",
  },
  {
    name: "Growth",
    price: 29,
    included: 100000,
    overage: 0.3,
    note: "For growing lists and regular campaigns.",
    popular: true,
  },
  {
    name: "Scale",
    price: 99,
    included: 500000,
    overage: 0.2,
    note: "For large audiences sending often.",
  },
];

const everyPlan = [
  "Campaigns and recurring newsletters",
  "Visual automations",
  "Signup forms, lists and contact CRM",
  "Shared inbox",
  "Analytics and send-time suggestions",
  "200+ editable templates and AI drafting",
  "API, webhooks and MCP server",
  "Public newsletter archive and RSS",
];

// Monthly cost of a weekly newsletter to 10,000 subscribers (40,000 emails),
// using the prices in /blog/mailchimp-alternative, checked October 30, 2026.
const comparison = [
  {
    name: "Xem Starter",
    price: 17,
    how: "$9 + 20,000 emails at $0.40 per 1,000",
  },
  { name: "MailerLite Comfort", price: 89, how: "per subscriber" },
  { name: "Mailchimp Standard", price: 135, how: "per contact" },
  { name: "beehiiv Lite", price: 139, how: "per subscriber" },
  { name: "Kit Creator", price: 139, how: "per subscriber" },
];

const faqs = [
  {
    q: "What counts as an email?",
    a: "Each recipient of each email Xem sends for you counts once: a campaign to 1,000 people is 1,000 emails. Test sends to yourself are free, and storing contacts costs nothing.",
  },
  {
    q: "What happens when I go past my plan?",
    a: "On Starter, Growth and Scale, sending never stops: emails past your included amount are billed at your plan's rate per 1,000 on your next invoice. On Free, sending pauses at 3,000 emails until the next month or until you upgrade.",
  },
  {
    q: "Do I pay for contacts or subscribers?",
    a: "No. You pay for emails sent, not for the size of your list, so unsubscribed and quiet contacts don't add to your bill.",
  },
  {
    q: "Is anything held back on cheaper plans?",
    a: "No. Every plan, including Free, has the full product. Plans differ only in how many emails are included.",
  },
  {
    q: "Can I run Xem myself instead?",
    a: "Yes. Xem is open source under GPL-3.0. Self-hosting is free; you pay your server and your email provider, such as Amazon SES at $0.10 to $0.16 per 1,000 emails.",
  },
  {
    q: "How is billing handled?",
    a: "Payments, invoices and sales tax are handled by Dodo Payments as merchant of record. You can change or cancel your plan from Settings → Billing at any time.",
  },
];

const money = (n: number) => (n === 0 ? "$0" : `$${n}`);
const count = (n: number) => n.toLocaleString("en-US");

export default function PricingPage() {
  const structuredData = {
    "@context": "https://schema.org",
    "@graph": [
      {
        "@type": "SoftwareApplication",
        name: "Xem",
        url: siteUrl,
        applicationCategory: "BusinessApplication",
        operatingSystem: "Web",
        offers: plans.map((p) => ({
          "@type": "Offer",
          name: `Xem ${p.name}`,
          price: p.price.toFixed(2),
          priceCurrency: "USD",
          description: `${count(p.included)} emails a month${p.overage ? `, then $${p.overage.toFixed(2)} per 1,000` : ""}.`,
        })),
      },
      {
        "@type": "FAQPage",
        mainEntity: faqs.map((f) => ({
          "@type": "Question",
          name: f.q,
          acceptedAnswer: { "@type": "Answer", text: f.a },
        })),
      },
    ],
  };
  return (
    <>
      <SiteHeader active="pricing" />
      <main id="main">
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{
            __html: JSON.stringify(structuredData).replace(/</g, "\\u003c"),
          }}
        />
        <section className="mx-auto max-w-[1160px] px-5 pb-12 pt-16 md:pt-24">
          <p className="text-xs font-medium uppercase tracking-[0.17em] text-iris">
            Pricing
          </p>
          <h1 className="mt-5 max-w-4xl font-editorial text-[clamp(3rem,6.5vw,6rem)] leading-[1.02] tracking-[-0.045em]">
            Pay for the emails you send.
            <br />
            <em>Not for the list you keep.</em>
          </h1>
          <p className="mt-7 max-w-xl text-lg leading-relaxed text-muted">
            Every plan has the whole of Xem. Start free with 3,000 emails a
            month, and upgrade when your list grows.
          </p>
        </section>

        <section
          aria-label="Plans"
          className="mx-auto grid max-w-[1160px] gap-4 px-5 md:grid-cols-2 lg:grid-cols-4"
        >
          {plans.map((p) => (
            <article
              key={p.name}
              className={`flex flex-col rounded-2xl border p-7 ${p.popular ? "border-iris bg-[#ebe7f2]" : "border-ink/15 bg-cream"}`}
            >
              <div className="flex flex-wrap items-center justify-between gap-2">
                <h2 className="font-editorial text-3xl tracking-tight">
                  {p.name}
                </h2>
                {p.popular && (
                  <span className="rounded-full bg-iris px-3 py-1 text-[11px] font-medium text-white">
                    Most popular
                  </span>
                )}
              </div>
              <p className="mt-6">
                <span className="font-editorial text-5xl tracking-tight">
                  {money(p.price)}
                </span>
                <span className="text-sm text-muted"> / month</span>
              </p>
              <p className="mt-3 text-sm font-medium">
                {count(p.included)} emails a month
              </p>
              <p className="mt-1 text-sm text-muted">
                {p.overage
                  ? `then $${p.overage.toFixed(2)} per 1,000`
                  : "hard cap"}
              </p>
              <p className="mt-5 flex-1 text-sm leading-relaxed text-muted">
                {p.note}
              </p>
              <a
                href={appLink(
                  p.price === 0 ? "/auth/register" : "/settings/billing",
                )}
                className={
                  p.popular
                    ? `${primaryButton} mt-7`
                    : "mt-7 inline-flex min-h-12 items-center justify-center gap-2 rounded-lg border border-ink/20 px-6 py-3 text-sm font-semibold hover:border-iris hover:text-iris"
                }
              >
                {p.price === 0 ? "Start free" : `Choose ${p.name}`}{" "}
                <ArrowUpRight size={16} />
              </a>
            </article>
          ))}
        </section>

        <section className="mx-auto max-w-[1160px] px-5 py-14">
          <h2 className="font-editorial text-3xl tracking-tight md:text-4xl">
            Included in every plan
          </h2>
          <ul className="mt-6 grid gap-3 text-sm sm:grid-cols-2 lg:grid-cols-4">
            {everyPlan.map((item) => (
              <li key={item} className="flex gap-2">
                <Check
                  size={16}
                  className="mt-0.5 shrink-0 text-iris"
                  aria-hidden="true"
                />
                {item}
              </li>
            ))}
          </ul>
        </section>

        <section
          aria-labelledby="compare-title"
          className="bg-[#ebe7f2] px-5 py-16 md:py-24"
        >
          <div className="mx-auto max-w-[1160px]">
            <h2
              id="compare-title"
              className="font-editorial text-4xl tracking-tight md:text-5xl"
            >
              A weekly newsletter to 10,000 readers.
            </h2>
            <p className="mt-4 max-w-2xl text-sm leading-relaxed text-muted">
              That is about 40,000 emails a month. Monthly prices on monthly
              billing, checked October 30, 2026; see{" "}
              <Link
                href="/blog/mailchimp-alternative"
                className="text-iris underline"
              >
                our comparison
              </Link>{" "}
              for sources.
            </p>
            <div className="mt-8 overflow-x-auto rounded-2xl border border-ink/10 bg-cream">
              <table className="w-full text-left text-sm">
                <thead className="border-b border-ink/10 text-xs uppercase tracking-wider text-muted">
                  <tr>
                    <th className="px-5 py-4 font-medium">Tool</th>
                    <th className="px-5 py-4 font-medium">Per month</th>
                    <th className="hidden px-5 py-4 font-medium sm:table-cell">
                      How it&apos;s priced
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {comparison.map((row, i) => (
                    <tr
                      key={row.name}
                      className={i ? "border-t border-ink/10" : "font-medium"}
                    >
                      <td className="px-5 py-4">{row.name}</td>
                      <td className="px-5 py-4">{money(row.price)}</td>
                      <td className="hidden px-5 py-4 text-muted sm:table-cell">
                        {row.how}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </section>

        <section className="mx-auto grid max-w-[1160px] gap-6 px-5 py-16 md:grid-cols-[1.2fr_1fr] md:items-center md:py-24">
          <div>
            <p className="text-xs uppercase tracking-widest text-iris">
              Open source
            </p>
            <h2 className="mt-4 font-editorial text-4xl tracking-tight md:text-5xl">
              Or run it yourself, free.
            </h2>
            <p className="mt-5 max-w-xl text-sm leading-relaxed text-muted">
              Xem is GPL-3.0. Install it on your own server with one command and
              pay only your email provider: through Amazon SES that is $0.10 to
              $0.16 per 1,000 emails.
            </p>
          </div>
          <div className="flex flex-wrap gap-5 md:justify-end">
            <a
              href={`${docsUrl}/guides/self-hosting`}
              className={primaryButton}
            >
              Self-host Xem <ArrowUpRight size={17} />
            </a>
            <Link
              href="/blog/amazon-ses-newsletter-cost"
              className="self-center text-sm underline underline-offset-4"
            >
              What SES costs
            </Link>
          </div>
        </section>

        <section
          aria-labelledby="faq-title"
          className="mx-auto max-w-3xl px-5 pb-20"
        >
          <h2
            id="faq-title"
            className="mb-8 font-editorial text-4xl tracking-tight"
          >
            Pricing questions.
          </h2>
          {faqs.map((f) => (
            <details key={f.q} className="border-b border-ink/15 py-5">
              <summary className="cursor-pointer text-base font-medium">
                {f.q}
              </summary>
              <p className="mt-4 text-sm leading-relaxed text-muted">{f.a}</p>
            </details>
          ))}
        </section>
      </main>
      <SiteFooter />
    </>
  );
}
