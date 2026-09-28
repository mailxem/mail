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
import { useSMTP } from "@/app/providers/smtp-provider";
import { SMTPConfig, formSchema } from "@/lib/validations/smtp-provider";
import { SMTPProviders } from "./smtp-providers";
import { ApiError, SMTPProviderType } from "@/lib";
import Link from "next/link";
import { toast } from "sonner";
import { useApi } from "@/hooks/use-api";
import { CollectionCard } from "@/components/ui/collection-card";
import { QueryState, Empty, Metric } from "@/components/marketing/shared";
import { workspaceClassName } from "@/lib/workspace-styles";

const DEFAULT_PROVIDERS: Record<SMTPProviderType, SMTPConfig> = {
  [SMTPProviderType.CUSTOM]: {
    provider: SMTPProviderType.CUSTOM,
    host: "",
    port: 587,
    username: "",
    password: "",
    requiresAuth: true,
    supportsTls: true,
    maxSendRate: 14,
    documentation: "",
  },
  [SMTPProviderType.GMAIL]: {
    provider: SMTPProviderType.GMAIL,
    host: "smtp.gmail.com",
    port: 587,
    username: "",
    password: "",
    requiresAuth: true,
    supportsTls: true,
    maxSendRate: 14,
    documentation: "https://support.google.com/mail/answer/7126229?hl=en",
  },
  [SMTPProviderType.OUTLOOK]: {
    provider: SMTPProviderType.OUTLOOK,
    host: "smtp.office365.com",
    port: 587,
    username: "",
    password: "",
    requiresAuth: true,
    supportsTls: true,
    maxSendRate: 14,
    documentation:
      "https://support.microsoft.com/en-us/office/set-up-a-connection-to-an-smtp-server-for-outlook-com-d088d509-d54c-4036-a5c3-21d483f2f017",
  },
  [SMTPProviderType.AMAZON]: {
    provider: SMTPProviderType.AMAZON,
    host: "email-smtp.us-east-1.amazonaws.com",
    port: 587,
    username: "",
    password: "",
    requiresAuth: true,
    supportsTls: true,
    maxSendRate: 14,
    documentation:
      "https://docs.aws.amazon.com/ses/latest/DeveloperGuide/smtp-connect.html",
  },
};

