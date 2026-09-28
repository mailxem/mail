> Delivery update: the user approved this plan. The video is rendered; see [DELIVERY.md](DELIVERY.md) for implementation details, verification, and known fidelity differences.

# Production plan

This is the implementation plan for the proposed storyboard. The finished video does not exist yet.

## Build order

1. **Review the story and copy.** Confirm the 28 scene beats in `STORYBOARD.md`, the Xem color mapping, and the restrained placement of the two requested blocks. The source order and frame count are already mapped.
2. **Layout pass.** Create `storyboard.html` with one static Xem keyframe per beat, using real fonts, colors, and proposed copy. Draw media placeholders only for assets not yet produced. Review the whole sheet before animation.
3. **Prepare reusable assets.** Extract a faithful vector/transparent mark from the supplied current symbol; freeze DM Sans and EB Garamond; capture the real website; build the capsule, layered terrain, laptop, icon family, and three brand mockups. Keep editable sources and a provenance inventory.
4. **Build animation by connected sequence.** Group scenes 01–06, 07–10, 11–13, 14–18, 19–20, 21–22, 23–25, and 26–28. Shared-object sequences use one master trajectory and shared asset geometry. If using per-scene workers, give adjacent scenes the same numerical boundary state and retain one owner for the capsule/camera rig.
5. **Wire the two installed blocks.** Customize the carousel's twelve image variables and map its useful movement into scene 13. Mask and recolor Halftone Field behind scenes 11, 15, and 26. Check deterministic seeks and GPU availability before committing to the shader treatment.
6. **Assemble the exact edit.** Every scene begins at `start / 30` and lasts `(end - start) / 30`. Root duration is `901 / 30`. No default transition may shift a cut or consume extra frames. Preserve scenes that visually continue across an administrative storyboard boundary.
7. **Sound pass.** Audition the source audio and mark its accents against the cut list. Use the source locally as a timing reference, then settle the final music/SFX choice. No voiceover or captions are planned. Align logo fills, type changes, scene cuts, beam expansion, and final mark resolve to the existing edit. Mix decisions and final audio choice remain open.
8. **Compare and refine.** Follow the acceptance checks below. Deliver a review preview, then the 1080p master, 720p comparison, and editable Hyperframes project.

## Asset inventory

| Asset | Status | Use |
| --- | --- | --- |
| Current Xem mark PNG | Staged from website | Identity reference; scenes 01–04, 11–13, 23–28 |
| Current-mark vector / transparent symbol | To prepare faithfully | Path animation and clean recoloring; scene assets named `assets/brand/xem-logo.svg` are planned |
| DM Sans + EB Garamond | Website families verified; local font files to stage | All display and editorial copy |
| Product WebP captures | Staged from `public/images/product/` | Montage and content reference; illustrative workspace data |
| Twelve template WebPs | Staged from `public/images/templates/` | Circle Carousel 4 |
| Fresh 1920-wide Xem homepage capture | To capture | Scene 21 laptop screen |
| Mail capsule | To model and rig | Scenes 14–18, 26–27 |
| Three layered landscape treatments | To author as editable layers | Scenes 11–18, 21, 23–26 |
| Thin-line icon family | To author from real feature concepts | Scenes 19–20 and sweatshirt |
| Laptop screen-plane model/mockup | To prepare | Scene 21 |
| Cap, poster, sweatshirt mockups | To produce | Scene 22; conceptual brand applications |
| Circle Carousel 4 | Installed, not customized or rendered | Scene 13 nested template tile |
| Halftone Field | Installed, not customized or rendered | Masked light layer in scenes 11, 15, 26 |
| Source audio | Embedded in local source MP4 | Timing reference; audition/mix to do |

The carousel image order is: Show Your Work; Reading's Good for You; August's Most Wanted; Welcome to MyGreenhouse; Unleashing the Fizz; Meet the Leaders; Connect Your Tools; Thanks for Digging In; Let's Raise the Standard; What's New in Softr; The Family Journal; Your Next Teammate. Template brand names are design content, not customer endorsements.

## Technical contract

- Framework: Hyperframes, pinned `0.8.78` in project scripts.
- Delivery master: 1920×1080, 30 fps, 901 frames, duration 30.033333 s; H.264 MP4 with AAC audio when the soundtrack is settled.
- Comparison: 1280×720 to align directly with the source.
- Seek time comes from the integer frame index. No wall-clock animation, unseeded randomness, infinite loops, or render-time asset fetching.
- Use local fonts/assets. One registered paused timeline per composition; registry key matches its composition ID.
- Keep full-bleed backgrounds on explicit layers where shader composition requires them. Prefix all IDs by scene. Let the framework own clip visibility.
- Scene-boundary poses must specify position, scale, rotation, opacity, velocity, camera, and relevant mask state. The plan supplies timing; exact numerical values are established from full-resolution reference measurements in the layout/animation pass.
- Mark hard cuts explicitly. A storyboard boundary that continues the same shot does not automatically imply a visual transition.
- Keep the generated blank `index.html` clearly separate from the eventual film. Do not present a successful check of that starter as validation of the Xem animation.

## Acceptance checks

| Check | Pass condition |
| --- | --- |
| Duration and coverage | All 901 frames accounted for once, with no gaps or overlapping scene intervals |
| Shot sequence | All 28 beats retained in reference order, including 11-, 14-, and 15-frame transitions |
| Cut timing | Every observed hard cut lands on the source frame index; shared shots continue smoothly |
| Geometry | Source and output side by side at 720p; horizon, viewport bounds, object path, and type footprint align to the agreed Xem adaptation |
| Continuity | Inspect cut −1, cut, and cut +1 for every boundary, and full-speed playback for all shared-object sequences |
| Motion | Inspect the listed keyframes plus slow-motion playback; no unwanted pops, missing layers, or independent drifting |
| Identity | Current Xem mark and casing throughout; no Cyera branding or legacy bird logo survives |
| Copy | Exact proposed lines; no unverified adoption, assistant availability, or deliverability claims |
| Product proof | Real local product imagery and real website capture, clearly understood as sample workspace content |
| Registry effects | Correct carousel variant, all twelve Xem images, no stock artist imagery; halftone deterministic and behind content |
| Render QA | Hyperframes lint/check pass on the assembled film, followed by inspected snapshot sheet and full-motion preview |
| Audio | Final soundtrack selected, no unintended silence/clipping, exact timing, and end matched to picture |
| Deliverables | MP4 master, comparison proof/contact sheet, editable HTML/assets, and final scene/copy manifest |

## Known limits of the current planning pass

The reference frames have been extracted and inspected, and the scene intervals validated arithmetically. No recreation has been animated or visually compared yet. Source easing, exact camera transforms, detailed mesh geometry, and soundtrack characteristics have not been reverse-engineered into production values. Stock registry blocks are installed but untested in this composition.

The browser connector was unavailable during research, so the website was read from its live HTML and local source. The current site mark and supplied video frames were visually inspected from local files. A fresh website screenshot remains a production task.

HeyGen is signed out. The CLI reports that its offline voice/music engines are missing dependencies. This does not block planning or reuse of supplied local media; generated audio has not been attempted.
