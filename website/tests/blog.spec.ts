import { test, expect } from "@playwright/test";
import AxeBuilder from "@axe-core/playwright";
import fs from "node:fs";
import path from "node:path";
import matter from "gray-matter";
const posts = fs
  .readdirSync(path.join(process.cwd(), "content/blog"))
  .filter((p) => p.endsWith(".md"))
  .map((p) => ({
    slug: p.slice(0, -3),
    ...matter(
      fs.readFileSync(path.join(process.cwd(), "content/blog", p), "utf8"),
    ),
  }));

test("journal contains complete published Markdown stories and filters by topic and search", async ({
  page,
}) => {
  expect(posts.length).toBeGreaterThan(0);
  for (const p of posts)
    expect(p.content.split(/\s+/).length).toBeGreaterThan(700);
  await page.goto("/blog");
  await expect(page.locator("main article")).toHaveCount(posts.length);
  await page.getByRole("button", { name: "AI & writing", exact: true }).click();
  await expect(page.locator("main article")).toHaveCount(2);
  await page.getByRole("button", { name: "All stories", exact: true }).click();
  await page
    .getByRole("searchbox", { name: "Search articles" })
    .fill("deliverability");
  await expect(page.locator("main article")).toHaveCount(1);
  await page
    .getByRole("searchbox", { name: "Search articles" })
    .fill("unfindabletopic");
  await expect(page.getByText("No stories here just yet.")).toBeVisible();
  await page.getByRole("button", { name: "Clear filters" }).click();
  await expect(page.locator("main article")).toHaveCount(posts.length);
});

test("every article has unique metadata, structured data, working contents, and a cover", async ({
  page,
  request,
}) => {
  test.setTimeout(90000);
  const titles = new Set<string>();
  for (const p of posts) {
    const response = await page.goto(`/blog/${p.slug}`);
    expect(response?.status()).toBe(200);
    await expect(page.locator("h1")).toHaveCount(1);
    titles.add(await page.title());
    expect(
      await page.locator("link[rel=canonical]").getAttribute("href"),
    ).toMatch(new RegExp(`/blog/${p.slug}$`));
    expect(
      await page.locator("meta[name=description]").getAttribute("content"),
    ).toBe(p.data.description);
    expect(
      await page
        .locator('meta[property="og:image"]')
        .first()
        .getAttribute("content"),
    ).toContain(p.data.cover);
    expect(
      await page.locator('meta[property="og:type"]').getAttribute("content"),
    ).toBe("article");
    const json = JSON.parse(
      await page
        .locator('script[type="application/ld+json"]')
        .last()
        .innerText(),
    );
    expect(json["@graph"][0].headline).toBe(p.data.title);
    expect(json["@graph"][0].author.name).toBe("Xem editorial");
    const toc = await page
      .getByRole("navigation", { name: "Table of contents" })
      .locator("a")
      .evaluateAll((es) => es.map((e) => e.getAttribute("href")!));
    for (const hash of toc) expect(await page.locator(hash).count()).toBe(1);
    const image = page.locator("figure img");
    await expect
      .poll(() =>
        image.evaluate(
          (e: HTMLImageElement) => e.complete && e.naturalWidth > 0,
        ),
      )
      .toBe(true);
    expect((await request.get(p.data.cover)).ok()).toBe(true);
    const internal = await page
      .locator('.prose a[href^="/blog/"]')
      .evaluateAll((es) => es.map((e) => e.getAttribute("href")!));
    for (const href of new Set(internal))
      expect(
        posts.some((p) => href === `/blog/${p.slug}`) ||
          href === "/blog/editorial",
      ).toBe(true);
  }
  expect(titles.size).toBe(posts.length);
  const sitemap = await (await request.get("/sitemap.xml")).text();
  const feed = await (await request.get("/feed.xml")).text();
  expect(feed.match(/<item>/g)).toHaveLength(posts.length);
  for (const p of posts) {
    expect(sitemap).toContain(`/blog/${p.slug}`);
    expect(feed).toContain(`/blog/${p.slug}`);
  }
  expect((await request.get("/blog/not-a-real-post")).status()).toBe(404);
});

test("feature stories, testimonials, and iris calls to action work", async ({
  page,
}) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  await page.goto("/");
  const features = page.getByRole("tablist", { name: "More Xem features" });
  await features.getByRole("tab", { name: "Lead forms" }).focus();
  await page.keyboard.press("ArrowRight");
  await expect(
    features.getByRole("tab", { name: "Audience & CRM" }),
  ).toBeFocused();
  await expect(page.locator("#feature-panel img")).toHaveAttribute(
    "src",
    /crm/,
  );
  await features.getByRole("tab", { name: "Inbox & outbox" }).click();
  await expect(page.locator("#feature-panel img")).toHaveAttribute(
    "src",
    /inbox/,
  );
  await expect
    .poll(() =>
      page
        .locator("#feature-panel img")
        .evaluate((e: HTMLImageElement) => e.complete && e.naturalWidth > 0),
    )
    .toBe(true);
  await page.getByRole("button", { name: "Next testimonial" }).click();
  await expect(page.locator("#voices blockquote")).toContainText(
    "three different vendors",
  );
  await page.getByRole("button", { name: /Vivek Kaushik/ }).click();
  await expect(page.locator("#voices blockquote")).toContainText(
    "clean and intuitive",
  );
  await expect(page.locator("#journal article")).toHaveCount(3);
  const cta = page
    .getByRole("link", { name: "Get started", exact: false })
    .first();
  expect(await cta.evaluate((e) => getComputedStyle(e).backgroundColor)).toBe(
    "rgb(91, 60, 196)",
  );
});

test("journal and articles remain accessible on desktop and narrow phones", async ({
  page,
}) => {
  test.setTimeout(90000);
  await page.emulateMedia({ reducedMotion: "reduce" });
  for (const route of ["/blog", "/blog/email-deliverability"]) {
    await page.goto(route);
    for (const width of [1440, 390, 320]) {
      await page.setViewportSize({ width, height: 900 });
      expect(
        await page.evaluate(
          () => document.documentElement.scrollWidth <= innerWidth,
        ),
      ).toBe(true);
      const axe = await new AxeBuilder({ page })
        .withTags(["wcag2a", "wcag2aa", "wcag21aa"])
        .analyze();
      expect(axe.violations).toEqual([]);
    }
  }
});

test("article text and navigation render without JavaScript", async ({
  browser,
}) => {
  const context = await browser.newContext({ javaScriptEnabled: false });
  const page = await context.newPage();
  await page.goto(
    `${process.env.PLAYWRIGHT_BASE_URL || "http://localhost:3002"}/blog/ai-email-marketing`,
  );
  await expect(page.locator("h1")).toContainText("AI email marketing");
  await expect(
    page.getByRole("navigation", { name: "Table of contents" }),
  ).toBeVisible();
  await expect(page.locator(".prose")).toContainText("Start with a decision");
  await context.close();
});
