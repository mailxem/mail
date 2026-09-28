"use client";
import { useConfirmSheet } from "@/components/ui/confirm-sheet";
import { useState } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Edit, Mail, MoreHorizontal, Trash } from "lucide-react";
import { format } from "date-fns";
import { toast } from "sonner";
import { useCampaigns } from "@/app/providers/campaigns-provider";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { useApi } from "@/hooks/use-api";
import { QueryState, Status, Empty } from "@/components/marketing/shared";
import { CollectionCard } from "@/components/ui/collection-card";
import { CollectionPagination } from "@/components/ui/collection-pagination";

export function CampaignsList() {
  const confirm = useConfirmSheet();
  const { campaigns, refetch, loading, error, total, page, limit, setPage } = useCampaigns();
  const { apiFetch } = useApi();
  const [busy, setBusy] = useState("");
  const deleteCampaign = async (id: string) => {
    if (!(await confirm({ title: "Delete campaign?", description: "This cannot be undone.", confirmLabel: "Delete campaign", variant: "destructive" }))) return;
    setBusy(id);
    try {
      const response = await apiFetch(`campaigns/${id}`, { method: "DELETE" });
      if (!response.ok) throw new Error("Failed to delete campaign");
      if (campaigns.length === 1 && page > 1) setPage(page - 1);
      else await refetch();
      toast.success("Campaign deleted");
    } catch { toast.error("Failed to delete campaign"); }
    finally { setBusy(""); }
  };
  if (loading || error) return <QueryState loading={loading} error={error} retry={refetch}/>;
  if (!campaigns.length) return <Empty title="Create your first campaign" description="Connect an existing template, choose your audience, and schedule your email." action={<Button asChild><Link href="/campaigns/new">Create campaign</Link></Button>}/>;
  return <>
    <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
      {campaigns.map(campaign => <CollectionCard key={campaign.id} title={campaign.name} href={`/campaigns/${campaign.id}`} description={campaign.description} icon={<Mail size={22}/>} badge={<Status value={campaign.status}/>} action="View campaign" menu={
        <DropdownMenu><DropdownMenuTrigger asChild><Button variant="ghost" size="icon" disabled={busy === campaign.id} className="size-8" aria-label={`Actions for ${campaign.name}`}><MoreHorizontal size={18}/></Button></DropdownMenuTrigger><DropdownMenuContent align="end">
          <DropdownMenuItem asChild><Link href={`/campaigns/${campaign.id}`}><Edit className="mr-2 size-4"/>View campaign</Link></DropdownMenuItem>
          {campaign.status !== "SENDING" && <><DropdownMenuSeparator/><DropdownMenuItem className="text-destructive" onClick={() => deleteCampaign(campaign.id)}><Trash className="mr-2 size-4"/>Delete campaign</DropdownMenuItem></>}
        </DropdownMenuContent></DropdownMenu>
      }>
        <div className="mt-5 grid grid-cols-2 gap-3">
          <div className="rounded-xl bg-muted p-3"><p className="text-xs text-muted-foreground">Schedule</p><p className="mt-2 text-xs font-medium">{campaign.scheduledFor ? format(new Date(campaign.scheduledFor), "MMM d, p") : campaign.schedule || "Not scheduled"}</p></div>
          <div className="rounded-xl bg-muted p-3"><p className="text-xs text-muted-foreground">Created</p><p className="mt-2 text-xs font-medium">{format(new Date(campaign.createdAt), "MMM d, yyyy")}</p></div>
        </div>
      </CollectionCard>)}
    </div>
    <CollectionPagination page={page} limit={limit} total={total} onPageChange={setPage}/>
  </>;
}
