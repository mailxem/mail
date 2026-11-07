"use client";
import { useEffect, useState } from "react";
import { ArrowUpRight, Zap } from "lucide-react";
import { useMarketing } from "@/lib/marketing/api";
import type { SendingDomain } from "@/lib/sending/types";

type DNSHost = {
  id: "cloudflare" | "godaddy" | "namecheap";
  name: string;
  zone: string;
  detected: boolean;
  connect?: string;
  manage: string;
};

// Brand marks from Simple Icons (CC0), in each registrar's own colours.
const BRAND: Record<DNSHost["id"], { path: string; bg: string; fg: string }> = {
  cloudflare: {
    path: "M16.5088 16.8447c.1475-.5068.0908-.9707-.1553-1.3154-.2246-.3164-.6045-.499-1.0615-.5205l-8.6592-.1123a.1559.1559 0 0 1-.1333-.0713c-.0283-.042-.0351-.0986-.021-.1553.0278-.084.1123-.1484.2036-.1562l8.7359-.1123c1.0351-.0489 2.1601-.8868 2.5537-1.9136l.499-1.3013c.0215-.0561.0293-.1128.0147-.168-.5625-2.5463-2.835-4.4453-5.5499-4.4453-2.5039 0-4.6284 1.6177-5.3876 3.8614-.4927-.3658-1.1187-.5625-1.794-.499-1.2026.119-2.1665 1.083-2.2861 2.2856-.0283.31-.0069.6128.0635.894C1.5683 13.171 0 14.7754 0 16.752c0 .1748.0142.3515.0352.5273.0141.083.0844.1475.1689.1475h15.9814c.0909 0 .1758-.0645.2032-.1553l.12-.4268zm2.7568-5.5634c-.0771 0-.1611 0-.2383.0112-.0566 0-.1054.0415-.127.0976l-.3378 1.1744c-.1475.5068-.0918.9707.1543 1.3164.2256.3164.6055.498 1.0625.5195l1.8437.1133c.0557 0 .1055.0263.1329.0703.0283.043.0351.1074.0214.1562-.0283.084-.1132.1485-.204.1553l-1.921.1123c-1.041.0488-2.1582.8867-2.5527 1.914l-.1406.3585c-.0283.0713.0215.1416.0986.1416h6.5977c.0771 0 .1474-.0489.169-.126.1122-.4082.1757-.837.1757-1.2803 0-2.6025-2.125-4.727-4.7344-4.727",
    bg: "#F38020",
    fg: "#ffffff",
  },
  godaddy: {
    path: "M20.702 2.29c-2.494-1.554-5.778-1.187-8.706.654C9.076 1.104 5.79.736 3.3 2.29c-3.941 2.463-4.42 8.806-1.07 14.167 2.47 3.954 6.333 6.269 9.77 6.226 3.439.043 7.301-2.273 9.771-6.226 3.347-5.361 2.872-11.704-1.069-14.167zM4.042 15.328a12.838 12.838 0 01-1.546-3.541 10.12 10.12 0 01-.336-3.338c.15-1.98.956-3.524 2.27-4.345 1.315-.822 3.052-.87 4.903-.137.281.113.556.24.825.382A15.11 15.11 0 007.5 7.54c-2.035 3.255-2.655 6.878-1.945 9.765a13.247 13.247 0 01-1.514-1.98zm17.465-3.541a12.866 12.866 0 01-1.547 3.54 13.25 13.25 0 01-1.513 1.984c.635-2.589.203-5.76-1.353-8.734a.39.39 0 00-.563-.153l-4.852 3.032a.397.397 0 00-.126.546l.712 1.139a.395.395 0 00.547.126l3.145-1.965c.101.306.203.606.28.916.296 1.086.41 2.214.335 3.337-.15 1.982-.956 3.525-2.27 4.347a4.437 4.437 0 01-2.25.65h-.101a4.432 4.432 0 01-2.25-.65c-1.314-.822-2.121-2.365-2.27-4.347-.074-1.123.039-2.251.335-3.337a13.212 13.212 0 014.05-6.482 10.148 10.148 0 012.849-1.765c1.845-.733 3.586-.685 4.9.137 1.316.822 2.122 2.365 2.271 4.345a10.146 10.146 0 01-.33 3.334z",
    bg: "#1BDBDB",
    fg: "#111111",
  },
  namecheap: {
    path: "M17.295 17.484c.227.403.57.728.985.931-.309.15-.647.229-.99.232h-3.068a2.26 2.26 0 0 1-1.957-1.143L6.705 6.511a2.27 2.27 0 0 0-.974-.922c.309-.153.652-.233.997-.232h3.05c.81.003 1.558.438 1.959 1.143l5.558 10.984zm-9.329-7.392L6.269 6.755c-.209-.392-.582-.657-.984-.829-.204.165-.391.35-.522.581-.184.349-4.391 8.648-4.569 8.987a2.245 2.245 0 0 0 4.016 1.999l3.756-7.401zm15.846-1.593a2.245 2.245 0 0 0-1.162-2.955v-.001a2.243 2.243 0 0 0-.892-.187l-.003-.011c-.816 0-1.569.443-1.965 1.157l-3.749 7.414 1.689 3.323c.213.399.59.664.998.839.252-.2.473-.444.605-.742l4.479-8.837z",
    bg: "#DE3723",
    fg: "#ffffff",
  },
};

