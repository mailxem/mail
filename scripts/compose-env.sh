#!/usr/bin/env bash
# Write .env for docker compose with fresh credentials, the same kinds install.py generates.
set -euo pipefail
cd "$(dirname "$0")/.."
host=${1:-}
if [[ -z $host || $host == localhost || $host == 127.0.0.1 ]]; then
  echo 'Usage: bash scripts/compose-env.sh <server LAN IP or hostname> (not localhost)' >&2
  exit 1
fi
if [[ -e .env ]]; then
  echo '.env already exists and holds your credentials; refusing to overwrite it.' >&2
  exit 1
fi
umask 077
key=$(openssl genrsa 2048 2>/dev/null | base64 | tr -d '\n')
sed -e "s|192\.0\.2\.10|$host|" \
  -e "s|^POSTGRES_PASSWORD=$|POSTGRES_PASSWORD=$(openssl rand -hex 32)|" \
  -e "s|^JWT_SECRET=$|JWT_SECRET=$(openssl rand -hex 32)|" \
  -e "s|^AUTH_SECRET=$|AUTH_SECRET=$(openssl rand -hex 32)|" \
  -e "s|^STORAGE_PASSWORD=$|STORAGE_PASSWORD=$(openssl rand -hex 32)|" \
  -e "s|^PRIVATE_KEY=$|PRIVATE_KEY=$key|" \
  .env.example > .env
echo "Wrote .env for http://$host:3000. Next: docker compose up -d"
