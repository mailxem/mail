"use client";
import { useEffect, useState } from "react";
import { useMarketing, useMarketingQuery } from "@/lib/marketing/api";
import { useTeam } from "@/app/providers/team-provider";
import { QueryState } from "@/components/marketing/shared";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { workspaceClassName } from "@/lib/workspace-styles";
import { toast } from "sonner";
interface Branding { dashboardName:string; logoUrl:string }
export function BrandingSettings() {
 const query = useMarketingQuery<Branding>("marketing/branding");
 const { request } = useMarketing();
 const { refreshTeam } = useTeam();
 const [draft,setDraft] = useState<Branding>({dashboardName:"",logoUrl:""});
 const [dirty,setDirty] = useState(false);
 const [saving,setSaving] = useState(false);
 useEffect(()=>{if(query.data && !dirty)setDraft(query.data);},[query.data,dirty]);
 const save = async (event:React.FormEvent) => {
  event.preventDefault(); setSaving(true);
  try { await request("marketing/branding","PUT",draft); await query.refetch(); setDirty(false); await refreshTeam(); toast.success("Workspace branding saved"); }
  catch(error){toast.error((error as Error).message);}finally{setSaving(false);}
 };
 if(query.isPending || query.error)return <QueryState loading={query.isPending} error={query.error} retry={()=>void query.refetch()}/>;
 return <section className={workspaceClassName("product-panel")}><h2 className="text-lg font-semibold">Workspace branding</h2><p className="mt-1 text-sm text-muted-foreground">Your workspace identity within Xem.</p><form onSubmit={save} className="mt-6 max-w-2xl space-y-5"><div className="space-y-2"><Label htmlFor="workspace-name">Workspace name</Label><Input id="workspace-name" required minLength={2} maxLength={80} value={draft.dashboardName} onChange={e=>{setDirty(true);setDraft({...draft,dashboardName:e.target.value});}}/></div><div className="space-y-2"><Label htmlFor="workspace-logo">Logo URL</Label><Input id="workspace-logo" type="url" maxLength={2048} placeholder="https://example.com/logo.png" value={draft.logoUrl || ""} onChange={e=>{setDirty(true);setDraft({...draft,logoUrl:e.target.value});}}/><p className="text-xs text-muted-foreground">Use a publicly accessible HTTPS image URL.</p></div><Button disabled={saving || !dirty} type="submit">{saving ? "Saving…":"Save branding"}</Button></form></section>;
}
type ConsentMode = "none" | "opt_out" | "opt_in";
const consentModes: { value: ConsentMode; label: string; hint: string }[] = [
 { value: "none", label: "Follow each email’s setting", hint: "Opens and clicks are tracked wherever “Track opens and link clicks” is on." },
 { value: "opt_out", label: "Track unless the reader objects", hint: "Readers can turn tracking off from the unsubscribe page." },
 { value: "opt_in", label: "Track only readers who agree", hint: "Required in some EU countries. Readers agree from the unsubscribe page." },
];
export function TrackingConsentSettings() {
 const query = useMarketingQuery<{ mode: ConsentMode }>("marketing/tracking-consent");
 const { request } = useMarketing();
 const [saving,setSaving] = useState(false);
 const choose = async (mode:ConsentMode) => {
  setSaving(true);
  try { await request("marketing/tracking-consent","PUT",{mode}); await query.refetch(); toast.success("Tracking consent saved"); }
  catch(error){toast.error((error as Error).message);}finally{setSaving(false);}
 };
 if(query.isPending || query.error)return <QueryState loading={query.isPending} error={query.error} retry={()=>void query.refetch()}/>;
 return <section className={workspaceClassName("product-panel")}><h2 className="text-lg font-semibold">Tracking consent</h2><p className="mt-1 text-sm text-muted-foreground">Applies to campaigns, newsletters and marketing automations. Transactional email follows each email’s setting.</p><fieldset disabled={saving} className="mt-6 max-w-2xl space-y-3">{consentModes.map(m=><label key={m.value} className="flex items-start gap-2 text-sm"><input type="radio" name="tracking-consent" className="mt-1" checked={query.data.mode===m.value} onChange={()=>void choose(m.value)}/><span>{m.label}<span className="block text-xs text-muted-foreground">{m.hint}</span></span></label>)}</fieldset></section>;
}
