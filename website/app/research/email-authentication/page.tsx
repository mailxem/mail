import type { Metadata } from "next";
import Link from "next/link";
import { SiteHeader, SiteFooter } from "@/components/site-chrome";
import { siteUrl } from "@/lib/site";
import study from "@/lib/research/email-auth-2026.json";

const m = study.results.mail;
const top100 = study.results.tiers["100"];
const top1k = study.results.tiers["1000"];
const unprotected = Math.round((100 - m.dmarcEnforced) * 10) / 10;
// Share of published DMARC records that only monitor (p=none).
const monitorOnly = Math.round((m.dmarcPolicy.none / m.dmarc) * 100);
const scannedOn = new Date(study.scannedAt).toLocaleDateString("en-US", {
  month: "long",
  day: "numeric",
  year: "numeric",
  timeZone: "UTC",
});

const title = `Email authentication on the top 10,000 websites (${study.year}) | Xem research`;
const description = `We checked SPF, DMARC, BIMI and MTA-STS for the 10,000 most popular websites. ${unprotected}% of those that receive email don't ask Gmail, Outlook and Yahoo to stop mail forged in their name.`;
const path = "/research/email-authentication";

export const metadata: Metadata = {
  title,
  description,
  alternates: { canonical: path },
  openGraph: { type: "article", url: path, title, description },
  twitter: { card: "summary_large_image", title, description },
};

function Bar({
  label,
  value,
  note,
}: {
  label: string;
  value: number;
  note?: string;
}) {
  return (
    <div className="grid grid-cols-[minmax(0,9rem)_1fr_3.5rem] items-center gap-3 text-sm sm:grid-cols-[12rem_1fr_4rem]">
      <span className="truncate">{label}</span>
      <span className="h-3 overflow-hidden rounded-full bg-ink/10" aria-hidden>
        <span
          className="block h-full rounded-full bg-iris"
          style={{ width: `${Math.min(100, value)}%` }}
        />
      </span>
      <span className="text-right tabular-nums">
        {value}%{note ? <span className="sr-only"> {note}</span> : null}
      </span>
    </div>
  );
}

function Stat({ value, label }: { value: string; label: string }) {
  return (
    <div className="rounded-2xl border border-ink/10 bg-white/60 p-5">
      <p className="font-editorial text-5xl tracking-tight">{value}</p>
      <p className="mt-2 text-sm leading-relaxed text-muted">{label}</p>
    </div>
  );
}

function Section({
  id,
  title,
  children,
}: {
  id: string;
  title: string;
  children: React.ReactNode;
}) {
  return (
    <section aria-labelledby={id} className="mx-auto max-w-3xl px-5 py-12">
      <h2 id={id} className="font-editorial text-4xl tracking-tight">
        {title}
      </h2>
      <div className="mt-6 space-y-5 text-[17px] leading-relaxed text-ink/85">
        {children}
      </div>
    </section>
  );
}

