import type { MetadataRoute } from "next";

// The app is behind a login; only published newsletter archives are public pages.
export default function robots(): MetadataRoute.Robots {
  return {
    rules: { userAgent: "*", allow: "/n/", disallow: "/" },
  };
}
