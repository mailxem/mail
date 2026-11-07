"use client";
import { useState } from "react";
import type { UIMessage } from "ai";
import {
  Check,
  ChevronDown,
  CircleSlash,
  Loader2,
  Mail,
  PencilLine,
  ShieldCheck,
  X,
} from "lucide-react";
import type { ActionOutput } from "@/lib/assistant/types";
import styles from "./chat.module.css";

type Part = UIMessage["parts"][number];
type ToolishPart = Part & { toolCallId: string; state: string };

const sentence = (s: string) =>
  s.replace(/[_-]+/g, " ").replace(/([a-z])([A-Z])/g, "$1 $2").toLowerCase();
const cap = (s: string) => s.charAt(0).toUpperCase() + s.slice(1);

export const toolName = (part: Part) =>
  "toolName" in part ? String(part.toolName) : part.type.replace(/^tool-/, "");
export const actionOf = (part: Part) => {
  const output = "output" in part ? part.output : undefined;
  return (output as ActionOutput)?.kind === "xem-action"
    ? (output as ActionOutput)
    : undefined;
};

/** "get_contact_lists" while running reads "Looking up contact lists". */
function stepTitle(name: string, done: boolean) {
  const [verb, ...rest] = name.split("_");
  const thing = sentence(rest.join("_"));
  const verbs: Record<string, [string, string]> = {
    get: ["Looking up", "Looked up"],
    list: ["Listing", "Listed"],
    search: ["Searching", "Searched"],
    find: ["Finding", "Found"],
    preview: ["Previewing", "Previewed"],
    check: ["Checking", "Checked"],
  };
  const pair = verbs[verb];
  return pair ? `${pair[done ? 1 : 0]} ${thing}` : cap(sentence(name));
}

/** A short, human summary of a read tool's result, never the raw JSON. */
function resultSummary(output: unknown): string {
  if (output == null) return "";
  if (Array.isArray(output)) return `${output.length} found`;
  if (typeof output === "object") {
    const o = output as Record<string, unknown>;
    if (typeof o.error === "string") return "Couldn’t complete";
    if (typeof o.notice === "string") return "Too much to show";
    for (const key of ["total", "count"])
      if (typeof o[key] === "number") return `${o[key]} found`;
    for (const value of Object.values(o))
      if (Array.isArray(value)) return `${value.length} found`;
  }
  return "Done";
}

function actionResultSummary(a: ActionOutput) {
  if (a.status === "declined") return "Declined · nothing changed";
  if (a.status === "running") return "Running…";
  if (a.status === "unknown") return "Outcome unconfirmed · check before retrying";
  const r = a.result as Record<string, unknown> | undefined;
  const name = r && (r.name || r.subject || r.email);
  for (const key of ["created", "count", "added", "imported"])
    if (r && typeof r[key] === "number") return `Done · ${r[key]} ${key}`;
  return name ? `Done · ${String(name)}` : "Done";
}

/** The steps an answer took: lookups and approved changes, collapsed into one line. */
export function RunSummary({ parts, live }: { parts: Part[]; live: boolean }) {
  const [open, setOpen] = useState(false);
  const steps = parts.filter((p): p is ToolishPart => "toolCallId" in p);
  if (!steps.length) return null;
  const running = steps.find(
    (p) => p.state === "input-streaming" || p.state === "input-available",
  );
  const changes = steps.filter((p) => {
    const a = actionOf(p);
    return a && a.status === "completed";
  }).length;
  const failed = steps.some((p) => p.state === "output-error");
  return (
    <div className={styles.run}>
      <button
        type="button"
        className={styles.runLine}
        aria-expanded={open}
        onClick={() => setOpen(!open)}
      >
        {running && live ? (
          <Loader2 size={14} className={styles.spin} />
        ) : failed ? (
          <CircleSlash size={14} />
        ) : (
          <Check size={14} />
        )}
        <span className={styles.runStatus}>
          {running && live ? stepTitle(toolName(running), false) + "…" : "Worked"}
        </span>
        <span className={styles.runMeta}>
          {steps.length} {steps.length === 1 ? "step" : "steps"}
          {changes > 0 && ` · ${changes} ${changes === 1 ? "change" : "changes"}`}
        </span>
        <span className={styles.runDetails}>
          Details <ChevronDown size={13} />
        </span>
      </button>
      {open && (
        <ol className={styles.steps}>
          {steps.map((p) => (
            <Step key={p.toolCallId} part={p} />
          ))}
        </ol>
      )}
    </div>
  );
}

