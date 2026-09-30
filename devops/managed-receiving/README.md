# AWS managed receiving stack

`cloudformation.yaml` is an additive, opt-in Amazon SES receiving stack. It creates private S3 storage, an SES-restricted SNS topic, an encrypted SQS queue and 14-day DLQ, queue alarms, a separate runtime IAM policy, and an empty **inactive** receipt rule set. It does not change the account's active receipt rule set, DNS, existing managed-sending stack, or backend role attachment.

Follow the complete [managed receiving operations guide](../../server/docs/managed-receiving.md) for preflight inspection, change-set review, deliberate policy attachment and rule-set activation, DNS migration, acceptance, monitoring, recovery, and rollback. Local template tests are in `devops/scripts/test-infrastructure.py`; they perform no AWS calls.
