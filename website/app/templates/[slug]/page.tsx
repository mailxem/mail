import type { Metadata } from "next";
import Image from "next/image";
import Link from "next/link";
import { notFound } from "next/navigation";
import { ArrowUpRight } from "lucide-react";
import { SiteHeader, SiteFooter } from "@/components/site-chrome";
import { primaryButton } from "@/lib/brand";
import { siteUrl } from "@/lib/site";
import {
  galleryTemplates,
  getGalleryTemplate,
  usage,
  registerHref,
} from "@/lib/template-gallery";

export const dynamicParams = false;
export function generateStaticParams() {
  return galleryTemplates.map((t) => ({ slug: t.key }));
}

const kind = (category: string) =>
  category === "Newsletters" ? "newsletter" : category.toLowerCase();

export async function generateMetadata({
  params,
}: {
  params: Promise<{ slug: string }>;
}): Promise<Metadata> {
  const t = getGalleryTemplate((await params).slug);
  if (!t) return {};
  const title = `${t.name}: free ${kind(t.category)} email template | Xem`;
  const description = `${t.description} An editable ${kind(t.category)} email template you can open and change in Xem.`;
  const path = `/templates/${t.key}`;
  const image = `/images/gallery/${t.key}.webp`;
  return {
    title,
    description,
    alternates: { canonical: path },
    openGraph: {
      type: "website",
      url: path,
      title,
      description,
      images: [{ url: image, width: 400, height: 500, alt: t.name }],
    },
    twitter: { card: "summary", title, description, images: [image] },
  };
}

export default async function TemplatePage({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const t = getGalleryTemplate((await params).slug);
  if (!t) notFound();
  const related = galleryTemplates
    .filter((r) => r.category === t.category && r.key !== t.key)
    .slice(0, 4);
  const url = `${siteUrl}/templates/${t.key}`;
  const structuredData = {
    "@context": "https://schema.org",
    "@graph": [
      {
        "@type": "CreativeWork",
        "@id": url,
        name: t.name,
        description: t.description,
        genre: `${t.category} email template`,
        image: `${siteUrl}/images/gallery/${t.key}-full.webp`,
        isAccessibleForFree: true,
        creator: { "@type": "Organization", name: "Xem", url: siteUrl },
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
          { "@type": "ListItem", position: 3, name: t.name, item: url },
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
        <div className="mx-auto grid max-w-[1160px] gap-12 px-5 pb-20 pt-16 md:grid-cols-[minmax(0,1fr)_minmax(0,1.1fr)] md:pt-24">
          <section className="md:sticky md:top-8 md:self-start">
            <p className="text-xs font-medium uppercase tracking-[0.17em] text-iris">
              <Link href="/templates">Templates</Link> · {t.category}
            </p>
            <h1 className="mt-5 font-editorial text-[clamp(2.5rem,5vw,4.5rem)] leading-[1.02] tracking-[-0.045em]">
              {t.name}
            </h1>
            <p className="mt-6 text-lg leading-relaxed">{t.description}</p>
            <p className="mt-4 leading-relaxed text-muted">{usage(t)}</p>
            <dl className="mt-8 grid gap-4 rounded-xl border border-ink/10 bg-cream p-5 text-sm">
              <div>
                <dt className="text-[10px] uppercase tracking-widest text-ink/70">
                  Category
                </dt>
                <dd className="mt-1">{t.category}</dd>
              </div>
              <div>
                <dt className="text-[10px] uppercase tracking-widest text-ink/70">
                  Sample subject line
                </dt>
                <dd className="mt-1">{t.subject}</dd>
              </div>
              <div>
                <dt className="text-[10px] uppercase tracking-widest text-ink/70">
                  Sample preview text
                </dt>
                <dd className="mt-1">{t.preheader}</dd>
              </div>
            </dl>
            <a href={registerHref} className={`${primaryButton} mt-8`}>
              Use this template in Xem <ArrowUpRight size={17} />
            </a>
            <p className="mt-3 text-xs text-muted">
              Create a free account, then pick {t.name} from the template
              library. Every block can be edited.
            </p>
          </section>
          <figure className="rounded-2xl bg-lavender p-4 md:p-8">
            <Image
              src={`/images/gallery/${t.key}-full.webp`}
              alt={`Full preview of the ${t.name} email template: ${t.description}`}
              width={640}
              height={t.height}
              priority
              className="mx-auto h-auto w-full max-w-[640px] rounded shadow-lg"
            />
          </figure>
        </div>
        {related.length > 0 && (
          <section
            aria-labelledby="related-title"
            className="mx-auto max-w-[1160px] px-5 pb-24"
          >
            <h2
              id="related-title"
              className="font-editorial text-3xl tracking-tight md:text-4xl"
            >
              More {kind(t.category)} templates
            </h2>
            <ul className="mt-8 grid grid-cols-2 gap-5 md:grid-cols-4">
              {related.map((r) => (
                <li key={r.key}>
                  <Link href={`/templates/${r.key}`} className="group block">
                    <div className="overflow-hidden rounded-xl bg-lavender p-3 pb-0">
                      <Image
                        src={`/images/gallery/${r.key}.webp`}
                        alt={`Preview of the ${r.name} email template`}
                        width={400}
                        height={500}
                        loading="lazy"
                        className="h-auto w-full rounded-t shadow-md"
                      />
                    </div>
                    <h3 className="mt-3 font-editorial text-lg group-hover:text-iris">
                      {r.name}
                    </h3>
                  </Link>
                </li>
              ))}
            </ul>
          </section>
        )}
      </main>
      <SiteFooter />
    </>
  );
}
