import { test, expect } from "@playwright/test";
import AxeBuilder from "@axe-core/playwright";

test("product tour supports chapters, focus containment, and returning to its trigger", async ({
  page,
}) => {
  const errors: string[] = [];
  page.on("pageerror", (e) => errors.push(e.message));
  await page.goto("/");
  const trigger = page.getByRole("button", {
    name: "Take a little tour",
    exact: true,
  });
  await trigger.click();
  const dialog = page.getByRole("dialog");
  await expect(dialog).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Close preview" }),
  ).toBeFocused();
  await dialog.getByRole("button", { name: "Next chapter" }).click();
  await expect(
    dialog.getByRole("heading", { name: "Give good ideas a rhythm." }),
  ).toBeVisible();
  await dialog.getByRole("button", { name: /Automate/ }).click();
  await expect(dialog.getByRole("img")).toHaveAttribute("src", /automations/);
  for (let i = 0; i < 14; i++) {
    await page.keyboard.press("Tab");
    expect(
      await dialog.evaluate((e) => e.contains(document.activeElement)),
    ).toBe(true);
  }
  await page.keyboard.press("Escape");
  await expect(dialog).not.toBeVisible();
  await expect(trigger).toBeFocused();
  expect(await page.evaluate(() => document.body.style.overflow)).toBe("");
  expect(errors).toEqual([]);
});

test("feature tabs work with keyboard and expose the right product view", async ({
  page,
}) => {
  await page.goto("/");
  const tabs = page.locator("#product").getByRole("tablist");
  const design = tabs.getByRole("tab", { name: /Design/ });
  await design.focus();
  await page.keyboard.press("ArrowRight");
  await expect(tabs.getByRole("tab", { name: /Send/ })).toHaveAttribute(
    "aria-selected",
    "true",
  );
  await expect(page.locator("#product-panel").getByRole("img")).toHaveAttribute(
    "src",
    /newsletters/,
  );
  await tabs.getByRole("tab", { name: /Understand/ }).click();
  await expect(page.locator("#product-panel").getByRole("img")).toHaveAttribute(
    "src",
    /analytics/,
  );
});

test("template gallery scrolls and opens complete editable-template previews", async ({
  page,
}) => {
  await page.goto("/");
  const rail = page.getByRole("region", { name: "Template collection" });
  await page
    .getByRole("button", { name: "Next templates", exact: true })
    .click();
  await expect
    .poll(() => rail.evaluate((e) => e.scrollLeft))
    .toBeGreaterThan(100);
  await page
    .getByRole("button", { name: "Preview Show Your Work", exact: true })
    .click();
  const dialog = page.getByRole("dialog");
  await expect(dialog.getByRole("img")).toHaveAttribute(
    "src",
    /figma-show-work-full/,
  );
  await expect(
    dialog.getByRole("link", { name: "Use this template" }),
  ).toHaveAttribute("href", /\/templates\/new\?starter=figma-show-work$/);
  await dialog.getByRole("button", { name: "Next template preview" }).click();
  await expect(
    dialog.getByRole("heading", { name: "Meet the Leaders", exact: true }),
  ).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(dialog).not.toBeVisible();
});

test("mobile navigation, FAQ, and previews fit narrow screens", async ({
  page,
}) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto("/");
  const menu = page.getByRole("button", { name: "Open menu" });
  await menu.click();
  await expect(
    page.getByRole("button", { name: "Close menu" }),
  ).toHaveAttribute("aria-expanded", "true");
  await page
    .getByRole("navigation", { name: "Mobile navigation" })
    .getByRole("link", { name: "FAQs" })
    .click();
  await expect(
    page.getByRole("navigation", { name: "Mobile navigation" }),
  ).not.toBeVisible();
  const question = page.getByRole("button", {
    name: "Are the templates actually editable?",
  });
  await question.click();
  await expect(question).toHaveAttribute("aria-expanded", "true");
  await expect(page.locator("#faq-mobile-2")).toBeVisible();
  await page
    .getByRole("button", { name: "Preview Thanks for Digging In", exact: true })
    .click();
  await expect(page.getByRole("dialog")).toBeVisible();
  expect(
    await page
      .getByRole("dialog")
      .evaluate((e) => e.scrollWidth <= e.clientWidth),
  ).toBe(true);
  await page.keyboard.press("Escape");
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  ).toBe(true);
  await page.setViewportSize({ width: 320, height: 740 });
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  ).toBe(true);
});

