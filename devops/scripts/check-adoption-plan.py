#!/usr/bin/env python3
"""Fail closed on an adoption plan with creates/deletes, incomplete checks, or unreviewed updates.
Usage: terraform show -json adoption.tfplan | python3 scripts/check-adoption-plan.py
Plans may contain secrets: pipe them, do not commit or upload them.
"""
import argparse
import json
import sys


def violations(plan, allow_updates=False):
    if plan.get("errored") or plan.get("complete") is False or plan.get("deferred_changes"):
        yield "Plan is errored, incomplete, or deferred"
    if "resource_changes" not in plan and "planned_values" not in plan:
        yield "Input is not a Terraform plan"
    for check in plan.get("checks", []):
        if check.get("status") in ("fail", "error", "unknown"):
            yield "Plan contains failed or unknown checks"
    allowed = [["no-op"], ["read"]]
    if allow_updates:
        allowed.append(["update"])
    for change in plan.get("resource_changes", []):
        actions = change["change"]["actions"]
        if actions not in allowed:
            yield f"{change['address']}: {','.join(actions)}"


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--allow-updates", action="store_true", help="Only after reviewing every update; replacement is always rejected")
    args = parser.parse_args()
    problems = list(violations(json.load(sys.stdin), args.allow_updates))
    if problems:
        print("Adoption guard rejected the plan:\n" + "\n".join(problems), file=sys.stderr)
        sys.exit(1)
    print("Adoption guard passed; review the full plan before any apply.")
