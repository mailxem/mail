"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { format } from "date-fns";
import { ChevronDown, Copy, FilePlus, Plus, Search, Users } from "lucide-react";
import { toast } from "sonner";
import type { Campaign, CampaignAudience } from "@/lib";
import { useApi } from "@/hooks/use-api";
import { useTeam } from "@/app/providers/team-provider";
import { resourceEntity } from "@/lib/resource-response";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Status } from "@/components/marketing/shared";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";

type Source = Pick<Campaign, "id" | "name" | "status" | "createdAt" | "processed">;
type Audience = CampaignAudience | "";

/** New on the Campaigns page: start blank or from a past campaign. */
export function NewCampaignMenu() {
  const [open, setOpen] = useState(false);
  return <>
    <DropdownMenu modal={false}>
      <DropdownMenuTrigger asChild><Button><Plus size={15} strokeWidth={1.65}/>New campaign<ChevronDown size={15}/></Button></DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem asChild><Link href="/campaigns/new"><FilePlus className="mr-2 size-4"/>Blank</Link></DropdownMenuItem>
        <DropdownMenuItem onSelect={() => setOpen(true)}><Copy className="mr-2 size-4"/>Copy a campaign</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
    <CopyCampaignDialog open={open} onOpenChange={setOpen}/>
  </>;
}

/** Picks a campaign to copy (or copies campaign) and, for one that has sent, who the copy goes to. */
export function CopyCampaignDialog({ open, onOpenChange, campaign }: { open: boolean; onOpenChange: (open: boolean) => void; campaign?: Source }) {
  const router = useRouter();
  const { team } = useTeam();
  const { apiFetch, session } = useApi();
  const [search, setSearch] = useState("");
  const [picked, setPicked] = useState<Source | undefined>(campaign);
  const [audience, setAudience] = useState<Audience>("");
  useEffect(() => {
    if (open) { setPicked(campaign); setAudience(""); setSearch(""); }
  }, [open, campaign]);
  // ponytail: the 200 newest campaigns, searched here; add a server name filter if workspaces outgrow it.
  const sources = useQuery<Source[]>({
    queryKey: ["campaign-copy-sources", team?.id],
    enabled: open && !campaign && !!team?.id && !!session?.accessToken,
    queryFn: async ({ signal }) => {
      const response = await apiFetch(`campaigns?team_id=${team!.id}&sort=created_at&order=desc&limit=200`, { signal });
      if (!response.ok) throw new Error("Unable to load your campaigns. Please try again.");
      return (await response.json()).data ?? [];
    },
  });
  const term = search.trim().toLowerCase();
  const matches = (sources.data ?? []).filter(c => c.name.toLowerCase().includes(term));
  const sent = (picked?.processed ?? 0) > 0;
  const openCopy = (source: Source, to: Audience) => {
    onOpenChange(false);
    router.push(`/campaigns/new?copy=${source.id}${to ? `&audience=${to}` : ""}`);
  };
  const pick = (source: Source) => (source.processed ?? 0) > 0 ? setPicked(source) : openCopy(source, "");
  const choices: [Audience, string, string][] = picked ? [
    ["", "Everyone on the list", "The same list as before, as it is today."],
    ["not_opened", `People who didn’t open “${picked.name}”`, "Readers it reached who never opened it."],
    ["not_clicked", `People who didn’t click “${picked.name}”`, "Readers it reached who never clicked a link, including those who opened it."],
  ] : [];
  return <Dialog open={open} onOpenChange={onOpenChange}>
    <DialogContent className="sm:max-w-lg">
      <DialogHeader>
        <DialogTitle>{picked && sent ? "Who gets the copy?" : "Copy a campaign"}</DialogTitle>
        <DialogDescription>{picked && sent ? "The copy opens as a new draft with the same content, sender and settings. Unsubscribed and suppressed contacts are always left out." : "The copy opens as a new draft with the same content, list, sender and settings."}</DialogDescription>
      </DialogHeader>
      {picked && sent ? <>
        <div role="radiogroup" aria-label="Audience" className="space-y-2">
          {choices.map(([value, label, hint]) => <button key={value} type="button" role="radio" aria-checked={audience === value} onClick={() => setAudience(value)} className={cn("flex w-full items-start gap-3 rounded-xl border p-3 text-left text-sm", audience === value ? "border-primary bg-primary/5" : "border-border hover:bg-muted/50")}>
            <Users className="mt-0.5 size-4 shrink-0 text-muted-foreground"/>
            <span><span className="block font-medium">{label}</span><span className="block text-xs text-muted-foreground">{hint}</span></span>
          </button>)}
        </div>
        <div className="flex justify-end gap-2">
          {!campaign && <Button variant="ghost" onClick={() => setPicked(undefined)}>Back</Button>}
          <Button onClick={() => openCopy(picked, audience)}><Copy className="size-4"/>Open copy</Button>
        </div>
      </> : <>
        <div className="relative">
          <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground"/>
          <Input autoFocus value={search} onChange={e => setSearch(e.target.value)} placeholder="Search campaigns" aria-label="Search campaigns" className="pl-9"/>
        </div>
        <div className="max-h-80 space-y-1 overflow-y-auto">
          {sources.isPending ? <p role="status" className="p-3 text-sm text-muted-foreground">Loading campaigns…</p>
            : sources.error ? <p role="alert" className="p-3 text-sm text-destructive">{(sources.error as Error).message}</p>
            : !matches.length ? <p className="p-3 text-sm text-muted-foreground">{term ? "No campaigns match that search." : "No campaigns yet."}</p>
            : matches.map(c => <button key={c.id} type="button" onClick={() => pick(c)} className="flex w-full items-center justify-between gap-3 rounded-lg px-3 py-2 text-left text-sm hover:bg-muted">
              <span className="min-w-0"><span className="block truncate font-medium">{c.name}</span><span className="block text-xs text-muted-foreground">Created {format(new Date(c.createdAt), "MMM d, yyyy")}</span></span>
              <Status value={c.status}/>
            </button>)}
        </div>
      </>}
    </DialogContent>
  </Dialog>;
}

