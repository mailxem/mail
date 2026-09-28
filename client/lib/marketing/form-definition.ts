import type { FormField } from "./types";

export const FORM_FIELD_TYPES = [
  "TEXT",
  "EMAIL",
  "PHONE",
  "TEXTAREA",
  "SELECT",
  "RADIO",
  "CHECKBOX",
  "NUMBER",
  "DATE",
  "HIDDEN",
] as const;
export type FormFieldType = (typeof FORM_FIELD_TYPES)[number];
export type FormAnswers = Record<string, string>;
export type FormCondition = {
  field: string;
  operator: "equals" | "not_equals" | "contains" | "is_set";
  value?: string;
};
export type FormQuestion = {
  key: string;
  label: string;
  type: FormFieldType;
  required: boolean;
  placeholder?: string;
  helpText?: string;
  options?: string[];
  condition?: FormCondition;
  defaultValue?: string;
};
export type FormPage = {
  id: string;
  title: string;
  description?: string;
  condition?: FormCondition;
  fields: FormQuestion[];
};
export type FormDefinition = {
  schemaVersion: 1;
  pages: FormPage[];
  consent: { mode: "required" | "optional" | "none"; label: string };
  saveProgress: boolean;
  successRedirectUrl?: string;
};
export const DEFAULT_CONSENT_LABEL =
  "I agree to receive emails and understand that I can unsubscribe at any time.";
const reservedKeys = new Set([
  "__proto__",
  "prototype",
  "constructor",
  "consent",
  "website",
  "requestId",
  "version",
  "sessionId",
  "resumeToken",
  "attribution",
  "fields",
  "pageId",
  "token",
]);
export const isValidFormFieldKey = (key: string) =>
  /^[a-zA-Z][a-zA-Z0-9_]{0,63}$/.test(key) && !reservedKeys.has(key);

/** Adapt old forms without changing their required marketing consent. */
export function getFormDefinition(form: {
  definition?: FormDefinition | null;
  fields?: FormField[];
  consentText?: string;
}): FormDefinition {
  if (
    form.definition?.schemaVersion === 1 &&
    Array.isArray(form.definition.pages)
  )
    return form.definition;
  return {
    schemaVersion: 1,
    pages: [
      {
        id: "details",
        title: "Your details",
        fields: (form.fields || []).map((field) => ({
          key: field.mapToContactField,
          label: field.Label,
          type: FORM_FIELD_TYPES.includes(field.FieldType as FormFieldType)
            ? (field.FieldType as FormFieldType)
            : "TEXT",
          required: field.Required,
        })),
      },
    ],
    consent: {
      mode: "required",
      label: form.consentText || DEFAULT_CONSENT_LABEL,
    },
    saveProgress: false,
  };
}

export function evaluateCondition(
  condition: FormCondition | undefined,
  answers: FormAnswers,
): boolean {
  if (!condition) return true;
  // A skipped question cannot make a later question reachable, even with not_equals.
  if (!Object.prototype.hasOwnProperty.call(answers, condition.field))
    return false;
  const value = answers[condition.field] || "";
  switch (condition.operator) {
    case "equals":
      return value === (condition.value || "");
    case "not_equals":
      return value !== (condition.value || "");
    case "contains":
      return value.includes(condition.value || "");
    case "is_set":
      return value.trim() !== "" && value !== "false";
    default:
      return false;
  }
}

function effectiveValue(field: FormQuestion, answers: FormAnswers): string {
  if (field.type === "HIDDEN") return field.defaultValue || "";
  const value = (answers[field.key] ?? field.defaultValue ?? "").trim();
  return field.type === "EMAIL" ? value.toLowerCase() : value;
}

/** Walk in order so stale answers from skipped branches cannot activate downstream branches. */
export function getReachablePages(
  definition: FormDefinition,
  answers: FormAnswers,
): FormPage[] {
  const effective: FormAnswers = Object.create(null);
  const pages: FormPage[] = [];
  for (const page of definition.pages) {
    if (!evaluateCondition(page.condition, effective)) continue;
    const fields: FormQuestion[] = [];
    for (const field of page.fields) {
      if (
        !isValidFormFieldKey(field.key) ||
        !evaluateCondition(field.condition, effective)
      )
        continue;
      fields.push(field);
      effective[field.key] = effectiveValue(field, answers);
    }
    if (fields.length) pages.push({ ...page, fields });
  }
  return pages;
}

export function getReachableAnswers(
  definition: FormDefinition,
  answers: FormAnswers,
): FormAnswers {
  return Object.fromEntries(
    getReachablePages(definition, answers).flatMap((page) =>
      page.fields.map((field) => [field.key, effectiveValue(field, answers)]),
    ),
  );
}

