import type { Metadata } from "next";
import Link from "next/link";
import { ArrowUpRight } from "lucide-react";
import { SiteHeader, SiteFooter } from "@/components/site-chrome";
import { CostCalculator } from "@/components/cost-calculator";
import { appLink } from "@/lib/site";
import { primaryButton } from "@/lib/brand";
import { checked, competitors, sizes } from "@/lib/compare";
import { weeklyEmails, xemMonthlyCost } from "@/lib/xem-pricing";

const title =
  "Compare Xem with Mailchimp, Substack, beehiiv, Kit and more | Xem";
const description =
  "What a weekly newsletter costs on Xem and on Mailchimp, Substack, beehiiv, Kit, MailerLite, Brevo and Ghost, at 10,000, 50,000 and 100,000 subscribers.";
export const metadata: Metadata = {
  title,
  description,
  alternates: { canonical: "/compare" },
  openGraph: { type: "website", url: "/compare", title, description },
  twitter: { card: "summary", title, description },
};

const money = (n: number | string) =>
  typeof n === "number" ? `$${Math.round(n).toLocaleString("en-US")}` : n;
const count = (n: number) => n.toLocaleString("en-US");

export default function CompareIndex() {
  return (
    <>
      <SiteHeader active="none" />
      <main id="main">
        <section className="mx-auto max-w-[1160px] px-5 pb-10 pt-16 md:pt-24">
          <p className="text-xs font-medium uppercase tracking-[0.17em] text-iris">
            Compare
          </p>
          <h1 className="mt-5 max-w-4xl font-editorial text-[clamp(2.75rem,6vw,5.5rem)] leading-[1.02] tracking-[-0.045em]">
            Most tools charge for your list.
            <br />
            <em>Xem charges for your sends.</em>
          </h1>
          <p className="mt-7 max-w-2xl text-lg leading-relaxed text-muted">
            Work out what your newsletter costs on Xem, then see how that
            compares with the tools you might be using now.
          </p>
        </section>

        <section
          aria-label="Cost calculator"
          className="mx-auto max-w-[1160px] px-5 pb-16"
        >
          <CostCalculator />
        </section>

        <section
          aria-labelledby="table-title"
          className="bg-[#ebe7f2] px-5 py-16 md:py-20"
        >
          <div className="mx-auto max-w-[1160px]">
            <h2
              id="table-title"
              className="font-editorial text-4xl tracking-tight md:text-5xl"
            >
              A weekly newsletter, every tool.
            </h2>
            <p className="mt-4 max-w-2xl text-sm leading-relaxed text-muted">
              Monthly cost in US dollars on monthly billing, four sends a month
              to the whole list, on each tool&apos;s cheapest plan for
              newsletters. Checked {checked}; each comparison page lists its
              sources.
            </p>
            <div className="mt-8 overflow-x-auto rounded-2xl border border-ink/10 bg-cream">
              <table className="w-full min-w-[640px] text-left text-sm">
                <thead className="border-b border-ink/10 text-xs uppercase tracking-wider text-muted">
                  <tr>
                    <th className="px-5 py-4 font-medium">Tool</th>
                    {sizes.map((n) => (
                      <th key={n} className="px-5 py-4 font-medium">
                        {count(n)}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  <tr className="font-medium">
                    <td className="px-5 py-4">Xem</td>
                    {sizes.map((n) => (
                      <td key={n} className="px-5 py-4">
                        {money(xemMonthlyCost(weeklyEmails(n)).cost)}
                      </td>
                    ))}
                  </tr>
                  {competitors.map((c) => (
                    <tr key={c.id} className="border-t border-ink/10">
                      <td className="px-5 py-4">
                        <Link
                          href={`/compare/xem-vs-${c.id}`}
                          className="underline underline-offset-4 hover:text-iris"
                        >
                          {c.name}
                        </Link>
                      </td>
                      {c.prices.map((p, i) => (
                        <td key={i} className="px-5 py-4 text-muted">
                          {money(p)}
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </section>

        <section className="mx-auto max-w-[1160px] px-5 py-16">
          <h2 className="font-editorial text-3xl tracking-tight md:text-4xl">
            Moving from another tool?
          </h2>
          <ul className="mt-6 flex flex-wrap gap-3">
            {competitors.map((c) => (
              <li key={c.id}>
                <Link
                  href={`/migrate/${c.id}`}
                  className="inline-flex rounded-full border border-ink/20 px-4 py-2 text-sm hover:border-iris hover:text-iris"
                >
                  From {c.name}
                </Link>
              </li>
            ))}
          </ul>
          <a
            href={appLink("/auth/register")}
            className={`${primaryButton} mt-10`}
          >
            Start free <ArrowUpRight size={17} />
          </a>
        </section>
      </main>
      <SiteFooter />
    </>
  );
}
