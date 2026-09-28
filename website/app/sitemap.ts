export const dynamic = "force-static";
import type { MetadataRoute } from "next";
import { siteUrl } from "@/lib/site";
import { getPosts } from "@/lib/blog";
export default function sitemap(): MetadataRoute.Sitemap {
  return [
    { url: siteUrl, changeFrequency: "monthly", priority: 1 },
    { url: `${siteUrl}/mcp`, changeFrequency: "monthly", priority: 0.8 },
    { url: `${siteUrl}/blog`, changeFrequency: "weekly", priority: 0.8 },
    {
      url: `${siteUrl}/blog/editorial`,
      changeFrequency: "monthly",
      priority: 0.3,
    },
    ...getPosts().map((p) => ({
      url: `${siteUrl}/blog/${p.slug}`,
      lastModified: `${p.updated}T00:00:00Z`,
      changeFrequency: "monthly" as const,
      priority: 0.7,
    })),
  ];
}
