"use client";
import { useDeferredValue, useState } from "react";
import { useRouter } from "next/navigation";
import { Bar, BarChart, ResponsiveContainer, Tooltip, XAxis } from "recharts";
import { toast } from "sonner";
import { FilePlus2, Search, UsersRound } from "lucide-react";
import { useMarketing, useMarketingQuery } from "@/lib/marketing/api";
import { QueryState } from "@/components/marketing/shared";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import type { Campaign } from "@/lib";

type Hour = { hour: string; opens: number; clicks: number };
type Timeline = { buckets: Hour[]; later: Hour };

export type Recipient = {
  emailId: string;
  contactId?: string;
  email: string;
  delivery: "waiting" | "received" | "bounced" | "not_delivered";
  error?: string;
  tracked: boolean;
  sentAt?: string;
  bouncedAt?: string;
  openedAt?: string;
  opens: number;
  clickedAt?: string;
  links?: string[];
  spam: boolean;
  spamAt?: string;
  unsubscribedAt?: string;
};

type RecipientPage = {
  recipients: Recipient[];
  total: number;
  page: number;
  limit: number;
  counts: Record<string, number>;
  links: string[] | null;
};

export const FILTERS = [
  ["", "Everyone"],
  ["received", "Received"],
  ["opened", "Opened"],
  ["not_opened", "Didn't open"],
  ["clicked", "Clicked"],
  ["waiting", "Waiting"],
  ["not_delivered", "Not delivered"],
  ["bounced", "Bounced"],
  ["spam", "Marked as spam"],
  ["unsubscribed", "Unsubscribed"],
] as const;

