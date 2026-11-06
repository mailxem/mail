"use client";
import { resourceEntity } from "@/lib/resource-response";

import { useState } from "react";
import { ColumnDef, PaginationState } from "@tanstack/react-table";
import { DataTable } from "@/components/ui/data-table";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";
import { Button } from "@/components/ui/button";
import { MoreHorizontal, Mail, UserX, UserCheck } from "lucide-react";
import { useTeam } from "@/app/providers/team-provider";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { toast } from "sonner";
import { useApi } from "@/hooks/use-api";
import { useQuery } from "@tanstack/react-query";
import { MailingList } from "@/lib";
import { useResourcePage } from "@/hooks/use-resource-page";
import { QueryState, Metric } from "@/components/marketing/shared";
import { workspaceClassName } from "@/lib/workspace-styles";
import { CollectionCard } from "@/components/ui/collection-card";
import { CollectionPagination } from "@/components/ui/collection-pagination";
import { Status } from "@/components/marketing/shared";
import { ContactTagEditor } from "@/components/contacts/contact-tag-editor";
import { ContactImport } from "@/components/contacts/contact-import";
import { PageHeader } from "@/components/page-header";
import { Users, ListFilter, CalendarDays, MoonStar } from "lucide-react";
import { Switch } from "@/components/ui/switch";

interface Contact {
  id: string;
  email: string;
  firstName: string | null;
  lastName: string | null;
  isDeleted: boolean;
  status: string;
  metadata: Record<string, any>;
  tags: { id: string; name: string }[] | null;
  createdAt: string;
  updatedAt: string;
  listId: string;
  teamId: string;
  importId: string | null;
}

