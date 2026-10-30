# Xem analytics: product review and proposed implementation

Reviewed September 10, 2026. Status: proposal, not implemented.

## Recommendation

Rebuild audience analytics around four decisions: **Is my audience growing? Who is responding? Where are we losing subscribers? What should I do next?** Correct the metric definitions and tenant isolation before adding visual polish. Reuse the existing contacts, lists, tags, forms, campaigns, newsletter editions, and CRM.

Evidence comes from inspection of the audience dashboard in the user's Chrome and the local client/backend source. Backend findings describe this checkout; the deployed API was not separately audited or tested across tenants. No production data, campaigns, or application code were changed during this review.

## What is wrong today

| Finding | User impact | Source |
| --- | --- | --- |
| “Audience Segments” uses contact UUIDs as categories and tracking-event counts as contact counts. | The chart describes neither segments nor audience size. | `client/components/analytics/audience-insights.tsx:102` |
| Engagement is `(opens + clicks) / all tracking events`. Contacts without events disappear. | One click after 100 sends can still appear 100% engaged; the audience most in need of attention is omitted. | `server/internal/handlers/tracking_handler.go:891` |
| One component multiplies the ratio by 100; another appends `%` directly. | The same value can appear as 100% and 1.0%. | `client/components/analytics/audience-insights.tsx:106`; `audience-behavior.tsx:126` |
| Demographics and preferences are initialized but never populated. Empty preference confidence divides by zero. | Empty charts imply a broken product; confidence can display NaN%. | `server/internal/handlers/tracking_handler.go:885`; `client/components/analytics/audience-preferences.tsx:147` |
| Lines/areas connect unrelated contact IDs; count and percentage share an axis. | Visual shapes imply trends and comparisons that have no meaning. | Audience insights and behavior components |
| The audience endpoint loads every contact, then queries tracking separately for each. All three panels request this same endpoint. | Work grows with audience size; redundant requests amplify the cost. | `GetAudienceInsights`; three audience components |
| The authenticated audience route accepts an arbitrary `teamId` query value; inspected middleware validates token-team membership but does not bind this GET query to that team. | Code-level cross-tenant data-access risk. Authentication alone does not establish ownership of the queried resource. | `client/app/api/analytics/audience/route.ts`; `server/internal/api/middleware/auth.go`; `server/internal/routes/tracking_routes.go`; `GetAudienceInsights` |

The current page also spends prominent space on Contact lists, Tags, and Inbox shortcuts instead of an analytical summary. Audience analytics appears under Settings & workspace even though Analytics already exists in navigation.

## Problems shared with the wider analytics feature

- **Inconsistent denominators:** `GetTeamOverview` filters tracking activity by date but counts all-time email records, including records that may not have been sent. This makes date comparisons unreliable.
- **CTR confused with CTOR:** `processEmailAnalytics` calls unique clickers divided by unique openers “ClickRate.” Team analytics uses a different denominator. These are different metrics.
- **Incorrect per-link uniqueness:** every link is assigned the global set of unique clicking contacts, rather than its own clickers (`tracking_handler.go:465`).
- **Fragile recipient identity:** tracking aggregation keys uniqueness by ContactID, which is nullable on email records. Messages lacking contacts can collapse into one identity.
- **Misleading labels:** “Top campaigns” selects the five newest campaigns; “average read time” is time between open and click; a URL count distribution is presented as a heatmap.
- **Unreliable trend ordering:** `calculateEngagementTrend` builds monthly rates by iterating a Go map without sorting months before comparing halves.
- **Overclaimed recommendations:** proposed optimal send times combine separate day/hour activity totals rather than actual day-hour observations, without accounting for send volume. The confidence values are heuristic.
- **SMTP acceptance is not confirmed delivery:** `internal/utils/smtp.go` marks SENT after successful SMTP submission. The inspected tracking event definitions do not supply confirmed delivery. Do not label this “inbox placement” or assume it proves delivery.
- **Incomplete unsubscribe history:** the current marketing unsubscribe handler changes contact status without appending an unsubscribe analytics event. Counting legacy tracking events therefore misses this path.

## Proposed audience page

One page title, aligned with the current app design. Controls: last 30 days by default, previous-period comparison, workspace timezone, list and tag filters, and data freshness. Keep the existing dashboard scroll container. Use Tailwind and existing shadcn chart primitives.

