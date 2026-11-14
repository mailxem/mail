// Contact fields every email can use, filled per reader when sent
// (server/internal/tasks/handlers.go). {{first_name|there}} prints "there"
// for a reader with no first name.
export const CONTACT_FIELDS = [
  { key: "first_name", label: "First name" },
  { key: "last_name", label: "Last name" },
  { key: "name", label: "Full name" },
  { key: "email", label: "Email" },
  { key: "company", label: "Company" },
  { key: "city", label: "City" },
  { key: "state", label: "State" },
  { key: "country", label: "Country" },
  { key: "phone", label: "Phone" },
] as const;

/** The partial field name being typed after an unclosed "{{" before the caret. */
export function fieldQuery(text: string, caret: number) {
  return /{{\s*(\w*)$/.exec(text.slice(0, caret))?.[1];
}

/** Fields whose key or label starts with the query. */
export function matchFields(query: string) {
  const q = query.toLowerCase();
  return CONTACT_FIELDS.filter(
    (f) => f.key.startsWith(q) || f.label.toLowerCase().startsWith(q),
  );
}

/** Completes the "{{" before the caret with field; returns text and new caret. */
export function insertField(text: string, caret: number, field: string) {
  const start = text.slice(0, caret).lastIndexOf("{{");
  const rest = text.slice(caret).replace(/^\w*\s*}}/, "");
  const token = `{{${field}}}`;
  return {
    text: text.slice(0, start) + token + rest,
    caret: start + token.length,
  };
}

export type UsedField = { key: string; fallback?: string };

/** Contact fields used in the email, each with its fallback if every use has one. */
export function usedFields(...sources: (string | undefined)[]): UsedField[] {
  const found = new Map<string, UsedField>();
  const re = /{{\s*(\w+)\s*(?:\|([^{}]*))?}}/g;
  for (const source of sources)
    for (const [, key, fallback] of (source ?? "").matchAll(re)) {
      const seen = found.get(key);
      const fb = fallback === undefined ? undefined : fallback.trim();
      // One use without a fallback means some readers see a blank.
      if (!seen) found.set(key, { key, fallback: fb });
      else if (fb === undefined) seen.fallback = undefined;
    }
  return [...found.values()];
}
