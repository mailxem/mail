"use client";

import { useState } from "react";
import Link from "next/link";
import {
  ArrowUpRight,
  Check,
  CheckCheck,
  Globe2,
  Mail,
  Send,
} from "lucide-react";
import { appLink } from "@/lib/site";
import { primaryButton } from "@/lib/brand";

const steps = [
  {
    label: "Make it yours",
    description: "Choose a sender and add your domain.",
    icon: Globe2,
    title: "Your name in their inbox.",
    detail:
      "Send from your own domain. Keep replies going to the mailbox you already use.",
    status: "Sender selected",
    rows: [
      ["From", "The Sunday Edit"],
      ["Email", "hello@example.com"],
      ["Reply to", "hello@example.com"],
    ],
    note: "Example sender details. Your existing inbox stays where it is.",
  },
  {
    label: "Give it the green light",
    description: "See which domain checks still need attention.",
    icon: CheckCheck,
    title: "A little clarity goes a long way.",
    detail:
      "Copy the DNS records for your domain and check their status. See when you are waiting on DNS or workspace review.",
    status: "Example: domain verified",
    rows: [
      ["Domain ownership", "Verified"],
      ["Email authentication", "Verified"],
      ["Workspace review", "Approved"],
    ],
    note: "Managed sending requires domain verification and workspace approval.",
  },
  {
    label: "Say a first hello",
    description: "Send yourself a test before your audience hears from you.",
    icon: Mail,
    title: "A first send, just for you.",
    detail:
      "Check your sender with a test, then follow its status. Provider acceptance is the first step; delivery events tell you what happens next.",
    status: "Example: test accepted",
    rows: [
      ["To", "you@example.com"],
      ["Subject", "A little hello from Xem"],
      ["Provider status", "Accepted"],
    ],
    note: "This is a preview. No email is sent, and acceptance does not guarantee inbox placement.",
  },
  {
    label: "Make something worth opening",
    description: "Start a campaign draft. Make the next send your own.",
    icon: Send,
    title: "Now, the good part.",
    detail:
      "Give your campaign a name and start shaping your message. Review the content and your opted-in audience before you send.",
    status: "Example: draft created",
    rows: [
      ["Campaign", "Our first Sunday Edit"],
      ["Status", "Draft"],
      ["Next step", "Choose a template"],
    ],
    note: "Creating a draft never sends a campaign automatically.",
  },
] as const;