| Position | Module | Decision it supports |
| --- | --- | --- |
| Top row | Active subscribers now; new contact records in period; recipients reached; recipients who clicked | How large is the reachable audience, and are people responding? |
| Primary chart | Daily unique recipients reached and daily unique clicking recipients, clearly labeled as activity in each day | Is activity changing? Counts share a unit; daily unique values are not added to obtain period uniqueness. |
| Secondary chart | Horizontal bars for mutually exclusive engagement cohorts | Which audience groups should I examine? |
| Main table | Lists: active subscribers, reached recipients, unique clicking recipients, audience click rate, known suppression counts | Which lists need attention? Server-side sorting, pagination, total count, and drilldown. |
| Lower section | Acquisition sources, beginning with forms/imports/unknown and explicit coverage | Which sources bring contacts who later respond? |
| Contextual actions | View cohort in CRM; open list; create a campaign draft for a reviewed audience | What useful next step can I take? No automatic send or suppression. |

“New contact records” must not be mislabeled “new opt-ins”: imports and list duplicates are not necessarily new subscribers. Until history is reliable, show current suppression counts separately from period unsubscribe rates. Replace these interim metrics with true net subscriber growth and opt-out rates after instrumentation ships.

Remove the unsupported demographics and preference panels. Optional geography can return later with a clear distinction between contact-provided location and location inferred from tracking IP, plus coverage. Neither is evidence of demographic attributes.

### Cohort rules

Default lookback: 90 days, ending at the report cutoff. Apply these rules in order so every eligible recipient belongs to exactly one cohort:

1. **Clicked recently:** at least one tracked click in the last 30 days.
2. **Clicked earlier:** no click in the last 30 days, but at least one in days 31–90.
3. **No tracked clicks after repeated sends:** no click in 90 days and at least three accepted marketing messages with known click tracking enabled in that window.
4. **Not contacted:** no accepted marketing messages in the 90-day window.
5. **Insufficient evidence:** everyone else, including unknown tracking coverage and fewer than three observed send opportunities.

Three sends is a transparent product rule, not a validated churn model. Until tracking capability is recorded reliably, assign uncertain cases to insufficient evidence. Opens remain secondary directional signals; do not treat missing opens as proof of inactivity. Clicks are “tracked clicks,” not verified human clicks until bot/proxy classification is implemented.

## Metric contract

Use explicit populations. An audience metric counts distinct recipients; a campaign performance metric counts message-recipient deliveries. Never average campaign percentages to produce a workspace percentage.

| Metric | Definition |
| --- | --- |
| Active subscribers now | Distinct recipients with an active membership in the selected lists, excluding applicable suppression and deleted records. Label as current, not historical. Establish recipient identity and list/workspace consent precedence before implementation. |
| Recipients reached | Distinct recipients with an accepted, non-test marketing message sent in the selected period. “Reached” tooltip must say SMTP accepted; it does not prove inbox delivery. |
| Audience click rate | Distinct reached recipients who clicked one of those messages by the report cutoff / distinct reached recipients. |
| Campaign click rate | Distinct accepted message-recipient pairs clicked by cutoff / accepted message-recipient pairs in the send cohort. Label the accepted-message denominator until verified delivery is available. |
| Click-to-open rate | Distinct message-recipient pairs both opened and clicked / distinct opened message-recipient pairs. Secondary metric; disclose tracking limitations. |
| Net subscriber growth | First eligible subscriptions + reactivations − transitions out of eligibility, within the same identity and consent scope. Requires historical transitions. |
| Campaign unsubscribe rate | Distinct accepted message-recipient pairs with attributable opt-outs / accepted message-recipient pairs. Requires message attribution and complete opt-out capture; otherwise unavailable. |
| Known bounce / complaint rates | Deduplicated attributable outcomes / explicitly defined accepted-message cohort. Show event-source coverage; unknown outcome is not successful delivery. |

Store rate values as fractions from 0 to 1 and format once in the client. Include numerator and denominator in API responses and tooltips. Use percentage-point differences for rates. Zero denominator returns null with an explanation; missing data is never silently rendered as zero. A comparison with a zero baseline does not become an infinite growth percentage.

For identity, contacts are currently list-scoped. A stable workspace recipient key must map existing duplicate memberships without merging away list-specific consent. Do not blindly deduplicate by display name or use provider-specific email normalization. Message-recipient identity is needed for multi-recipient messages and missing ContactID cases; report unresolved historical mappings as coverage gaps.

Separate **send-cohort reports** (messages accepted during the period, outcomes observed by cutoff) from **activity reports** (events occurring during the period, possibly for older messages). Do not silently mix them. Compare equally mature send cohorts, or explicitly label unequal observation time in live comparisons.

## API and data changes

Reuse the Go analytics route group and Next proxy; replace the audience response with versioned aggregate data rather than a map of contact IDs.

