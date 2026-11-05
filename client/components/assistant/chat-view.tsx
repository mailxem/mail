"use client";
import { useEffect, useRef, useState, useId } from "react";
import type { UIMessage } from "ai";
import {
  ArrowDown,
  ArrowUp,
  Check,
  ChevronDown,
  Copy,
  History,
  Mail,
  MessageSquare,
  Plus,
  Square,
  Users,
  X,
  ChartNoAxesCombined,
  Loader2,
  Trash2,
  RefreshCw,
} from "lucide-react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import type { ActionOutput, ConversationSummary } from "@/lib/assistant/types";
import styles from "./chat.module.css";
import { GlanceLine } from "./workspace-glance";
import { useSession } from "next-auth/react";
import { XemMarkTile } from "./xem-mark";

const starters = [
  {
    icon: Mail,
    title: "Draft a newsletter",
    color: "#d6457c",
    text: "Help me draft a newsletter. First ask me about the audience and what I want to share.",
    detail: "Find the words for your next newsletter",
  },
  {
    icon: ChartNoAxesCombined,
    title: "Review results",
    color: "#d08a12",
    text: "How are my recent campaigns performing? Use my workspace data and suggest one useful next step.",
    detail: "Understand what’s working",
  },
  {
    icon: Users,
    title: "Know my audience",
    color: "#2f9e5b",
    text: "Give me an overview of my audience using aggregate metrics. What should I pay attention to?",
    detail: "Turn audience insights into a next step",
  },
];
const sources = [
  { name: "Mailchimp", logo: "mailchimp", tile: "#FFE01B" },
  { name: "Substack", logo: "substack", tile: "#FF6719" },
  { name: "Kit", logo: "kit", tile: "#111111" },
  { name: "Ghost", logo: "ghost", tile: "#15171A" },
];
function greeting() {
  const h = new Date().getHours();
  return h < 12 ? "Good morning" : h < 18 ? "Good afternoon" : "Good evening";
}
export const messageText = (message: UIMessage) =>
  message.parts.flatMap((p) => (p.type === "text" ? [p.text] : [])).join("\n");
