"use client";

import { WebhookDeliveries } from "@/components/settings/webhook-deliveries";
import { PageHeader } from "@/components/page-header";
import { TeamProvider } from "@/app/providers/team-provider";
import { useEffect, useState } from "react";
import { Webhook } from "@/lib";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";
import { useApi } from "@/hooks/use-api";
import { resourceEntity } from "@/lib/resource-response";
export default function WebhookDeliveriesPage({
  params,
}: {
  params: Promise<{ webhookId: string }>;
}) {
  const { apiFetch, session } = useApi();
  const [webhook, setWebhook] = useState<Webhook | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchWebhook = async () => {
      try {
        const { webhookId } = await params;
        const response = await apiFetch(`webhooks/${webhookId}`);
        if (!response.ok) throw new Error("Failed to fetch webhook");
        const data = await response.json();
        setWebhook(resourceEntity<Webhook>(data));
      } catch (error) {
        toast.error("Failed to load webhook details");
      } finally {
        setLoading(false);
      }
    };

    if (session?.accessToken) void fetchWebhook();
  }, [params, session?.accessToken]);

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-64">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    );
  }

  if (!webhook) {
    return (
      <div className="flex items-center justify-center min-h-64">
        <div className="text-center">
          <h2 className="text-lg font-medium">Webhook not found</h2>
          <p className="text-muted-foreground">
            The webhook you're looking for doesn't exist or you don't have access to it.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col">
      <PageHeader
        heading={`Webhook Deliveries: ${webhook.name}`}
        description="View webhook delivery history and debug delivery issues"
      />
      <div className="px-4 py-6">
        <TeamProvider>
          <WebhookDeliveries webhookId={webhook.id} />
        </TeamProvider>
      </div>
    </div>
  );
} 