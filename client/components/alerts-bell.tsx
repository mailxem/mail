"use client";

import { useState } from "react";
import Link from "next/link";
import { Bell, CircleCheck } from "lucide-react";
import { formatDistanceToNow } from "date-fns";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { useMarketing, useMarketingQuery } from "@/lib/marketing/api";
import { cn } from "@/lib/utils";

export interface WorkspaceAlert {
  id: string;
  kind: string;
  title: string;
  reason: string;
  count: number;
  action: string;
  href: string;
  resolved: boolean;
  readAt: string | null;
  createdAt: string;
}

// The workspace's alerts: the same problems Xem emails the owner about, plus
// news such as a finished import. Each one links to where it can be fixed.
export function AlertsBell({
  className,
  label = false,
  side = "bottom",
}: {
  className?: string;
  label?: boolean;
  side?: "bottom" | "right";
}) {
  const [open, setOpen] = useState(false);
  const query = useMarketingQuery<{ alerts: WorkspaceAlert[]; unread: number }>(
    "marketing/alerts",
    true,
    60_000,
  );
  const { request } = useMarketing();
  const unread = query.data?.unread ?? 0;
  const read = (id?: string) =>
    request("marketing/alerts/read", "POST", id ? { id } : {}).then(
      () => query.refetch(),
      () => undefined,
    );

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          type="button"
          className={className}
          aria-label={unread ? `Alerts, ${unread} unread` : "Alerts"}
          title={label ? undefined : "Alerts"}
        >
          <span className="relative inline-flex">
            <Bell size={15} strokeWidth={1.65} />
            {unread > 0 && (
              <span
                aria-hidden="true"
                className="absolute -right-2 -top-2 min-w-4 rounded-full bg-red-600 px-1 text-center text-[10px] font-medium leading-4 text-white"
              >
                {unread > 9 ? "9+" : unread}
              </span>
            )}
          </span>
          {label && <span>Alerts</span>}
        </button>
      </PopoverTrigger>
      <PopoverContent
        side={side}
        align="end"
        sideOffset={8}
        collisionPadding={12}
        className="w-[360px] max-w-[calc(100vw-24px)] rounded-xl p-0"
      >
        <div className="flex items-center justify-between border-b px-4 py-3">
          <p className="text-sm font-medium">Alerts</p>
          {unread > 0 && (
            <button
              type="button"
              className="text-xs text-muted-foreground hover:text-foreground"
              onClick={() => read()}
            >
              Mark all as read
            </button>
          )}
        </div>
        <div className="max-h-[420px] overflow-y-auto">
          {query.isError ? (
            <p className="px-4 py-6 text-sm text-muted-foreground">
              Alerts couldn’t load. Try again in a moment.
            </p>
          ) : !query.data ? (
            <p className="px-4 py-6 text-sm text-muted-foreground">Loading…</p>
          ) : query.data.alerts.length === 0 ? (
            <p className="px-4 py-6 text-sm text-muted-foreground">
              Nothing needs you right now. Problems with sending, imports or
              automations will show up here.
            </p>
          ) : (
            <ul>
              {query.data.alerts.map((alert) => (
                <li
                  key={alert.id}
                  className={cn(
                    "border-b px-4 py-3 last:border-b-0",
                    !alert.readAt && "bg-muted/50",
                  )}
                >
                  <div className="flex items-start justify-between gap-3">
                    <p className="text-sm font-medium leading-5">
                      {!alert.readAt && <span className="sr-only">Unread: </span>}
                      {alert.title}
                    </p>
                    {alert.resolved && (
                      <span className="inline-flex shrink-0 items-center gap-1 rounded-full bg-emerald-50 px-2 py-0.5 text-[11px] font-medium text-emerald-700 dark:bg-emerald-950 dark:text-emerald-300">
                        <CircleCheck size={11} />
                        Resolved
                      </span>
                    )}
                  </div>
                  {alert.reason && (
                    <p className="mt-1 text-xs text-muted-foreground">{alert.reason}</p>
                  )}
                  <div className="mt-2 flex items-center justify-between gap-3 text-xs text-muted-foreground">
                    <time
                      dateTime={alert.createdAt}
                      title={new Date(alert.createdAt).toLocaleString()}
                    >
                      {formatDistanceToNow(new Date(alert.createdAt), { addSuffix: true })}
                      {alert.count > 1 && ` · ${alert.count} events`}
                    </time>
                    {alert.href && !alert.resolved && (
                      <Link
                        href={alert.href}
                        className="shrink-0 rounded-md border px-2 py-1 font-medium text-foreground hover:bg-accent"
                        onClick={() => {
                          setOpen(false);
                          if (!alert.readAt) read(alert.id);
                        }}
                      >
                        {alert.action}
                      </Link>
                    )}
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>
      </PopoverContent>
    </Popover>
  );
}
