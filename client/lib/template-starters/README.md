# Xem email template library

218 editable templates: 20 reference-inspired designs, 18 Xem originals, and 180 editorial compositions. The old 180 industry reskins have been removed. Each current template has a different native block structure; editorial designs also have individual briefs, palettes, image choices, proportions, and reading order. Components are shared by the generator, but complete designs are not palette-only copies.

## Source and generation

- `scripts/generate_template_starters.py`: native Unlayer schema 18 and static gallery previews.
- `scripts/reference_templates.py`: layouts based on the supplied screenshots and linked emails.
- `scripts/composed_template_collection.py`: 180 editorial briefs and composition recipes.
- `scripts/crop-template-references.ts`: image crops from retained source screenshots. Temporary clipboard paths are optional once a source screenshot has been saved.
- `public/assets/template-starters/<key>/design.json`: editable rows, columns, headings, paragraphs, images, buttons, and dividers.
- `public/assets/template-starters/<key>/preview.html`: a lightweight visual preview generated from the same blocks.
- `manifest.json`: metadata only. `registry.ts` fetches the selected design on demand.

Regenerate from the client directory:

```sh
bun scripts/crop-template-references.ts
python3 scripts/generate_template_starters.py
bunx jest --runInBand
bunx tsc --noEmit --incremental false
bun run build
```

Native IDs and counters are generated deterministically. Tests validate all assets, allowed block types, IDs, counters, unique structural signatures, filtering, unsafe keys, empty responses, and local editor asset round trips. Review contact sheets under `design-references/email-library/review/` as well: structure checks cannot judge visual quality.

## Editor and gallery

The gallery reveals 12 templates at a time through an intersection observer, with lazy thumbnail iframes and an accessible load-more fallback. Search, category, and collection changes reset the batch. The starter library does not require a marketing API request.

Use-template opens the existing Unlayer editor. Name and subject are prefilled; users edit a copy and save through the existing template endpoint. Delivery HTML is exported by Unlayer, not by the lightweight gallery renderer. Native designs are hydrated once so query refreshes do not erase edits. Save/export waits for `design:loaded`.

On localhost, Chrome prevents the remote Unlayer iframe from fetching loopback images. `editor-assets.ts` reads only Xem's same-origin raster assets into temporary data URLs, then restores their absolute URLs during export. Production origins use normal URLs directly. This does not change browser security settings. Deploy the public assets alongside the application; images saved from localhost are not suitable for a real send.

## Content and references

Reference layouts are editable reconstructions or interpretations, not verified pixel-identical copies. Source screenshots and URLs are retained. Raster illustrations and logos remain replaceable image blocks; text and actions are native editable blocks. Stock photos use Unsplash URLs. Replace sample brand names, images, claims, addresses, dates, and example.com links before sending. Quiz buttons are links to editable destinations, not an embedded quiz application.

Every design includes Built with Xem. Marketing delivery paths add the actual recipient unsubscribe link. Do not invent an unsupported unsubscribe merge variable in the generator.

## AI and personal messages

Template AI requests `format: "design"` from the existing authenticated `marketing/email-draft` endpoint. The Go service asks the configured model for a bounded layout description and compiles it into schema-18 Unlayer JSON plus a safe preview. Text is escaped; colors, fonts, sizes, block counts, columns, and HTTPS URLs are validated. The model cannot inject raw HTML or call delivery tools.

Outbox AI uses `format: "text"`. The original Maily rich-message editor supports formatting and images, and its renderer exports the email HTML. Images suggested by the text assistant must come from HTTPS URLs included by the user. Drafting never sends a message.

Deploy the updated Go email-writing service to enable design mode against a deployed API. The existing server configuration remains `AI_PROXY_BASE_URL`, `AI_PROXY_API_KEY` (optional for the supplied open proxy), and `AI_PROXY_MODEL` (default `gpt-4.1`). Provider calls have bounded output, a timeout, no redirects, and the existing authenticated per-team rate limit.
