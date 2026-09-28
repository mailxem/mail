"use client";

import { useMemo, useState } from "react";
import { ArrowUpRight, Plus, RefreshCw } from "lucide-react";
import { EmailOverview } from "@/components/analytics/dashboard-home";
import type { EmailReport, EmailTotals } from "@/lib/analytics/types";
import { PageHeading, QueryState } from "@/components/marketing/shared";
import { Button } from "@/components/ui/button";
import { useConfirmSheet } from "@/components/ui/confirm-sheet";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import styles from "@/components/analytics/analytics-surface.module.css";

// Deliberately synthetic, development-only fixtures. Never used by live analytics.
function sampleReport(days: number, empty: boolean, campaigns: boolean): EmailReport {
  const series = Array.from({ length: days }, (_, index) => {
    const date = new Date(Date.UTC(2026, 8, 22 - days + index + 1)).toISOString().slice(0, 10);
    const sent = empty ? 0 : Math.round((170 + index * 13 + Math.sin(index * .85) * 95) * (campaigns ? .72 : 1));
    const queued = empty ? 0 : index % 7 === 0 ? 12 : 0;
    const failed = empty ? 0 : index % 5 === 0 ? 3 : 0;
    return { date, sent, queued, failed, drafts: 0, unknown: 0, total: sent + queued + failed, opened: Math.round(sent * .43), clicked: Math.round(sent * .08), bounced: index % 9 === 0 && !empty ? 1 : 0, campaigns: Math.round(sent * .72), other: sent - Math.round(sent * .72) };
  });
  const summary: EmailTotals = { total: 0, sent: 0, queued: 0, failed: 0, drafts: 0, unknown: 0, opened: 0, clicked: 0, bounced: 0, campaigns: 0, other: 0 };
  for (const day of series) for (const key of Object.keys(summary) as (keyof EmailTotals)[]) summary[key] += day[key];
  const rate = (numerator: number) => ({ value: summary.sent ? numerator / summary.sent : null, numerator, denominator: summary.sent });
  return { metricVersion: "email-v1", from: series[0].date, to: "2026-09-23", asOf: "2026-09-22T09:30:00Z", timezone: "UTC", summary, series, openRate: rate(summary.opened), clickRate: rate(summary.clicked), bounceRate: rate(summary.bounced) };
}

export function DashboardPreview({ onNavigate }: { onNavigate: (path: string) => void }) {
  const [days, setDays] = useState("30");
  const [tab, setTab] = useState("email");
  const [state, setState] = useState("populated");
  const confirm = useConfirmSheet();
  const report = useMemo(() => sampleReport(Number(days), state === "empty", tab === "campaigns"), [days, state, tab]);
  return (
    <div className={styles.workspace}>
      <PageHeading title="Dashboard" description="Your email activity and campaign performance, in one place." action={<div className="flex flex-wrap gap-2"><Button variant="outline" onClick={() => onNavigate("/developer/logs/emails")}>Open outbox <ArrowUpRight /></Button><Button onClick={() => onNavigate("/newsletters")}><Plus />Create newsletter</Button></div>} />
      <Tabs value={tab} onValueChange={setTab}>
        <TabsList className={styles.tabs} aria-label="Preview report source"><TabsTrigger className={styles.tab} value="email">Email</TabsTrigger><TabsTrigger className={styles.tab} value="campaigns">Campaign sample</TabsTrigger></TabsList>
      </Tabs>
      <div className={styles.filters}>
        <label className={styles.filter}><span>Period</span><Select value={days} onValueChange={setDays}><SelectTrigger className="w-36" aria-label="Report period"><SelectValue /></SelectTrigger><SelectContent>{[7,30,90].map(n => <SelectItem key={n} value={String(n)}>Last {n} days</SelectItem>)}</SelectContent></Select></label>
        <label className={styles.filter}><span>From</span><Input type="date" aria-label="Sample from date" className="w-40" value={report.from} readOnly /></label>
        <label className={styles.filter}><span>To</span><Input type="date" aria-label="Sample to date" className="w-40" value="2026-09-22" readOnly /></label>
        <span className={styles.filterNote}>Sample data · UTC</span>
        <Button variant="outline" className={styles.refresh} onClick={() => setState("populated")}><RefreshCw />Refresh</Button>
      </div>
      {state === "error" ? <QueryState loading={false} error={new Error("This is a sample connection error. Your workspace data is unchanged.")} retry={() => setState("populated")} /> : state === "loading" ? <QueryState loading /> : <EmailOverview report={report} />}
      <details className="mt-2 border-t border-border pt-4 text-xs text-muted-foreground"><summary>Preview controls</summary><div className="mt-3 flex flex-wrap gap-2">{["populated","empty","error","loading"].map(value => <Button key={value} size="sm" variant={value === state ? "secondary" : "outline"} onClick={() => setState(value)}>{value.charAt(0).toUpperCase()+value.slice(1)}</Button>)}<Button size="sm" variant="outline" onClick={async () => { await confirm({ title: "Remove sample item?", description: "This tests the inset confirmation sheet. No workspace data will be changed.", confirmLabel: "Remove sample", variant: "destructive" }); }}>Test confirmation</Button></div></details>
    </div>
  );
}
