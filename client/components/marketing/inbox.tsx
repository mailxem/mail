"use client";
import { workspaceClassName } from "@/lib/workspace-styles";
import { useEffect, useMemo, useRef, useState } from "react";
import { useInfiniteQuery } from "@tanstack/react-query";
import {
  Mail,
  Inbox,
  Send,
  Search,
  PencilLine,
  Reply,
  ArrowLeft,
  ArrowRight,
  Paperclip,
  RefreshCw,
  X,
  ChevronDown,
  Pin,
  Archive,
  FileText,
  Star,
  MailCheck,
} from "lucide-react";
import { useMarketing, useMarketingQuery } from "@/lib/marketing/api";
import { IMAPEmail, IMAPEmailResponse } from "@/types/imap";
import { Button } from "@/components/ui/button";
import { QueryState, Empty } from "./shared";
import { MailCompose } from "./mail-compose";
import { MailSummary } from "./mail-summary";
import { MailboxSelect } from "./mailbox-select";
import {
  ComposeValue,
  OutgoingAttachment,
  replySubject,
} from "@/lib/connected-mail";
import { toast } from "sonner";
export function senderName(value: string) {
  const source = value?.trim();
  if (!source) return "Unknown sender";
  const addressOnly = source.match(/^<\s*([^<>]+?)\s*>$/);
  if (addressOnly) return addressOnly[1];
  const address = source.match(/<[^<>]*>\s*$/);
  if (!address) return source.replace(/^"|"$/g, "").trim() || source;
  return (
    source.slice(0, address.index).trim().replace(/^"|"$/g, "").trim() ||
    address[0].slice(1, -1).trim()
  );
}
function text(body: string) {
  return (
    body
      ?.replace(/<style[^>]*>[\s\S]*?<\/style>/gi, "")
      .replace(/<[^>]*>/g, " ")
      .replace(/\s+/g, " ")
      .slice(0, 120) || ""
  );
}
function hasRemoteImages(body: string) {
  return /<(?:img|source)\b[^>]+(?:src|srcset)\s*=\s*["']\s*https?:|background(?:-image)?\s*:\s*url\(\s*["']?\s*https?:/i.test(
    body,
  );
}
function attachmentSize(data: string) {
  const bytes = Math.max(
    0,
    Math.floor((data.length * 3) / 4) -
      (data.endsWith("==") ? 2 : data.endsWith("=") ? 1 : 0),
  );
  return bytes < 1024
    ? `${bytes} B`
    : `${(bytes / 1024).toFixed(bytes < 10240 ? 1 : 0)} KB`;
}
function byteSize(bytes: number) {
  return bytes < 1024
    ? `${bytes} B`
    : bytes < 1024 * 1024
      ? `${(bytes / 1024).toFixed(bytes < 10240 ? 1 : 0)} KB`
      : `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}
type MailMessage = IMAPEmail & { status?: string; error?: string };
type Mailbox = {
  id: string;
  username: string;
  host: string;
  smtpConfigId?: string;
  provider?: string;
};
export const messageKey = (m: MailMessage) =>
  m.id ||
  (m.uid != null && m.uidValidity != null
    ? `uid:${m.uidValidity}:${m.uid}`
    : m.messageId);
export const sortMailMessages = (messages: MailMessage[]) =>
  [...messages].sort(
    (a, b) =>
      Number(b.flags?.includes("\\Flagged")) -
      Number(a.flags?.includes("\\Flagged")),
  );
export function dedupeMailMessages(messages: MailMessage[]) {
  const seen = new Set<string>();
  return messages.filter((message, index) => {
    const key =
      messageKey(message) || `${message.date}:${message.subject}:${index}`;
    if (seen.has(key)) return false;
    seen.add(key);
    return true;
  });
}
export function InboxPage() {
  return <MailboxPage mode="inbox" />;
}
export function OutboxPage() {
  return <MailboxPage mode="outbox" />;
}
function MailboxPage({ mode }: { mode: "inbox" | "outbox" }) {
  const outbox = mode === "outbox";
  const title = outbox ? "Outbox" : "Inbox";
  const { request, scope, ready } = useMarketing();
  const mailboxes = useMarketingQuery<Mailbox[]>(
    "mail-connections/mailboxes",
    !outbox,
  );
  const [chosenMailbox, setChosenMailbox] = useState("");
  const mailbox =
    mailboxes.data?.find((m) => m.id === chosenMailbox) ?? mailboxes.data?.[0];
  const configId = mailbox?.id ?? "";
  const folders = useMarketingQuery<
    {
      Name: string;
      DisplayName?: string;
      Total?: number;
      Attributes?: string[];
    }[]
  >(
    `imap/folders?${new URLSearchParams({ config_id: configId })}`,
    !outbox && !!configId,
  );
  const [folder, setFolder] = useState(outbox ? "SENT" : "INBOX");
  const [search, setSearch] = useState("");
  const [query, setQuery] = useState("");
  const [selected, setSelected] = useState<MailMessage | null>(null);
  const [detailState, setDetailState] = useState<
    "idle" | "loading" | "ready" | "error"
  >("idle");
  const [compose, setCompose] = useState<ComposeValue | null>(null);
  const [flagBusy, setFlagBusy] = useState(false);
  const [attachmentBusy, setAttachmentBusy] = useState("");
  const [newMailAvailable, setNewMailAvailable] = useState(false);
  const [epochVersion, setEpochVersion] = useState(0);
  const epochRef = useRef<number | undefined>(undefined);
  const historyRef = useRef<string | undefined>(undefined);
  const historyBaselineIdentityRef = useRef("");
  const pollInFlight = useRef(false);
  const requestRef = useRef(request);
  const selectedRef = useRef(selected);
  const attachmentRequestRef = useRef(0);
  const listScrollRef = useRef<HTMLDivElement>(null);
  const loadSentinelRef = useRef<HTMLDivElement>(null);
  const generation = `${scope}:${configId}:${folder}:${query}`;
  const generationRef = useRef(generation);
  const detailRequestRef = useRef(0);
  generationRef.current = generation;
  requestRef.current = request;
  selectedRef.current = selected;
  useEffect(() => {
    setCompose(null);
  }, [scope]);
  useEffect(() => {
    setSelected(null);
    setDetailState("idle");
    detailRequestRef.current += 1;
    attachmentRequestRef.current += 1;
    setImages(false);
    setFlagBusy(false);
    setAttachmentBusy("");
    setNewMailAvailable(false);
    epochRef.current = undefined;
    historyRef.current = undefined;
    historyBaselineIdentityRef.current = "";
  }, [scope, configId, folder, query]);
  const [images, setImages] = useState(false);
  const emails = useInfiniteQuery({
    queryKey: ["marketing", scope, mode, configId, folder, query, epochVersion],
    enabled: ready && (outbox || !!configId),
    initialPageParam: outbox ? 1 : (undefined as number | string | undefined),
    queryFn: async ({ pageParam }): Promise<IMAPEmailResponse> => {
      if (!outbox) {
        const queryGeneration = generation;
        const google = mailbox?.provider === "GOOGLE_OAUTH";
        const params = new URLSearchParams({
          folder,
          limit: "20",
          q: query,
          config_id: configId,
        });
        if (pageParam)
          params.set(google ? "page_token" : "before_uid", String(pageParam));
        const response = await request<IMAPEmailResponse>(
          `imap/emails?${params}`,
        );
        if (
          !google &&
          generationRef.current === queryGeneration &&
          epochRef.current === undefined
        )
          epochRef.current = response.uidValidity;
        else if (
          !google &&
          generationRef.current === queryGeneration &&
          pageParam &&
          response.uidValidity !== undefined &&
          response.uidValidity !== epochRef.current
        ) {
          epochRef.current = response.uidValidity;
          queueMicrotask(() => {
            if (generationRef.current === queryGeneration) {
              setSelected(null);
              setEpochVersion((value) => value + 1);
            }
          });
          throw new Error(
            "The mailbox changed while loading. Refreshing the message list.",
          );
        }
        return response;
      }
      const result = await request<{
        data: (Omit<MailMessage, "attachments"> & {
          id: string;
          createdAt: string;
          sentAt?: string;
          attachments?: Omit<OutgoingAttachment, "size">[];
        })[];
        total: number;
        page: number;
      }>(
        `emails?${new URLSearchParams({ page: String(pageParam), sort: "created_at", order: "desc", status: folder, limit: "20" })}`,
      );
      return {
        emails: (result.data || []).map((email) => {
          let body = "";
          try {
            body = new TextDecoder().decode(
              Uint8Array.from(atob(email.body || ""), (c) => c.charCodeAt(0)),
            );
          } catch {
            body = email.body || "";
          }
          return {
            ...email,
            body,
            messageId: email.id,
            date:
              email.sentAt && !email.sentAt.startsWith("0001")
                ? email.sentAt
                : email.createdAt,
            flags: [],
            attachments: (email.attachments ?? []).map((file) => ({
              Filename: file.filename,
              Data: file.content,
            })),
          };
        }),
        total_emails: result.total,
        offset: (result.page - 1) * 20,
        limit: 20,
      };
    },
    getNextPageParam: (last, pages) => {
      if (outbox)
        return last.emails.length > 0 &&
          last.offset + last.limit < last.total_emails
          ? pages.length + 1
          : undefined;
      const cursor =
        mailbox?.provider === "GOOGLE_OAUTH"
          ? last.next_page_token
          : last.next_before_uid;
      if (
        cursor === undefined ||
        pages
          .slice(0, -1)
          .some((page) =>
            mailbox?.provider === "GOOGLE_OAUTH"
              ? page.next_page_token === cursor
              : page.next_before_uid === cursor,
          )
      )
        return undefined;
      return cursor;
    },
    retry: 1,
  });
  const allRows: MailMessage[] = useMemo(
    () => dedupeMailMessages(emails.data?.pages.flatMap((p) => p.emails) || []),
    [emails.data],
  );
  const firstPageHistory = emails.data?.pages[0]?.history_id;
  useEffect(() => {
    if (mailbox?.provider !== "GOOGLE_OAUTH" || !firstPageHistory) return;
    const identity = `${generation}:${epochVersion}:${firstPageHistory}`;
    if (historyBaselineIdentityRef.current === identity) return;
    historyBaselineIdentityRef.current = identity;
    historyRef.current = firstPageHistory;
  }, [epochVersion, firstPageHistory, generation, mailbox?.provider]);
  const rows =
    outbox && query
      ? allRows.filter((m) =>
          `${m.to} ${m.subject} ${text(m.body)}`
            .toLowerCase()
            .includes(query.toLowerCase()),
        )
      : allRows;

  useEffect(() => {
    if (outbox || !ready || !configId || query || !emails.isSuccess) return;
    let stopped = false;
    let timer: number | undefined;
    let failures = 0;
    const schedule = (delay: number) => {
      window.clearTimeout(timer);
      timer = window.setTimeout(() => void poll(), delay);
    };
    const poll = async () => {
      if (
        document.visibilityState !== "visible" ||
        !navigator.onLine ||
        stopped ||
        pollInFlight.current
      )
        return schedule(30_000);
      pollInFlight.current = true;
      try {
        const google = mailbox?.provider === "GOOGLE_OAUTH";
        if (google && !historyRef.current) return;
        const params = new URLSearchParams({
          folder,
          q: query,
          config_id: configId,
        });
        if (google) params.set("history_id", historyRef.current!);
        const head = await requestRef.current<{
          latest_uid: number;
          uidValidity?: number;
          history_id?: string;
          changed?: boolean;
          reset_required?: boolean;
        }>(`imap/head?${params}`);
        if (stopped) return;
        failures = 0;
        if (google) {
          if (head.history_id) historyRef.current = head.history_id;
          if (head.reset_required) {
            setSelected(null);
            setNewMailAvailable(false);
            setEpochVersion((value) => value + 1);
          } else if (head.changed) setNewMailAvailable(true);
          return;
        }
        if (
          epochRef.current !== undefined &&
          head.uidValidity !== undefined &&
          head.uidValidity !== epochRef.current
        ) {
          epochRef.current = head.uidValidity;
          setSelected(null);
          setNewMailAvailable(false);
          setEpochVersion((value) => value + 1);
          return;
        }
        const highest = Math.max(
          0,
          ...allRows.map((message) => message.uid ?? 0),
        );
        setNewMailAvailable(head.latest_uid > highest);
      } catch {
        failures += 1;
        /* Poll failures stay quiet; the normal inbox query owns visible errors. */
      } finally {
        pollInFlight.current = false;
        if (!stopped)
          schedule(Math.min(300_000, 30_000 * 2 ** Math.min(failures, 4)));
      }
    };
    void poll();
    const resume = () => void poll();
    document.addEventListener("visibilitychange", resume);
    window.addEventListener("online", resume);
    return () => {
      stopped = true;
      window.clearTimeout(timer);
      document.removeEventListener("visibilitychange", resume);
      window.removeEventListener("online", resume);
    };
  }, [
    outbox,
    ready,
    configId,
    folder,
    query,
    allRows,
    emails.isSuccess,
    mailbox?.provider,
  ]);

  useEffect(() => {
    const sentinel = loadSentinelRef.current;
    const root = listScrollRef.current;
    if (
      typeof IntersectionObserver === "undefined" ||
      !sentinel ||
      !root ||
      !emails.hasNextPage ||
      emails.isFetching ||
      emails.isFetchNextPageError
    )
      return;
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) void emails.fetchNextPage();
      },
      { root, rootMargin: "160px" },
    );
    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [
    emails.hasNextPage,
    emails.isFetching,
    emails.isFetchNextPageError,
    emails.fetchNextPage,
  ]);
  const total = emails.data?.pages[0]?.total_emails || 0;
  const displayedFolders: {
    Name: string;
    DisplayName?: string;
    Total?: number;
    Attributes?: string[];
  }[] = outbox
    ? [
        "SENT",
        "ACCEPTED",
        "PARTIAL",
        "PENDING",
        "SENDING",
        "FAILED",
        "DELIVERY_UNKNOWN",
        "SUPPRESSED",
        "BOUNCED",
        "OPENED",
        "CLICKED",
      ].map((Name) => ({ Name, Total: 0 }))
    : folders.data?.filter(
        (item) => !item.Attributes?.includes("\\Noselect"),
      ) || [{ Name: "INBOX", Total: total }];
  const displayedRows = sortMailMessages(rows);
  const requiresDetail =
    mailbox?.provider === "CLOUDFLARE" ||
    mailbox?.provider === "MANAGED" ||
    mailbox?.provider === "GOOGLE_OAUTH";
  const index = displayedRows.findIndex(
    (e) => messageKey(e) === (selected ? messageKey(selected) : undefined),
  );
  const current = selected;
  const folderDisplayName = displayedFolders.find(
    (item) => item.Name === folder,
  )?.DisplayName;
  useEffect(() => {
    if (
      selected &&
      !emails.isFetching &&
      emails.data &&
      !rows.some((row) => messageKey(row) === messageKey(selected))
    ) {
      setSelected(null);
    }
  }, [emails.data, emails.isFetching, rows, selected]);
  const loadMessageDetail = async (m: MailMessage) => {
    const google = mailbox?.provider === "GOOGLE_OAUTH";
    const numericDetail =
      mailbox?.provider === "CLOUDFLARE" || mailbox?.provider === "MANAGED";
    if (
      outbox ||
      (!google && !numericDetail) ||
      (google ? !m.providerMessageId : m.uid == null || m.uidValidity == null)
    ) {
      setDetailState("ready");
      return;
    }
    const requestID = ++detailRequestRef.current;
    const requestGeneration = generation;
    const key = messageKey(m);
    setDetailState("loading");
    try {
      const params = new URLSearchParams({ config_id: configId, folder });
      if (google) params.set("message_id", m.providerMessageId!);
      else {
        params.set("uid", String(m.uid));
        params.set("uid_validity", String(m.uidValidity));
      }
      const detail = await request<MailMessage>(`imap/message?${params}`);
      if (
        detailRequestRef.current !== requestID ||
        generationRef.current !== requestGeneration
      )
        return;
      if (
        (google && detail.providerMessageId !== m.providerMessageId) ||
        (!google &&
          (detail.uid !== m.uid || detail.uidValidity !== m.uidValidity))
      ) {
        setDetailState("error");
        return;
      }
      setSelected((previous) =>
        previous && messageKey(previous) === key ? detail : previous,
      );
      setDetailState("ready");
    } catch {
      if (
        detailRequestRef.current === requestID &&
        generationRef.current === requestGeneration
      )
        setDetailState("error");
    }
  };
  const choose = (m: MailMessage) => {
    detailRequestRef.current += 1;
    attachmentRequestRef.current += 1;
    setAttachmentBusy("");
    setSelected(m);
    setImages(false);
    void loadMessageDetail(m);
  };
  async function changeFlag(flag: string) {
    const google = mailbox?.provider === "GOOGLE_OAUTH";
    if (
      !current ||
      flagBusy ||
      (google
        ? !current.providerMessageId
        : !current.uid || !current.uidValidity)
    )
      return;
    const message = current;
    const requestGeneration = generation;
    const enabled = !message.flags?.includes(flag);
    setFlagBusy(true);
    try {
      await request(
        `imap/flags?${new URLSearchParams({ config_id: configId })}`,
        "PATCH",
        google
          ? {
              folder,
              providerMessageId: message.providerMessageId,
              flag,
              enabled,
            }
          : {
              folder,
              uid: message.uid,
              uidValidity: message.uidValidity,
              flag,
              enabled,
            },
      );
      if (generationRef.current !== requestGeneration) return;
      const flags = enabled
        ? [...(message.flags ?? []), flag]
        : (message.flags ?? []).filter((f) => f !== flag);
      setSelected((previous) =>
        previous && messageKey(previous) === messageKey(message)
          ? { ...previous, flags }
          : previous,
      );
      await emails.refetch();
    } catch (error) {
      if (generationRef.current === requestGeneration)
        toast.error((error as Error).message);
    } finally {
      if (generationRef.current === requestGeneration) setFlagBusy(false);
    }
  }
  async function downloadAttachment(
    message: MailMessage,
    attachment: MailMessage["attachments"][number],
  ) {
    const google = mailbox?.provider === "GOOGLE_OAUTH";
    if (
      !attachment.AttachmentID ||
      (google
        ? !message.providerMessageId
        : message.uid == null || message.uidValidity == null)
    )
      return;
    const requestGeneration = generation;
    const requestID = ++attachmentRequestRef.current;
    const attachmentKey = `${message.providerMessageId || `${message.uid}:${message.uidValidity}`}:${attachment.AttachmentID}`;
    setAttachmentBusy(attachmentKey);
    try {
      const result = await request<{ Data: string }>(
        `imap/attachment?${new URLSearchParams({
          config_id: configId,
          folder,
          ...(google
            ? { message_id: message.providerMessageId! }
            : {
                uid: String(message.uid),
                uid_validity: String(message.uidValidity),
              }),
          attachment_id: attachment.AttachmentID,
        })}`,
      );
      if (
        attachmentRequestRef.current !== requestID ||
        generationRef.current !== requestGeneration ||
        !selectedRef.current ||
        messageKey(selectedRef.current) !== messageKey(message) ||
        typeof result.Data !== "string"
      )
        return;
      const anchor = document.createElement("a");
      anchor.download = attachment.Filename;
      anchor.href = `data:${attachment.MIMEType || "application/octet-stream"};base64,${result.Data}`;
      anchor.click();
    } catch (error) {
      if (generationRef.current === requestGeneration)
        toast.error((error as Error).message);
    } finally {
      if (
        attachmentRequestRef.current === requestID &&
        generationRef.current === requestGeneration
      )
        setAttachmentBusy("");
    }
  }
  function reply(message: MailMessage): ComposeValue {
    const address = outbox ? message.to : message.reply_to || message.from;
    const rawId = message.messageId;
    return {
      to: address.match(/<([^>]+)>/)?.[1] || address,
      subject: replySubject(message.subject),
      smtpConfigId: mailbox?.smtpConfigId,
      requireExplicitSender: !mailbox?.smtpConfigId,
      ...(!outbox && rawId
        ? { inReplyTo: rawId.startsWith("<") ? rawId : `<${rawId}>` }
        : {}),
    };
  }
  return (
    <div
      className={workspaceClassName(
        `mail-workspace ${current ? "mail-open" : "mail-idle"}`,
      )}
    >
      <aside className={workspaceClassName("mail-folders")}>
        <div className={workspaceClassName("mail-workspace-title")}>
          <span className={workspaceClassName("mail-account-mark")}>
            <Mail size={17} />
          </span>
          <strong>{outbox ? "Outgoing mail" : "Team mailbox"}</strong>
          <small>
            {outbox
              ? "Delivery and message history."
              : "Your conversations, together."}
          </small>
        </div>
        <Button
          className={workspaceClassName("compose-button")}
          onClick={() =>
            setCompose({
              to: "",
              subject: "",
              smtpConfigId: mailbox?.smtpConfigId,
            })
          }
        >
          <PencilLine />
          Compose
        </Button>
        {!outbox && (
          <div className="pb-3 pt-3">
            <MailboxSelect
              value={configId}
              mailboxes={mailboxes.data ?? []}
              disabled={mailboxes.isLoading}
              onChange={(value) => {
                setChosenMailbox(value);
                setFolder("INBOX");
                setSelected(null);
              }}
            />
            {mailboxes.error && (
              <p role="alert" className="mt-2 text-xs text-destructive">
                {mailboxes.error.message}
              </p>
            )}
            {!mailboxes.isLoading && !configId && !mailboxes.error && (
              <p className="mt-2 text-xs text-muted-foreground">
                Connect a mailbox in settings to read your mail.
              </p>
            )}
            {folders.error && (
              <p role="alert" className="mt-2 text-xs text-destructive">
                {folders.error.message}
              </p>
            )}
          </div>
        )}
        <div className={workspaceClassName("mail-folder-list")}>
          {displayedFolders.map((f) => (
            <button
              key={f.Name}
              className={folder === f.Name ? "active" : ""}
              onClick={() => {
                setFolder(f.Name);
                setSelected(null);
              }}
            >
              {f.Name.toLowerCase().includes("sent") ? (
                <Send size={15} />
              ) : f.Name.toLowerCase().includes("draft") ? (
                <FileText size={15} />
              ) : f.Name.toLowerCase().includes("archive") ? (
                <Archive size={15} />
              ) : (
                <Inbox size={15} />
              )}
              <span>
                {f.Name === "INBOX"
                  ? "Inbox"
                  : f.Name === "DELIVERY_UNKNOWN"
                    ? "Needs review"
                    : f.DisplayName ||
                      f.Name.charAt(0) + f.Name.slice(1).toLowerCase()}
              </span>
              <small>{f.Total || ""}</small>
            </button>
          ))}
        </div>
        <div className={workspaceClassName("mail-folder-note")}>
          {outbox ? (
            <>
              Manage your senders in{" "}
              <a href="/settings/smtp">SMTP settings ↗</a>
            </>
          ) : (
            <>
              Connect your mailbox in{" "}
              <a href="/settings/imap">IMAP settings ↗</a>
            </>
          )}
        </div>
      </aside>
      <section
        className={workspaceClassName(
          `mail-list-pane ${current ? "mail-list-collapsed" : ""}`,
        )}
      >
        <div className={workspaceClassName("mail-list-heading")}>
          <div>
            <h1>
              {outbox
                ? "Outbox"
                : folderDisplayName || (folder === "INBOX" ? "Inbox" : folder)}
              <ChevronDown size={16} />
            </h1>
            <p>
              {emails.data?.pages[0]?.total_is_estimate ? "About " : ""}
              {total.toLocaleString()} messages
            </p>
          </div>
          <button
            className={workspaceClassName("icon-button")}
            aria-label={`Refresh ${title.toLowerCase()}`}
            onClick={() => emails.refetch()}
          >
            <RefreshCw />
          </button>
        </div>
        {!outbox && (
          <div className={workspaceClassName("mail-mobile-actions")}>
            <MailboxSelect
              value={configId}
              mailboxes={mailboxes.data ?? []}
              disabled={mailboxes.isLoading}
              onChange={(value) => {
                setChosenMailbox(value);
                setFolder("INBOX");
                setSelected(null);
              }}
            />
          </div>
        )}
        <select
          aria-label="Choose mail folder"
          className={workspaceClassName("mail-mobile-folder")}
          value={folder}
          onChange={(event) => {
            setFolder(event.target.value);
            setSelected(null);
          }}
        >
          {displayedFolders.map((item) => (
            <option key={item.Name} value={item.Name}>
              {item.DisplayName || item.Name}
            </option>
          ))}
        </select>
        <form
          className={workspaceClassName("mail-search")}
          onSubmit={(e) => {
            e.preventDefault();
            setQuery(search);
            setSelected(null);
          }}
        >
          <Search size={17} />
          <input
            aria-label={`Search ${title.toLowerCase()}`}
            placeholder={
              outbox ? "Search loaded messages…" : "Search messages…"
            }
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          <kbd>↵</kbd>
        </form>
        <div
          ref={listScrollRef}
          className={workspaceClassName("mail-list-scroll")}
        >
          {newMailAvailable && (
            <button
              className={workspaceClassName("mail-new-banner")}
              onClick={() => {
                setNewMailAvailable(false);
                setSelected(null);
                epochRef.current = undefined;
                setEpochVersion((value) => value + 1);
                listScrollRef.current?.scrollTo({ top: 0, behavior: "smooth" });
              }}
            >
              {mailbox?.provider === "GOOGLE_OAUTH"
                ? "Mailbox updated · Refresh"
                : "New messages available · Show"}
            </button>
          )}
          <QueryState
            loading={emails.isLoading}
            error={emails.data ? null : emails.error}
            retry={() => emails.refetch()}
          />
          {emails.isFetchNextPageError && (
            <p role="alert" className="px-3 py-2 text-xs text-muted-foreground">
              Older messages could not be loaded. Use the button below to retry.
            </p>
          )}
          {!emails.isLoading && !emails.error && !rows.length && (
            <Empty
              title="A little peace and quiet"
              description={
                outbox
                  ? "No messages in this delivery status. Choose another status or load more messages."
                  : "New messages will appear here. Try another folder or connect a mailbox."
              }
            />
          )}
          {rows.some((e) => e.flags?.includes("\\Flagged")) && (
            <div className={workspaceClassName("mail-group-label")}>
              <Pin size={12} />
              PINNED
            </div>
          )}
          {displayedRows.map((m, i) => (
            <button
              key={messageKey(m) || `${m.date}-${i}`}
              className={workspaceClassName(
                `mail-list-item ${!m.flags?.includes("\\Seen") ? "unread" : ""} ${current && messageKey(current) === messageKey(m) ? "selected" : ""}`,
              )}
              onClick={() => choose(m)}
            >
              <span
                className={workspaceClassName("mail-avatar")}
                style={{
                  background: [
                    "var(--muted)",
                    "var(--secondary)",
                    "var(--muted)",
                    "var(--secondary)",
                  ][i % 4],
                }}
              >
                {senderName(outbox ? m.to : m.from)
                  .slice(0, 1)
                  .toUpperCase()}
              </span>
              <div>
                <div className={workspaceClassName("mail-item-line")}>
                  <strong>{senderName(outbox ? m.to : m.from)}</strong>
                  <time>
                    {m.date
                      ? new Date(m.date).toLocaleDateString(undefined, {
                          month: "short",
                          day: "numeric",
                        })
                      : ""}
                  </time>
                </div>
                <span className={workspaceClassName("mail-subject")}>
                  {m.subject || "(No subject)"}
                </span>
                <p>
                  {outbox && m.status ? `${m.status} · ` : ""}
                  {text(m.body)}
                </p>
              </div>
            </button>
          ))}
          {emails.hasNextPage && (
            <>
              <div ref={loadSentinelRef} aria-hidden="true" className="h-px" />
              <Button
                className="m-4"
                variant="outline"
                disabled={emails.isFetchingNextPage}
                onClick={() => emails.fetchNextPage()}
              >
                {emails.isFetchingNextPage
                  ? "Loading…"
                  : emails.isFetchNextPageError
                    ? "Retry loading older messages"
                    : "Load more messages"}
              </Button>
            </>
          )}
        </div>
      </section>
      <section className={workspaceClassName("mail-detail-pane")}>
        {current ? (
          <>
            <div className={workspaceClassName("mail-detail-toolbar")}>
              <button
                className={workspaceClassName("icon-button mail-back")}
                aria-label={`Back to ${title.toLowerCase()}`}
                onClick={() => setSelected(null)}
              >
                <ArrowLeft />
              </button>
              <button
                className={workspaceClassName("icon-button")}
                aria-label={outbox ? "Write to recipient" : "Reply"}
                disabled={requiresDetail && detailState !== "ready"}
                onClick={() => setCompose(reply(current))}
              >
                <Reply />
              </button>
              {!outbox && (current.uid || current.providerMessageId) && (
                <>
                  <button
                    className={workspaceClassName("icon-button")}
                    disabled={
                      flagBusy || (requiresDetail && detailState !== "ready")
                    }
                    aria-label={
                      current.flags?.includes("\\Seen")
                        ? "Mark unread"
                        : "Mark read"
                    }
                    title={
                      current.flags?.includes("\\Seen")
                        ? "Mark unread"
                        : "Mark read"
                    }
                    onClick={() => void changeFlag("\\Seen")}
                  >
                    <MailCheck />
                  </button>
                  <button
                    className={workspaceClassName("icon-button")}
                    disabled={
                      flagBusy || (requiresDetail && detailState !== "ready")
                    }
                    aria-label={
                      current.flags?.includes("\\Flagged")
                        ? "Unstar message"
                        : "Star message"
                    }
                    aria-pressed={current.flags?.includes("\\Flagged") ?? false}
                    onClick={() => void changeFlag("\\Flagged")}
                  >
                    <Star
                      fill={
                        current.flags?.includes("\\Flagged")
                          ? "currentColor"
                          : "none"
                      }
                    />
                  </button>
                </>
              )}
              <span className="ml-auto text-xs text-muted-foreground">
                {index + 1} of {total}
              </span>
              <button
                className={workspaceClassName("icon-button")}
                aria-label="Previous message"
                disabled={index <= 0}
                onClick={() => choose(displayedRows[index - 1])}
              >
                <ArrowLeft />
              </button>
              <button
                className={workspaceClassName("icon-button")}
                aria-label="Next message"
                disabled={index >= rows.length - 1}
                onClick={() => choose(displayedRows[index + 1])}
              >
                <ArrowRight />
              </button>
              <button
                className={workspaceClassName("icon-button")}
                aria-label="Close message"
                onClick={() => setSelected(null)}
              >
                <X />
              </button>
            </div>
            <div className={workspaceClassName("mail-subject-heading")}>
              <p>
                {new Date(current.date).toLocaleString(undefined, {
                  dateStyle: "long",
                  timeStyle: "short",
                })}
              </p>
              <h2>{current.subject}</h2>
              {outbox && (
                <p>
                  Delivery: {current.status?.replaceAll("_", " ")} · Message ID:{" "}
                  {current.messageId}
                </p>
              )}
              {outbox && current.error && <p role="alert">{current.error}</p>}
            </div>
            <div className={workspaceClassName("mail-message")}>
              <div className={workspaceClassName("mail-sender")}>
                <span className={workspaceClassName("mail-avatar")}>
                  {senderName(current.from).slice(0, 1)}
                </span>
                <div>
                  <strong>{senderName(current.from)}</strong>
                  <p>
                    {(outbox ? current.to : current.from).match(
                      /<([^>]+)>/,
                    )?.[1] || (outbox ? current.to : current.from)}
                  </p>
                </div>
                <time>
                  {new Date(current.date).toLocaleTimeString(undefined, {
                    hour: "2-digit",
                    minute: "2-digit",
                  })}
                </time>
              </div>
              <p className={workspaceClassName("mail-recipients")}>
                To <span>{current.to}</span>
                {current.cc && (
                  <>
                    {" "}
                    · Cc <span>{current.cc}</span>
                  </>
                )}
              </p>
              {requiresDetail && detailState === "loading" && (
                <p role="status" className="text-sm text-muted-foreground">
                  Loading the complete message…
                </p>
              )}
              {requiresDetail && detailState === "error" && (
                <div role="alert" className="space-y-2 rounded-lg border p-3">
                  <p className="text-sm">
                    Couldn’t load the complete message. Try again before
                    replying or using its AI summary.
                  </p>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => void loadMessageDetail(current)}
                  >
                    Retry loading message
                  </Button>
                </div>
              )}
              {(!requiresDetail || detailState === "ready") && (
                <>
                  {!outbox && current.providerMessageId ? (
                    <MailSummary
                      configId={configId}
                      folder={folder}
                      providerMessageId={current.providerMessageId}
                    />
                  ) : !outbox &&
                    current.uid != null &&
                    current.uidValidity != null ? (
                    <MailSummary
                      configId={configId}
                      folder={folder}
                      uid={current.uid}
                      uidValidity={current.uidValidity}
                    />
                  ) : null}
                  {current.warning && (
                    <p
                      role="note"
                      className="rounded-lg border px-3 py-2 text-sm text-muted-foreground"
                    >
                      {current.warning}
                    </p>
                  )}
                  {hasRemoteImages(current.body) && (
                    <div className={workspaceClassName("remote-images-note")}>
                      {images
                        ? "Remote images enabled for this message."
                        : "Remote images are hidden to protect your privacy."}{" "}
                      {!images && (
                        <button onClick={() => setImages(true)}>
                          Load images
                        </button>
                      )}
                    </div>
                  )}
                  <iframe
                    className={workspaceClassName("mail-body")}
                    title="Email content"
                    sandbox=""
                    referrerPolicy="no-referrer"
                    srcDoc={`<meta name="color-scheme" content="light"><meta http-equiv="Content-Security-Policy" content="default-src 'none'; img-src ${images ? "https:" : ""} data:; style-src 'unsafe-inline'; base-uri 'none'; form-action 'none'"><style>:root{color-scheme:light}html,body{background:#fff}body{font:14px/1.8 Arial;color:#3f3b43;overflow-wrap:anywhere;margin:0;padding:16px}img{max-width:100%}</style>${current.body}`}
                  />
                  {current.attachments?.length > 0 && (
                    <section
                      className={workspaceClassName("mail-attachment-section")}
                    >
                      <h3>Attachments ({current.attachments.length})</h3>
                      <div className={workspaceClassName("mail-attachments")}>
                        {current.attachments.map((a, i) =>
                          typeof a.Data === "string" ? (
                            <a
                              key={i}
                              className={workspaceClassName("mail-attachment")}
                              download={a.Filename}
                              href={`data:application/octet-stream;base64,${a.Data}`}
                            >
                              <Paperclip size={14} />
                              <span>
                                <strong>{a.Filename}</strong>
                                <small>{attachmentSize(a.Data)}</small>
                              </span>
                              <em>Download</em>
                            </a>
                          ) : (
                            <button
                              key={a.AttachmentID || i}
                              className={workspaceClassName("mail-attachment")}
                              type="button"
                              disabled={
                                !a.AttachmentID ||
                                attachmentBusy ===
                                  `${current.providerMessageId || `${current.uid}:${current.uidValidity}`}:${a.AttachmentID}`
                              }
                              onClick={() =>
                                void downloadAttachment(current, a)
                              }
                            >
                              <Paperclip size={14} />
                              <span>
                                <strong>{a.Filename}</strong>
                                {typeof a.Size === "number" && (
                                  <small>{byteSize(a.Size)}</small>
                                )}
                              </span>
                              <em>
                                {attachmentBusy ===
                                `${current.providerMessageId || `${current.uid}:${current.uidValidity}`}:${a.AttachmentID}`
                                  ? "Loading…"
                                  : "Download"}
                              </em>
                            </button>
                          ),
                        )}
                      </div>
                    </section>
                  )}
                  <Button
                    variant="outline"
                    className="mt-5"
                    onClick={() => setCompose(reply(current))}
                  >
                    <Reply />
                    {outbox ? "Write to recipient" : "Reply to conversation"}
                  </Button>
                </>
              )}
            </div>
          </>
        ) : (
          <div className={workspaceClassName("mail-empty")}>
            <Mail size={35} strokeWidth={1} />
            <h2>
              {outbox ? "Your outgoing messages" : "A space for conversation"}
            </h2>
            <p>Select a message to read it here.</p>
          </div>
        )}
      </section>
      {!current && !compose && (
        <Button
          aria-label="Compose message"
          className={workspaceClassName("mail-compose-fab")}
          onClick={() =>
            setCompose({
              to: "",
              subject: "",
              smtpConfigId: mailbox?.smtpConfigId,
            })
          }
        >
          <PencilLine />
          Compose
        </Button>
      )}
      {compose && (
        <MailCompose
          key={scope}
          value={compose}
          close={() => setCompose(null)}
        />
      )}
    </div>
  );
}
