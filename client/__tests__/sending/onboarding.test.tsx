import React from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { Onboarding } from "@/components/sending/onboarding";
import { useMarketingQuery } from "@/lib/marketing/api";
import { state } from "@/test/fixtures/sending";

jest.mock("@/lib/marketing/api", () => ({
  useMarketingQuery: jest.fn(),
  useMarketing: () => ({ request: jest.fn(), refresh: jest.fn() }),
}));
jest.mock("@/hooks/use-api", () => ({ useApi: () => ({}) }));

function render(s: ReturnType<typeof state>) {
  jest
    .mocked(useMarketingQuery)
    .mockReturnValue({ data: s, isLoading: false } as ReturnType<
      typeof useMarketingQuery
    >);
  return renderToStaticMarkup(<Onboarding />);
}

test("approval pending shows the DNS stage and locks sender navigation", () => {
  const s = state();
  s.account.approved = false;
  Object.assign(s.domains[0], {
    ready: false,
    provisioned: false,
    identityStatus: "AWAITING_APPROVAL",
    records: [
      { type: "TXT", name: "_xem.example.com", value: "ownership-token" },
    ],
  });
  const html = render(s);
  expect(html).toContain("2. Add and verify delivery records");
  expect(html).toContain("Contact your Xem operator");
  expect(html).toContain("ownership-token");
  expect(html).toMatch(/<button[^>]*disabled=""[^>]*>Complete DNS setup first/);
  expect(html).toMatch(
    /<button[^>]*disabled=""[^>]*>[\s\S]*?Send yourself a little hello/,
  );
  expect(html).not.toContain("The address people will see");
});

test("provisioned domains keep DKIM, bounce and DMARC guidance on the DNS step", () => {
  const s = state();
  Object.assign(s.domains[0], {
    ready: false,
    dmarcStatus: "MISSING_OR_INVALID",
    records: [
      { type: "TXT", name: "_xem.example.com", value: "ownership-token" },
      {
        type: "CNAME",
        name: "one._domainkey.example.com",
        value: "one.dkim.amazonses.com",
      },
      {
        type: "MX",
        name: "bounce.example.com",
        value: "10 feedback-smtp.us-east-1.amazonses.com",
      },
      {
        type: "TXT",
        name: "bounce.example.com",
        value: "v=spf1 include:amazonses.com ~all",
      },
    ],
  });
  const html = render(s);
  for (const value of [
    "one.dkim.amazonses.com",
    "bounce.example.com",
    "v=spf1",
    "_dmarc.example.com",
    "v=DMARC1; p=none",
  ])
    expect(html).toContain(value);
  expect(html).toContain("Complete DNS setup first");
  expect(html).not.toContain("The address people will see");
});

test("ready DNS resumes at sender setup; accepted test resumes at campaign", () => {
  const s = state();
  expect(render(s)).toContain("The address people will see");
  s.journey.testedDomainIds = [s.domains[0].id];
  expect(render(s)).toContain("Create my first campaign");
  s.journey.campaigns = 1;
  expect(render(s)).toContain("Your setup is complete");
});
