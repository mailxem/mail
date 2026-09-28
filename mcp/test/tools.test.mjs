import { test } from "node:test";
import assert from "node:assert/strict";
import { XemClient } from "../build/client.js";
import { createTools } from "../build/tools.js";
import { mapCSV, encodeCSV } from "../build/csv.js";
const id = "11111111-1111-4111-8111-111111111111";
const calls = [];
const api = new XemClient({ apiKey: "secret" }, async (url, options) => {
  calls.push({ url, options });
  return new Response(JSON.stringify({ ok: true }));
});
const tools = createTools(api);
test("tool schemas do not expose credentials or workspace overrides", () => {
  assert.equal(tools.size, 42);
  for (const { tool } of tools.values()) {
    assert.equal(tool.inputSchema.type, "object");
    assert.equal(tool.inputSchema.additionalProperties, false);
    assert.ok(!("token" in tool.inputSchema.properties));
    assert.ok(!("teamId" in tool.inputSchema.properties));
  }
});
test("reject invalid IDs, unknown fields and pagination before network", async () => {
  for (const [name, args] of [
    ["get_contact", { contactId: "../secret" }],
    ["get_contacts", { page: 0 }],
    ["get_contacts", { teamId: id }],
    ["create_contact", { listId: id, contact: { email: "bad" } }],
  ]) {
    const before = calls.length;
    await assert.rejects(tools.get(name).execute(args), /Invalid arguments/);
    assert.equal(calls.length, before);
  }
});
test("API key header, correct list endpoint and pagination", async () => {
  await tools.get("get_contact_lists").execute({});
  const { url, options } = calls.at(-1);
  assert.equal(
    url,
    "https://api.xem.email/api/v1/mailing-lists?page=1&limit=20",
  );
  assert.equal(options.headers["X-API-Key"], "secret");
  assert.equal(options.headers.Authorization, undefined);
  assert.equal(options.redirect, "error");
  assert.ok(options.signal);
});
test("campaign creation stays a draft and cannot smuggle status", async () => {
  const input = {
    name: "Launch",
    subject: "Hello",
    templateId: id,
    listId: id,
    smtpConfigId: id,
  };
  await tools.get("create_campaign").execute(input);
  assert.ok(calls.at(-1).url.endsWith("/marketing/campaign-drafts"));
  await assert.rejects(
    tools.get("create_campaign").execute({ ...input, status: "SENDING" }),
  );
});
test("CSV handles BOM, quoted commas/newlines and mapping", () => {
  assert.deepEqual(
    mapCSV('\uFEFFEmail,Name,Company\r\nADA@EXAMPLE.COM,"Ada, A","Acme\nInc"', {
      email: "Email",
      firstName: "Name",
      company: "Company",
    }),
    [{ email: "ada@example.com", firstName: "Ada, A", company: "Acme\nInc" }],
  );
});
test("CSV rejects malformed rows, duplicate/missing headers and oversized batches", () => {
  for (const csv of [
    "Email,Email\na@b.com,a@b.com",
    "Email,Name\na@b.com",
    "Email\ninvalid",
    "Other\na@b.com",
    "Email\n" + Array(501).fill("a@b.com").join("\n"),
  ])
    assert.throws(() => mapCSV(csv, { email: "Email" }));
});
test("invalid CSV never writes and imports default to preview", async () => {
  const before = calls.length;
  await assert.rejects(
    tools
      .get("import_contacts")
      .execute({ listId: id, csv: "Email\nbad", mappings: { email: "Email" } }),
  );
  assert.equal(calls.length, before);
  await tools.get("import_contacts").execute({
    listId: id,
    csv: "Email\na@b.com",
    mappings: { email: "Email" },
  });
  assert.deepEqual(JSON.parse(calls.at(-1).options.body), {
    listId: id,
    contacts: [{ email: "a@b.com" }],
    dryRun: true,
  });
});
test("CSV exports neutralize formulas and escape quotes", () => {
  const csv = encodeCSV([
    { email: "a@b.com", firstName: ' =HYPERLINK("bad")', phone: "+123" },
  ]);
  assert.ok(csv.includes('"\' =HYPERLINK(""bad"")"'));
  assert.ok(csv.includes('"\'+123"'));
});
test("redact upstream failures; handle 204 without JSON; forbid remote HTTP", async () => {
  const failing = new XemClient(
    { token: "secret" },
    async () => new Response("secret contact@example.com", { status: 500 }),
  );
  await assert.rejects(
    failing.request("/contacts"),
    (e) =>
      !e.message.includes("secret") &&
      !e.message.includes("contact@example.com"),
  );
  const empty = new XemClient(
    { token: "secret" },
    async () => new Response(null, { status: 204 }),
  );
  assert.deepEqual(await empty.request("/contacts", "DELETE"), { ok: true });
  assert.throws(
    () => new XemClient({ token: "s", baseUrl: "http://remote.example" }),
  );
  assert.throws(
    () =>
      new XemClient({ token: "s", baseUrl: "https://user:pass@example.com" }),
  );
});

test("assistant tools cannot publish forms or override workspace scope", async () => {
  const input = {
    name: "Signup",
    listId: id,
    fields: [{ label: "Email", key: "email", type: "EMAIL", required: true }],
  };
  await tools.get("create_form").execute(input);
  assert.equal(JSON.parse(calls.at(-1).options.body).status, "DRAFT");
  await assert.rejects(
    tools.get("create_form").execute({ ...input, status: "PUBLISHED" }),
  );
  await assert.rejects(tools.get("get_sending_status").execute({ teamId: id }));
  await assert.rejects(
    tools.get("pause_automation").execute({ automationId: "../../other" }),
  );
});
test("tag replacement and dispatch resumption are annotated as destructive", () => {
  for (const name of ["set_contact_tags", "set_sending_paused"]) {
    assert.equal(tools.get(name).tool.annotations.readOnlyHint, false);
    assert.equal(tools.get(name).tool.annotations.destructiveHint, true);
  }
});
