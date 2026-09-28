"use client";
import { OnboardingBanner } from "@/components/sending/onboarding";

import { useMemo, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import {
  Bar,
  BarChart,
  Area,
  AreaChart,
  CartesianGrid,
  XAxis,
  YAxis,
} from "recharts";
import {
  ArrowUpRight,
  BarChart3,
  Clock3,
  Mail,
  MousePointer2,
  RefreshCw,
  Plus,
  Send,
  ShieldAlert,
} from "lucide-react";
import { AnalyticsWorkspace } from "./analytics-workspace";
import { useTeam } from "@/app/providers/team-provider";
import { useAnalytics } from "@/lib/analytics/use-analytics";
import { EmailReport, formatRate } from "@/lib/analytics/types";
import { PageHeading, QueryState } from "@/components/marketing/shared";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from "@/components/ui/card";
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Input } from "@/components/ui/input";
import styles from "./analytics-surface.module.css";

const number = (n: number) => n.toLocaleString();
const compact = (n: number) =>
  Intl.NumberFormat("en", { notation: "compact" }).format(n);
const chartConfig = {
  sent: { label: "Sent", color: "var(--chart-1)" },
  queued: { label: "Pending", color: "var(--chart-4)" },
  failed: { label: "Failed", color: "var(--chart-5)" },
  drafts: { label: "Drafts", color: "var(--muted-foreground)" },
  unknown: { label: "Unknown", color: "var(--chart-3)" },
  opened: { label: "Opened", color: "var(--chart-4)" },
  clicked: { label: "Clicked", color: "var(--chart-2)" },
};
function dateString(d: Date) {
  return `${d.getUTCFullYear()}-${String(d.getUTCMonth() + 1).padStart(2, "0")}-${String(d.getUTCDate()).padStart(2, "0")}`;
}
function period(days: number) {
  const to = new Date();
  const from = new Date();
  from.setUTCDate(from.getUTCDate() - days + 1);
  return { from: dateString(from), to: dateString(to) };
}
function nextDay(value: string) {
  const d = new Date(`${value}T12:00:00Z`);
  if (Number.isNaN(d.getTime())) return "";
  d.setUTCDate(d.getUTCDate() + 1);
  return dateString(d);
}

