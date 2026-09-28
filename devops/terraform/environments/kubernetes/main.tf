terraform {
  required_version = ">= 1.10, < 2.0"
  backend "s3" {}
  required_providers {
    aws = { source = "hashicorp/aws", version = "~> 6.64.0" }
  }
}
provider "aws" {
  region              = var.region
  allowed_account_ids = [var.account_id]
}
variable "region" { default = "us-east-2" }
variable "account_id" { type = string }
variable "name" {
  type    = string
  default = "xem-kubernetes"
  validation {
    condition     = var.name != "xem-managed-sending"
    error_message = "Use a distinct Kubernetes stack name; the live Dokploy stack belongs to its own state."
  }
}
variable "oidc_provider_arn" { type = string }
variable "oidc_issuer_url" { type = string }
variable "namespace" { default = "xem" }
variable "service_account" { default = "xem-backend" }
variable "feedback_endpoint" { default = "" }
module "runtime" {
  source            = "../../modules/runtime-role"
  kind              = "web_identity"
  name              = "${var.name}-runtime"
  oidc_provider_arn = var.oidc_provider_arn
  oidc_issuer_url   = var.oidc_issuer_url
  namespace         = var.namespace
  service_account   = var.service_account
  tags              = { Project = "Xem", Purpose = "ManagedSending" }
}
module "sending" {
  source            = "../../modules/sending-stack"
  name              = var.name
  runtime_role_name = module.runtime.name
  feedback_endpoint = var.feedback_endpoint
}
output "runtime_role_arn" { value = module.runtime.arn }
output "topic_arn" { value = module.sending.topic_arn }
