/** @jest-environment jsdom */
import * as React from "react";
import { act } from "react";
import { createRoot, Root } from "react-dom/client";
import { CloudflareSetupSheet } from "@/components/settings/cloudflare-setup";

const mockRequest = jest.fn();
const mockRefresh = jest.fn();
let mockPaid = true;
jest.mock("@/lib/marketing/api", () => ({
  useMarketing: () => ({ request: mockRequest, refresh: mockRefresh }),
}));
jest.mock("@/components/billing/paid-plan", () => ({
  usePaidPlan: () => mockPaid,
  PaidPlanNotice: ({ feature }: { feature: string }) => <p>{feature} comes with a paid plan.</p>,
}));
jest.mock("@/components/ui/sheet", () => {
  const Pass = ({ children }: { children?: React.ReactNode }) => <div>{children}</div>;
  return { Sheet: ({ open, children }: { open: boolean; children: React.ReactNode }) => (open ? <div>{children}</div> : null), SheetPopup: Pass, SheetHeader: Pass, SheetTitle: Pass, SheetDescription: Pass, SheetPanel: Pass, SheetFooter: Pass };
});
jest.mock("sonner", () => ({
  toast: { success: jest.fn(), error: jest.fn() },
}));

let root: Root;
let container: HTMLDivElement;
beforeAll(() => {
  (globalThis as any).IS_REACT_ACT_ENVIRONMENT = true;
});
beforeEach(() => {
  jest.clearAllMocks();
  mockPaid = true;
  container = document.createElement("div");
  document.body.appendChild(container);
  root = createRoot(container);
});
afterEach(async () => {
  await act(async () => root.unmount());
  container.remove();
});

async function fill(values: Record<string, string>) {
  for (const [id, value] of Object.entries(values)) {
    const input = container.querySelector(`#${id}`) as HTMLInputElement;
    await act(async () => {
      Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, "value")!.set!.call(input, value);
      input.dispatchEvent(new Event("input", { bubbles: true }));
    });
  }
  await act(async () =>
    container.querySelector("form")!.dispatchEvent(new Event("submit", { bubbles: true, cancelable: true })),
  );
}

test("connects a customer-owned Worker and closes the sheet", async () => {
  const onOpenChange = jest.fn();
  mockRequest.mockResolvedValue({});
  await act(async () => root.render(<CloudflareSetupSheet mode="receive" open onOpenChange={onOpenChange} />));
  expect(container.textContent).toContain("stored encrypted");
  await fill({
    "cf-worker-url": "https://mailbox.example.workers.dev",
    "cf-inbox-address": "support@example.com",
    "cf-inbox-secret": "preview-secret-1234567890-abcdefghij",
  });
  expect(mockRequest).toHaveBeenCalledWith("mail-connections/cloudflare-relays", "POST", {
    workerUrl: "https://mailbox.example.workers.dev",
    address: "support@example.com",
    secret: "preview-secret-1234567890-abcdefghij",
  });
  expect(mockRefresh).toHaveBeenCalled();
  expect(onOpenChange).toHaveBeenCalledWith(false);
});

test("saves a Cloudflare sender", async () => {
  mockRequest.mockResolvedValue({});
  await act(async () => root.render(<CloudflareSetupSheet mode="send" open onOpenChange={jest.fn()} />));
  await fill({
    "cf-sender-address": "notifications@example.com",
    "cf-account-id": "0123456789abcdef0123456789abcdef",
    "cf-token": "token-1234567890abcdef",
  });
  expect(mockRequest).toHaveBeenCalledWith("mail-connections/cloudflare", "POST", {
    address: "notifications@example.com",
    accountId: "0123456789abcdef0123456789abcdef",
    token: "token-1234567890abcdef",
  });
});

test("shows the upgrade prompt instead of the form on Free", async () => {
  mockPaid = false;
  await act(async () => root.render(<CloudflareSetupSheet mode="receive" open onOpenChange={jest.fn()} />));
  expect(container.querySelector("form")).toBeNull();
  expect(container.textContent).toContain("comes with a paid plan");
});
