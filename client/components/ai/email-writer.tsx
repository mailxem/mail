"use client";
import { useState } from "react";
import { Sparkles, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Modal, Field } from "@/components/marketing/shared";
import { useMarketing } from "@/lib/marketing/api";

export interface EmailDraft {
  subject: string;
  body: string;
  images?: { url: string; alt: string }[];
  design?: { schemaVersion: number; body: { rows: unknown[] }; [key: string]: unknown };
  previewHtml?: string;
}
export function EmailWriter({ subject = "", body = "", format = "text", onApply, applyLabel = "Use draft" }: {
  subject?: string; body?: string; format?: "text" | "design";
  onApply: (draft: EmailDraft) => void | Promise<void>; applyLabel?: string;
}) {
  const { request } = useMarketing();
  const [open, setOpen] = useState(false);
  const [instruction, setInstruction] = useState("");
  const [tone, setTone] = useState("Professional");
  const [draft, setDraft] = useState<EmailDraft | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const designing = format === "design";
  async function generate() {
    setBusy(true); setError("");
    try {
      const result = await request<EmailDraft>("marketing/email-draft", "POST", {
        instruction, tone, format, subject: draft?.subject ?? subject,
        body: (draft?.body ?? body).slice(0, 16000),
      });
      if (!result?.subject || !result?.body) throw new Error("The writing service returned an empty draft. Please try again.");
      if (designing && (!Array.isArray(result.design?.body?.rows) || !result.design.body.rows.length || result.design.schemaVersion !== 18 || !result.previewHtml)) {
        throw new Error("The writing service did not return an editable design. Update the email service to enable template generation, then try again.");
      }
      setDraft(result);
    } catch (error) { setError((error as Error).message); }
    finally { setBusy(false); }
  }
  return <>
    <Button type="button" variant="outline" onClick={() => { setDraft(null); setError(""); setOpen(true); }}><Sparkles size={16}/>Write with AI</Button>
    <Modal open={open} onOpenChange={value => { if (!busy) setOpen(value); }} wide
      title={designing ? "Design with Xem" : "Write with Xem"}
      description={designing ? "Create an email layout with editable sections, typography, images, and buttons." : "Draft a personal email, refine its tone, and review the copy before using it."}>
      <div className="grid gap-6 md:grid-cols-[0.8fr_1.2fr]">
        <div className="space-y-5">
          <Field label={designing ? "Describe your email design" : "What would you like to write?"}><Textarea rows={7} maxLength={4000} value={instruction} onChange={e => setInstruction(e.target.value)} placeholder={designing ? "Create a modern coffee newsletter. Use warm cream, espresso typography, a large photo, two article columns, and a bold shop button. Brand: Sunday Roast. Audience: home brewers." : "Write a warm follow-up after our meeting. Summarize the next steps and invite questions. You can include image URLs."}/></Field>
          <Field label="Tone"><select className="h-10 w-full rounded-lg border border-input bg-white px-3 text-sm" value={tone} onChange={e => setTone(e.target.value)}>{["Professional", "Friendly", "Concise", "Warm", "Persuasive"].map(item => <option key={item}>{item}</option>)}</select></Field>
          <Button type="button" disabled={busy || instruction.trim().length < 3} onClick={generate} className="w-full">{busy ? <Loader2 className="animate-spin" size={16}/> : <Sparkles size={16}/>}{busy ? designing ? "Designing…" : "Writing…" : draft ? "Generate a revision" : designing ? "Generate email design" : "Generate draft"}</Button>
          <p className="text-xs leading-relaxed text-muted-foreground">Your instructions and current draft are sent to the workspace’s AI writing service. Review facts and placeholders before using the result.</p>
          {error && <p role="alert" className="rounded-xl bg-red-50 p-3 text-sm text-red-700">{error}</p>}
        </div>
        <div className="min-w-0 rounded-2xl border border-border bg-muted/30 p-5">
          {draft ? <div className="space-y-4">
            <Field label="Subject"><Input value={draft.subject} maxLength={200} onChange={e => setDraft({ ...draft, subject: e.target.value })}/></Field>
            {designing ? <iframe title="Generated email design" srcDoc={draft.previewHtml} sandbox="" referrerPolicy="no-referrer" className="h-[420px] w-full rounded-lg border bg-white"/> : <Field label="Email copy"><Textarea rows={13} maxLength={16000} value={draft.body} onChange={e => setDraft({ ...draft, body: e.target.value })}/></Field>}
            {!designing && !!draft.images?.length && <p className="text-xs text-muted-foreground">Includes {draft.images.length} image{draft.images.length === 1 ? "" : "s"} from your instructions.</p>}
            <Button type="button" className="w-full" disabled={busy} onClick={async () => {
              setBusy(true); setError("");
              try { await onApply(draft); setOpen(false); } catch (e) { setError((e as Error).message); } finally { setBusy(false); }
            }}>{applyLabel}</Button>
          </div> : <div className="flex min-h-80 flex-col items-center justify-center gap-3 text-center text-muted-foreground"><Sparkles size={28} className="text-violet-400"/><p className="text-sm">Your {designing ? "design" : "draft"} will appear here.</p><p className="max-w-64 text-xs leading-relaxed">{designing ? "Describe the brand, colors, layout, and content. Refine every block in the template editor." : "Give the assistant a purpose, audience, and the details your email should include."}</p></div>}
        </div>
      </div>
    </Modal>
  </>;
}
