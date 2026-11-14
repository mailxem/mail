/** @jest-environment jsdom */
import * as React from "react";
import { act } from "react";
import { createRoot, Root } from "react-dom/client";
import { AlertsBell, type WorkspaceAlert } from "@/components/alerts-bell";

const mockRequest = jest.fn();
const mockRefetch = jest.fn();
let mockData: { alerts: WorkspaceAlert[]; unread: number } | undefined;

jest.mock("@/lib/marketing/api", () => ({
  useMarketing: () => ({ request: mockRequest }),
  useMarketingQuery: () => ({ data: mockData, isError: false, refetch: mockRefetch }),
}));
jest.mock("next/link", () => ({
  __esModule: true,
  default: ({ children, href, onClick, className }: any) => (
    <a href={href} onClick={(e) => { e.preventDefault(); onClick?.(e); }} className={className}>{children}</a>
  ),
}));

const alert = (values: Partial<WorkspaceAlert>): WorkspaceAlert => ({
  id: "a1",
  kind: "sending_paused",
  title: "A campaign was paused to protect your reputation",
  reason: "Paused after 9 of 100 emails bounced",
  count: 1,
  action: "Resume campaign",
  href: "/campaigns/c1",
  resolved: false,
  readAt: null,
  createdAt: new Date(Date.now() - 5 * 60_000).toISOString(),
  ...values,
});

let root: Root;
let container: HTMLDivElement;
beforeAll(() => {
  (globalThis as any).IS_REACT_ACT_ENVIRONMENT = true;
});
beforeEach(() => {
  jest.clearAllMocks();
  mockRequest.mockResolvedValue(null);
  container = document.createElement("div");
  document.body.appendChild(container);
  root = createRoot(container);
});
afterEach(async () => {
  await act(async () => root.unmount());
  container.remove();
});

const open = async () => {
  await act(async () => root.render(<AlertsBell label />));
  const trigger = container.querySelector("button")!;
  await act(async () => trigger.click());
  return trigger;
};

test("shows the unread count, the fix and resolved alerts", async () => {
  mockData = {
    unread: 1,
    alerts: [
      alert({}),
      alert({ id: "a2", kind: "domain_unready", title: "Your sending domain stopped verifying", action: "Check DNS", href: "/settings/sending", resolved: true, readAt: new Date().toISOString() }),
    ],
  };
  const trigger = await open();
  expect(trigger.getAttribute("aria-label")).toBe("Alerts, 1 unread");
  const text = document.body.textContent;
  expect(text).toContain("5 minutes ago");
  expect(text).toContain("Resolved");
  const links = Array.from(document.body.querySelectorAll("a"));
  // A resolved problem has nothing left to fix.
  expect(links.map((a) => a.textContent)).toEqual(["Resume campaign"]);
  expect(links[0].getAttribute("href")).toBe("/campaigns/c1");

  await act(async () => links[0].click());
  expect(mockRequest).toHaveBeenCalledWith("marketing/alerts/read", "POST", { id: "a1" });
});

test("marks every alert read", async () => {
  mockData = { unread: 2, alerts: [alert({}), alert({ id: "a2" })] };
  await open();
  const all = Array.from(document.body.querySelectorAll("button")).find((b) => b.textContent === "Mark all as read")!;
  await act(async () => all.click());
  expect(mockRequest).toHaveBeenCalledWith("marketing/alerts/read", "POST", {});
  expect(mockRefetch).toHaveBeenCalled();
});

test("says when nothing needs attention", async () => {
  mockData = { unread: 0, alerts: [] };
  const trigger = await open();
  expect(trigger.getAttribute("aria-label")).toBe("Alerts");
  expect(document.body.textContent).toContain("Nothing needs you right now");
});
