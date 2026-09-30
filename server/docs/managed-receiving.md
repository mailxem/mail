# Managed receiving on a custom domain

Xem's managed receiving implementation is an opt-in deployment path under development. It receives mail for explicitly created Xem mailbox addresses through Amazon SES in `us-east-2`, stores raw MIME privately in S3, sends SES receipt notifications through SNS and SQS, and indexes mailbox metadata in PostgreSQL. It does not expose IMAP, create a catch-all address, or import an existing external mailbox. Replies use the same address through Xem's existing managed SES sender.

No AWS resource or DNS record is changed by this repository alone. The receiving stack creates an empty, inactive receipt rule set and never calls `SetActiveReceiptRuleSet`. The backend can add, update, and remove only its named exact-recipient rules in the configured active set; it has no permission to activate a rule set. Xem-owned rules enable SES spam and virus scanning and contain one S3 action that publishes its receipt notification to SNS, without a Stop action, so one SMTP transaction addressed to mailboxes on multiple managed domains can reach every matching rule.

## Architecture and retention

SES writes raw messages under `incoming/` in a private, TLS-only S3 bucket and publishes a notification to a dedicated SNS topic. SNS delivers its signed envelope to an encrypted SQS queue; raw message delivery is deliberately disabled. The backend validates the configured topic ARN, SNS signature and signing-certificate host, then checks the receipt's S3 action against the configured bucket and prefix before using the event. AWS account, region, and rule-set provenance are enforced by the SNS and S3 resource policies because SES receipt notifications do not include a `mail.sourceArn` field. It deletes the SQS message only after an idempotent database commit.

SES can place messages up to its 40 MB S3-action limit, but this implementation defaults to and caps accepted raw MIME at 10 MiB. Oversized and invalid MIME events are recorded as rejections and acknowledged without creating mailbox messages. Over-quota and transient processing failures remain unacknowledged and eventually reach the DLQ after repeated delivery. The queue retains events for four days, retries five times with a five-minute visibility timeout, and the DLQ retains exhausted events for 14 days.

Only a trusted SES `receipt.virusVerdict.status` of `PASS` can become visible mailbox content. A missing verdict or `FAIL`, `GRAY`, or `PROCESSING_FAILED` result is recorded as quarantined and acknowledged without creating a mailbox message. SES spam results are recorded separately; they never override a non-passing virus verdict.

Virus-PASS mail is copied idempotently to `mail/<SES-message-id>` before the database commit. The `incoming/` staging prefix expires after 30 days; the durable `mail/` prefix has no automatic expiry. One provider message and durable object can be referenced by several mailbox rows, so deleting or pausing one mailbox must not delete that object. Accepted-mail retention, backups, quota recovery, and any future purge must preserve every remaining database reference. The initial quota for newly created mailboxes is controlled by `MANAGED_RECEIVING_MAILBOX_QUOTA_BYTES` and defaults to 1 GiB; changing that environment value does not update existing mailbox rows. A full mailbox is removed from its SES receipt rule; an event that races with the quota boundary remains retryable or moves to the DLQ rather than being silently lost. Automated quarantine/orphan purge and self-service quota management are not included in this release.

## Prerequisites and DNS boundary

Use an AWS account and domain configured for Xem managed sending. The workspace must own and verify the exact domain and must not be suspended. A sending-only account pause does not stop receiving, though it leaves the linked sender unavailable for replies until sending resumes. Workspace administrators with the existing managed-mail permissions can create mailbox addresses, and mailbox contents are workspace-wide. Private personal mailboxes, delegated access, and per-mailbox roles are not provided.

Start with a dedicated subdomain and controlled addresses. In `us-east-2`, the receiving MX is:

```text
10 inbound-smtp.us-east-2.amazonaws.com
```

This is different from the custom MAIL FROM MX used for bounces, typically:

```text
10 feedback-smtp.us-east-2.amazonses.com
```

Publishing the bounce MX does not route incoming mailbox mail to Xem. Publish the receiving MX at the exact receiving domain. Do not mix SES receiving MX records with another mailbox provider at the same domain: equal or different MX priorities create delivery, fallback, and migration behavior that Xem cannot reconcile. To preserve an existing provider, test on a new subdomain. For a migration, inventory every address, lower DNS TTL in advance, create and test explicit Xem mailboxes, then replace the old receiving MX during a controlled window. Never rely on a catch-all; Xem provisions exact recipients only.

## Create and inspect the AWS stack

Use short-lived operator credentials. These commands are read-only until the change-set creation step and print no secrets:

```bash
export AWS_PROFILE=xem
export AWS_REGION=us-east-2
aws sts get-caller-identity
aws configure get region --profile "$AWS_PROFILE"
aws ses describe-active-receipt-rule-set --region "$AWS_REGION"
aws ses list-receipt-rule-sets --region "$AWS_REGION"
aws cloudformation validate-template \
  --region "$AWS_REGION" \
  --template-body file://devops/managed-receiving/cloudformation.yaml
```

