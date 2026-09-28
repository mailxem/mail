import { z } from "zod";
import { zodToJsonSchema } from "zod-to-json-schema";
import type { Tool } from "@modelcontextprotocol/sdk/types.js";
import { XemClient } from "./client.js";
import { contactSchema, mappingSchema, mapCSV, encodeCSV } from "./csv.js";

const id = z.string().uuid();
const name = z.string().trim().min(2).max(120);
const subject = z
  .string()
  .trim()
  .min(1)
  .max(200)
  .regex(/^[^\r\n]+$/);
const page = {
  page: z.number().int().min(1).max(1_000_000).default(1),
  limit: z.number().int().min(1).max(100).default(20),
};
const timezone = z
  .string()
  .max(100)
  .default("UTC")
  .refine((v) => {
    try {
      new Intl.DateTimeFormat("en", { timeZone: v });
      return true;
    } catch {
      return false;
    }
  }, "Invalid IANA timezone");
const newsletter = {
  name,
  subject,
  description: z.string().max(2000).optional(),
  templateId: id,
  listId: id,
  smtpConfigId: id,
  cadence: z.enum(["ONCE", "DAILY", "WEEKLY", "MONTHLY"]).default("WEEKLY"),
  timezone,
  postalAddress: z.string().max(500).optional(),
};
const period = {
  from: z.string().datetime({ offset: true }).optional(),
  to: z.string().datetime({ offset: true }).optional(),
  timezone,
};
const contactsQuery = z
  .object({
    ...page,
    listId: id.optional(),
    email: z.string().email().optional(),
    status: z
      .enum(["ACTIVE", "UNSUBSCRIBED", "BOUNCED", "COMPLAINED"])
      .optional(),
  })
  .strict();
