"use client";

import { useEffect, useId, useMemo, useState } from "react";
import Link from "next/link";
import {
  Area,
  AreaChart,
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Pie,
  PieChart,
  XAxis,
  YAxis,
} from "recharts";
import { ArrowDownToLine, ArrowUpRight, BarChart3, Table2 } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import type { Report } from "@/lib/analytics/types";
import styles from "./analytics-surface.module.css";

const colors = ["var(--chart-1)", "var(--chart-2)", "var(--chart-5)", "var(--chart-4)", "var(--muted-foreground)"];
const heatmapColors = [
  "var(--muted)",
  "color-mix(in srgb, var(--chart-1) 22%, var(--card))",
  "color-mix(in srgb, var(--chart-1) 44%, var(--card))",
  "color-mix(in srgb, var(--chart-1) 68%, var(--card))",
  "var(--chart-1)",
];
const deviceLabels: Record<string, string> = {
  desktop: "Desktop",
  mobile: "Mobile",
  tablet: "Tablet",
  other: "Other",
  unknown: "Unknown",
};
const cohortColors: Record<string, string> = {
  recent: "var(--chart-2)",
  earlier: "var(--chart-4)",
  no_clicks: "var(--chart-5)",
  not_contacted: "var(--chart-1)",
  insufficient: "var(--muted-foreground)",
};
const labels = {
  campaigns: { label: "Campaigns" },
  newsletters: { label: "Newsletters" },
  campaignClicks: { label: "Campaigns" },
  newsletterClicks: { label: "Newsletters" },
  reached: { label: "Recipients reached" },
  clicked: { label: "Recipients who clicked" },
  active: { label: "Active subscriptions" },
  count: { label: "Recorded click events" },
};
const number = (n: number) => n.toLocaleString();
const compact = (n: number) =>
  Intl.NumberFormat("en", {
    notation: "compact",
    maximumFractionDigits: 1,
  }).format(n);
const shortDate = (date: string) => date.slice(5, 10);
const dayNames = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
function useChartAnimation() {
  const [animate, setAnimate] = useState(false);
  useEffect(() => {
    const media = window.matchMedia("(prefers-reduced-motion: reduce)");
    const update = () => setAnimate(!media.matches);
    update();
    media.addEventListener("change", update);
    return () => media.removeEventListener("change", update);
  }, []);
  return animate;
}
function EmptyChart({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-48 items-center justify-center rounded-xl border border-dashed bg-muted/20 p-8 text-center text-sm leading-relaxed text-muted-foreground">
      {children}
    </div>
  );
}

