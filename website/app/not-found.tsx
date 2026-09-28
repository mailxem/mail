import Link from "next/link";
import { SiteHeader, SiteFooter } from "@/components/site-chrome";
import { primaryButton } from "@/lib/brand";
export default function NotFound() {
  return (
    <>
      <SiteHeader />
      <main id="main" className="mx-auto max-w-3xl px-5 py-28 text-center">
        <p className="text-xs uppercase tracking-widest text-iris">
          404 / A small detour
        </p>
        <h1 className="mt-6 font-editorial text-6xl tracking-tight">
          This page slipped
          <br />
          out of the inbox.
        </h1>
        <p className="mx-auto mt-6 max-w-md text-sm leading-relaxed text-muted">
          The link may have changed, or the story may not exist. There's plenty
          more to explore in the journal.
        </p>
        <Link href="/blog" className={`${primaryButton} mt-8`}>
          Back to the journal ↗
        </Link>
      </main>
      <SiteFooter />
    </>
  );
}
