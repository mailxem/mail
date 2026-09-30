/** @jest-environment jsdom */
import * as React from "react";
import { act } from "react";
import { createRoot, Root } from "react-dom/client";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { InboxPage, senderName } from "@/components/marketing/inbox";

const mockRequest = jest.fn();
const mockSummaryProps = jest.fn();
let mailboxBFirst = false;
let cloudflareMailbox = false;
let gmailMailbox = false;
let managedMailbox = false;
jest.mock("@/lib/marketing/api", () => ({
  PreviewTransport: React.createContext(null),
  useMarketing: () => ({ request: mockRequest, scope: "team", ready: true }),
  useMarketingQuery: (path: string) =>
    path === "mail-connections/mailboxes"
      ? {
          data: mailboxBFirst
            ? [
                {
                  id: "mailbox-b",
                  username: "b@example.com",
                  host: "imap.example.com",
                },
                {
                  id: "mailbox-a",
                  username: "a@example.com",
                  host: "imap.example.com",
                  provider: gmailMailbox
                    ? "GOOGLE_OAUTH"
                    : managedMailbox
                      ? "MANAGED"
                      : cloudflareMailbox
                        ? "CLOUDFLARE"
                        : "CUSTOM",
                },
              ]
            : [
                {
                  id: "mailbox-a",
                  username: "a@example.com",
                  host: "imap.example.com",
                  provider: gmailMailbox
                    ? "GOOGLE_OAUTH"
                    : managedMailbox
                      ? "MANAGED"
                      : cloudflareMailbox
                        ? "CLOUDFLARE"
                        : "CUSTOM",
                },
                {
                  id: "mailbox-b",
                  username: "b@example.com",
                  host: "imap.example.com",
                },
              ],
          isLoading: false,
          error: null,
        }
      : { data: [{ Name: "INBOX", Total: 2 }], isLoading: false, error: null },
}));
jest.mock("@/components/marketing/mail-compose", () => ({
  MailCompose: () => null,
}));
jest.mock("@/components/marketing/mail-summary", () => ({
  MailSummary: (props: unknown) => {
    mockSummaryProps(props);
    return <div data-testid="mail-summary" />;
  },
}));
jest.mock("sonner", () => ({ toast: { error: jest.fn() } }));

const base = {
  uid: 8,
  uidValidity: 1,
  from: "Sender <sender@example.com>",
  to: "Reader <reader@example.com>",
  cc: "",
  bcc: "",
  body: "<p>Message body</p>",
  date: "2026-09-30T09:00:00Z",
  attachments: [],
};
const messages = {
  "mailbox-a": [
    {
      ...base,
      id: "mailbox-a:INBOX:1:8",
      messageId: "a-plain",
      subject: "A plain",
      flags: [],
    },
    {
      ...base,
      uid: 9,
      id: "mailbox-a:INBOX:1:9",
      messageId: "a-star",
      subject: "A starred",
      flags: ["\\Flagged"],
    },
  ],
  "mailbox-b": [
    {
      ...base,
      id: "mailbox-b:INBOX:1:8",
      messageId: "b-plain",
      subject: "B plain",
      flags: [],
    },
  ],
};

let root: Root;
let container: HTMLDivElement;
let queryClient: QueryClient;
beforeAll(() => {
  (globalThis as any).IS_REACT_ACT_ENVIRONMENT = true;
});
beforeEach(() => {
  jest.clearAllMocks();
  mailboxBFirst = false;
  cloudflareMailbox = false;
  gmailMailbox = false;
  managedMailbox = false;
  queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0 } },
  });
  container = document.createElement("div");
  document.body.appendChild(container);
  root = createRoot(container);
});
afterEach(async () => {
  await act(async () => root.unmount());
  queryClient.clear();
  container.remove();
});

async function renderInbox() {
  await act(async () =>
    root.render(
      <QueryClientProvider client={queryClient}>
        <InboxPage />
      </QueryClientProvider>,
    ),
  );
  await settle();
}
async function settle() {
  await act(async () => {
    await new Promise((resolve) => setTimeout(resolve, 0));
  });
}
function button(label: string) {
  const found = Array.from(container.querySelectorAll("button")).find(
    (node) =>
      node.textContent?.includes(label) ||
      node.getAttribute("aria-label") === label,
  );
  if (!found) throw new Error(`Missing button ${label}`);
  return found as HTMLButtonElement;
}

