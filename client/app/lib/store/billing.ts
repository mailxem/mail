/**
 * Billing Store using React Query
 * Provides efficient caching and state management for billing operations
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
    BillingDetails,
    CheckoutRequired,
    PaymentsClient,
    Subscription,
} from '@/app/lib/payments-client';
import { toast } from 'sonner';
import { useLocalStorage } from 'usehooks-ts';
import { FREE_PLAN_ID, QUERIES_FEATURE, SUBSCRIPTION_STATUS_ACTIVE, SUBSCRIPTION_STATUS_CANCELLED, SUBSCRIPTION_STATUS_EXPIRED, SUBSCRIPTION_STATUS_FAILED, SUBSCRIPTION_STATUS_ON_HOLD, SUBSCRIPTION_STATUS_PAUSED, SUBSCRIPTION_STATUS_PENDING, UNLIMITED_QUOTA } from '@/app/lib/constants';

// Query Keys Factory - Ensures consistent cache keys
export const billingKeys = {
    all: ['billing'] as const,
    plans: () => [...billingKeys.all, 'plans'] as const,
    subscriptions: () => [...billingKeys.all, 'subscriptions'] as const,
    subscription: (teamId: string) => [...billingKeys.subscriptions(), teamId] as const,
    usageStats: (subscriptionId: string) => [...billingKeys.all, 'usage', subscriptionId] as const,
    paymentHistory: (subscriptionId: string) => [...billingKeys.all, 'payments', subscriptionId] as const,
    usageCheck: (teamId: string, feature: string) => [...billingKeys.all, 'usage-check', teamId, feature] as const,
    health: () => [...billingKeys.all, 'health'] as const,
    billingDetails: (teamId: string) => [...billingKeys.all, 'billing-details', teamId] as const,
};

// Initialize payments client
const paymentsClient = new PaymentsClient();

// ============================================================================
// QUERY HOOKS
// ============================================================================

/**
 * Get all available plans
 */
export const usePlans = () => {
    return useQuery({
        queryKey: billingKeys.plans(),
        queryFn: async () => {
            const result = await paymentsClient.getPlans();
            return result.plans;
        },
        refetchOnWindowFocus: true,
    });
};

/**
 * Get subscription for a team
 */
export const useSubscription = (teamId: string) => {
    return useQuery({
        queryKey: billingKeys.subscription(teamId),
        queryFn: async () => {
            try {
                return await paymentsClient.getSubscription(teamId);
            } catch (error: any) {
                // If no subscription found, return null instead of throwing
                if (error.message?.includes('404') || error.message?.includes('not found')) {
                    return null;
                }
                throw error;
            }
        },
        enabled: !!teamId,
        staleTime: 15 * 60 * 1000, // 15 minutes
        gcTime: 30 * 60 * 1000, // 30 minutes
        refetchOnWindowFocus: true,
        retry: (failureCount, error: any) => {
            // Don't retry for 404s (no subscription found)
            if (error.message?.includes('404') || error.message?.includes('not found')) {
                return false;
            }
            return failureCount < 3;
        },
    });
};

/**
 * Get usage statistics for a subscription
 */
export const useUsageStats = (subscriptionId?: string) => {
    return useQuery({
        queryKey: billingKeys.usageStats(subscriptionId!),
        queryFn: () => paymentsClient.getUsageStats(subscriptionId!),
        enabled: !!subscriptionId,
        refetchOnWindowFocus: true,
        staleTime: 10 * 60 * 1000, // 10 minutes
        gcTime: 20 * 60 * 1000, // 20 minutes
    });
};

/**
 * Get payment history for a subscription
 */
export const usePaymentHistory = (subscriptionId?: string) => {
    return useQuery({
        queryKey: billingKeys.paymentHistory(subscriptionId!),
        queryFn: async () => {
            const result = await paymentsClient.getPaymentHistory(subscriptionId!);
            return result.payments;
        },
        enabled: !!subscriptionId,
        refetchOnWindowFocus: true,
        staleTime: 10 * 60 * 1000, // 10 minutes
        gcTime: 30 * 60 * 1000, // 30 minutes
    });
};

/**
 * Check usage limit for a team
 */
