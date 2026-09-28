import { chromium } from "playwright";
import sharp from "sharp";
import fs from "node:fs/promises";
import { fileURLToPath } from "node:url";
const root = fileURLToPath(new URL("../", import.meta.url));
const templates = JSON.parse(
  await fs.readFile(root + "lib/templates.json", "utf8"),
);
const app = process.env.XEM_CAPTURE_APP_URL || "http://localhost:3000";
const browser = await chromium.launch({ channel: "chrome", headless: true });
const page = await browser.newPage({ viewport: { width: 640, height: 860 } });
await fs.mkdir(root + "design/template-captures", { recursive: true });
for (const t of templates) {
  await page.goto(`${app}/assets/template-starters/${t.key}/preview.html`, {
    waitUntil: "networkidle",
  });
  await page.evaluate(async () => {
    await document.fonts.ready;
    await Promise.all(
      Array.from(document.images).map((i) => i.decode().catch(() => {})),
    );
  });
  const broken = await page
    .locator("img")
    .evaluateAll((es) =>
      es.filter((i) => !i.complete || i.naturalWidth === 0).map((i) => i.src),
    );
  if (broken.length)
    throw new Error(`${t.key}: ${broken.length} broken images`);
  const path = `${root}design/template-captures/${t.key}.png`;
  await page.screenshot({ path, fullPage: true });
  await sharp(path)
    .resize({ width: 640 })
    .webp({ quality: 86 })
    .toFile(`${root}public/images/templates/${t.key}-full.webp`);
  await sharp(path)
    .resize({ width: 640 })
    .extract({ left: 0, top: 0, width: 640, height: 820 })
    .webp({ quality: 86 })
    .toFile(`${root}public/images/templates/${t.key}.webp`);
  console.log(`Captured ${t.name}`);
}
await browser.close();