Record the caller account, configured region, active rule-set name and every existing rule. If another rule set is active, stop and plan its migration into the dedicated set. SES supports one active receipt rule set per region/account; activating the new set replaces the old active choice even though it does not delete the old set. Preserve unrelated rules, ordering, recipients and actions. The backend preserves unrelated rules inside the configured set, but cannot merge two rule sets.

Create a change set without executing it:

```bash
aws cloudformation create-change-set \
  --region "$AWS_REGION" \
  --stack-name xem-managed-receiving \
  --change-set-name initial-review \
  --change-set-type CREATE \
  --capabilities CAPABILITY_IAM \
  --template-body file://devops/managed-receiving/cloudformation.yaml \
  --parameters \
    ParameterKey=RuleSetName,ParameterValue=xem-managed-receiving \
    ParameterKey=AlarmTopicArn,ParameterValue=EXISTING_OPERATOR_ALARM_TOPIC_ARN

aws cloudformation describe-change-set \
  --region "$AWS_REGION" \
  --stack-name xem-managed-receiving \
  --change-set-name initial-review
```

Review every resource and IAM statement. Execution is a separate operator decision:

```bash
aws cloudformation execute-change-set \
  --region "$AWS_REGION" \
  --stack-name xem-managed-receiving \
  --change-set-name initial-review
aws cloudformation wait stack-create-complete \
  --region "$AWS_REGION" \
  --stack-name xem-managed-receiving
aws cloudformation describe-stacks \
  --region "$AWS_REGION" \
  --stack-name xem-managed-receiving \
  --query 'Stacks[0].Outputs'
```

The outputs provide `BucketName`, `TopicArn`, `QueueUrl`, `DeadLetterQueueUrl`, `RuleSetName`, `RuntimePolicyArn`, `Region`, and `AccountId`. `AlarmTopicArn` must name an existing operator-owned SNS topic with confirmed responders before production; leaving it blank creates alarm state without sending a notification. Confirm both alarm action configurations after deployment. The managed policy is intentionally not attached automatically. After reviewing it, attach it to the existing backend workload role, never a node-wide role:

```bash
aws iam attach-role-policy \
  --role-name EXISTING_XEM_BACKEND_ROLE \
  --policy-arn RUNTIME_POLICY_ARN
aws iam list-attached-role-policies --role-name EXISTING_XEM_BACKEND_ROLE
```

The policy can read only `incoming/` and `mail/` in this bucket, write only `mail/`, consume only this queue, and manage SES receipt rules. SES Classic receipt-rule APIs do not provide reliable resource-level IAM scoping, so those actions require `Resource: "*"`; the policy restricts them to the deployed region and deliberately omits rule-set activation and deletion. The backend additionally requires the configured rule-set name and Xem-owned rule-name prefix. Keep this role isolated from unrelated applications. The S3 and SNS resource policies accept SES only from this account and rules in the configured rule set, including SES's rule-creation validation write.

## Configure, deploy, and activate deliberately

Set these backend values through the deployment secret/configuration system. They are identifiers and limits, not AWS credentials:

```dotenv
MANAGED_RECEIVING_ENABLED=false
MANAGED_SES_REGION=us-east-2
MANAGED_AWS_ACCOUNT_ID=123456789012
MANAGED_RECEIVING_BUCKET=STACK_BUCKET_NAME
MANAGED_RECEIVING_TOPIC_ARN=STACK_TOPIC_ARN
MANAGED_RECEIVING_QUEUE_URL=STACK_QUEUE_URL
MANAGED_RECEIVING_RULE_SET=xem-managed-receiving
MANAGED_RECEIVING_MAX_MESSAGE_BYTES=10485760
MANAGED_RECEIVING_MAILBOX_QUOTA_BYTES=1073741824
```

Use workload identity; do not add static AWS keys. Deploy the backend migration and backend worker before the client. Keep `MANAGED_RECEIVING_ENABLED=false` while validating startup, database migration, role assumption, queue attributes and S3 access. A successful stack deployment does not prove receipt.

After reconciling every pre-existing rule into the dedicated set and verifying that no active receiving flow will be displaced, activation must be a separate explicit operator action:

```bash
aws ses describe-active-receipt-rule-set --region "$AWS_REGION"
aws ses describe-receipt-rule-set \
  --region "$AWS_REGION" \
  --rule-set-name xem-managed-receiving
aws ses set-active-receipt-rule-set \
  --region "$AWS_REGION" \
  --rule-set-name xem-managed-receiving
aws ses describe-active-receipt-rule-set --region "$AWS_REGION"
```

Only then set `MANAGED_RECEIVING_ENABLED=true`, redeploy the backend, and create one controlled mailbox on the test subdomain. Rule creation must produce one enabled, TLS-required, scan-enabled exact-recipient rule with an S3 `incoming/` action that publishes to the configured SNS topic and no Stop action. Do not publish receiving MX until that rule, queue consumer and database are ready.

