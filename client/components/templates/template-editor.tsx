"use client";
import { EmailWriter } from "@/components/ai/email-writer";

import { workspaceClassName } from "@/lib/workspace-styles";
import { useState, useRef, useEffect } from "react";
import { useRouter } from "next/navigation";
import EmailEditor, { EditorRef, EmailEditorProps } from "react-email-editor";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { EmailCategory, EmailTemplate, TemplateEditorProps } from "@/lib";
import { useTeam } from "@/app/providers/team-provider";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Settings, Mail, Send } from "lucide-react";
import { deepCompare } from "@/lib/utils";
import { toast } from "sonner";
import { useApi } from "@/hooks/use-api";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { resourceEntity } from "@/lib/resource-response";
import {
  getTemplateStarter,
  loadStarterDesign,
} from "@/lib/template-starters/registry";
import Link from "next/link";
import { prepareEditorAssets, restoreEditorAssets } from "@/lib/template-starters/editor-assets";
import { QueryState } from "@/components/marketing/shared";

const testEmailSchema = z.object({
  to: z.string().email(),
});

export function TemplateEditor({
  templateId,
  starterKey,
}: TemplateEditorProps) {
  const router = useRouter();
  const starter =
    templateId === "new" ? getTemplateStarter(starterKey) : undefined;
  const emailEditorRef = useRef<EditorRef | null>(null);
  const { team } = useTeam();
  const { apiFetch, session } = useApi();
  const queryClient = useQueryClient();
  const hydrated = useRef("");
  const [isSaving, setIsSaving] = useState(false);
  const [editorReady, setEditorReady] = useState(false);
  const [designLoading, setDesignLoading] = useState(false);
  const [designError, setDesignError] = useState("");
  const [designAttempt, setDesignAttempt] = useState(0);
  const editorAssetSources = useRef(new Map<string, string>());

  const [template, setTemplate] = useState<EmailTemplate>({
    id: "",
    name: starter?.name ?? "",
    subject: starter?.subject ?? "",
    variables: [],
    teamId: team?.id || "",
    html: "",
    design: {},
    createdAt: new Date(),
    updatedAt: new Date(),
    categoryId: "Transactional",
    designJson: "",
  });

  const [activeTab, setActiveTab] = useState(
    starterKey ? "design" : "settings",
  );
  const [isSendingTestEmail, setIsSendingTestEmail] = useState(false);
  const [showTestEmailDialog, setShowTestEmailDialog] = useState(false);
  const categoriesQuery = useQuery({
    queryKey: ["emailCategories", team?.id],
    enabled: !!team?.id && !!session?.accessToken,
    queryFn: async ({ signal }) => {
      const response = await apiFetch("categories", { signal });
      if (!response.ok) throw new Error("Unable to load email categories");
      const payload = await response.json();
      return (payload.data || []) as EmailCategory[];
    },
  });
  const starterQuery = useQuery({
    queryKey: ["template-starter", starterKey, starter?.version],
    enabled: templateId === "new" && !!starterKey,
    queryFn: ({ signal }) => loadStarterDesign(starterKey!, signal),
    staleTime: Infinity,
    retry: 1,
  });
  useEffect(() => {
    if (
      templateId === "new" &&
      starterQuery.data &&
      hydrated.current !== `starter:${starterKey}`
    ) {
      setTemplate((current) => ({
        ...current,
        design: structuredClone(starterQuery.data),
      }));
      hydrated.current = `starter:${starterKey}`;
    }
  }, [starterQuery.data, starterKey, templateId]);
  const templateQuery = useQuery({
    queryKey: ["template", team?.id, templateId],
    enabled: !!team?.id && !!session?.accessToken && templateId !== "new",
    queryFn: async ({ signal }) => {
      const response = await apiFetch(`templates/${templateId}`, { signal });
      if (!response.ok) throw new Error("Unable to load this template");
      const data = resourceEntity<EmailTemplate>(await response.json());
      const design = data.designJson
        ? JSON.parse(Buffer.from(data.designJson, "base64").toString("utf-8"))
        : {};
      return { ...data, design };
    },
  });
  const emailCategories = categoriesQuery.data || [];
  useEffect(() => {
    const key = `${team?.id}:${templateId}`;
    // Hydrate once per record, including cache hits; background refetches must not erase edits.
    if (templateQuery.data && hydrated.current !== key) {
      setTemplate(templateQuery.data);
      hydrated.current = key;
    }
  }, [templateQuery.data, team?.id, templateId]);
  useEffect(() => {
    if (templateId === "new" && categoriesQuery.data?.length) {
      setTemplate((current) =>
        current.categoryId === "Transactional"
          ? {
              ...current,
              categoryId: (
                categoriesQuery.data.find(
                  (category) =>
                    category.name.toLowerCase() ===
                    (starter?.marketing ? "marketing" : "transactional"),
                ) ?? categoriesQuery.data[0]
              ).id,
            }
          : current,
      );
    }
  }, [categoriesQuery.data, templateId, starter]);

  useEffect(() => {
    if (!editorReady || !template.design || !Object.keys(template.design).length) return;
    const controller = new AbortController();
    const editor = emailEditorRef.current?.editor;
    let timeout: ReturnType<typeof setTimeout> | undefined;
    const loaded = () => {
      if (!controller.signal.aborted) { setDesignLoading(false); setDesignError(""); }
      if (timeout) clearTimeout(timeout);
    };
    editor?.addEventListener("design:loaded", loaded);
    setDesignLoading(true); setDesignError("");
    prepareEditorAssets(template.design, controller.signal).then(({ design, sources }) => {
      if (controller.signal.aborted) return;
      editorAssetSources.current = sources;
      timeout = setTimeout(() => {
        if (!controller.signal.aborted) { setDesignLoading(false); setDesignError("The editor did not finish loading this design. Please retry."); }
      }, 20000);
      editor?.loadDesign(design as any);
    }).catch((error) => {
      if (!controller.signal.aborted) { setDesignLoading(false); setDesignError(error.message || "This design could not be opened."); }
    });
    return () => { controller.abort(); if (timeout) clearTimeout(timeout); editor?.removeEventListener("design:loaded"); };
  }, [template.design, editorReady, designAttempt]);

  const onReady: EmailEditorProps["onReady"] = () => {
    setEditorReady(true);
  };

  const templateCategories = emailCategories.map((category) => ({
    label: category.name,
    value: category.id,
  }));

  const selectedCategory = templateCategories.find(
    (category) => category.value === template.categoryId,
  );

  const exportTemplate = () =>
    new Promise<{ html: string; design: any }>((resolve, reject) => {
      const editor = emailEditorRef.current?.editor;
      if (!editor || !editorReady || designLoading || designError)
        return reject(
          new Error("The editor is still loading. Please try again."),
        );
      const timeout = setTimeout(
        () =>
          reject(new Error("The editor did not respond. Please try again.")),
        20000,
      );
      try {
        editor.exportHtml((data) => {
          clearTimeout(timeout);
          resolve(restoreEditorAssets(data, editorAssetSources.current));
        });
      } catch (error) {
        clearTimeout(timeout);
        reject(error);
      }
    });
  const handleSave = async () => {
    if (isSaving) return;
    if (
      !template.name.trim() ||
      !template.subject.trim() ||
      !selectedCategory
    ) {
      toast.error(
        "Enter a template name, subject, and category before saving.",
      );
      return;
    }
    setIsSaving(true);
    try {
      const { html, design } = await exportTemplate();
      const response = await fetch(
        templateId === "new"
          ? "/api/templates"
          : `/api/templates/${templateId}`,
        {
          method: templateId === "new" ? "POST" : "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            ...template,
            teamId: team?.id,
            html,
            designJson: design,
            changed: !deepCompare(template.design, design),
            categoryId: selectedCategory.value,
          }),
        },
      );
      if (!response.ok)
        throw new Error(
          "Unable to save this template. Your edits are still here.",
        );
      await queryClient.invalidateQueries({ queryKey: ["templates"] });
      await queryClient.invalidateQueries({
        queryKey: ["template", team?.id, templateId],
      });
      toast.success("Template saved successfully");
      router.push("/templates");
    } catch (error) {
      toast.error((error as Error).message);
    } finally {
      setIsSaving(false);
    }
  };
  const sendTestEmail = async (email: string) => {
    setIsSendingTestEmail(true);
    try {
      const { html } = await exportTemplate();
      const response = await apiFetch("emails", {
        method: "POST",
        body: JSON.stringify({ to: email, html, test: true }),
      });
      if (!response.ok) throw new Error("Failed to send test email");
      toast.success("Test email sent successfully");
      setShowTestEmailDialog(false);
    } catch (error) {
      toast.error((error as Error).message);
    } finally {
      setIsSendingTestEmail(false);
    }
  };

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm({
    resolver: zodResolver(testEmailSchema),
  });

  const onSubmit = (data: { to: string }) => {
    void sendTestEmail(data.to);
  };

  if (
    templateId === "new" &&
    starterKey &&
    (starterQuery.isPending || starterQuery.error)
  ) {
    return (
      <div className="space-y-4">
        <QueryState
          loading={starterQuery.isPending}
          error={starterQuery.error}
          retry={() => starterQuery.refetch()}
        />
        <Button variant="outline" asChild>
          <Link href="/templates">Back to template library</Link>
        </Button>
      </div>
    );
  }

  if (
    categoriesQuery.isPending ||
    categoriesQuery.error ||
    (templateId !== "new" && (templateQuery.isPending || templateQuery.error))
  )
    return (
      <QueryState
        loading={
          categoriesQuery.isPending ||
          (templateId !== "new" && templateQuery.isPending)
        }
        error={categoriesQuery.error || templateQuery.error}
        retry={() => {
          void categoriesQuery.refetch();
          if (templateId !== "new") void templateQuery.refetch();
        }}
      />
    );

  return (
    <div className={workspaceClassName("template-editor-workspace")}>
      {starter && (
        <div className="mx-4 mt-4 rounded-xl border border-violet-100 bg-violet-50 px-4 py-3 text-sm text-violet-900">
          <strong>{starter.name}</strong> · Your editable copy. Update the
          brand, links, and sample details before saving.
        </div>
      )}
      <div className="flex justify-between items-center p-4">
        <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
          <div className="flex flex-wrap justify-between items-center gap-3">
            <TabsList>
              <TabsTrigger value="settings" className="flex items-center gap-2">
                <Settings className="h-4 w-4" />
                Settings
              </TabsTrigger>
              <TabsTrigger value="design" className="flex items-center gap-2">
                <Mail className="h-4 w-4" />
                Design
              </TabsTrigger>
            </TabsList>
            <div className="flex flex-wrap items-center gap-2">
              <Button
                variant="secondary"
                disabled={!editorReady || designLoading || !!designError || isSendingTestEmail}
                onClick={() => setShowTestEmailDialog(true)}
                className="flex items-center gap-2"
              >
                <Send className="h-4 w-4" />
                {isSendingTestEmail ? "Sending..." : "Send Test"}
              </Button>

              <EmailWriter
                format="design"
                subject={template.subject}
                applyLabel="Use this design in editor"
                onApply={(draft) => {
                  if (!draft.design) throw new Error("No editable design was returned.");
                  setTemplate((current) => ({ ...current, subject: draft.subject, design: structuredClone(draft.design!) }));
                  setActiveTab("design");
                }}
              />
              <Dialog>
                <DialogTrigger asChild>
                  <Button>Save Template</Button>
                </DialogTrigger>
                <DialogContent className="max-w-sm">
                  <DialogHeader>
                    <DialogTitle>Save Template</DialogTitle>
                    <DialogDescription>
                      {templateId === "new"
                        ? "Save your design as a new template for campaigns, newsletters, and automations."
                        : "Save your changes to this template."}
                    </DialogDescription>
                  </DialogHeader>
                  <Button
                    onClick={handleSave}
                    disabled={isSaving || !editorReady || designLoading || !!designError}
                  >
                    {isSaving ? "Saving…" : "Save Template"}
                  </Button>
                </DialogContent>
              </Dialog>
            </div>
          </div>

          {
            <Dialog
              open={showTestEmailDialog}
              onOpenChange={setShowTestEmailDialog}
            >
              <DialogContent className="max-w-sm">
                <DialogHeader>
                  <DialogTitle>Send Test Email</DialogTitle>
                </DialogHeader>
                <form
                  className="flex flex-col space-y-4"
                  onSubmit={handleSubmit(onSubmit)}
                >
                  <Input
                    name="to"
                    required
                    type="email"
                    placeholder="Email"
                    {...register("to")}
                  />
                  {errors.to && (
                    <span className="text-red-500">
                      {errors.to.message as string}
                    </span>
                  )}
                  {isSendingTestEmail ? (
                    <Button type="submit" disabled>
                      Sending...
                    </Button>
                  ) : (
                    <Button type="submit">Send</Button>
                  )}
                </form>
              </DialogContent>
            </Dialog>
          }

          <div className="flex-1 overflow-hidden">
            <TabsContent value="settings" className="h-full">
              <div className="max-w-2xl mx-auto p-6 space-y-6">
                <div className="space-y-2">
                  <Label htmlFor="name">Template Name</Label>
                  <Input
                    id="name"
                    value={template.name}
                    onChange={(e) =>
                      setTemplate({ ...template, name: e.target.value })
                    }
                    placeholder="Enter template name"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="subject">Subject</Label>
                  <Input
                    id="subject"
                    value={template.subject}
                    onChange={(e) =>
                      setTemplate({ ...template, subject: e.target.value })
                    }
                    placeholder="Enter email subject"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="emailCategory">Email Category</Label>
                  <Select
                    value={selectedCategory?.value}
                    onValueChange={(value) =>
                      setTemplate({ ...template, categoryId: value })
                    }
                    defaultValue={selectedCategory?.value}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select category" />
                    </SelectTrigger>
                    <SelectContent>
                      {templateCategories.map((category) => (
                        <SelectItem key={category.value} value={category.value}>
                          {category.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </TabsContent>

            <TabsContent
              value="design"
              forceMount
              className="h-full data-[state=inactive]:hidden"
            >
              <div className="h-full">
                {designLoading && <p role="status" className="py-3 text-sm text-muted-foreground">Loading your design and images…</p>}
                {designError && <div role="alert" className="flex items-center justify-between gap-3 rounded-lg bg-red-50 p-3 text-sm text-red-700"><span>{designError}</span><Button variant="outline" onClick={() => setDesignAttempt(n => n + 1)}>Retry design</Button></div>}
                <EmailEditor
                  ref={emailEditorRef}
                  onReady={onReady}
                  options={{
                    locale: "en",
                  }}
                  minHeight="calc(100vh - 120px)"
                />
              </div>
            </TabsContent>
          </div>
        </Tabs>
      </div>
    </div>
  );
}
