import Image from "next/image";
import Link from "next/link";
import { ArrowUpRight } from "lucide-react";
import {
  appLink,
  githubUrl,
  selfHostUrl,
  contributeUrl,
  docsUrl,
} from "@/lib/site";
import { primaryButton } from "@/lib/brand";

export function Brand({ light = false }: { light?: boolean }) {
  return (
    <Link
      href="/"
      aria-label="Xem home"
      className={`inline-flex shrink-0 items-center gap-2.5 ${light ? "text-cream" : "text-ink"}`}
    >
      <Image
        src="/brand/xem-mark.png"
        width={36}
        height={36}
        alt=""
        className="rounded-md"
      />
      <span className="text-[28px] font-semibold leading-none tracking-[-1.5px]">
        Xem
      </span>
    </Link>
  );
}
/** `active` highlights the current section's link; the journal by default. */
export function SiteHeader({
  active = "journal",
}: {
  active?: "journal" | "pricing" | "mcp" | "templates" | "compare" | "none";
}) {
  const link = (section: string, extra = "") =>
    `${extra} ${active === section ? "font-medium text-iris" : "hover:text-iris"}`.trim();
  return (
    <header className="relative z-30 mx-auto w-[calc(100%-32px)] max-w-[1160px] pt-5">
      <div className="flex min-h-[66px] flex-wrap items-center justify-between gap-4 rounded-xl border border-ink/15 bg-cream/95 px-4 py-3 md:px-5">
        <Brand />
        <nav
          aria-label="Main navigation"
          className="flex items-center gap-4 text-xs sm:gap-7 sm:text-sm"
        >
          <Link href="/#product" className="hover:text-iris">
            Product
          </Link>
          <Link
            href="/templates"
            className={link("templates", "hidden sm:block")}
          >
            Templates
          </Link>
          <Link href="/pricing" className={link("pricing")}>
            Pricing
          </Link>
          <Link href="/compare" className={link("compare", "hidden sm:block")}>
            Compare
          </Link>
          <Link href="/blog" className={link("journal")}>
            Journal
          </Link>
          <a href={docsUrl} className="hidden hover:text-iris sm:block">
            Docs
          </a>
        </nav>
        <div className="flex items-center gap-4">
          <a
            href={appLink("/auth/login")}
            className="hidden text-sm font-medium hover:underline lg:block"
          >
            Log in
          </a>
          <a
            href={appLink("/auth/register")}
            className={`${primaryButton} min-h-10 px-3 py-2 text-xs sm:px-4`}
          >
            Start free <ArrowUpRight size={15} />
          </a>
        </div>
      </div>
    </header>
  );
}
const footerColumns: [string, [string, string][]][] = [
  [
    "Product",
    [
      ["Explore Xem", "/#product"],
      ["Templates", "/templates"],
      ["Pricing", "/pricing"],
      ["MCP server", "/mcp"],
      ["Documentation", docsUrl],
      ["Email deliverability checker", "/tools/email-checker"],
    ],
  ],
  [
    "Switching",
    [
      ["Compare Xem", "/compare"],
      ["Xem vs Mailchimp", "/compare/xem-vs-mailchimp"],
      ["Xem vs beehiiv", "/compare/xem-vs-beehiiv"],
      ["Email authentication report", "/research/email-authentication"],
      ["Journal", "/blog"],
      ["RSS feed", "/feed.xml"],
    ],
  ],
  [
    "Xem",
    [
      ["Open source on GitHub ↗", githubUrl],
      ["Self-host Xem ↗", selfHostUrl],
      ["Contribute to Xem ↗", contributeUrl],
      ["Affiliates · earn 30%", "/affiliates"],
      ["Editorial approach", "/blog/editorial"],
      ["Contact the Xem team", "mailto:team@xem.email"],
    ],
  ],
];

export function SiteFooter() {
  return (
    <footer className="bg-[#222d25] px-5 pb-7 pt-14 text-cream md:px-10 md:pt-16">
      <div className="mx-auto max-w-[1160px]">
        <div className="flex flex-wrap justify-between gap-12 border-b border-cream/20 pb-12">
          <div>
            <Brand light />
            <p className="mt-5 max-w-[260px] text-sm leading-relaxed text-cream/70">
              An open-source home for your email marketing. A little more human.
            </p>
            <a
              href={appLink("/auth/register")}
              className="mt-6 inline-flex items-center gap-2 rounded-full bg-lemon px-5 py-2.5 text-sm font-medium text-ink"
            >
              Start free <ArrowUpRight size={15} />
            </a>
          </div>
          <nav
            aria-label="Footer navigation"
            className="grid grid-cols-2 gap-x-12 gap-y-10 text-sm sm:grid-cols-3 md:gap-x-16"
          >
            {footerColumns.map(([title, links]) => (
              <div key={title}>
                <p className="mb-5 text-[11px] font-medium uppercase tracking-[0.17em] text-cream/70">
                  {title}
                </p>
                <ul className="space-y-3">
                  {links.map(([label, href]) => (
                    <li key={href}>
                      <a href={href} className="hover:underline">
                        {label}
                      </a>
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </nav>
        </div>
        <div className="flex flex-wrap items-center justify-between gap-x-6 gap-y-3 pt-6 text-xs text-cream/70">
          <p>
            © {new Date().getFullYear()} Xem. Made for the people on the other
            end.
          </p>
          <p className="flex flex-wrap gap-x-5 gap-y-2">
            <a href="/privacy">Privacy policy</a>
            <a href="/data-deletion">Delete your data</a>
            <a href="/google-mail">Gmail connection</a>
            <a href={appLink("/auth/login")}>Log in ↗</a>
          </p>
        </div>
      </div>
    </footer>
  );
}
