# Kubernetes deployment

This is an alternative runtime, not an in-place upgrade of Dokploy. The chart deploys Xem onto an **existing cluster**. Test on isolated infrastructure before considering production cutover.

## Prerequisites

- Supported Kubernetes >=1.30; choose a currently maintained release. Manifests are schema-tested against 1.34.
- An ingress controller such as Traefik, with the correct namespace selector in `networkPolicy.ingressNamespaceSelector`.
- A CNI enforcing NetworkPolicy, and a LoadBalancer implementation supporting direct TCP. On AWS use AWS Load Balancer Controller with NLB IP targets and source-IP preservation.
- cert-manager >=1.18 with its CRDs for the optional Certificate/Issuer, or a pre-provisioned trusted TLS Secret.
- An existing IAM OIDC provider for the cluster; exact service-account trust, no application permission on the node role. Block pod access to node IMDS at cluster level. No static AWS keys in application Secrets or Infisical.
- External PostgreSQL and Redis with authenticated, encrypted connections, tested backups and appropriate security groups/network routes. This chart creates no PVCs or data services.
- Application images pinned to tested digests, matching database migration expectations. Optional payments/MCP images need their own runtime settings and smoke checks.

## Create AWS workload resources

Use `terraform/environments/kubernetes` with its own encrypted state key. Copy `cluster.tfvars.example` to a local file and replace the account, cluster OIDC ARN/issuer and namespace. The issuer URL must exactly correspond to the provider ARN. The role's `sub` is `system:serviceaccount:NAMESPACE:SERVICE_ACCOUNT` and `aud` is `sts.amazonaws.com`.

Use a distinct stack name from Dokploy. Review the Terraform plan before creating the role and stack. Output `runtime_role_arn` goes into Helm values; `topic_arn` goes into the backend's **separate** Infisical environment. Leave the feedback endpoint blank until the backend knows that exact topic ARN and is reachable over HTTPS, then configure and verify the subscription in a separate change.

This does not configure OIDC on a cluster, install a load-balancer controller, or grant that controller its own AWS permissions. Those remain cluster administration prerequisites.

## Prepare namespace and existing Secrets

Create a namespace such as `xem`. Provision Secrets using your existing secret manager/operator or a local protected file, not `--set` arguments with secrets. Secret values are deliberately absent from Helm templates and examples.

| Default Secret | Contents / purpose |
| --- | --- |
| `xem-backend-bootstrap` | `INFISICAL_CLIENT_ID`, `INFISICAL_CLIENT_SECRET`, `INFISICAL_PROJECT_ID`, `INFISICAL_ENV`, and `INFISICAL_API_URL` when needed |
| `xem-frontend-runtime` | Existing frontend authentication/provider settings, including the application's public URL and auth secret |
| `xem-payments-bootstrap` | Optional payments Infisical bootstrap and runtime settings required by its image |
| `xem-mcp-runtime` | Optional MCP runtime settings; set `XEM_API_BASE_URL`, `XEM_MCP_PUBLIC_URL`, and allowed origins for the target environment |
| `cloudflare-dns-token` | `api-token`, scoped to Zone Read and DNS Edit for the one certificate zone, in the Issuer namespace |
| `xem-smtp-tls` | `kubernetes.io/tls` Secret with `tls.crt` (full chain) and `tls.key`; cert-manager creates this when enabled |
| `xem-web-tls` | Trusted HTTPS certificate covering the enabled HTTP ingress hostnames; provision separately or use your ingress/cert-manager annotation workflow |

Secret names are configurable. Enable etcd encryption and least-privilege namespace RBAC. Do not export Secrets to the repository or CI logs. Infisical bootstrap credentials are environment-specific and read-only. The chart does not run an Infisical Operator; the backend itself loads settings at startup.

Set these non-secret backend values in the target Infisical environment to match chart infrastructure:

```dotenv
SERVER_HOST=0.0.0.0
SERVER_PORT=9001
MANAGED_SENDING_ENABLED=true
MANAGED_SMTP_ENABLED=true
MANAGED_SMTP_ADDR=:2525
MANAGED_SMTP_HOST=smtp.YOUR_DOMAIN
MANAGED_SMTP_TLS_CERT=/etc/xem/smtp/fullchain.pem
MANAGED_SMTP_TLS_KEY=/etc/xem/smtp/privkey.pem
MANAGED_SES_REGION=us-east-2
MANAGED_AWS_ACCOUNT_ID=<target account>
MANAGED_SNS_TOPIC_ARN=<this environment's stack output>
```

Set backend public API/CORS/auth URLs to the target environment as required by the application. Infisical can override process environment values at startup, so mismatched ports or flags there cannot be fixed by Helm alone. Restart the deployment after Infisical changes. If managed sending is disabled in Helm, disable both managed application flags in Infisical too.

