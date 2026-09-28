import Image from "next/image";
import Link from "next/link";
import { ArrowUpRight } from "lucide-react";
import { appLink, githubUrl, selfHostUrl, contributeUrl } from "@/lib/site";
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
export function SiteHeader() {
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
          <Link href="/#templates" className="hidden hover:text-iris sm:block">
            Templates
          </Link>
          <Link href="/mcp" className="hover:text-iris">
            MCP
          </Link>
          <Link href="/blog" className="font-medium text-iris">
            Journal
          </Link>
        </nav>
        <a
          href={appLink("/auth/register")}
          className={`${primaryButton} min-h-10 px-3 py-2 text-xs sm:px-4`}
        >
          Get started <ArrowUpRight size={15} />
        </a>
      </div>
    </header>
  );
}
export function SiteFooter() {
  return (
    <footer className="bg-[#222d25] px-5 py-12 text-cream md:px-10">
      <div className="mx-auto max-w-[1160px]">
        <div className="flex flex-wrap justify-between gap-10">
          <div>
            <Brand light />
            <p className="mt-4 max-w-xs text-sm leading-relaxed text-cream/70">
              An open-source home for your email marketing.
              <br />A little more human.
            </p>
          </div>
          <nav
            aria-label="Footer navigation"
            className="grid grid-cols-2 gap-x-10 gap-y-3 text-sm"
          >
            <Link href="/#product">Explore Xem</Link>
            <Link href="/#assistant">Ask Xem</Link>
            <Link href="/#sending">Sending & setup</Link>
            <Link href="/blog">Journal</Link>
            <Link href="/#templates">Email templates</Link>
            <Link href="/mcp">MCP server</Link>
            <Link href="/blog/editorial">Editorial approach</Link>
            <a href="https://docs.xem.email">Documentation</a>
            <a href="/feed.xml">RSS feed</a>
            <a href={githubUrl}>Open source on GitHub ↗</a>
            <a href={selfHostUrl}>Self-host Xem ↗</a>
            <a href={contributeUrl}>Contribute to Xem ↗</a>
          </nav>
        </div>
        <div className="mt-10 flex flex-wrap justify-between gap-4 border-t border-cream/20 pt-6 text-xs text-cream/70">
          <p>
            © {new Date().getFullYear()} Xem. Made for the people on the other
            end.
          </p>
          <a href={appLink("/auth/login")}>Log in to your workspace ↗</a>
        </div>
      </div>
    </footer>
  );
}
