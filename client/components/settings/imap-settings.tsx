"use client";
import { useConfirmSheet } from "@/components/ui/confirm-sheet";

import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@/components/ui/button";
import { DataTable } from "@/components/ui/data-table";
import { ColumnDef } from "@tanstack/react-table";
import { Badge } from "@/components/ui/badge";
import { MoreHorizontal, Pencil, Trash, TestTube, Mail, Server, ShieldCheck } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useTeam } from "@/app/providers/team-provider";
import { toast } from "sonner";
import { useIMAP } from "@/app/providers/imap-provider";
import { IMAPConfig, IMAPConfigSchema } from "@/lib/validations/imap-provider";
import { ApiError } from "@/lib";
import { IMAPProviders } from "./imap-providers";
import { useApi } from "@/hooks/use-api";
import { CollectionCard } from "@/components/ui/collection-card";
import { QueryState, Empty, Metric } from "@/components/marketing/shared";
import { workspaceClassName } from "@/lib/workspace-styles";

export function IMAPSettings({
  isDialogOpen,
  setIsDialogOpen,
}: {
  isDialogOpen: boolean;
  setIsDialogOpen: (open: boolean) => void;
}) {
  const confirm = useConfirmSheet();
  const [editConfig, setEditConfig] = useState<IMAPConfig | null>(null);
  const { configs: imapConfigs, isLoading, refresh, error } = useIMAP();
  const { apiFetch } = useApi();
  const form = useForm<IMAPConfig>({
    resolver: zodResolver(IMAPConfigSchema),
    defaultValues: { id: null, host: "", port: 993, username: "", password: "" },
  });

  const onSubmit = async (data: IMAPConfig) => {
    try {
      // first test the configuration
      await testConfiguration(data);

      const response = await apiFetch(data.id ? "imap/" + data.id : "imap", {
        method: data.id ? "PUT" : "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          ...data,
          port: Number(data.port),
        }),
      });

      if (!response.ok) {
        const apiError = (await response.json()) as ApiError;
        throw new Error(apiError.message);
      }

      refresh();
      toast.success("IMAP configuration saved successfully");
      setEditConfig(null);
      setIsDialogOpen(false);
      form.reset({ id: null, host: "", port: 993, username: "", password: "" });
    } catch (error: any) {
      toast.error("Failed to save IMAP configuration");
    }
  };

  const removeSMTPConfig = async (id: string) => {
    try {
      const response = await apiFetch("imap/" + id, { method: "DELETE" });
      if (!response.ok) throw new Error("Unable to delete connection");
      refresh();
      toast.success("IMAP configuration removed successfully");
    } catch (error) {
      toast.error("Failed to remove IMAP configuration");
    }
  };

  const testConfiguration = async (config: IMAPConfig) => {
    try {
      const response = await apiFetch("imap/test", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          ...config,
          port: Number(config.port),
        }),
      });

      if (!response.ok) {
        const apiError = (await response.json()) as ApiError;
        throw new Error(apiError.message);
      }

      toast.success("IMAP configuration test successful");
    } catch (error: any) {
      toast.error("IMAP configuration test failed");
      throw error;
    }
  };

  useEffect(() => {
    if (editConfig) {
      form.reset(editConfig);
      setIsDialogOpen(true);
    }
  }, [editConfig, form]);

  const edit = (config: IMAPConfig) => { form.reset(config); setEditConfig(config); setIsDialogOpen(true); };
  const closeDialog = (open: boolean) => {
    setIsDialogOpen(open);
    if (!open) { setEditConfig(null); form.reset({ id: null, host: "", port: 993, username: "", password: "" }); }
  };
  return (
    <div className="space-y-6">
      <div className="grid gap-4 sm:grid-cols-2">
        <Metric label="IMAP mailboxes" value={isLoading ? "—" : imapConfigs.length} icon={<Mail size={18}/>}/>
        <Metric label="Configured servers" value={isLoading ? "—" : new Set(imapConfigs.map(c => c.host)).size} icon={<Server size={18}/>}/>
      </div>
      <section className={workspaceClassName("product-panel")}>
        <div className={workspaceClassName("panel-toolbar")}><h2>IMAP mailboxes</h2><span className="text-xs text-muted-foreground">{imapConfigs.length} connections</span></div>
        {isLoading || error ? <QueryState loading={isLoading} error={error} retry={refresh}/> : imapConfigs.length === 0 ? <Empty title="Connect your first mailbox" description="Use your existing email provider to read and reply from your inbox." action={<Button onClick={() => setIsDialogOpen(true)}>Add IMAP connection</Button>}/> : <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {imapConfigs.map(config => <CollectionCard key={config.id} icon={<Mail size={22}/>} badge={"IMAP"} title={config.username} description={`Server: ${config.host}:${config.port}`} action="Edit connection" onAction={() => edit(config)} menu={
            <DropdownMenu><DropdownMenuTrigger asChild><Button variant="ghost" size="icon" className="size-8" aria-label={`Actions for ${config.host}`}><MoreHorizontal size={18}/></Button></DropdownMenuTrigger><DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => edit(config)}><Pencil className="mr-2 size-4"/>Edit connection</DropdownMenuItem>
              <DropdownMenuItem onClick={() => { void testConfiguration(config).catch(() => {}); }}><TestTube className="mr-2 size-4"/>Test connection</DropdownMenuItem>
              <DropdownMenuSeparator/>
              <DropdownMenuItem className="text-destructive" onClick={async () => { if (await confirm({ title: "Delete IMAP connection?", description: "This connection will be removed from your workspace.", confirmLabel: "Delete connection", variant: "destructive" })) void removeSMTPConfig(config.id as string); }}><Trash className="mr-2 size-4"/>Delete connection</DropdownMenuItem>
            </DropdownMenuContent></DropdownMenu>
          }>
            <div className="mt-5 rounded-xl bg-muted p-3"><p className="text-xs text-muted-foreground">Mailbox server</p><p className="mt-1 break-all text-sm font-medium">{config.host}:{config.port}</p></div>
          </CollectionCard>)}
        </div>}
      </section>

      <IMAPProviders
        onTestConnection={testConfiguration}
        onSaveProvider={onSubmit}
        isDialogOpen={isDialogOpen}
        form={form}
        setIsDialogOpen={closeDialog}
      />
    </div>
  );
}
