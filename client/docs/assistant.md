# Ask Xem

The dashboard home is a conversation with Xem. The existing overview is at `/dashboard`; `/chat` redirects to `/`. AI orchestration and streaming live in Next.js. The Go API only authenticates people, provisions short-lived scoped credentials, and serves product operations. Xem's hosted MCP is the tool transport.

## Run

Use Bun 1.3.5 (the container runtime) or Node 22+. Install with `bun install --frozen-lockfile`. Set these server-only runtime variables:

| Variable                             | Default / purpose                                                          |
| ------------------------------------ | -------------------------------------------------------------------------- |
| `XEM_ASSISTANT_ENABLED`              | `true`; set `false` to disable chat and approvals                          |
| `AI_PROXY_BASE_URL`                  | `https://ai-proxy.synehq.com/v1`                                           |
| `ASSISTANT_TEAM_DAILY_REQUEST_LIMIT` | `500` model turns per workspace per UTC day; configurable from 1 to 10,000 |
| `AI_PROXY_MODEL`                     | `gpt-4.1`, verified for streaming and tool calls                           |
| `AI_PROXY_API_KEY`                   | Optional credential if the proxy requires it                               |
| `XEM_MCP_URL`                        | `https://mcp.xem.email/mcp`                                                |
| `ASSISTANT_REDIS_URL`                | Required in production; private, authenticated Redis with persistence      |
| `INTERNAL_API_URL`                   | Go API base including `/api/v1`; falls back to `NEXT_PUBLIC_API_URL`       |
| `NEXTAUTH_URL`                       | The public app origin; also used to check mutation origins                 |

No AI, MCP or Redis secrets belong in `NEXT_PUBLIC_*`. `INTERNAL_API_URL` is an operator-trusted, server-only address and accepts private HTTP service URLs such as `http://backend:9001/api/v1` on the deployment network. Use HTTPS when backend traffic crosses untrusted networks. Public API, AI and MCP endpoints require HTTPS except on loopback. All endpoints reject embedded credentials, query strings and fragments; redirects are rejected. The proxy's model list includes entries that currently reject inference requests, so test both streaming and tools before changing the model. Its tool-only assistant messages require an empty content string rather than null; the provider adapter normalizes this without dropping tool calls.

Deploy the Go credential endpoint and the expanded MCP catalog before enabling the Next.js assistant. Existing hosted MCP versions still expose their own tools through discovery; newly added tools become available after the MCP rollout. Keep backend and MCP access on trusted hosts. Configure Redis persistence (AOF), backups, a private network, and TLS/auth where traffic crosses hosts. Use `noeviction` and alert on storage errors, memory pressure and failed writes. Redis holds the authoritative transcript and action state; a storage error fails closed. Do not restore a stale Redis backup and resume pending actions: clear the assistant namespace after a disaster recovery restore so already executed writes cannot be replayed.

In development only, missing Redis uses bounded process memory. This is for local development, resets on restart, and is never a production fallback. `/preview` uses the shadcn AI SDK helper to stream a scripted fixture and never calls an API or sends email. Production builds return 404 for that preview.

## Access and action review

Every request verifies the NextAuth session against the Go API's current user record. Workspace/user identifiers, role, model, tool catalog and tool arguments cannot be supplied by the browser. The Go API issues a five-minute `xem_bot_` credential for the current user and team. Only its SHA-256 digest is stored. Each API/MCP request rechecks expiry, deletion, membership and (for write credentials) current admin role. A stored digest cannot be used as a credential. The legacy email-write grant is not applied to assistant keys.

Members receive read-only tools. Admins can propose every mutation exposed by the hosted MCP. No mutation executes during a model turn: it produces a review card containing saved arguments. Approval checks the user/team, a ten-minute expiry and a fingerprint of the tool definition. A distributed user/workspace lock serializes requests, and `running` is persisted before dispatch. Replays are rejected. If a process, connection or provider fails after dispatch, the outcome remains running/unknown; inspect the resource before preparing another action. This is not a claim of exactly-once execution across network or storage failures.

The bot cannot issue reusable SMTP credentials or administer developer API keys, users, billing or operator sending approval. Those operations remain in their dedicated screens. Managed-sending status is available to admin-bound assistant credentials; normal customer API keys remain denied. The assistant cannot enable a managed sender that has not passed existing readiness and operator checks.

The current expanded catalog contains 42 tools across audiences, contacts, CRM notes/stages/tags, campaign drafts, newsletters, analytics, templates, signup-form drafts, automation inspection/pause, outgoing messages and managed-domain readiness/pause. Unknown tools require review regardless of their remote read-only annotations. MCP definitions and results are treated as data, not instructions. Remote images and raw HTML are not rendered in chat; secrets are redacted from results. Read-only tools can reveal workspace data, so prompt instructions minimize bulk retrieval and exports.

## Limits and data

- 20 requests/minute per user/workspace; 100 model turns/minute and 500/day per workspace, enforced in Redis. One active turn/action per user/workspace.
- 8,000 characters per user message; 20 KB request body; up to six model steps and 2,400 output tokens per step; 60-second generation timeout; no automatic provider or MCP write retries.
- At most 60 messages/actions per conversation, four outstanding proposals, 30 saved conversations, and 500 KB per stored conversation. Tool inputs/results are bounded.
- Conversations and reviewed arguments/results are private to the authenticated user/team and expire seven days after their last save. History deletion removes that record and its pending proposals. Redis backups need a matching retention policy.
- Messages and relevant tool results are sent to the configured AI proxy. The UI explains this. No assumption is made about the proxy's training or retention policy; the operator must disclose its actual policy before public rollout.
- Backend API-key usage records remain under the backend's audit retention policy. Expired assistant keys are soft-deleted on the next issuance for that user; audit rows are retained.

## Verification

`npx jest --runInBand` covers origin checks, untrusted history/arguments, workspace isolation, role checks, locks, limits, schema drift, expired review, decline, replay and unknown outcomes. The Go tests exercise digest authentication, default-grant exclusion, membership changes, demotion and managed-sending restrictions. MCP tests cover protocol isolation and strict schemas.

With the sibling MCP repo built, `bun run scripts/verify-assistant.mjs` exercises the real proxy and real MCP HTTP transport against synthetic local workspace records. It performs a read, stages a list creation, approves it once and rejects replay. It does not send email or access production workspace data. This is an explicit live proxy test, not part of routine unit CI.