export default function EmailAuthenticationResearch() {
  const citation = `Xem (${study.year}). Email authentication on the top 10,000 websites. ${siteUrl}${path}`;
  return (
    <>
      <SiteHeader active="none" />
      <main id="main">
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{
            __html: JSON.stringify({
              "@context": "https://schema.org",
              "@type": "Dataset",
              name: "Email authentication on the top 10,000 websites",
              description,
              url: `${siteUrl}${path}`,
              creator: { "@type": "Organization", name: "Xem", url: siteUrl },
              datePublished: study.publishedAt,
              temporalCoverage: study.scannedAt.slice(0, 10),
              license: "https://creativecommons.org/licenses/by/4.0/",
              isBasedOn: study.source.url,
              distribution: {
                "@type": "DataDownload",
                encodingFormat: "application/json",
                contentUrl: `${siteUrl}/research/email-authentication-${study.year}.json`,
              },
            }).replace(/</g, "\\u003c"),
          }}
        />
        <header className="mx-auto max-w-[1160px] px-5 pb-10 pt-16 md:pt-24">
          <p className="text-xs font-medium uppercase tracking-[0.17em] text-iris">
            Xem research · {scannedOn}
          </p>
          <h1 className="mt-5 max-w-4xl font-editorial text-[clamp(2.5rem,5.5vw,5rem)] leading-[1.03] tracking-[-0.04em]">
            {Math.round(unprotected)}% of the web&apos;s top sites don&apos;t
            ask Gmail to stop email forged in their name.
          </h1>
          <p className="mt-7 max-w-2xl text-lg leading-relaxed text-muted">
            We checked the email authentication records of the 10,000 most
            popular websites. Of the{" "}
            {study.results.receivesMail.toLocaleString()} that receive email,{" "}
            {m.dmarc}% publish a DMARC record, but only {m.dmarcEnforced}% tell
            mailbox providers to quarantine or reject mail that fails it.
          </p>
        </header>

        <section
          aria-label="Key findings"
          className="mx-auto grid max-w-[1160px] gap-4 px-5 pb-6 sm:grid-cols-2 lg:grid-cols-4"
        >
          <Stat
            value={`${m.spf}%`}
            label="of sites that receive email publish an SPF record"
          />
          <Stat
            value={`${m.dmarc}%`}
            label="publish a DMARC record of any kind"
          />
          <Stat
            value={`${m.dmarcReject}%`}
            label="set DMARC to reject, the strongest protection"
          />
          <Stat
            value={`${m.bimi}%`}
            label="publish a BIMI record to show their logo in the inbox"
          />
        </section>

        <Section
          id="dmarc"
          title={`${monitorOnly}% of DMARC records don't stop anything`}
        >
          <p>
            DMARC tells Gmail, Outlook and Yahoo what to do with mail that
            claims to come from your domain but fails SPF and DKIM. A policy of{" "}
            <code>p=none</code> only asks for reports and tells providers to
            take no action on failing mail. Only <code>quarantine</code> and{" "}
            <code>reject</code> ask them to keep forged mail away from readers.
          </p>
          <div className="space-y-3 rounded-2xl border border-ink/10 bg-white/60 p-5">
            <p className="text-sm font-medium">
              DMARC policy, sites that receive email
            </p>
            <Bar
              label="No DMARC record"
              value={Math.round((100 - m.dmarc) * 10) / 10}
            />
            <Bar label="p=none (monitor only)" value={m.dmarcPolicy.none} />
            <Bar label="p=quarantine" value={m.dmarcPolicy.quarantine} />
            <Bar label="p=reject" value={m.dmarcPolicy.reject} />
          </div>
          <p>
            {m.dmarcReports}% of DMARC records ask for aggregate reports (a{" "}
            <code>rua</code> address), the first step towards enforcing. Among
            sites that do enforce, {m.dmarcPartial}% apply the policy to only
            part of their failing mail (<code>pct</code> below 100).
          </p>
        </Section>

        <Section id="tiers" title="Protection drops off outside the top 1,000">
          <p>
            The top 100 and the top 1,000 sites enforce DMARC at about the same
            rate. Below that, enforcement falls: across the full top 10,000,{" "}
            {Math.round(100 - m.dmarcEnforced)}% of sites that receive email
            don&apos;t enforce it.
          </p>
          <div className="space-y-3 rounded-2xl border border-ink/10 bg-white/60 p-5">
            <p className="text-sm font-medium">
              DMARC set to quarantine or reject
            </p>
            <Bar
              label={`Top 100 (${top100.domains} with email)`}
              value={top100.dmarcEnforced}
            />
            <Bar
              label={`Top 1,000 (${top1k.domains.toLocaleString()} with email)`}
              value={top1k.dmarcEnforced}
            />
            <Bar
              label={`Top 10,000 (${m.domains.toLocaleString()} with email)`}
              value={m.dmarcEnforced}
            />
          </div>
        </Section>

        <Section id="spf" title="SPF is nearly universal, but often soft">
          <p>
            {m.spf}% of sites that receive email publish SPF, the list of
            servers allowed to send for them. How the record ends decides what
            happens to everything else: <code>-all</code> says fail it,{" "}
            <code>~all</code> says treat it as suspicious, and <code>?all</code>{" "}
            says nothing at all.
          </p>
          <div className="space-y-3 rounded-2xl border border-ink/10 bg-white/60 p-5">
            <p className="text-sm font-medium">
              How SPF records end, among sites with SPF
            </p>
            {Object.entries(m.spfAll)
              .filter(([, v]) => v > 0)
              .map(([k, v]) => (
                <Bar
                  key={k}
                  label={
                    {
                      "-all": "-all (fail)",
                      "~all": "~all (soft fail)",
                      "?all": "?all (neutral)",
                      "+all": "+all (allow anyone)",
                      redirect: "redirect= (another domain's)",
                      none: "No all mechanism",
                      multiple: "Two or more SPF records (invalid)",
                    }[k] ?? k
                  }
                  value={v}
                />
              ))}
          </div>
        </Section>

        <Section
          id="parked"
          title="Domains that never send mail are the easiest to fake"
        >
          <p>
            {study.results.noMail.domains.toLocaleString()} of the 10,000
            domains publish no MX record. They can still be spoofed unless they
            say they never send it: an SPF record of <code>v=spf1 -all</code>{" "}
            and a DMARC policy of <code>p=reject</code>. Only{" "}
            {study.results.noMail.lockedDown}% of them do both.
          </p>
        </Section>

        <Section id="newer" title="The newer standards are rare">
          <p>
            BIMI, which shows a verified brand logo next to your mail in Gmail,
            Apple Mail and Yahoo, appears on {m.bimi}% of sites that receive
            email. MTA-STS, which forces encrypted delivery to your own mail
            servers, appears on {m.mtaSts}%.
          </p>
        </Section>

        <section className="mx-auto max-w-3xl px-5 py-12">
          <div className="rounded-2xl bg-ink p-8 text-cream">
            <h2 className="font-editorial text-3xl tracking-tight">
              Check your own domain
            </h2>
            <p className="mt-3 max-w-xl text-sm leading-relaxed text-cream/80">
              Xem&apos;s free checker runs the same lookups for your domain, in
              your browser, and gives you the exact record to add for anything
              missing.
            </p>
            <Link
              href="/tools/email-checker"
              className="mt-5 inline-block rounded-full bg-cream px-5 py-2.5 text-sm font-medium text-ink"
            >
              Check your domain
            </Link>
          </div>
        </section>

        <Section id="method" title="How we did this">
          <p>
            We took the top 10,000 domains of the Tranco list{" "}
            <a href={study.source.url} className="underline underline-offset-4">
              {study.source.id}
            </a>
            , a research ranking that averages several popularity lists. On{" "}
            {scannedOn} we looked up each domain&apos;s MX records and the TXT
            records for SPF (the domain itself), DMARC (<code>_dmarc</code>),
            BIMI (<code>default._bimi</code>) and MTA-STS (<code>_mta-sts</code>
            ) through Cloudflare, Google and Quad9 public resolvers, retrying
            failed lookups twice. Domains with any failed lookup were then
            scanned again, more slowly, the same day.
          </p>
          <p>
            A domain &ldquo;receives email&rdquo; if it has at least one MX
            record that isn&apos;t a null MX. Lookups that still failed (
            {study.results.errors.toLocaleString()} domains had at least one)
            are left out of that check rather than counted as missing. A domain
            with two DMARC records is counted as invalid, as mailbox providers
            ignore both. We didn&apos;t check DKIM, because its selectors
            can&apos;t be discovered from DNS alone, and we read only public
            DNS: no mail was sent.
          </p>
          <p>
            The aggregate numbers are{" "}
            <a
              href={`/research/email-authentication-${study.year}.json`}
              className="underline underline-offset-4"
            >
              available as JSON
            </a>{" "}
            under CC BY 4.0. Please cite and link this page if you use them.
          </p>
          <div className="rounded-xl border border-ink/15 bg-white/60 p-4 text-sm">
            <p className="text-xs font-medium uppercase tracking-[0.14em] text-muted">
              Cite this research
            </p>
            <p className="mt-2 select-all">{citation}</p>
          </div>
        </Section>
      </main>
      <SiteFooter />
    </>
  );
}