export const useUsageCheck = (teamId: string, feature: string) => {
    return useQuery({
        queryKey: billingKeys.usageCheck(teamId, feature),
        queryFn: () => paymentsClient.checkUsageLimit(teamId, feature),
        enabled: !!teamId,
        staleTime: 60 * 1000, // 1 minute - check frequently for limits
        gcTime: 5 * 60 * 1000, // 5 minutes
    });
};

/**
 * Health check for billing service
 */
export const useBillingHealth = () => {
    return useQuery({
        queryKey: billingKeys.health(),
        queryFn: () => paymentsClient.healthCheck(),
        refetchOnWindowFocus: true,
        staleTime: 10 * 60 * 1000, // 10 minutes
        gcTime: 20 * 60 * 1000, // 20 minutes
        retry: 3,
    });
};

// ============================================================================
// COMBINED HOOKS FOR CONVENIENCE
// ============================================================================

/**
 * Get all billing data for a team - combines subscription, usage, and payments
 */
export const useBillingData = (teamId: string) => {
    const subscription = useSubscription(teamId);
    const usageStats = useUsageStats(subscription.data?.id);
    const paymentHistory = usePaymentHistory(subscription.data?.id);
    const usageCheck = useUsageCheck(teamId, 'queries');
    const billingDetails = useGetBillingDetails(teamId);

    return {
        subscription,
        usageStats,
        paymentHistory,
        usageCheck,
        billingDetails,
        isLoading: subscription.isLoading ||
            (subscription.data && usageStats.isLoading) ||
            (subscription.data && paymentHistory.isLoading),
        error: subscription.error || usageStats.error || paymentHistory.error,
        hasSubscription: !!subscription.data,
    };
};

// ============================================================================
// MUTATION HOOKS
// ============================================================================

/**
 * Create a new subscription
 */
export const useCreateSubscription = () => {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (data: {
            team_id: string;
            plan_id: string;
            seats: number;
        }) => {
            return await paymentsClient.createSubscription(data);
        },
        onSuccess: (data, variables) => {
            // Invalidate and refetch subscription data
            queryClient.invalidateQueries({ queryKey: billingKeys.subscription(variables.team_id) });
            queryClient.invalidateQueries({ queryKey: billingKeys.usageCheck(variables.team_id, 'queries') });
            toast.success('Subscription created successfully');
        },
        onError: (error: Error) => {
            toast.error(`Failed to create subscription: ${error.message}`);
        },
    });
};

/** Type guard for checkout_required response from updateSubscription */
export function isCheckoutRequired(
    result: Subscription | CheckoutRequired
): result is CheckoutRequired {
    return 'checkout_required' in result && !!result.checkout_required;
}

/**
 * Update an existing subscription (plan change, seat change).
 * - Switching to free plan: cancels old subscription, creates new free one → 200 with Subscription.
 * - Switching to paid from free/inactive: deletes old, returns 202 with checkout_required.
 * Callers should check for checkout_required and redirect to checkout when needed.
 */
export const useUpdateSubscription = () => {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (data: {
            subscriptionId: string;
            teamId: string;
            updates: { plan_id?: string; seats?: number };
        }) => {
            return await paymentsClient.updateSubscription(data.subscriptionId, data.updates);
        },
        onSuccess: (result, variables) => {
            if (isCheckoutRequired(result)) {
                // Old subscription was removed; invalidate so UI refetches
                queryClient.invalidateQueries({ queryKey: billingKeys.subscription(variables.teamId) });
                queryClient.invalidateQueries({ queryKey: billingKeys.usageCheck(variables.teamId, 'queries') });
                // Caller handles checkout redirect
                return;
            }
            // Update the cache with new subscription
            queryClient.setQueryData(
                billingKeys.subscription(variables.teamId),
                result
            );
            queryClient.invalidateQueries({ queryKey: billingKeys.usageStats(result.id) });
            queryClient.invalidateQueries({ queryKey: billingKeys.usageCheck(variables.teamId, 'queries') });
            toast.success('Subscription updated successfully');
        },
        onError: (error: Error) => {
            toast.error(`Failed to update subscription: ${error.message}`);
        },
    });
};

/**
 * Cancel a subscription
 */
