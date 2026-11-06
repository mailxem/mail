"use client";
import { useState } from "react";
import Link from "next/link";
import { ArrowUpRight } from "lucide-react";
import { CostCalculator } from "./cost-calculator";
import { checked, competitors, sizes } from "@/lib/compare";
import { weeklyEmails, xemMonthlyCost } from "@/lib/xem-pricing";

const MONTHS = "Jan Feb Mar Apr May Jun Jul Aug Sep Oct Nov Dec".split(" ");
const count = (n: number) => n.toLocaleString("en-US");
const money = (n: number) => `$${Math.round(n).toLocaleString("en-US")}`;

/** Competitors with a published price for list size i, cheapest bill first. */
const pricedAt = (i: number) =>
  competitors
    .filter((c) => typeof c.prices[i] === "number")
    .map((c) => ({ ...c, price: c.prices[i] as number }))
    .sort((a, b) => b.price - a.price);

function Invoices({
  name,
  price,
  muted,
}: {
  name: string;
  price: number;
  muted?: boolean;
}) {
  return (
    <div>
      <div className="flex items-baseline justify-between gap-4">
        <p className="text-sm font-medium">{name}</p>
        <p className="font-editorial text-3xl tracking-tight md:text-4xl">
          {money(price * 12)}
          <span className="ml-1 font-sans text-xs text-muted">a year</span>
        </p>
      </div>
      <ol className="mt-3 grid grid-cols-6 gap-1.5 md:grid-cols-12">
        {MONTHS.map((m) => (
          <li
            key={m}
            className={`rounded-md border px-1.5 py-2 text-center ${muted ? "border-ink/10 bg-ink/[.04] text-muted" : "border-iris/30 bg-[#ebe7f2]"}`}
          >
            <span className="block text-[9px] uppercase tracking-wider">
              {m}
            </span>
            <span className="mt-1 block text-xs font-medium tabular-nums">
              {money(price)}
            </span>
          </li>
        ))}
      </ol>
    </div>
  );
}

/** The largest saving at 10,000 subscribers, as a whole percent, e.g. 86. */
export const biggestSaving = () => {
  const xem = xemMonthlyCost(weeklyEmails(sizes[0])).cost;
  return Math.floor(100 * (1 - xem / pricedAt(0)[0].price));
};

/** A year of invoices on another tool next to Xem, then the calculator. */
export function PriceSection() {
  const [size, setSize] = useState(0);
  const options = pricedAt(size);
  const [pick, setPick] = useState("beehiiv");
  const them = options.find((c) => c.id === pick) ?? options[0];
  const xem = xemMonthlyCost(weeklyEmails(sizes[size])).cost;
  return (
    <section
      id="price"
      aria-labelledby="price-title"
      className="scroll-mt-20 px-5 py-20 md:px-10 md:py-28"
    >
      <div className="mx-auto max-w-[1120px]">
        <p className="text-[11px] font-medium uppercase tracking-[0.17em] text-iris">
          What it costs
        </p>
        <h2
          id="price-title"
          className="mt-5 max-w-3xl font-editorial text-[clamp(2.6rem,5vw,4.6rem)] leading-[1.02] tracking-[-.04em]"
        >
          Most tools charge for your list.
          <br />
          <em>Xem charges for your sends.</em>
        </h2>
        <p className="mt-6 max-w-2xl leading-relaxed text-muted">
          Here is a year of a weekly newsletter on another tool, invoice by
          invoice, next to the same year on Xem.
        </p>

        <div className="mt-10 rounded-2xl border border-ink/15 bg-cream p-6 md:p-8">
          <div className="flex flex-wrap gap-x-8 gap-y-4">
            <fieldset>
              <legend className="text-xs uppercase tracking-wider text-muted">
                Subscribers
              </legend>
              <div className="mt-2 flex flex-wrap gap-2">
                {sizes.map((n, i) => (
                  <button
                    key={n}
                    type="button"
                    aria-pressed={size === i}
                    onClick={() => setSize(i)}
                    className={`rounded-full border px-3 py-1.5 text-xs ${size === i ? "border-iris bg-iris text-white" : "border-ink/20 hover:border-iris"}`}
                  >
                    {count(n)}
                  </button>
                ))}
              </div>
            </fieldset>
            <fieldset>
              <legend className="text-xs uppercase tracking-wider text-muted">
                Compared with
              </legend>
              <div className="mt-2 flex flex-wrap gap-2">
                {options.map((c) => (
                  <button
                    key={c.id}
                    type="button"
                    aria-pressed={them.id === c.id}
                    onClick={() => setPick(c.id)}
                    className={`rounded-full border px-3 py-1.5 text-xs ${them.id === c.id ? "border-ink bg-ink text-cream" : "border-ink/20 hover:border-ink"}`}
                  >
                    {c.name}
                  </button>
                ))}
              </div>
            </fieldset>
          </div>

          <div className="mt-8 grid gap-6" aria-live="polite">
            <Invoices name={`${them.name} ${them.plan}`} price={them.price} muted />
            <Invoices name="Xem" price={xem} />
          </div>

          <div className="mt-8 flex flex-wrap items-end justify-between gap-4 border-t border-ink/10 pt-6">
            <p className="font-editorial text-3xl tracking-tight md:text-4xl">
              You keep {money((them.price - xem) * 12)} a year.
            </p>
            <Link
              href={`/compare/xem-vs-${them.id}`}
              className="inline-flex items-center gap-2 border-b border-ink pb-1 text-sm font-medium"
            >
              Xem vs {them.name} <ArrowUpRight size={16} />
            </Link>
          </div>
          <p className="mt-4 text-xs leading-relaxed text-muted">
            Monthly billing, four sends a month to the whole list, each
            tool&apos;s cheapest plan for newsletters. Prices checked {checked};
            every comparison page lists its sources.
          </p>
        </div>

        <h3 className="mt-16 font-editorial text-3xl tracking-tight md:text-4xl">
          Work out your own bill.
        </h3>
        <div className="mt-6">
          <CostCalculator />
        </div>

        <div className="mt-12">
          <p className="text-sm font-medium">Switching from another tool?</p>
          <ul className="mt-4 flex flex-wrap gap-2">
            {competitors.map((c) => (
              <li key={c.id}>
                <Link
                  href={`/migrate/${c.id}`}
                  className="inline-flex rounded-full border border-ink/20 px-4 py-2 text-sm hover:border-iris hover:text-iris"
                >
                  From {c.name}
                </Link>
              </li>
            ))}
          </ul>
          <p className="mt-4 text-xs text-muted">
            Unsubscribed contacts stay unsubscribed when you import them.
          </p>
        </div>
      </div>
    </section>
  );
}
