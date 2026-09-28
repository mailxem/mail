import "server-only";
import { auth } from "@/auth";
import { createMCPClient } from "@ai-sdk/mcp";
import { createOpenAICompatible } from "@ai-sdk/openai-compatible";
import { dynamicTool, type ToolSet } from "ai";
import { randomUUID } from "node:crypto";
import { READ_TOOLS, redact, safeEndpoint, schemaHash } from "./policy";
import { saveConversation } from "./store";
import type { AssistantScope, Conversation } from "./types";

export class AssistantError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
  }
}
export function enabled() {
  return (
    process.env.XEM_ASSISTANT_ENABLED !== "false" &&
    (process.env.NODE_ENV !== "production" || !!process.env.ASSISTANT_REDIS_URL)
  );
}
function backendEndpoint() {
  const internal = process.env.INTERNAL_API_URL;
  // This server-only setting is the operator's trusted service address.
  // Public API, MCP and model endpoints keep the default HTTPS requirement.
  return safeEndpoint(
    internal ||
      process.env.NEXT_PUBLIC_API_URL ||
      "https://api.xem.email/api/v1",
    { allowHTTP: Boolean(internal) },
  );
}
export async function authenticate(): Promise<AssistantScope> {
  const session = await auth();
  if (!session?.accessToken || session.error)
    throw new AssistantError(401, "Please sign in again.");
  const base = backendEndpoint();
  const response = await fetch(`${base}/users/me`, {
    headers: { Authorization: `Bearer ${session.accessToken}` },
    cache: "no-store",
    redirect: "error",
    signal: AbortSignal.timeout(10_000),
  });
  if (!response.ok)
    throw new AssistantError(
      response.status === 401 ? 401 : 503,
      "Your workspace could not be verified. Please try again.",
    );
  const user = await response.json();
  if (!user.id || !user.teamId)
    throw new AssistantError(403, "Join a workspace to use Xem assistant.");
  return {
    userId: user.id,
    teamId: user.teamId,
    canWrite: ["ADMIN", "SUPER_ADMIN"].includes(user.role),
    accessToken: session.accessToken,
  };
}
export const ownerKey = (s: AssistantScope) => `${s.teamId}:${s.userId}`;
export async function connectMCP(scope: AssistantScope) {
  const base = backendEndpoint();
  const response = await fetch(`${base}/assistant/credential`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${scope.accessToken}`,
      "Content-Type": "application/json",
    },
    body: "{}",
    cache: "no-store",
    redirect: "error",
    signal: AbortSignal.timeout(10_000),
  });
  if (!response.ok)
    throw new AssistantError(
      503,
      "Workspace tools are not connected yet. Please try again shortly.",
    );
  const grant = await response.json();
  if (
    !grant.key?.startsWith("xem_bot_") ||
    grant.teamId !== scope.teamId ||
    grant.userId !== scope.userId ||
    grant.canWrite !== scope.canWrite
  )
    throw new AssistantError(
      403,
      "Workspace access changed. Refresh and try again.",
    );
  return createMCPClient({
    clientName: "xem-dashboard",
    maxRetries: 0,
    protocolVersionDiscovery: false,
    initializationOptions: { timeout: 10_000 },
    transport: {
      type: "http",
      url: safeEndpoint(process.env.XEM_MCP_URL || "https://mcp.xem.email/mcp"),
      headers: { Authorization: `Bearer ${grant.key}` },
    },
    onUncaughtError: () => {},
  });
}
export function model() {
  return createOpenAICompatible({
    name: "synehq",
    baseURL: safeEndpoint(
      process.env.AI_PROXY_BASE_URL || "https://ai-proxy.synehq.com/v1",
    ),
    apiKey: process.env.AI_PROXY_API_KEY || undefined,
    transformRequestBody: (body) => ({
      ...body,
      messages: body.messages?.map((message: Record<string, unknown>) => ({
        ...message,
        content: message.content ?? "",
      })),
    }),
    fetch: (url, init) => fetch(url, { ...init, redirect: "error" }),
  }).chatModel(process.env.AI_PROXY_MODEL || "gpt-4.1");
}
export function toolResult(result: unknown) {
  const value = result as {
    content?: { type: string; text?: string }[];
    isError?: boolean;
  };
  if (value.isError)
    return {
      error:
        "Xem could not complete this request. Check permissions and inputs; do not retry a write without checking its outcome.",
    };
  const text =
    value.content
      ?.filter((p) => p.type === "text")
      .map((p) => p.text)
      .join("\n") || "{}";
  try {
    const clean = redact(JSON.parse(text));
    return JSON.stringify(clean).length > 32000
      ? { notice: "Result too large. Request a smaller page or a single item." }
      : clean;
  } catch {
    return redact(text);
  }
}
export async function workspaceTools(
  mcp: Awaited<ReturnType<typeof connectMCP>>,
  scope: AssistantScope,
  row: Conversation,
): Promise<ToolSet> {
  const definitions = await mcp.listTools({ options: { timeout: 10_000 } });
  if (definitions.nextCursor)
    throw new AssistantError(
      503,
      "Tool catalog is incomplete. Try again later.",
    );
  const original = mcp.toolsFromDefinitions(definitions);
  const result: ToolSet = {};
  for (const definition of definitions.tools) {
    const name = definition.name;
    if (
      !scope.canWrite &&
      (!READ_TOOLS.has(name) || name === "get_sending_status")
    )
      continue;
    result[name] = dynamicTool({
      description: definition.description,
      inputSchema: original[name].inputSchema,
      execute: async (input) => {
        if (JSON.stringify(input).length > 32000)
          return { error: "This action is too large. Use smaller batches." };
        if (READ_TOOLS.has(name))
          return toolResult(
            await mcp.callTool({
              name,
              arguments: input as Record<string, unknown>,
              options: { timeout: 20_000 },
            }),
          );
        if (row.actions.filter((a) => a.status === "pending").length >= 4)
          return { error: "Review the existing proposed actions first." };
        const action = {
          id: randomUUID(),
          tool: name,
          input: input as Record<string, unknown>,
          schemaHash: schemaHash(definition),
          status: "pending" as const,
          createdAt: Date.now(),
        };
        row.actions.push(action);
        await saveConversation(ownerKey(scope), row);
        return {
          kind: "xem-action",
          actionId: action.id,
          tool: name,
          status: "pending",
          input: redact(input),
        };
      },
    });
  }
  return result;
}
export function errorResponse(error: unknown) {
  const e =
    error instanceof AssistantError
      ? error
      : new AssistantError(
          503,
          "Xem assistant is temporarily unavailable. Your saved conversation is safe; try again shortly.",
        );
  return Response.json(
    { error: e.message },
    { status: e.status, headers: { "Cache-Control": "no-store" } },
  );
}
