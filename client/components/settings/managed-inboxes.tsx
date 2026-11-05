"use client";

import Link from "next/link";
import { PaidPlanNotice, usePaidPlan } from "@/components/billing/paid-plan";
import { useMemo, useState } from "react";
import {
  AlertTriangle,
  Clipboard,
  Inbox,
  Plus,
  RefreshCw,
} from "lucide-react";
import { toast } from "sonner";
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
  const paid = usePaidPlan();
  const data = useMemo(() => parseReceiving(query.data), [query.data]);
  const [domainId, setDomainId] = useState("");
  const [localPart, setLocalPart] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [busy, setBusy] = useState("");
  const [composing, setComposing] = useState(false);
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
      const result = await action();
      await refresh();
      if ((result as { status?: string } | null)?.status === "error")
        toast.warning(
          "Mailbox created, but receiving isn’t set up yet. Use Check DNS to retry; mail won’t arrive until it shows as ready.",
        );
      else toast.success(message);
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
  const needsDNS = data?.domains.filter((domain) =>
    ["pending_mx", "mx_conflict", "provisioning", "error"].includes(
      domain.receivingStatus,
    ) && data.mailboxes.some((mailbox) => mailbox.domainId === domain.id),
  ) ?? [];
  const showForm = composing || (!!data && data.mailboxes.length === 0);
  const body = query.isLoading ? (
    <p className="text-sm text-muted-foreground">Loading…</p>
  ) : query.error instanceof MarketingRequestError &&
    query.error.status === 403 ? (
    <p className="text-sm text-muted-foreground">
      A workspace admin can manage receiving mailboxes.
    </p>
  ) : query.error || !data ? (
    <div role="alert" className="flex items-center justify-between gap-3">
      <p className="text-sm text-muted-foreground">
        Couldn’t load receiving setup.
        {query.error ? ` ${query.error.message}` : ""}
      </p>
      <Button size="sm" variant="outline" onClick={() => query.refetch()}>
        Retry
      </Button>
    </div>
  ) : !paid ? (
    <PaidPlanNotice feature="The Xem inbox" />
  ) : !data.enabled ? (
    <p className="text-sm text-muted-foreground">
      Managed receiving is unavailable on this installation.
    </p>
  ) : data.domains.length === 0 ? (
    <div className="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-dashed p-4">
      <div>
        <p className="text-sm font-medium">Add a domain first</p>
        <p className="mt-0.5 text-xs text-muted-foreground">
          Addresses are created on a domain you’ve verified for sending.
        </p>
      </div>
      <Button asChild size="sm" variant="outline">
        <Link href="/settings/sending">Set up a domain</Link>
      </Button>
    </div>
  ) : (
    <div className="space-y-4">
      {needsDNS.map((domain) => (
        <DNSStep
          key={domain.id}
          domain={domain}
          busy={busy === `check-${domain.id}`}
          disabled={!!busy}
          onCopy={copy}
          onCheck={() =>
            void mutate(
              `check-${domain.id}`,
              () =>
                request(
                  `mail-connections/receiving/domains/${domain.id}/check`,
                  "POST",
                ),
              "DNS check complete",
            )
          }
        />
      ))}

      {data.mailboxes.length > 0 && (
        <ul className="divide-y rounded-lg border">
          {data.mailboxes.map((mailbox) => {
            const domain = data.domains.find((d) => d.id === mailbox.domainId);
            const pill = mailboxPill(mailbox, domain);
            return (
              <li
                key={mailbox.id}
                className="flex flex-wrap items-center gap-3 px-4 py-3"
              >
                <span className="grid size-8 shrink-0 place-items-center rounded-full bg-muted text-xs font-medium uppercase">
                  {(mailbox.displayName || mailbox.address).charAt(0)}
                </span>
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium">
                    {mailbox.address}
                  </p>
                  <p className="truncate text-xs text-muted-foreground">
                    {[
                      mailbox.displayName,
                      mailbox.quotaBytes > 0
                        ? `Storage: ${bytes(mailbox.usageBytes)} of ${bytes(mailbox.quotaBytes)} used`
                        : "",
                    ]
                      .filter(Boolean)
                      .join(" · ")}
                  </p>
                </div>
                <span
                  className={`inline-flex items-center gap-1.5 rounded-md px-2 py-1 text-[11px] font-medium ${pill.tone}`}
                >
                  <i className="size-1.5 rounded-full bg-current" />
                  {pill.label}
                </span>
                <Button
                  size="sm"
                  variant="ghost"
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
              </li>
            );
          })}
        </ul>
      )}

      {showForm && selectedDomain ? (
        <form
          className="space-y-3 rounded-lg border bg-muted/30 p-4"
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
                setComposing(false);
              }
            });
          }}
        >
          <div className="grid gap-3 sm:grid-cols-[1.4fr_1fr]">
            <div className="space-y-1.5">
              <Label htmlFor="managed-local-part">Address</Label>
              <div className="flex min-w-0 items-center rounded-md border bg-background focus-within:ring-2 focus-within:ring-ring/30">
                <Input
                  id="managed-local-part"
                  className="border-0 shadow-none focus-visible:ring-0"
                  value={localPart}
                  onChange={(event) => setLocalPart(event.target.value)}
                  placeholder="hello"
                  maxLength={64}
                  autoFocus={composing}
                  aria-invalid={!!localPart && !localPartValid}
                />
                {data.domains.length > 1 ? (
                  <select
                    aria-label="Domain"
                    className="mr-1 max-w-[55%] shrink-0 truncate rounded bg-transparent py-1 pr-1 text-sm text-muted-foreground outline-none"
                    value={selectedDomain.id}
                    onChange={(event) => setDomainId(event.target.value)}
                  >
                    {data.domains.map((domain) => (
                      <option key={domain.id} value={domain.id}>
                        @{domain.name}
                      </option>
                    ))}
                  </select>
                ) : (
                  <span
                    className="max-w-[55%] shrink-0 truncate pr-3 text-sm text-muted-foreground"
                    title={`@${selectedDomain.name}`}
                  >
                    @{selectedDomain.name}
                  </span>
                )}
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
                <span className="font-normal text-muted-foreground">
                  (optional)
                </span>
              </Label>
              <Input
                id="managed-display-name"
                value={displayName}
                onChange={(event) => setDisplayName(event.target.value)}
                placeholder="Support"
                maxLength={120}
              />
            </div>
          </div>
          <div className="flex flex-wrap items-center justify-between gap-3">
            <p className="text-xs text-muted-foreground">
              {mailboxLimitReached
                ? `This domain has reached its ${data.maxMailboxesPerDomain}-mailbox limit.`
                : !domainCanCreate
                  ? statusText[selectedDomain.receivingStatus]
                  : <>Full address: <strong className="font-medium text-foreground">{normalizedLocalPart || "mailbox"}@{selectedDomain.name}</strong></>}
            </p>
            <div className="flex gap-2">
              {data.mailboxes.length > 0 && (
                <Button
                  type="button"
                  size="sm"
                  variant="ghost"
                  onClick={() => setComposing(false)}
                >
                  Cancel
                </Button>
              )}
              <Button
                type="submit"
                size="sm"
                disabled={
                  !localPartValid ||
                  !!busy ||
                  !domainCanCreate ||
                  mailboxLimitReached
                }
              >
                {busy === "create" ? "Creating…" : "Create mailbox"}
              </Button>
            </div>
          </div>
        </form>
      ) : null}
    </div>
  );
  const canAdd =
    paid && !!data?.enabled && !!data?.domains.length && !showForm;
  return (
    <section
      id="managed-inboxes"
      className="mb-6 space-y-4 rounded-xl border bg-card p-5"
    >
      <div className="flex items-start justify-between gap-4">
        <div className="flex min-w-0 gap-3">
          <span className="grid size-9 shrink-0 place-items-center rounded-lg bg-primary text-primary-foreground">
            <Inbox size={18} />
          </span>
          <div className="min-w-0">
            <h2 className="font-semibold">Xem inbox</h2>
            <p className="mt-0.5 text-sm text-muted-foreground">
              Get mail at your own domain, stored and shared in this workspace.
            </p>
          </div>
        </div>
        {canAdd && (
          <Button size="sm" variant="outline" onClick={() => setComposing(true)}>
            <Plus /> New address
          </Button>
        )}
      </div>
      {body}
    </section>
  );
}

