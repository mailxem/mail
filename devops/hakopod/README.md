# Xem on Hakopod

`xem.toml` uses the updated schema v1 and the four requested image references. It targets **self-hosted Hakopod** with public SMTP on TCP 587, a backend certificate mount, and an operator-approved AWS workload identity. Managed-cloud Hakopod rejects public TCP. This configuration includes verified STARTTLS readiness, automatic renewal from the backend HTTP ingress certificate, and explicit stop-first updates. It is a migration target requiring operator setup, not an already provisioned deployment. Saving it changes no infrastructure or DNS.

## Prepare the application

1. Confirm self-hosted mode and stage a private bootstrap revision as described below. Create the `xem` application in the chosen Hakopod project/environment, then verify the five HTTP hostnames (including `smtp.xem.email` for certificate ownership) in Custom domains and restore their mappings. Do not move production DNS before testing the destination.
2. Confirm that `hakopod-acme` is an available cert-manager issuer, or change each service's `tls` reference to a configured issuer/managed certificate. Domain ownership and certificate coverage must both pass. The TOML does not create an issuer.
3. Create the application-scoped secrets referenced below. Keep the backend and payments settings in separate, dedicated Infisical environments. The frontend does not load Infisical itself; supply its explicit secret references through Hakopod.
4. Check database/Redis connectivity from the platform. They remain external; general egress excludes private/management CIDRs and metadata endpoints. Existing EC2-private endpoints are not automatically reachable.
5. Verify the image build configuration, OAuth callbacks and payment webhook URLs for the target domains. Public frontend API/payment URLs are baked in at image build time; TOML environment values cannot rewrite the existing browser bundle. The payments hostname/API path here uses `payments.xem.email/api/v1`; adjust it to match the intended image and verified domain if different.

Required Hakopod secrets:

| Reference | Value supplied through the secret manager |
| --- | --- |
| `infisical-api-url` | The existing Infisical HTTPS API origin |
| `backend-infisical-client-id`, `backend-infisical-client-secret` | Read-only backend machine identity credentials |
| `backend-infisical-project-id`, `backend-infisical-environment` | Backend project ID and dedicated environment slug |
| `payments-infisical-client-id`, `payments-infisical-client-secret` | Read-only payments machine identity credentials |
| `payments-infisical-project-id`, `payments-infisical-environment` | Payments project ID and dedicated environment slug |
| `frontend-auth-secret` | Frontend session secret; used by both Auth.js naming conventions |
| `frontend-google-client-id`, `frontend-google-client-secret` | Google OAuth application credentials |

Backend Infisical must provide its existing PostgreSQL, Redis, JWT, storage and integration settings. Payments Infisical must provide its database, strong JWT/admin secrets and payment-provider settings. Infisical loads into process environment at startup, so its ports, URLs, flags and storage paths must agree with this file. Use a dedicated environment and test data; do not run two backend installations against the production queue.

For the final SMTP revision, mirror the managed-sending settings from `xem.toml` in that Infisical environment, including `:2525`, `smtp.xem.email` and the new `/app/certificates/smtp/tls.crt` / `tls.key` paths. Supply `MANAGED_AWS_ACCOUNT_ID` and the exact `MANAGED_SNS_TOPIC_ARN` belonging to the chosen account, region and environment. Configure and verify its SNS feedback endpoint separately. The file uses SES region `us-east-2`; the approved AWS binding must agree. Keep managed sending and SMTP disabled during the initial private bootstrap.

Do not set static AWS credentials, `AWS_PROFILE`, `AWS_ROLE_ARN`, `AWS_REGION` or `AWS_WEB_IDENTITY_TOKEN_FILE` in TOML/Secrets/Infisical. Hakopod injects its approved identity settings. Its spec validator rejects conflicting TOML bindings, but cannot inspect settings that the application later imports from Infisical.

If an image is private, create a scoped managed registry credential and add `registry_credential = "NAME"` to that service. No registry password belongs in the file. Hakopod resolves the supplied mutable image tags to digests during deployment; image pull/architecture availability is not established by local validation.

## Compose translation

