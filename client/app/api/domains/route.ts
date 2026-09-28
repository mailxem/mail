import { NextResponse } from "next/server";
import { APIService } from "@/lib/services/api";
import { logger } from "@/app/lib/logger";
import { ApiError, DomainResponse } from "@/lib";
import { auth } from "@/auth";

const FILE_NAME = "app/api/domains/route.ts";

interface GetDomainsParams {
  verified: boolean;
  active: boolean;
}

interface GetDomainsResponse {
  data: DomainResponse[];
  error?: string;
}

/**
 * @swagger
 * /api/domains:
 *   get:
 *     summary: Get all whitelisted domains
 *     tags: [Domains]
 *     security:
 *       - BearerAuth: []
 *     responses:
 *       200:
 *         description: List of whitelisted domains
 *         content:
 *           application/json:
 *             schema:
 *               type: array
 *               items:
 *                 type: string
 *       401:
 *         description: Unauthorized
 */
export async function GET(): Promise<NextResponse<GetDomainsResponse | { error: string }>> {
  try {
    const session = await auth();
    if (!session?.user) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "🚫",
        action: "authenticate",
        label: "domains",
        value: {},
        message: "Unauthorized"
      });
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    logger.info({
      fileName: FILE_NAME,
      emoji: "🔍",
      action: "fetch",
      label: "domains",
      value: { verified: true, active: true },
      message: "Fetching verified and active domains"
    });

    const apiService = new APIService("domains", session);
    const params: GetDomainsParams = {
      verified: true,
      active: true
    };

    const data = await apiService.get<DomainResponse[]>("", params);
    return NextResponse.json({ data });

  } catch (error) {
    const apiError = error as ApiError;
    logger.error({
      fileName: FILE_NAME,
      emoji: "❌",
      action: "fetch",
      label: "domains",
      value: { error: apiError.message || "Unknown error" },
      message: "Failed to fetch domains"
    });
    return NextResponse.json({ error: "Failed to fetch domains" }, { status: 500 });
  }
}
