import type { CSSProperties, ReactNode } from "react";
import { resolveFormTheme, type FormTheme } from "@/lib/marketing/form-theme";
import styles from "./form-surface.module.css";
export function FormSurface({
  theme: value,
  preview = false,
  embedded = false,
  name,
  children,
}: {
  theme?: Partial<FormTheme>;
  preview?: boolean;
  embedded?: boolean;
  name?: string;
  children: ReactNode;
}) {
  const theme = resolveFormTheme(value);
  const style = {
    "--form-bg": theme.backgroundColor,
    "--form-card": theme.cardColor,
    "--form-text": theme.textColor,
    "--form-button": theme.buttonColor,
    "--form-button-text": theme.buttonTextColor,
    "--form-radius": theme.corners === "square" ? "0px" : "22px",
    "--form-input-radius": theme.corners === "square" ? "0px" : "8px",
    "--form-font":
      theme.font === "serif"
        ? "Georgia, serif"
        : theme.font === "mono"
          ? "ui-monospace, monospace"
          : "system-ui, sans-serif",
  } as CSSProperties;
  return (
    <div data-form-surface className={preview ? styles.preview : `${styles.page}${embedded ? ` ${styles.embedded}` : ""}`} style={style}>
      <div className={styles.card}>
        {theme.logoUrl && (
          <img
            className={styles.logo}
            src={theme.logoUrl}
            alt={name ? `${name} logo` : "Form logo"}
            referrerPolicy="no-referrer"
          />
        )}
        {children}
      </div>
    </div>
  );
}
