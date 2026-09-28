# Terraform adoption

No Terraform imports, plans against live infrastructure, or applies were performed when this package was created. Validation was local. Import updates Terraform state; applying a plan changes infrastructure.

## Prepare state and access

Use an operator role with short-lived credentials and the target account guard in each root. Do not use AWS root credentials for routine operations or give them to the application. Keep Cloudflare credentials in `CLOUDFLARE_API_TOKEN`, populated through your secret manager; the token needs Zone Read and DNS Edit scoped to the selected zone.

Pre-create a private S3 state bucket with versioning, encryption, public access blocked, and tightly scoped IAM. Terraform state and saved plans can contain sensitive data even when console output is redacted. The backend example enables encryption and native S3 locking; the operator needs Get/Put on the state object and Get/Put/Delete on its `.tflock` object. With KMS encryption, configure the KMS key and necessary key permissions as well. Do not commit local state, plans, backend configs or real tfvars.

Copy `terraform/backend.hcl.example` to a local `state.backend.hcl` **inside each environment**. Use distinct state keys (`xem/dokploy/terraform.tfstate`, `xem/domain/terraform.tfstate`, `xem/kubernetes/terraform.tfstate`). Initialize from each root:

```bash
terraform init -backend-config=state.backend.hcl
```

After local validation with `-backend=false`, a real initialization must explicitly configure the backend. Never assume a local validation state is the production state.

## Adopt existing Dokploy infrastructure

From `terraform/environments/dokploy`, copy `production.tfvars.example` to `production.tfvars`, verify the inventory against AWS, and initialize the remote backend. The example is the 2026-09-12 snapshot; it is not live discovery.

Import each existing resource once, using the same variable file and state. Skip resources already present in this state; never import the same object into another state.

```bash
terraform import -var-file=production.tfvars module.runtime.aws_iam_role.runtime xem-managed-sending-ec2
terraform import -var-file=production.tfvars 'module.runtime.aws_iam_instance_profile.runtime[0]' xem-managed-sending-ec2
terraform import -var-file=production.tfvars aws_security_group.smtp sg-07658f0e546336d23
terraform import -var-file=production.tfvars aws_instance.dokploy i-00aea05ab70b00750
terraform import -var-file=production.tfvars aws_eip.smtp eipalloc-05ef4e0c998579542
terraform import -var-file=production.tfvars aws_eip_association.smtp eipassoc-00971304f7e75161c
terraform import -var-file=production.tfvars module.sending.aws_cloudformation_stack.sending xem-managed-sending
terraform plan -var-file=production.tfvars -out=adoption.tfplan
terraform show -json adoption.tfplan | python3 ../../../scripts/check-adoption-plan.py
```

The guard rejects create/delete/replacement and unreviewed updates. Initial adoption should converge to a no-op plan. If a benign update is intentional, inspect every changed attribute and then use `--allow-updates`; this still rejects replacement. Do not ignore drift wholesale to make a plan pass. `prevent_destroy` protects critical resources while their configuration remains present; it does not protect resources removed from configuration/state or prevent a nested CloudFormation template update from replacing a child. Review the CloudFormation template diff too.

The imported EC2 root disk is **unencrypted**, 128 GiB gp3. Adoption preserves it. Encrypting it requires a separate snapshot/volume or instance migration with backups and downtime planning. Changing the AMI, subnet, encryption or disk configuration during adoption can request replacement and is blocked. The existing administrative security group is referenced rather than managed; this root neither opens SSH nor installs Dokploy. Egress on the SMTP SG remains unrestricted so the application can reach external dependencies; tighten it only with a complete dependency inventory.

The CloudFormation stack remains the sole owner of its generated SNS topic, runtime policy and conditional subscription. Do **not** import those children into Terraform resources or create a second copy with guessed names. `cloudformation.json` mirrors the backend's managed-sending template. Review both copies together when changing IAM. The runtime role has no static access keys.

Keep `feedback_endpoint = ""` during adoption because that is the existing stack value. After the backend is configured with the exact topic ARN, set the endpoint to `https://API_HOST/api/v1/sending/events/ses` in a separate reviewed change. Verify SNS subscription confirmation and signed event handling; a successful stack update alone does not prove feedback works.

## Platform domain and Cloudflare

This root owns only the platform domain (`xem.email`). Customer identities belong to Xem and must not be adopted here. It does not change the root domain's existing mailbox MX, SPF or iCloud DKIM records.

From `terraform/environments/domain`, copy `platform.tfvars.example` to `platform.tfvars`, set the **zone ID** (not the Cloudflare account ID), initialize its separate backend, and authenticate both providers. For existing production records:

```bash
terraform import -var-file=platform.tfvars module.domain.aws_sesv2_email_identity.platform xem.email
terraform import -var-file=platform.tfvars module.domain.aws_sesv2_email_identity_mail_from_attributes.platform xem.email
terraform console -var-file=platform.tfvars
# Evaluate module.domain.dkim_tokens to match each index to its existing DNS record.
```

Read the actual Cloudflare record IDs and match each DKIM token to its exact record; do not assume the UI order is the provider's token order. Substitute the verified IDs below:

```bash
terraform import -var-file=platform.tfvars 'module.domain.cloudflare_dns_record.dkim[0]' 'ZONE_ID/DKIM_RECORD_ID_0'
terraform import -var-file=platform.tfvars 'module.domain.cloudflare_dns_record.dkim[1]' 'ZONE_ID/DKIM_RECORD_ID_1'
terraform import -var-file=platform.tfvars 'module.domain.cloudflare_dns_record.dkim[2]' 'ZONE_ID/DKIM_RECORD_ID_2'
terraform import -var-file=platform.tfvars module.domain.cloudflare_dns_record.mail_from_mx 'ZONE_ID/BOUNCE_MX_RECORD_ID'
terraform import -var-file=platform.tfvars module.domain.cloudflare_dns_record.mail_from_spf 'ZONE_ID/BOUNCE_SPF_RECORD_ID'
terraform import -var-file=platform.tfvars module.domain.cloudflare_dns_record.smtp 'ZONE_ID/SMTP_RECORD_ID'
# Only when manage_dmarc=true; import and preserve the existing policy.
terraform import -var-file=platform.tfvars 'module.domain.cloudflare_dns_record.dmarc[0]' 'ZONE_ID/DMARC_RECORD_ID'
terraform plan -var-file=platform.tfvars -out=adoption.tfplan
terraform show -json adoption.tfplan | python3 ../../../scripts/check-adoption-plan.py
```

Defaults: 2048-bit DKIM; `bounce.DOMAIN` MX and SPF; SES MAIL FROM failure rejects the message. DMARC management is opt-in and starts at `p=none`; adopting an existing stricter policy requires setting its exact content, never weakening it automatically. An SMTP A record targets the EC2 EIP; a Kubernetes CNAME targets the NLB DNS name. Both must remain DNS-only (`proxied=false`). Do not move SMTP DNS until certificate, readiness and send tests pass on the destination.

For a genuinely new domain, the plan will intentionally create resources and the **adoption** guard will reject it. Review that creation plan separately rather than pretending it is an import. SES production access and account sending/suppression settings are account-level operations; this module neither requests production access nor resets those policies.

## Kubernetes IAM

The Kubernetes root takes an **existing** IAM OIDC provider and issuer URL. Its exact trust condition must match the Helm namespace and `workloadIdentity.serviceAccountName`. Use a distinct stack and role name, and a separate state key. See [Kubernetes deployment](kubernetes.md). Do not reuse the live Dokploy stack in both states.
