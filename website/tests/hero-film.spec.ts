import { expect, test } from "@playwright/test";
import AxeBuilder from "@axe-core/playwright";

test("film preview pauses offscreen and the full player loads on request", async ({
  page,
}) => {
  const requests: string[] = [];
  const errors: string[] = [];
  page.on("request", (request) => requests.push(request.url()));
  page.on("pageerror", (error) => errors.push(error.message));
  await page.goto("/");
  const preview = page.locator('#hero video[aria-hidden="true"]');
  await expect(preview).toHaveJSProperty("paused", false);
  await expect(preview).toHaveJSProperty("muted", true);
  expect(requests.some((url) => url.endsWith("/xem-brand-film.mp4"))).toBe(
    false,
  );

  await page.locator("#templates").scrollIntoViewIfNeeded();
  await expect(preview).toHaveJSProperty("paused", true);
  await page.locator("#hero").scrollIntoViewIfNeeded();
  await expect(preview).toHaveJSProperty("paused", false);
  await page.getByRole("button", { name: "Pause film preview" }).click();
  await expect(preview).toHaveJSProperty("paused", true);
  await page.locator("#templates").scrollIntoViewIfNeeded();
  await page.locator("#hero").scrollIntoViewIfNeeded();
  await expect(
    page.getByRole("button", { name: "Play film preview" }),
  ).toBeVisible();
  await expect(preview).toHaveJSProperty("paused", true);
  await page.getByRole("button", { name: "Play film preview" }).click();
  await expect(preview).toHaveJSProperty("paused", false);

  const trigger = page.getByRole("link", {
    name: "Play the Xem brand film, 30 seconds",
  });
  await trigger.click();
  const dialog = page.getByRole("dialog");
  await expect(dialog).toBeVisible();
  await expect(preview).toHaveJSProperty("paused", true);
  await expect(dialog.locator("video")).toHaveJSProperty("controls", true);
  await expect(dialog.locator("video")).toHaveJSProperty("muted", false);
  await expect(dialog.locator("video")).toHaveJSProperty("paused", false);
  expect(requests.some((url) => url.endsWith("/xem-brand-film.mp4"))).toBe(
    true,
  );

  await expect(
    dialog.getByRole("button", { name: "Close preview" }),
  ).toBeFocused();
  await page.keyboard.press("Shift+Tab");
  await expect(
    dialog.getByRole("link", { name: "Download film" }),
  ).toBeFocused();
  await page.keyboard.press("Tab");
  await expect(
    dialog.getByRole("button", { name: "Close preview" }),
  ).toBeFocused();
  await page.keyboard.press("Escape");
  await expect(dialog).toHaveCount(0);
  await expect(trigger).toBeFocused();
  await expect(preview).toHaveJSProperty("paused", false);

  await page
    .getByRole("button", { name: "Take a little tour", exact: true })
    .click();
  await expect(preview).toHaveJSProperty("paused", true);
  await page.keyboard.press("Escape");
  expect(errors).toEqual([]);
});

for (const preference of ["reduced motion", "save data"]) {
  test(`${preference} keeps the poster until playback is requested`, async ({
    page,
  }) => {
    if (preference === "reduced motion") {
      await page.emulateMedia({ reducedMotion: "reduce" });
    } else {
      await page.addInitScript(() => {
        Object.defineProperty(navigator, "connection", {
          value: Object.assign(new EventTarget(), { saveData: true }),
          configurable: true,
        });
      });
    }
    const mediaRequests: string[] = [];
    page.on("request", (request) => {
      if (request.url().endsWith(".mp4")) mediaRequests.push(request.url());
    });
    await page.goto("/");
    const play = page.getByRole("button", { name: "Play film preview" });
    await expect(play).toBeVisible();
    const preview = page.locator('#hero video[aria-hidden="true"]');
    await expect(preview).not.toHaveAttribute("src");
    expect(mediaRequests).toEqual([]);
    await play.click();
    await expect(preview).toHaveJSProperty("paused", false);
  });
}

test("a failed film request leaves a direct link and a working close button", async ({
  page,
}) => {
  await page.route("**/videos/xem-brand-film.mp4", (route) =>
    route.fulfill({ status: 503 }),
  );
  await page.goto("/");
  const trigger = page.getByRole("link", {
    name: "Play the Xem brand film, 30 seconds",
  });
  await trigger.click();
  const dialog = page.getByRole("dialog");
  await expect(dialog.getByRole("status")).toContainText(
    "The film couldn’t load.",
  );
  await expect(
    dialog.getByRole("link", { name: "Open the video directly." }),
  ).toHaveAttribute("href", "/videos/xem-brand-film.mp4");
  await dialog.getByRole("button", { name: "Close preview" }).click();
  await expect(dialog).toHaveCount(0);
  await expect(trigger).toBeFocused();
});

test("film card and accessible player fit a 320px screen", async ({
  page,
  request,
}) => {
  await page.setViewportSize({ width: 320, height: 740 });
  await page.emulateMedia({ reducedMotion: "reduce" });
  await page.goto("/");
  const trigger = page.getByRole("link", {
    name: "Play the Xem brand film, 30 seconds",
  });
  await trigger.scrollIntoViewIfNeeded();
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  ).toBe(true);
  await trigger.click();
  const dialog = page.getByRole("dialog");
  await expect(dialog).toBeVisible();
  expect(
    await dialog.evaluate(
      (element) => element.scrollWidth <= element.clientWidth,
    ),
  ).toBe(true);
  await dialog.locator("summary").click();
  await expect(dialog.locator("#film-description")).toBeVisible();
  const audit = await new AxeBuilder({ page })
    .withTags(["wcag2a", "wcag2aa", "wcag21aa"])
    .analyze();
  expect(audit.violations).toEqual([]);
  const player = dialog.locator("video");
  await expect
    .poll(() => player.evaluate((video: HTMLVideoElement) => video.readyState))
    .toBeGreaterThanOrEqual(2);
  await player.evaluate((video: HTMLVideoElement) => {
    video.pause();
    video.currentTime = 20;
  });
  await expect(player).toHaveJSProperty("seeking", false);
  await expect(player).toHaveJSProperty("currentTime", 20);
  await dialog.getByRole("button", { name: "Close preview" }).click();
  await expect(dialog).toHaveCount(0);

  const response = await request.get("/videos/xem-brand-film.mp4");
  expect(response.ok()).toBe(true);
  expect(response.headers()["content-type"]).toContain("video/mp4");
  expect((await request.get("/videos/xem-brand-film.vtt")).ok()).toBe(true);
});