test("reduced motion retains all content without initializing WebGL", async ({
  page,
}) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  await page.goto("/");
  await expect(page.getByRole("heading", { level: 1 })).toBeVisible();
  await expect(page.locator("canvas")).toHaveCount(0);
  await page.locator("#templates").scrollIntoViewIfNeeded();
  await expect(
    page.getByRole("heading", { name: /A head start/ }),
  ).toBeVisible();
  await page
    .getByRole("button", { name: "Preview Thanks for Digging In", exact: true })
    .click();
  await expect(page.getByRole("dialog")).toBeVisible();
});

test("page, mobile FAQ, and dialog have no serious accessibility violations", async ({
  page,
}) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  await page.goto("/");
  const audit = async () => {
    const result = await new AxeBuilder({ page })
      .withTags(["wcag2a", "wcag2aa", "wcag21aa"])
      .analyze();
    expect(result.violations).toEqual([]);
  };
  await audit();
  await page
    .getByRole("button", { name: "Take a little tour", exact: true })
    .click();
  await audit();
  await page.keyboard.press("Escape");
  await page.setViewportSize({ width: 390, height: 844 });
  await page
    .getByRole("button", { name: "Can I use my own email provider?" })
    .click();
  await audit();
});

test("public metadata, headers, links, and every image resolve", async ({
  page,
  request,
}) => {
  const response = await page.goto("/");
  expect(response?.headers()["x-content-type-options"]).toBe("nosniff");
  expect(response?.headers()["x-frame-options"]).toBe("DENY");
  await expect(page).toHaveTitle("Xem — Every email, a little more human.");
  const hrefs = await page
    .locator("a")
    .evaluateAll((es) => es.map((e) => e.getAttribute("href") || ""));
  expect(hrefs.filter((h) => !h || h.startsWith("javascript:"))).toEqual([]);
  for (const hash of new Set(
    hrefs.filter((h) => h.startsWith("#") && h.length > 1),
  )) {
    expect(await page.locator(hash).count()).toBe(1);
  }
  for (const path of ["/robots.txt", "/sitemap.xml", "/opengraph-image.png"]) {
    const r = await request.get(path);
    expect(r.ok()).toBe(true);
  }
  await page.evaluate(async () => {
    const images = Array.from(document.images);
    await Promise.all(
      images.map((img) => {
        img.loading = "eager";
        return img.decode().catch(() => {});
      }),
    );
  });
  expect(
    await page
      .locator("img")
      .evaluateAll((es) =>
        es
          .filter(
            (e) =>
              e instanceof HTMLImageElement &&
              (!e.complete || e.naturalWidth === 0),
          )
          .map((e) => e.getAttribute("src")),
      ),
  ).toEqual([]);
});

test("content and signup remain available when JavaScript is unavailable", async ({
  browser,
}) => {
  const context = await browser.newContext({ javaScriptEnabled: false });
  const page = await context.newPage();
  await page.goto(process.env.PLAYWRIGHT_BASE_URL || "http://localhost:3002");
  await expect(page.getByRole("heading", { level: 1 })).toBeVisible();
  await expect(
    page.getByRole("link", { name: "Create your first email" }),
  ).toHaveAttribute("href", /\/auth\/register$/);
  await expect(
    page.getByRole("link", { name: "Play the Xem brand film, 30 seconds" }),
  ).toHaveAttribute("href", "/videos/xem-brand-film.mp4");
  await expect(page.locator("#hero video")).toHaveAttribute(
    "poster",
    "/videos/xem-brand-poster.webp",
  );
  await expect(
    page.getByRole("heading", { name: /Your whole email world/ }),
  ).toBeVisible();
  await context.close();
});
