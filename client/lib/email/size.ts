// Gmail clips messages whose HTML is over about 102 KB and hides the rest,
// including the unsubscribe link and the open pixel.
const KB = 1000;
export const GMAIL_CLIP_BYTES = 102 * KB;
export const GMAIL_WARN_BYTES = 90 * KB;

// What sending adds (server/internal/utils/parser.go, marketing/service.go):
// each http(s) link becomes /t/click/<base64 target>?token=<signed JWT>.
const TRACKED_LINK_BYTES = 220;
// Unsubscribe and postal-address footer, open pixel, "Sent with Xem" badge.
// ponytail: badge counted for every plan; ~400 bytes, not worth a plan lookup.
const SEND_EXTRA_BYTES = 1200;

export type GmailClipState = "ok" | "close" | "clipped";

/** Estimated size in bytes of the HTML as it will reach the inbox. */
export function sentEmailBytes(html: string) {
  const links = html.match(/<a\b[^>]*\bhref\s*=\s*["']?https?:\/\//gi) ?? [];
  return (
    new TextEncoder().encode(html).length +
    links.length * TRACKED_LINK_BYTES +
    SEND_EXTRA_BYTES
  );
}

export function gmailClipState(bytes: number): GmailClipState {
  if (bytes > GMAIL_CLIP_BYTES) return "clipped";
  if (bytes >= GMAIL_WARN_BYTES) return "close";
  return "ok";
}

export function formatKB(bytes: number) {
  return `${Math.round(bytes / KB)} KB`;
}
