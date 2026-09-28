import { NextRequest } from "next/server";
import { middleware } from "@/middleware";
import { auth } from "@/auth";

jest.mock("@/auth", () => ({ auth: jest.fn() }));
const session = jest.mocked(auth) as jest.Mock;

beforeEach(() => {
  session.mockReset();
  session.mockResolvedValue(null);
});

test.each(["/forms/embed.js", "/forms/embed.js?version=1", "/f/example"]) (
  "visitors can load public form content at %s without a session",
  async (path) => {
    const response = await middleware(new NextRequest(`https://app.example${path}`));
    expect(response.status).toBe(200);
    expect(response.headers.get("x-middleware-next")).toBe("1");
    expect(session).not.toHaveBeenCalled();
  },
);

test.each(["/forms", "/forms/new", "/forms/embed.js/private", "/forms/embed.js.bak"]) (
  "the public asset exception does not expose workspace pages at %s",
  async (path) => {
    const response = await middleware(new NextRequest(`https://app.example${path}`));
    expect(response.status).toBe(302);
    expect(response.headers.get("location")).toBe("https://app.example/login");
    expect(session).toHaveBeenCalledTimes(1);
  },
);
