"use client";
import { useMemo, useState } from "react";
import {
  Area,
  AreaChart,
  Bar,
  BarChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { ArrowUpRight, BarChart3, LayoutGrid, Mail } from "lucide-react";
import { appLink } from "@/lib/site";

// The Xem dashboard (Overview and Email tabs) on sample data. Numbers come
// from a fixed sequence, so every visit shows the same demo.
const RANGES = [7, 30, 90] as const;
type Range = (typeof RANGES)[number];

const wave = (i: number, seed: number) =>
  0.55 +
  0.45 * Math.abs(Math.sin(i * 1.7 + seed) * Math.cos(i * 0.6 + seed / 2));

function series(days: Range, offset = 0) {
  const start = new Date(Date.UTC(2026, 9, 31 - days - offset));
  return Array.from({ length: days }, (_, i) => {
    const d = new Date(start);
    d.setUTCDate(start.getUTCDate() + i + 1);
    const k = i + offset * 3;
    // Weekly issue on Sundays, smaller sends in between.
    const big = d.getUTCDay() === 0;
    const sent = Math.round((big ? 2150 : 180) * wave(k, 1));
    return {
      date: d.toISOString().slice(0, 10),
      sent,
      pending: Math.round(sent * 0.02 * wave(k, 3)),
      failed: Math.round(sent * 0.006 * wave(k, 5)),
      opened: Math.round(sent * (0.42 + 0.1 * wave(k, 7))),
      clicked: Math.round(sent * (0.05 + 0.03 * wave(k, 9))),
      unsubscribed: Math.round(sent * 0.0016 * wave(k, 11)),
      subscribers: 0,
    };
  });
}

function totals(rows: ReturnType<typeof series>) {
  const sum = (
    k: "sent" | "pending" | "failed" | "opened" | "clicked" | "unsubscribed",
  ) => rows.reduce((n, r) => n + r[k], 0);
  const sent = sum("sent");
  return {
    sent,
    pending: sum("pending"),
    failed: sum("failed"),
    unsubscribed: sum("unsubscribed"),
    openRate: sent ? sum("opened") / sent : 0,
    clickRate: sent ? sum("clicked") / sent : 0,
  };
}

const number = (n: number) => n.toLocaleString("en-US");
const compact = (n: number) =>
  Intl.NumberFormat("en", { notation: "compact" }).format(n);
const pct = (r: number) => `${(r * 100).toFixed(1)}%`;
const shortDate = (v: string) =>
  new Date(`${v}T12:00:00Z`).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
  });

function change(
  now: number,
  before: number,
  kind: "count" | "rate",
  good: "up" | "down" = "up",
) {
  if (!before) return null;
  const v =
    kind === "rate" ? (now - before) * 100 : ((now - before) / before) * 100;
  const r = Math.round(v * 10) / 10;
  return {
    label: `${r >= 0 ? "+" : ""}${r.toFixed(1)}${kind === "rate" ? " pp" : "%"}`,
    tone:
      r === 0
        ? "text-muted"
        : r > 0 === (good === "up")
          ? "text-[#15803d]"
          : "text-[#b91c1c]",
  };
}

function Kpi({
  label,
  value,
  delta,
  note,
}: {
  label: string;
  value: string;
  delta?: ReturnType<typeof change>;
  note?: string;
}) {
  return (
    <div
      className="min-w-0 rounded-[14px] border border-ink/10 bg-white px-4 py-3.5"
      data-stat={label}
    >
      <p className="text-[12px] text-muted">{label}</p>
      <p data-value className="mt-0.5 text-[20px] font-medium tabular-nums">
        {value}
      </p>
      <p
        className={`text-[12px] tabular-nums ${delta ? delta.tone : "text-muted"}`}
      >
        {delta ? `${delta.label} vs. previous ${note}` : note}
      </p>
    </div>
  );
}

const campaigns = [
  ["The Sunday Edit #42", 2184, 0.081],
  ["Welcome to the Sunday Edit", 412, 0.124],
  ["The Sunday Edit #41", 2160, 0.071],
  ["Spring reading list", 1730, 0.066],
  ["The Sunday Edit #40", 2141, 0.058],
] as const;

