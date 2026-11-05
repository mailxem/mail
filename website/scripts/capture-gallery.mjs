// Builds the public /templates gallery from the app's starter manifest.
// Serve client/public first, e.g. `python3 -m http.server 4100 -d ../client/public`.
import { chromium } from "playwright";
import sharp from "sharp";
import fs from "node:fs/promises";
import { fileURLToPath } from "node:url";
const root = fileURLToPath(new URL("../", import.meta.url));
const manifest = JSON.parse(
  await fs.readFile(
    root + "../client/lib/template-starters/manifest.json",
    "utf8",
  ),
);
// Only the hand-built "Xem originals" are published. "Reference designs"
// recreate other companies' emails, the generated "Editorial collection" is
// being redesigned, and the `xem-*` starters are Xem's own account notices.
const gallery = manifest
  .filter((t) => t.collection === "Xem originals")
  .filter((t) => !t.key.startsWith("xem-"))
  .map(
    ({
      key,
      name,
      category,
      description,
      subject,
      preheader,
      tags,
      collection,
    }) => ({
      key,
      name,
      category,
      description,
      subject,
      preheader,
      tags,
      collection,
    }),
  );
const app = process.env.XEM_CAPTURE_APP_URL || "http://localhost:4100";
const out = root + "public/images/gallery/";
await fs.mkdir(out, { recursive: true });
const browser = await chromium.launch({ channel: "chrome", headless: true });
const page = await browser.newPage({ viewport: { width: 640, height: 860 } });
for (const t of gallery) {
  if (!process.env.FORCE) {
    try {
      await fs.access(`${out}${t.key}-full.webp`);
      continue;
    } catch {}
  }
  await page.goto(`${app}/assets/template-starters/${t.key}/preview.html`, {
    waitUntil: "load",
    timeout: 60000,
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
  if (broken.length) throw new Error(`${t.key}: broken images ${broken}`);
  const png = await page.screenshot({ fullPage: true });
  await sharp(png)
    .resize({ width: 640 })
    .webp({ quality: 70 })
    .toFile(`${out}${t.key}-full.webp`);
  await sharp(png)
    .resize({ width: 400 })
    .extract({ left: 0, top: 0, width: 400, height: 500 })
    .webp({ quality: 68 })
    .toFile(`${out}${t.key}.webp`);
  console.log(`Captured ${t.name}`);
}
await browser.close();
for (const t of gallery)
  t.height = (await sharp(`${out}${t.key}-full.webp`).metadata()).height;
await fs.writeFile(
  root + "lib/template-gallery.json",
  JSON.stringify(gallery, null, 2) + "\n",
);
console.log(`${gallery.length} templates`);
