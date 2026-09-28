"use client";
import { useEffect, useRef, useState } from "react";
import dynamic from "next/dynamic";
const Showcase = dynamic(() => import("./analytics-showcase"), {
  ssr: false,
  loading: () => (
    <div className="min-h-[1300px] rounded-[30px] border border-ink/10 bg-[#f7f6fa] p-6">
      <div className="h-12 w-1/3 rounded-xl bg-[#e9e3f0] motion-safe:animate-pulse" />
      <div className="mt-7 grid grid-cols-2 gap-4 md:grid-cols-4">
        {[0, 1, 2, 3].map((n) => (
          <div key={n} className="h-28 rounded-xl bg-white" />
        ))}
      </div>
      <p role="status" className="mt-14 text-center text-sm text-muted">
        Bringing the numbers into focus…
      </p>
    </div>
  ),
});
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
              A clearer picture. A more thoughtful next move.
            </p>
            <h2 className="mt-5 font-editorial text-[clamp(3rem,5.5vw,5rem)] leading-[1.02] tracking-[-.045em]">
              Good emails connect.
              <br />
              <em>The numbers tell the story.</em>
            </h2>
          </div>
          <p className="max-w-[290px] text-sm leading-relaxed text-muted">
            See what’s resonating, who’s engaging, and where to go next. Take
            the controls below — there’s a whole story in these numbers.
          </p>
        </div>
        {ready ? (
          <Showcase />
        ) : (
          <div className="min-h-[1300px] rounded-[30px] border border-ink/10 bg-[#f7f6fa] p-7">
            <p className="text-sm text-muted">
              Explore campaigns, audience growth, devices, and engagement.
            </p>
            <noscript>
              <p className="mt-4">
                Enable JavaScript to interact with this synthetic analytics
                demo. The Xem analytics workspace includes campaign comparisons,
                date filters, audience cohorts, and delivery reports.
              </p>
            </noscript>
          </div>
        )}
      </div>
    </section>
  );
}
