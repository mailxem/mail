const nextConfig = require("../next.config.js");

describe("optional billing integration", () => {
  const originalPaywall = process.env.NEXT_PUBLIC_PAYWALL_URL;

  afterEach(() => {
    if (originalPaywall === undefined) delete process.env.NEXT_PUBLIC_PAYWALL_URL;
    else process.env.NEXT_PUBLIC_PAYWALL_URL = originalPaywall;
  });

  it.each([undefined, "", "  "])(
    "omits the provider redirect when billing is not configured (%p)",
    async (value) => {
      if (value === undefined) delete process.env.NEXT_PUBLIC_PAYWALL_URL;
      else process.env.NEXT_PUBLIC_PAYWALL_URL = value;
      const redirects = await nextConfig.redirects();
      expect(redirects.some((route: { source: string }) =>
        route.source.startsWith("/api/billing"),
      )).toBe(false);
      expect(redirects).toContainEqual({
        source: "/login",
        destination: "/auth/login",
        permanent: true,
      });
    },
  );

  it.each([
    "https://payments.example.test/api/v1",
    "https://payments.example.test/api/v1/",
  ])("preserves the configured hosted billing redirect (%p)", async (value) => {
    process.env.NEXT_PUBLIC_PAYWALL_URL = value;
    expect(await nextConfig.redirects()).toContainEqual({
      source: "/api/billing/:path*",
      destination: "https://payments.example.test/api/v1/:path*",
      permanent: false,
    });
  });
});
