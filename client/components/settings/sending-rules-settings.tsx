"use client";
import { useEffect, useState } from "react";
import { useMarketing, useMarketingQuery } from "@/lib/marketing/api";
import type { SendWindow } from "@/lib/marketing/types";
import { Field, QueryState } from "@/components/marketing/shared";
import { Button } from "@/components/ui/button";
import { workspaceClassName } from "@/lib/workspace-styles";
import { toast } from "sonner";

const days = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];
export const defaultWindow: SendWindow = { enabled: true, start: "09:00", end: "17:00", days: [1, 2, 3, 4, 5] };

export function WindowFields({ value, onChange }: { value: SendWindow; onChange: (w: SendWindow) => void }) {
  return (
    <>
      <div className="grid grid-cols-2 gap-3">
        <Field label="From"><input type="time" required value={value.start} onChange={(e) => onChange({ ...value, start: e.target.value })} /></Field>
        <Field label="Until"><input type="time" required value={value.end} onChange={(e) => onChange({ ...value, end: e.target.value })} /></Field>
      </div>
      <div role="group" aria-label="Sending days" className="flex flex-wrap gap-3 text-sm">
        {days.map((day, i) => (
          <label key={day} className="flex items-center gap-1">
            <input className="!w-4" type="checkbox" checked={value.days.includes(i)} onChange={(e) => onChange({ ...value, days: e.target.checked ? [...value.days, i].sort() : value.days.filter((d) => d !== i) })} />
            {day}
          </label>
        ))}
      </div>
      <small className="text-xs text-muted-foreground">In each reader’s time zone when known. An end before the start runs past midnight.</small>
    </>
  );
}

// Per-automation window: null follows the workspace default.
export function AutomationWindowSetting({ value, onChange }: { value: SendWindow | null; onChange: (w: SendWindow | null) => void }) {
  const mode = !value ? "workspace" : value.enabled ? "custom" : "any";
  return (
    <>
      <Field label="Send emails" hint="Emails that come up outside the window wait for the next opening.">
        <select value={mode} onChange={(e) => onChange(e.target.value === "workspace" ? null : e.target.value === "any" ? { ...defaultWindow, enabled: false } : defaultWindow)}>
          <option value="workspace">Using the workspace sending window</option>
          <option value="any">At any time</option>
          <option value="custom">Only within a window</option>
        </select>
      </Field>
      {value?.enabled && <WindowFields value={value} onChange={onChange} />}
    </>
  );
}

interface Rules { timezone: string; sendWindow: SendWindow | null; frequencyCap: number; frequencyCapPeriod: "day" | "week" }
export function SendingRulesSettings() {
  const query = useMarketingQuery<Rules>("marketing/sending-rules");
  const { request } = useMarketing();
  const [draft, setDraft] = useState<Rules>({ timezone: "UTC", sendWindow: null, frequencyCap: 0, frequencyCapPeriod: "day" });
  const [dirty, setDirty] = useState(false);
  const [saving, setSaving] = useState(false);
  useEffect(() => { if (query.data && !dirty) setDraft(query.data); }, [query.data, dirty]);
  const set = (values: Partial<Rules>) => { setDirty(true); setDraft({ ...draft, ...values }); };
  const save = async (event: React.FormEvent) => {
    event.preventDefault(); setSaving(true);
    try { await request("marketing/sending-rules", "PUT", draft); await query.refetch(); setDirty(false); toast.success("Sending rules saved"); }
    catch (error) { toast.error((error as Error).message); } finally { setSaving(false); }
  };
  if (query.isPending || query.error) return <QueryState loading={query.isPending} error={query.error} retry={() => void query.refetch()} />;
  const windowOn = !!draft.sendWindow?.enabled;
  return (
    <section className={workspaceClassName("product-panel")}>
      <h2 className="text-lg font-semibold">Automation sending rules</h2>
      <p className="mt-1 text-sm text-muted-foreground">Defaults for every automation. Steps marked transactional ignore them.</p>
      <form onSubmit={save} className={`${workspaceClassName("product-form")} mt-6 max-w-2xl space-y-5`}>
        <Field label="Workspace time zone" hint="Used for contacts without a time zone of their own.">
          <input required value={draft.timezone} placeholder="Europe/London" onChange={(e) => set({ timezone: e.target.value })} />
        </Field>
        <label className="flex items-start gap-2 text-sm">
          <input className="!w-4" type="checkbox" checked={windowOn} onChange={(e) => set({ sendWindow: e.target.checked ? defaultWindow : null })} />
          <span>Only send automation emails within a window<small className="mt-1 block text-muted-foreground">Automations can choose their own window instead.</small></span>
        </label>
        {windowOn && draft.sendWindow && <WindowFields value={draft.sendWindow} onChange={(w) => set({ sendWindow: w })} />}
        <div className="grid grid-cols-2 gap-3">
          <Field label="At most (emails per contact)" hint="0 turns the frequency cap off.">
            <input type="number" min={0} max={100} value={draft.frequencyCap} onChange={(e) => set({ frequencyCap: Number(e.target.value) })} />
          </Field>
          <Field label="Per">
            <select value={draft.frequencyCapPeriod} onChange={(e) => set({ frequencyCapPeriod: e.target.value as Rules["frequencyCapPeriod"] })}>
              <option value="day">24 hours</option>
              <option value="week">7 days</option>
            </select>
          </Field>
        </div>
        <Button disabled={saving || !dirty} type="submit">{saving ? "Saving…" : "Save sending rules"}</Button>
      </form>
    </section>
  );
}
