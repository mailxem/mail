jest.mock("@/app/providers/team-provider", () => ({ useTeam: () => ({}) }));
jest.mock("@/app/lib/store/billing", () => ({ useSubscription: () => ({}) }));
import { isPaidSubscription } from "@/components/billing/paid-plan";

test("only a real, active, non-Free plan counts as paid", () => {
  expect(isPaidSubscription({ plan_id: "xem-starter", status: "active" })).toBe(true);
  expect(isPaidSubscription({ plan_id: "xem-free", status: "active" })).toBe(false);
  expect(isPaidSubscription({ plan_id: "", status: "active" })).toBe(false); // a just-created Free row
  expect(isPaidSubscription({ status: "active" })).toBe(false);
  expect(isPaidSubscription({ plan_id: "xem-growth", status: "pending" })).toBe(false);
  expect(isPaidSubscription(null)).toBe(false);
});
