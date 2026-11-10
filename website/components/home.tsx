"use client";
import { useEffect, useRef, useState } from "react";
import Image from "next/image";
import {
  ArrowDown,
  ArrowLeft,
  ArrowRight,
  ArrowUpRight,
  ChartColumn,
  Check,
  ClipboardList,
  Code2,
  ExternalLink,
  GitBranch,
  Github,
  Inbox,
  LayoutTemplate,
  Mail,
  Menu,
  Play,
  Plus,
  Star,
  Users,
  Workflow,
  X,
} from "lucide-react";
import { gsap } from "gsap";
import { ScrollTrigger } from "gsap/ScrollTrigger";
import { productViews, templates, faqs } from "@/lib/content";
import {
  appLink,
  githubUrl,
  selfHostUrl,
  contributeUrl,
  docsUrl,
} from "@/lib/site";
import { SiteDialog } from "./site-dialog";
import { McpSection } from "./mcp-section";
import { AnalyticsSection } from "./analytics-section";
import { AssistantShowcase } from "./assistant-showcase";
import { SendingSection } from "./sending-section";
import { MoreFeatures, Testimonials, JournalPreview } from "./home-expansion";
import type { PostMeta } from "@/lib/blog-types";
import { HeroFilm } from "./hero-film";
import { HeroWheel } from "./hero-wheel";
import { PriceSection, biggestSaving } from "./price-section";
import { primaryButton as primary } from "@/lib/brand";
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
export function Home({ latestPosts }: { latestPosts: PostMeta[] }) {
  const root = useRef<HTMLDivElement>(null);
  const rail = useRef<HTMLDivElement>(null);
  const [menu, setMenu] = useState(false);
  const [view, setView] = useState(0);
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
  const current = productViews[view];
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
              Get started <Arrow />
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
          className="relative isolate pt-[116px] md:pt-[120px]"
        >
          <div className="relative z-10 mx-auto max-w-6xl px-5 text-center">
            <a
              href="#open-source"
              className={`${sectionLabel} inline-flex items-center gap-2.5 rounded-full border border-forest/20 bg-forest/5 px-4 py-2 transition-colors hover:bg-forest/10`}
            >
              <GitBranch size={14} aria-hidden="true" /> Proudly open source
              <ArrowUpRight size={14} aria-hidden="true" />
            </a>
            <h1
              id="hero-title"
              className="mt-5 font-editorial text-[clamp(3.25rem,7.8vw,6rem)] font-normal leading-[.94] tracking-[-.055em]"
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
                Create your first email <Arrow />
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
              Prefer your own server?{" "}
              <a
                href={selfHostUrl}
                className="font-medium text-forest underline decoration-forest/30 underline-offset-4 hover:decoration-forest"
              >
                Self-host Xem <span aria-hidden="true">↗</span>
              </a>
            </p>
          </div>
          <HeroFilm suspended={tour !== null || template !== null || menu} />
          <div className="mx-auto flex max-w-[1240px] items-center justify-between px-5 pb-7 text-[10px] uppercase tracking-[.13em] text-muted md:px-10">
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

        <PriceSection />
        <Testimonials />

        <section
          id="product"
          className="relative scroll-mt-20 rounded-t-[36px] bg-forest px-5 pb-20 pt-16 text-cream md:rounded-t-[64px] md:px-10 md:pb-24 md:pt-24"
        >
          <div className="mx-auto max-w-[1180px]">
            <div
              data-reveal
              className="flex flex-wrap items-center justify-center gap-x-8 gap-y-4 border-b border-cream/20 pb-10 text-xs text-cream/80 md:gap-x-12"
            >
              {[
                [Mail, "Campaigns & newsletters"],
                [Workflow, "Automations"],
                [Users, "Audience & CRM"],
                [ClipboardList, "Forms"],
                [Inbox, "Shared inbox"],
                [ChartColumn, "Analytics"],
                [LayoutTemplate, "Templates"],
                [Code2, "API & webhooks"],
              ].map(([Icon, label]) => {
                const I = Icon as typeof Mail;
                return (
                  <span
                    key={label as string}
                    className="flex items-center gap-2"
                  >
                    <I size={16} strokeWidth={1.4} />
                    {label as string}
                  </span>
                );
              })}
            </div>
            <div
              data-reveal
              className="mx-auto mb-10 mt-14 max-w-3xl text-center"
            >
              <p className={sectionLabel}>Less busywork. More possibility.</p>
              <h2 className={`${heading} mt-5`}>
                Your whole email world.
                <br />
                <em>Finally, together.</em>
              </h2>
              <p className="mx-auto mt-6 max-w-[510px] text-base leading-relaxed text-cream/75">
                The canvas, the audience, the perfect moment.
                <br className="hidden md:block" /> Everything you need to turn a
                message into a relationship.
              </p>
            </div>
            <div
              role="tablist"
              aria-label="Explore Xem features"
              className="mx-auto mb-7 flex w-fit max-w-full rounded-full border border-cream/20 p-1"
            >
              {productViews.map((p, i) => (
                <button
                  key={p.id}
                  role="tab"
                  id={`product-tab-${i}`}
                  aria-controls="product-panel"
                  aria-selected={view === i}
                  tabIndex={view === i ? 0 : -1}
                  onKeyDown={(e) => {
                    if (e.key === "ArrowRight" || e.key === "ArrowLeft") {
                      e.preventDefault();
                      const next = (i + (e.key === "ArrowRight" ? 1 : 3)) % 4;
                      setView(next);
                      document.getElementById(`product-tab-${next}`)?.focus();
                    }
                  }}
                  onClick={() => setView(i)}
                  className={`min-h-10 rounded-full px-3.5 text-xs transition-colors md:px-7 md:text-sm ${view === i ? "bg-lemon text-forest" : "text-cream/80 hover:bg-cream/10"}`}
                >
                  <span className="mr-2 hidden text-[10px] opacity-90 sm:inline">
                    0{i + 1}
                  </span>
                  {p.label}
                </button>
              ))}
            </div>
            <div
              id="product-panel"
              role="tabpanel"
              aria-labelledby={`product-tab-${view}`}
              className="overflow-hidden rounded-xl border border-cream/20 bg-[#dce4cd] p-2 pb-0 md:rounded-2xl md:p-3 md:pb-0"
            >
              <div className="flex items-center justify-between px-2 pb-3 pt-1 text-[10px] text-forest/80">
                <span className="flex gap-1.5">
                  <i className="h-2 w-2 rounded-full bg-forest/25" />
                  <i className="h-2 w-2 rounded-full bg-forest/20" />
                  <i className="h-2 w-2 rounded-full bg-forest/15" />
                </span>
                <span>Xem / {current.label}</span>
                <button
                  onClick={() => setTour(view)}
                  className="flex items-center gap-1.5 hover:underline"
                >
                  Explore this view <ExternalLink size={11} />
                </button>
              </div>
              <button
                aria-label={`Open ${current.label.toLowerCase()} product tour`}
                onClick={() => setTour(view)}
                className="group relative block aspect-[1.53] w-full overflow-hidden rounded-t-lg bg-white"
              >
                <Image
                  key={current.image}
                  quality={90}
                  src={`/images/product/${current.image}.webp`}
                  alt={current.alt}
                  width={1512}
                  height={1100}
                  sizes="(max-width: 768px) 95vw, 1160px"
                  className="h-full w-full object-cover object-top transition-transform duration-700 group-hover:scale-[1.015] motion-reduce:transform-none"
                />
                <span className="absolute bottom-5 right-5 flex items-center gap-2 rounded-full bg-ink px-4 py-2.5 text-xs text-cream shadow-md md:bottom-8 md:right-8">
                  Take a closer look <Plus size={14} />
                </span>
              </button>
            </div>
            <div className="mt-7 grid items-start gap-4 md:grid-cols-[1fr_1.1fr]">
              <h3 className="font-editorial text-3xl leading-tight tracking-tight">
                {current.title}
              </h3>
              <div>
                <p className="text-sm leading-relaxed text-cream/80">
                  {current.description}
                </p>
                <p className="mt-3 text-[10px] text-cream/75">
                  Actual Xem interface. Demo workspace with sample data.
                </p>
              </div>
            </div>
          </div>
        </section>

        <AssistantShowcase />
        <SendingSection />

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
                <p className={sectionLabel}>The Xem collection</p>
                <h2 className={`${heading} mt-5`}>
                  A head start.
                  <br />
                  <em>Never a blank canvas.</em>
                </h2>
              </div>
              <div className="max-w-[315px]">
                <p className="text-sm leading-relaxed text-ink/75">
                  200+ editable email templates.
                  <br />
                  Different moods, different stories. All yours to make your
                  own.
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
        <MoreFeatures />
        <JournalPreview posts={latestPosts} />
        <section
          id="open-source"
          aria-labelledby="open-source-title"
          className="scroll-mt-28 border-t border-ink/10 bg-lavender/30 px-5 py-20 md:px-10 md:py-28"
        >
          <div className="mx-auto grid max-w-[1160px] gap-10 md:grid-cols-2 md:items-center md:gap-20">
            <div data-reveal>
              <span
                className={`${sectionLabel} inline-flex items-center gap-2`}
              >
                <GitBranch size={14} aria-hidden="true" /> Proudly open source
              </span>
              <h2 id="open-source-title" className={`${heading} mt-6`}>
                Good things grow
                <br />
                <em>in the open.</em>
              </h2>
            </div>
            <div data-reveal>
              <p className="text-base leading-relaxed text-muted md:text-lg">
                The app, Go backend, SDKs, and self-hosting tools live in one
                repository. Run Xem on your own server, see how it works, and
                help shape what comes next. There’s room for your ideas here.
              </p>
              <ul className="mt-6 space-y-3 text-sm">
                {[
                  "Clone once to explore the whole product",
                  "Make a small fix, improve a guide, or share an idea",
                  "Keep your SMTP provider and your infrastructure",
                ].map((item) => (
                  <li key={item} className="flex items-center gap-2.5">
                    <Check
                      size={16}
                      className="shrink-0 text-forest"
                      aria-hidden="true"
                    />
                    {item}
                  </li>
                ))}
              </ul>
              <div className="mt-8 flex flex-wrap items-center gap-x-6 gap-y-4">
                <a href={githubUrl} className={primary}>
                  <Github size={18} aria-hidden="true" /> Explore the repository
                  <Arrow />
                </a>
                <a href={contributeUrl} className={textLink}>
                  Find a way to contribute <Arrow />
                </a>
              </div>
              <p className="mt-5 text-sm text-muted">
                Ready to try it on your server?{" "}
                <a
                  href={selfHostUrl}
                  className="font-medium text-forest underline decoration-forest/30 underline-offset-4 hover:decoration-forest"
                >
                  Follow the self-hosting guide{" "}
                  <span aria-hidden="true">↗</span>
                </a>
              </p>
            </div>
          </div>
        </section>

        <section
          id="questions"
          className="scroll-mt-24 px-5 py-20 md:px-10 md:py-28"
        >
          <div className="mx-auto max-w-[1040px]">
            <div data-reveal className="text-center">
              <p className={sectionLabel}>
                A few things you might be wondering
              </p>
              <h2 className={`${heading} mt-5`}>
                Good <em>questions.</em>
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
      </main>
      <footer className="bg-[#222d25] px-5 pb-7 pt-14 text-cream md:px-10 md:pt-16">
        <div className="mx-auto max-w-[1160px]">
          <div className="flex flex-wrap justify-between gap-12 border-b border-cream/20 pb-12">
            <div>
              <Brand light />
              <p className="mt-5 max-w-[240px] text-sm leading-relaxed text-cream/60">
                An open-source home for your email marketing. A little more
                human.
              </p>
            </div>
            <div className="grid grid-cols-2 gap-x-12 gap-y-8 text-sm sm:grid-cols-3 md:gap-x-20">
              <div>
                <p className={`${sectionLabel} mb-5 text-cream/65`}>
                  Make something
                </p>
                <a
                  className="mb-3 block hover:underline"
                  href={appLink("/campaigns")}
                >
                  Campaigns
                </a>
                <a
                  className="mb-3 block hover:underline"
                  href={appLink("/newsletters")}
                >
                  Newsletters
                </a>
                <a
                  className="mb-3 block hover:underline"
                  href={appLink("/automations")}
                >
                  Automations
                </a>
                <a className="block hover:underline" href="#templates">
                  Templates
                </a>
              </div>
              <div>
                <p className={`${sectionLabel} mb-5 text-cream/65`}>
                  Go a little deeper
                </p>
                <a
                  className="mb-3 block hover:underline"
                  href={appLink("/crm")}
                >
                  Audience & CRM
                </a>
                <a
                  className="mb-3 block hover:underline"
                  href={appLink("/forms")}
                >
                  Lead forms
                </a>
                <a
                  className="mb-3 block hover:underline"
                  href={appLink("/analytics")}
                >
                  Analytics
                </a>
                <a className="mb-3 block hover:underline" href="/mcp">
                  MCP server
                </a>
                <a className="block hover:underline" href="#sending">
                  Sending & setup
                </a>
              </div>
              <div>
                <p className={`${sectionLabel} mb-5 text-cream/65`}>
                  Come on in
                </p>
                <a
                  className="mb-3 block hover:underline"
                  href={appLink("/auth/login")}
                >
                  Log in <span aria-hidden="true">↗</span>
                </a>
                <a
                  className="mb-3 block hover:underline"
                  href={appLink("/auth/register")}
                >
                  Get started <span aria-hidden="true">↗</span>
                </a>
                <button
                  className="mb-3 block hover:underline"
                  onClick={() => setTour(0)}
                >
                  Product tour
                </button>
                <a className="mb-3 block hover:underline" href="/blog">
                  Journal
                </a>
                <a
                  className="mb-3 block hover:underline"
                  href="/blog/editorial"
                >
                  Editorial approach
                </a>
                <a className="mb-3 block hover:underline" href={githubUrl}>
                  Open source on GitHub <span aria-hidden="true">↗</span>
                </a>
                <a className="mb-3 block hover:underline" href={selfHostUrl}>
                  Self-host Xem <span aria-hidden="true">↗</span>
                </a>
                <a className="mb-3 block hover:underline" href={contributeUrl}>
                  Contribute to Xem <span aria-hidden="true">↗</span>
                </a>
                <a className="mb-3 block hover:underline" href="/google-mail">
                  Google mail · limited testing
                </a>
                <a className="mb-3 block hover:underline" href="/privacy">
                  Privacy policy · draft
                </a>
                <a className="mb-3 block hover:underline" href="/data-deletion">
                  Delete your data
                </a>
                <a
                  className="mb-3 block hover:underline"
                  href="mailto:team@xem.email"
                >
                  Contact the Xem team
                </a>
                <a className="block hover:underline" href="#questions">
                  FAQs
                </a>
              </div>
            </div>
          </div>
          <div className="flex flex-wrap items-center justify-between gap-4 pt-6 text-[11px] text-cream/50">
            <p>
              © {new Date().getFullYear()} Xem. Made for meaningful connections.
            </p>
            <p className="flex items-center gap-2">
              <span className="h-1.5 w-1.5 rounded-full bg-lemon" /> An email is
              only the beginning.
            </p>
          </div>
        </div>
      </footer>

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
