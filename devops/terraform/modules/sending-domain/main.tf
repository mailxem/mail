terraform {
  required_version = ">= 1.10, < 2.0"
  required_providers {
    aws        = { source = "hashicorp/aws", version = ">= 6.0, < 7.0" }
    cloudflare = { source = "cloudflare/cloudflare", version = ">= 5.0, < 6.0" }
  }
}

variable "domain" { type = string }
variable "zone_id" { type = string }
variable "region" { type = string }
variable "smtp_hostname" { type = string }
variable "smtp_record_type" {
  type = string
  validation {
    condition     = contains(["A", "CNAME"], var.smtp_record_type)
    error_message = "Use A for the EC2 Elastic IP or CNAME for the Kubernetes NLB hostname."
  }
}
variable "smtp_target" { type = string }
variable "manage_dmarc" {
  type    = bool
  default = false
}
variable "dmarc_record" {
  type    = string
  default = "v=DMARC1; p=none"
}

# Platform-owned domain only. Customer/workspace identities are provisioned by
# the application; never import them into this module.
resource "aws_sesv2_email_identity" "platform" {
  email_identity = var.domain
  dkim_signing_attributes { next_signing_key_length = "RSA_2048_BIT" }
  lifecycle { prevent_destroy = true }
}

resource "aws_sesv2_email_identity_mail_from_attributes" "platform" {
  email_identity         = aws_sesv2_email_identity.platform.email_identity
  mail_from_domain       = "bounce.${var.domain}"
  behavior_on_mx_failure = "REJECT_MESSAGE"
  lifecycle { prevent_destroy = true }
}

resource "cloudflare_dns_record" "dkim" {
  count   = 3
  zone_id = var.zone_id
  name    = "${aws_sesv2_email_identity.platform.dkim_signing_attributes[0].tokens[count.index]}._domainkey.${var.domain}"
  type    = "CNAME"
  content = "${aws_sesv2_email_identity.platform.dkim_signing_attributes[0].tokens[count.index]}.dkim.amazonses.com"
  proxied = false
  ttl     = 300
  lifecycle { prevent_destroy = true }
}

resource "cloudflare_dns_record" "mail_from_mx" {
  zone_id  = var.zone_id
  name     = "bounce.${var.domain}"
  type     = "MX"
  content  = "feedback-smtp.${var.region}.amazonses.com"
  priority = 10
  ttl      = 300
  lifecycle { prevent_destroy = true }
}
resource "cloudflare_dns_record" "mail_from_spf" {
  zone_id = var.zone_id
  name    = "bounce.${var.domain}"
  type    = "TXT"
  content = "v=spf1 include:amazonses.com ~all"
  ttl     = 300
  lifecycle { prevent_destroy = true }
}
resource "cloudflare_dns_record" "dmarc" {
  count   = var.manage_dmarc ? 1 : 0
  zone_id = var.zone_id
  name    = "_dmarc.${var.domain}"
  type    = "TXT"
  content = var.dmarc_record
  ttl     = 300
  lifecycle { prevent_destroy = true }
}
resource "cloudflare_dns_record" "smtp" {
  zone_id = var.zone_id
  name    = var.smtp_hostname
  type    = var.smtp_record_type
  content = var.smtp_target
  proxied = false
  ttl     = 300
  lifecycle { prevent_destroy = true }
}

output "dkim_tokens" { value = aws_sesv2_email_identity.platform.dkim_signing_attributes[0].tokens }
output "identity_arn" { value = aws_sesv2_email_identity.platform.arn }
