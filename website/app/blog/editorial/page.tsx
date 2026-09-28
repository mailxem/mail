import type { Metadata } from "next";
import Link from "next/link";
import { SiteHeader, SiteFooter } from "@/components/site-chrome";
export const metadata: Metadata = {
  title: "Our editorial approach | Xem Journal",
  description:
    "How the Xem journal chooses topics, uses AI assistance, checks sources, and explains the limits of email marketing data.",
  alternates: { canonical: "/blog/editorial" },
  openGraph: {
    type: "website",
    url: "/blog/editorial",
    title: "Our editorial approach | Xem",
    description:
      "How the Xem journal researches, writes, and checks its guides.",
    images: ["/images/blog/ai-email-prompts.webp"],
  },
  twitter: {
    card: "summary_large_image",
    title: "Our editorial approach | Xem",
    description:
      "How the Xem journal researches, writes, and checks its guides.",
    images: ["/images/blog/ai-email-prompts.webp"],
  },
};
export default function EditorialPage() {
  return (
    <>
      <SiteHeader />
      <main id="main" className="mx-auto max-w-3xl px-5 py-20">
        <p className="text-xs uppercase tracking-widest text-iris">
          Behind the journal
        </p>
        <h1 className="mt-5 font-editorial text-6xl tracking-tight">
          Useful beats louder.
        </h1>
        <div className="prose mt-10 max-w-none prose-headings:font-editorial prose-headings:font-normal prose-a:text-iris">
          <p>
            The Xem journal is a collection of practical guides for people
            writing, sending, and measuring email. Our goal is to give you a
            useful next step, with enough context to understand the tradeoffs.
          </p>
          <h2>How these articles are made</h2>
          <p>
            This launch collection was produced with AI assistance. It is
            published under Xem editorial, not a fictional individual author. We
            check technical claims against the sources linked in each article,
            use original examples, and avoid invented first-hand experience,
            research results, or customer quotes.
          </p>
          <h2>Sources and dates</h2>
          <p>
            Provider documentation takes priority for sender requirements.
            Research for the launch collection was checked on September 10,
            2026. Older sources are used for established principles, not
            presented as new research. Requirements can change; check the linked
            provider page before changing a sending system.
          </p>
          <h2>Examples and product claims</h2>
          <p>
            Sample campaigns, numbers, and workflows explain an idea. They are
            not customer results. When a guide describes a general marketing
            practice rather than a current Xem capability, it says so. Open
            tracking is directional, SMTP acceptance is not inbox placement, and
            an observed click is not proof of human attention.
          </p>
          <h2>Images and corrections</h2>
          <p>
            Journal covers are original Xem editorial illustrations, created
            with AI assistance using our brand colors and typography. Charts and
            email layouts in the artwork are illustrative, not customer data. To
            propose a correction, use the issue tracker linked from the{" "}
            <a href="https://github.com/mailxem/xem-app.ts/issues">
              Xem repository
            </a>{" "}
            and include the article URL and a source.
          </p>
          <p>
            <Link href="/blog">Explore the journal →</Link>
          </p>
        </div>
      </main>
      <SiteFooter />
    </>
  );
}
