# Production snapshot — 2026-09-12

This is an operational snapshot, not a health guarantee. The Terraform/Helm package has not been applied or deployed. Live changes below were made earlier during the managed-SMTP setup.

## Existing resources

| Item | Value |
| --- | --- |
| AWS account / region | `831009601947` / `us-east-2` |
| EC2 | `i-00aea05ab70b00750`, `xem-server`, `t3.medium`, Debian 13 |
| Elastic IP | `3.151.225.168`, allocation `eipalloc-05ef4e0c998579542` |
| VPC / subnet | `vpc-07691bd7e3877f067` / `subnet-03995a2e70364d56f` |
| Existing administrative SG | `sg-07f60a270312ebf96` (preserved, referenced only) |
| SMTP SG | `sg-07658f0e546336d23`, public TCP 587 |
| Root disk | 128 GiB gp3, 3000 IOPS / 125 MB/s, **unencrypted** |
| Runtime IAM role/profile | `xem-managed-sending-ec2` |
| CloudFormation stack | `xem-managed-sending` |
| Feedback topic | `arn:aws:sns:us-east-2:831009601947:xem-managed-sending-FeedbackTopic-fctHOxASbCWx` |
| SES platform identity | `xem.email`, verified, DKIM 2048, `bounce.xem.email`, reject on MAIL FROM MX failure |
| Authoritative DNS | Cloudflare; SMTP DNS-only; existing root iCloud mail records preserved |
| Dokploy / Traefik | `v0.29.13` / `v3.6.7` |
| Backend service | `xem-backend-j0iqkf`, one replica, stop-first |
| Published SMTP | TCP 587 host mode → container 2525 |
| Certificate mount | host `/etc/xem/smtp-readonly` → container `/etc/xem/smtp` |

SES production access is enabled; the account reported 50,000 recipients/day and 14/second. Application pilot limits remain separate (200/day at setup). Account suppression covers bounces and complaints. No account-level SES settings are managed by this Terraform package.

## Verified

- DNS resolves `smtp.xem.email` to the Elastic IP.
- The persistent Dokploy port and mount survive redeployment.
- The backend can read the certificate as its non-root user; writes to the host mirror fail read-only, including from a root container.
- External `smtp-smoke.py` passed again during this package's validation: public certificate trust and hostname, TLS 1.3, no advertised AUTH before TLS, rejection of pre-TLS AUTH and unauthenticated MAIL FROM.
- The Let's Encrypt certificate expires **2026-12-10 19:47:02 UTC**. Renewal export runs every five minutes; this timestamp is not a substitute for monitoring.
- The API health endpoint returned HTTP 200 during initial setup.

No message was sent by the transport test.

## Required before claiming full delivery readiness

1. Approve a test workspace through the supported sending-admin workflow, verify its custom domain and sender, and create a revocable SMTP credential through Xem. Do not reuse AWS/SES credentials as Xem SMTP credentials.
2. Send one authorized test message to a controlled mailbox, securely supplying credentials without logging them. Confirm queue acceptance, SES message acceptance, actual receipt and SPF/DKIM/DMARC results. Check the runtime role's real access and tenant/configuration-set behavior.
3. Configure the CloudFormation `FeedbackEndpoint` after the backend knows the exact topic ARN. It is currently blank. Confirm the SNS subscription and verify signed delivery, bounce and complaint events using the SES mailbox simulator and a test workspace. Confirm suppression and repeat-recipient rejection.
4. Verify revocation, domain isolation, pilot limits and retry behavior with that test workspace. Review quota and abuse controls before widening access.
5. Test the onboarding checklist and domain verification UI against these real results. Revoke the test credential after testing.

IAM Access Analyzer found no policy findings, but an earlier policy-simulator run reported unexpected implicit denials for some SES actions. The actual authenticated SES send/provision path remains unverified. The IaC deliberately preserves the existing scoped policy; do not broaden it to `ses:*`/`Resource:*` based on speculation. Diagnose any real denial against AWS's action/resource support and retest.

Some earlier setup screens exposed bootstrap/Redis credentials; credential rotation was recommended but is not recorded as completed. Rotate through their owning systems, redeploy dependants and verify access. Never copy those values into this repository.

## Package validation

Terraform 1.16.2 with AWS provider 6.64.0 / Cloudflare 5.25.0: all three roots validate, and four local plan-only IAM trust tests pass. Helm 3.18.6: default and managed/all-components examples lint and render. Ten Python tests cover adoption guard behavior, certificate rotation/key mismatch/hostname checks, and chart safeguards. Kubernetes 1.34 schemas validate 19 built-in resource instances across both examples; the two cert-manager resources validate separately against upstream v1.18.2 CRD schemas. No Terraform live plan/apply, Helm cluster installation, or production migration was performed.
