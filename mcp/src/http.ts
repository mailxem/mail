import {
  createServer as createHTTPServer,
  type IncomingMessage,
  type ServerResponse,
} from "node:http";
import { createHash } from "node:crypto";
import { StreamableHTTPServerTransport } from "@modelcontextprotocol/sdk/server/streamableHttp.js";
import { APIError, XemClient } from "./client.js";
import { createServer } from "./server.js";

export interface HTTPConfig {
  publicUrl: string;
  apiBaseUrl?: string;
  allowedOrigins?: string[];
  fetch?: typeof fetch;
}
const MAX_BODY = 2 * 1024 * 1024;
function reply(res: ServerResponse, status: number, message: string) {
  res.writeHead(status, { "Content-Type": "application/json" });
  res.end(JSON.stringify({ error: message }));
}
async function readBody(req: IncomingMessage) {
  const chunks: Buffer[] = [];
  let bytes = 0;
  for await (const chunk of req) {
    bytes += chunk.length;
    if (bytes > MAX_BODY) throw new APIError("Request exceeds 2 MiB", 413);
    chunks.push(chunk);
  }
  try {
    return JSON.parse(Buffer.concat(chunks).toString("utf8"));
  } catch {
    throw new APIError("Invalid JSON", 400);
  }
}

export function createHostedServer(config: HTTPConfig) {
  const publicUrl = new URL(config.publicUrl);
  if (
    publicUrl.protocol !== "https:" &&
    !(
      publicUrl.protocol === "http:" &&
      ["localhost", "127.0.0.1", "[::1]"].includes(publicUrl.hostname)
    )
  )
    throw new Error("Hosted public URL requires HTTPS");
  if (
    publicUrl.username ||
    publicUrl.password ||
    publicUrl.search ||
    publicUrl.hash ||
    publicUrl.pathname !== "/mcp"
  )
    throw new Error("Public URL must end in /mcp without credentials or query");
  const origins = new Set([
    publicUrl.origin,
    ...(config.allowedOrigins ?? []).map((value) => new URL(value).origin),
  ]);
  // Store only credential digests for a bounded, per-process rate limit.
  const limits = new Map<string, { count: number; until: number }>();
  let active = 0;
  const http = createHTTPServer(async (req, res) => {
    res.setHeader("Cache-Control", "no-store");
    res.setHeader("X-Content-Type-Options", "nosniff");
    if (publicUrl.protocol === "https:")
      res.setHeader("Strict-Transport-Security", "max-age=31536000");
    if (req.url === "/health" && req.method === "GET") {
      res.writeHead(200);
      res.end("ok");
      return;
    }
    // Trust neither forwarded hosts nor cookies. TLS terminates at the ingress.
    if (req.headers.host !== publicUrl.host) {
      reply(res, 403, "Invalid host");
      return;
    }
    if (req.url !== "/mcp") {
      reply(res, 404, "Not found");
      return;
    }
    const origin = req.headers.origin;
    if (origin && !origins.has(origin)) {
      reply(res, 403, "Origin not allowed");
      return;
    }
    if (origin) {
      res.setHeader("Access-Control-Allow-Origin", origin);
      res.setHeader("Vary", "Origin");
      res.setHeader(
        "Access-Control-Allow-Headers",
        "Authorization, Content-Type, MCP-Protocol-Version",
      );
      res.setHeader("Access-Control-Allow-Methods", "POST, OPTIONS");
    }
    if (req.method === "OPTIONS") {
      res.writeHead(204);
      res.end();
      return;
    }
    if (req.method !== "POST") {
      res.setHeader("Allow", "POST, OPTIONS");
      reply(res, 405, "Stateless MCP supports POST only");
      return;
    }
    const match = /^Bearer ([^\s,]{1,4096})$/.exec(
      req.headers.authorization ?? "",
    );
    if (!match) {
      res.setHeader("WWW-Authenticate", 'Bearer realm="Xem MCP"');
      reply(res, 401, "A Xem API key is required as a Bearer credential");
      return;
    }
    if (
      !req.headers["content-type"]?.toLowerCase().startsWith("application/json")
    ) {
      reply(res, 415, "Content-Type must be application/json");
      return;
    }
    if (Number(req.headers["content-length"] ?? 0) > MAX_BODY) {
      reply(res, 413, "Request exceeds 2 MiB");
      return;
    }
    const now = Date.now();
    for (const [key, limit] of limits)
      if (limit.until <= now) limits.delete(key);
    const digest = createHash("sha256").update(match[1]).digest("hex");
    const limit = limits.get(digest) ?? { count: 0, until: now + 60_000 };
    if (
      active >= 128 ||
      limit.count >= 120 ||
      (!limits.has(digest) && limits.size >= 10_000)
    ) {
      res.setHeader("Retry-After", "60");
      reply(res, 429, "Rate limit exceeded");
      return;
    }
    limit.count++;
    limits.set(digest, limit);
    active++;
    let released = false;
    const release = () => {
      if (!released) {
        released = true;
        active--;
      }
    };
    res.once("close", release);
    try {
      // A new API client, protocol server and transport for EVERY request.
      // No process credential fallback, sessions, or shared tenant state.
      const api = new XemClient(
        { apiKey: match[1], baseUrl: config.apiBaseUrl },
        config.fetch,
      );
      await api.request("/mcp/authorize");
      const body = await readBody(req);
      const server = createServer(api);
      const transport = new StreamableHTTPServerTransport({
        sessionIdGenerator: undefined,
        enableJsonResponse: true,
      });
      res.once("close", () => {
        void server.close().catch(() => {});
      });
      await server.connect(transport);
      await transport.handleRequest(req, res, body);
    } catch (error) {
      if (!res.headersSent && !res.destroyed) {
        const status =
          error instanceof APIError &&
          [400, 401, 403, 413].includes(error.status ?? 0)
            ? error.status!
            : 502;
        if (status === 401)
          res.setHeader("WWW-Authenticate", 'Bearer realm="Xem MCP"');
        reply(
          res,
          status,
          status === 401
            ? "Invalid or expired Xem API key"
            : status === 502
              ? "Xem service unavailable"
              : "Request rejected",
        );
      } else if (!res.destroyed) res.end();
    } finally {
      if (res.writableEnded || res.destroyed) release();
    }
  });
  http.requestTimeout = 60_000;
  http.headersTimeout = 10_000;
  http.timeout = 45_000;
  http.keepAliveTimeout = 5_000;
  http.maxHeadersCount = 50;
  return http;
}
export async function startHTTPServer() {
  const port = Number(process.env.PORT ?? 3000);
  if (!Number.isInteger(port) || port < 1 || port > 65535)
    throw new Error("Invalid PORT");
  const server = createHostedServer({
    publicUrl: process.env.XEM_MCP_PUBLIC_URL ?? "https://mcp.xem.email/mcp",
    apiBaseUrl: process.env.XEM_API_BASE_URL,
    allowedOrigins: process.env.XEM_MCP_ALLOWED_ORIGINS?.split(",")
      .map((v) => v.trim())
      .filter(Boolean),
  });
  await new Promise<void>((resolve, reject) => {
    server.once("error", reject);
    server.listen(port, process.env.HOST ?? "127.0.0.1", resolve);
  });
  console.error("Xem hosted MCP listening");
  const shutdown = () => {
    server.close(() => {
      process.exitCode = 0;
    });
    server.closeIdleConnections();
    const timer = setTimeout(() => server.closeAllConnections(), 35_000);
    timer.unref();
  };
  process.once("SIGTERM", shutdown);
  process.once("SIGINT", shutdown);
}
