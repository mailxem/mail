"use client";
import { useState } from "react";
import { useMarketing, useMarketingQuery } from "@/lib/marketing/api";
import { Field, QueryState } from "@/components/marketing/shared";
import { Button } from "@/components/ui/button";
import { workspaceClassName } from "@/lib/workspace-styles";
import { toast } from "sonner";

type Frequency = "daily" | "weekly" | "off";
interface Settings { frequency: Frequency; includeMe: boolean; isOwner: boolean; timezone: string }
interface Preview { subject: string; html: string; quiet: boolean }

export function SummaryEmailSettings() {
  const query = useMarketingQuery<Settings>("marketing/summary-email");
  const { request } = useMarketing();
  const [saving, setSaving] = useState(false);
  const [preview, setPreview] = useState<Preview | null>(null);
  const [loadingPreview, setLoadingPreview] = useState(false);
  const save = async (values: Partial<Settings>) => {
    if (!query.data) return;
    setSaving(true);
    try {
      await request("marketing/summary-email", "PUT", { frequency: query.data.frequency, includeMe: query.data.includeMe, ...values });
      await query.refetch();
      setPreview(null);
      toast.success("Summary email saved");
    } catch (error) { toast.error((error as Error).message); } finally { setSaving(false); }
  };
  const showPreview = async () => {
    setLoadingPreview(true);
    try { setPreview(await request<Preview>(`marketing/summary-email/preview?frequency=${query.data?.frequency === "daily" ? "daily" : "weekly"}`)); }
    catch (error) { toast.error((error as Error).message); } finally { setLoadingPreview(false); }
  };
  if (query.isPending || query.error) return <QueryState loading={query.isPending} error={query.error} retry={() => void query.refetch()} />;
  const { frequency, includeMe, isOwner, timezone } = query.data;
  const period = frequency === "daily" ? "yesterday" : "last week";
  return (
    <section className={workspaceClassName("product-panel")}>
      <h2 className="text-lg font-semibold">Summary email</h2>
      <p className="mt-1 text-sm text-muted-foreground">
        How sending went: emails sent, opens, clicks, bounces, complaints, new contacts, unsubscribes and your top campaign, compared with the period before. It goes to the workspace owner at 8:00 in {timezone}, and is skipped when nothing was sent and no subscribers changed.
      </p>
      <div className={`${workspaceClassName("product-form")} mt-6 max-w-2xl space-y-5`}>
        <Field label="Send it" hint="Weekly summaries arrive on Monday and cover Monday to Sunday. The time zone comes from Sending rules.">
          <select value={frequency} disabled={saving} onChange={(e) => void save({ frequency: e.target.value as Frequency })}>
            <option value="daily">Daily</option>
            <option value="weekly">Weekly</option>
            <option value="off">Off</option>
          </select>
        </Field>
        {!isOwner && (
          <label className="flex items-start gap-2 text-sm">
            <input className="!w-4" type="checkbox" disabled={saving} checked={includeMe} onChange={(e) => void save({ includeMe: e.target.checked })} />
            <span>Send me a copy too<small className="mt-1 block text-muted-foreground">Available to workspace admins.</small></span>
          </label>
        )}
        <Button type="button" variant="outline" disabled={loadingPreview} onClick={() => void showPreview()}>
          {loadingPreview ? "Loading preview…" : "Preview"}
        </Button>
      </div>
      {preview && (
        <div className="mt-6 space-y-3">
          <p className="text-sm"><strong>Subject:</strong> {preview.subject}</p>
          {preview.quiet && (
            <p className="text-sm text-muted-foreground">Nothing was sent and no subscribers changed {period}, so this one would be skipped.</p>
          )}
          <iframe title="Summary email preview" sandbox="" srcDoc={preview.html} className="h-[720px] w-full rounded-lg border" />
        </div>
      )}
    </section>
  );
}
