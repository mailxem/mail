"use client";
import { useState, type ReactNode } from "react";
import { ArrowUpRight, Check, Copy, Eye, EyeOff, Loader2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Sheet,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetPanel,
  SheetPopup,
  SheetTitle,
} from "@/components/ui/sheet";
import { useMarketing } from "@/lib/marketing/api";
import { PaidPlanNotice, usePaidPlan } from "@/components/billing/paid-plan";
import { cn } from "@/lib/utils";

export type CloudflareMode = "receive" | "send";

export function CloudflareLogo({ className }: { className?: string }) {
  return (
    <svg viewBox="0 -3 66 33" aria-hidden className={cn("h-4 w-auto", className)}>
      <path fill="#F38020" d="M43.9 29.1l.6-2a5.4 5.4 0 0 0-.3-4.6 4.9 4.9 0 0 0-3.4-2L13 20.1a.5.5 0 0 1-.4-.3.6.6 0 0 1 0-.5.8.8 0 0 1 .6-.5l28.4-.4a10.2 10.2 0 0 0 8.4-6.2l1.6-4.2a1 1 0 0 0 .1-.6A18.5 18.5 0 0 0 16 5.4a8.4 8.4 0 0 0-13.1 8.8A11.8 11.8 0 0 0 0 28.6a.5.5 0 0 0 .5.5h43a.6.6 0 0 0 .4-.5z" />
      <path fill="#FAAE40" d="M51.5 13.5h-.6a.5.5 0 0 0-.4.3l-.9 3.1a5.4 5.4 0 0 0 .3 4.6 4.9 4.9 0 0 0 3.4 2l6 .4a.5.5 0 0 1 .4.3.6.6 0 0 1 0 .5.8.8 0 0 1-.6.5l-6.3.4a10.2 10.2 0 0 0-8.4 6.2l-.3.8a.3.3 0 0 0 .3.5h21.6a.5.5 0 0 0 .5-.4 15.5 15.5 0 0 0 .6-4.2 15.3 15.3 0 0 0-15.6-15z" transform="translate(-1.5 -3)" />
    </svg>
  );
}

/** A Cloudflare-branded button that opens the setup sheet. */
export function CloudflareButton({ mode, children, className }: { mode: CloudflareMode; children: ReactNode; className?: string }) {
  const [open, setOpen] = useState(false);
  return (
    <>
      <Button
        type="button"
        onClick={() => setOpen(true)}
        className={cn("gap-2 border-0 bg-[#F38020] text-white shadow-none hover:bg-[#E06D10] hover:brightness-100 focus-visible:ring-[#F38020]/40", className)}
      >
        <span className="grid size-5 place-items-center rounded bg-white">
          <CloudflareLogo className="h-2.5" />
        </span>
        {children}
      </Button>
      <CloudflareSetupSheet mode={mode} open={open} onOpenChange={setOpen} />
    </>
  );
}

const copy = {
  receive: {
    title: "Connect a Cloudflare inbox",
    description: "Mail for your address lands in your own Cloudflare account. Xem reads it through a signed Worker API, with no IMAP server.",
    submit: "Verify and connect",
    busy: "Verifying Worker…",
  },
  send: {
    title: "Send through Cloudflare",
    description: "Send transactional mail with Cloudflare Email Sending, using a token you create. Cloudflare bills sending separately.",
    submit: "Save sender",
    busy: "Saving…",
  },
} as const;

