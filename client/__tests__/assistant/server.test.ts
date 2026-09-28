jest.mock("server-only", () => ({}), { virtual: true });
jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("@ai-sdk/mcp", () => ({ createMCPClient: jest.fn() }));
jest.mock("@ai-sdk/openai-compatible", () => ({
  createOpenAICompatible: jest.fn(),
}));
jest.mock("ai", () => ({ dynamicTool: jest.fn() }));
jest.mock("@/lib/assistant/store", () => ({ saveConversation: jest.fn() }));

import { auth } from "@/auth";
import { createMCPClient } from "@ai-sdk/mcp";
import { authenticate, connectMCP, model } from "@/lib/assistant/server";
import { safeEndpoint } from "@/lib/assistant/policy";

const originalEnv = process.env;
const originalFetch = global.fetch;
const mockFetch = jest.fn();
const scope = {
  userId: "user-a",
  teamId: "team-a",
  canWrite: true,
  accessToken: "test-access-token",
};

beforeEach(() => {
  process.env = { ...originalEnv };
  delete process.env.INTERNAL_API_URL;
  delete process.env.NEXT_PUBLIC_API_URL;
  delete process.env.XEM_MCP_URL;
  delete process.env.AI_PROXY_BASE_URL;
  jest.clearAllMocks();
  mockFetch.mockReset();
  global.fetch = mockFetch;
  jest.mocked(auth as () => Promise<unknown>).mockResolvedValue({
    accessToken: scope.accessToken,
  });
});

afterEach(() => {
  process.env = originalEnv;
  global.fetch = originalFetch;
});

test("authentication and scoped credentials use the configured private HTTP backend", async () => {
  process.env.INTERNAL_API_URL = "http://backend:9001/api/v1/";
  process.env.NEXT_PUBLIC_API_URL = "https://api.xem.email/api/v1";
  mockFetch
    .mockResolvedValueOnce(
      Response.json({ id: scope.userId, teamId: scope.teamId, role: "ADMIN" }),
    )
    .mockResolvedValueOnce(
      Response.json({ ...scope, key: "xem_bot_test-credential" }),
    );

  const authenticated = await authenticate();
  expect(authenticated).toEqual(scope);
  await connectMCP(authenticated);

  expect(mockFetch).toHaveBeenNthCalledWith(
    1,
    "http://backend:9001/api/v1/users/me",
    expect.objectContaining({
      headers: { Authorization: `Bearer ${scope.accessToken}` },
      redirect: "error",
      cache: "no-store",
    }),
  );
  expect(mockFetch).toHaveBeenNthCalledWith(
    2,
    "http://backend:9001/api/v1/assistant/credential",
    expect.objectContaining({ method: "POST", redirect: "error" }),
  );
  expect(createMCPClient).toHaveBeenCalledWith(
    expect.objectContaining({
      transport: expect.objectContaining({
        url: "https://mcp.xem.email/mcp",
        headers: { Authorization: "Bearer xem_bot_test-credential" },
      }),
    }),
  );
});

test.each([undefined, "https://api.example.test/api/v1/"])(
  "authentication falls back to the public API or default: %s",
  async (publicUrl) => {
    if (publicUrl) process.env.NEXT_PUBLIC_API_URL = publicUrl;
    mockFetch.mockResolvedValueOnce(
      Response.json({ id: scope.userId, teamId: scope.teamId, role: "MEMBER" }),
    );
    expect((await authenticate()).canWrite).toBe(false);
    expect(mockFetch).toHaveBeenCalledWith(
      `${(publicUrl || "https://api.xem.email/api/v1").replace(/\/$/, "")}/users/me`,
      expect.anything(),
    );
  },
);

test.each([
  "ftp://backend:9001/api/v1",
  "http://user:password@backend:9001/api/v1",
  "http://backend:9001/api/v1?token=secret",
  "http://backend:9001/api/v1#fragment",
])("invalid private API configuration fails before sending credentials: %s", async (url) => {
  process.env.INTERNAL_API_URL = url;
  await expect(authenticate()).rejects.toThrow("Invalid assistant endpoint");
  expect(mockFetch).not.toHaveBeenCalled();
});

test("HTTP permission is limited to the explicitly configured internal API", async () => {
  process.env.NEXT_PUBLIC_API_URL = "http://public-api.example.test/api/v1";
  await expect(authenticate()).rejects.toThrow("Invalid assistant endpoint");
  expect(mockFetch).not.toHaveBeenCalled();
  expect(() => safeEndpoint("http://backend:9001/api/v1")).toThrow();

  process.env.AI_PROXY_BASE_URL = "http://proxy.example.test/v1";
  expect(() => model()).toThrow("Invalid assistant endpoint");

  process.env.INTERNAL_API_URL = "https://api.xem.email/api/v1";
  process.env.XEM_MCP_URL = "http://mcp.example.test/mcp";
  mockFetch.mockResolvedValueOnce(
    Response.json({ ...scope, key: "xem_bot_test-credential" }),
  );
  await expect(connectMCP(scope)).rejects.toThrow("Invalid assistant endpoint");
  expect(createMCPClient).not.toHaveBeenCalled();
});

test("private API support preserves scoped credential identity checks", async () => {
  process.env.INTERNAL_API_URL = "http://backend:9001/api/v1";
  mockFetch.mockResolvedValueOnce(
    Response.json({ ...scope, teamId: "another-team", key: "xem_bot_wrong-team" }),
  );
  await expect(connectMCP(scope)).rejects.toThrow("Workspace access changed");
  expect(createMCPClient).not.toHaveBeenCalled();
});
