# Xem — Make email more human

The delivered film is `renders/xem-brand-film-1080p.mp4`: 1920×1080, 30 fps,
901 frames, 30.033333 seconds, stereo soundtrack.

This is the approved 28-scene Xem adaptation of the supplied Cyera brand film.
It retains the mapped cut frames and sequence. Xem identity, copy, palette,
product screens, and recreated illustrations replace the reference branding.
The rebuilt illustration and object geometry are not pixel-identical copies of
the original studio's production assets.

## Edit and render

- `assets/film.js`: editable scene drawings, text, camera/object poses and effects.
- `scene-plan.json`: approved zero-based scene ranges, copy, cues and source notes.
- `index.html`: 28 labelled clips grouped by chapter in the Studio timeline.
- `assets/vendor/carousel-rig.js`: original Circle Carousel 4 path/pose mathematics,
  adapted to canvas and filled with twelve Xem template previews in scene 13.
- `assets/vendor/halftone-field.js`: original Halftone Field WebGL shader, with
  Xem palette and a deterministic frame driver in the illustrated scenes.
- `compositions/`: unchanged installed registry originals for reference.

```sh
python3 scripts/build-index.py
npm run check
npx --yes hyperframes@0.8.78 preview --background
npm run render -- --quality delivery --fps 30 --workers 2 --output renders/xem-brand-film-1080p.mp4
```

Rendering uses local fonts and assets. Internet is needed only if the pinned CLI
has not yet been installed. Scene graphics are code-editable canvas layers;
individual words are not separate draggable Studio text elements.

## Media provenance

The live Xem website was captured on 2026-09-26. The laptop shows that captured
site. Product and template images come from the current website repository.
The official current Xem mark was prepared from its PNG, not the obsolete bird.

The soundtrack comes from the user-supplied reference, trimmed to the picture
duration. The cap and sweatshirt photographic plates also come from the supplied
reference; their print regions are replaced with Xem artwork. The poster,
landscapes, capsule, icons, typography, laptop frame and other graphic scenes
are recreated drawings. Apparel is a conceptual brand application.

`SOURCES.md`, `BRIEF.md`, `STORYBOARD.md`, and `frame.md` retain the research and
approved creative plan. `DELIVERY.md` records the final render and validation.
Source-video frames and review captures are local research assets, separate from
the source ZIP needed to render the film.

## Website hero

The parent website embeds this film through `components/hero-film.tsx`. Optimized
copies live in `public/videos/`: a 2.35 MB 720p film with sound, a 489 KB silent
832px preview, a WebP poster from 26.6 seconds, and English music captions.
The standalone 1080p master and its editable sources remain here.

From the website repository root, run `node scripts/prepare-hero-film.mjs` with
FFmpeg installed to regenerate the web copies after rendering a new master.
