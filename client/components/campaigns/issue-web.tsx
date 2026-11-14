"use client";
import { useMemo, useState } from "react";
import { Bar, BarChart, ResponsiveContainer, Tooltip, XAxis } from "recharts";
import { ArrowUpRight, CalendarDays, Eye, Pencil, Sun, Trophy } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { AuthorSelect, useHasAuthors } from "@/components/marketing/authors";
import { liveUrl } from "@/components/marketing/newsletters";
import { EmailPreview, Field, Metric, Modal } from "@/components/marketing/shared";
import type { Campaign } from "@/lib";
import { useMarketing, useMarketingQuery } from "@/lib/marketing/api";
import type { Newsletter } from "@/lib/marketing/types";
import { diffLines, readableLines } from "@/lib/text-diff";
import { workspaceClassName } from "@/lib/workspace-styles";

type Page = { title: string; description: string; cover: string; html: string; updatedAt?: string };
type WebPage = { newsletter: Newsletter; emailed: Page; page: Page | null };
type Count = { key: string; views: number; visitors: number };
type Views = { total: number; today: number; week: number; best: Count | null; daily: Count[] };

const day = (key: string) =>
  new Date(`${key}T00:00:00Z`).toLocaleDateString(undefined, { month: "short", day: "numeric", timeZone: "UTC" });

/** Who signs the campaign; changing it after sending changes the web byline only. */
export function CampaignAuthor({ campaign, saved }: { campaign: Campaign; saved: () => unknown }) {
  const { request } = useMarketing();
  const hasAuthors = useHasAuthors();
  const [busy, setBusy] = useState(false);
  if (!hasAuthors) return null;
  const choose = async (authorId: string) => {
    setBusy(true);
    try {
      await request(`marketing/campaigns/${campaign.id}/author`, "PUT", { authorId });
      await saved();
      toast.success(campaign.status === "COMPLETED" ? "Author changed on the web page. The email stays as sent." : "Author saved");
    } catch (e) {
      toast.error((e as Error).message);
    } finally {
      setBusy(false);
    }
  };
  return (
    <label className="flex flex-wrap items-center gap-3 text-sm">
      <span className="text-muted-foreground">Author</span>
      <fieldset disabled={busy}>
        <AuthorSelect
          className="h-9 rounded-md border border-input bg-background px-3 text-sm"
          value={campaign.authorId || ""}
          onChange={(v) => void choose(v)}
        />
      </fieldset>
    </label>
  );
}

