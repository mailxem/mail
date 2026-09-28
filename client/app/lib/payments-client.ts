/**
 * Payments Service Client
 * Client for interacting with the Diagonal Payments microservice
 */

import {
    CHATS_FEATURE,
    DATA_SOURCES_FEATURE,
    MEMBERS_FEATURE,
    QUERIES_FEATURE,
  } from "@/app/lib/constants";
  
  export interface PlanFeature {
    id: string;
    plan_id: string;
    quota: number;
    feature: string;
    created_at: string;
    updated_at: string;
    trail: boolean;
  }
  
  export interface Plan {
    id: string;
    dodo_plan_id?: string;
    name: string;
    description?: string;
    status: string;
    price: string;
    currency: string;
    price_per_seat: string;
    billing_cycle: string;
    trial_period_days: number;
    trial_enabled: boolean;
    display_order: number;
    is_popular: boolean;
    is_enterprise: boolean;
    features: PlanFeature[];
    trail_features: PlanFeature[];
    metadata?: any;
    created_at: string;
    updated_at: string;
  }
  
  export interface Subscription {
    id: string;
    team_id: string;
    plan_id: string;
    status: string;
    dodo_customer_id?: string;
    dodo_subscription_id?: string;
    price_per_seat: string;
    total_seats: number;
    billing_cycle: string;
    start_date: string;
    end_date?: string;
    trial_ends_at?: string;
    next_billing_date?: string;
    created_at: string;
    updated_at: string;
    plan: Plan;
  }
  
  export interface UsageHistoryItem {
    subscription_id: string;
    period: string;
    feature: string;
    usage: number;
    created_at: string;
  }
  
  export interface UsageStats {
    feature_usage: {
      [QUERIES_FEATURE]: number;
      [DATA_SOURCES_FEATURE]: number;
      [CHATS_FEATURE]: number;
      [MEMBERS_FEATURE]: number;
    };
    current_period: string;
    usage_history: Array<UsageHistoryItem>;
  }
  
  export interface CheckoutSession {
    checkout_url: string;
    session_id: string;
  }
  
  /** Returned when switching to paid plan from free/inactive requires checkout */
  export interface CheckoutRequired {
    checkout_required: {
      team_id: string;
      plan_id: string;
      seats: number;
      message: string;
    };
  }
  
  export interface Payment {
    id: string;
    subscription_id: string;
    dodo_payment_id: string;
    amount: number;
    currency: string;
    status: string;
    description?: string;
    processed_at?: string;
    created_at: string;
    updated_at: string;
  }
  
  export interface BillingDetails {
    id?: string;
    street: string;
    city: string;
    state: string;
    zip: string;
    country: string;
    team_id: string;
  }
  
  export class PaymentsClient {
    private baseUrl: string;
  
    constructor(baseUrl: string = "/api/billing") {
      this.baseUrl = baseUrl;
    }
  
    private async request<T>(
      endpoint: string,
      options: RequestInit = {}
    ): Promise<T> {
      const url = `${this.baseUrl}${endpoint}`;
      const response = await fetch(url, {
        headers: {
          "Content-Type": "application/json",
          ...options.headers,
        },
        ...options,
      });
  
      if (!response.ok) {
        const error = await response
          .json()
          .catch(() => ({ error: "Unknown error" }));
        throw new Error(error.error || `HTTP ${response.status}`);
      }
  
      return response.json();
    }
  
    // Plans
    async getPlans(): Promise<{ plans: Plan[] }> {
      return this.request("/plans");
    }
  
    // Subscriptions
    async createSubscription(data: {
      team_id: string;
      plan_id: string;
      seats: number;
    }): Promise<Subscription> {
      return this.request("/subscriptions", {
        method: "POST",
        body: JSON.stringify(data),
      });
    }
  
    async getSubscription(teamId: string): Promise<Subscription> {
      return this.request(`/subscriptions/team/${teamId}`);
    }
  
    async updateSubscription(
      subscriptionId: string,
      data: { plan_id?: string; seats?: number }
    ): Promise<Subscription | CheckoutRequired> {
      const url = `${this.baseUrl}/subscriptions/${subscriptionId}`;
      const response = await fetch(url, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data),
      });
  
      const json = await response.json().catch(() => ({}));
  
      if (response.status === 202 && json.checkout_required) {
        return json as CheckoutRequired;
      }
  
      if (!response.ok) {
        throw new Error(json.error || `HTTP ${response.status}`);
      }
  
      return json as Subscription;
    }
  
    async cancelSubscription(
      subscriptionId: string
    ): Promise<{ message: string }> {
      return this.request(`/subscriptions/${subscriptionId}`, {
        method: "DELETE",
      });
    }
  
    async getUsageStats(subscriptionId: string): Promise<UsageStats> {
      return this.request(`/subscriptions/${subscriptionId}/usage`);
    }
  
    async getPaymentHistory(
      subscriptionId: string
    ): Promise<{ payments: Payment[] }> {
      return this.request(`/subscriptions/${subscriptionId}/payments`);
    }
  
    // Usage tracking
    async checkUsageLimit(
      teamId: string,
      feature: string
    ): Promise<{ can_query: boolean }> {
      return this.request(`/usage/check/${teamId}/${feature}`);
    }
  
    async recordUsage(
      teamId: string,
      queries: number
    ): Promise<{ message: string }> {
      return this.request(`/usage/record/${teamId}`, {
        method: "POST",
        body: JSON.stringify({ queries }),
      });
    }
  
    // Payments
    async createCheckoutSession(data: {
      team_id: string;
      plan_id: string;
      seats: number;
      success_url: string;
      cancel_url: string;
      customer_name: string;
      customer_email: string;
      billing_details?: BillingDetails;
    }): Promise<CheckoutSession> {
      return this.request("/payments/checkout", {
        method: "POST",
        body: JSON.stringify(data),
      });
    }
  
    // Health check
    async healthCheck(): Promise<{
      status: string;
      service: string;
      timestamp: string;
    }> {
      return this.request("/health");
    }
  
    // Customer portal
    async getCustomerPortalSession(
      subscriptionID: string
    ): Promise<{ session: string }> {
      return this.request(`/subscriptions/${subscriptionID}/session`);
    }
  
    // Billing details
    async createBillingDetails(
      teamId: string,
      billingDetails: BillingDetails
    ): Promise<BillingDetails> {
      return this.request(`/billing/details`, {
        method: "POST",
        body: JSON.stringify({ ...billingDetails, team_id: teamId }),
      });
    }
  
    async getBillingDetails(teamId: string): Promise<BillingDetails> {
      return this.request(`/billing/details/${teamId}`);
    }
  
    async updateBillingDetails(
      id: string,
      billingDetails: BillingDetails
    ): Promise<BillingDetails> {
      return this.request(`/billing/details/${id}`, {
        method: "PUT",
        body: JSON.stringify({ ...billingDetails }),
      });
    }
  }
  
  // Hooks for React components
  export function usePaymentsClient() {
    const paymentsClient = new PaymentsClient();
    return paymentsClient;
  }
  
  // Utility functions
  export function formatPrice(amount: number, currency: string = "USD"): string {
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency,
    }).format(amount);
  }
  
  export function getPlanDisplayName(planType: string): string {
    switch (planType) {
      case "starter":
        return "Starter";
      case "scale":
        return "SyneHQ Scale";
      case "enterprise":
        return "Enterprise";
      default:
        return planType;
    }
  }
  
  export function getStatusDisplayName(status: string): string {
    switch (status) {
      case "active":
        return "Active";
      case "trial":
        return "Trial";
      case "past_due":
        return "Past Due";
      case "cancelled":
        return "Cancelled";
      case "suspended":
        return "Suspended";
      default:
        return status;
    }
  }
  
  export function isSubscriptionActive(subscription: Subscription): boolean {
    return ["active", "trial"].includes(subscription.status);
  }
  
  export function getUsagePercentage(used: number, limit?: number): number {
    if (!limit) return 0;
    return Math.min((used / limit) * 100, 100);
  }
  
  export function isUsageLimitReached(used: number, limit?: number): boolean {
    if (!limit) return false;
    return used >= limit;
  }