export function DashboardHome() {
  const search = useSearchParams();
  const router = useRouter();
  const tab = search.get("tab") === "campaigns" ? "campaigns" : "email";
  const [dates, setDates] = useState(() => period(30));
  const [days, setDays] = useState("30");
  return (
    <div className={styles.workspace}>
      <OnboardingBanner />
      <PageHeading
        title="Dashboard"
        description="Your email activity and campaign performance, in one place."
        action={
          <div className="flex flex-wrap items-center gap-2">
            <Button variant="outline" asChild>
              <Link href="/developer/logs/emails">
                Open outbox
                <ArrowUpRight className="ml-2 h-4 w-4" />
              </Link>
            </Button>
            <Button asChild>
              <Link href="/campaigns/new">
                <Plus className="mr-2 h-4 w-4" />
                Create campaign
              </Link>
            </Button>
          </div>
        }
      />
      <Tabs
        value={tab}
        onValueChange={(value) => {
          const q = new URLSearchParams(search);
          q.set("tab", value);
          router.replace(`/?${q}`, { scroll: false });
        }}
      >
        <TabsList aria-label="Dashboard analytics" className={styles.tabs}>
          <TabsTrigger value="email" className={styles.tab}>
            <Mail className="h-4 w-4" />
            Email
          </TabsTrigger>
          <TabsTrigger value="campaigns" className={styles.tab}>
            <BarChart3 className="h-4 w-4" />
            Campaigns
          </TabsTrigger>
        </TabsList>
        <TabsContent value="email" className="mt-6">
          <EmailActivity
            dates={dates}
            setDates={setDates}
            days={days}
            setDays={setDays}
          />
        </TabsContent>
        <TabsContent value="campaigns" className="mt-6">
          <AnalyticsWorkspace view="campaigns" embedded />
        </TabsContent>
      </Tabs>
    </div>
  );
}
function EmailActivity({
  dates,
  setDates,
  days,
  setDays,
}: {
  dates: { from: string; to: string };
  setDates: (dates: { from: string; to: string }) => void;
  days: string;
  setDays: (days: string) => void;
}) {
  const params = useMemo(
    () =>
      new URLSearchParams({
        from: dates.from,
        to: nextDay(dates.to),
        timezone: "UTC",
      }),
    [dates],
  );
  const valid = !!dates.from && !!dates.to && dates.from <= dates.to;
  const query = useAnalytics<EmailReport>("email-overview", params, valid);
  const r = query.data;
  const {
    team,
    loading: teamLoading,
    error: teamError,
    refreshTeam,
  } = useTeam();
  return (
    <div className={styles.report}>
      <div className={styles.filters}>
        <label className={styles.filter}>
          <span>Period</span>
          <Select
            value={days}
            onValueChange={(v) => {
              setDays(v);
              if (v !== "custom") setDates(period(Number(v)));
            }}
          >
            <SelectTrigger className="w-36">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {[7, 30, 90].map((n) => (
                <SelectItem key={n} value={String(n)}>
                  Last {n} days
                </SelectItem>
              ))}
              <SelectItem value="custom">Custom range</SelectItem>
            </SelectContent>
          </Select>
        </label>
        <label className={styles.filter}>
          <span>From</span>
          <Input
            type="date"
            aria-label="Email from date"
            className="w-40"
            value={dates.from}
            onChange={(e) => {
              setDays("custom");
              setDates({ ...dates, from: e.target.value });
            }}
          />
        </label>
        <label className={styles.filter}>
          <span>To</span>
          <Input
            type="date"
            aria-label="Email to date"
            className="w-40"
            value={dates.to}
            onChange={(e) => {
              setDays("custom");
              setDates({ ...dates, to: e.target.value });
            }}
          />
        </label>
        <span className={styles.filterNote}>All email sources · UTC</span>
        <Button
          variant="outline"
          className={styles.refresh}
          disabled={!valid || query.isFetching}
          onClick={() => void query.refetch()}
        >
          <RefreshCw
            className={`mr-2 h-4 w-4 ${query.isFetching ? "animate-spin" : ""}`}
          />
          Refresh
        </Button>
      </div>
      {!valid && (
        <p role="alert" className="text-sm text-destructive">
          Choose an end date on or after the start date.
        </p>
      )}
      <QueryState
        loading={teamLoading || (valid && !!team && query.isPending)}
        error={teamError || query.error}
        retry={() => void (teamError ? refreshTeam() : query.refetch())}
      />
      {!teamLoading && !team && !teamError && (
        <p className="text-sm text-muted-foreground">
          Select a workspace to view email activity.
        </p>
      )}
      {valid && team && r && !query.isError && <EmailOverview report={r} />}
    </div>
  );
}

