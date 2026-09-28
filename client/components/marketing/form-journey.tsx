"use client";

import { useState } from "react";
import Link from "next/link";
import { Loader2, Plus, Sparkles, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { MarketingRequestError, useMarketing, useMarketingQuery } from "@/lib/marketing/api";
import type { Options } from "@/lib/marketing/types";
import { Field } from "./shared";

type EmailDraft = { name: string; subject: string; body: string; delayHours: number };
const initialEmail = (): EmailDraft => ({name:"Welcome",subject:"Thanks for joining us",body:"Hi {{first_name}},\n\nThanks for joining us. We're glad you're here.",delayHours:0});

export function FormJourneyPanel({formId,formName}:{formId:string;formName:string}) {
  const {request,refresh}=useMarketing();
  const options=useMarketingQuery<Options>("marketing/options");
  const [name,setName]=useState(`${formName} follow-up`.slice(0,120));
  const [instruction,setInstruction]=useState("");
  const [emails,setEmails]=useState<EmailDraft[]>([initialEmail()]);
  const [proposed,setProposed]=useState<EmailDraft[]|null>(null);
  const [sender,setSender]=useState("");
  const [address,setAddress]=useState("");
  const [busy,setBusy]=useState<"generate"|"save"|null>(null);
  const [error,setError]=useState("");
  const [saved,setSaved]=useState(false);
  const [pending,setPending]=useState<Record<string,unknown>|null>(null);
  const update=(index:number,patch:Partial<EmailDraft>)=>setEmails(rows=>rows.map((row,i)=>i===index?{...row,...patch}:row));
  async function generate(){
    setBusy("generate");setError("");
    try {
      const draft=await request<{emails:EmailDraft[]}>("marketing/form-draft","POST",{instruction:`Create follow-up emails for the form named ${formName}. ${instruction}`});
      setProposed(draft.emails);
    } catch(e){setError((e as Error).message);} finally{setBusy(null);}
  }
  async function save(e:React.FormEvent){
    e.preventDefault();if(busy)return;setBusy("save");setError("");
    const payload=pending||{requestId:crypto.randomUUID(),name,smtpConfigId:sender,postalAddress:address,emails};
    setPending(payload);
    try {
      await request(`marketing/forms/${formId}/journey`,"POST",payload);
      setSaved(true);setPending(null);await refresh();
    } catch(e){
      if(e instanceof MarketingRequestError && e.status >= 400 && e.status < 500) setPending(null);
      setError((e as Error).message);
    }finally{setBusy(null);}
  }
  if(saved)return <div className="space-y-4 py-4"><h3 className="text-lg font-semibold">Your journey is ready to review</h3><p className="text-sm text-muted-foreground">The email templates and automation were saved as a draft. Review the steps in Automations and activate when you’re ready.</p><Button asChild><Link href="/automations">Review in Automations</Link></Button></div>;
  return <form onSubmit={save} className="space-y-6">
    <p className="text-sm text-muted-foreground">Follow up when someone completes this form and opts into emails. Returning subscribers can enter again. Unsubscribed contacts stay excluded.</p>
    <fieldset disabled={!!busy||!!pending} className="space-y-5 disabled:opacity-70">
      <Field label="Journey name"><input required minLength={2} maxLength={120} value={name} onChange={e=>setName(e.target.value)}/></Field>
      <div className="space-y-3 rounded-xl border bg-muted/20 p-4">
        <Field label="Draft with AI" hint="Describe the resource, tone, and follow-up you want. Include factual details and links the draft should use."><textarea value={instruction} maxLength={3500} rows={3} onChange={e=>setInstruction(e.target.value)}/></Field>
        <Button type="button" variant="outline" disabled={instruction.trim().length<3} onClick={generate}>{busy==="generate"?<Loader2 className="animate-spin"/>:<Sparkles/>}Suggest emails</Button>
        {proposed&&<div className="space-y-3 border-t pt-3"><p className="text-sm font-medium">Suggested sequence</p>{proposed.map((email,i)=><div key={i} className="rounded-lg border bg-background p-3 text-sm"><strong>{email.subject}</strong><p className="mt-2 whitespace-pre-wrap">{email.body}</p><p className="mt-2 text-xs text-muted-foreground">Wait {email.delayHours} hours before this email</p></div>)}<Button type="button" variant="outline" onClick={()=>{setEmails(proposed);setProposed(null);}}>Use these emails</Button></div>}
      </div>
      {emails.map((email,index)=><section key={index} className="space-y-3 rounded-xl border p-4" aria-label={`Email ${index+1}`}>
        <div className="flex items-center justify-between"><h3 className="font-medium">Email {index+1}</h3>{emails.length>1&&<Button type="button" variant="ghost" size="icon" aria-label={`Remove email ${index+1}`} onClick={()=>setEmails(rows=>rows.filter((_,i)=>i!==index))}><Trash2 size={16}/></Button>}</div>
        <Field label="Template name"><input required minLength={2} maxLength={120} value={email.name} onChange={e=>update(index,{name:e.target.value})}/></Field>
        <Field label="Subject"><input required maxLength={200} value={email.subject} onChange={e=>update(index,{subject:e.target.value})}/></Field>
        <Field label="Email copy" hint="Use {{first_name}} or {{form_your_field_key}} to personalize."><textarea required rows={6} maxLength={12000} value={email.body} onChange={e=>update(index,{body:e.target.value})}/></Field>
        <Field label="Wait before this email (hours)"><input type="number" required min={0} max={8760} step={1} value={email.delayHours} onChange={e=>update(index,{delayHours:Number(e.target.value)})}/></Field>
      </section>)}
      {emails.length<4&&<Button type="button" variant="outline" onClick={()=>setEmails(rows=>[...rows,{...initialEmail(),name:"Follow-up",delayHours:48}])}><Plus/>Add follow-up</Button>}
      <Field label="Send from"><select required value={sender} onChange={e=>setSender(e.target.value)}><option value="">{options.isLoading ? "Loading senders…" : "Choose a sender"}</option>{options.data?.senders.map(row=><option key={row.id} value={row.id}>{row.fromEmail}</option>)}</select></Field>
      {options.error && <div role="alert" className="text-sm text-destructive">Senders could not be loaded. <button type="button" className="underline" onClick={()=>void options.refetch()}>Try again</button></div>}
      {options.data && !options.data.senders.length && <p className="text-sm text-muted-foreground">Add a sender in <Link className="underline" href="/settings/sending">sending settings</Link> before saving this journey.</p>}
      <Field label="Sender postal address" hint="Included with the unsubscribe link in these marketing emails."><textarea required minLength={8} maxLength={500} value={address} onChange={e=>setAddress(e.target.value)}/></Field>
    </fieldset>
    {error&&<p role="alert" className="text-sm text-destructive">{error}</p>}
    {pending&&error&&<p className="text-sm text-muted-foreground">The save could not be confirmed. Retry the same draft to check or finish saving it safely.</p>}
    <Button type="submit" disabled={!!busy}>{busy==="save"&&<Loader2 className="animate-spin"/>}{pending?"Retry saving draft":"Save journey draft"}</Button>
  </form>;
}