## Acceptance checks

For a controlled address, verify all of the following before broader DNS migration:

1. `dig +short MX receive.example.com` returns only the intended SES inbound endpoint.
2. The active rule set still contains every unrelated rule unchanged and the Xem rule names only the requested recipient. There is no catch-all and no Stop action.
3. A controlled message creates `incoming/<SES-message-id>`, one SNS envelope in SQS, a durable `mail/<SES-message-id>` object, and one idempotent PostgreSQL message. No S3 object is public.
4. Duplicate delivery of the same queue event does not create another message. The SQS message disappears only after commit.
5. A virus verdict other than PASS is recorded as quarantined, and oversized or invalid MIME is recorded as rejected; these events are acknowledged without mailbox content. An over-quota event remains retryable and reaches the DLQ after the configured attempts.
6. The new message appears in the workspace inbox, its attachment bytes match, and a reply sends from the same address through managed sending.
7. Pausing a mailbox removes or disables its exact receipt rule and sender while retaining its indexed messages and durable raw MIME. Resume restores the exact address without changing unrelated rules.
8. CloudWatch shows a healthy main queue, empty DLQ, and both alarms in `OK`. Trigger a controlled alarm notification and confirm receipt through `AlarmTopicArn`; an unrouted CloudWatch alarm is not an operational alert.

Also test the SES receipt-rule quota before onboarding. AWS currently allows 200 receipt rules per account/region and 100 recipients per rule. Xem uses per-domain rules with explicit recipients, so one domain can contain at most 100 active mailbox addresses in one rule and the shared account can exhaust its rule quota before storage quota. Treat AWS Service Quotas, SES rule count, S3 bytes, mailbox quota, SQS age, DLQ depth, database growth, and bounce/complaint reputation as separate limits.

## Operations and recovery

- **Backlog:** inspect `ApproximateAgeOfOldestMessage`, visible/in-flight counts, backend health, PostgreSQL capacity and S3 access. Do not purge the queue. Restore the consumer; duplicate processing is safe.
- **Poison/DLQ:** retain DLQ messages for investigation. Correlate the SES message ID, S3 key and database ingestion record without logging body content. Oversized, invalid MIME, and non-passing virus events normally have rejection rows and do not enter the DLQ. Fix quota, storage, database, or other retryable failures before redriving a bounded sample. Never redrive an event whose durable commit cannot be determined.
- **Full mailbox:** confirm the exact receipt rule no longer includes that address. Recovery requires an operator to adjust the existing mailbox quota record or otherwise restore capacity, resume and reconcile the mailbox rule, then inspect each affected S3 staging object and database state before a bounded DLQ redrive. Changing `MANAGED_RECEIVING_MAILBOX_QUOTA_BYTES` alone affects only new mailboxes. There is no self-service purge or quota-change endpoint, and freeing quota does not automatically prove old DLQ events are safe to replay.
- **S3:** back up durable `mail/` objects and PostgreSQL together and test a paired restore. Database-only restore can point at missing MIME; bucket-only restore cannot reconstruct tenant metadata. Do not delete shared raw objects when deleting one mailbox row.
- **Quarantine/orphans:** there is no automated purge in this release. `incoming/` expires after 30 days as a staging safety bound; accepted `mail/` does not. Export evidence needed for an incident before expiry.
- **Quotas:** alarm before SES rule, S3, SQS, PostgreSQL or mailbox limits. A provider quota increase does not change Xem's per-mailbox cap or onboarding policy.

## Rollback without mail loss

Pause new mailbox creation first. Remove the receiving MX to stop new internet delivery, but keep the backend worker, active rule set, S3 bucket, SNS topic, queues, database and IAM policy until DNS caches have expired and both the main queue and DLQ are reconciled. Pausing an individual mailbox retains its messages but disables receiving and its linked sender.

Do not delete the CloudFormation stack as a first rollback step. The bucket, topic, queues and receipt rule set carry retention policies, but their subscriptions and policies can still be removed by stack deletion and strand mail. To restore a previous provider, reinstate its complete MX set in one controlled change; do not leave mixed providers as a fallback. After the queues are empty, all quarantined/ambiguous events are resolved, backups are verified, and the old provider is accepting mail, detach the runtime policy and disable `MANAGED_RECEIVING_ENABLED`. Preserve the database and durable `mail/` objects for the retention period chosen by the operator.

## AWS references

- [Amazon SES S3 receipt action](https://docs.aws.amazon.com/ses/latest/dg/receiving-email-action-s3.html)
- [Publishing an MX record for SES receiving](https://docs.aws.amazon.com/ses/latest/dg/receiving-email-mx-record.html)
- [AWS service endpoints and quotas: SES](https://docs.aws.amazon.com/general/latest/gr/ses.html)
