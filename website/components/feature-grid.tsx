import type { ReactNode } from "react";
import Image from "next/image";
import { Check, Inbox, Sparkles, Tag } from "lucide-react";

// Plain cards, one claim each, with a small piece of the real UI (sample data).
const TONES = {
  lavender: "bg-lavender text-ink",
  forest: "bg-forest text-cream",
  ink: "bg-[#222d25] text-cream",
  lemon: "bg-lemon text-ink",
  sage: "bg-[#dce4cd] text-ink",
  peach: "bg-[#faeee6] text-ink",
  mist: "bg-[#ebe7f2] text-ink",
  sand: "bg-[#fbf7ef] text-ink",
} as const;

function Card({
  id,
  title,
  body,
  tone,
  wide,
  children,
}: {
  id?: string;
  title: string;
  body: string;
  tone: keyof typeof TONES;
  wide?: boolean;
  children: ReactNode;
}) {
  const dark = tone === "forest" || tone === "ink";
  return (
    <article
      id={id}
      className={`flex scroll-mt-28 flex-col gap-6 rounded-[28px] p-6 md:p-7 ${TONES[tone]} ${wide ? "md:col-span-2" : ""}`}
    >
      <div>
        <h3 className="font-editorial text-[1.9rem] leading-none tracking-[-.02em]">
          {title}
        </h3>
        <p
          className={`mt-3 text-[14.5px] leading-relaxed ${dark ? "text-cream/85" : "text-ink/80"}`}
        >
          {body}
        </p>
      </div>
      <div className="mt-auto text-[12.5px]">{children}</div>
    </article>
  );
}

const Pill = ({ ok, children }: { ok?: boolean; children: ReactNode }) => (
  <span
    className={`inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-[12px] ${ok ? "bg-cream/15 text-cream" : "bg-lemon text-ink"}`}
  >
    {ok ? (
      <Check size={12} aria-hidden="true" />
    ) : (
      <span className="h-1.5 w-1.5 rounded-full bg-current" />
    )}
    {children}
  </span>
);

