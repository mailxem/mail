"use client";
import { useState } from "react";
import {
  Check,
  Copy,
  ArrowUpRight,
  Globe,
  RefreshCw,
  Mail,
  Loader2,
  CircleHelp,
} from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useMarketing } from "@/lib/marketing/api";
import type { SendingDomain } from "@/lib/sending/types";

export function CopyButton({
  value,
  label = "Copy",
}: {
  value: string;
  label?: string;
}) {
  const [copied, setCopied] = useState(false);
  async function copy() {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      setTimeout(() => setCopied(false), 1800);
    } catch {
      toast.error("Could not copy. Select the text and copy it manually.");
    }
  }
  return (
    <Button
      type="button"
      variant="ghost"
      size="sm"
      aria-label={label}
      onClick={copy}
    >
      {copied ? <Check size={15} /> : <Copy size={15} />}
      <span className="ml-2">{copied ? "Copied" : label}</span>
    </Button>
  );
}
export function DomainForm({ onDone }: { onDone: () => void | Promise<void> }) {
  const { request } = useMarketing();
  const [name, setName] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      await request("sending/domains", "POST", { name });
      setName("");
      await onDone();
      toast.success("Your domain is saved. Let’s connect it.");
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }
  return (
    <form onSubmit={submit} className="space-y-5">
      <div className="space-y-2">
        <Label htmlFor="sending-domain">Your sending domain</Label>
        <Input
          id="sending-domain"
          autoComplete="off"
          placeholder="updates.yourcompany.com"
          value={name}
          onChange={(e) => setName(e.target.value)}
          required
          maxLength={253}
          className="h-12"
        />
        <p className="text-sm text-muted-foreground">
          Use a domain you own. A subdomain such as{" "}
          <code>updates.yourcompany.com</code> keeps your marketing setup
          organized.
        </p>
      </div>
      {error && (
        <p role="alert" className="text-sm text-destructive">
          {error}
        </p>
      )}
      <Button disabled={busy || !name.trim()} className="h-11">
        {busy ? (
          <Loader2
            className="mr-2 animate-spin motion-reduce:animate-none"
            size={16}
          />
        ) : (
          <Globe className="mr-2" size={16} />
        )}
        Add my domain
      </Button>
      <p className="text-xs text-muted-foreground">
        Your existing inbox keeps working. You won’t need to move it.
      </p>
    </form>
  );
}
export function DomainDNS({
  domain,
  onDone,
}: {
  domain: SendingDomain;
  onDone: () => void | Promise<void>;
}) {
  const { request } = useMarketing();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  async function check() {
    setBusy(true);
    setError("");
    try {
      const next = await request<SendingDomain>(
        `sending/domains/${domain.id}/check`,
        "POST",
      );
      await onDone();
      if (next.ready) {
        toast.success(
          "All connected. Continue to your sender when you’re ready.",
        );
      } else {
        toast.info(
          next.identityStatus === "AWAITING_APPROVAL"
            ? "Ownership confirmed. Managed sending approval is still needed."
            : next.ownership
              ? "Ownership confirmed. Add and verify the delivery records below."
              : "Still waiting for your ownership record. Your progress is saved.",
        );
      }
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }
  const states: [[string, boolean], ...Array<[string, boolean]>] = [
    ["Ownership", domain.ownership],
    ["DKIM", domain.dkimStatus === "SUCCESS"],
    ["Bounce domain", domain.mailFromStatus === "SUCCESS"],
    ["DMARC", domain.dmarcStatus === "VALID"],
  ];
  return (
    <div className="space-y-5">
      <div className="flex flex-wrap gap-2">
        {states.map(([label, ok]) => (
          <span
            key={label}
            className={`inline-flex items-center gap-1.5 rounded-full px-3 py-1.5 text-xs ${ok ? "bg-emerald-50 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-200" : "bg-muted text-muted-foreground"}`}
          >
            {ok ? (
              <Check size={13} />
            ) : (
              <span className="size-1.5 rounded-full bg-current" />
            )}
            {label}
          </span>
        ))}
      </div>
      <p className="text-sm leading-6 text-muted-foreground">
        Add these records where you manage DNS for{" "}
        <strong className="text-foreground">{domain.name}</strong>. Copy each
        value exactly. For CNAME records, choose “DNS only” if your provider
        offers proxying.
      </p>
      <section className="space-y-3" aria-label="Domain ownership">
        <h3 className="text-sm font-semibold">1. Verify domain ownership</h3>
        {domain.ownership && (
          <p className="text-sm text-emerald-700 dark:text-emerald-300">
            Ownership confirmed. Keep this TXT record in place.
          </p>
        )}
        {domain.records
          .filter((record) => record.name === `_xem.${domain.name}`)
          .map((record) => (
            <DNSRecordCard key={record.name + record.type} record={record} />
          ))}
      </section>
      <section className="space-y-3" aria-label="Delivery DNS records">
        <h3 className="text-sm font-semibold">
          2. Add and verify delivery records
        </h3>
        {!domain.ownership ? (
          <p className="rounded-xl bg-muted p-4 text-sm">
            Add the ownership TXT record above, then select “Check my records”.
            Once ownership and managed sending approval are confirmed, your DKIM
            and bounce-domain records appear here.
          </p>
        ) : domain.identityStatus === "AWAITING_APPROVAL" ? (
          <p
            role="status"
            className="rounded-xl bg-amber-50 p-4 text-sm leading-6 text-amber-950 dark:bg-amber-950 dark:text-amber-100"
          >
            Ownership is confirmed. Contact your Xem operator to approve managed
            sending for this workspace. After approval, select “Check my
            records” to generate your delivery records. Stay on this step until
            DKIM, the bounce domain, and DMARC are verified.
          </p>
        ) : !domain.provisioned ? (
          <p className="rounded-xl bg-muted p-4 text-sm">
            Ownership is confirmed. Select “Check my records” to prepare your
            delivery records.
          </p>
        ) : (
          <p className="text-sm leading-6 text-muted-foreground">
            Add these records alongside your ownership record, then check again.
            Sender setup unlocks after all DNS checks pass.
          </p>
        )}
        {domain.records
          .filter((record) => record.name !== `_xem.${domain.name}`)
          .map((record) => (
            <DNSRecordCard key={record.name + record.type} record={record} />
          ))}
      </section>
      {domain.provisioned && domain.dmarcStatus !== "VALID" && (
        <div className="rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-950 dark:border-amber-900 dark:bg-amber-950 dark:text-amber-100">
          <strong>Check your DMARC policy</strong>
          <p className="mt-1 leading-6">
            Look for a TXT record at{" "}
            <code className="break-all">_dmarc.{domain.name}</code>. If there
            isn’t one, a starting policy is <code>v=DMARC1; p=none</code>. Keep
            an existing policy; ask your domain administrator before changing
            it.
          </p>
        </div>
      )}
      <div className="flex flex-wrap items-center gap-3">
        <Button onClick={check} disabled={busy}>
          {busy ? (
            <Loader2
              size={16}
              className="mr-2 animate-spin motion-reduce:animate-none"
            />
          ) : (
            <RefreshCw size={16} className="mr-2" />
          )}
          {busy ? "Checking DNS…" : "Check my records"}
        </Button>
        <span role="status" className="text-xs text-muted-foreground">
          {domain.ready
            ? "Everything looks good."
            : domain.identityStatus === "AWAITING_APPROVAL"
              ? "Waiting for managed sending approval."
              : "DNS can take a little time. You can come back later."}
        </span>
      </div>
      {error && (
        <p role="alert" className="text-sm text-destructive">
          {error}
        </p>
      )}
      <details className="rounded-xl bg-muted/60 p-4 text-sm">
        <summary className="cursor-pointer font-medium">
          <CircleHelp size={15} className="mr-2 inline" />A little help with DNS
        </summary>
        <div className="mt-3 space-y-3 text-muted-foreground">
          <p>
            Some providers add your domain to the record name automatically.
            Avoid entering it twice. For MX, put <strong>10</strong> in the
            priority field and the remaining hostname in the mail-server field.
          </p>
          <p>
            Only add the bounce subdomain’s MX record. Keep the MX records that
            receive your regular email.
          </p>
          <p>
            <a
              href="https://developers.cloudflare.com/dns/manage-dns-records/how-to/create-dns-records/"
              target="_blank"
              rel="noreferrer"
              className="underline"
            >
              Cloudflare DNS guide <ArrowUpRight size={12} className="inline" />
            </a>
          </p>
        </div>
      </details>
    </div>
  );
}
function DNSRecordCard({
  record,
}: {
  record: SendingDomain["records"][number];
}) {
  return (
    <div className="rounded-xl border bg-background p-4">
      <div className="mb-2 flex items-center justify-between gap-2">
        <span className="rounded bg-muted px-2 py-1 text-xs font-semibold">
          {record.type}
        </span>
        <CopyButton value={record.value} label="Copy value" />
      </div>
      <p className="text-[11px] uppercase tracking-wider text-muted-foreground">
        Name
      </p>
      <code className="my-1 block break-all text-xs">{record.name}</code>
      <p className="mt-3 text-[11px] uppercase tracking-wider text-muted-foreground">
        Value
      </p>
      <code className="mt-1 block break-all text-xs">{record.value}</code>
    </div>
  );
}

export function SenderSetup({
  domain,
  approved,
  onDone,
}: {
  domain: SendingDomain;
  approved: boolean;
  onDone: () => void | Promise<void>;
}) {
  const { request } = useMarketing();
  const [from, setFrom] = useState(domain.fromEmail || `hello@${domain.name}`);
  const [busy, setBusy] = useState("");
  const [error, setError] = useState("");
  async function act(kind: "sender" | "test") {
    setBusy(kind);
    setError("");
    try {
      await request(
        kind === "sender"
          ? `sending/domains/${domain.id}/sender`
          : "sending/test",
        "POST",
        { from, domainId: domain.id, key: crypto.randomUUID() },
      );
      await onDone();
      toast.success(
        kind === "sender"
          ? "Your sender is ready to use in campaigns."
          : "Your test is queued. Follow its progress in Sending.",
      );
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy("");
    }
  }
  return (
    <div className="space-y-5">
      <div className="space-y-2">
        <Label htmlFor={`sender-${domain.id}`}>
          The address people will see
        </Label>
        <Input
          id={`sender-${domain.id}`}
          type="email"
          value={from}
          onChange={(e) => setFrom(e.target.value)}
          className="h-12"
        />
        <p className="text-sm text-muted-foreground">
          Use an address at {domain.name}. Replies should go to an inbox you can
          read.
        </p>
      </div>
      <div className="flex flex-wrap gap-3">
        <Button
          onClick={() => void act("sender")}
          disabled={!!busy || !domain.ready}
        >
          {busy === "sender"
            ? "Saving…"
            : domain.smtpConfigId
              ? "Update sender"
              : "Save sender"}
        </Button>
        <Button
          variant="outline"
          onClick={() => void act("test")}
          disabled={!!busy || !domain.ready || !approved}
        >
          <Mail size={16} className="mr-2" />
          {busy === "test" ? "Queueing…" : "Send me a test"}
        </Button>
      </div>
      <p className="text-xs text-muted-foreground">
        The test goes only to your signed-in account’s email address. Provider
        acceptance is shown separately from confirmed delivery.
      </p>
      {!approved && (
        <p className="rounded-xl bg-muted p-4 text-sm">
          Managed sending is awaiting operator approval, paused, or suspended.
          You can finish your domain setup now.
        </p>
      )}
      {error && (
        <p role="alert" className="text-sm text-destructive">
          {error}
        </p>
      )}
    </div>
  );
}
