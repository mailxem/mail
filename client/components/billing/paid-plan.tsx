"use client";

import Link from "next/link";
import { useTeam } from "@/app/providers/team-provider";
import { useSubscription } from "@/app/lib/store/billing";
import { FREE_PLAN_ID } from "@/app/lib/constants";
import { Button } from "@/components/ui/button";

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
