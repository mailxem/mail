jest.mock("@/hooks/use-api", () => ({ useApi: () => ({}) }));
import { contactVariables, parseRecipients } from "@/components/campaigns/send-test";

test("test recipients are deduplicated and checked", () => {
  const r = parseRecipients("Me@x.com, me@x.com\nboss@y.org  nope");
  expect(r.list).toEqual(["me@x.com", "boss@y.org", "nope"]);
  expect(r.invalid).toEqual(["nope"]);
});

test("preview-as fills the same fields a campaign send does", () => {
  const v = contactVariables({ id: "1", email: "ana@x.com", firstName: "Ana", lastName: "Lima", city: "Porto" });
  expect(v).toMatchObject({ email: "ana@x.com", first_name: "Ana", name: "Ana Lima", full_name: "Ana Lima", city: "Porto" });
  expect(contactVariables(undefined)).toEqual({});
});
