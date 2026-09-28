# Forms and follow-up journeys

Forms support up to 12 ordered pages and 50 questions, conditional pages/questions, editable branding, hosted and embedded rendering, explicit save/resume, campaign attribution, and follow-up automation drafts. A form submission is recorded independently from marketing subscription consent.

## Publishing and capture

Authenticated workspace operations use `/api/v1/marketing` and the existing `marketing:read` or `marketing:create` API-key grants. Creating a journey additionally requires `templates:create` and `automations:create`. Session requests retain the existing role/scope checks. Every form, audience, sender, and workflow lookup is scoped to its workspace.

| Method and route | Behavior |
| --- | --- |
| `GET /forms` | List workspace forms and normalized definitions. |
| `POST /forms`, `PUT /forms/:id` | Save form text, theme, audience, status, and a validated definition; append a revision. Bodies are limited to 256 KiB. |
| `GET /forms/:id/submissions` | Return the latest 200 submissions for this workspace. |
| `GET /forms/:id/analytics` | Return session counts, completion rate, saved partials, step counts, and sources. |
| `POST /form-draft` | Generate a proposed form and follow-up email copy; save, publish, and sending do not occur. |
| `POST /forms/:id/journey` | Save editable email templates and an inactive workflow together. |

Public endpoints are under `/public/forms/:slug`, require a published form for new activity, and have body/rate limits. `GET` returns the current public definition and version; `/manifest` provides its schema. `POST` accepts JSON or native HTML submissions. `/events` records views, starts, and steps. `/progress` explicitly saves partial answers; `/resume` exchanges a private token for the saved form state.

Definitions use `schemaVersion: 1`, stable page IDs and field keys, `pages`, `consent`, and `saveProgress`. Supported questions are `TEXT`, `EMAIL`, `PHONE`, `TEXTAREA`, `SELECT`, `RADIO`, `CHECKBOX`, `NUMBER`, `DATE`, and `HIDDEN`. Each definition includes an unconditional required `email` question. Conditions use `equals`, `not_equals`, `contains`, or `is_set` and reference earlier questions only. The server reevaluates visibility and discards answers for skipped questions; hidden values come from the configured defaults.

Submission requests contain string-valued `fields`, explicit boolean `consent`, the displayed `version`, a random UUID `requestId`, and a random UUID `sessionId`. Optional attribution contains bounded UTM values and referrer. Retries of an uncertain request must retain the same request ID and payload. Reusing an ID with different answers is rejected. A committed receipt can be acknowledged after a subsequent form edit or archive without creating another submission.

Rich forms require the current version. Legacy simple native HTML forms remain compatible. When the version changes before a submission commits, the client reloads the definition, clears old consent and resume state, and asks the visitor to review the current form.

## Consent and saved progress

Consent modes are required, optional, or none. A visitor who does not opt in can submit an optional/no-subscription form without becoming a marketing subscriber. Existing unsubscribe, bounce, complaint, and suppression states remain protected. Only eligible, explicitly consenting subscribers enter marketing follow-up workflows.

Saving answers requires an explicit visitor action. Resume links carry an opaque private token in the URL fragment, rather than answers in query parameters. The database stores only the token hash. Tokens are scoped to a form, session, revision, and seven-day lifetime; saved progress requires a signing secret of at least 32 bytes. A resumed save does not extend the original expiry. Completion removes saved copies only for that form/session. Expired partial answers are removed opportunistically and by the periodic worker in bounded batches.

## Follow-up processing

The submission transaction writes its receipt, response, eligible contact updates, analytics, and a completion outbox event together. The periodic `forms:completion-tick` task runs every 15 seconds, processes a bounded batch, and retries events if enqueueing fails. It uses a stable workflow execution ID so a queue retry does not create another execution.

Dispatch rechecks the form, workspace, active workflow, contact status, and suppression. Workflows activated after the response are not applied retroactively. Trusted event identifiers cannot be overwritten by answers. Questions are exposed to email/condition steps as `form_<fieldKey>` variables while retaining the trusted form and submission identifiers.

Creating a journey requires a workspace sender, postal address, one to four reviewed emails, and a stable request UUID. It creates editable templates and a draft workflow atomically; it never publishes the form or activates the workflow. AI-generated content is validated and cannot introduce a completion redirect. Activation remains a separate action in the existing workflow editor.

## Sharing and measurement

The frontend serves the public `/forms/embed.js` asset without a session while keeping `/forms` and its editor private. Hosted links, inline iframes, inset side sheets, and simple native HTML examples are available. Existing `popup` and `sidebar` embed aliases render inset sheets. Embeds accept messages only from their own frame and the expected origin; campaign attribution is allowlisted.

Views, starts, and completions are deduplicated by random visitor session, rather than presented as verified people. Step reporting uses the current revision. SMTP acceptance, tracking signals, and form completion do not establish inbox delivery, bookings, or revenue.

## Validation and rollout

The local checks cover schema/default/branch validation, consent, suppression, tenant isolation, version changes, secure resume, receipt replay, draft transactions, queue failure/retry identity, and analytics. Database integration tests use SQLite; production PostgreSQL migration and actual provider delivery are separate deployment checks.

September 22 validation passed: the full Go test suite, `go vet ./...`, `go build ./...`, and race tests for marketing, handlers, AI, automation services, and sending. The paired frontend passed 85 tests across 18 suites and its production build.

Startup migrations add form revisions, partial progress, analytics events, and completion events, and extend the existing form/submission/receipt records. Roll out the backend and its task worker before the matching frontend. This document describes the implementation; it is not a production deployment record.
