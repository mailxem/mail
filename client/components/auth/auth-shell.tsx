"use client";

import { useEffect, useState, type ReactNode } from "react";
import Link from "next/link";
import { useTheme } from "next-themes";
import { Monitor, Moon, Sun } from "lucide-react";
import styles from "./auth.module.css";

export function AuthShell({ children }: { children: ReactNode }) {
  const { theme, setTheme } = useTheme();
  const [mounted, setMounted] = useState(false);
  useEffect(() => setMounted(true), []);
  return (
    <div className={styles.screen}>
      <div className={styles.appearance} role="group" aria-label="Appearance">
        {[
          { value: "light", label: "Light theme", Icon: Sun },
          { value: "dark", label: "Dark theme", Icon: Moon },
          { value: "system", label: "System theme", Icon: Monitor },
        ].map(({ value, label, Icon }) => (
          <button
            key={value}
            type="button"
            aria-label={label}
            title={label}
            aria-pressed={mounted && theme === value}
            onClick={() => setTheme(value)}
          >
            <Icon size={14} />
          </button>
        ))}
      </div>
      <main className={styles.card}>
        <div className={styles.pixels} aria-hidden="true" />
        <section className={styles.formPanel}>
          <Link href="/" aria-label="Xem home" className={styles.logo}>
            <svg viewBox="0 0 64 36" aria-hidden="true">
              <path
                fill="currentColor"
                d="M4 2h6v6H4zm12 0h6v6h-6zm36 0h6v6h-6zm-12 0h6v6h-6zM16 14h6v6h-6zm12 0h6v6h-6zm12 0h6v6h-6zM4 26h6v6H4zm12 0h6v6h-6zm36 0h6v6h-6zm-12 0h6v6h-6z"
              />
            </svg>
          </Link>
          {children}
        </section>
        <aside className={styles.featurePanel} aria-label="About Xem">
          {/* Licensed artwork, cropped and upscaled 3× (nearest-neighbour) from the original. */}
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img src="/illustrations/auth.webp" alt="" className={styles.artwork} />
          <p className={styles.tagline}>
            Email built for <em>better</em> conversations
          </p>
        </aside>
      </main>
    </div>
  );
}
