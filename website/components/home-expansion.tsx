"use client";
import { useState } from "react";
import Image from "next/image";
import Link from "next/link";
import {
  ArrowLeft,
  ArrowRight,
  ArrowUpRight,
  Braces,
  Check,
  FileInput,
  Inbox,
  Users,
  Webhook,
} from "lucide-react";
import { appLink } from "@/lib/site";
import { primaryButton } from "@/lib/brand";
import type { PostMeta } from "@/lib/blog-types";
import { PostCard } from "./blog/post-card";
const features = [
  {
    label: "Lead forms",
    icon: FileInput,
    title: "A better first hello.",
    description:
      "Turn an interested visitor into a contact. Build a lead form, connect it to a list, and keep the signup experience connected to your email workflow.",
    image: "forms",
    path: "/forms",
    points: [
      "Forms connected to audience lists",
      "Fields and consent settings you control",
      "Submission records in your workspace",
    ],
  },
  {
    label: "Audience & CRM",
    icon: Users,
    title: "Remember the person behind the address.",
    description:
      "Keep contacts, lists, and tags close to the work. Find the people you need, understand their details, and organize the audience for your next send.",
    image: "crm",
    path: "/crm",
    points: [
      "Searchable contact records",
      "Lists and tags that stay useful",
      "A shared view for your team",
    ],
  },
  {
    label: "Inbox & outbox",
    icon: Inbox,
    title: "Keep the conversation going.",
    description:
      "Bring your connected mailbox into the same workspace. Read incoming messages, compose a reply, and review outgoing email without losing the thread.",
    image: "inbox",
    path: "/inbox",
    points: [
      "Connect an IMAP mailbox",
      "Compose with rich email content",
      "Review outgoing messages in the outbox",
    ],
  },
];
export function MoreFeatures() {
  const [active, setActive] = useState(0);
  const f = features[active];
  return (
    <section
      id="more-features"
      className="scroll-mt-24 px-5 py-20 md:px-10 md:py-24"
    >
      <div className="mx-auto max-w-[1160px]">
        <div
          data-reveal
          className="flex flex-wrap items-end justify-between gap-6"
        >
          <div>
            <p className="text-[11px] font-medium uppercase tracking-[.17em] text-iris">
              The rest of the relationship
            </p>
            <h2 className="mt-5 font-editorial text-[clamp(3rem,5vw,4.8rem)] leading-[1.02] tracking-[-.04em]">
              More than
              <br />
              <em>the send button.</em>
            </h2>
          </div>
          <p className="max-w-sm text-sm leading-relaxed text-muted">
            The form that starts it. The contact you remember. The reply that
            turns into a conversation. Give every part a home.
          </p>
        </div>
        <div
          data-reveal
          className="mt-10 overflow-hidden rounded-[24px] border border-ink/15 bg-[#f2eee7]"
        >
          <div
            role="tablist"
            aria-label="More Xem features"
            className="flex overflow-x-auto border-b border-ink/15 bg-cream/60 p-3"
          >
            {features.map((item, i) => {
              const Icon = item.icon;
              return (
                <button
                  key={item.label}
                  id={`feature-tab-${i}`}
                  role="tab"
                  aria-selected={active === i}
                  aria-controls="feature-panel"
                  tabIndex={active === i ? 0 : -1}
                  onClick={() => setActive(i)}
                  onKeyDown={(e) => {
                    let next = i;
                    if (e.key === "ArrowRight") next = (i + 1) % 3;
                    else if (e.key === "ArrowLeft") next = (i + 2) % 3;
                    else if (e.key === "Home") next = 0;
                    else if (e.key === "End") next = 2;
                    else return;
                    e.preventDefault();
                    setActive(next);
                    document.getElementById(`feature-tab-${next}`)?.focus();
                  }}
                  className={`flex min-h-11 shrink-0 items-center gap-2 rounded-lg px-4 text-xs transition-colors sm:px-6 sm:text-sm ${active === i ? "bg-iris text-white shadow-sm" : "text-muted hover:bg-ink/5"}`}
                >
                  <Icon size={16} />
                  {item.label}
                </button>
              );
            })}
          </div>
          <div
            id="feature-panel"
            role="tabpanel"
            aria-labelledby={`feature-tab-${active}`}
            className="grid lg:grid-cols-[.8fr_1.4fr]"
          >
            <div className="flex flex-col justify-center p-7 md:p-10">
              <span className="text-[10px] text-muted">
                0{active + 1} / A connected workspace
              </span>
              <h3 className="mt-6 font-editorial text-4xl leading-[1.08] tracking-tight">
                {f.title}
              </h3>
              <p className="mt-5 text-sm leading-relaxed text-muted">
                {f.description}
              </p>
              <ul className="mt-6 space-y-3">
                {f.points.map((p) => (
                  <li key={p} className="flex items-start gap-2 text-xs">
                    <Check size={14} className="mt-px shrink-0 text-iris" />
                    {p}
                  </li>
                ))}
              </ul>
              <a
                href={appLink(f.path)}
                className={`${primaryButton} mt-8 self-start`}
              >
                Explore {f.label.toLowerCase()} <ArrowUpRight size={17} />
              </a>
            </div>
            <div className="flex min-w-0 items-center overflow-hidden bg-[#e4ddec] p-5 pb-0 md:p-8 md:pb-0 lg:pt-12">
              <div className="w-full self-end overflow-hidden rounded-t-xl border border-b-0 border-ink/20 bg-white shadow-[0_0_40px_#30233c12]">
                <div className="flex gap-1.5 border-b bg-[#faf9fb] px-4 py-3">
                  <i className="h-1.5 w-1.5 rounded-full bg-ink/20" />
                  <i className="h-1.5 w-1.5 rounded-full bg-ink/20" />
                  <i className="h-1.5 w-1.5 rounded-full bg-ink/20" />
                  <span className="ml-auto text-[9px] text-muted">
                    Your Xem workspace
                  </span>
                </div>
                <Image
                  src={`/images/product/${f.image}.webp`}
                  alt={`Xem ${f.label.toLowerCase()} interface with illustrative workspace data`}
                  width={1512}
                  height={1100}
                  quality={90}
                  sizes="(max-width:1024px) 92vw, 720px"
                  className="h-auto w-full"
                />
              </div>
            </div>
          </div>
        </div>
        <p className="mt-3 text-right text-[10px] text-muted">
          Actual Xem interfaces, shown with illustrative workspace data.
        </p>
        <div data-reveal className="mt-7 grid gap-5 md:grid-cols-2">
          {[
            {
              icon: Braces,
              title: "Your sender. Your setup.",
              body: "Connect your SMTP provider and IMAP mailbox. Keep the email tools you use, with one place to manage the work.",
              href: appLink("/settings"),
              link: "Connect your workspace",
            },
            {
              icon: Webhook,
              title: "Fits into the way you work.",
              body: "Use API keys and webhooks to connect Xem to your product. Manage the integrations alongside your workspace settings.",
              href: "https://docs.xem.email",
              link: "Explore the documentation",
            },
          ].map((c) => (
            <div
              key={c.title}
              className="flex items-start gap-5 rounded-2xl border border-ink/15 p-6 md:p-7"
            >
              <span className="rounded-xl bg-lavender/60 p-3 text-iris">
                <c.icon size={22} />
              </span>
              <div>
                <h3 className="font-editorial text-3xl tracking-tight">
                  {c.title}
                </h3>
                <p className="mt-3 text-sm leading-relaxed text-muted">
                  {c.body}
                </p>
                <a
                  className="mt-5 inline-flex items-center gap-2 text-xs font-medium text-iris"
                  href={c.href}
                >
                  {c.link}
                  <ArrowUpRight size={14} />
                </a>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

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
            In good company
          </p>
          <h2
            id="voices-heading"
            className="mt-5 font-editorial text-5xl leading-[1.04] tracking-tight"
          >
            Real people.
            <br />
            <em>Better email days.</em>
          </h2>
          <p className="mt-6 max-w-xs text-sm leading-relaxed text-cream/65">
            A few words from the teams who have made Xem part of their work.
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
          <p className="mt-5 text-[10px] text-cream/55">
            Testimonials supplied by Xem. Individual experiences may vary.
          </p>
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
              Fresh from the journal
            </p>
            <h2 className="mt-5 font-editorial text-[clamp(3rem,5vw,4.8rem)] leading-[1.02] tracking-[-.04em]">
              Ideas worth
              <br />
              <em>opening.</em>
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
