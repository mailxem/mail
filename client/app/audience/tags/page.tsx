"use client";
import { useConfirmSheet } from "@/components/ui/confirm-sheet";
import { useState } from "react";
import { Tag, Plus, MoreHorizontal, Pencil, Trash } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { PageHeading, QueryState, Empty, Modal, Field } from "@/components/marketing/shared";
import { CollectionCard } from "@/components/ui/collection-card";
import { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from "@/components/ui/dropdown-menu";
import { useMarketing, useMarketingQuery } from "@/lib/marketing/api";
import { workspaceClassName } from "@/lib/workspace-styles";
import { toast } from "sonner";

export interface ContactTag { id: string; name: string; value?: string; contactCount: number; }
export default function TagsPage() {
  const confirm = useConfirmSheet();
  const query = useMarketingQuery<ContactTag[]>("marketing/tags");
  const { request, refresh } = useMarketing();
  const [editing, setEditing] = useState<Partial<ContactTag> | null>(null);
  const [busy, setBusy] = useState(false);
  async function save(event: React.FormEvent) {
    event.preventDefault(); if (!editing) return; setBusy(true);
    try { await request(`marketing/tags${editing.id ? `/${editing.id}` : ""}`, editing.id ? "PUT" : "POST", { name: editing.name, value: editing.value || "" }); await refresh(); setEditing(null); toast.success("Tag saved"); }
    catch (error) { toast.error((error as Error).message); }
    finally { setBusy(false); }
  }
  async function remove(tag: ContactTag) {
    if (!(await confirm({ title: `Delete “${tag.name}”?`, description: "It will be removed from contacts without deleting them.", confirmLabel: "Delete tag", variant: "destructive" }))) return;
    try { await request(`marketing/tags/${tag.id}`, "DELETE"); await refresh(); toast.success("Tag deleted"); }
    catch (error) { toast.error((error as Error).message); }
  }
  return <>
    <PageHeading title="Tags" description="Organize your existing contacts with reusable audience labels." action={<Button onClick={() => setEditing({ name: "" })}><Plus size={16}/>Create tag</Button>}/>
    <section className={workspaceClassName("product-panel")}>
      <div className={workspaceClassName("panel-toolbar")}><h2>Your tags</h2><span className="text-xs text-muted-foreground">{query.data?.length ?? 0} tags</span></div>
      {query.isPending || query.error ? <QueryState loading={query.isPending} error={query.error} retry={() => void query.refetch()}/> : !query.data?.length ? <Empty title="Give your audience some structure" description="Create a tag, then assign it to contacts from a contact list." action={<Button onClick={() => setEditing({ name: "" })}>Create tag</Button>}/> : <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
        {query.data.map(tag => <CollectionCard key={tag.id} icon={<Tag size={22}/>} title={tag.name} badge={`${tag.contactCount} contacts`} description={tag.value || "An audience label for your contacts."} action="Edit tag" onAction={() => setEditing(tag)} menu={<DropdownMenu><DropdownMenuTrigger asChild><Button variant="ghost" size="icon" className="size-8" aria-label={`Actions for ${tag.name}`}><MoreHorizontal size={18}/></Button></DropdownMenuTrigger><DropdownMenuContent align="end"><DropdownMenuItem onClick={() => setEditing(tag)}><Pencil className="mr-2 size-4"/>Edit tag</DropdownMenuItem><DropdownMenuItem className="text-destructive" onClick={() => remove(tag)}><Trash className="mr-2 size-4"/>Delete tag</DropdownMenuItem></DropdownMenuContent></DropdownMenu>}/>)}</div>}
    </section>
    <Modal open={!!editing} onOpenChange={open => { if (!open && !busy) setEditing(null); }} title={editing?.id ? "Edit tag" : "Create tag"} description="Use a short, recognizable name for this audience label.">
      <form onSubmit={save} className="space-y-5"><Field label="Name"><Input required maxLength={80} value={editing?.name || ""} onChange={e => setEditing({ ...editing, name: e.target.value })}/></Field><Field label="Description"><Input maxLength={120} value={editing?.value || ""} onChange={e => setEditing({ ...editing, value: e.target.value })}/></Field><div className="flex justify-end gap-2"><Button type="button" variant="outline" onClick={() => setEditing(null)} disabled={busy}>Cancel</Button><Button disabled={busy}>{busy ? "Saving…" : "Save tag"}</Button></div></form>
    </Modal>
  </>;
}
