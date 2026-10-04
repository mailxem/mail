# Portable Xem images

The self-hosted app can use bundled or existing PostgreSQL, Redis and S3-compatible storage independently. MinIO is supported for bundled object storage. PostgreSQL and Redis are still required; choosing an existing service removes the need to run another instance.

This configuration covers the frontend and backend. Managed SES, public SMTP submission, payments, MCP and hosted AI require separate operator configuration. The adjacent `xem.toml` describes Xem's managed offering; it is not the self-hosted preset.

## Images

After a frontend release, use `theboringhumane/xemapp:sudo-self-hosted` with `theboringhumane/xemgo:sudo`. The frontend release also publishes `<version>-self-hosted` and `sha-<commit>-self-hosted` tags for both amd64 and arm64. Pin the resolved digests when deploying. Public images can be pulled without Docker login, subject to registry rate limits.

The self-hosted frontend is compiled with `NEXT_PUBLIC_API_URL=/api/v1` and no billing redirect. Set `INTERNAL_API_URL=http://backend:9001/api/v1` at runtime for server-side authentication and API calls. Set `NEXTAUTH_URL` to the installation's HTTPS origin and provide a strong `AUTH_SECRET` through the secret manager. Disable `XEM_ASSISTANT_ENABLED` unless its separate dependencies are configured.

The existing frontend `sudo`, `latest`, version and commit tags retain their configured hosted API/payment origins. The release workflow continues deploying that hosted digest to Xem Cloud. Operators building their own image can still override the public URLs with Docker build arguments; runtime `NEXT_PUBLIC_*` variables cannot rewrite the browser bundle.

## Route the public origin

Place an HTTPS reverse proxy in front of the self-hosted frontend. Preserve these paths:

| Public path | Private upstream |
| --- | --- |
| `/api/v1` and `/api/v1/*` | Backend port 9001 |
| `/public` and `/public/*` | Backend port 9001, including exported forms |
| `/t` and `/t/*` | Backend port 9001 for tracking |
| Other app paths, including `/api/auth/*` | Frontend port 3000 |

The frontend does not proxy `/api/v1` itself. Set backend `PUBLIC_URL`, `DASHBOARD_URL` and `CORS_ALLOWED_ORIGINS` to the installation's public origin. Keep database, Redis and object-storage administration ports private.

## Choose PostgreSQL and Redis

Supply the backend's `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD` and `POSTGRES_SSLMODE` for either the bundled database or an existing PostgreSQL database. Use the existing database's required TLS mode and a dedicated database/user.

Choose Redis independently through `REDIS_HOST`, `REDIS_PORT`, `REDIS_USERNAME`, `REDIS_PASSWORD`, `REDIS_DB` and `REDIS_USE_TLS`. Use the endpoint's required authentication and TLS settings. See `server/.env.example` for the remaining application settings, including signing keys.

## Bundled MinIO

Configure a private, persistent MinIO service and pass its credentials as secrets to the backend. These example values describe a private service named `storage` and an HTTPS proxy on the app's public origin:

```dotenv
STORAGE_PROVIDER=s3
S3_BUCKET_NAME=xem-files
S3_REGION=us-east-1
S3_ENDPOINT_URL=http://storage:9000
S3_PUBLIC_ENDPOINT_URL=https://app.example.com
S3_DISABLE_ACL=true
S3_CREATE_BUCKET=true
```

Supply `S3_ACCESS_KEY` and `S3_SECRET_KEY` separately. The backend uploads through the private endpoint and signs browser reads for the public endpoint. Proxy only object `GET` and `HEAD` requests under `/xem-files/` to MinIO; preserve the full incoming Host (including any port), path and query string so signatures remain valid. Do not expose bucket listing, writes or the console. Strip app cookies and Authorization headers and return `Cache-Control: private, no-store` on this object route. MinIO validates the presigned query; no public bucket policy is needed.

`S3_CREATE_BUCKET=true` allows startup to create a missing dedicated bucket. Initialization is limited to 30 seconds; access-denied responses remain errors. An existing bucket is checked without changing its policy. `S3_DISABLE_ACL=true` omits object ACL headers for MinIO and ACL-disabled buckets; the default preserves existing provider/caller ACL behavior.

## Existing object storage

Use the provider's bucket, region, endpoint and secret credentials. Leave `S3_CREATE_BUCKET=false` for a bucket managed elsewhere. Enable `S3_DISABLE_ACL` when the provider or bucket disables ACLs. Keep private buckets private and use signed reads.

`S3_PUBLIC_ENDPOINT_URL` is optional. Set it only when browser reads use a different HTTPS origin; it requires `S3_ENDPOINT_URL`. With no public override, the existing endpoint is used for both uploads and signed reads and must be reachable by browsers. Existing configurations without these new options retain their endpoint and ACL behavior.
