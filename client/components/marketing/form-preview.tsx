"use client";

import { useId, useState } from "react";
import { Button } from "@/components/ui/button";
import { FormSurface } from "./form-surface";
import { Field } from "./shared";
import {
  getReachablePages,
  type FormDefinition,
  type FormQuestion,
} from "@/lib/marketing/form-definition";
import type { FormTheme } from "@/lib/marketing/form-theme";
import { workspaceClassName } from "@/lib/workspace-styles";

/** A local sandbox: interacting here never stores answers or emits analytics. */
export function FormPreview({
  name,
  description,
  definition,
  button,
  theme,
}: {
  name: string;
  description: string;
  definition: FormDefinition;
  button: string;
  theme?: Partial<FormTheme>;
}) {
  const prefix = useId();
  const [answers, setAnswers] = useState<Record<string, string>>({});
  const [pageIndex, setPageIndex] = useState(0);
  const pages = getReachablePages(definition, answers);
  const index = Math.min(pageIndex, Math.max(0, pages.length - 1));
  const page = pages[index];
  const answer = (field: FormQuestion, value: string) =>
    setAnswers((current) => ({ ...current, [field.key]: value }));
  return (
    <FormSurface theme={theme} preview name={name}>
      <h2>{name || "Let’s stay in touch"}</h2>
      {description && <p>{description}</p>}
      <div className={workspaceClassName("product-form")}>
        {definition.pages.length > 1 && (
          <div className="space-y-2">
            <p className="text-xs">
              Step {index + 1} of {pages.length}
            </p>
            <progress
              className="h-1 w-full"
              max={pages.length}
              value={index + 1}
              aria-label="Form progress"
            />
          </div>
        )}
        {page && (
          <>
            <div>
              <h3 className="font-medium">{page.title}</h3>
              {page.description && (
                <p className="text-sm">{page.description}</p>
              )}
            </div>
            {page.fields
              .filter((field) => field.type !== "HIDDEN")
              .map((field, fieldIndex) => {
                const value = answers[field.key] ?? field.defaultValue ?? "";
                const label = `${field.label}${field.required ? " *" : ""}`;
                if (field.type === "CHECKBOX")
                  return (
                    <label
                      className="flex items-start gap-2 text-sm"
                      key={fieldIndex}
                    >
                      <input
                        className="!w-4"
                        type="checkbox"
                        checked={value === "true"}
                        onChange={(event) =>
                          answer(field, String(event.target.checked))
                        }
                      />
                      <span>
                        {label}
                        {field.helpText && (
                          <small className="block">{field.helpText}</small>
                        )}
                      </span>
                    </label>
                  );
                if (field.type === "RADIO")
                  return (
                    <fieldset className="space-y-2" key={fieldIndex}>
                      <legend className="mb-2 text-sm">{label}</legend>
                      {field.options?.map((option, optionIndex) => (
                        <label
                          className="flex items-center gap-2 text-sm"
                          key={optionIndex}
                        >
                          <input
                            className="!w-4"
                            type="radio"
                            name={`${prefix}-${field.key}`}
                            checked={value === option}
                            onChange={() => answer(field, option)}
                          />
                          {option}
                        </label>
                      ))}
                      {field.helpText && <small>{field.helpText}</small>}
                    </fieldset>
                  );
                return (
                  <Field key={fieldIndex} label={label} hint={field.helpText}>
                    {field.type === "TEXTAREA" ? (
                      <textarea
                        rows={3}
                        value={value}
                        placeholder={field.placeholder}
                        onChange={(event) => answer(field, event.target.value)}
                      />
                    ) : field.type === "SELECT" ? (
                      <select
                        value={value}
                        onChange={(event) => answer(field, event.target.value)}
                      >
                        <option value="">
                          {field.placeholder || "Choose an answer"}
                        </option>
                        {field.options?.map((option, optionIndex) => (
                          <option key={optionIndex} value={option}>
                            {option}
                          </option>
                        ))}
                      </select>
                    ) : (
                      <input
                        type={
                          field.type === "EMAIL"
                            ? "email"
                            : field.type === "PHONE"
                              ? "tel"
                              : field.type === "NUMBER"
                                ? "number"
                                : field.type === "DATE"
                                  ? "date"
                                  : "text"
                        }
                        value={value}
                        placeholder={
                          field.placeholder ||
                          (field.type === "EMAIL" ? "you@example.com" : "")
                        }
                        onChange={(event) => answer(field, event.target.value)}
                      />
                    )}
                  </Field>
                );
              })}
          </>
        )}
        {index === pages.length - 1 && definition.consent.mode !== "none" && (
          <label className="flex items-start gap-2 text-sm">
            <input type="checkbox" className="!w-4" />
            <span>
              {definition.consent.label}
              {definition.consent.mode === "optional" && " (optional)"}
            </span>
          </label>
        )}
        <div className="flex flex-wrap gap-2">
          {index > 0 && (
            <Button
              type="button"
              variant="outline"
              onClick={() => setPageIndex(index - 1)}
            >
              Back
            </Button>
          )}
          <Button
            type="button"
            data-form-submit
            disabled={index === pages.length - 1}
            onClick={() => setPageIndex(index + 1)}
          >
            {index < pages.length - 1 ? "Continue" : button || "Submit"}
          </Button>
        </div>
        <p className="text-xs">Preview only. Answers are not saved.</p>
      </div>
    </FormSurface>
  );
}
