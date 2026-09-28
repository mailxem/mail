import Link from "next/link";
import { ArrowUpRight } from "lucide-react";
import type { ReactNode } from "react";
import { cn } from "@/lib/utils";
import styles from "./collection-card.module.css";

/** A quiet, shared resource surface for collections and connected services. */
export function CollectionCard({ icon, badge, title, href, description, children, action, menu, onAction }: {
  icon: ReactNode;
  badge?: ReactNode;
  title: string;
  href?: string;
  onAction?: () => void;
  description?: string;
  children?: ReactNode;
  action: string;
  menu?: ReactNode;
}) {
  const textBadge = typeof badge === "string" || typeof badge === "number";
  return (
    <article className={styles.card}>
      <div className={styles.header}>
        <span className={styles.icon}>{icon}</span>
        {badge && <span className={cn(styles.badge, textBadge && styles.textBadge)}>{badge}</span>}
      </div>
      <div className={styles.body}>
        <h3>{href ? <Link href={href}>{title}</Link> : title}</h3>
        {description && <p className={styles.description}>{description}</p>}
        {children && <div className={styles.details}>{children}</div>}
      </div>
      <div className={styles.footer}>
        {href
          ? <Link href={href} className={styles.action}>{action}<ArrowUpRight size={14} strokeWidth={1.65} /></Link>
          : <button type="button" onClick={onAction} className={styles.action}>{action}<ArrowUpRight size={14} strokeWidth={1.65} /></button>}
        {menu}
      </div>
    </article>
  );
}
