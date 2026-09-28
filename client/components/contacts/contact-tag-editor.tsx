"use client";
import { useEffect, useState } from "react";
import Link from "next/link";
import { Modal, QueryState } from "@/components/marketing/shared";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { useMarketing, useMarketingQuery } from "@/lib/marketing/api";
import type { ContactTag } from "@/app/audience/tags/page";
import { toast } from "sonner";

export function ContactTagEditor({ contactId, close, onSaved }: { contactId: string; close: () => void; onSaved: () => void }) {
  const all = useMarketingQuery<ContactTag[]>("marketing/tags");
  const current = useMarketingQuery<ContactTag[]>(`marketing/contacts/${contactId}/tags`);
  const { request, refresh } = useMarketing();
  const [selected, setSelected] = useState<string[]>([]);
  const [busy, setBusy] = useState(false);
  useEffect(() => { if (current.data) setSelected(current.data.map(tag => tag.id)); }, [current.data]);
  async function save() {
    setBusy(true);
    try { await request(`marketing/contacts/${contactId}/tags`, "PUT", { tagIds: selected }); await refresh(); onSaved(); close(); toast.success("Contact tags updated"); }
    catch (error) { toast.error((error as Error).message); }
    finally { setBusy(false); }
  }
  const loading = all.isPending || current.isPending;
  return <Modal open onOpenChange={open => { if (!open && !busy) close(); }} title="Manage contact tags" description="Choose the labels that describe this contact.">
    <QueryState loading={loading} error={all.error || current.error} retry={() => { void all.refetch(); void current.refetch(); }}/>
    {!loading && !all.error && !current.error && <div className="max-h-72 space-y-2 overflow-y-auto">{all.data?.map(tag => <label key={tag.id} className="flex items-center gap-3 rounded-xl border border-border p-3 text-sm"><Checkbox checked={selected.includes(tag.id)} onCheckedChange={checked => setSelected(ids => checked ? [...ids, tag.id] : ids.filter(id => id !== tag.id))}/>{tag.name}</label>)}{!all.data?.length && <p className="text-sm text-muted-foreground">Create tags in <Link className="text-primary underline" href="/audience/tags">Tags</Link> first.</p>}</div>}
    <div className="mt-5 flex justify-end gap-2"><Button variant="outline" onClick={close} disabled={busy}>Cancel</Button><Button disabled={busy || loading || !!all.error || !!current.error} onClick={save}>{busy ? "Saving…" : "Save tags"}</Button></div>
  </Modal>;
}
