let mockConfiguration: any;

jest.mock("next-auth", () => ({
  __esModule: true,
  default: (configuration: unknown) => {
    mockConfiguration = configuration;
    return { handlers: {}, auth: jest.fn(), signIn: jest.fn(), signOut: jest.fn() };
  },
}));
jest.mock("next-auth/providers/credentials", () => ({
  __esModule: true,
  default: (configuration: unknown) => configuration,
}));
jest.mock("next-auth/providers/google", () => ({
  __esModule: true,
  default: () => ({ id: "google" }),
}));
jest.mock("@/auth.config", () => ({ authConfig: { callbacks: {} } }));
jest.mock("@/app/lib/logger", () => ({ logger: { error: jest.fn() } }));
jest.mock("@/lib/utils", () => ({ isJwtExpired: () => true }));

describe("server authentication API origin", () => {
  const originalFetch = global.fetch;
  const originalInternal = process.env.INTERNAL_API_URL;
  const originalPublic = process.env.NEXT_PUBLIC_API_URL;

  beforeEach(() => {
    jest.resetModules();
    global.fetch = jest.fn();
    delete process.env.INTERNAL_API_URL;
    delete process.env.NEXT_PUBLIC_API_URL;
  });

  afterEach(() => {
    global.fetch = originalFetch;
    if (originalInternal === undefined) delete process.env.INTERNAL_API_URL;
    else process.env.INTERNAL_API_URL = originalInternal;
    if (originalPublic === undefined) delete process.env.NEXT_PUBLIC_API_URL;
    else process.env.NEXT_PUBLIC_API_URL = originalPublic;
    jest.clearAllMocks();
  });

  it.each([
    { internal: "http://backend:9001/api/v1", public: "/api/v1" },
    { internal: "http://backend:9001/api/v1/", public: "https://api.example.test/api/v1" },
    { internal: undefined, public: "https://api.example.test/api/v1/" },
  ])("uses the configured server origin for login, identity and refresh (%p)", async ({ internal, public: publicAPI }) => {
    if (internal) process.env.INTERNAL_API_URL = internal;
    process.env.NEXT_PUBLIC_API_URL = publicAPI;
    const expected = (internal || publicAPI).replace(/\/$/, "");
    const { API_URL } = await import("@/auth");
    expect(API_URL).toBe(expected);
    (global.fetch as jest.Mock)
      .mockResolvedValueOnce(new Response(JSON.stringify({
        token: "fixture-access", refresh_token: "fixture-refresh",
      })))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        id: "fixture-user", email: "admin@example.test", firstName: "Test",
        lastName: "Admin", role: "admin", teamId: "fixture-team",
      })))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        token: "fixture-renewed", exp: 1234567890,
      })));

    const credentials = mockConfiguration.providers.find(
      (provider: { id: string }) => provider.id === "credentials",
    );
    const user = await credentials.authorize({
      email: "admin@example.test", password: "fixture-password",
    });
    expect(user).toMatchObject({ id: "fixture-user", accessToken: "fixture-access" });
    expect(global.fetch).toHaveBeenNthCalledWith(1, `${expected}/auth/login`,
      expect.objectContaining({ method: "POST" }),
    );
    expect(global.fetch).toHaveBeenNthCalledWith(2, `${expected}/users/me`,
      expect.objectContaining({ headers: { Authorization: "Bearer fixture-access" } }),
    );

    const refreshed = await mockConfiguration.callbacks.jwt({
      token: { accessToken: "fixture-access", refreshToken: "fixture-refresh" },
    });
    expect(refreshed.accessToken).toBe("fixture-renewed");
    expect(global.fetch).toHaveBeenNthCalledWith(3, `${expected}/auth/refresh`,
      expect.objectContaining({ method: "POST" }),
    );
  });
});
