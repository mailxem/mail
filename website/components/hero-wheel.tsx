import { getCompetitor, sizes } from "@/lib/compare";
import { weeklyEmails, xemMonthlyCost } from "@/lib/xem-pricing";

// Savings for a weekly newsletter to the smallest compare size, from the
// same data as /compare, so the headline never drifts from the tables.
// Logos are each company's own icon, from its website.
const rivals = [
  { id: "mailchimp", logo: "/logos/mailchimp.svg" },
  { id: "beehiiv", logo: "/logos/beehiiv.png" },
  { id: "mailcheap", logo: "/logos/mailcheap.svg" },
  { id: "plunk", logo: "/logos/plunk.svg" },
  { id: "sequenzy", logo: "/logos/sequenzy.png" },
  { id: "ghost", logo: "/logos/ghost.png" },
];
const xem = xemMonthlyCost(weeklyEmails(sizes[0])).cost;
type Line = { text: string; name?: string; logo?: string };
const lines: Line[] = [
  { text: "at the price of email." },
  ...rivals.flatMap(({ id, logo }) => {
    const c = getCompetitor(id);
    const theirs = c?.prices[0];
    if (!c || typeof theirs !== "number") return [];
    return [
      {
        text: `for ${Math.floor(100 * (1 - xem / theirs))}% less than`,
        name: c.name,
        logo,
      },
    ];
  }),
];

// Sized in cqw of the heading: the longest line is about 10em wide, so 9.6cqw
// always fits. Each line holds, then rolls up like a wheel; the first line repeats at the
// end so the loop restarts without a jump.
const hold = 2.6;
const roll = 0.6;
const step = hold + roll;
const total = step * lines.length;
const keyframes = lines
  .map((_, i) => {
    const start = ((i * step) / total) * 100;
    const end = ((i * step + hold) / total) * 100;
    return `${start.toFixed(3)}%,${end.toFixed(3)}%{transform:translateY(${-i * 1.1}em)}`;
  })
  .concat(`100%{transform:translateY(${-lines.length * 1.1}em)}`)
  .join("");

export function HeroWheel() {
  return (
    <>
      <style>{`@keyframes hero-wheel{${keyframes}}`}</style>
      <span className="sr-only">{lines[0].text}</span>
      <span
        aria-hidden="true"
        className="relative block h-[1.1em] overflow-hidden text-[min(.86em,9.6cqw)] [mask-image:linear-gradient(transparent,#000_22%,#000_78%,transparent)]"
      >
        <span
          className="block animate-[hero-wheel_var(--wheel)_cubic-bezier(.65,0,.25,1.15)_infinite] motion-reduce:animate-none"
          style={{ "--wheel": `${total}s` } as React.CSSProperties}
        >
          {[...lines, lines[0]].map(({ text, name, logo }, i) => (
            <em
              key={i}
              className="block h-[1.1em] whitespace-nowrap px-[.15em] font-normal leading-[1.1em]"
            >
              {text}
              {name && (
                <>
                  {" "}
                  {/* eslint-disable-next-line @next/next/no-img-element */}
                  <img
                    src={logo}
                    alt=""
                    className="mr-[.14em] inline-block h-[.72em] w-auto -translate-y-[.04em] rounded-[.14em] align-baseline"
                  />
                  {name}.
                </>
              )}
            </em>
          ))}
        </span>
      </span>
    </>
  );
}
