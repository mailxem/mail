import { Button } from "@/components/ui/button";
import { CreditCard, Loader2 } from "lucide-react";
import { ArrowLeftRight } from "lucide-react";
import { FREE_PLAN_ID } from "@/app/lib/constants";
import {
  useBillingData,
  useCustomerPortalSession,
} from "@/app/lib/store/billing";
import { useTeam } from "@/app/providers/team-provider";
import { useRouter } from "next/navigation";

export function BillingHeader() {
  const { team } = useTeam();
  const teamId = team?.id || "";

  const { subscription } = useBillingData(teamId);
  const customerPortal = useCustomerPortalSession();
  const router = useRouter();

  const setShowComparison = () => {
    router.push(`/settings/billing?compare=true`);
  };

  const handleManageBilling = () => {
    if (subscription.data && subscription.data.plan.id !== FREE_PLAN_ID) {
      customerPortal.mutate(subscription.data.id);
    }
  };

  return (
    <div className="flex justify-end gap-2">
      <Button variant="outline" onClick={setShowComparison}>
        <ArrowLeftRight className="h-4 w-4 mr-2" />
        Compare Plans
      </Button>
      {subscription.data && subscription.data.plan.id !== FREE_PLAN_ID && (
        <Button
          variant="outline"
          onClick={handleManageBilling}
          disabled={customerPortal.isPending}
        >
          {customerPortal.isPending && (
            <Loader2 className="h-4 w-4 mr-2 animate-spin" />
          )}
          <CreditCard className="h-4 w-4 mr-2" />
          Manage Billing
        </Button>
      )}
    </div>
  );
}
