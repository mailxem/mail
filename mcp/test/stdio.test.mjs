import { test } from "node:test";
import assert from "node:assert/strict";
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StdioClientTransport } from "@modelcontextprotocol/sdk/client/stdio.js";

test("stdio handshake, tool discovery and invalid call stay protocol-compatible", async () => {
  const transport = new StdioClientTransport({
    command: process.execPath,
    args: ["build/index.js"],
    env: { XEM_API_KEY: "test-only-not-a-real-key" },
    stderr: "pipe",
  });
  const client = new Client({ name: "test", version: "1.0.0" });
  try {
    await client.connect(transport);
    const { tools } = await client.listTools();
    assert.equal(tools.length, 42);
    const response = await client.callTool({
      name: "get_contact",
      arguments: { contactId: "invalid" },
    });
    assert.equal(response.isError, true);
    assert.ok(!JSON.stringify(response).includes("test-only-not-a-real-key"));
  } finally {
    await client.close();
  }
});
