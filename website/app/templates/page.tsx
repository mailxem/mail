import type { Metadata } from "next";
import { ArrowUpRight } from "lucide-react";
import { SiteHeader, SiteFooter } from "@/components/site-chrome";
import { TemplateGallery } from "@/components/template-gallery";
import { primaryButton } from "@/lib/brand";
import { siteUrl } from "@/lib/site";
import { galleryTemplates, registerHref } from "@/lib/template-gallery";

const title = `${galleryTemplates.length} free email templates: newsletters, welcome, product and more | Xem`;
const description = `Browse ${galleryTemplates.length} editable email templates for newsletters, welcome emails, product updates, shops, events and receipts. Open any of them in Xem and change every block.`;
export const metadata: Metadata = {
  title,
  description,
  alternates: { canonical: "/templates" },
  openGraph: { type: "website", url: "/templates", title, description },
  twitter: { card: "summary", title, description },
};

export default function TemplatesIndex() {
  const structuredData = {
    "@context": "https://schema.org",
    "@graph": [
      {
        "@type": "CollectionPage",
        "@id": `${siteUrl}/templates`,
        name: "Xem email templates",
        description,
        hasPart: galleryTemplates.map((t) => ({
          "@type": "CreativeWork",
          name: t.name,
          url: `${siteUrl}/templates/${t.key}`,
        })),
      },
      {
        "@type": "BreadcrumbList",
        itemListElement: [
          { "@type": "ListItem", position: 1, name: "Home", item: siteUrl },
          {
            "@type": "ListItem",
            position: 2,
            name: "Templates",
            item: `${siteUrl}/templates`,
          },
        ],
      },
    ],
  };
  return (
    <>
      <SiteHeader active="templates" />
      <main id="main">
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{
            __html: JSON.stringify(structuredData).replace(/</g, "\\u003c"),
          }}
        />
        <section className="mx-auto max-w-[1160px] px-5 pb-10 pt-16 md:pt-24">
          <p className="text-xs font-medium uppercase tracking-[0.17em] text-iris">
            Email templates
          </p>
          <h1 className="mt-5 max-w-4xl font-editorial text-[clamp(2.75rem,6vw,5.5rem)] leading-[1.02] tracking-[-0.045em]">
            {galleryTemplates.length} email templates.
            <br />
            <em>Every block editable.</em>
          </h1>
          <p className="mt-7 max-w-2xl text-lg leading-relaxed text-muted">
            Layouts for newsletters, welcome series, product updates, shops,
            events, account emails and win-back campaigns, built from the same
            blocks you edit in Xem. Pick one, change the words, photos and
            colours, and send it from your own domain.
          </p>
          <a href={registerHref} className={`${primaryButton} mt-8`}>
            Start free and pick a template <ArrowUpRight size={17} />
          </a>
        </section>
        <section
          aria-label="Templates"
          className="mx-auto max-w-[1160px] px-5 pb-24"
        >
          <TemplateGallery />
        </section>
      </main>
      <SiteFooter />
    </>
  );
}
