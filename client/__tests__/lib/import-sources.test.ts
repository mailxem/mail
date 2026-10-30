import {
  detectSource,
  guessField,
  prepareImport,
} from "@/lib/contacts/import-sources";

test("recognises common newsletter exports", () => {
  expect(detectSource(["email", "active_subscription", "email_disabled"])).toBe("Substack");
  expect(detectSource(["email", "status", "subscription_tier", "utm_source"])).toBe("beehiiv");
  expect(detectSource(["Email Address", "First Name", "MEMBER_RATING"])).toBe("Mailchimp");
  expect(detectSource(["email", "name", "subscribed_to_emails"])).toBe("Ghost");
  expect(detectSource(["email"], "contacts.csv")).toBeNull();
});

test("maps headers to the fields the import worker reads", () => {
  expect(guessField("Email Address")).toBe("email");
  expect(guessField("First Name")).toBe("first_name");
  expect(guessField("Postal Code")).toBe("zip");
  // Substring matches used to map "Last Changed" to last_name.
  expect(guessField("LAST_CHANGED")).toBeNull();
});

test("keeps only people who are still subscribed", () => {
  const { keep, skipped } = prepareImport(
    [
      { email: "Ana@Example.com", status: "active" },
      { email: "ben@example.com", status: "unsubscribed" },
      { email: "cy@example.com", email_disabled: "true" },
      { email: "dee@example.com", subscribed_to_emails: "false" },
      { email: "not-an-email", status: "active" },
      { email: "ana@example.com", status: "active" },
      { email: "eve@example.com", status: "customer" },
    ],
    "email",
  );
  expect(keep.map((r) => r.email)).toEqual(["ana@example.com", "eve@example.com"]);
  expect(skipped).toEqual({
    'status "unsubscribed"': 1,
    unsubscribed: 2,
    "invalid email": 1,
    duplicate: 1,
  });
});

test("splits a single name column", () => {
  const { keep } = prepareImport([{ email: "a@b.co", name: "Mei Ling Chen" }], "email");
  expect(keep[0]).toMatchObject({ first_name: "Mei", last_name: "Ling Chen" });
});
