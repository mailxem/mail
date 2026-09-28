# SES production access request

Target: `xem` AWS profile, `us-east-2`. Mail type: MARKETING. Website: https://xem.email.

Submitted on 12 September 2026 (Asia/Kolkata) after SES confirmed identity, DKIM, and custom MAIL FROM verification, as authorized by the project owner. AWS granted production access under case `178915636800118`. A subsequent account check confirmed production access and sending enabled, with a quota of 50,000 recipients/day and 14/second. No additional contact addresses were supplied; AWS uses the account contact. The approved provider quota does not change the controlled pilot target described below.

## Submitted use case

Xem is an open-source email campaign and customer communication application at https://xem.email. We are preparing a limited, manually approved pilot of an optional Amazon SES managed-sending integration in us-east-2. The initial use is permission-based newsletters, product updates, and customer communications sent by approved workspaces from domains they own. We also plan to support transactional application mail through authenticated SMTP submission. This is not an unauthenticated relay.

Before enabling a workspace, an operator will review its use case, domain ownership, anticipated volume, and consent practices. The pilot will be restricted to small reviewed audiences, with an initial account-wide operating target of at most 200 recipient sends per day and single-recipient test messages while we validate delivery. We do not intend to enable unrestricted self-service sending. We will not permit purchased, rented, or scraped recipient lists. Each customer domain must pass a DNS ownership challenge, SES identity verification, DKIM, a custom MAIL FROM domain, and DMARC checks before sending is allowed.

The implementation includes authenticated workspace administration, domain-scoped expiring SMTP credentials, TLS, a durable outbox, per-workspace recipient and conservative spending limits, pause/suspension controls, and dispatch-time suppression checks. Campaign mail requires HTTPS one-click unsubscribe headers. Unsubscribed contacts are blocked. SES SNS bounce and complaint notifications are signature-verified, processed idempotently, and permanent bounces and complaints suppress future sends to the affected address. SES account suppression is enabled for both bounce and complaint. Tenant separation is used for customer sending resources.

This integration is still being activated; it has passed local PostgreSQL and SMTP/TLS integration tests but is not yet open to customers. Before starting the pilot, we will validate SES simulator feedback, notification delivery, monitoring, and operator response procedures in the AWS account. We understand that deliverability and complaint/bounce management remain our responsibility. Production access is requested so that the controlled pilot can send to opted-in recipients without individually verifying every recipient address.
