"use client";

import { useEffect, useRef, useState } from "react";
import { Paperclip, Send, X } from "lucide-react";
import type { Editor as TiptapEditor } from "@tiptap/core";
import { Maily } from "@maily-to/render";
import { RichMessageEditor } from "@/components/rich-message-editor";
import { EmailWriter } from "@/components/ai/email-writer";
import { Button } from "@/components/ui/button";
import { useConfirmSheet } from "@/components/ui/confirm-sheet";
import {
  MarketingRequestError,
  useMarketing,
  useMarketingQuery,
} from "@/lib/marketing/api";
import {
  ComposeValue,
  MailSender,
  OutgoingAttachment,
  readMailAttachments,
  resolveMailSender,
} from "@/lib/connected-mail";
import { workspaceClassName } from "@/lib/workspace-styles";
import { Modal, Field } from "./shared";
import { toast } from "sonner";

type Submission = {
  requestId: string;
  to: string;
  subject: string;
  html: string;
  data: Record<string, string>;
  smtpConfigId: string;
  inReplyTo?: string;
  attachments: Omit<OutgoingAttachment, "size">[];
};

export function MailCompose({
  value,
  close,
}: {
  value: ComposeValue;
  close: () => void;
}) {
  const { request } = useMarketing();
  const confirm = useConfirmSheet();
  const senders = useMarketingQuery<MailSender[]>("mail-connections/senders");
  const [chosenSender, setChosenSender] = useState(value.smtpConfigId ?? "");
  const selectedSender =
    value.requireExplicitSender && !chosenSender
      ? undefined
      : resolveMailSender(senders.data, chosenSender);
  const [to, setTo] = useState(value.to);
  const [subject, setSubject] = useState(value.subject);
  const [body, setBody] = useState("");
  const [hasContent, setHasContent] = useState(false);
  const [editor, setEditor] = useState<TiptapEditor>();
  const [attachments, setAttachments] = useState<OutgoingAttachment[]>([]);
  const [readingFiles, setReadingFiles] = useState(false);
  const preparingFiles = useRef(false);
  const [busy, setBusy] = useState(false);
  const submitting = useRef(false);
  const [error, setError] = useState("");
  const [submission, setSubmission] = useState<Submission | null>(null);
  const locked = busy || !!submission;
  const dirty =
    hasContent ||
    attachments.length > 0 ||
    to !== value.to ||
    subject !== value.subject;

  useEffect(() => {
    if (!dirty && !submission) return;
    const warn = (event: BeforeUnloadEvent) => {
      event.preventDefault();
    };
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  }, [dirty, submission]);

  async function requestClose() {
    if (submitting.current || preparingFiles.current) return;
    if (
      (dirty || submission) &&
      !(await confirm({
        title: submission ? "Close this message?" : "Discard this draft?",
        description: submission
          ? "This message may already be in your outbox. Check its status there before composing it again. Retrying here uses the same request and cannot create a second message."
          : "Your message and attachments have not been saved. Keep editing or discard them.",
        confirmLabel: submission ? "Close message" : "Discard draft",
        cancelLabel: "Keep editing",
      }))
    )
      return;
    close();
  }

  return (
    <Modal
      open
      onOpenChange={(open) => {
        if (!open) void requestClose();
      }}
      wide
      title={value.inReplyTo ? "Reply to conversation" : "New message"}
      description="Choose a sender, write your message, and review before sending."
    >
      <form
        className={workspaceClassName("product-form mail-compose-form")}
        onSubmit={async (event) => {
          event.preventDefault();
          if (submitting.current || preparingFiles.current) return;
          submitting.current = true;
          setBusy(true);
          setError("");
          let attempted = false;
          try {
            let payload = submission;
            if (!payload) {
              if (!selectedSender)
                throw new Error("Choose an available sender before sending.");
              if (!editor || editor.isEmpty)
                throw new Error("Write a message before sending.");
              payload = {
                requestId: crypto.randomUUID(),
                to: to.trim(),
                subject,
                html: await new Maily(editor.getJSON()).render(),
                data: {},
                smtpConfigId: selectedSender.id,
                inReplyTo: value.inReplyTo,
                attachments: attachments.map(({ size, ...file }) => file),
              };
            }
            setSubmission(payload);
            attempted = true;
            const receipt = await request<{ id: string }>(
              "emails",
              "POST",
              payload,
            );
            if (!receipt?.id)
              throw new Error(
                "The server did not return a message receipt. Check the outbox before retrying.",
              );
            toast.success("Message saved to outbox", {
              description:
                "Xem will send it through your selected provider. Follow delivery in Outbox.",
            });
            close();
          } catch (error) {
            if (
              !attempted ||
              (!submission &&
                error instanceof MarketingRequestError &&
                error.status >= 400 &&
                error.status < 500)
            ) {
              setSubmission(null);
              setError((error as Error).message);
            } else {
              setError(
                "We couldn’t confirm the outbox receipt. Your message is preserved below. Retry safely with the same request, or check Outbox before sending another copy.",
              );
            }
          } finally {
            submitting.current = false;
            setBusy(false);
          }
        }}
      >
        <Field label="From">
          <select
            required
            disabled={locked}
            value={selectedSender?.id ?? ""}
            onChange={(event) => setChosenSender(event.target.value)}
          >
            <option value="" disabled>
              {senders.isLoading ? "Loading senders…" : "Choose a sender"}
            </option>
            {senders.data?.map((sender) => (
              <option key={sender.id} value={sender.id}>
                {sender.fromEmail} ·{" "}
                {sender.provider === "GOOGLE_OAUTH"
                  ? "Google"
                  : sender.provider === "CLOUDFLARE"
                    ? "Cloudflare"
                    : sender.provider === "MANAGED"
                      ? "Xem managed sending"
                      : sender.provider}
              </option>
            ))}
          </select>
        </Field>
        {senders.error && (
          <p role="alert" className="text-sm text-destructive">
            {senders.error.message}{" "}
            <button
              type="button"
              className="underline"
              onClick={() => senders.refetch()}
            >
              Try again
            </button>
          </p>
        )}
        {!senders.isLoading && !senders.error && !selectedSender && (
          <p role="alert" className="text-sm text-destructive">
            {chosenSender
              ? "The original sender is unavailable. Reconnect it in settings or explicitly choose another sender."
              : "Connect a sender in SMTP settings before sending."}
          </p>
        )}
        {value.requireExplicitSender && !chosenSender && (
          <p className="text-xs text-muted-foreground">
            This receive-only mailbox has no linked sender. Choose the address
            you want to reply from; Xem will not substitute a default identity.
          </p>
        )}
        {selectedSender?.provider === "CLOUDFLARE" && (
          <p className="rounded-lg border bg-muted/40 p-3 text-sm text-muted-foreground">
            Cloudflare sends transactional messages such as receipts and account
            notifications. Choose another sender for marketing or newsletters.
          </p>
        )}
        <Field label="To">
          <input
            required
            type="email"
            disabled={locked}
            value={to}
            onChange={(event) => setTo(event.target.value)}
            autoComplete="off"
          />
        </Field>
        <Field label="Subject">
          <input
            required
            maxLength={200}
            disabled={locked}
            value={subject}
            onChange={(event) => setSubject(event.target.value)}
          />
        </Field>
        {!locked && (
          <div className="flex justify-end">
            <EmailWriter
              subject={subject}
              body={body}
              onApply={(draft) => {
                if (!editor)
                  throw new Error("The message editor is still loading.");
                const content: any[] = draft.body
                  .split(/\n\n+/)
                  .map((paragraph) => ({
                    type: "paragraph",
                    content: paragraph
                      .split("\n")
                      .flatMap((line, index) => [
                        ...(index ? [{ type: "hardBreak" }] : []),
                        ...(line ? [{ type: "text", text: line }] : []),
                      ]),
                  }));
                for (const image of draft.images ?? []) {
                  if (!/^https:\/\//i.test(image.url))
                    throw new Error("Images must use an HTTPS URL.");
                  content.push({
                    type: "image",
                    attrs: { src: image.url, alt: image.alt },
                  });
                }
                editor.commands.setContent({ type: "doc", content });
                setSubject(draft.subject);
                setBody(draft.body);
              }}
            />
          </div>
        )}
        <div
          role="group"
          aria-labelledby="compose-message-label"
          className="space-y-2"
        >
          <span
            id="compose-message-label"
            className="text-xs text-muted-foreground"
          >
            Message
          </span>
          <RichMessageEditor
            editable={!locked}
            onReady={(editor) => {
              setEditor(editor);
              setHasContent(!editor.isEmpty);
            }}
            onChange={(editor) => {
              setBody(editor.getText());
              setHasContent(!editor.isEmpty);
            }}
          />
        </div>
        <div className="space-y-3">
          <label className="flex flex-wrap items-center gap-2 text-sm font-medium">
            <Paperclip className="size-4" /> Attach files
            <input
              aria-label="Attach files"
              type="file"
              multiple
              disabled={locked || readingFiles}
              className="min-w-0 w-full max-w-full text-xs"
              onChange={async (event) => {
                if (preparingFiles.current || submitting.current || submission)
                  return;
                const files = Array.from(event.target.files ?? []);
                event.target.value = "";
                if (!files.length) return;
                preparingFiles.current = true;
                setReadingFiles(true);
                setError("");
                try {
                  const added = await readMailAttachments(files, attachments);
                  setAttachments((previous) => [...previous, ...added]);
                } catch (error) {
                  setError((error as Error).message);
                } finally {
                  preparingFiles.current = false;
                  setReadingFiles(false);
                }
              }}
            />
          </label>
          <p className="text-xs text-muted-foreground">
            Up to 10 files, 3 MiB combined. Files stay with this message.
          </p>
          {readingFiles && (
            <p role="status" className="text-xs">
              Preparing attachments…
            </p>
          )}
          {attachments.length > 0 && (
            <ul className="space-y-2">
              {attachments.map((file, index) => (
                <li
                  key={`${file.filename}-${index}`}
                  className="flex min-w-0 items-center gap-2 rounded-lg border px-3 py-2 text-sm"
                >
                  <Paperclip className="size-4 shrink-0" />
                  <span className="min-w-0 flex-1 break-all">
                    {file.filename}
                  </span>
                  <span className="shrink-0 text-xs text-muted-foreground">
                    {Math.max(1, Math.ceil(file.size / 1024))} KiB
                  </span>
                  <Button
                    variant="ghost"
                    size="icon"
                    type="button"
                    disabled={locked || readingFiles}
                    aria-label={`Remove ${file.filename}`}
                    onClick={() =>
                      setAttachments((files) =>
                        files.filter((_, i) => i !== index),
                      )
                    }
                  >
                    <X className="size-4" />
                  </Button>
                </li>
              ))}
            </ul>
          )}
        </div>
        {error && (
          <div role="alert" className={workspaceClassName("product-error")}>
            {error}
          </div>
        )}
        {submission && !busy && (
          <p className="break-all text-xs text-muted-foreground">
            Request receipt: {submission.requestId}
          </p>
        )}
        <div className={workspaceClassName("modal-actions")}>
          <Button
            type="button"
            variant="outline"
            disabled={busy || readingFiles}
            onClick={() => void requestClose()}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            disabled={
              busy ||
              readingFiles ||
              (!submission && (!editor || !selectedSender || !hasContent))
            }
          >
            <Send />
            {busy
              ? "Saving to outbox…"
              : submission
                ? "Retry safely"
                : "Send message"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
