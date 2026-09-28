#!/usr/bin/env bash
# Static/local validation only: no cloud credentials, terraform apply, or helm install.
set -euo pipefail
cd "$(dirname "$0")/.."
tf_bin=${TF_BIN:-terraform}
"$tf_bin" fmt -check -recursive terraform
for env in dokploy kubernetes domain; do
  "$tf_bin" -chdir="terraform/environments/$env" init -backend=false -input=false -lockfile=readonly
  "$tf_bin" -chdir="terraform/environments/$env" validate
 done
"$tf_bin" -chdir=terraform/modules/runtime-role init -backend=false -input=false -lockfile=readonly
"$tf_bin" -chdir=terraform/modules/runtime-role test
helm lint charts/xem --strict
helm lint charts/xem --strict -f charts/xem/ci/managed.yaml
python3 scripts/test-infrastructure.py
bash -n scripts/install-dokploy-smtp-tls.sh
python3 -m py_compile scripts/*.py
mkdir -p .validation
helm template test charts/xem > .validation/default.yaml
helm template test charts/xem -f charts/xem/ci/managed.yaml > .validation/managed.yaml
# cert-manager CRDs are checked separately with their upstream OpenAPI schema.
kubeconform -strict -summary -kubernetes-version 1.34.0 -skip Certificate,Issuer .validation/default.yaml .validation/managed.yaml
python3 scripts/validate-cert-manager.py .validation/managed.yaml
git diff --check
