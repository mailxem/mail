# Xem MCP server

Connect an MCP client to Xem for newsletter drafts and schedules, campaigns, analytics, contact lists, and CSV imports/exports. Supports hosted HTTPS (Streamable HTTP) and npm-based local stdio.

## Connect over HTTPS (hosted)

Use a client that supports Streamable HTTP and custom Authorization headers:

```json
{
  "mcpServers": {
    "xem": {
      "url": "https://mcp.xem.email/mcp",
      "headers": { "Authorization": "Bearer your-workspace-api-key" }
    }
  }
}
```

This is API-key authentication, not an OAuth login flow. Clients that only support OAuth must use local stdio instead. Use your client's secret store where available. Each request validates the supplied Xem API key and creates an isolated protocol server and API client. No shared service key, sessions, or cross-client state. Revoked and expired keys are rejected on the next request. Tool endpoints enforce resource permissions independently. The hosted URL becomes available after the Compose deployment below.

## Local stdio through npm

Requires Node.js 20 or newer. After publishing this 2.x release, configure:

```json
{
  "mcpServers": {
    "xem": {
      "command": "npx",
      "args": ["-y", "@xem.email/mcp@2", "--stdio"],
      "env": { "XEM_API_KEY": "your-workspace-api-key" }
    }
  }
}
```

Alternatively install `npm install -g @xem.email/mcp@2` and use `xem-email-mcp --stdio`. To work from source, run `npm ci && npm run build` and launch `node /absolute/path/to/mcp/build/index.js --stdio`.

`XEM_API_KEY` uses the upstream `X-API-Key` header. In stdio mode only, `XEM_API_TOKEN` can supply a JWT bearer token; the API key wins if both exist. `XEM_API_BASE_URL` defaults to `https://api.xem.email/api/v1`. HTTPS is required except on loopback for development. Redirects are rejected. Credentials and workspace IDs cannot be supplied as tool arguments.

## Host with Docker Compose

1. Deploy the accompanying Xem API changes first, including `/api/v1/mcp/authorize`, the batch-contact endpoint and campaign-draft endpoint.
2. Point `mcp.xem.email` (or your own hostname) at the Docker host. Allow incoming TCP ports 80 and 443 for Caddy's certificates and HTTPS. These ports must be available.
3. Set `MCP_DOMAIN` and `XEM_API_BASE_URL` in `.env` if overriding defaults. **Do not set a shared XEM_API_KEY or XEM_API_TOKEN for the hosted service.**
4. After the first successful container workflow, run:

```sh
docker compose pull
docker compose up -d
docker compose ps
curl --fail https://mcp.xem.email/health
```

Caddy provisions and renews TLS certificates; keep the `caddy_data` volume across upgrades. Only Caddy publishes ports. The MCP container runs as an unprivileged user with a read-only filesystem and resource limits. It uses a 2 MiB request limit, per-key rate limits (120 requests/minute per process), and a 128-request concurrency cap. Add an edge rate limiter for large public deployments. Native clients may omit Origin; browser clients must match the public origin or an explicit comma-separated `XEM_MCP_ALLOWED_ORIGINS` allowlist. Cookies and forwarded host headers are not used for authentication.

The HTTP service is stateless: POST `/mcp`; GET/DELETE return 405. `/health` checks process health, not API availability. For an end-to-end check, connect with a restricted test key and call `get_contact_lists`. Read/write behavior and key revocation are covered by integration tests. No real contacts or mail are used in tests.

For an existing TLS proxy, run `xem-email-mcp --http`, set `HOST`, `PORT`, `XEM_MCP_PUBLIC_URL`, and proxy `/mcp` while preserving Host and Authorization. Keep the backend on a private network. All incoming API keys go only to the fixed, operator-configured Xem API base URL; never use a user-selected upstream.

## Container builds in GitHub Actions

`.github/workflows/container.yml` runs package tests, validates Compose, builds the image, and smoke-tests its health and authentication checks on pull requests. It publishes only from this repository's `main` branch or version tags, after validation succeeds. Manual runs can publish from `main` or a version tag too.

Images are pushed to `ghcr.io/mailxem/mcp` for both `linux/amd64` and `linux/arm64`, with an SBOM and build provenance:

- `main`: the current successful main-branch build (Compose default).
- `v2.0.0`: a release tag; the tag must match the package version.
- `sha-<full-commit-sha>`: a build identified by its source commit.

The workflow summary provides a digest for reproducible deployments. Set `MCP_IMAGE=ghcr.io/mailxem/mcp@sha256:...` in `.env`, then run `docker compose pull && docker compose up -d`. Keep the previous digest to roll back by changing `MCP_IMAGE` and running those commands again. Publishing an image does not automatically restart your production service.

Publishing uses the repository's `GITHUB_TOKEN` with job-scoped `packages:write`; no Docker Hub secret is needed. After the first publish, make the GHCR package **public** in its package settings for unauthenticated pulls, or authenticate the deployment host with a read-only package credential. Organization package policies may require an administrator to grant publishing access. This workflow does not publish the npm package.

For local development, build directly with `docker build -t xem-mcp:local .`, then use `MCP_IMAGE=xem-mcp:local docker compose up -d` without pulling. Actions are pinned to commit SHAs and Dependabot proposes updates weekly.