function label(name: string) {
  return name.replace(/_/g, " ").replace(/^./, (c) => c.toUpperCase());
}
function Markdown({ text }: { text: string }) {
  return (
    <div className={styles.markdown}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        skipHtml
        components={{
          img: () => null,
          a: ({ href, children }) => (
            <a
              href={href}
              target={href?.startsWith("/") ? undefined : "_blank"}
              rel="noopener noreferrer"
            >
              {children}
            </a>
          ),
          table: ({ children }) => (
            <div className={styles.table}>
              <table>{children}</table>
            </div>
          ),
        }}
      >
        {text}
      </ReactMarkdown>
    </div>
  );
}
function Output({ value }: { value: unknown }) {
  return (
    <pre className={styles.output}>
      {typeof value === "string" ? value : JSON.stringify(value, null, 2)}
    </pre>
  );
}
function ToolPart({
  part,
  busy,
  onAction,
}: {
  part: UIMessage["parts"][number];
  busy: boolean;
  onAction: (id: string, approved: boolean) => void;
}) {
  if (!("toolCallId" in part)) return null;
  const name =
    "toolName" in part
      ? String(part.toolName)
      : part.type.replace(/^tool-/, "");
  const output = "output" in part ? part.output : undefined;
  const action =
    (output as ActionOutput)?.kind === "xem-action"
      ? (output as ActionOutput)
      : undefined;
  const loading =
    part.state === "input-streaming" || part.state === "input-available";
  if (action) {
    const details = (
      <>
        <dl className={styles.arguments}>
          {Object.entries(action.input).map(([key, value]) => (
            <div key={key}>
              <dt>{label(key)}</dt>
              <dd>
                {typeof value === "object" ? (
                  <Output value={value} />
                ) : (
                  String(value)
                )}
              </dd>
            </div>
          ))}
        </dl>
        {action.result != null && (
          <details className={styles.result}>
            <summary>
              View result <ChevronDown size={14} />
            </summary>
            <Output value={action.result} />
          </details>
        )}
      </>
    );
    if (action.status === "completed" || action.status === "declined") {
      return (
        <details className={styles.resolved}>
          <summary>
            {action.status === "completed" ? (
              <Check size={16} />
            ) : (
              <X size={16} />
            )}
            <span>
              <strong>{label(action.tool)}</strong>
              <span>
                {action.status === "completed"
                  ? "Completed in your workspace"
                  : "Declined · nothing changed"}
              </span>
            </span>
            <ChevronDown size={14} />
          </summary>
          {details}
        </details>
      );
    }
    return (
      <section
        className={styles.action}
        aria-label={`Review ${label(action.tool)}`}
      >
        <div className={styles.actionHeading}>
          <span className={styles.actionIcon}>
            <Mail size={17} />
          </span>
          <div>
            <strong>{label(action.tool)}</strong>
            <p>
              {action.status === "pending"
                ? "Ready for your review"
                : action.status === "running"
                  ? "Outcome pending · check your workspace"
                  : "Outcome unconfirmed · check before retrying"}
            </p>
          </div>
        </div>
        {details}
        {action.status === "pending" && (
          <div className={styles.actionFooter}>
            <span>Only these details will be used.</span>
            <button
              disabled={busy}
              onClick={() => onAction(action.actionId, false)}
            >
              Decline
            </button>
            <button
              className={styles.approve}
              disabled={busy}
              onClick={() => onAction(action.actionId, true)}
            >
              <Check size={14} /> Approve
            </button>
          </div>
        )}
      </section>
    );
  }
  return (
    <details className={styles.tool}>
      <summary>
        {loading ? (
          <Loader2 size={14} className={styles.spin} />
        ) : part.state === "output-error" ? (
          <X size={14} />
        ) : (
          <Check size={14} />
        )}
        <span>{label(name)}</span>
        <ChevronDown size={13} />
      </summary>
      {output != null && <Output value={output} />}{" "}
      {part.state === "output-error" && (
        <p>Couldn’t complete this lookup. Try again in a moment.</p>
      )}
    </details>
  );
}
export interface ChatViewProps {
  compact?: boolean;
  messages: UIMessage[];
  input: string;
  setInput: (v: string) => void;
  busy: boolean;
  loading?: boolean;
  error?: string;
  enabled?: boolean;
  canWrite?: boolean;
  demo?: boolean;
  onSend: (text: string) => void;
  onStop: () => void;
  onNew: () => void;
  onAction: (id: string, approved: boolean) => void;
  conversations?: ConversationSummary[];
  onOpen?: (id: string) => void;
  onDelete?: (id: string) => void;
  onRecover?: () => void;
}
export function ChatView({
  compact = false,
  messages,
  input,
  setInput,
  busy,
  loading,
  error,
  enabled = true,
  canWrite = true,
  demo,
  onSend,
  onStop,
  onNew,
  onAction,
  conversations = [],
  onOpen,
  onDelete,
  onRecover,
}: ChatViewProps) {
  const empty = messages.length === 0;
  const { data: session } = useSession();
  const firstName = session?.user?.name?.split(" ")[0] ?? "";
  const viewport = useRef<HTMLDivElement>(null);
  const follow = useRef(true);
  const field = useRef<HTMLTextAreaElement>(null);
  const [showLatest, setShowLatest] = useState(false);
  const [copied, setCopied] = useState<string | null>(null);
  useEffect(() => {
    follow.current = true;
    setShowLatest(false);
  }, [messages[0]?.id]);
  useEffect(() => {
    if (follow.current && viewport.current)
      viewport.current.scrollTop = viewport.current.scrollHeight;
  }, [messages, busy]);
  useEffect(() => {
    if (field.current) {
      field.current.style.height = "auto";
      field.current.style.height = `${Math.min(field.current.scrollHeight, 180)}px`;
    }
  }, [input]);
  function send(text = input) {
    if (!text.trim() || busy || !enabled) return;
    follow.current = true;
    onSend(text.trim());
  }
  return (
    <div
      className={`${styles.screen} ${compact ? styles.embedded : ""} ${empty ? "" : styles.transcript}`}
      data-assistant
    >
      <header className={styles.header}>
        <div className={styles.identity}>
          {empty && <XemMarkTile size={24} />}
          <span
            className={styles.wordmark}
            title={empty ? "Ask Xem" : messageText(messages[0])}
          >
            {empty ? "Ask Xem" : messageText(messages[0]).slice(0, 80)}
          </span>
          <span className={styles.beta}>{demo ? "Preview" : "Beta"}</span>
        </div>
        <div className={styles.headerActions}>
          <details className={styles.history}>
            <summary aria-label="Conversation history">
              <History size={17} />
              <span>History</span>
            </summary>
            <div className={styles.historyMenu}>
              <p>
                Your conversations <span>Saved for 7 days</span>
              </p>
              {conversations.length === 0 ? (
                <div className={styles.historyEmpty}>
                  A fresh page. Your chats will appear here.
                </div>
              ) : (
                conversations.map((c) => (
                  <div className={styles.historyItem} key={c.id}>
                    <button
                      disabled={busy}
                      onClick={(e) => {
                        onOpen?.(c.id);
                        e.currentTarget
                          .closest("details")
                          ?.removeAttribute("open");
                      }}
                    >
                      <MessageSquare size={14} />
                      <span>{c.title}</span>
                    </button>
                    <button
                      aria-label={`Delete ${c.title}`}
                      title="Delete conversation"
                      disabled={busy}
                      onClick={() => onDelete?.(c.id)}
                    >
                      <Trash2 size={14} />
                    </button>
                  </div>
                ))
              )}
            </div>
          </details>
          <button
            onClick={() => {
              onNew();
              field.current?.focus();
            }}
            disabled={busy}
            title="New conversation"
            aria-label="New conversation"
          >
            <Plus size={17} />
            <span>New chat</span>
          </button>
        </div>
      </header>
      <div
        className={styles.viewport}
        ref={viewport}
        onScroll={() => {
          const e = viewport.current;
          if (e) {
            follow.current =
              e.scrollHeight - e.scrollTop - e.clientHeight < 100;
            setShowLatest(!follow.current);
          }
        }}
      >
        {empty ? (
          <div className={styles.welcome}>
            <XemMarkTile size={48} />
            <h1>
              <span className={styles.greeting}>
                {greeting()}
                {firstName ? `, ${firstName}` : ""}.
              </span>{" "}
              What will you get done <em>today</em>?
            </h1>
            {compact || demo ? (
              <p className={styles.subtitle}>
                Ask about your campaigns, audience or inbox.
              </p>
            ) : (
              <GlanceLine />
            )}
          </div>
        ) : (
          <div
            className={styles.messages}
            aria-label="Conversation"
            aria-busy={busy}
          >
            {messages.map((message) => (
              <article
                className={
                  message.role === "user"
                    ? styles.userMessage
                    : styles.assistantMessage
                }
                key={message.id}
                aria-label={message.role === "user" ? "You" : "Xem"}
              >
                <div className={styles.messageContent}>
                  {message.parts.map((part, i) =>
                    part.type === "text" ? (
                      message.role === "user" ? (
                        <div key={i} className={styles.userText}>
                          {part.text}
                        </div>
                      ) : (
                        <Markdown key={i} text={part.text} />
                      )
                    ) : "toolCallId" in part ? (
                      <ToolPart
                        key={i}
                        part={part}
                        busy={busy}
                        onAction={onAction}
                      />
                    ) : null,
                  )}
                  {message.role === "assistant" &&
                    messageText(message) &&
                    !busy && (
                      <button
                        className={styles.copy}
                        aria-label={
                          copied === message.id
                            ? "Answer copied"
                            : "Copy answer"
                        }
                        title={copied === message.id ? "Copied" : "Copy answer"}
                        onClick={async () => {
                          try {
                            await navigator.clipboard.writeText(
                              messageText(message),
                            );
                            setCopied(message.id);
                            setTimeout(() => setCopied(null), 1500);
                          } catch {}
                        }}
                      >
                        {copied === message.id ? (
                          <Check size={16} />
                        ) : (
                          <Copy size={16} />
                        )}
                        <span className="sr-only" role="status">
                          {copied === message.id ? "Copied" : "Copy"}
                        </span>
                      </button>
                    )}
                </div>
              </article>
            ))}
            {busy && (
              <div className={styles.thinking} role="status">
                <span />
                <span />
                <span />
                <span className="sr-only">Xem is working</span>
              </div>
            )}
          </div>
        )}
        {empty && (
          <div className={styles.emptyComposer}>
            <Composer
              {...{
                input,
                setInput,
                busy,
                enabled,
                loading,
                field,
                send,
                onStop,
              }}
            />
            <div className={styles.starters}>
              {starters.map((s) => (
                <button
                  key={s.title}
                  disabled={busy || loading || !enabled}
                  onClick={() => send(s.text)}
                >
                  <s.icon size={18} strokeWidth={1.75} style={{ color: s.color }} />
                  {s.title}
                </button>
              ))}
              {!compact && !demo && (
                <button
                  disabled={busy || loading || !enabled}
                  onClick={() =>
                    send(
                      "I’m moving to Xem from another newsletter tool. Walk me through exporting my subscribers and importing them here.",
                    )
                  }
                >
                  <span className={styles.logoStack} aria-hidden>
                    {sources.map((source) => (
                      <span key={source.name} style={{ background: source.tile }}>
                        {/* eslint-disable-next-line @next/next/no-img-element */}
                        <img src={`/logos/${source.logo}.svg`} alt="" />
                      </span>
                    ))}
                  </span>
                  Import your list
                </button>
              )}
            </div>
          </div>
        )}
      </div>
      <div className={empty ? styles.emptyFooter : styles.composerDock}>
        {!empty && showLatest && (
          <button
            className={styles.latest}
            aria-label="Scroll to latest message"
            onClick={() => {
              follow.current = true;
              setShowLatest(false);
              viewport.current?.scrollTo({
                top: viewport.current.scrollHeight,
                behavior: "smooth",
              });
            }}
          >
            <ArrowDown size={17} />
          </button>
        )}
        {error && (
          <div role="alert" className={styles.error}>
            {error}
            {onRecover && (
              <button disabled={busy} onClick={onRecover}>
                <RefreshCw size={13} /> Reopen saved chat
              </button>
            )}
          </div>
        )}
        {!enabled && !loading && (
          <p role="status" className={styles.unavailable}>
            Your assistant is being connected. You can keep using the workspace
            from the sidebar.
          </p>
        )}
        {!empty && (
          <Composer
            compact
            {...{
              input,
              setInput,
              busy,
              enabled,
              loading,
              field,
              send,
              onStop,
            }}
          />
        )}
        <p className={styles.footnote}>
          {demo
            ? "Scripted preview · Sample data · No real actions"
            : canWrite
              ? "Connected to your workspace. You review changes before they happen."
              : "Read-only workspace access. Ask an admin to make changes."}
          <span>AI can make mistakes.</span>
        </p>
        {!demo && (
          <details className={styles.dataNote}>
            <summary>About your data</summary>
            <p>
              Your messages and relevant workspace results are sent to Xem’s
              configured AI provider. Conversations are private to your account
              and workspace, kept for seven days after activity, and can be
              deleted from History. Avoid sharing passwords or secrets.
            </p>
          </details>
        )}
      </div>
    </div>
  );
}
function Composer({
  input,
  setInput,
  busy,
  enabled,
  loading,
  field,
  send,
  onStop,
  compact = false,
}: {
  compact?: boolean;
  input: string;
  setInput: (s: string) => void;
  busy: boolean;
  enabled: boolean;
  loading?: boolean;
  field: React.RefObject<HTMLTextAreaElement>;
  send: () => void;
  onStop: () => void;
}) {
  const inputId = useId();
  return (
    <form
      className={`${styles.composer} ${compact ? styles.compactComposer : ""}`}
      onSubmit={(e) => {
        e.preventDefault();
        send();
      }}
    >
      <label className="sr-only" htmlFor={inputId}>
        Message Xem
      </label>
      <textarea
        ref={field}
        id={inputId}
        value={input}
        maxLength={8000}
        rows={compact ? 1 : 2}
        disabled={loading || !enabled}
        placeholder={
          loading
            ? "Connecting your workspace…"
            : compact
              ? "Reply to Xem…"
              : "Ask anything about your workspace…"
        }
        onChange={(e) => setInput(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === "Enter" && !e.shiftKey && !e.nativeEvent.isComposing) {
            e.preventDefault();
            send();
          }
        }}
      />
      <div className={styles.composerBottom}>
        {!compact && (
          <>
            <span className={styles.shortcut}>
              {loading ? "Connecting…" : "Shift ↵ for a new line"}
            </span>
          </>
        )}
        {busy ? (
          <button
            type="button"
            onClick={onStop}
            aria-label="Stop response"
            className={styles.send}
          >
            <Square size={13} fill="currentColor" />
          </button>
        ) : (
          <button
            type="submit"
            disabled={!input.trim() || loading || !enabled}
            aria-label="Send message"
            className={styles.send}
          >
            <ArrowUp size={18} />
          </button>
        )}
      </div>
    </form>
  );
}
