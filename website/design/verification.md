# Website verification

The website is a separate project in `posthoot/website`; the client and server working trees were checked and remain clean.

Verified in actual Google Chrome:

- Floating navigation, anchor destinations, mobile menu, template gallery and complete previews.
- Every homepage email preview uses the first 15 entries in the existing Xem gallery. The product editor uses Show Your Work; the AI section uses August’s Most Wanted. Both editor screenshots were captured at 2× display resolution.
- Product-tour keyboard focus, Escape, focus restoration, and scroll locking.
- Narrow layouts at 390px and 320px, plus desktop at 1440px.
- Reduced motion, WebGL fallback, and server-rendered content without JavaScript.
- Interactive analytics: 180 days × 6 campaigns of deterministic sample records; 7/30/90-day filtering; campaign filtering; animated metrics; stacked series; message/click switching; growth chart; device selection; heatmap; sortable table; CSV export; accessible daily data table.
- The accepted-message summary reconciles to the selected daily data. Device counts and time buckets reconcile to clicks. Prior-period comparisons use a separate period of the same length.
- Automated axe checks cover the page, product dialog, mobile FAQ, and live analytics.
- Production build, standalone server, public metadata, image loading, response headers, and 11 browser tests.

Local URLs are configured in the ignored `.env.local`. Set production public URLs before building for deployment. No deployment or GitHub publication was performed for this project.

## Feature, journal, and real-dashboard extension — September 10

- Added homepage forms/CRM/inbox tabs, SMTP/API/webhook feature cards, three sourced legacy testimonials, and recent journal stories.
- Centralized primary CTAs in Xem iris (#5b3cc4), white text, with Tailwind hover/focus states. No custom CSS was added.
- Ten complete, original AI-assisted Markdown articles, 870–983 words, with source links and locally optimized photographic covers. Published static routes, searchable/category-filtered index, article contents, related reading, RSS, per-page canonical/social metadata, BlogPosting/BreadcrumbList and site Organization/WebSite structured data.
- Content and primary navigation available in server-rendered HTML without JavaScript. Raw Markdown HTML skipped; frontmatter and asset paths validated.
- Website production build/typecheck passed. All 16 Chrome Playwright tests passed: original tour/gallery/demo analytics coverage plus article rendering, metadata/JSON-LD, RSS/sitemap inclusion, ten covers, internal references, category/search/empty states, feature keyboard controls, testimonials, iris CTA color, no-JS article, and 1440/390/320px checks. Axe checks had no violations for the audited website/journal/article surfaces.
- Client production build/typecheck passed after chart changes. Browser checks exercised the real chart components using reports produced by isolated PostgreSQL test queries: stack switches, device selection, heatmap cells, table, CSV/summary reconciliation, cohort cutoff/pagination, invalid dates/retry, old-API availability messaging, and empty states. No real messages were sent.
- New dashboard labels were darkened after axe found insufficient contrast in the app's muted color token. Donut slices now carry accessible labels. Final chart audit passed, including 390/320px layout.
- Backend analytics and handler tests passed against isolated PostgreSQL. New tests cover cross-workspace exclusion, list/campaign filters, timezone boundaries, repeated/older clicks, invalid event times, and reconciled totals. Representative report: 10k contacts, 50k messages, ~0.8s locally. This is not a production latency guarantee.
- Website demo remains explicitly synthetic. Actual dashboard code reads the API; device/time charts count observed events and do not claim verified humans or optimal send times. Subscriber growth requires recorded history.
- Final captures in design/review include homepage-iris, features-final, testimonials-final, blog-article-final, blog-reading-final, blog-mobile-final, and dashboard-real-*.

Deployment is separate: the new website has no remote/hosting target configured. Set production origins at build time; local .env.local uses localhost for review. The hosted backend must receive the additive report fields before those charts can appear with live workspace data. Current changes remain local.

## Static Cloudflare deployment — September 10

Converted production output from a Next.js standalone server to `output: export`. Image assets are served directly, metadata routes and RSS are prerendered, and all ten article HTML files are present. Wrangler 4.130.0 validates the assets-only configuration with a successful deployment dry run (206 exported files). Production builds override local preview origins with explicit HTTPS deployment origins. Cloudflare `_headers` carries security, caching, and RSS headers. Unknown routes use 404-page handling rather than an SPA fallback. The optional Docker image now serves the same output with unprivileged Nginx.

The repository includes Wrangler deployment commands and a GitHub Actions static-build check. No Cloudflare deployment or production domain change was performed as part of configuring it.

All 16 Chrome Playwright tests passed against the actual Wrangler static preview on port 3003. Direct HTTP checks confirmed RSS content type, security headers, 404 status for unknown articles, and trailing-slash redirects. The exported homepage contains production origins and no Next image-optimizer URLs.
