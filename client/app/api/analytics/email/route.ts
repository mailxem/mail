import { NextResponse } from "next/server";
import { auth } from "@/auth";
import { logger } from "@/app/lib/logger";
import { APIService } from "@/lib/services/api";
import { ApiError } from "@/lib";

const FILE_NAME = "app/api/analytics/email/route.ts";

// Using the same EmailAnalytics interface since the schema is identical to campaign analytics
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

interface EmailAnalyticsResponse {
  data?: EmailAnalytics;
  error?: string;
}

export async function GET(request: Request): Promise<NextResponse<EmailAnalyticsResponse>> {
  try {
    const session = await auth();
    if (!session?.user) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "🚫",
        action: "fetch",
        label: "email_analytics",
        value: { userId: null },
        message: "Unauthorized access attempt",
      });
      return NextResponse.json(
        { error: "Unauthorized" },
        { status: 401 }
      );
    }

    const { searchParams } = new URL(request.url);
    const emailId = searchParams.get("emailId");

    if (!emailId) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "❓",
        action: "fetch",
        label: "email_analytics",
        value: { emailId },
        message: "Missing email ID parameter",
      });
      return NextResponse.json(
        { error: "Email ID is required" },
        { status: 400 }
      );
    }

    const apiService = new APIService("analytics/email", session);
    const data = await apiService.get<EmailAnalytics>(null, {
      emailId,
    });

    logger.info({
      fileName: FILE_NAME,
      emoji: "📧",
      action: "fetch",
      label: "email_analytics",
      value: { emailId },
      message: "Email analytics retrieved successfully",
    });

    return NextResponse.json({
      data,
      message: "Email analytics retrieved successfully",
    });
  } catch (error) {
    const apiError = error as ApiError;
    logger.error({
      fileName: FILE_NAME,
      emoji: "❌",
      action: "fetch",
      label: "email_analytics",
      value: { error: apiError.message || "Unknown error" },
      message: "Failed to fetch email analytics",
    });
    return NextResponse.json(
      { error: "Failed to fetch email analytics" },
      { status: 500 }
    );
  }
}
