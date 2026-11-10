import type { ReactNode } from "react";
import Image from "next/image";
import { githubUrl } from "@/lib/site";
import { agents } from "./agent-strip";

// In the founder's own voice. Every claim here is checkable: prices come
// from the comparison pages, tool counts from the MCP server.
const GITHUB =
  "M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12";
const SES =
  "M11.9996 0C5.3833 0 0 5.3834 0 11.9998c0 2.5316.7813 4.9544 2.2599 7.0051l.6955-.5014C1.5827 16.5993.8571 14.3505.8571 11.9998.8571 5.856 5.856.8572 12.0004.8572c6.144 0 11.1425 4.999 11.1425 11.1426 0 2.3508-.7256 4.5995-2.0983 6.5037l.6955.5014C23.2187 16.9542 24 14.5314 24 11.9998 24 5.3834 18.6163 0 11.9996 0zM6 16.7142a.4285.4285 0 0 0-.4286.4285v1.7598c-.9643.2048-1.7143 1.0822-1.7143 2.0974 0 1.1615.9815 2.143 2.1429 2.143s2.1429-.9815 2.1429-2.143c0-1.0152-.75-1.8926-1.7143-2.0974v-1.3312h5.1428v2.1883c-.9643.2049-1.7143 1.0822-1.7143 2.0975C9.8571 23.0186 10.8386 24 12 24s2.1429-.9814 2.1429-2.1429c0-1.0153-.75-1.8926-1.7143-2.0975v-2.1883h5.1428v1.3312c-.9643.2048-1.7143 1.0822-1.7143 2.0974 0 1.1615.9815 2.143 2.1429 2.143s2.1429-.9815 2.1429-2.143c0-1.0152-.75-1.8926-1.7143-2.0974v-1.7598A.4285.4285 0 0 0 18 16.7142h-5.5714v-2.5715H18c.237 0 .4286-.192.4286-.4286V5.9997A.4285.4285 0 0 0 18 5.571H6a.4285.4285 0 0 0-.4286.4286v7.7144c0 .2366.1916.4286.4286.4286h5.5714v2.5715H6zm1.2857 4.2857c0 .697-.5889 1.2858-1.2857 1.2858s-1.2857-.5889-1.2857-1.2858c0-.6968.5889-1.2857 1.2857-1.2857S7.2857 20.3031 7.2857 21zm12 0c0 .697-.5889 1.2858-1.2857 1.2858s-1.2857-.5889-1.2857-1.2858c0-.6968.5889-1.2857 1.2857-1.2857s1.2857.5889 1.2857 1.2857zm-1.7143-8.248L14.259 9.7703l3.3124-2.8389v5.8205zm-.7298-6.3236-4.842 4.1499-4.8412-4.15h9.6832zm-10.413.5031L9.741 9.7707 6.4286 12.752V6.9314zm.6878 6.3541 3.2807-2.9525 1.3239 1.135a.4253.4253 0 0 0 .2786.1032.4253.4253 0 0 0 .2785-.1033l1.3243-1.1349 3.2812 2.9525H7.1164zM12 20.5714c.6968 0 1.2857.5888 1.2857 1.2857 0 .6969-.5889 1.2857-1.2857 1.2857s-1.2857-.5888-1.2857-1.2857c0-.6969.5889-1.2857 1.2857-1.2857z";

const Mark = ({
  path,
  bg,
  fg,
  label,
}: {
  path: string;
  bg: string;
  fg: string;
  label: string;
}) => (
  <span
    className="mx-0.5 inline-flex h-[1.2em] w-[1.2em] items-center justify-center rounded-[.3em] align-[-0.22em]"
    style={{ backgroundColor: bg, color: fg }}
  >
    <svg
      viewBox="0 0 24 24"
      className="h-[.85em] w-[.85em]"
      fill="currentColor"
      {...(label
        ? { role: "img", "aria-label": label }
        : { "aria-hidden": true })}
    >
      <path d={path} />
    </svg>
  </span>
);
const Logo = ({
  src,
  label,
  className = "",
}: {
  src: string;
  label: string;
  className?: string;
}) => (
  <Image
    src={src}
    alt={label}
    width={22}
    height={22}
    className={`mx-0.5 inline-block h-[1.2em] w-[1.2em] rounded-[.3em] object-contain align-[-0.22em] ${className}`}
  />
);
const B = ({ children }: { children: ReactNode }) => (
  <strong className="font-semibold text-ink">{children}</strong>
);

