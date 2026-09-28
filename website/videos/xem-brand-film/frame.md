---
name: Xem reference-matched brand film
status: approved; implemented
source: website theme and supplied Cyera film
canvas: "#ffffef"
ink: "#22251f"
accent: "#5b3cc4"
highlight: "#edf09b"
lavender: "#e7d8fa"
forest: "#164b3f"
sans: DM Sans
editorial: EB Garamond
---

# Visual direction

The supplied film determines composition and motion. Xem's website determines identity, colors, typography, claims, and product images. This is a custom reference direction, not a generic preset.

## Color roles

| Reference role | Xem role | Source |
| --- | --- | --- |
| Warm white construction boards | Cream `#ffffef` | `tailwind.config.ts` |
| Near-black stage | Ink `#22251f` | `tailwind.config.ts` |
| Saturated purple surfaces | Iris `#5b3cc4` | `tailwind.config.ts`, `lib/brand.ts` |
| Pale purple type and light | Lavender `#e7d8fa` | `tailwind.config.ts` |
| Lime controls and final mark | Lemon `#edf09b` | `tailwind.config.ts` |
| Green illustrated landscape | Forest `#164b3f` | `tailwind.config.ts` |

Preserve the reference's strong tonal separation and scene-to-scene color changes. Gradients interpolate these colors; sample their rasterized result for banding. The current colors are a Xem adaptation, not a claim of a pixel-identical palette match to Cyera.

## Type

- DM Sans is the geometric display/wordmark role. Use the site's wordmark casing, `Xem`, with optical tracking rather than nonuniform scaling.
- EB Garamond is the editorial contrast role. It replaces the reference's serif during the typographic switch and newsletter covers.
- Fonts come from the website's `app/layout.tsx`. Freeze local font files before rendering; local font files are staged in assets/fonts.
- Scale glyph groups to match their source layout footprint; different wording requires optical fitting. Hero type will likely be 120–180 px on the 1920-wide master; edge-cropped type can be larger. These are starting ranges, not measured final values.
- Tiny construction labels are secondary design texture. The main marketing claim never depends on reading them.

## Composition and motion

Retain source object placement, aspect ratios, horizon lines, entry order, held intervals, and cut frame numbers. Use a 1920×1080 master and a 1280×720 comparison export. Frame `n` samples time `n / 30`.

Primary continuing objects are the logo module, rounded viewport, newsletter cover, mail capsule, perspective rays, and descriptor anchor. Continuity must be shared numerically between adjacent scenes during the animation pass. Do not estimate a fresh pose in each scene.

Use flat/vector construction graphics and layered 2.5D or 3D for the illustrated worlds. The capsule uses one mesh in wireframe and shaded modes. Derive the background motion from the source; scene 28 is deliberately static.

The source is the authority where generic skill defaults differ: do not insert a pain-point intro, stretch transitions to 1.5 seconds, reserve an unused subtitle band, apply a universal crossfade, or add constant ambient movement. No narration or subtitle strip is planned.

## Requested blocks

- `carousel-circle-4`: installed 6-second, 1920×1080 block with twelve images, speed, and background controls. Use within the template panel in scene 13. Replace all bundled music-artist images with staged Xem template previews. Keep its circular depth/size progression and fast-slow-fast swing. The host scene stays 48 frames. Select and retime a useful movement interval instead of playing all six seconds.
- `halftone-field`: installed experimental 10-second WebGL block. Use masked light in scenes 11, 15, and 26. Starting trial: frequency 2, speed 1, cell size 12–18 px, gamma 10–12, Xem palette stops, ink background, time sampled at the output frame. Tune on moving comparison frames; these values are not validated settings.
- Both are planned additions. Keep an effect-off comparison so they can be judged against the reference. They must not obscure the source's silhouette, type, or rhythm.

## Asset fidelity

The authoritative current website mark is the two-circle/two-square symbol in `assets/brand/xem-mark.png`. The sibling client's `public/assets/logo.svg` is an older bird mark and is not suitable. A faithful transparent/vector preparation of the current mark is still needed; do not substitute another mark or generate a new logo.

Use actual Xem product captures and template previews. Their sample data is not evidence of customer usage. The laptop shot requires a fresh real website capture. The cap/poster/sweatshirt are proposed brand applications, not a merchandise launch.
