/** @jest-environment jsdom */
import * as React from "react";
import { randomUUID } from "node:crypto";
import { act } from "react";
import { createRoot, Root } from "react-dom/client";
import { PublicForm } from "@/components/marketing/public-form";
import { FormPreview } from "@/components/marketing/form-preview";
import type { FormDefinition } from "@/lib/marketing/form-definition";

jest.mock("@/components/marketing/public-form.module.css", () => ({}));
jest.mock("@/components/marketing/form-surface", () => ({ FormSurface: ({ children }: { children: React.ReactNode }) => <div>{children}</div> }));
jest.mock("@/components/marketing/shared", () => ({ Field: ({ children, label }: { children: React.ReactNode; label: string }) => <label>{label}{children}</label> }));
let root: Root;
let container: HTMLDivElement;
let fetcher: jest.Mock;
let frames: FrameRequestCallback[] = [];
const definition = (): FormDefinition => ({ schemaVersion: 1, pages: [{ id: "identity", title: "Your details", fields: [{ key: "email", label: "Email", type: "EMAIL", required: true, defaultValue: "reader@example.com" }] }], consent: { mode: "none", label: "" }, saveProgress: true });
const form = (version = 1, value = definition()) => ({ name: "Join us", description: "", fields: [], definition: value, version, buttonText: "Send response", successMessage: "Response saved" });
const response = (body: unknown, status = 200) => ({ ok: status < 400, status, json: async () => body });
async function flush() { await act(async () => { await Promise.resolve(); await Promise.resolve(); }); await act(async () => { frames.splice(0).forEach((callback) => callback(0)); }); }
async function submit() { await act(async () => { container.querySelector("form")!.dispatchEvent(new Event("submit", { bubbles: true, cancelable: true })); }); await flush(); }
beforeAll(() => { (globalThis as any).IS_REACT_ACT_ENVIRONMENT = true; Object.defineProperty(crypto, "randomUUID", { configurable: true, value: randomUUID }); window.requestAnimationFrame = (callback) => { frames.push(callback); return frames.length; }; });
beforeEach(() => {
  container = document.createElement("div"); document.body.appendChild(container); root = createRoot(container);
  frames = []; sessionStorage.clear(); window.history.replaceState(null, "", "/f/signup");
  fetcher = jest.fn(); global.fetch = fetcher;
});
afterEach(async () => { await act(async () => root.unmount()); container.remove(); });

test("a revised resumed form clears the obsolete token before a new submission", async () => {
  window.history.replaceState(null, "", "/f/signup#resume=private-token");
  let loaded = 0; const submissions: any[] = [];
  fetcher.mockImplementation(async (url: string, init?: RequestInit) => {
    if (url.endsWith("/events")) return response({});
    if (url.endsWith("/resume")) return response({ fields: { email: "reader@example.com" }, pageId: "identity", version: 1, sessionId: "64bdb84e-372c-4c2d-af61-488e9784a820" });
    if (!init?.method) return response(form(++loaded));
    submissions.push(JSON.parse(String(init.body)));
    return submissions.length === 1 ? response({ message: "This form changed. Reload it before continuing." }, 409) : response({});
  });
  await act(async () => root.render(<PublicForm slug="signup" />)); await flush();
  expect(window.location.hash).toBe("");
  await submit(); await submit();
  expect(submissions).toHaveLength(2);
  expect(submissions[0].resumeToken).toBe("private-token");
  expect(submissions[1].resumeToken).toBeUndefined();
  expect(submissions[1].version).toBe(2);
  expect(submissions[1].fields.email).toBe("reader@example.com");
  expect(container.textContent).toContain("Response saved");
});

test("an uncertain submission is retried without changing its request identifier or answers", async () => {
  const submissions: any[] = [];
  sessionStorage.setItem("xem-form-session:signup", "not-a-valid-session");
  fetcher.mockImplementation(async (url: string, init?: RequestInit) => {
    if (url.endsWith("/events")) return response({});
    if (!init?.method) return response(form());
    submissions.push(JSON.parse(String(init.body)));
    if (submissions.length === 1) throw new Error("Connection lost");
    return response({});
  });
  await act(async () => root.render(<PublicForm slug="signup" />)); await flush();
  await submit();
  expect(container.querySelector("fieldset")?.disabled).toBe(true);
  expect(container.textContent).toContain("Retry submission");
  await submit();
  expect(submissions[1]).toEqual(submissions[0]);
  expect(submissions[0].sessionId).toMatch(/^[0-9a-f-]{36}$/);
  expect(submissions[0].sessionId).not.toBe("not-a-valid-session");
});

test("invalid radio answers focus the first input rather than the nonfocusable fieldset", async () => {
  const value = definition();
  value.pages[0].fields.push({ key: "topic", label: "Topic", type: "RADIO", required: true, options: ["Product", "Company"] });
  fetcher.mockImplementation(async (url: string) => response(url.endsWith("/events") ? {} : form(1, value)));
  await act(async () => root.render(<PublicForm slug="signup" />)); await flush();
  await submit();
  expect(document.activeElement).toBe(container.querySelector('input[type="radio"]'));
  expect(container.querySelector("h2")).toBeNull();
});

test("editor preview shows the same default answers used by conditional reachability", async () => {
  const value = definition();
  value.pages[0].fields.push({ key: "updates", label: "Updates", type: "CHECKBOX", required: false, defaultValue: "true" });
  await act(async () => root.render(<FormPreview name="Join" description="" definition={value} button="Send" />));
  expect(container.querySelector<HTMLInputElement>('input[type="email"]')?.value).toBe("reader@example.com");
  expect(container.querySelector<HTMLInputElement>('input[type="checkbox"]')?.checked).toBe(true);
});
