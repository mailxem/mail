"use client";

import { Plan } from "@/app/lib/payments-client";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { Check, X } from "lucide-react";
import { UNLIMITED_QUOTA } from "@/app/lib/constants";

interface PlanComparisonProps {
  plans: Plan[];
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function PlanComparison({
  plans,
  open,
  onOpenChange,
}: PlanComparisonProps) {
  const formatPrice = (price: string) => {
    const numPrice = parseFloat(price);
    return numPrice === 0 ? "Free" : `$${numPrice.toFixed(2)}`;
  };

  // Get all unique features across all plans
  const allFeatures = Array.from(
    new Set(
      plans.flatMap((plan) => plan.features?.map((f) => f.feature) || [])
    )
  );

  const getFeatureQuota = (plan: Plan, feature: string) => {
    const planFeature = plan.features?.find((f) => f.feature === feature);
    if (!planFeature) return null;
    return planFeature.quota === UNLIMITED_QUOTA
      ? "Unlimited"
      : planFeature.quota;
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-6xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Compare Plans</DialogTitle>
          <DialogDescription>
            Compare features across all available plans
          </DialogDescription>
        </DialogHeader>

        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b">
                <th className="text-left p-4 font-semibold">Feature</th>
                {plans.map((plan) => (
                  <th key={plan.id} className="p-4 text-center min-w-[150px]">
                    <div className="space-y-2">
                      <div className="font-semibold">{plan.name}</div>
                      <div className="text-sm font-normal text-muted-foreground">
                        {formatPrice(plan.price)}
                        {plan.price !== "0" && (
                          <span className="text-xs">
                            /{plan.billing_cycle === "monthly" ? "mo" : "yr"}
                          </span>
                        )}
                      </div>
                      {plan.is_popular && (
                        <Badge className="text-xs">Popular</Badge>
                      )}
                    </div>
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              <tr className="border-b bg-muted/50">
                <td className="p-4 font-medium">Price</td>
                {plans.map((plan) => (
                  <td key={plan.id} className="p-4 text-center">
                    <div className="font-semibold">
                      {formatPrice(plan.price)}
                    </div>
                    <div className="text-xs text-muted-foreground">
                      per {plan.billing_cycle === "monthly" ? "month" : "year"}
                    </div>
                  </td>
                ))}
              </tr>

              {plans.some((p) => p.trial_enabled) && (
                <tr className="border-b">
                  <td className="p-4 font-medium">Trial Period</td>
                  {plans.map((plan) => (
                    <td key={plan.id} className="p-4 text-center">
                      {plan.trial_enabled ? (
                        <div className="flex flex-col items-center gap-1">
                          <Check className="h-5 w-5 text-green-500" />
                          <span className="text-xs text-muted-foreground">
                            {plan.trial_period_days} days
                          </span>
                        </div>
                      ) : (
                        <X className="h-5 w-5 text-muted-foreground mx-auto" />
                      )}
                    </td>
                  ))}
                </tr>
              )}

              <tr className="bg-muted/30">
                <td className="p-4 font-semibold" colSpan={plans.length + 1}>
                  Features
                </td>
              </tr>

              {allFeatures.map((feature) => (
                <tr key={feature} className="border-b">
                  <td className="p-4 capitalize">
                    {feature.replace(/_/g, " ")}
                  </td>
                  {plans.map((plan) => {
                    const quota = getFeatureQuota(plan, feature);
                    return (
                      <td key={plan.id} className="p-4 text-center">
                        {quota !== null ? (
                          <div className="flex flex-col items-center gap-1">
                            <Check className="h-5 w-5 text-green-500" />
                            <span className="text-xs text-muted-foreground">
                              {quota}
                            </span>
                          </div>
                        ) : (
                          <X className="h-5 w-5 text-muted-foreground mx-auto" />
                        )}
                      </td>
                    );
                  })}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </DialogContent>
    </Dialog>
  );
}
