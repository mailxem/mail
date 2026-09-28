import { NextResponse } from "next/server";
import { auth } from "@/auth";
import { APIService } from "@/lib/services/api";
import { z } from "zod";
import { logger } from "@/app/lib/logger";
import { ApiError } from "@/lib";

const FILE_NAME = "api/settings/domains/route.ts";

// Response interfaces
interface DomainResponse {
  id: string;
  domain: string;
  sslStatus: string;
  isVerified: boolean;
  isActive: boolean;
  verificationToken: string;
  sslExpiresAt?: string;
}

interface GetDomainsResponse {
  data: DomainResponse[];
  error?: string;
}

interface CreateDomainRequest {
  domain: string;
}

interface CreateDomainResponse {
  data: DomainResponse;
  error?: string;
}

const domainSchema = z.object({
  domain: z
    .string()
    .min(1)
    .regex(/^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$/),
});

/**
 * @openapi
 * /api/settings/domains:
 *   get:
 *     summary: List custom domains
 *     description: Retrieve all custom domains for the authenticated team
 *     tags:
 *       - Domains
 *     responses:
 *       200:
 *         description: List of custom domains
 *         content:
 *           application/json:
 *             schema:
 *               type: array
 *               items:
 *                 type: object
 *                 properties:
 *                   id:
 *                     type: string
 *                   domain:
 *                     type: string
 *                   sslStatus:
 *                     type: string
 *                   isVerified:
 *                     type: boolean
 *                   isActive:
 *                     type: boolean
 *                   verificationToken:
 *                     type: string
 *                   sslExpiresAt:
 *                     type: string
 *                     format: date-time
 *                     nullable: true
 *   post:
 *     summary: Add custom domain
 *     description: Add a new custom domain to the authenticated team
 *     tags:
 *       - Domains
 *     requestBody:
 *       required: true
 *       content:
 *         application/json:
 *           schema:
 *             type: object
 *             required:
 *               - domain
 *             properties:
 *               domain:
 *                 type: string
 *                 pattern: ^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$
 *                 example: example.com
 *     responses:
 *       200:
 *         description: Custom domain created successfully
 *         content:
 *           application/json:
 *             schema:
 *               type: object
 *               properties:
 *                 id:
 *                   type: string
 *                 domain:
 *                   type: string
 *                 sslStatus:
 *                   type: string
 *                 isVerified:
 *                   type: boolean
 *                 isActive:
 *                   type: boolean
 *                 verificationToken:
 *                   type: string
 *       400:
 *         description: Invalid domain or domain already exists
 *         content:
 *           application/json:
 *             schema:
 *               type: object
 *               properties:
 *                 message:
 *                   type: string
 */

export async function GET(
  request: Request
): Promise<NextResponse<GetDomainsResponse | { error: string }>> {
  try {
    const session = await auth();
    if (!session?.user) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "🚫",
        action: "list",
        label: "domains",
        value: {},
        message: "Unauthorized access attempt"
      });
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    const apiService = new APIService("domains", session);
    const data = await apiService.get<DomainResponse[]>("");

    return NextResponse.json({ data });
  } catch (error) {
    const apiError = error as ApiError;
    logger.error({
      fileName: FILE_NAME,
      emoji: "❌",
      action: "list",
      label: "domains",
      value: { error: apiError.message || "Unknown error" },
      message: "Failed to fetch domains"
    });
    return NextResponse.json({ error: "Failed to fetch domains" }, { status: 500 });
  }
}

export async function POST(
  req: Request
): Promise<NextResponse<CreateDomainResponse | { error: string }>> {
  try {
    const session = await auth();
    if (!session?.user) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "🚫",
        action: "create",
        label: "domain",
        value: { userId: session?.user?.id || 'unknown' },
        message: "Unauthorized access attempt"
      });
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    const json = await req.json();
    const result = domainSchema.safeParse(json);
    
    if (!result.success) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "❌",
        action: "validate",
        label: "domain",
        value: { error: result.error.message },
        message: "Invalid domain format"
      });
      return NextResponse.json({ error: "Invalid domain format" }, { status: 400 });
    }

    const { domain } = result.data;
    const apiService = new APIService("domains", session);
    const data = await apiService.post<DomainResponse>("", { domain });

    return NextResponse.json({ data });
  } catch (error: unknown) {
    const apiError = error as ApiError;
    logger.error({
      fileName: FILE_NAME,
      emoji: "❌",
      action: "create",
      label: "domain",
      value: { error: apiError.message || "Unknown error" },
      message: "Failed to create domain"
    });
    return NextResponse.json({ error: "Failed to create domain" }, { status: 500 });
  }
}