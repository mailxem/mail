"use client";

import { Plan } from "@/app/lib/payments-client";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Check, Zap } from "lucide-react";
import { cn } from "@/lib/utils";

interface PricingCardProps {
  plan: Plan;
  currentPlanId?: string;
  onSelectPlan: (plan: Plan) => void;
  isLoading?: boolean;
  isPending?: boolean;
}

export function PricingCard({
  plan,
  currentPlanId,
  onSelectPlan,
  isLoading = false,
  isPending = false,
}: PricingCardProps) {
  const isCurrentPlan = currentPlanId === plan.id;
  const isFree = plan.price === "0" || plan.price === "0.00";
  const isEnterprise = plan.is_enterprise;

  const getIcon = () => {
    if (isEnterprise) return "🏢";
    if (plan.is_popular) return "🔥";
    if (isFree) return "⚡";
    return "🚀";
  };

  const formatPrice = (price: string) => {
    const numPrice = parseFloat(price);
    if (numPrice === -1) return "Contact us";
    return numPrice === 0 ? "0" : "$" + numPrice.toFixed(2);
  };

  return (
    <Card
      className={cn(
        "relative p-6 flex flex-col h-full border-2 transition-all hover:shadow-lg",
        plan.is_popular && "border-primary shadow-lg scale-105",
        isCurrentPlan && "border-green-500 bg-green-50/5",
      )}
    >
      {plan.is_popular && (
        <Badge className="absolute -top-3 left-1/2 -translate-x-1/2 bg-primary">
          Most popular
        </Badge>
      )}

      <div className="flex items-start gap-3 mb-4">
        <div className="text-3xl">{getIcon()}</div>
        <div className="flex-1">
          <h3 className="text-xl font-semibold">{plan.name}</h3>
          {plan.description && (
            <p className="text-sm text-muted-foreground mt-1">
              {plan.description}
            </p>
          )}
        </div>
      </div>

      <div className="mb-6">
        <div className="flex items-baseline gap-1">
          <span className="text-4xl font-bold">{formatPrice(plan.price)}</span>
          {!isEnterprise && (
            <span className="text-muted-foreground">
              /{plan.billing_cycle === "monthly" ? "month" : "year"}
            </span>
          )}
        </div>
        {isEnterprise && (
          <p className="text-sm text-muted-foreground mt-1">Custom Pricing</p>
        )}
      </div>

      <div className="flex-1 space-y-3 mb-6">
        {plan.features && plan.features.length > 0 ? (
          plan.features.map((feature, index) => (
            <div key={index} className="flex items-start gap-2">
              <Check className="h-5 w-5 text-primary shrink-0 mt-0.5" />
              <span className="text-sm">
                {feature.quota === -1
                  ? "Unlimited"
                  : feature.quota === 1
                    ? ""
                    : feature.quota}{" "}
                {feature.feature.replace(/_/g, " ")}
              </span>
            </div>
          ))
        ) : (
          <div className="flex items-start gap-2">
            <Check className="h-5 w-5 text-primary shrink-0 mt-0.5" />
            <span className="text-sm">All features included</span>
          </div>
        )}
      </div>

      <Button
        onClick={() => onSelectPlan(plan)}
        disabled={isLoading || isCurrentPlan || isPending}
        variant={plan.is_popular ? "default" : "outline"}
        className="w-full"
      >
        {isCurrentPlan
          ? "Current Plan"
          : isFree
            ? "Try for free"
            : isEnterprise
              ? "Get Started"
              : "Try for free"}
      </Button>

      {isCurrentPlan && (
        <p className="text-xs text-center text-muted-foreground mt-2">
          You're currently on this plan
        </p>
      )}
    </Card>
  );
}
