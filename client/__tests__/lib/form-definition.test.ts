import {
  evaluateCondition,
  getFormDefinition,
  getReachableAnswers,
  getReachablePages,
  validateFormDefinition,
  validateFormPage,
  type FormDefinition,
} from "@/lib/marketing/form-definition";

const definition = (): FormDefinition => ({
  schemaVersion: 1,
  saveProgress: true,
  consent: { mode: "optional", label: "Send product news" },
  pages: [
    {
      id: "identity",
      title: "About you",
      fields: [
        { key: "email", label: "Email", type: "EMAIL", required: true },
        {
          key: "role",
          label: "Role",
          type: "SELECT",
          required: true,
          options: ["Developer", "Designer"],
        },
      ],
    },
    {
      id: "engineering",
      title: "Engineering",
      condition: { field: "role", operator: "equals", value: "Developer" },
      fields: [
        { key: "language", label: "Language", type: "TEXT", required: true },
        {
          key: "source",
          label: "Source",
          type: "HIDDEN",
          required: false,
          defaultValue: "engineering-form",
        },
      ],
    },
    {
      id: "followup",
      title: "Follow up",
      fields: [
        {
          key: "details",
          label: "Details",
          type: "TEXTAREA",
          required: true,
          condition: {
            field: "language",
            operator: "not_equals",
            value: "Rust",
          },
        },
        {
          key: "feedback",
          label: "Feedback",
          type: "TEXTAREA",
          required: false,
        },
      ],
    },
  ],
});

describe("conditional form reachability", () => {
  it("omits skipped answers and prevents stale hidden branches activating downstream questions", () => {
    const answers = {
      email: "reader@example.com",
      role: "Designer",
      language: "Go",
      source: "injected",
      details: "stale",
      feedback: "Hello",
      unknown: "injected",
    };
    const pages = getReachablePages(definition(), answers);
    expect(pages.map((p) => p.id)).toEqual(["identity", "followup"]);
    expect(pages[1].fields.map((f) => f.key)).toEqual(["feedback"]);
    expect(getReachableAnswers(definition(), answers)).toEqual({
      email: "reader@example.com",
      role: "Designer",
      feedback: "Hello",
    });
    // The caller keeps local values for Back/edit; pruning only affects the submitted payload.
    expect(answers.language).toBe("Go");
  });
  it("uses configured hidden defaults and preserves reachable answers when navigating back", () => {
    const answers = getReachableAnswers(definition(), {
      email: "reader@example.com",
      role: "Developer",
      language: "Go",
      source: "attacker",
      details: "Build tools",
    });
    expect(answers.source).toBe("engineering-form");
    expect(answers.details).toBe("Build tools");
  });
  it("does not validate required fields on unreachable branches", () => {
    const answers = { email: "reader@example.com", role: "Designer" };
    for (const page of getReachablePages(definition(), answers))
      expect(validateFormPage(page, answers)).toEqual({});
  });
  it("does not regard unchecked checkboxes or whitespace as set", () => {
    expect(
      evaluateCondition(
        { field: "choice", operator: "is_set" },
        { choice: "false" },
      ),
    ).toBe(false);
    expect(
      evaluateCondition(
        { field: "choice", operator: "is_set" },
        { choice: "  " },
      ),
    ).toBe(false);
    expect(
      evaluateCondition(
        { field: "missing", operator: "not_equals", value: "x" },
        {},
      ),
    ).toBe(false);
  });
});

describe("form validation", () => {
  it("requires an unconditional email and rejects forward references and duplicate keys", () => {
    expect(validateFormDefinition(definition())).toEqual([]);
    const invalid = definition();
    invalid.pages[0].fields[0].condition = {
      field: "language",
      operator: "is_set",
    };
    invalid.pages[1].fields[0].key = "role";
    expect(validateFormDefinition(invalid)).toEqual(
      expect.arrayContaining([
        "Conditions must reference a question in an earlier step or an earlier question.",
        "Include an always-visible, required email question with the key email.",
        "Use a unique, valid key for Language.",
      ]),
    );
  });
  it("checks real date values, finite numbers, options, email and required checkbox", () => {
    const page = {
      id: "test",
      title: "Test",
      fields: [
        {
          key: "birthday",
          label: "Birthday",
          type: "DATE" as const,
          required: false,
        },
        {
          key: "count",
          label: "Count",
          type: "NUMBER" as const,
          required: false,
        },
        {
          key: "role",
          label: "Role",
          type: "RADIO" as const,
          required: false,
          options: ["Developer"],
        },
        {
          key: "terms",
          label: "Terms",
          type: "CHECKBOX" as const,
          required: true,
        },
        {
          key: "email",
          label: "Email",
          type: "EMAIL" as const,
          required: true,
        },
      ],
    };
    expect(
      Object.keys(
        validateFormPage(page, {
          birthday: "2026-02-30",
          count: "Infinity",
          role: "Other",
          terms: "false",
          email: "x@",
        }),
      ),
    ).toHaveLength(5);
    expect(
      validateFormPage(page, {
        birthday: "2024-02-29",
        count: "1.5",
        role: "Developer",
        terms: "true",
        email: "x@example.com",
      }),
    ).toEqual({});
  });
  it("rejects unsafe redirects and protocol field keys", () => {
    const invalid = definition();
    invalid.successRedirectUrl = "javascript:alert(1)";
    invalid.pages[0].fields[1].key = "constructor";
    expect(validateFormDefinition(invalid)).toEqual(
      expect.arrayContaining([
        "Use a full HTTPS URL for the success redirect.",
        "Use a unique, valid key for Role.",
      ]),
    );
  });
  it("adapts legacy fields with required explicit consent", () => {
    expect(
      getFormDefinition({
        fields: [
          {
            Label: "Email",
            FieldType: "EMAIL",
            Required: true,
            mapToContactField: "email",
          },
        ],
      }),
    ).toMatchObject({
      schemaVersion: 1,
      saveProgress: false,
      consent: { mode: "required" },
      pages: [
        {
          id: "details",
          fields: [{ key: "email", type: "EMAIL", required: true }],
        },
      ],
    });
  });
});


test("required hidden questions must have a configured answer", () => {
  const value = definition();
  value.pages[0].fields.push({ key: "source", label: "Source", type: "HIDDEN", required: true });
  expect(validateFormDefinition(value)).toContain("Source needs a default value because it is hidden and required.");
});


test("choice options reject whitespace that the server trims from answers", () => {
  const value = definition();
  value.pages[0].fields[1].options = [" Developer ", "Designer"];
  expect(validateFormDefinition(value)).toContain("Role needs unique, non-empty options without surrounding spaces.");
});
