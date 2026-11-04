/** @jest-environment jsdom */
import * as React from "react";
import { act } from "react";
import { createRoot, Root } from "react-dom/client";
import { BillingDetailsDialog, hasCompleteBillingDetails } from "@/components/billing/billing-details-dialog";

let root: Root;
let container: HTMLDivElement;
beforeAll(() => {
  (globalThis as any).IS_REACT_ACT_ENVIRONMENT = true;
  globalThis.ResizeObserver = class { observe() {} unobserve() {} disconnect() {} };
});
beforeEach(() => {
  container = document.createElement("div");
  document.body.appendChild(container);
  root = createRoot(container);
});
afterEach(async () => {
  await act(async () => { root.unmount(); });
  container.remove();
});

function type(id: string, value: string) {
  const input = document.getElementById(id) as HTMLInputElement;
  const setter = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, "value")!.set!;
  setter.call(input, value);
  input.dispatchEvent(new Event("input", { bubbles: true }));
}

test("an address with any blank field is incomplete", () => {
  const full = { team_id: "t", street: "1 Main St", city: "Austin", state: "TX", zip: "78701", country: "US" };
  expect(hasCompleteBillingDetails(full)).toBe(true);
  expect(hasCompleteBillingDetails({ ...full, zip: " " })).toBe(false);
  expect(hasCompleteBillingDetails(undefined)).toBe(false);
});

test("submitting keeps what was saved and sends an upper-case country code", async () => {
  const onSubmit = jest.fn();
  await act(async () => {
    root.render(
      <BillingDetailsDialog
        open
        onOpenChange={() => {}}
        teamId="team"
        initial={{ id: "bd1", team_id: "team", street: "1 Main St", city: "Austin", state: "", zip: "", country: "" }}
        saving={false}
        onSubmit={onSubmit}
      />,
    );
  });
  expect((document.getElementById("billing-street") as HTMLInputElement).value).toBe("1 Main St");
  await act(async () => {
    type("billing-state", "TX");
    type("billing-zip", "78701");
    type("billing-country", "us");
  });
  await act(async () => {
    document.querySelector("form")!.dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
  });
  expect(onSubmit).toHaveBeenCalledWith({ id: "bd1", team_id: "team", street: "1 Main St", city: "Austin", state: "TX", zip: "78701", country: "US" });
});