export function CloudflareSetupSheet({ mode, open, onOpenChange }: { mode: CloudflareMode; open: boolean; onOpenChange: (open: boolean) => void }) {
  const paid = usePaidPlan();
  const { request, refresh } = useMarketing();
  const [fields, setFields] = useState({ workerUrl: "", address: "", secret: "", accountId: "", token: "" });
  const [busy, setBusy] = useState(false);
  const text = copy[mode];
  const set = (key: keyof typeof fields) => (event: React.ChangeEvent<HTMLInputElement>) => setFields((f) => ({ ...f, [key]: event.target.value.trim() }));

  const submit = async (event: React.FormEvent) => {
    event.preventDefault();
    if (busy) return;
    setBusy(true);
    try {
      if (mode === "receive") {
        await request("mail-connections/cloudflare-relays", "POST", { workerUrl: fields.workerUrl, address: fields.address, secret: fields.secret });
      } else {
        await request("mail-connections/cloudflare", "POST", { address: fields.address, accountId: fields.accountId, token: fields.token });
      }
      await refresh();
      toast.success(mode === "receive" ? "Cloudflare inbox verified and connected" : "Cloudflare sender saved");
      setFields({ workerUrl: "", address: "", secret: "", accountId: "", token: "" });
      onOpenChange(false);
    } catch (error) {
      toast.error((error as Error).message);
    } finally {
      setBusy(false);
    }
  };

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetPopup variant="inset" className="max-w-xl">
        <SheetHeader>
          <div className="mb-3 flex size-10 items-center justify-center rounded-xl border bg-[#F38020]/10">
            <CloudflareLogo className="h-4" />
          </div>
          <SheetTitle>{text.title}</SheetTitle>
          <SheetDescription>{text.description}</SheetDescription>
        </SheetHeader>
        {!paid ? (
          <SheetPanel>
            <PaidPlanNotice feature="Cloudflare email" />
          </SheetPanel>
        ) : (
          <form onSubmit={submit} className="flex min-h-0 min-w-0 flex-1 flex-col">
            <SheetPanel className="min-w-0 space-y-6">
              <ol className="space-y-5">
                {mode === "receive" ? (
                  <>
                    <Step n={1} title="Deploy the mailbox Worker" link={{ href: "https://github.com/mailxem/mail/blob/undefined/devops/cloudflare-mailbox/README.md", label: "Deployment guide" }}>
                      It stores mail in a private R2 bucket and D1 database in your account. Set <Code>MAILBOX_ADDRESS</Code> to the address below.
                      <Command value="npx wrangler secret put MAILBOX_API_SECRET" />
                      <Command value="npx wrangler deploy" />
                    </Step>
                    <Step n={2} title="Route the address to the Worker" link={{ href: "https://dash.cloudflare.com/?to=/:account/:zone/email/routing/routes", label: "Email Routing" }}>
                      Add a custom address rule whose action is <em>Send to a Worker</em>.
                    </Step>
                    <Step n={3} title="Connect it here" last>
                      <div className="mt-3 space-y-4">
                        <Field id="cf-worker-url" label="Worker URL">
                          <Input id="cf-worker-url" required type="url" pattern="https://.*" placeholder="https://mailbox.your-account.workers.dev" value={fields.workerUrl} onChange={set("workerUrl")} autoComplete="url" />
                        </Field>
                        <Field id="cf-inbox-address" label="Mailbox address" hint="Must match MAILBOX_ADDRESS and the routing rule exactly.">
                          <Input id="cf-inbox-address" required type="email" placeholder="support@yourdomain.com" value={fields.address} onChange={set("address")} />
                        </Field>
                        <Field id="cf-inbox-secret" label="Mailbox API secret" hint="The same secret you stored in the Worker. Sent once to verify, stored encrypted.">
                          <Secret id="cf-inbox-secret" minLength={32} value={fields.secret} onChange={set("secret")} />
                        </Field>
                      </div>
                    </Step>
                  </>
                ) : (
                  <>
                    <Step n={1} title="Copy your account ID" link={{ href: "https://dash.cloudflare.com/?to=/:account/home", label: "Dashboard" }}>
                      It’s on your account home page, under <em>Account details</em>.
                    </Step>
                    <Step n={2} title="Create an Email Sending token" link={{ href: "https://dash.cloudflare.com/profile/api-tokens", label: "API tokens" }}>
                      Scope it to Email Sending for this account only. Sending to any recipient needs Workers Paid.
                    </Step>
                    <Step n={3} title="Connect it here" last>
                      <div className="mt-3 space-y-4">
                        <Field id="cf-sender-address" label="Sender address" hint="A verified sending address on your Cloudflare domain.">
                          <Input id="cf-sender-address" required type="email" placeholder="notifications@yourdomain.com" value={fields.address} onChange={set("address")} />
                        </Field>
                        <Field id="cf-account-id" label="Account ID" hint={fields.accountId && !/^[a-fA-F0-9]{32}$/.test(fields.accountId) ? "Account IDs are 32 hex characters." : undefined}>
                          <Input id="cf-account-id" required pattern="[a-fA-F0-9]{32}" maxLength={32} className="font-mono" placeholder="0123456789abcdef0123456789abcdef" value={fields.accountId} onChange={set("accountId")} autoComplete="off" />
                        </Field>
                        <Field id="cf-token" label="API token" hint="Stored encrypted and never shown again.">
                          <Secret id="cf-token" minLength={16} value={fields.token} onChange={set("token")} />
                        </Field>
                      </div>
                    </Step>
                  </>
                )}
              </ol>
            </SheetPanel>
            <SheetFooter>
              <Button type="button" variant="ghost" onClick={() => onOpenChange(false)}>Cancel</Button>
              <Button type="submit" disabled={busy}>
                {busy && <Loader2 className="size-4 animate-spin" />}
                {busy ? text.busy : text.submit}
              </Button>
            </SheetFooter>
          </form>
        )}
      </SheetPopup>
    </Sheet>
  );
}

