/** @jest-environment jsdom */
import React, { act, useEffect } from "react";
import { createRoot, type Root } from "react-dom/client";
jest.mock("next/navigation", () => ({
  usePathname: () => "/templates/fixture/edit",
}));
jest.mock("next/link", () => ({
  __esModule: true,
  default: ({ children, ...props }: any) => <a {...props}>{children}</a>,
}));
jest.mock("@/app/providers/team-provider", () => ({
  useTeam: () => ({ team: { id: "team" } }),
}));
jest.mock("next-auth/react", () => ({
  useSession: () => ({ data: { user: { email: "owner@example.test" } } }),
}));
jest.mock("@/components/assistant/assistant-home", () => ({
  AssistantHome: () => <p>Workspace assistant</p>,
}));
jest.mock(
  "@/components/assistant/workspace-ai.module.css",
  () => new Proxy({}, { get: (_, key) => String(key) }),
);
import {
  WorkspaceAIProvider,
  useWorkspaceAI,
  type DesignEditor,
} from "@/components/assistant/workspace-ai";
const snapshot = {
  key: "team:fixture",
  revision: 1,
  subject: "Hello",
  html: "Previous content",
  design: {},
};
let editor: DesignEditor;
let root: Root;
let container: HTMLDivElement;
let fetchMock: jest.Mock;
function Register() {
  const ai = useWorkspaceAI();
  useEffect(() => ai!.register(editor), [ai?.register]);
  return null;
}
const click = async (label: string) => {
  await act(async () => {
    (
      container.querySelector(`[aria-label="${label}"]`) as HTMLButtonElement
    ).click();
  });
};
beforeEach(async () => {
  (globalThis as any).IS_REACT_ACT_ENVIRONMENT = true;
  fetchMock = jest.fn();
  global.fetch = fetchMock;
  editor = {
    key: "team:fixture",
    name: "Welcome",
    snapshot: jest.fn().mockResolvedValue(snapshot),
    apply: jest.fn().mockResolvedValue(undefined),
    undo: jest.fn().mockResolvedValue(undefined),
  };
  container = document.createElement("div");
  document.body.appendChild(container);
  root = createRoot(container);
  await act(async () => {
    root.render(
      <WorkspaceAIProvider>
        <Register />
      </WorkspaceAIProvider>,
    );
  });
  await click("Open Xem AI");
});
afterEach(async () => {
  await act(async () => root.unmount());
  container.remove();
});
const generate = async () => {
  await act(async () => {
    const field = container.querySelector("textarea")!;
    Object.getOwnPropertyDescriptor(
      HTMLTextAreaElement.prototype,
      "value",
    )!.set!.call(field, "Make this warmer");
    field.dispatchEvent(new Event("input", { bubbles: true }));
  });
  await click("Design and save revision");
};
const draft = {
  subject: "Welcome",
  body: "Hello",
  previewHtml: "<p>Hello</p>",
  design: { schemaVersion: 18, body: { rows: [{}] } },
};
test("uses current editor context and reports saved only after persistence resolves", async () => {
  let saved!: () => void;
  editor.apply = jest.fn().mockImplementation(
    () =>
      new Promise<void>((resolve) => {
        saved = resolve;
      }),
  );
  fetchMock.mockResolvedValue({ ok: true, json: async () => draft });
  await generate();
  expect(JSON.parse(fetchMock.mock.calls[0][1].body)).toMatchObject({
    teamId: "team",
    body: "Previous content",
    instruction: "Make this warmer",
  });
  expect(editor.apply).toHaveBeenCalledWith(draft, snapshot);
  expect(container.textContent).toContain("Updating and saving");
  expect(container.textContent).not.toContain("design is updated and saved");
  await act(async () => saved());
  expect(container.textContent).toContain("design is updated and saved");
});
test("navigation cancels a pending design and never applies it to another editor", async () => {
  let finish!: (value: unknown) => void;
  fetchMock.mockImplementation(
    () =>
      new Promise((resolve) => {
        finish = resolve;
      }),
  );
  await generate();
  const signal = fetchMock.mock.calls[0][1].signal;
  await act(async () =>
    root.render(
      <WorkspaceAIProvider>
        <div>Different page</div>
      </WorkspaceAIProvider>,
    ),
  );
  expect(signal.aborted).toBe(true);
  await act(async () => finish({ ok: true, json: async () => draft }));
  expect(editor.apply).not.toHaveBeenCalled();
});
test("failed saves keep the generated revision and never claim success", async () => {
  editor.apply = jest
    .fn()
    .mockRejectedValue(new Error("This template changed in another session."));
  fetchMock.mockResolvedValue({ ok: true, json: async () => draft });
  await generate();
  expect(container.textContent).toContain("another session");
  expect(container.textContent).toContain("Download revision");
  expect(container.textContent).not.toContain("design is updated and saved");
});
test("Escape minimizes the non-modal panel and returns focus", async () => {
  await act(async () => {
    container
      .querySelector('[role="dialog"]')!
      .dispatchEvent(
        new KeyboardEvent("keydown", { key: "Escape", bubbles: true }),
      );
  });
  expect(
    container.querySelector('[role="dialog"]')!.hasAttribute("hidden"),
  ).toBe(true);
  expect(document.activeElement?.getAttribute("aria-label")).toBe(
    "Open Xem AI",
  );
});
