# Dashboard redesign validation

Reviewed locally on September 22, 2026. See [design-system.md](./design-system.md) for the reference extraction and implementation rules.

## Automated checks

- The configured Turbopack production build and TypeScript validation passed, including in the isolated checkout prepared for pushing.
- The initial shared-workspace run passed all 67 tests across 15 suites, including 9 sheet interaction tests.
- Before pushing, the redesign was separated from the unfinished form features and exported into an isolated checkout. All 45 tests across its 12 suites passed, including the same 9 sheet interaction tests.
- After completing the form features, the combined frontend passed all 85 tests across 18 suites and the configured production build, including TypeScript validation.
- After integrating the latest `sudo` DNS-onboarding changes, a clean checkout passed `bun install --frozen-lockfile`, all 90 tests across 19 suites, and `bun run build` with TypeScript validation. The committed Bun lockfile includes the Coss and sheet-test dependencies used by the release build.
- Sheet coverage includes cancellation, focus restoration, dropdown handoff, nested selects, and busy states.
- The existing ESLint configuration fails while loading; lint could not complete. Build, type checking, and tests passed independently.
- An additional build using the unconfigured Webpack alternative failed to resolve the existing editor dependency `y-protocols/awareness`. The project's configured Turbopack build passed; Webpack compatibility is not claimed.

## Browser review

The development-only [preview](http://127.0.0.1:3100/preview) uses visibly labeled synthetic data and disables mutations. Desktop, 390px, and 320px layouts were reviewed. The final preview had no browser console errors.

- Reviewed the dashboard, shared navigation, newsletters, form editor, CRM, inbox, outbox, automation editor, assistant, and sending surfaces.
- Checked empty, error, and retry states, newsletter editing and template selection, and form editing.
- Verified inset sheets with 12px desktop and 8px mobile viewport margins, scrolling content, and confirmation focus initially on Cancel.
- Inbox, Outbox, and the automation editor fill the content area beneath the application header without outer padding, borders, or rounded containers.
- Browser measurements confirmed identical bounds for the Outbox content area and mail workspace, with zero outer padding; the automation editor also matched its content area's bounds and reached the bottom of the workspace.
- Mail folders, message lists, and readers retain independent scrolling. Automation side panels scroll independently of the canvas.
- The completed form editor opens as an inset sheet; selecting the Demo request starter updates its interactive two-step preview. The browser console remained free of errors. Anonymous requests to the public embed script return JavaScript successfully; workspace form routes remain authenticated.

## Saved visual evidence

- [Dashboard at 1440 × 1000](./screenshots/dashboard-desktop.png)
- [Automation editor at 1280 × 720](./screenshots/automation-desktop.png)

## Scope and limitations

The first redesign commit separated the form features from the visual changes. The follow-up forms release completes and includes those frontend and backend features, with additional coverage for resume/version recovery, immutable retries, consent-aware native HTML sharing, public asset access, and inset embeds. The form editor and multi-step starter were reviewed again in the browser.

This validates the local interface and automated behavior. Authenticated production workflows, real sending, and deployment were not verified. Backend integration tests use local storage and an injected queue; the backend forms documentation records the rollout order and data semantics.