test.each([
  ["alex@example.com", "alex@example.com"],
  ["<alex@example.com>", "alex@example.com"],
  ['"Alex Morgan" <alex@example.com>', "Alex Morgan"],
  ["Alex Morgan <alex@example.com>", "Alex Morgan"],
  ["", "Unknown sender"],
])("renders sender %p as %p", (value, expected) => {
  expect(senderName(value)).toBe(expected);
});

test("previous navigation follows the starred-first order shown in the list", async () => {
  mockRequest.mockImplementation(async (path: string) => {
    const config = new URLSearchParams(path.split("?")[1]).get(
      "config_id",
    ) as keyof typeof messages;
    return {
      emails: messages[config],
      total_emails: messages[config].length,
      offset: 0,
      limit: 20,
    };
  });
  await renderInbox();
  await act(async () => button("A plain").click());
  await act(async () => button("Previous message").click());
  expect(
    container.querySelector(".mail-subject-heading")?.textContent,
  ).toContain("A starred");
});

test("a stale flag completion cannot mutate the next mailbox selection", async () => {
  let finishPatch!: () => void;
  mockRequest.mockImplementation((path: string, method = "GET") => {
    if (method === "PATCH")
      return new Promise<void>((resolve) => {
        finishPatch = resolve;
      });
    const config = new URLSearchParams(path.split("?")[1]).get(
      "config_id",
    ) as keyof typeof messages;
    return Promise.resolve({
      emails: messages[config],
      total_emails: messages[config].length,
      offset: 0,
      limit: 20,
    });
  });
  await renderInbox();
  await act(async () => button("A plain").click());
  await act(async () => button("Mark read").click());
  mailboxBFirst = true;
  await act(async () => {
    root.render(
      <QueryClientProvider client={queryClient}>
        <InboxPage />
      </QueryClientProvider>,
    );
  });
  await settle();
  await act(async () => button("B plain").click());
  await act(async () => finishPatch());
  expect(
    container.querySelector(".mail-subject-heading")?.textContent,
  ).toContain("B plain");
  expect(button("Mark read").disabled).toBe(false);
  expect(container.querySelector('[aria-label="Mark unread"]')).toBeNull();
});

test("loads older mail with the server cursor", async () => {
  mockRequest.mockImplementation(async (path: string) => {
    const params = new URLSearchParams(path.split("?")[1]);
    if (params.get("before_uid") === "20")
      return {
        emails: [
          {
            ...base,
            uid: 19,
            id: "mailbox-a:INBOX:1:19",
            messageId: "older",
            subject: "Older message",
            flags: ["\\Seen"],
          },
        ],
        total_emails: 21,
        offset: 0,
        limit: 20,
        uidValidity: 1,
      };
    return {
      emails: Array.from({ length: 20 }, (_, index) => ({
        ...base,
        uid: 39 - index,
        id: `mailbox-a:INBOX:1:${39 - index}`,
        messageId: `page-${index}`,
        subject: `Page message ${index}`,
        flags: ["\\Seen"],
      })),
      total_emails: 21,
      offset: 0,
      limit: 20,
      uidValidity: 1,
      next_before_uid: 20,
    };
  });
  await renderInbox();
  await act(async () => button("Load more messages").click());
  await settle();
  expect(
    mockRequest.mock.calls.some(([path]) =>
      String(path).includes("before_uid=20"),
    ),
  ).toBe(true);
  expect(container.textContent).toContain("Older message");
});

test("ignores a stale Cloudflare detail response after selecting another message", async () => {
  cloudflareMailbox = true;
  const details = new Map<number, (value: unknown) => void>();
  mockRequest.mockImplementation((path: string) => {
    const params = new URLSearchParams(path.split("?")[1]);
    if (path.startsWith("imap/message?"))
      return new Promise((resolve) =>
        details.set(Number(params.get("uid")), resolve),
      );
    return Promise.resolve({
      emails: messages["mailbox-a"],
      total_emails: 2,
      offset: 0,
      limit: 20,
      uidValidity: 1,
    });
  });
  await renderInbox();
  await act(async () => button("A plain").click());
  expect(container.textContent).toContain("Loading the complete message");
  await act(async () => button("A starred").click());
  await act(async () =>
    details.get(8)?.({
      ...messages["mailbox-a"][0],
      body: "<p>Stale full body</p>",
    }),
  );
  expect(
    container.querySelector(".mail-subject-heading")?.textContent,
  ).toContain("A starred");
  expect(container.querySelector('iframe[title="Email content"]')).toBeNull();
  await act(async () =>
    details.get(9)?.({
      ...messages["mailbox-a"][1],
      body: "<p>Current full body</p>",
    }),
  );
  expect(
    container
      .querySelector('iframe[title="Email content"]')
      ?.getAttribute("srcdoc"),
  ).toContain("Current full body");
  expect(
    container.querySelector('[data-testid="mail-summary"]'),
  ).not.toBeNull();
});