export function FeatureGrid() {
  return (
    <section
      aria-labelledby="features-title"
      className="px-5 py-20 md:px-10 md:py-24"
    >
      <div className="mx-auto max-w-[1160px]">
        <h2
          id="features-title"
          className="max-w-[640px] font-editorial text-[clamp(2.6rem,5vw,4.2rem)] leading-[1.02] tracking-[-.04em]"
        >
          Everything you need to run your email
        </h2>
        <p className="mt-4 max-w-[560px] text-[16px] leading-relaxed text-ink/75">
          The parts of Xem beyond writing and sending.
        </p>

        <div className="mt-10 grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          <Card
            id="assistant"
            wide
            title="Ask Xem"
            tone="lavender"
            body="Ask which campaign did best last month, or have it draft next week’s newsletter. It changes nothing until you approve."
          >
            <div className="space-y-2.5">
              <p className="ml-auto w-fit max-w-[85%] rounded-2xl rounded-br-md bg-iris px-3.5 py-2 text-cream">
                Which campaign did best last month?
              </p>
              <p className="w-fit max-w-[90%] rounded-2xl rounded-bl-md bg-cream px-3.5 py-2 shadow-sm">
                <Sparkles
                  size={12}
                  className="mr-1.5 inline text-iris"
                  aria-hidden="true"
                />
                “The Sunday Edit #41”: 52% opened, 7.1% clicked. Want a draft of
                #43 in the same style?
              </p>
            </div>
          </Card>

          <Card
            id="sending"
            title="Set up in four steps"
            tone="forest"
            body="Verify your domain, add an API key or SMTP, send yourself a test. A checklist stays on your dashboard until you’re done."
          >
            <ol className="space-y-1.5">
              {[
                ["Verify your domain", true],
                ["Add an API key or SMTP", true],
                ["Send yourself a test", false],
                ["Send your first campaign", false],
              ].map(([label, done], i) => (
                <li key={label as string} className="flex items-center gap-2">
                  <span
                    className={`flex h-5 w-5 items-center justify-center rounded-full text-[10px] ${done ? "bg-lemon text-forest" : "border border-cream/40"}`}
                  >
                    {done ? <Check size={11} aria-hidden="true" /> : i + 1}
                  </span>
                  <span className={done ? "text-cream/70 line-through" : ""}>
                    {label as string}
                  </span>
                </li>
              ))}
            </ol>
          </Card>

          <Card
            title="Domain checks"
            tone="ink"
            body="Xem checks ownership, DKIM, the bounce domain and DMARC, and tells you which record is still missing."
          >
            <div className="flex flex-wrap gap-1.5">
              <Pill ok>Ownership</Pill>
              <Pill ok>DKIM</Pill>
              <Pill ok>Bounce domain</Pill>
              <Pill>DMARC</Pill>
            </div>
            <p className="mt-3 font-mono text-[11px] text-cream/80">
              _dmarc.yourdomain.com · TXT · v=DMARC1; p=none
            </p>
          </Card>

          <Card
            title="Signup forms"
            tone="lemon"
            body="Put a form on your site. New subscribers land in a list and can get a welcome email straight away."
          >
            <div className="flex items-center gap-2 rounded-full bg-cream p-1 pl-3.5 shadow-sm">
              <span className="flex-1 text-ink/70">you@example.com</span>
              <span className="rounded-full bg-ink px-3 py-1.5 text-cream">
                Subscribe
              </span>
            </div>
          </Card>

          <Card
            title="Contacts and CRM"
            tone="mist"
            body="Lists, tags and lifecycle stages, with notes on each contact."
          >
            <ul className="space-y-1.5">
              {[
                ["Priya Shah", "Customer", "engaged"],
                ["Tom Becker", "Lead", "webinar"],
                ["Lea Dubois", "Subscriber", "new"],
              ].map(([name, stage, tag]) => (
                <li
                  key={name}
                  className="flex items-center justify-between gap-2 rounded-xl bg-cream px-3 py-2"
                >
                  <span className="font-medium">{name}</span>
                  <span className="flex items-center gap-1.5 text-[11.5px] text-ink/70">
                    {stage}
                    <span className="inline-flex items-center gap-1 rounded-full bg-lavender px-2 py-0.5">
                      <Tag size={10} aria-hidden="true" />
                      {tag}
                    </span>
                  </span>
                </li>
              ))}
            </ul>
          </Card>

          <Card
            title="Shared inbox"
            tone="sage"
            body="Replies to your campaigns land in one inbox your team can read and answer, from your own mailbox."
          >
            <ul className="space-y-1.5">
              {[
                [
                  "Re: The Sunday Edit #42",
                  "Loved the piece on slow mornings!",
                ],
                ["Re: Your order has shipped", "Can I change the address?"],
              ].map(([subject, line]) => (
                <li
                  key={subject}
                  className="flex gap-2 rounded-xl bg-cream px-3 py-2"
                >
                  <Inbox
                    size={14}
                    className="mt-0.5 shrink-0 text-ink/60"
                    aria-hidden="true"
                  />
                  <span className="min-w-0">
                    <span className="block truncate font-medium">
                      {subject}
                    </span>
                    <span className="block truncate text-ink/70">{line}</span>
                  </span>
                </li>
              ))}
            </ul>
          </Card>

          <Card
            title="Import your list"
            tone="peach"
            body="Bring a CSV from Mailchimp, beehiiv, Substack or Kit. Unsubscribed contacts stay unsubscribed."
          >
            <div className="flex items-center gap-2">
              <Image
                src="/logos/mailchimp.svg"
                alt="Mailchimp"
                width={28}
                height={28}
                className="h-8 w-8 rounded-lg bg-cream p-1"
              />
              <Image
                src="/logos/beehiiv.png"
                alt="beehiiv"
                width={28}
                height={28}
                className="h-8 w-8 rounded-lg"
              />
              <span className="rounded-lg bg-cream px-2 py-1.5 text-[11.5px] font-medium">
                CSV
              </span>
              <span className="text-ink/60" aria-hidden="true">
                →
              </span>
              <Image
                src="/brand/xem-mark.png"
                alt="Xem"
                width={28}
                height={28}
                className="h-8 w-8 rounded-lg"
              />
              <span className="ml-auto text-ink/70">2,184 contacts</span>
            </div>
          </Card>

          <Card
            title="Click map"
            tone="sand"
            body="See each campaign as it went out, with every link coloured by how many people clicked it."
          >
            <p className="pt-3 leading-loose">
              This week:{" "}
              <span className="relative whitespace-nowrap rounded px-1 underline outline outline-2 outline-[#facc15]">
                Read the essay
                <span className="absolute -right-3 -top-3 rounded-full bg-[#facc15] px-1.5 text-[10px] font-semibold no-underline">
                  41
                </span>
              </span>{" "}
              or{" "}
              <span className="relative ml-2 whitespace-nowrap rounded bg-ink px-2 py-0.5 text-cream outline outline-2 outline-offset-2 outline-[#b91c1c]">
                Join the launch
                <span className="absolute -right-3 -top-3 rounded-full bg-[#b91c1c] px-1.5 text-[10px] font-semibold text-white">
                  84
                </span>
              </span>
            </p>
          </Card>
        </div>
        <p className="mt-4 text-[12px] text-ink/70">
          Previews use sample data.
        </p>
      </div>
    </section>
  );
}
