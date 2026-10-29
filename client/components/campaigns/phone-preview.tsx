"use client";

import { useEffect, useMemo, useState, type SyntheticEvent } from "react";
import { useQuery } from "@tanstack/react-query";
import { ChevronLeft } from "lucide-react";
import { useApi } from "@/hooks/use-api";
import styles from "./phone-preview.module.css";

/** First readable text of an email, as a mail app shows it under the subject. */
export function previewText(html: string, max = 90) {
  // Keep block boundaries as spaces, the way a mail app's preview text reads.
  const spaced = html.replace(/<\/(p|h[1-6]|div|td|li|tr)>|<br\s*\/?>/gi, "$& ");
  const doc = new DOMParser().parseFromString(spaced, "text/html");
  doc.querySelectorAll("style,script,head,title").forEach((n) => n.remove());
  const text = (doc.body?.textContent ?? "").replace(/\s+/g, " ").trim();
  return text.length > max ? `${text.slice(0, max - 1)}…` : text;
}

const PHONE_WIDTH = 375;
const SCREEN_WIDTH = 264;

/** Mail apps zoom wide (e.g. 600px) emails out to fit the screen; measure and do the same. */
function measure(event: SyntheticEvent<HTMLIFrameElement>) {
  const doc = event.currentTarget.contentDocument;
  return Math.max(PHONE_WIDTH, doc?.documentElement.scrollWidth ?? PHONE_WIDTH);
}

const FILLER = [
  ["Calendar", "Reminder: team sync at 11:00", "Tomorrow, 11:00 to 11:30"],
  ["Julia", "Dinner on Saturday?", "We booked the place by the river"],
  ["Order updates", "Your order has shipped", "It should arrive by Thursday"],
];

export function PhonePreview({
  templateId,
  subject,
  sender,
}: {
  templateId?: string;
  subject?: string;
  sender?: string;
}) {
  const { apiFetch, session } = useApi();
  const [open, setOpen] = useState(false);
  const [contentWidth, setContentWidth] = useState(PHONE_WIDTH);
  const html = useQuery<string>({
    queryKey: ["template-preview-html", templateId],
    enabled: !!templateId && !!session?.accessToken,
    staleTime: 60000,
    queryFn: async ({ signal }) => {
      const res = await apiFetch(`marketing/templates/${templateId}/preview`, {
        signal,
      });
      if (!res.ok) throw new Error("Unable to load the email preview");
      return (await res.json()).htmlBody ?? "";
    },
  });
  // Re-measure from phone width whenever the email changes.
  useEffect(() => setContentWidth(PHONE_WIDTH), [html.data, open]);
  const snippet = useMemo(
    () => (html.data ? previewText(html.data) : ""),
    [html.data],
  );
  const from = sender?.split("@")[0] || "You";
  const policy = `<meta http-equiv="Content-Security-Policy" content="default-src 'none'; img-src https: data:; style-src 'unsafe-inline'; font-src 'none'; base-uri 'none'; form-action 'none'"><meta name="viewport" content="width=device-width">`;

  return (
    <div className={styles.wrap}>
      <div className={styles.toggle} role="group" aria-label="Phone preview view">
        <button type="button" aria-pressed={!open} onClick={() => setOpen(false)}>
          Inbox
        </button>
        <button
          type="button"
          aria-pressed={open}
          disabled={!html.data}
          onClick={() => setOpen(true)}
        >
          Opened
        </button>
      </div>
      <div className={styles.phone} aria-label="Phone preview">
        <div className={styles.notch} aria-hidden="true" />
        <div className={styles.status} aria-hidden="true">
          <span>9:41</span>
        </div>
        {open && html.data ? (
          <div className={styles.screen}>
            <button
              type="button"
              className={styles.back}
              onClick={() => setOpen(false)}
            >
              <ChevronLeft size={16} /> Inbox
            </button>
            <p className={styles.openSubject}>{subject || "Your subject line"}</p>
            <p className={styles.openFrom}>
              <span className={styles.avatar}>{from[0]?.toUpperCase()}</span>
              {from}
            </p>
            <div className={styles.frameBox}>
              <iframe
                title="Email on a phone"
                // Same origin only so the width can be measured; scripts stay disabled.
                sandbox="allow-same-origin"
                onLoad={(e) => setContentWidth(measure(e))}
                style={{
                  width: contentWidth,
                  height: `${(100 * contentWidth) / SCREEN_WIDTH}%`,
                  transform: `scale(${SCREEN_WIDTH / contentWidth})`,
                }}
                referrerPolicy="no-referrer"
                srcDoc={policy + html.data}
                className={styles.frame}
              />
            </div>
          </div>
        ) : (
          <ul className={styles.inbox}>
            <li className={styles.inboxTitle}>Inbox</li>
            <li>
              <button
                type="button"
                className={styles.row}
                disabled={!html.data}
                onClick={() => setOpen(true)}
              >
                <span className={styles.avatar}>{from[0]?.toUpperCase()}</span>
                <span className={styles.rowText}>
                  <span className={styles.rowTop}>
                    <strong>{from}</strong>
                    <time>9:41</time>
                  </span>
                  <strong className={styles.rowSubject}>
                    {subject || "Your subject line goes here"}
                  </strong>
                  <span className={styles.rowSnippet}>
                    {html.isPending && templateId
                      ? "Loading…"
                      : snippet || "Choose a template to see the preview text"}
                  </span>
                </span>
              </button>
            </li>
            {FILLER.map(([who, subj, text]) => (
              <li key={who} className={styles.filler} aria-hidden="true">
                <span className={styles.avatar}>{who[0]}</span>
                <span className={styles.rowText}>
                  <span className={styles.rowTop}>
                    <span>{who}</span>
                    <time>Yesterday</time>
                  </span>
                  <span className={styles.rowSubject}>{subj}</span>
                  <span className={styles.rowSnippet}>{text}</span>
                </span>
              </li>
            ))}
          </ul>
        )}
      </div>
      {html.isError && (
        <p className={styles.note} role="alert">
          {html.error.message}
        </p>
      )}
      <p className={styles.note}>
        Your subject and preview text, as most readers first see them.
      </p>
    </div>
  );
}
