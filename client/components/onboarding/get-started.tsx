"use client";
import Link from "next/link";
import { Check, ChevronRight, FileText, Megaphone, Send, Users, X, type LucideIcon } from "lucide-react";
import { useLocalStorage } from "usehooks-ts";
import { Button } from "@/components/ui/button";
import { useResourcePage } from "@/hooks/use-resource-page";
import { useMarketingQuery } from "@/lib/marketing/api";
import type { LeadForm } from "@/lib/marketing/types";
import type { MailSender } from "@/lib/connected-mail";
import { GridIllustration } from "./grid-illustration";

type Step = {
  done: boolean;
  title: string;
  body: string;
  href: string;
  cta: string;
  icon: LucideIcon;
  /** Tile background, icon colour and dark-mode tile background. */
  tint: [string, string, string];
};

/** First-run guide for every member; each step ticks itself off from workspace data. */
export function GetStarted() {
  const [hidden, setHidden] = useLocalStorage("xem-get-started-hidden", false);
  const senders = useMarketingQuery<MailSender[]>("mail-connections/senders");
  const contacts = useResourcePage("contacts", 1, 1);
  const campaigns = useResourcePage("campaigns", 1, 1);
  const forms = useMarketingQuery<LeadForm[]>("marketing/forms");
  const loading = [senders, contacts, campaigns, forms].some((q) => q.isLoading);
  if (hidden || loading) return null;

  const steps: Step[] = [
    { done: !!senders.data?.length, title: "Connect a sender", body: "Use Xem’s managed sending or your own SMTP so email can leave your workspace.", href: "/settings/smtp", cta: "Connect sender", icon: Send, tint: ["#e6f0ff", "#1d5fd6", "rgb(59 111 216 / 0.22)"] },
    { done: !!contacts.data?.total, title: "Add your contacts", body: "Create a list, then import a CSV or an export from Mailchimp, Substack or beehiiv.", href: "/audience/lists", cta: "Add contacts", icon: Users, tint: ["#e5f6ec", "#1f7a45", "rgb(47 158 91 / 0.22)"] },
    { done: !!campaigns.data?.total, title: "Send your first campaign", body: "Pick one of 80+ editable templates, choose a list and send now or schedule it.", href: "/campaigns/new", cta: "Create campaign", icon: Megaphone, tint: ["#ffece3", "#c2410c", "rgb(240 100 74 / 0.22)"] },
    { done: !!forms.data?.length, title: "Grow with a signup form", body: "Publish a form so new subscribers land in your lists on their own.", href: "/forms/new", cta: "Create form", icon: FileText, tint: ["#efe7ff", "#6b3fd4", "rgb(139 92 246 / 0.24)"] },
  ];
  const done = steps.filter((s) => s.done).length;
  if (done === steps.length) return null;
  const next = steps.findIndex((s) => !s.done);

  return (
    <section
      aria-labelledby="get-started"
      className="relative grid overflow-hidden rounded-2xl border border-border bg-card md:grid-cols-[1fr_30%]"
    >
      <div className="p-5">
        <div className="flex items-start justify-between gap-4">
          <div>
            <h2 id="get-started" className="text-base font-semibold">Get started with Xem</h2>
            <p className="mt-1 text-sm text-muted-foreground">
              {done} of {steps.length} done. Follow these in order and you’ll have your first campaign out in minutes.
            </p>
          </div>
          <Button variant="ghost" size="icon" className="size-8 shrink-0 md:hidden" aria-label="Hide getting started guide" onClick={() => setHidden(true)}>
            <X size={16} />
          </Button>
        </div>
        <div
          className="mt-3 grid grid-cols-4 gap-1"
          role="progressbar"
          aria-label="Setup progress"
          aria-valuemin={0}
          aria-valuemax={steps.length}
          aria-valuenow={done}
        >
          {steps.map((s) => (
            <span
              key={s.title}
              className="h-1.5 rounded-full bg-muted transition-colors"
              style={s.done ? { background: s.tint[1] } : undefined}
            />
          ))}
        </div>
        <ol className="mt-3 divide-y divide-border">
          {steps.map((s, i) => (
            <li key={s.title} className="flex flex-wrap items-center gap-3 py-3">
              <span
                className="grid size-8 shrink-0 place-items-center rounded-[10px] [background:var(--step-bg)] dark:[background:var(--step-bg-dark)]"
                style={{ "--step-bg": s.tint[0], "--step-bg-dark": s.tint[2], color: s.tint[1] } as React.CSSProperties}
              >
                {s.done ? <Check size={16} strokeWidth={2.25} /> : <s.icon size={16} strokeWidth={1.75} />}
              </span>
              <div className="min-w-0 flex-1">
                <p className={`text-sm font-medium ${s.done ? "text-muted-foreground line-through" : ""}`}>{s.title}</p>
                {!s.done && <p className="mt-0.5 text-xs text-muted-foreground">{s.body}</p>}
              </div>
              {!s.done && (
                <Button asChild size="sm" variant={i === next ? "default" : "outline"}>
                  <Link href={s.href}>
                    {s.cta}
                    <ChevronRight className="ml-1 size-4" />
                  </Link>
                </Button>
              )}
            </li>
          ))}
        </ol>
      </div>
      <div className="relative hidden border-l border-border md:block">
        <GridIllustration className="absolute inset-0 size-full" />
        <Button
          variant="secondary"
          size="icon"
          className="absolute right-3 top-3 size-8 rounded-full bg-background/90 shadow-sm backdrop-blur"
          aria-label="Hide getting started guide"
          onClick={() => setHidden(true)}
        >
          <X size={16} />
        </Button>
      </div>
    </section>
  );
}
