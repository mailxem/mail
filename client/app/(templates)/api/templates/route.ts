import { editorProxy } from "@/lib/assistant/editor-proxy";
import { NextResponse } from "next/server";
import { auth } from "@/auth";
import { APIService } from "@/lib/services/api";
import { logger } from "@/app/lib/logger";
import { ApiError } from "@/lib";

const FILE_NAME = "app/(templates)/api/templates/route.ts";

// Response interfaces
interface TemplateResponse {
  data: {
    id: string;
    html: string;
    variables: string[];
    teamId: string;
    emailCategoryId: string;
  };
  error?: string;
}

// Request interfaces
interface TemplateRequest {
  id: string;
  html: string;
  teamId: string;
  emailCategoryId: string;
  duplicate?: boolean;
  [key: string]: any;
}

export async function GET(
  request: Request,
): Promise<NextResponse<TemplateResponse | { error: string }>> {
  try {
    const session = await auth();
    if (!session?.user) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "🚫",
        action: "GET",
        label: "templates",
        value: {},
        message: "Unauthorized access attempt",
      });
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    const url = new URL(request.url);
    const page = url.searchParams.get("page");
    const limit = url.searchParams.get("limit");

    const apiService = new APIService("templates", session);
    const data = await apiService.get<TemplateResponse["data"][]>("", {
      include: "Category",
      limit: limit ? parseInt(limit) : 50,
      page: page ? parseInt(page) : 1,
      exclude: "designJson",
    });

    logger.info({
      fileName: FILE_NAME,
      emoji: "🔍",
      action: "GET",
      label: "templates",
      value: { count: data.length },
      message: "Retrieved email templates",
    });

    return NextResponse.json({
      ...data,
      error: null,
    });
  } catch (error) {
    const apiError = error as ApiError;
    logger.error({
      fileName: FILE_NAME,
      emoji: "❌",
      action: "GET",
      label: "templates",
      value: { error: apiError.message || "Unknown error" },
      message: "Failed to fetch templates",
    });
    return NextResponse.json(
      { error: "Failed to fetch templates" },
      { status: 500 },
    );
  }
}

export const POST = (request: Request) => editorProxy(request, "save");
