let mockSubscription: { data?: { plan_id?: string; status?: string } | null; isLoading: boolean } = { isLoading: false };
const mockUseSubscription = jest.fn((_teamId: string) => mockSubscription);
jest.mock("@/app/providers/team-provider", () => ({ useTeam: () => ({ team: { id: "team-1" } }) }));
jest.mock("@/app/lib/store/billing", () => ({ useSubscription: (teamId: string) => mockUseSubscription(teamId) }));
import { isPaidSubscription, usePaidPlan } from "@/components/billing/paid-plan";

const originalPaywall = process.env.NEXT_PUBLIC_PAYWALL_URL;
afterEach(() => {
  if (originalPaywall === undefined) delete process.env.NEXT_PUBLIC_PAYWALL_URL;
  else process.env.NEXT_PUBLIC_PAYWALL_URL = originalPaywall;
  mockUseSubscription.mockClear();
});

test("only a real, active, non-Free plan counts as paid", () => {
  expect(isPaidSubscription({ plan_id: "xem-starter", status: "active" })).toBe(true);
  expect(isPaidSubscription({ plan_id: "xem-free", status: "active" })).toBe(false);
  expect(isPaidSubscription({ plan_id: "", status: "active" })).toBe(false); // a just-created Free row
  expect(isPaidSubscription({ status: "active" })).toBe(false);
  expect(isPaidSubscription({ plan_id: "xem-growth", status: "pending" })).toBe(false);
  expect(isPaidSubscription(null)).toBe(false);
});

test.each([undefined, "", "  "])("self-hosted builds without billing are never gated (%p)", (value) => {
  if (value === undefined) delete process.env.NEXT_PUBLIC_PAYWALL_URL;
  else process.env.NEXT_PUBLIC_PAYWALL_URL = value;
  // What a self-hosted build sees: the billing request fails, nothing loads.
  mockSubscription = { data: undefined, isLoading: false };
  expect(usePaidPlan()).toBe(true);
  expect(mockUseSubscription).toHaveBeenCalledWith(""); // query disabled, no billing request
});

test("hosted builds still gate Free workspaces", () => {
  process.env.NEXT_PUBLIC_PAYWALL_URL = "https://payments.example.test/api/v1";
  mockSubscription = { data: { plan_id: "xem-free", status: "active" }, isLoading: false };
  expect(usePaidPlan()).toBe(false);
  expect(mockUseSubscription).toHaveBeenCalledWith("team-1");
  mockSubscription = { data: { plan_id: "xem-starter", status: "active" }, isLoading: false };
  expect(usePaidPlan()).toBe(true);
});