/** A sent issue's page on the newsletter website: its views, and editing it. */
export function IssueWebPage({ campaign }: { campaign: Campaign }) {
  const onSite = campaign.status === "COMPLETED" && !!(campaign.newsletterId || campaign.archiveNewsletterId);
  const page = useMarketingQuery<WebPage>(`marketing/campaigns/${campaign.id}/web-page`, onSite);
  const views = useMarketingQuery<Views>(`marketing/campaigns/${campaign.id}/web-views`, onSite && !!page.data);
  const { request } = useMarketing();
  const [editing, setEditing] = useState(false);
  if (!page.data) return null;
  const { newsletter, page: edited } = page.data;
  const url = `${liveUrl(newsletter)}/${campaign.id}`;
  const v = views.data;
  const restore = async () => {
    if (!confirm("Put the emailed version back on the website? Your edits will be removed.")) return;
    try {
      await request(`marketing/campaigns/${campaign.id}/web-page`, "DELETE");
      await page.refetch();
      toast.success("The web page shows the emailed version again");
    } catch (e) {
      toast.error((e as Error).message);
    }
  };
  return (
    <Card>
      <CardHeader>
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <CardTitle>Web page</CardTitle>
            <CardDescription>
              On {newsletter.name}’s website.{" "}
              {edited
                ? `Edited ${new Date(edited.updatedAt!).toLocaleString(undefined, { dateStyle: "medium", timeStyle: "short" })}. The email stays as sent.`
                : "It shows the email as sent."}
            </CardDescription>
          </div>
          <div className="flex flex-wrap gap-2">
            <Button size="sm" variant="outline" asChild>
              <a href={url} target="_blank" rel="noreferrer">
                View page
                <ArrowUpRight className="ml-1 size-4" />
              </a>
            </Button>
            {edited && (
              <Button size="sm" variant="ghost" onClick={() => void restore()}>
                Restore emailed version
              </Button>
            )}
            <Button size="sm" onClick={() => setEditing(true)}>
              <Pencil className="mr-2 size-4" />
              Edit web page
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {v ? (
          <>
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
              <Metric label="Views" value={v.total.toLocaleString()} icon={<Eye size={18} />} detail="All time" />
              <Metric label="Today" value={v.today.toLocaleString()} icon={<Sun size={18} />} />
              <Metric label="Last 7 days" value={v.week.toLocaleString()} icon={<CalendarDays size={18} />} />
              <Metric
                label="Best day"
                value={v.best ? v.best.views.toLocaleString() : "—"}
                icon={<Trophy size={18} />}
                detail={v.best ? day(v.best.key) : undefined}
              />
            </div>
            {v.daily.length ? (
              <div className="h-40">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={v.daily} margin={{ top: 4, right: 0, left: 0, bottom: 0 }}>
                    <XAxis dataKey="key" tickFormatter={day} tick={{ fontSize: 11 }} tickLine={false} axisLine={false} minTickGap={24} />
                    <Tooltip
                      cursor={{ fill: "var(--muted)" }}
                      contentStyle={{ fontSize: 12, borderRadius: 8 }}
                      labelFormatter={(k: string) => day(k)}
                      formatter={(n: number) => [n, "Views"]}
                    />
                    <Bar dataKey="views" fill="var(--chart-1, #6366f1)" radius={[4, 4, 0, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">No views yet. Share the page link and they’ll show up here.</p>
            )}
            <p className="text-xs text-muted-foreground">Views by day, in UTC.</p>
          </>
        ) : (
          <p className="text-sm text-muted-foreground" role="status">
            {views.error ? "Views couldn’t be loaded." : "Loading views…"}
          </p>
        )}
      </CardContent>
      {editing && (
        <WebPageEditor
          campaignId={campaign.id}
          data={page.data}
          close={() => setEditing(false)}
          saved={() => void page.refetch()}
        />
      )}
    </Card>
  );
}

function WebPageEditor({
  campaignId,
  data,
  close,
  saved,
}: {
  campaignId: string;
  data: WebPage;
  close: () => void;
  saved: () => void;
}) {
  const { request } = useMarketing();
  const emailed = data.emailed;
  const [draft, setDraft] = useState<Page>(data.page ?? emailed);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [showEmailed, setShowEmailed] = useState(false);
  const set = (key: keyof Page, v: string) => setDraft((d) => ({ ...d, [key]: v }));
  const content = useMemo(
    () => diffLines(readableLines(emailed.html), readableLines(draft.html)),
    [emailed.html, draft.html],
  );
  const changed = (
    [
      ["Title", emailed.title, draft.title],
      ["Description", emailed.description, draft.description],
      ["Card image", emailed.cover, draft.cover],
    ] as const
  ).filter(([, a, b]) => a.trim() !== b.trim());
  const contentChanged = content.some((l) => l.op !== " ");

  async function save(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      await request(`marketing/campaigns/${campaignId}/web-page`, "PUT", draft);
      saved();
      toast.success("Web page updated. The email stays as sent.");
      close();
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <Modal open onOpenChange={close} title="Edit web page" description="Changes show on your newsletter website only. Readers who got the email keep the version that was sent." wide>
      <form className={workspaceClassName("product-form")} onSubmit={save}>
        <div className={workspaceClassName("editor-columns")}>
          <div className={workspaceClassName("product-form")}>
            <Field label="Title">
              <input required maxLength={200} value={draft.title} onChange={(e) => set("title", e.target.value)} />
            </Field>
            <Field label="Description (optional)" hint="Shown in search results and link previews.">
              <textarea maxLength={300} rows={2} value={draft.description} onChange={(e) => set("description", e.target.value)} />
            </Field>
            <Field label="Card image (optional)" hint="An https:// image for the issue’s card and link previews. Empty uses the first image in the content.">
              <input type="url" maxLength={500} placeholder="https://" value={draft.cover} onChange={(e) => set("cover", e.target.value)} />
            </Field>
            <Field label="Content (HTML)">
              <textarea required rows={14} className="font-mono text-xs" value={draft.html} onChange={(e) => set("html", e.target.value)} />
            </Field>
          </div>
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-semibold">{showEmailed ? "Emailed version" : "Preview"}</h3>
              <button type="button" className="text-xs text-muted-foreground underline" onClick={() => setShowEmailed((s) => !s)}>
                {showEmailed ? "Show your edit" : "Show the emailed version"}
              </button>
            </div>
            <EmailPreview html={showEmailed ? emailed.html : draft.html} title={showEmailed ? "Emailed version" : "Web page preview"} />
          </div>
        </div>

        <section className="space-y-2 rounded-xl border p-4" aria-labelledby="web-page-changes">
          <h3 id="web-page-changes" className="text-sm font-semibold">
            Changes from the emailed version
          </h3>
          {!changed.length && !contentChanged && <p className="text-sm text-muted-foreground">No changes yet.</p>}
          {changed.map(([label, before, after]) => (
            <div key={label} className="text-sm">
              <p className="font-medium">{label}</p>
              <p className="break-words bg-red-500/10 px-2 py-0.5 text-red-700 line-through dark:text-red-400">{before || "(none)"}</p>
              <p className="break-words bg-green-500/10 px-2 py-0.5 text-green-700 dark:text-green-400">{after || "(none)"}</p>
            </div>
          ))}
          {contentChanged && (
            <div className="text-sm">
              <p className="font-medium">Content</p>
              <ol className="mt-1 max-h-72 overflow-auto rounded-md border font-mono text-xs">
                {content.map((l, i) => (
                  <li
                    key={i}
                    className={
                      l.op === "+"
                        ? "bg-green-500/10 px-2 py-0.5 text-green-700 dark:text-green-400"
                        : l.op === "-"
                          ? "bg-red-500/10 px-2 py-0.5 text-red-700 line-through dark:text-red-400"
                          : "px-2 py-0.5 text-muted-foreground"
                    }
                  >
                    <span aria-hidden="true" className="mr-2 select-none">{l.op}</span>
                    <span className="sr-only">{l.op === "+" ? "Added: " : l.op === "-" ? "Removed: " : ""}</span>
                    {l.text}
                  </li>
                ))}
              </ol>
            </div>
          )}
        </section>

        {error && (
          <p role="alert" className="text-sm text-destructive">
            {error}
          </p>
        )}
        <div className="flex justify-end gap-2">
          <Button type="button" variant="ghost" onClick={close}>
            Cancel
          </Button>
          <Button type="submit" disabled={busy}>
            {busy ? "Saving…" : "Save web page"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
