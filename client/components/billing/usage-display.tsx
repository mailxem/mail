"use client";

import { UsageStats } from "@/app/lib/payments-client";
import { Card } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Badge } from "@/components/ui/badge";
import {
  QUERIES_FEATURE,
  DATA_SOURCES_FEATURE,
  CHATS_FEATURE,
  MEMBERS_FEATURE,
  UNLIMITED_QUOTA,
} from "@/app/lib/constants";
import { TrendingUp, Database, MessageSquare, Users } from "lucide-react";

interface UsageDisplayProps {
  usageStats: UsageStats | undefined;
  planFeatures?: Array<{ feature: string; quota: number }>;
}

export function UsageDisplay({ usageStats, planFeatures }: UsageDisplayProps) {
  if (!usageStats || !planFeatures) {
    return null;
  }

  const getFeatureInfo = (feature: string) => {
    switch (feature) {
      case QUERIES_FEATURE:
        return {
          label: "Queries",
          icon: TrendingUp,
          color: "text-blue-500",
          bgColor: "bg-blue-500/10",
        };
      case DATA_SOURCES_FEATURE:
        return {
          label: "Data Sources",
          icon: Database,
          color: "text-green-500",
          bgColor: "bg-green-500/10",
        };
      case CHATS_FEATURE:
        return {
          label: "Chats",
          icon: MessageSquare,
          color: "text-purple-500",
          bgColor: "bg-purple-500/10",
        };
      case MEMBERS_FEATURE:
        return {
          label: "Team Members",
          icon: Users,
          color: "text-orange-500",
          bgColor: "bg-orange-500/10",
        };
      default:
        return {
          label: feature,
          icon: TrendingUp,
          color: "text-gray-500",
          bgColor: "bg-gray-500/10",
        };
    }
  };

  const getUsageValue = (feature: string): number => {
    return (usageStats.feature_usage as any)[feature] || 0;
  };

  const formatNumber = (num: number) => {
    return new Intl.NumberFormat().format(num);
  };

  return (
    <Card className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h3 className="font-semibold text-lg">Usage Statistics</h3>
        <Badge variant="outline">{usageStats.current_period}</Badge>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        {planFeatures.map((feature) => {
          const info = getFeatureInfo(feature.feature);
          const Icon = info.icon;
          const used = getUsageValue(feature.feature);
          const quota = feature.quota;
          const isUnlimited = quota === UNLIMITED_QUOTA;
          const percentage = isUnlimited ? 0 : (used / quota) * 100;

          return (
            <div key={feature.feature} className="space-y-3">
              <div className="flex items-center gap-3">
                <div
                  className={`flex items-center justify-center w-10 h-10 rounded-lg ${info.bgColor}`}
                >
                  <Icon className={`h-5 w-5 ${info.color}`} />
                </div>
                <div className="flex-1">
                  <p className="font-medium">{info.label}</p>
                  <p className="text-sm text-muted-foreground">
                    {formatNumber(used)} /{" "}
                    {isUnlimited ? "Unlimited" : formatNumber(quota)}
                  </p>
                </div>
              </div>
              {!isUnlimited && (
                <div className="space-y-1">
                  <Progress value={Math.min(percentage, 100)} />
                  <p className="text-xs text-muted-foreground text-right">
                    {percentage.toFixed(1)}% used
                  </p>
                </div>
              )}
            </div>
          );
        })}
      </div>
    </Card>
  );
}
