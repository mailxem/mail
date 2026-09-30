import { editorProxy } from "@/lib/assistant/editor-proxy";
import { NextResponse } from "next/server";
import { logger } from "@/app/lib/logger";
import { APIService } from "@/lib/services/api";
import { ApiError } from "@/lib";
import { auth } from "@/auth";
import { EmailTemplate } from "@/lib";
const FILE_NAME = "app/(templates)/api/templates/[id]/route.ts";

/**
 * @openapi
 * /api/templates/{id}:
 *   get:
 *     summary: Get a specific email template
 *     tags: [Templates]
 *     security:
 *       - BearerAuth: []
 *     parameters:
 *       - in: path
 *         name: id
 *         required: true
 *         schema:
 *           type: string
 *         description: Template ID
 *     responses:
 *       200:
 *         description: Email template details
 *         content:
 *           application/json:
 *             schema:
 *               type: object
 *               properties:
 *                 id:
 *                   type: string
 *                 name:
 *                   type: string
 *                 html:
 *                   type: string
 *                 variables:
 *                   type: array
 *                   items:
 *                     type: string
 *       401:
 *         description: Unauthorized
 *       404:
 *         description: Template not found
 *       500:
 *         description: Internal Server Error
 */
interface GetTemplateResponse {
  data?: EmailTemplate;
  error?: string;
}

export async function GET(
  req: Request,
  { params }: { params: Promise<{ id: string }> },
): Promise<NextResponse<GetTemplateResponse>> {
  const { id } = await params;

  try {
    const session = await auth();
    if (!session?.user) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "🚫",
        action: "authenticate",
        label: "template",
        value: {},
        message: "Unauthorized",
      });
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    const apiService = new APIService("templates", session);
    const data = await apiService.get<EmailTemplate>(`${id}`);

    if (!data) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "❓",
        action: "fetch",
        label: "template",
        value: { templateId: id },
        message: "Template not found",
      });
      return NextResponse.json(
        { error: "Template not found" },
        { status: 404 },
      );
    }

    // @ts-ignore
    data.design = JSON.parse(
      Buffer.from(data.designJson, "base64").toString("utf-8"),
    );

    logger.info({
      fileName: FILE_NAME,
      emoji: "📧",
      action: "fetch",
      label: "template",
      value: { templateId: data.id },
      message: "Retrieved email template",
    });

    return NextResponse.json({ data });
  } catch (error) {
    const apiError = error as ApiError;
    console.log("error", apiError);
    logger.error({
      fileName: FILE_NAME,
      emoji: "❌",
      action: "fetch",
      label: "template",
      value: { error: apiError.message || "Unknown error" },
      message: "Error retrieving template",
    });
    return NextResponse.json(
      { error: "Internal Server Error" },
      { status: 500 },
    );
  }
}

export async function PUT(
  request: Request,
  context: { params: Promise<{ id: string }> },
) {
  return editorProxy(request, "save", (await context.params).id);
}

interface DeleteTemplateResponse {
  error?: string;
}

export async function DELETE(
  req: Request,
  { params }: { params: Promise<{ id: string }> },
): Promise<NextResponse<DeleteTemplateResponse>> {
  const id = (await params).id;

  try {
    const session = await auth();
    if (!session?.user) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    const apiService = new APIService("templates", session);
    await apiService.delete(`${id}`);

    logger.info({
      fileName: FILE_NAME,
      emoji: "✅",
      action: "delete",
      label: "template",
      value: { templateId: id },
      message: "Deleted email template",
    });

    return NextResponse.json({});
  } catch (error) {
    logger.error({
      fileName: FILE_NAME,
      emoji: "❌",
      action: "delete",
      label: "template",
      value: { error, templateId: id },
      message: "Error deleting template",
    });

    return NextResponse.json(
      { error: "Failed to delete template" },
      { status: 500 },
    );
  }
}
