import { NextResponse } from "next/server";
import { auth } from "@/auth";
import { logger } from "@/app/lib/logger";
import { APIService } from "@/lib/services/api";
import { ApiError } from "@/lib";

const FILE_NAME = "app/api/analytics/export/email/route.ts";

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
        label: "analytics_export_email",
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
        action: "export",
        label: "analytics_export_email",
        value: { emailId },
        message: "Missing required emailId parameter",
      });
      return NextResponse.json(
        { error: "emailId is required" },
        { status: 400 }
      );
    }

    const apiService = new APIService("analytics/export/email", session);
    const data = await apiService.get<Uint8Array>(null, {
      emailId,
    });

    logger.info({
      fileName: FILE_NAME,
      emoji: "📊",
      action: "export",
      label: "analytics_export_email", 
      value: { emailId },
      message: "Email analytics exported successfully",
    });

    // Set appropriate headers for file download
    return new NextResponse(data, {
      headers: {
        'Content-Type': 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
        'Content-Disposition': `attachment; filename=email_analytics_${emailId}.xlsx`
      },
    });
  } catch (error) {
    const apiError = error as ApiError;
    logger.error({
      fileName: FILE_NAME,
      emoji: "❌",
      action: "export",
      label: "analytics_export_email",
      value: { error: apiError.message || "Unknown error" },
      message: "Failed to export email analytics",
    });
    return NextResponse.json(
      { error: "Failed to export email analytics" },
      { status: 500 }
    );
  }
}