export function DNSHostButtons({ domain }: { domain: SendingDomain }) {
  const { request } = useMarketing();
  const [hosts, setHosts] = useState<DNSHost[]>([]);
  useEffect(() => {
    request<DNSHost[]>(`sending/domains/${domain.id}/dns-hosts`)
      .then(setHosts)
      .catch(() => setHosts([]));
    // request is rebuilt every render; refetch only when the domain changes.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [domain.id, domain.provisioned]);
  if (!hosts.length) return null;
  const detected = hosts.find((h) => h.detected);
  const ordered = detected
    ? [detected, ...hosts.filter((h) => h !== detected)]
    : hosts;
  return (
    <section
      aria-label="Add records at your DNS provider"
      className="space-y-3 rounded-xl border bg-background p-4"
    >
      <div>
        <h3 className="text-sm font-semibold">
          {detected?.connect
            ? `Add every record with one click`
            : "Open your DNS provider"}
        </h3>
        <p className="mt-1 text-xs leading-5 text-muted-foreground">
          {detected?.connect
            ? `${detected.name} adds the records for ${detected.zone} after you approve. You come straight back here.`
            : detected
              ? `${detected.zone} uses ${detected.name}. Open its DNS page and paste the records below.`
              : `Pick where you manage DNS for ${ordered[0].zone} and paste the records below.`}
        </p>
      </div>
      <div className="flex flex-wrap gap-2">
        {ordered.map((h) => {
          const b = BRAND[h.id];
          const oneClick = !!h.connect;
          return (
            <a
              key={h.id}
              href={h.connect || h.manage}
              target={oneClick ? undefined : "_blank"}
              rel="noreferrer"
              style={{ backgroundColor: b.bg, color: b.fg }}
              className={`inline-flex h-10 items-center gap-2 rounded-lg px-4 text-sm font-semibold shadow-sm transition hover:brightness-95 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 ${detected && !h.detected ? "opacity-70" : ""}`}
            >
              <svg viewBox="0 0 24 24" aria-hidden className="size-4" fill="currentColor">
                <path d={b.path} />
              </svg>
              {oneClick ? `Add with ${h.name}` : h.name}
              {oneClick ? (
                <Zap size={14} aria-hidden />
              ) : (
                <ArrowUpRight size={14} aria-hidden />
              )}
              {h.detected && (
                <span
                  className="rounded-full px-1.5 py-0.5 text-[10px] font-medium"
                  style={{ backgroundColor: "rgba(0,0,0,.14)" }}
                >
                  Your DNS
                </span>
              )}
            </a>
          );
        })}
      </div>
    </section>
  );
}
