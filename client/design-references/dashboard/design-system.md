# Xem workspace design system

Reference: the user-supplied Ace Studio dashboard screenshot, reviewed September 22, 2026. Component source: [Coss UI](https://coss.com/ui/docs/get-started), [Sheet](https://coss.com/ui/docs/components/sheet.md), [registry source](https://coss.com/ui/r/sheet.json).

The screenshot provides visual evidence, not exact source CSS. The measurements below are implementation targets inferred from its proportions; the precise original font cannot be identified from a bitmap. The screenshot's Slack copy, workspace name, and numbers are reference content, not Xem product requirements.

## Philosophy

The workspace should feel calm, precise, and continuous. Work sits on one charcoal canvas. Navigation is visible but quiet. Thin rules and spacing establish groups; shadows and bright colors are reserved for a reason. The active page is identified by a soft neutral fill, not a colorful block. Text supplies the hierarchy before decoration does.

Keep task context while editing. A create, edit, detail, search, or confirmation action opens an inset side sheet with a visible margin from the viewport. Preserve the current page behind it. Keep the existing information architecture and real data semantics: a sent email is SMTP acceptance, not proof of inbox delivery.

## Reference extraction and Xem translation

| Element | Reference observation | Xem implementation |
| --- | --- | --- |
| Canvas | Almost uniform charcoal, subtly rounded outer frame | #191919 workspace; full viewport application shell |
| Sidebar | Roughly 18% of the displayed app width; narrow text and icons | 236px desktop, compact collapsed rail, mobile navigation sheet |
| Top context | Small breadcrumb, generous content lead-in | 56px header with breadcrumb, compact search, account menu |
| Content | Centered, bounded reading column for integration settings | Fluid operational pages; bounded form and sheet widths |
| Hierarchy | Small neutral headings and muted secondary text | 26px page title; 16px section title; 14px body; 12px labels |
| Panels | Barely raised surface, thin low-contrast borders | #1c1c1c panels, #2d2d2a borders, 12px radius |
| Accent | Lime mark; functional green/red states | Lime primary action/brand; semantic status colors |
| Grouping | Details table and stacked permission rows | Divided metrics, grouped tables, compact settings tabs |
| Interactions | Restrained buttons and selectors | Coss-derived controls; all blocking overlays use inset sheets |

## Typography

Use the locally bundled Geist Sans family for body, navigation, and headings, with Geist Mono for identifiers and shortcuts. This visually matches the reference's restrained grotesk typography without claiming to reproduce an unidentified proprietary font.

| Role | Size / line height | Weight | Tracking |
| --- | --- | --- | --- |
| Page title | 26px / 1.2; 24px on mobile | 500 | -0.8px |
| Sheet title | 20px / 1.4 | 500 | -0.4px |
| Section title | 16px / 1.5 | 500 | -0.2px |
| Body and forms | 14px / 1.5 | 400 | normal |
| Navigation | 13px / 1.4 | 400–500 | normal |
| Labels and metadata | 12px / 1.5 | 400–500 | normal |
| Metrics | 26–32px / 1.1 | 500 | -1.2px; tabular figures |

Avoid all-caps labels except short preview/category markers. Long descriptions wrap naturally; counts and identifiers use tabular or monospace treatment where helpful.

## Color tokens

Tokens live in `app/globals.css`. Components consume semantic names rather than embedding pastel palettes. Dark is the new default; the account menu supports Dark, Light, and System. Hosted form artwork and email previews preserve their own content theme.

| Token | Dark | Light | Use |
| --- | --- | --- | --- |
| background | #191919 | #fafaf8 | Canvas |
| foreground | #ededeb | #242520 | Primary text |
| card | #1c1c1c | #ffffff | Grouped content |
| popover | #222222 | #ffffff | Sheets and menus |
| muted | #242424 | #f1f1ed | Soft surfaces |
| muted-foreground | #a3a39e | #71736a | Supporting text |
| border | #2d2d2a | #e3e4de | Dividers |
| input | #3a3b35 | #d6d8cf | Control borders |
| primary | #d4ed55 | #d4ed55 | Primary action background |
| primary-foreground | #202610 | #202610 | Text on lime |
| primary-text | #d4ed55 | #586d1e | Accessible accent text |
| success / foreground | #183626 / #83c89b | #e4f1e8 / #2e744a | Positive state |
| warning / foreground | #3b311b / #dfbf79 | #faf0d8 / #906713 | Needs attention |
| info / foreground | #23303e / #a2bedf | #e7edf5 / #496d98 | Scheduled/informational |
| destructive | #d43b40 | #d43b40 | Destructive action |

Charts use related lime, sage, cyan, amber, and rose hues with labels and table alternatives. Do not invent trends or replace unavailable data with zeros.

## Spacing, margins, and padding

Use a 4px base rhythm: 4, 8, 12, 16, 20, 24, 28, and 32px. Small optical adjustments are acceptable for borders and icon alignment.

- Desktop content gutter: 32px; narrower desktop: 24px; mobile: 16px.
- Sidebar: 12px horizontal inset, 32px navigation rows, 24px between groups.
- Page heading to content: 28px; section spacing: 24–32px.
- Panels: 20px padding, 16px on mobile, 12px outer corner radius.
- Tables: 12px header padding and 14–16px cell padding; horizontal scrolling stays inside the table.
- Form groups: 16–20px between fields; 8px between label and input; 36–40px control height.
- Inset sheets: 12px viewport margin on desktop, 8px on mobile, 16px radius; 24px content padding.
- Icon-only desktop controls: 32–36px, expanded to 40px on small screens where space allows.

## Product and UX layout

**Workspace:** Overview, Ask Xem, Inbox, Outbox. These are the daily entry points.

**Create:** Campaigns, Newsletters, Automations, Templates, Forms. Keep production tasks grouped together.

**Audience:** Contact lists, CRM, Analytics. Group contact organization with measurement.

**Persistent utilities:** Getting started and Settings at the sidebar foot. Settings includes sending, SMTP, IMAP, keys, webhooks, tags, team, billing, and account. Search is available from the header and Command/Ctrl K. Appearance belongs in the account menu.

Inbox, Outbox, and the automation editor fill the area beneath the application header edge to edge, with no outer padding, border, or corner radius. Folder navigation, message list, and reader scroll independently. The automation toolbar sits above a flexible canvas with independently scrolling palette and inspector.

Operational dashboards use a compact filter rail followed by a divided metric strip and related charts/details. Settings use compact section tabs followed by grouped field rows. Resource collections share quiet titles, neutral icons, status badges, and a divided action row.

## Inset sheets and interaction rules

The sheet implementation manually adapts the official Coss Base UI composition (Portal → Backdrop → Viewport → Popup) to the project's existing Tailwind 3 build. Coss's current registry uses Tailwind 4; importing its classes unchanged would silently omit styling. The component uses `@base-ui/react`, with source attribution, rather than a visual-only imitation of a centered Radix dialog.

- `SheetPopup` is the modern composition; `SheetContent` preserves existing consumers.
- `SheetHeader`, `SheetPanel`, and `SheetFooter` provide title, scrollable body, and actions.
- Default side is right; navigation enters from the left. Inset is the default variant.
- Existing `Dialog` imports are compatibility aliases to sheets. AlertDialog imports retain their accessible confirmation semantics but render as sheets. The ARIA dialog role remains appropriate for a modal sheet; the visual interaction is not a centered dialog.
- Native `confirm()` calls become awaited `useConfirmSheet()` decisions. Cancellation performs no mutation. Critical confirmations do not dismiss on backdrop click and initially focus Cancel.
- Escape, focus trapping, focus restoration, nested selects, and dropdown-to-sheet transitions must work.
- Long content scrolls within the sheet. Forms with previews use a wide sheet and collapse to one column at narrow sizes.
- Respect reduced motion, readable error states, empty states, and keyboard interaction. Never fake success after an unsuccessful request.

## Implementation map

- `app/globals.css`: shared light/dark and semantic tokens.
- `lib/workspace-styles.ts`: shared form, table, inbox, automation, and resource recipes.
- `components/workspace-shell.module.css`: shell, navigation, responsive behavior.
- `components/analytics/analytics-surface.module.css`: analytics information hierarchy.
- `components/ui/sheet.tsx`: Coss inset sheets.
- `components/ui/confirm-sheet.tsx`: awaited confirmation queue.
- `components/ui/collection-card.module.css`: shared resource presentation.
- `app/preview`: development-only, synthetic visual review; mutations disabled.

See `validation.md` for the verified review scope and limitations.
