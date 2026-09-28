# Delivered Xem brand film

- Master: `renders/xem-brand-film-1080p.mp4`
- Format: H.264 High, Rec.709, 1920×1080, 30 fps, 901 frames.
- Picture duration: 30.033333 seconds. Audio: AAC stereo, 48 kHz.
- Size: 14.50 MB.
- SHA-256: `65dbaf5f4c211ce0f538fb1ced8bb05c39af0fc94b2e293c2a243c5365a039ad`
- Renderer: Hyperframes 0.8.78, delivery quality, 2 workers, hardware GPU
  screenshot capture. Final render completed in 32.5 seconds.

## Validation

`npm run check` passes: no lint errors, no runtime errors or warnings, no layout
issues across nine samples, no motion errors or warnings. Three architectural
warnings recommend smaller sub-compositions for the chapter tracks. The film
intentionally keeps all source frame decisions in one shared renderer so objects
stay continuous across scene boundaries. Canvas-drawn text has no DOM contrast
checks; typography was reviewed in rendered images.

All 901 frames decode successfully, with no uniform blank frame. The last 47
frames form a visually static logo hold (mean temporal pixel standard deviation
0.00076/255 after lossy encoding). Every cut boundary and
scene midpoint was inspected. Full frame sheets are in `snapshots/final-audit`.
The soundtrack was checked for speech with local Whisper Tiny plus VAD; it
returned no speech segments. This is automated evidence, not a human audition.

## Implementation and fidelity

The 28 approved scene ranges and 901-frame duration are preserved. Xem's mark,
DM Sans / EB Garamond, product screenshots, template previews and truthful
marketing language are used. Circle Carousel 4's installed path rig drives the
12-template montage. Halftone Field's installed WebGL shader supplies the purple
light textures. All motion is deterministic from the output frame index.

The recreated terrain, capsule, icons and camera poses follow the approved
reference composition, but are adaptations, not pixel-identical production
assets. The current Xem mark is prepared as a transparent raster rather than an
invented replacement SVG. The cap and sweatshirt use photographic plates from
the supplied reference with Xem print regions; the poster is redrawn. The music
is the supplied reference soundtrack. These choices are explicit in README.

The editable source package contains the HTML timeline, scene renderer, frozen
fonts/assets, installed effect originals, effect adapters, approved plan and
re-render instructions. The website hero now uses optimized copies of this
master, with a silent preview and an on-demand full player; see the README's
website section. No production site deployment was performed as part of this
delivery.