/** The copied fields, as the new-campaign form takes them. Mirrors marketing.CopyCampaign on the server. */
export function copiedValues(source: Campaign) {
  return {
    name: `${source.name} (copy)`,
    description: source.description ?? "",
    templateId: source.templateId,
    listId: source.listId,
    smtpConfigId: source.smtpConfigId,
    subject: source.subject ?? "",
    replyTo: source.replyTo ?? "",
    trackEngagement: source.trackEngagement ?? true,
    data: source.data ?? {},
    archiveNewsletterId: source.archiveNewsletterId ?? "",
  };
}

export type CampaignCopy = { id: string; name: string; audience: Audience };

/** Fills the new-campaign form from ?copy=<id>, once. Returns the copy and a setter to change its audience. */
export function useCampaignCopy(fill: (values: ReturnType<typeof copiedValues>) => void) {
  const { apiFetch, session } = useApi();
  const [copy, setCopy] = useState<CampaignCopy>();
  const started = useRef(false);
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const id = params.get("copy");
    if (!id || !session?.accessToken || started.current) return;
    started.current = true;
    void (async () => {
      try {
        const response = await apiFetch(`campaigns/${id}`);
        if (!response.ok) throw new Error();
        const source = resourceEntity<Campaign>(await response.json());
        fill(copiedValues(source));
        const audience = params.get("audience");
        setCopy({ id, name: source.name, audience: audience === "not_opened" || audience === "not_clicked" ? audience : "" });
      } catch {
        toast.error("Couldn’t load the campaign to copy. Start from blank or try again.");
      }
    })();
  }, [session?.accessToken, apiFetch, fill]);
  return [copy, setCopy] as const;
}

/** Says who a follow-up copy goes to, with a way back to the whole list. */
export function FollowUpNotice({ copy, onWholeList }: { copy?: CampaignCopy; onWholeList: () => void }) {
  if (!copy?.audience) return null;
  return <div role="status" className="flex flex-wrap items-start justify-between gap-3 rounded-xl border border-primary/30 bg-primary/5 p-3 text-sm">
    <p>Only people on this list who didn’t {copy.audience === "not_opened" ? "open" : "click"} “{copy.name}” get this email. Unsubscribed and suppressed contacts are left out.</p>
    <Button type="button" size="sm" variant="outline" onClick={onWholeList}>Send to the whole list</Button>
  </div>;
}
