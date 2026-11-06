#!/usr/bin/env bash
# Turns on the SES feedback queue on a managed-sending CloudFormation stack that
# was created with the AWS CLI, then stores the queue URL in Infisical.
#
# It only ever adds the queue: it previews a change set and refuses to apply it
# if anything other than the queue resources would be added, or anything other
# than the runtime policy would change. Running it again is safe.
#
# Usage: devops/scripts/enable-ses-events-queue.sh [--dry-run] [--yes]
#   --dry-run  preview the change set, then delete it without applying
#   --yes      apply without asking
#
# Settings (environment variables, defaults in brackets):
#   AWS_PROFILE [xem]  AWS_REGION [us-east-2]  STACK [xem-managed-sending]
#   TEMPLATE [server/deploy/managed-sending/cloudformation.json]
#   INFISICAL_PROJECT [XEM BACKEND project id]  INFISICAL_ENV [prod]
#   SKIP_INFISICAL=1 to leave Infisical alone (prints the URL instead)
#
# Needs: aws CLI v2 signed in (aws login --profile "$AWS_PROFILE"), python3,
# and the infisical CLI unless SKIP_INFISICAL=1.
#
# Stacks managed by Terraform must use `feedback_queue = true` in the
# sending-stack module instead; this script refuses Terraform-tagged stacks.
set -euo pipefail

export AWS_PROFILE="${AWS_PROFILE:-xem}" AWS_REGION="${AWS_REGION:-us-east-2}"
STACK="${STACK:-xem-managed-sending}"
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
TEMPLATE="${TEMPLATE:-$ROOT/server/deploy/managed-sending/cloudformation.json}"
INFISICAL_PROJECT="${INFISICAL_PROJECT:-7350b754-2ba4-46c3-9551-6c307415627d}"
INFISICAL_ENV="${INFISICAL_ENV:-prod}"
CHANGE_SET="enable-events-queue-$(date +%Y%m%d%H%M%S)"
DRY_RUN=0 YES=0
for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=1 ;;
    --yes) YES=1 ;;
    *) echo "unknown option: $arg" >&2; exit 2 ;;
  esac
done

say() { printf '\n== %s\n' "$*"; }
die() { printf 'error: %s\n' "$*" >&2; exit 1; }
output() { aws cloudformation describe-stacks --stack-name "$STACK" --query "Stacks[0].Outputs[?OutputKey=='$1'].OutputValue" --output text; }
cleanup() { aws cloudformation delete-change-set --stack-name "$STACK" --change-set-name "$CHANGE_SET" >/dev/null 2>&1 || true; }

say "Checking AWS sign-in ($AWS_PROFILE, $AWS_REGION)"
aws sts get-caller-identity --query Arn --output text >/dev/null 2>&1 ||
  die "not signed in: run 'aws login --profile $AWS_PROFILE' first"

say "Checking stack $STACK"
status=$(aws cloudformation describe-stacks --stack-name "$STACK" --query 'Stacks[0].StackStatus' --output text) ||
  die "stack $STACK not found in $AWS_REGION"
case "$status" in
  *_IN_PROGRESS|*FAILED|ROLLBACK_COMPLETE) die "stack is $status; wait or fix it first" ;;
esac
if aws cloudformation describe-stacks --stack-name "$STACK" --query 'Stacks[0].Tags[].Key' --output text | grep -qiw terraform; then
  die "stack is tagged as Terraform-managed; set feedback_queue = true in Terraform instead"
fi
[ -f "$TEMPLATE" ] || die "template not found: $TEMPLATE"
grep -q '"CreateEventsQueue"' "$TEMPLATE" || die "template has no CreateEventsQueue parameter (needs the SQS feedback change)"
aws cloudformation validate-template --template-body "file://$TEMPLATE" >/dev/null

queue_url=$(output FeedbackQueueURL)
if [ -n "$queue_url" ] && [ "$queue_url" != "None" ]; then
  echo "queue already enabled"
