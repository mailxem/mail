import { NextResponse } from "next/server";
import { auth } from "@/auth";
import { z } from "zod";
import { APIService } from "@/lib/services/api";
import { logger } from "@/app/lib/logger";
import { ApiError, WebhookEventType } from "@/lib";

const FILE_NAME = "app/api/webhooks/[webhookId]/route.ts";

interface GetWebhookResponse {
  id: string;
  name: string;
  url: string;
  isActive: boolean;
  events: WebhookEventType[];
  teamId: string;
}

interface UpdateWebhookRequest {
  name?: string;
  url?: string;
  isActive?: boolean;
  events?: WebhookEventType[];
}

interface UpdateWebhookResponse extends GetWebhookResponse {}

const updateWebhookSchema = z.object({
  name: z.string().min(1).optional(),
  url: z.string().url().optional(),
  isActive: z.boolean().optional(),
  events: z.array(z.nativeEnum(WebhookEventType)).optional(),
});

/**
 * @openapi
 * /api/webhooks/{webhookId}:
 *   get:
 *     summary: Get a webhook
 *     tags: [Webhooks]
 *     parameters:
 *       - in: path
 *         name: webhookId
 *         required: true
 *         schema:
 *           type: string
 *     responses:
 *       200:
 *         description: Webhook details
 *       404:
 *         description: Webhook not found
 *       500:
 *         description: Internal Server Error
 */
export async function GET(
  req: Request,
  { params }: { params: Promise<{ webhookId: string }> }
): Promise<NextResponse<GetWebhookResponse | { error: string }>> {
  try {
    const session = await auth();
    if (!session?.user) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "🚫",
        action: "get",
        label: "webhook",
        value: { webhookId: (await params).webhookId },
        message: "Unauthorized access attempt",
      });
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    const apiService = new APIService("webhooks", session);
    const webhook = await apiService.get<GetWebhookResponse>(
      `/${(await params).webhookId}`,
      {
        teamId: session.user.teamId,
      }
    );

    if (!webhook) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "❓",
        action: "get",
        label: "webhook",
        value: { webhookId: (await params).webhookId },
        message: "Webhook not found",
      });
      return NextResponse.json({ error: "Webhook not found" }, { status: 404 });
    }

    return NextResponse.json({ data: webhook, error: null });
  } catch (error) {
    const apiError = error as ApiError;
    logger.error({
      fileName: FILE_NAME,
      emoji: "❌",
      action: "get",
      label: "webhook",
      value: { webhookId: (await params).webhookId },
      message: apiError.message || "Failed to get webhook",
    });
    return NextResponse.json(
      { error: "Internal Server Error" },
      { status: 500 }
    );
  }
}

/**
 * @openapi
 * /api/webhooks/{webhookId}:
 *   patch:
 *     summary: Update a webhook
 *     tags: [Webhooks]
 *     parameters:
 *       - in: path
 *         name: webhookId
 *         required: true
 *         schema:
 *           type: string
 *     requestBody:
 *       required: true
 *       content:
 *         application/json:
 *           schema:
 *             $ref: '#/components/schemas/Webhook'
 *     responses:
 *       200:
 *         description: Webhook updated
 *       400:
 *         description: Invalid request body
 *       404:
 *         description: Webhook not found
 *       500:
 *         description: Internal Server Error
 */
export async function PATCH(
  req: Request,
  { params }: { params: Promise<{ webhookId: string }> }
): Promise<NextResponse<UpdateWebhookResponse | { error: string }>> {
  try {
    const session = await auth();
    if (!session?.user) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "🚫",
        action: "update",
        label: "webhook",
        value: { webhookId: (await params).webhookId },
        message: "Unauthorized access attempt",
      });
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    const json = await req.json();
    const body = updateWebhookSchema.parse(json) as UpdateWebhookRequest;

    const apiService = new APIService("webhooks", session);
    const updatedWebhook = await apiService.post<UpdateWebhookResponse>(
      `/${(await params).webhookId}/update`,
      {
        ...body,
        teamId: session.user.teamId,
      }
    );

    if (!updatedWebhook) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "❓",
        action: "update",
        label: "webhook",
        value: { webhookId: (await params).webhookId },
        message: "Webhook not found",
      });
      return NextResponse.json({ error: "Webhook not found" }, { status: 404 });
    }

    return NextResponse.json({ data: updatedWebhook, error: null });
  } catch (error) {
    const apiError = error as ApiError;
    logger.error({
      fileName: FILE_NAME,
      emoji: "❌",
      action: "update",
      label: "webhook",
      value: { webhookId: (await params).webhookId },
      message: apiError.message || "Failed to update webhook",
    });
    return NextResponse.json(
      { error: "Internal Server Error" },
      { status: 500 }
    );
  }
}

export async function DELETE(
  req: Request,
  { params }: { params: Promise<{ webhookId: string }> }
) {
  try {
    const session = await auth();
    if (!session?.user) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "🚫",
        action: "delete",
        label: "webhook",
        value: { webhookId: (await params).webhookId },
        message: "Unauthorized access attempt",
      });
      return new NextResponse("Unauthorized", { status: 401 });
    }

    logger.info({
      fileName: FILE_NAME,
      emoji: "🔍",
      action: "delete",
      label: "webhook",
      value: { webhookId: (await params).webhookId },
      message: "Attempting to delete webhook",
    });

    const deleted = await new APIService("webhooks", session).delete(
      `/${(await params).webhookId}`
    );

    if (!deleted) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "❓",
        action: "delete",
        label: "webhook",
        value: { webhookId: (await params).webhookId },
        message: "Webhook not found",
      });
      return new NextResponse("Webhook not found", { status: 404 });
    }

    logger.info({
      fileName: FILE_NAME,
      emoji: "✅",
      action: "delete",
      label: "webhook",
      value: { webhookId: (await params).webhookId },
      message: "Successfully deleted webhook",
    });

    return new NextResponse(null, { status: 204 });
  } catch (error: unknown) {
    const apiError = error as ApiError;
    logger.error({
      fileName: FILE_NAME,
      emoji: "❌",
      action: "delete",
      label: "webhook",
      value: { webhookId: (await params).webhookId },
      message: apiError.message || "Failed to delete webhook",
    });
    return new NextResponse("Internal Server Error", { status: 500 });
  }
}
