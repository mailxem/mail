// Captures the website's product screenshots from the client's dev-only
// /preview page (sample data, no account). Run the client with `next dev`
// first: node scripts/capture-product.mjs http://127.0.0.1:3000 out-dir [light|dark]
import { chromium } from "playwright";
import sharp from "sharp";

const [
  base = "http://127.0.0.1:3000",
  out = "public/images/product",
  theme = "dark",
] = process.argv.slice(2);
const pages = [
  ["analytics", "/dashboard"],
  ["newsletters", "/newsletters"],
  // The builder, opened from a starter journey.
  [
    "automations",
    "/automations",
    (p) => p.getByText("Welcome new subscribers").first().click(),
  ],
  ["crm", "/crm"],
  ["forms", "/forms"],
  ["inbox", "/inbox", (p) => p.getByText("Ava Thompson").first().click()],
];
const browser = await chromium.launch({ channel: "chrome", headless: true });
const context = await browser.newContext({
  // 26px taller than the shot: the preview's own "Local preview" bar is cropped off.
  viewport: { width: 1440, height: 1026 },
  deviceScaleFactor: 2,
  colorScheme: theme,
});
await context.addInitScript(
  (t) => localStorage.setItem("xem-appearance", t),
  theme,
);
const page = await context.newPage();
await page.goto(`${base}/preview`, { waitUntil: "networkidle" });
// Hide the Next.js dev overlay.
await page.addStyleTag({ content: "nextjs-portal{display:none!important}" });
for (const [name, path, open] of pages) {
  await page.locator(`a[href="${path}"]`).first().click();
  await page.waitForTimeout(1500);
  if (open) {
    await open(page);
    await page.waitForTimeout(1500);
  }
  const shot = await page.screenshot({
    clip: { x: 0, y: 26, width: 1440, height: 1000 },
  });
  await sharp(shot).webp({ quality: 82 }).toFile(`${out}/${name}.webp`);
  console.log(name);
}
await browser.close();
