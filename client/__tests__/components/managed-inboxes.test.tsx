/** @jest-environment jsdom */
import * as React from "react";
import { act } from "react";
import { createRoot, Root } from "react-dom/client";
import { ManagedInboxes } from "@/components/settings/managed-inboxes";

const mockRequest = jest.fn();
const mockRefresh = jest.fn();
const mockRefetch = jest.fn();
let result: any;
jest.mock("@/lib/marketing/api", () => ({
  MarketingRequestError: class MarketingRequestError extends Error {
    constructor(
      message: string,
      public status: number,
    ) {
      super(message);
    }
  },
  useMarketing: () => ({ request: mockRequest, refresh: mockRefresh }),
  useMarketingQuery: () => result,
}));
jest.mock("sonner", () => ({
  toast: { success: jest.fn(), error: jest.fn() },
}));

const receiving = {
  enabled: true,
  region: "test-1",
  maxMailboxesPerDomain: 10,
  maxMessageBytes: 1024,
  domains: [
    {
      id: "domain-1",
      name: "example.com",
      sendingReady: false,
      receivingStatus: "mx_conflict",
      mx: {
        name: "example.com",
        type: "MX",
        value: "inbound.sample.invalid",
        priority: 10,
      },
      existingMX: [{ host: "mx.provider.example", priority: 10 }],
    },
  ],
  mailboxes: [
    {
      id: "box-1",
      domainId: "domain-1",
      address: "alex@example.com",
      displayName: "Alex",
      active: false,
      status: "inactive",
      createdAt: "2026-10-01T00:00:00Z",
    },
  ],
};
let root: Root;
let container: HTMLDivElement;
beforeAll(() => {
  (globalThis as any).IS_REACT_ACT_ENVIRONMENT = true;
});
beforeEach(() => {
  jest.clearAllMocks();
  result = {
    data: receiving,
    isLoading: false,
    error: null,
    refetch: mockRefetch,
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
  await act(async () => root.render(<ManagedInboxes />));
}
function button(label: string) {
  const found = Array.from(container.querySelectorAll("button")).find((node) =>
    node.textContent?.includes(label),
  );
  if (!found) throw new Error(`Missing button ${label}`);
  return found as HTMLButtonElement;
}

test("shows receiving and sending independently with honest MX migration guidance", async () => {
  await render();
  expect(container.textContent).toContain("Receiving: setup needed");
  expect(container.textContent).toContain("Sending: not ready");
  expect(container.textContent).toContain("cannot split or mirror delivery");
  expect(container.textContent).toContain("separate from sending bounce");
  expect(container.textContent).toContain("Host");
  expect(container.textContent).toContain("Priority");
  expect(container.textContent).toContain("Value");
  expect(container.textContent).toContain("Full address: mailbox@example.com");
  expect(container.textContent).toContain(
    "Paused. Existing mail is preserved.",
  );
});

test("shows mailbox quota and does not promise an automatic storage retry", async () => {
  result = {
    data: {
      ...receiving,
      mailboxes: [
        {
          ...receiving.mailboxes[0],
          active: true,
          status: "storage_full",
          usageBytes: 104857600,
          quotaBytes: 104857600,
        },
      ],
    },
    isLoading: false,
    error: null,
    refetch: mockRefetch,
  };
  await render();
  expect(container.textContent).toContain(
    "Storage full. New mail cannot be received",
  );
  expect(container.textContent).toContain("Storage: 100.0 MB of 100.0 MB used");
  expect(container.textContent).not.toContain("automatically retry");
});

test("normalizes a valid local part and resumes without deleting mail", async () => {
  mockRequest.mockResolvedValue({});
  mockRefresh.mockResolvedValue(undefined);
  await render();
  const input = container.querySelector(
    "#managed-local-part",
  ) as HTMLInputElement;
  await act(async () => {
    Object.getOwnPropertyDescriptor(
      HTMLInputElement.prototype,
      "value",
    )?.set?.call(input, "  Alex  ");
    input.dispatchEvent(new Event("input", { bubbles: true }));
  });
  await act(async () => button("Create mailbox").click());
  expect(mockRequest).toHaveBeenCalledWith(
    "mail-connections/receiving/mailboxes",
    "POST",
    { domainId: "domain-1", localPart: "alex" },
  );
  await act(async () => button("Resume").click());
  expect(mockRequest).toHaveBeenCalledWith(
    "mail-connections/receiving/mailboxes/box-1",
    "PATCH",
    { active: true },
  );
});

test("renders disabled and retry states", async () => {
  result = {
    data: { ...receiving, enabled: false },
    isLoading: false,
    error: null,
    refetch: mockRefetch,
  };
  await render();
  expect(container.textContent).toContain("unavailable on this installation");
  result = {
    data: undefined,
    isLoading: false,
    error: new Error("offline"),
    refetch: mockRefetch,
  };
  await act(async () => root.render(<ManagedInboxes />));
  await act(async () => button("Retry").click());
  expect(mockRefetch).toHaveBeenCalled();
});

test("rechecks receiving DNS", async () => {
  mockRequest.mockResolvedValue({});
  mockRefresh.mockResolvedValue(undefined);
  await render();
  await act(async () => button("Check DNS").click());
  expect(mockRequest).toHaveBeenCalledWith(
    "mail-connections/receiving/domains/domain-1/check",
    "POST",
  );
});

test("keeps form values when mailbox creation fails", async () => {
  mockRequest.mockRejectedValue(new Error("Try again"));
  await render();
  const input = container.querySelector(
    "#managed-local-part",
  ) as HTMLInputElement;
  await act(async () => {
    Object.getOwnPropertyDescriptor(
      HTMLInputElement.prototype,
      "value",
    )?.set?.call(input, "alex");
    input.dispatchEvent(new Event("input", { bubbles: true }));
  });
  await act(async () => button("Create mailbox").click());
  expect(input.value).toBe("alex");
});