export function SendingSection() {
  const [active, setActive] = useState(0);
  const step = steps[active];
  const Icon = step.icon;

  return (
    <section
      id="sending"
      aria-labelledby="sending-title"
      className="scroll-mt-28 bg-[#f0f0e5] px-5 py-20 md:px-10 md:py-28"
    >
      <div className="mx-auto max-w-[1160px]">
        <div
          data-reveal
          className="grid gap-6 lg:grid-cols-[1.2fr_.8fr] lg:items-end"
        >
          <div>
            <p className="text-[11px] font-medium uppercase tracking-[.17em] text-forest">
              A preview of your first steps
            </p>
            <h2
              id="sending-title"
              className="mt-5 font-editorial text-[clamp(3rem,5.5vw,5rem)] leading-[1.02] tracking-[-.045em]"
            >
              A good start.
              <br />
              <em>One small step at a time.</em>
            </h2>
          </div>
          <p className="max-w-md text-sm leading-relaxed text-muted">
            We’re building a calmer path to your first send: a checklist you can
            come back to, clear domain checks, and a little progress at every
            step. Take a look at the new setup flow below.
          </p>
        </div>

        <div
          data-reveal
          className="mt-10 grid overflow-hidden rounded-[28px] border border-forest/20 bg-cream lg:grid-cols-[.85fr_1.15fr]"
        >
          <div className="p-5 sm:p-8">
            <p className="mb-5 text-xs font-medium text-muted">
              Explore the checklist
            </p>
            <ol className="space-y-2">
              {steps.map((item, index) => (
                <li key={item.label}>
                  <button
                    type="button"
                    aria-pressed={active === index}
                    aria-controls="sending-preview"
                    onClick={() => setActive(index)}
                    className={`flex w-full items-start gap-3 rounded-2xl border p-4 text-left transition-colors sm:gap-4 ${active === index ? "border-forest bg-forest text-cream" : "border-transparent hover:border-forest/20 hover:bg-forest/5"}`}
                  >
                    <span
                      aria-hidden="true"
                      className={`flex h-7 w-7 shrink-0 items-center justify-center rounded-full border text-xs ${active === index ? "border-cream/40" : "border-forest/25"}`}
                    >
                      {index + 1}
                    </span>
                    <span>
                      <span className="block text-sm font-medium">
                        {item.label}
                      </span>
                      <span
                        className={`mt-1.5 block text-xs leading-relaxed ${active === index ? "text-cream/80" : "text-muted"}`}
                      >
                        {item.description}
                      </span>
                    </span>
                  </button>
                </li>
              ))}
            </ol>
          </div>
          <div
            id="sending-preview"
            role="region"
            aria-label="Setup preview"
            aria-live="polite"
            aria-atomic="true"
            className="flex min-w-0 flex-col border-t border-forest/15 bg-[#e3e9d9] p-6 sm:p-10 lg:border-l lg:border-t-0"
          >
            <div className="flex items-center justify-between text-[11px] text-forest">
              <span className="rounded-full border border-forest/25 px-3 py-1.5">
                Illustrative preview
              </span>
              <span>Step {active + 1} of 4</span>
            </div>
            <Icon
              size={28}
              strokeWidth={1.3}
              aria-hidden="true"
              className="mt-8 text-forest"
            />
            <h3 className="mt-4 font-editorial text-[clamp(2rem,3.5vw,2.7rem)] leading-[1.08] tracking-tight">
              {step.title}
            </h3>
            <p className="mt-4 min-h-[4.5rem] text-sm leading-relaxed text-ink/80">
              {step.detail}
            </p>
            <div className="mt-7 rounded-2xl border border-forest/15 bg-cream p-5">
              <p className="flex items-center gap-2 text-xs font-medium text-forest">
                <Check size={15} aria-hidden="true" />
                {step.status}
              </p>
              <dl className="mt-4 space-y-3 border-t border-forest/15 pt-4 text-xs">
                {step.rows.map(([label, value]) => (
                  <div
                    key={label}
                    className="flex flex-wrap justify-between gap-x-4 gap-y-1"
                  >
                    <dt className="text-muted">{label}</dt>
                    <dd className="break-words font-medium">{value}</dd>
                  </div>
                ))}
              </dl>
            </div>
            <p className="mt-5 text-xs leading-relaxed text-ink/75">
              {step.note}
            </p>
          </div>
        </div>

        <div data-reveal className="mt-8 grid gap-5 md:grid-cols-2">
          <article className="rounded-2xl border border-ink/15 p-6 sm:p-8">
            <p className="text-[11px] font-medium uppercase tracking-widest text-forest">
              Available today
            </p>
            <h3 className="mt-4 font-editorial text-3xl tracking-tight">
              Your provider. Your choice.
            </h3>
            <p className="mt-4 text-sm leading-relaxed text-muted">
              Already have an email provider? Connect your SMTP sender to Xem
              and keep the setup that works for you.
            </p>
            <a
              href={appLink("/settings")}
              className="mt-5 inline-flex min-h-11 items-center gap-2 text-sm font-medium text-iris"
            >
              Connect your provider{" "}
              <ArrowUpRight size={16} aria-hidden="true" />
            </a>
          </article>
          <article className="rounded-2xl border border-iris/20 bg-lavender/40 p-6 sm:p-8">
            <p className="text-[11px] font-medium uppercase tracking-widest text-iris">
              Available today · Managed sending
            </p>
            <h3 className="mt-4 font-editorial text-3xl tracking-tight">
              Your domain. Sent with Xem.
            </h3>
            <p className="mt-4 text-sm leading-relaxed text-muted">
              Send through Xem’s managed Amazon SES delivery, with verified
              domains, delivery history, and revocable SMTP credentials for your
              other tools. Domain verification and workspace approval are
              required.
            </p>
            <p className="mt-4 text-xs leading-relaxed text-muted">
              Managed sending handles outgoing email. Your existing inbox and
              replies stay with your mailbox provider.
            </p>
            <Link
              href="/blog/managed-smtp-custom-domains"
              className="mt-5 inline-flex min-h-11 items-center gap-2 text-sm font-medium text-iris"
            >
              Explore managed sending{" "}
              <ArrowUpRight size={16} aria-hidden="true" />
            </Link>
          </article>
        </div>
        <div data-reveal className="mt-8 flex flex-wrap items-center gap-5">
          <a href={appLink("/auth/register")} className={primaryButton}>
            Make yourself at home <ArrowUpRight size={17} aria-hidden="true" />
          </a>
          <p className="text-xs text-muted">
            Choose your own SMTP provider or Xem’s managed sending.
          </p>
        </div>
      </div>
    </section>
  );
}
