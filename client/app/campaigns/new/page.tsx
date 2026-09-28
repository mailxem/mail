"use client";

import { useEffect, useState, type ReactNode } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useForm, type FieldErrors } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { format } from "date-fns";
import * as z from "zod";
import {
  ArrowRight,
  ArrowUpRight,
  CalendarDays,
  Check,
  CheckCheck,
  ChevronRight,
  Clock3,
  FileText,
  LayoutTemplate,
  Loader2,
  Mail,
  Plus,
  Save,
  Send,
  Users,
} from "lucide-react";
import { toast } from "sonner";
import {
  TemplatesProvider,
  useTemplates,
} from "@/app/providers/templates-provider";
import {
  MailingListProvider,
  useMailingLists,
} from "@/app/providers/mailinglist-provider";
import { SMTPProvider, useSMTP } from "@/app/providers/smtp-provider";
import { useTeam } from "@/app/providers/team-provider";
import { useApi } from "@/hooks/use-api";
import { resourceEntity } from "@/lib/resource-response";
import { workspaceClassName } from "@/lib/workspace-styles";
import { cn } from "@/lib/utils";
import { PageHeader } from "@/components/page-header";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

const campaignSchema = z.object({
  name: z.string().trim().min(1, "Give your campaign a name"),
  description: z.string().optional(),
  data: z.record(z.string(), z.string()).optional(),
  templateId: z.string().uuid("Choose an email template"),
  status: z.literal("DRAFT").default("DRAFT"),
  scheduledFor: z.date().optional(),
  schedule: z.enum(["ONE_TIME", "RECURRING"]).default("ONE_TIME"),
  listId: z.string().uuid("Choose a contact list"),
  recurringSchedule: z
    .enum(["DAILY", "WEEKLY", "MONTHLY", "CUSTOM"])
    .optional(),
  cronExpression: z.string().optional(),
  batchSize: z.number().min(1).default(100),
  processed: z.number().default(0),
  batchDelay: z.number().min(1).optional(),
  timezone: z.string().default("UTC"),
  subject: z.string().trim().min(1, "Add a subject line"),
  smtpConfigId: z.string().uuid("Choose a sender"),
});
type CampaignFormValues = z.infer<typeof campaignSchema>;
type ContactPreview = {
  id: string;
  email: string;
  firstName?: string;
  lastName?: string;
};
const timezones = [
  "UTC",
  "America/New_York",
  "America/Chicago",
  "America/Los_Angeles",
  "Europe/London",
  "Asia/Kolkata",
  "Asia/Tokyo",
  "Asia/Shanghai",
];

export default function NewCampaignPage() {
  return (
    <SMTPProvider>
      <TemplatesProvider>
        <MailingListProvider>
          <NewCampaignForm />
        </MailingListProvider>
      </TemplatesProvider>
    </SMTPProvider>
  );
}

function ResourceFeedback({
  loading,
  error,
  retry,
  empty,
  href,
  label,
}: {
  loading: boolean;
  error: Error | null;
  retry: () => void;
  empty: boolean;
  href: string;
  label: string;
}) {
  if (loading)
    return (
      <p
        role="status"
        className="flex items-center gap-2 text-sm text-muted-foreground"
      >
        <Loader2 className="size-4 animate-spin" />
        Loading {label.toLowerCase()}…
      </p>
    );
  if (error)
    return (
      <div
        role="alert"
        className="flex flex-wrap items-center justify-between gap-2 rounded-xl border border-destructive/20 bg-destructive/5 p-3 text-sm"
      >
        <span>Unable to load {label.toLowerCase()}.</span>
        <Button type="button" size="sm" variant="outline" onClick={retry}>
          Try again
        </Button>
      </div>
    );
  if (empty)
    return (
      <div className="rounded-xl border border-dashed p-4 text-sm text-muted-foreground">
        No {label.toLowerCase()} yet.{" "}
        <Link
          href={href}
          target="_blank"
          rel="noopener noreferrer"
          className="font-medium text-primary underline underline-offset-4"
        >
          Add one to get started
        </Link>
        .
      </div>
    );
  return null;
}

