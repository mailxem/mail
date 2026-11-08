"use client";
import { useEffect, useMemo, useRef, useState } from "react";
import { Bar, BarChart, ResponsiveContainer, Tooltip, XAxis } from "recharts";
import { Eye, MousePointerClick, UserPlus, Users } from "lucide-react";
import { useMarketingQuery } from "@/lib/marketing/api";
import type { Newsletter } from "@/lib/marketing/types";
import { Metric, Modal, QueryState } from "./shared";

type Count = { key: string; views: number; visitors: number };
type Device = "desktop" | "tablet" | "mobile";
type Report = {
  days: number;
  views: number;
  visitors: number;
  clicks: number;
  signups: number;
  daily: Count[];
  pages: Count[];
  referrers: Count[];
  devices: Count[];
  links: { link: string; clicks: number }[];
  heatmap: {
    path: string;
    device: Device;
    clicks: number;
    height: number;
    cols: number;
    rowPx: number;
    cells: { X: number; Y: number; N: number }[];
  };
};

const FRAME: Record<Device, number> = { desktop: 1280, tablet: 820, mobile: 390 };
const pct = (n: number, of: number) => (of ? `${Math.round((n / of) * 1000) / 10}%` : "—");

/** A newsletter website's traffic, sign-ups and click heatmap. */
export function SiteAnalytics({
  newsletter,
  url,
  close,
}: {
  newsletter: Newsletter;
  url: string;
  close: () => void;
}) {
  const [days, setDays] = useState(30);
  const [path, setPath] = useState("");
  const [device, setDevice] = useState<Device>("desktop");
  const q = new URLSearchParams({ days: String(days), device, ...(path ? { path } : {}) });
  const report = useMarketingQuery<Report>(`marketing/newsletters/${newsletter.id}/site/analytics?${q}`);
  const r = report.data;

  return (
    <Modal
      open
      onOpenChange={close}
      title="Website analytics"
      description={`${newsletter.name}: who visits the website, where they come from, and what they click.`}
      wide
    >
      <div className="space-y-6">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <Segmented
            label="Period"
            value={String(days)}
            options={[["7", "7 days"], ["30", "30 days"], ["90", "90 days"]]}
            onChange={(v) => setDays(Number(v))}
          />
          <a href={url} target="_blank" rel="noreferrer" className="text-xs text-muted-foreground hover:text-foreground hover:underline">
            {url.replace(/^https?:\/\//, "")}
          </a>
        </div>
        {!r ? (
          <QueryState loading={report.isLoading} error={report.error} retry={() => void report.refetch()} />
        ) : (
          <>
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
              <Metric label="Visitors" value={r.visitors.toLocaleString()} icon={<Users size={18} />} detail={`${r.views.toLocaleString()} page views`} />
              <Metric label="Sign-ups" value={r.signups.toLocaleString()} icon={<UserPlus size={18} />} detail={`${pct(r.signups, r.visitors)} of visitors`} />
              <Metric label="Clicks" value={r.clicks.toLocaleString()} icon={<MousePointerClick size={18} />} detail={`${r.views ? (r.clicks / r.views).toFixed(1) : "0"} per page view`} />
              <Metric label="Pages per visit" value={r.visitors ? (r.views / r.visitors).toFixed(1) : "—"} icon={<Eye size={18} />} detail={`Last ${r.days} days`} />
            </div>

            <section className="rounded-xl border p-4">
              <h3 className="text-sm font-semibold">Visitors by day</h3>
              {r.daily.length ? (
                <div className="mt-3 h-40">
                  <ResponsiveContainer width="100%" height="100%">
                    <BarChart data={r.daily} margin={{ top: 4, right: 0, left: 0, bottom: 0 }}>
                      <XAxis dataKey="key" tickFormatter={(d: string) => d.slice(5)} tick={{ fontSize: 11 }} tickLine={false} axisLine={false} minTickGap={24} />
                      <Tooltip cursor={{ fill: "var(--muted)" }} contentStyle={{ fontSize: 12, borderRadius: 8 }} formatter={(v: number, k: string) => [v, k === "visitors" ? "Visitors" : "Page views"]} />
                      <Bar dataKey="visitors" fill="var(--chart-1, #6366f1)" radius={[4, 4, 0, 0]} />
                    </BarChart>
                  </ResponsiveContainer>
                </div>
              ) : (
                <p className="mt-2 text-sm text-muted-foreground">No visits yet. Share your website link and they’ll show up here.</p>
              )}
            </section>

            <div className="grid gap-4 lg:grid-cols-3">
              <Ranked title="Top pages" rows={r.pages.map((p) => ({ label: p.key === "/" ? "Home" : p.key, value: p.views, active: p.key === r.heatmap.path, onClick: () => setPath(p.key) }))} empty="No page views yet." />
              <Ranked title="Where visitors come from" rows={r.referrers.map((p) => ({ label: p.key, value: p.visitors }))} empty="Visitors so far came directly or from email." />
              <Ranked title="Devices" rows={r.devices.map((p) => ({ label: p.key.charAt(0).toUpperCase() + p.key.slice(1), value: p.visitors }))} empty="No visits yet." />
            </div>
            {r.links.length > 0 && (
              <Ranked title="Most-clicked links" rows={r.links.map((l) => ({ label: l.link.replace(/^https?:\/\//, ""), value: l.clicks }))} empty="" />
            )}

            <section className="space-y-3 rounded-xl border p-4">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <h3 className="text-sm font-semibold">Click heatmap</h3>
                  <p className="text-xs text-muted-foreground">
                    {r.heatmap.path === "/" ? "Home page" : r.heatmap.path} · {r.heatmap.clicks.toLocaleString()} clicks. Pick a page in Top pages to switch.
                  </p>
                </div>
                <Segmented
                  label="Device"
                  value={device}
                  options={[["desktop", "Desktop"], ["tablet", "Tablet"], ["mobile", "Mobile"]]}
                  onChange={(v) => setDevice(v as Device)}
                />
              </div>
              <Heatmap url={url + (r.heatmap.path === "/" ? "" : r.heatmap.path)} heatmap={r.heatmap} device={device} />
            </section>
          </>
        )}
      </div>
    </Modal>
  );
}

function Segmented({ label, value, options, onChange }: { label: string; value: string; options: [string, string][]; onChange: (v: string) => void }) {
  return (
    <div role="radiogroup" aria-label={label} className="inline-flex rounded-lg border bg-muted/50 p-0.5">
      {options.map(([v, text]) => (
        <button
          key={v}
          type="button"
          role="radio"
          aria-checked={value === v}
          onClick={() => onChange(v)}
          className={`rounded-md px-3 py-1 text-xs font-medium transition ${value === v ? "bg-background shadow-sm" : "text-muted-foreground hover:text-foreground"}`}
        >
          {text}
        </button>
      ))}
    </div>
  );
}

function Ranked({ title, rows, empty }: { title: string; rows: { label: string; value: number; active?: boolean; onClick?: () => void }[]; empty: string }) {
  const top = Math.max(1, ...rows.map((r) => r.value));
  return (
    <section className="rounded-xl border p-4">
      <h3 className="text-sm font-semibold">{title}</h3>
      {!rows.length ? (
        <p className="mt-2 text-sm text-muted-foreground">{empty}</p>
      ) : (
        <ul className="mt-3 space-y-1.5">
          {rows.map((row) => {
            const inner = (
              <>
                <span className="absolute inset-y-0 left-0 rounded-md bg-primary/10" style={{ width: `${(row.value / top) * 100}%` }} aria-hidden />
                <span className="relative truncate">{row.label}</span>
                <span className="relative tabular-nums text-muted-foreground">{row.value.toLocaleString()}</span>
              </>
            );
            const cls = `relative flex w-full items-center justify-between gap-3 rounded-md px-2 py-1 text-left text-sm ${row.active ? "ring-1 ring-primary/40" : ""}`;
            return (
              <li key={row.label}>
                {row.onClick ? (
                  <button type="button" onClick={row.onClick} className={`${cls} hover:bg-muted/60`} aria-pressed={row.active}>
                    {inner}
                  </button>
                ) : (
                  <div className={cls}>{inner}</div>
                )}
              </li>
            );
          })}
        </ul>
      )}
    </section>
  );
}

// Blue through red, by click density.
const PALETTE = [
  [37, 99, 235],
  [6, 182, 212],
  [132, 204, 22],
  [250, 204, 21],
  [239, 68, 68],
];
function color(t: number) {
  const i = Math.min(PALETTE.length - 2, Math.floor(t * (PALETTE.length - 1)));
  const f = t * (PALETTE.length - 1) - i;
  return PALETTE[i].map((c, k) => Math.round(c + (PALETTE[i + 1][k] - c) * f));
}

/** The live page in a frame, with click density painted over it. */
function Heatmap({ url, heatmap, device }: { url: string; heatmap: Report["heatmap"]; device: Device }) {
  const box = useRef<HTMLDivElement>(null);
  const canvas = useRef<HTMLCanvasElement>(null);
  const [width, setWidth] = useState(0);
  const frame = FRAME[device];
  const height = Math.min(Math.max(heatmap.height || 0, 1600), 9000);
  const scale = width ? Math.min(1, width / frame) : 0;

  useEffect(() => {
    const el = box.current;
    if (!el) return;
    const ro = new ResizeObserver(([e]) => setWidth(e.contentRect.width));
    ro.observe(el);
    return () => ro.disconnect();
  }, []);

  const max = useMemo(() => Math.max(1, ...heatmap.cells.map((c) => c.N)), [heatmap.cells]);
  useEffect(() => {
    const c = canvas.current;
    const ctx = c?.getContext("2d", { willReadFrequently: true });
    if (!c || !ctx) return;
    ctx.clearRect(0, 0, frame, height);
    const radius = device === "mobile" ? 22 : 34;
    for (const cell of heatmap.cells) {
      const x = ((cell.X + 0.5) / heatmap.cols) * frame;
      const y = (cell.Y + 0.5) * heatmap.rowPx;
      const g = ctx.createRadialGradient(x, y, 0, x, y, radius);
      g.addColorStop(0, `rgba(0,0,0,${0.25 + 0.75 * (cell.N / max)})`);
      g.addColorStop(1, "rgba(0,0,0,0)");
      ctx.fillStyle = g;
      ctx.fillRect(x - radius, y - radius, radius * 2, radius * 2);
    }
    const img = ctx.getImageData(0, 0, frame, height);
    for (let i = 3; i < img.data.length; i += 4) {
      const a = img.data[i];
      if (!a) continue;
      const [r, g, b] = color(Math.min(1, a / 255));
      img.data[i - 3] = r;
      img.data[i - 2] = g;
      img.data[i - 1] = b;
      img.data[i] = Math.min(220, a * 1.4);
    }
    ctx.putImageData(img, 0, 0);
  }, [heatmap, frame, height, max, device]);

  return (
    <div ref={box} className="max-h-[640px] overflow-y-auto overflow-x-hidden rounded-lg border bg-muted/30">
      <div className="relative mx-auto" style={{ width: frame * scale, height: height * scale }}>
        <div className="absolute left-0 top-0 origin-top-left" style={{ width: frame, height, transform: `scale(${scale})` }}>
          <iframe
            key={url + device}
            src={url}
            title="Your website"
            className="pointer-events-none absolute inset-0 size-full border-0 bg-white"
            style={{ width: frame, height }}
            loading="lazy"
          />
          <canvas ref={canvas} width={frame} height={height} className="pointer-events-none absolute inset-0" aria-hidden />
        </div>
      </div>
      {!heatmap.clicks && (
        <p className="sticky bottom-0 bg-background/90 px-4 py-2 text-center text-xs text-muted-foreground">
          No clicks recorded on this page and device yet.
        </p>
      )}
    </div>
  );
}
