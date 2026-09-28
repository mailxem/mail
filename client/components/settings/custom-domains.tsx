"use client";
import { useConfirmSheet } from "@/components/ui/confirm-sheet";
import { useState } from "react";
import { Globe, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { useResourcePage } from "@/hooks/use-resource-page";
import { useMarketing } from "@/lib/marketing/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { CollectionCard } from "@/components/ui/collection-card";
import { CollectionPagination } from "@/components/ui/collection-pagination";
import { QueryState, Empty } from "@/components/marketing/shared";
import { workspaceClassName } from "@/lib/workspace-styles";
interface Domain { id:string; domain:string; dnsRecord:string; isVerified:boolean }
export function CustomDomains() {
  const confirm = useConfirmSheet();
 const [page,setPage] = useState(1);
 const query = useResourcePage<Domain>("domains",page,12);
 const { request } = useMarketing();
 const [name,setName] = useState("");
 const [busy,setBusy] = useState("");
 const add = async(event:React.FormEvent)=>{
  event.preventDefault();setBusy("new");
  try{await request("marketing/domains","POST",{domain:name});setName("");setPage(1);await query.refetch();toast.success("Domain added. Add the TXT record to verify ownership.");}catch(error){toast.error((error as Error).message);}finally{setBusy("");}
 };
 const verify = async(id:string)=>{
  setBusy(id);try{await request(`marketing/domains/${id}/verify`,"POST");toast.success("Domain ownership verified");}catch(error){toast.error((error as Error).message);}finally{await query.refetch();setBusy("");}
 };
 const remove = async(domain:Domain)=>{
  if (!(await confirm({ title: `Remove ${domain.domain}?`, description: "This domain will be removed from your workspace.", confirmLabel: "Remove domain", variant: "destructive" }))) return;
  setBusy(domain.id);try{await request(`domains/${domain.id}`,"DELETE");await query.refetch();toast.success("Domain removed");}catch(error){toast.error((error as Error).message);}finally{setBusy("");}
 };
 return <section className={workspaceClassName("product-panel")}><h2 className="text-lg font-semibold">Custom domains</h2><p className="mt-1 text-sm text-muted-foreground">Verify ownership with a DNS TXT record. Your sending provider manages SPF, DKIM, and email authentication.</p><form onSubmit={add} className="my-6 flex flex-wrap items-end gap-3"><div className="min-w-0 flex-1 space-y-2"><Label htmlFor="new-domain">Domain name</Label><Input id="new-domain" required value={name} onChange={e=>setName(e.target.value)} placeholder="example.com" maxLength={253}/></div><Button type="submit" disabled={!!busy}>{busy==="new"?"Adding…":"Add domain"}</Button></form>
 {query.isLoading || query.error ? <QueryState loading={query.isLoading} error={query.error} retry={()=>void query.refetch()}/> : !query.data?.data.length ? <Empty title="No domains yet" description="Add your domain to verify that it belongs to your workspace."/> : <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">{query.data.data.map(domain=><CollectionCard key={domain.id} title={domain.domain} icon={<Globe size={22}/>} badge={domain.isVerified?"Verified":"Awaiting verification"} description="Domain ownership" action={busy===domain.id?"Checking…":"Check DNS"} onAction={()=>{if(!busy)void verify(domain.id);}} menu={<Button variant="ghost" size="icon" aria-label={`Remove ${domain.domain}`} disabled={!!busy} onClick={()=>void remove(domain)}><Trash2 size={16}/></Button>}><div className="mt-5 space-y-3 rounded-xl bg-muted p-3 text-xs"><div><p className="text-muted-foreground">TXT record name</p><code className="break-all">_xem.{domain.domain}</code></div><div><p className="text-muted-foreground">TXT record value</p><code className="break-all">{domain.dnsRecord || "Click Check DNS to create a verification record."}</code></div></div></CollectionCard>)}</div>}
 <CollectionPagination page={page} limit={12} total={query.data?.total || 0} onPageChange={setPage}/></section>;
}