/** Presentational report shared with the development-only visual preview. */
export function EmailOverview({ report: r }: { report: EmailReport }) {
  return (
    <div className={styles.report}>
      <div className={styles.context}>
        <p>All email sources · Test sends excluded</p>
        <time dateTime={r.asOf}>
          Updated{" "}
          {new Date(r.asOf).toLocaleString([], {
            month: "short",
            day: "numeric",
            hour: "2-digit",
            minute: "2-digit",
            timeZone: "UTC",
            timeZoneName: "short",
          })}
        </time>
      </div>
      <div className={styles.metrics}>
        <Stat
          label="Emails sent"
          value={number(r.summary.sent)}
          detail="Accepted by SMTP in this period."
          icon={Send}
          tone="accent"
        />
        <Stat
          label="Pending emails"
          value={number(r.summary.queued)}
          detail="Queued, scheduled, or sending; created in this period."
          icon={Clock3}
        />
        <Stat
          label="Click rate"
          value={formatRate(r.clickRate)}
          detail={`${number(r.summary.clicked)} clicked messages / ${number(r.summary.sent)} sent.`}
          icon={MousePointer2}
        />
        <Stat
          label="Failed emails"
          value={number(r.summary.failed)}
          detail="Current failed status; created in this period."
          icon={ShieldAlert}
          tone={r.summary.failed > 0 ? "danger" : undefined}
        />
      </div>
      {r.summary.total === 0 && (
        <div className={styles.empty}>
          <div>
            <h2>No email activity in this period</h2>
            <p>
              Try a wider date range, or send an email from the outbox or API.
              You don’t need a campaign to see activity here.
            </p>
          </div>
          <Button variant="outline" asChild>
            <Link href="/developer/logs/emails">
              Go to outbox <ArrowUpRight className="ml-2 h-4 w-4" />
            </Link>
          </Button>
        </div>
      )}
      <div className={styles.primaryGrid}>
        <Card className={styles.panel}>
          <CardHeader className={styles.panelHeader}>
            <div>
              <CardTitle className={styles.panelTitle}>
                Email activity
              </CardTitle>
              <CardDescription className={styles.panelDescription}>
                Your sending volume, day by day.
              </CardDescription>
            </div>
          </CardHeader>
          <CardContent className={styles.panelBody}>
            <div className={styles.legend}>
              {(["sent", "queued", "failed", "drafts", "unknown"] as const).map(
                (k) => (
                  <span key={k} className={styles.legendItem}>
                    <svg width="7" height="7" aria-hidden="true">
                      <circle
                        cx="3.5"
                        cy="3.5"
                        r="3.5"
                        fill={chartConfig[k].color}
                      />
                    </svg>
                    {chartConfig[k].label}
                  </span>
                ),
              )}
            </div>
            <ChartContainer
              config={chartConfig}
              className={styles.chart}
              role="img"
              aria-label="Daily email records by status"
            >
              <BarChart
                accessibilityLayer
                data={r.series}
                margin={{ left: -8, right: 0, top: 8, bottom: 0 }}
                barCategoryGap="28%"
              >
                <CartesianGrid vertical={false} strokeDasharray="3 5" />
                <XAxis
                  dataKey="date"
                  tickFormatter={(d) => d.slice(5)}
                  tickLine={false}
                  axisLine={false}
                  minTickGap={28}
                  tickMargin={12}
                />
                <YAxis
                  allowDecimals={false}
                  tickFormatter={compact}
                  tickLine={false}
                  axisLine={false}
                  width={42}
                />
                <ChartTooltip content={<ChartTooltipContent />} />
                {(
                  ["sent", "queued", "failed", "drafts", "unknown"] as const
                ).map((k) => (
                  <Bar
                    key={k}
                    dataKey={k}
                    stackId="status"
                    fill={`var(--color-${k})`}
                    maxBarSize={22}
                    isAnimationActive={false}
                  />
                ))}
              </BarChart>
            </ChartContainer>
            <details className={styles.dataDetails}>
              <summary>View daily counts</summary>
              <div className={styles.dataScroll}>
                <table>
                  <thead>
                    <tr>
                      {[
                        "Date",
                        "Sent",
                        "Pending",
                        "Failed",
                        "Drafts",
                        "Unknown",
                      ].map((s) => (
                        <th key={s} scope="col">
                          {s}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {r.series.map((d) => (
                      <tr key={d.date}>
                        <td>{d.date}</td>
                        {(
                          [
                            "sent",
                            "queued",
                            "failed",
                            "drafts",
                            "unknown",
                          ] as const
                        ).map((k) => (
                          <td key={k}>{number(d[k])}</td>
                        ))}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </details>
          </CardContent>
          <div className={styles.panelFooter}>
            Sent emails use their send date. Other records use their creation
            date.
          </div>
        </Card>
        <Card className={styles.panel}>
          <CardHeader className={styles.panelHeader}>
            <div>
              <CardTitle className={styles.panelTitle}>At a glance</CardTitle>
              <CardDescription className={styles.panelDescription}>
                Across API, outbox, and marketing.
              </CardDescription>
            </div>
          </CardHeader>
          <CardContent className={styles.detailRows}>
            <dl>
              {[
                ["Open rate", formatRate(r.openRate)],
                [
                  "Known bounces",
                  `${number(r.summary.bounced)} (${formatRate(r.bounceRate)})`,
                ],
                ["Campaign-linked sends", number(r.summary.campaigns)],
                ["Other email sends", number(r.summary.other)],
                ["Drafts", number(r.summary.drafts)],
                ["Unknown outcomes", number(r.summary.unknown)],
              ].map(([label, value]) => (
                <div key={label} className={styles.detailRow}>
                  <dt>{label}</dt>
                  <dd>{value}</dd>
                </div>
              ))}
            </dl>
          </CardContent>
          <div className={styles.panelFooter}>
            <Link href="/developer/logs/emails">
              Inspect individual emails <ArrowUpRight className="h-3 w-3" />
            </Link>
          </div>
        </Card>
      </div>
      <Card className={styles.panel}>
        <CardHeader className={styles.panelHeader}>
          <div>
            <CardTitle className={styles.panelTitle}>
              Email engagement
            </CardTitle>
            <CardDescription className={styles.panelDescription}>
              Unique opened and clicked messages, grouped by send date.
            </CardDescription>
          </div>
          <div className={styles.engagementHeader}>
            {(["opened", "clicked"] as const).map((key) => (
              <span key={key} className={styles.engagementStat}>
                <svg width="7" height="7" aria-hidden="true">
                  <circle
                    cx="3.5"
                    cy="3.5"
                    r="3.5"
                    fill={chartConfig[key].color}
                  />
                </svg>
                {chartConfig[key].label}
                <strong>{number(r.summary[key])}</strong>
              </span>
            ))}
          </div>
        </CardHeader>
        <CardContent className={styles.panelBody}>
          <ChartContainer
            config={chartConfig}
            className={styles.chart}
            role="img"
            aria-label="Opened and clicked emails by send date"
          >
            <AreaChart
              accessibilityLayer
              data={r.series}
              margin={{ left: -8, right: 0, top: 8, bottom: 0 }}
            >
              <CartesianGrid vertical={false} strokeDasharray="3 5" />
              <XAxis
                dataKey="date"
                tickFormatter={(d) => d.slice(5)}
                tickLine={false}
                axisLine={false}
                minTickGap={28}
                tickMargin={12}
              />
              <YAxis
                allowDecimals={false}
                width={42}
                tickFormatter={compact}
                axisLine={false}
                tickLine={false}
              />
              <ChartTooltip content={<ChartTooltipContent />} />
              <Area
                dataKey="opened"
                type="monotone"
                stroke="var(--color-opened)"
                fill="var(--color-opened)"
                fillOpacity={0.08}
                strokeWidth={2}
                isAnimationActive={false}
              />
              <Area
                dataKey="clicked"
                type="monotone"
                stroke="var(--color-clicked)"
                fill="var(--color-clicked)"
                fillOpacity={0.08}
                strokeWidth={2}
                isAnimationActive={false}
              />
            </AreaChart>
          </ChartContainer>
          <details className={styles.dataDetails}>
            <summary>View engagement counts</summary>
            <div className={styles.dataScroll}>
              <table>
                <thead>
                  <tr>
                    <th scope="col">Date</th>
                    <th scope="col">Opened</th>
                    <th scope="col">Clicked</th>
                  </tr>
                </thead>
                <tbody>
                  {r.series.map((d) => (
                    <tr key={d.date}>
                      <td>{d.date}</td>
                      <td>{number(d.opened)}</td>
                      <td>{number(d.clicked)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </details>
        </CardContent>
        <div className={styles.panelFooter}>
          Open tracking is directional. Clicks can include scanners. Missing
          tracking does not prove a message was unread.
        </div>
      </Card>
      <p className={styles.note}>
        Counts are email records, not individual recipients in CC/BCC. Sent
        means accepted by SMTP, not verified inbox delivery. Bounces are a
        subset of sent emails. Statuses reflect the latest recorded state.
      </p>
    </div>
  );
}

function Stat({
  label,
  value,
  detail,
  icon: Icon,
  tone,
}: {
  label: string;
  value: string;
  detail: string;
  icon: typeof Mail;
  tone?: "accent" | "danger";
}) {
  return (
    <div className={styles.metric} data-tone={tone}>
      <div className={styles.metricLabel}>
        <Icon aria-hidden="true" />
        <span>{label}</span>
      </div>
      <p className={styles.metricValue}>{value}</p>
      <p className={styles.metricDetail}>{detail}</p>
    </div>
  );
}
