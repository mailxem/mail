"use client";
import { useEffect, useRef, useState } from "react";
import Image from "next/image";
import {
  ArrowDown,
  ArrowLeft,
  ArrowRight,
  ArrowUpRight,
  Check,
  GitBranch,
  Mail,
  Menu,
  Play,
  Star,
  X,
} from "lucide-react";
import { gsap } from "gsap";
import { ScrollTrigger } from "gsap/ScrollTrigger";
import { productViews, templates, faqs } from "@/lib/content";
import { appLink, githubUrl, selfHostUrl, docsUrl } from "@/lib/site";
import { SiteDialog } from "./site-dialog";
import { HeroDemo } from "./hero-demo";
import { FeatureGrid } from "./feature-grid";
import { AgentStrip } from "./agent-strip";
import { FounderNote } from "./founder-note";
import { McpSection } from "./mcp-section";
import { AnalyticsSection } from "./analytics-section";
import { Testimonials, JournalPreview } from "./home-expansion";
import type { PostMeta } from "@/lib/blog-types";
import { HeroFilm } from "./hero-film";
import { HeroWheel } from "./hero-wheel";
import { PriceSection, biggestSaving } from "./price-section";
import { primaryButton as primary } from "@/lib/brand";
import { SiteFooter } from "./site-chrome";
import { DeliveredCount, LiveNumbers } from "./live-numbers";
import { cn } from "@/lib/utils";
const textLink =
  "group inline-flex items-center gap-2 border-b border-current pb-1 text-sm font-medium transition-opacity hover:opacity-65";
const sectionLabel = "text-[11px] font-medium uppercase tracking-[0.17em]";
const heading =
  "font-editorial text-[clamp(3rem,5.5vw,5rem)] font-normal leading-[1.02] tracking-[-0.045em]";
