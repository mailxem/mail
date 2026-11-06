import { getPosts } from "@/lib/blog";
import { faqs } from "@/lib/content";
import { competitors } from "@/lib/compare";
import { galleryTemplates } from "@/lib/template-gallery";
import { docsUrl, githubUrl, siteUrl } from "@/lib/site";
export const dynamic = "force-static";
// Plain-text map of the site for AI assistants (llmstxt.org).
export function GET() {
  const posts = getPosts();
  const body = `# Xem

> Xem is an open-source email marketing suite: campaigns, newsletters, visual automations, signup forms, a contact CRM, a shared inbox and analytics in one app. It sends through Amazon SES or any SMTP provider, so sending costs what your provider charges. It can be self-hosted under GPL-3.0 or used at app.xem.email.

## Product

- [Home](${siteUrl}/): what Xem does, templates, analytics and FAQs
- [Pricing](${siteUrl}/pricing): per email sent. Free 3,000 a month, up to 100 a day; Starter $9 with 20,000, then $0.50 per 1,000; Growth $29 with 100,000, then $0.30; Scale $99 with 500,000, then $0.20. Self-hosting is free.
- [MCP server](${siteUrl}/mcp): 44 tools that let AI assistants run Xem campaigns, contacts and analytics
- [Email deliverability checker](${siteUrl}/tools/email-checker): free browser-only check of a domain's MX, SPF (including the 10-lookup limit), DMARC and DKIM records, with the record to add for each problem
- [Source code](${githubUrl}): app, Go backend, SDKs and self-hosting tools

## Compare and switch

- [Compare Xem](${siteUrl}/compare): cost calculator and a weekly-newsletter price table across tools
${competitors.map((c) => `- [Xem vs ${c.name}](${siteUrl}/compare/xem-vs-${c.id}): ${c.name} ${c.modelShort}; Xem charges per email sent. [Move from ${c.name}](${siteUrl}/migrate/${c.id})`).join("\n")}

## Email templates

- [Template gallery](${siteUrl}/templates): ${galleryTemplates.length} editable email templates by category, each opens in the Xem editor
${galleryTemplates.map((t) => `- [${t.name}](${siteUrl}/templates/${t.key}): ${t.category}. ${t.description}`).join("\n")}

## Docs

- [Introduction](${docsUrl}/introduction): what Xem includes and how sending costs work
- [Send with Amazon SES](${docsUrl}/guides/amazon-ses): verify a domain, leave the SES sandbox and connect SES as an SMTP sender
- [Import contacts](${docsUrl}/guides/import-contacts): import CSVs and exports from beehiiv, Substack, Mailchimp, Ghost and Kit
- [Self-host Xem](${docsUrl}/guides/self-hosting): install on your own Linux server with Docker Swarm

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