export function SMTPSettings({
  isDialogOpen,
  setIsDialogOpen,
}: {
  isDialogOpen: boolean;
  setIsDialogOpen: (open: boolean) => void;
}) {
  const confirm = useConfirmSheet();
  const [editConfig, setEditConfig] = useState<SMTPConfig | null>(null);
  const { team } = useTeam();
  const { configs: smtpConfigs, isLoading, refresh, error } = useSMTP();
  const { apiFetch } = useApi();

  const form = useForm<SMTPConfig>({
    resolver: zodResolver(formSchema),
    defaultValues: { ...DEFAULT_PROVIDERS[SMTPProviderType.CUSTOM], id: null, isActive: true, fromEmail: "" },
  });

  const onSubmit = async (data: SMTPConfig) => {
    try {
      // first test the configuration
      await testConfiguration(data);

      const response = await apiFetch(
        data.id ? "smtp-configs/" + data.id : "smtp-configs",
        {
          method: data.id ? "PUT" : "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            ...data,
            port: Number(data.port),
            requireTls: Number(data.port) === 587,
            maxSendRate: Number(data.maxSendRate),
          }),
        }
      );

      if (!response.ok) {
        const apiError = (await response.json()) as ApiError;
        throw new Error(apiError.message);
      }

      refresh();
      toast.success("SMTP configuration saved successfully");
      setEditConfig(null);
      setIsDialogOpen(false);
      form.reset({ ...DEFAULT_PROVIDERS[SMTPProviderType.CUSTOM], id: null, isActive: true, fromEmail: "" });
    } catch (error: any) {
      toast.error("Failed to save SMTP configuration");
    }
  };

  const removeSMTPConfig = async (id: string) => {
    try {
      const response = await apiFetch("smtp-configs/" + id, { method: "DELETE" });
      if (!response.ok) throw new Error("Unable to delete connection");
      refresh();
      toast.success("SMTP configuration removed successfully");
    } catch (error) {
      toast.error("Failed to remove SMTP configuration");
    }
  };

  const testConfiguration = async (config: SMTPConfig) => {
    try {
      const response = await apiFetch("smtp/test", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          ...config,
          from: config.fromEmail,
          port: Number(config.port),
          requireTls: Number(config.port) === 587,
          maxSendRate: Number(config.maxSendRate),
        }),
      });

      if (!response.ok) {
        const apiError = (await response.json()) as ApiError;
        throw new Error(apiError.message);
      }

      toast.success("SMTP configuration test successful");
    } catch (error: any) {
      toast.error("SMTP configuration test failed");
      throw error;
    }
  };

  const onProviderChange = (value: SMTPProviderType) => {
    form.setValue("provider", value);
    if (value !== SMTPProviderType.CUSTOM) {
      const provider = DEFAULT_PROVIDERS[value];
      form.setValue("host", provider.host);
      form.setValue("port", provider.port);
    }
  };

  useEffect(() => {
    if (editConfig) {
      form.reset(editConfig);
      setIsDialogOpen(true);
    }
  }, [editConfig, form]);

  const edit = (config: SMTPConfig) => { form.reset(config); setEditConfig(config); setIsDialogOpen(true); };
  const closeDialog = (open: boolean) => {
    setIsDialogOpen(open);
    if (!open) { setEditConfig(null); form.reset({ ...DEFAULT_PROVIDERS[SMTPProviderType.CUSTOM], id: null, isActive: true, fromEmail: "" }); }
  };
  return (
    <div className="space-y-6">
      <div className="grid gap-4 sm:grid-cols-2">
        <Metric label="SMTP senders" value={isLoading ? "—" : smtpConfigs.length} icon={<Mail size={18}/>}/>
        <Metric label="Configured servers" value={isLoading ? "—" : new Set(smtpConfigs.map(c => c.host)).size} icon={<Server size={18}/>}/>
      </div>
      <section className={workspaceClassName("product-panel")}>
        <div className={workspaceClassName("panel-toolbar")}><h2>SMTP senders</h2><span className="text-xs text-muted-foreground">{smtpConfigs.length} connections</span></div>
        {isLoading || error ? <QueryState loading={isLoading} error={error} retry={refresh}/> : smtpConfigs.length === 0 ? <Empty title="Connect your first sender" description="Use your existing email provider to send emails with Xem." action={<Button onClick={() => setIsDialogOpen(true)}>Add SMTP connection</Button>}/> : <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {smtpConfigs.map(config => String(config.provider) === "MANAGED" ? <CollectionCard key={config.id} icon={<ShieldCheck size={22}/>} badge="Xem managed" title={config.fromEmail} description="Delivery through your verified domain" action="Manage sending" onAction={() => { window.location.href = "/settings/sending"; }} /> : <CollectionCard key={config.id} icon={<Mail size={22}/>} badge={config.provider} title={config.fromEmail || config.host} description={`Server: ${config.host}:${config.port}`} action="Edit connection" onAction={() => edit(config)} menu={
            <DropdownMenu><DropdownMenuTrigger asChild><Button variant="ghost" size="icon" className="size-8" aria-label={`Actions for ${config.host}`}><MoreHorizontal size={18}/></Button></DropdownMenuTrigger><DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => edit(config)}><Pencil className="mr-2 size-4"/>Edit connection</DropdownMenuItem>
              <DropdownMenuItem onClick={() => { void testConfiguration(config).catch(() => {}); }}><TestTube className="mr-2 size-4"/>Test connection</DropdownMenuItem>
              <DropdownMenuSeparator/>
              <DropdownMenuItem className="text-destructive" onClick={async () => { if (await confirm({ title: "Delete SMTP connection?", description: "This connection will be removed from your workspace.", confirmLabel: "Delete connection", variant: "destructive" })) void removeSMTPConfig(config.id as string); }}><Trash className="mr-2 size-4"/>Delete connection</DropdownMenuItem>
            </DropdownMenuContent></DropdownMenu>
          }>
            <div className="mt-5 grid grid-cols-2 gap-3"><div className="rounded-xl bg-muted p-3"><p className="text-xs text-muted-foreground">Send rate</p><p className="mt-1 text-sm font-medium">{config.maxSendRate}/sec</p></div><div className="rounded-xl bg-muted p-3"><p className="text-xs text-muted-foreground">Status</p><p className="mt-1 text-sm font-medium">{config.isActive ? "Active" : "Inactive"}</p></div></div>
          </CollectionCard>)}
        </div>}
      </section>

      <SMTPProviders
        onTestConnection={testConfiguration}
        onSaveProvider={onSubmit}
        isDialogOpen={isDialogOpen}
        form={form}
        onProviderChange={onProviderChange}
        providers={DEFAULT_PROVIDERS}
        setIsDialogOpen={closeDialog}
      />
    </div>
  );
}
