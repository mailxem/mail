"use client";
import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { Check, Copy, ShieldCheck } from "lucide-react";
import { toast } from "sonner";
import { Button } from "../ui/button";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import { Switch } from "../ui/switch";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "../ui/sheet";
import { useApi } from "@/hooks/use-api";
import type { APIKey } from "@/lib";

const EXPIRY = [
  { value: "never", label: "No expiry" },
  { value: "7d", label: "7 days" },
  { value: "30d", label: "30 days" },
  { value: "90d", label: "90 days" },
  { value: "1y", label: "1 year" },
] as const;

export function CreateKey() {
  const queryClient = useQueryClient();
  const { apiFetch, session } = useApi();
  const isAdmin = ["ADMIN", "SUPER_ADMIN"].includes(session?.user?.role || "");
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [expiresIn, setExpiresIn] = useState<string>("90d");
  const [admin, setAdmin] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [created, setCreated] = useState<APIKey | null>(null);
  const [copied, setCopied] = useState(false);

  function reset() {
    setName("");
    setExpiresIn("90d");
    setAdmin(false);
    setError("");
    setCreated(null);
    setCopied(false);
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!name.trim()) return setError("Give the key a name.");
    setBusy(true);
    setError("");
    try {
      const response = await apiFetch("api-keys", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: name.trim(), expiresIn, admin }),
      });
      const data = await response.json().catch(() => null);
      if (!response.ok)
        throw new Error(data?.message || "Unable to create the API key");
      setCreated(data as APIKey);
      await queryClient.invalidateQueries({ queryKey: ["api-keys"] });
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }

  async function copy() {
    if (!created) return;
    try {
      await navigator.clipboard.writeText(created.key);
      setCopied(true);
      toast.success("API key copied");
    } catch {
      toast.error("Unable to copy API key");
    }
  }

  return (
    <Sheet
      open={open}
      onOpenChange={(next) => {
        setOpen(next);
        if (!next) reset();
      }}
    >
      <SheetTrigger asChild>
        <Button variant="default">Create a key</Button>
      </SheetTrigger>
      <SheetContent>
        {created ? (
          <>
            <SheetHeader>
              <SheetTitle>Your new API key</SheetTitle>
              <SheetDescription>
                Copy it now and keep it somewhere safe. Anyone with this key can
                {created.admin
                  ? " manage your workspace’s campaigns, contacts and sending."
                  : " send email from your workspace."}
              </SheetDescription>
            </SheetHeader>
            <div className="mt-6 space-y-3">
              <div className="flex items-center gap-2 rounded-xl border bg-muted p-3">
                <code className="min-w-0 flex-1 break-all font-mono text-xs">
                  {created.key}
                </code>
                <Button
                  size="icon"
                  variant="ghost"
                  className="size-8 shrink-0"
                  onClick={copy}
                  aria-label="Copy API key"
                >
                  {copied ? <Check size={15} /> : <Copy size={15} />}
                </Button>
              </div>
              <p className="text-xs text-muted-foreground">
                {created.admin ? "Admin key · " : ""}
                {expiresIn === "never"
                  ? "Never expires"
                  : `Expires in ${EXPIRY.find((o) => o.value === expiresIn)?.label}`}
              </p>
            </div>
            <SheetFooter className="mt-6">
              <Button onClick={() => setOpen(false)}>Done</Button>
            </SheetFooter>
          </>
        ) : (
          <>
            <SheetHeader>
              <SheetTitle>Create an API key</SheetTitle>
              <SheetDescription>
                Name it after the app or service that will use it.
              </SheetDescription>
            </SheetHeader>
            <form onSubmit={submit} className="mt-6 space-y-6">
              <div className="space-y-2">
                <Label htmlFor="key-name">Name</Label>
                <Input
                  id="key-name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="Production app"
                  maxLength={100}
                  autoFocus
                />
              </div>
              <fieldset className="space-y-2">
                <legend className="text-sm font-medium">Expires</legend>
                <div className="grid grid-cols-2 gap-2 sm:grid-cols-3">
                  {EXPIRY.map((o) => (
                    <label
                      key={o.value}
                      className={`flex cursor-pointer items-center justify-center rounded-lg border px-3 py-2 text-sm transition has-[:focus-visible]:ring-2 has-[:focus-visible]:ring-ring ${expiresIn === o.value ? "border-foreground bg-foreground text-background" : "bg-muted/40 hover:bg-muted"}`}
                    >
                      <input
                        type="radio"
                        name="expiresIn"
                        value={o.value}
                        checked={expiresIn === o.value}
                        onChange={() => setExpiresIn(o.value)}
                        className="sr-only"
                      />
                      {o.label}
                    </label>
                  ))}
                </div>
                {expiresIn === "never" && (
                  <p className="text-xs text-muted-foreground">
                    The key works until you delete it. Rotate it if it’s ever
                    shared by mistake.
                  </p>
                )}
              </fieldset>
              {isAdmin && (
                <div className="flex items-start justify-between gap-4 rounded-xl border p-4">
                  <div className="space-y-1">
                    <Label htmlFor="key-admin" className="flex items-center gap-1.5">
                      <ShieldCheck size={15} aria-hidden />
                      Admin key
                    </Label>
                    <p className="text-xs leading-5 text-muted-foreground">
                      Full access to campaigns, contacts, lists, templates,
                      automations, forms, domains and sending. It can’t manage
                      API keys, members or billing.
                    </p>
                  </div>
                  <Switch
                    id="key-admin"
                    checked={admin}
                    onCheckedChange={setAdmin}
                  />
                </div>
              )}
              {!admin && (
                <p className="text-xs text-muted-foreground">
                  Standard keys can send email and read email status.
                </p>
              )}
              {error && (
                <p role="alert" className="text-sm text-destructive">
                  {error}
                </p>
              )}
              <SheetFooter>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setOpen(false)}
                >
                  Cancel
                </Button>
                <Button type="submit" disabled={busy}>
                  {busy ? "Creating…" : "Create key"}
                </Button>
              </SheetFooter>
            </form>
          </>
        )}
      </SheetContent>
    </Sheet>
  );
}
