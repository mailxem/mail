"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from "recharts";
import { ArrowUpRight } from "lucide-react";
import { useTeam } from "@/app/providers/team-provider";
import { useAnalytics } from "@/lib/analytics/use-analytics";
import {
  AnalyticsPage,
  BreakdownRow,
  Rate,
  Report,
  formatRate,
} from "@/lib/analytics/types";
import { QueryState } from "@/components/marketing/shared";
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import styles from "./overview-home.module.css";

const RANGES = [7, 30, 90] as const;
const number = (n: number) => n.toLocaleString();
const compact = (n: number) =>
  Intl.NumberFormat("en", { notation: "compact" }).format(n);

function day(offset: number) {
  const d = new Date();
  d.setUTCHours(0, 0, 0, 0);
  d.setUTCDate(d.getUTCDate() + offset);
  return d.toISOString().slice(0, 10);
}

/** Percent change for counts, percentage points for rates. `good` says which direction is good news. */
export function delta(
  current: number | null,
  previous: number | null,
  kind: "count" | "rate",
  good: "up" | "down" = "up",
) {
  if (current == null || previous == null) return null;
  if (kind === "count" && previous === 0)
    return { label: current === 0 ? "No change" : "New this period", tone: "flat" } as const;
  const diff =
    kind === "rate"
      ? (current - previous) * 100
      : previous === 0
        ? null
        : ((current - previous) / previous) * 100;
  if (diff == null || !Number.isFinite(diff)) return null;
  const rounded = Math.round(diff * 10) / 10;
  return {
    label: `${rounded >= 0 ? "+" : ""}${rounded.toFixed(1)}${kind === "rate" ? " pp" : "%"}`,
    tone:
      rounded === 0
        ? "flat"
        : (rounded > 0) === (good === "up")
          ? "good"
          : "bad",
  } as const;
}

