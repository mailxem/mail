import { NextResponse } from "next/server";
import { auth } from "@/auth";
import { encodeToBase64, extractVariables } from "@/lib/utils";
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
  request: Request
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
      { status: 500 }
    );
  }
}

export async function POST(
  req: Request
): Promise<NextResponse<TemplateResponse | { error: string }>> {
  try {
    const session = await auth();

    if (!session) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "🚫",
        action: "POST",
        value: {
          emailCategoryId: ``,
        },
        label: "templates",
        message: "Unauthorized access attempt",
      });
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    const json: TemplateRequest = await req.json();

    let variables: string[] = [];
    let htmlFileId: string = "";
    let designJson: string = "";
    if (!json.duplicate) {
      const htmlFile = `${json.id}.html`;
      variables = extractVariables(json.html);

      const createTempFile = require("@/lib/email").createTempFile;

      // Create blob storage for HTML content
      const tempFile = await createTempFile(json.html, htmlFile);
      const fileService = new APIService("files", session);
      const { file } = await fileService.upload(tempFile);
      htmlFileId = file;
      designJson = encodeToBase64(JSON.stringify(json.designJson));
    } else {
      htmlFileId = json.htmlFileId;
      variables = json.variables;
      designJson = json.designJson;
    }

    const templateData = {
      ...json,
      htmlFileId,
      htmlBody: json.html,
      variables,
      designJson,
    };

    const apiService = new APIService("templates", session);
    const data = await apiService.post<TemplateResponse["data"]>(
      "",
      templateData
    );

    logger.info({
      fileName: FILE_NAME,
      emoji: "✅",
      action: "POST",
      label: "template",
      value: data.id,
      message: "Created new email template",
    });

    return NextResponse.json({ data });
  } catch (error) {
    const apiError = error as ApiError;
    logger.error({
      fileName: FILE_NAME,
      emoji: "❌",
      action: "POST",
      label: "template",
      value: apiError.message || "Unknown error",
      message: "Failed to create template",
    });
    return NextResponse.json(
      { error: "Failed to create template" },
      { status: 500 }
    );
  }
}
