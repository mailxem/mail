/** @jest-environment jsdom */
import * as React from "react";
import { act } from "react";
import { createRoot, Root } from "react-dom/client";
import {
  GMAIL_CLIP_BYTES,
  GMAIL_WARN_BYTES,
  gmailClipState,
  sentEmailBytes,
} from "@/lib/email/size";
import { GmailClipNotice } from "@/components/campaigns/gmail-clip-notice";

// Builds HTML whose estimated sent size is exactly `bytes`.
const sized = (bytes: number) => {
  const base = sentEmailBytes("");
  return "x".repeat(bytes - base);
};

describe("sentEmailBytes", () => {
  it("counts UTF-8 bytes, tracked links and what sending adds", () => {
    const empty = sentEmailBytes("");
    expect(empty).toBeGreaterThan(0);
    expect(sentEmailBytes("é")).toBe(empty + 2);
    const plain = "<p>hi</p>";
    const linked = '<a href="https://example.com">hi</a>';
    const relative = '<a href="mailto:a@b.co">hi</a>';
    expect(sentEmailBytes(linked) - linked.length).toBeGreaterThan(
      sentEmailBytes(plain) - plain.length,
    );
    expect(sentEmailBytes(relative) - relative.length).toBe(empty);
  });

  it("maps sizes to the Gmail thresholds", () => {
    expect(gmailClipState(GMAIL_WARN_BYTES - 1)).toBe("ok");
    expect(gmailClipState(GMAIL_WARN_BYTES)).toBe("close");
    expect(gmailClipState(GMAIL_CLIP_BYTES)).toBe("close");
    expect(gmailClipState(GMAIL_CLIP_BYTES + 1)).toBe("clipped");
  });
});

describe("GmailClipNotice", () => {
  let root: Root;
  let container: HTMLDivElement;
  beforeAll(() => {
    (globalThis as any).IS_REACT_ACT_ENVIRONMENT = true;
  });
  beforeEach(() => {
    container = document.createElement("div");
    root = createRoot(container);
  });
  afterEach(() => act(() => root.unmount()));
  const render = (html?: string) =>
    act(() => root.render(<GmailClipNotice html={html} />));

  it("renders nothing without HTML", () => {
    render(undefined);
    expect(container.innerHTML).toBe("");
  });

  it("stays quiet under 90 KB", () => {
    render(sized(50_000));
    expect(container.querySelector("[data-state=ok]")?.textContent).toBe(
      "About 50 KB when sent",
    );
    expect(container.querySelector("[role=status]")).toBeNull();
  });

  it("warns between 90 and 102 KB", () => {
    render(sized(95_000));
    const notice = container.querySelector("[data-state=close]");
    expect(notice?.textContent).toContain("Close to Gmail's 102 KB limit");
    expect(notice?.textContent).toContain("data URLs");
  });

  it("warns clearly above 102 KB", () => {
    render(sized(150_000));
    expect(
      container.querySelector("[data-state=clipped]")?.textContent,
    ).toContain("Gmail will clip this email and hide your unsubscribe link");
  });
});
