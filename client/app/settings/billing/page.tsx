"use client";

import { workspaceClassName } from "@/lib/workspace-styles";
import { useState } from "react";
import { PricingCard, includedEmails, isFreePlan, isYearlyPlan } from "@/components/billing/pricing-card";
import { PlanComparison } from "@/components/billing/plan-comparison";
import { Button } from "@/components/ui/button";
import { Metric, FilterTabs, QueryState, Empty } from "@/components/marketing/shared";
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
import { Loader2, CreditCard, CircleCheck, Mail, CalendarDays, ArrowUpRight } from "lucide-react";
import { toast } from "sonner";
import { format } from "date-fns";
import { EMAILS_FEATURE, FREE_PLAN_ID } from "@/app/lib/constants";
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

  const [view, setView] = useState("Plans");

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

  const sub = subscription.data;
  const currentPlan = plans.find((p) => p.id === sub?.plan_id) || sub?.plan;
  const onPaidPlan = !!sub && sub.plan_id !== FREE_PLAN_ID;
  const displayPlans = plans.filter((plan) =>
    isFreePlan(plan) || isYearlyPlan(plan) === (billingCycle === "yearly"),
  );
  const included = currentPlan ? includedEmails(currentPlan) : 0;
  const emailsUsed = (usageStats.data?.feature_usage as Record<string, number> | undefined)?.[EMAILS_FEATURE] ?? 0;
  const renewal = sub?.trial_ends_at
    ? { label: "Trial ends", date: sub.trial_ends_at }
    : sub?.next_billing_date
      ? { label: "Renews", date: sub.next_billing_date }
      : null;

  return (
    <div className={workspaceClassName("workspace-page-body")}>
      <div className={workspaceClassName("metrics-grid")}>
        <Metric label="Plan" value={currentPlan?.name.replace(/^Xem /, "") || "Free"} icon={<CreditCard />} />
        <Metric label="Status" value={sub ? sub.status.charAt(0).toUpperCase() + sub.status.slice(1) : "Free"} icon={<CircleCheck />} />
        <Metric
          label="Emails this month"
          value={emailsUsed.toLocaleString("en-US")}
          detail={included ? `of ${included.toLocaleString("en-US")}${currentPlan && isYearlyPlan(currentPlan) ? " this year" : ""}` : undefined}
          icon={<Mail />}
        />
        <Metric
          label={renewal?.label || "Renews"}
          value={renewal ? format(new Date(renewal.date), "MMM d") : "—"}
          detail={renewal ? format(new Date(renewal.date), "yyyy") : undefined}
          icon={<CalendarDays />}
        />
      </div>

      <section className={workspaceClassName("product-panel")}>
        <div className={workspaceClassName("panel-toolbar")}>
          <FilterTabs items={["Plans", "Payment history"]} value={view} onChange={setView} />
          {onPaidPlan && (
            <div className="flex items-center gap-2">
              <Button variant="ghost" size="sm" className="text-muted-foreground" onClick={() => setShowCancelDialog(true)}>
                Cancel plan
              </Button>
              <Button variant="outline" size="sm" onClick={handleManageBilling} disabled={customerPortal.isPending}>
                Manage billing <ArrowUpRight className="ml-1 size-3.5" />
              </Button>
            </div>
          )}
        </div>

        {view === "Plans" ? (
          <>
            <div className="mb-5 flex flex-wrap items-center justify-between gap-3">
              <p className="text-xs text-muted-foreground">
                Paid plans start with a 7-day free trial. Extra emails are billed per 1,000.
              </p>
              <FilterTabs
                items={["Monthly", "Yearly · 2 months free"]}
                value={billingCycle === "monthly" ? "Monthly" : "Yearly · 2 months free"}
                onChange={(v) => setBillingCycle(v === "Monthly" ? "monthly" : "yearly")}
              />
            </div>
            <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
              {displayPlans.map((plan) => (
                <PricingCard
                  key={plan.id}
                  plan={plan}
                  currentPlanId={sub?.plan_id}
                  onSelectPlan={handleSelectPlan}
                  isLoading={updateSubscription.isPending || createCheckout.isPending}
                  isPending={sub?.status === "pending"}
                />
              ))}
            </div>
          </>
        ) : paymentHistory.isLoading || paymentHistory.error ? (
          <QueryState loading={paymentHistory.isLoading} error={paymentHistory.error} retry={() => void paymentHistory.refetch()} />
        ) : !paymentHistory.data?.length ? (
          <Empty title="No payments yet" description="Invoices show up here after your first charge." />
        ) : (
          <div className="divide-y divide-border rounded-xl border border-border">
            {paymentHistory.data.map((payment) => (
              <div key={payment.id} className="flex items-center justify-between gap-4 px-4 py-3 text-[13px]">
                <div>
                  <p className="font-medium">{payment.description || "Payment"}</p>
                  <p className="text-xs text-muted-foreground">
                    {payment.processed_at ? format(new Date(payment.processed_at), "MMM d, yyyy") : "Pending"}
                  </p>
                </div>
                <div className="flex items-center gap-3">
                  <span className={workspaceClassName("status-pill")}>{payment.status}</span>
                  <span className="tabular-nums font-medium">{formatCurrency(payment.amount, payment.currency)}</span>
                </div>
              </div>
            ))}
          </div>
        )}
      </section>

      {/* Plan Change Confirmation Dialog */}
      <AlertDialog open={showConfirmDialog} onOpenChange={setShowConfirmDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Change plan?</AlertDialogTitle>
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
                      per {isYearlyPlan(selectedPlan) ? "year" : "month"}.
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
