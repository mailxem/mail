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
} from "lucide-react";
import { Button } from "@/components/ui/button";
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
export function NewslettersPage() {
  const q = useMarketingQuery<Newsletter[]>("marketing/newsletters");
  const options = useMarketingQuery<Options>("marketing/options");
  const { request, refresh } = useMarketing();
  const [tab, setTab] = useState("All newsletters");
  const [editing, setEditing] = useState<Partial<Newsletter> | null>(null);
  const [history, setHistory] = useState<Newsletter | null>(null);
  const rows = q.data || [];
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
          <Button className={workspaceClassName("product-primary")} onClick={() => setEditing({})}>
            <Plus />
            Create Newsletter
          </Button>
        }
      />
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
              <Button
                className={workspaceClassName("product-primary")}
                onClick={() => setEditing({})}
              >
                <Plus />
                Create a newsletter
              </Button>
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
                          className={workspaceClassName("newsletter-title text-left")}
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
  });
  const [at, setAt] = useState(localDate(value.nextSendAt));
  const [status, setStatus] = useState(
    value.status === "SCHEDULED" ? "SCHEDULED" : "DRAFT",
  );
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [review, setReview] = useState(false);
  const template = options?.templates.find((t) => t.id === draft.templateId);
  const preview = useMarketingQuery<{htmlBody:string}>(`marketing/templates/${draft.templateId}/preview`, !!draft.templateId && !template?.htmlBody);
  const templateHTML = template?.htmlBody || preview.data?.htmlBody;
  const set = (key: string, val: string) => {
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
          <Button className={workspaceClassName("product-primary")} disabled={busy}>
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
            </tr>
          ))}
        </tbody>
      </table>
    </Modal>
  );
}
