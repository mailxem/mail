#!/usr/bin/env bash
# Download the source and execute the installer from the same selected revision.
set -euo pipefail
for tool in git python3 docker openssl; do
  command -v "$tool" >/dev/null || { echo "Install $tool first." >&2; exit 1; }
done
if [[ ${1:-} == --help ]]; then
  echo 'Usage: bash install.sh [--config /path/to/config.json] [--state /path/to/state]'
  echo 'Set XEM_REF to a published branch, tag, or commit (default: undefined).'
  exit 0
fi
checkout=$(mktemp -d)
trap 'rm -rf "$checkout"' EXIT
git clone --quiet --no-checkout https://github.com/mailxem/mail.git "$checkout"
git -C "$checkout" checkout --quiet --detach "${XEM_REF:-origin/undefined}"
python3 "$checkout/devops/swarm/install.py" "$@"
