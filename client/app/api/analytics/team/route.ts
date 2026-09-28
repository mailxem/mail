import { NextResponse } from "next/server";
import { auth } from "@/auth";
import { logger } from "@/app/lib/logger";
import { APIService } from "@/lib/services/api";
import { ApiError } from "@/lib";

const FILE_NAME = "app/api/analytics/team/route.ts";

// Based on Swagger schema
interface TeamOverview {
  totalEmails: number;
  totalOpens: number;
  totalClicks: number;
  averageOpenRate: number;
  averageClickRate: number;
  deviceStats: Record<string, number>;
  geoStats: Record<string, number>;
  monthlyStats: Array<{
    month: string;
    totalEmails: number;
    openRate: number;
    clickRate: number;
  }>;
  topCampaigns: Array<{
    campaignId: string;
    name: string;
    openRate: number;
    clickRate: number;
    engagementScore: number;
  }>;
  topPerformers: Array<{
    type: string;
    value: string;
    engagementRate: number;
    sampleSize: number;
  }>;
}

interface TeamOverviewResponse {
  data?: TeamOverview;
  error?: string;
}

export async function GET(request: Request): Promise<NextResponse<TeamOverviewResponse>> {
  try {
    const session = await auth();
    if (!session?.user) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "🚫",
        action: "fetch",
        label: "team_analytics",
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
    if (searchParams.get("teamId") && searchParams.get("teamId") !== teamId) return NextResponse.json({ error: "Forbidden" }, {status: 403});
    const type = searchParams.get("type") || "overview"; // overview or trends

    if (!teamId) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "❓",
        action: "fetch",
        label: "team_analytics",
        value: { teamId },
        message: "Missing team ID parameter",
      });
      return NextResponse.json(
        { error: "Team ID is required" },
        { status: 400 }
      );
    }

    const endpoint = type === "trends" ? "analytics/trends" : "analytics/team/overview";
    const apiService = new APIService(endpoint, session);
    const data = await apiService.get<TeamOverview>(null, {
      teamId,
      from: searchParams.get("from") || searchParams.get("startDate") || "",
      to: searchParams.get("to") || searchParams.get("endDate") || "",
      timezone: searchParams.get("timezone") || "UTC",
    });

    logger.info({
      fileName: FILE_NAME,
      emoji: "📈",
      action: "fetch",
      label: "team_analytics",
      value: { teamId, type },
      message: `Team ${type} retrieved successfully`,
    });

    return NextResponse.json({
      data,
      message: `Team ${type} retrieved successfully`,
    });
  } catch (error) {
    const apiError = error as ApiError;
    logger.error({
      fileName: FILE_NAME,
      emoji: "❌",
      action: "fetch",
      label: "team_analytics",
      value: { error: apiError.message || "Unknown error" },
      message: "Failed to fetch team analytics",
    });
    return NextResponse.json(
      { error: "Failed to fetch team analytics" },
      { status: 500 }
    );
  }
}
