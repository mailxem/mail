"use client";

import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";

interface BillingCycleToggleProps {
  cycle: "monthly" | "yearly";
  onCycleChange: (cycle: "monthly" | "yearly") => void;
  yearlyDiscount?: number;
}

export function BillingCycleToggle({
  cycle,
  onCycleChange,
  yearlyDiscount = 20,
}: BillingCycleToggleProps) {
  return (
    <div className="flex items-center justify-center gap-4 p-6 bg-muted/30 rounded-lg border">
      <div className="flex items-center gap-3">
        <Label
          htmlFor="billing-cycle"
          className={`text-sm font-medium cursor-pointer transition-colors ${
            cycle === "monthly" ? "text-foreground" : "text-muted-foreground"
          }`}
        >
          Monthly
        </Label>
        <Switch
          id="billing-cycle"
          checked={cycle === "yearly"}
          onCheckedChange={(checked) =>
            onCycleChange(checked ? "yearly" : "monthly")
          }
        />
        <div className="flex items-center gap-2">
          <Label
            htmlFor="billing-cycle"
            className={`text-sm font-medium cursor-pointer transition-colors ${
              cycle === "yearly" ? "text-foreground" : "text-muted-foreground"
            }`}
          >
            Yearly
          </Label>
          {cycle === "yearly" && yearlyDiscount > 0 && (
            <Badge variant="secondary" className="bg-green-500/10 text-green-600 border-green-500/20">
              Save {yearlyDiscount}%
            </Badge>
          )}
        </div>
      </div>
    </div>
  );
}