export function FounderNote() {
  return (
    <section
      aria-labelledby="founder-title"
      className="px-5 py-20 md:px-10 md:py-28"
    >
      <div data-reveal className="mx-auto max-w-[860px]">
        <h2
          id="founder-title"
          className="font-editorial text-[clamp(2.8rem,5.5vw,4.6rem)] leading-[1.02] tracking-[-.04em]"
        >
          Why I built Xem
        </h2>
        <div className="mt-10 space-y-6 text-[18px] leading-[1.8] text-ink/80 md:text-[19px]">
          <p>
            Hi, I’m{" "}
            <Image
              src="/founder/harsh-avatar.jpg"
              alt="Harsh"
              width={40}
              height={40}
              className="mx-0.5 inline-block h-[1.4em] w-[1.4em] rounded-full object-cover align-[-0.32em] ring-2 ring-lemon"
            />{" "}
            <B>Harsh</B>, and I build{" "}
            <Logo src="/brand/xem-mark.png" label="Xem" /> <B>Xem</B>.
          </p>
          <p>
            Sending an email is close to free.{" "}
            <Mark path={SES} bg="#232f3e" fg="#ff9900" label="Amazon SES" />{" "}
            <B>Amazon SES</B> charges about ten cents per thousand. Yet most
            email tools don’t charge you for sending. They charge for the size
            of your list, so every month you pay for people who stopped opening
            your emails a long time ago.
          </p>
          <p>
            <Logo src="/logos/mailchimp.svg" label="Mailchimp" />{" "}
            <B>Mailchimp</B> is the clearest case. A weekly newsletter to 10,000
            people costs about <B>$135 a month</B> on its Standard plan, and
            this year it cut its free plan to 250 contacts and 500 emails a
            month. <Logo src="/logos/beehiiv.png" label="beehiiv" />{" "}
            <B>beehiiv</B> wants $139 for the same list.
          </p>
          <p>
            So <Logo src="/brand/xem-mark.png" label="Xem" /> Xem charges for
            the one thing it actually does, which is send your email. That same
            newsletter costs <B>$19 a month</B>, and the first{" "}
            <B>3,000 emails every month are free</B>. Teams at{" "}
            <Logo src="/logos/hakopod.svg" label="" /> <B>Hakopod</B>,{" "}
            <Logo src="/logos/zunofy.png" label="" className="scale-[1.4]" />{" "}
            <B>Zunofy</B> and <Logo src="/logos/synehq.png" label="" />{" "}
            <B>SyneHQ</B> already send through it.
          </p>
          <p>I built it around three things:</p>
          <ol className="list-decimal space-y-4 pl-7 marker:text-ink/60">
            <li>
              <B>You pay for sends, not contacts.</B> Quiet subscribers cost
              nothing, and unsubscribed ones never count against you.
            </li>
            <li>
              <B>It’s open source.</B> Read every line on{" "}
              <a
                href={githubUrl}
                className="font-medium text-ink underline decoration-ink/30 underline-offset-4 hover:decoration-ink"
              >
                <Mark path={GITHUB} bg="#181717" fg="#ffffff" label="" />
                GitHub
              </a>
              , or run it on your own server and pay Amazon directly.
            </li>
            <li>
              <B>Your AI agent can use it too.</B> 43 MCP tools, so{" "}
              <span className="whitespace-nowrap">
                {agents.map((a) => (
                  <Mark
                    key={a.name}
                    path={a.path}
                    bg={a.bg}
                    fg={a.fg}
                    label={a.name}
                  />
                ))}
              </span>{" "}
              Claude, Codex or Cursor can import your list, draft a campaign or
              read last week’s numbers. Nothing goes out until you schedule it.
            </li>
          </ol>
          <p>
            Oh, and on paid plans there’s no “Sent with Xem” line at the bottom
            of your emails. The free plan carries a small one; that’s the trade
            for not paying.
          </p>
        </div>
      </div>
    </section>
  );
}
