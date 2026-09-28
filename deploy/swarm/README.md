# Self-host Xem with Docker Swarm

This starter builds the monorepo's frontend and backend at the same commit and deploys them alongside PostgreSQL 16, Redis 7, and RustFS. It creates one replica per long-running service on the selected manager and keeps data in named Docker volumes. A short-lived `storage-init` service creates the private `xem` bucket and exits successfully.

## Requirements

- A Linux host running Docker Engine with the build plugin, Git, Python 3, and OpenSSL. The invoking user needs Docker access and write access to the state directory. Docker installation and system firewall changes are left to the operator.
- Start with 4 CPU cores, 8 GB RAM, and 20 GB free disk for builds and data; usage will determine your actual capacity needs.
- An IP or hostname reachable from **both your browser and containers**. Do not use `localhost`: the frontend's authentication also calls the public API URL from inside its container, and signed storage URLs must work in the browser.
- Available app/API/storage ports, defaulting to 3000/9001/9000. PostgreSQL, Redis, and the storage console are not published.
- Outbound HTTPS for source/image downloads. Sending email also requires connectivity to your SMTP provider.

The installer can initialize a new Swarm or use an existing manager. It refuses a worker, an unrelated stack already named `xem`, or saved state from another node. It does not leave an existing Swarm, remove stacks, or move persistent data to another node. Multi-interface hosts may need an explicit `docker swarm init --advertise-addr <manager-IP>` before installation.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/mailxem/mail/undefined/scripts/install.sh | bash
```

Follow the address prompts. For example, on a trusted LAN, enter `http://192.168.1.10:3000`, `http://192.168.1.10:9001`, and `http://192.168.1.10:9000`, replacing the IP with your server's address.

Before merge, preview the branch with:

```bash
curl -fsSL https://raw.githubusercontent.com/mailxem/mail/codex/swarm-bootstrap/scripts/install.sh \
  | XEM_REF=origin/codex/swarm-bootstrap bash
```

For an unattended install, copy [`config.example.json`](config.example.json), replace its documentation-only IP, then run:

```bash
bash install-xem.sh --config ./config.json --state "$HOME/.local/share/xem"
```

The JSON file is data, never shell code. The installer generates database/auth/storage credentials and a base64-encoded RSA key. It stores them in the state directory with owner-only permissions. Repeated runs reuse these credentials and existing volumes; **do not delete this directory to update**. Docker administrators can inspect service environment variables; this initial version does not use Swarm secrets.

The installer only prints success after the API health route and app login page return HTTP 200. This confirms HTTP readiness, not SMTP deliverability or a complete signup flow. PostgreSQL migrations run automatically in the backend, and Swarm restarts the API while its dependencies initialize.

## First use

1. Open `<app_url>/auth/register` and create an account. There is no shared default admin password.
2. Log in and configure your sending provider with your own SMTP credentials.
3. Follow your provider's domain verification, SPF, DKIM, and DMARC instructions before sending.
4. Add an audience and template, then test with an address you control.

The bootstrap does not send test emails. Managed sending and managed SMTP ingress are disabled. Google login, hosted billing/payments, MCP, and the AI assistant are not configured. Some UI entry points may remain visible but need separate integrations.

## HTTPS and public deployment

The default HTTP endpoints are suitable for a trusted local network. For a public server, put all three endpoints behind a TLS reverse proxy and use origins such as `https://app.example.com`, `https://api.example.com`, and `https://files.example.com` in the config. Configure DNS, certificates, forwarded headers, and routing in your proxy before running the readiness check. Storage requests must preserve their original host/path/query for S3 signatures.

Restrict direct access to the published ports at the host/network boundary so external traffic goes through the proxy. These are Docker-published ports: verify your firewall rules apply to Docker traffic. The installer does not provision certificates, DNS, a proxy, or firewall rules. Allow Swarm control/overlay ports only between trusted cluster nodes if you use an existing cluster.

