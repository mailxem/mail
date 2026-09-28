"use client";
import { useConfirmSheet } from "@/components/ui/confirm-sheet";
import { useRef, useState } from "react";
import Link from "next/link";
import {
  Check,
  KeyRound,
  Mail,
  Plus,
  RefreshCw,
  ShieldCheck,
  ArrowRight,
  Pause,
  Play,
  ChevronDown,
} from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useMarketing, useMarketingQuery } from "@/lib/marketing/api";
import type { SendingState, SendingCredential } from "@/lib/sending/types";
import { CopyButton, DomainDNS, DomainForm, SenderSetup } from "./controls";

export function SendingDashboard() {
  const { scope } = useMarketing();
  return <DashboardContent key={scope || "signed-out"} />;
}
function DashboardContent() {
  const confirm = useConfirmSheet();
  const credentialTrigger = useRef<HTMLButtonElement>(null);
  const q = useMarketingQuery<SendingState>("sending", true, 15000);
  const { request, refresh } = useMarketing();
  const [add, setAdd] = useState(false);
  const [busy, setBusy] = useState("");
  const [name, setName] = useState("");
  const [domainID, setDomainID] = useState("");
  const [secret, setSecret] = useState<{
    credential: SendingCredential;
    password: string;
  } | null>(null);
  const [events, setEvents] = useState<
    { id: string; kind: string; detail: string; createdAt: string }[] | null
  >(null);
  const sync = () => {
    void q.refetch();
    void refresh();
  };
  async function action(path: string, method: string, body?: unknown) {
    setBusy(path);
    try {
      await request(path, method, body);
      sync();
      toast.success("Sending settings updated.");
    } catch (e) {
      toast.error((e as Error).message);
    } finally {
      setBusy("");
    }
  }
  async function create(e: React.FormEvent) {
    e.preventDefault();
    setBusy("credential");
    try {
      const result = await request<{
        credential: SendingCredential;
        password: string;
      }>("sending/credentials", "POST", {
        name,
        domainId: domainID || q.data?.domains.find((d) => d.ready)?.id,
      });
      setSecret(result);
      setName("");
      sync();
    } catch (e) {
      toast.error((e as Error).message);
    } finally {
      setBusy("");
    }
  }
  async function showEvents(id: string) {
    setBusy(id);
    try {
      setEvents(await request(`sending/messages/${id}/events`));
    } catch (e) {
      toast.error((e as Error).message);
    } finally {
      setBusy("");
    }
  }
  if (q.isLoading)
    return (
      <p role="status" className="p-8 text-muted-foreground">
        Loading your sending workspace…
      </p>
    );
  if (q.error || !q.data)
    return (
      <div className="space-y-4 rounded-xl border p-8">
        <h1 className="text-2xl font-medium">Managed sending</h1>
        <p role="alert">{q.error?.message || "Sign in to continue."}</p>
        <Button variant="outline" onClick={() => void q.refetch()}>
          Try again
        </Button>
      </div>
    );
  const state = q.data;
  const eligible =
    state.enabled &&
    state.account.approved &&
    !state.account.suspended &&
    !state.account.paused;
  const ready = state.domains.filter((d) => d.ready);
  return (
    <div className="mx-auto max-w-6xl space-y-7 pb-10">
      <header className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <p className="mb-2 text-xs text-muted-foreground">
            Workspace settings
          </p>
          <h1 className="text-2xl font-medium tracking-tight">
            Managed sending
          </h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Domains, credentials, and the journey of every email.
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={() => void q.refetch()}
            aria-label="Refresh sending status"
          >
            <RefreshCw size={16} />
          </Button>
          <Button variant="outline" asChild>
            <Link href="/onboarding">
              Setup checklist <ArrowRight size={15} className="ml-2" />
            </Link>
          </Button>
        </div>
      </header>
      {!state.enabled ? (
        <div className="rounded-xl border bg-muted/40 p-6">
          <h2 className="font-medium">
            Managed sending isn’t enabled here yet.
          </h2>
          <p className="mt-2 text-sm leading-6 text-muted-foreground">
            Your installation operator can enable the SES integration. You can
            keep sending with your existing provider.
          </p>
          <Button asChild variant="outline" className="mt-4">
            <Link href="/settings/smtp">Connect my own provider</Link>
          </Button>
        </div>
      ) : (
        <>
          <section className="flex flex-wrap items-center justify-between gap-4 rounded-xl border bg-card p-5">
            <div className="flex items-center gap-3">
              <span
                className={`rounded-lg p-2 ${eligible ? "bg-emerald-500/10 text-emerald-500" : "bg-amber-500/10 text-amber-500"}`}
              >
                <ShieldCheck size={20} />
              </span>
              <div>
                <h2 className="text-sm font-medium">
                  {state.account.suspended
                    ? "Sending suspended"
                    : !state.account.approved
                      ? "Awaiting operator approval"
                      : state.account.paused
                        ? "Sending paused"
                        : "Your sending account is active"}
                </h2>
                <p className="mt-1 text-xs text-muted-foreground">
                  {eligible
                    ? "Verified domains can send within your workspace limits."
                    : "You can prepare your domains while sending is unavailable."}
                </p>
              </div>
            </div>
            <Button
              variant="outline"
              disabled={
                !!busy || !state.account.approved || state.account.suspended
              }
              onClick={() =>
                void action("sending/pause", "PUT", {
                  paused: !state.account.paused,
                })
              }
            >
              {state.account.paused ? (
                <Play size={14} className="mr-2" />
              ) : (
                <Pause size={14} className="mr-2" />
              )}
              {state.account.paused ? "Resume sending" : "Pause sending"}
            </Button>
          </section>
          <div className="grid divide-y divide-border border-y border-border sm:grid-cols-3 sm:divide-x sm:divide-y-0">
            {[
              {
                label: "Today’s recipient sends",
                used: state.account.dailyUsed,
                limit: state.account.dailyLimit,
              },
              {
                label: "This month’s recipient sends",
                used: state.account.monthlyUsed,
                limit: state.account.monthlyLimit,
              },
              {
                label: "Reserved delivery allowance",
                used: state.account.budgetUsedMicros / 1e6,
                limit: state.account.monthlyBudgetMicros / 1e6,
                money: true,
              },
            ].map((metric) => (
              <div
                key={metric.label}
                className="min-w-0 px-5 py-5 sm:first:pl-0 sm:last:pr-0"
              >
                <p className="text-xs text-muted-foreground">{metric.label}</p>
                <p className="mt-3 text-2xl font-medium tabular-nums">
                  {metric.money ? "$" : ""}
                  {metric.used.toLocaleString()}
                  <span className="ml-2 text-sm font-normal text-muted-foreground">
                    / {metric.money ? "$" : ""}
                    {metric.limit.toLocaleString()}
                  </span>
                </p>
                <div className="mt-4 h-1.5 overflow-hidden rounded-full bg-muted">
                  <div
                    className="h-full rounded-full bg-primary"
                    style={{
                      width: `${Math.min(100, metric.limit ? (metric.used / metric.limit) * 100 : 0)}%`,
                    }}
                  />
                </div>
              </div>
            ))}
          </div>
          <p className="-mt-4 text-xs leading-5 text-muted-foreground">
            Limits count recipients when a message is accepted into the queue,
            including tests. Reserved allowance is a conservative sending cap,
            not an invoice. Failed or suppressed queued messages keep their
            reservation.
          </p>
          <section className="space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-sm font-medium">Sending domains</h2>
              <Button size="sm" variant="outline" onClick={() => setAdd(!add)}>
                <Plus size={15} className="mr-1" />
                Add domain
              </Button>
            </div>
            {(add || !state.domains.length) && (
              <div className="rounded-xl border border-border bg-card p-5">
                <DomainForm
                  onDone={() => {
                    setAdd(false);
                    sync();
                  }}
                />
              </div>
            )}
            {state.domains.map((domain) => (
              <details
                key={domain.id}
                className="rounded-xl border border-border bg-card"
              >
                <summary className="flex cursor-pointer list-none items-center justify-between gap-3 p-5">
                  <div className="flex min-w-0 items-center gap-3">
                    <Mail size={20} className="text-muted-foreground" />
                    <div>
                      <span className="break-all text-sm font-medium">
                        {domain.name}
                      </span>
                      <p className="mt-1 text-xs text-muted-foreground">
                        {domain.fromEmail || "Set up your sender address"}
                      </p>
                    </div>
                  </div>
                  <span className="flex items-center gap-3 text-xs text-muted-foreground">
                    {domain.ready ? (
                      <>
                        <Check size={14} className="text-emerald-600" />
                        Ready
                      </>
                    ) : (
                      "Setup in progress"
                    )}
                    <ChevronDown size={16} />
                  </span>
                </summary>
                <div className="grid gap-8 border-t p-5 lg:grid-cols-2">
                  <DomainDNS domain={domain} onDone={sync} />
                  <div className="space-y-8">
                    <SenderSetup
                      key={domain.id}
                      domain={domain}
                      approved={eligible}
                      onDone={sync}
                    />
                    <div className="border-t pt-5">
                      <p className="mb-3 text-xs leading-5 text-muted-foreground">
                        Disconnecting revokes this domain’s SMTP credentials and
                        blocks queued mail. Domain history is retained.
                      </p>
                      <Button
                        variant="outline"
                        size="sm"
                        disabled={!!busy}
                        onClick={async () => {
                          if (
                            await confirm({
                              title: `Disconnect ${domain.name}?`,
                              description:
                                "The domain’s credentials will be revoked.",
                              confirmLabel: "Disconnect domain",
                              variant: "destructive",
                            })
                          )
                            void action(
                              `sending/domains/${domain.id}`,
                              "DELETE",
                            );
                        }}
                      >
                        Disconnect domain
                      </Button>
                    </div>
                  </div>
                </div>
              </details>
            ))}
          </section>
          <section className="rounded-xl border border-border bg-card p-5">
            <div className="mb-5 flex items-center gap-3">
              <KeyRound size={20} className="text-muted-foreground" />
              <div>
                <h2 className="text-sm font-medium">SMTP credentials</h2>
                <p className="mt-1 text-xs text-muted-foreground">
                  Connect external apps. Each credential belongs to one domain
                  and expires after 90 days.
                </p>
              </div>
            </div>
            {state.smtp.enabled ? (
              <>
                <div className="mb-6 flex flex-wrap items-center justify-between gap-3 rounded-xl bg-muted p-4">
                  <code className="break-all text-xs">
                    {state.smtp.host}:{state.smtp.port} · {state.smtp.security}
                  </code>
                  <CopyButton value={state.smtp.host} label="Copy host" />
                </div>
                <form
                  onSubmit={create}
                  className="mb-6 grid items-end gap-3 sm:grid-cols-[1fr_1fr_auto]"
                >
                  <div className="space-y-2">
                    <Label htmlFor="credential-name">
                      App or credential name
                    </Label>
                    <Input
                      id="credential-name"
                      value={name}
                      onChange={(e) => setName(e.target.value)}
                      maxLength={80}
                      placeholder="My WordPress site"
                      required
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="credential-domain">Sending domain</Label>
                    <select
                      id="credential-domain"
                      className="h-9 w-full rounded-lg border border-border bg-background px-3 text-sm"
                      value={domainID || ready[0]?.id || ""}
                      onChange={(e) => setDomainID(e.target.value)}
                      required
                    >
                      {!ready.length && (
                        <option value="">Verify a domain first</option>
                      )}
                      {ready.map((d) => (
                        <option key={d.id} value={d.id}>
                          {d.name}
                        </option>
                      ))}
                    </select>
                  </div>
                  <Button
                    disabled={!!busy || !eligible || !ready.length}
                    type="submit"
                    ref={credentialTrigger}
                  >
                    {busy === "credential" ? "Creating…" : "Create credential"}
                  </Button>
                </form>
                <div className="divide-y">
                  {state.credentials.length ? (
                    state.credentials.map((cred) => (
                      <div
                        key={cred.id}
                        className="flex flex-wrap items-center justify-between gap-3 py-4"
                      >
                        <div>
                          <p className="text-sm font-medium">{cred.name}</p>
                          <p className="mt-1 text-xs text-muted-foreground">
                            {cred.revokedAt
                              ? "Revoked"
                              : `Expires ${new Date(cred.expiresAt).toLocaleDateString()}`}{" "}
                            ·{" "}
                            {state.domains.find((d) => d.id === cred.domainId)
                              ?.name || "Disconnected domain"}
                          </p>
                        </div>
                        {!cred.revokedAt && (
                          <Button
                            size="sm"
                            variant="ghost"
                            disabled={!!busy}
                            onClick={async () => {
                              if (
                                await confirm({
                                  title: `Revoke ${cred.name}?`,
                                  description:
                                    "Apps using it will stop sending.",
                                  confirmLabel: "Revoke credential",
                                  variant: "destructive",
                                })
                              )
                                void action(
                                  `sending/credentials/${cred.id}`,
                                  "DELETE",
                                );
                            }}
                          >
                            Revoke
                          </Button>
                        )}
                      </div>
                    ))
                  ) : (
                    <p className="text-sm text-muted-foreground">
                      No external apps connected yet. Sending inside Xem doesn’t
                      need SMTP credentials.
                    </p>
                  )}
                </div>
              </>
            ) : (
              <p className="text-sm leading-6 text-muted-foreground">
                The external SMTP listener is not enabled on this installation.
                Verified managed senders can still be used inside Xem.
              </p>
            )}
          </section>
        </>
      )}
      <section className="rounded-xl border border-border bg-card p-5">
        <div className="mb-5 flex items-center justify-between">
          <h2 className="text-sm font-medium">Recent delivery activity</h2>
          <span className="text-xs text-muted-foreground">
            Latest 50 messages
          </span>
        </div>
        {!state.messages.length ? (
          <div className="rounded-xl border border-dashed p-8 text-center">
            <Mail size={25} className="mx-auto mb-3 text-muted-foreground" />
            <p className="text-sm font-medium">
              Your first hello will appear here.
            </p>
            <p className="mt-2 text-xs text-muted-foreground">
              Send a test from a verified domain to follow its progress.
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full min-w-[600px] text-left text-sm">
              <thead className="border-b text-xs text-muted-foreground">
                <tr>
                  <th className="pb-3 font-medium">Message</th>
                  <th className="pb-3 font-medium">Status</th>
                  <th className="pb-3 font-medium">What happened</th>
                  <th className="pb-3 font-medium">History</th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {state.messages.map((m) => (
                  <tr key={m.id}>
                    <td className="max-w-[220px] py-4 pr-4">
                      <p className="truncate font-medium">
                        {m.subject || "No subject"}
                      </p>
                      <p className="mt-1 truncate text-xs text-muted-foreground">
                        {m.recipients}
                      </p>
                    </td>
                    <td className="pr-4 text-xs">
                      {m.status.replaceAll("_", " ")}
                    </td>
                    <td className="max-w-xs pr-4 text-xs leading-5 text-muted-foreground">
                      {m.detail || "Durably queued for delivery."}
                    </td>
                    <td>
                      <Button
                        size="sm"
                        variant="ghost"
                        disabled={!!busy}
                        onClick={() => void showEvents(m.id)}
                      >
                        View events
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
      <Dialog
        open={!!secret}
        onOpenChange={(open) => {
          if (!open) setSecret(null);
        }}
      >
        <DialogContent
          className="max-w-lg"
          onCloseAutoFocus={(event) => {
            event.preventDefault();
            credentialTrigger.current?.focus();
          }}
        >
          <DialogHeader>
            <DialogTitle>Your app’s new sending key</DialogTitle>
            <DialogDescription>
              Copy the password now. We store only its hash, so it cannot be
              shown again.
            </DialogDescription>
          </DialogHeader>
          {secret && (
            <div className="space-y-4">
              <div className="rounded-xl bg-muted p-4">
                <p className="text-xs text-muted-foreground">Username</p>
                <code className="mt-1 block break-all text-xs">
                  {secret.credential.id}
                </code>
                <CopyButton
                  value={secret.credential.id}
                  label="Copy username"
                />
              </div>
              <div className="rounded-xl bg-muted p-4">
                <p className="text-xs text-muted-foreground">
                  Password · shown once
                </p>
                <code className="mt-1 block break-all text-xs">
                  {secret.password}
                </code>
                <CopyButton value={secret.password} label="Copy password" />
              </div>
              <p className="text-xs leading-5 text-muted-foreground">
                Use {state.smtp.host}, port {state.smtp.port}, and STARTTLS.
                Keep the password in your application’s secret store.
              </p>
              <Button className="w-full" onClick={() => setSecret(null)}>
                I’ve saved it securely
              </Button>
            </div>
          )}
        </DialogContent>
      </Dialog>
      <Dialog
        open={events !== null}
        onOpenChange={(open) => {
          if (!open) setEvents(null);
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delivery events</DialogTitle>
            <DialogDescription>
              Authenticated provider feedback for this message.
            </DialogDescription>
          </DialogHeader>
          <div className="max-h-80 space-y-4 overflow-y-auto">
            {events?.length ? (
              events.map((e) => (
                <div key={e.id} className="border-l-2 border-border pl-4">
                  <p className="text-sm font-medium">{e.kind}</p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {e.detail} · {new Date(e.createdAt).toLocaleString()}
                  </p>
                </div>
              ))
            ) : (
              <p className="text-sm text-muted-foreground">
                No provider events yet. Refresh shortly after sending.
              </p>
            )}
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
