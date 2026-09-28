"use client";
import Link from "next/link";
import { AnalyticsCharts } from "./analytics-charts";
import styles from "./analytics-surface.module.css";
import { useMemo, useState, useEffect, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import {
  ArrowDownToLine,
  ArrowUpRight,
  BarChart3,
  ChevronLeft,
  ChevronRight,
  Info,
  RefreshCw,
  Users,
  MousePointer2,
  Send,
  ShieldCheck,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { PageHeading, QueryState } from "@/components/marketing/shared";
import { useAnalytics } from "@/lib/analytics/use-analytics";
import {
  AnalyticsPage,
  BreakdownRow,
  formatRate,
  Rate,
  rateDifference,
  Report,
} from "@/lib/analytics/types";

export type AnalyticsView = "overview" | "audience" | "campaigns" | "delivery";
const views = [
  { value: "overview", label: "Overview", href: "/analytics" },
  { value: "audience", label: "Audience", href: "/audience/dashboard" },
  {
    value: "campaigns",
    label: "Campaigns & newsletters",
    href: "/analytics/campaigns",
  },
  { value: "delivery", label: "Delivery health", href: "/analytics/delivery" },
];
const number = (value: number) => value.toLocaleString();
function shiftDate(date: string, days: number) {
  if (!date) return "";
  const d = new Date(`${date}T00:00:00Z`);
  if (Number.isNaN(d.getTime())) return "";
  d.setUTCDate(d.getUTCDate() + days);
  return d.toISOString().slice(0, 10);
}
function range(days: number) {
  const today = new Date().toISOString().slice(0, 10);
  return { from: shiftDate(today, -(days - 1)), to: shiftDate(today, 1) };
}
function MetricCard({
  label,
  value,
  detail,
  icon: Icon,
  change,
}: {
  label: string;
  value: string;
  detail: string;
  icon: typeof Users;
  change?: string;
}) {
  return (
    <div className={styles.metric}>
      <div className={styles.metricLabel}>
        <Icon aria-hidden="true" />
        <span>{label}</span>
      </div>
      <p className={styles.metricValue}>{value}</p>
      <p className={styles.metricDetail}>{detail}</p>
      {change && <p className={styles.metricChange}>{change}</p>}
    </div>
  );
}

function RateValue({ rate }: { rate: Rate }) {
  return (
    <span
      className="tabular-nums"
      title={
        rate.denominator
          ? `${number(rate.numerator)} / ${number(rate.denominator)}`
          : "No eligible messages in this period"
      }
    >
      {formatRate(rate)}
      <span className="mt-0.5 block text-xs text-muted-foreground">
        {number(rate.numerator)} / {number(rate.denominator)}
      </span>
    </span>
  );
}
export function AnalyticsWorkspace({
  view = "audience",
  embedded = false,
}: {
  view?: AnalyticsView;
  embedded?: boolean;
}) {
  return (
    <Suspense fallback={<p role="status">Loading analytics…</p>}>
      <Workspace view={view} embedded={embedded} />
    </Suspense>
  );
}
function Workspace({
  view,
  embedded,
}: {
  view: AnalyticsView;
  embedded: boolean;
}) {
  const search = useSearchParams();
  const [localTimezone, setLocalTimezone] = useState("UTC");
  useEffect(
    () => setLocalTimezone(Intl.DateTimeFormat().resolvedOptions().timeZone),
    [],
  );
  const [dates, setDates] = useState(() => ({
    from: search.get("from") || range(30).from,
    to: search.get("to") || range(30).to,
  }));
  const [days, setDays] = useState(
    search.get("from") || search.get("to") ? "custom" : "30",
  );
  const [timezone, setTimezone] = useState(search.get("timezone") || "UTC");
  const [listId, setListId] = useState(search.get("listId") || "all");
  const [tagId, setTagId] = useState(search.get("tagId") || "all");
  useEffect(() => {
    setListId(search.get("listId") || "all");
    setTagId(search.get("tagId") || "all");
    if (search.get("from") && search.get("to"))
      setDates({ from: search.get("from")!, to: search.get("to")! });
    if (search.get("timezone")) setTimezone(search.get("timezone")!);
  }, [search]);
  const campaignId = search.get("campaignId") || "";
  const [compare, setCompare] = useState(true);
  const params = useMemo(() => {
    const q = new URLSearchParams({ ...dates, timezone });
    if (listId !== "all") q.set("listId", listId);
    if (tagId !== "all") q.set("tagId", tagId);
    if (campaignId) q.set("campaignId", campaignId);
    return q;
  }, [dates, timezone, listId, tagId, campaignId]);
  const valid = !!dates.from && !!dates.to && dates.from < dates.to;
  const query = useAnalytics<Report>("report", params, valid);
  const options = useAnalytics<{
    lists: { id: string; name: string }[];
    tags: { id: string; name: string }[];
  }>("options", new URLSearchParams());
  const report = query.data;
  const summary = report?.summary;
  const title = views.find((v) => v.value === view)!.label;
  const scoped = new URLSearchParams(params);
  if (report) scoped.set("asOf", report.asOf);
  return (
    <div className={styles.workspace}>
      {!embedded && (
        <>
          <PageHeading
            title={
              view === "audience"
                ? "Audience analytics"
                : view === "overview"
                  ? "Analytics"
                  : title
            }
            description={
              view === "audience"
                ? "Understand who you reach, who responds, and where to focus next."
                : "Clear measures of your email marketing, from audience to delivery."
            }
            action={
              <Button
                variant="outline"
                onClick={() => void query.refetch()}
                disabled={query.isFetching || !valid}
              >
                <RefreshCw
                  className={`mr-2 h-4 w-4 ${query.isFetching ? "animate-spin" : ""}`}
                />
                Refresh
              </Button>
            }
          />
          <nav aria-label="Analytics reports" className={styles.reportNav}>
            {views.map((v) => (
              <Link
                key={v.value}
                href={`${v.href}?${params}`}
                aria-current={view === v.value ? "page" : undefined}
                className="hover:text-foreground"
              >
                {v.label}
              </Link>
            ))}
          </nav>
        </>
      )}
      <div className={styles.filters}>
        <label className={styles.filter}>
          <span>Period</span>
          <Select
            value={days}
            onValueChange={(v) => {
              setDays(v);
              if (v !== "custom") setDates(range(Number(v)));
            }}
          >
            <SelectTrigger className="w-full sm:w-36">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {[7, 30, 90].map((n) => (
                <SelectItem value={String(n)} key={n}>
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
            aria-label="From date"
            type="date"
            value={dates.from}
            onChange={(e) => {
              setDays("custom");
              setDates({ ...dates, from: e.target.value });
            }}
            className="w-full sm:w-36"
          />
        </label>
        <label className={styles.filter}>
          <span>To</span>
          <Input
            aria-label="To date"
            type="date"
            value={shiftDate(dates.to, -1)}
            onChange={(e) => {
              setDays("custom");
              setDates({ ...dates, to: shiftDate(e.target.value, 1) });
            }}
            className="w-full sm:w-36"
          />
        </label>
        <FilterSelect
          label="Timezone"
          value={timezone}
          onChange={setTimezone}
          options={[
            ...new Set([
              "UTC",
              localTimezone,
              "America/New_York",
              "America/Los_Angeles",
              "Europe/London",
              "Asia/Kolkata",
              "Asia/Singapore",
              "Australia/Sydney",
              timezone,
            ]),
          ].map((id) => ({ id, name: id.replaceAll("_", " ") }))}
        />
        <FilterSelect
          label="Contact list"
          value={listId}
          onChange={setListId}
          options={[
            { id: "all", name: "All lists" },
            ...(options.data?.lists || []),
          ]}
        />
        <FilterSelect
          label="Tag"
          value={tagId}
          onChange={setTagId}
          options={[
            { id: "all", name: "All tags" },
            ...(options.data?.tags || []),
          ]}
        />
        <label className="col-span-2 flex h-9 items-center gap-2 text-xs text-muted-foreground">
          <input
            className="accent-primary"
            type="checkbox"
            checked={compare}
            onChange={(e) => setCompare(e.target.checked)}
          />
          Compare previous period
        </label>
        {embedded && (
          <Button
            variant="outline"
            disabled={query.isFetching || !valid}
            onClick={() => void query.refetch()}
          >
            <RefreshCw
              className={`mr-2 h-4 w-4 ${query.isFetching ? "animate-spin" : ""}`}
            />
            Refresh
          </Button>
        )}
      </div>
      {options.error && (
        <p role="alert" className="text-sm text-destructive">
          Filters unavailable.{" "}
          <button className="underline" onClick={() => void options.refetch()}>
            Try again
          </button>
        </p>
      )}
      {!valid ? (
        <p role="alert" className="text-sm text-destructive">
          Choose an end date after the start date.
        </p>
      ) : query.isPending || query.error ? (
        <QueryState
          loading={query.isPending}
          error={query.error}
          retry={() => void query.refetch()}
        />
      ) : (
        report &&
        summary && (
          <>
            <div className={styles.context}>
              <span>
                Updated{" "}
                {new Date(report.asOf).toLocaleString(undefined, {
                  timeZone: timezone,
                })}{" "}
                · {timezone} · Marketing campaigns only
              </span>
              {campaignId && (
                <Link
                  className="text-foreground underline"
                  href="/analytics/campaigns"
                >
                  Clear campaign filter
                </Link>
              )}
            </div>
            <div className={styles.metrics}>
              {view === "audience" ? (
                <>
                  <MetricCard
                    label="Active subscribers"
                    value={number(summary.activeSubscribers)}
                    detail="Unique addresses active now, excluding workspace suppressions."
                    icon={Users}
                  />
                  <MetricCard
                    label="New contact records"
                    value={number(summary.newContactRecords)}
                    detail="Includes imports and new list memberships; not necessarily new opt-ins."
                    icon={Users}
                    change={
                      compare
                        ? `${number(summary.newContactRecords - report.previous.newContactRecords)} vs previous period`
                        : undefined
                    }
                  />
                  <MetricCard
                    label="Recipients reached"
                    value={number(summary.recipientsReached)}
                    detail="Unique recipients of SMTP-accepted messages in this period."
                    icon={Send}
                  />
                  <MetricCard
                    label="Audience click rate"
                    value={formatRate(summary.audienceClickRate)}
                    detail={`${number(summary.recipientsClicked)} of ${number(summary.recipientsReached)} reached recipients clicked.`}
                    icon={MousePointer2}
                    change={
                      compare
                        ? `${rateDifference(summary.audienceClickRate, report.previous.audienceClickRate)} vs previous period`
                        : undefined
                    }
                  />
                </>
              ) : view === "delivery" ? (
                <>
                  <MetricCard
                    label="SMTP accepted"
                    value={number(summary.accepted)}
                    detail="Messages acknowledged by your SMTP server."
                    icon={Send}
                  />
                  <MetricCard
                    label="Known bounce rate"
                    value={formatRate(summary.bounceRate)}
                    detail={`${number(summary.bounced)} known bounces / ${number(summary.accepted)} accepted messages.`}
                    icon={ShieldCheck}
                  />
                  <MetricCard
                    label="Known complaints"
                    value={number(summary.complaints)}
                    detail={`${formatRate(summary.complaintRate)} of accepted messages; captured feedback only.`}
                    icon={ShieldCheck}
                  />
                  <MetricCard
                    label="Needs investigation"
                    value={number(summary.failed + summary.deliveryUnknown)}
                    detail={`${number(summary.failed)} failed and ${number(summary.deliveryUnknown)} unknown outcomes, created in this period.`}
                    icon={Info}
                  />
                </>
              ) : (
                <>
                  <MetricCard
                    label="Messages accepted"
                    value={number(summary.accepted)}
                    detail="Non-test marketing messages accepted by SMTP."
                    icon={Send}
                  />
                  <MetricCard
                    label="Recipients reached"
                    value={number(summary.recipientsReached)}
                    detail="Deduplicated across selected campaigns and lists."
                    icon={Users}
                  />
                  <MetricCard
                    label="Message click rate"
                    value={formatRate(summary.clickRate)}
                    detail={`${number(summary.clickedMessages)} clicked / ${number(summary.accepted)} accepted messages.`}
                    icon={MousePointer2}
                    change={
                      compare
                        ? `${rateDifference(summary.clickRate, report.previous.clickRate)} vs previous period`
                        : undefined
                    }
                  />
                  <MetricCard
                    label="Tracked open rate"
                    value={formatRate(summary.openRate)}
                    detail="Directional only; privacy proxies can inflate opens."
                    icon={BarChart3}
                  />
                </>
              )}
            </div>
            {summary.accepted === 0 && (
              <div className={styles.empty}>
                <div>
                  <h2 className="font-medium">
                    No accepted messages in this period
                  </h2>
                  <p className="mt-1 text-sm text-muted-foreground">
                    Try another date range or prepare your next campaign.
                    Unavailable rates appear as —.
                  </p>
                </div>
                <Button asChild variant="outline">
                  <Link href="/campaigns/new">
                    Create campaign draft
                    <ArrowUpRight className="ml-2 h-4 w-4" />
                  </Link>
                </Button>
              </div>
            )}
            {view !== "delivery" && (
              <AnalyticsCharts
                key={params.toString()}
                report={report}
                scope={scoped.toString()}
                tagFiltered={tagId !== "all"}
              />
            )}
            <Breakdown
              key={`${view}-${params}`}
              params={scoped}
              kind={
                view === "audience"
                  ? "lists"
                  : view === "delivery"
                    ? "domains"
                    : "campaigns"
              }
              title={
                view === "audience"
                  ? "List health"
                  : view === "delivery"
                    ? "Sending domains"
                    : "Campaigns & newsletter editions"
              }
            />
            {view === "audience" && (
              <Breakdown
                key={`sources-${params}`}
                params={scoped}
                kind="sources"
                title="Acquisition sources"
              />
            )}
            {(view === "campaigns" || campaignId) && (
              <Breakdown
                key={`links-${params}`}
                params={scoped}
                kind="links"
                title="Links people clicked"
              />
            )}
            {view === "delivery" && (
              <Card className={styles.panel}>
                <CardHeader>
                  <CardTitle className={styles.panelTitle}>
                    Tracking coverage
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-3 text-sm">
                  <p>
                    {number(summary.trackingKnown)} of{" "}
                    {number(summary.accepted)} accepted messages have a recorded
                    tracking setting; {number(summary.trackingEnabled)} had
                    click tracking enabled.
                  </p>
                  <p className="text-muted-foreground">
                    No inbox-placement or verified-human metrics are available.
                    Older messages may have working tracking without a recorded
                    setting.
                  </p>
                  <Button asChild variant="outline">
                    <Link href="/developer/logs/emails">
                      Review outbox
                      <ArrowUpRight className="ml-2 h-4 w-4" />
                    </Link>
                  </Button>
                </CardContent>
              </Card>
            )}
            <details className="rounded-xl border bg-muted/20 p-4 text-xs text-muted-foreground">
              <summary className="cursor-pointer font-medium text-foreground">
                How to read these numbers
              </summary>
              <ul className="mt-3 list-disc space-y-2 pl-5">
                {report.warnings.map((w) => (
                  <li key={w}>{w}</li>
                ))}
                <li>
                  Summary rates follow messages sent in the period, with
                  outcomes observed up to its cutoff. The previous period is
                  closed at its own cutoff; recent sends have had less time to
                  accumulate activity.
                </li>
                <li>
                  “No clicks after 3+ tracked sends” is a transparent review
                  rule, not a churn prediction. Cohort actions open contacts for
                  review and never send automatically.
                </li>
              </ul>
            </details>
          </>
        )
      )}
    </div>
  );
}
function FilterSelect({
  label,
  value,
  onChange,
  options,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
  options: { id: string; name: string }[];
}) {
  return (
    <label className={styles.filter}>
      <span>{label}</span>
      <Select value={value} onValueChange={onChange}>
        <SelectTrigger aria-label={label} className="w-full sm:w-44">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {options.map((o) => (
            <SelectItem key={o.id} value={o.id}>
              {o.name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </label>
  );
}
function Breakdown({
  params,
  kind,
  title,
}: {
  params: URLSearchParams;
  kind: string;
  title: string;
}) {
  const [page, setPage] = useState(1);
  const [sort, setSort] = useState(kind === "sources" ? "new" : "accepted");
  const [direction, setDirection] = useState("desc");
  const queryParams = new URLSearchParams(params);
  queryParams.set("kind", kind);
  queryParams.set("page", String(page));
  queryParams.set("limit", "10");
  queryParams.set("sort", sort);
  queryParams.set("direction", direction);
  const query = useAnalytics<AnalyticsPage<BreakdownRow>>(
    "breakdown",
    queryParams,
  );
  const data = query.data;
  function header(label: string, key: string) {
    return (
      <button
        className="whitespace-nowrap text-left font-medium hover:text-foreground"
        onClick={() => {
          setSort(key);
          setDirection(sort === key && direction === "desc" ? "asc" : "desc");
          setPage(1);
        }}
      >
        {label}
        {sort === key ? (direction === "desc" ? " ↓" : " ↑") : ""}
      </button>
    );
  }
  function csv() {
    if (!data) return;
    const rows = [
      [
        "Name",
        "Accepted",
        "Unique clicking recipients",
        "Clicked messages",
        "Known bounces",
        "Known opt-outs",
        "Known complaints",
      ],
      ...data.items.map((r) => [
        r.name,
        r.accepted,
        r.clicked,
        r.clickedMessages,
        r.bounced,
        r.unsubscribes,
        r.complaints,
      ]),
    ];
    const escaped = rows
      .map((row) =>
        row
          .map(
            (v) =>
              '"' +
              String(v)
                .replace(/^[=+@-]/, "'$&")
                .replaceAll('"', '""') +
              '"',
          )
          .join(","),
      )
      .join("\r\n");
    const url = URL.createObjectURL(
      new Blob([escaped], { type: "text/csv;charset=utf-8" }),
    );
    const a = document.createElement("a");
    a.href = url;
    a.download = `xem-${kind}-page-${page}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  }
  return (
    <Card className={styles.panel}>
      <CardHeader className={styles.panelHeader}>
        <div>
          <CardTitle className={styles.panelTitle}>{title}</CardTitle>
          <CardDescription className={styles.panelDescription}>
            {kind === "sources"
              ? "New records grouped by known source, with subsequent campaign engagement. Unknown sources remain visible."
              : kind === "links"
                ? "Unique message clicks for each destination. Links can share clickers."
                : "Rates include their numerator and denominator. Select a name to investigate."}
          </CardDescription>
        </div>
        <Button
          size="sm"
          variant="ghost"
          onClick={csv}
          disabled={!data?.items.length}
        >
          <ArrowDownToLine className="mr-2 h-4 w-4" />
          Export page
        </Button>
      </CardHeader>
      <CardContent className="p-0">
        {query.isPending || query.error ? (
          <div className="px-5 pb-5">
            <QueryState
              loading={query.isPending}
              error={query.error}
              retry={() => void query.refetch()}
            />
          </div>
        ) : !data?.items.length ? (
          <p className="p-6 text-sm text-muted-foreground">
            No matching {kind} in this period.
          </p>
        ) : (
          <>
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="pl-6">
                      {header(
                        kind === "links" ? "Destination" : "Name",
                        "name",
                      )}
                    </TableHead>
                    {kind === "lists" && (
                      <TableHead>
                        {header("Active subscribers", "active")}
                      </TableHead>
                    )}
                    {kind === "sources" && (
                      <TableHead>{header("New records", "new")}</TableHead>
                    )}
                    <TableHead>{header("Accepted", "accepted")}</TableHead>
                    <TableHead>
                      {header("Clicking recipients", "clicked")}
                    </TableHead>
                    <TableHead>
                      {header("Message click rate", "clickRate")}
                    </TableHead>
                    {kind === "lists" && (
                      <TableHead>Audience click rate</TableHead>
                    )}
                    {kind !== "links" && (
                      <TableHead>
                        {header("Known bounces", "bounced")}
                      </TableHead>
                    )}
                    {kind !== "links" && (
                      <>
                        <TableHead>Known opt-outs</TableHead>
                        <TableHead>Known complaints</TableHead>
                      </>
                    )}
                    {kind === "lists" && (
                      <TableHead>Suppressed memberships</TableHead>
                    )}
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data.items.map((r) => {
                    const scope = new URLSearchParams(params);
                    if (kind === "campaigns") scope.set("campaignId", r.id);
                    if (kind === "lists") scope.set("listId", r.id);
                    return (
                      <TableRow key={r.id}>
                        <TableCell className="max-w-72 pl-6">
                          <div className="truncate font-medium" title={r.name}>
                            {kind === "campaigns" || kind === "lists" ? (
                              <Link
                                className="hover:text-primary"
                                href={`${kind === "campaigns" ? "/analytics/campaigns" : "/audience/dashboard"}?${scope}`}
                              >
                                {r.name} ↗
                              </Link>
                            ) : (
                              r.name
                            )}
                          </div>
                          {r.kind === "newsletter" && (
                            <span className="text-xs text-muted-foreground">
                              Newsletter edition
                            </span>
                          )}
                        </TableCell>
                        {kind === "lists" && (
                          <TableCell>{number(r.active)}</TableCell>
                        )}
                        {kind === "sources" && (
                          <TableCell>{number(r.new)}</TableCell>
                        )}
                        <TableCell>{number(r.accepted)}</TableCell>
                        <TableCell>{number(r.clicked)}</TableCell>
                        <TableCell>
                          <RateValue rate={r.clickRate} />
                        </TableCell>
                        {kind === "lists" && (
                          <TableCell>
                            <RateValue rate={r.audienceClickRate} />
                          </TableCell>
                        )}
                        {kind !== "links" && (
                          <TableCell>{number(r.bounced)}</TableCell>
                        )}
                        {kind !== "links" && (
                          <>
                            <TableCell>{number(r.unsubscribes)}</TableCell>
                            <TableCell>{number(r.complaints)}</TableCell>
                          </>
                        )}
                        {kind === "lists" && (
                          <TableCell>{number(r.suppressed)}</TableCell>
                        )}
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            </div>
            <div className="flex items-center justify-between border-t px-6 py-4 text-xs text-muted-foreground">
              <span>
                {number(data.totalCount)} results · Page {data.page} of{" "}
                {Math.max(1, Math.ceil(data.totalCount / data.pageSize))}
              </span>
              <div className="flex gap-2">
                <Button
                  size="sm"
                  variant="outline"
                  aria-label={`Previous ${kind} page`}
                  disabled={data.page <= 1 || query.isFetching}
                  onClick={() => setPage(data.page - 1)}
                >
                  <ChevronLeft className="h-4 w-4" />
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  aria-label={`Next ${kind} page`}
                  disabled={
                    data.page * data.pageSize >= data.totalCount ||
                    query.isFetching
                  }
                  onClick={() => setPage(data.page + 1)}
                >
                  <ChevronRight className="h-4 w-4" />
                </Button>
              </div>
            </div>
          </>
        )}
      </CardContent>
    </Card>
  );
}