- Derive the team from verified authentication context. Reject conflicting legacy teamId parameters; verify ownership of every list, tag, campaign, email, and export. Enforce `analytics:read` as the actual resource permission, not merely a generic permission for the HTTP method.
- Validate dates, timezone, maximum range, bucket granularity, pagination, and allowed sort columns. Use half-open ranges `[from, to)` and timezone-aware day boundaries.
- Aggregate in SQL with bounded result sets. Avoid loading all contacts/events into Go. Index according to query plans; keep raw-event details out of summary responses.
- Use one shared React Query request keyed by team, period, timezone, filters, and metric version; propagate cancellation and isolate stale data during team switches.
- Return summary/cohorts/series from the main endpoint, with a separate paginated list breakdown. Include `asOf`, `metricVersion`, `coverage`, and per-section availability; never expose unsupported fields as invented confidence scores.
- Record idempotent subscription transitions for forms, imports, manual changes, unsubscribe, resubscribe, bounces, and complaints. Preserve source and scope. Record message attribution when known; do not invent it when absent.
- Capture tracking capability and event source per send. Add bot/proxy classification with raw and filtered counts distinguishable. Extend current event processing rather than introducing another competing analytics store without need.
- Reuse newsletter edition-to-campaign relationships for edition comparison. Reuse form submissions and import IDs for partial acquisition attribution. Historical unknown source and missing consent history remain explicit unknowns.

Example proposed response shape (not real workspace data):

```ts
type Rate = {
  value: number | null; // Fraction, never preformatted percent.
  numerator: number;
  denominator: number;
  unavailableReason?: string;
};

type AudienceAnalytics = {
  metricVersion: "audience-v2";
  asOf: string;
  period: { from: string; to: string; timezone: string };
  filters: { listIds: string[]; tagIds: string[] };
  summary: {
    activeSubscribersNow: number;
    newContactRecords: number;
    recipientsReached: number;
    recipientsClicked: number;
    audienceClickRate: Rate;
  };
  activitySeries: Array<{
    date: string;
    recipientsReached: number;
    recipientsClicked: number;
  }>;
  cohorts: Array<{ key: string; label: string; count: number }>;
  coverage: {
    subscriptionHistoryFrom: string | null;
    trackingCapabilityKnown: Rate;
    recipientIdentityResolved: Rate;
    botFilteringAvailable: boolean;
  };
};
```

Validate chart and summary contracts independently: daily activity includes events from older sends, while the summary click rate refers to its stated send cohort. Tooltips must explain that distinction. Paginated tables return `items`, `page`, `pageSize`, and server-calculated `totalCount` from the same filters.

## Delivery order and acceptance

**P0 — trust and isolation.** Bind analytics queries to the authenticated tenant; validate resource permissions and ownership including exports. Correct denominator/rate units, link uniqueness, chronological trends, and unavailable states. Remove unsupported panels and rename misleading metrics. Regression fixtures cover cross-team requests, insufficient permissions, repeated tracking, missing contact IDs, zero sends, date boundaries, and test/failed messages.

**P1 — useful audience dashboard.** Ship the proposed controls, current audience summary, factual activity chart, transparent cohorts, and paginated list table. Reuse CRM filters for drilldowns, preserving filter semantics and cutoff. Validate filter changes, retry, team switching, empty/new accounts, mobile layout, accessibility, tooltip units, and scroll containment in the browser. Fixture totals must reconcile between summary and drilldown; aggregates must not perform one query per contact. Record query plans and performance on a representative large dataset before release.

**P2 — trustworthy growth and health.** Add append-only subscription/suppression history, complete opt-out tracking, acquisition attribution, and tracking coverage. Backfill only facts supported by stored records. Show the start of reliable historical coverage. Tests cover duplicate/retried events, list-specific vs workspace suppression, imports, resubscription, and out-of-order outcomes. Unlock net growth and period unsubscribe charts only when the source paths reconcile.

**Later — advanced optimization.** Add content affinity only with tagged content and sufficient observed behavior; send-time suggestions only with send-opportunity normalization and credible evidence. Revenue attribution requires conversion instrumentation and a declared attribution window. These should follow reliable core reports.

Unify navigation under Analytics: Overview, Audience, Campaigns & newsletters, and Delivery health. Campaign reports should emphasize comparable rates with counts, sample sizes, edition trends, and per-link tables. Keep unsupported revenue, inbox placement, read-time, and confidence claims out of the interface.

The first release succeeds when a marketer can identify a specific audience problem, see the evidence and its limitations, and open the exact affected cohort without translating UUIDs or guessing what a percentage means.

## Implementation follow-up

The core reporting, tenant isolation, dashboard, cohort drilldowns, and subscription-history instrumentation are implemented. See [the release notes](/analytics/xem-analytics-release) for exact behavior, validation, deployment requirements, and remaining data limitations.
