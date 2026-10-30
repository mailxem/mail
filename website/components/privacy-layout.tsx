import type { ReactNode } from "react";
import Link from "next/link";
import { ArrowUpRight, Mail } from "lucide-react";
import { SiteFooter, SiteHeader } from "@/components/site-chrome";
import { privacyContact, privacyEmailLink } from "@/lib/privacy";

export type PrivacySection = { id: string; title: string; content: ReactNode };

export function PrivacyLayout({
  eyebrow,
  title,
  description,
  summary,
  sections,
}: {
  eyebrow: string;
  title: string;
  description: string;
  summary: ReactNode;
  sections: PrivacySection[];
}) {
  return (
    <>
      <SiteHeader active="none" />
      <main
        id="main"
        className="mx-auto max-w-[1160px] px-5 pb-20 pt-16 md:pt-24"
      >
        <div className="max-w-3xl">
          <p className="text-xs font-medium uppercase tracking-[0.18em] text-iris">
            {eyebrow}
          </p>
          <h1 className="mt-5 font-editorial text-5xl leading-[1.05] tracking-tight sm:text-7xl">
            {title}
          </h1>
          <p className="mt-6 max-w-2xl text-lg leading-relaxed text-muted">
            {description}
          </p>
        </div>
        <aside
          aria-label="Document status"
          className="mt-9 rounded-xl border border-iris/25 bg-lavender/40 px-5 py-4 text-sm leading-relaxed"
        >
          <p className="font-semibold">Draft for review · September 28, 2026</p>
          <p className="mt-1 text-ink/80">
            The privacy commitments and retention deadlines on these pages are
            proposed and are not yet in effect. We are validating the supporting
            processes before publication. Connected Google mail is in limited
            testing; Google verification is not complete.
          </p>
        </aside>
        <div className="mt-12 grid gap-12 lg:grid-cols-[220px_minmax(0,1fr)] lg:gap-16">
          <aside className="min-w-0">
            <div className="lg:sticky lg:top-8">
              <nav
                aria-label="On this page"
                className="border-l border-ink/15 pl-5"
              >
                <p className="mb-4 text-xs font-semibold uppercase tracking-wider text-muted">
                  On this page
                </p>
                <ol className="space-y-3 text-sm leading-relaxed">
                  {sections.map((section) => (
                    <li key={section.id}>
                      <a
                        href={`#${section.id}`}
                        className="hover:text-iris hover:underline"
                      >
                        {section.title}
                      </a>
                    </li>
                  ))}
                </ol>
              </nav>
              <div className="mt-8 rounded-xl border border-ink/15 p-5">
                <Mail size={20} aria-hidden="true" className="text-iris" />
                <p className="mt-3 text-sm font-medium">A person to talk to.</p>
                <a
                  href={privacyEmailLink}
                  className="mt-2 inline-flex min-h-8 items-center gap-1 break-all text-sm text-iris underline underline-offset-4"
                >
                  {privacyContact}
                  <ArrowUpRight size={13} aria-hidden="true" />
                </a>
              </div>
            </div>
          </aside>
          <div className="min-w-0 max-w-3xl">
            <div className="rounded-2xl border border-ink/10 bg-white/60 p-6 text-base leading-relaxed sm:p-8">
              {summary}
            </div>
            <div className="mt-12 space-y-12">
              {sections.map((section) => (
                <section
                  key={section.id}
                  id={section.id}
                  aria-labelledby={`${section.id}-heading`}
                  className="scroll-mt-8"
                >
                  <h2
                    id={`${section.id}-heading`}
                    className="font-editorial text-3xl tracking-tight sm:text-4xl"
                  >
                    {section.title}
                  </h2>
                  <div className="prose mt-5 max-w-none break-words text-ink/85 prose-headings:font-editorial prose-a:text-iris prose-a:decoration-iris/40 prose-a:underline-offset-4 prose-strong:text-ink">
                    {section.content}
                  </div>
                </section>
              ))}
            </div>
            <nav
              aria-label="Privacy resources"
              className="mt-14 flex flex-wrap gap-x-6 gap-y-3 border-t border-ink/15 pt-6 text-sm text-iris"
            >
              <Link href="/privacy" className="underline underline-offset-4">
                Privacy policy
              </Link>
              <Link
                href="/data-deletion"
                className="underline underline-offset-4"
              >
                Delete your data
              </Link>
              <Link
                href="/google-mail"
                className="underline underline-offset-4"
              >
                Google mail & your data
              </Link>
            </nav>
          </div>
        </div>
      </main>
      <SiteFooter />
    </>
  );
}
