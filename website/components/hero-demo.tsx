"use client";
import { useEffect, useRef, useState, type ReactNode } from "react";
import Image from "next/image";
import {
  BarChart3,
  CalendarClock,
  Check,
  Clock,
  Eye,
  FlaskConical,
  Globe,
  LayoutGrid,
  Mail,
  Monitor,
  Newspaper,
  PenLine,
  Search,
  Send,
  Smartphone,
  Sparkles,
  Tag,
  Users,
  Workflow,
  Zap,
} from "lucide-react";
import { PhonePreview } from "./phone-preview";

// A coded, animated walk through the Xem app with sample data. Each tab plays
// a short scene; the tabs advance on their own until someone picks one.
const TABS = [
  {
    id: "write",
    label: "Write",
    icon: PenLine,
    path: "campaigns/new",
    nav: "Campaigns",
  },
  {
    id: "schedule",
    label: "Schedule",
    icon: CalendarClock,
    path: "campaigns/new",
    nav: "Campaigns",
  },
  {
    id: "automate",
    label: "Automate",
    icon: Workflow,
    path: "automations/welcome",
    nav: "Automations",
  },
  {
    id: "stats",
    label: "Stats",
    icon: BarChart3,
    path: "analytics",
    nav: "Analytics",
  },
  {
    id: "website",
    label: "Website",
    icon: Globe,
    path: "newsletters/sunday-edit/site",
    nav: "Newsletters",
  },
] as const;
const SCENE_MS = 7000;

function useReducedMotion() {
  const [reduced, setReduced] = useState(false);
  useEffect(() => {
    const m = window.matchMedia("(prefers-reduced-motion: reduce)");
    setReduced(m.matches);
    const on = () => setReduced(m.matches);
    m.addEventListener("change", on);
    return () => m.removeEventListener("change", on);
  }, []);
  return reduced;
}

// Milliseconds since the scene started, ticking while it plays.
function useSceneClock(key: number, still: boolean) {
  const [t, setT] = useState(still ? 1e9 : 0);
  useEffect(() => {
    if (still) {
      setT(1e9);
      return;
    }
    setT(0);
    const start = performance.now();
    const id = setInterval(() => setT(performance.now() - start), 60);
    return () => clearInterval(id);
  }, [key, still]);
  return t;
}

const typed = (text: string, t: number, from: number, perChar = 45) =>
  text.slice(0, Math.max(0, Math.floor((t - from) / perChar)));

