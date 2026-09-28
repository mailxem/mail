# Choose a host for Xem

Xem's single-node Docker Swarm starter runs the app, Go backend, PostgreSQL, Redis and RustFS on a Linux VPS. The README buttons lead here so you can choose a server that can also reach your email provider.

The setup path is **choose a provider → create a Linux VPS → prepare Docker and HTTPS → run the Xem installer**. You create and pay for the server in your own account. These are setup guides, not published provider marketplace images or automatic provisioning templates.

## Compare providers

| Provider | Sending through this starter | Get started |
| --- | --- | --- |
| **Vultr** | Its published default outbound block list includes SMTP port 25. Use your mail provider's authenticated submission service on 587 or 465 and verify connectivity for your account. | [Vultr setup](#vultr) |
| **Hetzner Cloud** | Port 587 is explicitly allowed. Ports 25 and 465 are blocked by default; unblocking is subject to eligibility and review. Choose SMTP with STARTTLS on 587 when your sending provider supports it. | [Hetzner setup](#hetzner) |
| **DigitalOcean Droplets** | Ports 25, 465 and 587 are blocked by default. The standard SMTP starter cannot send from a Droplet. Resolve delivery compatibility before choosing it. | [DigitalOcean requirements](#digitalocean) |
| **Your existing Linux server** | Check its network policy and access to your SMTP provider. Keep the database, Redis and storage console private. | [Swarm guide](../swarm/README.md) |

Provider policies were checked against the sources below on **September 28, 2026**. Account restrictions can change. A hosting guide does not establish sender approval or successful email delivery.

## Vultr

**[Choose a Vultr cloud server][vultr-signup]**

1. Create a Linux cloud-compute instance in your chosen region; Ubuntu 24.04 LTS is a straightforward choice. Add your SSH public key during server creation.
2. Choose at least **2 CPU cores, 2 GB RAM and 20 GB free disk** for the basic runtime target. Source builds can need additional RAM or swap; provider plan sizes may exceed the minimum.
3. Use your email provider's SMTP submission endpoint on 587 with STARTTLS, or 465 with TLS if supported and reachable. Vultr blocks port 25 by default; an account-specific exception requires provider approval.
4. Follow [prepare and install](#prepare-and-install). Configure your sender's domain authentication, then validate a send to an address you control before scheduling real mail.

Source: [Vultr's blocked-port policy](https://docs.vultr.com/what-ports-are-blocked).

## Hetzner

**[Choose a Hetzner Cloud server][hetzner-signup]**

1. Create an Ubuntu 24.04 LTS cloud server in your chosen region and add your SSH public key.
2. Choose at least **2 CPU cores, 2 GB RAM and 20 GB free disk**, with extra memory or swap for source builds. Use persistent server storage for the Docker volumes.
3. Configure a sending provider that supports **SMTP port 587 with STARTTLS**. Hetzner states that 587 is not blocked. Ports 25 and 465 are blocked by default; do not assume a support request will be approved immediately.
4. Follow [prepare and install](#prepare-and-install), configure your sender and verify delivery to an address you control.

If a particular mailbox integration fixes its SMTP connection to port 465, it needs that port unblocked or a compatible host. Changing a different sender's SMTP settings does not change the integration's connection.

Sources: [Hetzner Cloud server FAQ](https://docs.hetzner.com/cloud/servers/faq/) and [general FAQ](https://docs.hetzner.com/cloud/general/faq/). Hetzner's former referral-credit program is discontinued; this guide does not offer a referral credit.

## DigitalOcean

**[Explore DigitalOcean Droplets][digitalocean-signup]**

**Check the sending limitation first:** DigitalOcean blocks outbound ports **25, 465 and 587** on Droplets by default, including traffic through Reserved IPs. A Droplet can host Xem's app and data services, but the standard SMTP configuration cannot send email from it. Opening a cloud firewall rule does not remove a provider-level SMTP block.

For an installation that needs SMTP sending now, use a host that permits your sender's submission port. HTTPS-based email sending is a different integration and is not configured by this starter. Do not buy a Droplet expecting the install command to make Gmail SMTP or another blocked SMTP service work.

If you have separately validated an email-delivery configuration compatible with your DigitalOcean account:

1. Create an Ubuntu 24.04 LTS Droplet, add your SSH public key and select at least **2 CPU cores, 2 GB RAM and 20 GB free disk**, with additional memory or swap for source builds.
2. Follow [prepare and install](#prepare-and-install). Plan persistent storage and backups for all data services.
3. Validate the separately configured delivery path before inviting users or starting campaigns. App readiness alone is not email readiness.

Source: [DigitalOcean: Why is SMTP blocked?](https://docs.digitalocean.com/support/why-is-smtp-blocked/).

## Prepare and install

On the VPS you selected:

1. Install [Docker Engine with the build plugin](https://docs.docker.com/engine/install/ubuntu/), Git, Python 3, OpenSSL and curl. Confirm your SSH user can run Docker. The installer expects these tools; it does not install the operating system or Docker for you.
2. Prepare app, API and storage hostnames, such as `app.example.com`, `api.example.com` and `files.example.com`. Configure DNS, a TLS reverse proxy and access rules using the [HTTPS guide](../swarm/README.md#https-and-public-deployment). All three origins must work from browsers and containers.
3. Keep SSH restricted to trusted administrator addresses. Allow public HTTPS through your proxy and restrict direct access to Docker-published app/API/storage ports. Do not expose PostgreSQL, Redis, the storage console or single-node Swarm control ports publicly. Verify firewall behavior for Docker traffic.
4. Install Xem:

```bash
curl -fsSL https://raw.githubusercontent.com/mailxem/mail/undefined/scripts/install.sh | bash
```

The wizard asks for the three origins, generates credentials and starts the services. Keep its protected state directory: it contains credentials required for updates and recovery. The [full Swarm guide](../swarm/README.md) covers inspecting/pinning the script, unattended configuration, updates, backup and restore.

After installation, open the printed registration URL, create your workspace and configure an authorized SMTP provider. Complete its domain verification, SPF, DKIM and DMARC requirements. Test using a recipient you control. The installer checks HTTP readiness and does not send a test email.

**Runtime target:** 2 CPU cores, 2 GB RAM and 20 GB free disk for light use. This is not a measured capacity guarantee. Source builds, updates, attachments, volumes and backups need additional capacity. Keep persistent volumes on the selected node; this is a single-node deployment.

## Hosting links and supporting Xem

The provider signup links in this guide are currently standard links, with no affiliate or referral identifier. Any future compensated link will be labeled next to the signup link. Hosting remains your choice; using another provider does not limit the open-source software.

Maintainers: [partner-program options and link activation](partners.md) explains how to add approved links without implying a nonexistent partnership, credit or one-click template.

<!-- Affiliate URLs belong in these three definitions. Label the corresponding
signup link above before replacing a standard URL with an approved referral URL.
Do not put affiliate dashboard credentials or unpublished access tokens here. -->

[vultr-signup]: https://www.vultr.com/
[hetzner-signup]: https://www.hetzner.com/cloud/
[digitalocean-signup]: https://www.digitalocean.com/products/droplets
