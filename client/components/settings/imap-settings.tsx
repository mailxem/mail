"use client";
import { useConfirmSheet } from "@/components/ui/confirm-sheet";

import * as React from "react";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@/components/ui/button";
import { DataTable } from "@/components/ui/data-table";
import { ColumnDef } from "@tanstack/react-table";
import { Badge } from "@/components/ui/badge";
import {
  MoreHorizontal,
  Pencil,
  Trash,
  TestTube,
  Mail,
  Server,
  ShieldCheck,
} from "lucide-react";
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
import { CloudflareButton, CloudflareLogo } from "./cloudflare-setup";
import { useApi } from "@/hooks/use-api";
import { CollectionCard } from "@/components/ui/collection-card";
import { QueryState } from "@/components/marketing/shared";
import { useMarketingQuery } from "@/lib/marketing/api";
import { workspaceClassName } from "@/lib/workspace-styles";

export function IMAPSettings({
  isDialogOpen,
  setIsDialogOpen,
  leading,
}: {
  isDialogOpen: boolean;
  setIsDialogOpen: (open: boolean) => void;
  /** Provider tiles shown before Cloudflare and IMAP, e.g. Gmail. */
  leading?: React.ReactNode;
}) {
  const confirm = useConfirmSheet();
  const linkedMailboxes = useMarketingQuery<
    { id: string; smtpConfigId?: string }[]
  >("mail-connections/mailboxes");
  const [editConfig, setEditConfig] = useState<IMAPConfig | null>(null);
  const { configs: imapConfigs, isLoading, refresh, error } = useIMAP();
  const { apiFetch } = useApi();
  const form = useForm<IMAPConfig>({
    resolver: zodResolver(IMAPConfigSchema),
    defaultValues: {
      id: null,
      host: "",
      port: 993,
      username: "",
      password: "",
    },
  });

  const onSubmit = async (data: IMAPConfig) => {
    try {
      // first test the configuration
      await testConfiguration(data, false);

      const response = await apiFetch(data.id ? "imap/" + data.id : "imap", {
        method: data.id ? "PUT" : "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          ...data,
          port: Number(data.port),
        }),
      });

      if (!response.ok) {
        const apiError = (await response.json()) as ApiError & {
          error?: unknown;
        };
        throw new Error(
          typeof apiError.error === "string"
            ? apiError.error
            : apiError.message || "Unable to save IMAP configuration",
        );
      }

      refresh();
      toast.success("IMAP configuration saved successfully");
      setEditConfig(null);
      setIsDialogOpen(false);
      form.reset({ id: null, host: "", port: 993, username: "", password: "" });
    } catch (error: any) {
      toast.error("Failed to save IMAP configuration", {
        description: error.message,
      });
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

  const testConfiguration = async (config: IMAPConfig, notify = true) => {
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
        const apiError = (await response.json()) as ApiError & {
          error?: unknown;
        };
        throw new Error(
          typeof apiError.error === "string"
            ? apiError.error
            : apiError.message || "Check the IMAP connection details",
        );
      }

      if (notify) toast.success("IMAP configuration test successful");
    } catch (error: any) {
      if (notify)
        toast.error("IMAP configuration test failed", {
          description: error.message,
        });
      throw error;
    }
  };

  useEffect(() => {
    if (editConfig) {
      form.reset({ ...editConfig, password: "" });
      setIsDialogOpen(true);
    }
  }, [editConfig, form]);

  const edit = (config: IMAPConfig) => {
    form.reset({ ...config, password: "" });
    setEditConfig(config);
    setIsDialogOpen(true);
  };
  const closeDialog = (open: boolean) => {
    setIsDialogOpen(open);
    if (!open) {
      setEditConfig(null);
      form.reset({ id: null, host: "", port: 993, username: "", password: "" });
    }
  };
  return (
    <div id="other-imap" className="space-y-6">
      <div>
        <h2 className="text-sm font-medium">Connect another provider</h2>
        <div className="mt-3 grid gap-4 md:grid-cols-3">
          {leading}
          <ProviderTile
            icon={<CloudflareLogo className="h-3.5" />}
            iconClassName="bg-[#F38020]/10"
            title="Cloudflare"
            description="Keep incoming mail in your own Cloudflare account."
            action={<CloudflareButton mode="receive">Connect Cloudflare</CloudflareButton>}
          />
          <ProviderTile
            icon={<Server size={16} />}
            title="Other IMAP"
            description="Fastmail, Zoho, Outlook or any IMAP server."
            action={
              <Button variant="outline" onClick={() => setIsDialogOpen(true)}>
                Connect IMAP
              </Button>
            }
            meta={imapConfigs.length ? `${imapConfigs.length} connected` : undefined}
          />
        </div>
      </div>
      {(isLoading || error || imapConfigs.length > 0) && (
      <section className={workspaceClassName("product-panel")}>
        <div className={workspaceClassName("panel-toolbar")}>
          <h2>IMAP mailboxes</h2>
          <span className="text-xs text-muted-foreground">
            {imapConfigs.length} connections
          </span>
        </div>
        {isLoading || error ? (
          <QueryState loading={isLoading} error={error} retry={refresh} />
        ) : (
          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
            {imapConfigs.map((config) =>
              linkedMailboxes.data?.some(
                (mailbox) => mailbox.id === config.id && mailbox.smtpConfigId,
              ) ? (
                <CollectionCard
                  key={config.id}
                  icon={<ShieldCheck size={22} />}
                  badge="Google OAuth"
                  title={config.username}
                  description="Google manages authentication for this mailbox."
                  action="Manage Google connection"
                  onAction={() =>
                    document
                      .getElementById("connected-mail-google")
                      ?.scrollIntoView({ behavior: "smooth" })
                  }
                />
              ) : (
                <CollectionCard
                  key={config.id}
                  icon={<Mail size={22} />}
                  badge={"IMAP"}
                  title={config.username}
                  description={`Server: ${config.host}:${config.port}`}
                  action="Edit connection"
                  onAction={() => edit(config)}
                  menu={
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="size-8"
                          aria-label={`Actions for ${config.host}`}
                        >
                          <MoreHorizontal size={18} />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => edit(config)}>
                          <Pencil className="mr-2 size-4" />
                          Edit connection
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          onClick={() => {
                            void testConfiguration(config).catch(() => {});
                          }}
                        >
                          <TestTube className="mr-2 size-4" />
                          Test connection
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          className="text-destructive"
                          onClick={async () => {
                            if (
                              await confirm({
                                title: "Delete IMAP connection?",
                                description:
                                  "This connection will be removed from your workspace.",
                                confirmLabel: "Delete connection",
                                variant: "destructive",
                              })
                            )
                              void removeSMTPConfig(config.id as string);
                          }}
                        >
                          <Trash className="mr-2 size-4" />
                          Delete connection
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  }
                >
                  <div className="mt-5 rounded-xl bg-muted p-3">
                    <p className="text-xs text-muted-foreground">
                      Mailbox server
                    </p>
                    <p className="mt-1 break-all text-sm font-medium">
                      {config.host}:{config.port}
                    </p>
                  </div>
                </CollectionCard>
              ),
            )}
          </div>
        )}
      </section>
      )}

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

export function ProviderTile({
  icon,
  iconClassName = "bg-muted",
  title,
  description,
  action,
  meta,
  children,
}: {
  icon: React.ReactNode;
  iconClassName?: string;
  title: string;
  description: string;
  action: React.ReactNode;
  meta?: string;
  children?: React.ReactNode;
}) {
  return (
    <div className="flex h-full flex-col rounded-xl border bg-card p-5">
      <div className="flex items-center justify-between gap-2">
        <span className={`grid size-9 place-items-center rounded-lg border ${iconClassName}`}>
          {icon}
        </span>
        {meta && <span className="text-xs text-muted-foreground">{meta}</span>}
      </div>
      <h3 className="mt-4 text-sm font-medium">{title}</h3>
      <p className="mt-1 text-sm text-muted-foreground">{description}</p>
      {children}
      <div className="mt-auto pt-4 [&>button]:w-full">{action}</div>
    </div>
  );
}