export function HeroDemo() {
  const reduced = useReducedMotion();
  const [tab, setTab] = useState(0);
  const [auto, setAuto] = useState(true);
  const [phone, setPhone] = useState(false);
  const [run, setRun] = useState(0);
  const box = useRef<HTMLDivElement>(null);
  const [visible, setVisible] = useState(true);

  // Only play while the demo is on screen.
  useEffect(() => {
    const el = box.current;
    if (!el) return;
    const io = new IntersectionObserver(([e]) => setVisible(e.isIntersecting), {
      threshold: 0.2,
    });
    io.observe(el);
    return () => io.disconnect();
  }, []);

  useEffect(() => {
    if (!auto || reduced || !visible) return;
    const id = setTimeout(() => {
      setTab((n) => (n + 1) % TABS.length);
      setRun((r) => r + 1);
      setPhone(false);
    }, SCENE_MS);
    return () => clearTimeout(id);
  }, [auto, reduced, visible, run]);

  const pick = (n: number) => {
    setAuto(false);
    setTab(n);
    setRun((r) => r + 1);
    if (n !== 0) setPhone(false);
  };
  const t = useSceneClock(run, reduced);
  const current = TABS[tab];

  return (
    <div ref={box}>
      <div
        role="tablist"
        aria-label="Explore Xem features"
        className="mx-auto mb-6 flex w-fit max-w-full overflow-x-auto rounded-full border border-cream/20 bg-cream/5 p-1"
      >
        {TABS.map((x, i) => {
          const Icon = x.icon;
          return (
            <button
              key={x.id}
              role="tab"
              id={`product-tab-${i}`}
              aria-controls="product-panel"
              aria-selected={tab === i}
              tabIndex={tab === i ? 0 : -1}
              onClick={() => pick(i)}
              onKeyDown={(e) => {
                if (e.key !== "ArrowRight" && e.key !== "ArrowLeft") return;
                e.preventDefault();
                const next =
                  (i + (e.key === "ArrowRight" ? 1 : TABS.length - 1)) %
                  TABS.length;
                pick(next);
                document.getElementById(`product-tab-${next}`)?.focus();
              }}
              className={`relative flex min-h-10 shrink-0 items-center gap-2 overflow-hidden rounded-full px-3.5 text-xs transition-colors md:px-5 md:text-sm ${tab === i ? "bg-cream text-forest" : "text-cream/85 hover:bg-cream/10"}`}
            >
              <Icon size={15} aria-hidden="true" />
              {x.label}
              {tab === i && auto && !reduced && (
                <span
                  key={run}
                  aria-hidden="true"
                  className="absolute bottom-0 left-0 h-[2px] bg-forest/40"
                  style={{
                    animation: `demo-progress ${SCENE_MS}ms linear forwards`,
                  }}
                />
              )}
            </button>
          );
        })}
      </div>
      <style>{`@keyframes demo-progress{from{width:0}to{width:100%}}@keyframes demo-in{from{opacity:0;transform:translateY(8px)}to{opacity:1;transform:none}}`}</style>

      <div
        id="product-panel"
        role="tabpanel"
        aria-labelledby={`product-tab-${tab}`}
        className="overflow-hidden rounded-2xl border border-cream/20 bg-[#f6f6f3] text-ink shadow-[0_30px_80px_-30px_rgba(0,0,0,.6)]"
      >
        <div className="flex h-12 items-center gap-3 border-b border-black/5 bg-[#ececE6] px-4">
          <span className="flex gap-1.5" aria-hidden="true">
            <i className="h-2.5 w-2.5 rounded-full bg-black/15" />
            <i className="h-2.5 w-2.5 rounded-full bg-black/15" />
            <i className="h-2.5 w-2.5 rounded-full bg-black/15" />
          </span>
          <span className="mx-auto truncate rounded-md bg-white/80 px-3 py-1 font-mono text-[11px] text-ink/70">
            app.xem.email/{current.path}
          </span>
          {tab === 0 ? (
            <span
              role="group"
              aria-label="Preview size"
              className="flex shrink-0 rounded-full bg-black/5 p-0.5 text-[11px]"
            >
              {(
                [
                  [false, "Desktop", Monitor],
                  [true, "Mobile", Smartphone],
                ] as const
              ).map(([value, label, Icon]) => (
                <button
                  key={label}
                  type="button"
                  aria-label={label}
                  aria-pressed={phone === value}
                  onClick={() => {
                    setAuto(false);
                    setPhone(value);
                  }}
                  className={`flex min-h-7 items-center gap-1.5 rounded-full px-2.5 ${phone === value ? "bg-white shadow-sm" : "text-ink hover:bg-white/60"}`}
                >
                  <Icon size={12} aria-hidden="true" />
                  <span className="hidden sm:inline">{label}</span>
                </button>
              ))}
            </span>
          ) : (
            <span
              className="w-[74px] shrink-0 sm:w-[150px]"
              aria-hidden="true"
            />
          )}
        </div>

        <div className="grid h-[460px] md:h-[520px] md:grid-cols-[200px_1fr]">
          <Sidebar active={current.nav} />
          <div className="relative min-w-0 overflow-hidden bg-white">
            <div
              key={`${tab}-${run}`}
              className="h-full"
              style={{
                animation: reduced ? undefined : "demo-in .45s ease-out",
              }}
            >
              {current.id === "write" && <WriteScene t={t} phone={phone} />}
              {current.id === "schedule" && <ScheduleScene t={t} />}
              {current.id === "automate" && <AutomateScene t={t} />}
              {current.id === "stats" && <StatsScene t={t} />}
              {current.id === "website" && <WebsiteScene t={t} />}
            </div>
          </div>
        </div>
      </div>
      <p className="mt-3 text-center text-[11px] text-cream/70">
        A demo of the Xem app with sample data.
      </p>
    </div>
  );
}

