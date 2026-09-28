export class APIError extends Error {
  constructor(
    message: string,
    readonly status?: number,
  ) {
    super(message);
  }
}
export interface Config {
  apiKey?: string;
  token?: string;
  baseUrl?: string;
}

export class XemClient {
  private readonly base: string;
  constructor(
    private readonly config: Config,
    private readonly transport: typeof fetch = fetch,
  ) {
    const url = new URL(config.baseUrl ?? "https://api.xem.email/api/v1");
    if (
      url.username ||
      url.password ||
      url.search ||
      url.hash ||
      (url.protocol !== "https:" &&
        !(
          url.protocol === "http:" &&
          ["localhost", "127.0.0.1", "[::1]"].includes(url.hostname)
        ))
    ) {
      throw new Error(
        "XEM_API_BASE_URL must use HTTPS (HTTP is allowed only on loopback) without credentials, query or fragment",
      );
    }
    if (!config.apiKey && !config.token)
      throw new Error("Set XEM_API_KEY or XEM_API_TOKEN");
    this.base = url.toString().replace(/\/$/, "");
  }
  async request(
    path: string,
    method = "GET",
    body?: unknown,
  ): Promise<unknown> {
    if (!path.startsWith("/") || path.startsWith("//"))
      throw new Error("Invalid API path");
    const headers: Record<string, string> = { Accept: "application/json" };
    if (this.config.apiKey) headers["X-API-Key"] = this.config.apiKey;
    else headers.Authorization = `Bearer ${this.config.token}`;
    if (body !== undefined) headers["Content-Type"] = "application/json";
    try {
      const response = await this.transport(this.base + path, {
        method,
        headers,
        body: body === undefined ? undefined : JSON.stringify(body),
        redirect: "error",
        signal: AbortSignal.timeout(30_000),
      });
      // Do not echo upstream bodies: they can contain secrets or contact data.
      if (!response.ok) {
        await response.body?.cancel();
        throw new APIError(
          `Xem API returned HTTP ${response.status}. ${response.status === 403 ? "Check resource permissions." : response.status === 401 ? "Check configured credentials." : "Check inputs and server availability; writes are not automatically retried."}`,
          response.status,
        );
      }
      if (response.status === 204) return { ok: true };
      const reader = response.body?.getReader();
      if (!reader) return { ok: true };
      const chunks: Uint8Array[] = [];
      let size = 0;
      for (;;) {
        const { done, value } = await reader.read();
        if (done) break;
        size += value.byteLength;
        if (size > 2 * 1024 * 1024) {
          await reader.cancel();
          throw new APIError("Response exceeds 2 MiB. Request a smaller page.");
        }
        chunks.push(value);
      }
      const text = Buffer.concat(chunks).toString("utf8");
      return text ? JSON.parse(text) : { ok: true };
    } catch (error) {
      if (error instanceof APIError) throw error;
      throw new APIError(
        "Xem request failed or timed out. Check connectivity and server state before retrying a write.",
      );
    }
  }
}
