"use client";
import { useEffect, useState } from "react";
import { Loader2, Send } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useApi } from "@/hooks/use-api";

export type TestContact = {
  id: string;
  email: string;
  firstName?: string;
  lastName?: string;
  company?: string;
  city?: string;
  country?: string;
  state?: string;
  phone?: string;
};

const KEY = "xem:test-recipients";
const MAX = 5;

/** The variables a campaign send merges for a contact (server: tasks/handlers.go). */
export function contactVariables(c?: TestContact): Record<string, string> {
  if (!c) return {};
  const name = `${c.firstName ?? ""} ${c.lastName ?? ""}`;
  return {
    email: c.email,
    first_name: c.firstName ?? "",
    last_name: c.lastName ?? "",
    name,
    full_name: name,
    company: c.company ?? "",
    city: c.city ?? "",
    country: c.country ?? "",
    state: c.state ?? "",
    phone: c.phone ?? "",
  };
}

export function parseRecipients(raw: string) {
  const list = [
    ...new Set(
      raw
        .split(/[,\s]+/)
        .map((s) => s.trim().toLowerCase())
        .filter(Boolean),
    ),
  ];
  const invalid = list.filter((a) => !/^[^@\s]+@[^@\s]+\.[^@\s]+$/.test(a));
  return { list, invalid };
}

// Sends the campaign as it stands to up to five people, optionally read as one
// of the list's contacts would get it. Only those addresses receive it.
export function SendTest({
  templateId,
  smtpConfigId,
  subject,
  data,
  contacts,
}: {
  templateId?: string;
  smtpConfigId?: string;
  subject: string;
  data?: Record<string, string>;
  contacts: TestContact[];
}) {
  const { apiFetch, session } = useApi();
  const [to, setTo] = useState("");
  const [as, setAs] = useState("");
  const [save, setSave] = useState(false);
  const [busy, setBusy] = useState(false);
  useEffect(() => {
    let saved = "";
    try {
      saved = localStorage.getItem(KEY) ?? "";
    } catch {}
    setTo((current) => current || saved || session?.user?.email || "");
  }, [session?.user?.email]);

  const { list, invalid } = parseRecipients(to);
  const ready = !!templateId && !!smtpConfigId && !!subject.trim();
  const problem = !ready
    ? "Choose a template, a sender and a subject first."
    : invalid.length
      ? `Check ${invalid[0]}`
      : list.length > MAX
        ? `Up to ${MAX} addresses at a time`
        : "";

  async function send() {
    setBusy(true);
    const contact = contacts.find((c) => c.id === as);
    const variables = { ...contactVariables(contact), ...(data ?? {}) };
    try {
      const results = await Promise.all(
        list.map((address) =>
          apiFetch("emails", {
            method: "POST",
            body: JSON.stringify({
              to: address,
              templateId,
              smtpConfigId,
              subject: `[Test] ${subject.trim()}`,
              data: variables,
              test: true,
            }),
          }),
        ),
      );
      const failed = results.filter((r) => !r.ok);
      if (failed.length) {
        const body = await failed[0].json().catch(() => null);
        throw new Error(body?.message || "Couldn’t send the test");
      }
      if (save) {
        try {
          localStorage.setItem(KEY, list.join(", "));
        } catch {}
      }
      toast.success(
        `Test sent to ${list.length === 1 ? list[0] : `${list.length} people`}${contact ? `, as ${contact.firstName || contact.email} would see it` : ""}.`,
      );
    } catch (e) {
      toast.error((e as Error).message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="space-y-3 border-t border-border p-5">
      <h3 className="text-sm font-medium">Send a test</h3>
      <div className="space-y-1.5">
        <label htmlFor="test-to" className="text-xs text-muted-foreground">
          To (up to {MAX}, separated by commas)
        </label>
        <Input
          id="test-to"
          value={to}
          onChange={(e) => setTo(e.target.value)}
          placeholder="you@example.com, editor@example.com"
        />
      </div>
      {contacts.length > 0 && (
        <div className="space-y-1.5">
          <label htmlFor="test-as" className="text-xs text-muted-foreground">
            Preview as
          </label>
          <select
            id="test-as"
            value={as}
            onChange={(e) => setAs(e.target.value)}
            className="h-9 w-full rounded-md border border-input bg-background px-2 text-sm"
          >
            <option value="">Nobody in particular (fields stay empty)</option>
            {contacts.map((c) => (
              <option key={c.id} value={c.id}>
                {[c.firstName, c.lastName].filter(Boolean).join(" ") || c.email}{" "}
                · {c.email}
              </option>
            ))}
          </select>
          <p className="text-xs text-muted-foreground">
            They don’t get the email. Their name and fields fill it in.
          </p>
        </div>
      )}
      <label className="flex items-center gap-2 text-xs">
        <input
          type="checkbox"
          checked={save}
          onChange={(e) => setSave(e.target.checked)}
        />
        Save as my default test addresses
      </label>
      {problem && list.length > 0 && (
        <p role="alert" className="text-xs text-destructive">
          {problem}
        </p>
      )}
      <Button
        type="button"
        variant="outline"
        size="sm"
        disabled={busy || !!problem || !list.length}
        onClick={() => void send()}
      >
        {busy ? (
          <Loader2 className="size-4 animate-spin motion-reduce:animate-none" />
        ) : (
          <Send className="size-4" />
        )}
        {busy ? "Sending…" : "Send test"}
      </Button>
    </div>
  );
}
