"use client";

import { Field } from "./shared";
import { Button } from "@/components/ui/button";
import { useMarketingQuery } from "@/lib/marketing/api";
import {
  formPresets,
  resolveFormTheme,
  type FormTheme,
} from "@/lib/marketing/form-theme";
import type { LeadForm } from "@/lib/marketing/types";

export function FormThemeEditor({
  theme,
  onChange,
  forms,
}: {
  theme: FormTheme;
  onChange: (theme: FormTheme) => void;
  forms: LeadForm[];
}) {
  const branding = useMarketingQuery<{ logoUrl?: string }>(
    "marketing/branding",
  );
  const brandedForms = forms.filter((form) => form.theme);
  return (
    <fieldset className="space-y-4 rounded-xl border p-4">
      <legend className="px-1 text-sm font-medium">Appearance</legend>
      {brandedForms.length > 0 && (
        <Field
          label="Reuse a form’s design"
          hint="Copies the saved design into this form; later changes stay independent."
        >
          <select
            value=""
            onChange={(event) => {
              const source = brandedForms.find(
                (form) => form.id === event.target.value,
              );
              if (source) onChange(resolveFormTheme(source.theme));
            }}
          >
            <option value="">Choose an existing form</option>
            {brandedForms.map((form) => (
              <option key={form.id} value={form.id}>
                {form.Name}
              </option>
            ))}
          </select>
        </Field>
      )}
      <Field label="Theme">
        <select
          value={theme.preset}
          onChange={(event) =>
            onChange({
              ...formPresets[event.target.value as FormTheme["preset"]],
              logoUrl: theme.logoUrl,
            })
          }
        >
          <option value="light">Light</option>
          <option value="dark">Dark</option>
          <option value="warm">Warm</option>
        </select>
      </Field>
      <div className="grid grid-cols-2 gap-3">
        {(
          [
            ["backgroundColor", "Page background"],
            ["cardColor", "Form background"],
            ["textColor", "Text color"],
            ["buttonColor", "Button color"],
            ["buttonTextColor", "Button text color"],
          ] as const
        ).map(([key, label]) => (
          <Field key={key} label={label}>
            <input
              type="color"
              value={theme[key]}
              onChange={(event) =>
                onChange({ ...theme, [key]: event.target.value })
              }
            />
          </Field>
        ))}
      </div>
      <Field label="Logo URL (HTTPS)">
        <input
          type="url"
          pattern="https://.*"
          maxLength={2048}
          value={theme.logoUrl}
          placeholder="https://example.com/logo.png"
          onChange={(event) =>
            onChange({ ...theme, logoUrl: event.target.value })
          }
        />
      </Field>
      {branding.data?.logoUrl && (
        <Button
          type="button"
          size="sm"
          variant="outline"
          onClick={() =>
            onChange(
              resolveFormTheme({ ...theme, logoUrl: branding.data?.logoUrl }),
            )
          }
        >
          Use workspace logo
        </Button>
      )}
      <div className="grid grid-cols-2 gap-3">
        <Field label="Font">
          <select
            value={theme.font}
            onChange={(event) =>
              onChange({
                ...theme,
                font: event.target.value as FormTheme["font"],
              })
            }
          >
            <option value="sans">Sans serif</option>
            <option value="serif">Serif</option>
            <option value="mono">Monospace</option>
          </select>
        </Field>
        <Field label="Corners">
          <select
            value={theme.corners}
            onChange={(event) =>
              onChange({
                ...theme,
                corners: event.target.value as FormTheme["corners"],
              })
            }
          >
            <option value="rounded">Rounded</option>
            <option value="square">Square</option>
          </select>
        </Field>
      </div>
    </fieldset>
  );
}
