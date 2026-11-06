import type { Metadata } from "next";
import { SiteHeader, SiteFooter } from "@/components/site-chrome";
import { EmailChecker } from "@/components/email-checker";

const title = "Email deliverability checker: SPF, DKIM, DMARC and MX | Xem";
const description =
  "Free check of the DNS records Gmail, Yahoo and Outlook look at before accepting your email: MX, SPF and its 10-lookup limit, DMARC and DKIM, with the exact record to fix each problem.";
export const metadata: Metadata = {
  title,
  description,
  alternates: { canonical: "/tools/email-checker" },
  openGraph: {
    type: "website",
    url: "/tools/email-checker",
    title,
    description,
  },
  twitter: { card: "summary", title, description },
};

const sources = [
  {
    name: "Gmail email sender guidelines",
    href: "https://support.google.com/a/answer/81126",
  },
  {
    name: "Yahoo sender best practices",
    href: "https://senders.yahooinc.com/best-practices/",
  },
  {
    name: "RFC 7208: SPF and its 10-lookup limit",
    href: "https://www.rfc-editor.org/rfc/rfc7208#section-4.6.4",
  },
  {
    name: "RFC 7489: DMARC",
    href: "https://www.rfc-editor.org/rfc/rfc7489",
  },
  {
    name: "Amazon SES Easy DKIM",
    href: "https://docs.aws.amazon.com/ses/latest/dg/send-email-authentication-dkim-easy.html",
  },
];

const faqs = [
  {
    q: "Where do I find my DKIM selector?",
    a: "Open a message you sent from the domain, show its original headers, and find the DKIM-Signature line. The s= value is the selector and d= is the domain it signs for. If you leave the field empty, the checker tries google, selector1, selector2, k1, s1, s2 and resend.",
  },
  {
    q: "I send with Amazon SES. Why isn't my DKIM key found?",
    a: "SES Easy DKIM gives each domain three random selectors and asks you to publish them as CNAME records like abcd1234._domainkey.example.com pointing to abcd1234.dkim.amazonses.com. No list of common selectors can guess them. Copy one of the three tokens from the SES console (or the s= value of a sent message) into the selector field; the lookup follows the CNAME to the key.",
  },
  {
    q: "Does passing every check mean my email reaches the inbox?",
    a: "No. These records prove the mail is really from you, which Gmail and Yahoo require before anything else. After that, inbox placement depends on whether people want your email: low spam complaints, a working one-click unsubscribe, and sending only to people who signed up.",
  },
  {
    q: "What isn't checked here?",
    a: "Alignment between your From address and the domains your provider signs and bounces with, reverse DNS of your sending IPs, blocklists and the one-click unsubscribe header. Those depend on the messages themselves, not just your DNS. A DMARC report (the rua address) shows alignment for real mail.",
  },
];

export default function EmailCheckerPage() {
  return (
    <>
      <SiteHeader active="none" />
      <main id="main">
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{
            __html: JSON.stringify({
              "@context": "https://schema.org",
              "@type": "FAQPage",
              mainEntity: faqs.map((f) => ({
                "@type": "Question",
                name: f.q,
                acceptedAnswer: { "@type": "Answer", text: f.a },
              })),
            }).replace(/</g, "\\u003c"),
          }}
        />
        <section className="mx-auto max-w-[1160px] px-5 pb-10 pt-16 md:pt-24">
          <p className="text-xs font-medium uppercase tracking-[0.17em] text-iris">
            Free tool
          </p>
          <h1 className="mt-5 max-w-4xl font-editorial text-[clamp(2.75rem,6vw,5.5rem)] leading-[1.02] tracking-[-0.045em]">
            Will Gmail accept your email?
            <br />
            <em>Check your domain.</em>
          </h1>
          <p className="mt-7 max-w-2xl text-lg leading-relaxed text-muted">
            Mailbox providers look at your MX, SPF, DKIM and DMARC records
            before they take bulk mail. Enter the domain you send from to see
            what they&apos;ll find, and the exact record to fix anything
            missing.
          </p>
        </section>

        <section
          aria-label="Checker"
          className="mx-auto max-w-[1160px] px-5 pb-16"
        >
          <EmailChecker />
          <p className="mt-6 max-w-2xl text-xs leading-relaxed text-muted">
            The lookups run in your browser, straight to Cloudflare&apos;s
            public DNS resolver. Xem has no server behind this page: we
            don&apos;t see, store or count the domains you check.
          </p>
        </section>

        <section
          aria-labelledby="sources-title"
          className="bg-[#ebe7f2] px-5 py-16 md:py-20"
        >
          <div className="mx-auto max-w-[1160px]">
            <h2
              id="sources-title"
              className="font-editorial text-4xl tracking-tight md:text-5xl"
            >
              Where the rules come from.
            </h2>
            <p className="mt-4 max-w-2xl text-sm leading-relaxed text-muted">
              Since February 2024, Gmail and Yahoo require SPF or DKIM from
              every sender, and SPF, DKIM and DMARC from bulk senders. The
              checks above follow their published requirements and the RFCs
              behind them.
            </p>
            <ul className="mt-6 grid gap-3 text-sm sm:grid-cols-2">
              {sources.map((s) => (
                <li key={s.href}>
                  <a
                    href={s.href}
                    className="underline underline-offset-4 hover:text-iris"
                  >
                    {s.name} ↗
                  </a>
                </li>
              ))}
            </ul>
          </div>
        </section>

        <section
          aria-labelledby="faq-title"
          className="mx-auto max-w-3xl px-5 py-16 md:py-20"
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
      </main>
      <SiteFooter />
    </>
  );
}
