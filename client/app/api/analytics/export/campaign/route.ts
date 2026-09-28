import { NextResponse } from "next/server";
import { auth } from "@/auth";
import { logger } from "@/app/lib/logger";
import { APIService } from "@/lib/services/api";
import { ApiError } from "@/lib";

const FILE_NAME = "app/api/analytics/export/campaign/route.ts";

interface ExportResponse {
  data?: Uint8Array;
  error?: string;
}

export async function GET(request: Request): Promise<NextResponse<ExportResponse>> {
  try {
    const session = await auth();
    if (!session?.user) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "🚫",
        action: "export",
        label: "analytics_export_campaign",
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

    if (!campaignId) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "❓",
        action: "export",
        label: "analytics_export_campaign",
        value: { campaignId },
        message: "Missing required campaignId parameter",
      });
      return NextResponse.json(
        { error: "campaignId is required" },
        { status: 400 }
      );
    }

    const apiService = new APIService("analytics/export/campaign", session);
    const data = await apiService.get<Uint8Array>(null, {
      campaignId,
    });

    logger.info({
      fileName: FILE_NAME,
      emoji: "📊",
      action: "export",
      label: "analytics_export_campaign",
      value: { campaignId },
      message: "Campaign analytics exported successfully",
    });

    // Set appropriate headers for file download
    return new NextResponse(data, {
      headers: {
        'Content-Type': 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
        'Content-Disposition': `attachment; filename=campaign_analytics_${campaignId}.xlsx`
      },
    });
  } catch (error) {
    const apiError = error as ApiError;
    logger.error({
      fileName: FILE_NAME,
      emoji: "❌",
      action: "export",
      label: "analytics_export_campaign",
      value: { error: apiError.message || "Unknown error" },
      message: "Failed to export campaign analytics",
    });
    return NextResponse.json(
      { error: "Failed to export campaign analytics" },
      { status: 500 }
    );
  }
}