function Step({ part }: { part: ToolishPart }) {
  const [raw, setRaw] = useState(false);
  const name = toolName(part);
  const action = actionOf(part);
  const output = "output" in part ? part.output : undefined;
  const loading = part.state === "input-streaming" || part.state === "input-available";
  const error = part.state === "output-error";
  const title = action ? cap(sentence(action.tool)) : stepTitle(name, !loading);
  const summary = action
    ? actionResultSummary(action)
    : error
      ? "Couldn’t complete this lookup"
      : loading
        ? ""
        : resultSummary(output);
  return (
    <li className={styles.step}>
      <span className={styles.stepIcon} data-tone={error || action?.status === "declined" ? "off" : action ? "change" : "read"}>
        {loading ? (
          <Loader2 size={12} className={styles.spin} />
        ) : error ? (
          <X size={12} />
        ) : action?.status === "declined" ? (
          <CircleSlash size={12} />
        ) : action ? (
          <PencilLine size={12} />
        ) : (
          <Check size={12} />
        )}
      </span>
      <div className={styles.stepBody}>
        <p>
          {title}
          {summary && <span> · {summary}</span>}
        </p>
        {(output != null || action?.result != null) && (
          <button type="button" className={styles.stepRaw} onClick={() => setRaw(!raw)}>
            {raw ? "Hide data" : "Data"}
          </button>
        )}
        {raw && (
          <pre className={styles.output}>
            {JSON.stringify(action ? action.result ?? action.input : output, null, 2)}
          </pre>
        )}
      </div>
    </li>
  );
}

const SENDS = new Set(["send_email", "schedule_newsletter"]);
// IDs the person approving knows by another name.
const FIELD: Record<string, string> = {
  listId: "List",
  smtpConfigId: "Sender",
  templateId: "Template",
  contactId: "Contact",
  newsletterId: "Newsletter",
  campaignId: "Campaign",
  domainId: "Domain",
  formId: "Form",
  automationId: "Automation",
};

/** Readable value for an approval card: lists are counted and previewed, IDs shortened. */
function Value({ name, value }: { name: string; value: unknown }) {
  const [all, setAll] = useState(false);
  if (Array.isArray(value)) {
    const label = (v: unknown) => {
      if (v && typeof v === "object") {
        const o = v as Record<string, unknown>;
        const who = [o.firstName, o.lastName].filter(Boolean).join(" ");
        return String(o.email || o.name || o.label || who || JSON.stringify(o));
      }
      return String(v);
    };
    const shown = all ? value : value.slice(0, 3);
    return (
      <span>
        <strong>{value.length}</strong> {sentence(name)}
        <span className={styles.valueList}>
          {shown.map((v, i) => (
            <span key={i}>{label(v)}</span>
          ))}
          {value.length > 3 && (
            <button type="button" onClick={() => setAll(!all)}>
              {all ? "Show less" : `+${value.length - 3} more`}
            </button>
          )}
        </span>
      </span>
    );
  }
  if (typeof value === "boolean") return <span>{value ? "Yes" : "No"}</span>;
  if (value && typeof value === "object")
    return <span className={styles.mono}>{JSON.stringify(value)}</span>;
  const text = String(value ?? "");
  if (/id$/i.test(name) && /^[0-9a-f-]{20,}$/i.test(text))
    return <span className={styles.mono} title={text}>…{text.slice(-8)}</span>;
  if (text.length > 280 || /<\w+[^>]*>/.test(text))
    return <span className={styles.muted}>{text.length.toLocaleString()} characters of content</span>;
  return <span>{text}</span>;
}