const shortUrl = (url: string) => url.replace(/^https?:\/\//, "");
const when = (iso?: string) =>
  iso
    ? new Date(iso).toLocaleString(undefined, {
        dateStyle: "medium",
        timeStyle: "short",
      })
    : "";

/** The segment name for a filtered recipient view, at most 100 characters. */
export function audienceName(
  campaign: string,
  filter: string,
  url = "",
  q = "",
) {
  const label =
    filter === "clicked" && url
      ? `clicked ${shortUrl(url)}`
      : (FILTERS.find(([f]) => f === filter)?.[1] ?? "Everyone").toLowerCase();
  const tail = `: ${label}${q.trim() ? ` matching “${q.trim()}”` : ""}`;
  const room = 100 - Math.min(tail.length, 60);
  const name =
    campaign.length > room ? campaign.slice(0, room - 1) + "…" : campaign;
  return (name + tail).slice(0, 100);
}

/** People who first opened and first clicked, hour by hour since sending. */
export function CampaignTimeline({ campaign }: { campaign: Campaign }) {
  const live = campaign.status === "SENDING";
  const q = useMarketingQuery<Timeline>(
    `marketing/campaigns/${campaign.id}/timeline`,
    true,
    live ? 60000 : undefined,
  );
  const t = q.data;
  if (!t)
    return (
      <QueryState
        loading={q.isLoading}
        error={q.error}
        retry={() => q.refetch()}
      />
    );
  const opens = t.buckets.reduce((s, b) => s + b.opens, 0);
  const clicks = t.buckets.reduce((s, b) => s + b.clicks, 0);
  const peak = t.buckets.reduce<Hour | null>(
    (best, b) => (b.opens > (best?.opens ?? 0) ? b : best),
    null,
  );
  const hourLabel = (iso: string) =>
    new Date(iso).toLocaleString(undefined, {
      weekday: t.buckets.length > 24 ? "short" : undefined,
      hour: "numeric",
    });
  return (
    <Card>
      <CardHeader>
        <CardTitle>Opens and clicks by hour</CardTitle>
        <CardDescription>
          {t.buckets.length === 0
            ? "Nothing has been sent yet."
            : peak
              ? `${opens.toLocaleString()} people opened and ${clicks.toLocaleString()} clicked. Most opened in the hour from ${when(peak.hour)}.`
              : "No opens yet."}
        </CardDescription>
      </CardHeader>
      {t.buckets.length > 0 && (
        <CardContent className="space-y-3">
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart
                data={t.buckets}
                margin={{ top: 4, right: 0, left: 0, bottom: 0 }}
              >
                <XAxis
                  dataKey="hour"
                  tickFormatter={hourLabel}
                  tick={{ fontSize: 11 }}
                  tickLine={false}
                  axisLine={false}
                  minTickGap={24}
                />
                <Tooltip
                  cursor={{ fill: "var(--muted)" }}
                  contentStyle={{ fontSize: 12, borderRadius: 8 }}
                  labelFormatter={(iso: string) => when(iso)}
                  formatter={(v: number, k: string) => [
                    v,
                    k === "opens" ? "First opens" : "First clicks",
                  ]}
                />
                <Bar dataKey="opens" fill="#10b981" radius={[3, 3, 0, 0]} />
                <Bar dataKey="clicks" fill="#6366f1" radius={[3, 3, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>
          <div className="flex flex-wrap gap-4 text-xs text-muted-foreground">
            <span className="flex items-center gap-1.5">
              <span className="size-2 rounded-full bg-emerald-500" /> Opened
            </span>
            <span className="flex items-center gap-1.5">
              <span className="size-2 rounded-full bg-indigo-500" /> Clicked
            </span>
            <span className="basis-full sm:basis-auto sm:ml-auto">
              Each person counts once, in the hour they first opened or clicked.
              Bot opens and link-scanner clicks are left out.
              {t.later.opens + t.later.clicks > 0 &&
                ` After the first week: ${t.later.opens} more opened, ${t.later.clicks} more clicked.`}
            </span>
          </div>
        </CardContent>
      )}
    </Card>
  );
}

/** Every recipient and what they did, with saving a view as an audience. */
export function CampaignRecipients({ campaign }: { campaign: Campaign }) {
  const router = useRouter();
  const { request, refresh } = useMarketing();
  const [filter, setFilter] = useState("");
  const [url, setUrl] = useState("");
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [busy, setBusy] = useState<"" | "segment" | "follow-up">("");
  const q = useDeferredValue(search.trim());
  const params = new URLSearchParams({
    page: String(page),
    limit: "50",
    ...(filter ? { filter } : {}),
    ...(filter === "clicked" && url ? { url } : {}),
    ...(q ? { q } : {}),
  });
  const list = useMarketingQuery<RecipientPage>(
    `marketing/campaigns/${campaign.id}/recipients?${params}`,
  );
  const r = list.data;
  const audience = {
    name: audienceName(campaign.name, filter, url, q),
    filter,
    url: filter === "clicked" ? url : "",
    q,
  };

  const save = async (kind: "segment" | "follow-up") => {
    setBusy(kind);
    try {
      if (kind === "segment") {
        const s = await request<{ name: string; contactCount: number }>(
          `marketing/campaigns/${campaign.id}/recipients/segment`,
          "POST",
          audience,
        );
        toast.success(
          `Saved ${s.contactCount.toLocaleString()} ${s.contactCount === 1 ? "person" : "people"} as the segment “${s.name}”.`,
        );
      } else {
        const draft = await request<{ id: string }>(
          `marketing/campaigns/${campaign.id}/follow-up`,
          "POST",
          audience,
        );
        await refresh();
        toast.success("Follow-up drafted. Check it over, then send it.");
        router.push(`/campaigns/${draft.id}`);
      }
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Something went wrong.");
    } finally {
      setBusy("");
    }
  };

  const from = r ? (r.page - 1) * r.limit : 0;
  return (
    <Card>
      <CardHeader>
        <CardTitle>Recipients</CardTitle>
        <CardDescription>
          Everyone this campaign went to, and what they did. Bot opens and
          link-scanner clicks are left out.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div
          role="group"
          aria-label="Filter by outcome"
          className="flex flex-wrap gap-1.5"
        >
          {FILTERS.map(([f, label]) => (
            <button
              key={f}
              type="button"
              aria-pressed={filter === f}
              onClick={() => {
                setFilter(f);
                setPage(1);
              }}
              className={`rounded-full border px-3 py-1 text-xs font-medium transition-colors ${
                filter === f
                  ? "border-foreground bg-foreground text-background"
                  : "bg-background text-muted-foreground hover:text-foreground"
              }`}
            >
              {label}
              {r && (
                <span className="ml-1.5 tabular-nums opacity-70">
                  {(r.counts[f || "all"] ?? 0).toLocaleString()}
                </span>
              )}
            </button>
          ))}
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <div className="relative min-w-[12rem] flex-1">
            <Search className="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              type="search"
              aria-label="Search by email address"
              placeholder="Search by email address"
              className="pl-8"
              value={search}
              onChange={(e) => {
                setSearch(e.target.value);
                setPage(1);
              }}
            />
          </div>
          {filter === "clicked" && !!r?.links?.length && (
            <select
              aria-label="Link"
              className="h-9 max-w-xs rounded-md border bg-background px-2 text-sm"
              value={url}
              onChange={(e) => {
                setUrl(e.target.value);
                setPage(1);
              }}
            >
              <option value="">Any link</option>
              {r.links.map((l) => (
                <option key={l} value={l}>
                  {shortUrl(l)}
                </option>
              ))}
            </select>
          )}
          <Button
            variant="outline"
            size="sm"
            disabled={!r?.total || !!busy}
            onClick={() => void save("segment")}
            title={`Save these ${r?.total ?? 0} people as a segment`}
          >
            <UsersRound className="mr-1.5 size-4" />
            {busy === "segment" ? "Saving…" : "Save as segment"}
          </Button>
          <Button
            size="sm"
            disabled={!r?.total || !!busy}
            onClick={() => void save("follow-up")}
            title="Draft a copy of this campaign that goes only to these people"
          >
            <FilePlus2 className="mr-1.5 size-4" />
            {busy === "follow-up" ? "Drafting…" : "Draft follow-up"}
          </Button>
        </div>

        {!r ? (
          <QueryState
            loading={list.isLoading}
            error={list.error}
            retry={() => list.refetch()}
          />
        ) : r.recipients.length === 0 ? (
          <p className="py-8 text-center text-sm text-muted-foreground">
            {q ? `No one matches “${q}”.` : "No one here yet."}
          </p>
        ) : (
          <div className="overflow-x-auto rounded-lg border">
            <table className="w-full text-sm">
              <thead className="bg-muted/50 text-left text-xs text-muted-foreground">
                <tr>
                  <th className="px-3 py-2 font-medium">Email</th>
                  <th className="px-3 py-2 font-medium">Delivery</th>
                  <th className="px-3 py-2 font-medium">Opened</th>
                  <th className="px-3 py-2 font-medium">Clicked</th>
                  <th className="px-3 py-2 font-medium">Unsubscribed</th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {r.recipients.map((p) => (
                  <RecipientRow key={p.emailId} p={p} />
                ))}
              </tbody>
            </table>
          </div>
        )}

        {r && r.total > r.limit && (
          <div className="flex items-center justify-between text-sm text-muted-foreground">
            <span className="tabular-nums">
              {(from + 1).toLocaleString()}–
              {Math.min(from + r.limit, r.total).toLocaleString()} of{" "}
              {r.total.toLocaleString()}
            </span>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                disabled={page <= 1}
                onClick={() => setPage(page - 1)}
              >
                Previous
              </Button>
              <Button
                variant="outline"
                size="sm"
                disabled={from + r.limit >= r.total}
                onClick={() => setPage(page + 1)}
              >
                Next
              </Button>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

const DELIVERY: Record<Recipient["delivery"], [string, string]> = {
  received: ["Received", "bg-emerald-500"],
  waiting: ["Waiting", "bg-sky-400"],
  bounced: ["Bounced", "bg-amber-500"],
  not_delivered: ["Not delivered", "bg-red-600"],
};

function RecipientRow({ p }: { p: Recipient }) {
  const [label, color] = DELIVERY[p.delivery];
  const none = <span className="text-muted-foreground">—</span>;
  return (
    <tr className="align-top">
      <td className="max-w-[16rem] truncate px-3 py-2.5 font-medium">
        {p.email}
      </td>
      <td className="px-3 py-2.5">
        <span className="flex items-center gap-1.5 whitespace-nowrap">
          <span className={`size-2 rounded-full ${color}`} aria-hidden />
          {label}
        </span>
        <span
          className="block max-w-[14rem] truncate text-xs text-muted-foreground"
          title={p.error}
        >
          {p.delivery === "bounced"
            ? when(p.bouncedAt) || p.error
            : p.delivery === "not_delivered"
              ? p.error
              : when(p.sentAt)}
        </span>
        {p.spam && (
          <span className="mt-1 block text-xs font-medium text-rose-600">
            Marked as spam{p.spamAt ? ` ${when(p.spamAt)}` : ""}
          </span>
        )}
      </td>
      <td className="whitespace-nowrap px-3 py-2.5">
        {p.openedAt ? (
          <>
            {when(p.openedAt)}
            {p.opens > 1 && (
              <span className="block text-xs text-muted-foreground">
                {p.opens} times
              </span>
            )}
          </>
        ) : p.tracked ? (
          none
        ) : (
          <span className="text-xs text-muted-foreground">Not tracked</span>
        )}
      </td>
      <td className="px-3 py-2.5">
        {p.clickedAt ? (
          <>
            <span className="whitespace-nowrap">{when(p.clickedAt)}</span>
            {p.links?.map((l) => (
              <a
                key={l}
                href={l}
                target="_blank"
                rel="noreferrer"
                className="block max-w-[16rem] truncate text-xs text-muted-foreground hover:underline"
              >
                {shortUrl(l)}
              </a>
            ))}
          </>
        ) : (
          none
        )}
      </td>
      <td className="whitespace-nowrap px-3 py-2.5">
        {p.unsubscribedAt ? when(p.unsubscribedAt) : none}
      </td>
    </tr>
  );
}
