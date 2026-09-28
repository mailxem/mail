import { NextResponse } from "next/server";
import { auth } from "@/auth";
import { APIService } from "@/lib/services/api";
import { logger } from "@/app/lib/logger";
import { ApiError, Webhook } from "@/lib";

const FILE_NAME = "app/api/webhooks/[webhookId]/deliveries/route.ts";

interface WebhookDelivery {
  id: string;
  status: number;
  // ... other delivery properties
}

interface GetDeliveriesResponse {
  deliveries: WebhookDelivery[];
  total: number;
}

interface GetDeliveriesParams {
  webhookId: string;
  status?: number;
  limit: number;
  offset: number;
}

/**
 * @openapi
 * /api/webhooks/{webhookId}/deliveries:
 *   get:
 *     summary: Get webhook deliveries
 *     tags: [Webhooks]
 *     parameters:
 *       - in: path
 *         name: webhookId
 *         required: true
 *         schema:
 *           type: string
 *     responses:
 *       200:
 *         description: List of webhook deliveries
 *       404:
 *         description: Webhook not found
 *       500:
 *         description: Internal Server Error
 */
export async function GET(
  req: Request,
  { params }: { params: Promise<{ webhookId: string }> }
): Promise<NextResponse<GetDeliveriesResponse | { error: string }>> {
  try {
    const session = await auth();
    const { webhookId } = await params;

    if (!session?.user) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "🚫",
        action: "authenticate",
        label: "webhook_deliveries",
        value: { webhookId },
        message: "Unauthorized access attempt",
      });
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    const apiService = new APIService("marketing", session);
    const { searchParams } = new URL(req.url);
    const limit = Math.min(parseInt(searchParams.get("limit") ?? "50"), 100);
    const offset = Math.max(parseInt(searchParams.get("offset") ?? "0"), 0);
    const status = searchParams.get("status") 
      ? parseInt(searchParams.get("status")) 
      : undefined;

    const response = await apiService.get<GetDeliveriesResponse>(
      `webhooks/${webhookId}/deliveries`,
      { ...(status ? { status } : {}), limit, offset }
    );

    logger.info({
      fileName: FILE_NAME,
      emoji: "✅",
      action: "fetch",
      label: "webhook_deliveries",
      value: { webhookId, count: response.deliveries.length },
      message: "Successfully retrieved webhook deliveries",
    });

    return NextResponse.json(response);
  } catch (error) {
    logger.error({
      fileName: FILE_NAME,
      emoji: "❌",
      action: "fetch",
      label: "webhook_deliveries",
      value: { error: (error as ApiError).message },
      message: "Failed to fetch webhook deliveries",
    });
    return NextResponse.json({ error: "Internal Server Error" }, { status: 500 });
  }
} 