export function ContactsList({ listId }: { listId: string }) {
  const [tagContact, setTagContact] = useState<string | null>(null);
  const [pagination, setPagination] = useState<PaginationState>({ pageIndex: 0, pageSize: 20 });
  const [inactiveOnly, setInactiveOnly] = useState(false);
  const { team } = useTeam();
  const { apiFetch, session } = useApi();
  const contactsQuery = useResourcePage<Contact>("contacts", pagination.pageIndex + 1, pagination.pageSize, { list_id: listId, include: "Tags", ...(inactiveOnly ? { status: "INACTIVE" } : {}) });
  const inactiveQuery = useResourcePage<Contact>("contacts", 1, 1, { list_id: listId, status: "INACTIVE" });
  const listQuery = useQuery<MailingList>({
    queryKey: ["list", team?.id, listId],
    enabled: !!listId && !!team?.id && !!session?.accessToken,
    queryFn: async ({ signal }) => {
      const response = await apiFetch(`mailing-lists/${listId}`, { signal });
      if (!response.ok) throw new Error("Unable to load this contact list.");
      const result = await response.json();
      return resourceEntity<MailingList>(result);
    },
  });
  const contacts = contactsQuery.data?.data ?? [];
  const list = listQuery.data;
  const total = contactsQuery.data?.total ?? 0;
  const inactive = inactiveQuery.data?.total ?? 0;
  const refresh = async () => { await Promise.all([contactsQuery.refetch(), listQuery.refetch(), inactiveQuery.refetch()]); };
  const showInactive = (on: boolean) => { setInactiveOnly(on); setPagination({ ...pagination, pageIndex: 0 }); };

  const setListCleanup = async (listCleanup: boolean) => {
    const response = await apiFetch(`mailing-lists/${listId}`, { method: "PUT", body: JSON.stringify({ name: list?.name, listCleanup }) });
    if (!response.ok) return toast.error("Couldn't change list cleanup");
    toast.success(listCleanup ? "List cleanup is on" : "List cleanup is off");
    await listQuery.refetch();
  };

  const updateContactStatus = async (
    contactId: string,
    newStatus: "ACTIVE" | "UNSUBSCRIBED"
  ) => {
    try {
      const response = await apiFetch("contacts/" + contactId, {
        method: "PUT",
        body: JSON.stringify({
          email: contacts.find(contact => contact.id === contactId)?.email,
          listId,
          status: newStatus,
        }),
      });

      if (!response.ok) throw new Error("Failed to update contact status");

      toast.success(
        `Contact ${
          newStatus === "ACTIVE" ? "resubscribed" : "unsubscribed"
        } successfully`
      );

      await refresh();
    } catch (error) {
      toast.error("Failed to update contact status");
    }
  };

  return (
    <>
      {tagContact && <ContactTagEditor contactId={tagContact} close={() => setTagContact(null)} onSaved={() => void refresh()}/>}
      <PageHeader heading={list?.name || "Contact list"} description={list?.description || "Manage the contacts in this audience."} backButton={{ href: "/audience/lists", label: "Back to contact lists" }}>
        <ContactImport listId={listId} onImportComplete={refresh}/>
      </PageHeader>
      <div className="mb-6 grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <Metric label="Total contacts" value={contactsQuery.isLoading ? "—" : total} icon={<Users size={18}/>}/>
        <Metric label="Subscribers" value={listQuery.isPending ? "—" : list?.subscribersCount ?? 0} icon={<UserCheck size={18}/>}/>
        <Metric label="Inactive" value={inactiveQuery.isLoading ? "—" : inactive} icon={<MoonStar size={18}/>} detail="Not emailed, so not billed"/>
        <Metric label="Created" value={list?.createdAt ? new Date(list.createdAt).toLocaleDateString(undefined, { month: "short", day: "numeric", year: "numeric" }) : "—"} icon={<CalendarDays size={18}/>}/>
      </div>
      <section className={workspaceClassName("product-panel")}>
        <div className={workspaceClassName("panel-toolbar")}><h2>{inactiveOnly ? "Inactive contacts" : "All contacts"}</h2><div className="flex flex-wrap items-center gap-3">
          <label className="flex items-center gap-2 text-xs text-muted-foreground" title="Readers who ignore 5 emails over 60 days get one “Still want these?” email. If they don't answer within 14 days, they stop getting mail until they click or you resume them."><Switch checked={list?.listCleanup !== false} disabled={!list} onCheckedChange={value => void setListCleanup(value)} aria-label="List cleanup"/>List cleanup</label>
          <Button variant={inactiveOnly ? "secondary" : "ghost"} size="sm" onClick={() => showInactive(!inactiveOnly)}><ListFilter size={14}/>{inactiveOnly ? "Show all" : `Inactive (${inactive.toLocaleString()})`}</Button>
          <span className="text-xs text-muted-foreground">{total.toLocaleString()} contacts</span>
        </div></div>
        <div className="px-5 pb-5">
          {contactsQuery.error || listQuery.error ? <QueryState loading={false} error={contactsQuery.error || listQuery.error} retry={refresh}/> : contactsQuery.isLoading ? <QueryState loading/> : <>
            <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">{contacts.map(contact => <CollectionCard key={contact.id} icon={<Users size={22}/>} title={[contact.firstName, contact.lastName].filter(Boolean).join(" ") || contact.email} description={contact.firstName || contact.lastName ? contact.email : undefined} badge={<Status value={contact.status}/>} action="Manage tags" onAction={() => setTagContact(contact.id)} menu={
              <DropdownMenu><DropdownMenuTrigger asChild><Button variant="ghost" size="icon" className="size-8" aria-label={`Actions for ${contact.email}`}><MoreHorizontal size={18}/></Button></DropdownMenuTrigger><DropdownMenuContent align="end"><DropdownMenuItem onClick={() => setTagContact(contact.id)}>Manage tags</DropdownMenuItem><DropdownMenuItem onClick={() => updateContactStatus(contact.id, contact.status === "ACTIVE" ? "UNSUBSCRIBED" : "ACTIVE")}>{contact.status === "ACTIVE" ? "Unsubscribe" : contact.status === "INACTIVE" ? "Resume sending" : "Resubscribe"}</DropdownMenuItem></DropdownMenuContent></DropdownMenu>
            }><div className="mt-4 flex min-h-7 flex-wrap gap-1">{contact.tags?.map(tag => <span key={tag.id} className="rounded-full bg-violet-50 px-2 py-1 text-xs text-violet-600">{tag.name}</span>)}</div><p className="mt-2 text-xs text-muted-foreground">Joined {new Date(contact.createdAt).toLocaleDateString()}</p></CollectionCard>)}</div>
            {!contacts.length && <p className="py-10 text-center text-sm text-muted-foreground">{inactiveOnly ? "No inactive contacts. Everyone here still gets your emails." : "No contacts in this list yet. Import contacts to get started."}</p>}
            <CollectionPagination page={pagination.pageIndex + 1} limit={pagination.pageSize} total={total} onPageChange={page => setPagination({ ...pagination, pageIndex: page - 1 })}/>
          </>}
        </div>
      </section>
    </>
  );
}