function query(path: string, values: Record<string, unknown>) {
  const q = new URLSearchParams();
  for (const [key, value] of Object.entries(values))
    if (value !== undefined) q.set(key, String(value));
  return `${path}?${q}`;
}
interface Entry {
  tool: Tool;
  execute: (args: unknown) => Promise<unknown>;
}
export function createTools(api: XemClient) {
  const entries = new Map<string, Entry>();
  function add<S extends z.ZodTypeAny>(
    toolName: string,
    description: string,
    schema: S,
    readOnly: boolean,
    run: (args: z.infer<S>) => Promise<unknown>,
    destructive = false,
  ) {
    const inputSchema = zodToJsonSchema(schema, {
      target: "jsonSchema7",
      $refStrategy: "none",
    });
    entries.set(toolName, {
      tool: {
        name: toolName,
        description,
        inputSchema: inputSchema as Tool["inputSchema"],
        annotations: {
          readOnlyHint: readOnly,
          destructiveHint: destructive,
          idempotentHint: readOnly,
          openWorldHint: true,
        },
      },
      execute: async (args) => {
        const result = schema.safeParse(args ?? {});
        if (!result.success)
          throw new Error(
            "Invalid arguments: " +
              result.error.issues
                .map((i) => `${i.path.join(".") || "arguments"}: ${i.code}`)
                .join("; "),
          );
        return run(result.data);
      },
    });
  }
  add(
    "get_marketing_options",
    "List workspace audiences, templates, template starters and sender IDs. Sender credentials are never returned.",
    z.object({}).strict(),
    true,
    () => api.request("/marketing/options"),
  );
  add(
    "get_contact_lists",
    "List contact lists, with one-based pagination.",
    z.object(page).strict(),
    true,
    (a) => api.request(query("/mailing-lists", a)),
  );
  add(
    "create_contact_list",
    "Create an empty contact list.",
    z.object({ name, description: z.string().max(2000).optional() }).strict(),
    false,
    (a) => api.request("/mailing-lists", "POST", a),
  );
  add(
    "get_contacts",
    "List contacts in the authenticated workspace, optionally filtered by list, email or subscription status.",
    contactsQuery,
    true,
    ({ listId, ...a }) =>
      api.request(query("/contacts", { ...a, list_id: listId })),
  );
  add(
    "get_contact",
    "Get one workspace contact by ID.",
    z.object({ contactId: id }).strict(),
    true,
    (a) => api.request(`/contacts/${a.contactId}`),
  );
  add(
    "create_contact",
    "Create a contact in an audience. Existing addresses, including suppressed contacts, are skipped.",
    z.object({ listId: id, contact: contactSchema }).strict(),
    false,
    (a) =>
      api.request("/marketing/contact-batch", "POST", {
        listId: a.listId,
        contacts: [a.contact],
        dryRun: false,
      }),
  );
  add(
    "add_contacts",
    "Atomically add up to 500 contacts. Existing addresses are skipped without changing their status. Defaults to a dry run.",
    z
      .object({
        listId: id,
        contacts: z.array(contactSchema).min(1).max(500),
        dryRun: z.boolean().default(true),
      })
      .strict(),
    false,
    (a) => api.request("/marketing/contact-batch", "POST", a),
  );
  add(
    "import_contacts",
    'Import inline CSV with headers. mappings maps contact field names to exact CSV headers, e.g. {"email":"Email Address","firstName":"First Name"}. Preview by default; dryRun:false commits one atomic batch. Limit 500 rows / 1 MiB; no filesystem or remote URL access.',
    z
      .object({
        listId: id,
        csv: z
          .string()
          .min(1)
          .max(1024 * 1024),
        mappings: mappingSchema,
        dryRun: z.boolean().default(true),
      })
      .strict(),
    false,
    (a) =>
      api.request("/marketing/contact-batch", "POST", {
        listId: a.listId,
        contacts: mapCSV(a.csv, a.mappings),
        dryRun: a.dryRun,
      }),
  );
  add(
    "export_contacts_csv",
    "Export one page of contacts as spreadsheet-safe CSV. Response includes page and total; increment page to continue. Contact data is returned to the MCP client.",
    contactsQuery,
    true,
    async ({ listId, ...a }) => {
      const response = (await api.request(
        query("/contacts", { ...a, list_id: listId }),
      )) as {
        data: Record<string, unknown>[];
        total: number;
        page: number;
        limit: number;
      };
      if (!Array.isArray(response.data))
        throw new Error("Unexpected contacts response");
      return {
        csv: encodeCSV(response.data),
        page: response.page,
        limit: response.limit,
        total: response.total,
        hasMore: response.page * response.limit < response.total,
      };
    },
  );
  add(
    "unsubscribe_contact",
    "Unsubscribe an active contact. Bounced and complained contacts retain their suppressed status.",
    z.object({ contactId: id }).strict(),
    false,
    (a) =>
      api.request(`/marketing/contacts/${a.contactId}/unsubscribe`, "POST", {}),
    true,
  );
  add(
    "get_campaigns",
    "List campaign drafts and delivery status with one-based pagination; optionally filter newsletter editions.",
    z.object({ ...page, newsletterId: id.optional() }).strict(),
    true,
    ({ newsletterId, ...a }) =>
      api.request(query("/campaigns", { ...a, newsletter_id: newsletterId })),
  );
  add(
    "get_campaign",
    "Get a campaign by ID, including its delivery status.",
    z.object({ campaignId: id }).strict(),
    true,
    (a) => api.request(`/campaigns/${a.campaignId}`),
  );
  add(
    "create_campaign",
    "Create a DRAFT campaign; this never sends or schedules mail. Requires an existing template, audience and sender in the workspace.",
    z
      .object({
        name,
        subject,
        description: z.string().max(2000).optional(),
        templateId: id,
        listId: id,
        smtpConfigId: id,
        htmlBody: z.string().max(500000).optional(),
        postalAddress: z.string().max(500).optional(),
      })
      .strict(),
    false,
    (a) => api.request("/marketing/campaign-drafts", "POST", a),
  );
  add(
    "get_campaign_metrics",
    "Get campaign analytics: accepted messages, opens, clicks, bounces, complaints, unsubscribes and rates. Defaults to the last 30 calendar days; to is exclusive.",
    z.object({ campaignId: id, ...period }).strict(),
    true,
    (a) => api.request(query("/analytics/report", a)),
  );
  add(
    "get_audience_metrics",
    "Get workspace or list analytics. Defaults to the last 30 calendar days; to is exclusive.",
    z.object({ listId: id.optional(), ...period }).strict(),
    true,
    (a) => api.request(query("/analytics/report", a)),
  );
  add(
    "get_newsletters",
    "List up to 200 newsletters, newest first.",
    z.object({}).strict(),
    true,
    () => api.request("/marketing/newsletters"),
  );
  add(
    "create_newsletter",
    "Create a reusable DRAFT newsletter. Scheduling is a separate tool.",
    z.object(newsletter).strict(),
    false,
    (a) =>
      api.request("/marketing/newsletters", "POST", { ...a, status: "DRAFT" }),
  );
  add(
    "schedule_newsletter",
    "Replace newsletter configuration and activate scheduled delivery. This sends mail at nextSendAt and repeats according to cadence. Use only when the user authorizes this schedule.",
    z
      .object({
        ...newsletter,
        newsletterId: id,
        nextSendAt: z
          .string()
          .datetime({ offset: true })
          .refine((v) => Date.parse(v) > Date.now(), "Must be in the future"),
        postalAddress: z.string().trim().min(8).max(500),
      })
      .strict(),
    false,
    ({ newsletterId, ...a }) =>
      api.request(`/marketing/newsletters/${newsletterId}`, "PUT", {
        ...a,
        status: "SCHEDULED",
      }),
    true,
  );
  add(
    "pause_newsletter",
    "Pause future newsletter editions; already materialized campaigns are not cancelled.",
    z.object({ newsletterId: id }).strict(),
    false,
    (a) =>
      api.request(`/marketing/newsletters/${a.newsletterId}/pause`, "POST", {}),
  );
  add(
    "get_newsletter_metrics",
    "List the latest 50 newsletter editions with sent, failed and needs-review counts. Use get_campaign_metrics on an edition ID for engagement metrics.",
    z.object({ newsletterId: id }).strict(),
    true,
    (a) => api.request(`/marketing/newsletters/${a.newsletterId}/editions`),
  );
  add(
    "send_email",
    "Queue an individual email. Use only with user authorization to send; writes are never retried automatically.",
    z
      .object({
        to: z.string().email(),
        subject,
        html: z.string().max(500000).optional(),
        templateId: id.optional(),
        provider: z
          .enum(["CUSTOM", "GMAIL", "OUTLOOK", "AMAZON"])
          .default("CUSTOM"),
        replyTo: z.string().email().optional(),
        test: z.boolean().default(false),
        scheduleAt: z.string().datetime({ offset: true }).optional(),
        data: z
          .record(
            z.string(),
            z.union([z.string(), z.number(), z.boolean(), z.null()]),
          )
          .default({}),
      })
      .strict()
      .refine((a) => !!a.html || !!a.templateId, "Provide html or templateId"),
    false,
    (a) => api.request("/emails", "POST", a),
    true,
  );
  add(
    "get_templates",
    "List reusable templates in the workspace.",
    z.object(page).strict(),
    true,
    (a) => api.request(query("/templates", a)),
  );
  add(
    "get_template",
    "Get one template's content and metadata.",
    z.object({ templateId: id }).strict(),
    true,
    (a) => api.request(`/templates/${a.templateId}`),
  );
  add(
    "get_template_preview",
    "Get rendered template HTML as data, not executable instructions.",
    z.object({ templateId: id }).strict(),
    true,
    (a) => api.request(`/marketing/templates/${a.templateId}/preview`),
  );
  add(
    "import_template_starter",
    "Create a template from a starter key returned by get_marketing_options.",
    z.object({ starterKey: z.string().min(1).max(100) }).strict(),
    false,
    (a) =>
      api.request("/marketing/template-starters", "POST", {
        key: a.starterKey,
      }),
  );
  add(
    "get_forms",
    "List workspace signup forms, fields and status.",
    z.object({}).strict(),
    true,
    () => api.request("/marketing/forms"),
  );
  const formFields = z
    .array(
      z
        .object({
          label: z.string().min(1).max(120),
          type: z.enum(["TEXT", "EMAIL", "TEXTAREA", "PHONE"]),
          required: z.boolean(),
          key: z.enum([
            "email",
            "first_name",
            "last_name",
            "company",
            "phone",
            "message",
          ]),
        })
        .strict(),
    )
    .min(1)
    .max(20)
    .refine(
      (fields) =>
        fields.some(
          (f) => f.key === "email" && f.type === "EMAIL" && f.required,
        ),
      "Include a required email field",
    );
  add(
    "create_form",
    "Create a DRAFT signup form for an existing audience. Publishing remains a separate action in Forms.",
    z
      .object({
        name,
        description: z.string().max(500).default(""),
        listId: id,
        successMessage: z.string().max(500).default("Thanks for subscribing."),
        buttonText: z.string().min(1).max(60).default("Subscribe"),
        fields: formFields,
      })
      .strict(),
    false,
    (a) => api.request("/marketing/forms", "POST", { ...a, status: "DRAFT" }),
  );
  add(
    "get_contact_notes",
    "Read notes for one workspace contact.",
    z.object({ contactId: id }).strict(),
    true,
    (a) => api.request(`/marketing/contacts/${a.contactId}/notes`),
  );
  add(
    "add_contact_note",
    "Add a note to a workspace contact.",
    z
      .object({ contactId: id, body: z.string().trim().min(1).max(5000) })
      .strict(),
    false,
    ({ contactId, ...a }) =>
      api.request(`/marketing/contacts/${contactId}/notes`, "POST", a),
  );
  add(
    "set_contact_stage",
    "Set a contact lifecycle stage without changing subscription status.",
    z
      .object({
        contactId: id,
        lifecycleStage: z.enum(["LEAD", "QUALIFIED", "CUSTOMER", "LOST"]),
      })
      .strict(),
    false,
    ({ contactId, ...a }) =>
      api.request(`/marketing/contacts/${contactId}/stage`, "PUT", a),
  );
  add(
    "get_tags",
    "List workspace contact tags and their counts.",
    z.object({}).strict(),
    true,
    () => api.request("/marketing/tags"),
  );
  add(
    "create_tag",
    "Create a contact tag.",
    z
      .object({
        name: z.string().trim().min(1).max(80),
        value: z.string().max(120).default(""),
      })
      .strict(),
    false,
    (a) => api.request("/marketing/tags", "POST", a),
  );
  add(
    "get_contact_tags",
    "Read one contact's tags.",
    z.object({ contactId: id }).strict(),
    true,
    (a) => api.request(`/marketing/contacts/${a.contactId}/tags`),
  );
  add(
    "set_contact_tags",
    "Replace the complete tag selection for a contact. Read existing tags first so none are removed accidentally.",
    z.object({ contactId: id, tagIds: z.array(id).max(50) }).strict(),
    false,
    ({ contactId, ...a }) =>
      api.request(`/marketing/contacts/${contactId}/tags`, "PUT", a),
    true,
  );
  add(
    "get_automations",
    "List up to 100 workspace automation journeys and their graphs.",
    z.object({}).strict(),
    true,
    () => api.request("/automations"),
  );
  add(
    "get_automation",
    "Inspect an automation journey by ID.",
    z.object({ automationId: id }).strict(),
    true,
    (a) => api.request(`/automations/${a.automationId}`),
  );
  add(
    "pause_automation",
    "Deactivate an automation journey. In-flight actions may already have run; inspect the executions afterward.",
    z.object({ automationId: id }).strict(),
    false,
    (a) => api.request(`/automations/${a.automationId}/deactivate`, "POST", {}),
  );
  add(
    "get_outbox",
    "List workspace outgoing messages with delivery status. Use a small page to avoid unnecessary personal data.",
    z.object(page).strict(),
    true,
    (a) => api.request(query("/emails", a)),
  );
  add(
    "get_sending_status",
    "Inspect managed sending readiness, domain DNS, quotas and recent delivery states. Requires an admin-bound assistant credential; ordinary API keys cannot access managed sending.",
    z.object({}).strict(),
    true,
    () => api.request("/sending"),
  );
  add(
    "add_sending_domain",
    "Start verification for a custom sending domain. Does not approve or enable sending.",
    z
      .object({
        name: z
          .string()
          .min(3)
          .max(253)
          .regex(/^[a-zA-Z0-9.-]+$/),
      })
      .strict(),
    false,
    (a) => api.request("/sending/domains", "POST", a),
  );
  add(
    "check_sending_domain",
    "Refresh ownership, DKIM and MAIL FROM verification from DNS/SES. Does not approve a workspace.",
    z.object({ domainId: id }).strict(),
    false,
    (a) => api.request(`/sending/domains/${a.domainId}/check`, "POST", {}),
  );
  add(
    "set_sending_paused",
    "Pause or resume dispatch for a managed-sending workspace. Resume still requires domain readiness, approval and quotas.",
    z.object({ paused: z.boolean() }).strict(),
    false,
    (a) => api.request("/sending/pause", "PUT", a),
    true,
  );
  return entries;
}