export function validateFormPage(
  page: FormPage,
  answers: FormAnswers,
): Record<string, string> {
  const errors: Record<string, string> = {};
  for (const field of page.fields) {
    const value = effectiveValue(field, answers);
    if (
      field.required &&
      (!value.trim() || (field.type === "CHECKBOX" && value !== "true"))
    )
      errors[field.key] = `${field.label} is required.`;
    else if (value && value.length > 2000)
      errors[field.key] = "Use 2,000 characters or fewer.";
    else if (
      value &&
      field.type === "EMAIL" &&
      !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)
    )
      errors[field.key] = "Enter a valid email address.";
    else if (
      value &&
      field.type === "NUMBER" &&
      (!/^[+-]?(?:\d+\.?\d*|\.\d+)(?:[eE][+-]?\d+)?$/.test(value) ||
        !Number.isFinite(Number(value)))
    )
      errors[field.key] = "Enter a valid number.";
    else if (
      value &&
      field.type === "DATE" &&
      (!/^\d{4}-\d{2}-\d{2}$/.test(value) ||
        Number.isNaN(Date.parse(value)) ||
        new Date(value).toISOString().slice(0, 10) !== value)
    )
      errors[field.key] = "Enter a valid date.";
    else if (
      value &&
      (field.type === "SELECT" || field.type === "RADIO") &&
      !field.options?.includes(value)
    )
      errors[field.key] = "Choose one of the available options.";
    else if (
      value &&
      field.type === "CHECKBOX" &&
      !["true", "false"].includes(value)
    )
      errors[field.key] = "Choose a valid checkbox value.";
  }
  return errors;
}

export function validateFormDefinition(definition: FormDefinition): string[] {
  const errors: string[] = [];
  if (definition.schemaVersion !== 1)
    errors.push("Unsupported form definition version.");
  if (!definition.pages.length || definition.pages.length > 12)
    errors.push("Use between 1 and 12 steps.");
  const seen = new Set<string>();
  const pageIds = new Set<string>();
  let email = false;
  let fieldCount = 0;
  const checkCondition = (condition: FormCondition | undefined) => {
    if (!condition) return;
    if (!seen.has(condition.field) || (condition.value?.length || 0) > 2000)
      errors.push(
        "Conditions must reference a question in an earlier step or an earlier question.",
      );
    if (
      !["equals", "not_equals", "contains", "is_set"].includes(
        condition.operator,
      )
    )
      errors.push("Unsupported condition operator.");
  };
  for (const page of definition.pages) {
    if (!/^[a-zA-Z][a-zA-Z0-9_]{0,63}$/.test(page.id) || pageIds.has(page.id))
      errors.push("Each step needs a unique ID.");
    pageIds.add(page.id);
    if (!page.title.trim() || page.title.length > 120)
      errors.push("Every step needs a title of up to 120 characters.");
    if ((page.description?.length || 0) > 1000)
      errors.push("Step descriptions must be 1,000 characters or fewer.");
    if (!page.fields.length)
      errors.push("Every step needs at least one question.");
    checkCondition(page.condition);
    for (const field of page.fields) {
      fieldCount++;
      if (!isValidFormFieldKey(field.key) || seen.has(field.key))
        errors.push(
          `Use a unique, valid key for ${field.label || "each question"}.`,
        );
      if (!field.label.trim() || field.label.length > 120)
        errors.push("Every question needs a label of up to 120 characters.");
      if (
        (field.placeholder?.length || 0) > 200 ||
        (field.helpText?.length || 0) > 500
      )
        errors.push(`${field.label} needs shorter placeholder or help text.`);
      if (
        (field.options?.length || 0) > 100 ||
        field.options?.some((option) => option.length > 200)
      )
        errors.push(
          `${field.label} supports up to 100 options of 200 characters each.`,
        );
      if (!FORM_FIELD_TYPES.includes(field.type))
        errors.push("Unsupported question type.");
      if (field.type === "HIDDEN" && field.required && !field.defaultValue?.trim())
        errors.push(`${field.label} needs a default value because it is hidden and required.`);
      checkCondition(field.condition);
      if (
        field.key === "email" &&
        field.type === "EMAIL" &&
        field.required &&
        !field.condition &&
        !page.condition
      )
        email = true;
      if (
        ["SELECT", "RADIO"].includes(field.type) &&
        (!field.options?.length ||
          field.options.some((o) => !o.trim() || o !== o.trim()) ||
          new Set(field.options).size !== field.options.length)
      )
        errors.push(`${field.label} needs unique, non-empty options without surrounding spaces.`);
      if (
        field.defaultValue &&
        Object.keys(
          validateFormPage(
            { ...page, fields: [{ ...field, required: false }] },
            { [field.key]: field.defaultValue },
          ),
        ).length
      )
        errors.push(`${field.label} has an invalid default value.`);
      seen.add(field.key);
    }
  }
  if (fieldCount > 50) errors.push("Use 50 questions or fewer.");
  if (!email)
    errors.push(
      "Include an always-visible, required email question with the key email.",
    );
  if (!["required", "optional", "none"].includes(definition.consent.mode))
    errors.push("Choose a valid consent mode.");
  if (definition.consent.label.length > 500)
    errors.push("Consent wording must be 500 characters or fewer.");
  if (definition.consent.mode !== "none" && !definition.consent.label.trim())
    errors.push("Add a marketing consent label.");
  if (definition.successRedirectUrl) {
    try {
      const url = new URL(definition.successRedirectUrl);
      if (definition.successRedirectUrl.length > 2048 || url.protocol !== "https:" || url.username || url.password)
        throw new Error();
    } catch {
      errors.push("Use a full HTTPS URL for the success redirect.");
    }
  }
  return [...new Set(errors)];
}