test("retries a failed Cloudflare detail request before enabling reply and downloads", async () => {
  cloudflareMailbox = true;
  let attempts = 0;
  mockRequest.mockImplementation((path: string) => {
    if (path.startsWith("imap/message?")) {
      attempts += 1;
      if (attempts === 1) return Promise.reject(new Error("offline"));
      return Promise.resolve({
        ...messages["mailbox-a"][0],
        body: "<p>Complete body</p>",
        attachments: [{ Filename: "brief.txt", Data: "aGVsbG8=" }],
      });
    }
    return Promise.resolve({
      emails: messages["mailbox-a"],
      total_emails: 2,
      offset: 0,
      limit: 20,
      uidValidity: 1,
    });
  });
  await renderInbox();
  await act(async () => button("A plain").click());
  await settle();
  expect(container.textContent).toContain("Couldn’t load the complete message");
  expect(button("Reply").disabled).toBe(true);
  expect(container.querySelector('[data-testid="mail-summary"]')).toBeNull();
  await act(async () => button("Retry loading message").click());
  await settle();
  expect(button("Reply").disabled).toBe(false);
  expect(
    container.querySelector('[data-testid="mail-summary"]'),
  ).not.toBeNull();
  const download = container.querySelector('a[download="brief.txt"]');
  expect(download?.getAttribute("href")).toBe(
    "data:application/octet-stream;base64,aGVsbG8=",
  );
});

test("uses numeric detail and flag payloads for a managed mailbox", async () => {
  managedMailbox = true;
  const click = jest
    .spyOn(HTMLAnchorElement.prototype, "click")
    .mockImplementation(() => {});
  mockRequest.mockImplementation((path: string, method = "GET") => {
    if (path.startsWith("imap/attachment?"))
      return Promise.resolve({ Data: "aGVsbG8=" });
    if (path.startsWith("imap/message?"))
      return Promise.resolve({
        ...messages["mailbox-a"][0],
        body: "<p>Managed full body</p>",
        attachments: [
          {
            AttachmentID: "managed-attachment",
            Filename: "brief.txt",
            MIMEType: "text/plain",
            Size: 5,
          },
        ],
      });
    if (method === "PATCH") return Promise.resolve(undefined);
    return Promise.resolve({
      emails: messages["mailbox-a"],
      total_emails: 2,
      offset: 0,
      limit: 20,
      uidValidity: 1,
    });
  });
  await renderInbox();
  await act(async () => button("A plain").click());
  await settle();
  expect(mockRequest).toHaveBeenCalledWith(expect.stringContaining("uid=8"));
  expect(mockRequest).toHaveBeenCalledWith(
    expect.stringContaining("uid_validity=1"),
  );
  await act(async () => button("Download").click());
  await settle();
  expect(mockRequest).toHaveBeenCalledWith(expect.stringContaining("uid=8"));
  expect(mockRequest).toHaveBeenCalledWith(
    expect.stringContaining("attachment_id=managed-attachment"),
  );
  expect(click).toHaveBeenCalled();
  await act(async () => button("Mark read").click());
  expect(mockRequest).toHaveBeenCalledWith(
    expect.stringContaining("imap/flags?"),
    "PATCH",
    { folder: "INBOX", uid: 8, uidValidity: 1, flag: "\\Seen", enabled: true },
  );
  click.mockRestore();
});

