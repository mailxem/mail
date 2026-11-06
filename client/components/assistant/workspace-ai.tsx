"use client";

import {
  createContext,
  useCallback,
  useContext,
  useDeferredValue,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { usePathname } from "next/navigation";
import Link from "next/link";
import { useTeam } from "@/app/providers/team-provider";
import { useSession } from "next-auth/react";
import {
  ArrowUp,
  Check,
  FilePenLine,
  Loader2,
  Minus,
  Square,
} from "lucide-react";
import { AssistantHome } from "./assistant-home";
import type { EmailDraft } from "@/components/ai/email-writer";
import { XemMark, XemMarkTile } from "./xem-mark";
import styles from "./workspace-ai.module.css";
import {
  applySlots,
  extractSlots,
  rankReferences,
  designImages,
  rowsOf,
} from "@/lib/assistant/reference-design";
import { loadStarterDesign } from "@/lib/template-starters/registry";
import { GmailClipNotice } from "@/components/campaigns/gmail-clip-notice";

export type DesignSnapshot = {
  key: string;
  revision: number;
  subject: string;
  html: string;
  design: unknown;
};
export type DesignEditor = {
  key: string;
  name: string;
  snapshot: () => Promise<DesignSnapshot>;
  apply: (draft: EmailDraft, previous: DesignSnapshot) => Promise<void>;
  undo: () => Promise<void>;
};
type WorkspaceAI = {
  register: (editor: DesignEditor) => () => void;
  openDesigner: () => void;
};
const Context = createContext<WorkspaceAI | null>(null);
export const useWorkspaceAI = () => useContext(Context);

export function WorkspaceAIProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  const { team } = useTeam();
  const { data: session } = useSession();
  return (
    <WorkspaceAI key={`${session?.user?.email}:${team?.id}`} teamId={team?.id}>
      {children}
    </WorkspaceAI>
  );
}
function WorkspaceAI({
  children,
  teamId,
}: {
  children: React.ReactNode;
  teamId?: string;
}) {
  const pathname = usePathname();
  const [editor, setEditor] = useState<DesignEditor | null>(null);
  const [open, setOpen] = useState(false);
  const [started, setStarted] = useState(false);
  const [mode, setMode] = useState<"design" | "workspace">("workspace");
  const launcher = useRef<HTMLButtonElement>(null);
  const panel = useRef<HTMLDivElement>(null);
  // The design conversation survives a new template getting its id on first
  // save (/templates/new -> /templates/<id>/edit); any other switch starts fresh.
  const thread = useRef({ key: "", current: "" });
  if (editor) {
    // Editor keys end in ":new" until a template is first saved.
    const current = editor.key;
    if (current !== thread.current.current) {
      const fromNew = thread.current.current.endsWith(":new");
      thread.current = {
        key: fromNew && thread.current.key ? thread.current.key : current,
        current,
      };
    }
  }
  const register = useCallback((next: DesignEditor) => {
    setEditor(next);
    return () => setEditor((current) => (current === next ? null : current));
  }, []);
  const openDesigner = useCallback(() => {
    setMode("design");
    setOpen(true);
  }, []);
  useEffect(() => {
    if (open) {
      setStarted(true);
      panel.current?.focus();
    }
  }, [open]);
  const close = () => {
    setOpen(false);
    launcher.current?.focus();
  };
  return (
    <Context.Provider value={{ register, openDesigner }}>
      {children}
      {teamId && (
        <div className={styles.widget}>
          <div
            hidden={!open}
            className={styles.panel}
            role="dialog"
            aria-label="Xem AI"
            tabIndex={-1}
            ref={panel}
            onKeyDown={(event) => {
              if (event.key === "Escape") {
                event.stopPropagation();
                close();
              }
            }}
          >
            <header className={styles.header}>
              <XemMarkTile size={32} />
              <div>
                <strong>Xem AI</strong>
                <small>A little help, right where you work</small>
              </div>
              <button aria-label="Minimize Xem AI" onClick={close}>
                <Minus size={18} />
              </button>
            </header>
            <div className={styles.modes} aria-label="Assistant mode">
              <button
                aria-pressed={mode === "workspace"}
                onClick={() => setMode("workspace")}
              >
                Workspace
              </button>
              <button
                aria-pressed={mode === "design"}
                onClick={() => setMode("design")}
              >
                <FilePenLine size={14} />
                Email design
              </button>
            </div>
            <div hidden={mode !== "workspace"} className={styles.modeBody}>
              {started && <AssistantHome embedded />}
            </div>
            <div hidden={mode !== "design"} className={styles.modeBody}>
              {editor ? (
                <DesignConversation
                  key={thread.current.key}
                  thread={thread.current.key}
                  editor={editor}
                  teamId={teamId}
                />
              ) : (
                <div className={styles.welcome}>
                  <XemMarkTile size={44} />
                  <h2>Let’s make something worth opening.</h2>
                  <p>
                    Open a template and tell Xem what you have in mind. Your
                    revisions appear in the editor and save as you go.
                  </p>
                  <Link href="/templates/new" onClick={close}>
                    Create an email template →
                  </Link>
                  <Link href="/templates" onClick={close}>
                    Browse your templates
                  </Link>
                </div>
              )}
            </div>
          </div>
          <button
            ref={launcher}
            hidden={pathname === "/" && !open}
            className={styles.launcher}
            aria-expanded={open}
            aria-label={open ? "Minimize Xem AI" : "Open Xem AI"}
            onClick={() =>
              open
                ? close()
                : (setMode(editor ? "design" : "workspace"), setOpen(true))
            }
          >
            <XemMark size={18} />
            <span>Xem AI</span>
          </button>
        </div>
      )}
    </Context.Provider>
  );
}

