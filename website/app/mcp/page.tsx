import type { Metadata } from "next";
import Link from "next/link";
import {
  ArrowUpRight,
  BarChart3,
  FileSpreadsheet,
  Mail,
  ShieldCheck,
  Terminal,
  Users,
} from "lucide-react";
import { SiteHeader, SiteFooter } from "@/components/site-chrome";
import { siteUrl } from "@/lib/site";
import { primaryButton } from "@/lib/brand";

const title = "Xem MCP Server — Email Marketing for AI Assistants";
const description =
  "Connect your AI assistant to Xem with Model Context Protocol. Create newsletter and campaign drafts, manage contacts, map CSV imports, and explore email analytics.";
export const metadata: Metadata = {
  title,
  description,
  alternates: { canonical: "/mcp" },
  openGraph: { type: "website", url: "/mcp", title, description },
  twitter: { card: "summary", title, description },
};
const workflows = [
  {
    icon: Mail,
    title: "Give the next newsletter a head start",
    text: "Create a campaign or reusable newsletter draft using an audience, template, and sender from your workspace. Schedule a newsletter when you’re ready.",
    prompt:
      "Create a draft of our weekly product newsletter using the community list.",
  },
  {
    icon: Users,
    title: "Keep your audience close",
    text: "Find contact lists, look up subscribers, create contacts, and unsubscribe people who no longer want your emails.",
    prompt: "Find our community list and show me the first 20 active contacts.",
  },
  {
    icon: FileSpreadsheet,
    title: "Make that CSV useful",
    text: "Map your column names to contact fields and preview the import before adding anyone. Existing addresses are skipped, including unsubscribed contacts.",
    prompt:
      "Map Email Address to email and Given Name to firstName. Preview this import.",
  },
  {
    icon: BarChart3,
    title: "Ask what happened after send",
    text: "Explore campaign opens, clicks, bounces, and unsubscribes. Review newsletter delivery counts by edition, then inspect an edition’s engagement.",
    prompt:
      "Show clicks and unsubscribes for our latest campaign over the last 30 days.",
  },
];
const faqs = [
  {
    question: "What is the Xem MCP server?",
    answer:
      "It is an open-source Model Context Protocol server that connects an MCP-compatible AI client to the Xem email API. Connect over hosted HTTPS or run locally from npm over stdio. It exposes tools for newsletters, campaigns, contacts, CSV workflows, and analytics.",
  },
  {
    question: "Which AI assistants can connect?",
    answer:
      "For hosted access, use an MCP client supporting Streamable HTTP and custom Authorization headers. Clients that only support OAuth can use the local npm server over stdio. Configuration formats vary; the examples use the common mcpServers format.",
  },
  {
    question: "Does creating a newsletter send an email?",
    answer:
      "No. Newsletter and campaign creation produce drafts. Scheduling a newsletter is a separate operation that activates delivery at a future time. The individual email tool can also queue a send. Use your MCP client’s approval controls for these actions.",
  },
  {
    question: "How does CSV mapping work?",
    answer:
      "Supply CSV text and map contact fields to your exact column headers. Each batch supports up to 500 rows and 1 MiB. Preview is the default; committing requires dryRun:false. Existing addresses are skipped without changing subscription status. You can also export contacts one page at a time as CSV.",
  },
  {
    question: "What credentials does it need?",
    answer:
      "For hosted access, supply a Xem workspace API key in the Authorization header. For local stdio, configure it in the server environment. The API checks resource permissions and derives workspace access from that key. Credentials are not tool arguments. Returned contact data is visible to your AI client, so choose a client configuration suitable for that data.",
  },
];
const hostedConfig = `{
  "mcpServers": {
    "xem": {
      "url": "https://mcp.xem.email/mcp",
      "headers": {
        "Authorization": "Bearer your-workspace-api-key"
      }
    }
  }
}`;
const localConfig = `{
  "mcpServers": {
    "xem": {
      "command": "npx",
      "args": ["-y", "@xem.email/mcp@2", "--stdio"],
      "env": { "XEM_API_KEY": "your-workspace-api-key" }
    }
  }
}`;
export default function McpPage() {
  const structuredData = {
    "@context": "https://schema.org",
    "@graph": [
      {
        "@type": "SoftwareSourceCode",
        "@id": `${siteUrl}/mcp#software`,
        name: "Xem MCP server",
        description,
        url: `${siteUrl}/mcp`,
        codeRepository: "https://github.com/mailxem/mcp",
        programmingLanguage: "TypeScript",
        runtimePlatform: "Node.js",
        license: "https://opensource.org/license/mit",
      },
      {
        "@type": "BreadcrumbList",
        itemListElement: [
          { "@type": "ListItem", position: 1, name: "Xem", item: siteUrl },
          {
            "@type": "ListItem",
            position: 2,
            name: "MCP server",
            item: `${siteUrl}/mcp`,
          },
        ],
      },
    ],
  };
  return (
    <>
      <SiteHeader />
      <main id="main">
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{
            __html: JSON.stringify(structuredData).replace(/</g, "\\u003c"),
          }}
        />
        <section className="mx-auto max-w-[1160px] px-5 pb-16 pt-16 md:pb-24 md:pt-24">
          <Link href="/" className="text-xs text-muted hover:underline">
            Xem /
          </Link>
          <span className="ml-2 text-xs text-muted">MCP server</span>
          <p className="mt-9 text-xs font-medium uppercase tracking-[0.17em] text-iris">
            Email marketing, in conversation
          </p>
          <h1 className="mt-5 max-w-4xl font-editorial text-[clamp(3rem,6.5vw,6rem)] leading-[1.02] tracking-[-0.045em]">
            Your assistant.
            <br />
            Your audience.
            <br />
            <em>A little more connected.</em>
          </h1>
          <p className="mt-7 max-w-xl text-lg leading-relaxed text-muted">
            The Xem MCP server brings newsletters, contacts, and campaign
            analytics into your AI workflow. Turn an idea into a draft, a CSV
            into an audience, and results into your next question.
          </p>
          <div className="mt-8 flex flex-wrap items-center gap-6">
            <a href="#setup" className={primaryButton}>
              Connect your workspace <ArrowUpRight size={17} />
            </a>
            <a
              href="https://github.com/mailxem/mcp"
              className="text-sm underline underline-offset-4"
            >
              Explore the source
            </a>
          </div>
        </section>
        <section
          aria-labelledby="workflows"
          className="bg-[#ebe7f2] px-5 py-16 md:py-24"
        >
          <div className="mx-auto max-w-[1160px]">
            <p className="text-xs uppercase tracking-widest text-iris">
              Less switching. More doing.
            </p>
            <h2
              id="workflows"
              className="mt-4 font-editorial text-4xl tracking-tight md:text-5xl"
            >
              Start with a simple ask.
            </h2>
            <div className="mt-10 grid gap-5 md:grid-cols-2">
              {workflows.map(({ icon: Icon, title, text, prompt }) => (
                <article
                  key={title}
                  className="rounded-2xl border border-ink/10 bg-cream p-7 md:p-9"
                >
                  <Icon size={25} strokeWidth={1.4} aria-hidden="true" />
                  <h3 className="mt-6 font-editorial text-3xl tracking-tight">
                    {title}
                  </h3>
                  <p className="mt-4 text-sm leading-relaxed text-muted">
                    {text}
                  </p>
                  <blockquote className="mt-6 border-l-2 border-iris/40 pl-4 text-sm leading-relaxed">
                    “{prompt}”
                  </blockquote>
                </article>
              ))}
            </div>
          </div>
        </section>
        <section
          className="mx-auto grid max-w-[1160px] gap-8 px-5 py-16 md:grid-cols-2 md:py-24"
          aria-labelledby="control"
        >
          <div>
            <ShieldCheck size={28} aria-hidden="true" />
            <h2
              id="control"
              className="mt-5 font-editorial text-4xl tracking-tight"
            >
              Useful access.
              <br />
              <em>Thoughtful defaults.</em>
            </h2>
          </div>
          <div className="space-y-5 text-sm leading-relaxed text-muted">
            <p>
              Access follows your workspace API key and its resource
              permissions. Grant the capabilities you need, from reading
              analytics to creating contacts.
            </p>
            <p>
              Campaigns start as drafts. CSV imports preview first and validate
              the whole batch before a write. Suppressed contacts stay
              suppressed. Scheduling and sending are separate, explicit tools.
            </p>
            <p>
              Contact data returned by the server is shared with your MCP
              client. Choose its approval and data settings to match your
              workflow.
            </p>
          </div>
        </section>
        <section
          id="setup"
          aria-labelledby="setup-title"
          className="scroll-mt-10 bg-[#222d25] px-5 py-16 text-cream md:py-24"
        >
          <div className="mx-auto max-w-[1160px]">
            <Terminal size={26} aria-hidden="true" />
            <h2
              id="setup-title"
              className="mt-5 font-editorial text-4xl tracking-tight md:text-5xl"
            >
              A small setup.
              <br />A connected workspace.
            </h2>
            <div className="mt-10 grid min-w-0 gap-10 md:grid-cols-2">
              <div>
                <p className="text-xs font-medium uppercase tracking-widest text-lemon">
                  Hosted · Recommended
                </p>
                <h3 className="mt-4 font-editorial text-3xl">
                  Just bring your workspace.
                </h3>
                <p className="mt-5 text-sm leading-relaxed text-cream/80">
                  Add the Xem HTTPS endpoint to your MCP client and connect with
                  your workspace API key. No local server to install or
                  maintain.
                </p>
                <p className="mt-5 break-all rounded-lg border border-cream/20 px-4 py-3 font-mono text-sm">
                  https://mcp.xem.email/mcp
                </p>
                <ol className="mt-6 list-decimal space-y-4 pl-5 text-sm leading-relaxed text-cream/80">
                  <li>Create a Xem API key with the permissions you need.</li>
                  <li>
                    Add the endpoint as a Streamable HTTP server. Set the
                    Authorization header using your client’s secret
                    configuration.
                  </li>
                  <li>
                    Connect and ask for your contact lists. Try a newsletter
                    draft or CSV preview next.
                  </li>
                </ol>
                <p className="mt-5 text-xs leading-relaxed text-cream/70">
                  Hosted access uses an API key, not an OAuth login. If your
                  client cannot set custom headers, use the local option below.
                </p>
              </div>
              <div className="min-w-0">
                <p className="mb-3 text-xs text-cream/70">
                  Hosted MCP client configuration
                </p>
                <pre className="overflow-x-auto rounded-xl border border-cream/20 bg-black/15 p-5 text-xs leading-6">
                  <code>{hostedConfig}</code>
                </pre>
              </div>
            </div>
            <div className="mt-12 grid min-w-0 gap-10 border-t border-cream/20 pt-10 md:grid-cols-2">
              <div>
                <p className="text-xs font-medium uppercase tracking-widest text-lemon">
                  Local · npm + stdio
                </p>
                <h3 className="mt-4 font-editorial text-3xl">
                  Prefer to run it yourself?
                </h3>
                <p className="mt-5 text-sm leading-relaxed text-cream/80">
                  Use Node.js 20 or newer and let your MCP client launch the npm
                  package. Your API key stays in the local server environment
                  and requests go directly to the Xem API.
                </p>
                <p className="mt-5 text-sm leading-relaxed text-cream/80">
                  You can also install it with{" "}
                  <code className="break-all">
                    npm install -g @xem.email/mcp@2
                  </code>{" "}
                  and launch <code>xem-email-mcp --stdio</code>.
                </p>
              </div>
              <div className="min-w-0">
                <p className="mb-3 text-xs text-cream/70">
                  Local MCP client configuration
                </p>
                <pre className="overflow-x-auto rounded-xl border border-cream/20 bg-black/15 p-5 text-xs leading-6">
                  <code>{localConfig}</code>
                </pre>
              </div>
            </div>
            <p className="mt-8 max-w-3xl text-xs leading-relaxed text-cream/70">
              Self-hosting Xem? Run the MCP service with Docker Compose and
              automatic HTTPS, or point the local package at your API with
              XEM_API_BASE_URL. Use matching 2.x MCP and Xem server releases.
              See the{" "}
              <a
                href="https://github.com/mailxem/mcp#readme"
                className="underline"
              >
                README for deployment, permissions, and migration instructions
              </a>
              .
            </p>
          </div>
        </section>
        <section
          aria-labelledby="faq-title"
          className="mx-auto max-w-3xl px-5 py-16 md:py-24"
        >
          <h2
            id="faq-title"
            className="mb-8 font-editorial text-4xl tracking-tight"
          >
            A few good questions.
          </h2>
          {faqs.map((faq) => (
            <details key={faq.question} className="border-b border-ink/15 py-5">
              <summary className="cursor-pointer text-base font-medium">
                {faq.question}
              </summary>
              <p className="mt-4 text-sm leading-relaxed text-muted">
                {faq.answer}
              </p>
            </details>
          ))}
          <p className="mt-10 text-sm text-muted">
            Looking for email strategy alongside the tools?{" "}
            <Link href="/blog" className="text-iris underline">
              Explore the Xem journal.
            </Link>
          </p>
        </section>
      </main>
      <SiteFooter />
    </>
  );
}
