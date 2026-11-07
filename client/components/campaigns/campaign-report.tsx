"use client";
import Link from "next/link";
import {
  ArrowUpRight,
  Check,
  Circle,
  Clock,
  Link2,
  MailCheck,
  MailOpen,
  MousePointerClick,
  ShieldAlert,
} from "lucide-react";
import { workspaceClassName } from "@/lib/workspace-styles";
import { useMarketingQuery } from "@/lib/marketing/api";
import { Metric, QueryState } from "@/components/marketing/shared";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import type { Campaign } from "@/lib";

export type CampaignReportData = {
  delivery: {
    total: number;
    queued: number;
    sent: number;
    delivered: number;
    bounced: number;
    complained: number;
    failed: number;
    suppressed: number;
    unknown: number;
  };
  unsubscribed: number;
  engagement: {
    uniqueOpens: number;
    uniqueClicks: number;
    openRate: number;
    clickRate: number;
    clickToOpenRate: number;
    untracked: number;
    botOpens: number;
    botClicks: number;
    proxyOpens: number;
    deviceBreakdown?: Record<string, number> | null;
    clickedLinks?: { url: string; uniqueClicks: number; clickRate: number }[] | null;
  };
  stages?: {
    returningLeft: number;
    new: number;
    sentToNew: number;
    bounds: [number, number];
  };
};

const pct = (n: number) => `${n >= 10 || n === 0 ? Math.round(n) : n.toFixed(1)}%`;
const share = (part: number, whole: number) => (whole ? (part / whole) * 100 : 0);
const count = (n: number, one: string, many = one + "s") =>
  `${n.toLocaleString()} ${n === 1 ? one : many}`;

