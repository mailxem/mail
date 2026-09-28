import { getPosts } from "@/lib/blog";
import { siteUrl } from "@/lib/site";
const xml = (value: string) =>
  value.replace(
    /[<>&"']/g,
    (c) =>
      ({
        "<": "&lt;",
        ">": "&gt;",
        "&": "&amp;",
        '"': "&quot;",
        "'": "&apos;",
      })[c]!,
  );
export const dynamic = "force-static";
export function GET() {
  const posts = getPosts();
  return new Response(
    `<?xml version="1.0" encoding="UTF-8"?><rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom"><channel><title>The Xem Journal</title><link>${xml(siteUrl)}/blog</link><description>Better email, thoughtfully made.</description><language>en</language><atom:link href="${xml(siteUrl)}/feed.xml" rel="self" type="application/rss+xml"/>${posts.map((p) => `<item><title>${xml(p.title)}</title><link>${xml(siteUrl)}/blog/${p.slug}</link><guid isPermaLink="true">${xml(siteUrl)}/blog/${p.slug}</guid><description>${xml(p.description)}</description><pubDate>${new Date(`${p.date}T00:00:00Z`).toUTCString()}</pubDate><category>${xml(p.category)}</category></item>`).join("")}</channel></rss>`,
    {
      headers: {
        "Content-Type": "application/rss+xml; charset=utf-8",
        "Cache-Control": "public, max-age=3600",
      },
    },
  );
}
