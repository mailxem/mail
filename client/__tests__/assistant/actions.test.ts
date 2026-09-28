import { randomUUID } from "crypto";
const mockScope = {
  userId: "user-a",
  teamId: "team-a",
  canWrite: true,
  accessToken: "test",
};
const mockDefinition = {
  name: "send_email",
  inputSchema: { type: "object", properties: { to: { type: "string" } } },
};
const mockCall = jest.fn();
jest.mock("@/lib/assistant/server", () => ({
  AssistantError: class extends Error {
    constructor(
      public status: number,
      message: string,
    ) {
      super(message);
    }
  },
  authenticate: async () => mockScope,
  enabled: () => true,
  ownerKey: (s: any) => `${s.teamId}:${s.userId}`,
  connectMCP: async () => ({
    listTools: async () => ({ tools: [mockDefinition] }),
    callTool: mockCall,
    close: async () => {},
  }),
  toolResult: (r: any) => r,
  errorResponse: (e: any) =>
    Response.json({ error: e.message }, { status: e.status || 503 }),
}));
import { POST } from "@/app/api/assistant/actions/route";
import { loadConversation, saveConversation } from "@/lib/assistant/store";
import { schemaHash } from "@/lib/assistant/policy";
import type { Conversation } from "@/lib/assistant/types";
const owner = "team-a:user-a";
async function setup() {
  const id = randomUUID();
  const actionId = randomUUID();
  const output = {
    kind: "xem-action",
    actionId,
    tool: "send_email",
    status: "pending",
    input: { to: "reader@example.test" },
  };
  const row: Conversation = {
    id,
    title: "Send",
    createdAt: Date.now(),
    updatedAt: Date.now(),
    messages: [
      {
        id: randomUUID(),
        role: "assistant",
        parts: [
          {
            type: "dynamic-tool",
            toolName: "send_email",
            toolCallId: randomUUID(),
            state: "output-available",
            input: output.input,
            output,
          },
        ],
      },
    ],
    actions: [
      {
        id: actionId,
        tool: "send_email",
        input: output.input,
        schemaHash: schemaHash(mockDefinition),
        status: "pending",
        createdAt: Date.now(),
      },
    ],
  };
  await saveConversation(owner, row);
  return row;
}
function request(row: Conversation, approved = true, extra = {}) {
  return new Request("http://localhost:3000/api/assistant/actions", {
    method: "POST",
    headers: {
      origin: "http://localhost:3000",
      "content-type": "application/json",
    },
    body: JSON.stringify({
      conversationId: row.id,
      actionId: row.actions[0].id,
      approved,
      ...extra,
    }),
  });
}
beforeEach(() => {
  process.env.NEXTAUTH_URL = "http://localhost:3000";
  mockScope.userId = "user-a";
  mockScope.canWrite = true;
  mockCall.mockReset();
  mockCall.mockResolvedValue({
    content: [{ type: "text", text: '{"ok":true}' }],
  });
});
test("execute only saved arguments once; replay is rejected", async () => {
  const row = await setup();
  expect(
    (await POST(request(row, true, { input: { to: "evil@example.test" } })))
      .status,
  ).toBe(400);
  expect((await POST(request(row))).status).toBe(200);
  expect(mockCall).toHaveBeenCalledWith(
    expect.objectContaining({ arguments: { to: "reader@example.test" } }),
  );
  expect((await POST(request(row))).status).toBe(409);
  expect(mockCall).toHaveBeenCalledTimes(1);
});
test("cross-user, member and expired approvals never call MCP", async () => {
  const row = await setup();
  mockScope.userId = "user-b";
  expect((await POST(request(row))).status).toBe(404);
  mockScope.userId = "user-a";
  mockScope.canWrite = false;
  expect((await POST(request(row))).status).toBe(403);
  mockScope.canWrite = true;
  row.actions[0].createdAt = 0;
  await saveConversation(owner, row);
  expect((await POST(request(row))).status).toBe(409);
  expect(mockCall).not.toHaveBeenCalled();
});
test("decline and changed tool schema never execute", async () => {
  const declined = await setup();
  expect((await POST(request(declined, false))).status).toBe(200);
  const row = await setup();
  row.actions[0].schemaHash = "changed";
  await saveConversation(owner, row);
  expect((await POST(request(row))).status).toBe(409);
  expect(mockCall).not.toHaveBeenCalled();
});
test("a timed-out mutation stays unknown and cannot be retried", async () => {
  const row = await setup();
  mockCall.mockRejectedValue(new Error("timeout"));
  expect((await POST(request(row))).status).toBe(200);
  expect((await loadConversation(owner, row.id))?.actions[0].status).toBe(
    "unknown",
  );
  expect((await POST(request(row))).status).toBe(409);
  expect(mockCall).toHaveBeenCalledTimes(1);
});
