#!/usr/bin/python3
"""Export only the SMTP certificate from Traefik, without exposing ACME keys."""
import base64
import fcntl
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile

HOST = os.environ.get("XEM_SMTP_HOST", "smtp.xem.email")
SOURCE = Path("/etc/dokploy/traefik/dynamic/acme.json")
DEST = Path("/etc/xem/smtp")
LOCK = Path("/run/lock/xem-smtp-certificate.lock")
TRUST_STORE = "/etc/ssl/certs/ca-certificates.crt"


def run(args, data=None):
    return subprocess.run(args, input=data, check=True, capture_output=True).stdout


def directory(path):
    path.mkdir(parents=True, exist_ok=True)
    os.chown(path, 0, 65532)
    os.chmod(path, 0o750)


def link(target, name):
    path = DEST / name
    if path.is_symlink() and os.readlink(path) == target:
        return
    if path.exists() and not path.is_symlink():
        raise RuntimeError(f"Refusing to replace non-symlink {path}")
    temporary = DEST / ("." + name + ".next")
    temporary.unlink(missing_ok=True)
    temporary.symlink_to(target)
    os.replace(temporary, path)


def main():
    os.umask(0o077)
    with open(LOCK, "w") as lock:
        fcntl.flock(lock, fcntl.LOCK_EX)
        candidates = []
        for resolver in json.loads(SOURCE.read_text()).values():
            for cert in resolver.get("Certificates", []):
                domain = cert.get("domain", {})
                if domain.get("main") == HOST and not domain.get("sans"):
                    candidates.append(cert)
        if len(candidates) != 1:
            raise RuntimeError("Expected one dedicated SMTP certificate")
        cert = base64.b64decode(candidates[0]["certificate"], validate=True)
        key = base64.b64decode(candidates[0]["key"], validate=True)
        directory(DEST)
        directory(DEST / "versions")
        revision = hashlib.sha256(cert + key).hexdigest()
        version = DEST / "versions" / revision
        with tempfile.TemporaryDirectory(prefix=".check-", dir=DEST) as work:
            chain = Path(work) / "fullchain.pem"
            private = Path(work) / "privkey.pem"
            leaf = Path(work) / "cert.pem"
            chain.write_bytes(cert)
            private.write_bytes(key)
            leaf.write_bytes(run(["openssl", "x509", "-in", str(chain), "-outform", "PEM"]))
            run(["openssl", "x509", "-in", str(chain), "-noout", "-checkhost", HOST, "-checkend", "86400"])
            run(["openssl", "verify", "-purpose", "sslserver", "-verify_hostname", HOST,
                 "-CAfile", TRUST_STORE, "-untrusted", str(chain), str(leaf)])
            cert_public = run(["openssl", "x509", "-in", str(chain), "-noout", "-pubkey"])
            key_public = run(["openssl", "pkey", "-in", str(private), "-pubout"])
            if cert_public != key_public:
                raise RuntimeError("Certificate and private key do not match")
            if not version.exists():
                staged = Path(tempfile.mkdtemp(prefix=".version-", dir=DEST / "versions"))
                try:
                    directory(staged)
                    for name, contents in [("fullchain.pem", cert), ("privkey.pem", key)]:
                        item = staged / name
                        item.write_bytes(contents)
                        os.chown(item, 0, 65532)
                        os.chmod(item, 0o640)
                    os.rename(staged, version)
                finally:
                    if staged.exists():
                        shutil.rmtree(staged)
        changed = not (DEST / "current").is_symlink() or os.readlink(DEST / "current") != f"versions/{revision}"
        link(f"versions/{revision}", "current")
        link("current/fullchain.pem", "fullchain.pem")
        link("current/privkey.pem", "privkey.pem")
        if changed:
            print("Validated and installed SMTP certificate; private key remains on this host")


if __name__ == "__main__":
    main()
