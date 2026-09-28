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
variable "ami" { type = string }
variable "subnet_id" { type = string }
variable "vpc_id" { type = string }
variable "existing_security_group_ids" { type = list(string) }
variable "key_name" { type = string }
variable "instance_type" { default = "t3.medium" }
variable "root_volume_size" { default = 128 }
# Adoption preserves the existing disk. Encryption requires a separate migration.
variable "root_volume_encrypted" { default = false }
variable "feedback_endpoint" { default = "" }

module "runtime" {
  source                = "../../modules/runtime-role"
  name                  = "xem-managed-sending-ec2"
  kind                  = "ec2"
  tags                  = { Project = "Xem", Purpose = "ManagedSending" }
  instance_profile_tags = { Project = "Xem" }
}
module "sending" {
  source            = "../../modules/sending-stack"
  runtime_role_name = module.runtime.name
  feedback_endpoint = var.feedback_endpoint
}
resource "aws_security_group" "smtp" {
  name        = "xem-smtp-submission"
  description = "Public authenticated SMTP submission with STARTTLS; Xem EC2 only"
  vpc_id      = var.vpc_id
  ingress {
    description = "Xem authenticated SMTP STARTTLS"
    from_port   = 587
    to_port     = 587
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  tags = { Project = "Xem", Name = "xem-smtp-submission" }
  lifecycle { prevent_destroy = true }
}
resource "aws_instance" "dokploy" {
  ami                    = var.ami
  instance_type          = var.instance_type
  subnet_id              = var.subnet_id
  key_name               = var.key_name
  vpc_security_group_ids = concat(var.existing_security_group_ids, [aws_security_group.smtp.id])
  iam_instance_profile   = module.runtime.instance_profile_name
  ebs_optimized          = true
  monitoring             = false
  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 2
    http_protocol_ipv6          = "disabled"
    instance_metadata_tags      = "disabled"
  }
  root_block_device {
    volume_type           = "gp3"
    volume_size           = var.root_volume_size
    iops                  = 3000
    throughput            = 125
    encrypted             = var.root_volume_encrypted
    delete_on_termination = true
  }
  tags = { Name = "xem-server" }
  lifecycle { prevent_destroy = true }
}
resource "aws_eip" "smtp" {
  domain = "vpc"
  tags   = { Name = "xem" }
  lifecycle { prevent_destroy = true }
}
resource "aws_eip_association" "smtp" {
  instance_id   = aws_instance.dokploy.id
  allocation_id = aws_eip.smtp.id
  lifecycle { prevent_destroy = true }
}
output "smtp_ip" { value = aws_eip.smtp.public_ip }
output "runtime_role_arn" { value = module.runtime.arn }
output "topic_arn" { value = module.sending.topic_arn }
