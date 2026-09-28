/** @jest-environment jsdom */
import * as React from "react";
import { act } from "react";
import { createRoot, Root } from "react-dom/client";
import { randomUUID } from "node:crypto";
import { AutomationBuilder } from "@/components/marketing/automations";
import type { Workflow } from "@/lib/marketing/types";
const mockRequest = jest.fn();
jest.mock("@/lib/marketing/api", () => ({
  useMarketing: () => ({ request: mockRequest, refresh: async () => {} }),
  useMarketingQuery: () => ({
    data: { templates: [], senders: [], lists: [] },
  }),
}));
jest.mock("sonner", () => ({
  toast: { success: jest.fn(), error: jest.fn() },
}));
jest.mock("@xyflow/react/dist/style.css", () => ({}));
jest.mock("@xyflow/react", () => {
  const R = require("react");
  return {
    ReactFlow: ({ nodes, onNodeClick, children }: any) => (
      <div>
        {nodes.map((n: any) => (
          <button key={n.id} onClick={() => onNodeClick(null, n)}>
            {n.id}
          </button>
        ))}
        {children}
      </div>
    ),
    Background: () => null,
    Controls: () => null,
    Handle: () => null,
    Position: { Top: "top", Bottom: "bottom" },
    MarkerType: { ArrowClosed: "arrow" },
    useNodesState: (v: any) => {
      const [s, set] = R.useState(v);
      return [s, set, () => {}];
    },
    useEdgesState: (v: any) => {
      const [s, set] = R.useState(v);
      return [s, set, () => {}];
    },
    addEdge: (e: any, es: any) => [...es, e],
  };
});
let root: Root, container: HTMLDivElement;
const flow: Workflow = {
  id: "",
  name: "Branch test",
  description: "Keep this description",
  triggerEvent: "manual",
  isActive: false,
  updatedAt: "",
  nodes: [
    { id: "start", type: "START", data: {} },
    {
      id: "condition",
      type: "CONDITION",
      data: {
        operator: "AND",
        conditions: [
          { variable: "contact_email", operator: "contains", value: "example" },
        ],
      },
    },
    { id: "yes-exit", type: "EXIT", data: {} },
    { id: "no-exit", type: "EXIT", data: {} },
  ],
  edges: [
    { sourceId: "start", targetId: "condition", label: "" },
    { sourceId: "condition", targetId: "yes-exit", label: "true" },
    { sourceId: "condition", targetId: "no-exit", label: "false" },
  ],
};
beforeAll(() => {
  (globalThis as any).IS_REACT_ACT_ENVIRONMENT = true;
  Object.defineProperty(crypto, "randomUUID", {
    configurable: true,
    value: randomUUID,
  });
});
beforeEach(() => {
  mockRequest.mockReset().mockResolvedValue({ id: "saved" });
  container = document.createElement("div");
  document.body.appendChild(container);
  root = createRoot(container);
});
afterEach(async () => {
  await act(async () => root.unmount());
  container.remove();
});
async function click(label: string) {
  const button = Array.from(container.querySelectorAll("button")).find(
    (b) => b.getAttribute("aria-label") === label || b.textContent === label,
  );
  expect(button).toBeDefined();
  await act(async () => button!.click());
}
test("inserting before a branch finish keeps its label and leaves the other branch intact", async () => {
  await act(async () =>
    root.render(<AutomationBuilder initial={flow} close={() => {}} />),
  );
  await click("no-exit");
  await click("Add Set a variable");
  await click("Save draft");
  const payload = mockRequest.mock.calls[0][2];
  const variable = payload.nodes.find((n: any) => n.type === "SET_VARIABLE");
  expect(payload.edges).toContainEqual({
    sourceId: "condition",
    targetId: variable.id,
    label: "false",
  });
  expect(payload.edges).toContainEqual({
    sourceId: variable.id,
    targetId: "no-exit",
    label: "",
  });
  expect(payload.edges).toContainEqual({
    sourceId: "condition",
    targetId: "yes-exit",
    label: "true",
  });
  expect(payload.description).toBe("Keep this description");
  expect(variable.data.variable).toBe("workflow_stage");
});
test("adding a second rule preserves the first on save", async () => {
  await act(async () =>
    root.render(<AutomationBuilder initial={flow} close={() => {}} />),
  );
  await click("condition");
  await click("Add rule");
  await click("Save draft");
  const condition = mockRequest.mock.calls[0][2].nodes.find(
    (n: any) => n.type === "CONDITION",
  );
  expect(condition.data.conditions).toHaveLength(2);
  expect(condition.data.conditions[0].value).toBe("example");
});
test("published workflows disable step additions and configuration", async () => {
  await act(async () =>
    root.render(
      <AutomationBuilder
        initial={{ ...flow, id: "published", isActive: true }}
        close={() => {}}
      />,
    ),
  );
  expect(
    container.querySelector('button[aria-label="Add Manage tags"]'),
  ).toHaveProperty("disabled", true);
  expect(container.querySelector("fieldset")).toHaveProperty("disabled", true);
});