function SetupSection({
  id,
  title,
  description,
  icon,
  complete,
  children,
}: {
  id: string;
  title: string;
  description: string;
  icon: ReactNode;
  complete: boolean;
  children: ReactNode;
}) {
  return (
    <AccordionItem
      value={id}
      className="overflow-hidden rounded-[20px] border border-border bg-white shadow-sm"
    >
      <AccordionTrigger className="min-h-[88px] gap-4 px-5 py-5 text-left hover:no-underline sm:px-6 [&>svg]:text-muted-foreground">
        <span className="flex min-w-0 flex-1 items-center gap-3.5">
          <span
            className={cn(
              "flex size-11 shrink-0 items-center justify-center rounded-xl border",
              complete
                ? "border-emerald-100 bg-emerald-50 text-emerald-600"
                : "border-primary/10 bg-primary/5 text-primary",
            )}
          >
            {complete ? <Check size={20} /> : icon}
          </span>
          <span className="min-w-0">
            <span className="block text-sm font-semibold text-foreground sm:text-base">
              {title}
            </span>
            <span className="mt-1 block truncate text-xs font-normal text-muted-foreground sm:text-sm">
              {description}
            </span>
          </span>
        </span>
        {complete && (
          <span className="hidden text-xs font-medium text-emerald-600 sm:block">
            Added
          </span>
        )}
      </AccordionTrigger>
      <AccordionContent className="border-t border-border px-5 pb-6 pt-5 sm:px-6">
        <div className="space-y-5">{children}</div>
      </AccordionContent>
    </AccordionItem>
  );
}

