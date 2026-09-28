// Run with `bun run scripts/verify-assistant.mjs`. Uses the real proxy and local
// MCP transport; all workspace records and actions are isolated synthetic fixtures.
import { mock } from "bun:test";
import { createServer } from "node:http";
import { randomUUID } from "node:crypto";
import assert from "node:assert/strict";
const root = process.cwd();
mock.module("server-only", () => ({}));
mock.module(`${root}/auth.ts`, () => ({
  auth: async () => ({
    accessToken: "local-fixture-session",
    user: { id: "fixture-user" },
  }),
}));
const { createHostedServer } = await import("../../mcp/build/http.js");
let writes = 0;
const fixture = createServer(async (req, res) => {
  const reply = (data) => {
    res.setHeader("Content-Type", "application/json");
    res.end(JSON.stringify(data));
  };
  if (req.url === "/api/v1/users/me")
    return reply({ id: "fixture-user", teamId: "fixture-team", role: "ADMIN" });
  if (req.url === "/api/v1/assistant/credential")
    return reply({
      key: "xem_bot_local_fixture_only",
      teamId: "fixture-team",
      userId: "fixture-user",
      canWrite: true,
    });
  if (req.url === "/api/v1/mcp/authorize") {
    res.writeHead(204);
    return res.end();
  }
  if (req.url?.startsWith("/api/v1/mailing-lists") && req.method === "POST") {
    writes++;
    return reply({
      id: randomUUID(),
      name: "Assistant verification list",
      contactsCount: 0,
    });
  }
  if (req.url?.startsWith("/api/v1/mailing-lists"))
    return reply({
      data: [
        {
          id: "11111111-1111-4111-8111-111111111111",
          name: "Example community",
          subscribersCount: 42,
        },
      ],
    });
  res.writeHead(404);
  res.end();
});
await new Promise((r) => fixture.listen(0, "127.0.0.1", r));
const api = `http://127.0.0.1:${fixture.address().port}/api/v1`;
// Reserve an ephemeral port before configuring the hosted server's host checks.
const reserve = createServer();
await new Promise((r) => reserve.listen(0, "127.0.0.1", r));
const port = reserve.address().port;
await new Promise((r) => reserve.close(r));
const mcp = createHostedServer({
  publicUrl: `http://127.0.0.1:${port}/mcp`,
  apiBaseUrl: api,
});
await new Promise((r) => mcp.listen(port, "127.0.0.1", r));
Object.assign(process.env, {
  NODE_ENV: "test",
  INTERNAL_API_URL: api,
  XEM_MCP_URL: `http://127.0.0.1:${port}/mcp`,
  NEXTAUTH_URL: "http://localhost:3000",
  XEM_ASSISTANT_ENABLED: "true",
});
delete process.env.ASSISTANT_REDIS_URL;
const { POST } = await import("../app/api/assistant/route.ts");
const { POST: approve } = await import("../app/api/assistant/actions/route.ts");
const { loadConversation } = await import("../lib/assistant/store.ts");
const id = randomUUID();
const req = (body) =>
  new Request("http://localhost:3000/api/assistant", {
    method: "POST",
    headers: {
      origin: "http://localhost:3000",
      "content-type": "application/json",
    },
    body: JSON.stringify(body),
  });
try {
  let response = await POST(
    req({
      conversationId: id,
      text: "Use get_contact_lists to list my audiences. Tell me the exact list name and subscriber count.",
    }),
  );
  assert.equal(response.status, 200);
  let stream = await response.text();
  assert.match(stream, /text-delta/);
  assert.match(stream, /get_contact_lists/);
  assert.match(stream, /42/);
  let row = await loadConversation("fixture-team:fixture-user", id);
  assert.ok(row.messages.length >= 2);
  assert.equal(writes, 0);
  console.log(
    "PASS live proxy streaming → 42-tool MCP → workspace lookup → saved transcript",
  );
  response = await POST(
    req({
      conversationId: id,
      text: "Create an empty contact list named Assistant verification list. Use create_contact_list and prepare the review card.",
    }),
  );
  assert.equal(response.status, 200);
  stream = await response.text();
  row = await loadConversation("fixture-team:fixture-user", id);
  const action = row.actions.find((a) => a.status === "pending");
  assert.ok(action, "model did not propose a review card");
  assert.equal(writes, 0);
  console.log("PASS model mutation is staged without executing");
  response = await approve(
    req({ conversationId: id, actionId: action.id, approved: true }),
  );
  assert.equal(response.status, 200);
  assert.equal(writes, 1);
  response = await approve(
    req({ conversationId: id, actionId: action.id, approved: true }),
  );
  assert.equal(response.status, 409);
  assert.equal(writes, 1);
  console.log(
    "PASS exact saved action approved once; replay blocked; no real email sent",
  );
} finally {
  await new Promise((r) => fixture.close(r));
  await new Promise((r) => mcp.close(r));
}
