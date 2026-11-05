import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { ArrowUpRight } from "lucide-react";
import { SiteHeader, SiteFooter } from "@/components/site-chrome";
import { appLink, docsUrl } from "@/lib/site";
import { primaryButton } from "@/lib/brand";
import { competitors, getCompetitor } from "@/lib/compare";

export const dynamicParams = false;
export function generateStaticParams() {
  return competitors.map((c) => ({ slug: c.id }));
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ slug: string }>;
}): Promise<Metadata> {
  const c = getCompetitor((await params).slug);
  if (!c) return {};
  const title = `Move from ${c.name} to Xem: export and import your list | Xem`;
  const description = `Export your subscribers from ${c.name}, import them into Xem and keep unsubscribes intact. Step by step, in about ten minutes.`;
  const path = `/migrate/${c.id}`;
  return {
    title,
    description,
    alternates: { canonical: path },
    openGraph: { type: "article", url: path, title, description },
    twitter: { card: "summary", title, description },
  };
}

export default async function MigratePage({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const c = getCompetitor((await params).slug);
  if (!c) notFound();
  const steps = [
    {
      title: `Export from ${c.name}`,
      body: c.exportSteps,
      link: c.exportHelp
        ? { href: c.exportHelp, label: `${c.name} help: exporting subscribers` }
        : null,
    },
    {
      title: "Create a free Xem workspace",
      body: "Sign up, then create a contact list for these subscribers. No card is needed and storing contacts is free.",
      link: { href: appLink("/auth/register"), label: "Create your workspace" },
    },
    {
      title: "Import the file",
      body: c.recognised
        ? `In Audience → Contact lists, open your list and choose Import. Xem recognises a ${c.name} export and maps its columns for you, including who has unsubscribed. Check the preview, then import.`
        : `In Audience → Contact lists, open your list and choose Import. Map the email column (and names, if you have them) in the preview, then import. People marked unsubscribed in the file stay unsubscribed.`,
      link: {
        href: `${docsUrl}/guides/import-contacts`,
        label: "Import guide",
      },
    },
    {
      title: "Connect your sending domain",
      body: "Verify your domain with Xem's managed sending, or connect your own SMTP or Amazon SES. Your first campaign can use one of 170+ editable templates.",
      link: {
        href: `${docsUrl}/guides/amazon-ses`,
        label: "Send with Amazon SES",
      },
    },
  ];
  return (
    <>
      <SiteHeader active="none" />
      <main id="main">
        <section className="mx-auto max-w-[1160px] px-5 pb-10 pt-16 md:pt-24">
          <p className="text-xs font-medium uppercase tracking-[0.17em] text-iris">
            Move to Xem · from {c.name}
          </p>
          <h1 className="mt-5 max-w-4xl font-editorial text-[clamp(2.75rem,6vw,5.5rem)] leading-[1.02] tracking-[-0.045em]">
            Bring your list from {c.name}.
            <br />
            <em>Keep every unsubscribe.</em>
          </h1>
          <p className="mt-7 max-w-2xl text-lg leading-relaxed text-muted">
            Export a CSV from {c.name}, import it into Xem, and send your next
            issue from your own domain. Contacts are free to store; you pay only
            for the emails you send.
          </p>
          <div className="mt-8 flex flex-wrap items-center gap-5">
            <a href={appLink("/auth/register")} className={primaryButton}>
              Start free <ArrowUpRight size={17} />
            </a>
            <Link
              href={`/compare/xem-vs-${c.id}`}
              className="text-sm underline underline-offset-4"
            >
              Compare Xem and {c.name}
            </Link>
          </div>
        </section>

        <section
          aria-labelledby="steps-title"
          className="mx-auto max-w-[1160px] px-5 pb-16"
        >
          <h2 id="steps-title" className="sr-only">
            Steps
          </h2>
          <ol className="grid gap-4 md:grid-cols-2">
            {steps.map((s, i) => (
              <li
                key={s.title}
                className="rounded-2xl border border-ink/15 bg-cream p-7"
              >
                <span className="font-editorial text-4xl tracking-tight text-iris">
                  {i + 1}
                </span>
                <h3 className="mt-3 text-lg font-semibold">{s.title}</h3>
                <p className="mt-2 text-sm leading-relaxed text-muted">
                  {s.body}
                </p>
                {s.link && (
                  <a
                    href={s.link.href}
                    className="mt-4 inline-flex items-center gap-1 text-sm text-iris underline underline-offset-4"
                  >
                    {s.link.label} <ArrowUpRight size={14} />
                  </a>
                )}
              </li>
            ))}
          </ol>
        </section>

        <section className="bg-[#ebe7f2] px-5 py-16">
          <div className="mx-auto grid max-w-[1160px] gap-10 md:grid-cols-2">
            <div>
              <h2 className="font-editorial text-3xl tracking-tight md:text-4xl">
                What comes with you
              </h2>
              <ul className="mt-5 list-disc space-y-2 pl-5 text-sm leading-relaxed text-muted">
                <li>Email addresses and names</li>
                <li>
                  Subscribed and unsubscribed status, so nobody is emailed again
                  by mistake
                </li>
                <li>
                  Any extra columns you map, such as tags or sign-up source
                </li>
              </ul>
            </div>
            <div>
              <h2 className="font-editorial text-3xl tracking-tight md:text-4xl">
                What to rebuild
              </h2>
              <ul className="mt-5 list-disc space-y-2 pl-5 text-sm leading-relaxed text-muted">
                <li>
                  Signup forms: publish a Xem form and swap the embed on your
                  site
                </li>
                <li>
                  Automations, such as a welcome sequence, in the visual builder
                </li>
                <li>
                  Past campaign stats stay in {c.name}; export them first if you
                  need them
                </li>
                {c.extraRebuild && <li>{c.extraRebuild}</li>}
              </ul>
            </div>
          </div>
        </section>

        <section className="mx-auto max-w-3xl px-5 py-16 text-center">
          <h2 className="font-editorial text-4xl tracking-tight">
            Stuck on a step?
          </h2>
          <p className="mt-4 text-sm leading-relaxed text-muted">
            Ask Xem inside your workspace, or email{" "}
            <a href="mailto:team@xem.email" className="underline">
              team@xem.email
            </a>{" "}
            and we&apos;ll help you move.
          </p>
        </section>
      </main>
      <SiteFooter />
    </>
  );
}