else
  say "Previewing change set $CHANGE_SET"
  params=()
  for key in $(aws cloudformation describe-stacks --stack-name "$STACK" --query 'Stacks[0].Parameters[].ParameterKey' --output text); do
    [ "$key" = CreateEventsQueue ] || params+=("ParameterKey=$key,UsePreviousValue=true")
  done
  params+=("ParameterKey=CreateEventsQueue,ParameterValue=true")
  aws cloudformation create-change-set --stack-name "$STACK" --change-set-name "$CHANGE_SET" \
    --template-body "file://$TEMPLATE" --parameters "${params[@]}" --capabilities CAPABILITY_NAMED_IAM >/dev/null
  if ! aws cloudformation wait change-set-create-complete --stack-name "$STACK" --change-set-name "$CHANGE_SET"; then
    aws cloudformation describe-change-set --stack-name "$STACK" --change-set-name "$CHANGE_SET" --query StatusReason --output text >&2
    cleanup; die "change set could not be created"
  fi
  changes=$(aws cloudformation describe-change-set --stack-name "$STACK" --change-set-name "$CHANGE_SET" \
    --query 'Changes[].ResourceChange.[Action,LogicalResourceId,ResourceType,Replacement]' --output text)
  printf '%s\n' "$changes"

  # Exactly what enabling the queue should do; anything else means the stack
  # has drifted from the template and needs a human look.
  if ! CHANGES="$changes" python3 - <<'PY'
import os, sys
allowed_adds = {"FeedbackQueue", "FeedbackDeadLetterQueue", "FeedbackQueuePolicy", "FeedbackQueueSubscription"}
bad = []
for line in os.environ["CHANGES"].splitlines():
    action, rid, _type, replace = (line.split("\t") + [""] * 4)[:4]
    if action == "Add" and rid in allowed_adds:
        continue
    if action == "Modify" and rid == "RuntimePolicy" and replace in ("False", "None", ""):
        continue
    bad.append(line)
if bad:
    print("unexpected changes:\n" + "\n".join(bad), file=sys.stderr)
    sys.exit(1)
PY
  then
    cleanup; die "refusing to apply: the change set does more than add the queue"
  fi

  if [ "$DRY_RUN" = 1 ]; then
    cleanup; echo "dry run: change set looks right and was deleted without applying"; exit 0
  fi
  if [ "$YES" != 1 ]; then
    read -r -p "Apply these changes to $STACK? [y/N] " answer
    [ "$answer" = y ] || [ "$answer" = Y ] || { cleanup; echo "not applied"; exit 0; }
  fi

  say "Applying"
  aws cloudformation execute-change-set --stack-name "$STACK" --change-set-name "$CHANGE_SET"
  aws cloudformation wait stack-update-complete --stack-name "$STACK" ||
    die "stack update failed; see the stack's events in the CloudFormation console"
  queue_url=$(output FeedbackQueueURL)
fi

say "Verifying"
topic=$(output TopicARN)
queue_arn=$(aws sqs get-queue-attributes --queue-url "$queue_url" --attribute-names QueueArn --query Attributes.QueueArn --output text)
sub=$(aws sns list-subscriptions-by-topic --topic-arn "$topic" --query "Subscriptions[?Endpoint=='$queue_arn'].SubscriptionArn" --output text)
case "$sub" in
  ""|None|PendingConfirmation) die "the queue's SNS subscription is missing or unconfirmed" ;;
esac
raw=$(aws sns get-subscription-attributes --subscription-arn "$sub" --query Attributes.RawMessageDelivery --output text)
[ "$raw" != true ] || die "raw message delivery is on; Xem needs the SNS envelope, turn it off"
aws sqs get-queue-attributes --queue-url "$queue_url" --attribute-names RedrivePolicy --query Attributes.RedrivePolicy --output text |
  grep -q deadLetterTargetArn || die "the queue has no dead-letter queue"
echo "queue subscribed (raw delivery off) with a dead-letter queue"

if [ "${SKIP_INFISICAL:-0}" = 1 ]; then
  say "Set SES_EVENTS_QUEUE_URL=$queue_url in the backend's environment"
  exit 0
fi
say "Saving SES_EVENTS_QUEUE_URL in Infisical ($INFISICAL_ENV)"
infisical secrets set "SES_EVENTS_QUEUE_URL=$queue_url" --projectId "$INFISICAL_PROJECT" --env "$INFISICAL_ENV" --silent >/dev/null
[ "$(infisical secrets get SES_EVENTS_QUEUE_URL --projectId "$INFISICAL_PROJECT" --env "$INFISICAL_ENV" --plain --silent)" = "$queue_url" ] ||
  die "Infisical didn't return the value just saved"
echo "done: redeploy the backend so it starts reading the queue"
