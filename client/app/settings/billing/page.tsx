"use client";

import { workspaceClassName } from "@/lib/workspace-styles";
import { useState } from "react";
import { PricingCard } from "@/components/billing/pricing-card";
import { SubscriptionStatus } from "@/components/billing/subscription-status";
import { UsageDisplay } from "@/components/billing/usage-display";
import { PlanComparison } from "@/components/billing/plan-comparison";
import { BillingCycleToggle } from "@/components/billing/billing-cycle-toggle";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import {
  usePlans,
  useBillingData,
  useUpdateSubscription,
  useCreateCheckoutSession,
  useCancelSubscription,
  useCustomerPortalSession,
  isCheckoutRequired,
} from "@/app/lib/store/billing";
import { Plan } from "@/app/lib/payments-client";
import { useTeam } from "@/app/providers/team-provider";
import { useSession } from "next-auth/react";
import {
  Loader2,
  CreditCard,
  History,
  BarChart3,
} from "lucide-react";
import { toast } from "sonner";
import { format } from "date-fns";
import { FREE_PLAN_ID } from "@/app/lib/constants";
import { useRouter, useSearchParams } from "next/navigation";

export default function BillingPage() {
  const { team } = useTeam();
  const { data: session } = useSession();
  const teamId = team?.id || "";
  const router = useRouter();
  const { data: plans, isLoading: plansLoading } = usePlans();
  const { subscription, usageStats, paymentHistory, billingDetails } =
    useBillingData(teamId);
  const updateSubscription = useUpdateSubscription();
  const createCheckout = useCreateCheckoutSession();
  const cancelSubscription = useCancelSubscription();
  const customerPortal = useCustomerPortalSession();

  const [selectedPlan, setSelectedPlan] = useState<Plan | null>(null);
  const [showConfirmDialog, setShowConfirmDialog] = useState(false);
  const [showCancelDialog, setShowCancelDialog] = useState(false);
  const [billingCycle, setBillingCycle] = useState<"monthly" | "yearly">(
    "monthly",
  );

  const searchParams = useSearchParams();
  const showComparison = searchParams.get("compare") === "true";

  const handleSelectPlan = (plan: Plan) => {
    setSelectedPlan(plan);
    setShowConfirmDialog(true);
  };

  const handleConfirmPlanChange = async () => {
    if (!selectedPlan || !teamId) return;

    try {
      // If user has no subscription, create checkout session
      if (!subscription.data) {
        await createCheckout.mutateAsync({
          team_id: teamId,
          plan_id: selectedPlan.id,
          seats: 1,
          success_url: `${window.location.origin}/settings/billing?success=true`,
          cancel_url: `${window.location.origin}/settings/billing?canceled=true`,
          customer_name: session?.user?.name || "",
          customer_email: session?.user?.email || "",
          billing_details: billingDetails.data || {
            team_id: teamId,
            street: "",
            city: "",
            state: "",
            zip: "",
            country: "",
          },
        });
        return;
      }

      // Update existing subscription
      const result = await updateSubscription.mutateAsync({
        subscriptionId: subscription.data.id,
        teamId,
        updates: { plan_id: selectedPlan.id },
      });

      // Check if checkout is required (switching to paid from free)
      if (isCheckoutRequired(result)) {
        await createCheckout.mutateAsync({
          team_id: teamId,
          plan_id: selectedPlan.id,
          seats: subscription.data.total_seats || 1,
          success_url: `${window.location.origin}/settings/billing?success=true`,
          cancel_url: `${window.location.origin}/settings/billing?canceled=true`,
          customer_name: session?.user?.name || "",
          customer_email: session?.user?.email || "",
          billing_details: billingDetails.data || {
            team_id: teamId,
            street: "",
            city: "",
            state: "",
            zip: "",
            country: "",
          },
        });
      }

      setShowConfirmDialog(false);
      setSelectedPlan(null);
    } catch (error: any) {
      console.error("Failed to change plan:", error);
      toast.error(error.message || "Failed to change plan");
    }
  };

  const handleCancelSubscription = async () => {
    if (!subscription.data) return;

    try {
      await cancelSubscription.mutateAsync({
        subscriptionId: subscription.data.id,
        teamId,
      });
      setShowCancelDialog(false);
    } catch (error: any) {
      console.error("Failed to cancel subscription:", error);
      toast.error(error.message || "Failed to cancel subscription");
    }
  };

  const handleManageBilling = () => {
    if (subscription.data?.id) {
      customerPortal.mutate(subscription.data.id);
    }
  };

  const formatCurrency = (
    amount: number | string,
    currency: string = "USD",
  ) => {
    const numAmount = typeof amount === "string" ? parseFloat(amount) : amount;
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency,
    }).format(numAmount);
  };

  if (plansLoading || !plans) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  // Debug: Log plans to console
  console.log("All plans:", plans);
  console.log("Current billing cycle:", billingCycle);

  // Filter plans by billing cycle
  // Show free plans in both views, and filter paid plans by cycle
  const filteredPlans = plans.filter((plan) => {
    const isFree = parseFloat(plan.price) === 0;
    const planCycle = plan.billing_cycle?.toLowerCase();
    const selectedCycle = billingCycle.toLowerCase();

    // Always show free plans
    if (isFree) return selectedCycle === "monthly";

    // For paid plans, match billing cycle (handle various formats)
    return (
      planCycle === selectedCycle ||
      planCycle?.includes(selectedCycle) ||
      (selectedCycle === "monthly" && planCycle?.includes("month")) ||
      (selectedCycle === "yearly" &&
        (planCycle?.includes("year") || planCycle?.includes("annual")))
    );
  });

  // Debug: Log filtered plans
  console.log("Filtered plans:", filteredPlans);

  // If no plans match the filter, show all plans (fallback)
  const displayPlans = filteredPlans.length > 0 ? filteredPlans : plans;

  // Calculate yearly discount (if applicable)
  const calculateYearlyDiscount = () => {
    const monthlyPlan = plans.find(
      (p) => p.billing_cycle === "monthly" && p.price !== "0",
    );
    const yearlyPlan = plans.find(
      (p) => p.billing_cycle === "yearly" && p.price !== "0",
    );

    if (monthlyPlan && yearlyPlan) {
      const monthlyAnnual = parseFloat(monthlyPlan.price) * 12;
      const yearly = parseFloat(yearlyPlan.price);
      const discount = ((monthlyAnnual - yearly) / monthlyAnnual) * 100;
      return Math.round(discount);
    }
    return 20; // Default discount
  };

  return (
    <div className={workspaceClassName("workspace-page-body space-y-6")}>
      <Tabs defaultValue="plans" className="space-y-6">
        <TabsList>
          <TabsTrigger value="plans">
            <CreditCard className="h-4 w-4 mr-2" />
            Plans
          </TabsTrigger>
          <TabsTrigger value="usage">
            <BarChart3 className="h-4 w-4 mr-2" />
            Usage
          </TabsTrigger>
          <TabsTrigger value="history">
            <History className="h-4 w-4 mr-2" />
            Payment History
          </TabsTrigger>
        </TabsList>

        <TabsContent value="plans" className="space-y-6">
          <SubscriptionStatus
            subscription={subscription.data || null}
            onManageBilling={handleManageBilling}
            isLoading={customerPortal.isPending}
          />

          {subscription.data && subscription.data.plan.id !== FREE_PLAN_ID && (
            <div className="flex justify-end">
              <Button
                variant="destructive"
                onClick={() => setShowCancelDialog(true)}
              >
                Cancel Subscription
              </Button>
            </div>
          )}

          {/* Billing Cycle Toggle */}
          <BillingCycleToggle
            cycle={billingCycle}
            onCycleChange={setBillingCycle}
            yearlyDiscount={calculateYearlyDiscount()}
          />

          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {displayPlans.length > 0 ? (
              displayPlans.map((plan) => (
                <PricingCard
                  key={plan.id}
                  plan={plan}
                  currentPlanId={subscription.data?.plan_id}
                  onSelectPlan={handleSelectPlan}
                  isLoading={
                    updateSubscription.isPending || createCheckout.isPending
                  }
                  isPending={subscription.data?.status === "pending"}
                />
              ))
            ) : (
              <div className="col-span-full text-center py-12 text-muted-foreground">
                <p>No plans available for {billingCycle} billing.</p>
                <p className="text-sm mt-2">
                  Please try switching to a different billing cycle.
                </p>
              </div>
            )}
          </div>
        </TabsContent>

        <TabsContent value="usage" className="space-y-6">
          <UsageDisplay
            usageStats={usageStats.data}
            planFeatures={subscription.data?.plan.features}
          />

          {!usageStats.data && (
            <Card className="p-6">
              <div className="text-center text-muted-foreground">
                <BarChart3 className="h-12 w-12 mx-auto mb-4 opacity-50" />
                <p>No usage data available yet</p>
              </div>
            </Card>
          )}
        </TabsContent>

        <TabsContent value="history" className="space-y-6">
          <Card className="p-6">
            <h3 className="font-semibold text-lg mb-4">Payment History</h3>
            {paymentHistory.isLoading ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
              </div>
            ) : paymentHistory.data && paymentHistory.data.length > 0 ? (
              <div className="space-y-4">
                {paymentHistory.data.map((payment) => (
                  <div
                    key={payment.id}
                    className="flex items-center justify-between p-4 border rounded-lg"
                  >
                    <div>
                      <p className="font-medium">
                        {payment.description || "Payment"}
                      </p>
                      <p className="text-sm text-muted-foreground">
                        {payment.processed_at
                          ? format(
                              new Date(payment.processed_at),
                              "MMM d, yyyy",
                            )
                          : "Pending"}
                      </p>
                    </div>
                    <div className="text-right">
                      <p className="font-semibold">
                        {formatCurrency(payment.amount, payment.currency)}
                      </p>
                      <p className="text-sm text-muted-foreground capitalize">
                        {payment.status}
                      </p>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8 text-muted-foreground">
                <History className="h-12 w-12 mx-auto mb-4 opacity-50" />
                <p>No payment history available</p>
              </div>
            )}
          </Card>
        </TabsContent>
      </Tabs>

      {/* Plan Change Confirmation Dialog */}
      <AlertDialog open={showConfirmDialog} onOpenChange={setShowConfirmDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Confirm Plan Change</AlertDialogTitle>
            <AlertDialogDescription>
              {selectedPlan && (
                <>
                  You are about to change your plan to{" "}
                  <strong>{selectedPlan.name}</strong>.
                  {selectedPlan.price !== "0" && (
                    <>
                      {" "}
                      This will cost{" "}
                      <strong>
                        {formatCurrency(
                          selectedPlan.price,
                          selectedPlan.currency,
                        )}
                      </strong>{" "}
                      per {selectedPlan.billing_cycle}.
                    </>
                  )}
                </>
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleConfirmPlanChange}>
              {updateSubscription.isPending || createCheckout.isPending ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Processing...
                </>
              ) : (
                "Confirm"
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Cancel Subscription Dialog */}
      <AlertDialog open={showCancelDialog} onOpenChange={setShowCancelDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Cancel Subscription</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to cancel your subscription? You will lose
              access to premium features at the end of your billing period.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Keep Subscription</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleCancelSubscription}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {cancelSubscription.isPending ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Cancelling...
                </>
              ) : (
                "Cancel Subscription"
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Plan Comparison Dialog */}
      <PlanComparison
        plans={plans}
        open={showComparison}
        onOpenChange={(open) => router.push(`/settings/billing?compare=${open}`)}
      />
    </div>
  );
}
