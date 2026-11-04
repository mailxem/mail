"use client";

import { useEffect, useState } from "react";
import { Loader2 } from "lucide-react";
import { BillingDetails } from "@/app/lib/payments-client";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

const FIELDS = ["street", "city", "state", "zip", "country"] as const;

/** Checkout is refused unless every address field is filled. */
export function hasCompleteBillingDetails(details?: BillingDetails | null): details is BillingDetails {
  return !!details && FIELDS.every((field) => details[field]?.trim());
}

export function BillingDetailsDialog({
  open,
  onOpenChange,
  teamId,
  initial,
  saving,
  onSubmit,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  teamId: string;
  initial?: BillingDetails | null;
  saving: boolean;
  onSubmit: (details: BillingDetails) => void;
}) {
  const [details, setDetails] = useState<BillingDetails>({ team_id: teamId, street: "", city: "", state: "", zip: "", country: "" });

  useEffect(() => {
    if (open) setDetails({ team_id: teamId, street: "", city: "", state: "", zip: "", country: "", ...initial });
  }, [open, initial, teamId]);

  const set = (field: (typeof FIELDS)[number]) => (event: React.ChangeEvent<HTMLInputElement>) =>
    setDetails((prev) => ({ ...prev, [field]: event.target.value }));

  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault();
    onSubmit({ ...details, country: details.country.trim().toUpperCase() });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Billing details</DialogTitle>
          <DialogDescription>
            Add your billing address to continue to checkout. It appears on your invoices.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="billing-street">Street address</Label>
            <Input id="billing-street" value={details.street} onChange={set("street")} autoComplete="street-address" required />
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="billing-city">City</Label>
              <Input id="billing-city" value={details.city} onChange={set("city")} autoComplete="address-level2" required />
            </div>
            <div className="space-y-2">
              <Label htmlFor="billing-state">State or region</Label>
              <Input id="billing-state" value={details.state} onChange={set("state")} autoComplete="address-level1" required />
            </div>
            <div className="space-y-2">
              <Label htmlFor="billing-zip">Postal code</Label>
              <Input id="billing-zip" value={details.zip} onChange={set("zip")} autoComplete="postal-code" required />
            </div>
            <div className="space-y-2">
              <Label htmlFor="billing-country">Country code</Label>
              <Input
                id="billing-country"
                value={details.country}
                onChange={set("country")}
                autoComplete="country"
                placeholder="US"
                pattern="[A-Za-z]{2}"
                title="Two-letter country code, such as US, IN or GB"
                maxLength={2}
                required
              />
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={saving}>
              Cancel
            </Button>
            <Button type="submit" disabled={saving}>
              {saving && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
              Continue to checkout
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
