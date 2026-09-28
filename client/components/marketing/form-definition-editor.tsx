"use client";

import { ArrowDown, ArrowUp, Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Field } from "./shared";
import type {
  FormCondition,
  FormDefinition,
  FormQuestion,
} from "@/lib/marketing/form-definition";
import {
  fieldIsReferenced,
  renameFormField,
} from "@/lib/marketing/form-starters";

const fieldTypes: { value: FormQuestion["type"]; label: string }[] = [
  { value: "TEXT", label: "Short text" },
  { value: "TEXTAREA", label: "Long text" },
  { value: "EMAIL", label: "Email" },
  { value: "PHONE", label: "Phone" },
  { value: "SELECT", label: "Dropdown" },
  { value: "RADIO", label: "Multiple choice" },
  { value: "CHECKBOX", label: "Checkbox" },
  { value: "NUMBER", label: "Number" },
  { value: "DATE", label: "Date" },
  { value: "HIDDEN", label: "Hidden value" },
];

function ConditionEditor({
  condition,
  fields,
  onChange,
}: {
  condition?: FormCondition;
  fields: FormQuestion[];
  onChange: (condition: FormCondition | undefined) => void;
}) {
  if (!fields.length && !condition) return null;
  const source = fields.find((field) => field.key === condition?.field);
  return (
    <div className="space-y-3 rounded-md bg-muted/40 p-3">
      <Field label="Show when">
        <select
          value={condition?.field || ""}
          onChange={(event) =>
            onChange(
              event.target.value
                ? { field: event.target.value, operator: "equals", value: "" }
                : undefined,
            )
          }
        >
          <option value="">Always visible</option>
          {fields.map((field, index) => (
            <option key={`${field.key}-${index}`} value={field.key}>
              {field.label || field.key}
            </option>
          ))}
          {condition && !source && (
            <option value={condition.field}>
              Missing question: {condition.field}
            </option>
          )}
        </select>
      </Field>
      {condition && (
        <div className="grid gap-3 sm:grid-cols-2">
          <Field label="Condition">
            <select
              value={condition.operator}
              onChange={(event) =>
                onChange({
                  ...condition,
                  operator: event.target.value as FormCondition["operator"],
                })
              }
            >
              <option value="equals">Equals</option>
              <option value="not_equals">Does not equal</option>
              <option value="contains">Contains</option>
              <option value="is_set">Has an answer</option>
            </select>
          </Field>
          {condition.operator !== "is_set" && (
            <Field label="Answer">
              {source?.options?.length ? (
                <select
                  value={condition.value || ""}
                  onChange={(event) =>
                    onChange({ ...condition, value: event.target.value })
                  }
                >
                  <option value="">Choose an answer</option>
                  {source.options.map((option, index) => (
                    <option key={index} value={option}>
                      {option}
                    </option>
                  ))}
                </select>
              ) : source?.type === "CHECKBOX" ? (
                <select
                  value={condition.value || ""}
                  onChange={(event) =>
                    onChange({ ...condition, value: event.target.value })
                  }
                >
                  <option value="">Choose an answer</option>
                  <option value="true">Checked</option>
                  <option value="false">Unchecked</option>
                </select>
              ) : (
                <input
                  maxLength={2000}
                  value={condition.value || ""}
                  onChange={(event) =>
                    onChange({ ...condition, value: event.target.value })
                  }
                />
              )}
            </Field>
          )}
        </div>
      )}
    </div>
  );
}

