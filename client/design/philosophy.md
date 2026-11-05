# Xem product design philosophy

This is the bar for every signed-in screen. **The gold standard is the Acme
"Development Pipeline" shell**; the Claude-powered inbox and Oripio "Product Launch
2026" are supporting references. Where they disagree, Acme wins. It also records
what is wrong with Xem today. When a screen and this document disagree,
the screen is wrong.

## What we are building toward

**Calm, precise, quietly confident.** The interface is a well-made tool, not a
marketing page. Neutral surfaces carry the structure; colour appears only where it
tells you something (state, identity, AI, the one action that matters). Every
pixel of chrome earns its place, so the user's work — their campaigns, their
audience, their mail — is the loudest thing on the screen.

Three tests for any screen:

1. **Squint test.** Blur your eyes. You should see one white work surface, one
   clear heading and one obvious next action. Nothing else should pop.
2. **First-minute test.** A new user with an empty workspace knows what this page
   is for and what to do next, without reading docs.
3. **Slop test.** No gradients behind text, no glowing borders, no emoji icons, no
   random accent colours, no cards inside cards inside cards, no copy that says
   "Supercharge". If it looks generated, cut it.

## What is wrong today

- The sidebar and the content share the same flat warm grey, so there is no
  "work surface". Everything reads as one undifferentiated page.
- The active nav item is barely distinguishable; section labels, items and icons
  all have similar weight.
- Type is set large and loose (the sidebar reads like 16px body text), which makes
  the app feel like a website instead of a tool.
- Lime is used as the primary button colour. On white it has poor contrast and
  makes every primary button shout.
- Empty pages are a blank white void (see the "Workspace / Workspace" screen).
- Breadcrumbs repeat themselves and say nothing ("Workspace / Workspace").

## Principles

### 1. Two layers: canvas and sheet
The app window is a **canvas** (cool light grey). The sidebar sits directly on the
canvas with no border of its own. The page lives on a **sheet**: a white panel with
a hairline border and a large radius, inset from the window edge. This single move
is what makes the references feel finished.

### 2. Hairlines over shadows
Structure comes from 1px borders in a very light neutral. Shadows exist only on
things that float (menus, popovers, sheets, toasts) and are soft and wide, never
dark. No shadow on cards that sit on the sheet.

### 3. Colour is information
Grey does the layout. Colour is reserved for:
- **State**: soft tinted pills (draft, scheduled, sending, sent, failed).
- **Identity**: small coloured tiles or avatars for lists, newsletters, senders.
- **AI**: the sparkle mark for anything Xem's assistant does.
- **Brand**: Xem lime appears in the logo and small highlights, not on buttons.
If removing a colour loses no meaning, remove it.

### 4. One primary action per view
The primary button is near-black with white text (the "Compose" button in the
inbox reference). Secondary actions are soft grey filled buttons. Everything else
is ghost. A page never has two black buttons.

### 5. Density with air
Rows and nav items are compact (32–36px), but groups breathe: generous gaps
between sections, consistent padding inside surfaces. Tight inside, loose between.

### 6. Show the shape of the data
Prefer structure that mirrors the object: key/value blocks on a muted inset
(label with icon on the left, value on the right), segmented progress, counts next
to tab labels, IDs in mono. Avoid big decorative stat cards full of zeros.

### 7. Guide, then get out of the way
Empty states teach the page in one sentence and offer one action. Upgrade prompts
are native, beautiful and dismissible (the "Upgrade with AI" card in the inbox
reference sits inside the list, it is not a modal). Help appears where the
question arises, not in a tour.

## Foundations

### Colour tokens (light)
| Token | Value | Use |
|---|---|---|
| `--canvas` | `#f4f4f5` | Window background behind sidebar and sheet |
| `--background` (sheet) | `#ffffff` | Main work surface |
| `--surface-muted` | `#f7f7f8` | Inset blocks, input fills, secondary buttons |
| `--border` | `#ececee` | Hairlines everywhere |
| `--border-strong` | `#e1e1e4` | Inputs, active nav pill |
| `--foreground` | `#18181b` | Headings, values, primary text |
| `--text-secondary` | `#52525b` | Nav items, body copy |
| `--muted-foreground` | `#8a8a93` | Labels, meta, placeholders, icons at rest |
| `--primary` | `#18181b` | Primary buttons |
| `--primary-foreground` | `#ffffff` | Text on primary |
| `--brand` | `#d4ed55` | Logo and small brand highlights only |

Status tints (background / text): neutral `#f4f4f5 / #52525b`, blue
`#eaf2ff / #1d5fd6`, green `#e9f8ef / #1f7a45`, amber `#fff4e5 / #a35a00`,
peach `#ffefe6 / #b5471b`, red `#feecec / #c2302b`. Same hue family in dark mode
at roughly 15% opacity over the surface.

Dark mode keeps the same structure: canvas `#111113`, sheet `#18181b`, hairlines
`#26262a`, primary becomes white with black text.

### Typography
- Family: Geist Sans (already loaded); Geist Mono for IDs, tokens and code.
- Base size **13px**, line height 1.5. Nav items 13px. Meta and labels 12px;
  key/value rows 12px.
- Page title **17px / 600**, letter spacing −0.01em, with a 13px muted one-line
  description. Card title 15px / 500. Big stat numbers 20px / 500 tabular.
  Large display headings are not used on work screens.
