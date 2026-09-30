/** @jest-environment jsdom */
import React, { act } from "react";
import { createRoot, type Root } from "react-dom/client";
import type { DesignEditor } from "@/components/assistant/workspace-ai";
let mockBridge: DesignEditor;
const mockRegister = jest.fn((entry: DesignEditor) => {
  mockBridge = entry;
  return () => {};
});
const mockListeners: Record<string, () => void> = {};
const mockEditor = {
  exportHtml: jest.fn((callback: (value: unknown) => void) =>
    callback({
      html: "<p>Original</p>",
      design: { schemaVersion: 18, body: { rows: [] } },
    }),
  ),
  loadDesign: jest.fn(() => mockListeners["design:loaded"]?.()),
  addEventListener: jest.fn((name, fn) => {
    mockListeners[name] = fn;
  }),
  removeEventListener: jest.fn((name) => {
    delete mockListeners[name];
  }),
};
jest.mock("react-email-editor", () => {
  const React = require("react");
  return {
    __esModule: true,
    default: React.forwardRef((props: any, ref: any) => {
      React.useImperativeHandle(ref, () => ({ editor: mockEditor }));
      React.useEffect(() => {
        props.onReady();
      }, []);
      return <div>Editor canvas</div>;
    }),
  };
});
jest.mock("@/components/assistant/workspace-ai", () => ({
  useWorkspaceAI: () => ({ register: mockRegister, openDesigner: jest.fn() }),
}));
jest.mock("next/navigation", () => ({
  useRouter: () => ({ replace: jest.fn() }),
}));
jest.mock("@/app/providers/team-provider", () => ({
  useTeam: () => ({ team: { id: "team" } }),
}));
jest.mock("@/hooks/use-api", () => ({
  useApi: () => ({ apiFetch: jest.fn(), session: { accessToken: "fixture" } }),
}));
jest.mock("@/components/marketing/shared", () => ({ QueryState: () => null }));
jest.mock("@/lib/template-starters/editor-assets", () => ({
  prepareEditorAssets: async (design: unknown) => ({
    design,
    sources: new Map(),
  }),
  restoreEditorAssets: (value: unknown) => value,
}));
jest.mock("sonner", () => ({
  toast: { success: jest.fn(), error: jest.fn() },
}));
const mockQueryClient = {
  setQueryData: jest.fn(),
  invalidateQueries: jest.fn().mockResolvedValue(undefined),
};
const record = {
  id: "template",
  teamId: "team",
  name: "Welcome",
  subject: "Hello",
  categoryId: "category",
  design: {},
  updatedAt: "2026-09-30T10:00:00Z",
};
jest.mock("@tanstack/react-query", () => ({
  useQueryClient: () => mockQueryClient,
  useQuery: ({ queryKey }: any) => ({
    data:
      queryKey[0] === "template"
        ? record
        : queryKey[0] === "emailCategories"
          ? [{ id: "category", name: "Transactional" }]
          : undefined,
    isPending: false,
    refetch: jest.fn(),
  }),
}));
import { TemplateEditor } from "@/components/templates/template-editor";
const draft = {
  subject: "Héllo",
  body: "Welcome",
  previewHtml: "<p>Welcome</p>",
  design: { schemaVersion: 18, body: { rows: [{}] } },
};
let root: Root;
let container: HTMLDivElement;
beforeEach(async () => {
  (globalThis as any).IS_REACT_ACT_ENVIRONMENT = true;
  global.fetch = jest.fn();
  global.structuredClone = (value) => JSON.parse(JSON.stringify(value));
  container = document.createElement("div");
  document.body.appendChild(container);
  root = createRoot(container);
  await act(async () => root.render(<TemplateEditor templateId="template" />));
});
afterEach(async () => {
  await act(async () => root.unmount());
  container.remove();
});
test("manual changes after generation starts prevent AI replacement and saving", async () => {
  const before = await mockBridge.snapshot();
  await act(async () => {
    const name = container.querySelector("#name")!;
    Object.getOwnPropertyDescriptor(
      HTMLInputElement.prototype,
      "value",
    )!.set!.call(name, "My manual edit");
    name.dispatchEvent(new Event("input", { bubbles: true }));
  });
  await expect(mockBridge.apply(draft, before)).rejects.toThrow(
    "You edited this template",
  );
  expect(global.fetch).not.toHaveBeenCalled();
  expect((container.querySelector("#name") as HTMLInputElement).value).toBe(
    "My manual edit",
  );
});
test("AI saves native design, HTML and the expected version to the open template URL", async () => {
  (global.fetch as jest.Mock).mockResolvedValue({
    ok: true,
    json: async () => ({
      ...record,
      subject: draft.subject,
      updatedAt: "2026-09-30T11:00:00Z",
    }),
  });
  const before = await mockBridge.snapshot();
  await act(async () => mockBridge.apply(draft, before));
  const [url, options] = (global.fetch as jest.Mock).mock.calls[0];
  expect(url).toBe("/api/templates/template");
  const body = JSON.parse(options.body);
  expect(body).toMatchObject({
    teamId: "team",
    subject: "Héllo",
    htmlBody: "<p>Welcome</p>",
    expectedUpdatedAt: record.updatedAt,
  });
  expect(
    JSON.parse(Buffer.from(body.designJson, "base64").toString("utf8")),
  ).toEqual(draft.design);
});
