"use client";

import { Plan } from "@/app/lib/payments-client";
import { Button } from "@/components/ui/button";
import { Check } from "lucide-react";
import { cn } from "@/lib/utils";
import { workspaceClassName } from "@/lib/workspace-styles";
import { FREE_PLAN_ID } from "@/app/lib/constants";

interface PricingCardProps {
  plan: Plan;
  currentPlanId?: string;
  onSelectPlan: (plan: Plan) => void;
  isLoading?: boolean;
  isPending?: boolean;
}

export const isFreePlan = (plan: Plan) => parseFloat(plan.price) === 0;
export const isYearlyPlan = (plan: Plan) => /year|annual/i.test(plan.billing_cycle || "");

// ponytail: the allowance lives in the plan description ("20000 emails a month included, ...");
// read it from a plan field once payments exposes one.
export const includedEmails = (plan: Plan) => Number(plan.description?.match(/\d+/)?.[0] ?? 0);

const withCommas = (text: string) => text.replace(/\d{4,}/g, (n) => Number(n).toLocaleString("en-US"));

export function PricingCard({
  plan,
  currentPlanId,
  onSelectPlan,
  isLoading = false,
  isPending = false,
}: PricingCardProps) {
  const isCurrentPlan = currentPlanId === plan.id;
  const isFree = isFreePlan(plan);
  const price = parseFloat(plan.price);
  const included = (plan.description || "").split(/[,.]/)[0];
  const overage = plan.description?.match(/then (.+?)\.?$/i)?.[1];
  const points = [
    withCommas(included),
    overage ? `${overage} after that` : "Stops at the limit, no overage",
    isFree ? "Up to 100 emails a day" : "No daily limit",
    isFree ? "No card needed" : "7-day free trial",
    isFree ? "“Sent with Xem” link under each email" : "No Xem link in your emails",
  ];

  return (
    <article
      className={cn(
        workspaceClassName("form-card"),
        "flex h-full flex-col",
        isCurrentPlan && "border-foreground/40 hover:border-foreground/40",
      )}
    >
      <div className={workspaceClassName("form-card-top")}>
        <h3 className="!mb-0">{plan.name.replace(/\s*\(yearly\)/i, "")}</h3>
        {isCurrentPlan ? (
          <span className={workspaceClassName("status-pill status-active")}><i />Current</span>
        ) : plan.is_popular ? (
          <span className={workspaceClassName("status-pill")}>Most popular</span>
        ) : null}
      </div>

      <div className="flex items-baseline gap-1">
        <span className="text-[28px] font-medium tracking-tight tabular-nums">
          {price < 0 ? "Custom" : `$${price.toLocaleString("en-US")}`}
        </span>
        {price >= 0 && (
          <span className="text-xs text-muted-foreground">/{isYearlyPlan(plan) ? "year" : "month"}</span>
        )}
      </div>

      <ul className="mt-5 flex-1 space-y-2.5 text-[13px]">
        {points.map((point) => (
          <li key={point} className="flex items-start gap-2">
            <Check className="mt-0.5 size-3.5 shrink-0 text-muted-foreground" strokeWidth={1.75} />
            <span>{point}</span>
          </li>
        ))}
      </ul>

      <Button
        onClick={() => onSelectPlan(plan)}
        disabled={isLoading || isCurrentPlan || isPending}
        variant={plan.is_popular && !isCurrentPlan ? "default" : "outline"}
        className="mt-6 w-full"
      >
        {isCurrentPlan ? "Current plan" : isFree ? "Switch to Free" : currentPlanId && currentPlanId !== FREE_PLAN_ID ? "Switch plan" : "Start free trial"}
      </Button>
    </article>
  );
}