export const useCancelSubscription = () => {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (data: { subscriptionId: string; teamId: string }) => {
            return await paymentsClient.cancelSubscription(data.subscriptionId);
        },
        onSuccess: (result, variables) => {
            // Invalidate subscription data to refetch updated status
            queryClient.invalidateQueries({ queryKey: billingKeys.subscription(variables.teamId) });
            queryClient.invalidateQueries({ queryKey: billingKeys.usageCheck(variables.teamId, QUERIES_FEATURE) });
            toast.success('Subscription cancelled successfully');
        },
        onError: (error: Error) => {
            toast.error(`Failed to cancel subscription: ${error.message}`);
        },
    });
};

/**
 * Create billing details
 */
export const useCreateBillingDetails = () => {
    return useMutation({
        mutationFn: async (data: BillingDetails) => {
            return await paymentsClient.createBillingDetails(data.team_id, data);
        },
        onSuccess: (result, variables) => {
            toast.success('Billing details created successfully');
        },
        onError: (error: Error) => {
            toast.error(`Failed to create billing details: ${error.message}`);
        },
    });
};

/**
 * Get billing details
 */
export const useGetBillingDetails = (teamId: string) => {
    return useQuery({
        queryKey: billingKeys.billingDetails(teamId),
        queryFn: async () => {
            return await paymentsClient.getBillingDetails(teamId);
        },
        enabled: !!teamId,
    });
};

/**
 * Update billing details
 */
export const useUpdateBillingDetails = () => {
    return useMutation({
        mutationFn: async (data: BillingDetails) => {
            return await paymentsClient.updateBillingDetails(data.id!, data);
        },
        onSuccess: (result, variables) => {
            toast.success('Billing details updated successfully');
        },
        onError: (error: Error) => {
            toast.error(`Failed to update billing details: ${error.message}`);
        },
    });
};

/**
 * Create checkout session for new subscription
 */
export const useCreateCheckoutSession = () => {
    return useMutation({
        mutationFn: async (data: {
            team_id: string;
            plan_id: string;
            seats: number;
            success_url: string;
            cancel_url: string;
            customer_name: string;
            customer_email: string;
            billing_details: BillingDetails;
        }) => {
            return await paymentsClient.createCheckoutSession(data);
        },
        onSuccess: (checkoutSession) => {
            // Redirect to checkout
            window.location.href = checkoutSession.checkout_url;
        },
        onError: (error: Error) => {
            toast.error(`Failed to create checkout session: ${error.message}`);
        },
    });
};

/**
 * Record usage for a team
 */
export const useRecordUsage = () => {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (data: { teamId: string; queries: number }) => {
            return await paymentsClient.recordUsage(data.teamId, data.queries);
        },
        onSuccess: (result, variables) => {
            // Invalidate usage-related queries to get fresh data
            queryClient.invalidateQueries({ queryKey: billingKeys.usageCheck(variables.teamId, QUERIES_FEATURE) });
            // Find and invalidate usage stats for this team's subscription
            const subscriptionQuery = queryClient.getQueryData(billingKeys.subscription(variables.teamId)) as Subscription | null;
            if (subscriptionQuery?.id) {
                queryClient.invalidateQueries({ queryKey: billingKeys.usageStats(subscriptionQuery.id) });
            }
        },
        onError: (error: Error) => {
            console.error('Failed to record usage:', error);
            // Don't show error toast for usage recording as it might be too noisy
        },
    });
};

/**
 * Get customer portal session
 */
export const useCustomerPortalSession = () => {
    return useMutation({
        mutationFn: async (subscriptionId: string) => {
            return await paymentsClient.getCustomerPortalSession(subscriptionId);
        },
        onSuccess: (session) => {
            // Open customer portal in new tab
            window.open(session.session, '_blank');
        },
        onError: (error: Error) => {
            toast.error(`Failed to open customer portal: ${error.message}`);
        },
    });
};

// ============================================================================
// UTILITY HOOKS
// ============================================================================

/**
 * Hook to prefetch billing data - useful for warming cache
 */
