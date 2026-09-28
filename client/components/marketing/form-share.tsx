"use client";

import { useState } from "react";
import { Copy } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { formHTMLSnippet, publicFormAction } from "@/lib/marketing/form-theme";
import type { LeadForm } from "@/lib/marketing/types";
import { getFormDefinition } from "@/lib/marketing/form-definition";
import { Field, Modal } from "./shared";

const escapeAttribute = (value: string) =>
  value.replace(
    /[&<>"']/g,
    (character) =>
      ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[
        character
      ]!,
  );

export function FormShare({
  form,
  close,
}: {
  form: LeadForm;
  close: () => void;
}) {
  const [mode, setMode] = useState("inline");
  const [source, setSource] = useState("");
  const [campaign, setCampaign] = useState("");
  const [label, setLabel] = useState("Open form");
  const origin = typeof window !== "undefined" ? window.location.origin : "";
  const path = `${origin}/f/${encodeURIComponent(form.Slug)}`;
  const params = new URLSearchParams();
  if (source.trim()) params.set("utm_source", source.trim());
  if (campaign.trim()) params.set("utm_campaign", campaign.trim());
  const url = `${path}${params.size ? `?${params.toString()}` : ""}`;
  const definition = getFormDefinition(form);
  const simplePage = definition.pages.length === 1 && !definition.pages[0].condition
    && definition.pages[0].fields.every((field) => !field.condition && !field.defaultValue && ["TEXT", "EMAIL", "PHONE", "TEXTAREA", "NUMBER", "DATE"].includes(field.type));
  const nativeFields = definition.pages.flatMap((page) => page.fields).map((field) => ({ Label: field.label, FieldType: field.type, Required: field.required, mapToContactField: field.key }));
  const action = origin
    ? new URL(publicFormAction(form.Slug), origin).href
    : publicFormAction(form.Slug);
  const iframe = `<iframe src="${escapeAttribute(url)}" title="${escapeAttribute(form.Name)}" width="100%" height="720" style="border:0" loading="lazy"></iframe>`;
  const script = `<script src="${escapeAttribute(origin)}/forms/embed.js" defer></script>\n<div data-xem-form="${escapeAttribute(url)}" data-xem-mode="${mode}" data-xem-label="${escapeAttribute(label)}" data-xem-title="${escapeAttribute(form.Name)}"></div>`;
  const copy = async (value: string) => {
    try {
      await navigator.clipboard.writeText(value);
      toast.success("Copied to clipboard");
    } catch {
      toast.error("Could not copy. Select and copy the text instead.");
    }
  };
  return (
    <Modal
      open
      onOpenChange={close}
      title={`Share ${form.Name}`}
      description="Share a hosted link, or bring the same form to your own site."
    >
      <div className="space-y-5">
        {form.Status !== "PUBLISHED" && (
          <p role="status" className="rounded-lg bg-amber-500/10 p-3 text-sm">
            This form is {form.Status === "DRAFT" ? "a draft" : "paused"}.
            Publish it before sharing with visitors.
          </p>
        )}
        <div className="grid gap-3 sm:grid-cols-2">
          <Field label="Campaign source (optional)">
            <input
              maxLength={200}
              value={source}
              placeholder="newsletter"
              onChange={(event) => setSource(event.target.value)}
            />
          </Field>
          <Field label="Campaign name (optional)">
            <input
              maxLength={200}
              value={campaign}
              placeholder="product-launch"
              onChange={(event) => setCampaign(event.target.value)}
            />
          </Field>
        </div>
        <Field label="Hosted form link">
          <input
            readOnly
            value={url}
            onFocus={(event) => event.target.select()}
          />
        </Field>
        <Button type="button" variant="outline" onClick={() => void copy(url)}>
          <Copy size={16} />
          Copy link
        </Button>
        <Field label="Embed placement">
          <select
            value={mode}
            onChange={(event) => setMode(event.target.value)}
          >
            <option value="inline">Inline iframe</option>
            <option value="sheet">Inset side sheet</option>
          </select>
        </Field>
        {mode !== "inline" && (
          <Field label="Launch button label">
            <input
              maxLength={80}
              value={label}
              onChange={(event) => setLabel(event.target.value)}
            />
          </Field>
        )}
        <Field label="Embed code">
          <textarea
            rows={5}
            readOnly
            value={mode === "inline" ? iframe : script}
            onFocus={(event) => event.target.select()}
          />
        </Field>
        <Button
          type="button"
          variant="outline"
          onClick={() => void copy(mode === "inline" ? iframe : script)}
        >
          <Copy size={16} />
          Copy embed
        </Button>
        <details className="space-y-4 rounded-lg border p-4">
          <summary className="cursor-pointer text-sm font-medium">
            Use your own HTML form
          </summary>
          <Field label="Form action URL">
            <input
              readOnly
              value={action}
              onFocus={(event) => event.target.select()}
            />
          </Field>
          <p className="text-xs text-muted-foreground">
            Use method="post" with the form’s field keys, version, and configured consent
            value. No API key is needed. Multi-step forms and conditional
            questions work best with the hosted form or embed.
          </p>
          <Button
            type="button"
            size="sm"
            variant="outline"
            onClick={() => void copy(action)}
          >
            Copy action URL
          </Button>
          {simplePage && (
            <>
              <Field label="Custom HTML form">
                <textarea
                  rows={10}
                  readOnly
                  value={formHTMLSnippet(
                    action,
                    nativeFields,
                    form.SubmitButtonText,
                    definition.consent,
                    form.version,
                  )}
                />
              </Field>
              <Button
                type="button"
                size="sm"
                variant="outline"
                onClick={() =>
                  void copy(
                    formHTMLSnippet(action, nativeFields, form.SubmitButtonText, definition.consent, form.version),
                  )
                }
              >
                Copy HTML example
              </Button>
            </>
          )}
        </details>
      </div>
    </Modal>
  );
}
