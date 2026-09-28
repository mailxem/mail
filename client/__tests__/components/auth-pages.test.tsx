/** @jest-environment jsdom */
import * as React from "react";
import { act } from "react";
import { createRoot, Root } from "react-dom/client";
import ForgotPasswordPage from "@/app/(auth)/auth/forgot-password/page";
import RegisterPage from "@/app/(auth)/auth/register/page";
import { AuthPasswordField } from "@/components/auth/auth-fields";
const mockFetch = jest.fn();
const mockSignIn = jest.fn();
jest.mock("@/hooks/use-api", () => ({
  useApi: () => ({ apiFetch: mockFetch }),
}));
jest.mock("next-auth/react", () => ({
  useSession: () => ({ data: null }),
  signIn: (...args: any[]) => mockSignIn(...args),
}));
jest.mock("next/navigation", () => ({ redirect: jest.fn() }));
jest.mock("sonner", () => ({ toast: { success: jest.fn() } }));
jest.mock("@/components/auth/auth.module.css", () => ({}));
let root: Root, container: HTMLDivElement;
beforeAll(() => {
  (globalThis as any).IS_REACT_ACT_ENVIRONMENT = true;
});
beforeEach(() => {
  mockFetch.mockReset();
  mockSignIn.mockReset();
  container = document.createElement("div");
  document.body.appendChild(container);
  root = createRoot(container);
});
afterEach(async () => {
  await act(async () => root.unmount());
  container.remove();
});
async function fill(name: string, value: string) {
  const input = container.querySelector(
    `input[name="${name}"]`,
  ) as HTMLInputElement;
  await act(async () => {
    Object.getOwnPropertyDescriptor(
      HTMLInputElement.prototype,
      "value",
    )!.set!.call(input, value);
    input.dispatchEvent(new Event("input", { bubbles: true }));
  });
}
async function submit() {
  await act(async () =>
    container
      .querySelector("form")!
      .dispatchEvent(new Event("submit", { bubbles: true, cancelable: true })),
  );
}
test("password visibility has a label and never submits the form", async () => {
  const submitted = jest.fn();
  await act(async () =>
    root.render(
      <form onSubmit={submitted}>
        <AuthPasswordField
          id="secret"
          label="Password"
          defaultValue="test-password"
        />
      </form>,
    ),
  );
  const reveal = container.querySelector(
    'button[aria-label="Show password"]',
  ) as HTMLButtonElement;
  expect(reveal.type).toBe("button");
  await act(async () => reveal.click());
  expect(container.querySelector("input")!.type).toBe("text");
  expect(submitted).not.toHaveBeenCalled();
  expect(
    container.querySelector('button[aria-label="Hide password"]'),
  ).not.toBeNull();
});
test("a rejected reset request shows an error instead of false success", async () => {
  mockFetch.mockResolvedValue({ ok: false });
  await act(async () => root.render(<ForgotPasswordPage />));
  await fill("email", "person@example.com");
  await submit();
  expect(mockFetch).toHaveBeenCalledWith(
    "auth/password-reset",
    expect.objectContaining({ requireAuth: false, method: "POST" }),
  );
  expect(container.querySelector('[role="alert"]')?.textContent).toContain(
    "couldn’t request",
  );
  expect(container.querySelector('[role="status"]')).toBeNull();
});
test("a successful reset request shows neutral confirmation", async () => {
  mockFetch.mockResolvedValue({ ok: true });
  await act(async () => root.render(<ForgotPasswordPage />));
  await fill("email", "person@example.com");
  await submit();
  expect(container.querySelector('[role="status"]')?.textContent).toContain(
    "If an account exists",
  );
});
test("registration keeps matching-password validation and accessible labels", async () => {
  await act(async () => root.render(<RegisterPage />));
  await fill("first_name", "Alex");
  await fill("last_name", "Morgan");
  await fill("email", "person@example.com");
  await fill("password", "first-test-password");
  await fill("confirmPassword", "different-password");
  await submit();
  expect(mockFetch).not.toHaveBeenCalled();
  expect(container.textContent).toContain("Passwords don’t match");
  expect(container.querySelector('label[for="email"]')?.textContent).toBe(
    "Email",
  );
});
