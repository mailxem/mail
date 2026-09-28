import { NextResponse } from "next/server";
import { auth } from "@/auth";
import { logger } from "@/app/lib/logger";
import { APIService } from "@/lib/services/api";
import { ApiError } from "@/lib";

const FILE_NAME = "app/api/analytics/export/route.ts";

interface ExportResponse {
  data?: Uint8Array;
  error?: string;
}

export async function GET(
  request: Request
): Promise<NextResponse<ExportResponse>> {
  try {
    const session = await auth();
    if (!session?.user) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "🚫",
        action: "export",
        label: "analytics_export",
        value: { userId: null },
        message: "Unauthorized access attempt",
      });
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    const { searchParams } = new URL(request.url);
    const campaignId = searchParams.get("campaignId");
    const emailId = searchParams.get("emailId");

    if (!campaignId && !emailId) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "❓",
        action: "export",
        label: "analytics_export",
        value: { campaignId, emailId },
        message: "Missing required ID parameter",
      });
      return NextResponse.json(
        { error: "Either campaignId or emailId is required" },
        { status: 400 }
      );
    }

    let endpoint: string;
    let id: string;
    let type: string;

    if (campaignId) {
      endpoint = "analytics/export/campaign";
      id = campaignId;
      type = "campaign";
    } else {
      endpoint = "analytics/export/email";
      id = emailId!;
      type = "email";
    }

    const apiService = new APIService(endpoint, session);
    const data = await apiService.get<Uint8Array>(`?${type}Id=${id}`);
    const blob = new Blob([data.buffer as ArrayBuffer]);
    logger.info({
      fileName: FILE_NAME,
      emoji: "📊",
      action: "export",
      label: "analytics_export",
      value: { type, id },
      message: `${type} analytics exported successfully`,
    });

    // Set appropriate headers for file download
    return new NextResponse(blob, {
      headers: {
        "Content-Type":
          "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
        "Content-Disposition": `attachment; filename=${type}_analytics_${id}.xlsx`,
      },
    });
  } catch (error) {
    const apiError = error as ApiError;
    logger.error({
      fileName: FILE_NAME,
      emoji: "❌",
      action: "export",
      label: "analytics_export",
      value: { error: apiError.message || "Unknown error" },
      message: "Failed to export analytics",
    });
    return NextResponse.json(
      { error: "Failed to export analytics" },
      { status: 500 }
    );
  }
}
