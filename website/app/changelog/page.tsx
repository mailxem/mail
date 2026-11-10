import type { Metadata } from "next";
import Link from "next/link";
import { SiteHeader, SiteFooter } from "@/components/site-chrome";
import { changes } from "@/lib/changelog";
import { githubUrl } from "@/lib/site";

const title = "Changelog | Xem";
const description =
  "What's new in Xem, the open-source email marketing tool: every feature and fix as it ships, newest first.";
export const metadata: Metadata = {
  title,
  description,
  alternates: { canonical: "/changelog" },
  openGraph: { type: "website", url: "/changelog", title, description },
  twitter: { card: "summary", title, description },
};

const day = (d: string) =>
  new Date(`${d}T12:00:00Z`).toLocaleDateString("en-US", {
    month: "long",
    day: "numeric",
    year: "numeric",
  });

export default function ChangelogPage() {
  const dates = [...new Set(changes.map((c) => c.date))];
  return (
    <>
      <SiteHeader active="none" />
      <main
        id="main"
        className="mx-auto max-w-[760px] px-5 pb-20 pt-16 md:pt-24"
      >
        <h1 className="font-editorial text-[clamp(2.75rem,6vw,4.5rem)] leading-[1.02] tracking-[-0.04em]">
          Changelog
        </h1>
        <p className="mt-5 text-lg leading-relaxed text-ink/75">
          What&apos;s new in Xem, newest first. Every entry links to the code
          that shipped it, on{" "}
          <a href={githubUrl} className="underline underline-offset-4">
            GitHub
          </a>
          .
        </p>
        <div className="mt-14 space-y-14">
          {dates.map((date) => (
            <section key={date} aria-labelledby={`d-${date}`}>
              <h2 id={`d-${date}`} className="text-sm font-medium text-muted">
                <time dateTime={date}>{day(date)}</time>
              </h2>
              <ul className="mt-5 space-y-8">
                {changes
                  .filter((c) => c.date === date)
                  .map((c) => (
                    <li key={c.pr}>
                      <h3 className="font-editorial text-[1.7rem] leading-tight tracking-[-0.02em]">
                        {c.href ? (
                          <Link href={c.href} className="hover:underline">
                            {c.title}
                          </Link>
                        ) : (
                          c.title
                        )}
                      </h3>
                      <p className="mt-2 leading-relaxed text-ink/80">
                        {c.body}
                      </p>
                      <a
                        href={`${githubUrl}/pull/${c.pr}`}
                        className="mt-2 inline-block text-sm text-muted underline underline-offset-4 hover:text-ink"
                      >
                        Pull request #{c.pr}
                      </a>
                    </li>
                  ))}
              </ul>
            </section>
          ))}
        </div>
      </main>
      <SiteFooter />
    </>
  );
}
