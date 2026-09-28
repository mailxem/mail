import type { Metadata } from "next";
import Image from "next/image";
import Link from "next/link";
import { ArrowUpRight, Rss } from "lucide-react";
import { getPosts } from "@/lib/blog";
import { siteUrl } from "@/lib/site";
import { SiteHeader, SiteFooter } from "@/components/site-chrome";
import { BlogExplorer } from "@/components/blog/blog-explorer";
export const metadata: Metadata = {
  title: "The Xem Journal — Better email, thoughtfully made",
  description:
    "Practical guides to AI email writing, newsletter growth, deliverability, and meaningful analytics. Ideas worth opening, from the Xem journal.",
  alternates: {
    canonical: "/blog",
    types: { "application/rss+xml": "/feed.xml" },
  },
  openGraph: {
    type: "website",
    url: "/blog",
    title: "The Xem Journal",
    description: "Better email, thoughtfully made.",
    images: [
      {
        url: "/images/blog/ai-email-marketing.webp",
        width: 1600,
        height: 1067,
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: "The Xem Journal",
    description: "Better email, thoughtfully made.",
    images: ["/images/blog/ai-email-marketing.webp"],
  },
};
export default function BlogPage() {
  const posts = getPosts();
  const featured = posts.find((p) => p.featured) ?? posts[0];
  return (
    <>
      <SiteHeader />
      <main id="main">
        <section className="mx-auto max-w-[1160px] px-5 pb-16 pt-16 md:px-0 md:pt-24">
          <div className="flex items-center justify-between">
            <p className="text-[11px] font-medium uppercase tracking-[0.17em] text-iris">
              The Xem journal
            </p>
            <a
              href="/feed.xml"
              className="inline-flex items-center gap-2 text-xs text-muted"
            >
              <Rss size={14} />
              Follow along
            </a>
          </div>
          <h1 className="mt-7 max-w-3xl font-editorial text-[clamp(3.5rem,7vw,6.8rem)] leading-[.98] tracking-[-.045em]">
            Good email starts
            <br />
            with a <em>good idea.</em>
          </h1>
          <p className="mt-7 max-w-xl text-base leading-relaxed text-muted">
            A field guide for people who care about what they send. Thoughtful
            writing, healthier lists, and numbers that mean something.
          </p>
          <Link
            href={`/blog/${featured.slug}`}
            className="group mt-12 grid overflow-hidden rounded-[24px] bg-[#e8e5f4] md:grid-cols-[1.1fr_1fr]"
          >
            <div className="relative aspect-[3/2] self-center">
              <Image
                src={featured.cover}
                alt={featured.coverAlt}
                fill
                priority
                sizes="(max-width:768px) 95vw, 620px"
                className="object-cover transition-transform duration-700 group-hover:scale-[1.02] motion-reduce:transform-none"
              />
            </div>
            <div className="flex flex-col items-start justify-center p-7 md:p-12">
              <p className="text-[10px] font-medium uppercase tracking-[.15em] text-iris">
                Start here · {featured.category}
              </p>
              <h2 className="mt-5 font-editorial text-[clamp(2.4rem,3.5vw,3.6rem)] leading-[1.04] tracking-tight">
                {featured.title}
              </h2>
              <p className="mt-5 text-sm leading-relaxed text-muted">
                {featured.description}
              </p>
              <span className="mt-8 inline-flex items-center gap-3 border-b border-ink pb-2 text-sm font-medium">
                Read the story <ArrowUpRight size={17} />
              </span>
              <p className="mt-6 text-[11px] text-muted">
                {featured.readingMinutes} minute read
              </p>
            </div>
          </Link>
        </section>
        <BlogExplorer posts={posts} />
      </main>
      <SiteFooter />
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{
          __html: JSON.stringify({
            "@context": "https://schema.org",
            "@type": "Blog",
            name: "The Xem Journal",
            url: `${siteUrl}/blog`,
            description: metadata.description,
            blogPost: posts.map((p) => ({
              "@type": "BlogPosting",
              headline: p.title,
              url: `${siteUrl}/blog/${p.slug}`,
              datePublished: `${p.date}T00:00:00Z`,
            })),
          }).replace(/</g, "\\u003c"),
        }}
      />
    </>
  );
}
