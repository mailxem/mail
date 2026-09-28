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
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Save, Trash2 } from "lucide-react";
import { useQuery } from "@tanstack/react-query";

const CampaignViewPage = ({ params }: { params: Promise<{ id: string }> }) => {
  const router = useRouter();
  const { id } = use(params);
  const { getCampaign, refetch } = useCampaigns();
  const { apiFetch, session } = useApi();
  const [template, setTemplate] = useState<EmailTemplate | null>(null);
  const [templateVariables, setTemplateVariables] = useState<
    Record<string, any>
  >({});
  const [isSaving, setIsSaving] = useState(false);
  const [loadingTemplate, setLoadingTemplate] = useState(false);

  const { data: campaign, error: campaignError } = useQuery({
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
            {} as Record<string, any>
          )
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
            {} as Record<string, any>
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

  if (campaignError) return <div role="alert" className="rounded-2xl border border-border bg-white p-8">Unable to load this campaign. Please try again.</div>;
  if (!campaign) return <div role="status" className="p-8 text-muted-foreground">Loading campaign…</div>;

  return (
    <div className="container mx-auto">
      <PageHeader heading={campaign.name}>
        <div className="flex gap-2 items-center">
          <Button variant="outline" onClick={() => router.back()}>
            Back
          </Button>
          <Button variant="destructive" onClick={handleDelete}>
            <Trash2 className="mr-2 h-4 w-4" />
            Delete Campaign
          </Button>
        </div>
      </PageHeader>

      <div className="mt-4 px-8 space-y-6">
        {/* Campaign Details Card */}
        <Card>
          <CardHeader>
            <CardTitle>Campaign Details</CardTitle>
            <CardDescription>
              General information about this campaign
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <Label className="text-muted-foreground">Status</Label>
                <div className="mt-1">
                  <Badge
                    variant={
                      campaign.status === "COMPLETED"
                        ? "default"
                        : campaign.status === "SENDING"
                          ? "secondary"
                          : campaign.status === "SCHEDULED"
                            ? "outline"
                            : campaign.status === "FAILED"
                              ? "destructive"
                              : "secondary"
                    }
                  >
                    {campaign.status}
                  </Badge>
                </div>
              </div>
              <div>
                <Label className="text-muted-foreground">Created At</Label>
                <p className="mt-1">{campaign.createdAt.toLocaleString()}</p>
              </div>
            </div>

            {campaign.description && (
              <div>
                <Label className="text-muted-foreground">Description</Label>
                <p className="mt-1">{campaign.description}</p>
              </div>
            )}

            <Separator />

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {campaign.schedule && (
                <div>
                  <Label className="text-muted-foreground">Schedule</Label>
                  <p className="mt-1">{campaign.schedule}</p>
                </div>
              )}

              {campaign.scheduledFor && (
                <div>
                  <Label className="text-muted-foreground">Scheduled For</Label>
                  <p className="mt-1">
                    {campaign.scheduledFor.toLocaleString()}
                  </p>
                </div>
              )}

              {campaign.recurringSchedule && (
                <div>
                  <Label className="text-muted-foreground">
                    Recurring Schedule
                  </Label>
                  <p className="mt-1">{campaign.recurringSchedule}</p>
                </div>
              )}

              {campaign.cronExpression && (
                <div>
                  <Label className="text-muted-foreground">
                    Cron Expression
                  </Label>
                  <p className="mt-1 font-mono text-sm">
                    {campaign.cronExpression}
                  </p>
                </div>
              )}
            </div>
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
                  Set values for the template variables. These will be used when
                  sending emails for this campaign.
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
                        placeholder={`Enter value for ${variable}`}
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

export default CampaignViewPageWrapper;