test("uses opaque Gmail pagination and native IDs for detail and flags", async () => {
  gmailMailbox = true;
  const native = Array.from({ length: 20 }, (_, index) => ({
    ...base,
    uid: undefined,
    uidValidity: undefined,
    id: `mailbox-a:gmail:native-${index}`,
    providerMessageId: `native-${index}`,
    messageId: `<rfc-${index}@example.com>`,
    subject: `Native message ${index}`,
    flags: [],
  }));
  mockRequest.mockImplementation(
    (path: string, method = "GET", body?: unknown) => {
      const params = new URLSearchParams(path.split("?")[1]);
      if (path.startsWith("imap/message?"))
        return Promise.resolve({
          ...native[0],
          body: "<p>Full Gmail API body</p>",
        });
      if (method === "PATCH") return Promise.resolve(undefined);
      if (params.get("page_token") === "opaque-next")
        return Promise.resolve({
          emails: [
            native[19],
            {
              ...native[19],
              id: "mailbox-a:gmail:native-20",
              providerMessageId: "native-20",
              subject: "Native message 20",
            },
          ],
          total_emails: 21,
          total_is_estimate: true,
          history_id: "history-2",
          limit: 20,
          offset: 0,
        });
      return Promise.resolve({
        emails: native,
        total_emails: 20,
        total_is_estimate: true,
        history_id: "history-1",
        next_page_token: "opaque-next",
        limit: 20,
        offset: 0,
      });
    },
  );
  await renderInbox();
  expect(container.textContent).toContain("About 20 messages");
  await act(async () => button("Load more messages").click());
  await settle();
  expect(
    mockRequest.mock.calls.some(([path]) =>
      String(path).includes("page_token=opaque-next"),
    ),
  ).toBe(true);
  expect(
    Array.from(container.querySelectorAll("button")).filter((node) =>
      node.textContent?.includes("Native message 19"),
    ),
  ).toHaveLength(1);
  expect(container.textContent).toContain("Native message 20");
  await act(async () => button("Native message 0").click());
  await settle();
  expect(
    mockRequest.mock.calls.some(([path]) =>
      String(path).includes("message_id=native-0"),
    ),
  ).toBe(true);
  expect(mockSummaryProps).toHaveBeenCalledWith(
    expect.objectContaining({ providerMessageId: "native-0" }),
  );
  await act(async () => button("Mark read").click());
  expect(mockRequest).toHaveBeenCalledWith(
    expect.stringContaining("imap/flags?"),
    "PATCH",
    {
      folder: "INBOX",
      providerMessageId: "native-0",
      flag: "\\Seen",
      enabled: true,
    },
  );
});

test("fetches Gmail attachment content only when Download is clicked", async () => {
  gmailMailbox = true;
  const native = {
    ...base,
    uid: undefined,
    uidValidity: undefined,
    id: "mailbox-a:gmail:native-attachment",
    providerMessageId: "native-attachment",
    messageId: "<attachment@example.com>",
    subject: "Native attachment",
    flags: [],
  };
  const click = jest
    .spyOn(HTMLAnchorElement.prototype, "click")
    .mockImplementation(() => {});
  mockRequest.mockImplementation((path: string) => {
    if (path.startsWith("imap/attachment?"))
      return Promise.resolve({ Data: "aGVsbG8=" });
    if (path.startsWith("imap/message?"))
      return Promise.resolve({
        ...native,
        body: "<p>Full body</p>",
        attachments: [
          {
            Filename: "brief.txt",
            MIMEType: "text/plain",
            Size: 5,
            AttachmentID: "attachment-opaque",
          },
        ],
      });
    return Promise.resolve({
      emails: [native],
      total_emails: 1,
      total_is_estimate: true,
      history_id: "history-1",
      limit: 20,
      offset: 0,
    });
  });
  await renderInbox();
  expect(
    mockRequest.mock.calls.some(([path]) =>
      String(path).startsWith("imap/attachment?"),
    ),
  ).toBe(false);
  await act(async () => button("Native attachment").click());
  await settle();
  await act(async () => button("Download").click());
  await settle();
  expect(mockRequest).toHaveBeenCalledWith(
    expect.stringContaining("message_id=native-attachment"),
  );
  expect(mockRequest).toHaveBeenCalledWith(
    expect.stringContaining("attachment_id=attachment-opaque"),
  );
  expect(click).toHaveBeenCalled();
  click.mockRestore();
});

