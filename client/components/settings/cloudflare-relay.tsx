"use client";
import { useState } from "react";
import {
  Check,
  Database,
  KeyRound,
  RotateCw,
  ShieldOff,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useConfirmSheet } from "@/components/ui/confirm-sheet";
import { useMarketing, useMarketingQuery } from "@/lib/marketing/api";
import { toast } from "sonner";

type Relay = {
  id: string;
  mailboxId: string;
  address: string;
  workerUrl: string;
  enabled: boolean;
};

/** Connected Cloudflare inboxes; new ones are added from the setup sheet. */
export function CloudflareRelays() {
  const result = useMarketingQuery<{ relays: Relay[] }>(
    "mail-connections/cloudflare-relays",
  );
  if (result.error)
    return (
      <p role="alert" className="text-sm text-muted-foreground">
        Couldn’t load Cloudflare inboxes.{" "}
        <button className="underline" onClick={() => result.refetch()}>
          Retry
        </button>
      </p>
    );
  return (
    <>
      {result.data?.relays.map((relay) => (
        <WorkerCard key={relay.id} relay={relay} />
      ))}
    </>
  );
}

function WorkerCard({ relay }: { relay: Relay }) {
  const { request, refresh } = useMarketing();
  const confirm = useConfirmSheet();
  const [secret, setSecret] = useState("");
  const [changing, setChanging] = useState(false);
  const [busy, setBusy] = useState(false);
  const replaceSecret = async () => {
    if (!secret || busy) return;
    setBusy(true);
    try {
      await request(
        `mail-connections/cloudflare-relays/${relay.id}/secret`,
        "POST",
        { secret },
      );
      setSecret("");
      setChanging(false);
      await refresh();
      toast.success(
        relay.enabled
          ? "Mailbox API secret updated"
          : "Cloudflare mailbox enabled",
      );
    } catch (error) {
      toast.error((error as Error).message);
    } finally {
      setBusy(false);
    }
  };
  const disable = async () => {
    if (
      !(await confirm({
        title: "Disconnect Cloudflare mailbox?",
        description:
          "Xem will stop reading this mailbox. Mail remains in your Cloudflare D1 and R2 storage.",
        confirmLabel: "Disconnect mailbox",
        variant: "destructive",
      }))
    )
      return;
    setBusy(true);
    try {
      await request(`mail-connections/cloudflare-relays/${relay.id}`, "DELETE");
      await refresh();
      toast.success("Cloudflare mailbox disconnected");
    } catch (error) {
      toast.error((error as Error).message);
    } finally {
      setBusy(false);
    }
  };
  return (
    <div className="space-y-3 rounded-lg border bg-card p-3">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="flex min-w-0 gap-2.5">
          <Database className="mt-0.5 size-4 shrink-0" />
          <div className="min-w-0">
            <p className="break-all text-sm font-medium">{relay.address}</p>
            <p className="mt-1 break-all text-xs text-muted-foreground">
              {relay.workerUrl}
            </p>
            <p className="mt-1 text-xs text-muted-foreground">
              Mail is stored in your Cloudflare account · sending is configured
              separately.
            </p>
          </div>
        </div>
        <span className="inline-flex items-center gap-1.5 rounded-full border px-2 py-1 text-[10px]">
          {relay.enabled ? (
            <>
              <Check className="size-3" /> Connected
            </>
          ) : (
            <>
              <ShieldOff className="size-3" /> Disconnected
            </>
          )}
        </span>
      </div>
      {changing && (
        <div className="space-y-2 rounded-lg border bg-muted/30 p-3">
          <Label htmlFor={`worker-secret-${relay.id}`}>
            New mailbox API secret
          </Label>
          <Input
            id={`worker-secret-${relay.id}`}
            type="password"
            minLength={32}
            maxLength={4096}
            value={secret}
            onChange={(event) => setSecret(event.target.value)}
            autoComplete="new-password"
          />
          <p className="text-xs text-muted-foreground">
            Rotate or create the secret in Cloudflare first, then enter the
            replacement here. Xem verifies it before saving.
          </p>
          <div className="flex gap-2">
            <Button
              size="sm"
              disabled={busy || secret.length < 32}
              onClick={() => void replaceSecret()}
            >
              <KeyRound className="size-3.5" /> Verify and save
            </Button>
            <Button
              size="sm"
              variant="ghost"
              onClick={() => {
                setChanging(false);
                setSecret("");
              }}
            >
              Cancel
            </Button>
          </div>
        </div>
      )}
      {!changing && (
        <div className="flex flex-wrap gap-2">
          <Button size="sm" variant="outline" onClick={() => setChanging(true)}>
            <RotateCw className="size-3.5" />{" "}
            {relay.enabled ? "Replace API secret" : "Reconnect"}
          </Button>
          {relay.enabled && (
            <Button
              size="sm"
              variant="ghost"
              disabled={busy}
              onClick={() => void disable()}
            >
              Disconnect
            </Button>
          )}
        </div>
      )}
    </div>
  );
}
