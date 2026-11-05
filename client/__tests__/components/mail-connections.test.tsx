/** @jest-environment jsdom */
import * as React from "react";
import { act } from "react";
import { createRoot, Root } from "react-dom/client";
import { MailConnections } from "@/components/settings/mail-connections";

const mockRequest = jest.fn();
const mockRefresh = jest.fn();
const mockRefetch = jest.fn();
let query = {
  data: {
    connections: [] as Array<{
      id: string;
      provider: string;
      address: string;
      smtpConfigId: string;
      imapConfigId?: string;
    }>,
    googleConfigured: true,
    googleAvailable: true,
  },
  error: null as Error | null,
};
jest.mock("@/lib/marketing/api", () => ({
  useMarketing: () => ({ request: mockRequest, refresh: mockRefresh }),
  useMarketingQuery: () => ({
    ...query,
    isLoading: false,
    refetch: mockRefetch,
  }),
}));
jest.mock("@/components/settings/cloudflare-relay", () => ({
  CloudflareRelays: () => null,
}));
jest.mock("@/components/billing/paid-plan", () => ({
  usePaidPlan: () => true,
  PaidPlanNotice: () => null,
}));
jest.mock("@/components/settings/cloudflare-setup", () => ({
  CloudflareButton: () => null,
  CloudflareLogo: () => null,
  CloudflareSetupSheet: () => null,
}));
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
  query = {
    data: { connections: [], googleConfigured: true, googleAvailable: true },
    error: null,
  };
  container = document.createElement("div");
  document.body.appendChild(container);
  root = createRoot(container);
});
afterEach(async () => {
  await act(async () => root.unmount());
  container.remove();
});
async function render() {
  await act(async () => root.render(<MailConnections provider="google" />));
}
function button(label: string) {
  const item = Array.from(container.querySelectorAll("button")).find((node) =>
    node.textContent?.includes(label),
  );
  if (!item) throw new Error(`Missing button ${label}`);
  return item as HTMLButtonElement;
}

test("explains unavailable Google OAuth and disables connection", async () => {
  query.data.googleConfigured = false;
  query.data.googleAvailable = false;
  await render();
  expect(button("Connect Google mailbox").disabled).toBe(true);
  expect(container.textContent).toContain(
    "administrator needs to configure Google mailbox OAuth",
  );
});

test("starts Google authorization once and exposes the pending state", async () => {
  mockRequest.mockImplementation(() => new Promise(() => {}));
  await render();
  await act(async () => button("Connect Google mailbox").click());
  expect(mockRequest).toHaveBeenCalledWith(
    "mail-connections/google/start",
    "POST",
  );
  expect(button("Opening Google").disabled).toBe(true);
});

test("supports canceling and confirming a connected mailbox disconnect", async () => {
  query.data.connections = [
    {
      id: "google-1",
      provider: "GOOGLE_OAUTH",
      address: "alex@example.com",
      smtpConfigId: "sender-1",
      imapConfigId: "mailbox-1",
    },
  ];
  mockRequest.mockResolvedValue(undefined);
  mockRefresh.mockResolvedValue(undefined);
  await render();
  expect(container.textContent).toContain("Connect another Google mailbox");
  await act(async () => button("Disconnect").click());
  expect(container.textContent).toContain("Disconnect this inbox and sender?");
  await act(async () => button("Cancel").click());
  expect(container.textContent).not.toContain(
    "Disconnect this inbox and sender?",
  );
  await act(async () => button("Disconnect").click());
  const confirmations = Array.from(container.querySelectorAll("button")).filter(
    (node) => node.textContent === "Disconnect",
  );
  await act(async () => (confirmations.at(-1) as HTMLButtonElement).click());
  expect(mockRequest).toHaveBeenCalledWith(
    "mail-connections/google-1",
    "DELETE",
  );
  expect(mockRefresh).toHaveBeenCalled();
});

test("offers retry when mailbox connections fail to load", async () => {
  query.error = new Error("Temporary failure");
  await render();
  await act(async () => button("Retry").click());
  expect(mockRefetch).toHaveBeenCalled();
});
