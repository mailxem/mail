import type { NextConfig } from "next";
const nextConfig: NextConfig = {
  output: "export",
  poweredByHeader: false,
  images: { unoptimized: true, qualities: [75, 90] },
  devIndicators: false,
  turbopack: { root: process.cwd() },
};
export default nextConfig;
