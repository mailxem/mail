import { test, expect } from "@playwright/test";
import AxeBuilder from "@axe-core/playwright";
import fs from "node:fs/promises";
async function open(page: import("@playwright/test").Page) {
  await page.emulateMedia({ reducedMotion: "reduce" });
  await page.goto("/");
  await page.locator("#analytics").scrollIntoViewIfNeeded();
  await expect(
    page.getByRole("heading", { name: "Your audience, in focus." }),
  ).toBeVisible();
  return page.locator("#analytics");
}

test("analytics filters recalculate charts and export the selected demo records", async ({
  page,
}) => {
  const section = await open(page);
  const metric = section.locator('[data-stat="Messages accepted"]');
  const original = await metric.innerText();
  await section.getByRole("button", { name: "7 days", exact: true }).click();
  await expect(metric).not.toHaveText(original);
  await section
    .getByLabel("Filter analytics by campaign")
    .selectOption("show-work");
  await expect(
    section.getByRole("button", { name: "Show all six campaigns" }),
  ).toBeVisible();
  await section
    .getByText("View the chart data as a table", { exact: true })
    .click();
  const data = section.getByRole("table", {
    name: "Daily values for the selected period and campaign",
  });
  await expect(data.getByRole("row")).toHaveCount(8);
  const messages = await data
    .locator("tbody tr td:nth-child(2)")
    .allTextContents();
  const sum = messages.reduce((n, s) => n + Number(s.replaceAll(",", "")), 0);
  expect(Number((await metric.innerText()).replaceAll(",", ""))).toBe(sum);
  const downloaded = page.waitForEvent("download");
  await section.getByRole("button", { name: "Export demo data" }).click();
  const file = await downloaded;
  const contents = await fs.readFile((await file.path())!, "utf8");
  expect(contents).toContain("synthetic demo analytics");
  expect(contents).toContain("Show Your Work");
  expect(contents).not.toContain("The Sunday Edit");
  await section
    .getByRole("button", { name: "Reset analytics filters" })
    .click();
  await expect(metric).toHaveText(original);
  await expect(section.getByLabel("Filter analytics by campaign")).toHaveValue(
    "all",
  );
});

test("stacked series, metric switch, device chart, heatmap and table are interactive", async ({
  page,
}) => {
  const section = await open(page);
  const bars = section.locator(".recharts-bar");
  await expect(bars).toHaveCount(3);
  await section
    .getByRole("button", { name: "Toggle newsletters series" })
    .click();
  await expect(bars).toHaveCount(2);
  await section
    .getByRole("button", { name: "Toggle newsletters series" })
    .click();
  await expect(bars).toHaveCount(3);
  await section
    .getByRole("button", { name: "Clicks", exact: true })
    .first()
    .click();
  await expect(
    section.getByText(
      "Tracked clicks, stacked by send type. Hover to explore.",
    ),
  ).toBeVisible();
  await section.getByRole("button", { name: /Mobile/ }).click();
  await expect(section.getByRole("button", { name: /Mobile/ })).toHaveAttribute(
    "aria-pressed",
    "true",
  );
  await section
    .getByRole("button", { name: "All devices", exact: true })
    .click();
  const cell = section.getByRole("button", { name: /Mon 08–12 UTC/ });
  await cell.click();
  await expect(cell).toHaveAttribute("aria-pressed", "true");
  await expect(
    section.getByText("Mon, 08–12 UTC", { exact: true }),
  ).toBeVisible();
  await section
    .getByRole("button", { name: "Click rate", exact: true })
    .click();
  await expect(
    section.getByRole("columnheader", { name: "Click rate" }),
  ).toHaveAttribute("aria-sort", "descending");
  const rect = section.locator(".recharts-bar-rectangle path").nth(8);
  await rect.hover();
  await expect(
    section.locator(".recharts-tooltip-wrapper").first(),
  ).toBeVisible();
});

test("live charts remain accessible and fit mobile screens", async ({
  page,
}) => {
  await page.setViewportSize({ width: 390, height: 844 });
  const section = await open(page);
  await section.getByRole("button", { name: "90 days", exact: true }).click();
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  ).toBe(true);
  for (const chart of await section.locator("[data-chart]").all()) {
    const dimensions = await chart.evaluate((e) => ({
      outer: e.clientWidth,
      svg: e.querySelector("svg")?.getBoundingClientRect().width || 0,
    }));
    expect(dimensions.svg).toBeLessThanOrEqual(dimensions.outer + 1);
  }
  const result = await new AxeBuilder({ page })
    .include("#analytics")
    .withTags(["wcag2a", "wcag2aa", "wcag21aa"])
    .analyze();
  expect(result.violations).toEqual([]);
});
