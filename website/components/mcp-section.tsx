import Link from "next/link";
import { ArrowUpRight, Terminal } from "lucide-react";

export function McpSection() {
  return (
    <section
      aria-labelledby="mcp-heading"
      className="px-5 py-16 md:px-10 md:py-24"
    >
      <div className="mx-auto grid max-w-[1160px] items-center gap-10 rounded-[28px] bg-[#ebe7f2] p-7 md:grid-cols-2 md:p-12">
        <div>
          <p className="text-[11px] font-medium uppercase tracking-[0.17em] text-iris">
            Xem × Model Context Protocol
          </p>
          <h2
            id="mcp-heading"
            className="mt-5 font-editorial text-4xl leading-[1.05] tracking-tight md:text-5xl"
          >
            Run Xem from
            <br />
            <em>Claude, Cursor or Codex.</em>
          </h2>
          <p className="mt-5 max-w-md text-sm leading-relaxed text-muted">
            43 tools over MCP: draft newsletters, import contacts, read campaign results. Connect over HTTPS, or run it locally from npm.
          </p>
          <p className="mt-5 text-xs font-medium text-muted">Works with</p>
          <ul className="mt-2 flex flex-wrap gap-2 text-sm">
            {["Claude", "Cursor", "Codex", "VS Code"].map((app) => (
              <li
                key={app}
                className="rounded-full border border-ink/15 bg-cream px-3 py-1.5"
              >
                {app}
              </li>
            ))}
            <li className="px-1 py-1.5 text-muted">and any MCP client</li>
          </ul>
          <Link
            href="/mcp"
            className="mt-7 inline-flex items-center gap-2 border-b border-ink pb-1 text-sm font-medium"
          >
            Explore the Xem MCP server <ArrowUpRight size={17} />
          </Link>
        </div>
        <div className="rounded-2xl border border-ink/10 bg-cream p-6 shadow-sm">
          <p className="flex items-center gap-2 text-xs font-medium text-muted">
            <Terminal size={16} /> A workflow you can ask for
          </p>
          <blockquote className="mt-6 font-editorial text-2xl leading-snug">
            “Preview this contact CSV, map the email and first-name columns,
            then help me draft our next newsletter.”
          </blockquote>
          <div className="mt-7 flex flex-wrap gap-2 text-xs">
            {["Map columns", "Preview import", "Create a draft"].map((step) => (
              <span
                key={step}
                className="rounded-full border border-ink/15 px-3 py-2"
              >
                {step}
              </span>
            ))}
          </div>
          <p className="mt-5 text-xs leading-relaxed text-muted">
            CSV imports preview first. Newsletter drafts stay drafts until you
            schedule them.
          </p>
        </div>
      </div>
    </section>
  );
}