function Sidebar({ active }: { active: string }) {
  const groups: [string, [string, typeof Mail][]][] = [
    [
      "Workspace",
      [
        ["Overview", LayoutGrid],
        ["Ask Xem", Sparkles],
        ["Inbox", Mail],
      ],
    ],
    [
      "Create",
      [
        ["Campaigns", Send],
        ["Newsletters", Newspaper],
        ["Automations", Workflow],
      ],
    ],
    [
      "Audience",
      [
        ["Contact lists", Users],
        ["Analytics", BarChart3],
      ],
    ],
  ];
  return (
    <aside className="hidden border-r border-black/5 bg-[#f6f6f3] px-3 py-4 text-[12.5px] md:block">
      <div className="mb-4 flex items-center gap-2 px-1">
        <Image
          src="/brand/xem-mark.png"
          alt=""
          width={26}
          height={26}
          className="rounded-md"
        />
        <span className="leading-tight">
          <span className="block font-semibold">Xem workspace</span>
          <span className="block text-[11px] text-ink/70">Alex Morgan</span>
        </span>
      </div>
      <div className="mb-4 flex items-center gap-2 rounded-lg bg-black/[.04] px-2.5 py-1.5 text-ink/70">
        <Search size={13} /> Search
      </div>
      {groups.map(([title, items]) => (
        <div key={title} className="mb-3">
          <p className="mb-1 px-2 text-[11px] font-medium text-ink/70">
            {title}
          </p>
          {items.map(([label, Icon]) => (
            <p
              key={label}
              className={`flex items-center gap-2 rounded-md px-2 py-1.5 ${label === active ? "bg-white font-medium shadow-sm" : "text-ink/75"}`}
            >
              <Icon size={14} strokeWidth={1.6} />
              {label}
            </p>
          ))}
        </div>
      ))}
    </aside>
  );
}

function TopBar({
  crumbs,
  children,
}: {
  crumbs: string[];
  children?: ReactNode;
}) {
  return (
    <div className="flex items-center justify-between gap-3 border-b border-black/5 px-4 py-3 text-[13px] md:px-6">
      <p className="truncate text-ink/70">
        {crumbs.map((c, i) => (
          <span key={c}>
            {i > 0 && <span className="mx-1.5 text-ink/70">/</span>}
            <span
              className={i === crumbs.length - 1 ? "font-medium text-ink" : ""}
            >
              {c}
            </span>
          </span>
        ))}
      </p>
      <div className="flex shrink-0 items-center gap-2">{children}</div>
    </div>
  );
}

const Btn = ({ children, dark }: { children: ReactNode; dark?: boolean }) => (
  <span
    className={`inline-flex items-center gap-1.5 rounded-lg border px-2.5 py-1.5 text-[12px] ${dark ? "border-ink bg-ink text-cream" : "border-black/10 bg-white"}`}
  >
    {children}
  </span>
);