- One exception: the Ask Xem home heading uses EB Garamond (`--font-editorial`,
  the website's editorial serif) at 400 weight, 34–46px, −0.03em, with one
  italic word. It is the only place the app greets you rather than shows data.
- Weights: 400, 500, 600 only. No bold 700, no all-caps labels except tiny
  status badges.
- Numbers are tabular (`font-variant-numeric: tabular-nums`).
- Sentence case everywhere ("Contact lists", not "Contact Lists").

### Spacing and size (4px grid)
- Window inset around the sheet: **8px**.
- Sidebar width **232px**, horizontal padding 12px, nav item height **28px**,
  item padding 4px 8px, icon–label gap 8px, gap between groups 20px.
- Sheet radius **20px**; card radius **12px**; inset block and input radius
  **10px**; buttons and pills radius **8px** (pills fully rounded).
- Page content max width 1200px, horizontal padding 32px (16px on mobile),
  top padding 24px.
- Top bar height **52px**, hairline below.
- Card padding 16px; gap between cards 16px.

### Icons
Lucide, outline, **16px** in nav and buttons, **14px** in meta rows, stroke
**1.5**. Icons at rest use `--muted-foreground`; they take the text colour when the
row is active or hovered. No filled icons except status dots and the AI sparkle.

## Components

### Sidebar
- Sits on the canvas, no border, no background of its own.
- Workspace switcher at top: 28px logo tile (dark rounded square with the Xem
  mark), name 14px / 500, plan 12px muted, chevron-up-down. Collapse control to the
  right.
- Search: 32px tall, `--surface-muted` fill, hairline, ⌘K hint at the right.
- Group labels: 13px / 500, `--foreground`, sentence case, with a small
  chevron to collapse the group. 6px above items.
- Items are full-width rows, 30px tall, padding 5px 8px, radius 8px, 13px
  `--foreground` text, 16px muted icon, 2px between items. No fill at rest
  (content-width filled pills looked ragged with Xem's short labels); hover adds
  a faint fill (`rgb(0 0 0 / 0.07)`).
  **Active: white pill with a very soft shadow
  (`0 1px 2px rgb(0 0 0 / 0.06)`).** Entity items (lists, newsletters) can use
  a small coloured tile or status icon instead of a line icon.
- Footer: Upgrade card (for Free workspaces) then the user row: 28px avatar,
  name 13px / 500, email 12px muted. Settings live in the user menu, not as a
  floating avatar over a nav item.

### Top bar (inside the sheet)
Breadcrumb with a 16px section icon, parent in muted text, `/` separator, current
page in foreground. Never repeat the same word twice. Optional search or page
actions on the right. 52px tall, hairline bottom.

### Page header
Title 17px / 600, optional 24px identity tile to its left, one-line description
13px muted beneath. Primary action on the right. Tabs (if any) sit below with icon
plus label; active tab is a white bordered pill or a 2px underline in foreground,
counts shown as small muted numbers next to the label.

### Buttons
- Primary: `--primary` fill, white text, 32px tall (36px for hero actions),
  radius 8px, 500 weight, optional leading icon.
- Secondary: `--surface-muted` fill, hairline, foreground text.
- Ghost: no fill, `--text-secondary`, hover `--surface-muted`.
- Destructive: red text ghost; filled red only inside a confirm sheet.

### Cards and lists
White, hairline border, radius 12px, padding 16px, no shadow. Header row: 32px
identity tile or avatar, title 15px / 500, ID or subtitle in mono 12px muted.
Details go in an inset block (`--surface-muted`, radius 10px, padding 12px) as
key/value rows: icon + label muted on the left, value foreground on the right.
Footer meta (counts, time ago) in 12px muted with 14px icons. Lists use 56–72px
rows separated by hairlines, avatar left, title + snippet, time right-aligned.

### Status pills and progress
Pills: fully rounded, 22px tall, 12px / 500, tinted background, optional 8px
dot or 12px icon. Progress: segmented bars (thin vertical ticks) or a 6px rounded
bar, tinted by state, with "12/32" in tabular numerals and a small trend delta.

### Menus and popovers
White, hairline, radius 12px, shadow `0 12px 32px rgb(0 0 0 / 0.08)`, 4px inner
padding, items 32px with 16px icons, dividers as hairlines with 4px vertical
margin, optional search input at the top.

### Empty states
Centred in the content area, max width 420px: 40px muted icon tile, 15px / 600
title that names the outcome ("Send your first campaign"), one 14px muted sentence
on what the page does, one primary action, optional ghost link to docs. No
illustrations of people, no confetti.

### Illustrations
Soft, painterly scene illustrations (sky, architecture, landscape) are allowed in
exactly two roles: as the background of a promotional or empty-state card that
fades into white at the bottom (the "Upgrade with AI" card in the inbox
reference), and as the canvas behind the whole app on marketing screenshots.
Text sits on the calm part of the image in foreground colour, never on busy
detail. Use only images we own or have licensed; no images lifted from Pinterest.
Store them in `public/illustrations/`, ship as WebP under 150 KB, and always pair
them with a gradient fade so the card still reads in dark mode.

### Getting started and upgrade
- Getting started is a card at the top of Overview: title, "2 of 5 done",
  segmented progress, then steps as rows (number or check, title, one line,
  action on the next unfinished step only). Dismissible.
- Upgrade lives in two places only: a quiet card at the bottom of the sidebar for
  Free workspaces, and inline "comes with a paid plan" notes where a paid feature
  is reached. The upgrade sheet lists concrete benefits, then one primary
  "Start free trial" action.

## Copy
Short, specific, human. Say what happens ("Send to 1,240 contacts"), not what the
feature is ("Campaign dispatch"). No exclamation marks, no "Oops", no
"Supercharge", no "Unlock". Errors say what went wrong and what to do next.

## Rollout order
1. Tokens and the canvas/sheet shell (affects every page at once).
2. Sidebar, top bar and page header.
3. Buttons, inputs, pills, menus.
4. Cards, lists and tables page by page, starting with Overview, Campaigns, Inbox.
5. Empty states, getting started and upgrade surfaces.