Do not place `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`, or conflicting workload identity variables in Secrets/Infisical. The chart projects a short-lived token with STS audience, sets the exact role ARN/token path, disables SDK EC2 metadata fallback, and disables automatic Kubernetes API token mounting. No Kubernetes API RoleBinding is required for the application.

## Configure chart values and images

Start from `charts/xem/ci/managed.yaml` as a **shape example**, copying it to an ignored `values.local.yaml`. Replace every example hostname, role ARN and certificate account setting. Disable optional payments/MCP components if unused. Update the ingress namespace selector if your controller is outside namespace `traefik`.

For production, set each `components.NAME.image` to `repository@sha256:...`. The frontend's `NEXT_PUBLIC_API_URL` and `NEXT_PUBLIC_PAYWALL_URL` are **build-time** Docker arguments; build the image for the public target endpoints (`https://API_HOST/api/v1`, and the appropriate payments API URL). Setting them only at pod runtime will not change browser code. The chart sets frontend server-side `INTERNAL_API_URL` to the backend Service on port 9001 with `/api/v1`.

HTTP ingress supports one hostname per enabled component. It requires an HTTPS Secret, but the ingress controller is responsible for HTTP-to-HTTPS redirect behavior. Configure and verify that redirect before exposing authenticated traffic. The SMTP certificate is independent of web HTTPS certificates.

The AWS SMTP example uses:

- `loadBalancerClass: service.k8s.aws/nlb`
- internet-facing NLB, TCP listener, IP targets
- `preserve_client_ip.enabled=true` and `externalTrafficPolicy: Local`
- published 587 to the backend's named `smtp` port 2525

Do **not** enable load-balancer TLS termination, PROXY protocol, or route SMTP through HTTP ingress. The backend expects a normal SMTP session upgraded with STARTTLS. Check controller support, subnet discovery, security groups and observed client IPs in the target cluster; a rendered Service is not proof that a cloud load balancer is configured correctly. Limit `loadBalancerSourceRanges` for private pilots where suitable; empty allows the provider's default public access.

## Render, install, verify

After prerequisites, reviewed AWS changes, Secrets, images and values are ready:

```bash
helm lint charts/xem -f values.local.yaml
helm template xem charts/xem -n xem -f values.local.yaml > rendered.local.yaml
# Inspect rendered resources without printing Secret contents.
helm upgrade --install xem charts/xem -n xem -f values.local.yaml --wait --timeout 10m
kubectl -n xem get deployments,pods,services,ingress,certificates
kubectl -n xem describe certificate xem-xem-smtp
kubectl -n xem rollout status deployment/xem-xem-backend
```

The backend can remain Pending until cert-manager has issued the TLS Secret; DNS-01 needs outbound DNS/HTTPS and the scoped Cloudflare token. For a pre-existing issuer, set `certificate.createIssuer=false` and provide its name/kind. For a pre-existing trusted TLS Secret, set `certificate.enabled=false`. Use ACME staging only in testing; its certificates will correctly fail the public-trust smoke test.

Once the NLB is ready and the backend is healthy, create/update the SMTP DNS-only CNAME through the domain Terraform root. Use a separate test hostname first. Verify external HTTPS health, the frontend login/onboarding flow, and then run:

```bash
python3 scripts/smtp-smoke.py smtp.YOUR_DOMAIN
```

Finish controlled authenticated delivery and SNS feedback checks before cutover; see [production status](production-status.md). Do not run both old and new backend workers on the same database during migration.

## Operations and limits

The backend uses **one replica and Recreate updates**. In managed mode the chart refuses more than one replica; no HPA or misleading availability budget is generated. The managed readiness probe checks that SMTP is listening; HTTP startup/liveness probes cover the API. These are not deep database or SES health checks. An upgrade pauses service while the pod restarts and migrations run. Keep client retries and a maintenance/rollback plan.

TLS is mounted as a Secret directory without `subPath`, permitting certificate renewal to propagate. The backend reloads the pair during new TLS handshakes. Pods run as non-root with all capabilities dropped, a read-only root filesystem, bounded temporary/cache storage, resource limits and no default service-account token. Check optional images in a real cluster before use; arbitrary images may require different filesystem behavior.

NetworkPolicy restricts HTTP ingress to this release's pods and the selected ingress-controller namespace, and permits SMTP ingress to the backend. **Egress is not restricted** by this chart: Infisical, SES, STS, SNS signature certificates, DNS and external databases need environment-specific rules. Add those rules using your CNI/firewall capabilities. NetworkPolicy has no effect without an enforcing CNI.

Monitor SMTP certificate expiry, submission failures, queue age, SES quota/throttling, bounce/complaint rates, SNS confirmation/processing, database health and backup restore tests. This chart does not install a monitoring stack or deliver alerts. To roll back, review database compatibility before `helm rollback`; Helm cannot undo data migrations or external SES side effects. Do not use an automatic rollback as a substitute for that review.
