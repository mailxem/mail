// Run with: bun test lib/email-check.test.ts
import { expect, test } from "bun:test";
import {
  checkDmarc,
  checkSpf,
  countSpfLookups,
  normalizeDomain,
  parseDmarc,
  parseSpf,
  unquoteTxt,
  type Resolve,
} from "./email-check";

const dns =
  (txt: Record<string, string[]>): Resolve =>
  async (name, type) =>
    type === "TXT" ? (txt[name] ?? []) : [];

test("unquoteTxt joins split TXT strings", () => {
  expect(unquoteTxt('"v=spf1 include:a.com" " -all"')).toBe(
    "v=spf1 include:a.com -all",
  );
  expect(unquoteTxt("v=DMARC1; p=none")).toBe("v=DMARC1; p=none");
});

test("normalizeDomain accepts URLs and addresses", () => {
  expect(normalizeDomain(" https://Example.com/pricing ")).toBe("example.com");
  expect(normalizeDomain("me@mail.example.co.uk")).toBe("mail.example.co.uk");
  expect(normalizeDomain("not a domain")).toBeNull();
});

test("parseSpf counts only lookup mechanisms", () => {
  const spf = parseSpf(
    "v=spf1 ip4:1.2.3.4 ip6:::1 a mx include:x.com exists:%{i}.y.com ptr ?all",
  );
  expect(spf.lookups).toBe(5);
  expect(spf.all).toBe("?");
  expect(spf.ptr).toBe(true);
  expect(spf.targets).toEqual(["x.com"]);
  expect(parseSpf("v=spf1 redirect=_spf.x.com").lookups).toBe(1);
  expect(parseSpf("v=spf1 all").all).toBe("+");
});

test("countSpfLookups follows includes and flags broken ones", async () => {
  const resolve = dns({
    "a.com": ["v=spf1 include:b.com include:c.com include:gone.com ~all"],
    "b.com": ["v=spf1 include:d.com mx -all"],
    "c.com": ["v=spf1 a a:x.c.com ~all"],
    "d.com": ["v=spf1 ip4:10.0.0.0/8 -all"],
    "gone.com": ["google-site-verification=abc"],
  });
  const { count, problems } = await countSpfLookups("a.com", resolve);
  // a.com 3 + b.com 2 + c.com 2 + d.com 0
  expect(count).toBe(7);
  expect(problems).toHaveLength(1);
  expect(problems[0]).toContain("gone.com");
});

test("checkSpf fails past 10 lookups and on +all", async () => {
  const many: Record<string, string[]> = {
    "big.com": [
      `v=spf1 ${Array.from({ length: 6 }, (_, i) => `include:i${i}.com`).join(" ")} +all`,
    ],
  };
  for (let i = 0; i < 6; i++) many[`i${i}.com`] = ["v=spf1 a ~all"];
  const findings = await checkSpf("big.com", dns(many));
  expect(findings.map((f) => f.title)).toEqual([
    "SPF ends in +all: anyone may send as you",
    "SPF needs 12 DNS lookups; the limit is 10",
  ]);
  expect(findings[0].fix).toEndWith("include:i5.com ~all");
});

test("parseDmarc reads tags and rejects other records", () => {
  expect(
    parseDmarc("v=DMARC1; p=quarantine; rua=mailto:r@x.com; pct=50"),
  ).toEqual({
    v: "DMARC1",
    p: "quarantine",
    rua: "mailto:r@x.com",
    pct: "50",
  });
  expect(parseDmarc("v=spf1 -all")).toBeNull();
});

test("checkDmarc reports missing records, rua and pct", async () => {
  const none = await checkDmarc("x.com", dns({}));
  expect(none[0].status).toBe("fail");
  expect(none[0].fix).toContain("_dmarc.x.com");
  const weak = await checkDmarc(
    "x.com",
    dns({ "_dmarc.x.com": ["v=DMARC1; p=reject; pct=25"] }),
  );
  expect(weak.map((f) => f.status)).toEqual(["pass", "warn", "warn"]);
});
