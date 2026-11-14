"use client";

import { Suspense, use, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import {
  CampaignsProvider,
  useCampaigns,
} from "@/app/providers/campaigns-provider";
import { Campaign, EmailTemplate } from "@/lib";
import { PageHeader } from "@/components/page-header";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";
import { useApi } from "@/hooks/use-api";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { CampaignReport } from "@/components/campaigns/campaign-report";
import { Status } from "@/components/marketing/shared";
import { useMarketingQuery } from "@/lib/marketing/api";
import type { Options } from "@/lib/marketing/types";
import { AlertTriangle, CalendarClock, Copy, Save, Trash2 } from "lucide-react";
import { CopyCampaignDialog } from "@/components/campaigns/copy-campaign";
import { useQuery } from "@tanstack/react-query";

const CampaignViewPage = ({ params }: { params: Promise<{ id: string }> }) => {
  const router = useRouter();
  const { id } = use(params);
  const { getCampaign, refetch } = useCampaigns();
  const { apiFetch, session } = useApi();
  const options = useMarketingQuery<Options>("marketing/options");
  const [template, setTemplate] = useState<EmailTemplate | null>(null);
  const [templateVariables, setTemplateVariables] = useState<
    Record<string, any>
  >({});
  const [isSaving, setIsSaving] = useState(false);
  const [loadingTemplate, setLoadingTemplate] = useState(false);
  const [copying, setCopying] = useState(false);

  const {
    data: campaign,
    error: campaignError,
    refetch: reloadCampaign,
  } = useQuery({
    queryKey: ["campaign", session?.user?.teamId, id],
    queryFn: () => getCampaign(id as string),
    enabled: !!id && !!session?.accessToken,
  });

  useEffect(() => {
    if (campaign) {
      // Initialize template variables from campaign
      if (campaign.template?.variables) {
        setTemplateVariables(
          campaign.template.variables.reduce(
            (acc, varName) => {
              acc[varName] = "";
              return acc;
            },
            {} as Record<string, any>,
          ),
        );
      }
      // Fetch the associated template
      if (campaign.templateId) {
        fetchTemplate(campaign.templateId);
      }
    }
  }, [campaign]);

  const fetchTemplate = async (templateId: string) => {
    try {
      setLoadingTemplate(true);
      const response = await apiFetch("templates/" + templateId, {
        method: "GET",
      });

      if (response.ok) {
        const payload = await response.json();
        const data = { data: payload.data ?? payload };
        setTemplate(data.data);

        // Initialize empty values for variables that don't have values yet
        if (data.data.variables && data.data.variables.length > 0) {
          const existingVars = data.data.variables.reduce(
            (acc, varName) => {
              acc[varName] = "";
              return acc;
            },
            {} as Record<string, any>,
          );
          const initialVars: Record<string, any> = {};

          data.data.variables.forEach((varName: string) => {
            initialVars[varName] = existingVars[varName] || "";
          });

          setTemplateVariables(initialVars);
        }
      } else {
        toast.error("Failed to load template");
      }
    } catch (error) {
      console.error("Error fetching template:", error);
      toast.error("Error loading template");
    } finally {
      setLoadingTemplate(false);
    }
  };

  const handleVariableChange = (variableName: string, value: string) => {
    setTemplateVariables((prev) => ({
      ...prev,
      [variableName]: value,
    }));
  };

  const handleSaveVariables = async () => {
    try {
      setIsSaving(true);
      const response = await apiFetch("campaigns/" + id, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          data: templateVariables,
        }),
      });

      if (response.ok) {
        toast.success("Template variables saved successfully");
        refetch();
      } else {
        toast.error("Error saving template variables");
      }
    } catch (error) {
      console.error("Error saving variables:", error);
      toast.error("Error saving template variables");
    } finally {
      setIsSaving(false);
    }
  };

  const [resuming, setResuming] = useState(false);
  const handleResume = async () => {
    setResuming(true);
    try {
      const response = await apiFetch(`marketing/campaigns/${id}/resume`, {
        method: "POST",
      });
      if (!response.ok) throw new Error();
      toast.success("Campaign resumed. The rest of the list sends now.");
      await reloadCampaign();
      refetch();
    } catch {
      toast.error("Couldn’t resume this campaign. Please try again.");
    } finally {
      setResuming(false);
    }
  };

  const handleCancel = async () => {
    setResuming(true);
    try {
      const response = await apiFetch(`marketing/campaigns/${id}/cancel`, {
        method: "POST",
      });
      if (!response.ok) throw new Error();
      toast.success(
        campaign?.status === "SCHEDULED"
          ? "Campaign cancelled. Nothing will be sent."
          : "Campaign cancelled. The rest of the list won’t be sent.",
      );
      await reloadCampaign();
      refetch();
    } catch {
      toast.error("Couldn’t cancel this campaign. Please try again.");
    } finally {
      setResuming(false);
    }
  };

  const handleDelete = async () => {
    const response = await apiFetch("campaigns/" + id, {
      method: "DELETE",
    });

    if (response.ok) {
      toast.success("Campaign deleted successfully");
      refetch();
      router.push("/campaigns");
    } else {
      toast.error("Error deleting campaign");
    }
  };

  if (campaignError)
    return (
      <div
        role="alert"
        className="rounded-2xl border border-border bg-white p-8"
      >
        Unable to load this campaign. Please try again.
      </div>
    );
  if (!campaign)
    return (
      <div role="status" className="p-8 text-muted-foreground">
        Loading campaign…
      </div>
    );

  return (
    <div className="container mx-auto">
      <PageHeader
        heading={campaign.name}
        description={campaign.subject || campaign.description}
        backButton={{ href: "/campaigns", label: "Campaigns" }}
      >
        <div className="flex gap-2 items-center">
          <Status value={campaign.status} />
          <Button variant="outline" onClick={() => setCopying(true)}>
            <Copy className="mr-2 h-4 w-4" />
            Copy
          </Button>
          <CopyCampaignDialog
            open={copying}
            onOpenChange={setCopying}
            campaign={campaign}
          />
          <Button variant="destructive" onClick={handleDelete}>
            <Trash2 className="mr-2 h-4 w-4" />
            Delete Campaign
          </Button>
        </div>
      </PageHeader>

      <div className="mt-4 px-8 space-y-6">
        {campaign.status === "PAUSED" && campaign.pauseReason && (
          <div
            role="alert"
            className="flex flex-wrap items-start gap-3 rounded-xl border border-amber-500/30 bg-amber-500/10 p-4 text-sm"
          >
            <AlertTriangle className="mt-0.5 size-4 shrink-0 text-amber-600" />
            <div className="min-w-0 flex-1">
              <p className="font-medium">
                Sending paused to protect your sender reputation
              </p>
              <p className="mt-1 text-muted-foreground">
                {campaign.pauseReason}
              </p>
              <p className="mt-1 text-muted-foreground">
                Resuming sends to the rest of the list at once, without further
                stages.
              </p>
            </div>
            <Button
              size="sm"
              variant="outline"
              onClick={() => void handleResume()}
              disabled={resuming}
            >
              {resuming ? "Resuming…" : "Resume sending"}
            </Button>
            <Button
              size="sm"
              variant="ghost"
              onClick={() => void handleCancel()}
              disabled={resuming}
            >
              Cancel the rest
            </Button>
          </div>
        )}
        {campaign.status === "SCHEDULED" && (
          <div
            role="status"
            className="flex flex-wrap items-center gap-3 rounded-xl border border-border bg-muted/50 p-4 text-sm"
          >
            <CalendarClock className="size-4 shrink-0 text-muted-foreground" />
            <p className="min-w-0 flex-1">
              {campaign.schedule === "RECURRING"
                ? "Sends on its recurring schedule."
                : campaign.scheduledFor &&
                    new Date(campaign.scheduledFor).getTime() > Date.now()
                  ? `Sends ${new Date(campaign.scheduledFor).toLocaleString(undefined, { dateStyle: "medium", timeStyle: "short" })}.`
                  : "Sending shortly."}
            </p>
            <Button
              size="sm"
              variant="outline"
              onClick={() => void handleCancel()}
              disabled={resuming}
            >
              Cancel campaign
            </Button>
          </div>
        )}
        <CampaignReport campaign={campaign} />

        <Card>
          <CardHeader>
            <CardTitle>Details</CardTitle>
          </CardHeader>
          <CardContent>
            <dl className="grid gap-x-8 gap-y-4 text-sm sm:grid-cols-2 lg:grid-cols-3">
              {(
                [
                  ["Subject", campaign.subject],
                  [
                    "Audience",
                    options.data?.lists.find((l) => l.id === campaign.listId)
                      ?.name,
                  ],
                  [
                    "From",
                    options.data?.senders.find(
                      (s) => s.id === campaign.smtpConfigId,
                    )?.fromEmail,
                  ],
                  ["Template", template?.name],
                  [
                    campaign.schedule === "RECURRING" ? "Repeats" : "Send time",
                    campaign.schedule === "RECURRING"
                      ? campaign.cronExpression ||
                        campaign.recurringSchedule?.toLowerCase()
                      : campaign.scheduledFor &&
                        new Date(campaign.scheduledFor).toLocaleString(
                          undefined,
                          { dateStyle: "medium", timeStyle: "short" },
                        ),
                  ],
                  [
                    "Created",
                    new Date(campaign.createdAt).toLocaleString(undefined, {
                      dateStyle: "medium",
                      timeStyle: "short",
                    }),
                  ],
                  ["Time zone", campaign.timezone],
                  [
                    "Open and click tracking",
                    campaign.trackEngagement === false ? "Off" : "On",
                  ],
                  [
                    "Staged sending",
                    campaign.sendInStages === false
                      ? "Off"
                      : campaign.sendInStages
                        ? "On"
                        : "Automatic",
                  ],
                ] as const
              ).map(([label, value]) => (
                <div key={label}>
                  <dt className="text-muted-foreground">{label}</dt>
                  <dd className="mt-0.5 break-words">{value || "—"}</dd>
                </div>
              ))}
            </dl>
            {campaign.description && campaign.subject && (
              <p className="mt-4 border-t pt-4 text-sm text-muted-foreground">
                {campaign.description}
              </p>
            )}
          </CardContent>
        </Card>

        {/* Template Variables Card */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>Template Variables</CardTitle>
                <CardDescription>
                  {template?.name
                    ? `Variables for template: ${template.name}`
                    : "Loading template..."}
                </CardDescription>
              </div>
              {template &&
                template.variables &&
                template.variables.length > 0 && (
                  <Button
                    onClick={handleSaveVariables}
                    disabled={isSaving}
                    size="sm"
                  >
                    <Save className="mr-2 h-4 w-4" />
                    {isSaving ? "Saving..." : "Save Variables"}
                  </Button>
                )}
            </div>
          </CardHeader>
          <CardContent>
            {loadingTemplate ? (
              <div className="text-center py-8 text-muted-foreground">
                Loading template variables...
              </div>
            ) : template &&
              template.variables &&
              template.variables.length > 0 ? (
              <div className="space-y-4">
                <p className="text-sm text-muted-foreground">
                  Set values for the template variables. Contact fields such as
                  first_name fill in from each contact; a value here replaces
                  them for everyone, and a blank one leaves them as they are.
                </p>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {template.variables.map((variable) => (
                    <div key={variable} className="space-y-2">
                      <Label htmlFor={`var-${variable}`}>
                        {variable}
                        <span className="ml-1 text-muted-foreground font-mono text-xs">
                          {`{{${variable}}}`}
                        </span>
                      </Label>
                      <Input
                        id={`var-${variable}`}
                        type="text"
                        placeholder={
                          contactFields.has(variable)
                            ? "Filled from each contact; leave blank"
                            : `Enter value for ${variable}`
                        }
                        value={templateVariables[variable] || ""}
                        onChange={(e) =>
                          handleVariableChange(variable, e.target.value)
                        }
                      />
                    </div>
                  ))}
                </div>
              </div>
            ) : (
              <div className="text-center py-8 text-muted-foreground">
                {template
                  ? "This template has no variables defined."
                  : "No template associated with this campaign."}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

const CampaignViewPageWrapper = ({
  params,
}: {
  params: Promise<{ id: string }>;
}) => {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <CampaignsProvider>
        <CampaignViewPage params={params} />
      </CampaignsProvider>
    </Suspense>
  );
};

// Filled per recipient from the contact when sending (server/internal/tasks).
const contactFields = new Set([
  "email",
  "first_name",
  "last_name",
  "name",
  "full_name",
  "company",
  "country",
  "city",
  "state",
  "zip",
  "address",
  "phone",
  "linkedin",
  "twitter",
  "facebook",
  "instagram",
]);

export default CampaignViewPageWrapper;
