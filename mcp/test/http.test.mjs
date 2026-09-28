import { test } from "node:test";
import assert from "node:assert/strict";
import { once } from "node:events";
import { createServer as createNetServer } from "node:net";
import { request as httpRequest } from "node:http";
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StreamableHTTPClientTransport } from "@modelcontextprotocol/sdk/client/streamableHttp.js";
import { createHostedServer } from "../build/http.js";

async function fixture() {
  const revoked = new Set();
  const reservation = createNetServer();
  reservation.listen(0, "127.0.0.1");
  await once(reservation, "listening");
  const port = reservation.address().port;
  await new Promise((resolve) => reservation.close(resolve));
  const url = new URL(`http://127.0.0.1:${port}/mcp`);
  const server = createHostedServer({
    publicUrl: url.toString(),
    fetch: async (url, options) => {
      const key = options.headers["X-API-Key"];
      if (!["alpha", "beta"].includes(key) || revoked.has(key))
        return new Response("private failure details", { status: 401 });
      if (url.endsWith("/mcp/authorize"))
        return new Response(null, { status: 204 });
      return new Response(
        JSON.stringify({
          data: [{ email: `${key}@example.com` }],
          total: 1,
          page: 1,
          limit: 20,
        }),
      );
    },
  });
  server.listen(port, "127.0.0.1");
  await once(server, "listening");
  return {
    server,
    url,
    revoked,
    close: () =>
      new Promise((resolve) => {
        server.close(resolve);
        server.closeAllConnections();
      }),
  };
}
test("hosted MCP isolates simultaneous clients and rechecks revocation", async () => {
  const f = await fixture();
  const clients = [];
  try {
    for (const key of ["alpha", "beta"]) {
      const client = new Client({ name: key, version: "1" });
      clients.push(client);
      await client.connect(
        new StreamableHTTPClientTransport(f.url, {
          requestInit: {
            headers: {
              Authorization: `Bearer ${key}`,
            },
          },
        }),
      );
    }
    const responses = await Promise.all(
      clients.map((client) =>
        client.callTool({ name: "get_contacts", arguments: {} }),
      ),
    );
    assert.match(responses[0].content[0].text, /alpha@example.com/);
    assert.doesNotMatch(responses[0].content[0].text, /beta/);
    assert.match(responses[1].content[0].text, /beta@example.com/);
    assert.doesNotMatch(responses[1].content[0].text, /alpha/);
    f.revoked.add("alpha");
    await assert.rejects(clients[0].listTools());
    assert.equal((await clients[1].listTools()).tools.length, 42);
  } finally {
    await Promise.all(clients.map((c) => c.close()));
    await f.close();
  }
});
test("HTTP rejects missing auth, untrusted origins, hosts and oversized input", async () => {
  const f = await fixture();
  const request = (headers = {}, method = "POST") =>
    fetch(f.url, {
      method,
      headers: {
        "Content-Type": "application/json",
        ...headers,
      },
      body: method === "POST" ? "{}" : undefined,
    });
  try {
    assert.equal((await request()).status, 401);
    const invalid = await request({ Authorization: "Bearer invalid" });
    assert.equal(invalid.status, 401);
    assert.ok(!(await invalid.text()).includes("private failure"));
    const wrongHost = await new Promise((resolve) => {
      const req = httpRequest(
        f.url,
        { method: "POST", headers: { Host: "attacker.example" } },
        (res) => {
          res.resume();
          resolve(res.statusCode);
        },
      );
      req.end();
    });
    assert.equal(wrongHost, 403);
    assert.equal(
      (
        await request({
          Origin: "https://attacker.example",
          Authorization: "Bearer alpha",
        })
      ).status,
      403,
    );
    assert.equal((await request({}, "GET")).status, 405);
    const large = await fetch(f.url, {
      method: "POST",
      headers: {
        Authorization: "Bearer alpha",
        "Content-Type": "application/json",
      },
      body: " ".repeat(2 * 1024 * 1024 + 1),
    });
    assert.equal(large.status, 413);
    const cors = await request({ Origin: f.url.origin }, "OPTIONS");
    assert.equal(cors.status, 204);
    assert.equal(cors.headers.get("access-control-allow-origin"), f.url.origin);
  } finally {
    await f.close();
  }
});
