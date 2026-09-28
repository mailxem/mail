/** @type {import('next').NextConfig} */
const nextConfig = {
  redirects: async () => {
    return [
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
      {
        source: "/api/billing/:path*",
        destination: `${process.env.NEXT_PUBLIC_PAYWALL_URL}/:path*`,
        permanent: false,
      },
    ];
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
