import { NextResponse } from "next/server";
import { auth } from "@/auth";
import { logger } from "@/app/lib/logger";
import { APIService } from "@/lib/services/api";
import { ApiError } from "@/lib";

const FILE_NAME = "app/api/analytics/campaign/route.ts";

// Based on Swagger schema
interface EmailAnalytics {
  openCount: number;
  openRate: number;
  uniqueOpens: number;
  clickCount: number;
  clickRate: number;
  uniqueClicks: number;
  bounceCount: number;
  complaintCount: number;
  repeatOpens: number;
  repeatClicks: number;
  firstOpenTime: number;
  averageReadTime: number;
  industryAvgOpenRate: number;
  industryAvgClickRate: number;
  engagementScore: number;
  deviceBreakdown: Record<string, number>;
  browserBreakdown: Record<string, number>;
  osBreakdown: Record<string, number>;
  geoBreakdown: Record<string, number>;
  cityBreakdown: Record<string, number>;
  regionBreakdown: Record<string, number>;
  dayOfWeekBreakdown: Record<string, number>;
  hourlyBreakdown: Record<string, number>;
  timelineData: Array<{
    timestamp: string;
    eventType: string;
    count: number;
    uniqueCount: number;
    deviceType: string;
    browser: string;
    country: string;
  }>;
  clickedLinks: Array<{
    url: string;
    clickCount: number;
    uniqueClicks: number;
    clickRate: number;
    firstClickTime: string;
    lastClickTime: string;
    averageTimeToClick: number;
    deviceBreakdown: Record<string, number>;
  }>;
}

interface CampaignAnalyticsResponse {
  data?: EmailAnalytics | Record<string, EmailAnalytics>;
  error?: string;
}

export async function GET(request: Request): Promise<NextResponse<CampaignAnalyticsResponse>> {
  try {
    const session = await auth();
    if (!session?.user) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "🚫",
        action: "fetch",
        label: "campaign_analytics",
        value: { userId: null },
        message: "Unauthorized access attempt",
      });
      return NextResponse.json(
        { error: "Unauthorized" },
        { status: 401 }
      );
    }

    const { searchParams } = new URL(request.url);
    const campaignId = searchParams.get("campaignId");
    const campaignIds = searchParams.get("campaignIds");

    // Handle campaign comparison endpoint
    if (campaignIds) {
      const ids = campaignIds.split(",");
      const apiService = new APIService("analytics/campaign/compare", session);
      const data = await apiService.get<Record<string, EmailAnalytics>>(null, {
        campaignIds,
      });

      logger.info({
        fileName: FILE_NAME,
        emoji: "📊",
        action: "fetch",
        label: "campaign_comparison",
        value: { campaignIds: ids },
        message: "Campaign comparison retrieved successfully",
      });

      return NextResponse.json({
        data,
        message: "Campaign comparison retrieved successfully",
      });
    }

    // Handle single campaign analytics endpoint
    if (!campaignId) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "❓",
        action: "fetch",
        label: "campaign_analytics",
        value: { campaignId },
        message: "Missing campaign ID parameter",
      });
      return NextResponse.json(
        { error: "Campaign ID is required" },
        { status: 400 }
      );
    }

    const apiService = new APIService("analytics/campaign", session);
    const data = await apiService.get<EmailAnalytics>(null, {
      campaignId,
    });

    logger.info({
      fileName: FILE_NAME,
      emoji: "📈",
      action: "fetch",
      label: "campaign_analytics",
      value: { campaignId },
      message: "Campaign analytics retrieved successfully",
    });

    return NextResponse.json({
      data,
      message: "Campaign analytics retrieved successfully",
    });
  } catch (error) {
    const apiError = error as ApiError;
    logger.error({
      fileName: FILE_NAME,
      emoji: "❌",
      action: "fetch",
      label: "campaign_analytics",
      value: { error: apiError.message || "Unknown error" },
      message: "Failed to fetch campaign analytics",
    });
    return NextResponse.json(
      { error: "Failed to fetch campaign analytics" },
      { status: 500 }
    );
  }
}
