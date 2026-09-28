/** @jest-environment jsdom */
import * as React from "react";
import { randomUUID } from "node:crypto";
import { act } from "react";
import { createRoot, Root } from "react-dom/client";
import { FormJourneyPanel } from "@/components/marketing/form-journey";
import { MarketingRequestError } from "@/lib/marketing/api";
const mockRequest = jest.fn();
jest.mock("@/lib/marketing/api", () => ({
  MarketingRequestError: class MarketingRequestError extends Error { constructor(message: string, public status: number) { super(message); } },
  useMarketing: () => ({ request: mockRequest, refresh: async () => {} }),
  useMarketingQuery: () => ({ data: { senders: [{ id: "sender", fromEmail: "hello@example.com" }] } }),
}));
jest.mock("@/components/marketing/shared", () => ({ Field: ({ children, label }: { children: React.ReactNode; label: string }) => <label>{label}{children}</label> }));
let root: Root; let container: HTMLDivElement;
beforeAll(() => { (globalThis as any).IS_REACT_ACT_ENVIRONMENT = true; Object.defineProperty(crypto, "randomUUID", { configurable: true, value: randomUUID }); });
beforeEach(async () => { mockRequest.mockReset(); container = document.createElement("div"); document.body.appendChild(container); root = createRoot(container); await act(async () => root.render(<FormJourneyPanel formId="form" formName="Signup" />)); });
afterEach(async () => { await act(async () => root.unmount()); container.remove(); });
async function submit() { await act(async () => container.querySelector("form")!.dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }))); }
test("a definite validation rejection unlocks the draft for correction", async () => {
  mockRequest.mockRejectedValueOnce(new MarketingRequestError("Choose a valid sender", 400));
  await submit();
  expect(container.querySelector("fieldset")?.disabled).toBe(false);
  expect(container.textContent).toContain("Choose a valid sender");
  expect(container.textContent).not.toContain("Retry saving draft");
});
test("an uncertain failure retains the immutable draft and request ID for retry", async () => {
  mockRequest.mockRejectedValueOnce(new MarketingRequestError("Temporary error", 503)).mockResolvedValueOnce({});
  await submit();
  expect(container.querySelector("fieldset")?.disabled).toBe(true);
  await submit();
  expect(mockRequest.mock.calls[1][2]).toEqual(mockRequest.mock.calls[0][2]);
  expect(container.textContent).toContain("Your journey is ready to review");
});
