// Deliverability checks for /tools/email-checker. Pure apart from the
// injected resolver, so tests run against a fake DNS map.

/** Returns the data of every answer of `type` for `name`; TXT data unquoted. */
export type Resolve = (name: string, type: "TXT" | "MX") => Promise<string[]>;

export type Finding = {
  area: "MX" | "SPF" | "DMARC" | "DKIM";
  status: "pass" | "warn" | "fail";
  title: string;
  detail: string;
  /** The record to add or change, ready to paste. */
  fix?: string;
};

const TYPES = { TXT: 16, MX: 15 } as const;

/** Joins the quoted strings of one TXT answer: `"v=spf1 a" " -all"` → `v=spf1 a -all`. */
export function unquoteTxt(data: string) {
  const parts = [...data.matchAll(/"((?:[^"\\]|\\.)*)"/g)].map((m) =>
    m[1].replace(/\\(.)/g, "$1"),
  );
  return parts.length ? parts.join("") : data;
}

/** DNS-over-HTTPS through Cloudflare's JSON API; runs in the browser. */
export const dohResolve: Resolve = async (name, type) => {
  const res = await fetch(
    `https://cloudflare-dns.com/dns-query?name=${encodeURIComponent(name)}&type=${type}`,
    { headers: { accept: "application/dns-json" } },
  );
  if (!res.ok) throw new Error(`DNS lookup for ${name} failed (${res.status})`);
  const json = (await res.json()) as {
    Status: number;
    Answer?: { type: number; data: string }[];
  };
  // 0 = found, 3 = no such name; anything else is a resolver failure.
  if (json.Status !== 0 && json.Status !== 3)
    throw new Error(`DNS lookup for ${name} failed (status ${json.Status})`);
  return (json.Answer ?? [])
    .filter((a) => a.type === TYPES[type])
    .map((a) => (type === "TXT" ? unquoteTxt(a.data) : a.data));
};

