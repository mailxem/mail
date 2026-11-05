"use client";
import { useId, useState } from "react";
import { sesMonthlyCost, xemMonthlyCost } from "@/lib/xem-pricing";

const SIZES = [1000, 2500, 5000, 10000, 25000, 50000, 100000, 250000];
const SENDS = [
  { label: "Monthly", perMonth: 1 },
  { label: "Every two weeks", perMonth: 2 },
  { label: "Weekly", perMonth: 4 },
  { label: "Twice a week", perMonth: 8 },
];
const count = (n: number) => n.toLocaleString("en-US");
const money = (n: number) =>
  `$${n < 10 ? n.toFixed(2) : Math.round(n).toLocaleString("en-US")}`;

/** What Xem costs for a list size and sending rhythm, on Cloud and self-hosted. */
export function CostCalculator() {
  const id = useId();
  const [size, setSize] = useState(3);
  const [sends, setSends] = useState(2);
  const subscribers = SIZES[size];
  const emails = subscribers * SENDS[sends].perMonth;
  const cloud = xemMonthlyCost(emails);
  return (
    <div className="rounded-2xl border border-ink/15 bg-cream p-6 md:p-8">
      <div className="grid gap-6 md:grid-cols-2">
        <div>
          <label htmlFor={`${id}-size`} className="text-sm font-medium">
            Subscribers: <span className="text-iris">{count(subscribers)}</span>
          </label>
          <input
            id={`${id}-size`}
            type="range"
            min={0}
            max={SIZES.length - 1}
            value={size}
            onChange={(e) => setSize(Number(e.target.value))}
            className="mt-3 w-full accent-[#5b3fc4]"
            aria-valuetext={`${count(subscribers)} subscribers`}
          />
        </div>
        <fieldset>
          <legend className="text-sm font-medium">How often you send</legend>
          <div className="mt-3 flex flex-wrap gap-2">
            {SENDS.map((s, i) => (
              <button
                key={s.label}
                type="button"
                aria-pressed={sends === i}
                onClick={() => setSends(i)}
                className={`rounded-full border px-3 py-1.5 text-xs ${sends === i ? "border-iris bg-iris text-white" : "border-ink/20 hover:border-iris"}`}
              >
                {s.label}
              </button>
            ))}
          </div>
        </fieldset>
      </div>
      <dl
        className="mt-8 grid gap-4 border-t border-ink/10 pt-6 sm:grid-cols-3"
        aria-live="polite"
      >
        <div>
          <dt className="text-xs uppercase tracking-wider text-muted">
            Emails a month
          </dt>
          <dd className="mt-2 font-editorial text-4xl tracking-tight">
            {count(emails)}
          </dd>
        </div>
        <div>
          <dt className="text-xs uppercase tracking-wider text-muted">
            Xem Cloud
          </dt>
          <dd className="mt-2 font-editorial text-4xl tracking-tight">
            {money(cloud.cost)}
            <span className="ml-1 text-sm font-sans text-muted">
              / mo · {cloud.plan}
            </span>
          </dd>
        </div>
        <div>
          <dt className="text-xs uppercase tracking-wider text-muted">
            Self-hosted on SES
          </dt>
          <dd className="mt-2 font-editorial text-4xl tracking-tight">
            {money(sesMonthlyCost(emails))}
            <span className="ml-1 text-sm font-sans text-muted">
              / mo + server
            </span>
          </dd>
        </div>
      </dl>
    </div>
  );
}
