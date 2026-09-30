jest.mock("server-only", () => ({}), { virtual: true });
jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("@ai-sdk/mcp", () => ({ createMCPClient: jest.fn() }));
jest.mock("@ai-sdk/openai-compatible", () => ({
  createOpenAICompatible: jest.fn(),
}));
jest.mock("ai", () => ({ dynamicTool: jest.fn() }));
jest.mock("@/lib/assistant/store", () => ({ saveConversation: jest.fn() }));
import { auth } from "@/auth";
import { editorProxy } from "@/lib/assistant/editor-proxy";

const oldFetch = global.fetch;
const fetchMock = jest.fn();
const originalEnv = process.env;
beforeEach(() => {
  process.env = {
    ...originalEnv,
    NEXTAUTH_URL: "https://app.xem.email",
    INTERNAL_API_URL: "http://backend:9001/api/v1",
  };
  jest.clearAllMocks();
  fetchMock.mockReset();
  global.fetch = fetchMock;
  jest
    .mocked(auth as () => Promise<unknown>)
    .mockResolvedValue({ accessToken: "fixture-token" });
});
afterEach(() => {
  global.fetch = oldFetch;
  process.env = originalEnv;
});
const request = (body: unknown, origin = "https://app.xem.email") =>
  new Request("https://app.xem.email/api/assistant/design", {
    method: "POST",
    headers: { origin, "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
const verify = () =>
  fetchMock.mockResolvedValueOnce(
    Response.json({ id: "user", teamId: "team", role: "ADMIN" }),
  );

test("blocks cross-origin and oversized requests before touching credentials", async () => {
  expect(
    (await editorProxy(request({}, "https://evil.example"), "draft")).status,
  ).toBe(403);
  expect(
    (await editorProxy(request({ body: "x".repeat(100_001) }), "draft")).status,
  ).toBe(413);
  expect(auth).not.toHaveBeenCalled();
  expect(fetchMock).not.toHaveBeenCalled();
});
test("rejects a stale workspace without generating or saving", async () => {
  verify();
  expect(
    (await editorProxy(request({ teamId: "old-team" }), "draft")).status,
  ).toBe(409);
  expect(fetchMock).toHaveBeenCalledTimes(1);
});
test("proxies through the fixed internal endpoint, retaining version and server authentication", async () => {
  verify();
  fetchMock.mockResolvedValueOnce(
    Response.json({ id: "saved", updatedAt: "new" }),
  );
  const id = "11111111-1111-4111-8111-111111111111";
  const response = await editorProxy(
    request({
      teamId: "team",
      expectedUpdatedAt: "old",
      htmlBody: "<p>Safe</p>",
    }),
    "save",
    id,
  );
  expect(response.status).toBe(200);
  expect(fetchMock).toHaveBeenLastCalledWith(
    `http://backend:9001/api/v1/marketing/templates/${id}`,
    expect.objectContaining({
      method: "PUT",
      redirect: "error",
      cache: "no-store",
      headers: {
        Authorization: "Bearer fixture-token",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        expectedUpdatedAt: "old",
        htmlBody: "<p>Safe</p>",
      }),
    }),
  );
});
test("preserves conflict errors and never forwards provider diagnostics", async () => {
  verify();
  fetchMock.mockResolvedValueOnce(
    Response.json({ secret: "provider-internal-details" }, { status: 409 }),
  );
  const conflict = await editorProxy(request({ teamId: "team" }), "save");
  expect(conflict.status).toBe(409);
  expect(await conflict.text()).toContain("another session");
  verify();
  fetchMock.mockResolvedValueOnce(
    Response.json({ secret: "provider-internal-details" }, { status: 502 }),
  );
  const failure = await editorProxy(request({ teamId: "team" }), "draft");
  expect(failure.status).toBe(503);
  expect(await failure.text()).not.toContain("provider-internal-details");
});
test("cancellation propagates to the backend request", async () => {
  verify();
  const controller = new AbortController();
  const req = new Request(request({ teamId: "team" }), {
    signal: controller.signal,
  });
  fetchMock.mockImplementationOnce(
    (_url, options) =>
      new Promise((_resolve, reject) => {
        options.signal.addEventListener("abort", () =>
          reject(new DOMException("Cancelled", "AbortError")),
        );
        controller.abort();
      }),
  );
  expect((await editorProxy(req, "draft")).status).toBe(503);
});
