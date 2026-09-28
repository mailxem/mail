terraform {
  required_version = ">= 1.10, < 2.0"
  backend "s3" {}
  required_providers {
    aws        = { source = "hashicorp/aws", version = "~> 6.64.0" }
    cloudflare = { source = "cloudflare/cloudflare", version = "~> 5.25.0" }
  }
}
provider "aws" {
  region              = var.region
  allowed_account_ids = [var.account_id]
}
provider "cloudflare" {}
variable "account_id" { type = string }
variable "region" { default = "us-east-2" }
variable "zone_id" { type = string }
variable "domain" { type = string }
variable "smtp_hostname" { type = string }
variable "smtp_target" { type = string }
variable "smtp_record_type" { default = "A" }
variable "manage_dmarc" { default = false }
variable "dmarc_record" { default = "v=DMARC1; p=none" }
module "domain" {
  source           = "../../modules/sending-domain"
  region           = var.region
  zone_id          = var.zone_id
  domain           = var.domain
  smtp_hostname    = var.smtp_hostname
  smtp_target      = var.smtp_target
  smtp_record_type = var.smtp_record_type
  manage_dmarc     = var.manage_dmarc
  dmarc_record     = var.dmarc_record
}
output "dkim_tokens" { value = module.domain.dkim_tokens }
