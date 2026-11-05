# Release: Xem runs with one docker compose up

Drafts only. Not posted. Every claim below was checked on a test run of this branch
(Docker Desktop on macOS, fresh volumes): all five containers reported healthy, the API
health route and login page returned 200, an account registered and signed in through
the app, and the account still logged in after `docker compose down` and `up`.

Not verified, so not claimed: a Linux server install, sending email, file uploads
through the browser, HTTPS behind a proxy, build time on small VPS hardware.

## X

Xem, our open-source email marketing app, no longer needs Docker Swarm to self-host.

git clone https://github.com/mailxem/mail.git xem && cd xem
bash scripts/compose-env.sh <your server IP>
docker compose up -d

App, Go API, Postgres, Redis and S3-compatible storage on one machine. Bring your own SMTP.

## r/selfhosted

**Title:** Xem (open-source email marketing, GPL-3.0) now self-hosts with plain docker compose, no Swarm

I work on Xem. It's an email marketing app: audiences, campaigns, templates, automations and forms, with a Next.js frontend and a Go backend.

Until now the only self-host path was an install script that set up Docker Swarm and ran a wizard. That's a lot of machinery for one box, so there's now a root `docker-compose.yml`:

```
git clone https://github.com/mailxem/mail.git xem && cd xem
bash scripts/compose-env.sh 192.168.1.10   # your server's LAN IP or hostname
docker compose up -d
```

What you get:

- Five containers: the app, the Go API (which also runs the background workers), PostgreSQL 16, Redis 7 and RustFS for S3-compatible file storage. No cloud storage account needed.
- Healthchecks on every service, `restart: unless-stopped`, and named volumes, so `docker compose down` keeps your data.
- `compose-env.sh` writes a `.env` with generated database, auth and storage passwords and an RSA key. Nothing secret is in the repo. Back that file up.

Things to know before trying it:

- The first `up` builds both images from source, which takes a while and wants some spare RAM for the Next.js build. The frontend bakes its API URL in at build time, so the prebuilt images don't fit an arbitrary host.
- Use an IP or hostname, not `localhost`. The API reaches object storage through the same address your browser does.
- Sending needs your own SMTP provider, plus SPF, DKIM and DMARC on your domain. Nothing is preconfigured for that.
- It's plain HTTP on ports 3000, 9001 and 9000. Put it behind your reverse proxy before exposing it.

The Swarm installer is still there if you want it. Repo: https://github.com/mailxem/mail

Happy to answer questions, and bug reports about the compose setup are welcome.
