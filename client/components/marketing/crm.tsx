"use client";
import { workspaceClassName } from "@/lib/workspace-styles";
import Link from "next/link";
import { useEffect, useState } from "react";
import {
  Search,
  Users,
  UserCheck,
  Building2,
  TrendingUp,
  ArrowUpRight,
  MessageSquare,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { CollectionPagination } from "@/components/ui/collection-pagination";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useMarketing, useMarketingQuery } from "@/lib/marketing/api";
import type { Contact, CRMContactsPage } from "@/lib/marketing/types";
import {
  PageHeading,
  Metric,
  QueryState,
  Empty,
  Modal,
  Field,
  FilterTabs,
  Status,
} from "./shared";
import { toast } from "sonner";
const stages = ["LEAD", "QUALIFIED", "CUSTOMER", "LOST"];
export function CRMPage() {
  const [filter, setFilter] = useState("Everyone");
  const [search, setSearch] = useState("");
  const [querySearch, setQuerySearch] = useState("");
  const [editing, setEditing] = useState<Partial<Contact> | null>(null);
  const [view, setView] = useState("List");
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  useEffect(() => { const value = new URLSearchParams(window.location.search).get("search"); if (value) setSearch(value); }, []);
  const searchPending = search.trim() !== querySearch;
  useEffect(() => {
    const timer = setTimeout(() => {
      setQuerySearch(search.trim());
      setPage(1);
    }, 300);
    return () => clearTimeout(timer);
  }, [search]);
  const params = new URLSearchParams({
    page: String(page),
    limit: String(pageSize),
  });
  if (querySearch) params.set("search", querySearch);
  if (filter !== "Everyone") params.set("stage", filter.toUpperCase());
  const q = useMarketingQuery<CRMContactsPage>(
    `marketing/contacts?${params}`,
    !searchPending,
  );
  const contacts = q.data?.data ?? [];
  const total = q.data?.total ?? 0;
  const currentPage = q.data?.page ?? page;
  const summary = q.error ? undefined : q.data?.summary;
  const loading = q.isPending || searchPending;
  useEffect(() => {
    if (q.data && !searchPending) setPage(q.data.page);
  }, [q.data, searchPending]);
  return (
    <>
      <PageHeading
        title="People, not just addresses."
        description="Meet your audience. Build relationships that go beyond the inbox."
        action={
          <Button className={workspaceClassName("product-primary")} asChild>
            <Link href="/audience/lists">
              <Users />
              Manage contact lists
            </Link>
          </Button>
        }
      />
      <div className={workspaceClassName("metrics-grid")}>
        <Metric
          label="Total Contacts"
          value={summary?.total.toLocaleString() ?? "—"}
          icon={<Users />}
        />
        <Metric
          label="Qualified Leads"
          value={summary?.qualified.toLocaleString() ?? "—"}
          icon={<TrendingUp />}
        />
        <Metric
          label="Customers"
          value={summary?.customers.toLocaleString() ?? "—"}
          icon={<Building2 />}
        />
        <Metric
          label="Subscribed"
          value={summary?.subscribed.toLocaleString() ?? "—"}
          icon={<UserCheck />}
        />
      </div>
      <section className={workspaceClassName("product-panel")}>
        <div className={workspaceClassName("panel-toolbar")}>
          <div>
            <h2>Audience & CRM</h2>
            <p>
              {!loading && !q.error
                ? `${total.toLocaleString()} ${filter !== "Everyone" || querySearch ? "matching contacts" : "contacts in your workspace"}`
                : "All your workspace contacts, in one place"}
            </p>
          </div>
          <FilterTabs
            items={["List", "Pipeline"]}
            value={view}
            onChange={setView}
          />
        </div>
        <div className={workspaceClassName("panel-toolbar")}>
          <div className="relative">
            <Search
              size={15}
              className="absolute left-3 top-3.5 text-muted-foreground"
            />
            <input
              aria-label="Search contacts"
              className={workspaceClassName("product-input !pl-9 !w-64")}
              placeholder="Search people or companies…"
              value={search}
              onChange={(e) => {
                setSearch(e.target.value);
              }}
              maxLength={200}
            />
          </div>
          <FilterTabs
            items={["Everyone", "Lead", "Qualified", "Customer", "Lost"]}
            value={filter}
            onChange={(value) => {
              setFilter(value);
              setPage(1);
            }}
          />
        </div>
        <QueryState
          loading={loading}
          error={q.error}
          retry={() => q.refetch()}
        />
        {!loading && !q.error && contacts.length === 0 && (
          <Empty
            title={
              summary?.total
                ? "No matching contacts"
                : "Your people are out there"
            }
            description={
              summary?.total
                ? "Try a different search or lifecycle stage."
                : "Add a contact or publish a lead form to start growing your audience."
            }
          />
        )}{" "}
        {!loading &&
          !q.error &&
          contacts.length > 0 &&
          (view === "List" ? (
            <div>
              <div className={workspaceClassName("table-wrap")}>
                <table className={workspaceClassName("product-table")}>
                  <thead>
                    <tr>
                      <th>Contact</th>
                      <th>Company</th>
                      <th>Stage</th>
                      <th>Subscription</th>
                      <th>Audience</th>
                      <th />
                    </tr>
                  </thead>
                  <tbody>
                    {contacts.map((c) => (
                      <tr key={c.id}>
                        <td>
                          <button
                            className="flex items-center gap-3 text-left"
                            onClick={() => setEditing(c)}
                          >
                            <span
                              className={workspaceClassName(
                                "user-avatar !bg-muted !text-foreground",
                              )}
                            >
                              {(c.firstName || c.email)
                                .slice(0, 1)
                                .toUpperCase()}
                              {c.lastName?.slice(0, 1)}
                            </span>
                            <div>
                              <strong>
                                {[c.firstName, c.lastName]
                                  .filter(Boolean)
                                  .join(" ") || c.email}
                              </strong>
                              <small>{c.email}</small>
                            </div>
                          </button>
                        </td>
                        <td>{c.company || "—"}</td>
                        <td>
                          <span
                            className={workspaceClassName(
                              `crm-stage stage-${c.lifecycleStage?.toLowerCase()}`,
                            )}
                          >
                            {c.lifecycleStage || "LEAD"}
                          </span>
                        </td>
                        <td>
                          <Status value={c.status} />
                        </td>
                        <td>{c.listName || "—"}</td>
                        <td>
                          <button
                            className={workspaceClassName("icon-button")}
                            aria-label={`Open ${c.email}`}
                            onClick={() => setEditing(c)}
                          >
                            <ArrowUpRight />
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          ) : (
            <div>
              <p className="mb-4 text-xs text-muted-foreground">
                Pipeline shows the current page. Column counts are for this
                page.
              </p>
              <div className={workspaceClassName("crm-pipeline")}>
                {stages.map((stage) => (
                  <div
                    className={workspaceClassName("pipeline-column")}
                    key={stage}
                  >
                    <h3>
                      <span
                        className={workspaceClassName(
                          `stage-dot stage-${stage.toLowerCase()}`,
                        )}
                      />
                      {stage.charAt(0) + stage.slice(1).toLowerCase()}
                      <small>
                        {
                          contacts.filter(
                            (c) => (c.lifecycleStage || "LEAD") === stage,
                          ).length
                        }
                      </small>
                    </h3>
                    {contacts
                      .filter((c) => (c.lifecycleStage || "LEAD") === stage)
                      .map((c) => (
                        <button
                          className={workspaceClassName("pipeline-contact")}
                          key={c.id}
                          onClick={() => setEditing(c)}
                        >
                          <strong>
                            {[c.firstName, c.lastName]
                              .filter(Boolean)
                              .join(" ") || c.email}
                          </strong>
                          <span>{c.company || c.email}</span>
                          <small>
                            <MessageSquare size={12} />
                            Open contact
                          </small>
                        </button>
                      ))}
                  </div>
                ))}
              </div>
            </div>
          ))}
        {!loading && !q.error && (
          <div className="mt-5 flex flex-wrap items-center gap-x-5 gap-y-3 border-t border-border pt-4">
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <label htmlFor="crm-page-size">Rows per page</label>
              <Select
                value={String(pageSize)}
                onValueChange={(value) => {
                  setPageSize(Number(value));
                  setPage(1);
                }}
              >
                <SelectTrigger id="crm-page-size" className="h-8 w-20">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent side="top">
                  {[10, 20, 50].map((size) => (
                    <SelectItem key={size} value={String(size)}>
                      {size}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="min-w-0 flex-1 [&>div]:mt-0 [&>div]:border-t-0 [&>div]:pt-0">
              <CollectionPagination
                page={currentPage}
                limit={q.data?.limit ?? pageSize}
                total={total}
                onPageChange={setPage}
              />
            </div>
          </div>
        )}
      </section>
      {editing && (
        <ContactEditor contact={editing} close={() => setEditing(null)} />
      )}
    </>
  );
}
function ContactEditor({
  contact,
  close,
}: {
  contact: Partial<Contact>;
  close: () => void;
}) {
  const { request, refresh } = useMarketing();
  const [stage, setStage] = useState(contact.lifecycleStage || "LEAD");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  async function save(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    try {
      await request(`marketing/contacts/${contact.id}/stage`, "PUT", {
        lifecycleStage: stage,
      });
      await refresh();
      toast.success("Lifecycle stage updated");
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
      title="Contact profile"
      description={
        contact.id
          ? "Keep the details and conversations that matter in one place."
          : "Start with the basics. Make sure you have permission to email this person."
      }
    >
      <form onSubmit={save} className={workspaceClassName("product-form")}>
        <div className="rounded-xl bg-muted/40 p-4 space-y-2">
          <strong>
            {[contact.firstName, contact.lastName].filter(Boolean).join(" ") ||
              contact.email}
          </strong>
          <p>{contact.email}</p>
          <p>{contact.company}</p>
          <p>{contact.phone}</p>
          <Link
            className="text-muted-foreground"
            href={`/audience/lists/${contact.listId}`}
          >
            Open in contact list ↗
          </Link>
        </div>
        <Field label="Lifecycle stage">
          <select value={stage} onChange={(e) => setStage(e.target.value)}>
            {stages.map((stage) => (
              <option key={stage}>{stage}</option>
            ))}
          </select>
        </Field>
        {error && (
          <div role="alert" className={workspaceClassName("product-error")}>
            {error}
          </div>
        )}
        <div className={workspaceClassName("modal-actions")}>
          <Button type="button" variant="outline" onClick={close}>
            Cancel
          </Button>
          <Button
            disabled={busy}
            className={workspaceClassName("product-primary")}
          >
            {busy ? "Saving…" : "Save stage"}
          </Button>
        </div>
      </form>
      {contact.id && <ContactNotes id={contact.id} />}
    </Modal>
  );
}
function ContactNotes({ id }: { id: string }) {
  const q = useMarketingQuery<any[]>(`marketing/contacts/${id}/notes`);
  const { request, refresh } = useMarketing();
  const [note, setNote] = useState("");
  const [busy, setBusy] = useState(false);
  const add = async () => {
    setBusy(true);
    try {
      await request(`marketing/contacts/${id}/notes`, "POST", { body: note });
      setNote("");
      await refresh();
    } catch (e) {
      toast.error((e as Error).message);
    } finally {
      setBusy(false);
    }
  };
  return (
    <div className="border-t pt-5 mt-2">
      <h3 className="font-medium mb-3">Notes & conversations</h3>
      <Field label="Add a note">
        <textarea
          rows={2}
          value={note}
          maxLength={5000}
          onChange={(e) => setNote(e.target.value)}
          placeholder="What should your team know?"
        />
      </Field>
      <Button
        variant="outline"
        size="sm"
        className="mt-2"
        disabled={!note.trim() || busy}
        onClick={add}
      >
        {busy ? "Adding…" : "Add note"}
      </Button>
      <QueryState loading={q.isLoading} error={q.error} />
      {q.data?.map((n) => (
        <div className={workspaceClassName("contact-note")} key={n.id}>
          <p>{n.body}</p>
          <small>{new Date(n.createdAt).toLocaleString()}</small>
        </div>
      ))}
    </div>
  );
}