export const usePrefetchBillingData = () => {
    const queryClient = useQueryClient();

    return {
        prefetchPlans: () => {
            queryClient.prefetchQuery({
                queryKey: billingKeys.plans(),
                queryFn: async () => {
                    const result = await paymentsClient.getPlans();
                    return result.plans;
                },
                staleTime: 30 * 60 * 1000,
            });
        },
        prefetchBillingDetails: (teamId: string) => {
            queryClient.prefetchQuery({
                queryKey: billingKeys.billingDetails(teamId),
                queryFn: async () => {
                    return await paymentsClient.getBillingDetails(teamId);
                },
            });
        },
        prefetchSubscription: (teamId: string) => {
            queryClient.prefetchQuery({
                queryKey: billingKeys.subscription(teamId),
                queryFn: async () => {
                    try {
                        return await paymentsClient.getSubscription(teamId);
                    } catch (error: any) {
                        if (error.message?.includes('404') || error.message?.includes('not found')) {
                            return null;
                        }
                        throw error;
                    }
                },
                staleTime: 5 * 60 * 1000,
            });
        },
    };
};

/**
 * Hook to clear billing cache - useful for forced refresh
 */
export const useClearBillingCache = () => {
    const queryClient = useQueryClient();

    return {
        clearAll: () => {
            queryClient.removeQueries({ queryKey: billingKeys.all });
        },
        clearTeamData: (teamId: string) => {
            queryClient.removeQueries({ queryKey: billingKeys.subscription(teamId) });
            queryClient.removeQueries({ queryKey: billingKeys.usageCheck(teamId, QUERIES_FEATURE) });
        },
        clearSubscriptionData: (subscriptionId: string) => {
            queryClient.removeQueries({ queryKey: billingKeys.usageStats(subscriptionId) });
            queryClient.removeQueries({ queryKey: billingKeys.paymentHistory(subscriptionId) });
        },
    };
};

// ============================================================================
// SELECTOR HOOKS FOR COMPUTED VALUES
// ============================================================================

/**
 * Get computed billing status for a team
 */
export const useBillingStatus = (teamId: string) => {
    const { subscription, usageStats, usageCheck, billingDetails } = useBillingData(teamId);
    const isActive = subscription.data?.status === SUBSCRIPTION_STATUS_ACTIVE || subscription.data?.status === SUBSCRIPTION_STATUS_PENDING || subscription.data?.status === "TRIAL";
    const isFreePlan = subscription.data?.plan?.id === FREE_PLAN_ID;
    const isTrial = subscription.data?.status === SUBSCRIPTION_STATUS_PENDING && subscription.data?.trial_ends_at && new Date(subscription.data.trial_ends_at) > new Date();
    const isPending = subscription.data?.status === SUBSCRIPTION_STATUS_PENDING && !isTrial && !isFreePlan;
    const isPastDue = subscription.data?.status === SUBSCRIPTION_STATUS_ON_HOLD;
    const isCancelled = subscription.data?.status === SUBSCRIPTION_STATUS_CANCELLED;
    const isSuspended = subscription.data?.status === SUBSCRIPTION_STATUS_PAUSED;
    const isInactive = subscription.data?.status === SUBSCRIPTION_STATUS_EXPIRED;
    const isFailed = subscription.data?.status === SUBSCRIPTION_STATUS_FAILED;
    const canQuery = usageCheck.data?.can_query ?? true;
    const [hideBanner, setHideBanner] = useLocalStorage(
        "hide-banner",
        false,
    );

    const hasUsageLimit = usageStats.data && subscription.data?.plan?.features?.some((f: any) => f.feature === QUERIES_FEATURE && f.quota !== UNLIMITED_QUOTA);

    return {
        hasSubscription: !!subscription.data,
        isActive,
        isTrial,
        isPastDue,
        isCancelled,
        isSuspended,
        isInactive,
        isFailed,
        canQuery,
        hasUsageLimit,
        isOverLimit: !canQuery && hasUsageLimit,
        displayStatus: subscription.data?.status || 'none',
        planName: subscription.data?.plan?.name || 'No Plan',
        isLoading: subscription.isLoading || usageCheck.isLoading,
        trialDaysLeft: subscription.data?.trial_ends_at ? Math.ceil((new Date(subscription.data.trial_ends_at).getTime() - Date.now()) / (1000 * 60 * 60 * 24)) : 0,
        hideBanner,
        setHideBanner,
        isFreePlan,
        isPending,
        billingDetails: billingDetails.data,
    };
};