"use client";
import { useConfirmSheet } from "@/components/ui/confirm-sheet";

import { Button } from "@/components/ui/button";
import { Trash, Users, UserCheck, ListFilter, MoreHorizontal, ArrowUpRight } from "lucide-react";
import { useMailingLists } from "@/app/providers/mailinglist-provider";
import { format } from "date-fns";
import { toast } from "sonner";
import { useApi } from "@/hooks/use-api";
import { Metric, QueryState, Empty } from "@/components/marketing/shared";
import { workspaceClassName } from "@/lib/workspace-styles";
import { CollectionCard } from "@/components/ui/collection-card";
import { CollectionPagination } from "@/components/ui/collection-pagination";
import { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from "@/components/ui/dropdown-menu";
import Link from "next/link";

export function ContactLists() {
  const confirm = useConfirmSheet();
  const { lists, refetch, pagination, setPagination, isLoading, error } = useMailingLists();
  const { apiFetch } = useApi();
  const deleteList = async (id: string) => {
    if (!(await confirm({ title: "Delete contact list?", description: "This cannot be undone.", confirmLabel: "Delete list", variant: "destructive" }))) return;
    try {
      const response = await apiFetch(`mailing-lists/${id}`, { method: "DELETE" });
      if (!response.ok) throw new Error("Failed to delete list");
      if (lists.length === 1 && pagination.page > 1) setPagination({ ...pagination, page: pagination.page - 1 });
      else await refetch();
      toast.success("Contact list deleted");
    } catch { toast.error("Failed to delete contact list"); }
  };
  return <>
    <div className="mb-6 grid gap-4 sm:grid-cols-3">
      <Metric label="Contact lists" value={isLoading ? "—" : pagination.total} icon={<Users size={18}/>}/>
      <Metric label="Subscribers in these lists" value={isLoading ? "—" : lists.reduce((sum, list) => sum + (list.subscribersCount || 0), 0).toLocaleString()} icon={<UserCheck size={18}/>}/>
      <Metric label="Lists on this page" value={isLoading ? "—" : lists.length} icon={<ListFilter size={18}/>}/>
    </div>
    <section className={workspaceClassName("product-panel")}>
      <div className={workspaceClassName("panel-toolbar")}><h2>All contact lists</h2><span className="text-xs text-muted-foreground">{pagination.total} lists</span></div>
      {isLoading || error ? <QueryState loading={isLoading} error={error} retry={refetch}/> : lists.length === 0 ? <Empty title="Build your first audience" description="Create a contact list to use with your campaigns, newsletters, and forms."/> : <>
        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {lists.map(list => <CollectionCard key={list.id} icon={<Users size={22}/>} badge="Contact list" title={list.name} description={list.description || "Your audience for email campaigns and newsletters."} href={`/audience/lists/${list.id}`} action="Open contact list" menu={
            <DropdownMenu><DropdownMenuTrigger asChild><Button variant="ghost" size="icon" className="size-8" aria-label={`Actions for ${list.name}`}><MoreHorizontal size={18}/></Button></DropdownMenuTrigger><DropdownMenuContent align="end">
              <DropdownMenuItem asChild><Link href={`/audience/lists/${list.id}`}><ArrowUpRight className="mr-2 size-4"/>Open list</Link></DropdownMenuItem>
              <DropdownMenuItem className="text-destructive" onClick={() => deleteList(list.id)}><Trash className="mr-2 size-4"/>Delete list</DropdownMenuItem>
            </DropdownMenuContent></DropdownMenu>
          }>
            <div className="mt-5 grid grid-cols-2 gap-3">
              <div className="rounded-xl bg-muted p-3"><p className="text-xs text-muted-foreground">Subscribers</p><p className="mt-1 text-xl font-semibold tracking-tight">{(list.subscribersCount || 0).toLocaleString()}</p></div>
              <div className="rounded-xl bg-muted p-3"><p className="text-xs text-muted-foreground">Last updated</p><p className="mt-2 text-xs font-medium">{format(new Date(list.updatedAt), "MMM d, yyyy")}</p></div>
            </div>
          </CollectionCard>)}
        </div>
        <CollectionPagination {...pagination} onPageChange={page => setPagination({ ...pagination, page })}/>
      </>}
    </section>
  </>;
}
