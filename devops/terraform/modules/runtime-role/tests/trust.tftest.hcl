# Plan only: policy-document construction is local. No AWS lookup or mutation.
provider "aws" {
  region                      = "us-east-2"
  access_key                  = "testing"
  secret_key                  = "testing"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  skip_metadata_api_check     = true
  skip_region_validation      = true
}
variables {
  name = "xem-test-runtime"
  kind = "ec2"
}
run "ec2_trust" {
  command = plan
  assert {
    condition     = jsondecode(aws_iam_role.runtime.assume_role_policy).Statement[0].Principal.Service == "ec2.amazonaws.com"
    error_message = "EC2 role must trust only EC2."
  }
  assert {
    condition     = length(aws_iam_instance_profile.runtime) == 1
    error_message = "EC2 requires an instance profile."
  }
}
run "exact_workload_trust" {
  command = plan
  variables {
    kind              = "web_identity"
    oidc_provider_arn = "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-2.amazonaws.com/id/EXAMPLE"
    oidc_issuer_url   = "https://oidc.eks.us-east-2.amazonaws.com/id/EXAMPLE"
    namespace         = "xem"
    service_account   = "xem-backend"
  }
  assert {
    condition     = jsondecode(aws_iam_role.runtime.assume_role_policy).Statement[0].Condition.StringEquals["oidc.eks.us-east-2.amazonaws.com/id/EXAMPLE:sub"] == "system:serviceaccount:xem:xem-backend"
    error_message = "Trust must use an exact service account, never a wildcard."
  }
  assert {
    condition     = jsondecode(aws_iam_role.runtime.assume_role_policy).Statement[0].Condition.StringEquals["oidc.eks.us-east-2.amazonaws.com/id/EXAMPLE:aud"] == "sts.amazonaws.com"
    error_message = "Trust must require the STS audience."
  }
  assert {
    condition     = length(aws_iam_instance_profile.runtime) == 0
    error_message = "Kubernetes must not create an EC2 instance profile."
  }
}
run "reject_mismatched_issuer" {
  command = plan
  variables {
    kind              = "web_identity"
    oidc_provider_arn = "arn:aws:iam::123456789012:oidc-provider/other.example.com"
    oidc_issuer_url   = "https://oidc.eks.us-east-2.amazonaws.com/id/EXAMPLE"
  }
  expect_failures = [aws_iam_role.runtime]
}
run "reject_missing_identity" {
  command = plan
  variables { kind = "web_identity" }
  expect_failures = [aws_iam_role.runtime]
}
