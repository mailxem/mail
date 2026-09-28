"use client";

import { useState } from "react";
import { Cloud, Mail, Unplug } from "lucide-react";
import { useMarketing, useMarketingQuery } from "@/lib/marketing/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { toast } from "sonner";

type Connection = {
  id: string;
  provider: string;
  address: string;
  smtpConfigId: string;
  imapConfigId?: string;
};

export function MailConnections({
  provider,
}: {
  provider: "google" | "cloudflare";
}) {
  const { request, refresh } = useMarketing();
  const result = useMarketingQuery<{
    connections: Connection[];
    googleConfigured: boolean;
  }>("mail-connections");
  const [busy, setBusy] = useState(false);
  const [address, setAddress] = useState("");
  const [accountId, setAccountId] = useState("");
  const [token, setToken] = useState("");
  const [disconnecting, setDisconnecting] = useState<string | null>(null);
  const google = provider === "google";
  const connections =
    result.data?.connections.filter(
      (c) => c.provider === (google ? "GOOGLE_OAUTH" : "CLOUDFLARE"),
    ) ?? [];
  async function run(action: () => Promise<void>) {
    if (busy) return;
    setBusy(true);
    try {
      await action();
    } catch (error) {
      toast.error((error as Error).message);
    } finally {
      setBusy(false);
    }
  }
  return (
    <section
      id={`connected-mail-${provider}`}
      className="rounded-xl border bg-card p-5 space-y-4 mb-6"
    >
      <div className="flex gap-3 items-start">
        {google ? (
          <Mail className="size-5 mt-1" />
        ) : (
          <Cloud className="size-5 mt-1" />
        )}
        <div>
          <h2 className="font-semibold">
            {google ? "Gmail & Google Workspace" : "Cloudflare Email Sending"}
          </h2>
          <p className="text-sm text-muted-foreground mt-1">
            {google
              ? "Connect a mailbox for inbox access and sending. This grants your workspace access to its mail; connect a shared business mailbox you intend your team to use."
              : "Send transactional emails through your Cloudflare account. Your domain must have Email Sending enabled. Campaigns and newsletters need a different sender."}
          </p>
        </div>
      </div>
      {result.error ? (
        <p role="alert" className="text-sm text-muted-foreground">
          {result.error.message}
        </p>
      ) : (
        <>
          {connections.map((connection) => (
            <div
              key={connection.id}
              className="flex flex-wrap items-center justify-between gap-3 border-t pt-3"
            >
              <div>
                <p className="text-sm font-medium break-all">
                  {connection.address}
                </p>
                <p className="text-xs text-muted-foreground">
                  {google
                    ? "Mailbox and sender connected"
                    : "Sender saved · delivery setup is checked when you send"}
                </p>
              </div>
              {disconnecting === connection.id ? (
                <div className="flex flex-wrap items-center gap-2">
                  <p className="text-xs">Stop inbox access and future sends?</p>
                  <Button
                    size="sm"
                    variant="destructive"
                    disabled={busy}
                    onClick={() =>
                      run(async () => {
                        await request(
                          `mail-connections/${connection.id}`,
                          "DELETE",
                        );
                        setDisconnecting(null);
                        await refresh();
                        toast.success(
                          "Disconnected. You can also revoke access in your provider account.",
                        );
                      })
                    }
                  >
                    Disconnect
                  </Button>
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => setDisconnecting(null)}
                  >
                    Cancel
                  </Button>
                </div>
              ) : (
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => setDisconnecting(connection.id)}
                >
                  <Unplug className="size-4" />
                  Disconnect
                </Button>
              )}
            </div>
          ))}
          {google ? (
            <div className="space-y-2">
              <Button
                disabled={busy || !result.data?.googleConfigured}
                onClick={() =>
                  run(async () => {
                    const response = await request<{ url: string }>(
                      "mail-connections/google/start",
                      "POST",
                    );
                    window.location.assign(response.url);
                  })
                }
              >
                {busy ? "Opening Google…" : "Connect Google mailbox"}
              </Button>
              {result.data && !result.data.googleConfigured && (
                <p className="text-xs text-muted-foreground">
                  Your administrator needs to configure Google mailbox OAuth on
                  this installation. Google sign-in uses separate credentials.
                </p>
              )}
              <p className="text-xs text-muted-foreground">
                Requires Google’s mail scope for IMAP/SMTP. Workspace
                administrators may need to allow access. Xem stores tokens
                encrypted and never asks for your Google password.
              </p>
            </div>
          ) : (
            <form
              className="grid gap-4 sm:grid-cols-2"
              onSubmit={(e) => {
                e.preventDefault();
                void run(async () => {
                  await request("mail-connections/cloudflare", "POST", {
                    address,
                    accountId,
                    token,
                  });
                  setToken("");
                  setAddress("");
                  await refresh();
                  toast.success("Cloudflare sender saved");
                });
              }}
            >
              <div className="space-y-2">
                <Label htmlFor="cf-sender">Sender address</Label>
                <Input
                  id="cf-sender"
                  required
                  type="email"
                  value={address}
                  onChange={(e) => setAddress(e.target.value)}
                  placeholder="notifications@yourdomain.com"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="cf-account">Cloudflare account ID</Label>
                <Input
                  id="cf-account"
                  required
                  pattern="[a-fA-F0-9]{32}"
                  maxLength={32}
                  value={accountId}
                  onChange={(e) => setAccountId(e.target.value)}
                  autoComplete="off"
                />
              </div>
              <div className="space-y-2 sm:col-span-2">
                <Label htmlFor="cf-token">Email Sending API token</Label>
                <Input
                  id="cf-token"
                  required
                  type="password"
                  minLength={16}
                  maxLength={4096}
                  value={token}
                  onChange={(e) => setToken(e.target.value)}
                  autoComplete="new-password"
                />
                <p className="text-xs text-muted-foreground">
                  Use a token scoped to Email Sending for this account. Sending
                  to arbitrary recipients requires Workers Paid; provider
                  charges are separate from Xem.
                </p>
              </div>
              <div>
                <Button type="submit" disabled={busy || !result.data}>
                  {busy ? "Saving…" : "Save Cloudflare sender"}
                </Button>
              </div>
            </form>
          )}
        </>
      )}
    </section>
  );
}
