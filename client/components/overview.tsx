"use client";
import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { subDays } from "date-fns";
import { Bar, BarChart, CartesianGrid, XAxis, YAxis } from "recharts";
import { ChartContainer, ChartTooltip, ChartTooltipContent, ChartLegend, ChartLegendContent } from "@/components/ui/chart";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { useTeam } from "@/app/providers/team-provider";
import { FilterTabs, QueryState, Empty } from "@/components/marketing/shared";

type TrendPoint = { date: string; openRate: number; clickRate: number; bounceRate: number; total: number; devices: Record<string, number> };
export function Overview() {
  const { team } = useTeam();
  const [period, setPeriod] = useState("This week");
  const query = useQuery<TrendPoint[]>({
    queryKey: ["overview-trends", team?.id, period], enabled: !!team?.id,
    queryFn: async ({ signal }) => {
      const days = period === "This week" ? 7 : period === "This month" ? 30 : 90;
      const response = await fetch(`/api/analytics/trends?teamId=${team!.id}&startDate=${subDays(new Date(), days).toISOString()}`, { signal });
      if (!response.ok) throw new Error("Unable to load performance trends");
      return (await response.json()).data ?? [];
    },
  });
  const data = (query.data ?? []).map(point => ({ ...point, openRate: point.openRate * 100, clickRate: point.clickRate * 100 }));
  const devices = new Map<string, number>();
  for (const point of query.data ?? []) for (const [name, count] of Object.entries(point.devices || {})) devices.set(name, (devices.get(name) || 0) + count);
  const deviceData = Array.from(devices, ([device, count]) => ({ device: device.charAt(0).toUpperCase() + device.slice(1), count }));
  return <div className="space-y-5">
    <FilterTabs items={["This week", "This month", "Last 90 days"]} value={period} onChange={setPeriod}/>
    {query.isPending || query.error ? <QueryState loading={query.isPending} error={query.error} retry={() => void query.refetch()}/> : !data.length ? <Empty title="No activity in this period" description="Your email performance will appear here as messages are delivered and opened."/> : <div className="grid gap-5 xl:grid-cols-2">
      <Card><CardHeader><CardTitle>Engagement</CardTitle><CardDescription>Open and click rates over time</CardDescription></CardHeader><CardContent>
        <ChartContainer config={{ openRate: { label: "Open rate" }, clickRate: { label: "Click rate" } }} className="h-[280px] w-full">
          <BarChart accessibilityLayer data={data}>
            <CartesianGrid vertical={false}/><XAxis dataKey="date" tickLine={false} axisLine={false} tickMargin={10} minTickGap={24} tickFormatter={value => new Date(value).toLocaleDateString(undefined, { month: "short", day: "numeric" })}/><YAxis tickLine={false} axisLine={false} width={40} tickFormatter={value => `${value}%`}/>
            <ChartTooltip content={<ChartTooltipContent labelFormatter={value => new Date(value).toLocaleDateString()} formatter={(value, name) => <span className="text-xs">{name === "openRate" ? "Open rate" : "Click rate"}: {Number(value).toFixed(1)}%</span>}/>}/>
            <Bar dataKey="openRate" fill="var(--chart-1)" radius={[4,4,0,0]}/><Bar dataKey="clickRate" fill="var(--chart-2)" radius={[4,4,0,0]}/><ChartLegend content={<ChartLegendContent/>}/>
          </BarChart>
        </ChartContainer>
      </CardContent></Card>
      <Card><CardHeader><CardTitle>Devices</CardTitle><CardDescription>Recorded email activity by device</CardDescription></CardHeader><CardContent>
        <ChartContainer config={{ count: { label: "Activity" } }} className="h-[280px] w-full">
          <BarChart accessibilityLayer data={deviceData} layout="vertical"><CartesianGrid horizontal={false}/><XAxis type="number" hide/><YAxis dataKey="device" type="category" tickLine={false} axisLine={false} width={75}/><ChartTooltip cursor={false} content={<ChartTooltipContent hideLabel/>}/><Bar dataKey="count" fill="var(--chart-1)" radius={[0,4,4,0]} maxBarSize={38}/></BarChart>
        </ChartContainer>
      </CardContent></Card>
    </div>}
  </div>;
}
