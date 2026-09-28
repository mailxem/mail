import { AssistantError } from "./server";
import { sameOrigin } from "./policy";
import type { Conversation } from "./types";
export function publicConversation(row: Conversation) {
  return {
    id: row.id,
    title: row.title,
    createdAt: row.createdAt,
    updatedAt: row.updatedAt,
    messages: row.messages,
  };
}
export async function readJSON(request: Request) {
  if (!sameOrigin(request))
    throw new AssistantError(403, "Request origin was not accepted.");
  if (!request.headers.get("content-type")?.startsWith("application/json"))
    throw new AssistantError(415, "Expected JSON.");
  const reader = request.body?.getReader();
  if (!reader) throw new AssistantError(400, "Missing request.");
  let size = 0;
  const chunks: Uint8Array[] = [];
  for (;;) {
    const { done, value } = await reader.read();
    if (done) break;
    size += value.length;
    if (size > 20_000) {
      await reader.cancel();
      throw new AssistantError(413, "Your message is too long.");
    }
    chunks.push(value);
  }
  try {
    return JSON.parse(Buffer.concat(chunks).toString("utf8"));
  } catch {
    throw new AssistantError(400, "Invalid request.");
  }
}
