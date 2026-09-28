#!/usr/bin/env bash
# Run on the Dokploy host AFTER reviewing docs/dokploy.md. Never run in CI.
set -euo pipefail
[[ $EUID -eq 0 ]] || { echo "Run this installer as root on the Dokploy host." >&2; exit 1; }
command -v openssl >/dev/null
command -v python3 >/dev/null
command -v systemctl >/dev/null
script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
[[ -f /etc/dokploy/traefik/dynamic/acme.json ]] || { echo "Traefik ACME storage is missing." >&2; exit 1; }
# Parent ownership matters: recursive mkdir under umask 077 otherwise creates /etc/xem as 0700.
install -d -o root -g 65532 -m 0750 /etc/xem /etc/xem/smtp
if mountpoint -q /etc/xem/smtp-readonly; then
  # Do not chmod/chown an already mounted read-only directory on repeated runs.
  [[ "$(stat -c '%d:%i' /etc/xem/smtp)" == "$(stat -c '%d:%i' /etc/xem/smtp-readonly)" ]] || {
    echo "Existing certificate mirror points at a different directory; inspect it before continuing." >&2
    exit 1
  }
else
  install -d -o root -g 65532 -m 0750 /etc/xem/smtp-readonly
fi
install -o root -g root -m 0755 "$script_dir/export-smtp-certificate.py" /usr/local/sbin/xem-export-smtp-cert
install -o root -g root -m 0644 "$script_dir/../host/xem-smtp-certificate.yml" /etc/dokploy/traefik/dynamic/xem-smtp-certificate.yml
install -o root -g root -m 0644 "$script_dir/../host/xem-smtp-certificate.service" /etc/systemd/system/xem-smtp-certificate.service
install -o root -g root -m 0644 "$script_dir/../host/xem-smtp-certificate.timer" /etc/systemd/system/xem-smtp-certificate.timer
mount_unit=$(systemd-escape --path --suffix=mount /etc/xem/smtp-readonly)
install -o root -g root -m 0644 "$script_dir/../host/smtp-readonly.mount" "/etc/systemd/system/$mount_unit"
systemctl daemon-reload
systemctl enable --now xem-smtp-certificate.timer
# Fail closed until Traefik has issued a trusted certificate. The timer retries.
systemctl start xem-smtp-certificate.service
systemctl enable --now "$mount_unit"
# Confirm the mirror really is read-only, including when adopting an existing mount.
findmnt -n -o OPTIONS /etc/xem/smtp-readonly | tr ',' '\n' | grep -qx ro
printf '%s\n' 'Certificate export and read-only mirror ready. Configure the persistent Dokploy mount and TCP port as documented.'
