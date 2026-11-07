"use client";
import Link from "next/link";
import { ArrowRight, Check, ShieldAlert, ShieldCheck } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useApi } from "@/hooks/use-api";
import { useMarketingQuery } from "@/lib/marketing/api";
import type { MailSender } from "@/lib/connected-mail";
import { onboardingDomain, type SendingState } from "@/lib/sending/types";

const dots = {
  backgroundImage: "radial-gradient(currentColor 1px, transparent 1px)",
  backgroundSize: "14px 14px",
  maskImage: "linear-gradient(to left, black, transparent 70%)",
  WebkitMaskImage: "linear-gradient(to left, black, transparent 70%)",
};

/** Tells people up front whether email will actually leave the workspace. */
export function SendingStatus() {
  const { session } = useApi();
  const isAdmin = ["ADMIN", "SUPER_ADMIN"].includes(session?.user?.role || "");
  const senders = useMarketingQuery<MailSender[]>("mail-connections/senders");
  const sending = useMarketingQuery<SendingState>("sending", isAdmin, 30000);
  if (senders.isLoading || !senders.data) return null;

  const list = senders.data;
  const domain = sending.data ? onboardingDomain(sending.data) : undefined;
  const verified = list.length > 0;
  const checks: [string, boolean][] = domain
    ? [
        ["Ownership", domain.ownership],
        ["DKIM", domain.dkimStatus === "SUCCESS"],
        ["Bounce domain", domain.mailFromStatus === "SUCCESS"],
        ["DMARC", domain.dmarcStatus === "VALID"],
      ]
    : [];
  const passed = checks.filter(([, ok]) => ok).length;
  const main = list.find((s) => s.isDefault) || list[0];

  const tone = verified
    ? {
        card: "border-emerald-200/80 from-emerald-50 via-emerald-50/60 to-card dark:border-emerald-900/70 dark:from-emerald-950/80 dark:via-emerald-950/40",
        tile: "bg-emerald-600 text-white shadow-emerald-600/30",
        glow: "bg-emerald-400/25",
        dots: "text-emerald-600/25",
        kicker: "text-emerald-700 dark:text-emerald-300",
      }
    : {
        card: "border-amber-200/80 from-amber-50 via-amber-50/60 to-card dark:border-amber-900/70 dark:from-amber-950/80 dark:via-amber-950/40",
        tile: "bg-amber-500 text-white shadow-amber-500/30",
        glow: "bg-amber-400/25",
        dots: "text-amber-600/25",
        kicker: "text-amber-700 dark:text-amber-300",
      };
  const Icon = verified ? ShieldCheck : ShieldAlert;

  return (
    <section
      role="status"
      aria-label={verified ? "Sending verified" : "Sending not verified"}
      className={`relative overflow-hidden rounded-2xl border bg-gradient-to-r p-5 sm:p-6 ${tone.card}`}
    >
      <div aria-hidden className={`pointer-events-none absolute inset-y-0 right-0 w-1/2 ${tone.dots}`} style={dots} />
      <div aria-hidden className={`pointer-events-none absolute -left-10 -top-16 size-48 rounded-full blur-3xl ${tone.glow}`} />
      <div className="relative flex flex-col gap-5 md:flex-row md:items-center">
        <div className={`flex size-12 shrink-0 items-center justify-center rounded-xl shadow-lg ${tone.tile}`}>
          <Icon size={24} aria-hidden />
        </div>
        <div className="min-w-0 flex-1">
          <p className={`text-[11px] font-semibold uppercase tracking-[0.14em] ${tone.kicker}`}>
            {verified ? "Sending verified" : domain ? `${passed} of ${checks.length} checks passed` : "Sending not set up"}
          </p>
          <h2 className="mt-1 text-lg font-semibold tracking-tight text-foreground">
            {verified
              ? domain?.ready
                ? `${domain.name} is ready to send`
                : "Your workspace is ready to send"
              : domain
                ? `Almost there with ${domain.name}`
                : "Verify a domain to start sending"}
          </h2>
          <p className="mt-1 max-w-2xl text-sm leading-6 text-muted-foreground">
            {verified
              ? `Every email goes out from ${main.fromEmail}${list.length > 1 ? ` (or ${list.length - 1} other sender${list.length > 2 ? "s" : ""})` : ""}, signed and authenticated so it lands in the inbox.`
              : !isAdmin
                ? "Emails won’t go out until a workspace admin verifies a sending domain."
                : domain
                  ? "Emails won’t go out until every check passes. DNS changes can take a few minutes to show up."
                  : "Emails won’t go out until you verify a sending domain or connect your own SMTP."}
          </p>
          {checks.length > 0 && (!verified || domain?.ready) && (
            <ol className="mt-4 flex flex-wrap items-center gap-y-2" aria-label="DNS checks">
              {checks.map(([label, ok], i) => (
                <li key={label} className="flex items-center">
                  {i > 0 && (
                    <span aria-hidden className={`mx-2 h-px w-6 sm:w-10 ${ok ? "bg-emerald-500/60" : "bg-foreground/15"}`} />
                  )}
                  <span
                    className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium ring-1 ${ok ? "bg-emerald-500/10 text-emerald-800 ring-emerald-500/30 dark:text-emerald-200" : "bg-background/70 text-muted-foreground ring-foreground/10"}`}
                  >
                    {ok ? (
                      <Check size={12} strokeWidth={3} aria-hidden />
                    ) : (
                      <span className="size-1.5 rounded-full bg-amber-500" />
                    )}
                    {label}
                  </span>
                </li>
              ))}
            </ol>
          )}
        </div>
        {isAdmin && !verified && (
          <div className="flex shrink-0 flex-wrap gap-2">
            <Button asChild>
              <Link href="/onboarding">
                {domain ? "Finish DNS setup" : "Verify a domain"}
                <ArrowRight size={16} className="ml-1" />
              </Link>
            </Button>
            {!domain && (
              <Button asChild variant="outline" className="bg-background/70">
                <Link href="/settings/smtp">Use my SMTP</Link>
              </Button>
            )}
          </div>
        )}
        {isAdmin && verified && (
          <Button asChild variant="outline" className="shrink-0 bg-background/70">
            <Link href="/onboarding">
              Send a test
              <ArrowRight size={16} className="ml-1" />
            </Link>
          </Button>
        )}
      </div>
    </section>
  );
}
