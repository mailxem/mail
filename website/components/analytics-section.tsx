"use client";
import { useEffect, useRef, useState } from "react";
import dynamic from "next/dynamic";
const Showcase = dynamic(
  () => import("./dashboard-demo").then((m) => m.DashboardDemo),
  {
    ssr: false,
    loading: () => (
      <div className="min-h-[640px] rounded-[28px] border border-ink/10 bg-[#fcfcfa] p-7">
        <p role="status" className="text-sm text-muted">
          Loading the dashboard…
        </p>
      </div>
    ),
  },
);
export function AnalyticsSection() {
  const host = useRef<HTMLElement>(null);
  const [ready, setReady] = useState(false);
  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setReady(true);
          observer.disconnect();
        }
      },
      { rootMargin: "600px" },
    );
    if (host.current) observer.observe(host.current);
    return () => observer.disconnect();
  }, []);
  return (
    <section
      ref={host}
      id="analytics"
      className="scroll-mt-24 px-5 py-20 md:px-10 md:py-28"
    >
      <div className="mx-auto max-w-[1240px]">
        <div className="mb-12 flex flex-wrap items-end justify-between gap-7 md:px-4">
          <div>
            <p className="text-[11px] font-medium uppercase tracking-[.17em]">
              Analytics
            </p>
            <h2 className="mt-5 max-w-[640px] font-editorial text-[clamp(2.6rem,5vw,4.2rem)] leading-[1.02] tracking-[-.04em]">
              See how every send did
            </h2>
          </div>
          <p className="max-w-[290px] text-sm leading-relaxed text-muted">
            This is the Xem dashboard with sample data. Bot opens and
            link-scanner clicks are left out of every number. Switch the period
            or the tab.
          </p>
        </div>
        {ready ? (
          <Showcase />
        ) : (
          <div className="min-h-[640px] rounded-[28px] border border-ink/10 bg-[#fcfcfa] p-7">
            <p className="text-sm text-muted">
              Sent, opened, clicked and subscriber growth, on sample data.
            </p>
            <noscript>
              <p className="mt-4">
                Enable JavaScript to try this dashboard demo.
              </p>
            </noscript>
          </div>
        )}
      </div>
    </section>
  );
}