export function AnalyticsCharts({
  report,
  scope,
  tagFiltered,
}: {
  report: Report;
  scope: string;
  tagFiltered: boolean;
}) {
  const animate = useChartAnimation();
  const gradient = useId().replace(/:/g, "");
  const [metric, setMetric] = useState<"accepted" | "clicked">("accepted");
  const [hidden, setHidden] = useState<string[]>([]);
  const [table, setTable] = useState(false);
  const [device, setDevice] = useState<string | null>(null);
  const [selectedHour, setSelectedHour] = useState<{
    day: number;
    block: number;
  } | null>(null);
  const charts = report.charts;
  const keys =
    metric === "accepted"
      ? (["campaigns", "newsletters"] as const)
      : (["campaignClicks", "newsletterClicks"] as const);
  const cohortTotal = report.cohorts.reduce((sum, c) => sum + c.count, 0);
  const selectedDevice = charts?.devices.find((d) => d.device === device);
  const hourBlocks = useMemo(
    () =>
      dayNames.flatMap((_, day) =>
        Array.from({ length: 6 }, (_, block) => ({
          day,
          block,
          count: (charts?.clickHours ?? [])
            .filter((v) => v.day === day && Math.floor(v.hour / 4) === block)
            .reduce((sum, v) => sum + v.count, 0),
        })),
      ),
    [charts],
  );
  const maxHour = Math.max(1, ...hourBlocks.map((h) => h.count));
  const hourSelection =
    selectedHour &&
    hourBlocks.find(
      (h) => h.day === selectedHour.day && h.block === selectedHour.block,
    );
  const toggle = (key: string) =>
    setHidden((old) =>
      old.includes(key)
        ? old.filter((k) => k !== key)
        : old.length < 1
          ? [...old, key]
          : old,
    );
  const exportDaily = () => {
    if (!charts) return;
    const lines = [
      [
        "Date",
        "Campaigns accepted",
        "Newsletters accepted",
        "Campaign messages clicked",
        "Newsletter messages clicked",
        "Timezone",
      ],
      ...charts.volume.map((d) => [
        d.date,
        d.campaigns,
        d.newsletters,
        d.campaignClicks,
        d.newsletterClicks,
        report.timezone,
      ]),
    ];
    const csv = lines
      .map((row) =>
        row.map((v) => `"${String(v).replace(/"/g, '""')}"`).join(","),
      )
      .join("\r\n");
    const url = URL.createObjectURL(
      new Blob(["\uFEFF" + csv], { type: "text/csv;charset=utf-8" }),
    );
    const link = document.createElement("a");
    link.href = url;
    link.download = `xem-daily-analytics-${report.from.slice(0, 10)}.csv`;
    link.click();
    setTimeout(() => URL.revokeObjectURL(url), 1000);
  };
  return (
    <div className={styles.chartCollection} data-testid="live-analytics-charts">
      <div className="grid gap-5 xl:grid-cols-[minmax(0,1.8fr)_minmax(0,1fr)]">
        <Card className={styles.panel}>
          <CardHeader className={styles.panelHeader}>
            <div>
              <p className="mb-2 text-xs text-muted-foreground">
                Campaigns & newsletters
              </p>
              <CardTitle className={styles.panelTitle}>
                Sending performance
              </CardTitle>
              <CardDescription className={styles.panelDescription}>
                {metric === "accepted"
                  ? "SMTP-accepted messages"
                  : "Messages with at least one tracked click"}
                , grouped by send date. Campaigns and newsletter editions.
              </CardDescription>
            </div>
            <div
              className="flex rounded-lg bg-muted/60 p-1"
              aria-label="Sending chart metric"
            >
              {(["accepted", "clicked"] as const).map((v) => (
                <button
                  key={v}
                  aria-pressed={metric === v}
                  onClick={() => setMetric(v)}
                  className={`rounded-md px-3 py-2 text-xs transition-colors ${metric === v ? "bg-background font-medium text-foreground shadow-none" : "text-muted-foreground"}`}
                >
                  {v === "accepted" ? "Messages" : "Clicks"}
                </button>
              ))}
            </div>
          </CardHeader>
          <CardContent className={styles.panelBody}>
            {charts ? (
              <>
                <div className="mb-5 flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <span className="text-3xl font-semibold tracking-tight tabular-nums">
                      {number(
                        metric === "accepted"
                          ? report.summary.accepted
                          : report.summary.clickedMessages,
                      )}
                    </span>
                    <span className="ml-2 text-xs text-muted-foreground">
                      in selected period
                    </span>
                  </div>
                  <div className="flex gap-2">
                    {["Campaigns", "Newsletters"].map((name, i) => (
                      <button
                        key={name}
                        aria-pressed={!hidden.includes(String(i))}
                        onClick={() => toggle(String(i))}
                        className={`inline-flex min-h-8 items-center gap-2 rounded-md px-2 text-xs transition-opacity hover:bg-muted ${hidden.includes(String(i)) ? "opacity-40" : ""}`}
                      >
                        <span
                          className="h-2 w-2 rounded-full"
                          style={{ backgroundColor: colors[i] }}
                        />
                        {name}
                      </button>
                    ))}
                  </div>
                </div>
                {report.summary.accepted > 0 ? (
                  <ChartContainer config={labels} className="h-[270px] w-full">
                    <BarChart
                      accessibilityLayer
                      data={charts.volume}
                      barCategoryGap="22%"
                      margin={{ left: -18, right: 4, top: 8, bottom: 0 }}
                    >
                      <CartesianGrid vertical={false} strokeDasharray="3 5" />
                      <XAxis
                        dataKey="date"
                        tickFormatter={shortDate}
                        tickLine={false}
                        axisLine={false}
                        minTickGap={30}
                        tickMargin={12}
                      />
                      <YAxis
                        width={54}
                        tickFormatter={compact}
                        allowDecimals={false}
                        tickLine={false}
                        axisLine={false}
                      />
                      <ChartTooltip
                        cursor={{ fill: "var(--muted)" }}
                        content={<ChartTooltipContent hideIndicator />}
                      />
                      {keys.map((key, i) => (
                        <Bar
                          key={key}
                          dataKey={key}
                          stackId="volume"
                          fill={colors[i]}
                          hide={hidden.includes(String(i))}
                          radius={
                            i === 1 || hidden.includes("1") ? [4, 4, 0, 0] : 0
                          }
                          isAnimationActive={animate}
                          animationDuration={650}
                        />
                      ))}
                    </BarChart>
                  </ChartContainer>
                ) : (
                  <EmptyChart>
                    No accepted messages in this period. Your campaign and
                    newsletter volume will appear here after sending.
                  </EmptyChart>
                )}
                <div className="mt-4 flex flex-wrap items-center justify-between gap-2 border-t pt-4">
                  <p className="text-[11px] text-muted-foreground">
                    Accepted is an SMTP handoff, not confirmed inbox delivery.
                  </p>
                  <div className="flex gap-1">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setTable(!table)}
                      aria-expanded={table}
                      aria-controls="daily-chart-data"
                    >
                      <Table2 className="mr-2 h-3.5 w-3.5" />
                      {table ? "Hide data" : "View data"}
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={exportDaily}
                      aria-label="Export daily analytics CSV"
                    >
                      <ArrowDownToLine className="h-3.5 w-3.5" />
                    </Button>
                  </div>
                </div>
                {table && (
                  <div
                    id="daily-chart-data"
                    className="mt-3 max-h-72 overflow-auto rounded-lg border"
                  >
                    <table className="w-full text-left text-xs">
                      <caption className="p-3 text-left text-muted-foreground">
                        Daily send cohorts · {report.timezone}. Clicks counted
                        up to the report cutoff.
                      </caption>
                      <thead className="sticky top-0 bg-muted">
                        <tr>
                          {[
                            "Date",
                            "Campaigns",
                            "Newsletters",
                            "Campaign clicks",
                            "Newsletter clicks",
                          ].map((h) => (
                            <th
                              scope="col"
                              key={h}
                              className="whitespace-nowrap px-3 py-2 font-medium"
                            >
                              {h}
                            </th>
                          ))}
                        </tr>
                      </thead>
                      <tbody>
                        {charts.volume.map((d) => (
                          <tr key={d.date} className="border-t">
                            {[
                              d.date,
                              d.campaigns,
                              d.newsletters,
                              d.campaignClicks,
                              d.newsletterClicks,
                            ].map((v, i) => (
                              <td
                                key={i}
                                className="whitespace-nowrap px-3 py-2 tabular-nums"
                              >
                                {typeof v === "number" ? number(v) : v}
                              </td>
                            ))}
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </>
            ) : (
              <EmptyChart>
                This API has not supplied the sending breakdown yet. The
                existing activity and audience data are shown below.
              </EmptyChart>
            )}
          </CardContent>
        </Card>
        <Card className={styles.panel}>
          <CardHeader className={styles.panelHeader}>
            <p className="mb-1 text-[10px] font-medium uppercase tracking-[.16em] text-muted-foreground">
              Your audience, understood
            </p>
            <CardTitle className={styles.panelTitle}>Engagement mix</CardTitle>
            <CardDescription className={styles.panelDescription}>
              Active subscribers grouped by tracked activity in the 90 days
              ending at the cutoff.
            </CardDescription>
          </CardHeader>
          <CardContent className={styles.panelBody}>
            {cohortTotal > 0 ? (
              <>
                <div className="relative">
                  <ChartContainer
                    config={{ count: { label: "Subscribers" } }}
                    className="mx-auto h-[205px] w-full"
                  >
                    <PieChart accessibilityLayer>
                      <ChartTooltip
                        content={
                          <ChartTooltipContent
                            nameKey="label"
                            hideLabel
                            hideIndicator
                          />
                        }
                      />
                      <Pie
                        data={report.cohorts.filter((c) => c.count > 0)}
                        dataKey="count"
                        nameKey="label"
                        innerRadius={65}
                        outerRadius={88}
                        paddingAngle={3}
                        stroke="none"
                        isAnimationActive={animate}
                        animationDuration={650}
                      >
                        {report.cohorts
                          .filter((c) => c.count > 0)
                          .map((c) => (
                            <Cell
                              key={c.key}
                              fill={cohortColors[c.key]}
                              aria-label={`${c.label}: ${number(c.count)} subscribers`}
                            />
                          ))}
                      </Pie>
                    </PieChart>
                  </ChartContainer>
                  <div className="pointer-events-none absolute inset-0 flex flex-col items-center justify-center">
                    <strong className="text-2xl font-semibold tabular-nums">
                      {compact(cohortTotal)}
                    </strong>
                    <span className="mt-1 text-[10px] text-muted-foreground">
                      active subscribers
                    </span>
                  </div>
                </div>
                <div className="mt-3 space-y-1">
                  {report.cohorts.map((c) => (
                    <Link
                      key={c.key}
                      href={`/analytics/audience/contacts?${scope}&cohort=${c.key}`}
                      className="flex min-h-10 items-center gap-2 rounded-lg px-2 text-[11px] hover:bg-muted/60"
                    >
                      <span
                        className="h-2 w-2 shrink-0 rounded-full"
                        style={{ backgroundColor: cohortColors[c.key] }}
                      />
                      <span className="flex-1">{c.label}</span>
                      <strong className="font-medium tabular-nums">
                        {number(c.count)}
                      </strong>
                      <ArrowUpRight className="h-3 w-3 shrink-0 text-muted-foreground" />
                    </Link>
                  ))}
                </div>
              </>
            ) : (
              <EmptyChart>
                No active subscribers in the selected audience.
              </EmptyChart>
            )}
          </CardContent>
        </Card>
      </div>
      <div className="grid gap-5 lg:grid-cols-2">
        <Card className={styles.panel}>
          <CardHeader className={styles.panelHeader}>
            <CardTitle className={styles.panelTitle}>
              Subscriber growth
            </CardTitle>
            <CardDescription className={styles.panelDescription}>
              Recorded active list subscriptions. Separate from current
              deliverability eligibility.
            </CardDescription>
          </CardHeader>
          <CardContent className={styles.panelBody}>
            {report.growth.available ? (
              <>
                <div className="mb-5 flex flex-wrap gap-8">
                  {[
                    ["At period start", report.growth.start],
                    ["At cutoff", report.growth.end],
                    ["Net change", report.growth.net],
                  ].map(([label, value]) => (
                    <div key={String(label)}>
                      <p className="text-[10px] text-muted-foreground">
                        {label}
                      </p>
                      <p className="mt-1 text-xl font-semibold tabular-nums">
                        {label === "Net change" && Number(value) > 0 ? "+" : ""}
                        {value == null ? "—" : number(Number(value))}
                      </p>
                    </div>
                  ))}
                </div>
                <ChartContainer config={labels} className="h-52 w-full">
                  <AreaChart
                    accessibilityLayer
                    data={report.growth.series}
                    margin={{ left: -15, right: 5, top: 5, bottom: 0 }}
                  >
                    <defs>
                      <linearGradient
                        id={`${gradient}-growth`}
                        x1="0"
                        y1="0"
                        x2="0"
                        y2="1"
                      >
                        <stop
                          offset="0%"
                          stopColor="var(--chart-2)"
                          stopOpacity={0.35}
                        />
                        <stop
                          offset="100%"
                          stopColor="var(--chart-2)"
                          stopOpacity={0.02}
                        />
                      </linearGradient>
                    </defs>
                    <CartesianGrid vertical={false} strokeDasharray="3 5" />
                    <XAxis
                      dataKey="date"
                      tickFormatter={shortDate}
                      tickLine={false}
                      axisLine={false}
                      minTickGap={30}
                    />
                    <YAxis
                      allowDecimals={false}
                      tickFormatter={compact}
                      tickLine={false}
                      axisLine={false}
                      width={54}
                    />
                    <ChartTooltip
                      content={<ChartTooltipContent hideIndicator />}
                    />
                    <Area
                      type="stepAfter"
                      dataKey="active"
                      stroke="var(--chart-2)"
                      fill={`url(#${gradient}-growth)`}
                      strokeWidth={2.5}
                      isAnimationActive={animate}
                      animationDuration={650}
                    />
                  </AreaChart>
                </ChartContainer>
                <details className="mt-3 text-xs">
                  <summary className="cursor-pointer text-muted-foreground">
                    View subscription history
                  </summary>
                  <div className="mt-3 max-h-52 overflow-auto">
                    <table className="w-full text-left">
                      <thead>
                        <tr>
                          <th scope="col" className="py-2">
                            Time ({report.timezone})
                          </th>
                          <th scope="col">Active subscriptions</th>
                        </tr>
                      </thead>
                      <tbody>
                        {report.growth.series.map((d) => (
                          <tr key={d.date} className="border-t">
                            <td className="py-2">{d.date}</td>
                            <td>{number(d.active)}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </details>
              </>
            ) : (
              <EmptyChart>
                <div>
                  <BarChart3 className="mx-auto mb-3 h-6 w-6 text-muted-foreground" />
                  <p className="font-medium text-foreground">
                    Building reliable subscriber history
                  </p>
                  <p className="mt-2 max-w-sm text-xs leading-relaxed">
                    {tagFiltered
                      ? "Historical tag membership is not recorded. Clear the tag filter to see available growth."
                      : report.growth.since
                        ? `Complete history starts ${new Date(report.growth.since).toLocaleString()}. Choose a range after that time.`
                        : "Subscription history is not available yet."}{" "}
                    Existing contacts are not treated as newly acquired
                    subscribers.
                  </p>
                </div>
              </EmptyChart>
            )}
          </CardContent>
        </Card>
        <Card className={styles.panel}>
          <CardHeader className={styles.panelHeader}>
            <CardTitle className={styles.panelTitle}>
              Audience activity
            </CardTitle>
            <CardDescription className={styles.panelDescription}>
              Daily unique recipients. Click activity may come from messages
              sent before this period.
            </CardDescription>
          </CardHeader>
          <CardContent className={styles.panelBody}>
            <div className="mb-5 flex flex-wrap gap-4 text-[11px]">
              <span className="flex items-center gap-2">
                <i className="h-2 w-2 rounded-full" style={{ backgroundColor: colors[0] }} />
                Recipients reached
              </span>
              <span className="flex items-center gap-2">
                <i className="h-2 w-2 rounded-full" style={{ backgroundColor: colors[1] }} />
                Recipients who clicked
              </span>
            </div>
            <ChartContainer config={labels} className="h-[250px] w-full">
              <AreaChart
                accessibilityLayer
                data={report.activity}
                margin={{ left: -15, right: 5, top: 5, bottom: 0 }}
              >
                <defs>
                  <linearGradient
                    id={`${gradient}-activity`}
                    x1="0"
                    y1="0"
                    x2="0"
                    y2="1"
                  >
                    <stop offset="0%" stopColor="var(--chart-1)" stopOpacity={0.2} />
                    <stop offset="100%" stopColor="var(--chart-1)" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid vertical={false} strokeDasharray="3 5" />
                <XAxis
                  dataKey="date"
                  tickFormatter={shortDate}
                  tickLine={false}
                  axisLine={false}
                  minTickGap={30}
                />
                <YAxis
                  allowDecimals={false}
                  tickFormatter={compact}
                  tickLine={false}
                  axisLine={false}
                  width={54}
                />
                <ChartTooltip content={<ChartTooltipContent hideIndicator />} />
                <Area
                  dataKey="reached"
                  type="linear"
                  stroke="var(--chart-1)"
                  fill={`url(#${gradient}-activity)`}
                  strokeWidth={2}
                  isAnimationActive={animate}
                />
                <Area
                  dataKey="clicked"
                  type="linear"
                  stroke="var(--chart-2)"
                  fill="transparent"
                  strokeWidth={2}
                  isAnimationActive={animate}
                />
              </AreaChart>
            </ChartContainer>
            <details className="mt-3 text-xs">
              <summary className="cursor-pointer text-muted-foreground">
                View daily activity values
              </summary>
              <div className="mt-3 max-h-52 overflow-auto">
                <table className="w-full text-left">
                  <thead>
                    <tr>
                      <th scope="col" className="py-2">
                        Date
                      </th>
                      <th scope="col">Reached</th>
                      <th scope="col">Clicked</th>
                    </tr>
                  </thead>
                  <tbody>
                    {report.activity.map((d) => (
                      <tr key={d.date} className="border-t">
                        <td className="py-2">{d.date}</td>
                        <td>{number(d.reached)}</td>
                        <td>{number(d.clicked)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </details>
            <p className="mt-3 text-[11px] leading-relaxed text-muted-foreground">
              Daily unique counts do not add up to period unique totals.
            </p>
          </CardContent>
        </Card>
      </div>
      {charts && (
        <div className="grid gap-5 xl:grid-cols-[minmax(0,1fr)_minmax(0,1.6fr)]">
          <Card className={styles.panel}>
            <CardHeader className={styles.panelHeader}>
              <CardTitle className={styles.panelTitle}>
                Devices behind the clicks
              </CardTitle>
              <CardDescription className={styles.panelDescription}>
                Reported device types on recorded click events.
              </CardDescription>
            </CardHeader>
            <CardContent className={styles.panelBody}>
              {charts.clickEvents > 0 ? (
                <>
                  <div className="relative">
                    <ChartContainer config={labels} className="h-52 w-full">
                      <PieChart accessibilityLayer>
                        <ChartTooltip
                          content={
                            <ChartTooltipContent hideLabel hideIndicator />
                          }
                        />
                        <Pie
                          data={charts.devices
                            .filter((d) => d.count > 0)
                            .map((d) => ({
                              ...d,
                              name: deviceLabels[d.device],
                            }))}
                          dataKey="count"
                          nameKey="name"
                          innerRadius={65}
                          outerRadius={88}
                          paddingAngle={3}
                          stroke="none"
                          isAnimationActive={animate}
                          onClick={(d) =>
                            setDevice(d.device === device ? null : d.device)
                          }
                        >
                          {charts.devices
                            .filter((d) => d.count > 0)
                            .map((d) => (
                              <Cell
                                key={d.device}
                                aria-label={`${deviceLabels[d.device]}: ${number(d.count)} recorded click events`}
                                fill={
                                  colors[
                                    charts.devices.findIndex(
                                      (v) => v.device === d.device,
                                    )
                                  ]
                                }
                                fillOpacity={
                                  !device || device === d.device ? 1 : 0.25
                                }
                              />
                            ))}
                        </Pie>
                      </PieChart>
                    </ChartContainer>
                    <div className="pointer-events-none absolute inset-0 flex flex-col items-center justify-center">
                      <strong className="text-2xl tabular-nums">
                        {selectedDevice
                          ? `${((selectedDevice.count / charts.clickEvents) * 100).toFixed(1)}%`
                          : compact(charts.clickEvents)}
                      </strong>
                      <span className="mt-1 text-[10px] text-muted-foreground">
                        {selectedDevice
                          ? deviceLabels[selectedDevice.device]
                          : "click events"}
                      </span>
                    </div>
                  </div>
                  <div className="space-y-1">
                    {charts.devices.map((d, i) => (
                      <button
                        key={d.device}
                        aria-pressed={device === d.device}
                        onClick={() =>
                          setDevice(device === d.device ? null : d.device)
                        }
                        className={`flex min-h-10 w-full items-center gap-2 rounded-lg px-3 text-xs ${device === d.device ? "bg-muted" : "hover:bg-muted/60"}`}
                      >
                        <span
                          className="h-2 w-2 rounded-full"
                          style={{ backgroundColor: colors[i] }}
                        />
                        <span className="flex-1 text-left">
                          {deviceLabels[d.device]}
                        </span>
                        <strong className="font-medium tabular-nums">
                          {number(d.count)}
                        </strong>
                        <span className="w-12 text-right text-muted-foreground">
                          {((d.count / charts.clickEvents) * 100).toFixed(1)}%
                        </span>
                      </button>
                    ))}
                  </div>
                </>
              ) : (
                <EmptyChart>
                  No recorded click events in this period.
                </EmptyChart>
              )}
              <p className="mt-4 text-[11px] leading-relaxed text-muted-foreground">
                Repeat clicks and scanners can be included. Unknown devices stay
                unknown.
              </p>
            </CardContent>
          </Card>
          <Card className={styles.panel}>
            <CardHeader className={styles.panelHeader}>
              <CardTitle className={styles.panelTitle}>
                Click activity by time
              </CardTitle>
              <CardDescription className={styles.panelDescription}>
                Observed events in {report.timezone}, grouped into four-hour
                windows. Select a cell to inspect it.
              </CardDescription>
            </CardHeader>
            <CardContent className={styles.panelBody}>
              <div className="grid grid-cols-[28px_repeat(6,minmax(0,1fr))] gap-1.5 sm:gap-2.5">
                <div />
                {["00–04", "04–08", "08–12", "12–16", "16–20", "20–24"].map(
                  (t) => (
                    <span
                      key={t}
                      className="pb-1 text-center text-[9px] text-muted-foreground sm:text-[10px]"
                    >
                      {t}
                    </span>
                  ),
                )}
                {dayNames.map((day, i) => (
                  <div key={day} className="contents">
                    <span className="flex items-center text-[10px] text-muted-foreground">
                      {day}
                    </span>
                    {hourBlocks
                      .filter((h) => h.day === i)
                      .map((h) => {
                        const intensity = h.count / maxHour;
                        return (
                          <button
                            key={h.block}
                            aria-label={`${day}, ${h.block * 4}:00 to ${(h.block + 1) * 4}:00: ${number(h.count)} recorded click events`}
                            aria-pressed={
                              selectedHour?.day === i &&
                              selectedHour.block === h.block
                            }
                            onClick={() =>
                              setSelectedHour({ day: i, block: h.block })
                            }
                            className={`h-10 rounded-md border border-transparent transition-all hover:ring-2 hover:ring-ring focus-visible:ring-2 focus-visible:ring-ring ${selectedHour?.day === i && selectedHour.block === h.block ? "ring-2 ring-ring ring-offset-2 ring-offset-background" : ""}`}
                            style={{ backgroundColor: heatmapColors[h.count === 0 ? 0 : intensity < 0.25 ? 1 : intensity < 0.5 ? 2 : intensity < 0.75 ? 3 : 4] }}
                          />
                        );
                      })}
                  </div>
                ))}
              </div>
              <div className="mt-5 flex justify-end gap-1.5 text-[10px] text-muted-foreground">
                <span className="mr-1">Fewer</span>
                {heatmapColors.map((color) => (
                  <span key={color} className="h-3 w-3 rounded-sm" style={{ backgroundColor: color }} />
                ))}
                <span className="ml-1">More</span>
              </div>
              <div
                className="mt-5 rounded-lg border bg-muted/20 p-4 text-xs"
                aria-live="polite"
              >
                {hourSelection ? (
                  <>
                    <strong>
                      {dayNames[hourSelection.day]} ·{" "}
                      {String(hourSelection.block * 4).padStart(2, "0")}:00–
                      {String((hourSelection.block + 1) * 4).padStart(2, "0")}
                      :00
                    </strong>
                    <p className="mt-1.5 text-muted-foreground">
                      {number(hourSelection.count)} recorded click events across
                      the selected period.
                    </p>
                  </>
                ) : (
                  <>
                    <strong>
                      {number(charts.clickEvents)} recorded click events
                    </strong>
                    <p className="mt-1.5 text-muted-foreground">
                      {charts.clickEvents
                        ? "Explore a time window above."
                        : "Time windows will fill as clicks are recorded."}
                    </p>
                  </>
                )}
              </div>
              <p className="mt-4 text-[11px] leading-relaxed text-muted-foreground">
                Includes clicks on older messages and repeat events. This
                describes observed activity, not a recommended or statistically
                optimal send time.
              </p>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}
