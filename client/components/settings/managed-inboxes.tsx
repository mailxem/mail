"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import {
  AlertTriangle,
  Check,
  Clipboard,
  Inbox,
  RefreshCw,
} from "lucide-react";
import { toast } from "sonner";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  MarketingRequestError,
  useMarketing,
  useMarketingQuery,
} from "@/lib/marketing/api";

type ReceivingStatus =
  | "unconfigured"
  | "needs_verification"
  | "needs_mailbox"
  | "pending_mx"
  | "mx_conflict"
  | "provisioning"
  | "ready"
  | "error";
type Domain = {
  id: string;
  name: string;
  sendingReady: boolean;
  receivingStatus: ReceivingStatus;
  mx: { name: string; type: "MX"; value: string; priority: number };
  existingMX: Array<{ host: string; priority: number }>;
  detail?: string;
};
type Mailbox = {
  id: string;
  domainId: string;
  address: string;
  displayName: string;
  active: boolean;
  smtpConfigId?: string;
  status: string;
  usageBytes: number;
  quotaBytes: number;
};
type Receiving = {
  enabled: boolean;
  region: string;
  maxMailboxesPerDomain: number;
  maxMessageBytes: number;
  domains: Domain[];
  mailboxes: Mailbox[];
};

const statusText: Record<ReceivingStatus, string> = {
  unconfigured: "Receiving is not configured",
  needs_verification: "Verify domain ownership first",
  needs_mailbox: "Ready to create a mailbox",
  pending_mx: "Waiting for receiving MX",
  mx_conflict: "Current mail provider MX detected",
  provisioning: "Provisioning receiving",
  ready: "Receiving ready",
  error: "Receiving needs attention",
};
const statusValues = new Set(Object.keys(statusText));
const text = (value: unknown) => (typeof value === "string" ? value : "");
const bytes = (value: number) => {
  if (value < 1024) return `${value} B`;
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`;
  if (value < 1024 * 1024 * 1024)
    return `${(value / 1024 / 1024).toFixed(1)} MB`;
  return `${(value / 1024 / 1024 / 1024).toFixed(1)} GB`;
};
function parseReceiving(value: unknown): Receiving | null {
  if (!value || typeof value !== "object") return null;
  const raw = value as Record<string, unknown>;
  if (
    typeof raw.enabled !== "boolean" ||
    !Array.isArray(raw.domains) ||
    !Array.isArray(raw.mailboxes)
  )
    return null;
  const domains = raw.domains.flatMap((entry) => {
    if (!entry || typeof entry !== "object") return [];
    const item = entry as Record<string, any>;
    if (
      !text(item.id) ||
      !text(item.name) ||
      !statusValues.has(item.receivingStatus) ||
      !item.mx
    )
      return [];
    return [
      {
        id: text(item.id),
        name: text(item.name),
        sendingReady: item.sendingReady === true,
        receivingStatus: item.receivingStatus as ReceivingStatus,
        mx: {
          name: text(item.mx.name),
          type: "MX" as const,
          value: text(item.mx.value),
          priority: Number(item.mx.priority) || 0,
        },
        existingMX: Array.isArray(item.existingMX)
          ? item.existingMX
              .map((mx: any) => ({
                host: text(mx?.host),
                priority: Number(mx?.priority) || 0,
              }))
              .filter((mx: any) => mx.host)
          : [],
        detail: text(item.detail) || undefined,
      },
    ];
  });
  if (raw.domains.length > 0 && domains.length === 0) return null;
  const mailboxes = raw.mailboxes.flatMap((entry) => {
    if (!entry || typeof entry !== "object") return [];
    const item = entry as Record<string, unknown>;
    if (!text(item.id) || !text(item.domainId) || !text(item.address))
      return [];
    return [
      {
        id: text(item.id),
        domainId: text(item.domainId),
        address: text(item.address),
        displayName: text(item.displayName),
        active: item.active === true,
        smtpConfigId: text(item.smtpConfigId) || undefined,
        status: text(item.status),
        usageBytes: Math.max(0, Number(item.usageBytes) || 0),
        quotaBytes: Math.max(0, Number(item.quotaBytes) || 0),
      },
    ];
  });
  return {
    enabled: raw.enabled,
    region: text(raw.region),
    maxMailboxesPerDomain: Number(raw.maxMailboxesPerDomain) || 0,
    maxMessageBytes: Number(raw.maxMessageBytes) || 0,
    domains,
    mailboxes,
  };
}

export function ManagedInboxes() {
  const query = useMarketingQuery<unknown>("mail-connections/receiving");
  const { request, refresh } = useMarketing();
  const data = useMemo(() => parseReceiving(query.data), [query.data]);
  const [domainId, setDomainId] = useState("");
  const [localPart, setLocalPart] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [busy, setBusy] = useState("");
  const selectedDomain =
    data?.domains.find((domain) => domain.id === domainId) ?? data?.domains[0];
  const selectedMailboxCount = selectedDomain
    ? (data?.mailboxes.filter(
        (mailbox) => mailbox.domainId === selectedDomain.id,
      ).length ?? 0)
    : 0;
  const mailboxLimitReached =
    !!data?.maxMailboxesPerDomain &&
    selectedMailboxCount >= data.maxMailboxesPerDomain;
  const domainCanCreate =
    !!selectedDomain &&
    ["needs_mailbox", "pending_mx", "mx_conflict", "ready"].includes(
      selectedDomain.receivingStatus,
    );
  const normalizedLocalPart = localPart.trim().toLowerCase();
  const localPartValid =
    normalizedLocalPart.length <= 64 &&
    /^[a-z0-9](?:[a-z0-9.!#$%&'*+/=?^_`{|}~-]{0,62}[a-z0-9])?$/.test(
      normalizedLocalPart,
    ) &&
    !normalizedLocalPart.includes("..");
  async function mutate(
    key: string,
    action: () => Promise<unknown>,
    message: string,
  ): Promise<boolean> {
    setBusy(key);
    try {
      await action();
      await refresh();
      toast.success(message);
      return true;
    } catch (error) {
      toast.error((error as Error).message);
      return false;
    } finally {
      setBusy("");
    }
  }
  async function copy(value: string) {
    try {
      await navigator.clipboard.writeText(value);
      toast.success("MX value copied");
    } catch {
      toast.error("Couldn’t copy the MX value");
    }
  }
  return (
    <section
      id="managed-inboxes"
      className="mb-6 space-y-5 rounded-xl border bg-card p-4 sm:p-5"
    >
      <div className="flex flex-col justify-between gap-3 sm:flex-row sm:items-start">
        <div className="flex min-w-0 gap-3">
          <span className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary text-primary-foreground">
            <Inbox size={18} />
          </span>
          <div>
            <h2 className="font-semibold">Xem inbox</h2>
            <p className="mt-1 max-w-2xl text-sm text-muted-foreground">
              Receive mail for your custom domain in this shared workspace.
              Mailbox content is shared with workspace members who can use the
              inbox. Workspace admins manage connections and lifecycle.
            </p>
          </div>
        </div>
        <nav
          aria-label="Mailbox providers"
          className="flex flex-wrap gap-x-3 gap-y-1 text-xs"
        >
          <a className="font-medium text-primary" href="#managed-inboxes">
            Xem inbox
          </a>
          <a
            className="text-muted-foreground hover:text-foreground"
            href="#connected-mail-google"
          >
            Gmail
          </a>
          <Link
            className="text-muted-foreground hover:text-foreground"
            href="/settings/smtp#connected-mail-cloudflare"
          >
            Cloudflare
          </Link>
          <a
            className="text-muted-foreground hover:text-foreground"
            href="#other-imap"
          >
            Other IMAP
          </a>
        </nav>
      </div>
      {query.isLoading ? (
        <p className="text-sm text-muted-foreground">
          Loading receiving setup…
        </p>
      ) : query.error instanceof MarketingRequestError &&
        query.error.status === 403 ? (
        <div className="rounded-lg border border-dashed p-4">
          <p className="text-sm font-medium">
            A workspace admin can manage receiving mailboxes.
          </p>
          <p className="mt-1 text-xs text-muted-foreground">
            You can still use mailboxes already shared with your workspace from
            the inbox.
          </p>
        </div>
      ) : query.error ? (
        <div role="alert" className="space-y-2">
          <p className="text-sm text-muted-foreground">
            Couldn’t load receiving setup. {query.error.message}
          </p>
          <Button size="sm" variant="outline" onClick={() => query.refetch()}>
            Retry
          </Button>
        </div>
      ) : !data ? (
        <div role="alert" className="space-y-2">
          <p className="text-sm text-muted-foreground">
            Receiving setup returned an unexpected response.
          </p>
          <Button size="sm" variant="outline" onClick={() => query.refetch()}>
            Retry
          </Button>
        </div>
      ) : !data.enabled ? (
        <div className="rounded-lg border border-dashed p-4">
          <p className="text-sm font-medium">
            Managed receiving is unavailable on this installation.
          </p>
          <p className="mt-1 text-xs text-muted-foreground">
            Your administrator can enable it. Gmail, Cloudflare, and other IMAP
            connections remain available.
          </p>
        </div>
      ) : data.domains.length === 0 ? (
        <div className="rounded-lg border border-dashed p-4">
          <p className="text-sm font-medium">Add and verify a domain first</p>
          <p className="mt-1 text-xs text-muted-foreground">
            Domain ownership and sending readiness are checked separately from
            receiving.
          </p>
          <Button asChild size="sm" variant="outline" className="mt-3">
            <Link href="/settings/sending">Set up a domain</Link>
          </Button>
        </div>
      ) : (
        <>
          <div className="grid gap-3 lg:grid-cols-2">
            {data.domains.map((domain) => {
              const selected = selectedDomain?.id === domain.id;
              return (
                <button
                  key={domain.id}
                  type="button"
                  onClick={() => setDomainId(domain.id)}
                  className={`rounded-lg border p-3 text-left transition-colors ${selected ? "border-primary bg-primary/5" : "hover:bg-muted/50"}`}
                >
                  <span className="flex items-start justify-between gap-3">
                    <span className="min-w-0">
                      <strong className="block truncate text-sm">
                        {domain.name}
                      </strong>
                      <span className="mt-1 block text-xs text-muted-foreground">
                        {statusText[domain.receivingStatus]}
                      </span>
                    </span>
                    {selected && <Check className="size-4 text-primary" />}
                  </span>
                  <span className="mt-3 flex flex-wrap gap-2">
                    <Badge
                      variant={
                        domain.receivingStatus === "ready"
                          ? "success"
                          : "outline"
                      }
                    >
                      Receiving:{" "}
                      {domain.receivingStatus === "ready"
                        ? "ready"
                        : "setup needed"}
                    </Badge>
                    <Badge
                      variant={domain.sendingReady ? "success" : "outline"}
                    >
                      Sending: {domain.sendingReady ? "ready" : "not ready"}
                    </Badge>
                  </span>
                </button>
              );
            })}
          </div>
          {selectedDomain && (
            <div className="space-y-4 rounded-lg bg-muted/45 p-3 sm:p-4">
              {selectedDomain.receivingStatus === "mx_conflict" && (
                <div className="flex gap-2 rounded-md border border-amber-500/30 bg-amber-500/10 p-3 text-xs">
                  <AlertTriangle className="size-4 shrink-0 text-amber-600" />
                  <div>
                    <p>
                      Your current provider’s MX records still receive mail.
                      Replacing them moves new inbound mail to Xem; MX records
                      cannot split or mirror delivery between providers. Test
                      safely on a subdomain before migrating.
                    </p>
                    {selectedDomain.existingMX.length > 0 && (
                      <p className="mt-1 text-muted-foreground">
                        Current MX:{" "}
                        {selectedDomain.existingMX
                          .map((record) => `${record.priority} ${record.host}`)
                          .join(", ")}
                      </p>
                    )}
                  </div>
                </div>
              )}
              {selectedDomain.mx.value && (
                <div>
                  <div className="flex flex-wrap items-end justify-between gap-2">
                    <div>
                      <p className="text-sm font-medium">Receiving MX record</p>
                      <p className="text-xs text-muted-foreground">
                        This routes inbox mail. It is separate from sending
                        bounce and MAIL FROM records.
                      </p>
                    </div>
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => void copy(selectedDomain.mx.value)}
                    >
                      <Clipboard />
                      Copy value
                    </Button>
                  </div>
                  <div className="mt-2 grid gap-3 rounded-md border bg-background p-3 text-xs sm:grid-cols-[1fr_auto_2fr]">
                    <div className="min-w-0">
                      <p className="mb-1 text-[10px] uppercase tracking-wide text-muted-foreground">
                        Host
                      </p>
                      <code className="break-all">
                        {selectedDomain.mx.name}
                      </code>
                    </div>
                    <div>
                      <p className="mb-1 text-[10px] uppercase tracking-wide text-muted-foreground">
                        Priority
                      </p>
                      <code>{selectedDomain.mx.priority}</code>
                    </div>
                    <div className="min-w-0">
                      <p className="mb-1 text-[10px] uppercase tracking-wide text-muted-foreground">
                        Value
                      </p>
                      <code className="break-all">
                        {selectedDomain.mx.value}
                      </code>
                    </div>
                  </div>
                </div>
              )}
              {selectedDomain.detail && (
                <p className="text-xs text-muted-foreground">
                  {selectedDomain.detail}
                </p>
              )}
              <Button
                size="sm"
                variant="outline"
                disabled={!!busy}
                onClick={() =>
                  void mutate(
                    `check-${selectedDomain.id}`,
                    () =>
                      request(
                        `mail-connections/receiving/domains/${selectedDomain.id}/check`,
                        "POST",
                      ),
                    "DNS check complete",
                  )
                }
              >
                <RefreshCw />
                {busy === `check-${selectedDomain.id}`
                  ? "Checking…"
                  : "Check DNS"}
              </Button>
            </div>
          )}
          {selectedDomain && (
            <form
              className="grid gap-3 rounded-lg border p-3 sm:grid-cols-[1fr_1fr_auto] sm:items-end"
              onSubmit={(event) => {
                event.preventDefault();
                if (!localPartValid || !domainCanCreate || mailboxLimitReached)
                  return;
                void mutate(
                  "create",
                  () =>
                    request("mail-connections/receiving/mailboxes", "POST", {
                      domainId: selectedDomain.id,
                      localPart: normalizedLocalPart,
                      ...(displayName.trim()
                        ? { displayName: displayName.trim() }
                        : {}),
                    }),
                  "Mailbox created",
                ).then((saved) => {
                  if (saved) {
                    setLocalPart("");
                    setDisplayName("");
                  }
                });
              }}
            >
              <div className="space-y-1.5">
                <Label htmlFor="managed-local-part">Mailbox address</Label>
                <div className="flex min-w-0 items-center rounded-md border bg-background">
                  <Input
                    id="managed-local-part"
                    className="border-0 shadow-none"
                    value={localPart}
                    onChange={(event) => setLocalPart(event.target.value)}
                    placeholder="alex"
                    maxLength={64}
                    aria-invalid={!!localPart && !localPartValid}
                  />
                  <span
                    className="max-w-[55%] shrink-0 truncate pr-3 text-sm text-muted-foreground"
                    title={`@${selectedDomain.name}`}
                  >
                    @{selectedDomain.name}
                  </span>
                </div>
                {localPart && !localPartValid && (
                  <p className="text-xs text-destructive">
                    Enter a valid mailbox name without spaces.
                  </p>
                )}
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="managed-display-name">
                  Display name{" "}
                  <span className="text-muted-foreground">(optional)</span>
                </Label>
                <Input
                  id="managed-display-name"
                  value={displayName}
                  onChange={(event) => setDisplayName(event.target.value)}
                  placeholder="Alex"
                  maxLength={120}
                />
              </div>
              <Button
                type="submit"
                disabled={
                  !localPartValid ||
                  !!busy ||
                  !domainCanCreate ||
                  mailboxLimitReached
                }
              >
                {busy === "create" ? "Creating…" : "Create mailbox"}
              </Button>
              <p className="break-all text-xs text-muted-foreground sm:col-span-3">
                Full address:{" "}
                <strong className="font-medium text-foreground">
                  {normalizedLocalPart || "mailbox"}@{selectedDomain.name}
                </strong>
              </p>
              {(mailboxLimitReached || !domainCanCreate) && (
                <p className="text-xs text-muted-foreground sm:col-span-3">
                  {mailboxLimitReached
                    ? `This domain has reached its ${data.maxMailboxesPerDomain}-mailbox limit.`
                    : "Finish domain verification or receiving setup before creating a mailbox."}
                </p>
              )}
            </form>
          )}
          <div className="space-y-2">
            {data.mailboxes.map((mailbox) => (
              <div
                key={mailbox.id}
                className="flex flex-wrap items-center justify-between gap-3 border-t pt-3"
              >
                <div>
                  <p className="text-sm font-medium break-all">
                    {mailbox.address}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    {mailbox.active
                      ? mailbox.status === "storage_full"
                        ? "Storage full. New mail cannot be received until a workspace admin restores capacity."
                        : mailbox.status === "active"
                          ? "Receiving is active"
                          : `Receiving status: ${mailbox.status || "setting up"}`
                      : "Paused. Existing mail is preserved."}{" "}
                    ·{" "}
                    {mailbox.smtpConfigId
                      ? "Sender linked"
                      : "No sender linked to this address"}
                  </p>
                  {mailbox.quotaBytes > 0 && (
                    <p
                      className={`mt-1 text-xs ${mailbox.status === "storage_full" ? "font-medium text-destructive" : "text-muted-foreground"}`}
                    >
                      Storage: {bytes(mailbox.usageBytes)} of{" "}
                      {bytes(mailbox.quotaBytes)} used
                    </p>
                  )}
                </div>
                <Button
                  size="sm"
                  variant="outline"
                  disabled={!!busy}
                  onClick={() =>
                    void mutate(
                      mailbox.id,
                      () =>
                        request(
                          `mail-connections/receiving/mailboxes/${mailbox.id}`,
                          "PATCH",
                          { active: !mailbox.active },
                        ),
                      mailbox.active
                        ? "Mailbox paused; existing mail is preserved"
                        : "Mailbox resumed",
                    )
                  }
                >
                  {busy === mailbox.id
                    ? "Saving…"
                    : mailbox.active
                      ? "Pause"
                      : "Resume"}
                </Button>
              </div>
            ))}
          </div>
        </>
      )}
      <p className="text-xs text-muted-foreground">
        Mailbox content is visible to workspace members with inbox access. Use a
        subdomain such as <code>inbox.example.com</code> to test without moving
        your primary domain.
      </p>
    </section>
  );
}
