terraform {
  required_version = ">= 1.10, < 2.0"
  required_providers {
    aws = { source = "hashicorp/aws", version = ">= 6.0, < 7.0" }
  }
}

variable "name" {
  type    = string
  default = "xem-managed-sending"
}
variable "runtime_role_name" { type = string }
variable "feedback_endpoint" {
  type    = string
  default = ""
  validation {
    condition     = var.feedback_endpoint == "" || can(regex("^https://[^/]+/api/v1/sending/events/ses$", var.feedback_endpoint))
    error_message = "Leave feedback empty until the backend is ready, then use its exact HTTPS SES event route."
  }
}
# SQS feedback ingestion: SES events reach the backend in batches through a
# queue (with a dead-letter queue after 5 receives) instead of one HTTPS request
# each. Set SES_EVENTS_QUEUE_URL to the feedback_queue_url output.
variable "feedback_queue" {
  type    = bool
  default = false
}
# Existing SNS topic for the feedback queue's dead-letter and backlog alarms.
variable "alarm_topic_arn" {
  type    = string
  default = ""
}
variable "tags" {
  type    = map(string)
  default = { Project = "Xem", Purpose = "ManagedSending" }
}

# Adopt the STACK, not its children. CloudFormation remains the single owner of
# its existing topic/policy/subscription and preserves generated resource ARNs.
# The template is the one in server/deploy, also used by
# devops/scripts/enable-ses-events-queue.sh, so the two can't drift apart.
resource "aws_cloudformation_stack" "sending" {
  name          = var.name
  template_body = file("${path.module}/../../../../server/deploy/managed-sending/cloudformation.json")
  capabilities  = ["CAPABILITY_NAMED_IAM"]
  parameters = {
    RuntimeRoleName   = var.runtime_role_name
    FeedbackEndpoint  = var.feedback_endpoint
    CreateEventsQueue = tostring(var.feedback_queue)
    AlarmTopicArn     = var.alarm_topic_arn
  }
  tags = var.tags
  lifecycle { prevent_destroy = true }
}

output "topic_arn" { value = aws_cloudformation_stack.sending.outputs["TopicARN"] }
output "runtime_policy_arn" { value = aws_cloudformation_stack.sending.outputs["RuntimePolicyARN"] }
output "region" { value = aws_cloudformation_stack.sending.outputs["Region"] }
output "account_id" { value = aws_cloudformation_stack.sending.outputs["AccountID"] }
output "feedback_queue_url" { value = lookup(aws_cloudformation_stack.sending.outputs, "FeedbackQueueURL", "") }
