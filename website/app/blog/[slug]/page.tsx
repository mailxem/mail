import type { Metadata } from "next";
import Image from "next/image";
import Link from "next/link";
import { notFound } from "next/navigation";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { ArrowLeft, ArrowUpRight } from "lucide-react";
import { getPosts, getPost, relatedPosts } from "@/lib/blog";
import { remarkHeadingIds } from "@/lib/markdown";
import { siteUrl, appLink } from "@/lib/site";
import { primaryButton } from "@/lib/brand";
import { SiteHeader, SiteFooter } from "@/components/site-chrome";
import { PostCard } from "@/components/blog/post-card";
import { ArticleActions } from "@/components/blog/article-actions";
export const dynamicParams = false;
export function generateStaticParams() {
  return getPosts().map((p) => ({ slug: p.slug }));
}
export async function generateMetadata({
  params,
}: {
  params: Promise<{ slug: string }>;
}): Promise<Metadata> {
  const post = getPost((await params).slug);
  if (!post) notFound();
  const p = post.meta;
  return {
    title: `${p.title} | Xem`,
    description: p.description,
    alternates: { canonical: `/blog/${p.slug}` },
    openGraph: {
      type: "article",
      title: p.title,
      description: p.description,
      url: `/blog/${p.slug}`,
      publishedTime: `${p.date}T00:00:00Z`,
      modifiedTime: `${p.updated}T00:00:00Z`,
      authors: [`${siteUrl}/blog/editorial`],
      section: p.category,
      tags: p.tags,
      images: [{ url: p.cover, width: 1600, height: 1067, alt: p.coverAlt }],
    },
    twitter: {
      card: "summary_large_image",
      title: p.title,
      description: p.description,
      images: [p.cover],
    },
  };
}
export default async function ArticlePage({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const post = getPost((await params).slug);
  if (!post) notFound();
  const { meta: p, content, toc } = post;
  const dateLabel = new Date(`${p.date}T12:00:00Z`).toLocaleDateString(
    "en-US",
    { month: "long", day: "numeric", year: "numeric", timeZone: "UTC" },
  );
  const jsonLd = {
    "@context": "https://schema.org",
    "@graph": [
      {
        "@type": "BlogPosting",
        "@id": `${siteUrl}/blog/${p.slug}#article`,
        headline: p.title,
        description: p.description,
        image: [`${siteUrl}${p.cover}`],
        datePublished: `${p.date}T00:00:00Z`,
        dateModified: `${p.updated}T00:00:00Z`,
        inLanguage: "en",
        author: {
          "@type": "Organization",
          name: "Xem editorial",
          url: `${siteUrl}/blog/editorial`,
        },
        publisher: {
          "@type": "Organization",
          name: "Xem",
          url: siteUrl,
          logo: {
            "@type": "ImageObject",
            url: `${siteUrl}/brand/xem-mark.png`,
          },
        },
        mainEntityOfPage: {
          "@type": "WebPage",
          "@id": `${siteUrl}/blog/${p.slug}`,
        },
        articleSection: p.category,
        keywords: p.tags.join(", "),
      },
      {
        "@type": "BreadcrumbList",
        itemListElement: [
          { "@type": "ListItem", position: 1, name: "Home", item: siteUrl },
          {
            "@type": "ListItem",
            position: 2,
            name: "Journal",
            item: `${siteUrl}/blog`,
          },
          {
            "@type": "ListItem",
            position: 3,
            name: p.title,
            item: `${siteUrl}/blog/${p.slug}`,
          },
        ],
      },
    ],
  };
  return (
    <>
      <SiteHeader />
      <main id="main">
        <article>
          <header className="mx-auto max-w-[980px] px-5 pb-12 pt-14 md:pt-20">
            <Link
              href="/blog"
              className="inline-flex items-center gap-2 text-xs text-muted hover:text-iris"
            >
              <ArrowLeft size={14} />
              Back to the journal
            </Link>
            <p className="mt-10 text-[11px] font-medium uppercase tracking-[.15em] text-iris">
              {p.category} · {p.readingMinutes} minute read
            </p>
            <h1 className="mt-5 max-w-[920px] font-editorial text-[clamp(2.8rem,5.5vw,5.4rem)] leading-[1.04] tracking-[-.04em]">
              {p.title}
            </h1>
            <p className="mt-6 max-w-[720px] text-lg leading-relaxed text-muted">
              {p.description}
            </p>
            <div className="mt-8 flex flex-wrap items-center justify-between gap-4 border-t border-ink/15 pt-6">
              <div className="flex items-center gap-3">
                <Image
                  src="/brand/xem-mark.png"
                  alt=""
                  width={32}
                  height={32}
                />
                <div className="text-xs">
                  <Link
                    className="font-medium hover:text-iris"
                    href="/blog/editorial"
                  >
                    Xem editorial
                  </Link>
                  <p className="mt-1 text-muted">
                    <time dateTime={p.date}>{dateLabel}</time>
                    {p.updated !== p.date && ` · Updated ${p.updated}`}
                  </p>
                </div>
              </div>
              <ArticleActions />
            </div>
          </header>
          <figure className="mx-auto max-w-[1160px] px-5 md:px-0">
            <div className="relative aspect-[3/2] overflow-hidden rounded-[22px]">
              <Image
                src={p.cover}
                alt={p.coverAlt}
                fill
                priority
                sizes="(max-width:1200px) 95vw, 1160px"
                quality={90}
                className="object-cover"
              />
            </div>
            <figcaption className="mt-3 text-right text-[10px] text-muted">
              Original artwork by{" "}
              <a
                href={p.coverSource}
                target="_blank"
                rel="noopener noreferrer"
                className="underline"
              >
                Xem editorial
              </a>
            </figcaption>
          </figure>
          <div className="mx-auto grid max-w-[1100px] gap-10 px-5 py-14 lg:grid-cols-[220px_minmax(0,720px)] lg:gap-16 lg:py-20">
            <aside className="self-start lg:sticky lg:top-8">
              <details
                open
                className="rounded-xl border border-ink/15 p-5 lg:border-0 lg:p-0"
              >
                <summary className="cursor-pointer text-[10px] font-medium uppercase tracking-[.15em]">
                  In this story
                </summary>
                <nav aria-label="Table of contents" className="mt-5 space-y-3">
                  {toc
                    .filter((t) => t.depth === 2)
                    .map((t) => (
                      <a
                        key={t.id}
                        href={`#${t.id}`}
                        className="block text-xs leading-relaxed text-muted transition-colors hover:text-iris"
                      >
                        {t.text}
                      </a>
                    ))}
                </nav>
              </details>
              <div className="mt-8 hidden rounded-xl bg-[#ece7f7] p-5 lg:block">
                <p className="font-editorial text-2xl leading-tight">
                  Put a good idea
                  <br />
                  into practice.
                </p>
                <p className="mt-3 text-xs leading-relaxed text-muted">
                  Your templates, audience, and next email. Together in Xem.
                </p>
                <a
                  href={appLink("/templates")}
                  className="mt-5 inline-flex items-center gap-2 text-xs font-medium text-iris"
                >
                  Explore Xem <ArrowUpRight size={14} />
                </a>
              </div>
            </aside>
            <div className="min-w-0">
              <div className="prose prose-base max-w-none prose-headings:scroll-mt-8 prose-headings:font-editorial prose-headings:font-normal prose-headings:tracking-tight prose-h2:mb-5 prose-h2:mt-12 prose-h2:text-4xl prose-h3:text-3xl prose-p:leading-[1.85] prose-a:font-medium prose-a:text-iris prose-a:decoration-iris/35 prose-a:underline-offset-4 prose-blockquote:border-iris prose-blockquote:bg-[#eeeaf7] prose-blockquote:px-6 prose-blockquote:py-1 prose-blockquote:font-normal prose-blockquote:not-italic prose-strong:font-semibold prose-code:break-words prose-code:text-iris prose-pre:bg-[#292432] prose-pre:text-cream prose-th:text-left prose-td:align-top">
                <ReactMarkdown
                  skipHtml
                  remarkPlugins={[remarkGfm, remarkHeadingIds]}
                  components={{
                    table: ({ children }) => (
                      <div className="overflow-x-auto rounded-lg border border-ink/15 px-4">
                        <table>{children}</table>
                      </div>
                    ),
                    a: ({ href, children, ...props }) => (
                      <a
                        href={href}
                        {...props}
                        {...(href?.startsWith("https://")
                          ? { target: "_blank", rel: "noopener noreferrer" }
                          : {})}
                      >
                        {children}
                      </a>
                    ),
                  }}
                >
                  {content}
                </ReactMarkdown>
              </div>
              <div className="mt-12 border-t border-ink/15 pt-6">
                <p className="text-xs leading-relaxed text-muted">
                  Written with AI assistance and checked against the linked
                  sources. Examples are illustrative.{" "}
                  <Link className="underline" href="/blog/editorial">
                    Our editorial approach
                  </Link>
                </p>
                <div className="mt-5 flex flex-wrap gap-2">
                  {p.tags.map((t) => (
                    <span
                      key={t}
                      className="rounded-full border border-ink/15 px-3 py-1.5 text-[11px] text-muted"
                    >
                      {t}
                    </span>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </article>
        <section className="bg-[#f0eedf] px-5 py-16 md:px-10">
          <div className="mx-auto max-w-[1160px]">
            <p className="text-[11px] font-medium uppercase tracking-[.15em] text-muted">
              A little more reading
            </p>
            <h2 className="mb-9 mt-4 font-editorial text-4xl tracking-tight">
              Keep the ideas coming.
            </h2>
            <div className="grid gap-10 sm:grid-cols-2 lg:grid-cols-3">
              {relatedPosts(p).map((post) => (
                <PostCard key={post.slug} post={post} />
              ))}
            </div>
          </div>
        </section>
        <section className="px-5 py-16 text-center">
          <h2 className="font-editorial text-4xl tracking-tight">
            Make your next email a good one.
          </h2>
          <a
            href={appLink("/auth/register")}
            className={`${primaryButton} mt-6`}
          >
            Start with Xem <ArrowUpRight size={17} />
          </a>
        </section>
      </main>
      <SiteFooter />
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{
          __html: JSON.stringify(jsonLd).replace(/</g, "\\u003c"),
        }}
      />
    </>
  );
}
