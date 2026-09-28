import { NextResponse } from "next/server";
import { auth } from "@/auth";
import { logger } from "@/app/lib/logger";
import { APIService } from "@/lib/services/api";
import { ApiError } from "@/lib";
import { removeUndefined } from "@/lib/utils";
const FILE_NAME = "app/api/analytics/heatmap/route.ts";

// Based on Swagger schema
interface HeatmapData {
  url: string;
  clicks: Array<{
    element: string;
    x: number;
    y: number;
    count: number;
  }>;
  segments: Array<{
    selector: string;
    clickRate: number;
    count: number;
  }>;
}

interface HeatmapResponse {
  data?: HeatmapData;
  error?: string;
}

export async function GET(request: Request): Promise<NextResponse<HeatmapResponse>> {
  try {
    const session = await auth();
    if (!session?.user) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "🚫",
        action: "fetch",
        label: "heatmap_analytics",
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
    const campaignId = searchParams.get("campaignId");

    if (!emailId && !campaignId) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "❓",
        action: "fetch",
        label: "heatmap_analytics",
        value: { emailId, campaignId },
        message: "Missing email ID or campaign ID parameter",
      });
      return NextResponse.json(
        { error: "Email ID or campaign ID is required" },
        { status: 400 }
      );
    }

    const apiService = new APIService("analytics/heatmap", session);
    const data = await apiService.get<HeatmapData>(null, removeUndefined({
      emailId,
      campaignId,
    }));

    logger.info({
      fileName: FILE_NAME,
      emoji: "🎯",
      action: "fetch",
      label: "heatmap_analytics",
      value: { emailId, campaignId },
      message: "Heatmap data retrieved successfully",
    });

    return NextResponse.json({
      data,
      message: "Heatmap data retrieved successfully",
    });
  } catch (error) {
    const apiError = error as ApiError;
    logger.error({
      fileName: FILE_NAME,
      emoji: "❌",
      action: "fetch",
      label: "heatmap_analytics",
      value: { error: apiError.message || "Unknown error" },
      message: "Failed to fetch heatmap data",
    });
    return NextResponse.json(
      { error: "Failed to fetch heatmap data" },
      { status: 500 }
    );
  }
}
