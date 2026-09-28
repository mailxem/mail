import { actionInput, schemaHash } from "@/lib/assistant/policy";
import {
  acquireLock,
  allowRequest,
  loadConversation,
  saveConversation,
} from "@/lib/assistant/store";
import {
  AssistantError,
  authenticate,
  connectMCP,
  enabled,
  errorResponse,
  ownerKey,
  toolResult,
} from "@/lib/assistant/server";
import { publicConversation, readJSON } from "@/lib/assistant/http";
import type { ActionOutput } from "@/lib/assistant/types";
export const runtime = "nodejs";
export const maxDuration = 60;
export async function POST(request: Request) {
  let release: (() => Promise<void>) | null = null;
  let mcp: Awaited<ReturnType<typeof connectMCP>> | undefined;
  try {
    const data = actionInput.safeParse(await readJSON(request));
    if (!data.success) throw new AssistantError(400, "Invalid action.");
    const scope = await authenticate();
    if (!scope.canWrite)
      throw new AssistantError(403, "A workspace admin must approve changes.");
    if (!enabled())
      throw new AssistantError(503, "Assistant actions are unavailable.");
    const owner = ownerKey(scope);
    if (!(await allowRequest(owner, 20)))
      throw new AssistantError(429, "Try again in a minute.");
    release = await acquireLock(owner);
    if (!release)
      throw new AssistantError(
        409,
        "Wait for the current reply or action to finish.",
      );
    const row = await loadConversation(owner, data.data.conversationId);
    const action = row?.actions.find((a) => a.id === data.data.actionId);
    if (!row || !action)
      throw new AssistantError(404, "Action not found in this conversation.");
    if (action.status !== "pending")
      throw new AssistantError(
        409,
        "This action has already been handled. Refresh the conversation to see its outcome.",
      );
    if (Date.now() - action.createdAt > 10 * 60_000)
      throw new AssistantError(
        409,
        "This review has expired. Ask Xem to prepare a fresh action.",
      );
    // A stopped stream can leave a saved proposal without a rendered card. Such a
    // proposal must never be executable through an ID guessed from another response.
    const part = row.messages
      .flatMap((m) => m.parts)
      .find(
        (p) =>
          "output" in p && (p.output as ActionOutput)?.actionId === action.id,
      );
    if (!part || !("output" in part))
      throw new AssistantError(
        409,
        "This proposal was interrupted. Ask Xem to prepare it again.",
      );
    if (!data.data.approved) {
      action.status = "declined";
    } else {
      mcp = await connectMCP(scope);
      const catalog = await mcp.listTools({ options: { timeout: 10_000 } });
      const definition = catalog.tools.find((t) => t.name === action.tool);
      if (!definition || schemaHash(definition) !== action.schemaHash)
        throw new AssistantError(
          409,
          "This tool has changed. Ask Xem to prepare a new proposal.",
        );
      action.status = "running";
      (part.output as ActionOutput).status = "running";
      await saveConversation(owner, row);
      // Persist running BEFORE dispatch. A crash or timeout can never replay a write.
      try {
        const output = await mcp.callTool({
          name: action.tool,
          arguments: action.input,
          options: { timeout: 30_000 },
        });
        action.status = output.isError ? "unknown" : "completed";
        action.result = toolResult(output);
      } catch {
        action.status = "unknown";
        action.result = {
          notice:
            "The connection ended before Xem could confirm the outcome. Check the workspace before trying again; it may have completed.",
        };
      }
    }
    part.output = {
      ...(part.output as ActionOutput),
      status: action.status,
      result: action.result,
    };
    row.updatedAt = Date.now();
    await saveConversation(owner, row);
    return Response.json(
      { conversation: publicConversation(row) },
      { headers: { "Cache-Control": "no-store" } },
    );
  } catch (error) {
    return errorResponse(error);
  } finally {
    await mcp?.close().catch(() => {});
    await release?.();
  }
}
