import { test, expect } from "@playwright/test";
import AxeBuilder from "@axe-core/playwright";

async function open(page: import("@playwright/test").Page) {
  await page.emulateMedia({ reducedMotion: "reduce" });
  await page.goto("/");
  await page.locator("#analytics").scrollIntoViewIfNeeded();
  const section = page.locator("#analytics");
  await expect(
    section.getByText("How email is doing across The Sunday Edit."),
  ).toBeVisible();
  return section;
}

test("the period buttons recalculate the dashboard numbers", async ({
  page,
}) => {
  const section = await open(page);
  const sent = section.locator('[data-stat="Sent"] [data-value]');
  const original = await sent.innerText();
  await section.getByRole("button", { name: "7 days", exact: true }).click();
  await expect(sent).not.toHaveText(original);
  await section.getByRole("button", { name: "30 days", exact: true }).click();
  await expect(sent).toHaveText(original);
});

test("the Email tab shows daily activity and the at-a-glance list", async ({
  page,
}) => {
  const section = await open(page);
  await section.getByRole("tab", { name: "Email" }).click();
  await expect(
    section.getByText("Email activity", { exact: true }),
  ).toBeVisible();
  await expect(section.getByText("At a glance")).toBeVisible();
  await expect(section.locator(".recharts-bar")).toHaveCount(3);
});

test("the dashboard demo has no serious accessibility violations", async ({
  page,
}) => {
  await open(page);
  const result = await new AxeBuilder({ page })
    .include("#analytics")
    .withTags(["wcag2a", "wcag2aa"])
    .analyze();
  expect(
    result.violations.filter(
      (v) => v.impact === "serious" || v.impact === "critical",
    ),
  ).toEqual([]);
});