function WriteScene({ t, phone }: { t: number; phone: boolean }) {
  const subject = "Good things, delivered weekly";
  const s = typed(subject, t, 400);
  const pre = typed(
    "Three slow reads for your Sunday.",
    t,
    400 + subject.length * 45 + 300,
    30,
  );
  return (
    <div className="relative h-full">
      <TopBar crumbs={["Campaigns", "The Sunday Edit #42"]}>
        <span className="hidden rounded-full bg-black/5 px-2 py-0.5 text-[11px] text-ink/70 sm:inline">
          ● Draft
        </span>
        <Btn>
          <FlaskConical size={13} /> Test
        </Btn>
        <Btn>
          <Eye size={13} /> Preview
        </Btn>
        <Btn dark>
          <Send size={13} /> Send
        </Btn>
      </TopBar>
      <div
        className={`h-full p-4 transition md:p-6 ${phone ? "blur-[1.5px] brightness-[.6]" : ""}`}
      >
        <div className="rounded-xl border border-black/10 text-[13px]">
          <p className="flex gap-6 border-b border-black/5 px-4 py-3">
            <span className="w-16 shrink-0 text-ink/70">Subject</span>
            <span className="font-medium">
              {s}
              {s.length < subject.length && (
                <i className="ml-px inline-block h-4 w-px animate-pulse bg-ink align-middle" />
              )}
            </span>
          </p>
          <p className="flex gap-6 border-b border-black/5 px-4 py-3">
            <span className="w-16 shrink-0 text-ink/70">Preview</span>
            <span className="text-ink/75">{pre}</span>
          </p>
          <div className="flex justify-center bg-[#f3f1ec] px-4 pt-5">
            <Image
              src="/images/templates/reading-good-for-you.webp"
              alt="The “Reading’s Good for You” template in the Xem editor"
              width={640}
              height={820}
              className="w-[min(100%,340px)] rounded-t-md shadow-sm"
            />
          </div>
        </div>
      </div>
      {phone && (
        <div className="absolute inset-x-0 top-14 z-10">
          <PhonePreview />
        </div>
      )}
    </div>
  );
}

