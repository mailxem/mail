import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
} from "@modelcontextprotocol/sdk/types.js";
import { XemClient } from "./client.js";
import { createTools } from "./tools.js";

export function createServer(api: XemClient) {
  const entries = createTools(api);
  const server = new Server(
    { name: "xem-email-mcp-server", version: "2.0.0" },
    { capabilities: { tools: {} } },
  );
  server.setRequestHandler(ListToolsRequestSchema, async () => ({
    tools: [...entries.values()].map((e) => e.tool),
  }));
  server.setRequestHandler(CallToolRequestSchema, async (request) => {
    try {
      const entry = entries.get(request.params.name);
      if (!entry) throw new Error("Unknown tool");
      const result = await entry.execute(request.params.arguments);
      return {
        content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
      };
    } catch (error) {
      return {
        isError: true,
        content: [
          {
            type: "text",
            text: error instanceof Error ? error.message : "Tool failed",
          },
        ],
      };
    }
  });
  return server;
}
