# Dokploy SMTP and certificate renewal

This codifies the setup verified on 2026-09-12. The installer is intended for the Debian/systemd Dokploy host and fixed hostname `smtp.xem.email`; it does not install Dokploy. Review and adapt paths/hostname/UID for other hosts before running it.

## Prerequisites

- DNS-only `smtp.xem.email` points at the host's Elastic IP.
- EC2 security groups allow public TCP 587 plus HTTP 80 for the existing Traefik HTTP-01 resolver. Never expose container port 2525 publicly in addition to 587.
- Traefik's dynamic file provider watches `/etc/dokploy/traefik/dynamic`, and the `letsencrypt` resolver stores ACME state in that directory's `acme.json`.
- The backend image runs as UID/GID 65532. The host has Python 3, OpenSSL, CA roots, systemd and util-linux.

The router in `host/xem-smtp-certificate.yml` requests a **public** certificate for SMTP via the existing Traefik ACME resolver. It routes HTTPS to `noop@internal`; an HTTPS 418 is expected and says nothing about SMTP health. It does not terminate STARTTLS. Cloudflare Origin CA is unsuitable for normal SMTP clients.

## Install/reconcile the host exporter

Copy this repository to the host, review the scripts, then run as root:

```bash
bash scripts/install-dokploy-smtp-tls.sh
```

The installer writes the certificate-only router, exporter, service, timer and read-only mirror mount. Traefik may need time to issue the certificate on the first run; a failed first export leaves the timer retrying. Inspect the service journal and rerun the installer after issuance. Do not disable certificate verification to work around an issuance error.

The exporter reads only the dedicated SMTP certificate from Traefik storage, validates the hostname, public trust chain, remaining lifetime and private-key match, and installs a versioned directory. Keys are root-owned, group 65532, mode 0640; directories including `/etc/xem` are 0750. An atomic `current` symlink selects the valid pair. A bad renewal leaves the previous certificate untouched and fails the service. Keep ACME storage private; never mount the full ACME JSON into the app. Monitor exporter failures and certificate expiration externally, as this script does not deliver alerts.

The timer runs every five minutes. The backend reloads the certificate from disk at each TLS handshake, so renewal does not require a backend restart. Concurrent handshakes crossing a rotation can retry; do not copy certificate/key files independently into the container. Old version directories are retained for recovery; clean obsolete ones deliberately after validating a renewal.

Dokploy v0.29.13 cannot persist a read-only flag for this bind mount. The installer creates `/etc/xem/smtp-readonly` as a **kernel read-only bind mirror** of `/etc/xem/smtp`; this keeps the source immutable inside redeployed containers even if Dokploy omits the Docker ReadOnly field. Check the effective mount after every host or Dokploy upgrade.

## Persistent Dokploy settings

In the backend application (production ID `ExMUJXUJHPyGKOB_cxn5m`):

| Setting | Value |
| --- | --- |
| Host bind path | `/etc/xem/smtp-readonly` |
| Container mount path | `/etc/xem/smtp` |
| Published port | `587`, TCP, **host** publish mode |
| Container target port | `2525` |
| Replica count | `1` |
| Swarm update order | `stop-first` |

Save the mount and published port in **Dokploy's application settings**, then redeploy. A one-off `docker service update` alone is not durable because Dokploy recreates the service spec. Do not publish SMTP through an HTTP router. Stop-first prevents the old task from occupying host port 587 while the new task starts, and introduces a brief interruption.

## Infisical settings

Keep these in the backend environment in Infisical, alongside existing DB/Redis/application settings:

```dotenv
MANAGED_SENDING_ENABLED=true
MANAGED_SMTP_ENABLED=true
MANAGED_SMTP_ADDR=:2525
MANAGED_SMTP_HOST=smtp.xem.email
MANAGED_SMTP_TLS_CERT=/etc/xem/smtp/fullchain.pem
MANAGED_SMTP_TLS_KEY=/etc/xem/smtp/privkey.pem
MANAGED_SES_REGION=us-east-2
MANAGED_AWS_ACCOUNT_ID=831009601947
MANAGED_SNS_TOPIC_ARN=<exact CloudFormation TopicARN output>
```

Use the EC2 runtime instance profile; do not add AWS access keys to Infisical. IMDSv2 is required and the current Docker host uses hop limit 2. The EC2 role is shared at host level; unrelated containers must not be treated as isolated IAM workloads. The Kubernetes path instead uses a dedicated projected workload token.

Infisical is loaded at application startup. A setting change needs a backend redeploy. The bootstrap Secret/machine identity must have read access to the correct project and environment; it does not need permission to edit production settings.

## Verify

On the host:

```bash
systemctl status xem-smtp-certificate.timer
journalctl -u xem-smtp-certificate.service --since today
findmnt /etc/xem/smtp-readonly
openssl x509 -in /etc/xem/smtp/fullchain.pem -noout -subject -issuer -dates
```

From an external machine:

```bash
python3 scripts/smtp-smoke.py smtp.xem.email
```

This verifies public trust and hostname, STARTTLS enforcement, and rejection of unauthenticated submission. It never logs in or sends mail. Follow [production status](production-status.md) for the remaining end-to-end checks. To roll back an application release, retain the persistent mount/port configuration and use a known-good image compatible with the database. Never roll back to plaintext SMTP.
