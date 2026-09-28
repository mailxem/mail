> Delivery update: the user approved this plan. The video is rendered; see [DELIVERY.md](DELIVERY.md) for implementation details, verification, and known fidelity differences.

# Source ledger

## Reference film

The user supplied `Motionimo_-_Cyera_Brand_Video_by_Ido_Sapir_amp_Team_6khr40.mp4` from Downloads. The local planning copy is `reference/source.mp4`; all 901 decoded frames are under `reference/frames/`. Source PTS values come from ffprobe and agree with `frame / 30` within 0.000000334 seconds.

The full-sequence contact sheets were visually inspected. Automated scene-detection scores helped locate cuts, but scene boundaries were assigned from the visible frame sequence. The soundtrack has not been audibly reviewed in this planning pass; audio production notes are proposals.

## Current Xem content

The live https://xem.email HTML and this website checkout were read on 2026-09-26.

| Source | Evidence used |
| --- | --- |
| `components/home.tsx`, live homepage hero | “Every email, a little more human”; open-source email marketing for real connections |
| `lib/content.ts` | Design, Send, Automate, Understand; reusable templates, newsletter cadence/timezones, triggers/delays/conditions, audience analytics |
| `components/home-expansion.tsx` | Forms, audience/CRM, IMAP inbox and outbox |
| `components/mcp-section.tsx` | MCP-assisted draft and contact workflows |
| `components/assistant-showcase.tsx`, live homepage | Ask Xem is labelled “In development”; demonstration is a scripted example |
| `components/sending-section.tsx`, live homepage | Managed sending is described as available with domain checks and workspace approval; own SMTP provider remains supported |
| `tailwind.config.ts`, `lib/brand.ts` | Exact Xem palette |
| `app/layout.tsx` | DM Sans and EB Garamond |
| `public/brand/xem-mark.png` | Current connected circle/square mark, visually inspected |
| `public/images/product/`, `public/images/templates/` | Existing product captures and template previews; staged copies under `assets/` |

The README's older “coming soon” guidance for managed sending differs from the current live page and component copy. The film avoids making a managed-sending availability claim. Live marketing text is not production-runtime verification.

The sibling client's bird logo SVG is a different identity and has not been adopted. The current website PNG is the reference for preparing the animation's vector symbol.

## Hyperframes

- User-requested command completed: `npx skills add heygen-com/hyperframes --full-depth`.
- `npx hyperframes skills update product-launch-video` reported the installed skills already up to date.
- Project scaffold: Hyperframes `0.8.78`; script version pinned in `package.json`.
- `carousel-circle-4`: installed from the official registry. Twelve image slots, circular depth scaling, fast-slow-fast motion, 6-second stock duration. Catalog evidence: `reference/catalog-carousel.json`.
- `halftone-field`: installed from the official registry. Experimental WebGL noise-driven dot field, 10-second stock duration. Catalog evidence: `reference/catalog-halftone.json`.
- Both blocks remain stock and have not been rendered in the proposed Xem edit. Their installed source and lock records are retained.

## Validation completed

Scene ranges, keyframe ranges, poster ranges, every extracted frame path, all review-page local links, unique HTML IDs, and viewer JavaScript syntax were checked. The 28 intervals sum to exactly 901 frames. No final film checks, animation renders, or visual comparisons have been run.

The browser connector was unavailable, and native browser actions were rejected with a “user changed” state message. The review page's live browser behavior has therefore not been verified; it opens from local files and its dependencies have been checked.
