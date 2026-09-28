# Managed SMTP article — 12 September 2026

The article describes the implementation in backend PR [#10](https://github.com/mailxem/xem.go/pull/10) and frontend PR [#4](https://github.com/mailxem/xem-app.ts/pull/4). Code and the backend deployment runbook were reviewed for each product claim. The domain/production-access status comes from this setup session's AWS responses, recorded in backend commit `2b09fa5`. The dedicated PostgreSQL/SMTP integration job passed on GitHub; repository-wide checks still fail on pre-existing AI test compilation and the Cori GitHub App integration. The article claims only the dedicated checks passed.

Official documentation retrieved for this article:

- https://docs.aws.amazon.com/ses/latest/dg/Welcome.html
- https://docs.aws.amazon.com/ses/latest/dg/mail-from.html
- https://docs.aws.amazon.com/ses/latest/dg/tenants.html
- https://docs.aws.amazon.com/ses/latest/dg/manage-sending-quotas.html
- https://docs.aws.amazon.com/ses/latest/dg/monitor-sending-activity-using-notifications-sns.html
- https://docs.aws.amazon.com/sns/latest/dg/sns-verify-signature-of-message.html

All four illustrations are original code-drawn vectors in Xem's existing palette. `scripts/render-managed-smtp-art.mjs` is their editable source; it writes three public SVG diagrams and a 1600×1067 WebP cover. There are no licensed stock inputs, provider logos, customer data, synthetic numerical measurements, or performance claims in these assets. The diagrams explain address roles, component flow, and status semantics. Article text provides equivalent descriptions, and the images have descriptive alternative text.

Keep the publication's pilot status in sync with actual deployment. SES production access is not approval of the Xem application. Runtime IAM, public SMTP/TLS, live feedback, monitoring, and the reviewed pilot are still operational prerequisites; the article does not advertise them as live services.

The project owner also supplied a correction to the second testimonial's name and role: Girish Raju, CTO @Zunofy. See `testimonial-provenance.md`; this change does not independently verify the quoted outcome.
