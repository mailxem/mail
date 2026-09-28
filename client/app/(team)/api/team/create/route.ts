import { NextResponse } from 'next/server';
import { logger } from '@/app/lib/logger';
import { z } from 'zod';
import { auth } from '@/auth';
import { APIService } from '@/lib/services/api';
import { ApiError, Team, User } from '@/lib';

const FILE_NAME = 'app/(team)/api/team/create/route.ts';

// Type definitions
interface CreateTeamRequest {
  name: string;
}

interface CreateTeamResponse {
  data?: {
    id: string;
    name: string;
    users: User[];
  };
  error?: string;
}

// Validation schema for team creation
const TeamCreateSchema = z.object({
  name: z.string().min(2, 'Team name must be at least 2 characters')
});

type TeamCreateInput = z.infer<typeof TeamCreateSchema>;

/**
 * @openapi
 * /api/team/create:
 *   post:
 *     summary: Create a new team
 *     tags: [Team]
 *     security:
 *       - BearerAuth: []
 *     responses:
 *       200:
 *         description: Team created successfully
 *         content:
 *           application/json:
 *             schema:
 *               type: object
 *               properties:
 *                 id:
 *                   type: string
 *                 name:
 *                   type: string
 *                 users:
 *                   type: array
 *                   items:
 *                     type: object
 *                     properties:
 *                       id:
 *                         type: string
 *                       email:
 *                         type: string
 *                       name:
 *                         type: string
 *                       role:
 *                         type: string
 *                       createdAt:
 *                         type: string
 *                         format: date-time
 *                       updatedAt:
 *                         type: string
 *                         format: date-time
 */
export async function POST(
  req: Request
): Promise<NextResponse<CreateTeamResponse>> {
  try {
    const session = await auth();

    if (!session?.user) {
      logger.warn({
        fileName: FILE_NAME,
        emoji: "🚫",
        action: "create",
        label: "team",
        value: { userId: null },
        message: "Unauthorized access attempt"
      });
      return NextResponse.json(
        { error: 'Unauthorized' },
        { status: 401 }
      );
    }

    const body: CreateTeamRequest = await req.json();
    
    const validatedData = TeamCreateSchema.safeParse(body);
    if (!validatedData.success) {
      logger.error({
        fileName: FILE_NAME,
        emoji: "❌",
        action: "validate",
        label: "team",
        value: { errors: validatedData.error.errors },
        message: "Invalid team creation data"
      });
      return NextResponse.json(
        { error: 'Invalid team data' },
        { status: 400 }
      );
    }

    const apiService = new APIService("teams", session);
    const data = await apiService.post<CreateTeamResponse['data']>('/create', validatedData.data);

    return NextResponse.json({ data });

  } catch (error) {
    const apiError = error as ApiError;
    logger.error({
      fileName: FILE_NAME,
      emoji: "❌",
      action: "create",
      label: "team",
      value: { error: apiError.message || "Unknown error" },
      message: "Failed to create team"
    });
    return NextResponse.json(
      { error: 'Failed to create team' },
      { status: 500 }
    );
  }
}