export function OverviewHome() {
  const [days, setDays] = useState<(typeof RANGES)[number]>(30);
  const { team, loading, error, refreshTeam } = useTeam();
  const params = useMemo(
    () =>
      new URLSearchParams({
        from: day(-days + 1),
        to: day(1),
        timezone: "UTC",
      }),
    [days],
  );
  const report = useAnalytics<Report>("report", params);
  const top = useAnalytics<AnalyticsPage<BreakdownRow>>(
    "breakdown",
    useMemo(() => {
      const q = new URLSearchParams(params);
      q.set("kind", "campaigns");
      q.set("limit", "5");
      q.set("sort", "accepted");
      q.set("direction", "desc");
      return q;
    }, [params]),
  );
  const r = report.data;

  return (
    <section className={styles.overview} aria-labelledby="overview-title">
      <header className={styles.head}>
        <div>
          <p className={styles.eyebrow}>{team?.name ?? "Your workspace"}</p>
          <h1 id="overview-title" className={styles.title}>
            {greeting()}
          </h1>
        </div>
        <div className={styles.headActions}>
          <div className={styles.range} role="group" aria-label="Period">
            {RANGES.map((n) => (
              <button
                key={n}
                type="button"
                aria-pressed={days === n}
                onClick={() => setDays(n)}
              >
                {n} days
              </button>
            ))}
          </div>
          <Link href="/campaigns/new" className={styles.primary}>
            New campaign
          </Link>
        </div>
      </header>

      <QueryState
        loading={loading || (!!team && report.isPending)}
        error={error || report.error}
        retry={() => void (error ? refreshTeam() : report.refetch())}
      />

      {r && (
        <>
          <div className={styles.kpis}>
            <Kpi
              label="Sent"
              value={compact(r.summary.accepted)}
              change={delta(r.summary.accepted, r.previous.accepted, "count")}
            />
            <Kpi
              label="Opened"
              value={formatRate(r.summary.openRate)}
              change={rateDelta(r.summary.openRate, r.previous.openRate)}
            />
            <Kpi
              label="Clicked"
              value={formatRate(r.summary.clickRate)}
              change={rateDelta(r.summary.clickRate, r.previous.clickRate)}
            />
            <Kpi
              label="Unsubscribed"
              value={number(r.summary.unsubscribes)}
              change={delta(
                r.summary.unsubscribes,
                r.previous.unsubscribes,
                "count",
                "down",
              )}
            />
          </div>

          <div className={styles.grid}>
            <div className={styles.panel}>
              <div className={styles.panelHead}>
                <span>Subscribers</span>
              </div>
              <div className={styles.big}>
                <strong>{number(r.summary.activeSubscribers)}</strong>
                {r.growth.available && r.growth.net != null && (
                  <span
                    className={
                      r.growth.net >= 0 ? styles.good : styles.bad
                    }
                  >
                    {r.growth.net >= 0 ? "+" : ""}
                    {number(r.growth.net)} in {days} days
                  </span>
                )}
              </div>
              {r.growth.series.length > 1 ? (
                <ChartContainer
                  config={{
                    active: { label: "Subscribers", color: "var(--chart-1)" },
                  }}
                  className={styles.chart}
                >
                  <AreaChart
                    data={r.growth.series}
                    margin={{ left: 0, right: 0, top: 8, bottom: 0 }}
                  >
                    <defs>
                      <linearGradient id="subsFill" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="0%" stopColor="var(--color-active)" stopOpacity={0.28} />
                        <stop offset="100%" stopColor="var(--color-active)" stopOpacity={0} />
                      </linearGradient>
                    </defs>
                    <CartesianGrid vertical={false} strokeDasharray="2 4" />
                    <XAxis
                      dataKey="date"
                      tickLine={false}
                      axisLine={false}
                      minTickGap={32}
                      tickFormatter={(v: string) =>
                        new Date(`${v.slice(0, 10)}T12:00:00Z`).toLocaleDateString([], {
                          month: "short",
                          day: "numeric",
                        })
                      }
                    />
                    <YAxis
                      width={36}
                      tickLine={false}
                      axisLine={false}
                      tickFormatter={compact}
                      domain={["dataMin", "auto"]}
                    />
                    <ChartTooltip content={<ChartTooltipContent />} />
                    <Area
                      dataKey="active"
                      type="monotone"
                      stroke="var(--color-active)"
                      strokeWidth={2}
                      fill="url(#subsFill)"
                    />
                  </AreaChart>
                </ChartContainer>
              ) : (
                <div className={styles.emptyChart}>
                  <p>Your growth chart appears once people subscribe.</p>
                  <Link href="/forms/new">
                    Create a signup form <ArrowUpRight size={14} />
                  </Link>
                </div>
              )}
            </div>

            <div className={styles.panel}>
              <div className={styles.panelHead}>
                <span>Top campaigns</span>
                <span>Clicked</span>
              </div>
              {top.data?.items.length ? (
                <ol className={styles.list}>
                  {top.data.items.map((c) => (
                    <li key={c.id}>
                      <Link href={`/campaigns/${c.id}`}>
                        <span className={styles.issue}>{c.name}</span>
                        <span className={styles.meta}>
                          {compact(c.accepted)} sent
                        </span>
                        <span className={styles.rate}>
                          {formatRate(c.clickRate)}
                        </span>
                      </Link>
                    </li>
                  ))}
                </ol>
              ) : (
                <div className={styles.emptyList}>
                  <p>
                    {top.isPending
                      ? "Loading…"
                      : `No campaigns sent in the last ${days} days.`}
                  </p>
                  {!top.isPending && (
                    <Link href="/campaigns/new">
                      Write your next issue <ArrowUpRight size={14} />
                    </Link>
                  )}
                </div>
              )}
              <Link href="/analytics/campaigns" className={styles.more}>
                All campaign stats
              </Link>
            </div>
          </div>
        </>
      )}
    </section>
  );
}

function rateDelta(current: Rate, previous: Rate) {
  // Compare only when the previous period actually sent something.
  return previous.denominator > 0
    ? delta(current.value, previous.value, "rate")
    : null;
}

function greeting() {
  const h = new Date().getHours();
  return h < 12 ? "Good morning" : h < 18 ? "Good afternoon" : "Good evening";
}

function Kpi({
  label,
  value,
  change,
}: {
  label: string;
  value: string;
  change: ReturnType<typeof delta>;
}) {
  return (
    <div className={styles.kpi}>
      <span className={styles.kpiLabel}>{label}</span>
      <strong className={styles.kpiValue}>{value}</strong>
      <span className={change ? styles[change.tone] : styles.flat}>
        {change?.label ?? "No earlier data"}
      </span>
    </div>
  );
}
