import { NextRequest, NextResponse } from "next/server";
import { auth } from "@/auth";
import { APIService } from "@/lib/services/api";
import { logger } from "@/app/lib/logger";
import { TeamInvite, ApiError } from "@/lib";

/**
 * @swagger
 * /api/team/invite:
 *   get:
 *     summary: Get team invites
 *     tags: [Team]
 *     security:
 *       - BearerAuth: []
 *     responses:
 *       200:
 *         description: List of team invites retrieved successfully
 *         content:
 *           application/json:
 *             schema:
 *               type: array
 *               items:
 *                 type: object
 *                 properties:
 *                   id:
 *                     type: string
 *                   email:
 *                     type: string
 *                   status:
 *                     type: string
 *                   inviter:
 *                     type: object
 *                     properties:
 *                       id:
 *                         type: string
 *                       name:
 *                         type: string
 *                       email:
 *                         type: string
 */
export async function GET(
  request: NextRequest
): Promise<NextResponse<TeamInvite[] | { error: string }>> {
  try {
    const session = await auth();

    if (!session?.user) {
      logger.warn({
        fileName: "team-invites.ts",
        emoji: "❌",
        action: "authenticate",
        label: "team invites",
        value: { userId: session.user.id },
        message: "Unauthorized access attempt",
      });
      return new NextResponse("Unauthorized", { status: 401 });
    }

    const apiService = new APIService("team-invitations", session);
    const invites = await apiService.get<{
        data: TeamInvite[];
    }>("", {
        limit: 100,
        include: "Inviter",
    });

    return NextResponse.json(invites.data);
  } catch (error) {
    const apiError = error as ApiError;
    logger.error({
      action: "fetch",
      label: "team invites",
      fileName: "team-invites.ts",
      emoji: "❌",
      value: { error: apiError.message || "Unknown error" },
      message: "Error fetching team invites",
    });
    return NextResponse.json(
      { error: "Internal Server Error" },
      { status: 500 }
    );
  }
}
