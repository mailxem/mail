# Xem onboarding design decisions

Reviewed 12 September 2026 while implementing managed sending.

Public references inspected:

- [Resend domain setup](https://resend.com/docs/add-a-domain): exact copyable DNS values, provider-specific help, realistic propagation states, and separation between sending and receiving configuration.
- [Linear Start Guide](https://linear.app/docs/start-guide): an outcome-focused path from a new workspace to useful work, optional exploration, and different guidance for administrators and members.

These were public documentation flows, not a review of their private signed-in onboarding screens. The Xem layout is original and follows the product's existing iris, rounded cards, quiet borders, and workspace shell.

## Applied patterns

1. A short promise: “A little setup. A lot of possibility.” The target is a real first email, not a tour of features.
2. Choose managed sending or an existing provider first; remove irrelevant steps from the checklist.
3. Persist the choice and dismissal in the workspace. Derive completion from actual domains, sender configuration, accepted messages, and campaign drafts; never mark a technical step complete just because its button was clicked.
4. Keep the checklist visible beside one focused panel. Allow users to inspect future steps and return later.
5. Explain DNS at the point of use. Provide copy buttons, precise record labels, a collapsed help section, a separate approval state, and reassurance that the existing inbox stays in place.
6. Celebrate real completion with warm copy and a next action, without fake confetti, countdowns, or forced invitations.
7. Support keyboard navigation, visible labels, readable errors, reduced-motion preferences, and narrow screens.
8. Surface a resume card on the dashboard and a persistent Getting started navigation entry.

## Verification plan

Check a fresh managed journey, a returning partially configured journey, a BYO journey, disabled managed sending, pending DNS/approval, completed setup, and mobile layout. Confirm that live credentials only appear once and never enter URL parameters or persistent browser storage.

## Implemented and checked

- Account registration and the new-user callback lead into `/onboarding`; the old `/auth/onboarding` address redirects there. Registration already creates the workspace, so this removes a duplicate team-creation form. Invited members go to their dashboard; sending setup is for workspace administrators.
- Fresh managed selection, adding a domain, waiting for DNS, testing a ready domain, completed setup, BYO selection, and disabled managed sending were exercised in the local preview.
- Desktop (1280 × 720) and mobile (390 × 844) layouts were inspected. On mobile the active panel precedes the full checklist; progress remains above it. Browser console inspection reported no errors during this pass. Live sign-in/DNS/SES transport is a separate deployment validation gate.
- All 27 frontend tests passed, including regression coverage that another domain's test or previous BYO mail cannot complete managed onboarding. TypeScript and the production build passed.
- Configuration and test-message progress refresh every 15 seconds while the page is active. Tests complete on provider acceptance, not merely queue submission; delivery confirmation remains visible separately in Sending.
