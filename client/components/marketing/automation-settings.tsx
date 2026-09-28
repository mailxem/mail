"use client";

import { Field } from "./shared";
import { Button } from "@/components/ui/button";
import type { LeadForm, Options } from "@/lib/marketing/types";

type Settings = {
  config: Record<string, any>;
  update: (values: Record<string, any>) => void;
};
const contactFields: Record<string, string> = {
  first_name: "First name",
  last_name: "Last name",
  company: "Company",
  country: "Country",
  city: "City",
  state: "State",
  zip: "Postal code",
  address: "Address",
  phone: "Phone",
  linkedin: "LinkedIn",
  twitter: "X / Twitter",
  facebook: "Facebook",
  instagram: "Instagram",
};
const operators: Record<string, string> = {
  "==": "Equals",
  "!=": "Does not equal",
  contains: "Contains",
  not_contains: "Does not contain",
  starts_with: "Starts with",
  ends_with: "Ends with",
  empty: "Is empty",
  not_empty: "Is not empty",
  exists: "Exists",
  not_exists: "Does not exist",
  ">": "Greater than",
  "<": "Less than",
  ">=": "At least",
  "<=": "At most",
};

export function ConditionSettings({
  config,
  update,
  form,
  variables,
}: Settings & { form?: LeadForm; variables: string[] }) {
  const rules: { variable: string; operator: string; value: string }[] =
    config.conditions || [];
  const change = (index: number, patch: Record<string, string>) =>
    update({
      conditions: rules.map((rule, i) =>
        i === index ? { ...rule, ...patch } : rule,
      ),
    });
  const choices: Record<string, string> = {
    contact_email: "Email address",
    ...Object.fromEntries(
      Object.entries(contactFields).map(([key, label]) => [
        `contact_${key}`,
        label,
      ]),
    ),
    current_list_id: "Current list ID",
    ...Object.fromEntries(variables.filter(Boolean).map((v) => [v, v])),
  };
  for (const field of form?.definition?.pages.flatMap((page) => page.fields) ||
    [])
    choices[`form_${field.key}`] = `Form: ${field.label}`;
  return (
    <>
      <Field label="Continue on the true path when">
        <select
          value={config.operator || "AND"}
          onChange={(e) => update({ operator: e.target.value })}
        >
          <option value="AND">All rules match (AND)</option>
          <option value="OR">Any rule matches (OR)</option>
        </select>
      </Field>
      {rules.map((rule, index) => (
        <div key={index} className="space-y-3 rounded-lg border p-3">
          <div className="flex items-center justify-between text-xs">
            <strong>Rule {index + 1}</strong>
            <button
              type="button"
              aria-label={`Remove rule ${index + 1}`}
              disabled={rules.length === 1}
              onClick={() =>
                update({ conditions: rules.filter((_, i) => i !== index) })
              }
              className="text-muted-foreground disabled:opacity-30"
            >
              Remove
            </button>
          </div>
          <Field label="Variable">
            <select
              value={choices[rule.variable] ? rule.variable : "custom"}
              onChange={(e) =>
                change(index, {
                  variable:
                    e.target.value === "custom"
                      ? "event_variable"
                      : e.target.value,
                })
              }
            >
              {Object.entries(choices).map(([key, label]) => (
                <option key={key} value={key}>
                  {label}
                </option>
              ))}
              <option value="custom">Custom event variable</option>
            </select>
          </Field>
          {!choices[rule.variable] && (
            <Field label="Variable key">
              <input
                maxLength={128}
                value={rule.variable}
                onChange={(e) => change(index, { variable: e.target.value })}
              />
            </Field>
          )}
          <Field label="Comparison">
            <select
              value={rule.operator}
              onChange={(e) => change(index, { operator: e.target.value })}
            >
              {Object.entries(operators).map(([key, label]) => (
                <option key={key} value={key}>
                  {label}
                </option>
              ))}
            </select>
          </Field>
          {!["empty", "not_empty", "exists", "not_exists"].includes(
            rule.operator,
          ) && (
            <Field label="Value">
              <input
                maxLength={1000}
                value={rule.value || ""}
                onChange={(e) => change(index, { value: e.target.value })}
              />
            </Field>
          )}
        </div>
      ))}
      <Button
        type="button"
        variant="outline"
        disabled={rules.length >= 20}
        onClick={() =>
          update({
            conditions: [
              ...rules,
              { variable: "contact_email", operator: "contains", value: "" },
            ],
          })
        }
      >
        Add rule
      </Button>
      <p className="text-xs text-muted-foreground">
        Text contains, starts with and ends with ignore case. Numeric
        comparisons also support ordered text. Missing values only match “is
        empty” or “does not exist”.
      </p>
    </>
  );
}

function localDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return new Date(date.getTime() - date.getTimezoneOffset() * 60000)
    .toISOString()
    .slice(0, 16);
}
export function WaitSettings({ config, update }: Settings) {
  const mode =
    config.until !== undefined && config.until !== null ? "date" : "duration";
  const match = /^(\d+(?:\.\d+)?)(s|m|h)$/.exec(String(config.duration || ""));
  return (
    <>
      <Field label="Wait mode">
        <select
          value={mode}
          onChange={(e) =>
            update(
              e.target.value === "date"
                ? {
                    until: new Date(Date.now() + 86400000).toISOString(),
                    duration: "",
                  }
                : { until: undefined, duration: "24h" },
            )
          }
        >
          <option value="duration">For a duration</option>
          <option value="date">Until a date and time</option>
        </select>
      </Field>
      {mode === "date" ? (
        <Field
          label="Resume at"
          hint={`Your timezone: ${Intl.DateTimeFormat().resolvedOptions().timeZone}. Contacts arriving after this time continue immediately.`}
        >
          <input
            type="datetime-local"
            value={localDate(String(config.until || ""))}
            onChange={(e) => {
              if (e.target.value)
                update({
                  until: new Date(e.target.value).toISOString(),
                  duration: "",
                });
            }}
          />
        </Field>
      ) : (
        <>
          <Field label="Quick duration">
            <select
              value={
                ["1h", "24h", "72h", "168h"].includes(config.duration)
                  ? config.duration
                  : "custom"
              }
              onChange={(e) => {
                if (e.target.value !== "custom")
                  update({ duration: e.target.value });
              }}
            >
              <option value="1h">1 hour</option>
              <option value="24h">1 day</option>
              <option value="72h">3 days</option>
              <option value="168h">1 week</option>
              <option value="custom">Custom duration</option>
            </select>
          </Field>
          {match ? (
            <div className="grid grid-cols-2 gap-2">
              <Field label="Amount">
                <input
                  type="number"
                  min="1"
                  value={match[1]}
                  onChange={(e) =>
                    update({ duration: `${e.target.value}${match[2]}` })
                  }
                />
              </Field>
              <Field label="Unit">
                <select
                  value={match[2]}
                  onChange={(e) =>
                    update({ duration: `${match[1]}${e.target.value}` })
                  }
                >
                  <option value="s">Seconds</option>
                  <option value="m">Minutes</option>
                  <option value="h">Hours</option>
                </select>
              </Field>
            </div>
          ) : (
            <Field label="Duration">
              <input
                value={String(config.duration || "")}
                onChange={(e) => update({ duration: e.target.value })}
              />
            </Field>
          )}
          <p className="text-xs text-muted-foreground">
            Between 1 second and 365 days. One day is 24 hours.
          </p>
        </>
      )}
    </>
  );
}

export function ActionSettings({
  kind,
  config,
  update,
  options,
}: Settings & { kind: string; options?: Options }) {
  if (kind === "TAG")
    return (
      <>
        <Field label="Action">
          <select
            value={config.action || "add"}
            onChange={(e) => update({ action: e.target.value })}
          >
            <option value="add">Add tags</option>
            <option value="remove">Remove tags</option>
          </select>
        </Field>
        <Field
          label="Tags"
          hint="One tag per line, up to 20. New tags are created when added."
        >
          <textarea
            rows={4}
            value={(config.tags || []).join("\n")}
            onChange={(e) => update({ tags: e.target.value.split("\n") })}
          />
        </Field>
      </>
    );
  if (kind === "ADD_TO_LIST")
    return (
      <Field
        label="Destination list"
        hint="Moves the contact out of their current list. Subscription status stays the same."
      >
        <select
          value={config.listId || ""}
          onChange={(e) => update({ listId: e.target.value })}
        >
          <option value="">Choose a list</option>
          {options?.lists.map((list) => (
            <option key={list.id} value={list.id}>
              {list.name}
            </option>
          ))}
        </select>
      </Field>
    );
  if (kind === "PERCENTAGE_SPLIT")
    return (
      <>
        <Field label="Percentage taking path A">
          <input
            type="number"
            min={1}
            max={99}
            value={config.percentage ?? 50}
            onChange={(e) => update({ percentage: Number(e.target.value) })}
          />
        </Field>
        <p className="text-xs text-muted-foreground">
          Path B receives {100 - Number(config.percentage ?? 50)}%. Each contact
          keeps the same path on retries. Small groups may not match the exact
          ratio.
        </p>
      </>
    );
  if (kind === "SET_VARIABLE")
    return (
      <>
        <Field
          label="Variable name"
          hint="Start with workflow_, then a letter. Use letters, numbers and underscores."
        >
          <input
            maxLength={73}
            value={config.variable || ""}
            onChange={(e) => update({ variable: e.target.value })}
          />
        </Field>
        <Field
          label="Value"
          hint="Stored as text for conditions later in this run."
        >
          <input
            maxLength={1000}
            value={config.value || ""}
            onChange={(e) => update({ value: e.target.value })}
          />
        </Field>
      </>
    );
  if (kind === "UPDATE_SUBSCRIBER") {
    const fields: Record<string, string> = config.fields || {};
    return (
      <>
        {Object.entries(fields).map(([field, value]) => (
          <div key={field} className="space-y-2 rounded-lg border p-3">
            <Field label={contactFields[field] || field}>
              <input
                maxLength={1000}
                value={value}
                onChange={(e) =>
                  update({ fields: { ...fields, [field]: e.target.value } })
                }
              />
            </Field>
            <button
              type="button"
              className="text-xs text-muted-foreground"
              onClick={() =>
                update({
                  fields: Object.fromEntries(
                    Object.entries(fields).filter(([key]) => key !== field),
                  ),
                })
              }
            >
              Remove field
            </button>
          </div>
        ))}
        <Field label="Add a profile field">
          <select
            value=""
            onChange={(e) => {
              if (e.target.value)
                update({ fields: { ...fields, [e.target.value]: "" } });
            }}
          >
            <option value="">Choose a field</option>
            {Object.entries(contactFields)
              .filter(([key]) => !(key in fields))
              .map(([key, label]) => (
                <option key={key} value={key}>
                  {label}
                </option>
              ))}
          </select>
        </Field>
        <p className="text-xs text-muted-foreground">
          Values replace the current text. An empty value clears a field.
          Subscription status and email address cannot be changed here.
        </p>
      </>
    );
  }
  return null;
}
