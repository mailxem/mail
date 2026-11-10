export const dynamic = "force-static";
import type { MetadataRoute } from "next";
import { siteUrl } from "@/lib/site";
import { getPosts } from "@/lib/blog";
import { competitors } from "@/lib/compare";
import { galleryTemplates } from "@/lib/template-gallery";
export default function sitemap(): MetadataRoute.Sitemap {
  return [
    { url: siteUrl, changeFrequency: "monthly", priority: 1 },
    { url: `${siteUrl}/pricing`, changeFrequency: "monthly", priority: 0.9 },
    { url: `${siteUrl}/affiliates`, changeFrequency: "monthly", priority: 0.6 },
    { url: `${siteUrl}/mcp`, changeFrequency: "monthly", priority: 0.8 },
    { url: `${siteUrl}/changelog`, changeFrequency: "weekly", priority: 0.6 },
    { url: `${siteUrl}/compare`, changeFrequency: "monthly", priority: 0.8 },
    {
      url: `${siteUrl}/tools/email-checker`,
      changeFrequency: "monthly",
      priority: 0.8,
    },
    ...competitors.flatMap((c) => [
      {
        url: `${siteUrl}/compare/xem-vs-${c.id}`,
        changeFrequency: "monthly" as const,
        priority: 0.7,
      },
      {
        url: `${siteUrl}/migrate/${c.id}`,
        changeFrequency: "monthly" as const,
        priority: 0.6,
      },
    ]),
    {
      url: `${siteUrl}/research/email-authentication`,
      changeFrequency: "yearly",
      priority: 0.8,
    },
    { url: `${siteUrl}/templates`, changeFrequency: "monthly", priority: 0.8 },
    ...galleryTemplates.map((t) => ({
      url: `${siteUrl}/templates/${t.key}`,
      changeFrequency: "monthly" as const,
      priority: 0.5,
    })),
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
