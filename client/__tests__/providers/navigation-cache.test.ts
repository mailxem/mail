import React from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { TemplatesProvider, useTemplates } from "@/app/providers/templates-provider";
import { MailingListProvider, useMailingLists } from "@/app/providers/mailinglist-provider";
import { CampaignsProvider, useCampaigns } from "@/app/providers/campaigns-provider";
import { resourceEntity } from "@/lib/resource-response";

jest.mock("@/app/providers/team-provider", () => ({ useTeam: () => ({ team: { id: "team-a" }, error: null }) }));
jest.mock("@/hooks/use-api", () => ({ useApi: () => ({ session: { accessToken: "test-only" }, apiFetch: jest.fn() }) }));

describe("navigation back to a cached collection", () => {
  const scenarios = [
    { name: "templates", Provider: TemplatesProvider, key: ["templates", "team-a", 1, 50, {}], read: () => useTemplates().templates },
    { name: "contact lists", Provider: MailingListProvider, key: ["mailing-lists", "team-a", 1, 20, { sort: "subscribers_count", order: "desc" }], read: () => useMailingLists().lists },
    { name: "campaigns", Provider: CampaignsProvider, key: ["campaigns", "team-a", 1, 20, {}], read: () => useCampaigns().campaigns },
  ];
  for (const scenario of scenarios) test(scenario.name, () => {
    const client = new QueryClient();
    client.setQueryData(scenario.key, { data: [{ id: "existing-record", name: "Existing record" }], total: 1, page: 1, limit: 20 });
    function Page() { return React.createElement("div", null, scenario.read().map(item => item.name).join(",")); }
    const render = () => renderToStaticMarkup(React.createElement(QueryClientProvider, { client }, React.createElement(scenario.Provider, { children: React.createElement(Page) })));
    expect(render()).toContain("Existing record");
    expect(render()).toContain("Existing record");
    client.clear();
  });
});

test("single-record responses never turn a successful list load into undefined query data", () => {
  const record = { id: "list-id", name: "All Users" };
  expect(resourceEntity(record)).toEqual(record);
  expect(resourceEntity({ data: record })).toEqual(record);
  expect(() => resourceEntity({ data: undefined })).toThrow("invalid record");
  expect(() => resourceEntity(null)).toThrow("invalid record");
});