function mailboxPill(
  mailbox: { active: boolean; status: string },
  domain?: { receivingStatus: ReceivingStatus },
) {
  if (!mailbox.active)
    return { label: "Paused", tone: "bg-muted text-muted-foreground" };
  if (mailbox.status === "storage_full")
    return { label: "Storage full", tone: "bg-destructive/10 text-destructive" };
  if (domain && domain.receivingStatus !== "ready")
    return {
      label: domain.receivingStatus === "error" ? "Needs attention" : "Waiting for DNS",
      tone: "bg-amber-500/10 text-amber-700 dark:text-amber-400",
    };
  return { label: "Receiving", tone: "bg-[var(--success)] text-[var(--success-foreground)]" };
}

function DNSStep({
  domain,
  busy,
  disabled,
  onCopy,
  onCheck,
}: {
  domain: {
    name: string;
    receivingStatus: ReceivingStatus;
    mx: { name: string; value: string; priority: number };
    existingMX: Array<{ host: string; priority: number }>;
    detail?: string;
  };
  busy: boolean;
  disabled: boolean;
  onCopy: (value: string) => void;
  onCheck: () => void;
}) {
  const conflict = domain.receivingStatus === "mx_conflict";
  return (
    <div className="rounded-lg border border-amber-500/30 bg-amber-500/5 p-4">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p className="text-sm font-medium">
            Finish setup for {domain.name}
          </p>
          <p className="mt-0.5 text-xs text-muted-foreground">
            Add this record at your DNS provider. Mail starts arriving once Xem
            sees it, usually within minutes.
          </p>
        </div>
        <Button size="sm" variant="outline" disabled={disabled} onClick={onCheck}>
          <RefreshCw className={busy ? "animate-spin" : ""} />
          {busy ? "Checking…" : "Check DNS"}
        </Button>
      </div>
      <div className="mt-3 overflow-hidden rounded-md border bg-background text-xs">
        <div className="grid grid-cols-[56px_1fr_64px_2fr_auto] gap-3 border-b bg-muted/40 px-3 py-2 text-[10px] font-medium uppercase tracking-wide text-muted-foreground max-sm:hidden">
          <span>Type</span>
          <span>Host</span>
          <span>Priority</span>
          <span>Value</span>
          <span />
        </div>
        <div className="grid items-center gap-3 px-3 py-2.5 sm:grid-cols-[56px_1fr_64px_2fr_auto]">
          <code>MX</code>
          <code className="break-all">{domain.mx.name}</code>
          <code>{domain.mx.priority}</code>
          <code className="break-all">{domain.mx.value}</code>
          <Button
            size="sm"
            variant="ghost"
            className="h-7 justify-self-start"
            onClick={() => void onCopy(domain.mx.value)}
          >
            <Clipboard /> Copy
          </Button>
        </div>
      </div>
      {conflict && (
        <p className="mt-3 flex gap-2 text-xs text-amber-800 dark:text-amber-300">
          <AlertTriangle className="size-4 shrink-0" />
          This domain still has MX records for another provider
          {domain.existingMX.length > 0
            ? ` (${domain.existingMX.map((r) => r.host).join(", ")})`
            : ""}
          . Replacing them moves all new mail to Xem; MX records cannot split
          delivery. Use a subdomain to test first.
        </p>
      )}
      {domain.detail && !conflict && (
        <p className="mt-3 text-xs text-muted-foreground">{domain.detail}</p>
      )}
    </div>
  );
}
