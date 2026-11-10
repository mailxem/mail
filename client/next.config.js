/** @type {import('next').NextConfig} */
const nextConfig = {
  // Public licensed brand fonts must be readable inside the hosted editor iframe.
  headers: async () => [
    {
      source: "/assets/template-starters/brand/:path*",
      headers: [{ key: "Access-Control-Allow-Origin", value: "*" }],
    },
    // Only the app may frame its pages (clickjacking). Signup forms under /f/
    // stay embeddable: forms/embed.js frames them on customers' sites.
    {
      source: "/:path((?!f/).*)",
      headers: [
        { key: "X-Frame-Options", value: "SAMEORIGIN" },
        { key: "Content-Security-Policy", value: "frame-ancestors 'self'" },
      ],
    },
  ],
  redirects: async () => {
    const redirects = [
      {
        source: "/login",
        destination: "/auth/login",
        permanent: true,
      },
      {
        source: "/signup",
        destination: "/auth/signup",
        permanent: true,
      },
      {
        source: "/billing",
        destination: "/billing/overview",
        permanent: true,
      },
    ];
    const paywallURL = process.env.NEXT_PUBLIC_PAYWALL_URL?.trim().replace(
      /\/$/,
      "",
    );
    if (paywallURL) {
      redirects.push({
        source: "/api/billing/:path*",
        destination: `${paywallURL}/:path*`,
        permanent: false,
      });
    }
    return redirects;
  },
  images: {
    remotePatterns: [
      {
        hostname: "**",
      },
    ],
  },
  output: "standalone",
  turbopack: {},
  experimental: { turbopackFileSystemCacheForDev: false },
  transpilePackages: ["bcryptjs", "@maily-to/core", "@maily-to/render"],
};

module.exports = nextConfig;
