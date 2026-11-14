/** @jest-environment jsdom */
import * as React from "react";
import { act } from "react";
import { createRoot, Root } from "react-dom/client";
import { SummaryEmailSettings } from "@/components/settings/summary-email-settings";

const mockRequest = jest.fn();
const mockRefetch = jest.fn();
let data = { frequency: "weekly", includeMe: false, isOwner: true, timezone: "Asia/Kolkata" };
jest.mock("@/lib/marketing/api", () => ({
  useMarketing: () => ({ request: mockRequest }),
  useMarketingQuery: () => ({ data, error: null, isPending: false, refetch: mockRefetch }),
}));
jest.mock("sonner", () => ({ toast: { success: jest.fn(), error: jest.fn() } }));

let root: Root;
let container: HTMLDivElement;
beforeAll(() => {
  (globalThis as any).IS_REACT_ACT_ENVIRONMENT = true;
});
beforeEach(() => {
  jest.clearAllMocks();
  data = { frequency: "weekly", includeMe: false, isOwner: true, timezone: "Asia/Kolkata" };
  container = document.createElement("div");
  document.body.appendChild(container);
  root = createRoot(container);
});
afterEach(() => {
  act(() => root.unmount());
  container.remove();
});

it("saves the frequency and keeps the admin copy setting", async () => {
  act(() => root.render(<SummaryEmailSettings />));
  expect(container.textContent).toContain("8:00 in Asia/Kolkata");
  expect(container.textContent).not.toContain("Send me a copy too");
  const select = container.querySelector("select")!;
  await act(async () => {
    select.value = "daily";
    select.dispatchEvent(new Event("change", { bubbles: true }));
  });
  expect(mockRequest).toHaveBeenCalledWith("marketing/summary-email", "PUT", { frequency: "daily", includeMe: false });
});

it("offers admins a copy and previews a quiet period", async () => {
  data = { ...data, isOwner: false };
  mockRequest.mockResolvedValueOnce({ subject: "Your weekly summary for Studio", html: "<p>Hi</p>", quiet: true });
  act(() => root.render(<SummaryEmailSettings />));
  expect(container.textContent).toContain("Send me a copy too");
  const preview = [...container.querySelectorAll("button")].find((b) => b.textContent === "Preview")!;
  await act(async () => preview.click());
  expect(mockRequest).toHaveBeenCalledWith("marketing/summary-email/preview?frequency=weekly");
  expect(container.textContent).toContain("Your weekly summary for Studio");
  expect(container.textContent).toContain("would be skipped");
  expect(container.querySelector("iframe")?.getAttribute("sandbox")).toBe("");
});