/** Accepts `example.com`, `https://example.com/x` or `me@example.com`. */
export function normalizeDomain(input: string) {
  let d = input.trim().toLowerCase();
  d = d.replace(/^[a-z]+:\/\//, "").split(/[/?#]/)[0];
  d = d.slice(d.lastIndexOf("@") + 1).replace(/\.$/, "");
  return /^(?=.{1,253}$)([a-z0-9-]{1,63}\.)+[a-z0-9-]{2,63}$/.test(d)
    ? d
    : null;
}

const isSpf = (r: string) => /^v=spf1(\s|$)/i.test(r);

/** Terms of one SPF record and the DNS lookups it costs by itself (RFC 7208 §4.6.4). */
export function parseSpf(record: string) {
  const terms = record.trim().split(/\s+/).slice(1);
  const targets: string[] = [];
  let lookups = 0;
  let all: string | null = null;
  let ptr = false;
  for (const term of terms) {
    const m = term.match(/^([+\-~?]?)([a-z0-9]+)(?:[:=](.*))?$/i);
    if (!m) continue;
    const [, qualifier, name, value] = m;
    const mech = name.toLowerCase();
    if (mech === "all") all = qualifier || "+";
    if (["include", "a", "mx", "ptr", "exists", "redirect"].includes(mech))
      lookups++;
    if (mech === "ptr") ptr = true;
    if ((mech === "include" || mech === "redirect") && value)
      targets.push(value);
  }
  const redirect = terms.some((t) => /^redirect=/i.test(t));
  return { lookups, targets, all, ptr, redirect };
}

/**
 * Total DNS lookups an SPF check of `domain` costs, following include and
 * redirect. Stops counting once past 20; the limit is 10.
 */
export async function countSpfLookups(
  domain: string,
  resolve: Resolve,
  seen = new Set<string>(),
): Promise<{ count: number; problems: string[] }> {
  let count = 0;
  const problems: string[] = [];
  const records = (await resolve(domain, "TXT")).filter(isSpf);
  if (records.length === 0) return { count, problems };
  const spf = parseSpf(records[0]);
  count += spf.lookups;
  for (const target of spf.targets) {
    if (count > 20) break;
    // Macros like %{i} expand per message, so the target can't be followed.
    if (target.includes("%")) continue;
    if (seen.has(target)) {
      problems.push(`${target} is included more than once or in a loop.`);
      continue;
    }
    seen.add(target);
    const own = (await resolve(target, "TXT")).filter(isSpf);
    if (own.length === 0) {
      problems.push(
        `${target} has no SPF record, so receivers fail the whole check (permerror).`,
      );
      continue;
    }
    const sub = await countSpfLookups(target, resolve, seen);
    count += sub.count;
    problems.push(...sub.problems);
  }
  return { count, problems };
}

/** Tags of a DMARC record, or null when it doesn't start with v=DMARC1. */
export function parseDmarc(record: string) {
  const tags: Record<string, string> = {};
  for (const part of record.split(";")) {
    const i = part.indexOf("=");
    if (i > 0)
      tags[part.slice(0, i).trim().toLowerCase()] = part.slice(i + 1).trim();
  }
  return /^v\s*=\s*DMARC1\s*(;|$)/i.test(record.trim()) ? tags : null;
}

export async function checkMx(domain: string, resolve: Resolve) {
  const mx = await resolve(domain, "MX");
  if (mx.length === 0)
    return {
      area: "MX",
      status: "fail",
      title: "No MX record",
      detail: `${domain} can't receive mail, so replies and bounces have nowhere to go, and some receivers reject mail from domains that can't take any back. Add the MX records your mailbox provider gives you.`,
    } satisfies Finding;
  if (mx.some((r) => /^0\s+\.$/.test(r.trim())))
    return {
      area: "MX",
      status: "warn",
      title: "Null MX: this domain says it accepts no mail",
      detail: `${domain} publishes "0 ." (RFC 7505), which tells everyone it never receives email. If you send from it, replies will bounce; send from a domain that can receive them.`,
    } satisfies Finding;
  return {
    area: "MX",
    status: "pass",
    title: `MX found: ${mx.length} mail server${mx.length > 1 ? "s" : ""}`,
    detail: mx.sort((a, b) => parseInt(a) - parseInt(b)).join(", "),
  } satisfies Finding;
}

export async function checkSpf(
  domain: string,
  resolve: Resolve,
): Promise<Finding[]> {
  const records = (await resolve(domain, "TXT")).filter(isSpf);
  const example = `${domain}  TXT  "v=spf1 include:<your sending provider> ~all"`;
  if (records.length === 0)
    return [
      {
        area: "SPF",
        status: "fail",
        title: "No SPF record",
        detail:
          "SPF lists the servers allowed to send as your domain. Gmail and Yahoo require SPF or DKIM for every sender, and both for bulk senders. Your sending provider's setup page has the exact include to use.",
        fix: example,
      },
    ];
  if (records.length > 1)
    return [
      {
        area: "SPF",
        status: "fail",
        title: `${records.length} SPF records; only one is allowed`,
        detail: `With more than one v=spf1 record, receivers treat SPF as broken (permerror). Merge them into a single record. Found: ${records.join(" | ")}`,
      },
    ];
  const record = records[0];
  const spf = parseSpf(record);
  const out: Finding[] = [];
  if (spf.all === "+")
    out.push({
      area: "SPF",
      status: "fail",
      title: "SPF ends in +all: anyone may send as you",
      detail: `"+all" (or a bare "all") passes every server on the internet, which makes the record worthless and looks like spam to filters. Your record: ${record}`,
      fix: record.replace(/\s\+?all\b/i, " ~all"),
    });
  else if (spf.all === "?")
    out.push({
      area: "SPF",
      status: "warn",
      title: "SPF ends in ?all (neutral)",
      detail: `"?all" says nothing about servers you didn't list, so it gives you no protection. Use ~all (soft fail) or -all (fail) once the record lists every service that sends for you. Your record: ${record}`,
      fix: record.replace(/\s\?all\b/i, " ~all"),
    });
  else if (!spf.all && !spf.redirect)
    out.push({
      area: "SPF",
      status: "warn",
      title: "SPF has no all at the end",
      detail: `Without an all mechanism, mail from unlisted servers is treated as neutral. End the record with ~all or -all. Your record: ${record}`,
      fix: `${record} ~all`,
    });
  if (spf.ptr)
    out.push({
      area: "SPF",
      status: "warn",
      title: "SPF uses ptr, which is deprecated",
      detail:
        "RFC 7208 says not to use ptr: it's slow and some receivers skip it. Replace it with the ip4, ip6 or include entries of your servers.",
    });
  const { count, problems } = await countSpfLookups(domain, resolve);
  if (count > 10)
    out.push({
      area: "SPF",
      status: "fail",
      title: `SPF needs ${count > 20 ? "more than 20" : count} DNS lookups; the limit is 10`,
      detail:
        "Each include, a, mx, ptr, exists and redirect costs a lookup, and so does everything inside each include. Past 10, receivers fail SPF outright (permerror). Remove services you no longer send through, or replace an include with the ip4/ip6 ranges it lists.",
    });
  for (const p of problems)
    out.push({
      area: "SPF",
      status: "fail",
      title: "Broken include",
      detail: p,
    });
  if (!out.some((f) => f.status !== "pass"))
    out.push({
      area: "SPF",
      status: "pass",
      title: `SPF is valid (${count} of 10 DNS lookups)`,
      detail: record,
    });
  return out;
}

export async function checkDmarc(
  domain: string,
  resolve: Resolve,
): Promise<Finding[]> {
  const name = `_dmarc.${domain}`;
  const starter = `${name}  TXT  "v=DMARC1; p=none; rua=mailto:dmarc-reports@${domain}"`;
  const records = (await resolve(name, "TXT")).filter((r) =>
    /^v\s*=\s*DMARC1/i.test(r.trim()),
  );
  if (records.length === 0)
    return [
      {
        area: "DMARC",
        status: "fail",
        title: "No DMARC record",
        detail:
          "Gmail requires a DMARC record from anyone sending 5,000 or more messages a day to Gmail accounts, and Yahoo requires one from bulk senders. Start with p=none, which changes nothing about delivery but sends you daily reports of who is mailing as your domain.",
        fix: starter,
      },
    ];
  if (records.length > 1)
    return [
      {
        area: "DMARC",
        status: "fail",
        title: `${records.length} DMARC records; only one is allowed`,
        detail: `With more than one, receivers ignore DMARC entirely. Keep one. Found: ${records.join(" | ")}`,
      },
    ];
  const record = records[0];
  const tags = parseDmarc(record);
  if (!tags)
    return [
      {
        area: "DMARC",
        status: "fail",
        title: "DMARC record isn't valid",
        detail: `It must start with exactly "v=DMARC1;". Your record: ${record}`,
        fix: starter,
      },
    ];
  const out: Finding[] = [];
  const p = tags.p?.toLowerCase();
  if (!p || !["none", "quarantine", "reject"].includes(p))
    out.push({
      area: "DMARC",
      status: "fail",
      title: p
        ? `Unknown DMARC policy p=${tags.p}`
        : "DMARC has no policy (p=)",
      detail: `The p tag must be none, quarantine or reject. Your record: ${record}`,
      fix: starter,
    });
  else if (p === "none")
    out.push({
      area: "DMARC",
      status: "pass",
      title: "DMARC is set to p=none (monitoring)",
      detail:
        "This meets Gmail's and Yahoo's requirement. It doesn't stop anyone spoofing your domain yet: once the reports show only your own services, move to p=quarantine, then p=reject.",
    });
  else
    out.push({
      area: "DMARC",
      status: "pass",
      title: `DMARC policy is p=${p}`,
      detail: `Mail that fails both SPF and DKIM alignment is ${p === "reject" ? "rejected" : "sent to spam"}. Your record: ${record}`,
    });
  if (!tags.rua)
    out.push({
      area: "DMARC",
      status: "warn",
      title: "No rua: you get no DMARC reports",
      detail:
        "Without a rua address, receivers don't tell you which servers send as your domain or whether they pass. Add one (a dedicated mailbox or a report service).",
      fix: `${name}  TXT  "${record.replace(/;?\s*$/, "")}; rua=mailto:dmarc-reports@${domain}"`,
    });
  const pct = Number(tags.pct ?? 100);
  if (pct < 100)
    out.push({
      area: "DMARC",
      status: "warn",
      title: `Policy applies to only ${pct}% of mail`,
      detail:
        "pct below 100 is meant for a gradual rollout. Raise it to 100 (or remove the tag) when you're done.",
    });
  return out;
}

export const commonSelectors = [
  { selector: "google", who: "Google Workspace" },
  { selector: "selector1", who: "Microsoft 365" },
  { selector: "selector2", who: "Microsoft 365" },
  { selector: "k1", who: "" },
  { selector: "s1", who: "SendGrid" },
  { selector: "s2", who: "SendGrid" },
  { selector: "resend", who: "Resend" },
];

/** Bytes of a base64 DKIM public key; 1024-bit RSA keys are about 162. */
const keyBytes = (p: string) =>
  Math.floor((p.replace(/[^A-Za-z0-9+/]/g, "").length * 3) / 4);

export async function checkDkim(
  domain: string,
  selector: string,
  resolve: Resolve,
): Promise<Finding> {
  const name = `${selector}._domainkey.${domain}`;
  const record = (await resolve(name, "TXT")).find((r) =>
    /(^|;)\s*p=/i.test(r),
  );
  if (!record)
    return {
      area: "DKIM",
      status: "warn",
      title: `No DKIM key at ${name}`,
      detail:
        "Either this selector isn't the one your provider signs with, or the key isn't published. Find the selector in a message you sent: open its headers, look for DKIM-Signature, and read the s= value.",
    };
  const tags = parseDmarc(`v=DMARC1;${record}`) ?? {};
  if (!tags.p)
    return {
      area: "DKIM",
      status: "fail",
      title: `DKIM key at ${name} is revoked (empty p=)`,
      detail:
        "An empty p= means the key was withdrawn, so any signature with this selector fails. Publish the current key from your provider, or sign with another selector.",
    };
  if ((tags.k ?? "rsa").toLowerCase() === "rsa" && keyBytes(tags.p) < 200)
    return {
      area: "DKIM",
      status: "warn",
      title: `DKIM key at ${name} is 1024-bit`,
      detail:
        "It works, but Google recommends 2048-bit keys. Rotate to a 2048-bit key in your provider's settings when you can.",
    };
  return {
    area: "DKIM",
    status: "pass",
    title: `DKIM key found at ${name}`,
    detail:
      "The public key is published. Whether messages pass also depends on your provider signing with this selector and your domain.",
  };
}