## Tools

| Tools | Behavior |
| --- | --- |
| `get_marketing_options` | Audiences, templates, starter designs and sender IDs without sender passwords |
| `get_contact_lists`, `create_contact_list` | Paginated lists and empty list creation |
| `get_contacts`, `get_contact` | Paginated contacts with optional list, email and status filters; individual lookup |
| `create_contact`, `add_contacts` | Create one contact or an atomic batch; batch defaults to preview |
| `import_contacts` | Inline CSV, explicit column mapping, preview by default |
| `export_contacts_csv` | One page of CSV with page/total/hasMore metadata and formula escaping |
| `unsubscribe_contact` | Suppress an active contact, preserving existing bounce/complaint suppression |
| `create_campaign`, `get_campaigns`, `get_campaign` | Safe campaign drafts and status lookup |
| `get_campaign_metrics`, `get_audience_metrics` | Opens, clicks, accepted messages, bounces, unsubscribes and rates |
| `create_newsletter`, `get_newsletters` | Reusable draft newsletters; listing capped at 200 |
| `schedule_newsletter`, `pause_newsletter` | Activate delivery on an explicit future schedule, or pause future editions |
| `get_newsletter_metrics` | Delivery counts for the latest 50 editions; use edition IDs for campaign engagement metrics |
| `send_email` | Queue an authorized individual email |

Campaign drafts require `templateId`, `listId`, and `smtpConfigId` from your workspace. Creating a draft never sends. Newsletter scheduling replaces its configuration: supply all required fields, including cadence, timezone, future `nextSendAt` and sender postal address. Pausing does not cancel editions already materialized. Analytics default to the last 30 calendar days; `to` is exclusive and ranges are limited to 366 days. SMTP acceptance does not guarantee inbox placement; opens and clicks may include automated activity.

## CSV import

```json
{
  "listId": "11111111-1111-4111-8111-111111111111",
  "csv": "Email Address,Given Name,Company\nada@example.com,Ada,Acme\n",
  "mappings": { "email": "Email Address", "firstName": "Given Name", "company": "Company" },
  "dryRun": true
}
```

Mapping direction is **contact field → CSV header**. Supported fields: email (required), firstName, lastName, phone, company, country, city. Unmapped columns are ignored. Each request accepts at most 500 rows / 1 MiB. All rows must validate before anything is written. Preview returns `wouldCreate`, `skipped` and `total`; repeat the input with `dryRun:false` to commit. Counts are recalculated on commit. Each batch is atomic; split larger datasets into batches. Existing addresses in the list, including deleted or suppressed contacts, are skipped and never reactivated. No remote URLs, local file paths, or uploaded file IDs are accepted. Export returns one page at a time to the MCP client; increment `page` while `hasMore` is true. Data may change between pages.

## Permissions and security

Grant only needed resources: `lists:read/create`, `contacts:read/create`, `campaigns:read/create`, `analytics:read`, and `emails:create`. The options tool additionally requires `templates:read` and `smtp_configs:read`. A resource's create grant also permits reads; admin grants are supported. Scheduling requires campaign write access and validates the template, audience, sender, TLS support and postal address server-side.

The MCP host controls user approval. Tool annotations identify read-only and delivery-affecting operations, but they are hints, not authorization. Contact data returned by tools is visible to the client; configure that client's data policies appropriately. Upstream API requests time out after 30 seconds, responses are capped at 2 MiB, upstream error bodies are redacted, and writes are never retried automatically. After a timeout, inspect server state before retrying. Batch contact inserts skip existing addresses.

## Migration from 1.x

This is a breaking 2.0 release. Remove `token`/`teamId` tool arguments and `--token`/`--team-id` CLI flags; use environment configuration. Pagination starts at 1. `create_campaign` uses `htmlBody` and requires template/audience/sender IDs; it creates drafts only. `create_contact` takes `{listId, contact}`; `add_contacts` takes `{listId, contacts, dryRun}`. CSV import now accepts inline CSV and field-to-header mappings instead of `fileId`. Contact `name` becomes `firstName`/`lastName`. Email `data` is a key/value object. Deploy the accompanying server changes before using this MCP release.

## Development

```sh
npm test
npm audit
```

Tests use a mock HTTP transport and do not send email or change real contacts.

## Release order

Deploy the matching Xem API, launch and verify the Compose service, then publish this package with `npm publish --access public` and deploy the website. Version 2.0 is breaking; do not direct users to the old 1.x package for these tools. Run `npm pack --dry-run` before publishing.

### Dashboard assistant tools

The catalog now includes 42 tools: existing campaign/audience/newsletter operations plus templates and starter import, draft signup forms, contact notes/stages/tags, automation inspection/pause, outbox history, and managed-sending readiness/domain checks/pause. All inputs reject workspace overrides. New writes are annotated for client review; the dashboard stages every write before approval.

Managed-sending tools require a current admin-bound Xem assistant credential. Ordinary API keys remain denied by those backend routes. The catalog deliberately does not expose credential issuance, billing, team administration or operator approval. Upgrade the backend with its assistant credential support before using these tools. This change does not deploy the hosted MCP or publish an npm release.
