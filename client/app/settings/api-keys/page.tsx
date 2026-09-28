"use client";
import { useConfirmSheet } from "@/components/ui/confirm-sheet";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Copy, KeyRound, MoreHorizontal, Trash, ShieldCheck, Clock3 } from "lucide-react";
import { format } from "date-fns";
import { toast } from "sonner";
import { APIKey } from "@/lib";
import { useApi } from "@/hooks/use-api";
import { useResourcePage } from "@/hooks/use-resource-page";
import { CollectionCard } from "@/components/ui/collection-card";
import { CollectionPagination } from "@/components/ui/collection-pagination";
import { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from "@/components/ui/dropdown-menu";
import { Metric, QueryState, Empty } from "@/components/marketing/shared";
import { workspaceClassName } from "@/lib/workspace-styles";

export default function APIKeysPage() {
  const confirm = useConfirmSheet();
  const [page, setPage] = useState(1);
  const query = useResourcePage<APIKey>("api-keys", page, 20);
  const apiKeys = query.data?.data ?? [];
  const { apiFetch } = useApi();
  const [busy, setBusy] = useState("");
  const expired = (key: APIKey) => !!key.expiresAt && new Date(key.expiresAt) < new Date();
  const remove = async (id: string) => {
    if (!(await confirm({ title: "Delete API key?", description: "Applications using it will lose access.", confirmLabel: "Delete API key", variant: "destructive" }))) return;
    setBusy(id);
    try {
      const response = await apiFetch(`api-keys/${id}`, { method: "DELETE" });
      if (!response.ok) throw new Error("Unable to delete API key");
      await query.refetch();
      toast.success("API key deleted");
    } catch { toast.error("Unable to delete API key"); }
    finally { setBusy(""); }
  };
  const copy = async (key: string) => {
    try { await navigator.clipboard.writeText(key); toast.success("API key copied"); }
    catch { toast.error("Unable to copy API key"); }
  };
  return <>
    <div className="mb-6 grid gap-4 sm:grid-cols-3">
      <Metric label="API keys" value={query.isLoading ? "—" : query.data?.total ?? 0} icon={<KeyRound size={18}/>}/>
      <Metric label="Active on this page" value={apiKeys.filter(key => !expired(key) && !key.isDeleted).length} icon={<ShieldCheck size={18}/>}/>
      <Metric label="Expired on this page" value={apiKeys.filter(expired).length} icon={<Clock3 size={18}/>}/>
    </div>
    <section className={workspaceClassName("product-panel")}>
      <div className={workspaceClassName("panel-toolbar")}><h2>Your API keys</h2><a href="https://docs.xem.email/api-reference/email/send-an-email" target="_blank" rel="noreferrer" className="text-xs font-medium text-primary">API documentation ↗</a></div>
      {query.isLoading || query.error ? <QueryState loading={query.isLoading} error={query.error} retry={() => void query.refetch()}/> : !apiKeys.length ? <Empty title="Connect your application" description="Create an API key to send email and access your workspace through the Xem API."/> : <>
        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {apiKeys.map(key => <CollectionCard key={key.id} icon={<KeyRound size={22}/>} badge={expired(key) ? "Expired" : key.isDeleted ? "Disabled" : "Active"} title={key.name} href={`/settings/api-keys/${key.id}`} action="View usage" description={`Created ${format(new Date(key.createdAt), "MMM d, yyyy")}`} menu={
            <DropdownMenu><DropdownMenuTrigger asChild><Button disabled={busy === key.id} variant="ghost" size="icon" className="size-8" aria-label={`Actions for ${key.name}`}><MoreHorizontal size={18}/></Button></DropdownMenuTrigger><DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => copy(key.key)} disabled={!key.key}><Copy className="mr-2 size-4"/>Copy key</DropdownMenuItem>
              <DropdownMenuItem className="text-destructive" onClick={() => remove(key.id)}><Trash className="mr-2 size-4"/>Delete key</DropdownMenuItem>
            </DropdownMenuContent></DropdownMenu>
          }>
            <div className="mt-5 flex items-center justify-between rounded-xl bg-muted p-3"><span className="font-mono text-sm tracking-wider" aria-label="API key hidden">••••••••••••••••</span><Button variant="ghost" size="icon" className="size-8" disabled={!key.key} onClick={() => copy(key.key)} aria-label={`Copy ${key.name}`}><Copy size={15}/></Button></div>
            <p className="mt-3 text-xs text-muted-foreground">{key.expiresAt ? `Expires ${format(new Date(key.expiresAt), "MMM d, yyyy")}` : "No expiration date"}</p>
          </CollectionCard>)}
        </div>
        <CollectionPagination page={page} limit={20} total={query.data?.total ?? 0} onPageChange={setPage}/>
      </>}
    </section>
  </>;
}
