"use client";

import { Subscription } from "@/app/lib/payments-client";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { AlertCircle, CheckCircle, Clock, XCircle, CreditCard } from "lucide-react";
import { format } from "date-fns";
import {
  SUBSCRIPTION_STATUS_ACTIVE,
  SUBSCRIPTION_STATUS_CANCELLED,
  SUBSCRIPTION_STATUS_EXPIRED,
  SUBSCRIPTION_STATUS_FAILED,
  SUBSCRIPTION_STATUS_ON_HOLD,
  SUBSCRIPTION_STATUS_PAUSED,
  SUBSCRIPTION_STATUS_PENDING,
} from "@/app/lib/constants";

interface SubscriptionStatusProps {
  subscription: Subscription | null;
  onManageBilling?: () => void;
  isLoading?: boolean;
}

export function SubscriptionStatus({
  subscription,
  onManageBilling,
  isLoading = false,
}: SubscriptionStatusProps) {
  if (!subscription) {
    return (
      <Card className="p-6 border-dashed">
        <div className="flex items-center gap-4">
          <div className="flex items-center justify-center w-12 h-12 rounded-full bg-muted">
            <CreditCard className="h-6 w-6 text-muted-foreground" />
          </div>
          <div className="flex-1">
            <h3 className="font-semibold text-lg">No Active Subscription</h3>
            <p className="text-sm text-muted-foreground">
              Choose a plan below to get started
            </p>
          </div>
        </div>
      </Card>
    );
  }

  const getStatusConfig = (status: string) => {
    switch (status) {
      case SUBSCRIPTION_STATUS_ACTIVE:
        return {
          icon: CheckCircle,
          color: "text-green-500",
          bgColor: "bg-green-500/10",
          badge: "success",
          label: "Active",
        };
      case SUBSCRIPTION_STATUS_PENDING:
        return {
          icon: Clock,
          color: "text-yellow-500",
          bgColor: "bg-yellow-500/10",
          badge: "warning",
          label: "Pending",
        };
      case SUBSCRIPTION_STATUS_CANCELLED:
        return {
          icon: XCircle,
          color: "text-orange-500",
          bgColor: "bg-orange-500/10",
          badge: "secondary",
          label: "Cancelled",
        };
      case SUBSCRIPTION_STATUS_EXPIRED:
      case SUBSCRIPTION_STATUS_FAILED:
        return {
          icon: AlertCircle,
          color: "text-red-500",
          bgColor: "bg-red-500/10",
          badge: "destructive",
          label: "Expired",
        };
      case SUBSCRIPTION_STATUS_ON_HOLD:
        return {
          icon: AlertCircle,
          color: "text-orange-500",
          bgColor: "bg-orange-500/10",
          badge: "warning",
          label: "Past Due",
        };
      case SUBSCRIPTION_STATUS_PAUSED:
        return {
          icon: Clock,
          color: "text-blue-500",
          bgColor: "bg-blue-500/10",
          badge: "secondary",
          label: "Paused",
        };
      default:
        return {
          icon: AlertCircle,
          color: "text-gray-500",
          bgColor: "bg-gray-500/10",
          badge: "secondary",
          label: status,
        };
    }
  };

  const statusConfig = getStatusConfig(subscription.status);
  const StatusIcon = statusConfig.icon;

  return (
    <Card className="p-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <div
            className={`flex items-center justify-center w-12 h-12 rounded-full ${statusConfig.bgColor}`}
          >
            <StatusIcon className={`h-6 w-6 ${statusConfig.color}`} />
          </div>
          <div>
            <div className="flex items-center gap-2 mb-1">
              <h3 className="font-semibold text-lg">{subscription.plan.name}</h3>
              <Badge variant={statusConfig.badge as any}>
                {statusConfig.label}
              </Badge>
            </div>
            <div className="flex items-center gap-4 text-sm text-muted-foreground">
              {subscription.total_seats && (
                <span>{subscription.total_seats} seats</span>
              )}
              {subscription.next_billing_date && (
                <span>
                  Renews{" "}
                  {format(new Date(subscription.next_billing_date), "MMM d, yyyy")}
                </span>
              )}
              {subscription.trial_ends_at && (
                <span>
                  Trial ends{" "}
                  {format(new Date(subscription.trial_ends_at), "MMM d, yyyy")}
                </span>
              )}
            </div>
          </div>
        </div>
        {onManageBilling && (
          <Button
            variant="outline"
            onClick={onManageBilling}
            disabled={isLoading}
          >
            Manage Billing
          </Button>
        )}
      </div>
    </Card>
  );
}
