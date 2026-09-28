"use client";

import { useId, useState, type FormEvent } from "react";
import { Sparkles } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { useMarketing } from "@/lib/marketing/api";
import {
  getFormDefinition,
  validateFormDefinition,
  type FormDefinition,
} from "@/lib/marketing/form-definition";
import {
  copyFormDefinition,
  formStarters,
  type FormStarter,
} from "@/lib/marketing/form-starters";
import { resolveFormTheme } from "@/lib/marketing/form-theme";
import type { LeadForm, Options } from "@/lib/marketing/types";
import { workspaceClassName } from "@/lib/workspace-styles";
import { Modal, Field } from "./shared";
import { FormDefinitionEditor } from "./form-definition-editor";
import { FormPreview } from "./form-preview";
import { FormThemeEditor } from "./form-theme-editor";

type GeneratedDraft = {
  name: string;
  description: string;
  buttonText: string;
  successMessage: string;
  definition: FormDefinition;
};

export function FormEditor({
  open,
  onOpenChange,
  form,
  options,
  forms,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  form: LeadForm | null;
  options?: Options;
  forms: LeadForm[];
}) {
  const formId = useId();
  const { request, refresh } = useMarketing();
  const [name, setName] = useState(form?.Name || "");
  const [description, setDescription] = useState(form?.description || "");
  const [listId, setListId] = useState(form?.AddToListID || "");
  const [theme, setTheme] = useState(() => resolveFormTheme(form?.theme));
  const [definition, setDefinition] = useState<FormDefinition>(() =>
    form
      ? copyFormDefinition(getFormDefinition(form))
      : copyFormDefinition(formStarters[0].definition),
  );
  const [button, setButton] = useState(form?.SubmitButtonText || "Subscribe");
  const [success, setSuccess] = useState(
    form?.successMessage || formStarters[0].success,
  );
  const [status, setStatus] = useState(form?.Status || "DRAFT");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [starterKey, setStarterKey] = useState("newsletter");
  const [instruction, setInstruction] = useState("");
  const [generating, setGenerating] = useState(false);
  const [generated, setGenerated] = useState<GeneratedDraft | null>(null);
  const [generationError, setGenerationError] = useState("");
  const [previewRevision, setPreviewRevision] = useState(0);
  const [nextStep, setNextStep] = useState("");
  const validationErrors = validateFormDefinition(definition);
  const applyStarter = (starter: FormStarter) => {
    setName(starter.name);
    setDescription(starter.description);
    setButton(starter.button);
    setSuccess(starter.success);
    setDefinition(copyFormDefinition(starter.definition));
    setStatus("DRAFT");
    setNextStep(starter.nextStep);
    setPreviewRevision((value) => value + 1);
  };
  const generate = async () => {
    setGenerating(true);
    setGenerationError("");
    setGenerated(null);
    try {
      const draft = await request<GeneratedDraft>(
        "marketing/form-draft",
        "POST",
        { instruction },
      );
      const errors = validateFormDefinition(draft.definition);
      if (errors.length)
        throw new Error(`The draft needs revision: ${errors[0]}`);
      setGenerated(draft);
    } catch (error) {
      setGenerationError((error as Error).message);
    } finally {
      setGenerating(false);
    }
  };
  const save = async (event: FormEvent) => {
    event.preventDefault();
    if (busy) return;
    setError("");
    if (validationErrors.length) {
      setError(validationErrors.join(" "));
      return;
    }
    setBusy(true);
    try {
      await request(
        `marketing/forms${form ? `/${form.id}` : ""}`,
        form ? "PUT" : "POST",
        {
          name,
          description,
          listId,
          status,
          theme,
          definition,
          successMessage: success,
          buttonText: button,
          fields: definition.pages
            .flatMap((page) => page.fields)
            .map((field) => ({
              label: field.label,
              type: field.type,
              required: field.required,
              key: field.key,
            })),
        },
      );
      await refresh();
      onOpenChange(false);
      toast.success(
        status === "PUBLISHED" ? "Your form is live" : "Form saved",
      );
    } catch (error) {
      setError((error as Error).message);
    } finally {
      setBusy(false);
    }
  };
  return (
    <Modal
      open={open}
      onOpenChange={(value) => {
        if (!busy && !generating) onOpenChange(value);
      }}
      title={form ? "Edit form" : "Create a new form"}
      description="Build a thoughtful signup, request, or conversation. Preview every step before publishing."
      wide
    >
      <div className="grid min-w-0 gap-8 lg:grid-cols-2">
        <form
          id={formId}
          onSubmit={save}
          className={workspaceClassName("product-form min-w-0")}
        >
          <fieldset disabled={busy} className="min-w-0 space-y-5">
            <details className="rounded-xl border p-4">
              <summary className="cursor-pointer text-sm font-medium">
                Start from a template or describe your form
              </summary>
              <div className="mt-4 space-y-4">
                <Field label="Form starter">
                  <select
                    value={starterKey}
                    onChange={(event) => setStarterKey(event.target.value)}
                  >
                    {formStarters.map((starter) => (
                      <option key={starter.key} value={starter.key}>
                        {starter.name}
                      </option>
                    ))}
                  </select>
                </Field>
                <p className="text-xs text-muted-foreground">
                  {
                    formStarters.find((starter) => starter.key === starterKey)
                      ?.description
                  }
                </p>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() =>
                    applyStarter(
                      formStarters.find(
                        (starter) => starter.key === starterKey,
                      )!,
                    )
                  }
                >
                  Replace questions with this starter
                </Button>
                <Field
                  label="Describe your form"
                  hint="AI creates a draft for you to review. Nothing is published."
                >
                  <textarea
                    rows={3}
                    maxLength={2000}
                    value={instruction}
                    onChange={(event) => setInstruction(event.target.value)}
                    placeholder="A two-step waitlist for a new email product, asking about company size and interests"
                  />
                </Field>
                <Button
                  type="button"
                  variant="outline"
                  disabled={generating || instruction.trim().length < 10}
                  onClick={() => void generate()}
                >
                  <Sparkles size={16} />
                  {generating ? "Drafting…" : "Generate a draft"}
                </Button>
                {generationError && (
                  <p role="alert" className="text-sm text-destructive">
                    {generationError}
                  </p>
                )}
                {generated && (
                  <div className="space-y-3 rounded-lg bg-muted/40 p-3">
                    <h4 className="font-medium">{generated.name}</h4>
                    <p className="text-sm">{generated.description}</p>
                    <ul className="list-disc pl-5 text-xs">
                      {generated.definition.pages.map((page) => (
                        <li key={page.id}>
                          {page.title}:{" "}
                          {page.fields.map((field) => field.label).join(", ")}
                        </li>
                      ))}
                    </ul>
                    <p className="text-xs text-muted-foreground">
                      Using this draft replaces the current questions and copy,
                      and sets the form to draft. Review consent and follow-up
                      messages before publishing.
                    </p>
                    <Button
                      type="button"
                      size="sm"
                      onClick={() => {
                        setName(generated.name);
                        setDescription(generated.description);
                        setButton(generated.buttonText);
                        setSuccess(generated.successMessage);
                        setDefinition(copyFormDefinition(generated.definition));
                        setStatus("DRAFT");
                        setGenerated(null);
                        setNextStep(
                          "Save this form, then open Email journey to review and connect follow-up emails.",
                        );
                        setPreviewRevision((value) => value + 1);
                      }}
                    >
                      Use this draft
                    </Button>
                  </div>
                )}
              </div>
            </details>
            {nextStep && (
              <p className="rounded-lg bg-muted/40 p-3 text-xs" role="status">
                {nextStep}
              </p>
            )}
            <Field label="Form name">
              <input
                required
                maxLength={120}
                placeholder="e.g. The weekly newsletter"
                value={name}
                onChange={(event) => setName(event.target.value)}
              />
            </Field>
            <Field label="Description">
              <textarea
                rows={2}
                maxLength={500}
                value={description}
                onChange={(event) => setDescription(event.target.value)}
              />
            </Field>
            <Field
              label="Audience"
              hint="Only visitors who explicitly consent are added as subscribers."
            >
              <select
                required
                value={listId}
                onChange={(event) => setListId(event.target.value)}
              >
                <option value="">Choose an audience</option>
                {options?.lists.map((list) => (
                  <option key={list.id} value={list.id}>
                    {list.name}
                  </option>
                ))}
              </select>
            </Field>
            <FormDefinitionEditor
              definition={definition}
              onChange={setDefinition}
            />
            <fieldset className="space-y-4 rounded-xl border p-4">
              <legend className="px-1 text-sm font-medium">
                Consent & completion
              </legend>
              <Field label="Marketing consent">
                <select
                  value={definition.consent.mode}
                  onChange={(event) =>
                    setDefinition({
                      ...definition,
                      consent: {
                        mode: event.target
                          .value as FormDefinition["consent"]["mode"],
                        label:
                          definition.consent.label ||
                          "I agree to receive emails and understand I can unsubscribe at any time.",
                      },
                    })
                  }
                >
                  <option value="required">Required to submit</option>
                  <option value="optional">Optional subscription</option>
                  <option value="none">No marketing subscription</option>
                </select>
              </Field>
              {definition.consent.mode !== "none" && (
                <Field label="Consent label">
                  <textarea
                    required
                    rows={2}
                    maxLength={500}
                    value={definition.consent.label}
                    onChange={(event) =>
                      setDefinition({
                        ...definition,
                        consent: {
                          ...definition.consent,
                          label: event.target.value,
                        },
                      })
                    }
                  />
                </Field>
              )}
              <p className="text-xs text-muted-foreground">
                Optional and no-subscription forms record submissions without
                silently subscribing people. Previously unsubscribed or
                suppressed contacts stay protected.
              </p>
              <label className="flex items-start gap-2 text-sm">
                <input
                  className="!w-4"
                  type="checkbox"
                  checked={definition.saveProgress}
                  onChange={(event) =>
                    setDefinition({
                      ...definition,
                      saveProgress: event.target.checked,
                    })
                  }
                />
                <span>
                  Let visitors explicitly save and resume later
                  <small className="mt-1 block text-muted-foreground">
                    Answers are saved only after the visitor chooses to save
                    progress.
                  </small>
                </span>
              </label>
              <Field label="Button text">
                <input
                  required
                  maxLength={60}
                  value={button}
                  onChange={(event) => setButton(event.target.value)}
                />
              </Field>
              <Field label="Success message">
                <textarea
                  required
                  rows={2}
                  maxLength={500}
                  value={success}
                  onChange={(event) => setSuccess(event.target.value)}
                />
              </Field>
              <Field
                label="Redirect after completion (optional)"
                hint="A full HTTPS URL. Leave empty to display the success message."
              >
                <input
                  type="url"
                  pattern="https://.*"
                  maxLength={2048}
                  placeholder="https://example.com/thank-you"
                  value={definition.successRedirectUrl || ""}
                  onChange={(event) =>
                    setDefinition({
                      ...definition,
                      successRedirectUrl: event.target.value || undefined,
                    })
                  }
                />
              </Field>
            </fieldset>
            <FormThemeEditor theme={theme} onChange={setTheme} forms={forms} />
            <Field label="Publication status">
              <select
                value={status}
                onChange={(event) => setStatus(event.target.value)}
              >
                <option value="DRAFT">Draft</option>
                <option value="PUBLISHED">Published</option>
                <option value="ARCHIVED">Paused</option>
              </select>
            </Field>
          </fieldset>
        </form>
        <aside className="min-w-0 space-y-3 lg:sticky lg:top-0 lg:self-start">
          <div className="flex items-center justify-between gap-2">
            <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              Interactive preview
            </span>
            <Button
              type="button"
              size="sm"
              variant="ghost"
              onClick={() => setPreviewRevision((value) => value + 1)}
            >
              Reset answers
            </Button>
          </div>
          <FormPreview
            key={previewRevision}
            name={name}
            description={description}
            definition={definition}
            button={button}
            theme={theme}
          />
          {validationErrors.length > 0 && (
            <div className="rounded-lg border border-amber-500/30 bg-amber-500/5 p-3 text-sm">
              <h4 className="font-medium">Before saving</h4>
              <ul className="mt-2 list-disc space-y-1 pl-4">
                {validationErrors.map((message, index) => (
                  <li key={index}>{message}</li>
                ))}
              </ul>
            </div>
          )}
        </aside>
      </div>
      {error && (
        <p role="alert" className={workspaceClassName("product-error")}>
          {error}
        </p>
      )}
      <div className={workspaceClassName("modal-actions")}>
        <Button
          type="button"
          variant="outline"
          disabled={busy || generating}
          onClick={() => onOpenChange(false)}
        >
          Cancel
        </Button>
        <Button
          type="submit"
          form={formId}
          disabled={busy || generating}
          className={workspaceClassName("product-primary")}
        >
          {busy
            ? "Saving…"
            : status === "PUBLISHED"
              ? "Save & publish"
              : "Save form"}
        </Button>
      </div>
    </Modal>
  );
}
