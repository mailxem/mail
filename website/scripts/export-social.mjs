import sharp from "sharp";
import fs from "node:fs/promises";
import { fileURLToPath } from "node:url";
const root = fileURLToPath(new URL("../", import.meta.url));
const mark = (await fs.readFile(root + "public/brand/xem-mark.png")).toString(
  "base64",
);
const svg = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="1200" height="630" viewBox="0 0 1200 630"><rect width="1200" height="630" fill="#ffffef"/><path d="M-100 500 Q200 340 350 540 T1300 360" fill="none" stroke="#a4b28a" stroke-width="2"/><image x="72" y="58" width="48" height="48" href="data:image/png;base64,${mark}"/><text x="134" y="96" font-size="40" font-family="Arial" fill="#22251f" letter-spacing="-2">Xem</text><text x="72" y="178" font-size="13" font-family="Arial" fill="#4e6250" letter-spacing="3">A THOUGHTFUL HOME FOR YOUR EMAIL MARKETING</text><text x="65" y="304" font-size="112" font-family="Georgia" fill="#22251f" letter-spacing="-6">Every email,</text><text x="68" y="414" font-size="106" font-family="Georgia" font-style="italic" fill="#22251f" letter-spacing="-6">a little more human.</text><rect x="74" y="473" width="220" height="53" rx="8" fill="#e7d8fa" stroke="#22251f"/><text x="100" y="507" font-size="18" font-family="Arial" fill="#22251f">Design. Connect. Grow.</text><text x="1042" y="568" font-size="16" font-family="Arial" fill="#4e6250">xem.email</text></svg>`;
await sharp(Buffer.from(svg))
  .png()
  .toFile(root + "app/opengraph-image.png");
