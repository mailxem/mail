import { getPosts } from "@/lib/blog";
import { faqs } from "@/lib/content";
import { githubUrl, siteUrl } from "@/lib/site";
export const dynamic = "force-static";
// Plain-text map of the site for AI assistants (llmstxt.org).
export function GET() {
  const posts = getPosts();
  const body = `# Xem

> Xem is an open-source email marketing suite: campaigns, newsletters, visual automations, signup forms, a contact CRM, a shared inbox and analytics in one app. It sends through Amazon SES or any SMTP provider, so sending costs what your provider charges. It can be self-hosted under GPL-3.0 or used at app.xem.email.

## Product

- [Home](${siteUrl}/): what Xem does, templates, analytics and FAQs
- [MCP server](${siteUrl}/mcp): 44 tools that let AI assistants run Xem campaigns, contacts and analytics
- [Source code](${githubUrl}): app, Go backend, SDKs and self-hosting tools

## Questions

${faqs.map((f) => `- ${f.q} ${f.a}`).join("\n")}

## Journal

${posts.map((p) => `- [${p.title}](${siteUrl}/blog/${p.slug}): ${p.description}`).join("\n")}
`;
  return new Response(body, {
    headers: {
      "Content-Type": "text/plain; charset=utf-8",
      "Cache-Control": "public, max-age=3600",
    },
  });
}
