"use client";

import Link from "next/link";
import { useTeam } from "@/app/providers/team-provider";
import { useSubscription } from "@/app/lib/store/billing";
import { FREE_PLAN_ID } from "@/app/lib/constants";
import { Button } from "@/components/ui/button";
import { Check, Sparkles } from "lucide-react";
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";

/** False only once the workspace is known to be on Free; the API enforces it either way. */
export function usePaidPlan() {
  const { team } = useTeam();
  const { data, isLoading } = useSubscription(team?.id || "");
  if (isLoading) return true;
  return isPaidSubscription(data);
}

/** Same rule as the backend: a real, non-Free plan in good standing. */
export function isPaidSubscription(data?: { plan_id?: string; status?: string } | null) {
  return !!data?.plan_id && data.plan_id !== FREE_PLAN_ID && data.status === "active";
}

export function PaidPlanNotice({ feature }: { feature: string }) {
  return (
    <div className="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-dashed p-4">
      <div>
        <p className="text-sm font-medium">{feature} comes with a paid plan.</p>
        <p className="mt-1 text-xs text-muted-foreground">
          Starter and up include it, with a 7-day free trial.
        </p>
      </div>
      <Button asChild size="sm">
        <Link href="/settings/billing">See plans</Link>
      </Button>
    </div>
  );
}

const PAID_FEATURES = [
  { title: "No daily sending cap", body: "Free stops at 100 emails a day. Paid plans only count your monthly volume." },
  { title: "More email every month", body: "From 20,000 emails on Starter up to 500,000 on Scale, with low per-email overage." },
  { title: "The Xem inbox", body: "Receive replies on your own domain inside Xem, no separate mailbox needed." },
  { title: "Cloudflare email", body: "Send and receive through your Cloudflare account with guided setup." },
];

/** Sidebar Upgrade: what paid plans add, then off to billing to pick one. Hidden for paid workspaces. */
export function UpgradeSheet() {
  const paid = usePaidPlan();
  if (paid) return null;
  return (
    <Sheet>
      <SheetTrigger asChild>
        <button
          type="button"
          className="w-full rounded-xl border border-border bg-background p-3 text-left transition-colors hover:bg-muted"
        >
          <span className="flex items-center gap-2 text-sm font-medium">
            <Sparkles className="size-4" />
            Upgrade to Pro
          </span>
          <span className="mt-1 block text-xs text-muted-foreground">
            No daily cap, the Xem inbox and more. Free for 7 days.
          </span>
        </button>
      </SheetTrigger>
      <SheetContent className="overflow-y-auto">
        <SheetHeader>
          <SheetTitle>Do more with a paid plan</SheetTitle>
          <SheetDescription>
            You’re on Free. Every paid plan starts with a 7-day free trial and you can cancel any time.
          </SheetDescription>
        </SheetHeader>
        <ul className="mt-6 space-y-4 px-4">
          {PAID_FEATURES.map((f) => (
            <li key={f.title} className="flex gap-3">
              <Check className="mt-0.5 size-4 shrink-0 text-primary" />
              <div>
                <p className="text-sm font-medium">{f.title}</p>
                <p className="mt-0.5 text-xs text-muted-foreground">{f.body}</p>
              </div>
            </li>
          ))}
        </ul>
        <div className="mt-8 grid gap-2 px-4">
          <SheetClose asChild>
            <Button asChild>
              <Link href="/settings/billing">Start free trial</Link>
            </Button>
          </SheetClose>
          <SheetClose asChild>
            <Button asChild variant="ghost">
              <Link href="/settings/billing">Compare all plans</Link>
            </Button>
          </SheetClose>
        </div>
      </SheetContent>
    </Sheet>
  );
}
