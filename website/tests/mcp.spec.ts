import { test, expect } from "@playwright/test";
import AxeBuilder from "@axe-core/playwright";

test("MCP page is discoverable, indexable and usable on mobile", async ({
  page,
}) => {
  await page.goto("/");
  await page.getByRole("link", { name: "Explore the Xem MCP server" }).click();
  await expect(page).toHaveURL(/\/mcp\/?$/);
  await expect(page).toHaveTitle(/Xem MCP Server/);
  await expect(page.locator('link[rel="canonical"]')).toHaveAttribute(
    "href",
    "https://xem.email/mcp",
  );
  await expect(page.getByRole("heading", { level: 1 })).toHaveCount(1);
  await expect(
    page.getByText("Hosted · Recommended", { exact: true }),
  ).toBeVisible();
  await expect(
    page.getByText("Local · npm + stdio", { exact: true }),
  ).toBeVisible();
  await expect(page.locator("#setup")).toContainText(
    "https://mcp.xem.email/mcp",
  );
  await expect(page.locator("#setup")).toContainText("@xem.email/mcp@2");
  const jsonld = await page
    .locator('script[type="application/ld+json"]')
    .textContent();
  expect(JSON.parse(jsonld!)["@graph"][0]["@type"]).toBe("SoftwareSourceCode");
  await page
    .getByText("Does creating a newsletter send an email?", { exact: true })
    .click();
  await expect(
    page.getByText(/No. Newsletter and campaign creation produce drafts/),
  ).toBeVisible();
  expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);
  await page.screenshot({
    path: "test-results/mcp-desktop.png",
    fullPage: true,
  });
  await page.setViewportSize({ width: 390, height: 844 });
  await expect(page.getByRole("heading", { level: 1 })).toBeVisible();
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= window.innerWidth,
    ),
  ).toBe(true);
  await page.screenshot({
    path: "test-results/mcp-mobile.png",
    fullPage: true,
  });
});
