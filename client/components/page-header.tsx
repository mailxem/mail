"use client";

import { ChevronLeft } from "lucide-react";
import Link from "next/link";
import { cn } from "@/lib/utils";
import styles from "@/components/workspace-shell.module.css";

interface PageHeaderProps {
  heading: string;
  description?: string;
  backButton?: { href: string; label: string };
  className?: string;
  children?: React.ReactNode;
}

export function PageHeader({ heading, description, backButton, className, children }: PageHeaderProps) {
  return (
    <div className={cn("workspace-page-heading", styles.pageHeading, className)}>
      <div>
        {backButton && <Link className={styles.backLink} href={backButton.href}><ChevronLeft size={14} strokeWidth={1.65} />{backButton.label}</Link>}
        <h1>{heading}</h1>
        {description && <p>{description}</p>}
      </div>
      {children && <div className={cn("workspace-heading-actions", styles.headingActions)}>{children}</div>}
    </div>
  );
}

export type { PageHeaderProps };
