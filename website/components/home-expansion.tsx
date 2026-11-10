"use client";
import { useState } from "react";
import Link from "next/link";
import { ArrowLeft, ArrowRight, ArrowUpRight } from "lucide-react";
import type { PostMeta } from "@/lib/blog-types";
import { PostCard } from "./blog/post-card";

const testimonials = [
  {
    name: "Sidharth Rathi",
    role: "CTO @thirdbase",
    initials: "SR",
    quote:
      "Xem's multi-tenant support is perfect for our agency. We manage 50+ clients' email in one dashboard now. No separate vendor accounts per client",
  },
  {
    name: "Girish Raju",
    role: "CTO @Zunofy",
    initials: "GR",
    quote:
      "We were paying three different vendors for email. Xem consolidated everything into one API. We cut our email costs by 60% and actually have better control now.",
  },
  {
    name: "Vivek Kaushik",
    role: "Head of Tech @Farmako.in",
    initials: "VK",
    quote:
      "The API is clean and intuitive. We integrated Xem into our SaaS product in less than a day. Our customers love having provider flexibility built-in",
  },
];
export function Testimonials() {
  const [selected, setSelected] = useState(0);
  const t = testimonials[selected];
  return (
    <section
      id="voices"
      aria-labelledby="voices-heading"
      className="bg-[#292336] px-5 py-20 text-cream md:px-10 md:py-24"
    >
      <div className="mx-auto grid max-w-[1160px] gap-12 lg:grid-cols-[.65fr_1.4fr]">
        <div data-reveal>
          <p className="text-[11px] font-medium uppercase tracking-[.17em] text-[#cbb9ef]">
            Customers
          </p>
          <h2
            id="voices-heading"
            className="mt-5 font-editorial text-5xl leading-[1.04] tracking-tight"
          >
            What teams on Xem
            <br />
            <em>tell us.</em>
          </h2>
          <p className="mt-6 max-w-xs text-sm leading-relaxed text-cream/65">
            In their own words.
          </p>
          <div className="mt-8 flex gap-2">
            <button
              aria-label="Previous testimonial"
              onClick={() => setSelected((selected + 2) % 3)}
              className="rounded-full border border-cream/30 p-3 hover:bg-cream/10"
            >
              <ArrowLeft size={18} />
            </button>
            <button
              aria-label="Next testimonial"
              onClick={() => setSelected((selected + 1) % 3)}
              className="rounded-full border border-cream/30 p-3 hover:bg-cream/10"
            >
              <ArrowRight size={18} />
            </button>
          </div>
        </div>
        <div data-reveal>
          <span
            aria-hidden="true"
            className="font-editorial text-7xl leading-none text-[#b89dea]"
          >
            “
          </span>
          <div aria-live="polite" aria-atomic="true">
            <blockquote className="min-h-[250px] font-editorial text-[clamp(1.9rem,3.1vw,2.8rem)] leading-[1.18] tracking-tight sm:min-h-[215px]">
              {t.quote}
            </blockquote>
            <p className="mt-6 text-sm font-medium">
              {t.name}
              <span className="ml-3 text-xs font-normal text-cream/60">
                {t.role}
              </span>
            </p>
          </div>
          <div
            aria-label="Choose a customer story"
            className="mt-8 flex flex-wrap gap-2 border-t border-cream/20 pt-6"
          >
            {testimonials.map((person, i) => (
              <button
                key={person.name}
                onClick={() => setSelected(i)}
                aria-pressed={selected === i}
                className={`flex items-center gap-2.5 rounded-full border px-3 py-2.5 text-xs transition-colors ${selected === i ? "border-[#cbb9ef] bg-[#cbb9ef] text-[#292336]" : "border-cream/20 text-cream/80 hover:bg-cream/10"}`}
              >
                <span className="flex h-6 w-6 items-center justify-center rounded-full border border-current/25 text-[9px]">
                  {person.initials}
                </span>
                {person.name}
              </button>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}
export function JournalPreview({ posts }: { posts: PostMeta[] }) {
  return (
    <section id="journal" className="px-5 py-20 md:px-10 md:py-24">
      <div className="mx-auto max-w-[1160px]">
        <div
          data-reveal
          className="mb-10 flex flex-wrap items-end justify-between gap-7"
        >
          <div>
            <p className="text-[11px] font-medium uppercase tracking-[.17em] text-iris">
              Journal
            </p>
            <h2 className="mt-5 font-editorial text-[clamp(3rem,5vw,4.8rem)] leading-[1.02] tracking-[-.04em]">
              Guides and
              <br />
              <em>research.</em>
            </h2>
          </div>
          <Link
            href="/blog"
            className="inline-flex items-center gap-3 border-b border-ink pb-2 text-sm font-medium"
          >
            Explore the journal <ArrowUpRight size={17} />
          </Link>
        </div>
        <div data-reveal className="grid gap-10 sm:grid-cols-2 lg:grid-cols-3">
          {posts.map((post) => (
            <PostCard key={post.slug} post={post} />
          ))}
        </div>
      </div>
    </section>
  );
}
