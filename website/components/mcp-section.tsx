"use client";
import { useEffect, useState } from "react";
import Link from "next/link";
import { ArrowRight, MousePointer2 } from "lucide-react";
import { docsUrl } from "@/lib/site";
import { agents } from "./agent-strip";

// Real tools from Xem's MCP server (mcp/src/tools.ts).
const examples = [
  { agent: 0, does: "imports your subscribers", tool: "import_contacts" },
  { agent: 1, does: "drafts your next campaign", tool: "create_campaign" },
  {
    agent: 3,
    does: "checks your sending domain",
    tool: "check_sending_domain",
  },
  { agent: 2, does: "reads last week’s results", tool: "get_campaign_metrics" },
  { agent: 0, does: "schedules the next issue", tool: "schedule_newsletter" },
  { agent: 1, does: "tags your best readers", tool: "set_contact_tags" },
];
// The strip's three marks, plus Cursor, which has no open mark to use.
const clients = [
  ...agents.map((a) => ({ ...a, name: a.name.replace(" in VS Code", "") })),
  { name: "Cursor", bg: "#1b1b1b", fg: "#ffffff", path: "" },
];

export function McpSection() {
  const [i, setI] = useState(0);
  useEffect(() => {
    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;
    const t = setInterval(() => setI((n) => (n + 1) % examples.length), 2600);
    return () => clearInterval(t);
  }, []);
  const now = examples[i];
  const who = clients[now.agent];

  return (
    <section
      aria-labelledby="mcp-heading"
      className="px-5 py-16 md:px-10 md:py-24"
    >
      <div className="relative mx-auto max-w-[1160px] overflow-hidden rounded-[28px] bg-forest px-6 py-16 text-center text-cream md:px-12 md:py-20">
        <div
          aria-hidden="true"
          className="pointer-events-none absolute inset-0 opacity-[.18] [background-image:radial-gradient(#fffef0_1px,transparent_1px)] [background-size:22px_22px] [mask-image:radial-gradient(ellipse_at_top,black,transparent_70%)]"
        />
        <div className="relative">
          <h2
            id="mcp-heading"
            className="mx-auto max-w-[620px] font-editorial text-[clamp(2.6rem,5vw,4.2rem)] leading-[1.02] tracking-[-.04em]"
          >
            Your AI agent can do everything you can
          </h2>

          <ul className="mt-10 flex items-center justify-center gap-3 md:gap-4">
            {clients.map((c, n) => (
              <li key={c.name}>
                <span
                  title={c.name}
                  className={`flex h-12 w-12 items-center justify-center rounded-2xl border transition duration-500 md:h-14 md:w-14 ${n === now.agent ? "scale-110 border-cream/60 opacity-100 shadow-[0_0_0_4px_rgba(255,254,240,.08)]" : "border-cream/10 opacity-45"}`}
                  style={{ backgroundColor: c.bg, color: c.fg }}
                >
                  {c.path ? (
                    <svg
                      viewBox="0 0 24 24"
                      className="h-6 w-6 md:h-7 md:w-7"
                      fill="currentColor"
                      role="img"
                      aria-label={c.name}
                    >
                      <path d={c.path} />
                    </svg>
                  ) : (
                    <MousePointer2 size={24} aria-label={c.name} />
                  )}
                </span>
              </li>
            ))}
          </ul>

          <p
            aria-live="polite"
            className="mt-8 min-h-[1.25em] font-editorial text-[clamp(1.6rem,3vw,2.4rem)] leading-tight tracking-tight"
          >
            <span
              key={i}
              className="inline-block animate-[mcp-in_.5s_ease-out]"
            >
              {who.name} {now.does}
            </span>
          </p>
          <code
            key={now.tool}
            className="mt-3 inline-block animate-[mcp-in_.5s_ease-out] rounded-full border border-cream/15 bg-cream/5 px-3.5 py-1.5 font-mono text-xs text-cream/75"
          >
            {now.tool}
          </code>
          <style>{`@keyframes mcp-in{from{opacity:0;transform:translateY(6px)}to{opacity:1;transform:none}}@media (prefers-reduced-motion:reduce){[class*="mcp-in"]{animation:none!important}}`}</style>

          <p className="mx-auto mt-12 max-w-[560px] text-sm leading-relaxed text-cream/70 md:text-base">
            43 tools over MCP. Claude Code, Codex, Cursor and VS Code connect
            with a Xem API key, over HTTPS or locally from npm. CSV imports
            preview first, and drafts stay drafts until you schedule them.
          </p>
          <nav
            aria-label="MCP links"
            className="mt-8 flex flex-wrap items-center justify-center gap-x-8 gap-y-3 text-sm font-medium"
          >
            {[
              ["Explore the Xem MCP server", "/mcp"],
              ["What your agent can do", "/mcp#workflows"],
              ["MCP setup", "/mcp#setup"],
              ["Read the docs", docsUrl],
            ].map(([label, href]) => (
              <Link
                key={label}
                href={href}
                className="inline-flex items-center gap-1.5 hover:underline"
              >
                {label} <ArrowRight size={14} aria-hidden="true" />
              </Link>
            ))}
          </nav>
        </div>
      </div>
    </section>
  );
}
