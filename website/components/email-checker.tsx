"use client";
import { useId, useState } from "react";
import { ArrowUpRight, Check, CircleAlert, CircleX } from "lucide-react";
import {
  checkDkim,
  checkDmarc,
  checkMx,
  checkSpf,
  commonSelectors,
  dohResolve,
  normalizeDomain,
  type Finding,
} from "@/lib/email-check";
import { appLink } from "@/lib/site";

const icon = {
  pass: <Check size={18} className="text-[#2f6b3d]" aria-label="Pass" />,
  warn: (
    <CircleAlert size={18} className="text-[#a3620a]" aria-label="Warning" />
  ),
  fail: <CircleX size={18} className="text-[#b3261e]" aria-label="Problem" />,
};

/** Domain form and report; every lookup goes from the browser to Cloudflare DNS. */
export function EmailChecker() {
  const id = useId();
  const [domain, setDomain] = useState("");
  const [selector, setSelector] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [report, setReport] = useState<{
    domain: string;
    findings: Finding[];
  } | null>(null);

  async function run(e: React.FormEvent) {
    e.preventDefault();
    const d = normalizeDomain(domain);
    if (!d) {
      setError("Enter a domain like example.com.");
      return;
    }
    setBusy(true);
    setError("");
    try {
      const sel = selector.trim().toLowerCase();
      // With no selector, try the common ones and report only those found.
      const dkim = sel
        ? [await checkDkim(d, sel, dohResolve)]
        : (
            await Promise.all(
              commonSelectors.map((s) => checkDkim(d, s.selector, dohResolve)),
            )
          ).filter((f) => !f.title.startsWith("No DKIM key"));
      const [mx, spf, dmarc] = await Promise.all([
        checkMx(d, dohResolve),
        checkSpf(d, dohResolve),
        checkDmarc(d, dohResolve),
      ]);
      if (!sel && dkim.length === 0)
        dkim.push({
          area: "DKIM",
          status: "warn",
          title: "No DKIM key at the common selectors",
          detail: `None of ${commonSelectors.map((s) => s.selector).join(", ")} has a key on ${d}. That's expected if your provider uses its own selector (Amazon SES always does). Find yours in a message you sent: open its headers, look for DKIM-Signature, and enter the s= value above.`,
        });
      setReport({ domain: d, findings: [mx, ...spf, ...dmarc, ...dkim] });
    } catch (err) {
      setError(
        `${err instanceof Error ? err.message : "DNS lookup failed"}. Check your connection and try again.`,
      );
    } finally {
      setBusy(false);
    }
  }

  const problems = report?.findings.filter((f) => f.status !== "pass").length;
  return (
    <div>
      <form
        onSubmit={run}
        className="grid gap-4 rounded-2xl border border-ink/15 bg-cream p-6 md:grid-cols-[1.4fr_1fr_auto] md:items-end md:p-8"
      >
        <div>
          <label htmlFor={`${id}-domain`} className="text-sm font-medium">
            Sending domain
          </label>
          <input
            id={`${id}-domain`}
            value={domain}
            onChange={(e) => setDomain(e.target.value)}
            placeholder="example.com"
            autoComplete="off"
            autoCapitalize="none"
            spellCheck={false}
            required
            className="mt-2 w-full rounded-lg border border-ink/20 bg-white px-4 py-3 text-base outline-none focus:border-iris"
          />
        </div>
        <div>
          <label htmlFor={`${id}-selector`} className="text-sm font-medium">
            DKIM selector{" "}
            <span className="font-normal text-muted">(optional)</span>
          </label>
          <input
            id={`${id}-selector`}
            value={selector}
            onChange={(e) => setSelector(e.target.value)}
            placeholder="tries common ones"
            list={`${id}-selectors`}
            autoComplete="off"
            autoCapitalize="none"
            spellCheck={false}
            className="mt-2 w-full rounded-lg border border-ink/20 bg-white px-4 py-3 text-base outline-none focus:border-iris"
          />
          <datalist id={`${id}-selectors`}>
            {commonSelectors.map((s) => (
              <option key={s.selector} value={s.selector}>
                {s.who}
              </option>
            ))}
          </datalist>
        </div>
        <button
          type="submit"
          disabled={busy}
          className="inline-flex min-h-12 items-center justify-center rounded-lg border border-[#49309e] bg-iris px-6 py-3 text-sm font-semibold text-white shadow-[0_2px_0_#362278] hover:bg-[#49309e] disabled:opacity-60"
        >
          {busy ? "Checking…" : "Check domain"}
        </button>
      </form>

      <div aria-live="polite">
        {error && (
          <p role="alert" className="mt-6 text-sm text-[#b3261e]">
            {error}
          </p>
        )}
        {report && (
          <section aria-label={`Report for ${report.domain}`} className="mt-10">
            <h2 className="font-editorial text-3xl tracking-tight md:text-4xl">
              {report.domain}:{" "}
              {problems
                ? `${problems} thing${problems > 1 ? "s" : ""} to look at.`
                : "all checks pass."}
            </h2>
            <ul className="mt-6 divide-y divide-ink/10 rounded-2xl border border-ink/15 bg-cream">
              {report.findings.map((f, i) => (
                <li
                  key={i}
                  className="grid grid-cols-[auto_1fr] gap-x-4 p-5 md:p-6"
                >
                  <span className="mt-0.5">{icon[f.status]}</span>
                  <div className="min-w-0">
                    <p className="text-sm font-medium">
                      <span className="mr-2 text-xs uppercase tracking-wider text-muted">
                        {f.area}
                      </span>
                      {f.title}
                    </p>
                    <p className="mt-2 break-words text-sm leading-relaxed text-muted">
                      {f.detail}
                    </p>
                    {f.fix && (
                      <pre className="mt-3 overflow-x-auto rounded-lg bg-ink px-4 py-3 text-xs text-cream">
                        <code>{f.fix}</code>
                      </pre>
                    )}
                  </div>
                </li>
              ))}
            </ul>
            <p className="mt-8 text-sm text-muted">
              Xem walks you through these records when you add a sending domain.{" "}
              <a
                href={appLink("/auth/register")}
                className="inline-flex items-center gap-1 text-iris underline underline-offset-4"
              >
                Try Xem free <ArrowUpRight size={14} />
              </a>
            </p>
          </section>
        )}
      </div>
    </div>
  );
}