export function FormDefinitionEditor({
  definition,
  onChange,
}: {
  definition: FormDefinition;
  onChange: (definition: FormDefinition) => void;
}) {
  const updatePage = (
    pageIndex: number,
    patch: Partial<FormDefinition["pages"][number]>,
  ) =>
    onChange({
      ...definition,
      pages: definition.pages.map((page, index) =>
        index === pageIndex ? { ...page, ...patch } : page,
      ),
    });
  const updateField = (
    pageIndex: number,
    fieldIndex: number,
    patch: Partial<FormQuestion>,
  ) =>
    updatePage(pageIndex, {
      fields: definition.pages[pageIndex].fields.map((field, index) =>
        index === fieldIndex ? { ...field, ...patch } : field,
      ),
    });
  const allFields = definition.pages.flatMap((page) => page.fields);
  const nextKey = () => {
    let count = allFields.length + 1;
    while (allFields.some((field) => field.key === `question_${count}`))
      count++;
    return `question_${count}`;
  };
  const moveField = (
    pageIndex: number,
    fieldIndex: number,
    direction: number,
  ) => {
    const fields = [...definition.pages[pageIndex].fields];
    [fields[fieldIndex], fields[fieldIndex + direction]] = [
      fields[fieldIndex + direction],
      fields[fieldIndex],
    ];
    updatePage(pageIndex, { fields });
  };
  return (
    <div className="space-y-5">
      <div>
        <h3 className="font-medium">Questions & steps</h3>
        <p className="mt-1 text-xs text-muted-foreground">
          Conditions can use answers from earlier questions. Keep the required
          email question visible to everyone.
        </p>
      </div>
      {definition.pages.map((page, pageIndex) => {
        const previousFields = definition.pages
          .slice(0, pageIndex)
          .flatMap((item) => item.fields);
        const pageRemovable =
          definition.pages.length > 1 &&
          page.fields.every(
            (field) =>
              field.key !== "email" &&
              !fieldIsReferenced(definition, field.key),
          );
        return (
          <fieldset
            key={page.id}
            className="min-w-0 space-y-4 rounded-xl border p-4"
          >
            <legend className="px-1 text-sm font-semibold">
              Step {pageIndex + 1}
            </legend>
            <div className="flex items-end gap-2">
              <div className="min-w-0 flex-1">
                <Field label="Step title">
                  <input
                    required
                    maxLength={120}
                    value={page.title}
                    onChange={(event) =>
                      updatePage(pageIndex, { title: event.target.value })
                    }
                  />
                </Field>
              </div>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                disabled={!pageRemovable}
                aria-label={`Remove step ${pageIndex + 1}`}
                title={
                  pageRemovable
                    ? "Remove this step"
                    : "Keep the email question and remove dependent conditions first"
                }
                onClick={() =>
                  onChange({
                    ...definition,
                    pages: definition.pages.filter(
                      (_, index) => index !== pageIndex,
                    ),
                  })
                }
              >
                <Trash2 size={16} />
              </Button>
            </div>
            <Field label="Step description (optional)">
              <textarea
                rows={2}
                maxLength={500}
                value={page.description || ""}
                onChange={(event) =>
                  updatePage(pageIndex, { description: event.target.value })
                }
              />
            </Field>
            {!page.fields.some((field) => field.key === "email") && (
              <ConditionEditor
                fields={previousFields}
                condition={page.condition}
                onChange={(condition) => updatePage(pageIndex, { condition })}
              />
            )}
            {page.fields.map((field, fieldIndex) => {
              const isEmail = field.key === "email";
              const referenced = fieldIsReferenced(definition, field.key);
              const prior = [
                ...previousFields,
                ...page.fields.slice(0, fieldIndex),
              ];
              return (
                <div
                  key={fieldIndex}
                  className="space-y-3 rounded-lg border bg-background p-3"
                >
                  <div className="flex items-center justify-between gap-2">
                    <span className="min-w-0 truncate text-xs font-medium text-muted-foreground">
                      Question {fieldIndex + 1}
                    </span>
                    <div className="flex shrink-0 items-center gap-1">
                      <Button
                        type="button"
                        size="icon"
                        variant="ghost"
                        disabled={fieldIndex === 0}
                        aria-label={`Move ${field.label} up`}
                        onClick={() => moveField(pageIndex, fieldIndex, -1)}
                      >
                        <ArrowUp size={14} />
                      </Button>
                      <Button
                        type="button"
                        size="icon"
                        variant="ghost"
                        disabled={fieldIndex === page.fields.length - 1}
                        aria-label={`Move ${field.label} down`}
                        onClick={() => moveField(pageIndex, fieldIndex, 1)}
                      >
                        <ArrowDown size={14} />
                      </Button>
                      <Button
                        type="button"
                        size="icon"
                        variant="ghost"
                        disabled={isEmail || referenced}
                        title={
                          referenced
                            ? "Remove dependent conditions before deleting this question"
                            : undefined
                        }
                        aria-label={`Remove ${field.label}`}
                        onClick={() =>
                          updatePage(pageIndex, {
                            fields: page.fields.filter(
                              (_, index) => index !== fieldIndex,
                            ),
                          })
                        }
                      >
                        <Trash2 size={14} />
                      </Button>
                    </div>
                  </div>
                  <Field label="Question">
                    <input
                      required
                      maxLength={120}
                      value={field.label}
                      onChange={(event) =>
                        updateField(pageIndex, fieldIndex, {
                          label: event.target.value,
                        })
                      }
                    />
                  </Field>
                  <div className="grid gap-3 sm:grid-cols-2">
                    <Field label="Answer type">
                      <select
                        value={field.type}
                        disabled={isEmail}
                        onChange={(event) => {
                          const type = event.target
                            .value as FormQuestion["type"];
                          updateField(pageIndex, fieldIndex, {
                            type,
                            options: ["SELECT", "RADIO"].includes(type)
                              ? field.options || ["Option 1", "Option 2"]
                              : undefined,
                          });
                        }}
                      >
                        {fieldTypes.map((type) => (
                          <option key={type.value} value={type.value}>
                            {type.label}
                          </option>
                        ))}
                      </select>
                    </Field>
                    <Field
                      label="Field key"
                      hint="Stable identifier used in journeys."
                    >
                      <input
                        required
                        disabled={isEmail}
                        pattern="[a-zA-Z][a-zA-Z0-9_]*"
                        maxLength={64}
                        value={field.key}
                        onChange={(event) =>
                          onChange(
                            renameFormField(
                              definition,
                              field.key,
                              event.target.value,
                              { pageIndex, fieldIndex },
                            ),
                          )
                        }
                      />
                    </Field>
                  </div>
                  {["SELECT", "RADIO"].includes(field.type) && (
                    <Field label="Options" hint="One unique answer per line.">
                      <textarea
                        required
                        rows={3}
                        maxLength={4000}
                        value={(field.options || []).join("\n")}
                        onChange={(event) =>
                          updateField(pageIndex, fieldIndex, {
                            options: event.target.value.split("\n"),
                          })
                        }
                      />
                    </Field>
                  )}
                  {field.type === "HIDDEN" ? (
                    <Field label="Hidden value">
                      <input
                        maxLength={2000}
                        value={field.defaultValue || ""}
                        onChange={(event) =>
                          updateField(pageIndex, fieldIndex, {
                            defaultValue: event.target.value,
                          })
                        }
                      />
                    </Field>
                  ) : (
                    <>
                      {!["CHECKBOX", "RADIO", "DATE"].includes(field.type) && (
                        <Field label="Placeholder (optional)">
                          <input
                            maxLength={200}
                            value={field.placeholder || ""}
                            onChange={(event) =>
                              updateField(pageIndex, fieldIndex, {
                                placeholder: event.target.value,
                              })
                            }
                          />
                        </Field>
                      )}
                      <Field label="Help text (optional)">
                        <input
                          maxLength={500}
                          value={field.helpText || ""}
                          onChange={(event) =>
                            updateField(pageIndex, fieldIndex, {
                              helpText: event.target.value,
                            })
                          }
                        />
                      </Field>
                      <label className="flex items-center gap-2 text-sm">
                        <input
                          type="checkbox"
                          className="!w-4"
                          checked={field.required}
                          disabled={isEmail}
                          onChange={(event) =>
                            updateField(pageIndex, fieldIndex, {
                              required: event.target.checked,
                            })
                          }
                        />
                        Required answer
                      </label>
                    </>
                  )}
                  {!isEmail && (
                    <ConditionEditor
                      fields={prior}
                      condition={field.condition}
                      onChange={(condition) =>
                        updateField(pageIndex, fieldIndex, { condition })
                      }
                    />
                  )}
                </div>
              );
            })}
            <Button
              type="button"
              variant="outline"
              className="w-full"
              disabled={allFields.length >= 50}
              onClick={() =>
                updatePage(pageIndex, {
                  fields: [
                    ...page.fields,
                    {
                      key: nextKey(),
                      label: "New question",
                      type: "TEXT",
                      required: false,
                    },
                  ],
                })
              }
            >
              <Plus size={16} />
              Add question
            </Button>
          </fieldset>
        );
      })}
      <Button
        type="button"
        variant="outline"
        disabled={definition.pages.length >= 12}
        onClick={() =>
          onChange({
            ...definition,
            pages: [
              ...definition.pages,
              {
                id: `step_${crypto.randomUUID().slice(0, 8)}`,
                title: `Step ${definition.pages.length + 1}`,
                fields: [
                  {
                    key: nextKey(),
                    label: "New question",
                    type: "TEXT",
                    required: false,
                  },
                ],
              },
            ],
          })
        }
      >
        <Plus size={16} />
        Add step
      </Button>
    </div>
  );
}
