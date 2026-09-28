"use client";
import { useConfirmSheet } from "@/components/ui/confirm-sheet";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { PaginationState } from "@tanstack/react-table";
import { Button } from "@/components/ui/button";
import {
  MoreHorizontal,
  Pencil,
  Trash,
  Copy,
  ChevronRight,
  ChevronLeft,
  PanelsTopLeft,
  ArrowUpRight,
} from "lucide-react";
import { useTeam } from "@/app/providers/team-provider";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { EmailTemplate } from "@/lib";
import { Input } from "../ui/input";
import { toast } from "sonner";
import { useApi } from "@/hooks/use-api";
import { Empty, QueryState } from "@/components/marketing/shared";
import Link from "next/link";
import { CollectionCard } from "@/components/ui/collection-card";
import { CollectionPagination } from "@/components/ui/collection-pagination";
import { useTemplates } from "@/app/providers/templates-provider";

interface Template {
  id: string;
  name: string;
  description: string;
  category: string;
  status: "draft" | "published";
  lastModified: string;
  teamId: string;
}

export function TemplatesList() {
  const confirm = useConfirmSheet();

  const router = useRouter();
  const { team } = useTeam();
  const { apiFetch } = useApi();
  const { templates, isLoading, error, refetch, pagination, setPagination } =
    useTemplates();
  const totalCount = pagination.total;

  const deleteTemplate = async (id: string) => {
    try {
      const response = await apiFetch(`templates/${id}`, {
        method: "DELETE",
      });
      if (!response.ok) throw new Error("Failed to delete template");

      toast.success("Template deleted successfully");

      await refetch();
    } catch (error) {
      toast.error("Failed to delete template");
    }
  };

  const duplicateTemplate = async (data: EmailTemplate) => {
    try {
      const createResponse = await apiFetch("templates", {
        method: "POST",
        body: JSON.stringify({
          ...data,
          id: undefined,
          name: `${data.name} - Copy`,
          duplicate: true,
        }),
      });

      if (!createResponse.ok) throw new Error("Failed to duplicate template");

      toast.success("Template duplicated successfully");

      await refetch();
    } catch (error) {
      toast.error("Failed to duplicate template");
    }
  };



  return <>
    {isLoading || error ? <QueryState loading={isLoading} error={error} retry={refetch}/> : templates.length === 0 ? <Empty title="Your next email starts here" description="Create a design in your template editor, or choose a newsletter starter." action={<Button asChild><Link href="/templates/new">Create template</Link></Button>}/> : <>
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
        {templates.map(template => <CollectionCard key={template.id} icon={<PanelsTopLeft size={22}/>} badge="Email template" title={template.name} href={`/templates/${template.id}/edit`} description={`Last updated ${new Date(template.updatedAt).toLocaleDateString()}`} action="Open editor" menu={
          <DropdownMenu><DropdownMenuTrigger asChild><Button variant="ghost" size="icon" className="size-8" aria-label={`Actions for ${template.name}`}><MoreHorizontal size={18}/></Button></DropdownMenuTrigger><DropdownMenuContent align="end">
            <DropdownMenuItem asChild><Link href={`/templates/${template.id}/edit`}><Pencil className="mr-2 size-4"/>Edit template</Link></DropdownMenuItem>
            <DropdownMenuItem onClick={() => duplicateTemplate(template)}><Copy className="mr-2 size-4"/>Duplicate</DropdownMenuItem>
            <DropdownMenuItem onClick={async () => { try { await navigator.clipboard.writeText(template.id); toast.success("Template ID copied"); } catch { toast.error("Unable to copy template ID"); } }}>Copy template ID</DropdownMenuItem>
            <DropdownMenuSeparator/>
            <DropdownMenuItem className="text-destructive" onClick={async () => { if (await confirm({ title: "Delete template?", description: "This cannot be undone.", confirmLabel: "Delete template", variant: "destructive" })) void deleteTemplate(template.id); }}><Trash className="mr-2 size-4"/>Delete</DropdownMenuItem>
          </DropdownMenuContent></DropdownMenu>
        }/>)}
      </div>
      <CollectionPagination {...pagination} onPageChange={page => setPagination({ ...pagination, page })}/>
    </>}
  </>;
}
