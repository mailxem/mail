// Recognises subscriber exports from other newsletter tools and keeps only
// people who are still subscribed, so a migration never re-mails someone who
// unsubscribed elsewhere.

export type Row = Record<string, string>;

/** Contact fields the import worker reads (server/internal/tasks/handlers.go). */
export const CONTACT_FIELDS = [
  "email",
  "first_name",
  "last_name",
  "phone",
  "company",
  "address",
  "city",
  "state",
  "country",
  "zip",
  "linkedin",
  "twitter",
] as const;

const ALIASES: Record<string, string[]> = {
  email: ["email", "emailaddress", "email_address", "e-mail", "subscriber"],
  first_name: ["firstname", "first", "givenname", "fname"],
  last_name: ["lastname", "last", "surname", "familyname", "lname"],
  phone: ["phone", "phonenumber", "mobile", "telephone", "tel"],
  company: ["company", "organization", "organisation", "companyname"],
  address: ["address", "addressline1", "streetaddress"],
  city: ["city", "town"],
  state: ["state", "province", "region"],
  country: ["country", "countryname", "countrycode"],
  zip: ["zip", "zipcode", "postalcode", "postcode"],
  linkedin: ["linkedin"],
  twitter: ["twitter", "x"],
};

const norm = (s: string) => s.toLowerCase().replace(/[^a-z0-9]/g, "");

export function guessField(header: string) {
  const h = norm(header);
  for (const [field, names] of Object.entries(ALIASES))
    if (names.some((n) => norm(n) === h)) return field;
  return null;
}

export type Source = "beehiiv" | "Substack" | "Mailchimp" | "Ghost" | "Kit";

export function detectSource(headers: string[], fileName = ""): Source | null {
  const h = new Set(headers.map(norm));
  const f = fileName.toLowerCase();
  if (h.has("emaildisabled") || h.has("activesubscription")) return "Substack";
  if (h.has("subscribedtoemails")) return "Ghost";
  if (h.has("emailaddress") && (h.has("memberrating") || h.has("optintime") || f.includes("members_export")))
    return "Mailchimp";
  if (h.has("subscriptiontier") || h.has("utmsource") || f.includes("beehiiv")) return "beehiiv";
  if (f.includes("convertkit") || f.includes("kit")) return "Kit";
  return null;
}

const NO = new Set(["false", "no", "0", "n"]);
const YES = new Set(["true", "yes", "1", "y"]);
// Deny-list: a generic CSV may use "status" for something else (lead, customer).
const NOT_SUBSCRIBED = new Set([
  "unsubscribed", "inactive", "cancelled", "canceled", "bounced", "cleaned",
  "complained", "spam", "invalid", "pending", "validating", "needs_attention",
  "unconfirmed", "archived",
]);

/** Why a row is left out, or null to import it. */
export function exclusionReason(row: Row, emailHeader: string): string | null {
  const email = (row[emailHeader] ?? "").trim();
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) return "invalid email";
  for (const [key, raw] of Object.entries(row)) {
    const k = norm(key);
    const v = (raw ?? "").trim().toLowerCase();
    if (k === "status" && NOT_SUBSCRIBED.has(v)) return `status "${v}"`;
    if ((k === "emaildisabled" || k === "unsubscribed") && YES.has(v)) return "unsubscribed";
    if (k === "subscribedtoemails" && NO.has(v)) return "unsubscribed";
    if (k === "unsubscribedat" || k === "unsubtime") if (v) return "unsubscribed";
  }
  return null;
}

/**
 * Splits rows into the ones to import (with a derived first/last name when the
 * export only has "name") and a count of the rest by reason. Duplicate emails
 * keep their first row.
 */
export function prepareImport(rows: Row[], emailHeader: string) {
  const keep: Row[] = [];
  const skipped: Record<string, number> = {};
  const seen = new Set<string>();
  const nameHeader = Object.keys(rows[0] ?? {}).find((h) => norm(h) === "name");
  for (const row of rows) {
    let reason = exclusionReason(row, emailHeader);
    const email = (row[emailHeader] ?? "").trim().toLowerCase();
    if (!reason && seen.has(email)) reason = "duplicate";
    if (reason) {
      skipped[reason] = (skipped[reason] ?? 0) + 1;
      continue;
    }
    seen.add(email);
    const out = { ...row, [emailHeader]: email };
    if (nameHeader && !("first_name" in row)) {
      const [first = "", ...rest] = (row[nameHeader] ?? "").trim().split(/\s+/);
      out.first_name = first;
      out.last_name = rest.join(" ");
    }
    keep.push(out);
  }
  return { keep, skipped };
}