- MCP keeps a read-only root filesystem, non-root UID/GID 1000, a 16 MiB memory-backed `/tmp`, `/health` readiness and a 40-second shutdown grace period. Its literal URLs replace Compose interpolation; Hakopod does not evaluate `${...}` expressions. An empty allowed-origin list retains the MCP server's same-origin default; add explicit origins if needed.
- Hakopod drops capabilities and manages pod restarts. Its v1 service schema has no fields for Compose `restart`, `cap_drop`, `security_opt`, `pids_limit`, Docker log rotation or tmpfs `noexec,nosuid` options. These were not silently translated into invented settings; exact parity requires operator/runtime controls outside this TOML.
- No profile matches 1 CPU plus 256 MiB. MCP uses `small` (0.5 CPU limit, 256 MiB); select `medium` for 1 CPU and 512 MiB. Temporary mounts remain below the 128 MiB per-service cap.
- API, frontend, payments and MCP each have a public HTTP route. Shared-network peer access is restricted separately; `from = []` does not disable public ingress. External egress stays available for Infisical and integrations. Startup dependencies sequence readiness but do not replace application retries.
- The backend has a 5 GiB persistent `/app/storage` directory for local storage. It explicitly sets `replicas = 1` and `update_strategy = "recreate"`, retaining stop-first behavior even if local storage is later removed. Other services use the platform's rolling updates and can temporarily have an extra pod. No HPA is configured.

## Self-hosted SMTP setup

The updated Hakopod implementation supports the required controls:

| Application setting | Xem configuration | Required preparation |
| --- | --- | --- |
| `ports` + `public_tcp` | Public TCP 587 → private SMTP 2525; HTTP stays on 9001 | Self-hosted mode, administrator-provisioned HAProxy listener, source-preserving exposure and firewall rule |
| `certificate_mounts` with `source = "ingress"` | Read-only `/app/certificates/smtp/tls.crt` and `tls.key`, readable by GID 65532; follows renewals | A valid certificate covering `smtp.xem.email` already exists on this backend's owned HTTP TLS ingress |
| `aws_identity` | Requested binding `xem-managed-sending` | Operator binding for the exact project/environment/application `xem`/service `backend`, a registered public OIDC issuer and a narrowly scoped SES role |
| `readiness` | HTTP `/health` plus verified STARTTLS on local port 2525 | Operator-configured, digest-pinned readiness helper image |
| `update_strategy` | `recreate`, one replica | Capacity and client retries for a brief stop-first interruption, including certificate rotations |

The existing `tls` fields still configure **HTTP** ingress. `smtp.xem.email` is now mapped to `backend` in `[domains]` so that this service owns the HTTP TLS certificate used by automatic renewal. This also exposes the backend HTTP route at that hostname. SMTP submission remains a separate raw TCP 587 listener; it is not routed by HTTP hostname. Keep its DNS-only record pointing at the existing server until destination tests pass. No public ports 25 or 465 are declared.

### Bootstrap and provision

1. The destination operator sets `HAKOPOD_DEPLOYMENT_MODE=self-hosted` and provisions `HAKOPOD_PUBLIC_TCP_PORTS=587` plus matching HAProxy container/Service/host or load-balancer exposure. Existing installations need the reviewed Helm changes too. Open only the required external port in the destination cloud/host firewall. A variable alone does not open the listener. For an initial restricted pilot, replace the TOML's explicit `0.0.0.0/0` with approved source CIDRs.
2. Configure the operator's `HAKOPOD_READINESS_PROBE_IMAGE` (or `[server] readiness_probe_image`) with a verified multi-platform digest. Use a Hakopod release that implements the readiness helper, `source = "ingress"`, and `update_strategy` fields. Plans reject helper checks without this operator setting; do not replace the STARTTLS check with a weaker probe to bypass that error. The helper's public root bundle must trust the issuer used for SMTP. A staging ACME certificate will correctly fail public-trust readiness.
3. Bootstrap with a local copy: remove `[domains]`, set every service's `public=false`, omit their HTTP `tls` settings, and remove backend `public_tcp`, `certificate_mounts`, `readiness` and `aws_identity`. Keep both managed flags false in that copy and its dedicated Infisical environment. Keep the singleton/Recreate settings. Deploy privately to create the application namespace with isolated test data.
4. Verify the five HTTP domains, then deploy an **HTTP-only TLS revision**: restore `[domains]`, `public` and `tls` settings while retaining the bootstrap omissions and disabled managed flags. Wait for the backend's owned ingress certificate to cover both `api.xem.email` and `smtp.xem.email`. Automatic certificate sources require that active ingress before they can be enabled. For pre-cutover issuance while production DNS still points at the old server, use an operator-configured DNS-01 issuer or provision a trusted ingress certificate through the supported certificate flow. Do not move production DNS just to satisfy HTTP-01. A separate test hostname can be used first; update all corresponding hostnames together.
5. Register the `xem-managed-sending` operator binding, or change `aws_identity` to its actual approved name. Bind only this backend service. Configure a separate workload role with the required scoped SES permissions and exact OIDC `aud=sts.amazonaws.com` / generated ServiceAccount `sub`. Obtain the actual namespace and generated ServiceAccount from Hakopod's prepared identity status rather than guessing names. The node instance profile is not the application identity. Do not reuse the Helm chart's fixed `xem-backend` ServiceAccount name for Hakopod's generated identity.
6. Restore the final TOML with `source = "ingress"`, STARTTLS readiness, public TCP and approved AWS identity. Match the enabled managed flags and certificate paths in Infisical, review the plan and deploy in the destination test environment. Hakopod protects `/etc`; the old Dokploy host mount is replaced by `/app/certificates/smtp`, not attached to the container.