type DesignMessage = { text: string; user?: boolean };
// Conversations outlive the panel's remount when a new template is first saved
// and the editor briefly unregisters while the page changes.
const designThreads = new Map<string, DesignMessage[]>();

function DesignConversation({
  editor,
  teamId,
  thread,
}: {
  editor: DesignEditor;
  teamId: string;
  thread: string;
}) {
  const [input, setInput] = useState("");
  const [messages, setMessages] = useState<DesignMessage[]>(
    () => designThreads.get(thread) ?? [],
  );
  useEffect(() => {
    designThreads.set(thread, messages);
  }, [thread, messages]);
  const [phase, setPhase] = useState<"" | "designing" | "saving">("");
  const [error, setError] = useState("");
  const [undo, setUndo] = useState(false);
  const [candidate, setCandidate] = useState<EmailDraft | null>(null);
  // Which design to build from: a library key, "current" for the open
  // template, or null to use the best match for the brief.
  const [reference, setReference] = useState<string | null>(null);
  const [hasDesign, setHasDesign] = useState(false);
  const brief = useDeferredValue(input);
  const suggested = useMemo(
    () => rankReferences(brief.trim() || "newsletter editorial weekly", 6),
    [brief],
  );
  useEffect(() => {
    editor
      .snapshot()
      .then((s) => {
        if (active.current)
          setHasDesign(extractSlots(s.design as never).length >= 3);
      })
      .catch(() => {});
  }, [editor]);
  const active = useRef(true);
  const pending = useRef<AbortController | null>(null);
  const transcript = useRef<HTMLDivElement>(null);
  useEffect(() => {
    active.current = true;
    return () => {
      active.current = false;
      pending.current?.abort();
    };
  }, []);
  useEffect(() => {
    if (transcript.current)
      transcript.current.scrollTop = transcript.current.scrollHeight;
  }, [messages, phase, error]);
  async function send(text = input) {
    if (pending.current || phase || text.trim().length < 3) return;
    const controller = new AbortController();
    pending.current = controller;
    setError("");
    setCandidate(null);
    setPhase("designing");
    setMessages((rows) => [...rows, { text, user: true }]);
    setInput("");
    try {
      const previous = await editor.snapshot();
      if (controller.signal.aborted) return;
      // Build from a finished design: the chosen library template, the open
      // template when it already has copy, or the closest match for the brief.
      const useCurrent =
        reference === "current" ||
        (!reference && extractSlots(previous.design as never).length >= 3);
      const starter = useCurrent
        ? null
        : reference
          ? (suggested.find((t) => t.key === reference) ??
            rankReferences(text, 1)[0])
          : rankReferences(text, 1)[0];
      const base = (
        useCurrent
          ? previous.design
          : await loadStarterDesign(starter!.key, controller.signal)
      ) as Parameters<typeof applySlots>[0];
      if (controller.signal.aborted) return;
      const slots = extractSlots(base);
      const images = await designImages(text, starter, controller.signal);
      if (controller.signal.aborted) return;
      const response = await fetch("/api/assistant/design", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        signal: controller.signal,
        body: JSON.stringify({
          teamId,
          format: "slots",
          instruction: text,
          reference: useCurrent
            ? `The user's current template: ${editor.name || "untitled"}`
            : `${starter!.name}: ${starter!.description}`,
          slots,
          rows: rowsOf(base),
          ...(images.length ? { images } : {}),
        }),
      });
      const answer = await response.json();
      if (!response.ok)
        throw new Error(
          answer.error || "Couldn’t generate this revision. Try again.",
        );
      if (!answer.subject || !answer.copy || !Object.keys(answer.copy).length)
        throw new Error(
          "Xem returned an incomplete design. Your template is unchanged.",
        );
      const draft: EmailDraft = {
        subject: answer.subject,
        body: answer.body,
        design: applySlots(base, {
          subject: answer.subject,
          preheader: answer.preheader,
          slots: answer.copy,
          remove: answer.remove,
        }) as EmailDraft["design"],
      };
      const usedName = useCurrent ? "your current design" : starter!.name;
      if (!active.current || controller.signal.aborted) return;
      setCandidate(draft);
      setPhase("saving");
      setUndo(false);
      await editor.apply(draft, previous);
      if (active.current) {
        setCandidate(null);
        setUndo(true);
        setMessages((rows) => [
          ...rows,
          {
            text: `Built from ${usedName} and saved. Take a look in the editor, then tell me what to change.`,
          },
        ]);
      }
    } catch (e) {
      if (active.current && !controller.signal.aborted) {
        setError(
          e instanceof TypeError
            ? "Couldn’t reach Xem. Check your connection and try again. Your local edits are still here."
            : (e as Error).message,
        );
        setInput(text);
      }
    } finally {
      pending.current = null;
      if (active.current) setPhase("");
    }
  }
  return (
    <>
      <div className={styles.context}>
        <FilePenLine size={15} />
        <span>{editor.name || "Untitled template"}</span>
        <small>Revisions auto-save</small>
      </div>
      <div className={styles.transcript} ref={transcript}>
        {!messages.length && (
          <div className={styles.welcome}>
            <XemMarkTile size={44} />
            <h2>Your next great email starts here.</h2>
            <p>
              Describe the look, the audience, and what you want to say. Then
              keep refining it together.
            </p>
            <div className={styles.references}>
              <p>Start from a design</p>
              <div role="radiogroup" aria-label="Design to start from">
                {hasDesign && (
                  <button
                    type="button"
                    role="radio"
                    aria-checked={reference === "current" || !reference}
                    className={styles.referenceCurrent}
                    onClick={() => setReference("current")}
                  >
                    <FilePenLine size={18} />
                    <span>Your current design</span>
                  </button>
                )}
                {suggested.slice(0, hasDesign ? 5 : 6).map((t, i) => (
                  <button
                    key={t.key}
                    type="button"
                    role="radio"
                    aria-checked={
                      reference === t.key || (!reference && !hasDesign && i === 0)
                    }
                    title={t.description}
                    onClick={() => setReference(t.key)}
                  >
                    <span className={styles.thumb}>
                      <iframe
                        src={t.previewUrl}
                        title=""
                        tabIndex={-1}
                        loading="lazy"
                        sandbox=""
                        aria-hidden="true"
                      />
                    </span>
                    <span>{t.name}</span>
                  </button>
                ))}
              </div>
              <small>
                Xem keeps the layout, fonts and colours, and writes your copy
                into it. {reference ? "" : "Picks the closest match if you don’t choose."}
              </small>
            </div>
            <div className={styles.suggestions}>
              {[
                "A warm welcome email for new newsletter subscribers",
                "Announce our new paid tier with deep-dive essays",
              ].map((text) => (
                <button key={text} onClick={() => void send(text)}>
                  {text}
                  <ArrowUp size={14} />
                </button>
              ))}
            </div>
          </div>
        )}
        {messages.map((message, index) => (
          <div
            key={index}
            className={message.user ? styles.user : styles.answer}
          >
            {!message.user && <Check size={15} />}
            <p>{message.text}</p>
          </div>
        ))}
        {phase && (
          <p role="status" className={styles.progress}>
            <Loader2 size={15} className="animate-spin" />
            {phase === "designing"
              ? "Designing your revision…"
              : "Updating and saving your template…"}
          </p>
        )}
        {error && (
          <p role="alert" className={styles.error}>
            {error}
          </p>
        )}
        {candidate && !phase && (
          <div className={styles.recovery}>
            <p>Your generated revision is available to keep or review.</p>
            {candidate.previewHtml && (
              <details>
                <summary>Preview revision</summary>
                <iframe
                  title="Unsaved AI revision"
                  sandbox=""
                  referrerPolicy="no-referrer"
                  srcDoc={candidate.previewHtml}
                />
              </details>
            )}
            {candidate.previewHtml && (
              <GmailClipNotice html={candidate.previewHtml} />
            )}
            <button
              onClick={() => {
                const url = URL.createObjectURL(
                  new Blob([JSON.stringify(candidate, null, 2)], {
                    type: "application/json",
                  }),
                );
                const a = document.createElement("a");
                a.href = url;
                a.download = "xem-email-revision.json";
                a.click();
                setTimeout(() => URL.revokeObjectURL(url), 1000);
              }}
            >
              Download revision
            </button>
            <button
              onClick={async () => {
                if (pending.current) return;
                setPhase("saving");
                setError("");
                try {
                  await editor.apply(candidate, await editor.snapshot());
                  setCandidate(null);
                  setUndo(true);
                  setMessages((rows) => [
                    ...rows,
                    { text: "Your reviewed revision is applied and saved." },
                  ]);
                } catch (e) {
                  setError((e as Error).message);
                } finally {
                  setPhase("");
                }
              }}
            >
              Replace current design with this revision
            </button>
          </div>
        )}
        {undo && (
          <button
            className={styles.undo}
            disabled={!!phase}
            onClick={async () => {
              setPhase("saving");
              setError("");
              try {
                await editor.undo();
                setUndo(false);
                setMessages((rows) => [
                  ...rows,
                  { text: "Previous design restored and saved." },
                ]);
              } catch (e) {
                setError((e as Error).message);
              } finally {
                setPhase("");
              }
            }}
          >
            Undo last revision
          </button>
        )}
      </div>
      <form
        className={styles.composer}
        onSubmit={(e) => {
          e.preventDefault();
          void send();
        }}
      >
        <label className="sr-only" htmlFor="xem-design-message">
          Describe your email or revision
        </label>
        <textarea
          id="xem-design-message"
          rows={2}
          value={input}
          maxLength={4000}
          placeholder="Describe your email, or ask for a change…"
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => {
            if (
              e.key === "Enter" &&
              !e.shiftKey &&
              !e.nativeEvent.isComposing
            ) {
              e.preventDefault();
              void send();
            }
          }}
        />
        {phase === "designing" ? (
          <button
            type="button"
            aria-label="Stop designing"
            onClick={() => pending.current?.abort()}
          >
            <Square size={15} />
          </button>
        ) : (
          <button
            aria-label="Design and save revision"
            disabled={!!phase || input.trim().length < 3}
          >
            <ArrowUp size={18} />
          </button>
        )}
      </form>
      <p className={styles.note}>
        Your prompt and current email go to the AI provider. Review facts and
        links before sending. AI never sends this email.
      </p>
    </>
  );
}