test("keeps the first-page Gmail history baseline while loading another page", async () => {
  gmailMailbox = true;
  Object.defineProperty(document, "visibilityState", {
    configurable: true,
    value: "visible",
  });
  const headPaths: string[] = [];
  const native = {
    ...base,
    uid: undefined,
    uidValidity: undefined,
    id: "mailbox-a:gmail:native-history",
    providerMessageId: "native-history",
    messageId: "<history@example.com>",
    subject: "History message",
    flags: [],
  };
  mockRequest.mockImplementation((path: string) => {
    const params = new URLSearchParams(path.split("?")[1]);
    if (path.startsWith("imap/head?")) {
      headPaths.push(path);
      return Promise.resolve({
        history_id: "history-1",
        changed: false,
      });
    }
    if (params.get("page_token") === "opaque-next")
      return Promise.resolve({
        emails: [],
        total_emails: 1,
        total_is_estimate: true,
        history_id: "history-2",
        limit: 20,
        offset: 0,
      });
    return Promise.resolve({
      emails: [native],
      total_emails: 1,
      total_is_estimate: true,
      history_id: "history-1",
      next_page_token: "opaque-next",
      limit: 20,
      offset: 0,
    });
  });
  await renderInbox();
  await act(async () => button("Load more messages").click());
  await settle();
  await act(async () => window.dispatchEvent(new Event("online")));
  await settle();
  expect(headPaths.length).toBeGreaterThanOrEqual(2);
  expect(headPaths.every((path) => path.includes("history_id=history-1"))).toBe(
    true,
  );
});

test("restores the Gmail history baseline from warm query data after remount", async () => {
  gmailMailbox = true;
  queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: Infinity } },
  });
  Object.defineProperty(document, "visibilityState", {
    configurable: true,
    value: "visible",
  });
  const headPaths: string[] = [];
  const native = {
    ...base,
    uid: undefined,
    uidValidity: undefined,
    id: "mailbox-a:gmail:native-warm",
    providerMessageId: "native-warm",
    messageId: "<warm@example.com>",
    subject: "Warm cache message",
    flags: [],
  };
  mockRequest.mockImplementation((path: string) => {
    if (path.startsWith("imap/head?")) {
      headPaths.push(path);
      return Promise.resolve({ history_id: "history-warm", changed: false });
    }
    return Promise.resolve({
      emails: [native],
      total_emails: 1,
      total_is_estimate: true,
      history_id: "history-warm",
      limit: 20,
      offset: 0,
    });
  });
  await renderInbox();
  const beforeRemount = headPaths.length;
  expect(beforeRemount).toBeGreaterThan(0);
  await act(async () => root.unmount());
  root = createRoot(container);
  await act(async () =>
    root.render(
      <QueryClientProvider client={queryClient}>
        <InboxPage />
      </QueryClientProvider>,
    ),
  );
  await settle();
  expect(headPaths.length).toBeGreaterThan(beforeRemount);
  expect(headPaths.at(-1)).toContain("history_id=history-warm");
});

test("does not download a late Gmail attachment after message selection changes", async () => {
  gmailMailbox = true;
  let finishAttachment!: (value: { Data: string }) => void;
  const click = jest
    .spyOn(HTMLAnchorElement.prototype, "click")
    .mockImplementation(() => {});
  const native = [0, 1].map((index) => ({
    ...base,
    uid: undefined,
    uidValidity: undefined,
    id: `mailbox-a:gmail:native-${index}`,
    providerMessageId: `native-${index}`,
    messageId: `<native-${index}@example.com>`,
    subject: `Attachment message ${index}`,
    flags: [],
  }));
  mockRequest.mockImplementation((path: string) => {
    const params = new URLSearchParams(path.split("?")[1]);
    if (path.startsWith("imap/attachment?"))
      return new Promise((resolve) => {
        finishAttachment = resolve;
      });
    if (path.startsWith("imap/message?")) {
      const message = native.find(
        (item) => item.providerMessageId === params.get("message_id"),
      )!;
      return Promise.resolve({
        ...message,
        body: "<p>Full body</p>",
        attachments: [
          {
            Filename: "brief.txt",
            Size: 5,
            AttachmentID: `attachment-${message.providerMessageId}`,
          },
        ],
      });
    }
    return Promise.resolve({
      emails: native,
      total_emails: 2,
      total_is_estimate: true,
      history_id: "history-1",
      limit: 20,
      offset: 0,
    });
  });
  await renderInbox();
  await act(async () => button("Attachment message 0").click());
  await settle();
  await act(async () => button("Download").click());
  await act(async () => button("Attachment message 1").click());
  await settle();
  await act(async () => finishAttachment({ Data: "aGVsbG8=" }));
  expect(click).not.toHaveBeenCalled();
  click.mockRestore();
});