/** A change the agent wants to make, waiting for a person. */
export function ApprovalCard({
  action,
  busy,
  onAction,
}: {
  action: ActionOutput;
  busy: boolean;
  onAction: (id: string, approved: boolean) => void;
}) {
  const { why, ...input } = action.input as Record<string, unknown>;
  const sends = SENDS.has(action.tool);
  const pending = action.status === "pending";
  return (
    <section className={styles.approval} aria-label={`Review ${sentence(action.tool)}`}>
      <header className={styles.approvalHead}>
        <ShieldCheck size={16} />
        <div>
          <strong>{pending ? "Needs your approval" : "Waiting on the outcome"}</strong>
          <span>Xem · your workspace</span>
        </div>
      </header>
      <div className={styles.approvalAction}>
        <span className={styles.approvalIcon}>{sends ? <Mail size={15} /> : <PencilLine size={15} />}</span>
        <strong>{cap(sentence(action.tool))}</strong>
        <span className={styles.badge} data-tone={sends ? "send" : "write"}>
          {sends ? "Sends email" : "Changes data"}
        </span>
      </div>
      <dl className={styles.approvalArgs}>
        {Object.entries(input).map(([key, value]) => (
          <div key={key}>
            <dt>{FIELD[key] || cap(sentence(key))}</dt>
            <dd>
              <Value name={key} value={value} />
            </dd>
          </div>
        ))}
      </dl>
      {typeof why === "string" && why && (
        <p className={styles.why}>
          <strong>Why:</strong> {why}
        </p>
      )}
      <footer className={styles.approvalFoot}>
        <span>
          {pending
            ? "Runs only with these details. Xem continues once you decide."
            : actionResultSummary(action)}
        </span>
        {pending && (
          <>
            <button type="button" disabled={busy} onClick={() => onAction(action.actionId, false)}>
              <X size={14} /> Deny
            </button>
            <button
              type="button"
              className={styles.approve}
              disabled={busy}
              onClick={() => onAction(action.actionId, true)}
            >
              <Check size={14} /> Approve
            </button>
          </>
        )}
      </footer>
    </section>
  );
}

export type AskOutput = {
  kind: "xem-ask";
  question?: string;
  options?: string[];
  fields?: {
    name: string;
    label: string;
    type?: "text" | "email" | "url" | "number" | "date" | "textarea" | "select";
    placeholder?: string;
    choices?: string[];
    required?: boolean;
  }[];
  submitLabel?: string;
};
export const askOf = (part: Part) => {
  const output = "output" in part ? part.output : undefined;
  return (output as AskOutput)?.kind === "xem-ask" ? (output as AskOutput) : undefined;
};

/** Buttons and a short form the agent asked for. Answering sends a message. */
export function AskCard({
  ask,
  active,
  onReply,
}: {
  ask: AskOutput;
  /** Only the latest question takes answers; older ones stay as a record. */
  active: boolean;
  onReply: (text: string) => void;
}) {
  const [values, setValues] = useState<Record<string, string>>({});
  const fields = ask.fields || [];
  return (
    <div className={styles.ask} data-active={active || undefined}>
      {ask.question && <p className={styles.askQuestion}>{ask.question}</p>}
      {fields.length > 0 && (
        <form
          className={styles.askForm}
          onSubmit={(e) => {
            e.preventDefault();
            onReply(
              fields
                .filter((f) => values[f.name]?.trim())
                .map((f) => `${f.label}: ${values[f.name].trim()}`)
                .join("\n"),
            );
          }}
        >
          {fields.map((f) => {
            const common = {
              id: `ask-${f.name}`,
              name: f.name,
              required: f.required,
              disabled: !active,
              placeholder: f.placeholder,
              value: values[f.name] || "",
              onChange: (e: { target: { value: string } }) =>
                setValues((v) => ({ ...v, [f.name]: e.target.value })),
            };
            return (
              <label key={f.name} htmlFor={common.id}>
                <span>{f.label}</span>
                {f.type === "textarea" ? (
                  <textarea rows={3} {...common} />
                ) : f.type === "select" ? (
                  <select {...common}>
                    <option value="">Choose…</option>
                    {f.choices?.map((c) => (
                      <option key={c}>{c}</option>
                    ))}
                  </select>
                ) : (
                  <input type={f.type || "text"} {...common} />
                )}
              </label>
            );
          })}
          <button type="submit" className={styles.askSubmit} disabled={!active}>
            {ask.submitLabel || "Continue"}
          </button>
        </form>
      )}
      {!!ask.options?.length && (
        <div className={styles.askOptions}>
          {ask.options.map((o) => (
            <button key={o} type="button" disabled={!active} onClick={() => onReply(o)}>
              {o}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
