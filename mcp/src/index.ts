#!/usr/bin/env node
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { XemClient } from "./client.js";
import { createServer } from "./server.js";

async function main() {
  const mode = process.argv[2] ?? "--stdio";
  if (mode === "--http") {
    const { startHTTPServer } = await import("./http.js");
    await startHTTPServer();
    return;
  }
  if (mode !== "--stdio") throw new Error("Use --stdio or --http");
  const server = createServer(
    new XemClient({
      apiKey: process.env.XEM_API_KEY,
      token: process.env.XEM_API_TOKEN,
      baseUrl: process.env.XEM_API_BASE_URL,
    }),
  );
  await server.connect(new StdioServerTransport());
  console.error("Xem MCP server connected over stdio");
}
main().catch(() => {
  console.error(
    "Unable to start Xem MCP. Check transport and environment configuration.",
  );
  process.exitCode = 1;
});
