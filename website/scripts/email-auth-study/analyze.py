"""Summarises results.jsonl into the numbers the report page shows."""
import json, re, sys

rows = [json.loads(l) for l in open(sys.argv[1])]
rows.sort(key=lambda r: r["rank"])

def tag(rec, key):
    m = re.search(r"(?:^|;)\s*" + key + r"\s*=\s*([^;\s]+)", rec, re.I)
    return m.group(1).lower() if m else None

def spf_all(rec):
    m = re.search(r"\s([-~?+]?)all\b", rec.lower())
    if not m:
        return "redirect" if "redirect=" in rec.lower() else "none"
    return {"-": "-all", "~": "~all", "?": "?all", "+": "+all", "": "+all"}[m.group(1)]

def pct(a, b):
    return round(100 * a / b, 1) if b else 0.0

def summarise(group):
    ok = lambda r, k: k not in (r.get("errors") or [])
    spf_g = [r for r in group if ok(r, "spf")]
    dm_g = [r for r in group if ok(r, "dmarc")]
    has_spf = [r for r in spf_g if r["spf"]]
    has_dm = [r for r in dm_g if r["dmarc"]]
    pol = {"none": 0, "quarantine": 0, "reject": 0, "invalid": 0}
    rua = 0
    partial = 0
    for r in has_dm:
        if len(r["dmarc"]) > 1:
            pol["invalid"] += 1
            continue
        p = tag(r["dmarc"][0], "p")
        pol[p if p in pol else "invalid"] += 1
        if tag(r["dmarc"][0], "rua"):
            rua += 1
        if p in ("quarantine", "reject") and (tag(r["dmarc"][0], "pct") or "100") != "100":
            partial += 1
    alls = {}
    for r in has_spf:
        if len(r["spf"]) > 1:
            alls["multiple"] = alls.get("multiple", 0) + 1
            continue
        k = spf_all(r["spf"][0])
        alls[k] = alls.get(k, 0) + 1
    enforced = pol["quarantine"] + pol["reject"]
    both = [r for r in group if ok(r, "spf") and ok(r, "dmarc")]
    return {
        "domains": len(group),
        "spf": pct(len(has_spf), len(spf_g)),
        "spfAll": {k: pct(v, len(has_spf)) for k, v in sorted(alls.items(), key=lambda x: -x[1])},
        "dmarc": pct(len(has_dm), len(dm_g)),
        "dmarcPolicy": {k: pct(v, len(dm_g)) for k, v in pol.items()},
        "dmarcEnforced": pct(enforced, len(dm_g)),
        "dmarcReject": pct(pol["reject"], len(dm_g)),
        "dmarcReports": pct(rua, len(has_dm)),
        "dmarcPartial": pct(partial, enforced),
        "spfAndDmarc": pct(sum(1 for r in both if r["spf"] and r["dmarc"]), len(both)),
        "bimi": pct(sum(1 for r in group if r["bimi"] and ok(r, "bimi")), sum(1 for r in group if ok(r, "bimi"))),
        "mtaSts": pct(sum(1 for r in group if r["sts"] and ok(r, "sts")), sum(1 for r in group if ok(r, "sts"))),
    }

mail = [r for r in rows if r["mx"]]
nomail = [r for r in rows if not r["mx"] and "mx" not in (r.get("errors") or [])]
locked = [r for r in nomail if r["spf"] and len(r["spf"]) == 1 and spf_all(r["spf"][0]) == "-all"
          and r["dmarc"] and len(r["dmarc"]) == 1 and tag(r["dmarc"][0], "p") == "reject"]
out = {
    "scanned": len(rows),
    "errors": sum(1 for r in rows if r.get("errors")),
    "receivesMail": len(mail),
    "all": summarise(rows),
    "mail": summarise(mail),
    "tiers": {str(t): summarise([r for r in mail if r["rank"] <= t]) for t in (100, 1000, 10000)},
    "noMail": {"domains": len(nomail), "nullMx": sum(1 for r in nomail if r["nullMx"]), "lockedDown": pct(len(locked), len(nomail)),
               "dmarc": pct(sum(1 for r in nomail if r["dmarc"]), len(nomail))},
}
print(json.dumps(out, indent=2))
