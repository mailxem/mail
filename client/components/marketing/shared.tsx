"use client";
import { workspaceClassName } from "@/lib/workspace-styles";
import { ReactNode } from "react";
import {
  AlertCircle,
  ArrowUpRight,
  Inbox,
  Loader2,
  MoreVertical,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetPopup,
  SheetHeader,
  SheetTitle,
  SheetDescription,
  SheetPanel,
} from "@/components/ui/sheet";
export function PageHeading({
  title,
  description,
  action,
  eyebrow,
}: {
  title: string;
  description: string;
  action?: ReactNode;
  eyebrow?: string;
}) {
  return (
    <div className={workspaceClassName("product-heading")}>
      <div>
        {eyebrow && <span className={workspaceClassName("eyebrow")}>{eyebrow}</span>}
        <h1>{title}</h1>
        <p>{description}</p>
      </div>
      {action}
    </div>
  );
}
export function Metric({
  label,
  value,
  icon,
  detail,
}: {
  label: string;
  value: string | number;
  icon: ReactNode;
  detail?: string;
}) {
  return (
    <div className={workspaceClassName("metric-card")}>
      <div className={workspaceClassName("metric-label")}>
        <span className={workspaceClassName("metric-icon")}>{icon}</span>
        {label}
      </div>
      <div className={workspaceClassName("metric-value")}>
        <strong>{value}</strong>
        {detail && <span>{detail}</span>}
      </div>
    </div>
  );
}
export function Status({ value }: { value: string }) {
  const label: Record<string, string> = {
    PUBLISHED: "Active",
    DRAFT: "Draft",
    ARCHIVED: "Paused",
    SCHEDULED: "Scheduled",
    COMPLETED: "Completed",
    ACTIVE: "Subscribed",
    UNSUBSCRIBED: "Unsubscribed",
  };
  return (
    <span className={workspaceClassName(`status-pill status-${value.toLowerCase()}`)}>
      <i />
      {label[value] || value.charAt(0) + value.slice(1).toLowerCase()}
    </span>
  );
}
export function Empty({
  title,
  description,
  action,
}: {
  title: string;
  description: string;
  action?: ReactNode;
}) {
  return (
    <div className={workspaceClassName("product-empty")}>
      <div className={workspaceClassName("empty-icon")}>
        <Inbox size={26} />
      </div>
      <h3>{title}</h3>
      <p>{description}</p>
      {action}
    </div>
  );
}
export function QueryState({
  loading,
  error,
  retry,
}: {
  loading: boolean;
  error?: Error | null;
  retry?: () => void;
}) {
  if (loading)
    return (
      <div className={workspaceClassName("product-empty")} role="status">
        <Loader2 className="animate-spin" />
        <p>Loading your workspace…</p>
      </div>
    );
  if (error)
    return (
      <div className={workspaceClassName("product-empty")} role="alert">
        <AlertCircle />
        <h3>We couldn’t load this page</h3>
        <p>{error.message}</p>
        <Button variant="outline" onClick={retry}>
          Try again
        </Button>
      </div>
    );
  return null;
}
export function Modal({
  open,
  onOpenChange,
  title,
  description,
  children,
  wide,
}: {
  open: boolean;
  onOpenChange: (v: boolean) => void;
  title: string;
  description: string;
  children: ReactNode;
  wide?: boolean;
}) {
  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetPopup variant="inset" className={wide ? "max-w-5xl" : "max-w-xl"}>
        <SheetHeader>
          <SheetTitle>{title}</SheetTitle>
          <SheetDescription>{description}</SheetDescription>
        </SheetHeader>
        <SheetPanel>{children}</SheetPanel>
      </SheetPopup>
    </Sheet>
  );
}
export function Field({
  label,
  children,
  hint,
}: {
  label: string;
  children: ReactNode;
  hint?: string;
}) {
  return (
    <label className={workspaceClassName("product-field")}>
      <span>{label}</span>
      {children}
      {hint && <small>{hint}</small>}
    </label>
  );
}
export function EmailPreview({
  html,
  title = "Email preview",
}: {
  html: string;
  title?: string;
}) {
  const policy = `<meta http-equiv="Content-Security-Policy" content="default-src 'none'; img-src https: data:; style-src 'unsafe-inline'; font-src 'none'; base-uri 'none'; form-action 'none'">`;
  return (
    <iframe
      title={title}
      sandbox=""
      referrerPolicy="no-referrer"
      srcDoc={policy + html}
      className={workspaceClassName("email-preview")}
    />
  );
}
export function FilterTabs({
  items,
  value,
  onChange,
}: {
  items: string[];
  value: string;
  onChange: (v: string) => void;
}) {
  return (
    <div className={workspaceClassName("filter-tabs")} role="tablist" aria-label="Filter results">
      {items.map((item) => (
        <button
          key={item}
          role="tab"
          aria-selected={item === value}
          tabIndex={item === value ? 0 : -1}
          onKeyDown={(event) => {
            const position = items.indexOf(item);
            const next = event.key === "ArrowRight" ? (position + 1) % items.length : event.key === "ArrowLeft" ? (position - 1 + items.length) % items.length : event.key === "Home" ? 0 : event.key === "End" ? items.length - 1 : -1;
            if (next < 0) return;
            event.preventDefault();
            onChange(items[next]);
            (event.currentTarget.parentElement?.querySelectorAll("button")[next] as HTMLButtonElement | undefined)?.focus();
          }}
          onClick={() => onChange(item)}
        >
          {item}
        </button>
      ))}
    </div>
  );
}
