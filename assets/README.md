# Repository branding

`xem-social.png` is an unchanged copy of `website/app/opengraph-image.png` from website commit `ae68d0c3c5d9339834d58c43eb99f6494d5b903a`.

The website is the source of truth for Xem's branding:

- Mark: `website/public/brand/xem-mark.png`
- Social-card generator: `website/scripts/export-social.mjs`
- Colors and typography: `website/tailwind.config.ts`
- Button treatment: `website/lib/brand.ts`

Keep the README aligned with those assets. A local copy is necessary because GitHub does not resolve README image paths through a Git submodule. When the website's social card changes, copy the asset from the parent repository's newly pinned website revision and update this provenance note.
