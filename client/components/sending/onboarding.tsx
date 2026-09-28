"use client";
import { useState } from "react";
import { useApi } from "@/hooks/use-api";
import Link from "next/link";
import {
  ArrowRight,
  Check,
  ChevronRight,
  Cloud,
  PlugZap,
  Send,
  CheckCheck,
  CircleHelp,
} from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { useMarketing, useMarketingQuery } from "@/lib/marketing/api";
import { DomainDNS, DomainForm, SenderSetup } from "./controls";
import {
  checklist,
  onboardingDomain,
  onboardingNavigation,
  type SendingState,
} from "@/lib/sending/types";

export function OnboardingBanner() {
  const { session } = useApi();
  const isAdmin = ["ADMIN", "SUPER_ADMIN"].includes(session?.user?.role || "");
  const q = useMarketingQuery<SendingState>("sending", isAdmin, 15000);
  if (!q.data || q.data.account.onboardingDismissed) return null;
  const steps = checklist(q.data);
  const done = steps.filter((s) => s.done).length;
  if (done === steps.length) return null;
  return (
    <Link
      href="/onboarding"
      className="group flex flex-wrap items-center justify-between gap-4 rounded-xl border border-border bg-card px-5 py-4 transition-colors hover:bg-muted/40"
    >
      <div className="flex items-center gap-4">
        <span className="grid size-9 shrink-0 place-items-center rounded-lg border border-border text-muted-foreground">
          <Send size={16} />
        </span>
        <div>
          <p className="text-sm font-medium">
            Finish setting up your workspace
          </p>
          <p className="mt-1 text-xs text-muted-foreground">
            {done} of {steps.length} steps complete. Pick up where you left off.
          </p>
        </div>
      </div>
      <span className="flex items-center gap-2 text-xs font-medium text-foreground">
        Continue setup{" "}
        <ArrowRight
          size={16}
          className="transition-transform group-hover:translate-x-1 motion-reduce:transform-none"
        />
      </span>
    </Link>
  );
}
export function Onboarding() {
  const q = useMarketingQuery<SendingState>("sending", true, 15000);
  const { request, refresh } = useMarketing();
  const [selected, setSelected] = useState<number | null>(null);
  const [busy, setBusy] = useState(false);
  const sync = async () => {
    await refresh();
  };
  if (q.isLoading)
    return (
      <div
        role="status"
        className="mx-auto max-w-5xl p-10 text-muted-foreground"
      >
        Getting your welcome checklist ready…
      </div>
    );
  if (q.error || !q.data)
    return (
      <div className="mx-auto max-w-xl space-y-4 rounded-xl border p-8">
        <h1 className="text-xl font-medium">Let’s get your workspace ready</h1>
        <p role="alert" className="text-sm text-muted-foreground">
          {q.error?.message || "Sign in to continue."} A workspace administrator
          can connect your sender.
        </p>
        <Button variant="outline" onClick={() => void q.refetch()}>
          Try again
        </Button>
        <Link href="/" className="ml-4 text-sm underline">
          Back to dashboard
        </Link>
      </div>
    );
  const state = q.data,
    steps = checklist(state),
    done = steps.filter((s) => s.done).length,
    complete = done === steps.length;
  const { active, unlocked } = onboardingNavigation(steps, selected);
  const managed = state.account.sendingMode === "MANAGED";
  const domain = onboardingDomain(state);
  async function choose(mode: string) {
    setBusy(true);
    try {
      await request("sending/onboarding", "PUT", { mode, dismissed: false });
      setSelected(null);
      await sync();
    } catch (e) {
      toast.error((e as Error).message);
    } finally {
      setBusy(false);
    }
  }
  async function dismiss() {
    if (state.account.sendingMode) {
      try {
        await request("sending/onboarding", "PUT", {
          mode: state.account.sendingMode,
          dismissed: true,
        });
        sync();
      } catch (e) {
        toast.error((e as Error).message);
      }
    }
  }
  return (
    <div className="mx-auto max-w-6xl space-y-7 pb-10">
      <div className="flex items-center justify-between gap-4">
        <Link
          href="/"
          className="text-sm text-muted-foreground hover:text-foreground"
        >
          ← Back to your workspace
        </Link>
        <Link
          href="/settings/sending"
          className="text-sm text-muted-foreground hover:text-foreground"
        >
          Sending settings <ChevronRight size={13} className="inline" />
        </Link>
      </div>
      <header className="flex flex-wrap items-end justify-between gap-6 border-b border-border pb-7">
        <div className="max-w-xl">
          <h1 className="text-2xl font-medium tracking-tight">
            {complete ? "Your workspace is ready" : "Set up your workspace"}
          </h1>
          <p className="mt-2 text-sm leading-6 text-muted-foreground">
            {complete
              ? "Your setup is complete. Review your campaign and choose when to send."
              : "Connect a sender, prepare your domain, and send your first email."}
          </p>
        </div>
        <div className="min-w-40 space-y-3">
          <p className="text-xs text-muted-foreground">
            {done} of {steps.length} steps complete
          </p>
          <div
            role="progressbar"
            aria-label="Onboarding progress"
            aria-valuemin={0}
            aria-valuemax={steps.length}
            aria-valuenow={done}
            className="h-1 w-40 overflow-hidden rounded-full bg-muted"
          >
            <div
              className="h-full rounded-full bg-primary transition-all duration-500 motion-reduce:transition-none"
              style={{ width: `${(done / steps.length) * 100}%` }}
            />
          </div>
        </div>
      </header>
      <div className="grid items-start gap-6 lg:grid-cols-[264px_minmax(0,1fr)]">
        <aside className="order-2 space-y-5 lg:order-1">
          <nav
            aria-label="Welcome checklist"
            className="overflow-hidden rounded-xl border border-border bg-card"
          >
            <p className="border-b border-border px-4 py-4 text-xs font-medium text-muted-foreground">
              Setup checklist
            </p>
            <ol className="divide-y divide-border">
              {steps.map((step, i) => (
                <li key={step.title}>
                  <button
                    onClick={() => setSelected(i)}
                    disabled={busy || i > unlocked}
                    aria-current={active === i ? "step" : undefined}
                    className={`flex w-full gap-3 p-4 text-left transition-colors disabled:cursor-not-allowed disabled:opacity-50 ${active === i ? "bg-muted/70" : "hover:bg-muted/40"}`}
                  >
                    <span
                      className={`mt-0.5 grid size-6 shrink-0 place-items-center rounded-full border text-xs ${step.done ? "border-border bg-muted text-foreground" : active === i ? "border-primary text-primary" : "border-border text-muted-foreground"}`}
                    >
                      {step.done ? <Check size={15} /> : i + 1}
                    </span>
                    <span>
                      <span className="block text-sm font-medium leading-5">
                        {step.title}
                      </span>
                      <span className="mt-1 block text-xs leading-5 text-muted-foreground">
                        {step.done ? "All set" : step.detail}
                      </span>
                    </span>
                  </button>
                </li>
              ))}
            </ol>
          </nav>
          <div className="px-4 text-xs leading-6 text-muted-foreground">
            <CircleHelp size={15} className="mb-2" />
            <p>
              Your progress saves as you go. Explore Xem whenever you like, then
              return here to finish.
            </p>
            <Link
              href="/templates"
              className="mt-2 inline-flex items-center gap-1 font-medium text-foreground"
            >
              Explore email templates <ArrowRight size={13} />
            </Link>
          </div>
        </aside>
        <section
          aria-live="polite"
          className="order-1 min-w-0 rounded-xl border border-border bg-card p-5 sm:p-6 lg:order-2"
        >
          {complete && selected === null ? (
            <div className="py-8 text-center">
              <span className="mx-auto grid size-12 place-items-center rounded-xl border border-border bg-muted text-foreground">
                <CheckCheck size={24} />
              </span>
              <h2 className="mt-5 text-xl font-medium">
                You’re ready to send.
              </h2>
              <p className="mx-auto mt-3 max-w-md text-sm leading-7 text-muted-foreground">
                Your setup is complete. Review your campaign, make it yours, and
                choose when to send it.
              </p>
              <Button asChild className="mt-6">
                <Link href="/campaigns">
                  Open my campaigns <ArrowRight className="ml-2" size={16} />
                </Link>
              </Button>
            </div>
          ) : (
            <>
              <p className="mb-2 text-xs text-muted-foreground">
                Step {active + 1} of {steps.length}
              </p>
              <h2 className="mb-2 text-lg font-medium tracking-tight">
                {steps[active].title}
              </h2>
              <p className="mb-6 text-sm leading-6 text-muted-foreground">
                {steps[active].detail}
              </p>
              {active === 0 ? (
                <div className="overflow-hidden rounded-xl border border-border divide-y divide-border">
                  {[
                    {
                      mode: "MANAGED",
                      title: "Let Xem handle delivery",
                      body: "Connect your domain. We’ll provide the sending service and keep delivery details in one place.",
                      icon: Cloud,
                    },
                    {
                      mode: "BYO",
                      title: "Bring your own provider",
                      body: "Already using SES, Postmark, or another SMTP provider? Keep your setup and connect it here.",
                      icon: PlugZap,
                    },
                  ].map((option) => (
                    <button
                      key={option.mode}
                      disabled={
                        busy || (option.mode === "MANAGED" && !state.enabled)
                      }
                      onClick={() => void choose(option.mode)}
                      aria-pressed={state.account.sendingMode === option.mode}
                      className={`group flex w-full gap-4 p-5 text-left transition-colors hover:bg-muted/40 disabled:opacity-50 ${state.account.sendingMode === option.mode ? "bg-muted/60 ring-1 ring-inset ring-primary" : ""}`}
                    >
                      <option.icon
                        size={20}
                        className="mt-0.5 shrink-0 text-muted-foreground"
                      />
                      <span className="min-w-0 flex-1">
                        <span className="block text-sm font-medium">
                          {option.title}
                        </span>
                        <span className="mt-2 block text-xs leading-6 text-muted-foreground">
                          {option.body}
                        </span>
                        <span className="mt-3 flex items-center gap-2 text-xs font-medium text-primary">
                          {option.mode === "MANAGED" && !state.enabled
                            ? "Not enabled on this installation"
                            : "Choose this path"}
                          <ArrowRight size={15} />
                        </span>
                      </span>
                    </button>
                  ))}
                </div>
              ) : managed && active === 1 ? (
                domain ? (
                  <div className="space-y-4">
                    <p className="font-medium break-all">{domain.name}</p>
                    <p className="text-sm text-muted-foreground">
                      Your domain is saved. Let’s connect its DNS records.
                    </p>
                    <Button onClick={() => setSelected(2)}>
                      Continue to DNS <ArrowRight size={15} className="ml-2" />
                    </Button>
                  </div>
                ) : (
                  <DomainForm
                    onDone={async () => {
                      await sync();
                      setSelected(2);
                    }}
                  />
                )
              ) : managed && active === 2 ? (
                domain ? (
                  <DomainDNS
                    key={domain.id}
                    domain={domain}
                    onDone={async () => {
                      setSelected(2);
                      await sync();
                    }}
                  />
                ) : (
                  <p className="text-sm text-muted-foreground">
                    Add your domain in the previous step to see its DNS records.
                  </p>
                )
              ) : managed && active === 3 ? (
                domain ? (
                  <div className="space-y-5">
                    <SenderSetup
                      key={domain.id}
                      domain={domain}
                      approved={
                        state.account.approved &&
                        !state.account.paused &&
                        !state.account.suspended
                      }
                      onDone={sync}
                    />
                    <p
                      role="status"
                      className="rounded-xl bg-muted/60 p-4 text-sm leading-6"
                    >
                      {steps[3].done
                        ? "Your sender is saved and your test was accepted by the provider. You can continue to your first campaign."
                        : !domain.smtpConfigId
                          ? "Save your sender and send yourself a test to complete this step."
                          : "Send yourself a test, then wait for the provider to accept it. This checklist updates automatically."}
                      <Link href="/settings/sending" className="ml-1 underline">
                        View sending activity
                      </Link>
                    </p>
                  </div>
                ) : (
                  <p className="text-sm text-muted-foreground">
                    Connect your domain first. We’ll prepare your sender here.
                  </p>
                )
              ) : !managed && active === 1 ? (
                <div className="rounded-xl border border-border bg-muted/30 p-5">
                  <PlugZap className="mb-4 text-muted-foreground" />
                  <p className="mb-5 text-sm leading-7">
                    You’ll need your provider’s SMTP host, port, username, and
                    password. Xem checks the connection before saving it.
                  </p>
                  <Button asChild>
                    <Link href="/settings/smtp">
                      Connect my provider{" "}
                      <ArrowRight size={15} className="ml-2" />
                    </Link>
                  </Button>
                </div>
              ) : (
                <div className="rounded-xl border border-border bg-muted/30 p-5">
                  <MailArt />
                  <p className="my-5 text-sm leading-7">
                    {!managed && active === 3
                      ? "Open your campaign, choose your connected sender, and review the audience before sending. Your checklist updates when the provider accepts an email."
                      : "Start with a template or a blank canvas. Save a draft now; you decide when it goes out."}
                  </p>
                  <Button asChild>
                    <Link
                      href={
                        !managed && active === 3
                          ? "/campaigns"
                          : "/campaigns/new"
                      }
                    >
                      {!managed && active === 3
                        ? "Open my campaigns"
                        : "Create my first campaign"}
                      <ArrowRight size={15} className="ml-2" />
                    </Link>
                  </Button>
                </div>
              )}
              <div className="mt-8 flex flex-wrap items-center justify-between gap-3 border-t pt-5">
                <Link
                  href="/"
                  onClick={() => void dismiss()}
                  className="text-xs text-muted-foreground hover:text-foreground"
                >
                  I’ll finish this later
                </Link>
                {active < steps.length - 1 && (
                  <Button
                    variant="ghost"
                    disabled={busy || !steps[active].done}
                    onClick={() => setSelected(active + 1)}
                  >
                    {managed && active === 2
                      ? domain?.ready
                        ? "Continue to sender"
                        : "Complete DNS setup first"
                      : "Next step"}{" "}
                    <ChevronRight size={15} className="ml-1" />
                  </Button>
                )}
              </div>
            </>
          )}
        </section>
      </div>
    </div>
  );
}
function MailArt() {
  return (
    <span className="inline-grid size-12 place-items-center rounded-xl border bg-background text-muted-foreground">
      <Send size={23} strokeWidth={1.4} />
    </span>
  );
}