function NewCampaignForm() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const { team } = useTeam();
  const templateResource = useTemplates();
  const listResource = useMailingLists();
  const senderResource = useSMTP();
  const { templates } = templateResource;
  const { lists } = listResource;
  const { configs: senders } = senderResource;
  const { apiFetch, session } = useApi();
  const [currentStep, setCurrentStep] = useState("audience");
  const [isScheduled, setIsScheduled] = useState(false);
  const [deviceTimezone, setDeviceTimezone] = useState("UTC");
  const form = useForm<CampaignFormValues>({
    resolver: zodResolver(campaignSchema),
    defaultValues: {
      name: "",
      description: "",
      templateId: "",
      subject: "",
      status: "DRAFT",
      listId: "",
      smtpConfigId: "",
      schedule: "ONE_TIME",
      data: {},
      batchSize: 100,
      processed: 0,
      timezone: "UTC",
    },
  });
  useEffect(() => {
    const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
    setDeviceTimezone(timezone);
    form.setValue("timezone", timezone);
  }, [form]);
  const draft = form.watch();
  const selectedList = lists.find((list) => list.id === draft.listId);
  const selectedSender = senders.find(
    (sender) => sender.id === draft.smtpConfigId,
  );
  const selectedTemplate = templates.find(
    (template) => template.id === draft.templateId,
  );
  const contactsQuery = useQuery<ContactPreview[]>({
    queryKey: ["campaign-contact-preview", team?.id, draft.listId],
    queryFn: async ({ signal }) => {
      const response = await apiFetch(
        `contacts?limit=5&sort=created_at&order=desc&list_id=${draft.listId}`,
        { signal },
      );
      if (!response.ok) throw new Error("Unable to load the audience preview");
      return (await response.json()).data || [];
    },
    enabled: !!draft.listId && !!team?.id && !!session?.accessToken,
  });
  const checklist = [
    {
      title: "Campaign details",
      detail: draft.name.trim() || "Give your campaign a name",
      done: !!draft.name.trim(),
      step: "details",
      icon: FileText,
    },
    {
      title: "Audience",
      detail: selectedList?.name || "Choose a contact list",
      done: !!selectedList,
      step: "audience",
      icon: Users,
    },
    {
      title: "Sender",
      detail: selectedSender?.fromEmail || "Choose a sender",
      done: !!selectedSender,
      step: "sender",
      icon: Send,
    },
    {
      title: "Email content",
      detail:
        selectedTemplate && draft.subject.trim()
          ? selectedTemplate.name
          : "Add a subject and template",
      done: !!selectedTemplate && !!draft.subject.trim(),
      step: "content",
      icon: LayoutTemplate,
    },
  ];
  const completed = checklist.filter((item) => item.done).length;
  const saving = form.formState.isSubmitting;
  const goToStep = (step: string) => {
    if (step === "details") {
      form.setFocus("name");
      return;
    }
    setCurrentStep(step);
  };
  const invalidSubmit = (errors: FieldErrors<CampaignFormValues>) => {
    if (errors.name) form.setFocus("name");
    else if (errors.listId) setCurrentStep("audience");
    else if (errors.smtpConfigId) setCurrentStep("sender");
    else if (errors.subject || errors.templateId) setCurrentStep("content");
    else setCurrentStep("delivery");
    toast.error("Complete the highlighted campaign details.");
  };
  const onSubmit = async (values: CampaignFormValues) => {
    if (
      isScheduled &&
      (!values.scheduledFor || values.scheduledFor.getTime() <= Date.now())
    ) {
      form.setError("scheduledFor", {
        message: "Choose a date and time in the future",
      });
      setCurrentStep("delivery");
      return;
    }
    if (
      isScheduled &&
      values.schedule === "RECURRING" &&
      !values.recurringSchedule
    ) {
      form.setError("recurringSchedule", {
        message: "Choose how often to repeat",
      });
      setCurrentStep("delivery");
      return;
    }
    if (
      isScheduled &&
      values.schedule === "RECURRING" &&
      values.recurringSchedule === "CUSTOM" &&
      !values.cronExpression?.trim()
    ) {
      form.setError("cronExpression", { message: "Enter a cron expression" });
      setCurrentStep("delivery");
      return;
    }
    try {
      const response = await apiFetch("campaigns", {
        method: "POST",
        body: JSON.stringify({
          ...values,
          teamId: team?.id,
          status: "DRAFT",
          scheduledFor: isScheduled ? values.scheduledFor : undefined,
          schedule: isScheduled ? values.schedule : "ONE_TIME",
          recurringSchedule:
            isScheduled && values.schedule === "RECURRING"
              ? values.recurringSchedule
              : undefined,
          cronExpression:
            isScheduled &&
            values.schedule === "RECURRING" &&
            values.recurringSchedule === "CUSTOM"
              ? values.cronExpression
              : undefined,
        }),
      });
      if (!response.ok)
        throw new Error(
          "Unable to save your campaign. Your changes are still here.",
        );
      const payload = await response.json();
      const campaign = resourceEntity<{ id: string }>(
        payload.campaign || payload,
      );
      await queryClient.invalidateQueries({ queryKey: ["campaigns"] });
      toast.success("Campaign draft saved");
      router.push(`/campaigns/${campaign.id}`);
    } catch (error) {
      toast.error((error as Error).message);
    }
  };

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit, invalidSubmit)}
        className="min-w-0"
      >
        <PageHeader
          heading="New campaign"
          description="A great email starts with the right details."
          backButton={{ href: "/campaigns", label: "Back to campaigns" }}
        >
          <Button type="button" variant="outline" asChild>
            <Link href="/campaigns">Cancel</Link>
          </Button>
          <Button type="submit" disabled={saving || !team?.id}>
            {saving ? (
              <Loader2 className="size-4 animate-spin" />
            ) : (
              <Save className="size-4" />
            )}
            {saving ? "Saving…" : "Save draft"}
          </Button>
        </PageHeader>

        <div className="grid min-w-0 items-start gap-6 xl:grid-cols-[minmax(0,1fr)_340px] 2xl:grid-cols-[minmax(0,1fr)_380px]">
          <div className="min-w-0 space-y-5">
            <section
              className={workspaceClassName("product-panel")}
              aria-labelledby="campaign-details-heading"
            >
              <div className="mb-6 flex items-center justify-between gap-3">
                <div className="flex items-center gap-3">
                  <span className="flex size-10 items-center justify-center rounded-xl bg-primary/5 text-primary">
                    <FileText size={19} />
                  </span>
                  <h2
                    id="campaign-details-heading"
                    className="text-base font-semibold tracking-tight"
                  >
                    Campaign details
                  </h2>
                </div>
                <Badge
                  variant="secondary"
                  className="rounded-full px-3 py-1 text-xs font-normal"
                >
                  Draft
                </Badge>
              </div>
              <div className="space-y-5">
                <FormField
                  control={form.control}
                  name="name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Campaign name</FormLabel>
                      <FormControl>
                        <Input
                          {...field}
                          placeholder="e.g. September product update"
                          className="h-11"
                        />
                      </FormControl>
                      <FormDescription>
                        A name to help you find this campaign. Only your team
                        sees it.
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="description"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>
                        Description{" "}
                        <span className="ml-1 font-normal text-muted-foreground">
                          (optional)
                        </span>
                      </FormLabel>
                      <FormControl>
                        <Textarea
                          {...field}
                          rows={2}
                          placeholder="What’s this campaign about?"
                          className="min-h-[80px] resize-y rounded-xl border-input bg-white"
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
            </section>

            <Accordion
              type="single"
              collapsible
              value={currentStep}
              onValueChange={setCurrentStep}
              className="space-y-4"
            >
              <SetupSection
                id="audience"
                title="Audience"
                description={
                  selectedList
                    ? `${selectedList.name} · ${selectedList.subscribersCount.toLocaleString()} subscribers`
                    : "Who would you like to reach?"
                }
                icon={<Users size={20} />}
                complete={!!selectedList}
              >
                <ResourceFeedback
                  loading={listResource.isLoading}
                  error={listResource.error}
                  retry={listResource.refetch}
                  empty={!lists.length}
                  href="/audience/lists"
                  label="Contact lists"
                />
                <FormField
                  control={form.control}
                  name="listId"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Contact list</FormLabel>
                      <Select
                        value={field.value}
                        onValueChange={field.onChange}
                        disabled={
                          listResource.isLoading ||
                          !!listResource.error ||
                          !lists.length
                        }
                      >
                        <FormControl>
                          <SelectTrigger className="h-11">
                            <SelectValue placeholder="Select your audience" />
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          {lists.map((list) => (
                            <SelectItem key={list.id} value={list.id}>
                              {list.name} ·{" "}
                              {list.subscribersCount.toLocaleString()}{" "}
                              subscribers
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                {selectedList && (
                  <div className="rounded-xl border border-border bg-muted/40 p-4">
                    <div className="mb-3 flex items-center justify-between gap-2">
                      <h3 className="text-xs font-medium text-muted-foreground">
                        Contact preview
                      </h3>
                      <Link
                        href={`/audience/lists/${selectedList.id}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="flex items-center gap-1 text-xs font-medium text-primary"
                      >
                        View list
                        <ArrowUpRight size={13} />
                      </Link>
                    </div>
                    {contactsQuery.isPending ? (
                      <p
                        role="status"
                        className="text-sm text-muted-foreground"
                      >
                        Loading contacts…
                      </p>
                    ) : contactsQuery.error ? (
                      <div className="flex items-center justify-between gap-2 text-sm text-muted-foreground">
                        Unable to load contacts.
                        <Button
                          type="button"
                          size="sm"
                          variant="ghost"
                          onClick={() => void contactsQuery.refetch()}
                        >
                          Try again
                        </Button>
                      </div>
                    ) : !contactsQuery.data?.length ? (
                      <p className="text-sm text-muted-foreground">
                        No contacts in this list yet.
                      </p>
                    ) : (
                      <div className="space-y-3">
                        {contactsQuery.data.map((contact) => (
                          <div
                            key={contact.id}
                            className="flex min-w-0 items-center gap-3"
                          >
                            <span className="flex size-8 shrink-0 items-center justify-center rounded-full border border-primary/10 bg-white text-xs font-medium text-primary">
                              {(contact.firstName || contact.email)
                                .slice(0, 1)
                                .toUpperCase()}
                            </span>
                            <div className="min-w-0">
                              <p className="truncate text-sm">
                                {contact.email}
                              </p>
                              {(contact.firstName || contact.lastName) && (
                                <p className="truncate text-xs text-muted-foreground">
                                  {[contact.firstName, contact.lastName]
                                    .filter(Boolean)
                                    .join(" ")}
                                </p>
                              )}
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                )}
                <div className="flex flex-wrap items-center justify-between gap-3 pt-1">
                  <Button type="button" variant="ghost" size="sm" asChild>
                    <Link
                      href="/audience/lists"
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      <Plus size={15} />
                      Manage lists
                    </Link>
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => setCurrentStep("sender")}
                  >
                    Sender details
                    <ArrowRight size={14} />
                  </Button>
                </div>
              </SetupSection>

              <SetupSection
                id="sender"
                title="Sender"
                description={
                  selectedSender?.fromEmail ||
                  "Make it clear who’s saying hello."
                }
                icon={<Send size={20} />}
                complete={!!selectedSender}
              >
                <ResourceFeedback
                  loading={senderResource.isLoading}
                  error={senderResource.error}
                  retry={senderResource.refresh}
                  empty={!senders.length}
                  href="/settings/smtp"
                  label="Senders"
                />
                <FormField
                  control={form.control}
                  name="smtpConfigId"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>From address</FormLabel>
                      <Select
                        value={field.value}
                        onValueChange={field.onChange}
                        disabled={
                          senderResource.isLoading ||
                          !!senderResource.error ||
                          !senders.length
                        }
                      >
                        <FormControl>
                          <SelectTrigger className="h-11">
                            <SelectValue placeholder="Choose a sender" />
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          {senders.map((sender) => (
                            <SelectItem key={sender.id} value={sender.id}>
                              {sender.fromEmail || sender.username} ·{" "}
                              {sender.provider}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      <FormDescription>
                        Use one of your connected email senders.
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                {selectedSender && (
                  <div className="flex items-center gap-3 rounded-xl border border-primary/10 bg-primary/5 p-4">
                    <span className="flex size-10 shrink-0 items-center justify-center rounded-full bg-white text-primary">
                      <Mail size={18} />
                    </span>
                    <div className="min-w-0">
                      <p className="truncate text-sm font-medium">
                        {selectedSender.fromEmail}
                      </p>
                      <p className="mt-1 text-xs text-muted-foreground">
                        Connected via {selectedSender.provider}
                      </p>
                    </div>
                  </div>
                )}
                <div className="flex flex-wrap items-center justify-between gap-3 pt-1">
                  <Button type="button" size="sm" variant="ghost" asChild>
                    <Link
                      href="/settings/smtp"
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      <Plus size={15} />
                      Manage senders
                    </Link>
                  </Button>
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    onClick={() => setCurrentStep("content")}
                  >
                    Email content
                    <ArrowRight size={14} />
                  </Button>
                </div>
              </SetupSection>

              <SetupSection
                id="content"
                title="Email content"
                description={
                  selectedTemplate && draft.subject.trim()
                    ? draft.subject
                    : "Choose your template and find the right words."
                }
                icon={<LayoutTemplate size={20} />}
                complete={!!selectedTemplate && !!draft.subject.trim()}
              >
                <FormField
                  control={form.control}
                  name="subject"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Subject line</FormLabel>
                      <FormControl>
                        <Input
                          {...field}
                          placeholder="Give your readers a reason to open"
                          className="h-11"
                        />
                      </FormControl>
                      <FormDescription>
                        Keep it clear, personal, and true to your message.
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <ResourceFeedback
                  loading={templateResource.isLoading}
                  error={templateResource.error}
                  retry={templateResource.refetch}
                  empty={!templates.length}
                  href="/templates/new"
                  label="Templates"
                />
                <FormField
                  control={form.control}
                  name="templateId"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Email template</FormLabel>
                      <Select
                        value={field.value}
                        onValueChange={(value) => {
                          field.onChange(value);
                          form.setValue("data", {});
                        }}
                        disabled={
                          templateResource.isLoading ||
                          !!templateResource.error ||
                          !templates.length
                        }
                      >
                        <FormControl>
                          <SelectTrigger className="h-11">
                            <SelectValue placeholder="Choose from your template library" />
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          {templates.map((template) => (
                            <SelectItem key={template.id} value={template.id}>
                              {template.name}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                {selectedTemplate && (
                  <div className="overflow-hidden rounded-xl border border-border">
                    <div className="flex items-center gap-3 bg-muted/40 p-4">
                      <span className="flex size-12 shrink-0 items-center justify-center rounded-xl border border-border bg-white text-primary">
                        <LayoutTemplate size={23} />
                      </span>
                      <div className="min-w-0 flex-1">
                        <p className="truncate text-sm font-medium">
                          {selectedTemplate.name}
                        </p>
                        <p className="mt-1 truncate text-xs text-muted-foreground">
                          {selectedTemplate.subject ||
                            "Your existing email design"}
                        </p>
                      </div>
                      <Button type="button" variant="ghost" size="sm" asChild>
                        <Link
                          href={`/templates/${selectedTemplate.id}/edit`}
                          target="_blank"
                          rel="noopener noreferrer"
                        >
                          Edit
                          <ArrowUpRight size={14} />
                        </Link>
                      </Button>
                    </div>
                    {!!selectedTemplate.variables?.length && (
                      <div className="space-y-4 border-t border-border p-4">
                        <h3 className="text-sm font-medium">Personalization</h3>
                        {selectedTemplate.variables.map((variable) => (
                          <div key={variable} className="space-y-1.5">
                            <label
                              htmlFor={`campaign-variable-${variable}`}
                              className="font-mono text-xs text-muted-foreground"
                            >{`{{${variable}}}`}</label>
                            {variable === "name" ? (
                              <p className="text-sm">
                                Filled from each contact’s name
                              </p>
                            ) : (
                              <Input
                                id={`campaign-variable-${variable}`}
                                value={draft.data?.[variable] || ""}
                                onChange={(event) =>
                                  form.setValue(
                                    "data",
                                    {
                                      ...form.getValues("data"),
                                      [variable]: event.target.value,
                                    },
                                    { shouldDirty: true },
                                  )
                                }
                                placeholder={`Value for ${variable}`}
                                className="h-10"
                              />
                            )}
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                )}
                <div className="flex flex-wrap items-center justify-between gap-3 pt-1">
                  <Button type="button" size="sm" variant="ghost" asChild>
                    <Link
                      href="/templates/new"
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      <Plus size={15} />
                      Create template
                    </Link>
                  </Button>
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    onClick={() => setCurrentStep("delivery")}
                  >
                    Delivery timing
                    <ArrowRight size={14} />
                  </Button>
                </div>
              </SetupSection>

              <SetupSection
                id="delivery"
                title="Delivery timing"
                description={
                  isScheduled && draft.scheduledFor
                    ? format(draft.scheduledFor, "MMM d, yyyy · h:mm a")
                    : "Choose your timing when you’re ready."
                }
                icon={<CalendarDays size={20} />}
                complete={false}
              >
                <div
                  role="group"
                  aria-label="Delivery timing"
                  className="grid gap-3 sm:grid-cols-2"
                >
                  {[
                    {
                      scheduled: false,
                      title: "Decide later",
                      description: "Keep this campaign as a draft.",
                      icon: FileText,
                    },
                    {
                      scheduled: true,
                      title: "Plan a send time",
                      description: "Save a date with your draft.",
                      icon: CalendarDays,
                    },
                  ].map((option) => (
                    <button
                      key={option.title}
                      type="button"
                      aria-pressed={isScheduled === option.scheduled}
                      onClick={() => {
                        setIsScheduled(option.scheduled);
                        form.clearErrors([
                          "scheduledFor",
                          "recurringSchedule",
                          "cronExpression",
                        ]);
                      }}
                      className={cn(
                        "rounded-xl border p-4 text-left transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
                        isScheduled === option.scheduled
                          ? "border-primary/40 bg-primary/5"
                          : "border-border bg-white hover:bg-muted/30",
                      )}
                    >
                      <span className="mb-3 flex items-center justify-between">
                        <option.icon
                          size={18}
                          className={
                            isScheduled === option.scheduled
                              ? "text-primary"
                              : "text-muted-foreground"
                          }
                        />
                        <span
                          className={cn(
                            "flex size-4 items-center justify-center rounded-full border",
                            isScheduled === option.scheduled
                              ? "border-primary bg-primary text-white"
                              : "border-border",
                          )}
                        >
                          {isScheduled === option.scheduled && (
                            <Check size={10} />
                          )}
                        </span>
                      </span>
                      <span className="block text-sm font-medium">
                        {option.title}
                      </span>
                      <span className="mt-1 block text-xs leading-relaxed text-muted-foreground">
                        {option.description}
                      </span>
                    </button>
                  ))}
                </div>
                {isScheduled && (
                  <div className="space-y-5 border-t border-border pt-5">
                    <FormField
                      control={form.control}
                      name="scheduledFor"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Date and time</FormLabel>
                          <FormControl>
                            <Input
                              type="datetime-local"
                              value={
                                field.value
                                  ? format(field.value, "yyyy-MM-dd'T'HH:mm")
                                  : ""
                              }
                              onChange={(event) =>
                                field.onChange(
                                  event.target.value
                                    ? new Date(event.target.value)
                                    : undefined,
                                )
                              }
                              className="h-11"
                            />
                          </FormControl>
                          <FormDescription>
                            Choose a time in your device’s timezone:{" "}
                            {deviceTimezone}.
                          </FormDescription>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <div className="grid gap-5 sm:grid-cols-2">
                      <FormField
                        control={form.control}
                        name="schedule"
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>Frequency</FormLabel>
                            <Select
                              value={field.value}
                              onValueChange={field.onChange}
                            >
                              <FormControl>
                                <SelectTrigger className="h-11">
                                  <SelectValue />
                                </SelectTrigger>
                              </FormControl>
                              <SelectContent>
                                <SelectItem value="ONE_TIME">
                                  One time
                                </SelectItem>
                                <SelectItem value="RECURRING">
                                  Recurring
                                </SelectItem>
                              </SelectContent>
                            </Select>
                            <FormMessage />
                          </FormItem>
                        )}
                      />
                      {draft.schedule === "RECURRING" && (
                        <FormField
                          control={form.control}
                          name="recurringSchedule"
                          render={({ field }) => (
                            <FormItem>
                              <FormLabel>Repeat</FormLabel>
                              <Select
                                value={field.value || ""}
                                onValueChange={field.onChange}
                              >
                                <FormControl>
                                  <SelectTrigger className="h-11">
                                    <SelectValue placeholder="Choose frequency" />
                                  </SelectTrigger>
                                </FormControl>
                                <SelectContent>
                                  <SelectItem value="DAILY">Daily</SelectItem>
                                  <SelectItem value="WEEKLY">Weekly</SelectItem>
                                  <SelectItem value="MONTHLY">
                                    Monthly
                                  </SelectItem>
                                  <SelectItem value="CUSTOM">
                                    Custom (cron)
                                  </SelectItem>
                                </SelectContent>
                              </Select>
                              <FormMessage />
                            </FormItem>
                          )}
                        />
                      )}
                    </div>
                    {draft.schedule === "RECURRING" && (
                      <FormField
                        control={form.control}
                        name="timezone"
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>Recurring timezone</FormLabel>
                            <Select
                              value={field.value}
                              onValueChange={field.onChange}
                            >
                              <FormControl>
                                <SelectTrigger className="h-11">
                                  <SelectValue />
                                </SelectTrigger>
                              </FormControl>
                              <SelectContent>
                                {Array.from(
                                  new Set([deviceTimezone, ...timezones]),
                                ).map((timezone) => (
                                  <SelectItem key={timezone} value={timezone}>
                                    {timezone.replaceAll("_", " ")}
                                  </SelectItem>
                                ))}
                              </SelectContent>
                            </Select>
                            <FormMessage />
                          </FormItem>
                        )}
                      />
                    )}
                    {draft.schedule === "RECURRING" &&
                      draft.recurringSchedule === "CUSTOM" && (
                        <FormField
                          control={form.control}
                          name="cronExpression"
                          render={({ field }) => (
                            <FormItem>
                              <FormLabel>Cron expression</FormLabel>
                              <FormControl>
                                <Input
                                  {...field}
                                  value={field.value || ""}
                                  placeholder="0 9 * * 1"
                                  className="h-11 font-mono"
                                />
                              </FormControl>
                              <FormDescription>
                                For example, 0 9 * * 1 means Mondays at 9am.
                              </FormDescription>
                              <FormMessage />
                            </FormItem>
                          )}
                        />
                      )}
                  </div>
                )}
                <p className="flex items-start gap-2 rounded-xl bg-muted/50 p-3 text-xs leading-relaxed text-muted-foreground">
                  <Clock3 className="mt-0.5 size-3.5 shrink-0" />
                  Saving creates a draft. Your campaign will not send until you
                  launch it.
                </p>
              </SetupSection>
            </Accordion>
          </div>

          <aside
            className="min-w-0 space-y-5 xl:sticky xl:top-6"
            aria-label="Campaign summary"
          >
            <section className={workspaceClassName("product-panel")}>
              <div className="flex items-center justify-between gap-3">
                <h2 className="text-base font-semibold tracking-tight">
                  Campaign checklist
                </h2>
                <span className="text-xs text-muted-foreground">
                  {completed} of 4
                </span>
              </div>
              <div
                className="mb-5 mt-4 flex gap-1.5"
                role="img"
                aria-label={`${completed} of 4 campaign details completed`}
              >
                {checklist.map((item, index) => (
                  <span
                    key={item.step}
                    className={cn(
                      "h-1.5 flex-1 rounded-full",
                      index < completed ? "bg-primary" : "bg-muted",
                    )}
                  />
                ))}
              </div>
              <div className="space-y-1">
                {checklist.map((item) => (
                  <button
                    key={item.step}
                    type="button"
                    onClick={() => goToStep(item.step)}
                    className="flex w-full min-w-0 items-center gap-3 rounded-xl px-1 py-3 text-left transition-colors hover:bg-muted/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  >
                    <span
                      className={cn(
                        "flex size-7 shrink-0 items-center justify-center rounded-full border",
                        item.done
                          ? "border-emerald-100 bg-emerald-50 text-emerald-600"
                          : "border-border text-muted-foreground",
                      )}
                    >
                      {item.done ? (
                        <Check size={14} />
                      ) : (
                        <item.icon size={13} />
                      )}
                    </span>
                    <span className="min-w-0 flex-1">
                      <span className="block text-xs font-medium">
                        {item.title}
                      </span>
                      <span className="mt-1 block truncate text-xs text-muted-foreground">
                        {item.detail}
                      </span>
                    </span>
                    <ChevronRight
                      size={13}
                      className="shrink-0 text-muted-foreground"
                    />
                  </button>
                ))}
              </div>
              <div className="mt-4 border-t border-border pt-4">
                <div className="flex items-start gap-2 text-xs leading-relaxed text-muted-foreground">
                  <CheckCheck
                    size={16}
                    className="mt-0.5 shrink-0 text-primary"
                  />
                  <p>
                    {completed === 4
                      ? "All the essentials are in place. Save your draft to continue."
                      : "Work through the details at your own pace. Save when the essentials are ready."}
                  </p>
                </div>
              </div>
            </section>

            <section
              className="overflow-hidden rounded-[20px] border border-border bg-white shadow-sm"
              aria-labelledby="inbox-preview-heading"
            >
              <div className="flex items-center justify-between border-b border-border bg-muted/30 px-5 py-4">
                <h2 id="inbox-preview-heading" className="text-sm font-medium">
                  Inbox preview
                </h2>
                <Mail size={16} className="text-muted-foreground" />
              </div>
              <div className="p-5">
                <div className="mb-4 flex items-center gap-3">
                  <span className="flex size-10 shrink-0 items-center justify-center rounded-full bg-primary/10 text-primary">
                    <Mail size={18} />
                  </span>
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-xs font-medium">
                      {selectedSender?.fromEmail || "Your sender address"}
                    </p>
                    <p className="mt-1 text-[11px] text-muted-foreground">
                      to {selectedList?.name || "your audience"}
                    </p>
                  </div>
                </div>
                <p
                  className={cn(
                    "break-words text-base font-semibold leading-relaxed tracking-tight",
                    !draft.subject && "text-muted-foreground",
                  )}
                >
                  {draft.subject || "Your subject line goes here"}
                </p>
                <div className="mt-5 flex items-center gap-3 rounded-xl border border-dashed border-border bg-muted/20 p-3">
                  <LayoutTemplate
                    size={20}
                    className="shrink-0 text-muted-foreground"
                  />
                  <div className="min-w-0">
                    <p className="truncate text-xs font-medium">
                      {selectedTemplate?.name || "No template selected"}
                    </p>
                    <p className="mt-1 text-[11px] text-muted-foreground">
                      {selectedTemplate
                        ? "Your saved email design"
                        : "Choose a design from your library"}
                    </p>
                  </div>
                </div>
              </div>
              <div className="border-t border-border px-5 py-3 text-[11px] text-muted-foreground">
                A preview of your sender and subject line.
              </div>
            </section>
            <div className="flex items-center gap-2 px-1 text-xs text-muted-foreground">
              <Clock3 size={14} />
              {isScheduled && draft.scheduledFor
                ? `Planned for ${format(draft.scheduledFor, "MMM d, h:mm a")}`
                : "Saved as draft. Nothing sends yet."}
            </div>
          </aside>
        </div>
        <div className="mt-7 flex flex-wrap items-center justify-between gap-4 border-t border-border pt-5">
          <p className="text-xs text-muted-foreground">
            Your next great connection starts with an email.
          </p>
          <Button type="submit" disabled={saving || !team?.id}>
            {saving ? (
              <Loader2 className="size-4 animate-spin" />
            ) : (
              <Save className="size-4" />
            )}
            {saving ? "Saving…" : "Save draft"}
          </Button>
        </div>
      </form>
    </Form>
  );
}
