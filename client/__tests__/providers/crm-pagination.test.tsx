import React from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { CRMPage } from "@/components/marketing/crm";
import { useMarketingQuery } from "@/lib/marketing/api";

jest.mock("@/lib/marketing/api", () => ({ useMarketingQuery: jest.fn() }));
const query = jest.mocked(useMarketingQuery);

test("CRM renders the API page and total without slicing the returned rows again", () => {
  query.mockReturnValue({
    data: {
      data: [{ id: "last-contact", email: "last@example.com", firstName: "Last", lastName: "Contact", company: "", listName: "Community", status: "ACTIVE", lifecycleStage: "LEAD" }],
      total: 201, page: 21, limit: 10,
      summary: { total: 1500, qualified: 420, customers: 350, subscribed: 1200 },
    },
    isPending: false, error: null,
  } as ReturnType<typeof useMarketingQuery>);
  const html = renderToStaticMarkup(<CRMPage />);
  expect(query).toHaveBeenCalledWith("marketing/contacts?page=1&limit=10", true);
  expect(html).toContain("last@example.com");
  expect(html).toContain("Community");
  expect(html).toContain("201–201 of 201");
  expect(html).toContain("Page 21 of 21");
  expect(html).toContain("1,500");
  expect(html).toContain("420");
  expect(html).toContain("350");
  expect(html).toContain("1,200");
});

test("failed CRM requests do not present zero contacts as a successful empty result", () => {
  query.mockReturnValue({ data: undefined, isPending: false, error: new Error("Could not load contacts"), refetch: jest.fn() } as unknown as ReturnType<typeof useMarketingQuery>);
  const html = renderToStaticMarkup(<CRMPage />);
  expect(html).toContain("Could not load contacts");
  expect(html).not.toContain("0 results");
  expect(html).not.toContain("Your people are out there");
});
