---
workflow: product-launch-video
flow: automation
storyboard: yes
message: "Every email, a little more human."
aspect: 1920x1080
language: en
length: 30.033333s
angle: reference-matched brand film
phase: delivered
plan_approved: true
---

# Xem brand film — brief

Recreate the supplied Cyera brand video for Xem, matching its scene order, cut frames, composition, camera movement, typographic choreography, illustrated depth, and closing rhythm. The scene plan was approved by the user. The final video render is authorized.

## Explicit user direction

- “exact frame by frame for xem”
- Read the Xem website to understand the product.
- Install `npx skills add heygen-com/hyperframes --full-depth`.
- Use Circle Carousel 4 and Halftone Field.
- “FIRST PLAN OUT EVEYTHING EVERY SCENE”

## Source material

- Reference: `/Users/theboringhumane/Downloads/Motionimo_-_Cyera_Brand_Video_by_Ido_Sapir_amp_Team_6khr40.mp4`.
- Reference video: 1280×720, nominal 30 fps, 901 decoded frames, video duration 30.033346 s. Container/audio duration: 30.058667 s.
- Website: https://xem.email, read live on 2026-09-26, and the current website checkout.
- Brand: `public/brand/xem-mark.png`, DM Sans, EB Garamond, and the colors in `tailwind.config.ts`.
- Existing product and template imagery: `public/images/product/` and `public/images/templates/`.

## Working decisions for review

- 1920×1080 master, same 16:9 composition as the reference. Reference measurements multiply by 1.5. Retain 901 frames at 30 fps; do not silently shorten to 900.
- Use the source edit as the timing authority. The master therefore lasts 30.033333 s, with the source's approximately 25 ms audio tail trimmed at final mux.
- Preserve the 28 mapped scene beats, including transitions lasting only a few frames.
- Replace the Cyera identity and security claims with Xem's actual identity and email-marketing story.
- Preserve the film's strong light/dark alternation, with Xem cream, iris, forest, lavender, and lemon mapped to the corresponding roles.
- Retain the illustrated traveling-object story. Adapt the orb into a mail capsule with Xem's connected-shape geometry; this is a visual metaphor, not a product assistant or autonomous sending claim.
- Keep the source audio as a local timing reference. Plan a music/SFX-led version without added narration; soundtrack audition and final mix are still to do.
- Use Circle Carousel 4 inside the montage's template panel, and Halftone Field as a masked light texture. These are deliberate additions, so they must not replace the source's overall layouts or lengthen the film.
- Public destination is unspecified; the supplied landscape reference determines the master format. No extra social crops are included in this first plan.

## Scope and evidence

The sequence has been inspected through contact sheets covering all 901 frames, plus larger overview frames. Cut indices are observed from decoded frames. Proposed copy, Xem-specific geometry, asset treatments, easing, and pixel coordinates are design decisions until compared against a built result.

Live website content supports campaigns, newsletters, editable templates, visual automations, audience/CRM, forms, analytics, SMTP/IMAP, and open source. The live site labels Ask Xem “In development” and its demonstration “Scripted example.” Managed sending is described as available subject to domain verification and workspace approval. This film does not need a managed-delivery or assistant-availability claim.

Hyperframes 0.8.78 is the pinned renderer. Production uses 28 timed canvas scenes, source-derived registry effect adapters, local Xem fonts/assets, and the supplied reference soundtrack. No production website code is changed.
