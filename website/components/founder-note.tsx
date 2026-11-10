import { githubUrl } from "@/lib/site";

// In the founder's own voice. Keep every line true and specific.
export function FounderNote() {
  return (
    <section
      aria-labelledby="founder-title"
      className="px-5 py-20 md:px-10 md:py-28"
    >
      <div data-reveal className="mx-auto max-w-[680px]">
        <h2
          id="founder-title"
          className="font-editorial text-[clamp(2.6rem,5vw,4rem)] leading-[1.02] tracking-[-.04em]"
        >
          Why I built Xem
        </h2>
        <div className="mt-8 space-y-5 text-[17px] leading-[1.75] text-ink/85">
          <p>
            Hi, I’m{" "}
            <span className="inline-flex translate-y-[3px] items-center justify-center rounded-full bg-lemon px-2 text-sm font-semibold text-ink">
              H
            </span>{" "}
            <strong className="font-semibold text-ink">Harsh</strong>, and I
            build Xem.
          </p>
          <p>
            Most email tools charge for the size of your list, so you pay every
            month for people who stopped opening long ago. Sending an email
            costs a fraction of a cent. I wanted a tool that charges for what it
            actually does.
          </p>
          <p>So Xem is built around three things:</p>
          <ol className="list-decimal space-y-3 pl-6 marker:text-ink/50">
            <li>
              <strong className="font-semibold text-ink">
                Pay for sends, not contacts.
              </strong>{" "}
              A weekly newsletter to 10,000 readers is $19 a month.
            </li>
            <li>
              <strong className="font-semibold text-ink">Open source.</strong>{" "}
              Every line is on{" "}
              <a
                href={githubUrl}
                className="font-medium underline decoration-ink/30 underline-offset-4 hover:decoration-ink"
              >
                GitHub
              </a>{" "}
              under GPL-3.0, and you can run it on your own server.
            </li>
            <li>
              <strong className="font-semibold text-ink">
                Built for AI agents.
              </strong>{" "}
              43 MCP tools, so Claude, Codex or Cursor can work in your account
              too.
            </li>
          </ol>
          <p>
            And on paid plans, there’s no “Sent with Xem” line at the bottom of
            your emails.
          </p>
        </div>
      </div>
    </section>
  );
}
