"use client";
import { useEffect } from "react";
import { apiUrl } from "@/lib/site";

const SIXTY_DAYS = 60 * 24 * 60 * 60;
let counted = false; // effects run twice in development; count the click once

/**
 * Remembers the affiliate code from a `?via=` link for 60 days (the last link
 * clicked wins) and counts the click. The cookie sits on the parent domain so
 * app.xem.email can credit the sign-up. Nothing about the visitor is sent.
 */
export function ReferralCapture() {
  useEffect(() => {
    const code = new URLSearchParams(location.search)
      .get("via")
      ?.trim()
      .toLowerCase();
    if (counted || !code || !/^[a-z0-9]{4,32}$/.test(code)) return;
    counted = true;
    const parent = location.hostname.endsWith("xem.email")
      ? "; domain=.xem.email"
      : "";
    document.cookie = `xem_via=${code}; path=/; max-age=${SIXTY_DAYS}; samesite=lax${parent}${location.protocol === "https:" ? "; secure" : ""}`;
    // A beacon needs no CORS preflight and survives the visitor leaving the page.
    navigator.sendBeacon?.(
      `${apiUrl}/api/v1/affiliates/clicks`,
      JSON.stringify({ code }),
    );
  }, []);
  return null;
}
