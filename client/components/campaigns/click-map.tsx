"use client";
import { useMemo, useRef, useState } from "react";

type LinkStat = { url: string; uniqueClicks: number; clickRate: number };

// Light blue for a few clicks through to red for the most-clicked link.
const HEAT = ["#60a5fa", "#22d3ee", "#a3e635", "#facc15", "#fb923c", "#ef4444"];
const heat = (t: number) =>
  HEAT[Math.min(HEAT.length - 1, Math.floor(t * HEAT.length))];

/**
 * The email as sent, with every link outlined by how many people clicked it.
 * Links are matched by their exact address, as click tracking records them.
 * The frame has no scripts and its links don't navigate.
 */
export function ClickMap({ html, links }: { html: string; links: LinkStat[] }) {
  const frame = useRef<HTMLIFrameElement>(null);
  const [height, setHeight] = useState(600);
  const { doc, unmatched } = useMemo(() => {
    const byUrl = new Map(links.map((l) => [l.url, l]));
    const top = Math.max(1, ...links.map((l) => l.uniqueClicks));
    const page = new DOMParser().parseFromString(html, "text/html");
    page.querySelectorAll("script").forEach((s) => s.remove());
    const seen = new Set<string>();
    page.querySelectorAll<HTMLAnchorElement>("a[href]").forEach((a) => {
      const url = a.getAttribute("href") || "";
      if (!/^https?:\/\//i.test(url) || /unsubscribe/i.test(url)) return;
      const stat = byUrl.get(url);
      seen.add(url);
      const color = stat ? heat(stat.uniqueClicks / top) : "#9ca3af";
      a.setAttribute(
        "style",
        `${a.getAttribute("style") || ""};position:relative;outline:3px ${stat ? "solid" : "dashed"} ${color};outline-offset:2px;border-radius:2px`,
      );
      const badge = page.createElement("span");
      badge.textContent = stat
        ? `${stat.uniqueClicks} · ${stat.clickRate.toFixed(1)}%`
        : "0";
      badge.setAttribute(
        "style",
        `position:absolute;top:-12px;right:-8px;z-index:10;padding:1px 6px;border-radius:999px;background:${color};color:#111;font:600 11px/16px system-ui,sans-serif;white-space:nowrap;text-decoration:none;box-shadow:0 1px 2px rgba(0,0,0,.25)`,
      );
      a.appendChild(badge);
    });
    const style = page.createElement("style");
    style.textContent = "a{pointer-events:none;cursor:default}";
    page.head.appendChild(style);
    return {
      doc: "<!doctype html>" + page.documentElement.outerHTML,
      unmatched: links.filter((l) => !seen.has(l.url) && l.uniqueClicks > 0),
    };
  }, [html, links]);

  return (
    <div className="space-y-3">
      <div className="flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
        <span>Fewer clicks</span>
        <span
          className="h-2 w-28 rounded-full"
          style={{ background: `linear-gradient(to right, ${HEAT.join(",")})` }}
          aria-hidden
        />
        <span>More clicks</span>
        <span className="ml-auto">
          Each badge: people who clicked · share of recipients
        </span>
      </div>
      <div className="overflow-hidden rounded-lg border bg-white">
        <iframe
          ref={frame}
          title="Click map"
          srcDoc={doc}
          sandbox="allow-same-origin"
          className="block w-full border-0"
          style={{ height }}
          onLoad={() => {
            // The document is at least as tall as the frame, so measure its content.
            const body = frame.current?.contentDocument?.body;
            if (!body) return;
            const bottom = Math.max(
              0,
              ...Array.from(
                body.children,
                (el) => el.getBoundingClientRect().bottom,
              ),
            );
            setHeight(Math.min(Math.max(Math.ceil(bottom) + 24, 200), 6000));
          }}
        />
      </div>
      {unmatched.length > 0 && (
        <p className="text-xs text-muted-foreground">
          Also clicked, but not found in this version of the email (personalised
          or edited links):{" "}
          {unmatched
            .slice(0, 5)
            .map(
              (l) => `${l.url.replace(/^https?:\/\//, "")} (${l.uniqueClicks})`,
            )
            .join(", ")}
          {unmatched.length > 5 ? `, and ${unmatched.length - 5} more` : ""}.
        </p>
      )}
    </div>
  );
}
