# Deployment guide

The current production runtime is EC2 + Dokploy. Adding this repository does not move it to Kubernetes or change cloud resources.

## Current deployment

1. Read [production status](docs/production-status.md), confirm resource IDs and take database/configuration backups.
2. Follow [Terraform adoption](docs/terraform-adoption.md) to import existing resources into separate, encrypted remote states. Resolve the complete plan before applying anything.
3. Follow [Dokploy SMTP TLS](docs/dokploy.md) to understand or reproduce certificate export, the durable read-only mount, and TCP port mapping.
4. Keep SMTP settings in Infisical. Redeploy the backend after changing those settings; the application reads them at startup.
5. Run the transport smoke test, then finish the controlled delivery and feedback checks in the status guide.

## Future Kubernetes deployment

Use [the Kubernetes runbook](docs/kubernetes.md). Test with a separate database, Redis, stack, service account and hostnames before planning a cutover. Application migrations run at startup; back up the database and review backend migration compatibility before an image upgrade or rollback.

Managed sending currently uses one backend replica and Recreate/stop-first updates. Deployments cause a short interruption. SMTP clients need retries; do not promise zero downtime. Enabling two active runtimes against the same queue is not an availability strategy.

## State and ownership

Terraform owns AWS infrastructure and platform DNS. CloudFormation owns the sending stack's SNS topic, subscription and IAM policy. Dokploy or Helm owns application runtime configuration. Infisical owns application settings. Traefik/ACME or cert-manager owns certificates. Xem owns customer/workspace identities and sending records.

The old Kops instructions are archived in `legacy/`. Do not run them against production.
