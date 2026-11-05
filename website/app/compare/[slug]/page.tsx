import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { ArrowUpRight, Check } from "lucide-react";
import { SiteHeader, SiteFooter } from "@/components/site-chrome";
import { appLink } from "@/lib/site";
import { primaryButton } from "@/lib/brand";
import { checked, competitors, getCompetitor, sizes } from "@/lib/compare";
import { weeklyEmails, xemMonthlyCost } from "@/lib/xem-pricing";

export const dynamicParams = false;
export function generateStaticParams() {
  return competitors.map((c) => ({ slug: `xem-vs-${c.id}` }));
}
const find = (slug: string) => getCompetitor(slug.replace(/^xem-vs-/, ""));

export async function generateMetadata({
  params,
}: {
  params: Promise<{ slug: string }>;
}): Promise<Metadata> {
  const c = find((await params).slug);
  if (!c) return {};
  const title = `Xem vs ${c.name}: pricing and features compared | Xem`;
  const description = `${c.name} ${c.modelShort}. Xem charges per email sent. Compare monthly costs at 10,000, 50,000 and 100,000 subscribers, and see how to move your list.`;
  const path = `/compare/xem-vs-${c.id}`;
  return {
    title,
    description,
    alternates: { canonical: path },
    openGraph: { type: "article", url: path, title, description },
    twitter: { card: "summary", title, description },
  };
}

const count = (n: number) => n.toLocaleString("en-US");
const money = (n: number) => `$${Math.round(n).toLocaleString("en-US")}`;

