import { NextRequest, NextResponse } from "next/server";
import { auth } from "@/auth";
import { APIService } from "@/lib/services/api";
import { logger } from "@/app/lib/logger";
import { Team, User, ApiError } from "@/lib";

/**
 * @swagger
 * /api/team/{userid}:
 *   get:
 *     summary: Get team member details
 *     tags: [Team]
 *     security:
 *       - BearerAuth: []
 *     parameters:
 *       - in: path
 *         name: userid
 *         required: true
 *         schema:
 *           type: string
 *         description: User ID of the team member
 *     responses:
 *       200:
 *         description: Team member details retrieved successfully
 *         content:
 *           application/json:
 *             schema:
 *               type: object
 *               properties:
 *                 id:
 *                   type: string
 *                 name:
 *                   type: string
 *                 email:
 *                   type: string
 *                 role:
 *                   type: string
 *       404:
 *         description: Team member not found
 *       401:
 *         description: Unauthorized
 */
interface UpdateTeamRequest {
  name?: string;
}

interface UpdateTeamResponse {
  user: User;
  team: Team;
}

const FILE_NAME = "team/[userid]/route.ts";

export async function GET(
  request: NextRequest
): Promise<NextResponse<Team | { error: string }>> {
  try {
    const session = await auth();

    if (!session?.user) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "🚫",
        action: "authenticate",
        label: "team",
        value: { userId: null },
        message: "Unauthorized access attempt",
      });
      return new NextResponse("Unauthorized", { status: 401 });
    }

    const apiServiceTeam = new APIService("teams", session);

    const team = await apiServiceTeam.get<{
      data: Team[];
    }>("", {
      include: "Users,Invites",
      id: session.user.teamId,
    });

    if (!team.data?.[0]) return NextResponse.json({error:"Workspace not found"},{status:404});
    return NextResponse.json(team.data[0]);
  } catch (error) {
    const apiError = error as ApiError;
    logger.error({
      fileName: FILE_NAME,
      emoji: "❌",
      action: "fetch",
      label: "team",
      value: { error: apiError.message || "Unknown error" },
      message: "Error fetching team data",
    });
    return NextResponse.json(
      { error: "Internal Server Error" },
      { status: 500 }
    );
  }
}

/**
 * @swagger
 * /api/team/{userid}:
 *   put:
 *     summary: Update team member details
 *     tags: [Team]
 *     security:
 *       - BearerAuth: []
 *     parameters:
 *       - in: path
 *         name: userid
 *         required: true
 *         schema:
 *           type: string
 *         description: User ID of the team member
 *     requestBody:
 *       required: true
 *       content:
 *         application/json:
 *           schema:
 *             type: object
 *             properties:
 *               name:
 *                 type: string
 *               role:
 *                 type: string
 *     responses:
 *       200:
 *         description: Team member updated successfully
 *       404:
 *         description: Team member not found
 *       401:
 *         description: Unauthorized
 */
export async function PUT(
  request: Request,
  { params }: { params: Promise<{ userid: string }> }
): Promise<NextResponse<UpdateTeamResponse | { error: string }>> {
  const { userid } = await params;
  try {
    const session = await auth();

    if (!session?.user) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "🚫",
        action: "authenticate",
        label: "team",
        value: { userId: userid },
        message: "Unauthorized access attempt",
      });
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    const body: UpdateTeamRequest = await request.json();
    const apiService = new APIService("team", session);

    logger.info({
      fileName: FILE_NAME,
      emoji: "📝",
      action: "update",
      label: "team",
      value: { userId: userid, updates: body },
      message: "Updating team member",
    });

    const response = await apiService.post<UpdateTeamResponse>(
      session.user.teamId,
      body
    );

    logger.info({
      fileName: FILE_NAME,
      emoji: "✅",
      action: "update",
      label: "team",
      value: { userId: userid },
      message: "Team member updated successfully",
    });

    return NextResponse.json(response);
  } catch (error) {
    const apiError = error as ApiError;
    logger.error({
      fileName: FILE_NAME,
      emoji: "❌",
      action: "update",
      label: "team",
      value: { error: apiError.message || "Unknown error", userId: userid },
      message: "Error updating team member",
    });
    return NextResponse.json(
      { error: "Internal Server Error" },
      { status: 500 }
    );
  }
}
