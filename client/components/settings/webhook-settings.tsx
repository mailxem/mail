"use client";
import { useConfirmSheet } from "@/components/ui/confirm-sheet";
import { useState, useEffect } from "react";
import { useTeam } from "@/app/providers/team-provider";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { WebhookEventType } from "@/lib";
import { Webhook } from "@/lib";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { format } from "date-fns";
import { ExternalLink, Loader2, Plus, RefreshCw, Trash, Webhook as WebhookIcon } from "lucide-react";
import { toast } from "sonner";
import Link from "next/link";
import { useApi } from "@/hooks/use-api";
import { useResourcePage } from "@/hooks/use-resource-page";
import { CollectionCard } from "@/components/ui/collection-card";
import { CollectionPagination } from "@/components/ui/collection-pagination";
import { PageHeading, QueryState, Empty } from "@/components/marketing/shared";
import { workspaceClassName } from "@/lib/workspace-styles";

interface WebhookWithEvents extends Webhook {
  events: WebhookEventType[];
  isActive: boolean;
  secret: string;
}

export function WebhookSettings() {
  const confirm = useConfirmSheet();
  const { team } = useTeam();
  const { apiFetch } = useApi();
  const [page, setPage] = useState(1);
  const query = useResourcePage<WebhookWithEvents>("webhooks", page, 20);
  const webhooks = query.data?.data ?? [];
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const fetchWebhooks = () => query.refetch();
  const toggle = async (webhook: WebhookWithEvents) => {
    try {
      const response = await apiFetch(`marketing/webhooks/${webhook.id}/status`, {method:"PUT",body:JSON.stringify({isActive:!webhook.isActive})});
      if (!response.ok) throw new Error("Unable to update webhook");
      await query.refetch();
    } catch { toast.error("Unable to update webhook"); }
  };

  // Delete webhook
  const deleteWebhook = async (id: string) => {
    try {
      const response = await apiFetch(`webhooks/${id}`, {
        method: "DELETE",
      });
      if (!response.ok) throw new Error("Failed to delete webhook");
      
      await fetchWebhooks();
      toast.success("Webhook deleted successfully");
    } catch (error) {
      toast.error("Failed to delete webhook");
    }
  };

  return (
    <div className="space-y-6">
      <PageHeading title="Webhooks" description="Send email events to your connected applications." action={<Button onClick={() => setIsAddDialogOpen(true)}><Plus size={16}/>Add webhook</Button>}/>
      <section className={workspaceClassName("product-panel")}>
        <div className={workspaceClassName("panel-toolbar")}><h2>Your webhooks</h2><span className="text-xs text-muted-foreground">{query.data?.total ?? 0} endpoints</span></div>
        {query.isLoading || query.error ? <QueryState loading={query.isLoading} error={query.error} retry={() => void fetchWebhooks()}/> : !webhooks.length ? <Empty title="Connect your first endpoint" description="Receive updates when contacts open, click, reply, bounce, or report an email." action={<Button onClick={() => setIsAddDialogOpen(true)}>Add webhook</Button>}/> : <>
          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">{webhooks.map(webhook => <CollectionCard key={webhook.id} icon={<WebhookIcon size={22}/>} title={webhook.name} badge={webhook.isActive ? "Active" : "Inactive"} description={webhook.url} href={`/settings/webhooks/${webhook.id}/deliveries`} action="View deliveries" menu={<Button variant="ghost" size="icon" className="size-8 text-destructive" aria-label={`Delete ${webhook.name}`} onClick={async () => { if (await confirm({ title: "Delete webhook?", description: "This endpoint will no longer receive events.", confirmLabel: "Delete webhook", variant: "destructive" })) void deleteWebhook(webhook.id); }}><Trash size={16}/></Button>}><div className="mt-4 flex items-center justify-between"><span className="text-xs text-muted-foreground">Receive events</span><Switch aria-label={`Enable ${webhook.name}`} checked={webhook.isActive} onCheckedChange={() => void toggle(webhook)}/></div><div className="mt-5 flex flex-wrap gap-1">{webhook.events.map(event => <span key={event} className="rounded-full bg-violet-50 px-2 py-1 text-xs text-violet-600">{event}</span>)}</div></CollectionCard>)}</div>
          <CollectionPagination page={page} limit={20} total={query.data?.total ?? 0} onPageChange={setPage}/>
        </>}
      </section>
      <AddWebhookDialog
        open={isAddDialogOpen}
        onOpenChange={setIsAddDialogOpen}
        onSuccess={() => {
          setIsAddDialogOpen(false);
          fetchWebhooks();
        }}
      />
    </div>
  );
}

function AddWebhookDialog({
  open,
  onOpenChange,
  onSuccess,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => void;
}) {
  const { apiFetch } = useApi();
  const [loading, setLoading] = useState(false);
  const [name, setName] = useState("");
  const [url, setUrl] = useState("");
  const [secret, setSecret] = useState("");
  const [selectedEvents, setSelectedEvents] = useState<WebhookEventType[]>([]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    try {
      const response = await apiFetch("webhooks", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          name,
          url,
          events: selectedEvents,
          secret,
          isActive: true,
        }),
      });

      if (!response.ok) throw new Error("Failed to create webhook");

      toast.success("Webhook created successfully");

      onSuccess();
    } catch (error) {
      toast.error("Failed to create webhook");
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Add Webhook</DialogTitle>
          <DialogDescription>
            Create a new webhook to receive event notifications
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Name</Label>
            <Input
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="My Webhook"
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="url">URL</Label>
            <Input
              id="url"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="https://api.example.com/webhook"
              required
              type="url"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="webhook-secret">Signing secret</Label>
            <Input id="webhook-secret" type="password" required minLength={16} value={secret} onChange={event => setSecret(event.target.value)} autoComplete="new-password" placeholder="At least 16 characters"/>
          </div>
          <div className="space-y-2">
            <Label>Events</Label>
            <Select
              onValueChange={(value) =>
                setSelectedEvents((prev) =>
                  prev.includes(value as WebhookEventType)
                    ? prev
                    : [...prev, value as WebhookEventType]
                )
              }
            >
              <SelectTrigger>
                <SelectValue placeholder="Select events" />
              </SelectTrigger>
              <SelectContent>
                {["open", "click", "reply", "bounce", "complaint"].map((event) => (
                  <SelectItem key={event} value={event}>
                    {event}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            <div className="flex flex-wrap gap-1 mt-2">
              {selectedEvents.map((event) => (
                <Badge
                  key={event}
                  variant="secondary"
                  className="cursor-pointer"
                  onClick={() =>
                    setSelectedEvents((prev) =>
                      prev.filter((e) => e !== event)
                    )
                  }
                >
                  {event}
                  <span className="ml-1">×</span>
                </Badge>
              ))}
            </div>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={loading || !selectedEvents.length}>
              {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Create Webhook
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
} 