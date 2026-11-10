"use client";
import { useEffect, useState } from "react";
import { apiUrl } from "@/lib/site";

type Stats = {
  delivered: number;
  delivered30d: number;
  deliveryRate: number;
  complaintRate: number;
};

// Real platform-wide totals from the API, refreshed hourly. Nothing here is
// typed in: with no data, or too little to mean anything, nothing shows.
function useStats() {
  const [stats, setStats] = useState<Stats | null>(null);
  useEffect(() => {
    const ctrl = new AbortController();
    fetch(`${apiUrl}/public/stats`, { signal: ctrl.signal })
      .then((r) => (r.ok ? r.json() : null))
      .then((s: Stats | null) => s && setStats(s))
      .catch(() => {});
    return () => ctrl.abort();
  }, []);
  return stats;
}

const count = (n: number) => n.toLocaleString("en-US");

/** "Over 5,000 emails", live when the API answers; `fallback` is the last known total. */
export function DeliveredCount({ fallback }: { fallback: number }) {
  const stats = useStats();
  const n = Math.max(fallback, stats?.delivered ?? 0);
  const step = n >= 100_000 ? 10_000 : 1_000;
  return <>{count(Math.floor(n / step) * step)} emails</>;
}

export function LiveNumbers() {
  const stats = useStats();
  if (!stats || stats.delivered30d < 500) return null;
  const items = [
    ["Delivery rate", `${(stats.deliveryRate * 100).toFixed(1)}%`],
    ["Spam complaints", `${(stats.complaintRate * 100).toFixed(3)}%`],
    ["Emails delivered", count(stats.delivered30d)],
  ];
  return (
    <section aria-labelledby="numbers-title" className="px-5 pb-6 md:px-10">
      <div className="mx-auto max-w-[860px] text-center">
        <h2 id="numbers-title" className="sr-only">
          Xem’s sending numbers
        </h2>
        <dl className="grid grid-cols-3 gap-4">
          {items.map(([label, value]) => (
            <div key={label} className="flex flex-col-reverse">
              <dt className="mt-2 text-[13px] text-muted">{label}</dt>
              <dd className="font-editorial text-[clamp(1.8rem,4.5vw,2.8rem)] leading-none tracking-[-.03em]">
                {value}
              </dd>
            </div>
          ))}
        </dl>
        <p className="mt-4 text-xs text-muted">
          The last 30 days, across every Xem workspace. Updated hourly.
        </p>
      </div>
    </section>
  );
}
