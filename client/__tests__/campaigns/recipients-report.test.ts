jest.mock("@/hooks/use-api", () => ({ useApi: () => ({}) }));
import { audienceName } from "@/components/campaigns/recipients-report";

test("a saved view is named after the outcome and campaign", () => {
  expect(audienceName("Launch", "not_opened")).toBe("Launch: didn't open");
  expect(audienceName("Launch", "clicked", "https://x.com/pricing")).toBe(
    "Launch: clicked x.com/pricing",
  );
  expect(audienceName("Launch", "", "", " ann ")).toBe(
    "Launch: everyone matching “ann”",
  );
});

test("a saved view's name fits a segment name", () => {
  const name = audienceName("x".repeat(300), "clicked", "https://x.com/" + "p".repeat(40));
  expect(name.length).toBeLessThanOrEqual(100);
  expect(name).toContain("…");
});
