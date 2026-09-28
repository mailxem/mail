"use client";
import { useEffect, useMemo, useRef, useState } from "react";
import { useChat } from "@ai-sdk/react";
import { createChat } from "@shadcn/helpers/ai-sdk";
import {
  ArrowUp,
  ArrowUpRight,
  Check,
  ChevronDown,
  Mail,
  Pause,
  Play,
  RotateCcw,
  Sparkles,
} from "lucide-react";
import { appLink } from "@/lib/site";

function ConversationDemo() {
  const [run, setRun] = useState(0);
  const [playing, setPlaying] = useState(false);
  const [started, setStarted] = useState(false);
  const section = useRef<HTMLDivElement>(null);
  const scroll = useRef<HTMLDivElement>(null);
  const story = useMemo(
    () =>
      createChat()
        .user("Help me turn our product update into a newsletter.")
        .assistant(({ writer }) => {
          writer
            .tool("get_marketing_options", { input: {} })
            .sleep(500)
            .output({
              audience: "The Xem community",
              template: "The weekly edit",
            });
          writer.text(
            "Let’s make it worth opening. A short note, one useful update, and a clear next step.\n\nWhat would you like your readers to take away?",
          );
        })
        .user(
          "Our custom-domain sending guide. Save a draft for the community.",
        )
        .assistant(({ writer }) => {
          writer.text(
            "Here’s a starting point. A little less setup, a little more room to create.",
          );
          writer
            .tool("create_campaign", {
              input: { name: "Your next great email" },
            })
            .sleep(450)
            .output({
              subject: "Your domain. A clearer path.",
              audience: "The Xem community",
              status: "Ready for your review",
            });
        }),
    [run],
  );
  const transport = useMemo(() => story.transport({ delayMs: 30 }), [story]);
  const { messages, sendMessage, status, stop, setMessages } = useChat({
    id: `showcase-${run}`,
    transport,
  });
  const busy = status === "submitted" || status === "streaming";
  useEffect(() => {
    const reduced = window.matchMedia(
      "(prefers-reduced-motion: reduce)",
    ).matches;
    if (reduced) {
      setMessages(story.get());
      return;
    }
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setPlaying(true);
          setStarted(true);
          observer.disconnect();
        }
      },
      { threshold: 0.35 },
    );
    if (section.current) observer.observe(section.current);
    return () => observer.disconnect();
  }, []);
  useEffect(() => {
    if (!playing || busy) return;
    const next = story.next(messages);
    if (!next) {
      setPlaying(false);
      return;
    }
    const timer = setTimeout(
      () => {
        void sendMessage(next);
      },
      messages.length ? 1700 : 700,
    );
    return () => clearTimeout(timer);
  }, [playing, busy, messages, story, sendMessage]);
  useEffect(() => {
    if (scroll.current) scroll.current.scrollTop = scroll.current.scrollHeight;
  }, [messages]);
  const restart = () => {
    void stop();
    setMessages([]);
    setRun((r) => r + 1);
    setPlaying(true);
    setStarted(true);
  };
  return (
    <div ref={section} className="relative mx-auto w-full max-w-[620px]">
      <div
        className="absolute -inset-5 rounded-[34px] bg-[#ded3ec]/50 md:-inset-7"
        aria-hidden
      />
      <div className="relative overflow-hidden rounded-[20px] border border-[#dcd8ce] bg-[#faf9f5] shadow-[0_24px_80px_#4730590d]">
        <div className="flex h-16 items-center justify-between border-b border-[#e8e4d9] px-5">
          <span className="flex items-center gap-2 text-xs font-medium text-[#514c41]">
            <Sparkles size={15} strokeWidth={1.5} /> Ask Xem
          </span>
          <span className="rounded border border-[#ddd9ce] px-1.5 py-0.5 text-[9px] text-[#696256]">
            Interactive preview
          </span>
        </div>
        <div
          ref={scroll}
          tabIndex={0}
          className="h-[375px] overflow-y-auto px-5 py-6 [scrollbar-width:thin] sm:px-7"
          aria-label="Illustrative assistant conversation"
          aria-live={playing ? "off" : "polite"}
        >
          {messages.length === 0 && (
            <div className="grid h-full content-center text-center">
              <Sparkles
                className="mx-auto mb-5 text-[#8868a6]"
                size={27}
                strokeWidth={1.2}
              />
              <p className="font-editorial text-3xl tracking-tight text-[#5a416b]">
                What’s on your mind?
              </p>
              <p className="mt-3 text-xs text-[#79658a]">
                Your next great email starts here.
              </p>
            </div>
          )}
          <div className="space-y-5">
            {messages.map((message) => (
              <div
                key={message.id}
                className={
                  message.role === "user"
                    ? "ml-9 rounded-xl bg-[#efeee8] px-4 py-3 text-xs leading-relaxed text-[#45433c]"
                    : "space-y-3 text-xs leading-[1.7] text-[#3d3b35]"
                }
              >
                {message.parts.map((part, i) => {
                  if (part.type === "text")
                    return (
                      <p
                        key={i}
                        className="whitespace-pre-wrap"
                        style={
                          message.role === "assistant"
                            ? { fontFamily: "Georgia, serif", fontSize: 15 }
                            : undefined
                        }
                      >
                        {part.text}
                      </p>
                    );
                  if (!("toolCallId" in part)) return null;
                  const name =
                    "toolName" in part
                      ? part.toolName
                      : part.type.replace("tool-", "");
                  if (name === "get_marketing_options")
                    return (
                      <div
                        key={i}
                        className="flex items-center gap-2 rounded-lg border border-[#e2dfd5] px-3 py-2 text-[10px] text-[#6e685d]"
                      >
                        <Check size={12} /> Checked your workspace{" "}
                        <ChevronDown className="ml-auto" size={12} />
                      </div>
                    );
                  if (part.state !== "output-available") return null;
                  return (
                    <div
                      key={i}
                      className="rounded-xl border border-[#dcd8ce] bg-white"
                    >
                      <div className="flex items-center gap-2 border-b border-[#eae6db] px-4 py-3 font-medium">
                        <Mail size={14} /> A newsletter, ready to begin
                      </div>
                      <dl className="space-y-2 px-4 py-3 text-[11px]">
                        <div className="flex gap-3">
                          <dt className="w-14 text-[#736b5d]">Subject</dt>
                          <dd>Your domain. A clearer path.</dd>
                        </div>
                        <div className="flex gap-3">
                          <dt className="w-14 text-[#736b5d]">Audience</dt>
                          <dd>The Xem community</dd>
                        </div>
                      </dl>
                      <div className="flex items-center justify-between gap-2 border-t border-[#eae6db] px-4 py-3">
                        <span className="text-[9px] text-[#736b5d]">
                          Nothing sent. Your call.
                        </span>
                        <span className="rounded-md bg-[#9d5135] px-3 py-1.5 text-[10px] text-white">
                          Review draft{" "}
                          <ArrowUpRight size={10} className="ml-1 inline" />
                        </span>
                      </div>
                    </div>
                  );
                })}
              </div>
            ))}
          </div>
        </div>
        <div className="mx-5 mb-4 flex items-center justify-between gap-3 rounded-2xl border border-[#d8d5ca] bg-[#fffefa] px-4 py-3 shadow-sm">
          <p className="text-xs text-[#746e62]">Reply to Xem…</p>
          <span className="grid h-7 w-7 place-items-center rounded-lg bg-[#9d5135] text-white">
            <ArrowUp size={14} />
          </span>
        </div>
      </div>
      <div className="relative mt-5 flex items-center justify-between gap-4 text-[10px] text-[#715680]">
        <span>Scripted example · No account data or real sends</span>
        <div className="flex gap-2">
          <button
            aria-label={
              playing ? "Pause chat animation" : "Play chat animation"
            }
            className="flex min-h-10 items-center gap-1.5 rounded-lg px-2 hover:bg-white/50"
            onClick={() => {
              if (playing) {
                setPlaying(false);
                void stop();
              } else if (started) restart();
              else {
                setStarted(true);
                setPlaying(true);
              }
            }}
          >
            {playing ? <Pause size={12} /> : <Play size={12} />}{" "}
            {playing ? "Pause" : "Play"}
          </button>
          <button
            aria-label="Replay chat animation"
            className="min-h-10 px-2 hover:text-iris"
            onClick={restart}
          >
            <RotateCcw size={12} />
          </button>
        </div>
      </div>
    </div>
  );
}
export function AssistantShowcase() {
  return (
    <section
      id="assistant"
      aria-labelledby="assistant-heading"
      className="overflow-hidden bg-[#eee6f5] px-5 py-24 text-[#372640] md:px-10 md:py-32"
    >
      <div className="mx-auto grid max-w-[1160px] items-center gap-16 lg:grid-cols-[.85fr_1.15fr] lg:gap-24">
        <div>
          <p className="text-[11px] font-medium uppercase tracking-[.17em] text-[#715080]">
            A little less clicking. A little more creating.
          </p>
          <h2
            id="assistant-heading"
            className="mt-6 font-editorial text-[clamp(3.2rem,5.4vw,4.9rem)] font-normal leading-[1.02] tracking-[-.045em]"
          >
            Your workspace.
            <br />
            <em>In a conversation.</em>
          </h2>
          <p className="mt-7 max-w-sm text-sm leading-[1.9] text-[#6b507b]">
            Find the story in your numbers. Give your next newsletter a starting
            point. Or just ask where to begin.
          </p>
          <p className="mt-4 max-w-sm text-sm leading-[1.9] text-[#6b507b]">
            Ask Xem connects to your team’s tools, so help comes with context.
            You review workspace changes before they happen.
          </p>
          <div className="mt-8 flex items-center gap-2 text-[10px] text-[#77528a]">
            <span className="h-1.5 w-1.5 rounded-full bg-[#9775ae]" /> In
            development · Preview the experience
          </div>
          <a
            href={appLink("/")}
            className="mt-7 inline-flex min-h-11 items-center gap-2 border-b border-[#8f6ba1] pb-1 text-sm font-medium text-[#68427c]"
          >
            Explore your workspace <ArrowUpRight size={16} />
          </a>
          <p className="mt-6 max-w-xs text-[11px] leading-relaxed text-[#72567e]">
            Built with the AI SDK and Xem’s hosted MCP. Team-scoped access, with
            a human at the send button.
          </p>
        </div>
        <ConversationDemo />
      </div>
    </section>
  );
}
