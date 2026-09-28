import { NextResponse } from "next/server";
import { auth } from "@/auth";
import { logger } from "@/app/lib/logger";
import { APIService } from "@/lib/services/api";
import { ApiError } from "@/lib";

const FILE_NAME = "app/api/analytics/team/overview/route.ts";

interface CampaignSummary {
  campaignId: string;
  name: string;
  openRate: number;
  clickRate: number;
  engagementScore: number;
}

interface MonthlyEngagement {
  month: string;
  totalEmails: number;
  openRate: number;
  clickRate: number;
}

interface PerformerMetrics {
  type: string;
  value: string;
  engagementRate: number;
  sampleSize: number;
}

interface TeamOverview {
  totalEmails: number;
  totalOpens: number;
  totalClicks: number;
  averageOpenRate: number;
  averageClickRate: number;
  deviceStats: Record<string, number>;
  geoStats: Record<string, number>;
  monthlyStats: MonthlyEngagement[];
  topCampaigns: CampaignSummary[];
  topPerformers: PerformerMetrics[];
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
        label: "team_overview",
        value: { userId: null },
        message: "Unauthorized access attempt",
      });
      return NextResponse.json(
        { error: "Unauthorized" },
        { status: 401 }
      );
    }

    const { searchParams } = new URL(request.url);
    const teamId = searchParams.get("teamId");

    if (!teamId) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "❓",
        action: "fetch",
        label: "team_overview",
        value: { teamId },
        message: "Missing team ID parameter",
      });
      return NextResponse.json(
        { error: "Team ID is required" },
        { status: 400 }
      );
    }

    const apiService = new APIService("analytics", session);
    const data = await apiService.get<TeamOverview>(`team/overview`, {
      teamId,
    });

    logger.info({
      fileName: FILE_NAME,
      emoji: "📈",
      action: "fetch",
      label: "team_overview",
      value: { teamId },
      message: "Team overview retrieved successfully",
    });

    return NextResponse.json({
      data,
      message: "Team overview retrieved successfully",
    });
  } catch (error) {
    const apiError = error as ApiError;
    logger.error({
      fileName: FILE_NAME,
      emoji: "❌",
      action: "fetch",
      label: "team_overview",
      value: { error: apiError.message || "Unknown error" },
      message: "Failed to fetch team overview",
    });
    return NextResponse.json(
      { error: "Failed to fetch team overview" },
      { status: 500 }
    );
  }
}
