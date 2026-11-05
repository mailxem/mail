"use client";
import Link from "next/link";
import { Check, ChevronRight, X } from "lucide-react";
import { useLocalStorage } from "usehooks-ts";
import { Button } from "@/components/ui/button";
import { useResourcePage } from "@/hooks/use-resource-page";
import { useMarketingQuery } from "@/lib/marketing/api";
import type { LeadForm } from "@/lib/marketing/types";
import type { MailSender } from "@/lib/connected-mail";

/** First-run guide for every member; each step ticks itself off from workspace data. */
export function GetStarted() {
  const [hidden, setHidden] = useLocalStorage("xem-get-started-hidden", false);
  const senders = useMarketingQuery<MailSender[]>("mail-connections/senders");
  const contacts = useResourcePage("contacts", 1, 1);
  const campaigns = useResourcePage("campaigns", 1, 1);
  const forms = useMarketingQuery<LeadForm[]>("marketing/forms");
  const loading = [senders, contacts, campaigns, forms].some((q) => q.isLoading);
  if (hidden || loading) return null;

  const steps = [
    { done: !!senders.data?.length, title: "Connect a sender", body: "Use Xem’s managed sending or your own SMTP so email can leave your workspace.", href: "/settings/smtp", cta: "Connect sender" },
    { done: !!contacts.data?.total, title: "Add your contacts", body: "Create a list, then import a CSV or an export from Mailchimp, Substack or beehiiv.", href: "/audience/lists", cta: "Add contacts" },
    { done: !!campaigns.data?.total, title: "Send your first campaign", body: "Pick one of 200+ editable templates, choose a list and send now or schedule it.", href: "/campaigns/new", cta: "Create campaign" },
    { done: !!forms.data?.length, title: "Grow with a signup form", body: "Publish a form so new subscribers land in your lists on their own.", href: "/forms/new", cta: "Create form" },
  ];
  const done = steps.filter((s) => s.done).length;
  if (done === steps.length) return null;
  const next = steps.findIndex((s) => !s.done);

  return (
    <section aria-labelledby="get-started" className="rounded-xl border border-border bg-card p-5">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h2 id="get-started" className="text-base font-semibold">Get started with Xem</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            {done} of {steps.length} done. Follow these in order and you’ll have your first campaign out in minutes.
          </p>
        </div>
        <Button variant="ghost" size="icon" className="size-8 shrink-0" aria-label="Hide getting started guide" onClick={() => setHidden(true)}>
          <X size={16} />
        </Button>
      </div>
      <div className="mt-3 h-1.5 overflow-hidden rounded-full bg-muted" role="progressbar" aria-valuemin={0} aria-valuemax={steps.length} aria-valuenow={done}>
        <div className="h-full bg-primary transition-all" style={{ width: `${(done / steps.length) * 100}%` }} />
      </div>
      <ol className="mt-4 divide-y divide-border">
        {steps.map((s, i) => (
          <li key={s.title} className="flex flex-wrap items-center gap-3 py-3">
            <span className={`grid size-6 shrink-0 place-items-center rounded-full border text-xs ${s.done ? "border-primary bg-primary text-primary-foreground" : "border-border text-muted-foreground"}`}>
              {s.done ? <Check size={14} /> : i + 1}
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
    </section>
  );
}
