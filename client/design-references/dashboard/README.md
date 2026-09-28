# Dashboard Email / Campaigns review

Validated September 10, 2026 in real Chrome with an isolated local test session and intercepted API responses. Screenshots contain synthetic data, not a customer workspace.

Verified: standalone sends with no campaigns; campaign empty state; switching tabs; one Dashboard heading; inclusive date ranges forwarded to the API as an exclusive upper bound; empty email state; API failure and retry; narrow mobile layout without horizontal overflow.

The API is independently covered by PostgreSQL tests for workspace isolation, no campaign records, duplicate tracking events, timezone boundaries, deleted/test records, missing send timestamps, and campaign-linked versus standalone totals. Browser fixtures do not replace those API tests.