export function DashboardDemo() {
  const [days, setDays] = useState<Range>(30);
  const [tab, setTab] = useState<"overview" | "email">("overview");
  const { rows, now, before, growth } = useMemo(() => {
    const rows = series(days);
    const prev = series(days, days);
    // Subscribers: a steady climb with dips on unsubscribes.
    let active = 8900 - days * 6;
    for (const r of rows) {
      active += Math.round(14 * wave(rows.indexOf(r), 13)) - r.unsubscribed;
      r.subscribers = active;
    }
    return {
      rows,
      now: totals(rows),
      before: totals(prev),
      growth: active - (8900 - days * 6),
    };
  }, [days]);
  const label = `${days} days`;

  return (
    <div className="overflow-hidden rounded-[28px] border border-ink/10 bg-[#fcfcfa] text-ink shadow-[0_20px_60px_-40px_rgba(0,0,0,.35)]">
      <div
        className="flex flex-wrap items-center gap-1 border-b border-ink/10 px-5 pt-3 md:px-7"
        role="tablist"
        aria-label="Dashboard view"
      >
        {(
          [
            ["overview", "Overview", LayoutGrid],
            ["email", "Email", Mail],
          ] as const
        ).map(([id, name, Icon]) => (
          <button
            key={id}
            role="tab"
            aria-selected={tab === id}
            onClick={() => setTab(id)}
            className={`-mb-px flex items-center gap-1.5 border-b-2 px-3 py-2.5 text-[13px] ${tab === id ? "border-ink font-medium" : "border-transparent text-muted hover:text-ink"}`}
          >
            <Icon size={14} aria-hidden="true" />
            {name}
          </button>
        ))}
        <span className="ml-auto pb-2 text-[11px] text-muted">Sample data</span>
      </div>

      <div className="space-y-5 p-5 md:p-7">
        <div className="flex flex-wrap items-end justify-between gap-4">
          <div>
            <p className="text-[17px] font-semibold tracking-[-.01em]">
              Good morning
            </p>
            <p className="mt-0.5 text-[13px] text-muted">
              How email is doing across The Sunday Edit.
            </p>
          </div>
          <div className="flex items-center gap-2.5">
            <div
              role="group"
              aria-label="Period"
              className="inline-flex rounded-[9px] border border-ink/10 bg-white p-[3px]"
            >
              {RANGES.map((n) => (
                <button
                  key={n}
                  type="button"
                  aria-pressed={days === n}
                  onClick={() => setDays(n)}
                  className={`rounded-md px-2.5 py-1 text-[12px] ${days === n ? "bg-lavender text-ink" : "text-muted hover:text-ink"}`}
                >
                  {n} days
                </button>
              ))}
            </div>
            <a
              href={appLink("/auth/register")}
              className="hidden whitespace-nowrap rounded-[9px] bg-ink px-3.5 py-[7px] text-[13px] font-medium text-cream sm:inline-block"
            >
              New campaign
            </a>
          </div>
        </div>

        {tab === "overview" ? (
          <>
            <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
              <Kpi
                label="Sent"
                value={compact(now.sent)}
                delta={change(now.sent, before.sent, "count")}
                note={label}
              />
              <Kpi
                label="Opened"
                value={pct(now.openRate)}
                delta={change(now.openRate, before.openRate, "rate")}
                note={label}
              />
              <Kpi
                label="Clicked"
                value={pct(now.clickRate)}
                delta={change(now.clickRate, before.clickRate, "rate")}
                note={label}
              />
              <Kpi
                label="Unsubscribed"
                value={number(now.unsubscribed)}
                delta={change(
                  now.unsubscribed,
                  before.unsubscribed,
                  "count",
                  "down",
                )}
                note={label}
              />
            </div>
            <div className="grid gap-3 lg:grid-cols-[1.7fr_1fr]">
              <div className="min-w-0 rounded-[14px] border border-ink/10 bg-white p-4">
                <p className="text-[12px] text-muted">Subscribers</p>
                <p className="mb-2 mt-0.5 flex items-baseline gap-2.5">
                  <strong className="text-[24px] font-medium tabular-nums">
                    {number(rows[rows.length - 1].subscribers)}
                  </strong>
                  <span
                    className={`text-[12px] ${growth >= 0 ? "text-[#15803d]" : "text-[#b91c1c]"}`}
                  >
                    {growth >= 0 ? "+" : ""}
                    {number(growth)} in {days} days
                  </span>
                </p>
                <div className="h-[240px]">
                  <ResponsiveContainer width="100%" height="100%">
                    <AreaChart
                      data={rows}
                      margin={{ left: 0, right: 4, top: 8, bottom: 0 }}
                    >
                      <defs>
                        <linearGradient
                          id="demoSubs"
                          x1="0"
                          y1="0"
                          x2="0"
                          y2="1"
                        >
                          <stop
                            offset="0%"
                            stopColor="#5b3cc4"
                            stopOpacity={0.25}
                          />
                          <stop
                            offset="100%"
                            stopColor="#5b3cc4"
                            stopOpacity={0}
                          />
                        </linearGradient>
                      </defs>
                      <CartesianGrid
                        vertical={false}
                        strokeDasharray="2 4"
                        stroke="#e5e5df"
                      />
                      <XAxis
                        dataKey="date"
                        tickLine={false}
                        axisLine={false}
                        minTickGap={32}
                        tickFormatter={shortDate}
                        tick={{ fontSize: 11, fill: "#66695e" }}
                      />
                      <YAxis
                        width={40}
                        tickLine={false}
                        axisLine={false}
                        tickFormatter={compact}
                        domain={["dataMin - 100", "auto"]}
                        tick={{ fontSize: 11, fill: "#66695e" }}
                      />
                      <Tooltip
                        labelFormatter={(v) => shortDate(String(v))}
                        formatter={(v: number) => [number(v), "Subscribers"]}
                        contentStyle={{ fontSize: 12, borderRadius: 10 }}
                      />
                      <Area
                        dataKey="subscribers"
                        type="monotone"
                        stroke="#5b3cc4"
                        strokeWidth={2}
                        fill="url(#demoSubs)"
                        isAnimationActive={false}
                      />
                    </AreaChart>
                  </ResponsiveContainer>
                </div>
              </div>
              <div className="flex min-w-0 flex-col rounded-[14px] border border-ink/10 bg-white p-4">
                <p className="flex justify-between text-[12px] text-muted">
                  <span>Top campaigns</span>
                  <span>Clicked</span>
                </p>
                <ol className="mt-2.5">
                  {campaigns.map(([name, sent, rate]) => (
                    <li
                      key={name}
                      className="grid grid-cols-[minmax(0,1fr)_auto] gap-x-3 rounded-lg px-2 py-2.5 hover:bg-ink/[.03]"
                    >
                      <span className="truncate text-[13.5px]">{name}</span>
                      <span className="row-span-2 self-center text-[13.5px] tabular-nums">
                        {pct(rate)}
                      </span>
                      <span className="text-[12px] text-muted">
                        {compact(sent)} sent
                      </span>
                    </li>
                  ))}
                </ol>
                <span className="mt-auto inline-flex items-center gap-1 pt-3 text-[13px] text-muted">
                  All campaign stats{" "}
                  <ArrowUpRight size={13} aria-hidden="true" />
                </span>
              </div>
            </div>
          </>
        ) : (
          <>
            <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
              <Kpi
                label="Emails sent"
                value={number(now.sent)}
                note="Accepted by SMTP in this period."
              />
              <Kpi
                label="Pending emails"
                value={number(now.pending)}
                note="Queued, scheduled or sending."
              />
              <Kpi
                label="Click rate"
                value={pct(now.clickRate)}
                note="Bot clicks left out."
              />
              <Kpi
                label="Failed emails"
                value={number(now.failed)}
                note="Current failed status."
              />
            </div>
            <div className="grid gap-3 lg:grid-cols-[1.7fr_1fr]">
              <div className="min-w-0 rounded-[14px] border border-ink/10 bg-white p-4">
                <p className="text-[14px] font-medium">Email activity</p>
                <p className="text-[12px] text-muted">
                  Your sending volume, day by day.
                </p>
                <div className="mt-3 h-[250px]">
                  <ResponsiveContainer width="100%" height="100%">
                    <BarChart
                      data={rows}
                      margin={{ left: 0, right: 4, top: 8, bottom: 0 }}
                    >
                      <CartesianGrid
                        vertical={false}
                        strokeDasharray="2 4"
                        stroke="#e5e5df"
                      />
                      <XAxis
                        dataKey="date"
                        tickLine={false}
                        axisLine={false}
                        minTickGap={28}
                        tickFormatter={shortDate}
                        tick={{ fontSize: 11, fill: "#66695e" }}
                      />
                      <YAxis
                        width={40}
                        tickLine={false}
                        axisLine={false}
                        tickFormatter={compact}
                        tick={{ fontSize: 11, fill: "#66695e" }}
                      />
                      <Tooltip
                        labelFormatter={(v) => shortDate(String(v))}
                        contentStyle={{ fontSize: 12, borderRadius: 10 }}
                      />
                      <Bar
                        dataKey="sent"
                        name="Sent"
                        stackId="a"
                        fill="#164b3f"
                        isAnimationActive={false}
                      />
                      <Bar
                        dataKey="pending"
                        name="Pending"
                        stackId="a"
                        fill="#5b3cc4"
                        isAnimationActive={false}
                      />
                      <Bar
                        dataKey="failed"
                        name="Failed"
                        stackId="a"
                        fill="#b91c1c"
                        radius={[3, 3, 0, 0]}
                        isAnimationActive={false}
                      />
                    </BarChart>
                  </ResponsiveContainer>
                </div>
              </div>
              <div className="min-w-0 rounded-[14px] border border-ink/10 bg-white p-4">
                <p className="flex items-center gap-1.5 text-[14px] font-medium">
                  <BarChart3 size={14} aria-hidden="true" /> At a glance
                </p>
                <p className="text-[12px] text-muted">
                  Across API, outbox and marketing.
                </p>
                <dl className="mt-3 divide-y divide-ink/10 text-[13px]">
                  {[
                    ["Open rate", pct(now.openRate)],
                    [
                      "Known bounces",
                      `${number(now.failed)} (${pct(now.failed / Math.max(1, now.sent))})`,
                    ],
                    [
                      "Campaign-linked sends",
                      number(Math.round(now.sent * 0.82)),
                    ],
                    [
                      "Other email sends",
                      number(now.sent - Math.round(now.sent * 0.82)),
                    ],
                    ["Drafts", "2"],
                  ].map(([k, v]) => (
                    <div key={k} className="flex justify-between py-2.5">
                      <dt className="text-muted">{k}</dt>
                      <dd className="tabular-nums">{v}</dd>
                    </div>
                  ))}
                </dl>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
