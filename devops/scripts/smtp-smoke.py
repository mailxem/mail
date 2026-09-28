#!/usr/bin/env python3
"""Check STARTTLS, public trust, and relay rejection. Never sends an email."""
import argparse
import base64
import smtplib
import ssl


def require(condition, message):
    if not condition:
        raise RuntimeError(message)


def check(host, port):
    with smtplib.SMTP(host, port, timeout=15) as smtp:
        require(smtp.ehlo()[0] == 250, "EHLO failed")
        require(smtp.has_extn("starttls"), "STARTTLS is missing")
        require(not smtp.has_extn("auth"), "AUTH must not be advertised before TLS")
        # Deliberately invalid, non-secret credentials; no user identity is used.
        dummy = base64.b64encode(b"\0xem-transport-smoke-invalid\0invalid").decode()
        code, _ = smtp.docmd("AUTH", "PLAIN " + dummy)
        require(code in (523, 538, 530), f"Pre-TLS AUTH was not rejected securely: {code}")
        smtp.starttls(context=ssl.create_default_context())
        require(smtp.ehlo()[0] == 250, "Post-TLS EHLO failed")
        require("PLAIN" in smtp.esmtp_features.get("auth", "").upper(), "AUTH PLAIN missing after TLS")
        code, _ = smtp.mail("transport-smoke@example.invalid")
        require(code in (502, 530), f"Unauthenticated MAIL FROM was not rejected: {code}")
        peer = smtp.sock.getpeercert()
        print(f"PASS {host}:{port}: trusted hostname, {smtp.sock.version()}, STARTTLS required, unauthenticated relay rejected")
        print(f"Certificate expires: {peer['notAfter']}")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("host")
    parser.add_argument("--port", type=int, default=587)
    args = parser.parse_args()
    check(args.host, args.port)
