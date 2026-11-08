"use client";
import { workspaceClassName } from "@/lib/workspace-styles";
import { useState } from "react";
import Link from "next/link";
import {
  Plus,
  Newspaper,
  CalendarDays,
  Send,
  Repeat2,
  MoreHorizontal,
  Pencil,
  Copy,
  Pause,
  ArrowUpRight,
  Eye,
  Clock,
  Users,
  Globe,
  Link2,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { GmailClipNotice } from "@/components/campaigns/gmail-clip-notice";
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from "@/components/ui/dropdown-menu";
import { useMarketing, useMarketingQuery } from "@/lib/marketing/api";
import type { Options, Newsletter } from "@/lib/marketing/types";
import {
  PageHeading,
  Metric,
  QueryState,
  Empty,
  Modal,
  Field,
  EmailPreview,
  FilterTabs,
  Status,
} from "./shared";
import { toast } from "sonner";
import { SiteAnalytics } from "./site-analytics";
import { PaidPlanNotice, usePaidPlan } from "@/components/billing/paid-plan";

// Newsletter websites live on the marketing site, not in the app.
const siteUrl = (
  process.env.NEXT_PUBLIC_SITE_URL || "https://xem.email"
).replace(/\/$/, "");
export function NewslettersPage() {
  const q = useMarketingQuery<Newsletter[]>("marketing/newsletters");
  const options = useMarketingQuery<Options>("marketing/options");
  const { request, refresh } = useMarketing();
  const [tab, setTab] = useState("All newsletters");
  const [editing, setEditing] = useState<Partial<Newsletter> | null>(null);
  const [history, setHistory] = useState<Newsletter | null>(null);
  const [site, setSite] = useState<Newsletter | null>(null);
  const [stats, setStats] = useState<Newsletter | null>(null);
  const rows = q.data || [];
  // Newsletters, and their websites, come with Starter and up.
  const paid = usePaidPlan();
  async function pause(n: Newsletter) {
    try {
      await request(`marketing/newsletters/${n.id}/pause`, "POST", {});
      await refresh();
      toast.success("Future editions paused");
    } catch (e) {
      toast.error((e as Error).message);
    }
  }
  return (
    <>
      <PageHeading
        title="Newsletters"
        description="Good stories deserve a regular place in the inbox."
        action={
          paid ? (
            <Button
              className={workspaceClassName("product-primary")}
              onClick={() => setEditing({})}
            >
              <Plus />
              Create Newsletter
            </Button>
          ) : undefined
        }
      />
      {!paid && (
        <div className="mb-6">
          <PaidPlanNotice feature="A newsletter with its own website" />
        </div>
      )}
      <div className={workspaceClassName("metrics-grid")}>
        <Metric
          label="All Newsletters"
          value={rows.length.toString().padStart(2, "0")}
          icon={<Newspaper />}
        />
        <Metric
          label="Scheduled"
          value={rows
            .filter((n) => n.status === "SCHEDULED")
            .length.toString()
            .padStart(2, "0")}
          icon={<CalendarDays />}
        />
        <Metric
          label="Editions Created"
          value={rows.reduce((s, n) => s + n.editions, 0)}
          icon={<Send />}
        />
        <Metric
          label="Recurring"
          value={rows
            .filter((n) => n.cadence !== "ONCE")
            .length.toString()
            .padStart(2, "0")}
          icon={<Repeat2 />}
        />
      </div>
      {(!paid || !rows.some((n) => n.publicSlug)) && !q.isLoading && (
        <WebsiteBanner
          paid={paid}
          setUp={() => {
            const first = rows.find((n) => !n.publicSlug);
            if (first) setSite(first);
            else setEditing({});
          }}
          hasNewsletter={rows.length > 0}
        />
      )}
      <section className={workspaceClassName("product-panel")}>
        <div className={workspaceClassName("panel-toolbar")}>
          <h2>Your newsletters</h2>
          <FilterTabs
            items={["All newsletters", "Scheduled", "Drafts", "Paused"]}
            value={tab}
            onChange={setTab}
          />
        </div>
        <QueryState
          loading={q.isLoading}
          error={q.error}
          retry={() => q.refetch()}
        />
        {!q.isLoading && !q.error && rows.length === 0 ? (
          <Empty
            title="Something worth opening"
            description="Connect a template, choose your audience, and send your first edition. Once, weekly, or whenever inspiration strikes."
            action={
              paid ? (
                <Button
                  className={workspaceClassName("product-primary")}
                  onClick={() => setEditing({})}
                >
                  <Plus />
                  Create a newsletter
                </Button>
              ) : undefined
            }
          />
        ) : (
          <div className={workspaceClassName("table-wrap")}>
            <table className={workspaceClassName("product-table")}>
              <thead>
                <tr>
                  <th>Newsletter</th>
                  <th>Status</th>
                  <th>Audience</th>
                  <th>Next edition</th>
                  <th>Cadence</th>
                  <th>Website</th>
                  <th />
                </tr>
              </thead>
              <tbody>
                {rows
                  .filter(
                    (n) =>
                      tab === "All newsletters" ||
                      n.status ===
                        (tab === "Drafts" ? "DRAFT" : tab.toUpperCase()),
                  )
                  .map((n) => (
                    <tr key={n.id}>
                      <td>
                        <button
                          className={workspaceClassName(
                            "newsletter-title text-left",
                          )}
                          onClick={() => setEditing(n)}
                        >
                          <span>
                            <Newspaper size={18} />
                          </span>
                          <div>
                            <strong>{n.name}</strong>
                            <small>{n.subject}</small>
                          </div>
                        </button>
                      </td>
                      <td>
                        <Status value={n.status} />
                      </td>
                      <td>
                        {options.data?.lists.find((l) => l.id === n.listId)
                          ?.name || "Audience"}
                      </td>
                      <td>
                        {n.nextSendAt ? (
                          <>
                            <strong className="!text-xs">
                              {new Date(n.nextSendAt).toLocaleDateString(
                                undefined,
                                { month: "short", day: "numeric" },
                              )}
                            </strong>
                            <small>
                              {new Date(n.nextSendAt).toLocaleTimeString(
                                undefined,
                                {
                                  hour: "2-digit",
                                  minute: "2-digit",
                                  timeZone: n.timezone,
                                },
                              )}{" "}
                              · {n.timezone}
                            </small>
                          </>
                        ) : (
                          "—"
                        )}
                      </td>
                      <td>
                        {n.cadence === "ONCE"
                          ? "One time"
                          : n.cadence.charAt(0) +
                            n.cadence.slice(1).toLowerCase()}
                      </td>
                      <td>
                        {n.publicSlug ? (
                          <div className="flex flex-col items-start gap-0.5">
                            <a
                              href={liveUrl(n)}
                              target="_blank"
                              rel="noreferrer"
                              className="inline-flex items-center gap-1.5 whitespace-nowrap text-xs font-medium hover:underline"
                            >
                              <span
                                className="size-1.5 rounded-full bg-emerald-500"
                                aria-hidden
                              />
                              {liveUrl(n).replace(/^https?:\/\//, "")}
                              <ArrowUpRight size={12} />
                            </a>
                            <span className="flex gap-3 text-xs text-muted-foreground">
                              <button
                                className="hover:text-foreground"
                                onClick={() => setStats(n)}
                              >
                                Analytics
                              </button>
                              <button
                                className="hover:text-foreground"
                                onClick={() => setSite(n)}
                              >
                                Settings
                              </button>
                            </span>
                          </div>
                        ) : paid ? (
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => setSite(n)}
                          >
                            <Globe />
                            Set up
                          </Button>
                        ) : (
                          "—"
                        )}
                      </td>
                      <td>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <button
                              className={workspaceClassName("icon-button")}
                              aria-label={`Actions for ${n.name}`}
                            >
                              <MoreHorizontal />
                            </button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem onClick={() => setEditing(n)}>
                              Edit newsletter
                            </DropdownMenuItem>
                            {paid && (
                              <DropdownMenuItem onClick={() => setSite(n)}>
                                Website
                              </DropdownMenuItem>
                            )}
                            {n.publicSlug && (
                              <DropdownMenuItem onClick={() => setStats(n)}>
                                Website analytics
                              </DropdownMenuItem>
                            )}
                            <DropdownMenuItem
                              onClick={() =>
                                setEditing({
                                  ...n,
                                  id: undefined,
                                  name: `${n.name} (copy)`,
                                  status: "DRAFT",
                                  nextSendAt: null,
                                  editions: 0,
                                })
                              }
                            >
                              Duplicate as draft
                            </DropdownMenuItem>
                            <DropdownMenuItem onClick={() => setHistory(n)}>
                              Edition history
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              disabled={n.status !== "SCHEDULED"}
                              onClick={() => pause(n)}
                            >
                              Pause future editions
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
      {editing && (
        <NewsletterEditor
          key={editing.id || "new"}
          value={editing}
          options={options.data}
          close={() => setEditing(null)}
        />
      )}{" "}
      {history && (
        <EditionHistory newsletter={history} close={() => setHistory(null)} />
      )}
      {stats && (
        <SiteAnalytics
          key={stats.id}
          newsletter={stats}
          url={liveUrl(stats)}
          close={() => setStats(null)}
        />
      )}
      {site && (
        <NewsletterWebsite
          key={site.id}
          newsletter={site}
          close={() => setSite(null)}
        />
      )}
    </>
  );
}
function localDate(iso?: string | null) {
  if (!iso) return "";
  const d = new Date(iso);
  return new Date(d.getTime() - d.getTimezoneOffset() * 60000)
    .toISOString()
    .slice(0, 16);
}
function NewsletterEditor({
  value,
  options,
  close,
}: {
  value: Partial<Newsletter>;
  options?: Options;
  close: () => void;
}) {
  const { request, refresh } = useMarketing();
  const [draft, setDraft] = useState({
    name: value.name || "",
    subject: value.subject || "",
    description: value.description || "",
    templateId: value.templateId || "",
    listId: value.listId || "",
    smtpConfigId: value.smtpConfigId || "",
    cadence: value.cadence || "ONCE",
    timezone:
      value.timezone || Intl.DateTimeFormat().resolvedOptions().timeZone,
    postalAddress: value.postalAddress || "",
    trackEngagement: value.trackEngagement ?? true,
  });
  const [at, setAt] = useState(localDate(value.nextSendAt));
  const [status, setStatus] = useState(
    value.status === "SCHEDULED" ? "SCHEDULED" : "DRAFT",
  );
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [review, setReview] = useState(false);
  const template = options?.templates.find((t) => t.id === draft.templateId);
  const preview = useMarketingQuery<{ htmlBody: string }>(
    `marketing/templates/${draft.templateId}/preview`,
    !!draft.templateId && !template?.htmlBody,
  );
  const templateHTML = template?.htmlBody || preview.data?.htmlBody;
  const set = (key: string, val: string | boolean) => {
    setDraft((d) => ({ ...d, [key]: val }));
    setReview(false);
  };
  async function save(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    if (status === "SCHEDULED" && !review) {
      setReview(true);
      return;
    }
    setBusy(true);
    try {
      await request(
        `marketing/newsletters${value.id ? `/${value.id}` : ""}`,
        value.id ? "PUT" : "POST",
        {
          ...draft,
          status,
          nextSendAt: at ? new Date(at).toISOString() : null,
        },
      );
      await refresh();
      toast.success(
        status === "SCHEDULED" ? "Newsletter scheduled" : "Draft saved",
      );
      close();
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }
  return (
    <Modal
      open
      onOpenChange={close}
      title={value.id ? "Edit newsletter" : "Create your newsletter"}
      description="A reusable template, the right audience, and a moment that works."
      wide
    >
      <form className={workspaceClassName("product-form")} onSubmit={save}>
        <div className={workspaceClassName("editor-columns")}>
          <div className={workspaceClassName("product-form")}>
            <Field label="Newsletter name">
              <input
                required
                minLength={2}
                maxLength={120}
                placeholder="e.g. The Sunday Edit"
                value={draft.name}
                onChange={(e) => set("name", e.target.value)}
              />
            </Field>
            <Field label="Email template">
              <select
                required
                value={draft.templateId}
                onChange={(e) => {
                  set("templateId", e.target.value);
                  if (!draft.subject)
                    set(
                      "subject",
                      options?.templates.find((t) => t.id === e.target.value)
                        ?.subject || "",
                    );
                }}
              >
                <option value="">Choose a reusable template</option>
                {options?.templates.map((t) => (
                  <option key={t.id} value={t.id}>
                    {t.name}
                  </option>
                ))}
              </select>
              <Link href="/templates" className="text-primary text-xs">
                Explore starter templates ↗
              </Link>
            </Field>
            <Field label="Subject line">
              <input
                required
                maxLength={200}
                placeholder="Give them a reason to open"
                value={draft.subject}
                onChange={(e) => set("subject", e.target.value)}
              />
            </Field>
            <div className={workspaceClassName("form-row")}>
              <Field label="Audience">
                <select
                  required
                  value={draft.listId}
                  onChange={(e) => set("listId", e.target.value)}
                >
                  <option value="">Choose audience</option>
                  {options?.lists.map((l) => (
                    <option value={l.id} key={l.id}>
                      {l.name}
                    </option>
                  ))}
                </select>
              </Field>
              <Field label="Sender">
                <select
                  required
                  value={draft.smtpConfigId}
                  onChange={(e) => set("smtpConfigId", e.target.value)}
                >
                  <option value="">Choose sender</option>
                  {options?.senders.map((s) => (
                    <option value={s.id} key={s.id}>
                      {s.fromEmail}
                    </option>
                  ))}
                </select>
              </Field>
            </div>
            <div className={workspaceClassName("form-row")}>
              <Field label="Frequency">
                <select
                  value={draft.cadence}
                  onChange={(e) => set("cadence", e.target.value)}
                >
                  <option value="ONCE">One time</option>
                  <option value="DAILY">Daily</option>
                  <option value="WEEKLY">Weekly</option>
                  <option value="MONTHLY">Monthly</option>
                </select>
              </Field>
              <Field label="Delivery">
                <select
                  value={status}
                  onChange={(e) => {
                    setStatus(e.target.value);
                    setReview(false);
                  }}
                >
                  <option value="DRAFT">Save as draft</option>
                  <option value="SCHEDULED">Schedule delivery</option>
                </select>
              </Field>
            </div>
            <Field
              label={`First send · ${Intl.DateTimeFormat().resolvedOptions().timeZone}`}
              hint="Choose a time in your device’s timezone. Future recurring editions follow the timezone below."
            >
              <input
                type="datetime-local"
                required={status === "SCHEDULED"}
                min={localDate(new Date(Date.now() + 60000).toISOString())}
                value={at}
                onChange={(e) => {
                  setAt(e.target.value);
                  setReview(false);
                }}
              />
            </Field>
            <Field label="Recurring timezone">
              <input
                required
                value={draft.timezone}
                onChange={(e) => set("timezone", e.target.value)}
                placeholder="Asia/Kolkata"
              />
            </Field>
            <Field
              label="Sender postal address"
              hint="Included in every edition with an unsubscribe link."
            >
              <input
                required={status === "SCHEDULED"}
                maxLength={500}
                value={draft.postalAddress}
                onChange={(e) => set("postalAddress", e.target.value)}
                placeholder="Company, street, city, postal code, country"
              />
            </Field>
            <label className="flex items-start gap-2 text-sm">
              <input
                type="checkbox"
                className="mt-1 !w-4"
                checked={draft.trackEngagement}
                onChange={(e) => set("trackEngagement", e.target.checked)}
              />
              <span>
                Track opens and link clicks
                <span className="block text-xs text-muted-foreground">
                  Off sends editions without a tracking pixel or rewritten
                  links. Deliveries, bounces, complaints and unsubscribes are
                  still recorded.
                </span>
              </span>
            </label>
          </div>
          <div className={workspaceClassName("editor-preview")}>
            <div className={workspaceClassName("editor-preview-label")}>
              {template?.name || "Your email preview"}
            </div>
            {templateHTML ? (
              <EmailPreview html={templateHTML} />
            ) : (
              <Empty
                title="Choose your starting point"
                description="Select a template to preview the email your audience will receive."
              />
            )}
            <div className="p-5 text-xs text-muted-foreground leading-relaxed">
              {draft.cadence !== "ONCE"
                ? "Each edition uses the latest saved template. You can pause future editions at any time."
                : "Your template is captured when the scheduled edition is prepared."}
            </div>
          </div>
        </div>
        {review && (
          <div className="rounded-xl border border-violet-200 bg-violet-50 p-4 text-sm leading-7">
            <strong>Ready to schedule?</strong>
            <p>
              “{draft.subject}” will go to active subscribers in{" "}
              {options?.lists.find((l) => l.id === draft.listId)?.name} from{" "}
              {
                options?.senders.find((s) => s.id === draft.smtpConfigId)
                  ?.fromEmail
              }
              . First send: {at && new Date(at).toLocaleString()}. Frequency:{" "}
              {draft.cadence.toLowerCase()}.
            </p>
            <GmailClipNotice html={templateHTML} className="mt-2" />
          </div>
        )}
        {error && (
          <div role="alert" className={workspaceClassName("product-error")}>
            {error}
          </div>
        )}
        <div className={workspaceClassName("modal-actions")}>
          <Button variant="outline" type="button" onClick={close}>
            Cancel
          </Button>
          <Button
            className={workspaceClassName("product-primary")}
            disabled={busy}
          >
            {busy
              ? "Saving…"
              : status === "DRAFT"
                ? "Save draft"
                : review
                  ? "Confirm schedule"
                  : "Review schedule"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
function EditionHistory({
  newsletter,
  close,
}: {
  newsletter: Newsletter;
  close: () => void;
}) {
  const q = useMarketingQuery<any[]>(
    `marketing/newsletters/${newsletter.id}/editions`,
  );
  return (
    <Modal
      open
      onOpenChange={close}
      title="Edition history"
      description={`${newsletter.name} · Most recent 50 editions. Delivery totals update as messages are processed.`}
    >
      <QueryState loading={q.isLoading} error={q.error} />
      {q.data?.length === 0 && (
        <Empty
          title="Your first edition is still ahead"
          description="Scheduled editions appear here when they are prepared for delivery."
        />
      )}
      <table className={workspaceClassName("product-table")}>
        <thead>
          <tr>
            <th>Edition</th>
            <th>Prepared</th>
            <th>Sent</th>
            <th>Failed</th>
            <th>In flight / review</th>
            <th>Status</th>
          </tr>
        </thead>
        <tbody>
          {q.data?.map((e) => (
            <tr key={e.id}>
              <td>{new Date(e.scheduledFor).toLocaleString()}</td>
              <td>{e.processed}</td>
              <td>{e.sent}</td>
              <td>{e.failed}</td>
              <td>{e.needsReview || 0}</td>
              <td>
                <EditionStatus edition={e} reload={q.refetch} />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </Modal>
  );
}
function EditionStatus({
  edition,
  reload,
}: {
  edition: any;
  reload: () => unknown;
}) {
  const { request } = useMarketing();
  const [busy, setBusy] = useState(false);
  const act = async (action: "resume" | "cancel") => {
    setBusy(true);
    try {
      await request(`marketing/campaigns/${edition.id}/${action}`, "POST", {});
      toast.success(
        action === "resume"
          ? "Edition resumed. The rest of the list sends now."
          : "Edition cancelled. The rest of the list won’t be sent.",
      );
      await reload();
    } catch (e) {
      toast.error((e as Error).message);
    } finally {
      setBusy(false);
    }
  };
  if (edition.status === "PAUSED")
    return (
      <div role="alert" className="max-w-xs text-xs">
        <p>{edition.pauseReason || "Paused"}</p>
        <div className="mt-2 flex gap-2">
          <Button
            size="sm"
            variant="outline"
            disabled={busy}
            onClick={() => void act("resume")}
          >
            Resume
          </Button>
          <Button
            size="sm"
            variant="ghost"
            disabled={busy}
            onClick={() => void act("cancel")}
          >
            Cancel the rest
          </Button>
        </div>
      </div>
    );
  if (edition.status === "SENDING" && edition.sendInStages)
    return <span>Sending in stages</span>;
  return <Status value={edition.status} />;
}

/** Where a newsletter's website is live: its own domain once active, else xem.email/{slug}. */
function liveUrl(n: Newsletter) {
  return n.siteDomain && n.siteDomainStatus === "active"
    ? `https://${n.siteDomain}`
    : `${siteUrl}/${n.publicSlug}`;
}

const siteHost = siteUrl.replace(/^https?:\/\//, "");

/** Introduces newsletter websites until the workspace has published one. */
function WebsiteBanner({
  paid,
  hasNewsletter,
  setUp,
}: {
  paid: boolean;
  hasNewsletter: boolean;
  setUp: () => void;
}) {
  return (
    <section className="relative mb-6 overflow-hidden rounded-2xl bg-foreground text-background">
      <div
        aria-hidden
        className="pointer-events-none absolute -right-24 -top-24 size-80 rounded-full bg-violet-500/30 blur-3xl"
      />
      <div className="relative grid items-center gap-8 p-6 sm:p-8 md:grid-cols-[1.15fr_1fr]">
        <div>
          <span className="inline-flex items-center gap-1.5 rounded-full border border-background/20 px-2.5 py-1 text-xs font-medium text-background/80">
            <Globe size={13} />
            Newsletter websites
          </span>
          <h2 className="mt-4 max-w-md text-2xl font-semibold leading-tight tracking-tight sm:text-3xl">
            Give your newsletter a home on the web
          </h2>
          <p className="mt-3 max-w-md text-sm leading-relaxed text-background/70">
            Every sent edition, published at {siteHost}/your-name or your own
            domain, with a subscribe box readers confirm by email. Pick from{" "}
            {siteTemplates.length} templates.
          </p>
          <div className="mt-6 flex flex-wrap items-center gap-3">
            {paid ? (
              <Button variant="secondary" onClick={setUp}>
                {hasNewsletter ? <Globe /> : <Plus />}
                {hasNewsletter ? "Set up a website" : "Create a newsletter"}
              </Button>
            ) : (
              <>
                <Button variant="secondary" asChild>
                  <Link href="/settings/billing">See plans</Link>
                </Button>
                <span className="text-xs text-background/60">
                  Included with Starter and up, 7-day free trial.
                </span>
              </>
            )}
          </div>
        </div>
        <div aria-hidden className="relative hidden h-56 md:block">
          {(["zine", "paper", "midnight"] as const).map((id, i) => (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              key={id}
              src={`/assets/site-templates/${id}.webp`}
              alt=""
              width={320}
              height={200}
              className="absolute aspect-[16/10] w-[62%] rounded-xl border border-background/15 object-cover object-top shadow-2xl"
              style={{
                left: `${4 + i * 17}%`,
                top: `${30 - i * 14}%`,
                transform: `rotate(${[-6, 0, 5][i]}deg)`,
                zIndex: i,
              }}
            />
          ))}
        </div>
      </div>
    </section>
  );
}

const suggestSlug = (name: string) =>
  name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 48);

/** Publishes and styles one newsletter's website, apart from its sending settings. */
function NewsletterWebsite({
  newsletter: n,
  close,
}: {
  newsletter: Newsletter;
  close: () => void;
}) {
  const { request, refresh } = useMarketing();
  const live = !!n.publicSlug;
  const [slug, setSlug] = useState(n.publicSlug || suggestSlug(n.name));
  const [theme, setTheme] = useState(n.siteTheme || "paper");
  const [logo, setLogo] = useState(n.siteLogoUrl || "");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const name = siteTemplates.find(([id]) => id === theme)?.[1];
  const address = live ? liveUrl(n) : `${siteUrl}/${slug.toLowerCase()}`;
  async function save(publicSlug: string) {
    setBusy(true);
    setError("");
    try {
      await request(`marketing/newsletters/${n.id}/site`, "PUT", {
        publicSlug,
        siteTheme: theme,
        siteLogoUrl: logo,
      });
      await refresh();
      if (!publicSlug) toast.success("Website taken down");
      else
        toast.success(live ? "Website updated" : "Your website is live", {
          action: {
            label: "Visit",
            onClick: () =>
              window.open(`${siteUrl}/${publicSlug.toLowerCase()}`, "_blank"),
          },
        });
      close();
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }
  return (
    <Modal
      open
      onOpenChange={close}
      title={live ? "Website" : "Set up a website"}
      description={`${n.name}: every sent edition and a subscribe box, on the web.`}
      wide
    >
      <form
        className={workspaceClassName("product-form")}
        onSubmit={(e) => {
          e.preventDefault();
          void save(slug);
        }}
      >
        <div className={workspaceClassName("editor-columns")}>
          <div className="space-y-5">
            <Field
              label="Web address"
              hint="3 to 48 letters, numbers or hyphens."
            >
              <div className="flex items-center overflow-hidden rounded-lg border border-[var(--border-strong)] bg-[var(--card)] focus-within:border-[var(--ring)]">
                <span className="shrink-0 border-r bg-muted px-3 py-2 text-sm text-muted-foreground">
                  {siteHost}/
                </span>
                <input
                  required
                  value={slug}
                  onChange={(e) =>
                    setSlug(e.target.value.replace(/\s+/g, "-"))
                  }
                  pattern="[a-zA-Z0-9][a-zA-Z0-9-]{1,46}[a-zA-Z0-9]"
                  title="3 to 48 letters, numbers or hyphens"
                  maxLength={48}
                  placeholder="weekly-notes"
                  className="!rounded-none !border-0 !outline-none"
                />
              </div>
            </Field>
            <Field label="Template">
              <SiteTemplatePicker value={theme} onChange={setTheme} />
            </Field>
            <Field
              label="Logo (optional)"
              hint="An https:// image, square works best."
            >
              <input
                type="url"
                value={logo}
                onChange={(e) => setLogo(e.target.value)}
                pattern="https://.*"
                maxLength={500}
                placeholder="https://example.com/logo.png"
              />
            </Field>
            {live ? (
              <SiteDomain newsletterId={n.id} />
            ) : (
              <p className="text-xs text-muted-foreground">
                You can connect your own domain, like news.example.com, once
                the website is live.
              </p>
            )}
          </div>
          <div
            className={`${workspaceClassName("editor-preview")} self-start md:sticky md:top-0`}
          >
            <div className="flex items-center gap-2 border-b px-4 py-3">
              <span className="flex gap-1.5" aria-hidden>
                <span className="size-2.5 rounded-full bg-border" />
                <span className="size-2.5 rounded-full bg-border" />
                <span className="size-2.5 rounded-full bg-border" />
              </span>
              <span className="min-w-0 flex-1 truncate rounded-md bg-muted px-3 py-1 text-xs text-muted-foreground">
                {address.replace(/^https?:\/\//, "")}
              </span>
              {live && (
                <>
                  <button
                    type="button"
                    className={workspaceClassName("icon-button")}
                    aria-label="Copy link"
                    onClick={() => {
                      void navigator.clipboard.writeText(address);
                      toast.success("Link copied");
                    }}
                  >
                    <Link2 size={15} />
                  </button>
                  <a
                    href={address}
                    target="_blank"
                    rel="noreferrer"
                    className={workspaceClassName("icon-button")}
                    aria-label="Visit website"
                  >
                    <ArrowUpRight size={15} />
                  </a>
                </>
              )}
            </div>
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={`/assets/site-templates/${theme}.webp`}
              alt={`${name} template`}
              className="aspect-[16/10] w-full object-cover object-top"
            />
            <div className="flex items-center gap-2 p-5 text-xs text-muted-foreground">
              <span
                className={`size-1.5 rounded-full ${live ? "bg-emerald-500" : "bg-muted-foreground/40"}`}
                aria-hidden
              />
              {live
                ? `Live with the ${name} template.`
                : `${name} template. Not published yet.`}
            </div>
          </div>
        </div>
        {error && (
          <p role="alert" className="text-sm text-destructive">
            {error}
          </p>
        )}
        <div className="flex flex-wrap items-center justify-end gap-3">
          {live && (
            <Button
              type="button"
              variant="ghost"
              disabled={busy}
              className="mr-auto"
              onClick={() => {
                if (confirm("Take this website down? Its address stops working."))
                  void save("");
              }}
            >
              Take website down
            </Button>
          )}
          <Button type="button" variant="outline" onClick={close}>
            Cancel
          </Button>
          <Button
            type="submit"
            disabled={busy}
            className={workspaceClassName("product-primary")}
          >
            <Globe />
            {live ? "Save changes" : "Publish website"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}

type DomainState = {
  domain: string;
  status: string;
  records?: { type: string; name: string; value: string }[];
};

/** Connects the team's own domain to a saved newsletter's website. */
function SiteDomain({ newsletterId }: { newsletterId: string }) {
  const { request } = useMarketing();
  const path = `marketing/newsletters/${newsletterId}/domain`;
  const state = useMarketingQuery<DomainState>(path);
  const [domain, setDomain] = useState("");
  const [busy, setBusy] = useState(false);
  const current = state.data;
  const save = async (next: string) => {
    setBusy(true);
    try {
      await request(path, "PUT", { domain: next });
      await state.refetch();
      setDomain("");
    } catch (e) {
      toast.error(
        e instanceof Error ? e.message : "Couldn't update the domain",
      );
    } finally {
      setBusy(false);
    }
  };
  if (current?.domain)
    return (
      <Field
        label="Custom domain"
        hint={
          current.status === "active"
            ? `Live at https://${current.domain}.`
            : "Add these DNS records at your domain provider. The certificate is issued a few minutes after they resolve."
        }
      >
        <div className="space-y-2 text-sm">
          <div className="flex items-center justify-between gap-3">
            <span className="font-medium">{current.domain}</span>
            <span className="text-xs text-muted-foreground">
              {current.status === "active" ? "Connected" : "Waiting for DNS"}
            </span>
          </div>
          {current.status !== "active" &&
            current.records?.map((r) => (
              <div
                key={r.type + r.name}
                className="grid grid-cols-[56px_1fr] gap-x-2 rounded-md border p-2 font-mono text-xs"
              >
                <span>{r.type}</span>
                <span className="break-all">{r.name}</span>
                <span />
                <span className="break-all text-muted-foreground">
                  {r.value}
                </span>
              </div>
            ))}
          <div className="flex gap-3">
            <button
              type="button"
              className="underline"
              disabled={busy}
              onClick={() => state.refetch()}
            >
              Check again
            </button>
            <button
              type="button"
              className="underline"
              disabled={busy}
              onClick={() => save("")}
            >
              Disconnect
            </button>
          </div>
        </div>
      </Field>
    );
  return (
    <Field
      label="Custom domain (optional)"
      hint="Serve the website on your own domain, like news.example.com."
    >
      <div className="flex gap-2">
        <input
          value={domain}
          onChange={(e) => setDomain(e.target.value)}
          placeholder="news.example.com"
          maxLength={253}
        />
        <button
          type="button"
          className="shrink-0 underline"
          disabled={busy || !domain.trim()}
          onClick={() => save(domain)}
        >
          Connect
        </button>
      </div>
    </Field>
  );
}

// Newsletter website templates, matching website/worker/templates.
const siteTemplates = [
  ["paper", "Paper", "Editorial journal, featured issue"],
  ["broadsheet", "Broadsheet", "A newspaper front page"],
  ["midnight", "Midnight", "Dark, dotted grid, cards"],
  ["atelier", "Atelier", "Magazine spread, numbered"],
  ["letter", "Letter", "One column of text"],
  ["index", "Index", "Huge nameplate, dated rows"],
  ["bento", "Bento", "Everything in tiles"],
  ["gallery", "Gallery", "Covers in a masonry"],
  ["cinema", "Cinema", "Full-screen latest cover"],
  ["creator", "Creator", "Personal page, big avatar"],
  ["soft", "Soft", "Pastel mesh, glass cards"],
  ["zine", "Zine", "Loud, inky, stickers"],
  ["terminal", "Terminal", "A shell session"],
  ["almanac", "Almanac", "Classic book contents"],
] as const;

/** Thumbnail grid for choosing a newsletter website template. */
function SiteTemplatePicker({
  value,
  onChange,
}: {
  value: string;
  onChange: (value: string) => void;
}) {
  return (
    <div
      role="radiogroup"
      aria-label="Website template"
      className="grid grid-cols-2 gap-3 sm:grid-cols-3"
    >
      {siteTemplates.map(([id, name, about]) => {
        const checked = value === id;
        return (
          <label
            key={id}
            className={`group relative cursor-pointer rounded-xl border p-1.5 transition has-[:focus-visible]:outline has-[:focus-visible]:outline-2 has-[:focus-visible]:outline-offset-2 has-[:focus-visible]:outline-foreground ${
              checked
                ? "border-foreground ring-2 ring-foreground/15"
                : "border-border hover:border-foreground/40"
            }`}
          >
            <input
              type="radio"
              name="siteTheme"
              value={id}
              checked={checked}
              onChange={() => onChange(id)}
              className="sr-only"
            />
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={`/assets/site-templates/${id}.webp`}
              alt=""
              width={320}
              height={200}
              loading="lazy"
              className="aspect-[16/10] w-full rounded-lg border border-border/60 object-cover object-top"
            />
            <span className="mt-2 flex items-center justify-between px-1 text-sm font-medium text-foreground">
              {name}
              {checked && (
                <span
                  className="size-2 rounded-full bg-foreground"
                  aria-hidden
                />
              )}
            </span>
            <span className="block px-1 pb-1 text-xs text-muted-foreground">
              {about}
            </span>
          </label>
        );
      })}
    </div>
  );
}
