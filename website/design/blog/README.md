# Xem journal artwork

Ten original, AI-assisted vector compositions, replacing the launch stock photos. Created for Xem on September 10, 2026. The illustrations contain no customer data; chart values are decorative.

- Palette: cream #ffffef, iris #5b3cc4, forest #164b3f, lavender #e7d8fa, lemon #edf09b, ink #22251f.
- Type: EB Garamond and DM Sans, matching the website. Latin font files came from the site's Next.js font build; both use the accompanying SIL Open Font License.
- Source: `scripts/generate-blog-covers.mjs` and editable SVGs in `design/blog/svg/`.
- Delivery: 1600 × 1067 WebP files under `public/images/blog/`, shared by article cards, article heroes, homepage stories, and social metadata.
- Rebuild: `bun run generate:blog-covers` (requires installed Google Chrome). The script uses Playwright for font-accurate SVG rendering and Sharp for WebP encoding; no running application server is needed.
- Review: `contact-sheet.png` shows all ten compositions.

The earlier photo sourcing notes in `design/research/` document the launch design, not the current cover artwork.
