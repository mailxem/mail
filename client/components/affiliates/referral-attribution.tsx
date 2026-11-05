"use client";
import { useEffect, useRef } from "react";
import { useApi } from "@/hooks/use-api";

// xem.email stores the affiliate code from a `?via=` link in this cookie, on
// the parent domain so the app can read it. It expires 60 days after the click.
const COOKIE = "xem_via";

export const readReferral = (cookie: string) =>
  cookie.split("; ").find((c) => c.startsWith(`${COOKIE}=`))?.slice(COOKIE.length + 1) || "";

function clearReferral() {
  const parent = location.hostname.split(".").slice(-2).join(".");
  for (const domain of ["", `; domain=.${parent}`])
    document.cookie = `${COOKIE}=; path=/; max-age=0${domain}`;
}

/**
 * Credits a brand-new workspace to the affiliate whose link it came through.
 * The API only accepts workspaces created in the last day, so this is safe to
 * run on every visit; the cookie is cleared after the first try either way.
 */
export function ReferralAttribution() {
  const { apiFetch, status } = useApi();
  const tried = useRef(false);
  useEffect(() => {
    if (status !== "authenticated" || tried.current) return;
    const code = decodeURIComponent(readReferral(document.cookie));
    if (!code) return;
    tried.current = true;
    apiFetch("affiliates/attribute", { method: "POST", body: JSON.stringify({ code }) })
      .catch(() => {})
      .finally(clearReferral);
  }, [status, apiFetch]);
  return null;
}
