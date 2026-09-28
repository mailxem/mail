"use client";
import {
  useEffect,
  useLayoutEffect,
  useId,
  useMemo,
  useRef,
  useState,
} from "react";
import {
  Area,
  AreaChart,
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Label,
  Pie,
  PieChart,
  XAxis,
  YAxis,
} from "recharts";
import {
  ArrowDownToLine,
  ArrowRight,
  ArrowUpRight,
  ArrowUpDown,
  Check,
  ChevronDown,
  ChartNoAxesCombined,
  MousePointer2,
  RotateCcw,
  Send,
  Users,
} from "lucide-react";
import { gsap } from "gsap";
import { ScrollTrigger } from "gsap/ScrollTrigger";
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from "./ui/chart";
import {
  campaigns,
  compact,
  getDemo,
  number,
  palette,
  type SendType,
} from "@/lib/analytics-demo";
import { appLink } from "@/lib/site";
const chartConfig = {
  newsletters: { label: "Newsletters", color: palette.newsletters },
  campaigns: { label: "Campaigns", color: palette.campaigns },
  automations: { label: "Automations", color: palette.automations },
  newslettersClicks: { label: "Newsletter clicks", color: palette.newsletters },
  campaignsClicks: { label: "Campaign clicks", color: palette.campaigns },
  automationsClicks: { label: "Automation clicks", color: palette.automations },
  total: { label: "Subscribers", color: palette.clicks },
  value: { label: "Tracked clicks" },
} satisfies ChartConfig;
const sources: SendType[] = ["newsletters", "campaigns", "automations"];
const daysOfWeek = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
const windows = ["00–04", "04–08", "08–12", "12–16", "16–20", "20–24"];
function Stat({
  label,
  value,
  detail,
  change,
  icon: Icon,
  percentage = false,
  reduced,
}: {
  label: string;
  value: number;
  detail: string;
  change?: number;
  icon: typeof Send;
  percentage?: boolean;
  reduced: boolean;
}) {
  const ref = useRef<HTMLSpanElement>(null);
  const previous = useRef(0);
  useLayoutEffect(() => {
    const current = { value: previous.current };
    const format = (n: number) => (percentage ? `${n.toFixed(1)}%` : number(n));
    if (reduced) {
      if (ref.current) ref.current.textContent = format(value);
      previous.current = value;
      return;
    }
    const tween = gsap.to(current, {
      value,
      duration: 0.85,
      ease: "power2.out",
      onUpdate: () => {
        if (ref.current) ref.current.textContent = format(current.value);
      },
    });
    previous.current = value;
    return () => {
      tween.kill();
    };
  }, [value, reduced, percentage]);
  return (
    <div className="min-w-0 rounded-2xl border border-[#e8e6ed] bg-white p-4 sm:p-5">
      <div className="flex items-center justify-between gap-2 text-[11px] text-muted">
        <span>{label}</span>
        <Icon size={15} className="text-[#80758c]" />
      </div>
      <div className="mt-3 flex flex-wrap items-center gap-2">
        <span
          ref={ref}
          data-stat={label}
          className="text-[28px] font-medium leading-none tracking-[-.055em] tabular-nums sm:text-[32px]"
        >
          {percentage ? `${value.toFixed(1)}%` : number(value)}
        </span>
        {change !== undefined && (
          <span
            className={`inline-flex items-center gap-0.5 rounded-full px-1.5 py-1 text-[9px] font-medium ${change >= 0 ? "bg-[#eaf3ed] text-[#2d6a50]" : "bg-[#faeee6] text-[#885133]"}`}
          >
            <ArrowUpRight size={11} className={change < 0 ? "rotate-90" : ""} />
            {Math.abs(change).toFixed(1)}%
          </span>
        )}
      </div>
      <p className="mt-2 text-[10px] leading-relaxed text-muted">{detail}</p>
    </div>
  );
}
function Dot({ color }: { color: string }) {
  return (
    <svg aria-hidden="true" width="9" height="9" className="shrink-0">
      <rect width="9" height="9" rx="2.5" fill={color} />
    </svg>
  );
}
export default function AnalyticsShowcase() {
  const [period, setPeriod] = useState(30);
  const [campaign, setCampaign] = useState("all");
  const [mode, setMode] = useState<"accepted" | "clicks">("accepted");
  const [active, setActive] = useState<SendType[]>([...sources]);
  const [device, setDevice] = useState<number | null>(null);
  const [sort, setSort] = useState<"accepted" | "clicks" | "rate">("accepted");
  const [ascending, setAscending] = useState(false);
  const [heat, setHeat] = useState<[number, number]>([3, 2]);
  const [reduced, setReduced] = useState(
    () =>
      typeof window === "undefined" ||
      window.matchMedia("(prefers-reduced-motion: reduce)").matches,
  );
  const gradient = useId().replace(/:/g, "");
  useEffect(() => {
    const frame = requestAnimationFrame(() => ScrollTrigger.refresh());
    return () => cancelAnimationFrame(frame);
  }, [campaign]);
  useEffect(() => {
    const media = window.matchMedia("(prefers-reduced-motion: reduce)");
    const update = () => setReduced(media.matches);
    update();
    media.addEventListener("change", update);
    return () => media.removeEventListener("change", update);
  }, []);
  const { daily, totals, table, devices, growth, heatmap, cutoff, previous } =
    useMemo(() => getDemo(period, campaign), [period, campaign]);
  const sorted = [...table].sort(
    (a, b) => (ascending ? 1 : -1) * (a[sort] - b[sort]),
  );
  const net = totals.subscribers - totals.unsubscribed;
  const clickRate = (totals.clicks / totals.accepted) * 100;
  const acceptance =
    (totals.accepted / (totals.accepted + totals.rejected + totals.deferred)) *
    100;
  const heatMax = Math.max(...heatmap.flat());
  const selectedDevice = device === null ? null : devices[device];
  const delta = (n: number, p: number) => ((n - p) / p) * 100;
  const reset = () => {
    setPeriod(30);
    setCampaign("all");
    setMode("accepted");
    setActive([...sources]);
    setDevice(null);
    setHeat([3, 2]);
    setSort("accepted");
    setAscending(false);
  };
  const exportCSV = () => {
    const body = [
      "Xem synthetic demo analytics — not customer results",
      `Period: ${cutoff} to 2026-09-10`,
      "Campaign,Messages accepted,Tracked clicks,Click rate",
      ...table.map(
        (r) => `"${r.name}",${r.accepted},${r.clicks},${r.rate.toFixed(2)}%`,
      ),
    ].join("\r\n");
    const url = URL.createObjectURL(
      new Blob([body], { type: "text/csv;charset=utf-8;" }),
    );
    const a = document.createElement("a");
    a.href = url;
    a.download = `xem-demo-analytics-${period}-days.csv`;
    a.click();
    window.setTimeout(() => URL.revokeObjectURL(url), 1000);
  };
  return (
    <div className="overflow-hidden rounded-[24px] border border-[#dedce3] bg-[#f7f6fa] text-ink shadow-[0_30px_80px_-50px_#45395465] md:rounded-[30px]">
      <div className="flex flex-wrap items-center justify-between gap-4 border-b border-[#e7e4ed] bg-white/70 px-5 py-5 md:px-7">
        <div className="flex items-center gap-3">
          <div className="rounded-xl bg-[#eee7fa] p-2.5 text-[#7254a9]">
            <ChartNoAxesCombined size={21} />
          </div>
          <div>
            <h3 className="text-sm font-semibold">Your audience, in focus.</h3>
            <p className="mt-1 text-[10px] text-muted">
              Explore the numbers. Find your next move.
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2 rounded-full border border-[#dadfcf] bg-[#f0f4e5] px-3 py-1.5 text-[10px] text-[#466343]">
          <span className="h-1.5 w-1.5 rounded-full bg-[#56804c]" /> Interactive
          demo · Synthetic data
        </div>
      </div>
      <div className="p-4 md:p-6">
        <div className="mb-5 flex flex-wrap items-center justify-between gap-3">
          <div className="flex flex-wrap items-center gap-3">
            <div
              aria-label="Analytics date range"
              className="flex rounded-lg border border-[#e2dfe9] bg-white p-1"
            >
              {[7, 30, 90].map((n) => (
                <button
                  key={n}
                  onClick={() => setPeriod(n)}
                  aria-pressed={period === n}
                  className={`rounded-md px-3 py-2 text-[11px] transition-colors ${period === n ? "bg-[#eee7fa] font-medium text-[#694798]" : "text-muted hover:bg-[#f5f3f9]"}`}
                >
                  {n} days
                </button>
              ))}
            </div>
            <div className="relative">
              <label htmlFor="demo-campaign" className="sr-only">
                Filter analytics by campaign
              </label>
              <select
                id="demo-campaign"
                value={campaign}
                onChange={(e) => {
                  setCampaign(e.target.value);
                  setActive([...sources]);
                }}
                className="max-w-full appearance-none rounded-lg border border-[#e2dfe9] bg-white py-3 pl-3 pr-9 text-[11px]"
              >
                <option value="all">All campaigns & newsletters</option>
                {campaigns.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
              <ChevronDown
                className="pointer-events-none absolute right-3 top-3.5"
                size={12}
              />
            </div>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={reset}
              aria-label="Reset analytics filters"
              title="Reset filters"
              className="rounded-lg border border-[#e2dfe9] bg-white p-2.5 text-muted hover:bg-[#f0eaf8]"
            >
              <RotateCcw size={14} />
            </button>
            <button
              onClick={exportCSV}
              className="flex items-center gap-2 rounded-lg border border-[#e2dfe9] bg-white px-3 py-2.5 text-[11px] hover:bg-[#f0eaf8]"
            >
              <ArrowDownToLine size={13} />
              Export demo data
            </button>
          </div>
        </div>
        <p aria-live="polite" className="mb-4 text-[10px] text-muted">
          {cutoff} — 2026-09-10 ·{" "}
          {campaign === "all"
            ? "All six demo campaigns"
            : campaigns.find((c) => c.id === campaign)?.name}{" "}
          · UTC
        </p>
        <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
          <Stat
            label="Messages accepted"
            value={totals.accepted}
            detail={`vs. the previous ${period} days`}
            change={delta(totals.accepted, previous.accepted)}
            icon={Send}
            reduced={reduced}
          />
          <Stat
            label="Message click rate"
            value={clickRate}
            detail={`${number(totals.clicks)} tracked clicks`}
            icon={MousePointer2}
            percentage
            reduced={reduced}
          />
          <Stat
            label="New subscribers"
            value={totals.subscribers}
            detail={`${number(net)} net after unsubscribes`}
            change={delta(totals.subscribers, previous.subscribers)}
            icon={Users}
            reduced={reduced}
          />
          <Stat
            label="SMTP acceptance"
            value={acceptance}
            detail={`${number(totals.rejected)} rejected · ${number(totals.deferred)} deferred`}
            icon={Check}
            percentage
            reduced={reduced}
          />
        </div>
        <div className="mt-4 rounded-2xl border border-[#e8e6ed] bg-white p-4 md:p-6">
          <div className="mb-6 flex flex-wrap items-start justify-between gap-4">
            <div>
              <h4 className="text-sm font-semibold">
                A little momentum, every day.
              </h4>
              <p className="mt-1 text-[11px] text-muted">
                {mode === "accepted" ? "Messages accepted" : "Tracked clicks"},
                stacked by send type. Hover to explore.
              </p>
            </div>
            <div
              className="flex rounded-lg bg-[#f3f1f7] p-1"
              aria-label="Chart metric"
            >
              {(["accepted", "clicks"] as const).map((m) => (
                <button
                  key={m}
                  onClick={() => setMode(m)}
                  aria-pressed={mode === m}
                  className={`rounded-md px-3 py-1.5 text-[10px] ${mode === m ? "bg-white font-medium text-ink shadow-sm" : "text-muted"}`}
                >
                  {m === "accepted" ? "Messages" : "Clicks"}
                </button>
              ))}
            </div>
          </div>
          <div className="mb-4 flex flex-wrap items-center gap-x-5 gap-y-2">
            {sources.map((s) => (
              <button
                key={s}
                onClick={() =>
                  setActive((items) =>
                    items.includes(s)
                      ? items.length > 1
                        ? items.filter((x) => x !== s)
                        : items
                      : [...items, s],
                  )
                }
                aria-pressed={active.includes(s)}
                aria-label={`Toggle ${s} series`}
                disabled={active.length === 1 && active.includes(s)}
                className={`flex items-center gap-2 rounded px-1 py-1 text-[11px] ${active.includes(s) ? "text-ink" : "text-muted line-through"}`}
              >
                <Dot color={active.includes(s) ? palette[s] : "#b8b6bc"} />
                {chartConfig[s].label}
              </button>
            ))}
            <span className="text-[9px] text-muted">
              Click a legend to toggle chart series
            </span>
          </div>
          <ChartContainer
            config={chartConfig}
            className="aspect-auto h-[260px] w-full md:h-[300px]"
          >
            <BarChart
              accessibilityLayer
              data={daily}
              margin={{ top: 12, right: 0, left: -18, bottom: 0 }}
              barCategoryGap={period === 90 ? "14%" : "22%"}
            >
              <CartesianGrid
                vertical={false}
                stroke="#eeeaf1"
                strokeDasharray="3 4"
              />
              <XAxis
                dataKey="label"
                tickLine={false}
                axisLine={false}
                minTickGap={35}
                tickMargin={12}
                fontSize={10}
              />
              <YAxis
                tickLine={false}
                axisLine={false}
                tickFormatter={compact}
                fontSize={10}
                width={55}
              />
              <ChartTooltip
                cursor={{ fill: "#f5f2fa" }}
                content={<ChartTooltipContent className="min-w-[180px] p-3" />}
              />
              {sources
                .filter((s) => active.includes(s))
                .map((s, i, all) => (
                  <Bar
                    key={`${s}-${mode}`}
                    dataKey={mode === "accepted" ? s : `${s}Clicks`}
                    stackId="volume"
                    fill={palette[s]}
                    radius={i === all.length - 1 ? [4, 4, 0, 0] : [0, 0, 0, 0]}
                    isAnimationActive={!reduced}
                    animationDuration={750}
                    animationBegin={i * 80}
                  />
                ))}
            </BarChart>
          </ChartContainer>
        </div>
        <div className="mt-4 grid gap-4 lg:grid-cols-[1.35fr_1fr]">
          <div className="rounded-2xl border border-[#e8e6ed] bg-white p-4 md:p-6">
            <div className="flex items-start justify-between">
              <div>
                <h4 className="text-sm font-semibold">
                  Room for a few more good people.
                </h4>
                <p className="mt-1 text-[11px] text-muted">
                  Audience growth · net subscribers over time
                </p>
              </div>
              <span className="rounded-full bg-[#edf5ed] px-2 py-1 text-[10px] font-medium text-[#3b704c]">
                +{compact(net)}
              </span>
            </div>
            <ChartContainer
              config={chartConfig}
              className="mt-6 aspect-auto h-[210px] w-full"
            >
              <AreaChart
                accessibilityLayer
                data={growth}
                margin={{ top: 8, right: 6, bottom: 0, left: -12 }}
              >
                <defs>
                  <linearGradient
                    id={`growth-${gradient}`}
                    x1="0"
                    y1="0"
                    x2="0"
                    y2="1"
                  >
                    <stop offset="0%" stopColor="#739e88" stopOpacity={0.45} />
                    <stop
                      offset="100%"
                      stopColor="#739e88"
                      stopOpacity={0.04}
                    />
                  </linearGradient>
                </defs>
                <CartesianGrid
                  vertical={false}
                  stroke="#eeeaf1"
                  strokeDasharray="3 4"
                />
                <XAxis
                  dataKey="label"
                  axisLine={false}
                  tickLine={false}
                  minTickGap={55}
                  tickMargin={10}
                  fontSize={10}
                />
                <YAxis
                  domain={["dataMin - 300", "dataMax + 300"]}
                  tickFormatter={compact}
                  axisLine={false}
                  tickLine={false}
                  fontSize={10}
                />
                <ChartTooltip content={<ChartTooltipContent />} />
                <Area
                  dataKey="total"
                  type="monotone"
                  stroke="#467d66"
                  strokeWidth={2.5}
                  fill={`url(#growth-${gradient})`}
                  isAnimationActive={!reduced}
                  animationDuration={900}
                />
              </AreaChart>
            </ChartContainer>
          </div>
          <div className="rounded-2xl border border-[#e8e6ed] bg-white p-4 md:p-6">
            <h4 className="text-sm font-semibold">Meet them where they are.</h4>
            <p className="mt-1 text-[11px] text-muted">
              Share of tracked clicks by device
            </p>
            <div className="flex flex-wrap items-center justify-center gap-4">
              <ChartContainer
                config={{
                  value: { label: "Clicks" },
                  Mobile: { label: "Mobile" },
                  Desktop: { label: "Desktop" },
                  Other: { label: "Other" },
                }}
                className="aspect-square h-[210px] w-[210px] shrink-0"
              >
                <PieChart
                  accessibilityLayer
                  aria-label="Tracked clicks by device"
                >
                  <ChartTooltip
                    content={<ChartTooltipContent nameKey="name" hideLabel />}
                  />
                  <Pie
                    data={devices}
                    dataKey="value"
                    nameKey="name"
                    innerRadius={66}
                    outerRadius={87}
                    paddingAngle={4}
                    stroke="white"
                    strokeWidth={3}
                    isAnimationActive={!reduced}
                    animationDuration={800}
                    onMouseEnter={(_, i) => setDevice(i)}
                  >
                    {devices.map((d, i) => (
                      <Cell
                        key={d.key}
                        aria-label={`${d.name}: ${number(d.value)} tracked clicks`}
                        fill={d.fill}
                        opacity={device === null || device === i ? 1 : 0.35}
                      />
                    ))}
                    <Label
                      content={({ viewBox }) => {
                        if (viewBox && "cx" in viewBox && "cy" in viewBox)
                          return (
                            <text
                              x={viewBox.cx}
                              y={viewBox.cy}
                              textAnchor="middle"
                              dominantBaseline="middle"
                            >
                              <tspan
                                x={viewBox.cx}
                                y={(viewBox.cy || 0) - 4}
                                className="fill-ink text-2xl font-medium"
                              >
                                {selectedDevice
                                  ? `${((selectedDevice.value / totals.clicks) * 100).toFixed(0)}%`
                                  : compact(totals.clicks)}
                              </tspan>
                              <tspan
                                x={viewBox.cx}
                                y={(viewBox.cy || 0) + 20}
                                className="fill-muted text-[10px]"
                              >
                                {selectedDevice
                                  ? selectedDevice.name
                                  : "tracked clicks"}
                              </tspan>
                            </text>
                          );
                        return null;
                      }}
                    />
                  </Pie>
                </PieChart>
              </ChartContainer>
              <div className="min-w-[120px] flex-1 space-y-2">
                {devices.map((d, i) => (
                  <button
                    key={d.key}
                    aria-pressed={device === i}
                    onClick={() => setDevice(device === i ? null : i)}
                    className={`flex w-full items-center justify-between gap-5 rounded-lg px-2 py-2 text-[11px] ${device === i ? "bg-[#f1edf8]" : "hover:bg-[#f8f7fa]"}`}
                  >
                    <span className="flex items-center gap-2">
                      <Dot color={d.fill} />
                      {d.name}
                    </span>
                    <span className="font-medium tabular-nums">
                      {((d.value / totals.clicks) * 100).toFixed(0)}%
                    </span>
                  </button>
                ))}
                <button
                  onClick={() => setDevice(null)}
                  className="px-2 text-[10px] text-muted underline underline-offset-4"
                >
                  All devices
                </button>
              </div>
            </div>
          </div>
        </div>
        <div className="mt-4 grid gap-4 xl:grid-cols-[1.3fr_1fr]">
          <div className="min-w-0 overflow-hidden rounded-2xl border border-[#e8e6ed] bg-white">
            <div className="p-5">
              <h4 className="text-sm font-semibold">
                The stories that started something.
              </h4>
              <p className="mt-1 text-[11px] text-muted">
                Sort the results. Select a campaign to explore it.
              </p>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full min-w-[490px] text-left text-[11px]">
                <thead className="border-y border-[#eeebf2] bg-[#fcfbfe] text-muted">
                  <tr>
                    <th className="px-5 py-3 font-normal">Campaign</th>
                    {(
                      [
                        ["accepted", "Messages"],
                        ["clicks", "Clicks"],
                        ["rate", "Click rate"],
                      ] as const
                    ).map(([key, label]) => (
                      <th
                        key={key}
                        aria-sort={
                          sort === key
                            ? ascending
                              ? "ascending"
                              : "descending"
                            : "none"
                        }
                        className="px-3 py-3 text-right font-normal"
                      >
                        <button
                          onClick={() => {
                            if (sort === key) setAscending(!ascending);
                            else {
                              setSort(key);
                              setAscending(false);
                            }
                          }}
                          className="inline-flex items-center gap-1"
                        >
                          {label}
                          <ArrowUpDown size={11} />
                        </button>
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {sorted.map((c) => (
                    <tr
                      key={c.id}
                      className="border-b border-[#f0edf4] last:border-0 hover:bg-[#faf9fd]"
                    >
                      <td className="px-5 py-3.5">
                        <button
                          onClick={() => setCampaign(c.id)}
                          className="group flex items-center gap-2.5 text-left"
                        >
                          <Dot color={c.color} />
                          <span>
                            <span className="block font-medium group-hover:underline">
                              {c.name}
                            </span>
                            <span className="mt-0.5 block text-[9px] capitalize text-muted">
                              {c.type}
                            </span>
                          </span>
                        </button>
                      </td>
                      <td className="px-3 py-3 tabular-nums text-right">
                        {number(c.accepted)}
                      </td>
                      <td className="px-3 py-3 tabular-nums text-right">
                        {number(c.clicks)}
                      </td>
                      <td className="px-3 py-3 text-right">
                        <span className="rounded-full bg-[#edf4ec] px-2 py-1 text-[10px] font-medium text-[#376d4c]">
                          {c.rate.toFixed(1)}%
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {campaign !== "all" && (
              <button
                className="m-5 text-[11px] text-[#715294] underline"
                onClick={() => setCampaign("all")}
              >
                Show all six campaigns
              </button>
            )}
          </div>
          <div className="rounded-2xl border border-[#e8e6ed] bg-white p-5">
            <h4 className="text-sm font-semibold">
              Find a moment that feels right.
            </h4>
            <p className="mt-1 text-[11px] text-muted">
              Click activity by day and time · UTC
            </p>
            <div className="mt-6 grid grid-cols-[30px_repeat(6,minmax(0,1fr))] gap-1.5">
              <span />
              {windows.map((w) => (
                <span key={w} className="text-center text-[8px] text-muted">
                  {w}
                </span>
              ))}
              {heatmap.map((row, d) => (
                <div key={d} className="contents">
                  <span className="flex items-center text-[9px] text-muted">
                    {daysOfWeek[d]}
                  </span>
                  {row.map((value, h) => {
                    const ratio = value / heatMax;
                    return (
                      <button
                        key={h}
                        onClick={() => setHeat([d, h])}
                        title={`${daysOfWeek[d]} ${windows[h]} UTC: ${number(value)} clicks`}
                        aria-label={`${daysOfWeek[d]} ${windows[h]} UTC: ${number(value)} demo clicks`}
                        aria-pressed={heat[0] === d && heat[1] === h}
                        className={`h-6 rounded-[4px] transition-transform hover:scale-110 focus-visible:scale-110 motion-reduce:transform-none ${ratio > 0.75 ? "bg-[#68548f]" : ratio > 0.5 ? "bg-[#9780bc]" : ratio > 0.25 ? "bg-[#c1aed9]" : "bg-[#e9e0f2]"} ${heat[0] === d && heat[1] === h ? "ring-2 ring-[#403150] ring-offset-1" : ""}`}
                      />
                    );
                  })}
                </div>
              ))}
            </div>
            <div className="mt-4 flex items-center justify-end gap-1 text-[8px] text-muted">
              Less <i className="ml-1 h-2 w-3 rounded-sm bg-[#e9e0f2]" />
              <i className="h-2 w-3 rounded-sm bg-[#c1aed9]" />
              <i className="h-2 w-3 rounded-sm bg-[#9780bc]" />
              <i className="mr-1 h-2 w-3 rounded-sm bg-[#68548f]" /> More
            </div>
            <div
              aria-live="polite"
              className="mt-5 rounded-lg bg-[#f4f0f9] px-3 py-3 text-[11px]"
            >
              <span className="font-medium">
                {daysOfWeek[heat[0]]}, {windows[heat[1]]} UTC
              </span>
              <span className="mx-2 text-muted">·</span>
              {number(heatmap[heat[0]][heat[1]])} tracked clicks
            </div>
          </div>
        </div>
        <details className="mt-4 rounded-xl border border-[#e6e1ed] bg-white">
          <summary className="cursor-pointer px-4 py-3 text-[11px] text-muted">
            View the chart data as a table
          </summary>
          <div className="max-h-64 overflow-auto border-t border-[#eee9f2]">
            <table className="w-full min-w-[450px] text-left text-[11px]">
              <caption className="sr-only">
                Daily values for the selected period and campaign
              </caption>
              <thead className="sticky top-0 bg-[#f6f3fa]">
                <tr>
                  {[
                    "Date",
                    "Messages accepted",
                    "Tracked clicks",
                    "New subscribers",
                  ].map((t) => (
                    <th key={t} className="px-4 py-2 font-medium">
                      {t}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {daily.map((d) => (
                  <tr key={d.date} className="border-t border-[#f2edf7]">
                    <td className="px-4 py-2">{d.date}</td>
                    <td className="px-4 py-2">{number(d.accepted)}</td>
                    <td className="px-4 py-2">{number(d.clicks)}</td>
                    <td className="px-4 py-2">{number(d.subscribers)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </details>
        <div className="mt-5 flex flex-wrap items-center justify-between gap-3 px-1">
          <p className="max-w-[620px] text-[10px] leading-relaxed text-muted">
            Illustrative demo data. Your workspace reports your own campaign and
            audience activity.
          </p>
          <a
            href={appLink("/analytics")}
            className="group inline-flex items-center gap-2 whitespace-nowrap text-[11px] font-medium text-[#614580]"
          >
            Open your analytics <ArrowRight size={13} />
          </a>
        </div>
      </div>
    </div>
  );
}