function Brand({ light = false }: { light?: boolean }) {
  return (
    <a
      href="#"
      aria-label="Xem home"
      className={`flex shrink-0 items-center gap-2.5 ${light ? "text-cream" : "text-ink"}`}
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
    </a>
  );
}
function Arrow() {
  return (
    <ArrowUpRight
      size={17}
      className="transition-transform group-hover:-translate-y-0.5 group-hover:translate-x-0.5 motion-reduce:transform-none"
    />
  );
}
// Teams the founder confirmed use Xem. researchandrise.com has no site yet,
// so it isn't linked.
const customers: [string, string, string?][] = [
  ["Hakopod", "/logos/hakopod.svg", "https://hakopod.com"],
  ["Zunofy", "/logos/zunofy.png", "https://zunofy.com"],
  ["SyneHQ", "/logos/synehq.png", "https://synehq.com"],
  ["Research and Rise", ""],
];
export function Home({ latestPosts }: { latestPosts: PostMeta[] }) {
  const root = useRef<HTMLDivElement>(null);
  const rail = useRef<HTMLDivElement>(null);
  const [menu, setMenu] = useState(false);
  const [tour, setTour] = useState<number | null>(null);
  const [template, setTemplate] = useState<number | null>(null);
  const [faq, setFaq] = useState(0);
  const [railIndex, setRailIndex] = useState(0);
  const [railEnd, setRailEnd] = useState(false);
  useEffect(() => {
    gsap.registerPlugin(ScrollTrigger);
    const mm = gsap.matchMedia();
    mm.add("(prefers-reduced-motion: no-preference)", () => {
      const ctx = gsap.context(() => {
        // The hero has no intro animation: hiding it until scripts run left
        // the first screen blank on slow phones.
        gsap.utils.toArray<HTMLElement>("[data-reveal]").forEach((el) =>
          gsap.from(el, {
            y: 40,
            opacity: 0,
            duration: 0.8,
            ease: "power2.out",
            scrollTrigger: { trigger: el, start: "top 93%", once: true },
            clearProps: "all",
          }),
        );
      }, root);
      return () => ctx.revert();
    });
    return () => mm.revert();
  }, []);
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setMenu(false);
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);
  const scrollTemplates = (direction: number) => {
    const el = rail.current;
    if (!el) return;
    const card = el.querySelector("article");
    el.scrollBy({
      left: direction * ((card?.getBoundingClientRect().width || 320) + 24),
      behavior: window.matchMedia("(prefers-reduced-motion: reduce)").matches
        ? "instant"
        : "smooth",
    });
  };
  return (
    <div ref={root} className="overflow-clip">
      <header className="fixed inset-x-0 top-3 z-40 mx-auto w-[calc(100%-32px)] max-w-[1080px] md:top-5">
        <div className="flex h-[66px] items-center justify-between gap-2 rounded-xl border border-ink/15 bg-cream/95 px-4 shadow-[0_2px_0_#22251f08] backdrop-blur-xl sm:gap-4 md:px-5">
          <Brand />
          <nav
            aria-label="Main navigation"
            className="hidden items-center gap-7 text-sm md:flex"
          >
            <a href="#product" className="transition-colors hover:text-forest">
              Product
            </a>
            <a
              href="#templates"
              className="transition-colors hover:text-forest"
            >
              Templates
            </a>
            <a href="/pricing" className="transition-colors hover:text-iris">
              Pricing
            </a>
            <a href="/compare" className="transition-colors hover:text-iris">
              Compare
            </a>
            <a href="/blog" className="transition-colors hover:text-iris">
              Journal
            </a>
            <a href={docsUrl} className="transition-colors hover:text-iris">
              Docs
            </a>
          </nav>
          <div className="flex items-center gap-1 sm:gap-5">
            <a
              href={appLink("/auth/login")}
              className="hidden text-sm font-medium hover:underline lg:block"
            >
              Log in
            </a>
            <a
              href={appLink("/auth/register")}
              className={cn(
                primary,
                "min-h-10 gap-2 whitespace-nowrap px-3 py-2.5 text-xs sm:px-4 md:text-sm",
              )}
            >
              Start free <Arrow />
            </a>
            <button
              aria-label={menu ? "Close menu" : "Open menu"}
              aria-expanded={menu}
              aria-controls="mobile-menu"
              onClick={() => setMenu(!menu)}
              className="-mr-1 p-2 md:hidden"
            >
              {menu ? <X size={23} /> : <Menu size={23} />}
            </button>
          </div>
        </div>
        {menu && (
          <nav
            id="mobile-menu"
            aria-label="Mobile navigation"
            className="mt-2 rounded-xl border border-ink/15 bg-cream p-3 shadow-xl md:hidden"
          >
            {[
              ["Product", "#product"],
              ["Templates", "#templates"],
              ["Pricing", "/pricing"],
              ["Compare", "/compare"],
              ["Journal", "/blog"],
              ["FAQs", "#questions"],
              ["Docs", docsUrl],
              ["Log in", appLink("/auth/login")],
            ].map(([label, href]) => (
              <a
                key={label}
                href={href}
                onClick={() => setMenu(false)}
                className="flex items-center justify-between rounded-lg px-4 py-3 text-lg hover:bg-lemon"
              >
                {label}
                <ArrowUpRight size={18} />
              </a>
            ))}
          </nav>
        )}
      </header>
      <main id="main">
        <section
          id="hero"
          aria-labelledby="hero-title"
          className="relative isolate pt-[100px] md:pt-[104px]"
        >
          <div className="relative z-10 mx-auto max-w-6xl px-5 text-center">
            <a
              href={githubUrl}
              className={`${sectionLabel} inline-flex items-center gap-2.5 rounded-full border border-forest/20 bg-forest/5 px-4 py-2 transition-colors hover:bg-forest/10`}
            >
              <GitBranch size={14} aria-hidden="true" /> Proudly open source
              <ArrowUpRight size={14} aria-hidden="true" />
            </a>
            <h1
              id="hero-title"
              className="mt-5 font-editorial text-[clamp(2.75rem,7.8vw,6rem)] font-normal leading-[.94] tracking-[-.055em] [container-type:inline-size]"
            >
              All your email marketing,
              <HeroWheel />
            </h1>
            <p className="mx-auto mt-6 max-w-[560px] text-[17px] leading-relaxed tracking-[-.02em] md:text-lg">
              Pay for the emails you send, not the list you keep.{" "}
              <strong className="font-semibold">
                Free for 3,000 emails a month, then from $9.
              </strong>{" "}
              That&apos;s up to{" "}
              <a
                href="#price"
                className="font-medium underline decoration-ink/30 underline-offset-4 hover:decoration-ink"
              >
                {biggestSaving()}% less than Mailchimp, beehiiv or Kit
              </a>{" "}
              for the same newsletter.
            </p>
            <div className="mt-6 flex flex-wrap items-center justify-center gap-x-5 gap-y-2">
              <a href={appLink("/auth/register")} className={primary}>
                Start free <Arrow />
              </a>
              <button
                onClick={() => setTour(0)}
                className="group inline-flex items-center gap-2.5 py-3 text-sm font-medium"
              >
                <span className="flex h-8 w-8 items-center justify-center rounded-full border border-ink/25 transition-colors group-hover:bg-lemon">
                  <Play size={11} fill="currentColor" />
                </span>
                Take a little tour
              </button>
            </div>
            <p className="mt-4 text-xs text-muted">
              No card needed. Prefer your own server?{" "}
              <a
                href={selfHostUrl}
                className="font-medium text-forest underline decoration-forest/30 underline-offset-4 hover:decoration-forest"
              >
                Self-host Xem <span aria-hidden="true">↗</span>
              </a>
            </p>
            <p className="mx-auto mt-6 max-w-xl pb-4 text-sm text-muted md:pb-0">
              Used by teams at{" "}
              {customers.map(([name, logo, href], i) => (
                <span key={name}>
                  {i > 0 && (i === customers.length - 1 ? " and " : ", ")}
                  {href ? (
                    <a
                      href={href}
                      className="font-medium text-ink hover:underline"
                    >
                      {logo && (
                        <Image
                          src={logo}
                          alt=""
                          width={20}
                          height={20}
                          className="mr-1.5 inline-block h-[1.25em] w-[1.25em] -translate-y-px rounded-[5px]"
                        />
                      )}
                      {name}
                    </a>
                  ) : (
                    <span className="font-medium text-ink">{name}</span>
                  )}
                </span>
              ))}
              . Over{" "}
              <span className="font-medium text-ink">
                <DeliveredCount fallback={5000} />
              </span>{" "}
              delivered so far.
            </p>
          </div>
          <div className="mx-auto hidden max-w-[1240px] items-center justify-between px-5 pb-7 text-[10px] uppercase tracking-[.13em] text-muted md:flex md:px-10">
            <a
              href="#product"
              className="flex items-center gap-2 hover:text-ink"
            >
              Keep exploring <ArrowDown size={13} />
            </a>
            <a
              href={githubUrl}
              className="inline-flex items-center gap-2 hover:text-ink"
            >
              <Star size={14} aria-hidden="true" /> Star on GitHub
              <ArrowUpRight size={13} aria-hidden="true" />
            </a>
          </div>
        </section>

        <section
          id="product"
          className="relative scroll-mt-20 rounded-t-[36px] bg-forest px-5 pb-20 pt-12 text-cream md:rounded-t-[64px] md:px-10 md:pb-24 md:pt-10"
        >
          <div className="mx-auto max-w-[1180px]">
            <h2 className="sr-only">The Xem app</h2>
            <HeroDemo />
            <AgentStrip />
          </div>
        </section>

        <Testimonials />
        <LiveNumbers />

        <FeatureGrid />

        <section
          id="templates"
          className="scroll-mt-20 rounded-[36px] bg-lavender py-16 md:rounded-[64px] md:py-24"
        >
          <div className="mx-auto max-w-[1240px] px-5 md:px-10">
            <div
              data-reveal
              className="flex flex-wrap items-end justify-between gap-8"
            >
              <div>
                <p className={sectionLabel}>Templates</p>
                <h2 className={`${heading} mt-5`}>
                  200+ templates.
                  <br />
                  <em>Change every word.</em>
                </h2>
              </div>
              <div className="max-w-[315px]">
                <p className="text-sm leading-relaxed text-ink/75">
                  Newsletters, welcomes, launches and product updates, built
                  from blocks in the visual editor. Nothing is locked to a
                  layout.
                </p>
                <a href={appLink("/templates")} className={`${textLink} mt-5`}>
                  Explore the whole collection <Arrow />
                </a>
              </div>
            </div>
          </div>
          <div
            ref={rail}
            role="region"
            aria-label="Template collection"
            tabIndex={0}
            onScroll={() => {
              const el = rail.current;
              const card = el?.querySelector("article");
              if (el && card) {
                setRailIndex(
                  Math.round(
                    el.scrollLeft / (card.getBoundingClientRect().width + 24),
                  ),
                );
                setRailEnd(
                  el.scrollLeft + el.clientWidth >= el.scrollWidth - 8,
                );
              }
            }}
            className="mt-12 flex snap-x snap-mandatory gap-6 overflow-x-auto overscroll-x-contain px-[max(20px,calc((100vw-1160px)/2))] pb-8 pt-3 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden md:mt-16"
          >
            {templates.map((t, i) => (
              <article
                key={t.key}
                className="w-[280px] shrink-0 snap-center md:w-[330px]"
              >
                <button
                  aria-label={`Preview ${t.name}`}
                  onClick={() => setTemplate(i)}
                  className={`group relative block h-[368px] w-full overflow-hidden rounded-xl p-5 pb-0 transition-transform duration-300 hover:-translate-y-2 motion-reduce:transform-none md:h-[430px] ${t.color}`}
                >
                  <Image
                    src={`/images/templates/${t.key}.webp`}
                    alt={`${t.name} email layout`}
                    width={640}
                    height={820}
                    sizes="330px"
                    className="h-full w-full rounded-t shadow-lg object-cover object-top"
                  />
                  <span className="absolute bottom-5 left-1/2 flex -translate-x-1/2 items-center gap-2 whitespace-nowrap rounded-full bg-cream px-4 py-2.5 text-xs shadow-lg transition-transform group-hover:scale-105">
                    Take a peek <ArrowUpRight size={14} />
                  </span>
                </button>
                <div className="mt-5 flex items-center justify-between gap-3">
                  <div>
                    <h3 className="font-editorial text-2xl leading-none tracking-tight">
                      {t.name}
                    </h3>
                    <p className="mt-2 text-[10px] uppercase tracking-widest text-ink/75">
                      {t.category}
                    </p>
                  </div>
                  <button
                    onClick={() => setTemplate(i)}
                    aria-label={`Details for ${t.name}`}
                    className="rounded-full border border-ink/20 p-2 hover:bg-cream/40"
                  >
                    <ArrowUpRight size={18} />
                  </button>
                </div>
              </article>
            ))}
          </div>
          <div className="mx-auto mt-3 flex max-w-[1240px] items-center justify-between px-5 md:px-10">
            <p className="text-xs text-ink/75">
              Made of editable blocks. Made to be you.
            </p>
            <div className="flex items-center gap-2">
              <button
                onClick={() => scrollTemplates(-1)}
                disabled={railIndex === 0}
                aria-label="Previous templates"
                className="rounded-full border border-ink/30 p-3 transition-colors hover:bg-cream/60 disabled:opacity-25"
              >
                <ArrowLeft size={18} />
              </button>
              <button
                onClick={() => scrollTemplates(1)}
                disabled={railEnd}
                aria-label="Next templates"
                className="rounded-full border border-ink/30 p-3 transition-colors hover:bg-cream/60 disabled:opacity-25"
              >
                <ArrowRight size={18} />
              </button>
            </div>
          </div>
        </section>

        <AnalyticsSection />

        <McpSection />

        <FounderNote />

        <PriceSection />

        <section
          id="questions"
          className="scroll-mt-24 px-5 py-20 md:px-10 md:py-28"
        >
          <div className="mx-auto max-w-[1040px]">
            <div data-reveal className="text-center">
              <p className={sectionLabel}>Before you switch</p>
              <h2 className={`${heading} mt-5`}>
                Questions people <em>ask us.</em>
              </h2>
            </div>
            <div className="mt-12 grid gap-2 rounded-[28px] bg-[#e8eadb] p-3 md:grid-cols-2 md:gap-4 md:p-4">
              <div className="rounded-[19px] bg-forest p-4 text-cream md:p-6">
                <p className="mb-5 px-3 font-editorial text-2xl">
                  Let’s talk about it.
                </p>
                {faqs.map((f, i) => (
                  <div key={f.q}>
                    <button
                      id={`question-${i}`}
                      onClick={() => setFaq(i)}
                      aria-expanded={faq === i}
                      aria-controls={`faq-answer faq-mobile-${i}`}
                      className={`flex w-full items-center justify-between gap-4 rounded-lg px-3 py-3.5 text-left text-[13px] leading-relaxed transition-colors ${faq === i ? "bg-cream/15" : "hover:bg-cream/5"}`}
                    >
                      {f.q}
                      <ArrowUpRight
                        size={15}
                        className={`shrink-0 transition-transform ${faq === i ? "rotate-45" : ""}`}
                      />
                    </button>
                    <div
                      id={`faq-mobile-${i}`}
                      role="region"
                      aria-labelledby={`question-${i}`}
                      hidden={faq !== i}
                      className="px-3 pb-5 pt-2 text-[13px] leading-relaxed text-cream/80 md:hidden"
                    >
                      {f.a}
                    </div>
                  </div>
                ))}
              </div>
              <div
                id="faq-answer"
                role="region"
                aria-labelledby={`question-${faq}`}
                aria-live="polite"
                className="hidden min-h-[260px] flex-col md:flex items-start justify-between p-5 md:p-8"
              >
                <div>
                  <span className="text-[10px] uppercase tracking-widest text-muted">
                    A little clarity
                  </span>
                  <h3 className="mt-7 font-editorial text-3xl leading-tight tracking-tight">
                    {faqs[faq].q}
                  </h3>
                  <p className="mt-5 text-sm leading-[1.85] text-muted">
                    {faqs[faq].a}
                  </p>
                </div>
                <a
                  href={appLink("/auth/register")}
                  className={`${textLink} mt-8`}
                >
                  Make yourself at home <Arrow />
                </a>
              </div>
            </div>
          </div>
        </section>

        <section className="relative isolate overflow-hidden rounded-t-[36px] bg-lemon px-5 pb-20 pt-20 text-center md:rounded-t-[64px] md:pb-24 md:pt-24">
          <div
            aria-hidden="true"
            className="pointer-events-none absolute inset-0 -z-10"
          >
            <svg viewBox="0 0 1440 600" className="h-full w-full opacity-40">
              <path
                d="M-100 200 C300 -200 130 700 490 420 S500 70 900 180 S1250 650 1580 270"
                stroke="#869867"
                strokeWidth="1"
                fill="none"
              />
              <path
                d="M-100 225 C300 -175 130 725 490 445 S500 95 900 205 S1250 675 1580 295"
                stroke="#869867"
                strokeWidth="1"
                fill="none"
              />
            </svg>
          </div>
          <div data-reveal>
            <span className={`${sectionLabel} inline-flex items-center gap-2`}>
              <Mail size={13} /> Free for 3,000 emails a month
            </span>
            <h2 className="mt-7 font-editorial text-[clamp(4rem,8vw,7.5rem)] leading-[.93] tracking-[-.055em]">
              Pay for your sends.
              <br />
              <em>Keep the rest.</em>
            </h2>
            <p className="mx-auto mt-7 max-w-xl text-sm text-ink/75 md:text-base">
              Then from $9 a month. Bring your list from Mailchimp, beehiiv,
              Substack or Kit; unsubscribed contacts stay unsubscribed.
            </p>
            <div className="mt-8 flex flex-wrap items-center justify-center gap-x-6 gap-y-3">
              <a href={appLink("/auth/register")} className={primary}>
                Start free <Arrow />
              </a>
              <a
                href="/compare"
                className="text-sm font-medium underline decoration-ink/30 underline-offset-4 hover:decoration-ink"
              >
                Compare your current bill
              </a>
            </div>
          </div>
        </section>
        <JournalPreview posts={latestPosts} />
        <section
          id="film"
          aria-label="The Xem film"
          className="scroll-mt-24 pb-6 pt-4"
        >
          <HeroFilm suspended={tour !== null || template !== null || menu} />
        </section>
      </main>
      <SiteFooter />

      {tour !== null && (
        <SiteDialog title="A little tour of Xem" onClose={() => setTour(null)}>
          <div className="grid md:grid-cols-[240px_1fr]">
            <div className="border-b border-ink/10 p-5 md:border-b-0 md:border-r">
              <p className={`${sectionLabel} mb-4 text-muted`}>
                One connected workspace
              </p>
              <div className="grid grid-cols-2 gap-2 md:grid-cols-1">
                {productViews.map((p, i) => (
                  <button
                    aria-pressed={tour === i}
                    onClick={() => setTour(i)}
                    key={p.id}
                    className={`flex items-center gap-3 rounded-lg px-3 py-3 text-left text-sm ${tour === i ? "bg-forest text-cream" : "hover:bg-lemon"}`}
                  >
                    <span className="text-[10px] opacity-90">0{i + 1}</span>
                    {p.label}
                  </button>
                ))}
              </div>
              <p className="mt-5 text-[11px] leading-relaxed text-muted">
                Actual product screenshots from a demo workspace. Sample data is
                used throughout.
              </p>
            </div>
            <div className="min-w-0 p-5 md:p-7">
              <h3 className="font-editorial text-4xl leading-tight tracking-tight">
                {productViews[tour].title}
              </h3>
              <p className="mt-3 text-sm leading-relaxed text-muted">
                {productViews[tour].description}
              </p>
              <div className="mt-6 overflow-hidden rounded-lg border border-ink/15 bg-white">
                <Image
                  src={`/images/product/${productViews[tour].image}.webp`}
                  quality={90}
                  alt={productViews[tour].alt}
                  width={1512}
                  height={1100}
                  sizes="850px"
                  className="h-auto w-full"
                />
              </div>
              <ul className="my-6 space-y-2 text-sm">
                {productViews[tour].bullets.map((b) => (
                  <li key={b} className="flex items-center gap-2">
                    <Check size={14} className="text-forest" />
                    {b}
                  </li>
                ))}
              </ul>
              <div className="flex flex-wrap items-center justify-between gap-4">
                <a className={primary} href={appLink(productViews[tour].path)}>
                  Open in Xem <Arrow />
                </a>
                <button
                  onClick={() => setTour((tour + 1) % 4)}
                  className="flex items-center gap-2 text-sm"
                >
                  {tour === 3 ? "Back to design" : "Next chapter"}
                  <ArrowRight size={16} />
                </button>
              </div>
            </div>
          </div>
        </SiteDialog>
      )}
      {template !== null && (
        <SiteDialog
          title={templates[template].name}
          onClose={() => setTemplate(null)}
        >
          <div className="grid md:grid-cols-[1fr_300px]">
            <div
              className={`max-h-[70dvh] overflow-y-auto overscroll-contain p-5 ${templates[template].color}`}
            >
              <Image
                src={`/images/templates/${templates[template].key}-full.webp`}
                alt={`Full ${templates[template].name} email template preview`}
                width={640}
                height={1400}
                sizes="640px"
                className="mx-auto h-auto w-full max-w-[640px] shadow-lg"
              />
            </div>
            <div className="p-6">
              <p className={`${sectionLabel} text-muted`}>
                {templates[template].category}
              </p>
              <h3 className="mt-5 font-editorial text-4xl leading-none tracking-tight">
                Make it
                <br />
                <em>your own.</em>
              </h3>
              <p className="mt-5 text-sm leading-relaxed text-muted">
                {templates[template].description}
              </p>
              <ul className="my-7 space-y-3 text-xs">
                {[
                  "Native, editable design blocks",
                  "Your words, images, and colors",
                  "Reusable across your sends",
                ].map((t) => (
                  <li key={t} className="flex items-center gap-2">
                    <Check size={13} />
                    {t}
                  </li>
                ))}
              </ul>
              <a
                href={appLink(
                  `/templates/new?starter=${templates[template].key}`,
                )}
                className={`${primary} w-full`}
              >
                Use this template <Arrow />
              </a>
              <div className="mt-8 flex items-center justify-between">
                <button
                  aria-label="Previous template preview"
                  onClick={() =>
                    setTemplate(
                      (template + templates.length - 1) % templates.length,
                    )
                  }
                  className="rounded-full border border-ink/20 p-2.5 hover:bg-lemon"
                >
                  <ArrowLeft size={16} />
                </button>
                <span className="text-xs text-muted">
                  {template + 1} of {templates.length}
                </span>
                <button
                  aria-label="Next template preview"
                  onClick={() => setTemplate((template + 1) % templates.length)}
                  className="rounded-full border border-ink/20 p-2.5 hover:bg-lemon"
                >
                  <ArrowRight size={16} />
                </button>
              </div>
            </div>
          </div>
        </SiteDialog>
      )}
    </div>
  );
}