function Step({ n, title, link, last, children }: { n: number; title: string; link?: { href: string; label: string }; last?: boolean; children: ReactNode }) {
  return (
    <li className="relative flex gap-3">
      {!last && <span aria-hidden className="absolute left-3 top-7 bottom-[-14px] w-px bg-border" />}
      <span className="grid size-6 shrink-0 place-items-center rounded-full border bg-card text-xs font-medium tabular-nums">{n}</span>
      <div className="min-w-0 flex-1 pt-0.5">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <p className="text-sm font-medium">{title}</p>
          {link && (
            <a href={link.href} target="_blank" rel="noreferrer" className="inline-flex items-center gap-1 text-xs font-medium text-[#C2580A] hover:underline dark:text-[#FAAE40]">
              {link.label} <ArrowUpRight className="size-3" />
            </a>
          )}
        </div>
        <div className="mt-1 text-[13px] leading-5 text-muted-foreground">{children}</div>
      </div>
    </li>
  );
}

function Field({ id, label, hint, children }: { id: string; label: string; hint?: string; children: ReactNode }) {
  return (
    <div className="space-y-1.5">
      <Label htmlFor={id} className="text-foreground">{label}</Label>
      {children}
      {hint && <p className="text-xs text-muted-foreground">{hint}</p>}
    </div>
  );
}

function Secret({ id, minLength, value, onChange }: { id: string; minLength: number; value: string; onChange: (e: React.ChangeEvent<HTMLInputElement>) => void }) {
  const [shown, setShown] = useState(false);
  return (
    <div className="relative">
      <Input id={id} required type={shown ? "text" : "password"} minLength={minLength} maxLength={4096} value={value} onChange={onChange} autoComplete="new-password" className="pr-10 font-mono" />
      <button type="button" onClick={() => setShown(!shown)} aria-label={shown ? "Hide" : "Show"} className="absolute inset-y-0 right-0 grid w-10 place-items-center text-muted-foreground hover:text-foreground">
        {shown ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
      </button>
    </div>
  );
}

function Code({ children }: { children: ReactNode }) {
  return <code className="rounded bg-muted px-1 py-0.5 font-mono text-[12px] text-foreground">{children}</code>;
}

function Command({ value }: { value: string }) {
  const [copied, setCopied] = useState(false);
  return (
    <div className="mt-2 flex items-center gap-2 rounded-lg bg-muted px-3 py-2 font-mono text-[12px] text-foreground">
      <span className="min-w-0 flex-1 break-all">{value}</span>
      <button
        type="button"
        aria-label="Copy command"
        className="shrink-0 text-muted-foreground hover:text-foreground"
        onClick={() => void navigator.clipboard.writeText(value).then(() => { setCopied(true); setTimeout(() => setCopied(false), 1500); })}
      >
        {copied ? <Check className="size-3.5" /> : <Copy className="size-3.5" />}
      </button>
    </div>
  );
}