/** A campaign's delivery, engagement and staged-sending progress. */
export function CampaignReport({ campaign }: { campaign: Campaign }) {
  const live = campaign.status === "SENDING" || campaign.status === "SCHEDULED";
  const q = useMarketingQuery<CampaignReportData>(
    `marketing/campaigns/${campaign.id}/report`,
    true,
    live ? 30000 : undefined,
  );
  const r = q.data;
  if (!r)
    return (
      <QueryState loading={q.isLoading} error={q.error} retry={() => q.refetch()} />
    );
  const d = r.delivery;
  const e = r.engagement;
  const reached = d.sent + d.delivered + d.bounced + d.complained;
  const machines = e.botOpens + e.botClicks + e.proxyOpens;
  return (
    <div className="space-y-6">
      <div className={workspaceClassName("metrics-grid") + " !mb-0"}>
        <Metric
          label="Delivered"
          value={(d.sent + d.delivered).toLocaleString()}
          detail={`of ${d.total.toLocaleString()} sent`}
          icon={<MailCheck />}
        />
        <Metric
          label="Opened"
          value={pct(e.openRate)}
          detail={count(e.uniqueOpens, "person", "people")}
          icon={<MailOpen />}
        />
        <Metric
          label="Clicked"
          value={pct(e.clickRate)}
          detail={`${count(e.uniqueClicks, "person", "people")} · ${pct(e.clickToOpenRate)} of opens`}
          icon={<MousePointerClick />}
        />
        <Metric
          label="Bounced"
          value={pct(share(d.bounced, reached))}
          detail={`${d.bounced} bounced · ${d.complained} spam · ${r.unsubscribed} unsubscribed`}
          icon={<ShieldAlert />}
        />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Delivery</CardTitle>
          <CardDescription>
            Where each of the {d.total.toLocaleString()} emails is now.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <DeliveryBar d={r.delivery} />
        </CardContent>
      </Card>

      {r.stages && <Stages campaign={campaign} r={r} />}

      <div className="grid gap-6 lg:grid-cols-[1.4fr_1fr]">
        <Card>
          <CardHeader>
            <CardTitle>Top links</CardTitle>
            <CardDescription>People who clicked each link.</CardDescription>
          </CardHeader>
          <CardContent>
            {e.clickedLinks?.length ? (
              <ul className="divide-y">
                {[...e.clickedLinks]
                  .sort((a, b) => b.uniqueClicks - a.uniqueClicks)
                  .slice(0, 8)
                  .map((l) => (
                    <li key={l.url} className="flex items-center gap-3 py-2.5 text-sm">
                      <Link2 className="size-4 shrink-0 text-muted-foreground" />
                      <a
                        href={l.url}
                        target="_blank"
                        rel="noreferrer"
                        className="min-w-0 flex-1 truncate hover:underline"
                      >
                        {l.url.replace(/^https?:\/\//, "")}
                      </a>
                      <span className="tabular-nums font-medium">{l.uniqueClicks}</span>
                      <span className="w-12 text-right tabular-nums text-xs text-muted-foreground">
                        {pct(l.clickRate)}
                      </span>
                    </li>
                  ))}
              </ul>
            ) : (
              <p className="py-6 text-center text-sm text-muted-foreground">
                No clicks yet.
              </p>
            )}
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Devices</CardTitle>
            <CardDescription>Where people opened and clicked.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <Bars data={e.deviceBreakdown || {}} />
            <p className="border-t pt-3 text-xs leading-relaxed text-muted-foreground">
              {machines > 0
                ? `Left out of these numbers: ${count(e.proxyOpens, "privacy-proxy open")}, ${count(e.botOpens, "bot open")} and ${count(e.botClicks, "link-scanner click")}.`
                : "Bot opens and link-scanner clicks are left out of these numbers."}
              {e.untracked > 0 && ` ${count(e.untracked, "email was", "emails were")} sent without tracking.`}
            </p>
          </CardContent>
        </Card>
      </div>

      <Link
        href={`/analytics/campaigns?campaignId=${campaign.id}`}
        className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
      >
        Audience report for this campaign <ArrowUpRight className="size-3.5" />
      </Link>
    </div>
  );
}

function DeliveryBar({ d }: { d: CampaignReportData["delivery"] }) {
  const parts = [
    ["Delivered", d.sent + d.delivered, "bg-emerald-500"],
    ["Waiting", d.queued + d.unknown, "bg-sky-400"],
    ["Bounced", d.bounced, "bg-amber-500"],
    ["Spam reports", d.complained, "bg-rose-500"],
    ["Failed", d.failed, "bg-red-600"],
    ["Suppressed", d.suppressed, "bg-zinc-400"],
  ] as const;
  if (!d.total)
    return (
      <p className="py-4 text-center text-sm text-muted-foreground">
        Nothing has been sent yet.
      </p>
    );
  return (
    <div className="space-y-4">
      <div className="flex h-3 overflow-hidden rounded-full bg-muted">
        {parts.map(([label, n, color]) =>
          n ? (
            <div
              key={label}
              className={color}
              style={{ width: `${share(n, d.total)}%` }}
              title={`${label}: ${n}`}
            />
          ) : null,
        )}
      </div>
      <dl className="grid grid-cols-2 gap-x-6 gap-y-2 text-sm sm:grid-cols-3">
        {parts.map(([label, n, color]) => (
          <div key={label} className="flex items-center gap-2">
            <span className={`size-2 rounded-full ${color}`} aria-hidden />
            <dt className="text-muted-foreground">{label}</dt>
            <dd className="ml-auto tabular-nums font-medium">{n.toLocaleString()}</dd>
          </div>
        ))}
      </dl>
    </div>
  );
}

function Bars({ data }: { data: Record<string, number> }) {
  const rows = Object.entries(data).sort((a, b) => b[1] - a[1]);
  const total = rows.reduce((s, [, n]) => s + n, 0);
  if (!total)
    return <p className="py-4 text-center text-sm text-muted-foreground">No opens yet.</p>;
  return (
    <ul className="space-y-2.5">
      {rows.slice(0, 5).map(([name, n]) => (
        <li key={name} className="text-sm">
          <div className="mb-1 flex justify-between">
            <span className="capitalize">{name || "Unknown"}</span>
            <span className="tabular-nums text-muted-foreground">{pct(share(n, total))}</span>
          </div>
          <div className="h-1.5 rounded-full bg-muted">
            <div className="h-full rounded-full bg-emerald-500" style={{ width: `${share(n, total)}%` }} />
          </div>
        </li>
      ))}
    </ul>
  );
}

/** Staged sending: returning readers first, then never-emailed readers in three stages. */
function Stages({ campaign, r }: { campaign: Campaign; r: CampaignReportData }) {
  const s = r.stages!;
  const returningSent = Math.max(0, r.delivery.total - s.sentToNew);
  const steps = [
    {
      title: "Readers you've emailed before",
      detail: `${count(returningSent + s.returningLeft, "person", "people")}, sent right away`,
      done: s.returningLeft === 0,
    },
    {
      title: "Stage 1",
      detail: `First ${count(s.bounds[0], "new reader")}`,
      done: s.sentToNew >= s.bounds[0],
    },
    {
      title: "Stage 2",
      detail: `Up to ${count(s.bounds[1], "new reader")}`,
      done: s.sentToNew >= s.bounds[1],
    },
    {
      title: "Stage 3",
      detail: `All ${count(s.new, "new reader")}`,
      done: s.sentToNew >= s.new,
    },
  ];
  const current = steps.findIndex((x) => !x.done);
  const waiting =
    campaign.status === "PAUSED"
      ? "Paused: check the warning above"
      : campaign.nextStageAt && new Date(campaign.nextStageAt).getTime() > Date.now()
        ? `Starts after ${new Date(campaign.nextStageAt).toLocaleTimeString(undefined, { hour: "numeric", minute: "2-digit" })}, if bounces and complaints stay low`
        : "Sending";
  return (
    <Card>
      <CardHeader>
        <CardTitle>Staged sending</CardTitle>
        <CardDescription>
          New readers get it in stages, so a bad list pauses before it hurts your
          sender reputation. {s.sentToNew.toLocaleString()} of{" "}
          {s.new.toLocaleString()} new readers sent so far.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <ol className="relative space-y-5">
          {steps.map((step, i) => {
            const active = i === current;
            return (
              <li key={step.title} className="relative flex gap-3">
                {i < steps.length - 1 && (
                  <span
                    aria-hidden
                    className={`absolute left-[9px] top-6 h-[calc(100%+4px)] w-px ${step.done ? "bg-emerald-500/60" : "bg-border"}`}
                  />
                )}
                <span
                  className={`z-10 grid size-5 shrink-0 place-items-center rounded-full ${
                    step.done
                      ? "bg-emerald-500 text-white"
                      : active
                        ? "bg-foreground text-background"
                        : "border bg-background text-muted-foreground"
                  }`}
                >
                  {step.done ? (
                    <Check className="size-3" />
                  ) : active ? (
                    <Clock className="size-3" />
                  ) : (
                    <Circle className="size-2" />
                  )}
                </span>
                <div className="min-w-0 text-sm">
                  <p className={step.done || active ? "font-medium" : "text-muted-foreground"}>
                    {step.title}
                  </p>
                  <p className="text-muted-foreground">
                    {step.detail}
                    {active && ` · ${waiting}`}
                  </p>
                </div>
              </li>
            );
          })}
        </ol>
      </CardContent>
    </Card>
  );
}
