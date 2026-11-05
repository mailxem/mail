"use client";
import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { format } from "date-fns";
import { toast } from "sonner";
import { BadgeDollarSign, Copy, Link2, MousePointerClick, UserPlus, Users } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Metric, QueryState, Empty } from "@/components/marketing/shared";
import { useApi } from "@/hooks/use-api";
import { workspaceClassName } from "@/lib/workspace-styles";

type Status = "pending" | "unlocked" | "paid" | "reversed";
type Summary = {
  code: string;
  clicks: number;
  signups: number;
  customers: number;
  currency: string;
  totals: Record<Status, number>;
  commissions: { customer: string; amount: number; currency: string; status: Status; paidAt: string; unlocksAt: string }[];
};
type Me = { joined: false } | { joined: true; summary: Summary };

const SITE = (process.env.NEXT_PUBLIC_WEBSITE_URL || "https://xem.email").replace(/\/$/, "");
const money = (cents: number, currency = "USD") =>
  new Intl.NumberFormat("en-US", { style: "currency", currency }).format(cents / 100);
const statusLabel: Record<Status, string> = { pending: "Pending", unlocked: "Ready to pay", paid: "Paid", reversed: "Refunded" };

export default function AffiliatesPage() {
  const { apiFetch, status } = useApi();
  const [joining, setJoining] = useState(false);
  const query = useQuery<Me | null>({
    queryKey: ["affiliates", "me"],
    enabled: status === "authenticated",
    queryFn: async () => {
      const res = await apiFetch("affiliates/me");
      if (res.status === 404) return null; // self-hosted: no Cloud billing
      if (!res.ok) throw new Error("Couldn't load your affiliate account");
      return res.json();
    },
  });

  const join = async () => {
    setJoining(true);
    try {
      const res = await apiFetch("affiliates", { method: "POST" });
      if (!res.ok) throw new Error();
      await query.refetch();
      toast.success("Your affiliate link is ready");
    } catch {
      toast.error("Couldn't create your affiliate link. Try again.");
    } finally {
      setJoining(false);
    }
  };

  if (query.isLoading || query.error)
    return <section className={workspaceClassName("product-panel")}><QueryState loading={query.isLoading} error={query.error} retry={() => void query.refetch()} /></section>;

  if (query.data === null)
    return <section className={workspaceClassName("product-panel")}><Empty title="Affiliates are part of Xem Cloud" description="Self-hosted installations don't take payments, so there are no commissions to earn here." /></section>;

  if (!query.data?.joined)
    return (
      <section className={workspaceClassName("product-panel")}>
        <Empty
          title="Earn 30% of every payment, for as long as they pay"
          description="Share your link. When someone signs up through it and pays for Xem Cloud, you get 30% of each payment, every month or year they stay. Commissions unlock 60 days after each payment and are paid by bank transfer once they reach $100."
          action={<div className="flex flex-wrap items-center justify-center gap-3">
            <Button onClick={join} disabled={joining}>{joining ? "Creating your link…" : "Get my affiliate link"}</Button>
            <a href={`${SITE}/affiliates`} target="_blank" rel="noreferrer" className="text-sm font-medium text-primary">Read the terms ↗</a>
          </div>}
        />
      </section>
    );

  const s = query.data.summary;
  const link = `${SITE}/?via=${s.code}`;
  const copy = async () => {
    try { await navigator.clipboard.writeText(link); toast.success("Link copied"); }
    catch { toast.error("Couldn't copy the link"); }
  };

  return <>
    <section className={`${workspaceClassName("product-panel")} mb-6`}>
      <div className={workspaceClassName("panel-toolbar")}><h2>Your link</h2><a href={`${SITE}/affiliates`} target="_blank" rel="noreferrer" className="text-xs font-medium text-primary">Program terms ↗</a></div>
      <div className="flex flex-wrap items-center gap-3 rounded-xl bg-muted p-3">
        <Link2 size={16} className="text-muted-foreground" aria-hidden />
        <code className="min-w-0 flex-1 truncate text-sm" aria-label="Your affiliate link">{link}</code>
        <Button size="sm" variant="secondary" onClick={copy}><Copy size={14} className="mr-1.5" />Copy link</Button>
      </div>
      <p className="mt-3 text-xs text-muted-foreground">On the Free plan, every email you send ends with a small “Sent with Xem” link that carries this code, so readers who sign up count as yours.</p>
      <p className="mt-1 text-xs text-muted-foreground">Works on any xem.email page: add <code>?via={s.code}</code> to a link to your favourite page. A sign-up within 60 days of a click counts, and the last link clicked wins.</p>
    </section>

    <div className="mb-6 grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <Metric label="Clicks" value={s.clicks} icon={<MousePointerClick size={18} />} />
      <Metric label="Sign-ups" value={s.signups} icon={<UserPlus size={18} />} />
      <Metric label="Paying customers" value={s.customers} icon={<Users size={18} />} />
      <Metric label="Ready to pay" value={money(s.totals.unlocked, s.currency)} icon={<BadgeDollarSign size={18} />}
        detail={`${money(s.totals.pending, s.currency)} pending · ${money(s.totals.paid, s.currency)} paid`} />
    </div>

    <section className={workspaceClassName("product-panel")}>
      <div className={workspaceClassName("panel-toolbar")}><h2>Commissions</h2></div>
      {!s.commissions.length ? (
        <Empty title="No commissions yet" description="When someone you referred pays for a plan, their payment shows up here. Customer names and emails stay private." />
      ) : (
        <Table>
          <TableHeader><TableRow><TableHead>Customer</TableHead><TableHead>Payment date</TableHead><TableHead>Commission</TableHead><TableHead>Status</TableHead></TableRow></TableHeader>
          <TableBody>
            {s.commissions.map((c, i) => (
              <TableRow key={i}>
                <TableCell>{c.customer}</TableCell>
                <TableCell>{format(new Date(c.paidAt), "MMM d, yyyy")}</TableCell>
                <TableCell>{money(c.amount, c.currency)}</TableCell>
                <TableCell>{statusLabel[c.status]}{c.status === "pending" && <span className="text-muted-foreground"> · unlocks {format(new Date(c.unlocksAt), "MMM d")}</span>}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
      <p className="mt-4 text-xs text-muted-foreground">Payouts go out at the start of each month by bank transfer once ready-to-pay commissions reach $100. We'll email you to arrange the first one. Refunded or disputed payments earn nothing.</p>
    </section>
  </>;
}
