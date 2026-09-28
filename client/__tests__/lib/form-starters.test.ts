import {
  copyFormDefinition,
  fieldIsReferenced,
  formSavePayload,
  formStarters,
  renameFormField,
} from "@/lib/marketing/form-starters";
import {
  getReachablePages,
  validateFormDefinition,
  type FormDefinition,
} from "@/lib/marketing/form-definition";
import { formPresets } from "@/lib/marketing/form-theme";
import type { LeadForm } from "@/lib/marketing/types";

describe("form editor starters and persisted definitions", () => {
  it.each(formStarters)(
    "$name starts as a valid, editable definition",
    (starter) => {
      expect(validateFormDefinition(starter.definition)).toEqual([]);
      const copy = copyFormDefinition(starter.definition);
      copy.pages[0].fields[0].label = "Changed";
      expect(starter.definition.pages[0].fields[0].label).not.toBe("Changed");
    },
  );

  it("does not turn feedback or preferences into marketing subscriptions", () => {
    for (const key of ["feedback", "preferences"]) {
      expect(
        formStarters.find((starter) => starter.key === key)?.definition.consent
          .mode,
      ).toBe("none");
    }
    expect(
      formStarters.find((starter) => starter.key === "lead-magnet")?.definition
        .consent.mode,
    ).toBe("optional");
  });

  it("preserves branching, consent, resume, redirect, and branding when duplicating or pausing", () => {
    const definition = copyFormDefinition(
      formStarters.find((starter) => starter.key === "demo")!.definition,
    );
    definition.successRedirectUrl = "https://example.com/thanks";
    const form: LeadForm = {
      id: "form-1",
      Name: "Demo",
      description: "Meet us",
      Status: "PUBLISHED",
      Slug: "demo",
      AddToListID: "list-1",
      fields: [],
      successMessage: "Thanks",
      SubmitButtonText: "Send",
      SubmissionCount: 2,
      ViewCount: 4,
      updatedAt: "2026-09-22",
      theme: formPresets.dark,
      definition,
    };
    const payload = formSavePayload(form);
    expect(payload.definition).toEqual(definition);
    expect(payload.theme).toEqual(formPresets.dark);
    expect(
      { ...payload, status: "ARCHIVED" }.definition?.pages[1].fields[1]
        .condition?.field,
    ).toBe("team_size");
    expect({ ...payload, status: "DRAFT" }.definition?.saveProgress).toBe(true);
    expect(
      formSavePayload({ ...form, definition: undefined }),
    ).not.toHaveProperty("definition");
  });

  it("keeps later page and question conditions connected when a field key changes", () => {
    const definition: FormDefinition = {
      schemaVersion: 1,
      consent: { mode: "none", label: "" },
      saveProgress: false,
      pages: [
        {
          id: "contact",
          title: "Contact",
          fields: [
            { key: "email", label: "Email", type: "EMAIL", required: true },
            { key: "role", label: "Role", type: "TEXT", required: false },
          ],
        },
        {
          id: "business",
          title: "Business",
          condition: { field: "role", operator: "equals", value: "owner" },
          fields: [
            {
              key: "company",
              label: "Company",
              type: "TEXT",
              required: false,
              condition: { field: "role", operator: "is_set" },
            },
          ],
        },
      ],
    };
    const renamed = renameFormField(definition, "role", "job_role", {
      pageIndex: 0,
      fieldIndex: 1,
    });
    expect(validateFormDefinition(renamed)).toEqual([]);
    expect(
      getReachablePages(renamed, {
        email: "test@example.com",
        job_role: "owner",
      }),
    ).toHaveLength(2);
    expect(renamed.pages[1].fields[0].condition?.field).toBe("job_role");
    expect(fieldIsReferenced(renamed, "job_role")).toBe(true);
    expect(fieldIsReferenced(renamed, "role")).toBe(false);
    expect(definition.pages[0].fields[1].key).toBe("role");
  });

  it("lets an author repair duplicate keys without renaming both questions", () => {
    const definition = copyFormDefinition(formStarters[0].definition);
    definition.pages[0].fields.push({
      key: "first_name",
      label: "Other",
      type: "TEXT",
      required: false,
    });
    const repaired = renameFormField(definition, "first_name", "other_name", {
      pageIndex: 0,
      fieldIndex: 2,
    });
    expect(repaired.pages[0].fields[1].key).toBe("first_name");
    expect(repaired.pages[0].fields[2].key).toBe("other_name");
    expect(validateFormDefinition(repaired)).toEqual([]);
  });
});