export default async function ComparePage({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const c = find((await params).slug);
  if (!c) notFound();
  const faqs = [
    {
      q: `Is Xem cheaper than ${c.name}?`,
      a: c.cheaperAnswer,
    },
    {
      q: `Can I move my ${c.name} list to Xem?`,
      a: `${c.exportShort} Xem ${c.recognised ? `recognises ${c.name} exports` : "imports any CSV with an email column"} and keeps unsubscribed people unsubscribed.`,
    },
    {
      q: `When is ${c.name} the better choice?`,
      a: c.chooseThem.join(" "),
    },
  ];
  const structuredData = {
    "@context": "https://schema.org",
    "@type": "FAQPage",
    mainEntity: faqs.map((f) => ({
      "@type": "Question",
      name: f.q,
      acceptedAnswer: { "@type": "Answer", text: f.a },
    })),
  };
  return (
    <>
      <SiteHeader active="none" />
      <main id="main">
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{
            __html: JSON.stringify(structuredData).replace(/</g, "\\u003c"),
          }}
        />
        <section className="mx-auto max-w-[1160px] px-5 pb-10 pt-16 md:pt-24">
          <p className="text-xs font-medium uppercase tracking-[0.17em] text-iris">
            <Link href="/compare">Compare</Link> · Xem vs {c.name}
          </p>
          <h1 className="mt-5 max-w-4xl font-editorial text-[clamp(2.75rem,6vw,5.5rem)] leading-[1.02] tracking-[-0.045em]">
            Xem vs {c.name}.
            <br />
            <em>{c.headline}</em>
          </h1>
          <p className="mt-7 max-w-2xl text-lg leading-relaxed text-muted">
            {c.summary}
          </p>
          <div className="mt-8 flex flex-wrap items-center gap-5">
            <a href={appLink("/auth/register")} className={primaryButton}>
              Start free <ArrowUpRight size={17} />
            </a>
            <Link
              href={`/migrate/${c.id}`}
              className="text-sm underline underline-offset-4"
            >
              Move from {c.name}
            </Link>
          </div>
        </section>

        <section
          aria-labelledby="cost-title"
          className="bg-[#ebe7f2] px-5 py-16 md:py-20"
        >
          <div className="mx-auto max-w-[1160px]">
            <h2
              id="cost-title"
              className="font-editorial text-4xl tracking-tight md:text-5xl"
            >
              A weekly newsletter, priced.
            </h2>
            <p className="mt-4 max-w-2xl text-sm leading-relaxed text-muted">
              Four sends a month to the whole list, on monthly billing in US
              dollars. {c.name} on {c.plan}; Xem on its cheapest plan for that
              volume. {c.name} prices checked {checked}.
            </p>
            <div className="mt-8 overflow-x-auto rounded-2xl border border-ink/10 bg-cream">
              <table className="w-full text-left text-sm">
                <thead className="border-b border-ink/10 text-xs uppercase tracking-wider text-muted">
                  <tr>
                    <th className="px-5 py-4 font-medium">Subscribers</th>
                    <th className="px-5 py-4 font-medium">Emails a month</th>
                    <th className="px-5 py-4 font-medium">{c.name}</th>
                    <th className="px-5 py-4 font-medium">Xem</th>
                  </tr>
                </thead>
                <tbody>
                  {sizes.map((n, i) => {
                    const theirs = c.prices[i];
                    const ours = xemMonthlyCost(weeklyEmails(n));
                    return (
                      <tr
                        key={n}
                        className="border-t border-ink/10 first:border-t-0"
                      >
                        <td className="px-5 py-4 font-medium">{count(n)}</td>
                        <td className="px-5 py-4 text-muted">
                          {count(weeklyEmails(n))}
                        </td>
                        <td className="px-5 py-4">
                          {typeof theirs === "number" ? money(theirs) : theirs}
                        </td>
                        <td className="px-5 py-4 font-medium">
                          {money(ours.cost)}{" "}
                          <span className="font-normal text-muted">
                            · {ours.plan}
                          </span>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
            {c.priceNote && (
              <p className="mt-4 max-w-3xl text-xs leading-relaxed text-muted">
                {c.priceNote}
              </p>
            )}
          </div>
        </section>

        <section className="mx-auto grid max-w-[1160px] gap-10 px-5 py-16 md:grid-cols-2 md:py-20">
          <div>
            <h2 className="font-editorial text-3xl tracking-tight md:text-4xl">
              How they charge
            </h2>
            <dl className="mt-6 space-y-5 text-sm leading-relaxed">
              <div>
                <dt className="font-medium">{c.name}</dt>
                <dd className="mt-1 text-muted">{c.model}</dd>
              </div>
              <div>
                <dt className="font-medium">Xem</dt>
                <dd className="mt-1 text-muted">
                  Per email sent. Free for 3,000 a month, then from $9. Contacts
                  are free to keep, every plan has every feature, and you can
                  self-host the open-source version and pay only your email
                  provider.
                </dd>
              </div>
            </dl>
          </div>
          <div>
            <h2 className="font-editorial text-3xl tracking-tight md:text-4xl">
              When {c.name} fits better
            </h2>
            <ul className="mt-6 space-y-3 text-sm leading-relaxed text-muted">
              {c.chooseThem.map((item) => (
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
          </div>
        </section>

        <section
          aria-labelledby="faq-title"
          className="mx-auto max-w-3xl px-5 pb-12"
        >
          <h2
            id="faq-title"
            className="mb-8 font-editorial text-4xl tracking-tight"
          >
            Questions.
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

        <section className="mx-auto max-w-3xl px-5 pb-20 text-xs leading-relaxed text-muted">
          <h2 className="text-sm font-medium text-ink">Sources</h2>
          <ul className="mt-3 space-y-1">
            {c.sources.map((s) => (
              <li key={s}>
                <a
                  href={s}
                  className="underline underline-offset-2"
                  rel="nofollow noopener"
                  target="_blank"
                >
                  {s}
                </a>
              </li>
            ))}
            <li>
              <Link href="/pricing" className="underline underline-offset-2">
                Xem pricing
              </Link>
            </li>
          </ul>
          <p className="mt-3">
            Prices change. We check these pages by hand; tell us at{" "}
            <a href="mailto:team@xem.email" className="underline">
              team@xem.email
            </a>{" "}
            if something is out of date.
          </p>
        </section>
      </main>
      <SiteFooter />
    </>
  );
}
