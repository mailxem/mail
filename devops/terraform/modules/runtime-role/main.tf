terraform {
  required_version = ">= 1.10, < 2.0"
  required_providers {
    aws = { source = "hashicorp/aws", version = ">= 6.0, < 7.0" }
  }
}

variable "name" { type = string }
variable "kind" {
  type = string
  validation {
    condition     = contains(["ec2", "web_identity"], var.kind)
    error_message = "Use ec2 or web_identity. Kubernetes must not inherit an EC2 node role."
  }
}
variable "oidc_provider_arn" {
  type    = string
  default = ""
}
variable "oidc_issuer_url" {
  type    = string
  default = ""
}
variable "namespace" {
  type    = string
  default = "xem"
}
variable "service_account" {
  type    = string
  default = "xem-backend"
}
variable "tags" {
  type    = map(string)
  default = {}
}
variable "instance_profile_tags" {
  type    = map(string)
  default = {}
}

data "aws_iam_policy_document" "trust" {
  dynamic "statement" {
    for_each = var.kind == "ec2" ? [1] : []
    content {
      actions = ["sts:AssumeRole"]
      principals {
        type        = "Service"
        identifiers = ["ec2.amazonaws.com"]
      }
    }
  }
  dynamic "statement" {
    for_each = var.kind == "web_identity" ? [1] : []
    content {
      actions = ["sts:AssumeRoleWithWebIdentity"]
      principals {
        type        = "Federated"
        identifiers = [var.oidc_provider_arn]
      }
      condition {
        test     = "StringEquals"
        variable = "${trimprefix(var.oidc_issuer_url, "https://")}:aud"
        values   = ["sts.amazonaws.com"]
      }
      condition {
        test     = "StringEquals"
        variable = "${trimprefix(var.oidc_issuer_url, "https://")}:sub"
        values   = ["system:serviceaccount:${var.namespace}:${var.service_account}"]
      }
    }
  }
}

resource "aws_iam_role" "runtime" {
  name               = var.name
  description        = var.kind == "ec2" ? "Xem EC2 managed sending runtime; scoped SES policy managed by CloudFormation" : "Xem managed sending workload identity; one Kubernetes service account"
  assume_role_policy = data.aws_iam_policy_document.trust.json
  tags               = var.tags
  lifecycle {
    prevent_destroy = true
    precondition {
      condition = var.kind != "web_identity" || (
        can(regex("^arn:aws:iam::[0-9]{12}:oidc-provider/", var.oidc_provider_arn)) &&
        startswith(var.oidc_issuer_url, "https://") &&
        try(split(":oidc-provider/", var.oidc_provider_arn)[1], "") == trimprefix(var.oidc_issuer_url, "https://") &&
        length(var.namespace) > 0 && length(var.service_account) > 0
      )
      error_message = "Web identity needs a matching IAM OIDC provider/HTTPS issuer and an exact namespace/service account."
    }
  }
}

resource "aws_iam_instance_profile" "runtime" {
  count = var.kind == "ec2" ? 1 : 0
  name  = var.name
  role  = aws_iam_role.runtime.name
  tags  = var.instance_profile_tags
  lifecycle { prevent_destroy = true }
}

output "name" { value = aws_iam_role.runtime.name }
output "arn" { value = aws_iam_role.runtime.arn }
output "instance_profile_name" { value = try(aws_iam_instance_profile.runtime[0].name, null) }
