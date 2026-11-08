# Email authentication study

Produces `lib/research/email-auth-2026.json`, the data behind
`/research/email-authentication`.

1. Download a Tranco list (the 2026 study used list `94P72`) and unzip it to `top-1m.csv`.
2. Scan the top 10,000 domains: `go run . top-1m.csv 10000 > results.jsonl`
3. Rescan domains with failed lookups more slowly: write their `rank,domain`
   lines to `retry.csv`, run `WORKERS=12 go run . retry.csv 10000 > retry.jsonl`,
   and keep each domain's result with fewer errors.
4. Summarise: `python3 analyze.py merged.jsonl`, then wrap the output in the
   metadata fields of `lib/research/email-auth-2026.json`.

Only public DNS is read; no mail is sent.