### Readiness, rotation and cutover

Hakopod combines the existing API `/health` check with a local `smtp_starttls` probe on port 2525. Both must pass. It checks the SMTP greeting, EHLO, STARTTLS, a trusted hostname-matching TLS handshake, post-TLS EHLO and NOOP. It never authenticates or sends mail. This closes the listener-readiness gap without requiring a new endpoint or rebuild of the Xem image. The helper is installed through an init container and runs short probe processes; no shell or persistent sidecar is required in the application image. Readiness failure removes ready endpoints after the configured threshold; it does not add a liveness restart policy or forcibly close existing sessions.

With `source = "ingress"`, Hakopod's reconciliation loop follows the backend's current valid HTTP certificate. When certificate bytes change, it creates a service-owned immutable copy and rolls the service at the same application revision. No manual upload or deployment approval is required after opting in. Maintenance pauses during active deployments. The upstream issuer must still renew the **source** certificate; this feature does not renew a manually uploaded HTTP certificate by itself.

The explicit `update_strategy = "recreate"` keeps one managed-sending runtime during normal deployments and automatic renewal. Rotation includes a brief stop-first interruption, so clients need retries. Keep certificate expiry/delivery alarms: an invalid source leaves the last valid copy in place and reports unhealthy/pending status, and a replacement pod can still fail to become ready. Automatic-source rollback follows the current valid ingress certificate rather than restoring an old certificate snapshot.

For operators who prefer manual pinned certificates, replace `source = "ingress"` with `certificate = "REFERENCE_RETURNED_BY_UPLOAD"` after a backend-owned certificate upload. Never set both fields. Pinned references require renewed uploads and deployments; `--from-ingress` creates a pinned snapshot, not automatic renewal. The default TOML uses automatic renewal and contains no certificate-reference placeholder.

After deploying with the real binding and certificate, inspect `hakopod delivery APP_ID --service backend`. A configured listener or prepared identity is not proof of internet reachability or successful AWS role assumption. Run the existing no-email transport check from outside the destination network once the test hostname resolves correctly:

```bash
python3 scripts/smtp-smoke.py smtp.YOUR_TEST_DOMAIN
```

For a parallel test hostname, update its `[domains]` entry, `MANAGED_SMTP_HOST`, the certificate mount's hostname and `readiness.tls_server_name` together. Then verify authenticated submission to a controlled recipient, receipt, SNS feedback, allowed/denied source behavior, actual SES permissions, automatic certificate rotation/rollback, and connection draining. Keep the current SMTP service and DNS unchanged until those destination checks pass. Local Hakopod tests are not external AWS delivery evidence.

## Validate and review

From the devops repository:

```bash
hakopod validate --file hakopod/xem.toml
hakopod plan --file hakopod/xem.toml --project YOUR_PROJECT --environment YOUR_ENVIRONMENT
```

`validate` checks the local canonical schema. `plan` needs authenticated project/environment context and checks the platform's actual resources. Review that plan before deploying. Local validation does not prove domain ownership, secret existence, issuer availability, image startup or external connectivity.