function ScheduleScene({ t }: { t: number }) {
  const days = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
  const times = ["8:00 AM", "9:00 AM", "12:00 PM", "6:00 PM"];
  const day = t > 1800 ? 6 : -1;
  const time = t > 2800 ? 1 : -1;
  const done = t > 4200;
  return (
    <div className="h-full">
      <TopBar crumbs={["Campaigns", "The Sunday Edit #42"]}>
        <Btn>
          <CalendarClock size={13} /> Schedule
        </Btn>
        <Btn dark>
          <Send size={13} /> Send
        </Btn>
      </TopBar>
      <div className="flex h-full items-start justify-center bg-[#f6f6f3] p-4 md:p-8">
        <div className="w-full max-w-[520px] rounded-2xl border border-black/10 bg-white p-5 shadow-sm md:p-6">
          <p className="text-base font-semibold">Schedule</p>
          <p className="mt-1 text-[13px] text-ink/70">
            Goes out to 2,184 contacts. You can keep editing until then.
          </p>
          <div className="mt-4 flex items-center justify-between gap-4 rounded-xl bg-[#f4f1ff] p-4">
            <span>
              <span className="flex items-center gap-1.5 text-[12px] text-iris">
                <Sparkles size={13} /> Best time to send
              </span>
              <span className="mt-1 block text-[13.5px] font-medium">
                Your readers open most on Sunday around 9 AM
              </span>
            </span>
            <span
              className="grid shrink-0 grid-cols-7 gap-[3px]"
              aria-hidden="true"
            >
              {Array.from({ length: 35 }, (_, i) => (
                <i
                  key={i}
                  className="block h-[6px] w-[6px] rounded-full"
                  style={{
                    background:
                      i % 7 === 6 && i > 6 && i < 21
                        ? "#5b3fd4"
                        : `rgba(91,63,212,${0.12 + ((i * 37) % 5) / 14})`,
                  }}
                />
              ))}
            </span>
          </div>
          <p className="mt-5 text-[12px] text-ink/70">Day</p>
          <div className="mt-2 grid grid-cols-7 gap-1 text-[12px]">
            {days.map((d, i) => (
              <span
                key={d}
                className={`rounded-lg py-2 text-center transition ${i === day ? "bg-ink text-cream" : "text-ink/70"}`}
              >
                {d}
              </span>
            ))}
          </div>
          <p className="mt-4 text-[12px] text-ink/70">Time</p>
          <div className="mt-2 flex flex-wrap gap-2 text-[12px]">
            {times.map((x, i) => (
              <span
                key={x}
                className={`rounded-full border px-3 py-1.5 transition ${i === time ? "border-ink bg-ink text-cream" : "border-black/10"}`}
              >
                {x}
              </span>
            ))}
          </div>
          <div className="mt-5 flex items-center justify-between">
            <span className="text-[13px] font-medium">
              {day >= 0 && time >= 0
                ? "Sun, 9:00 AM · your readers’ time zones"
                : "Pick a day and time"}
            </span>
            <span
              className={`inline-flex items-center gap-1.5 rounded-lg px-3 py-2 text-[12px] transition ${done ? "bg-emerald-600 text-white" : "bg-ink text-cream"}`}
            >
              {done ? <Check size={13} /> : <CalendarClock size={13} />}{" "}
              {done ? "Scheduled" : "Schedule"}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}

function AutomateScene({ t }: { t: number }) {
  const steps = [
    {
      icon: Zap,
      kicker: "When this happens",
      title: "Someone signs up",
      sub: "Form: Sunday Edit signup",
    },
    {
      icon: Mail,
      kicker: "Then do this",
      title: "Send “Welcome to the Sunday Edit”",
      sub: "Right away",
    },
    {
      icon: Clock,
      kicker: "Wait",
      title: "3 days",
      sub: "Or until they open it",
    },
    {
      icon: Tag,
      kicker: "If they opened",
      title: "Tag as “engaged”",
      sub: "Otherwise, send a gentle nudge",
    },
  ];
  const shown = Math.min(steps.length, 1 + Math.floor(t / 1100));
  return (
    <div className="h-full">
      <TopBar crumbs={["Automations", "A warm welcome"]}>
        <span className="hidden rounded-full bg-emerald-50 px-2 py-0.5 text-[11px] text-emerald-700 sm:inline">
          ● Live
        </span>
        <Btn dark>Publish</Btn>
      </TopBar>
      <div className="flex h-full justify-center overflow-hidden bg-[#f6f6f3] bg-[radial-gradient(#d8d8d0_1px,transparent_1px)] p-5 [background-size:18px_18px]">
        <ol className="w-full max-w-[400px]">
          {steps.slice(0, shown).map((s, i) => {
            const Icon = s.icon;
            return (
              <li key={s.title} style={{ animation: "demo-in .5s ease-out" }}>
                {i > 0 && (
                  <span
                    className="mx-auto block h-6 w-px bg-ink/25"
                    aria-hidden="true"
                  />
                )}
                <span
                  className={`flex items-center gap-3 rounded-xl border bg-white p-3.5 shadow-sm ${i === 0 ? "border-ink/40" : "border-black/10"}`}
                >
                  <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-black/[.04]">
                    <Icon size={16} />
                  </span>
                  <span className="min-w-0 leading-tight">
                    <span className="block text-[10px] uppercase tracking-[.12em] text-ink/70">
                      {s.kicker}
                    </span>
                    <span className="mt-0.5 block truncate text-[13.5px] font-semibold">
                      {s.title}
                    </span>
                    <span className="block text-[12px] text-ink/70">
                      {s.sub}
                    </span>
                  </span>
                </span>
              </li>
            );
          })}
        </ol>
      </div>
    </div>
  );
}

function StatsScene({ t }: { t: number }) {
  const k = Math.min(1, t / 1600);
  const ease = 1 - Math.pow(1 - k, 3);
  const n = (v: number) => Math.round(v * ease).toLocaleString();
  const pts = [8, 12, 11, 18, 16, 24, 23, 31, 29, 38, 44, 41, 52, 58];
  const W = 520;
  const H = 120;
  const path = pts
    .map(
      (p, i) =>
        `${i === 0 ? "M" : "L"}${((i / (pts.length - 1)) * W).toFixed(1)},${(H - (p / 60) * H).toFixed(1)}`,
    )
    .join(" ");
  const links = [
    ["Read the essay", 41],
    ["Join the launch", 84],
    ["Field notes", 12],
  ] as const;
  return (
    <div className="h-full">
      <TopBar crumbs={["Analytics", "Last 30 days"]} />
      <div className="space-y-4 p-4 md:p-6">
        <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
          {[
            ["Sent", n(12480)],
            ["Opened", `${(48.6 * ease).toFixed(1)}%`],
            ["Clicked", `${(6.9 * ease).toFixed(1)}%`],
            ["New subscribers", `+${n(1304)}`],
          ].map(([label, value]) => (
            <div key={label} className="rounded-xl border border-black/10 p-3">
              <p className="text-[11.5px] text-ink/70">{label}</p>
              <p className="mt-1 text-xl font-semibold tabular-nums">{value}</p>
            </div>
          ))}
        </div>
        <div className="rounded-xl border border-black/10 p-4">
          <p className="text-[12px] text-ink/70">Subscribers</p>
          <svg
            viewBox={`0 0 ${W} ${H}`}
            className="mt-2 h-28 w-full"
            preserveAspectRatio="none"
            aria-hidden="true"
          >
            <path
              d={`${path} L${W},${H} L0,${H} Z`}
              fill="#5b3fd4"
              opacity={0.08 * ease}
            />
            <path
              d={path}
              fill="none"
              stroke="#5b3fd4"
              strokeWidth="2"
              strokeDasharray="900"
              strokeDashoffset={900 * (1 - ease)}
              vectorEffect="non-scaling-stroke"
            />
          </svg>
        </div>
        <div className="hidden rounded-xl border border-black/10 p-4 sm:block">
          <p className="mb-2 text-[12px] text-ink/70">
            Top links · people who clicked
          </p>
          {links.map(([label, v]) => (
            <p
              key={label}
              className="relative mb-1.5 flex justify-between overflow-hidden rounded-md px-2 py-1 text-[12.5px]"
            >
              <span
                className="absolute inset-y-0 left-0 rounded-md bg-[#5b3fd4]/10"
                style={{ width: `${(v / 84) * 100 * ease}%` }}
              />
              <span className="relative">{label}</span>
              <span className="relative tabular-nums text-ink/70">
                {Math.round(v * ease)}
              </span>
            </p>
          ))}
        </div>
      </div>
    </div>
  );
}

function WebsiteScene({ t }: { t: number }) {
  const email = typed("sam@example.com", t, 900, 70);
  const done = t > 900 + 15 * 70 + 500;
  return (
    <div className="h-full">
      <TopBar crumbs={["Newsletters", "The Sunday Edit", "Website"]}>
        <span className="hidden font-mono text-[11px] text-ink/70 sm:inline">
          xem.email/sunday-edit
        </span>
      </TopBar>
      <div className="h-full bg-[#fbf7ef] px-5 py-6 md:px-10 md:py-8">
        <p className="font-editorial text-3xl tracking-tight md:text-4xl">
          The Sunday Edit
        </p>
        <p className="mt-1 text-[13px] text-ink/70">
          Good things, delivered weekly. 2,184 readers.
        </p>
        <div className="mt-4 flex max-w-[420px] items-center gap-2 rounded-full border border-black/15 bg-white p-1 pl-4 text-[13px]">
          <span className="min-w-0 flex-1 truncate text-ink/80">
            {email || <span className="text-ink/70">you@example.com</span>}
          </span>
          <span
            className={`shrink-0 rounded-full px-3.5 py-1.5 text-[12px] transition ${done ? "bg-emerald-600 text-white" : "bg-ink text-cream"}`}
          >
            {done ? "Check your inbox" : "Subscribe"}
          </span>
        </div>
        <div className="mt-6 grid gap-3 sm:grid-cols-3">
          {[
            ["reading-good-for-you", "Three slow reads"],
            ["summer-family-journal", "A summer journal"],
            ["most-wanted", "August’s most wanted"],
          ].map(([img, title]) => (
            <div
              key={img}
              className="overflow-hidden rounded-xl border border-black/10 bg-white"
            >
              <Image
                src={`/images/templates/${img}.webp`}
                alt=""
                width={640}
                height={820}
                className="h-24 w-full object-cover object-top"
              />
              <p className="px-3 py-2 text-[12.5px] font-medium">{title}</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