The object store uses a private bucket. Back up the storage credential and RSA key; changing credentials can invalidate existing access, and losing the RSA key can make previously encrypted data unreadable.

## Operate

```bash
docker stack services xem
docker stack ps xem --no-trunc
docker service logs --tail 100 xem_api
docker service logs --tail 100 xem_app
docker service logs --tail 100 xem_storage-init
```

`xem_storage-init` should complete with exit code 0; its `0/1` service count after completion is expected. All five long-running services should have one running replica. The installer pins placement to the initial node because local volumes do not follow tasks across nodes. Images are built on that node and deployed with `--resolve-image never`; this is not a multi-node image distribution workflow.

If readiness times out, inspect task errors and logs. Common causes are unavailable ports, wrong DNS, failed builds, insufficient memory, storage initialization failures, or containers being unable to reach the public API/storage origins. Fix the cause and rerun with the same state directory. No failed-install cleanup removes your data.

If a process is killed abruptly, its `.install-lock` directory may remain. After confirming no installer is running, remove only that empty directory with `rmdir "$HOME/.local/share/xem/.install-lock"` and rerun.

## Update

Back up before an update: the API applies database migrations at startup, so reverting images alone is not a database rollback.

Run the curl command again with the same user and state directory. It fetches a fresh parent revision, checks out the app and backend together, rebuilds, and updates the services. This uses stop-first updates and can briefly interrupt service. It does not rotate existing credentials or delete volumes. To change origins or published ports, pass an updated `--config`; the app is rebuilt because its public API URL is a build-time setting.

For reproducibility, set `XEM_REF` to a reviewed monorepo commit SHA and download the script from that same SHA:

```bash
XEM_REV=<reviewed-monorepo-commit-sha>
curl -fsSLo install-xem.sh "https://raw.githubusercontent.com/mailxem/mail/$XEM_REV/scripts/install.sh"
XEM_REF="$XEM_REV" bash install-xem.sh
```

The state directory's `release.json` records the last HTTP-ready monorepo revision and image tag. Keep the prior release information and backups before updating; the installer does not prune old images.

## Backup and restore

Save an encrypted copy of the entire state directory, the `xem_storage` volume, and a PostgreSQL logical dump. Redis contains queued jobs; take a consistent Redis backup as part of a planned maintenance window. Keep SMTP traffic stopped during restores to avoid replaying queued sends.

On the pinned manager, create a logical database backup:

```bash
umask 077
DB_CONTAINER=$(docker ps -q --filter label=com.docker.swarm.service.name=xem_postgres)
test -n "$DB_CONTAINER"
docker exec "$DB_CONTAINER" pg_dump -U xem -d xem -Fc > xem-postgres.dump
```

For a coordinated backup, pause app/API traffic and stop the workers before snapshotting database, object storage, and Redis volumes. Follow each service's backup guidance; copying a live PostgreSQL data directory is not a substitute for a consistent backup.

Restores are operator-managed. Restore the named volumes and credentials together on the target node before deploying. The installer refuses a saved `node_id` that differs from the current Swarm node to prevent silently starting with empty local volumes. After verifying the restored volume names and contents, deliberately update `node_id` in the saved config to the target's `docker info --format '{{.Swarm.NodeID}}'`. Restore a matching code revision and database backup when rolling back migrations.

## Stop or remove

```bash
docker stack rm xem
```

This removes Xem's services and network; it retains named data volumes and the state directory. It does not leave the Swarm. Do not remove `xem_postgres`, `xem_redis`, or `xem_storage` unless you intend to erase their data. Rerun the installer with the saved state to start again.

## Validation

Run the installer contract checks without a Docker daemon:

```bash
bash -n scripts/install.sh
shellcheck scripts/install.sh
python3 -m unittest discover -s deploy/swarm -p 'test_*.py'
```

A real deployment also needs image builds, dependency startup, signup/login, uploads, and an authorized SMTP test. The CI smoke workflow exercises the generated stack on a disposable Linux runner; it never sends email.
