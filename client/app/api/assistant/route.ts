import { publicConversation, readJSON } from "@/lib/assistant/http";
import { convertToModelMessages, streamText, isStepCount } from "ai";
import { randomUUID } from "node:crypto";
import { z } from "zod";
import { chatInput, SYSTEM_PROMPT } from "@/lib/assistant/policy";
import {
  acquireLock,
  allowRequest,
  deleteConversation,
  listConversations,
  loadConversation,
  saveConversation,
} from "@/lib/assistant/store";
import {
  AssistantError,
  authenticate,
  connectMCP,
  enabled,
  errorResponse,
  model,
  ownerKey,
  workspaceTools,
} from "@/lib/assistant/server";
import type { Conversation } from "@/lib/assistant/types";
export const runtime = "nodejs";
export const maxDuration = 90;
export const dynamic = "force-dynamic";

export async function GET(request: Request) {
  try {
    const scope = await authenticate();
    if (!enabled())
      return Response.json(
        { enabled: false, canWrite: scope.canWrite, conversations: [] },
        { headers: { "Cache-Control": "no-store" } },
      );
    const id = new URL(request.url).searchParams.get("id");
    if (id) {
      if (!z.string().uuid().safeParse(id).success)
        throw new AssistantError(400, "Invalid conversation.");
      const row = await loadConversation(ownerKey(scope), id);
      if (!row)
        throw new AssistantError(
          404,
          "This conversation is no longer available.",
        );
      return Response.json(
        { conversation: publicConversation(row) },
        { headers: { "Cache-Control": "no-store" } },
      );
    }
    return Response.json(
      {
        enabled: true,
        canWrite: scope.canWrite,
        conversations: await listConversations(ownerKey(scope)),
      },
      { headers: { "Cache-Control": "no-store" } },
    );
  } catch (error) {
    return errorResponse(error);
  }
}
export async function POST(request: Request) {
  let release: (() => Promise<void>) | null = null;
  let mcp: Awaited<ReturnType<typeof connectMCP>> | undefined;
  const cleanup = async () => {
    await mcp?.close().catch(() => {});
    await release?.();
    release = null;
  };
  try {
    const input = chatInput.safeParse(await readJSON(request));
    if (!input.success)
      throw new AssistantError(
        400,
        "Enter a message of up to 8,000 characters.",
      );
    const scope = await authenticate();
    if (!enabled())
      throw new AssistantError(
        503,
        "Your workspace assistant is being connected. Please try again later.",
      );
    const owner = ownerKey(scope);
    if (
      !(await allowRequest(owner, 20)) ||
      !(await allowRequest(scope.teamId, 100))
    )
      throw new AssistantError(
        429,
        "A little breather. Try again in a minute.",
      );
    const dailyLimit = Math.floor(
      Math.max(
        1,
        Math.min(
          10000,
          Number(process.env.ASSISTANT_TEAM_DAILY_REQUEST_LIMIT) || 500,
        ),
      ),
    );
    if (!(await allowRequest(`daily:${scope.teamId}`, dailyLimit, 86400)))
      throw new AssistantError(
        429,
        "Your workspace has reached its daily assistant allowance. It resets at midnight UTC.",
      );
    release = await acquireLock(owner);
    if (!release)
      throw new AssistantError(
        409,
        "Another reply or action is still finishing. Please wait a moment.",
      );
    const now = Date.now();
    const row: Conversation = (await loadConversation(
      owner,
      input.data.conversationId,
    )) || {
      id: input.data.conversationId,
      title: input.data.text.slice(0, 80),
      createdAt: now,
      updatedAt: now,
      messages: [],
      actions: [],
    };
    if (row.messages.length >= 60 || row.actions.length >= 60)
      throw new AssistantError(
        400,
        "Start a new chat to continue. This one has reached its limit.",
      );
    mcp = await connectMCP(scope);
    const tools = await workspaceTools(mcp, scope, row);
    row.messages.push({
      id: randomUUID(),
      role: "user",
      parts: [{ type: "text", text: input.data.text }],
    });
    row.updatedAt = now;
    await saveConversation(owner, row);
    const result = streamText({
      model: model(),
      instructions:
        SYSTEM_PROMPT +
        `\nCurrent workspace role: ${scope.canWrite ? "admin" : "read-only member"}. Current UTC date: ${new Date().toISOString().slice(0, 10)}.`,
      messages: await convertToModelMessages(row.messages, {
        ignoreIncompleteToolCalls: true,
      }),
      tools,
      onError: () => {
        console.warn("Xem assistant generation failed");
      },
      stopWhen: isStepCount(6),
      maxOutputTokens: 2400,
      maxRetries: 0,
      timeout: 60_000,
      abortSignal: request.signal,
    });
    return result.toUIMessageStreamResponse({
      originalMessages: row.messages,
      generateMessageId: randomUUID,
      sendReasoning: false,
      headers: { "Cache-Control": "no-store", "X-Accel-Buffering": "no" },
      consumeSseStream: async ({ stream }) => {
        const reader = stream.getReader();
        while (!(await reader.read()).done) {}
      },
      onError: () =>
        "The reply was interrupted. Your saved conversation can be reopened; no proposed action was executed.",
      onEnd: async ({ messages }) => {
        try {
          row.messages = messages;
          row.updatedAt = Date.now();
          await saveConversation(owner, row);
        } finally {
          await cleanup();
        }
      },
    });
  } catch (error) {
    await cleanup();
    return errorResponse(error);
  }
}
export async function DELETE(request: Request) {
  let release: (() => Promise<void>) | null = null;
  try {
    const data = z
      .object({ conversationId: z.string().uuid() })
      .strict()
      .safeParse(await readJSON(request));
    if (!data.success) throw new AssistantError(400, "Invalid conversation.");
    const scope = await authenticate();
    const owner = ownerKey(scope);
    release = await acquireLock(owner);
    if (!release)
      throw new AssistantError(409, "Wait for the current reply to finish.");
    await deleteConversation(owner, data.data.conversationId);
    return new Response(null, { status: 204 });
  } catch (error) {
    return errorResponse(error);
  } finally {
    await release?.();
  }
}
