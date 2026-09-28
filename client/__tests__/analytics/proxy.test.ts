import { GET } from "@/app/api/analytics/v2/[resource]/route";
import { auth } from "@/auth";
jest.mock("@/auth", () => ({ auth: jest.fn() }));
const session = jest.mocked(auth) as jest.Mock;
const context = { params: Promise.resolve({ resource: "report" }) };
const originalFetch = global.fetch;
beforeEach(() => {
  process.env.INTERNAL_API_URL = "https://analytics.example.test/api/v1";
  global.fetch = jest.fn();
});
afterEach(() => {
  global.fetch = originalFetch;
  jest.clearAllMocks();
});
test("anonymous and cross-workspace requests never reach the backend", async () => {
  session.mockResolvedValue(null);
  expect(
    (
      await GET(
        new Request("http://localhost/api/analytics/v2/report"),
        context,
      )
    ).status,
  ).toBe(401);
  session.mockResolvedValue({
    user: { teamId: "workspace-a" },
    accessToken: "test-token",
  });
  expect(
    (
      await GET(
        new Request(
          "http://localhost/api/analytics/v2/report?teamId=workspace-b",
        ),
        context,
      )
    ).status,
  ).toBe(403);
  expect(global.fetch).not.toHaveBeenCalled();
});
test("filters reach the authenticated backend and validation errors retain their status", async () => {
  session.mockResolvedValue({
    user: { teamId: "workspace-a" },
    accessToken: "test-token",
  });
  (global.fetch as jest.Mock).mockResolvedValue(
    new Response(JSON.stringify({ message: "Invalid timezone" }), {
      status: 400,
      headers: { "Content-Type": "application/json" },
    }),
  );
  const result = await GET(
    new Request(
      "http://localhost/api/analytics/v2/report?timezone=invalid&from=2026-09-01&listId=list-a",
    ),
    context,
  );
  expect(result.status).toBe(400);
  expect(await result.json()).toEqual({ error: "Invalid timezone" });
  const [url, options] = (global.fetch as jest.Mock).mock.calls[0];
  expect(new URL(url).searchParams.get("teamId")).toBe("workspace-a");
  expect(new URL(url).searchParams.get("listId")).toBe("list-a");
  expect(options.cache).toBe("no-store");
});
test("unknown resources cannot turn the proxy into an arbitrary authenticated fetch", async () => {
  session.mockResolvedValue({
    user: { teamId: "workspace-a" },
    accessToken: "test-token",
  });
  expect(
    (
      await GET(new Request("http://localhost/api/analytics/v2/secrets"), {
        params: Promise.resolve({ resource: "secrets" }),
      })
    ).status,
  ).toBe(404);
  expect(global.fetch).not.toHaveBeenCalled();
});

test("email overview forwards only the authenticated workspace with uncached results", async () => {
  session.mockResolvedValue({
    user: { teamId: "workspace-a" },
    accessToken: "test-token",
  });
  (global.fetch as jest.Mock).mockResolvedValue(
    new Response(
      JSON.stringify({ metricVersion: "email-v1", summary: { sent: 42 } }),
      { status: 200 },
    ),
  );
  const response = await GET(
    new Request(
      "http://localhost/api/analytics/v2/email-overview?from=2026-09-01&to=2026-09-10&timezone=UTC",
    ),
    { params: Promise.resolve({ resource: "email-overview" }) },
  );
  expect(response.status).toBe(200);
  expect(response.headers.get("Cache-Control")).toBe("private, no-store");
  const [url, options] = (global.fetch as jest.Mock).mock.calls[0];
  expect(new URL(url).pathname).toBe("/api/v1/analytics/email-overview");
  expect(new URL(url).searchParams.get("teamId")).toBe("workspace-a");
  expect(options.headers.Authorization).toBe("Bearer test-token");
});
