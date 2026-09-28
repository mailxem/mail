import { NextResponse } from "next/server";
import { auth } from "@/auth";
import { logger } from "@/app/lib/logger";
import { APIService } from "@/lib/services/api";
import { ApiError } from "@/lib";

const FILE_NAME = "app/api/analytics/trends/route.ts";

interface TrendPoint {
  period: string;
  openCount: number;
  clickCount: number;
  bounceCount: number;
  total: number;
  devices: Record<string, number>;
  locations: Record<string, number>;
  engagementRate: number;
  growth: number;
}

interface TrendAnalyticsResponse {
  data?: {
    date: string;
    openRate: number;
    clickRate: number;
    bounceRate: number;
    total: number;
    devices: Record<string, number>;
    locations: Record<string, number>;
  }[];
  error?: string;
}

export async function GET(request: Request): Promise<NextResponse<TrendAnalyticsResponse>> {
  try {
    const session = await auth();
    if (!session?.user) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "🚫",
        action: "trends",
        label: "analytics_trends",
        value: { userId: null },
        message: "Unauthorized access attempt",
      });
      return NextResponse.json(
        { error: "Unauthorized" },
        { status: 401 }
      );
    }

    const { searchParams } = new URL(request.url);
    const teamId = session.user.teamId;
    if (searchParams.get("teamId") && searchParams.get("teamId") !== teamId) return NextResponse.json({ error: "Forbidden" }, { status: 403 });
    const startDate = searchParams.get("startDate");

    if (!teamId) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "❓",
        action: "trends",
        label: "analytics_trends",
        value: { teamId },
        message: "Missing required teamId parameter",
      });
      return NextResponse.json(
        { error: "teamId is required" },
        { status: 400 }
      );
    }

    const apiService = new APIService("analytics/trends", session);    
    const data = await apiService.get<TrendPoint[]>(null, {
      teamId,
      ...(startDate ? { startDate } : {}),
    });

    logger.info({
      fileName: FILE_NAME,
      emoji: "📈",
      action: "trends",
      label: "analytics_trends", 
      value: { teamId, startDate },
      message: "Successfully retrieved trend analysis",
    });

    // Transform trend data into chart-friendly format
    const transformedData = data?.map(point => ({
      date: point.period,
      openRate: point.total > 0 ? point.openCount / point.total : 0,
      clickRate: point.total > 0 ? point.clickCount / point.total : 0,
      bounceRate: point.total > 0 ? point.bounceCount / point.total : 0,
      total: point.total,
      devices: point.devices,
      locations: point.locations
    }));

    return NextResponse.json({ data: transformedData });
  } catch (error) {
    const apiError = error as ApiError;
    logger.error({
      fileName: FILE_NAME,
      emoji: "❌",
      action: "trends",
      label: "analytics_trends",
      value: { error: apiError.message || "Unknown error" },
      message: "Failed to retrieve trend analysis",
    });
    return NextResponse.json(
      { error: "Failed to retrieve trend analysis" },
      { status: 500 }
    );
  }
}
