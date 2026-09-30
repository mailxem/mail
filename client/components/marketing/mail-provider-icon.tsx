import { Mail } from "lucide-react";
import { SiCloudflare, SiGmail, SiIcloud } from "react-icons/si";

export type MailProviderKind = "managed" | "gmail" | "icloud" | "cloudflare" | "imap";

export type MailProviderIdentity = {
  kind: MailProviderKind;
  label: "Xem inbox" | "Gmail" | "iCloud Mail" | "Cloudflare" | "IMAP";
};

const GMAIL_HOSTS = new Set(["imap.gmail.com", "imap.googlemail.com"]);
const ICLOUD_HOSTS = new Set(["imap.mail.me.com", "imap.mail.icloud.com"]);

export function inferMailProvider({
  provider,
  host,
}: {
  provider?: string | null;
  host?: string | null;
}): MailProviderIdentity {
  const explicit = provider?.trim().toUpperCase();
  const canonicalHost = host?.trim().toLowerCase().replace(/\.$/, "");

  if (explicit === "MANAGED") return { kind: "managed", label: "Xem inbox" };
  if (explicit === "GOOGLE_OAUTH" || explicit === "GMAIL")
    return { kind: "gmail", label: "Gmail" };
  if (explicit === "ICLOUD" || explicit === "APPLE")
    return { kind: "icloud", label: "iCloud Mail" };
  // Cloudflare Email Sending is not an inbox provider. Only identify it when
  // trusted provider metadata says so; never infer it from an arbitrary host.
  if (explicit === "CLOUDFLARE")
    return { kind: "cloudflare", label: "Cloudflare" };
  if (canonicalHost && GMAIL_HOSTS.has(canonicalHost))
    return { kind: "gmail", label: "Gmail" };
  if (canonicalHost && ICLOUD_HOSTS.has(canonicalHost))
    return { kind: "icloud", label: "iCloud Mail" };
  return { kind: "imap", label: "IMAP" };
}

export function MailProviderIcon({
  provider,
  host,
  className = "",
  showLabel = false,
}: {
  provider?: string | null;
  host?: string | null;
  className?: string;
  showLabel?: boolean;
}) {
  const identity = inferMailProvider({ provider, host });
  const Icon =
    identity.kind === "managed"
      ? Mail
      : identity.kind === "gmail"
      ? SiGmail
      : identity.kind === "icloud"
        ? SiIcloud
        : identity.kind === "cloudflare"
          ? SiCloudflare
          : Mail;
  const color =
    identity.kind === "managed"
      ? "#7C3AED"
      : identity.kind === "gmail"
      ? "#EA4335"
      : identity.kind === "icloud"
        ? "#1687F8"
        : identity.kind === "cloudflare"
          ? "#F48120"
          : "#66695E";

  return (
    <span
      className={`inline-flex min-w-0 items-center gap-2 ${className}`.trim()}
      title={identity.label}
    >
      <span
        role="img"
        aria-label={`${identity.label} provider`}
        className="inline-flex size-7 shrink-0 items-center justify-center rounded-md border border-black/10 bg-white shadow-sm dark:border-white/15 dark:bg-white"
      >
        <Icon aria-hidden="true" focusable="false" size={15} color={color} />
      </span>
      {showLabel && <span className="truncate">{identity.label}</span>}
    </span>
  );
}
