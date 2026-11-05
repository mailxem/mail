"use client";

import { useState } from "react";
import { Mail, Unplug } from "lucide-react";
import { useMarketing, useMarketingQuery } from "@/lib/marketing/api";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";
import { CloudflareRelays } from "./cloudflare-relay";
import { CloudflareButton, CloudflareLogo, CloudflareSetupSheet } from "./cloudflare-setup";

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
    googleAvailable: boolean;
  }>("mail-connections");
  const [busy, setBusy] = useState(false);
  const [disconnecting, setDisconnecting] = useState<string | null>(null);
  const google = provider === "google";
  const [receiving, setReceiving] = useState(false);
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
          <Mail className="mt-1 size-5 shrink-0" />
        ) : (
          <span className="mt-0.5 grid size-7 shrink-0 place-items-center rounded-lg border bg-[#F38020]/10">
            <CloudflareLogo className="h-3" />
          </span>
        )}
        <div>
          <h2 className="font-semibold">
            {google ? "Gmail & Google Workspace" : "Cloudflare email"}
          </h2>
          <p className="text-sm text-muted-foreground mt-1">
            {google
              ? "Send emails and replies through Gmail SMTP, and browse, search and manage mail through the Gmail API. This grants your workspace access to the connected mailbox; connect only an account you are authorized to share with your team."
              : "Receive mail through Email Routing without IMAP, or configure Cloudflare Email Sending for transactional outbound mail."}
          </p>
        </div>
      </div>
      {result.error ? (
        <div role="alert" className="space-y-2">
          <p className="text-sm text-muted-foreground">
            Couldn’t load mailbox connections. {result.error.message}
          </p>
          <Button size="sm" variant="outline" onClick={() => result.refetch()}>
            Retry
          </Button>
        </div>
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
                  <p className="max-w-sm text-xs">
                    {google
                      ? "Disconnect this inbox and sender? Saved messages stay in Xem. Mail already being sent may still arrive."
                      : "Disconnect this sender? Saved messages stay in Xem. Mail already being sent may still arrive."}
                  </p>
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
                disabled={
                  busy ||
                  !result.data?.googleConfigured ||
                  !result.data?.googleAvailable
                }
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
                {busy
                  ? "Opening Google…"
                  : connections.length
                    ? "Connect another Google mailbox"
                    : "Connect Google mailbox"}
              </Button>
              {result.data && !result.data.googleConfigured && (
                <p className="text-xs text-muted-foreground">
                  Your administrator needs to configure Google mailbox OAuth on
                  this installation. Google sign-in alone does not configure
                  mailbox access.
                </p>
              )}
              {result.data?.googleConfigured &&
                !result.data.googleAvailable && (
                  <p className="text-xs text-muted-foreground">
                    Google mailbox connections are available to approved testers
                    while verification is in progress. Your administrator can
                    manage access.
                  </p>
                )}
              <p className="text-xs text-muted-foreground">
                You’ll continue on Google to choose an account and approve
                mailbox access, then return to Xem. No email is sent when you
                connect.
              </p>
              <details className="rounded-lg border px-3 py-2 text-xs text-muted-foreground">
                <summary className="cursor-pointer font-medium text-foreground">
                  Privacy and access details
                </summary>
                <p className="mt-2 leading-5">
                  Requires Google’s Gmail read and update access. Workspace
                  administrators may need to allow access. Google sign-in is
                  separate. Xem stores tokens encrypted and never asks for your
                  Google password. Incoming mail is fetched on demand; outgoing
                  messages and attachments are saved in Xem. Connecting does not
                  automatically send your inbox to AI. If you choose an AI
                  writing or assistant action, its inputs and relevant workspace
                  data may be sent to the configured AI service.
                </p>
              </details>
            </div>
          ) : (
            <div className="space-y-3 border-t pt-4">
              <CloudflareRelays />
              <div className="flex flex-wrap gap-2">
                <CloudflareButton mode="send">Add Cloudflare sender</CloudflareButton>
                <Button variant="outline" onClick={() => setReceiving(true)}>
                  Connect Cloudflare inbox
                </Button>
              </div>
              <CloudflareSetupSheet mode="receive" open={receiving} onOpenChange={setReceiving} />
            </div>
          )}
        </>
      )}
    </section>
  );
}
